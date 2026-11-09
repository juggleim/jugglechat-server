package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/smsengines"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSetPasswordRules(t *testing.T) {
	for _, p := range []string{"", "12345", strings.Repeat("a", 65), "      ", "\xff123456"} {
		if validInitialPassword(p) {
			t.Fatal("accepted invalid password")
		}
	}
	for _, p := range []string{"123456", " 密码六个字符 ", strings.Repeat("密", 64), " 1234 "} {
		if !validInitialPassword(p) {
			t.Fatal("rejected valid password")
		}
	}
}

func TestPasswordSMSPepperRequired(t *testing.T) {
	old := configures.Config.PasswordSMSPepper
	defer func() { configures.Config.PasswordSMSPepper = old }()
	c := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, "app")
	c = context.WithValue(c, ctxs.CtxKey_RequesterId, "u")
	c = context.WithValue(c, ctxs.CtxKey_AuthKind, "legacy")
	for _, pepper := range []string{"", strings.Repeat("x", 31)} {
		configures.Config.PasswordSMSPepper = pepper
		if got := SendPasswordSMS(c); got != 17024 {
			t.Fatalf("send code=%d", got)
		}
		if got := SetPasswordBySMS(c, "123456", "abcdef"); got != 17024 {
			t.Fatalf("consume code=%d", got)
		}
	}
}

func TestPasswordSMSDigestBindings(t *testing.T) {
	s := passwordSMSService{Pepper: strings.Repeat("t", 32)}
	base := s.digest("app", "user", "phone", "nonce", "123456", 7)
	for _, other := range []string{s.digest("other", "user", "phone", "nonce", "123456", 7), s.digest("app", "other", "phone", "nonce", "123456", 7), s.digest("app", "user", "other", "nonce", "123456", 7), s.digest("app", "user", "phone", "other", "123456", 7), s.digest("app", "user", "phone", "nonce", "654321", 7), s.digest("app", "user", "phone", "nonce", "123456", 8)} {
		if other == base {
			t.Fatal("unbound digest field")
		}
	}
}

