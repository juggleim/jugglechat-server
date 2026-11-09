package apis

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	am "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/services/deviceauth"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSetAccountMySQL(t *testing.T) {
	if os.Getenv("ACCOUNT_TEST_MYSQL") != "1" {
		t.Skip("requires isolated localhost:13316/account_test")
	}
	gin.SetMode(gin.TestMode)
	configures.Config.Mysql.User = "root"
	configures.Config.Mysql.Password = ""
	configures.Config.Mysql.Address = "127.0.0.1:13316"
	configures.Config.Mysql.DbName = "account_test"
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
		ddl = ddl[:strings.Index(ddl, ";")+1]
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	key := "0123456789abcdef"
	for _, app := range []string{"acctapp", "otherapp"} {
		if err := db.Exec("INSERT INTO apps(app_key,app_secure_key,app_secret) VALUES (?,?,'test-secret')", app, key).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, row := range [][3]string{{"acctapp", "u1", "Initial1"}, {"acctapp", "u2", "Initial2"}, {"otherapp", "u1", "Initial3"}} {
		if err := db.Exec("INSERT INTO users(app_key,user_id,login_account,login_pass) VALUES (?,?,?,'unchanged')", row[0], row[1], row[2]).Error; err != nil {
			t.Fatal(err)
		}
	}
	set := func(app, user, kind, account string) errs.IMErrorCode {
		ctx := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, app)
		ctx = context.WithValue(ctx, ctxs.CtxKey_RequesterId, user)
		ctx = context.WithValue(ctx, ctxs.CtxKey("CtxKey_AuthKind"), kind)
		return services.SetUserAccount(ctx, &am.SetUserAccountReq{Account: account, Password: "ignored"})
	}
	t.Run("service independent of app cache", func(t *testing.T) {
		// Fresh app keys have never entered appinfos. A cache lookup selects the
		// full app record; UpdateAccount needs only the locked app ID.
		if err := db.Exec("INSERT INTO apps(app_key,app_secure_key,app_secret) VALUES ('nocacheapp',?,'test-secret')", key).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("INSERT INTO users(app_key,user_id,login_account) VALUES ('nocacheapp','u1','Before123')").Error; err != nil {
			t.Fatal(err)
		}
		name := "test:reject_app_cache_lookup"
		if err := db.Callback().Query().Before("gorm:query").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "apps" && len(tx.Statement.Selects) == 0 {
				tx.AddError(errors.New("app cache lookup unavailable"))
			}
		}); err != nil {
			t.Fatal(err)
		}
		defer db.Callback().Query().Remove(name)
		if got := set("nocacheapp", "u1", "legacy", "After123"); got != 0 {
			t.Fatalf("service depended on app cache: got %d", got)
		}
		var account string
		if err := db.Raw("SELECT login_account FROM users WHERE app_key='nocacheapp' AND user_id='u1'").Scan(&account).Error; err != nil {
			t.Fatal(err)
		}
		if account != "After123" {
			t.Fatalf("account was not updated: %q", account)
		}
	})
	t.Run("cold app database failure is internal error", func(t *testing.T) {
		name := "test:app_database_failure"
		if err := db.Callback().Query().Before("gorm:query").Register(name, func(tx *gorm.DB) {
			if tx.Statement.Table == "apps" {
				tx.AddError(errors.New("app database unavailable"))
			}
		}); err != nil {
			t.Fatal(err)
		}
		defer db.Callback().Query().Remove(name)
		if got := set("coldfailureapp", "u1", "legacy", "Valid123"); got != 17000 {
			t.Fatalf("database failure got %d, want 17000", got)
		}
	})
	t.Run("service guards", func(t *testing.T) {
		for _, tc := range []struct {
			app, user, kind string
			want            errs.IMErrorCode
		}{
			{"", "u1", "legacy", 17002}, {"missingapp", "u1", "legacy", 17002},
			{"acctapp", "", "legacy", 17005}, {"acctapp", "u1", "", 17005},
			{"acctapp", "u1", "api_key", 17005}, {"acctapp", "u1", "service", 17005},
		} {
			if got := set(tc.app, tc.user, tc.kind, "Valid123"); got != tc.want {
				t.Errorf("%+v got %d", tc, got)
			}
		}
	})
	t.Run("boundaries", func(t *testing.T) {
		for _, account := range []string{"", "abc12", "abcdefghijklmnopqrstu", "abc中文123", "ａｂｃ１２３", " abc123", "abc123 ", "abc\t123", "abc123\n", "abc_123", "abc😀123"} {
			if checkAccount(account) {
				t.Errorf("registration accepted %q", account)
			}
			if got := set("acctapp", "u1", "legacy", account); got != errs.IMErrorCode_APP_REQ_BODY_ILLEGAL {
				t.Errorf("invalid %q got %d", account, got)
			}
		}
		for _, account := range []string{"abc123", "ABCDEFGHIJKLMNOPQRST", "123456"} {
			if !checkAccount(account) {
				t.Errorf("registration rejected %q", account)
			}
			if got := set("acctapp", "u1", "legacy", account); got != 0 {
				t.Errorf("valid %q got %d", account, got)
			}
		}
	})
	t.Run("outcomes", func(t *testing.T) {
		for _, tc := range []struct {
			app, user, account string
			want               errs.IMErrorCode
		}{
			{"acctapp", "u1", "Mixed123", 0}, {"acctapp", "u1", "Mixed123", 0}, {"acctapp", "u1", "mIXED123", 0},
			{"acctapp", "absent", "Absent12", 17012}, {"otherapp", "u2", "Absent12", 17012},
			{"acctapp", "u2", "mixed123", 17011}, {"otherapp", "u1", "mixed123", 0},
		} {
			if got := set(tc.app, tc.user, "legacy", tc.account); got != tc.want {
				t.Errorf("%+v got %d", tc, got)
			}
		}
		var row struct{ Account, Pass string }
		if err := db.Raw("SELECT login_account AS account,login_pass AS pass FROM users WHERE app_key='acctapp' AND user_id='u1'").Scan(&row).Error; err != nil {
			t.Fatal(err)
		}
		if row.Account != "mIXED123" || row.Pass != "unchanged" {
			t.Fatalf("case/password: %+v", row)
		}
	})
	t.Run("other DB errors", func(t *testing.T) {
		for _, signal := range []string{"SIGNAL SQLSTATE '23000' SET MYSQL_ERRNO=1062, MESSAGE_TEXT='Duplicate entry for key uniq_phone'", "SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='private database failure'"} {
			if err := db.Exec("CREATE TRIGGER account_failure BEFORE UPDATE ON users FOR EACH ROW " + signal).Error; err != nil {
				t.Fatal(err)
			}
			got := set("acctapp", "u1", "legacy", "Newname1")
			if err := db.Exec("DROP TRIGGER account_failure").Error; err != nil {
				t.Fatal(err)
			}
			if got != 17000 {
				t.Errorf("non-account failure got %d", got)
			}
		}
	})
	t.Run("DAO verifies app and user", func(t *testing.T) {
		dao := dbs.UserDao{}
		if err := dao.UpdateAccount("missingapp", "u1", "Direct123"); err == nil {
			t.Fatal("missing app succeeded")
		}
		if err := dao.UpdateAccount("acctapp", "missing", "Direct123"); err == nil {
			t.Fatal("missing user succeeded")
		}
	})
	t.Run("handler guards and boundaries", func(t *testing.T) {
		for _, tc := range []struct {
			body, app, user, kind string
			want                  int
		}{
			{`{"account":"abc12"}`, "acctapp", "u1", "legacy", 17003},
			{`{"account":"abc123 "}`, "acctapp", "u1", "legacy", 17003},
			{`{"account":"abc中文123"}`, "acctapp", "u1", "legacy", 17003},
			{`{"account":123}`, "acctapp", "u1", "legacy", 17003},
			{`{"account":"Valid123"}`, "", "u1", "legacy", 17002},
			{`{"account":"Valid123"}`, "acctapp", "u1", "service", 17005},
			{`{"account":"Valid123"}`, "acctapp", "u1", "", 17005},
		} {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("POST", "/jim/users/setaccount", strings.NewReader(tc.body))
			ctx.Set(string(ctxs.CtxKey_AppKey), tc.app)
			ctx.Set(string(ctxs.CtxKey_RequesterId), tc.user)
			ctx.Set(string(ctxs.CtxKey_AuthKind), tc.kind)
			SetLoginAccount(ctx)
			var body struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tc.want {
				t.Errorf("%+v got %d", tc, body.Code)
			}
		}
	})
	t.Run("SQL logging suppressed", func(t *testing.T) {
		var output strings.Builder
		previous := db.Logger
		db.Logger = logger.New(accountLogWriter{&output}, logger.Config{LogLevel: logger.Info})
		defer func() { db.Logger = previous }()
		if got := set("acctapp", "u1", "legacy", "Private123"); got != 0 {
			t.Fatal(got)
		}
		if got := set("acctapp", "u2", "legacy", "Private123"); got != 17011 {
			t.Fatal(got)
		}
		if output.Len() != 0 {
			t.Fatalf("account SQL leaked to logger: %s", output.String())
		}
	})
	t.Run("concurrent claim", func(t *testing.T) {
		start := make(chan struct{})
		results := make(chan errs.IMErrorCode, 2)
		var wg sync.WaitGroup
		for _, user := range []string{"u1", "u2"} {
			wg.Add(1)
			go func(user string) { defer wg.Done(); <-start; results <- set("acctapp", user, "legacy", "Racing123") }(user)
		}
		close(start)
		wg.Wait()
		close(results)
		counts := map[errs.IMErrorCode]int{}
		for code := range results {
			counts[code]++
		}
		if counts[0] != 1 || counts[17011] != 1 {
			t.Fatal(counts)
		}
	})
	t.Run("Validate types", func(t *testing.T) {
		now := time.Now().UnixMilli()
		legacy, err := (services.ImToken{AppKey: "acctapp", UserId: "u1", TokenTime: now}).ToTokenString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		svc, err := deviceauth.SignService("acctapp", "u1", []byte(key), now)
		if err != nil {
			t.Fatal(err)
		}
		apiKey, err := services.GenerateApiKey("acctapp", key)
		if err != nil {
			t.Fatal(err)
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
		ds, err := deviceStore().Issue(context.Background(), "acctapp", "u1", deviceauth.DeviceParams{DeviceID: json.RawMessage(`"account-device"`), Platform: json.RawMessage(`"ios"`)}, []byte(key), now)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			name, token string
			allowed     bool
		}{{"apikey", "Bearer " + apiKey, false}, {"service", svc, false}, {"legacy", legacy, true}, {"ds1", ds, true}} {
			t.Run(tc.name, func(t *testing.T) {
				r := gin.New()
				r.Use(Validate)
				r.POST("/jim/users/setaccount", SetLoginAccount)
				w := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/jim/users/setaccount", strings.NewReader(`{"account":"Http1234"}`))
				req.Header.Set("appkey", "acctapp")
				req.Header.Set("Authorization", tc.token)
				r.ServeHTTP(w, req)
				var body struct {
					Code int `json:"code"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if (body.Code == 0) != tc.allowed {
					t.Fatalf("code %d allowed %v", body.Code, tc.allowed)
				}
			})
		}
	})
}

type accountLogWriter struct{ out *strings.Builder }

func (w accountLogWriter) Printf(format string, args ...interface{}) { w.out.WriteString(format) }

var _ logger.Writer = accountLogWriter{}
