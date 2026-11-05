package dbcommons

import (
	"errors"
	"testing"
)

var errMigrationFixture = errors.New("statement failed")

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
