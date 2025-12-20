package migrations

import "database/sql"

type CreateMCPActivityLogsTable struct{}

func (m *CreateMCPActivityLogsTable) Name() string {
	return "20250120000002_create_mcp_activity_logs_table"
}

func (m *CreateMCPActivityLogsTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS mcp_activity_logs (
			id TEXT PRIMARY KEY,
			token_id TEXT NOT NULL,
			tool_name TEXT NOT NULL,
			project_ids TEXT,
			request_params TEXT,
			success INTEGER NOT NULL,
			error_message TEXT,
			duration_ms INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (token_id) REFERENCES mcp_tokens(id) ON DELETE CASCADE
		)
	`
	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	// Create indexes
	indexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_mcp_activity_logs_token_id ON mcp_activity_logs(token_id)",
		"CREATE INDEX IF NOT EXISTS idx_mcp_activity_logs_created_at ON mcp_activity_logs(created_at)",
	}

	for _, indexQuery := range indexQueries {
		if _, err := tx.Exec(indexQuery); err != nil {
			return err
		}
	}

	return nil
}

func (m *CreateMCPActivityLogsTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS mcp_activity_logs")
	return err
}
