package apis

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/events"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	sm "github.com/juggleim/jugglechat-server/storages/models"
	"gorm.io/gorm"
)

func TestSMSLegacyMySQL(t *testing.T) {
	if os.Getenv("SMS_LEGACY_TEST_MYSQL") != "1" {
		t.Skip("isolated localhost:13316/sms_legacy_test")
	}
	gin.SetMode(gin.TestMode)
	configures.Config.Mysql.User = "root"
	configures.Config.Mysql.Password = ""
	configures.Config.Mysql.Address = "127.0.0.1:13316"
	configures.Config.Mysql.DbName = "sms_legacy_test"
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
	for _, table := range []string{"users", "apps", "smsrecords", "qrcoderecords", "userexts"} {
		must(db.Exec("DROP TABLE IF EXISTS " + table).Error)
		start := strings.Index(string(b), "CREATE TABLE IF NOT EXISTS `"+table+"`")
		if start < 0 {
			t.Fatal(table)
		}
		ddl := string(b)[start:]
		must(db.Exec(ddl[:strings.Index(ddl, ";")+1]).Error)
	}
	for _, file := range []string{"20261103.sql", "20261109.sql"} {
		b, e := os.ReadFile("../commons/dbcommons/sqls/" + file)
		must(e)
		must(db.Connection(func(tx *gorm.DB) error {
			for _, stmt := range strings.Split(string(b), ";") {
				if strings.TrimSpace(stmt) != "" {
					if e := tx.Exec(stmt).Error; e != nil {
						return e
					}
				}
			}
			return nil
		}))
	}
	must(dbcommons.EnsurePasswordSMSPhoneTrigger(db))
	phone := "13800138000"
	t.Run("new SMS account and NULL password", func(t *testing.T) {
		must(db.Exec("DELETE FROM users").Error)
		must(db.Exec("INSERT INTO smsrecords(app_key,phone,code,created_time) VALUES ('smslegacy',?,'654321',?)", phone, time.Now()).Error)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(string(ctxs.CtxKey_AppKey), "smslegacy")
		c.Request = httptest.NewRequest("POST", "/sms/login", strings.NewReader(`{"phone":"13800138000","code":"654321"}`))
		SmsLogin(c)
		var n int64
		must(db.Raw("SELECT COUNT(*) FROM users WHERE app_key='smslegacy' AND phone=? AND login_account=? AND login_pass IS NULL", phone, phone).Scan(&n).Error)
		if n != 1 {
			t.Fatalf("expected phone account with NULL password, found %d", n)
		}
		u, e := (dbs.UserDao{}).FindByPhone("smslegacy", phone)
		must(e)
		infoContext := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, "smslegacy")
		infoContext = context.WithValue(infoContext, ctxs.CtxKey_RequesterId, u.UserId)
		infoCode, info := services.QryUserInfo(infoContext, u.UserId)
		if infoCode != 0 || info.Account != phone {
			t.Fatalf("users/info missing default account: %d %+v", infoCode, info)
		}
		for _, password := range []string{"", "password", utils.SHA1("")} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(string(ctxs.CtxKey_AppKey), "smslegacy")
			body, _ := json.Marshal(map[string]string{"account": phone, "password": password})
			c.Request = httptest.NewRequest("POST", "/login", strings.NewReader(string(body)))
			Login(c)
			var out struct{ Code int }
			must(json.Unmarshal(w.Body.Bytes(), &out))
			want := 17013
			if password == "" {
				want = 17003
			}
			if out.Code != want {
				t.Fatalf("passwordless account password login code=%d want=%d", out.Code, want)
			}
		}
	})
	id := utils.ShortMd5(phone)
	call := func(code string) int {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(string(ctxs.CtxKey_AppKey), "smslegacy")
		c.Request = httptest.NewRequest("POST", "/sms/login", strings.NewReader(`{"phone":"`+phone+`","code":"`+code+`"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		SmsLogin(c)
		var out struct{ Code int }
		must(json.Unmarshal(w.Body.Bytes(), &out))
		return out.Code
	}
	t.Run("concurrent new phone one user and registration event", func(t *testing.T) {
		must(db.Exec("DELETE FROM users").Error)
		must(db.Exec("DELETE FROM smsrecords").Error)
		must(db.Exec("DELETE FROM userexts").Error)
		must(db.Exec("INSERT INTO smsrecords(app_key,phone,code,created_time) VALUES ('smslegacy',?,'654321',?)", phone, time.Now()).Error)
		var count atomic.Int32
		var eventUser atomic.Value
		events.RegisteUserRegisteEvent(func(app string, u sm.User) {
			if app == "smslegacy" {
				count.Add(1)
				eventUser.Store(u)
			}
		})
		var wg sync.WaitGroup
		results := make(chan int, 16)
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); results <- call("654321") }()
		}
		wg.Wait()
		close(results)
		for code := range results {
			if code != 17002 {
				t.Fatalf("login failed before local absent SDK: %d", code)
			}
		}
		var n int64
		must(db.Raw("SELECT COUNT(*) FROM users WHERE app_key='smslegacy'").Scan(&n).Error)
		if n != 1 || count.Load() != 1 {
			t.Fatalf("users=%d events=%d", n, count.Load())
		}
		u, e := (dbs.UserDao{}).FindByPhone("smslegacy", phone)
		must(e)
		if eventUser.Load().(sm.User).UserId != u.UserId || eventUser.Load().(sm.User).LoginAccount != phone {
			t.Fatal("event did not use actual new user")
		}
		must(db.Raw("SELECT COUNT(*) FROM userexts WHERE app_key='smslegacy' AND user_id=?", u.UserId).Scan(&n).Error)
		if n != 1 {
			t.Fatalf("default settings rows %d", n)
		}
	})
	for _, legacy := range []bool{false, true} {
		t.Run("empty old account preserved "+fmtBound(legacy), func(t *testing.T) {
			must(db.Exec("DELETE FROM users").Error)
			uid := "old-owner"
			var bound interface{} = phone
			if legacy {
				uid = utils.ShortMd5(phone)
				bound = nil
			}
			must(db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('smslegacy',?,?)", uid, bound).Error)
			if code := call("654321"); code != 17002 {
				t.Fatal(code)
			}
			var n int64
			must(db.Raw("SELECT COUNT(*) FROM users WHERE app_key='smslegacy' AND user_id=? AND phone=? AND login_account IS NULL AND login_pass IS NULL", uid, phone).Scan(&n).Error)
			if n != 1 {
				t.Fatal("old/legacy account or password changed")
			}
		})
	}
	reset := func(bound interface{}) {
		must(db.Exec("DELETE FROM user_devices").Error)
		must(db.Exec("DELETE FROM users").Error)
		must(db.Exec("DELETE FROM smsrecords").Error)
		must(db.Exec("INSERT INTO users(app_key,user_id,phone,nickname,login_account,login_pass) VALUES ('smslegacy',?,?,'Keep Nick','keep_account','keep_hash')", id, bound).Error)
		must(db.Exec("INSERT INTO smsrecords(app_key,phone,code,created_time) VALUES ('smslegacy',?,'654321',?)", phone, time.Now()).Error)
		must(db.Exec("INSERT INTO user_devices(app_key,user_id,device_id,platform,last_login_time,session_id,state,created_time,updated_time) VALUES ('smslegacy',?,'existing','ios',1,'keep-session','active',1,1)", id).Error)
	}
	for _, bound := range []interface{}{nil, "", phone} {
		t.Run("backfill_"+fmtBound(bound), func(t *testing.T) {
			reset(bound)
			if got := call("654321"); got != 17002 {
				t.Fatalf("expected stop at absent local IM app, got %d", got)
			}
			var row struct {
				Phone, Nickname, LoginAccount, LoginPass string
				PasswordSmsPhoneVersion                  int64
			}
			must(db.Raw("SELECT * FROM users WHERE app_key='smslegacy' AND user_id=?", id).Scan(&row).Error)
			if row.Phone != phone {
				t.Fatalf("verified legacy phone not backfilled: %q", row.Phone)
			}
			resolved, code := services.ResolveVerifiedSMSUser("smslegacy", phone)
			if code != 0 || resolved == nil || resolved.Phone != phone || resolved.UserId != id {
				t.Fatalf("resolved stale user: %d %+v", code, resolved)
			}
			if row.Nickname != "Keep Nick" || row.LoginAccount != "keep_account" || row.LoginPass != "keep_hash" {
				t.Fatalf("profile changed: %+v", row)
			}
			var session string
			must(db.Raw("SELECT session_id FROM user_devices WHERE app_key='smslegacy' AND user_id=?", id).Scan(&session).Error)
			if session != "keep-session" {
				t.Fatal("session changed")
			}
			want := int64(1)
			if bound == phone {
				want = 0
			}
			if row.PasswordSmsPhoneVersion != want {
				t.Fatalf("version %d want %d", row.PasswordSmsPhoneVersion, want)
			}
		})
	}
	t.Run("different phone rejects", func(t *testing.T) {
		reset("13900139000")
		if got := call("654321"); got != 17028 {
			t.Fatalf("got %d want phone conflict 17028", got)
		}
	})
	t.Run("invalid SMS never backfills", func(t *testing.T) {
		reset(nil)
		if got := call("000000"); got != 17009 {
			t.Fatal(got)
		}
		var n int64
		must(db.Raw("SELECT COUNT(*) FROM users WHERE phone IS NULL").Scan(&n).Error)
		if n != 1 {
			t.Fatal("mutated without SMS")
		}
	})
	for _, failPhone := range []bool{true, false} {
		t.Run("DB failure_"+fmtBound(failPhone), func(t *testing.T) {
			reset(nil)
			name := "test:sms_query_failure"
			queries := 0
			must(db.Callback().Query().Before("gorm:query").Register(name, func(tx *gorm.DB) {
				if tx.Statement.Table == "users" {
					queries++
					if failPhone || queries == 2 {
						tx.AddError(errors.New("injected DB failure"))
					}
				}
			}))
			defer db.Callback().Query().Remove(name)
			if got := call("654321"); got != 17000 {
				t.Fatalf("DB fault continued login: %d", got)
			}
		})
	}
	t.Run("password status after login", func(t *testing.T) {
		reset(nil)
		must(db.Exec("UPDATE users SET login_pass=NULL").Error)
		call("654321")
		c := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, "smslegacy")
		c = context.WithValue(c, ctxs.CtxKey_RequesterId, id)
		c = context.WithValue(c, ctxs.CtxKey_AuthKind, "legacy")
		code, status := services.QryPasswordStatus(c)
		if code != 0 || status == nil || status.HasPassword || !status.CanSetPasswordBySMS {
			t.Fatalf("status %d %+v", code, status)
		}
	})
	t.Run("phone owner wins over unbound legacy", func(t *testing.T) {
		reset(nil)
		must(db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('smslegacy','owner',?)", phone).Error)
		if got := call("654321"); got != 17002 {
			t.Fatalf("phone owner login blocked: %d", got)
		}
		u, code := services.ResolveVerifiedSMSUser("smslegacy", phone)
		if code != 0 || u == nil || u.UserId != "owner" || u.Phone != phone {
			t.Fatalf("wrong owner: %d %+v", code, u)
		}
		var n int64
		must(db.Raw("SELECT COUNT(*) FROM users WHERE app_key='smslegacy' AND user_id=? AND phone IS NULL", id).Scan(&n).Error)
		if n != 1 {
			t.Fatal("unbound legacy modified")
		}
		u, e := (dbs.UserDao{}).BindSMSLegacyPhone("smslegacy", id, phone)
		if u != nil || !errors.Is(e, sm.ErrSMSPhoneConflict) {
			t.Fatalf("%+v %v", u, e)
		}
	})
	t.Run("recycled old phone logs in current owner", func(t *testing.T) {
		reset("13900139000")
		must(db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('smslegacy','owner',?)", phone).Error)
		if got := call("654321"); got != 17002 {
			t.Fatalf("current owner login blocked: %d", got)
		}
		u, code := services.ResolveVerifiedSMSUser("smslegacy", phone)
		if code != 0 || u == nil || u.UserId != "owner" {
			t.Fatalf("wrong owner: %d %+v", code, u)
		}
		var bound string
		must(db.Raw("SELECT phone FROM users WHERE app_key='smslegacy' AND user_id=?", id).Scan(&bound).Error)
		if bound != "13900139000" {
			t.Fatalf("legacy binding changed: %q", bound)
		}
	})
	t.Run("phone hit skips unrelated legacy DB failure", func(t *testing.T) {
		reset(nil)
		must(db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('smslegacy','owner',?)", phone).Error)
		name := "test:sms_no_legacy_query"
		queries := 0
		must(db.Callback().Query().Before("gorm:query").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "users" {
				queries++
				if queries > 1 {
					tx.AddError(errors.New("legacy query unavailable"))
				}
			}
		}))
		defer db.Callback().Query().Remove(name)
		u, code := services.ResolveVerifiedSMSUser("smslegacy", phone)
		if code != 0 || u == nil || u.UserId != "owner" {
			t.Fatalf("unrelated legacy query blocked owner: %d %+v", code, u)
		}
		if queries != 1 {
			t.Fatalf("queried users %d times", queries)
		}
	})
	t.Run("cross app isolation", func(t *testing.T) {
		reset(nil)
		must(db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('other',?,?)", id, phone).Error)
		u, e := (dbs.UserDao{}).BindSMSLegacyPhone("smslegacy", id, phone)
		must(e)
		if u.AppKey != "smslegacy" || u.Phone != phone {
			t.Fatalf("%+v", u)
		}
		u, e = (dbs.UserDao{}).BindSMSLegacyPhone("absent", id, phone)
		if u != nil || !errors.Is(e, sm.ErrSMSLegacyMissing) {
			t.Fatalf("cross app: %+v %v", u, e)
		}
	})
	t.Run("concurrent change wins before locked read", func(t *testing.T) {
		reset(nil)
		tx := db.Begin()
		must(tx.Error)
		defer tx.Rollback()
		must(tx.Exec("UPDATE users SET phone='13900139000' WHERE app_key='smslegacy' AND user_id=?", id).Error)
		done := make(chan error, 1)
		go func() { _, e := (dbs.UserDao{}).BindSMSLegacyPhone("smslegacy", id, phone); done <- e }()
		must(tx.Commit().Error)
		if e := <-done; !errors.Is(e, sm.ErrSMSPhoneConflict) {
			t.Fatalf("stale bind accepted: %v", e)
		}
	})
	t.Run("concurrent claims unique index", func(t *testing.T) {
		reset(nil)
		must(db.Exec("INSERT INTO users(app_key,user_id) VALUES ('smslegacy','competitor')").Error)
		var wg sync.WaitGroup
		results := make(chan error, 2)
		for _, uid := range []string{id, "competitor"} {
			wg.Add(1)
			go func(uid string) {
				defer wg.Done()
				_, e := (dbs.UserDao{}).BindSMSLegacyPhone("smslegacy", uid, phone)
				results <- e
			}(uid)
		}
		wg.Wait()
		close(results)
		success := 0
		for e := range results {
			if e == nil {
				success++
			}
		}
		if success != 1 {
			t.Fatalf("%d successful claims", success)
		}
		var n int64
		must(db.Raw("SELECT COUNT(*) FROM users WHERE app_key='smslegacy' AND phone=?", phone).Scan(&n).Error)
		if n != 1 {
			t.Fatal(n)
		}
	})
	t.Run("unique error never returns user", func(t *testing.T) {
		reset(nil)
		name := "test:sms_unique"
		must(db.Callback().Update().Before("gorm:update").Register(name, func(tx *gorm.DB) { tx.AddError(&mysqlerr.MySQLError{Number: 1062, Message: "duplicate phone"}) }))
		defer db.Callback().Update().Remove(name)
		if u, code := services.ResolveVerifiedSMSUser("smslegacy", phone); u != nil || code != 17028 {
			t.Fatalf("%+v %d", u, code)
		}
	})
	t.Run("missing legacy keeps create path", func(t *testing.T) {
		reset(nil)
		must(db.Exec("DELETE FROM users").Error)
		u, code := services.ResolveVerifiedSMSUser("smslegacy", phone)
		if u != nil || code != 0 {
			t.Fatalf("%+v %d", u, code)
		}
		if got := call("654321"); got != 17002 {
			t.Fatal(got)
		}
		var row struct{ UserId, Phone string }
		must(db.Raw("SELECT user_id,phone FROM users WHERE app_key='smslegacy'").Scan(&row).Error)
		if row.Phone != phone || row.UserId == id || row.UserId == "" {
			t.Fatalf("new user changed: %+v", row)
		}
	})
}

func fmtBound(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	if v == "" {
		return "empty"
	}
	if v == true {
		return "phone"
	}
	if v == false {
		return "userid"
	}
	return "same"
}
