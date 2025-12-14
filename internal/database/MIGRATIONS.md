# Database Migrations - Laravel Style

Sistem migrasi database ini terinspirasi dari Laravel migrations, dengan tracking history dan support untuk rollback.

## Fitur

- ✅ **Migration Tracking**: Semua migration yang dijalankan dicatat dalam tabel `migrations`
- ✅ **Batch System**: Migration dijalankan dalam batch, memudahkan rollback
- ✅ **Up/Down Methods**: Setiap migration memiliki method Up (apply) dan Down (rollback)
- ✅ **Timestamp Naming**: Migration files menggunakan timestamp prefix untuk ordering
- ✅ **Transaction Support**: Setiap migration dijalankan dalam database transaction
- ✅ **Status Command**: Lihat status semua migrations (Ran/Pending)

## Struktur Migration

Setiap migration file harus implement interface `Migration`:

```go
type Migration interface {
    Name() string           // Nama migration dengan timestamp
    Up(tx *sql.Tx) error   // Jalankan migration
    Down(tx *sql.Tx) error // Rollback migration
}
```

## Cara Membuat Migration Baru

1. Buat file baru di `internal/database/migrations/` dengan format:
   ```
   YYYYMMDDHHMMSS_nama_migration.go
   ```

2. Implement interface Migration:

```go
package migrations

import "database/sql"

type CreateExampleTable struct{}

func (m *CreateExampleTable) Name() string {
    return "20240115120000_create_example_table"
}

func (m *CreateExampleTable) Up(tx *sql.Tx) error {
    query := `
        CREATE TABLE IF NOT EXISTS example (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        )
    `
    _, err := tx.Exec(query)
    return err
}

func (m *CreateExampleTable) Down(tx *sql.Tx) error {
    _, err := tx.Exec("DROP TABLE IF EXISTS example")
    return err
}
```

3. Register migration di `internal/database/migrations/registry.go`:

```go
func GetAll() []database.Migration {
    return []database.Migration{
        &CreateUsersTable{},
        &CreateProjectsTable{},
        // ... existing migrations ...
        &CreateExampleTable{}, // ← tambahkan disini
    }
}
```

## Menjalankan Migrations

Migrations otomatis dijalankan saat server start. Atau bisa manual:

```go
import (
    "central-logs/internal/database"
    "central-logs/internal/database/migrations"
)

// Connect to database
db, err := database.New("path/to/db.sqlite")
if err != nil {
    log.Fatal(err)
}

// Run all pending migrations
err = db.MigrateWithRegistry(migrations.GetAll())
if err != nil {
    log.Fatal(err)
}
```

## Rollback Migrations

Rollback batch terakhir:

```go
err = db.RollbackWithRegistry(migrations.GetAll())
if err != nil {
    log.Fatal(err)
}
```

## Migration Status

Lihat status semua migrations:

```go
err = db.MigrationStatusWithRegistry(migrations.GetAll())
// Output:
// Migration Status:
// ================
// [Ran] 20240101000001_create_users_table
// [Ran] 20240101000002_create_projects_table
// [Pending] 20240115120000_create_example_table
```

## Penggunaan dalam Tests

Gunakan helper functions untuk testing:

```go
import (
    "testing"
    "central-logs/internal/database"
    "central-logs/internal/database/migrations"
)

func setupTestDB(t *testing.T) *sql.DB {
    db := database.NewTestDB(t)
    database.RunTestMigrations(t, db, migrations.GetAll())
    return db
}

func TestSomething(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Your test code here...
}
```

## Migration Table

Tabel `migrations` otomatis dibuat dengan struktur:

```sql
CREATE TABLE migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration TEXT NOT NULL UNIQUE,  -- nama migration
    batch INTEGER NOT NULL,           -- nomor batch
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)
```

## Best Practices

1. **Never modify existing migrations** yang sudah dijalankan di production
2. **Always create new migrations** untuk perubahan schema
3. **Test rollback** sebelum deploy ke production
4. **Use transactions** untuk memastikan atomic operations
5. **Keep migrations small** dan focused pada satu perubahan
6. **Add indexes** di migration terpisah dari table creation

## Contoh Migration Patterns

### Menambah Kolom

```go
func (m *AddColumnToTable) Up(tx *sql.Tx) error {
    _, err := tx.Exec("ALTER TABLE users ADD COLUMN phone TEXT")
    return err
}

func (m *AddColumnToTable) Down(tx *sql.Tx) error {
    // SQLite doesn't support DROP COLUMN, need to recreate table
    // Or just leave it (depends on your requirements)
    return nil
}
```

### Membuat Index

```go
func (m *CreateUserEmailIndex) Up(tx *sql.Tx) error {
    _, err := tx.Exec("CREATE INDEX idx_users_email ON users(email)")
    return err
}

func (m *CreateUserEmailIndex) Down(tx *sql.Tx) error {
    _, err := tx.Exec("DROP INDEX IF EXISTS idx_users_email")
    return err
}
```

### Seed Data

```go
func (m *SeedDefaultRoles) Up(tx *sql.Tx) error {
    _, err := tx.Exec(`
        INSERT INTO roles (id, name) VALUES
        ('1', 'admin'),
        ('2', 'user')
    `)
    return err
}

func (m *SeedDefaultRoles) Down(tx *sql.Tx) error {
    _, err := tx.Exec("DELETE FROM roles WHERE id IN ('1', '2')")
    return err
}
```

## Troubleshooting

### Migration sudah jalan tapi ingin jalan ulang

Hapus record dari tabel migrations:

```sql
DELETE FROM migrations WHERE migration = '20240115120000_create_example_table';
```

### Reset semua migrations

```sql
DROP TABLE migrations;
```

Lalu jalankan ulang migrations.

### Rollback tidak berfungsi

Pastikan method `Down()` sudah diimplement dengan benar dan migration sudah terdaftar di registry.
