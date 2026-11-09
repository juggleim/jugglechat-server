package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/services/deviceauth"
	"gorm.io/gorm"
)

// Exercises the registered route, credential verification and isolated real MySQL.
func TestPasswordStatusMySQL(t *testing.T) {
	if os.Getenv("PASSWORD_STATUS_TEST_MYSQL") != "1" {
		t.Skip("requires isolated localhost:13316/password_status_test")
	}
	gin.SetMode(gin.TestMode)
	configures.Config.Mysql.User = "root"
	configures.Config.Mysql.Password = ""
	configures.Config.Mysql.Address = "127.0.0.1:13316"
	configures.Config.Mysql.DbName = "password_status_test"
	if err := dbcommons.InitMysql(); err != nil {
		t.Fatal(err)
	}
	db := dbcommons.GetDb()
	defer dbcommons.CloseDB()
	schema, err := os.ReadFile("../commons/dbcommons/sqls/20250201.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"apps", "users", "qrcoderecords"} {
		if err := db.Exec("DROP TABLE IF EXISTS " + table).Error; err != nil {
			t.Fatal(err)
		}
		start := strings.Index(string(schema), "CREATE TABLE IF NOT EXISTS `"+table+"`")
		if start < 0 {
			t.Fatal("missing schema", table)
		}
		ddl := string(schema)[start:]
		if err := db.Exec(ddl[:strings.Index(ddl, ";")+1]).Error; err != nil {
			t.Fatal(err)
		}
	}
	migration, err := os.ReadFile("../commons/dbcommons/sqls/20261103.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range strings.Split(string(migration), ";") {
		if strings.TrimSpace(stmt) != "" {
			if err := db.Exec(stmt).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	key := "0123456789abcdef"
	for _, app := range []string{"passapp", "passother"} {
		if err := db.Exec("INSERT INTO apps(app_key,app_secure_key,app_secret) VALUES (?,?,'test-secret')", app, key).Error; err != nil {
			t.Fatal(err)
		}
	}
	r := gin.New()
	Route(r, "jim")
	now := time.Now().UnixMilli()
	legacy := func(app, user string) string {
		token, err := (services.ImToken{AppKey: app, UserId: user, TokenTime: now}).ToTokenString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	request := func(t *testing.T, app, token, query string, want int, has, sms bool) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/jim/users/password/status"+query, nil)
		req.Header.Set("appkey", app)
		req.Header.Set("Authorization", token)
		r.ServeHTTP(w, req)
		var body struct {
			Code int
			Msg  string
			Data map[string]interface{}
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("HTTP %d: %s", w.Code, w.Body.String())
		}
		if body.Code != want {
			t.Fatalf("code=%d want=%d body=%s", body.Code, want, w.Body.String())
		}
		if want == 0 {
			if w.Code != 200 || body.Msg != "success" || len(body.Data) != 2 || body.Data["has_password"] != has || body.Data["can_set_password_by_sms"] != sms {
				t.Fatalf("unexpected status envelope: %s", w.Body.String())
			}
		} else if len(body.Data) != 0 {
			t.Fatalf("error leaked data: %s", w.Body.String())
		}
	}
	// Literal expectations catch NULL/empty handling, trimmed hashes and false SMS eligibility.
	for _, tc := range []struct {
		name        string
		pass, phone interface{}
		has, sms    bool
	}{
		{"null", nil, nil, false, false}, {"empty", "", "", false, false},
		{"null with phone", nil, "13800138000", false, true}, {"empty with phone", "", "13800138000", false, true},
		{"password without phone", "opaque-hash", "", true, false}, {"password with phone", "opaque-hash", "13800138000", true, false},
		{"whitespace password is nonempty", " ", "13800138000", true, false},
		{"invalid phone", "", "not-a-phone", false, false}, {"phone whitespace", "", " 13800138000", false, false},
		{"international phone", "", "+14155552671", false, true}, {"short phone", "", "123", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := db.Exec("DELETE FROM users WHERE app_key='passapp'").Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("INSERT INTO users(app_key,user_id,login_pass,phone) VALUES ('passapp','u1',?,?)", tc.pass, tc.phone).Error; err != nil {
				t.Fatal(err)
			}
			request(t, "passapp", legacy("passapp", "u1"), "?user_id=someone-else", 0, tc.has, tc.sms)
		})
	}
	if err := db.Exec("INSERT INTO users(app_key,user_id,login_pass,phone) VALUES ('passother','u1','hash','13800138000')").Error; err != nil {
		t.Fatal(err)
	}
	t.Run("app and requester isolation", func(t *testing.T) {
		request(t, "passother", legacy("passother", "u1"), "", 0, true, false)
		request(t, "passapp", legacy("passapp", "absent"), "?user_id=u1", 17012, false, false)
		request(t, "passother", legacy("passother", "absent"), "", 17012, false, false)
	})
	t.Run("minimal projection", func(t *testing.T) {
		name := "test:password_projection"
		if err := db.Callback().Query().Before("gorm:query").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "users" && (len(tx.Statement.Selects) != 2 || tx.Statement.Selects[0] != "login_pass" || tx.Statement.Selects[1] != "phone") {
				tx.AddError(errors.New("full user read forbidden"))
			}
		}); err != nil {
			t.Fatal(err)
		}
		defer db.Callback().Query().Remove(name)
		request(t, "passother", legacy("passother", "u1"), "", 0, true, false)
	})
	t.Run("database failure", func(t *testing.T) {
		name := "test:password_db_failure"
		if err := db.Callback().Query().Before("gorm:query").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "users" {
				tx.AddError(errors.New("database unavailable"))
			}
		}); err != nil {
			t.Fatal(err)
		}
		defer db.Callback().Query().Remove(name)
		request(t, "passother", legacy("passother", "u1"), "", 17000, false, false)
	})
	t.Run("credential kinds", func(t *testing.T) {
		svc, err := deviceauth.SignService("passapp", "u1", []byte(key), now)
		if err != nil {
			t.Fatal(err)
		}
		apiKey, err := services.GenerateApiKey("passapp", key)
		if err != nil {
			t.Fatal(err)
		}
		store := deviceauth.Store{DB: db}
		ds, err := store.Issue(context.Background(), "passapp", "u1", deviceauth.DeviceParams{DeviceID: json.RawMessage(`"password-device"`), Platform: json.RawMessage(`"ios"`)}, []byte(key), now)
		if err != nil {
			t.Fatal(err)
		}
		for _, token := range []string{"", "Bearer " + apiKey, svc, legacy("passapp", "")} {
			request(t, "passapp", token, "", 17005, false, false)
		}
		request(t, "passapp", ds, "", 0, false, false)
		request(t, "passapp", legacy("passapp", "u1"), "", 0, false, false)
		request(t, "passother", legacy("passapp", "u1"), "", 17005, false, false)
	})
}
