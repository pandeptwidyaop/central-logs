package migrations

import "database/sql"

type CreateNotificationHistoryTable struct{}

func (m *CreateNotificationHistoryTable) Name() string {
	return "20240101000007_create_notification_history_table"
}

func (m *CreateNotificationHistoryTable) Up(tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS notification_history (
			id TEXT PRIMARY KEY,
			log_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			error_message TEXT,
			sent_at DATETIME,
			FOREIGN KEY (log_id) REFERENCES logs(id) ON DELETE CASCADE,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		)
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateNotificationHistoryTable) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS notification_history")
	return err
}
