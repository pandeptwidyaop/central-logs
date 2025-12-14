package migrations

import "database/sql"

type CreateLogsTable struct{}

func (m *CreateLogsTable) Name() string {
	return "20240101000004_create_logs_table"
}

func (m *CreateLogsTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS logs (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata TEXT,
			source TEXT,
			timestamp DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateLogsTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS logs")
	return err
}
