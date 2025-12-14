package migrations

import "database/sql"

type CreateUserProjectsTable struct{}

func (m *CreateUserProjectsTable) Name() string {
	return "20240101000003_create_user_projects_table"
}

func (m *CreateUserProjectsTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_projects (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'MEMBER',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			UNIQUE(user_id, project_id)
		)
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateUserProjectsTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS user_projects")
	return err
}
