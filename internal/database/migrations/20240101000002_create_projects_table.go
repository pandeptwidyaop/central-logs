package migrations

import "database/sql"

type CreateProjectsTable struct{}

func (m *CreateProjectsTable) Name() string {
	return "20240101000002_create_projects_table"
}

func (m *CreateProjectsTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			api_key TEXT NOT NULL,
			api_key_prefix TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			retention_config TEXT,
			icon_type TEXT DEFAULT 'initials',
			icon_value TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateProjectsTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS projects")
	return err
}
