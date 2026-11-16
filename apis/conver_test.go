package apis

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
)

func TestConverConfigRejectsMalformedHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{
		`{"target_id":"bob","conver_type":4294967297,"confs":{"forbid_forward":true}}`,
		`{"target_id":"bob","conver_type":1,"confs":{"forbid_forward":"true"}}`,
		`{"target_id":"bob","conver_type":1,"confs":{"msg_life_time":9223372036854775808}}`,
		`{"target_id":"bob","conver_type":1,"confs":{"forbid_forward":true},"unknown":1}`,
		`{"target_id":"bob","conver_type":1,"confs":{"unknown":false}}`,
		`{"target_id":"bob","conver_type":1,"confs":{"msg_life_time":1.5}}`,
		`{"target_id":"bob","conver_type":1,"confs":{"watermark_enabled":null}}`,
		`{"target_id":"bob","conver_type":1,"confs":{"content_protection":1}}`,
		`{"target_id":"bob","conver_type":1,"confs":{"forbid_forward":true}} {}`,
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/jim/converconfs/set", bytes.NewBufferString(body))
		SetConverConfs(c)
		var ret struct{ Code int }
		if e := json.Unmarshal(w.Body.Bytes(), &ret); e != nil {
			t.Fatal(e)
		}
		if ret.Code != 17003 {
			t.Errorf("%s: %s", body, w.Body.String())
		}
	}
	for _, typ := range []string{"4294967297", "3", "0", "-1", "1.0", ""} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/jim/converconfs/get?target_id=bob&conver_type="+typ, nil)
		GetConverConfs(c)
		var ret struct{ Code int }
		json.Unmarshal(w.Body.Bytes(), &ret)
		if ret.Code != 17003 {
			t.Fatal(typ, w.Body.String())
		}
	}
}

func TestConverConfigHTTPMySQL(t *testing.T) {
	if os.Getenv("CONVERCONF_TEST_MYSQL") != "1" {
		t.Skip("isolated localhost:13316/converconf_http_test")
	}
	old := configures.Config.Mysql
	defer func() { configures.Config.Mysql = old }()
	configures.Config.Mysql.User = "root"
	configures.Config.Mysql.Password = ""
	configures.Config.Mysql.Address = "127.0.0.1:13316"
	configures.Config.Mysql.DbName = "converconf_http_test"
	if e := dbcommons.InitMysql(); e != nil {
		t.Fatal(e)
	}
	defer dbcommons.CloseDB()
	db := dbcommons.GetDb()
	for _, sql := range []string{
		"DROP TABLE IF EXISTS users", "DROP TABLE IF EXISTS converconfs",
		"CREATE TABLE users(id bigint PRIMARY KEY AUTO_INCREMENT,app_key varchar(20),user_id varchar(32)) ENGINE=InnoDB",
		"CREATE TABLE converconfs(id bigint PRIMARY KEY AUTO_INCREMENT,app_key varchar(20),conver_id varchar(100),conver_type int,sub_channel varchar(32),item_key varchar(100),item_value varchar(2000),item_type int,UNIQUE KEY(app_key,conver_id,conver_type,sub_channel,item_key)) ENGINE=InnoDB",
		"INSERT INTO users(app_key,user_id) VALUES ('http','alice'),('http','bob')",
	} {
		if e := db.Exec(sql).Error; e != nil {
			t.Fatal(e)
		}
	}
	if e := db.Exec("DROP TABLE IF EXISTS converconf_outbox, converconf_scope_versions").Error; e != nil {
		t.Fatal(e)
	}
	migration, e := os.ReadFile("../commons/dbcommons/sqls/20261111.sql")
	if e != nil {
		t.Fatal(e)
	}
	for _, statement := range strings.Split(string(migration), ";") {
		if strings.TrimSpace(statement) != "" {
			if e := db.Exec(statement).Error; e != nil {
				t.Fatal(e)
			}
		}
	}
	if e := db.Exec("DELETE FROM converconf_outbox").Error; e != nil {
		t.Fatal(e)
	}
	if e := db.Exec("DELETE FROM converconf_scope_versions").Error; e != nil {
		t.Fatal(e)
	}
	request := func(method, url, body, kind, user string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(method, url, bytes.NewBufferString(body))
		c.Set(string(ctxs.CtxKey_AppKey), "http")
		c.Set(string(ctxs.CtxKey_RequesterId), user)
		c.Set(string(ctxs.CtxKey_AuthKind), kind)
		if method == "POST" {
			SetConverConfs(c)
		} else {
			GetConverConfs(c)
		}
		return w
	}
	w := request("POST", "/jim/converconfs/set", `{"target_id":"bob","conver_type":1,"confs":{"msg_life_time":9223372036854775807,"forbid_forward":false}}`, "legacy", "alice")
	if !bytes.Contains(w.Body.Bytes(), []byte(`"code":0`)) {
		t.Fatal(w.Body.String())
	}
	w = request("GET", "/jim/converconfs/get?target_id=alice&conver_type=1", "", "user_session", "bob")
	if !bytes.Contains(w.Body.Bytes(), []byte(`"msg_life_time":9223372036854775807`)) || !bytes.Contains(w.Body.Bytes(), []byte(`"forbid_forward":false`)) {
		t.Fatal("HTTP precision/default lost", w.Body.String())
	}
	for _, tc := range []struct{ kind, user string }{{"api_key", ""}, {"service", "alice"}} {
		w = request("POST", "/jim/converconfs/set", `{"target_id":"bob","conver_type":1,"confs":{"forbid_forward":true}}`, tc.kind, tc.user)
		if !bytes.Contains(w.Body.Bytes(), []byte(`"code":17005`)) {
			t.Fatal(w.Body.String())
		}
	}
}