func TestSetPasswordMySQL(t *testing.T) {
	if os.Getenv("SET_PASSWORD_TEST_MYSQL") != "1" {
		t.Skip("requires isolated localhost:13316/set_password_test")
	}
	db, err := gorm.Open(mysql.Open("root@tcp(127.0.0.1:13316)/set_password_test?charset=utf8mb4&parseTime=true"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	schema, err := os.ReadFile("../commons/dbcommons/sqls/20250201.sql")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(schema), "CREATE TABLE IF NOT EXISTS `users`")
	ddl := string(schema)[start:]
	if err = db.Exec(ddl[:strings.Index(ddl, ";")+1]).Error; err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../commons/dbcommons/sqls/20261109.sql")
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("DROP TABLE IF EXISTS password_sms_challenges")
	db.Exec("DROP TABLE IF EXISTS password_sms_quotas")
	if err := db.Connection(func(conn *gorm.DB) error {
		for _, stmt := range strings.Split(string(migration), ";") {
			if strings.TrimSpace(stmt) != "" {
				if err = conn.Exec(stmt).Error; err != nil {
					return err
				}
			}
		}
		return dbcommons.EnsurePasswordSMSPhoneTrigger(conn)
	}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Truncate(time.Second)
	var sent string
	var fail bool
	s := passwordSMSService{Pepper: strings.Repeat("t", 32), DB: db, Now: func() time.Time { return now }, Send: func(_ context.Context, app, phone, code string) error {
		sent = code
		if phone != "13800138000" {
			return errors.New("wrong target")
		}
		if fail {
			return errors.New("delivery failed")
		}
		return nil
	}}
	ctx := func(app, user string) context.Context {
		c := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, app)
		c = context.WithValue(c, ctxs.CtxKey_RequesterId, user)
		return context.WithValue(c, ctxs.CtxKey_AuthKind, "legacy")
	}
	c := ctx("setapp", "u1")
	check := func(got, want errs.IMErrorCode) {
		t.Helper()
		if got != want {
			t.Fatalf("code=%d want=%d", got, want)
		}
	}
	reset := func(pass interface{}) {
		t.Helper()
		for _, table := range []string{"password_sms_challenges", "password_sms_quotas", "users"} {
			if e := db.Exec("DELETE FROM " + table).Error; e != nil {
				t.Fatal(e)
			}
		}
		if e := db.Exec("INSERT INTO users(app_key,user_id,phone,login_pass) VALUES ('setapp','u1','13800138000',?)", pass).Error; e != nil {
			t.Fatal(e)
		}
		fail = false
	}
	t.Run("HMAC binding restart and wrong pepper", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		var ch passwordSMSChallenge
		if e := db.Take(&ch).Error; e != nil {
			t.Fatal(e)
		}
		mac := hmac.New(sha256.New, []byte(strings.Repeat("t", 32)))
		mac.Write([]byte(`["set_password","setapp","u1","13800138000","` + ch.Nonce + `","` + code + `","0"]`))
		if ch.Digest != hex.EncodeToString(mac.Sum(nil)) {
			t.Fatal("digest does not bind all fields with HMAC")
		}
		other := s
		other.Pepper = strings.Repeat("z", 32)
		check(other.consume(c, code, "abcdef"), 17023)
		restarted := passwordSMSService{DB: db, Now: s.Now, Pepper: strings.Repeat("t", 32)}
		check(restarted.consume(c, code, "abcdef"), 0)
	})
	t.Run("null empty once and SHA1 compatibility", func(t *testing.T) {
		for _, pass := range []interface{}{nil, ""} {
			reset(pass)
			check(s.send(c), 0)
			check(s.consume(c, "000000", "abcdef"), 17023)
			check(s.consume(c, sent, " 1234 "), 0)
			var hash string
			db.Raw("SELECT login_pass FROM users WHERE user_id='u1'").Scan(&hash)
			if hash != utils.SHA1(" 1234 ") {
				t.Fatal("password trimmed or incompatible hash")
			}
			check(s.consume(c, sent, "another"), 17019)
			check(s.send(c), 17019)
		}
	})
	t.Run("identity", func(t *testing.T) {
		reset(nil)
		for _, kind := range []string{"", "api_key", "service"} {
			bad := context.WithValue(c, ctxs.CtxKey_AuthKind, kind)
			check(s.send(bad), 17005)
			check(s.consume(bad, "123456", "abcdef"), 17005)
		}
		check(s.send(ctx("setapp", "")), 17005)
		check(s.send(ctx("setapp", "missing")), 17012)
	})
	t.Run("no phone and already password", func(t *testing.T) {
		reset("hash")
		check(s.send(c), 17019)
		reset(nil)
		db.Exec("UPDATE users SET phone=NULL")
		check(s.send(c), 17020)
		check(s.consume(c, "123456", "abcdef"), 17020)
	})
	t.Run("attempts expiry isolation", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		check(s.consume(ctx("other", "u1"), code, "abcdef"), 17012)
		db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('setapp','u2','13900139000')")
		check(s.consume(ctx("setapp", "u2"), code, "abcdef"), 17023)
		for i := 0; i < 4; i++ {
			check(s.consume(c, "xxxxxx", "abcdef"), 17023)
		}
		check(s.consume(c, "xxxxxx", "abcdef"), 17022)
		check(s.consume(c, code, "abcdef"), 17022)
		reset(nil)
		check(s.send(c), 0)
		now = now.Add(5 * time.Minute)
		check(s.consume(c, sent, "abcdef"), 17023)
	})
	t.Run("resend rate failed delivery and phone binding", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		old := sent
		check(s.send(c), 17021)
		now = now.Add(time.Minute)
		check(s.send(c), 0)
		if sent == old {
			t.Fatal("resend reused old code")
		}
		check(s.consume(c, old, "abcdef"), 17023)
		db.Exec("UPDATE users SET phone='13900139000'")
		check(s.consume(c, sent, "abcdef"), 17023)
		reset(nil)
		check(s.send(c), 0)
		old = sent
		now = now.Add(time.Minute)
		fail = true
		check(s.send(c), 17008)
		check(s.consume(c, old, "abcdef"), 17023)
		check(s.consume(c, sent, "abcdef"), 17023)
		check(s.send(c), 17021)
	})
	t.Run("hour quota", func(t *testing.T) {
		reset(nil)
		for i := 0; i < 5; i++ {
			check(s.send(c), 0)
			now = now.Add(time.Minute)
		}
		check(s.send(c), 17021)
		now = now.Add(time.Hour)
		check(s.send(c), 0)
	})
	t.Run("explicit rejection permits immediate retry but retains hourly quota", func(t *testing.T) {
		reset(nil)
		local := s
		calls := 0
		local.Send = func(context.Context, string, string, string) error {
			calls++
			return &smsengines.SMSRejectedError{Code: "isv.SMS_SIGNATURE_ILLEGAL"}
		}
		for i := 0; i < 5; i++ {
			check(local.send(c), 17008)
		}
		check(local.send(c), 17021)
		if calls != 5 {
			t.Fatalf("provider calls=%d", calls)
		}
	})
	t.Run("rejection then immediate successful resend", func(t *testing.T) {
		reset(nil)
		local := s
		local.Send = func(context.Context, string, string, string) error {
			return &smsengines.SMSRejectedError{Code: "rejected"}
		}
		check(local.send(c), 17008)
		var rejected passwordSMSChallenge
		if err := db.Take(&rejected).Error; err != nil {
			t.Fatal(err)
		}
		if rejected.Delivered {
			t.Fatal("rejection activated challenge")
		}
		got, retry := s.sendWithRetry(c)
		check(got, 0)
		if retry != 60 {
			t.Fatalf("successful retry=%d", retry)
		}
		check(s.consume(c, sent, "abcdef"), 0)
	})
	t.Run("retry hints rejection unknown and max of both quotas", func(t *testing.T) {
		reset(nil)
		local := s
		local.Send = func(context.Context, string, string, string) error {
			return &smsengines.SMSRejectedError{Code: "rejected"}
		}
		got, retry := local.sendWithRetry(c)
		check(got, 17008)
		if retry != 0 {
			t.Fatalf("rejected retry=%d", retry)
		}
		local.Send = func(context.Context, string, string, string) error { return context.DeadlineExceeded }
		got, retry = local.sendWithRetry(c)
		check(got, 17008)
		if retry != 60 {
			t.Fatalf("unknown retry=%d", retry)
		}
		now = now.Add(17 * time.Second)
		got, retry = local.sendWithRetry(c)
		check(got, 17021)
		if retry != 43 {
			t.Fatalf("remaining retry=%d", retry)
		}
		phoneKey := passwordSMSKey("set_password", "setapp", "phone", "13800138000")
		if err := db.Model(&passwordSMSQuota{}).Where("id=?", phoneKey).Updates(map[string]interface{}{"day_count": 10, "day_start": now.Unix() - 100}).Error; err != nil {
			t.Fatal(err)
		}
		got, retry = local.sendWithRetry(c)
		check(got, 17021)
		if retry != 86300 {
			t.Fatalf("max quota retry=%d", retry)
		}
	})
	t.Run("rejected attempts still exhaust daily budget", func(t *testing.T) {
		reset(nil)
		local := s
		local.Send = func(context.Context, string, string, string) error {
			return &smsengines.SMSRejectedError{Code: "rejected"}
		}
		for i := 0; i < 10; i++ {
			if i == 5 {
				now = now.Add(time.Hour)
			}
			check(local.send(c), 17008)
		}
		now = now.Add(time.Hour)
		got, retry := local.sendWithRetry(c)
		check(got, 17021)
		if retry != 79200 {
			t.Fatalf("daily retry=%d", retry)
		}
	})
	t.Run("late rejection cannot release replacement quota or activate replacement", func(t *testing.T) {
		reset(nil)
		local := s
		var clock atomic.Int64
		clock.Store(now.Unix())
		local.Now = func() time.Time { return time.Unix(clock.Load(), 0) }
		started, finish := make(chan struct{}), make(chan struct{})
		var calls atomic.Int32
		local.Send = func(context.Context, string, string, string) error {
			if calls.Add(1) == 1 {
				close(started)
				<-finish
				return &smsengines.SMSRejectedError{Code: "rejected"}
			}
			return context.DeadlineExceeded
		}
		done := make(chan errs.IMErrorCode, 1)
		go func() { done <- local.send(c) }()
		<-started
		clock.Add(60)
		check(local.send(c), 17008)
		var before passwordSMSChallenge
		if err := db.Take(&before).Error; err != nil {
			t.Fatal(err)
		}
		close(finish)
		check(<-done, 17008)
		got, retry := local.sendWithRetry(c)
		check(got, 17021)
		if retry != 60 {
			t.Fatalf("replacement cooldown=%d", retry)
		}
		var after passwordSMSChallenge
		if err := db.Take(&after).Error; err != nil {
			t.Fatal(err)
		}
		if after.Nonce != before.Nonce || after.Delivered {
			t.Fatal("old failure modified/activated replacement")
		}
	})
	t.Run("preflight failure does not reserve or invalidate", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		local := s
		local.Prepare = func(context.Context, string) (smsengines.ISmsEngine, error) { return nil, errors.New("invalid config") }
		check(local.send(c), 17008)
		var quotas []passwordSMSQuota
		if err := db.Find(&quotas).Error; err != nil {
			t.Fatal(err)
		}
		for _, q := range quotas {
			if q.HourCount != 1 || q.DayCount != 1 {
				t.Fatal("preflight consumed budget")
			}
		}
		check(s.consume(c, code, "abcdef"), 0)
	})
	t.Run("rollback retry", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		name := "test:setpass_fail"
		if e := db.Callback().Update().Before("gorm:update").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "users" {
				tx.AddError(errors.New("injected write failure"))
			}
		}); e != nil {
			t.Fatal(e)
		}
		check(s.consume(c, code, "abcdef"), 17000)
		db.Callback().Update().Remove(name)
		check(s.consume(c, code, "abcdef"), 0)
	})
	t.Run("parallel consumption", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		var wins atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				got := s.consume(c, code, "abcdef")
				if got == 0 {
					wins.Add(1)
				} else if got != 17019 {
					t.Errorf("concurrent code=%d", got)
				}
			}()
		}
		wg.Wait()
		if wins.Load() != 1 {
			t.Fatalf("successes=%d", wins.Load())
		}
	})
	t.Run("daily quota", func(t *testing.T) {
		reset(nil)
		for i := 0; i < 10; i++ {
			if i == 5 {
				now = now.Add(time.Hour)
			}
			check(s.send(c), 0)
			now = now.Add(time.Minute)
		}
		now = now.Add(time.Hour)
		check(s.send(c), 17021)
		now = now.Add(24 * time.Hour)
		check(s.send(c), 0)
	})
	t.Run("phone quota follows phone and account quota follows user", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		db.Exec("UPDATE users SET phone=NULL WHERE user_id='u1'")
		db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('setapp','u2','13800138000')")
		check(s.send(ctx("setapp", "u2")), 17021)
		db.Exec("UPDATE users SET phone='13900139000' WHERE user_id='u1'")
		check(s.send(c), 17021)
	})
	t.Run("cross app and login record cannot authorize", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		db.Exec("INSERT INTO users(app_key,user_id,phone) VALUES ('other','u1','13800138000')")
		check(s.consume(ctx("other", "u1"), code, "abcdef"), 17023)
		check(s.consume(c, "123456", "abcdef"), 17023)
	})
	t.Run("parallel sends once", func(t *testing.T) {
		reset(nil)
		local := s
		var deliveries atomic.Int32
		local.Send = func(context.Context, string, string, string) error { deliveries.Add(1); return nil }
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				got := local.send(c)
				if got != 0 && got != 17021 {
					t.Errorf("send code=%d", got)
				}
			}()
		}
		wg.Wait()
		if deliveries.Load() != 1 {
			t.Fatalf("deliveries=%d", deliveries.Load())
		}
	})
	t.Run("challenge delete rollback", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		name := "test:setpass_delete_fail"
		db.Callback().Delete().Before("gorm:delete").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "password_sms_challenges" {
				tx.AddError(errors.New("delete failed"))
			}
		})
		check(s.consume(c, code, "abcdef"), 17000)
		db.Callback().Delete().Remove(name)
		var count int64
		db.Raw("SELECT COUNT(*) FROM users WHERE login_pass IS NULL OR login_pass=''").Scan(&count)
		if count != 1 {
			t.Fatal("password committed despite failed consume")
		}
		check(s.consume(c, code, "abcdef"), 0)
	})
	t.Run("phone change and restore still invalidates", func(t *testing.T) {
		reset(nil)
		check(s.send(c), 0)
		code := sent
		db.Exec("UPDATE users SET phone='13900139000' WHERE user_id='u1'")
		db.Exec("UPDATE users SET phone='13800138000' WHERE user_id='u1'")
		check(s.consume(c, code, "abcdef"), 17023)
	})
	t.Run("upgrade old quota schema is idempotent", func(t *testing.T) {
		if err := db.Exec("ALTER TABLE password_sms_quotas DROP COLUMN attempt_nonce").Error; err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile("../commons/dbcommons/sqls/20261110.sql")
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			if err := db.Connection(func(conn *gorm.DB) error {
				for _, stmt := range strings.Split(string(b), ";") {
					if strings.TrimSpace(stmt) != "" {
						if err := conn.Exec(stmt).Error; err != nil {
							return err
						}
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}
		reset(nil)
		check(s.send(c), 0)
	})
}
