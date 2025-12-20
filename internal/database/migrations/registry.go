package migrations

import "central-logs/internal/database"

// GetAll returns all migrations in order
func GetAll() []database.Migration {
	return []database.Migration{
		&CreateUsersTable{},
		&CreateProjectsTable{},
		&CreateUserProjectsTable{},
		&CreateLogsTable{},
		&CreateChannelsTable{},
		&CreatePushSubscriptionsTable{},
		&CreateNotificationHistoryTable{},
		&CreateIndexes{},
		&CreateMCPTokensTable{},
		&CreateMCPActivityLogsTable{},
	}
}
