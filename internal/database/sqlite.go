package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Migration interface is defined in migration.go

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// Migrate runs all pending migrations using the new migration system
func (db *DB) Migrate() error {
	return db.MigrateWithRegistry(nil)
}

// MigrateWithRegistry runs migrations with custom registry (for testing)
func (db *DB) MigrateWithRegistry(migrations []Migration) error {
	migrator := NewMigrator(db.DB)

	// If no custom migrations provided, this will be handled by the caller
	// that imports the migrations package
	if migrations != nil {
		for _, m := range migrations {
			migrator.Register(m)
		}
	}

	return migrator.Run()
}

// Rollback rolls back the last batch of migrations
func (db *DB) Rollback() error {
	return db.RollbackWithRegistry(nil)
}

// RollbackWithRegistry rolls back migrations with custom registry
func (db *DB) RollbackWithRegistry(migrations []Migration) error {
	migrator := NewMigrator(db.DB)

	if migrations != nil {
		for _, m := range migrations {
			migrator.Register(m)
		}
	}

	return migrator.Rollback()
}

// MigrationStatus shows the status of all migrations
func (db *DB) MigrationStatus() error {
	return db.MigrationStatusWithRegistry(nil)
}

// MigrationStatusWithRegistry shows migration status with custom registry
func (db *DB) MigrationStatusWithRegistry(migrations []Migration) error {
	migrator := NewMigrator(db.DB)

	if migrations != nil {
		for _, m := range migrations {
			migrator.Register(m)
		}
	}

	return migrator.Status()
}

func (db *DB) Close() error {
	return db.DB.Close()
}
