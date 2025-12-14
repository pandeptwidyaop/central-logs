package database

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration interface - setiap migration harus implement ini
type Migration interface {
	// Name returns the migration name (e.g., "2024_01_01_000000_create_users_table")
	Name() string
	// Up runs the migration
	Up(tx *sql.Tx) error
	// Down rolls back the migration
	Down(tx *sql.Tx) error
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	ID        int
	Migration string
	Batch     int
	CreatedAt time.Time
}

// Migrator handles running and rolling back migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: []Migration{},
	}
}

// Register adds a migration to the migrator
func (m *Migrator) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// createMigrationsTable creates the migrations tracking table if it doesn't exist
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			migration TEXT NOT NULL UNIQUE,
			batch INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := m.db.Exec(query)
	return err
}

// getExecutedMigrations returns a map of migration names that have been executed
func (m *Migrator) getExecutedMigrations() (map[string]bool, error) {
	executed := make(map[string]bool)

	rows, err := m.db.Query("SELECT migration FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		executed[name] = true
	}

	return executed, rows.Err()
}

// getLastBatch returns the last batch number
func (m *Migrator) getLastBatch() (int, error) {
	var batch int
	err := m.db.QueryRow("SELECT COALESCE(MAX(batch), 0) FROM migrations").Scan(&batch)
	return batch, err
}

// getMigrationsByBatch returns migrations for a specific batch
func (m *Migrator) getMigrationsByBatch(batch int) ([]string, error) {
	var migrations []string

	rows, err := m.db.Query("SELECT migration FROM migrations WHERE batch = ? ORDER BY id DESC", batch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations = append(migrations, name)
	}

	return migrations, rows.Err()
}

// recordMigration records a migration as executed
func (m *Migrator) recordMigration(tx *sql.Tx, name string, batch int) error {
	_, err := tx.Exec("INSERT INTO migrations (migration, batch) VALUES (?, ?)", name, batch)
	return err
}

// removeMigrationRecord removes a migration record
func (m *Migrator) removeMigrationRecord(tx *sql.Tx, name string) error {
	_, err := tx.Exec("DELETE FROM migrations WHERE migration = ?", name)
	return err
}

// Run executes all pending migrations
func (m *Migrator) Run() error {
	// Create migrations table
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get executed migrations
	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Sort migrations by name (timestamp)
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Name() < m.migrations[j].Name()
	})

	// Get next batch number
	batch, err := m.getLastBatch()
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}
	batch++

	// Run pending migrations
	pendingCount := 0
	for _, migration := range m.migrations {
		if executed[migration.Name()] {
			continue
		}

		pendingCount++
		fmt.Printf("Migrating: %s\n", migration.Name())

		// Start transaction
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", migration.Name(), err)
		}

		// Run migration
		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", migration.Name(), err)
		}

		// Record migration
		if err := m.recordMigration(tx, migration.Name(), batch); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Name(), err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Name(), err)
		}

		fmt.Printf("Migrated:  %s\n", migration.Name())
	}

	if pendingCount == 0 {
		fmt.Println("Nothing to migrate.")
	}

	return nil
}

// Rollback rolls back the last batch of migrations
func (m *Migrator) Rollback() error {
	// Create migrations table
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get last batch
	batch, err := m.getLastBatch()
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}

	if batch == 0 {
		fmt.Println("Nothing to rollback.")
		return nil
	}

	// Get migrations in the last batch
	migrationNames, err := m.getMigrationsByBatch(batch)
	if err != nil {
		return fmt.Errorf("failed to get migrations for batch %d: %w", batch, err)
	}

	// Create a map of migrations by name
	migrationMap := make(map[string]Migration)
	for _, migration := range m.migrations {
		migrationMap[migration.Name()] = migration
	}

	// Rollback migrations
	for _, name := range migrationNames {
		migration, ok := migrationMap[name]
		if !ok {
			return fmt.Errorf("migration %s not found in registered migrations", name)
		}

		fmt.Printf("Rolling back: %s\n", name)

		// Start transaction
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", name, err)
		}

		// Run rollback
		if err := migration.Down(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("rollback %s failed: %w", name, err)
		}

		// Remove migration record
		if err := m.removeMigrationRecord(tx, name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove migration record %s: %w", name, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit rollback %s: %w", name, err)
		}

		fmt.Printf("Rolled back:  %s\n", name)
	}

	return nil
}

// Status shows the status of all migrations
func (m *Migrator) Status() error {
	// Create migrations table
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get executed migrations
	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Sort migrations by name
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Name() < m.migrations[j].Name()
	})

	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	for _, migration := range m.migrations {
		status := "Pending"
		if executed[migration.Name()] {
			status = "Ran"
		}
		fmt.Printf("[%s] %s\n", status, migration.Name())
	}
	fmt.Println()

	return nil
}
