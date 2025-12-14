package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// NewTestDB creates an in-memory database for testing with migrations
func NewTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db
}

// RunTestMigrations runs all migrations on a test database
func RunTestMigrations(t *testing.T, db *sql.DB, migrations []Migration) {
	migrator := NewMigrator(db)
	for _, m := range migrations {
		migrator.Register(m)
	}

	if err := migrator.Run(); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}
}

// RunTestMigrationsWithCleanup runs migrations and registers cleanup
func RunTestMigrationsWithCleanup(t *testing.T, db *sql.DB, migrations []Migration) {
	RunTestMigrations(t, db, migrations)
	t.Cleanup(func() {
		db.Close()
	})
}
