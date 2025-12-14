package models_test

import (
	"database/sql"
	"testing"
	"time"

	"central-logs/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func setupLogTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create logs and projects table (needed for join)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			api_key TEXT UNIQUE NOT NULL,
			api_key_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS logs (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata TEXT,
			source TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_logs_project_id ON logs(project_id);
		CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
		CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at);
	`)
	if err != nil {
		t.Fatalf("Failed to create logs table: %v", err)
	}

	// Insert a test project
	_, err = db.Exec(`
		INSERT INTO projects (id, name, description, api_key, api_key_hash)
		VALUES ('proj-1', 'Test Project', 'Test', 'test-key', 'hash')
	`)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return db
}

func TestLogRepository_Create(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	log := &models.Log{
		ProjectID: "proj-1",
		Level:     models.LogLevelInfo,
		Message:   "Test log message",
		Source:    "test-source",
	}

	err := repo.Create(log)
	if err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	if log.ID == "" {
		t.Error("Log ID should be set after creation")
	}
}

func TestLogRepository_CreateBatch(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	logs := []*models.Log{
		{ProjectID: "proj-1", Level: models.LogLevelInfo, Message: "Log 1"},
		{ProjectID: "proj-1", Level: models.LogLevelWarn, Message: "Log 2"},
		{ProjectID: "proj-1", Level: models.LogLevelError, Message: "Log 3"},
	}

	err := repo.CreateBatch(logs)
	if err != nil {
		t.Fatalf("Failed to create batch logs: %v", err)
	}

	// Verify all logs were created
	filter := &models.LogFilter{ProjectIDs: []string{"proj-1"}}
	allLogs, total, err := repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected 3 logs, got %d", total)
	}

	_ = allLogs
}

func TestLogRepository_GetByID(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	// Create a log
	log := &models.Log{
		ProjectID: "proj-1",
		Level:     models.LogLevelError,
		Message:   "Test error message",
		Source:    "test-source",
	}
	err := repo.Create(log)
	if err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	// Get by ID
	found, err := repo.GetByID(log.ID)
	if err != nil {
		t.Fatalf("Failed to get log by ID: %v", err)
	}

	if found == nil {
		t.Fatal("Log should be found")
	}

	if found.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", found.Message)
	}

	if found.Level != models.LogLevelError {
		t.Errorf("Expected level ERROR, got %s", found.Level)
	}
}

func TestLogRepository_List_WithFilter(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	// Add second project
	db.Exec(`INSERT INTO projects (id, name, description, api_key, api_key_hash) VALUES ('proj-2', 'Project 2', 'Test', 'key-2', 'hash2')`)

	repo := models.NewLogRepository(db)

	// Create logs with different levels
	logs := []*models.Log{
		{ProjectID: "proj-1", Level: models.LogLevelDebug, Message: "Debug log"},
		{ProjectID: "proj-1", Level: models.LogLevelInfo, Message: "Info log"},
		{ProjectID: "proj-1", Level: models.LogLevelWarn, Message: "Warn log"},
		{ProjectID: "proj-1", Level: models.LogLevelError, Message: "Error log"},
		{ProjectID: "proj-2", Level: models.LogLevelError, Message: "Other project error"},
	}

	for _, log := range logs {
		err := repo.Create(log)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Filter by level
	filter := &models.LogFilter{Levels: []models.LogLevel{models.LogLevelError}}
	results, total, err := repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 error logs, got %d", total)
	}

	// Filter by project
	filter = &models.LogFilter{ProjectIDs: []string{"proj-1"}}
	results, total, err = repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 4 {
		t.Errorf("Expected 4 logs for project 1, got %d", total)
	}

	// Filter by project and level
	filter = &models.LogFilter{ProjectIDs: []string{"proj-1"}, Levels: []models.LogLevel{models.LogLevelError}}
	results, total, err = repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 error log for project 1, got %d", total)
	}

	_ = results
}

func TestLogRepository_List_WithSearch(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	// Create logs
	logs := []*models.Log{
		{ProjectID: "proj-1", Level: models.LogLevelInfo, Message: "User logged in successfully"},
		{ProjectID: "proj-1", Level: models.LogLevelError, Message: "Failed to connect to database"},
		{ProjectID: "proj-1", Level: models.LogLevelInfo, Message: "User logged out"},
	}

	for _, log := range logs {
		err := repo.Create(log)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Search for "logged"
	filter := &models.LogFilter{Search: "logged"}
	results, total, err := repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 logs with 'logged', got %d", total)
	}

	// Search for "database"
	filter = &models.LogFilter{Search: "database"}
	results, total, err = repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 log with 'database', got %d", total)
	}

	_ = results
}

func TestLogRepository_List_Pagination(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	// Create 15 logs
	for i := 0; i < 15; i++ {
		log := &models.Log{
			ProjectID: "proj-1",
			Level:     models.LogLevelInfo,
			Message:   "Log message",
		}
		err := repo.Create(log)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Page 1 with limit 10
	filter := &models.LogFilter{Limit: 10, Offset: 0}
	results, total, err := repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 results on page 1, got %d", len(results))
	}

	// Page 2
	filter = &models.LogFilter{Limit: 10, Offset: 10}
	results, _, err = repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results on page 2, got %d", len(results))
	}
}

func TestLogLevel_Priority(t *testing.T) {
	tests := []struct {
		level    models.LogLevel
		priority int
	}{
		{models.LogLevelDebug, 0},
		{models.LogLevelInfo, 1},
		{models.LogLevelWarn, 2},
		{models.LogLevelError, 3},
		{models.LogLevelCritical, 4},
	}

	for _, tt := range tests {
		if got := tt.level.Priority(); got != tt.priority {
			t.Errorf("LogLevel(%s).Priority() = %d, want %d", tt.level, got, tt.priority)
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected models.LogLevel
	}{
		{"DEBUG", models.LogLevelDebug},
		{"INFO", models.LogLevelInfo},
		{"WARN", models.LogLevelWarn},
		{"WARNING", models.LogLevelWarn},
		{"ERROR", models.LogLevelError},
		{"CRITICAL", models.LogLevelCritical},
		{"UNKNOWN", models.LogLevelInfo}, // Default
	}

	for _, tt := range tests {
		if got := models.ParseLogLevel(tt.input); got != tt.expected {
			t.Errorf("ParseLogLevel(%s) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestLog_WithMetadata(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	metadata := map[string]interface{}{
		"user_id":    float64(123),
		"request_id": "abc-123",
		"duration":   1.5,
	}

	log := &models.Log{
		ProjectID: "proj-1",
		Level:     models.LogLevelInfo,
		Message:   "Test with metadata",
		Metadata:  metadata,
	}

	err := repo.Create(log)
	if err != nil {
		t.Fatalf("Failed to create log with metadata: %v", err)
	}

	// Retrieve and verify
	found, err := repo.GetByID(log.ID)
	if err != nil {
		t.Fatalf("Failed to get log: %v", err)
	}

	if found.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	if found.Metadata["user_id"] != float64(123) {
		t.Errorf("Expected user_id 123, got %v", found.Metadata["user_id"])
	}
}

func TestLogFilter_TimeRange(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	// Create logs
	log := &models.Log{
		ProjectID: "proj-1",
		Level:     models.LogLevelInfo,
		Message:   "Test log",
	}
	repo.Create(log)

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)
	filter := &models.LogFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	results, total, err := repo.List(filter)
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 log in time range, got %d", total)
	}

	_ = results
}

func TestLogRepository_GetStats(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	// Create logs with different levels
	logs := []*models.Log{
		{ProjectID: "proj-1", Level: models.LogLevelDebug, Message: "Debug"},
		{ProjectID: "proj-1", Level: models.LogLevelInfo, Message: "Info 1"},
		{ProjectID: "proj-1", Level: models.LogLevelInfo, Message: "Info 2"},
		{ProjectID: "proj-1", Level: models.LogLevelWarn, Message: "Warn"},
		{ProjectID: "proj-1", Level: models.LogLevelError, Message: "Error 1"},
		{ProjectID: "proj-1", Level: models.LogLevelError, Message: "Error 2"},
		{ProjectID: "proj-1", Level: models.LogLevelError, Message: "Error 3"},
	}

	for _, log := range logs {
		repo.Create(log)
	}

	stats, err := repo.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	expected := map[string]int{
		"DEBUG": 1,
		"INFO":  2,
		"WARN":  1,
		"ERROR": 3,
	}

	for level, count := range expected {
		if stats[level] != count {
			t.Errorf("Expected %d %s logs, got %d", count, level, stats[level])
		}
	}
}

func TestLogRepository_GetProjectStats(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	// Create logs
	logs := []*models.Log{
		{ProjectID: "proj-1", Level: models.LogLevelInfo, Message: "Info 1"},
		{ProjectID: "proj-1", Level: models.LogLevelError, Message: "Error 1"},
	}

	for _, log := range logs {
		repo.Create(log)
	}

	stats, err := repo.GetProjectStats("proj-1")
	if err != nil {
		t.Fatalf("Failed to get project stats: %v", err)
	}

	if stats["INFO"] != 1 {
		t.Errorf("Expected 1 INFO log, got %d", stats["INFO"])
	}

	if stats["ERROR"] != 1 {
		t.Errorf("Expected 1 ERROR log, got %d", stats["ERROR"])
	}
}

func TestLogRepository_CountByProject(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	repo := models.NewLogRepository(db)

	// Create logs
	for i := 0; i < 5; i++ {
		log := &models.Log{
			ProjectID: "proj-1",
			Level:     models.LogLevelInfo,
			Message:   "Test",
		}
		repo.Create(log)
	}

	count, err := repo.CountByProject("proj-1")
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5, got %d", count)
	}
}
