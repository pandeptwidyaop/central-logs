package migrations

import "database/sql"

type CreateUsersTable struct{}

func (m *CreateUsersTable) Name() string {
	return "20240101000001_create_users_table"
}

func (m *CreateUsersTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'USER',
			is_active INTEGER NOT NULL DEFAULT 1,
			two_factor_secret TEXT DEFAULT '',
			two_factor_enabled INTEGER DEFAULT 0,
			backup_codes TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateUsersTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS users")
	return err
}
