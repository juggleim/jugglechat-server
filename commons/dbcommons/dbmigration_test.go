package dbcommons

import (
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"testing"
)

var errMigrationFixture = errors.New("statement failed")

func TestPasswordSMSMigrationMySQL80Syntax(t *testing.T) {
	b, e := sqlFs.ReadFile("sqls/20261109.sql")
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(strings.ToUpper(string(b)), "CREATE TRIGGER IF NOT EXISTS") {
		t.Fatal("migration requires MySQL 8.0.29")
	}
}

func TestMigrationStopsOnStatementFailure(t *testing.T) {
	calls := 0
	err := runSQLStatements("CREATE TABLE a(id INT);\nINVALID;\nCREATE TABLE b(id INT);", func(string) error {
		calls++
		if calls == 2 {
			return errMigrationFixture
		}
		return nil
	})
	if err == nil || calls != 2 {
		t.Fatalf("migration ignored failure: calls=%d err=%v", calls, err)
	}
}

func TestPasswordSMSMigrationMySQL(t *testing.T) {
	if os.Getenv("PASSWORD_SMS_MIGRATION_TEST_MYSQL") != "1" {
		t.Skip("isolated localhost:13316/password_sms_migration_test")
	}
	conn, e := gorm.Open(mysql.Open("root@tcp(127.0.0.1:13316)/password_sms_migration_test?parseTime=true"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	old := db
	db = conn
	defer func() { db = old; sqlDB, _ := conn.DB(); sqlDB.Close() }()
	for _, table := range []string{"password_sms_challenges", "password_sms_quotas", "users"} {
		if e = conn.Exec("DROP TABLE IF EXISTS " + table).Error; e != nil {
			t.Fatal(e)
		}
	}
	if e = conn.Exec("CREATE TABLE users(id bigint PRIMARY KEY,phone varchar(32)) ENGINE=InnoDB").Error; e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 2; i++ {
		if e = executeSqlFile("sqls/20261109.sql"); e != nil {
			t.Fatal(e)
		}
	}
	conn.Exec("INSERT INTO users(id,phone) VALUES (1,'13800138000')")
	conn.Exec("UPDATE users SET phone='13900139000' WHERE id=1")
	conn.Exec("UPDATE users SET phone='13800138000' WHERE id=1")
	var version int64
	if e = conn.Raw("SELECT password_sms_phone_version FROM users WHERE id=1").Scan(&version).Error; e != nil {
		t.Fatal(e)
	}
	if version != 2 {
		t.Fatalf("phone change version=%d", version)
	}
	if e = conn.Exec("DROP TRIGGER users_password_sms_phone_version").Error; e != nil {
		t.Fatal(e)
	}
	if e = conn.Exec("ALTER TABLE users DROP COLUMN password_sms_phone_version").Error; e != nil {
		t.Fatal(e)
	}
	if e = EnsurePasswordSMSPhoneTrigger(conn); e == nil {
		t.Fatal("trigger creation failure ignored")
	}
	if e = executeSqlFile("sqls/20261109.sql"); e != nil {
		t.Fatal("failed migration cannot retry", e)
	}
}
