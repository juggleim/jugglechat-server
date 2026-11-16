package services

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
)

func TestConversationConfigMySQL(t *testing.T) {
	if os.Getenv("CONVERCONF_TEST_MYSQL") != "1" {
		t.Skip("isolated localhost:13316/converconf_test")
	}
	old := configures.Config.Mysql
	defer func() { configures.Config.Mysql = old }()
	configures.Config.Mysql.User = "root"
	configures.Config.Mysql.Password = ""
	configures.Config.Mysql.Address = "127.0.0.1:13316"
	configures.Config.Mysql.DbName = "converconf_test"
	if e := dbcommons.InitMysql(); e != nil {
		t.Fatal(e)
	}
	defer dbcommons.CloseDB()
	db := dbcommons.GetDb()
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	b, e := os.ReadFile("../commons/dbcommons/sqls/20250201.sql")
	must(e)
	must(db.Exec("DROP TABLE IF EXISTS converconf_outbox, converconf_scope_versions").Error)
	for _, table := range []string{"users", "groupinfos", "groupmembers", "converconfs"} {
		must(db.Exec("DROP TABLE IF EXISTS " + table).Error)
		ddl := string(b)[strings.Index(string(b), "CREATE TABLE IF NOT EXISTS `"+table+"`"):]
		must(db.Exec(ddl[:strings.Index(ddl, ";")+1]).Error)
	}
	migration, e := os.ReadFile("../commons/dbcommons/sqls/20261111.sql")
	must(e)
	for _, statement := range strings.Split(string(migration), ";") {
		if strings.TrimSpace(statement) != "" {
			must(db.Exec(statement).Error)
		}
	}
	must(db.Exec("DELETE FROM converconf_outbox").Error)
	must(db.Exec("DELETE FROM converconf_scope_versions").Error)
	must(db.Exec("INSERT INTO users(app_key,user_id) VALUES ('a','alice'),('a','bob'),('a','owner'),('a','admin'),('a','member'),('a','outsider'),('b','alien')").Error)
	must(db.Exec("INSERT INTO groupinfos(app_key,group_id,creator_id) VALUES ('a','g','owner')").Error)
	must(db.Exec("INSERT INTO groupmembers(app_key,group_id,member_id,member_type) VALUES ('a','g','owner',1),('a','g','admin',2),('a','g','member',0)").Error)
	c := func(app, user, kind string) context.Context {
		ctx := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, app)
		ctx = context.WithValue(ctx, ctxs.CtxKey_RequesterId, user)
		return context.WithValue(ctx, ctxs.CtxKey_AuthKind, kind)
	}
	alice := c("a", "alice", "legacy")
	bob := c("a", "bob", "user_session")
	keys := []string{"forbid_forward", "content_protection", "watermark_enabled"}
	t.Run("defaults and old absent lifetime", func(t *testing.T) {
		code, v := GetConverConfItems(alice, "bob", "", 1)
		if code != 0 {
			t.Fatal(code)
		}
		for _, k := range keys {
			if v[k] != false {
				t.Errorf("%s=%v want false", k, v[k])
			}
		}
		if _, ok := v["msg_life_time"]; ok {
			t.Fatal("absent legacy field added")
		}
		meta, ok := v["_config_meta"].(map[string]interface{})
		if !ok || meta["version"] != int64(0) {
			t.Fatal("missing authoritative GET version", v)
		}
	})
	t.Run("private bidirectional true false and scope", func(t *testing.T) {
		for _, k := range keys {
			for _, v := range []bool{true, false} {
				if code := SetConverConfItem(alice, "bob", "", 1, map[string]interface{}{k: v}); code != 0 {
					t.Fatal(code)
				}
				code, got := GetConverConfItems(bob, "alice", "", 1)
				if code != 0 || got[k] != v {
					t.Fatalf("reverse %s: %d %+v", k, code, got)
				}
				if code := SetConverConfItem(bob, "alice", "", 1, map[string]interface{}{k: !v}); code != 0 {
					t.Fatal(code)
				}
				_, got = GetConverConfItems(alice, "bob", "", 1)
				if got[k] != !v {
					t.Fatal(got)
				}
			}
		}
		if code := SetConverConfItem(alice, "bob", "thread", 1, map[string]interface{}{"forbid_forward": false}); code != 0 {
			t.Fatal(code)
		}
		_, v := GetConverConfItems(bob, "alice", "thread", 1)
		if v["forbid_forward"] != false {
			t.Fatal(v)
		}
		_, v = GetConverConfItems(alice, "bob", "", 1)
		if v["forbid_forward"] != true {
			t.Fatal("subchannel scope lost", v)
		}
	})
	t.Run("all current group members may read write", func(t *testing.T) {
		for _, user := range []string{"owner", "admin", "member"} {
			for _, k := range keys {
				if code := SetConverConfItem(c("a", user, "legacy"), "g", "", 2, map[string]interface{}{k: true}); code != 0 {
					t.Fatalf("%s %s %d", user, k, code)
				}
				code, v := GetConverConfItems(c("a", "member", "legacy"), "g", "", 2)
				if code != 0 || v[k] != true {
					t.Fatal(code, v)
				}
			}
		}
	})
	t.Run("independent patch and application isolation", func(t *testing.T) {
		must(db.Exec("INSERT INTO users(app_key,user_id) VALUES ('b','alice'),('b','bob')").Error)
		if code := SetConverConfItem(alice, "bob", "isolation", 1, map[string]interface{}{"forbid_forward": true, "content_protection": true, "watermark_enabled": true}); code != 0 {
			t.Fatal(code)
		}
		if code := SetConverConfItem(bob, "alice", "isolation", 1, map[string]interface{}{"content_protection": false}); code != 0 {
			t.Fatal(code)
		}
		code, v := GetConverConfItems(alice, "bob", "isolation", 1)
		if code != 0 || v["forbid_forward"] != true || v["content_protection"] != false || v["watermark_enabled"] != true {
			t.Fatal(code, v)
		}
		code, v = GetConverConfItems(c("b", "alice", "legacy"), "bob", "isolation", 1)
		if code != 0 || v["forbid_forward"] != false || v["watermark_enabled"] != false {
			t.Fatal("app scope leak", code, v)
		}
	})
	t.Run("identity membership target type isolation", func(t *testing.T) {
		for _, tc := range []struct {
			ctx    context.Context
			target string
			typ    int32
		}{
			{c("a", "outsider", "legacy"), "g", 2}, {c("b", "alien", "legacy"), "g", 2},
			{c("a", "", "api_key"), "bob", 1}, {c("a", "ghost", "legacy"), "bob", 1},
			{alice, "alien", 1}, {alice, "missing", 1}, {alice, "missing", 2}, {alice, "bob", 3}, {alice, "", 1},
			{c("a", "alice", ""), "bob", 1},
		} {
			if code := SetConverConfItem(tc.ctx, tc.target, "", tc.typ, map[string]interface{}{"forbid_forward": true}); code == 0 {
				t.Error("unauthorized set", tc)
			}
			if code, _ := GetConverConfItems(tc.ctx, tc.target, "", tc.typ); code == 0 {
				t.Error("unauthorized get", tc)
			}
		}
		svc := c("a", "alice", "service")
		if code, _ := GetConverConfItems(svc, "bob", "", 1); code != 0 {
			t.Fatal("subject read", code)
		}
		if code := SetConverConfItem(svc, "bob", "", 1, map[string]interface{}{"forbid_forward": true}); code == 0 {
			t.Fatal("service wrote")
		}
	})
	t.Run("strict whole patch validation", func(t *testing.T) {
		for _, k := range keys {
			for _, value := range []interface{}{"true", "false", 0, 1, float64(0), nil, []interface{}{}, map[string]interface{}{}} {
				if code := SetConverConfItem(alice, "bob", "", 1, map[string]interface{}{k: value}); code != 17003 {
					t.Errorf("%s %#v code=%d", k, value, code)
				}
			}
		}
		for _, patch := range []map[string]interface{}{nil, {}, {"unknown": true}, {"msg_life_time": json.Number("1.5")}, {"msg_life_time": json.Number("9223372036854775808")}, {"msg_life_time": float64(9223372036854775808)}, {"msg_life_time": -1}, {"msg_life_time": nil}, {"msg_life_time": "1"}} {
			if code := SetConverConfItem(alice, "bob", "", 1, patch); code != 17003 {
				t.Errorf("%+v code=%d", patch, code)
			}
		}
		must(db.Exec("DELETE FROM converconfs").Error)
		if code := SetConverConfItem(alice, "bob", "", 1, map[string]interface{}{"msg_life_time": json.Number("42"), "forbid_forward": true, "unknown": false}); code != 17003 {
			t.Fatal(code)
		}
		var n int64
		must(db.Table("converconfs").Count(&n).Error)
		if n != 0 {
			t.Fatal("partial write", n)
		}
	})
	t.Run("exact lifetime no new menu restriction", func(t *testing.T) {
		for _, value := range []interface{}{json.Number("9223372036854775807"), int64(0), float64(42)} {
			if code := SetConverConfItem(alice, "bob", "", 1, map[string]interface{}{"msg_life_time": value, "content_protection": false}); code != 0 {
				t.Fatal(code)
			}
			code, v := GetConverConfItems(bob, "alice", "", 1)
			if code != 0 || v["content_protection"] != false {
				t.Fatal(code, v)
			}
			if value == json.Number("9223372036854775807") && v["msg_life_time"] != int64(9223372036854775807) {
				t.Fatal("precision lost", v)
			}
		}
	})
	t.Run("failed multi-key SQL rolls back and reads fail", func(t *testing.T) {
		must(db.Exec("DELETE FROM converconfs").Error)
		must(db.Exec("CREATE TRIGGER reject_config BEFORE INSERT ON converconfs FOR EACH ROW BEGIN IF NEW.item_key='watermark_enabled' THEN SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='fixture'; END IF; END").Error)
		code := SetConverConfItem(alice, "bob", "", 1, map[string]interface{}{"forbid_forward": true, "watermark_enabled": true, "msg_life_time": 42})
		if code != 17000 {
			t.Fatal("DB false success", code)
		}
		var n int64
		must(db.Table("converconfs").Count(&n).Error)
		if n != 0 {
			t.Fatal("not atomic", n)
		}
		must(db.Exec("DROP TRIGGER reject_config").Error)
		must(db.Exec("RENAME TABLE converconfs TO converconfs_unavailable").Error)
		defer db.Exec("RENAME TABLE converconfs_unavailable TO converconfs")
		if code, _ := GetConverConfItems(alice, "bob", "", 1); code != 17000 {
			t.Fatal("read false success", code)
		}
		if code := SetConverConfItem(alice, "bob", "", 1, map[string]interface{}{"forbid_forward": false}); code != 17000 {
			t.Fatal("write false success", code)
		}
	})
}
