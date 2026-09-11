package main

import (
	"database/sql"
	"testing"
)

func newTestSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("unable to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := Migrate(db); err != nil {
		t.Fatalf("unable to migrate test database: %v", err)
	}
	return db
}

func TestSQLiteStore(t *testing.T) {
	testGameStatsStore(t, func(t *testing.T) GameStatsStore {
		return NewSQLiteStore(newTestSQLiteDB(t))
	})
}
