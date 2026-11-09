package routers

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/services/deviceauth"
	"gorm.io/gorm"
)

func TestSetPasswordHTTPMySQL(t *testing.T) {
	if os.Getenv("SET_PASSWORD_HTTP_TEST_MYSQL") != "1" {
		t.Skip("isolated localhost:13316/set_password_http_test")
	}
	gin.SetMode(gin.TestMode)
	configures.Config.PasswordSMSPepper = strings.Repeat("t", 32)
	defer func() { configures.Config.PasswordSMSPepper = "" }()
	configures.Config.Mysql.User = "root"
	configures.Config.Mysql.Password = ""
	configures.Config.Mysql.Address = "127.0.0.1:13316"
	configures.Config.Mysql.DbName = "set_password_http_test"
	if e := dbcommons.InitMysql(); e != nil {
		t.Fatal(e)
	}
	defer dbcommons.CloseDB()
	db := dbcommons.GetDb()
	schema, e := os.ReadFile("../commons/dbcommons/sqls/20250201.sql")
	if e != nil {
		t.Fatal(e)
	}
	for _, table := range []string{"apps", "users", "qrcoderecords", "smsrecords", "appexts"} {
		start := strings.Index(string(schema), "CREATE TABLE IF NOT EXISTS `"+table+"`")
		if start < 0 {
			continue
		}
		ddl := string(schema)[start:]
		if e = db.Exec(ddl[:strings.Index(ddl, ";")+1]).Error; e != nil {
			t.Fatal(e)
		}
	}
	db.Exec("DROP TABLE IF EXISTS password_sms_challenges")
	// Quotas survive failed sends by design; reset the isolated fixture on reruns.
	if e = db.Exec("DROP TABLE IF EXISTS password_sms_quotas").Error; e != nil {
		t.Fatal(e)
	}
	for _, file := range []string{"20261103.sql", "20261109.sql", "20261110.sql"} {
		b, e := os.ReadFile("../commons/dbcommons/sqls/" + file)
		if e != nil {
			t.Fatal(e)
		}
		if e := db.Connection(func(conn *gorm.DB) error {
			for _, stmt := range strings.Split(string(b), ";") {
				if strings.TrimSpace(stmt) != "" {
					if e = conn.Exec(stmt).Error; e != nil {
						return e
					}
				}
			}
			if file == "20261109.sql" {
				return dbcommons.EnsurePasswordSMSPhoneTrigger(conn)
			}
			return nil
		}); e != nil {
			t.Fatal(e)
		}
	}
	key := "0123456789abcdef"
	db.Exec("DELETE FROM apps WHERE app_key='sethttp'")
	db.Exec("DELETE FROM users WHERE app_key='sethttp'")
	if e = db.Exec("INSERT INTO apps(app_key,app_secure_key,app_secret) VALUES ('sethttp',?,'fake-secret')", key).Error; e != nil {
		t.Fatal(e)
	}
	if e = db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('sethttp','u1','13800138000')").Error; e != nil {
		t.Fatal(e)
	}
	r := gin.New()
	Route(r, "jim")
	legacy := func(user string) string {
		token, e := (services.ImToken{AppKey: "sethttp", UserId: user, TokenTime: time.Now().UnixMilli()}).ToTokenString([]byte(key))
		if e != nil {
			t.Fatal(e)
		}
		return token
	}
	api, e := services.GenerateApiKey("sethttp", key)
	if e != nil {
		t.Fatal(e)
	}
	svc, e := deviceauth.SignService("sethttp", "u1", []byte(key), time.Now().UnixMilli())
	if e != nil {
		t.Fatal(e)
	}
	ds, e := (deviceauth.Store{DB: db}).Issue(context.Background(), "sethttp", "u1", deviceauth.DeviceParams{DeviceID: json.RawMessage(`"set-http"`), Platform: json.RawMessage(`"ios"`)}, []byte(key), time.Now().UnixMilli())
	if e != nil {
		t.Fatal(e)
	}
	call := func(path, body, token string, want int) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", path, strings.NewReader(body))
		req.Header.Set("appkey", "sethttp")
		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var out struct {
			Code int
			Data map[string]interface{}
		}
		if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil {
			t.Fatal(e)
		}
		if out.Code != want {
			t.Fatalf("%s: %s want %d", path, w.Body.String(), want)
		}
		if strings.Contains(body, `"purpose":"set_password"`) && want == 17008 {
			if value, ok := out.Data["retry_after_seconds"]; !ok || value != float64(0) {
				t.Fatalf("missing preflight retry hint: %s", w.Body.String())
			}
		}
		if (body == `{"phone":"13800138000"}` || strings.Contains(body, `"purpose":"login"`)) && out.Data != nil {
			t.Fatal("login SMS envelope changed")
		}
	}
	for _, token := range []string{"", "Bearer " + api, svc, legacy("")} {
		call("/jim/sms/send", `{"purpose":"set_password","phone":"13900139000"}`, token, 17005)
		call("/jim/users/setpass", `{"code":"000000","new_password":"abcdef"}`, token, 17005)
	}
	call("/jim/sms/send", `{"purpose":"unknown","phone":"13800138000"}`, "", 17003)
	call("/jim/sms/send", `{"phone":"13800138000"}`, "", 17008)
	call("/jim/sms/send", `{"purpose":"login","phone":"13800138000"}`, "", 17008)
	for _, token := range []string{legacy("u1"), ds} {
		call("/jim/users/setpass", `{"code":"000000","new_password":"abcdef","user_id":"someone"}`, token, 17023)
		call("/jim/users/setpass", `{"code":"123456","new_password":"     "}`, token, 17003)
	}
	c := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, "sethttp")
	if e := db.Exec("DELETE FROM smsrecords WHERE app_key='sethttp'").Error; e != nil {
		t.Fatal(e)
	}
	call("/jim/sms/login", `{"phone":"13800138000","code":"000000"}`, "", 17009)
	call("/jim/users/bindphone", `{"phone":"13900139000","code":"000000"}`, legacy("u1"), 17009)
	if e := db.Exec("INSERT INTO smsrecords(app_key,phone,code,created_time) VALUES ('sethttp','13900139000','000000',?)", time.Now().Add(-6*time.Minute)).Error; e != nil {
		t.Fatal(e)
	}
	call("/jim/users/bindphone", `{"phone":"13900139000","code":"000000"}`, legacy("u1"), 17009)
	if e := db.Exec("UPDATE smsrecords SET created_time=? WHERE app_key='sethttp' AND phone='13900139000'", time.Now()).Error; e != nil {
		t.Fatal(e)
	}
	call("/jim/users/bindphone", `{"phone":"13900139000","code":"000000"}`, legacy("u1"), 0)
	var bound string
	db.Raw("SELECT phone FROM users WHERE app_key='sethttp' AND user_id='u1'").Scan(&bound)
	if bound != "13900139000" {
		t.Fatal("normal SMS binding failed")
	}
	db.Exec("UPDATE users SET phone='13800138000' WHERE app_key='sethttp'")
	if got := services.CheckPhoneSmsCode(c, "13900139000", "000000"); got != 0 {
		t.Fatal("stored valid zero code must use normal validation")
	}
	if e := db.Exec("INSERT INTO smsrecords(app_key,phone,code,created_time) VALUES ('sethttp','13800138000','543210',?)", time.Now()).Error; e != nil {
		t.Fatal(e)
	}
	if got := services.CheckPhoneSmsCode(c, "13800138000", "543210"); got != 0 {
		t.Fatalf("legacy login SMS regressed: %d", got)
	}
	call("/jim/users/setpass", `{"code":"543210","new_password":"abcdef"}`, legacy("u1"), 17023)
	if got := services.CheckPhoneSmsCode(c, "13800138000", "123456"); got != 17009 {
		t.Fatalf("missing login SMS accepted: %d", got)
	}
	call("/jim/sms/send", `{"purpose":"set_password","phone":"13900139000"}`, legacy("u1"), 17008)
	var count int64
	if e := db.Table("password_sms_challenges").Count(&count).Error; e != nil {
		t.Fatal(e)
	}
	if count != 0 {
		t.Fatal("configuration failure created a challenge")
	}
	db.Exec("UPDATE users SET login_pass='hash' WHERE app_key='sethttp'")
	for _, token := range []string{legacy("u1"), ds} {
		call("/jim/sms/send", `{"purpose":"set_password"}`, token, 17019)
		call("/jim/users/setpass", `{"code":"000000","new_password":"abcdef"}`, token, 17019)
	}
}
