package migrations

import "database/sql"

type CreatePushSubscriptionsTable struct{}

func (m *CreatePushSubscriptionsTable) Name() string {
	return "20240101000006_create_push_subscriptions_table"
}

func (m *CreatePushSubscriptionsTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS push_subscriptions (
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
		)
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreatePushSubscriptionsTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS push_subscriptions")
	return err
}
