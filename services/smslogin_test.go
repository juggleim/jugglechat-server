package services

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"
	"gorm.io/gorm"
)

func TestSMSLegacyPasswordMySQL(t *testing.T) {
	if os.Getenv("SMS_LEGACY_TEST_MYSQL") != "1" {
		t.Skip("isolated localhost:13316/sms_legacy_service_test")
	}
	configures.Config.Mysql.User = "root"
	configures.Config.Mysql.Password = ""
	configures.Config.Mysql.Address = "127.0.0.1:13316"
	configures.Config.Mysql.DbName = "sms_legacy_service_test"
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
	must(db.Exec("DROP TABLE IF EXISTS users").Error)
	ddl := string(b)[strings.Index(string(b), "CREATE TABLE IF NOT EXISTS `users`"):]
	must(db.Exec(ddl[:strings.Index(ddl, ";")+1]).Error)
	b, e = os.ReadFile("../commons/dbcommons/sqls/20261109.sql")
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
	must(dbcommons.EnsurePasswordSMSPhoneTrigger(db))
	t.Run("new SMS accounts actual schema", func(t *testing.T) {
		dao := dbs.UserDao{}
		must(dao.Create(models.User{AppKey: "accounts", UserId: "occupied", LoginAccount: "13800138000", LoginPass: utils.SHA1("original password")}))
		must(dao.Create(models.User{AppKey: "accounts", UserId: "suffix", LoginAccount: "13800138000a1b2"}))
		values := []string{"a1b2", "c3d4"}
		n := 0
		u, created, code := createVerifiedSMSUser(dao, func() (string, error) { v := values[n]; n++; return v, nil }, models.User{AppKey: "accounts", UserId: "sms-new", Phone: "13800138000"})
		if code != 0 || !created || u == nil || u.LoginAccount != "13800138000c3d4" || n != 2 {
			t.Fatalf("%+v %v %d n=%d", u, created, code, n)
		}
		var pass sql.NullString
		must(db.Raw("SELECT login_pass FROM users WHERE app_key='accounts' AND user_id='sms-new'").Scan(&pass).Error)
		if pass.Valid {
			t.Fatalf("SMS password must be NULL: %+v", pass)
		}
		c := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, "accounts")
		c = context.WithValue(c, ctxs.CtxKey_RequesterId, "sms-new")
		c = context.WithValue(c, ctxs.CtxKey_AuthKind, "legacy")
		statusCode, status := QryPasswordStatus(c)
		if statusCode != 0 || status.HasPassword || !status.CanSetPasswordBySMS {
			t.Fatalf("%d %+v", statusCode, status)
		}
		for _, item := range []models.User{
			{AppKey: "accounts", UserId: "phone-reg", Phone: "13900139000", LoginPass: utils.SHA1("phone password")},
			{AppKey: "accounts", UserId: "email-reg", Email: "test@example.invalid", LoginPass: utils.SHA1("email password")},
		} {
			must(dao.Create(item))
			got, e := dao.FindByUserId(item.AppKey, item.UserId)
			must(e)
			if got.LoginPass != item.LoginPass {
				t.Fatalf("registration password lost: %+v", got)
			}
		}
		registered, e := dao.FindByAccount("accounts", "13800138000")
		must(e)
		if registered.LoginPass != utils.SHA1("original password") {
			t.Fatal("account registration hash changed")
		}
		cross, created, code := CreateVerifiedSMSUser(models.User{AppKey: "other-accounts", UserId: "cross", Phone: "13800138000"})
		if code != 0 || !created || cross.LoginAccount != "13800138000" {
			t.Fatalf("cross app: %+v %v %d", cross, created, code)
		}
		if e := dao.CreateSMSUser(models.User{AppKey: "accounts", UserId: "sms-new", Phone: "13700137000", LoginAccount: "uniqueaccount"}); e == nil || errors.Is(e, models.ErrAccountConflict) || errors.Is(e, models.ErrSMSPhoneConflict) {
			t.Fatalf("userid error misclassified: %v", e)
		}
		for _, signal := range []string{
			"SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='injected database failure'",
			"SIGNAL SQLSTATE '23000' SET MYSQL_ERRNO=1062, MESSAGE_TEXT=\"Duplicate entry 'uniq_account' for key 'users.uniq_userid'\"",
		} {
			must(db.Exec("CREATE TRIGGER sms_account_failure BEFORE INSERT ON users FOR EACH ROW " + signal).Error)
			calls := 0
			failed, created, code := createVerifiedSMSUser(dao, func() (string, error) { calls++; return "a1b2", nil }, models.User{AppKey: "accounts", UserId: "failure", Phone: "13700137000"})
			must(db.Exec("DROP TRIGGER sms_account_failure").Error)
			if failed != nil || created || code != 17000 || calls != 0 {
				t.Fatalf("DB error retried: %+v %v %d suffixes=%d", failed, created, code, calls)
			}
		}
		owner, created, code := CreateVerifiedSMSUser(models.User{AppKey: "accounts", UserId: "loser", Phone: "13800138000"})
		if code != 0 || created || owner.UserId != "sms-new" {
			t.Fatalf("phone race: %+v %v %d", owner, created, code)
		}
		// First password remains available for these automatically assigned accounts.
		now := time.Now()
		sent := ""
		s := passwordSMSService{DB: db, Pepper: strings.Repeat("t", 32), Now: func() time.Time { return now }, Send: func(_ context.Context, _, _, code string) error { sent = code; return nil }}
		if code := s.send(c); code != 0 {
			t.Fatal(code)
		}
		if code := s.consume(c, sent, "first password"); code != 0 {
			t.Fatal(code)
		}
		got, e := dao.FindByUserId("accounts", "sms-new")
		must(e)
		if got.LoginPass != utils.SHA1("first password") {
			t.Fatal("initial password not saved")
		}
	})
	for _, table := range []string{"password_sms_challenges", "password_sms_quotas"} {
		must(db.Exec("DELETE FROM " + table).Error)
	}
	phone := "13800138000"
	id := utils.ShortMd5(phone)
	must(db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('smsflow',?,NULL)", id).Error)
	c := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, "smsflow")
	c = context.WithValue(c, ctxs.CtxKey_RequesterId, id)
	c = context.WithValue(c, ctxs.CtxKey_AuthKind, "legacy")
	now := time.Now()
	sent := ""
	s := passwordSMSService{DB: db, Pepper: strings.Repeat("t", 32), Now: func() time.Time { return now }, Send: func(_ context.Context, app, target, code string) error {
		if app != "smsflow" || target != phone {
			t.Fatal("wrong recipient")
		}
		sent = code
		return nil
	}}
	u, code := ResolveVerifiedSMSUser("smsflow", phone)
	if code != 0 || u == nil || u.Phone != phone {
		t.Fatalf("backfill: %d %+v", code, u)
	}
	statusCode, status := QryPasswordStatus(c)
	if statusCode != 0 || status.HasPassword || !status.CanSetPasswordBySMS {
		t.Fatalf("status %d %+v", statusCode, status)
	}
	if code := s.send(c); code != 0 {
		t.Fatal(code)
	}
	old := sent
	// Simulate an old challenge surviving an unbind followed by verified backfill.
	must(db.Exec("UPDATE users SET phone=NULL WHERE app_key='smsflow' AND user_id=?", id).Error)
	u, code = ResolveVerifiedSMSUser("smsflow", phone)
	if code != 0 || u == nil {
		t.Fatal(code)
	}
	if code := s.consume(c, old, "abcdef"); code != 17023 {
		t.Fatalf("old challenge accepted: %d", code)
	}
	now = now.Add(time.Minute)
	if code := s.send(c); code != 0 {
		t.Fatal(code)
	}
	if code := s.consume(c, sent, "abcdef"); code != 0 {
		t.Fatal(code)
	}
	var hash string
	must(db.Raw("SELECT login_pass FROM users WHERE app_key='smsflow' AND user_id=?", id).Scan(&hash).Error)
	if hash != utils.SHA1("abcdef") {
		t.Fatal("password not set")
	}
	statusCode, status = QryPasswordStatus(c)
	if statusCode != 0 || !status.HasPassword || status.CanSetPasswordBySMS {
		t.Fatalf("status %d %+v", statusCode, status)
	}
}
