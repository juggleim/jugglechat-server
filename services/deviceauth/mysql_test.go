package deviceauth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func mysqlStore(t *testing.T) Store {
	t.Helper()
	if os.Getenv("DEVICEAUTH_TEST_MYSQL") != "1" {
		t.Skip("set DEVICEAUTH_TEST_MYSQL=1 for isolated localhost:13316/deviceauth_test integration database")
	}
	db, e := gorm.Open(mysql.Open("root@tcp(127.0.0.1:13316)/deviceauth_test?charset=utf8mb4&parseTime=true"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	// Apply the actual migration, not a test-created approximation.
	if e = db.Exec("CREATE TABLE IF NOT EXISTS qrcoderecords (app_key VARCHAR(64),code_id VARCHAR(64),user_id VARCHAR(64),status INT,created_time BIGINT)").Error; e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile("../../commons/dbcommons/sqls/20261103.sql")
	if e != nil {
		t.Fatal(e)
	}
	for _, q := range strings.Split(string(b), ";") {
		if strings.TrimSpace(q) == "" {
			continue
		}
		if e = db.Exec(q).Error; e != nil {
			t.Fatal(e)
		}
	}
	for _, table := range []string{"device_revoke_jobs", "user_devices", "device_login_qrs"} {
		if e = db.Exec("DELETE FROM " + table).Error; e != nil {
			t.Fatal(e)
		}
	}
	sql, _ := db.DB()
	t.Cleanup(func() { sql.Close() })
	return Store{DB: db, IM: UnsupportedIM{}}
}
func params(id string) DeviceParams {
	var p DeviceParams
	_ = json.Unmarshal([]byte(`{"device_id":"`+id+`","platform":"ios"}`), &p)
	return p
}

func TestMySQLDeviceIDWhitespaceCannotAlias(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	token, err := s.Issue(ctx, "app", "u", params("device-A"), key, 100)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Verify(token, key, "app")
	if err != nil {
		t.Fatal(err)
	}
	var aliases int64
	if err = deviceQuery(s.DB.Model(&Device{}), "app", "u", "device-A ").Count(&aliases).Error; err != nil || aliases != 1 {
		t.Fatalf("PAD SPACE fixture: count=%d err=%v", aliases, err)
	}
	for _, id := range []string{"device-A ", " device-A", "device-A\u00a0"} {
		if _, err = s.Issue(ctx, "app", "u", params(id), key, 101); !errors.Is(err, ErrParams) {
			t.Errorf("login id=%q: %v", id, err)
		}
		if _, err = s.Delete(ctx, c, id, 102); !errors.Is(err, ErrParams) {
			t.Errorf("delete id=%q: %v", id, err)
		}
	}
	secret, err := s.CreateQR(ctx, "app", "whitespace", 103)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.ConfirmQR(ctx, "app", "whitespace", "u", 104); err != nil {
		t.Fatal(err)
	}
	if _, err = s.PrepareQR(ctx, "app", "whitespace", params("device-A "), secret, 105); !errors.Is(err, ErrParams) {
		t.Errorf("QR prepare: %v", err)
	}
	if _, err = s.RedeemQR(ctx, "app", "whitespace", params("device-A "), secret, key, 105, LoginResult{UserID: "u"}, nil); !errors.Is(err, ErrParams) {
		t.Errorf("QR redeem: %v", err)
	}
	if err = s.Authenticate(ctx, c, 106); err != nil {
		t.Errorf("current session mutated: %v", err)
	}
	var jobs int64
	if err = s.DB.Model(&RevokeJob{}).Count(&jobs).Error; err != nil || jobs != 0 {
		t.Fatalf("unexpected revoke jobs=%d err=%v", jobs, err)
	}
}

func TestMySQLDeleteRechecksStoredCurrentDevice(t *testing.T) {
	s := mysqlStore(t)
	// Simulate a row/session issued before boundary-whitespace validation.
	d := Device{AppKey: "app", UserID: "u", DeviceID: "device-A ", SessionID: "old-session", Platform: "ios", State: "active"}
	if err := s.DB.Create(&d).Error; err != nil {
		t.Fatal(err)
	}
	c := Claims{AppKey: "app", UserID: "u", DeviceID: "device-A ", SessionID: "old-session", Kind: "user_session"}
	if _, err := s.Delete(context.Background(), c, "device-A", 100); !errors.Is(err, ErrCurrent) {
		t.Fatalf("stored current device alias not protected: %v", err)
	}
	if err := s.DB.First(&d, d.ID).Error; err != nil || d.State != "active" || d.RevokeOperationID != "" {
		t.Fatalf("device changed: %+v err=%v", d, err)
	}
	var jobs int64
	if err := s.DB.Model(&RevokeJob{}).Count(&jobs).Error; err != nil || jobs != 0 {
		t.Fatalf("unexpected revoke jobs=%d err=%v", jobs, err)
	}
}
func TestMySQLLifecycleAndPersistentRetry(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	token, e := s.Issue(ctx, "app", "u", params("d"), key, 100)
	if e != nil {
		t.Fatal(e)
	}
	c, e := Verify(token, key, "app")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Authenticate(ctx, c, 101); e != nil {
		t.Fatal(e)
	}
	newer, e := s.Issue(ctx, "app", "u", params("d"), key, 102)
	if e != nil {
		t.Fatal(e)
	}
	nc, _ := Verify(newer, key, "app")
	if e = s.Authenticate(ctx, c, 103); e == nil {
		t.Fatal("old session revived")
	}
	caller := Claims{AppKey: "app", UserID: "u", DeviceID: "current", SessionID: "current", Kind: "user_session"}
	j, e := s.Delete(ctx, caller, "d", 200)
	if !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	if j.BannedUntil != 0 {
		t.Fatal(j)
	}
	if e = s.Authenticate(ctx, nc, 201); e == nil {
		t.Fatal("revoked business token accepted")
	}
	retry, e := s.Delete(ctx, caller, "d", 300)
	if !errors.Is(e, ErrIM) || retry.OperationID != j.OperationID || retry.BannedUntil != 0 {
		t.Fatalf("%+v %v", retry, e)
	}
	if _, e = s.Issue(ctx, "app", "u", params("d"), key, 301); e != nil {
		t.Fatal("pending job blocked immediate login", e)
	}
	var persisted RevokeJob
	if e = s.DB.Where("operation_id=?", j.OperationID).Take(&persisted).Error; e != nil || persisted.Attempts != 2 || persisted.Done {
		t.Fatalf("%+v %v", persisted, e)
	}
	s.IM = verifiedIM{}
	if e = s.Process(ctx, persisted, 700001); e != nil {
		t.Fatal(e)
	}
	if e = s.Process(ctx, persisted, 700001); e != nil {
		t.Fatal("concurrent/replayed completion must be idempotent", e)
	}
	if _, e = s.Issue(ctx, "app", "u", params("d"), key, 700002); e != nil {
		t.Fatal(e)
	}
	s.IM = UnsupportedIM{}
	if e = s.Process(ctx, persisted, 700003); e != nil {
		t.Fatal("completed old task must not invoke IM after new login", e)
	}
	if e = s.Authenticate(ctx, nc, 700003); e == nil {
		t.Fatal("revoked session resurrected")
	}
	if _, e = s.Delete(ctx, caller, "current", 400); !errors.Is(e, ErrCurrent) {
		t.Fatal(e)
	}
	other := caller
	other.UserID = "other"
	if _, e = s.Delete(ctx, other, "d", 400); !errors.Is(e, ErrNotFound) {
		t.Fatal(e)
	}
}

func TestMySQLListPropagatesIMFailure(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	_, e := s.Issue(ctx, "app", "u", params("d"), []byte("key"), 100)
	if e != nil {
		t.Fatal(e)
	}
	c := Claims{AppKey: "app", UserID: "u", DeviceID: "d", SessionID: "irrelevant", Kind: "user_session"}
	if _, e = s.List(ctx, c); !errors.Is(e, ErrIM) {
		t.Fatal("IM failure became offline list", e)
	}
	s.IM = verifiedIM{}
	g, e := s.List(ctx, c)
	if e != nil || len(g.Online) != 1 || !g.Online[0].IsCurrent || len(g.Offline) != 0 {
		t.Fatalf("%+v %v", g, e)
	}
}

func TestMySQLLegacyQRPinsMode(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	secret, e := s.CreateQR(ctx, "a", "legacy", 100)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.ConfirmQR(ctx, "a", "legacy", "u", 101); e != nil {
		t.Fatal(e)
	}
	sign := func(string) (string, error) { return "legacy-token", nil }
	first, e := s.RedeemQR(ctx, "a", "legacy", DeviceParams{}, "", key, 102, LoginResult{UserID: "u"}, sign)
	if e != nil {
		t.Fatal(e)
	}
	again, e := s.RedeemQR(ctx, "a", "legacy", DeviceParams{}, "", key, 103, LoginResult{}, sign)
	if e != nil || first != again {
		t.Fatal(e)
	}
	if _, e = s.RedeemQR(ctx, "a", "legacy", params("d"), secret, key, 104, LoginResult{}, sign); e == nil {
		t.Fatal("legacy mode upgraded")
	}
	var n int64
	s.DB.Model(&Device{}).Count(&n)
	if n != 0 {
		t.Fatal("legacy fabricated a device")
	}
}

type verifiedIM struct{}

func (verifiedIM) Online(context.Context, string, string) (map[string]bool, error) {
	return map[string]bool{"d": true}, nil
}
func (verifiedIM) Reconcile(context.Context, RevokeJob) (bool, error) { return true, nil }
func (verifiedIM) Disconnect(context.Context, RevokeJob) error        { return nil }
func TestMySQLQRRetryAndRevocation(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	secret, e := s.CreateQR(ctx, "a", "qr", 100)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.ConfirmQR(ctx, "a", "qr", "u", 101); e != nil {
		t.Fatal(e)
	}
	r, e := s.RedeemQR(ctx, "a", "qr", params("d"), secret, key, 102, LoginResult{UserID: "u", IMToken: "im-secret"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	again, e := s.RedeemQR(ctx, "a", "qr", params("d"), secret, key, 103, LoginResult{}, nil)
	if e != nil || r != again {
		t.Fatalf("retry changed: %+v %v", again, e)
	}
	if e = s.ConfirmQR(ctx, "a", "qr", "u", 104); e != nil {
		t.Fatal(e)
	}
	if _, e = s.RedeemQR(ctx, "a", "qr", params("other"), secret, key, 104, LoginResult{}, nil); e == nil {
		t.Fatal("different device claimed")
	}
	if _, e = s.RedeemQR(ctx, "a", "qr", DeviceParams{}, "", key, 104, LoginResult{}, nil); e == nil {
		t.Fatal("downgraded")
	}
	caller := Claims{AppKey: "a", UserID: "u", DeviceID: "current", SessionID: "current", Kind: "user_session"}
	_, _ = s.Delete(ctx, caller, "d", 105)
	if _, e = s.RedeemQR(ctx, "a", "qr", params("d"), secret, key, 106, LoginResult{}, nil); e == nil {
		t.Fatal("QR reactivated revoked session")
	}
	var q QRRecord
	if e = s.DB.Take(&q).Error; e != nil {
		t.Fatal(e)
	}
	if q.ConsumedTime != 102 {
		t.Fatal(q.ConsumedTime)
	}
	if e = s.Recover(ctx, 600100); e != nil {
		t.Fatal(e)
	}
	if e = s.DB.Take(&q).Error; e != nil || q.ResultCipher != "" {
		t.Fatal("expired result retained", e)
	}
}
func TestMySQLConcurrentFirstLoginAndDelete(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = s.Issue(ctx, "a", "u", params("d"), key, 100) }()
	}
	wg.Wait()
	var count int64
	s.DB.Model(&Device{}).Count(&count)
	if count != 1 {
		t.Fatal(count)
	}
	caller := Claims{AppKey: "a", UserID: "u", DeviceID: "current", SessionID: "current", Kind: "user_session"}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = s.Delete(ctx, caller, "d", 200) }()
	}
	wg.Wait()
	s.DB.Model(&RevokeJob{}).Count(&count)
	if count != 1 {
		t.Fatal("conflicting revoke operations", count)
	}
	if _, e := s.Issue(ctx, "a", "u", params("d"), key, 201); e != nil {
		t.Fatal("immediate login blocked", e)
	}
}

