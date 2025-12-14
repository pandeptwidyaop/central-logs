package migrations

import "database/sql"

type CreateChannelsTable struct{}

func (m *CreateChannelsTable) Name() string {
	return "20240101000005_create_channels_table"
}

func (m *CreateChannelsTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS channels (
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
		)
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateChannelsTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS channels")
	return err
}
