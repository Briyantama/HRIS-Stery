package testdb_test

import (
	"testing"

	"github.com/Briyantama/HRIS-Stery/internal/testdb"
)

func TestNewMigrated_returnsIndependentMigratedDatabases(t *testing.T) {
	// Arrange
	first := testdb.NewMigrated(t)
	second := testdb.NewMigrated(t)

	// Act
	if _, err := first.Exec(`CREATE SCHEMA only_in_first`); err != nil {
		t.Fatalf("create schema in the first database: %v", err)
	}
	var inFirst, inSecond, bootstrapped, applied int
	_ = first.QueryRow(`SELECT count(*) FROM pg_namespace WHERE nspname = 'only_in_first'`).Scan(&inFirst)
	_ = second.QueryRow(`SELECT count(*) FROM pg_namespace WHERE nspname = 'only_in_first'`).Scan(&inSecond)
	_ = second.QueryRow(`SELECT count(*) FROM pg_namespace WHERE nspname = 'auth'`).Scan(&bootstrapped)
	_ = second.QueryRow(`SELECT count(*) FROM goose_db_version WHERE is_applied`).Scan(&applied)

	// Assert
	if inFirst != 1 || inSecond != 0 {
		t.Errorf("schema in first=%d second=%d, want 1 and 0: the copies must not share state", inFirst, inSecond)
	}
	if bootstrapped != 1 || applied < 1 {
		t.Errorf("auth schema=%d applied=%d: the copy is not migrated", bootstrapped, applied)
	}
}
