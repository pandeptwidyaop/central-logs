package migrations

import "database/sql"

type CreateMCPTokensTable struct{}

func (m *CreateMCPTokensTable) Name() string {
	return "20250120000001_create_mcp_tokens_table"
}

func (m *CreateMCPTokensTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS mcp_tokens (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			token_prefix TEXT NOT NULL,
			granted_projects TEXT,
			expires_at DATETIME,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_by TEXT NOT NULL,
			last_used_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
		)
	`
	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	// Create indexes
	indexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_mcp_tokens_token_hash ON mcp_tokens(token_hash)",
		"CREATE INDEX IF NOT EXISTS idx_mcp_tokens_created_by ON mcp_tokens(created_by)",
	}

	for _, indexQuery := range indexQueries {
		if _, err := tx.Exec(indexQuery); err != nil {
			return err
		}
	}

	return nil
}

func (m *CreateMCPTokensTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS mcp_tokens")
	return err
}
