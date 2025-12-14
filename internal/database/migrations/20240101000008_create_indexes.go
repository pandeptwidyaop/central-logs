package migrations

import "database/sql"

type CreateIndexes struct{}

func (m *CreateIndexes) Name() string {
	return "20240101000008_create_indexes"
}

func (m *CreateIndexes) Up(tx *sql.Tx) error {
	indexes := []string{
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

	for _, index := range indexes {
		if _, err := tx.Exec(index); err != nil {
			return err
		}
	}

	return nil
}

func (m *CreateIndexes) Down(tx *sql.Tx) error {
	indexes := []string{
		`DROP INDEX IF EXISTS idx_logs_project_id`,
		`DROP INDEX IF EXISTS idx_logs_level`,
		`DROP INDEX IF EXISTS idx_logs_created_at`,
		`DROP INDEX IF EXISTS idx_logs_project_level`,
		`DROP INDEX IF EXISTS idx_logs_project_created`,
		`DROP INDEX IF EXISTS idx_user_projects_user_id`,
		`DROP INDEX IF EXISTS idx_user_projects_project_id`,
		`DROP INDEX IF EXISTS idx_channels_project_id`,
		`DROP INDEX IF EXISTS idx_notification_history_log_id`,
		`DROP INDEX IF EXISTS idx_notification_history_status`,
	}

	for _, index := range indexes {
		if _, err := tx.Exec(index); err != nil {
			return err
		}
	}

	return nil
}
