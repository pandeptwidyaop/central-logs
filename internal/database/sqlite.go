package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

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

func (db *DB) Migrate() error {
	// Run migrations that may fail (like ALTER TABLE for existing columns)
	alterMigrations := []string{
		`ALTER TABLE users ADD COLUMN two_factor_secret TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN two_factor_enabled INTEGER DEFAULT 0`,
		`ALTER TABLE users ADD COLUMN backup_codes TEXT DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN icon_type TEXT DEFAULT 'initials'`,
		`ALTER TABLE projects ADD COLUMN icon_value TEXT DEFAULT ''`,
	}
	for _, migration := range alterMigrations {
		// Ignore errors for ALTER TABLE (column may already exist)
		db.Exec(migration)
	}

	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'USER',
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		// Projects table
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			api_key TEXT NOT NULL,
			api_key_prefix TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			retention_config TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		// User-Projects pivot table
		`CREATE TABLE IF NOT EXISTS user_projects (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'MEMBER',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			UNIQUE(user_id, project_id)
		)`,

		// Logs table
		`CREATE TABLE IF NOT EXISTS logs (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata TEXT,
			source TEXT,
			timestamp DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,

		// Channels table
		`CREATE TABLE IF NOT EXISTS channels (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			config TEXT NOT NULL,
			min_level TEXT NOT NULL DEFAULT 'ERROR',
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,

		// Push subscriptions table
		`CREATE TABLE IF NOT EXISTS push_subscriptions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			project_id TEXT,
			endpoint TEXT NOT NULL UNIQUE,
			p256dh TEXT NOT NULL,
			auth TEXT NOT NULL,
			user_agent TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,

		// Notification history table
		`CREATE TABLE IF NOT EXISTS notification_history (
			id TEXT PRIMARY KEY,
			log_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			error_message TEXT,
			sent_at DATETIME,
			FOREIGN KEY (log_id) REFERENCES logs(id) ON DELETE CASCADE,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		)`,

		// Indexes for better query performance
		`CREATE INDEX IF NOT EXISTS idx_logs_project_id ON logs(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_project_level ON logs(project_id, level)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_project_created ON logs(project_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_user_projects_user_id ON user_projects(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_projects_project_id ON user_projects(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_project_id ON channels(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_history_log_id ON notification_history(log_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_history_status ON notification_history(status)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