func TestMySQLLoginDeleteRaceSerializesSession(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	if _, e := s.Issue(ctx, "a", "u", params("d"), key, 100); e != nil {
		t.Fatal(e)
	}
	caller := Claims{AppKey: "a", UserID: "u", DeviceID: "current", SessionID: "current", Kind: "user_session"}
	start := make(chan struct{})
	var wg sync.WaitGroup
	var deleteErr error
	wg.Add(2)
	go func() { defer wg.Done(); <-start; _, _ = s.Issue(ctx, "a", "u", params("d"), key, 200) }()
	go func() { defer wg.Done(); <-start; _, deleteErr = s.Delete(ctx, caller, "d", 200) }()
	close(start)
	wg.Wait()
	if deleteErr != nil && !errors.Is(deleteErr, ErrIM) {
		t.Fatal(deleteErr)
	}
	var d Device
	if e := s.DB.Take(&d).Error; e != nil {
		t.Fatal(e)
	}
	if (d.State != "revoking" && d.State != "active") || d.BannedUntil != 0 {
		t.Fatalf("invalid race result: %+v", d)
	}
}

type blockingKickIM struct {
	entered, release chan struct{}
	calls            int
}

func (b *blockingKickIM) Online(context.Context, string, string) (map[string]bool, error) {
	return map[string]bool{}, nil
}
func (b *blockingKickIM) Reconcile(context.Context, RevokeJob) (bool, error) { return false, nil }
func (b *blockingKickIM) Disconnect(ctx context.Context, j RevokeJob) error {
	b.calls++
	close(b.entered)
	select {
	case <-b.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestMySQLKickAndLoginSerializeAcrossConnections(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	old, e := s.Issue(ctx, "a", "u", params("d"), key, 100)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := Verify(old, key, "a")
	adapter := &blockingKickIM{entered: make(chan struct{}), release: make(chan struct{})}
	s.IM = adapter
	deleted := make(chan error, 1)
	go func() {
		_, e := s.Delete(ctx, Claims{AppKey: "a", UserID: "u", DeviceID: "current"}, "d", 101)
		deleted <- e
	}()
	select {
	case <-adapter.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("kick did not start")
	}
	// Independent pool simulates another server process against the same MySQL.
	db, e := gorm.Open(mysql.Open("root@tcp(127.0.0.1:13316)/deviceauth_test?charset=utf8mb4&parseTime=true"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	pool, _ := db.DB()
	defer pool.Close()
	other := Store{DB: db, IM: UnsupportedIM{}}
	logged := make(chan error, 1)
	started := make(chan struct{})
	go func() { close(started); _, e := other.Issue(ctx, "a", "u", params("d"), key, 102); logged <- e }()
	<-started
	select {
	case e := <-logged:
		close(adapter.release)
		t.Fatalf("login escaped kick lock: %v", e)
	case <-time.After(100 * time.Millisecond):
	}
	close(adapter.release)
	if e := <-deleted; e != nil {
		t.Fatal(e)
	}
	if e := <-logged; e != nil {
		t.Fatal(e)
	}
	if e := s.Authenticate(ctx, c, 103); !errors.Is(e, ErrAuth) {
		t.Fatal("old token revived", e)
	}
	if adapter.calls != 1 {
		t.Fatal(adapter.calls)
	}
}

func TestMySQLNewLoginSupersedesExpiredHistoricalJob(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	old, e := s.Issue(ctx, "a", "u", params("d"), key, 100)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := Verify(old, key, "a")
	j, e := s.Delete(ctx, Claims{AppKey: "a", UserID: "u", DeviceID: "current"}, "d", 101)
	if !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	if e = s.DB.Model(&RevokeJob{}).Where("operation_id=?", j.OperationID).Update("banned_until", 1).Error; e != nil {
		t.Fatal(e)
	}
	if e = s.DB.Exec("UPDATE user_devices SET banned_until=999999 WHERE device_id=?", "d").Error; e != nil {
		t.Fatal(e)
	}
	fresh, e := s.Issue(ctx, "a", "u", params("d"), key, 102)
	if e != nil {
		t.Fatal(e)
	}
	// Unsupported adapter proves the obsolete job makes no remote call.
	if e = s.Process(ctx, j, 9999999); e != nil {
		t.Fatal(e)
	}
	var persisted RevokeJob
	if e = s.DB.Where("operation_id=?", j.OperationID).Take(&persisted).Error; e != nil || !persisted.Done || persisted.LastError != "superseded" {
		t.Fatal(persisted, e)
	}
	nc, _ := Verify(fresh, key, "a")
	if e = s.Authenticate(ctx, nc, 103); e != nil {
		t.Fatal(e)
	}
	var historical Device
	if e = s.DB.Where("device_id=?", "d").Take(&historical).Error; e != nil || historical.BannedUntil != 999999 {
		t.Fatal("historical ban column was overwritten", historical, e)
	}
	if e = s.Authenticate(ctx, c, 9999999); !errors.Is(e, ErrAuth) {
		t.Fatal("old token revived", e)
	}
	newer, e := s.Delete(ctx, Claims{AppKey: "a", UserID: "u", DeviceID: "current"}, "d", 104)
	if !errors.Is(e, ErrIM) || newer.OperationID == j.OperationID || newer.TargetSessionID == j.TargetSessionID {
		t.Fatal("new delete reused old target", newer, e)
	}
}

func TestMySQLRecoverExpiredHistoricalTarget(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	if _, e := s.Issue(ctx, "a", "u", params("d"), []byte("key"), 100); e != nil {
		t.Fatal(e)
	}
	j, e := s.Delete(ctx, Claims{AppKey: "a", UserID: "u", DeviceID: "current"}, "d", 101)
	if !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	if e = s.DB.Model(&RevokeJob{}).Where("operation_id=?", j.OperationID).Updates(map[string]interface{}{"banned_until": 1, "next_attempt": 102}).Error; e != nil {
		t.Fatal(e)
	}
	s.IM = verifiedIM{}
	if e = s.Recover(ctx, 9999999); e != nil {
		t.Fatal(e)
	}
	var result RevokeJob
	if e = s.DB.Where("operation_id=?", j.OperationID).Take(&result).Error; e != nil || !result.Done || result.LastError != "" {
		t.Fatal(result, e)
	}
	var d Device
	if e = s.DB.Take(&d).Error; e != nil || d.State != "revoked" {
		t.Fatal(d, e)
	}
}

func TestMySQLConcurrentWorkersKickOnce(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	if _, e := s.Issue(ctx, "a", "u", params("d"), []byte("key"), 100); e != nil {
		t.Fatal(e)
	}
	j, e := s.Delete(ctx, Claims{AppKey: "a", UserID: "u", DeviceID: "current"}, "d", 101)
	if !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	adapter := &blockingKickIM{entered: make(chan struct{}), release: make(chan struct{})}
	s.IM = adapter
	results := make(chan error, 2)
	go func() { results <- s.Process(ctx, j, 102) }()
	select {
	case <-adapter.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("no kick")
	}
	go func() { results <- s.Process(ctx, j, 102) }()
	close(adapter.release)
	for i := 0; i < 2; i++ {
		if e := <-results; e != nil {
			t.Fatal(e)
		}
	}
	if adapter.calls != 1 {
		t.Fatal("duplicate kick", adapter.calls)
	}
}

func TestMySQLQRNewSessionSupersedesPendingKick(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	key := []byte("key")
	if _, e := s.Issue(ctx, "a", "u", params("d"), key, 100); e != nil {
		t.Fatal(e)
	}
	j, e := s.Delete(ctx, Claims{AppKey: "a", UserID: "u", DeviceID: "current"}, "d", 101)
	if !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	secret, e := s.CreateQR(ctx, "a", "fresh-qr", 102)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.ConfirmQR(ctx, "a", "fresh-qr", "u", 103); e != nil {
		t.Fatal(e)
	}
	r, e := s.RedeemQR(ctx, "a", "fresh-qr", params("d"), secret, key, 104, LoginResult{UserID: "u"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Process(ctx, j, 105); e != nil {
		t.Fatal("obsolete QR target called unsupported IM", e)
	}
	c, e := Verify(r.Authorization, key, "a")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Authenticate(ctx, c, 106); e != nil {
		t.Fatal(e)
	}
}

func TestMySQLConcurrentQRClaimHasSingleWinner(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	secret, e := s.CreateQR(ctx, "a", "race", 100)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.ConfirmQR(ctx, "a", "race", "u", 101); e != nil {
		t.Fatal(e)
	}
	start := make(chan struct{})
	out := make(chan error, 2)
	for _, id := range []string{"one", "two"} {
		go func(id string) {
			<-start
			_, e := s.RedeemQR(ctx, "a", "race", params(id), secret, []byte("key"), 102, LoginResult{UserID: "u"}, nil)
			out <- e
		}(id)
	}
	close(start)
	success := 0
	for i := 0; i < 2; i++ {
		if <-out == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatal("QR winners", success)
	}
	var n int64
	s.DB.Model(&Device{}).Count(&n)
	if n != 1 {
		t.Fatal("loser created device", n)
	}
}
