package mcp

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"central-logs/internal/models"

	"github.com/mark3labs/mcp-go/mcp"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		name TEXT NOT NULL,
		role TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		api_key TEXT NOT NULL UNIQUE,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE logs (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		source TEXT,
		metadata TEXT,
		timestamp DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);
	CREATE INDEX idx_logs_project_id ON logs(project_id);
	CREATE INDEX idx_logs_level ON logs(level);
	CREATE INDEX idx_logs_timestamp ON logs(timestamp);

	CREATE TABLE mcp_tokens (
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
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE mcp_activity_logs (
		id TEXT PRIMARY KEY,
		token_id TEXT NOT NULL,
		tool_name TEXT NOT NULL,
		project_ids TEXT,
		request_params TEXT,
		success INTEGER NOT NULL,
		error_message TEXT,
		duration_ms INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// setupTestData creates test fixtures
func setupTestData(t *testing.T, db *sql.DB) (string, string, string, string) {
	// Create test user
	userID := "test-user-1"
	_, err := db.Exec(`INSERT INTO users (id, username, password, name, role) VALUES (?, ?, ?, ?, ?)`,
		userID, "testuser", "password", "Test User", "admin")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test projects
	project1ID := "test-project-1"
	project2ID := "test-project-2"

	_, err = db.Exec(`INSERT INTO projects (id, name, api_key) VALUES (?, ?, ?)`,
		project1ID, "Project 1", "key1")
	if err != nil {
		t.Fatalf("Failed to create project 1: %v", err)
	}

	_, err = db.Exec(`INSERT INTO projects (id, name, api_key) VALUES (?, ?, ?)`,
		project2ID, "Project 2", "key2")
	if err != nil {
		t.Fatalf("Failed to create project 2: %v", err)
	}

	// Create test logs
	now := time.Now()
	logs := []struct {
		id        string
		projectID string
		level     string
		message   string
	}{
		{"log-1", project1ID, "info", "Test info log 1"},
		{"log-2", project1ID, "error", "Test error log 1"},
		{"log-3", project1ID, "debug", "Test debug log 1"},
		{"log-4", project2ID, "info", "Test info log 2"},
		{"log-5", project2ID, "warn", "Test warn log 2"},
	}

	for _, log := range logs {
		_, err := db.Exec(`INSERT INTO logs (id, project_id, level, message, source, timestamp) VALUES (?, ?, ?, ?, ?, ?)`,
			log.id, log.projectID, log.level, log.message, "test-source", now)
		if err != nil {
			t.Fatalf("Failed to create log %s: %v", log.id, err)
		}
	}

	return userID, project1ID, project2ID, "log-1"
}

// createTestToken creates a test MCP token
func createTestToken(t *testing.T, db *sql.DB, userID string, grantedProjects string) (*models.MCPToken, string) {
	repo := models.NewMCPTokenRepository(db)

	token := &models.MCPToken{
		Name:            "Test Token",
		GrantedProjects: grantedProjects,
		IsActive:        true,
		CreatedBy:       userID,
	}

	rawToken, err := repo.Create(token)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	return token, rawToken
}

// createMockRequest creates a mock CallToolRequest
func createMockRequest(params map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}
}

// TestHandleGetLog tests the handleGetLog tool
func TestHandleGetLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID, _, _, logID := setupTestData(t, db)
	token, _ := createTestToken(t, db, userID, "*") // All projects access

	// Setup server
	server := &MCPServer{
		mcpTokenRepo:    models.NewMCPTokenRepository(db),
		mcpActivityRepo: models.NewMCPActivityLogRepository(db),
		logRepo:         models.NewLogRepository(db),
		projectRepo:     models.NewProjectRepository(db),
		userRepo:        models.NewUserRepository(db),
	}

	// Test successful retrieval
	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"log_id": logID,
		})

		result, err := server.handleGetLog(ctx, request)
		if err != nil {
			t.Fatalf("handleGetLog returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test access denied
	t.Run("AccessDenied", func(t *testing.T) {
		// Create token with access to project2 only
		token2, _ := createTestToken(t, db, userID, `["test-project-2"]`)
		ctx := context.WithValue(context.Background(), "mcp_token", token2)

		request := createMockRequest(map[string]interface{}{
			"log_id": logID, // This log belongs to project1
		})

		result, err := server.handleGetLog(ctx, request)
		if err != nil {
			t.Fatalf("handleGetLog returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for access denied")
		}
	})

	// Test log not found
	t.Run("NotFound", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"log_id": "non-existent-log",
		})

		result, err := server.handleGetLog(ctx, request)
		if err != nil {
			t.Fatalf("handleGetLog returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for not found")
		}
	})

	// Test missing parameter
	t.Run("MissingParameter", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{})

		result, err := server.handleGetLog(ctx, request)
		if err != nil {
			t.Fatalf("handleGetLog returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for missing parameter")
		}
	})

	// Verify activity was logged
	activityRepo := models.NewMCPActivityLogRepository(db)
	activities, _, err := activityRepo.GetByTokenID(token.ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get activities: %v", err)
	}

	if len(activities) == 0 {
		t.Errorf("Expected activities to be logged")
	}
}

// TestHandleListProjects tests the handleListProjects tool
func TestHandleListProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID, _, _, _ := setupTestData(t, db)

	server := &MCPServer{
		mcpTokenRepo:    models.NewMCPTokenRepository(db),
		mcpActivityRepo: models.NewMCPActivityLogRepository(db),
		logRepo:         models.NewLogRepository(db),
		projectRepo:     models.NewProjectRepository(db),
		userRepo:        models.NewUserRepository(db),
	}

	// Test with all projects access
	t.Run("AllProjects", func(t *testing.T) {
		token, _ := createTestToken(t, db, userID, "*")
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{})

		result, err := server.handleListProjects(ctx, request)
		if err != nil {
			t.Fatalf("handleListProjects returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test with specific projects access
	t.Run("SpecificProjects", func(t *testing.T) {
		token, _ := createTestToken(t, db, userID, `["test-project-1"]`)
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{})

		result, err := server.handleListProjects(ctx, request)
		if err != nil {
			t.Fatalf("handleListProjects returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})
}

// TestHandleGetProject tests the handleGetProject tool
func TestHandleGetProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID, project1ID, _, _ := setupTestData(t, db)
	token, _ := createTestToken(t, db, userID, "*")

	server := &MCPServer{
		mcpTokenRepo:    models.NewMCPTokenRepository(db),
		mcpActivityRepo: models.NewMCPActivityLogRepository(db),
		logRepo:         models.NewLogRepository(db),
		projectRepo:     models.NewProjectRepository(db),
		userRepo:        models.NewUserRepository(db),
	}

	// Test successful retrieval
	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"project_id": project1ID,
		})

		result, err := server.handleGetProject(ctx, request)
		if err != nil {
			t.Fatalf("handleGetProject returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test access denied
	t.Run("AccessDenied", func(t *testing.T) {
		token2, _ := createTestToken(t, db, userID, `["test-project-2"]`)
		ctx := context.WithValue(context.Background(), "mcp_token", token2)

		request := createMockRequest(map[string]interface{}{
			"project_id": project1ID, // Token only has access to project2
		})

		result, err := server.handleGetProject(ctx, request)
		if err != nil {
			t.Fatalf("handleGetProject returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for access denied")
		}
	})

	// Test project not found
	t.Run("NotFound", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"project_id": "non-existent-project",
		})

		result, err := server.handleGetProject(ctx, request)
		if err != nil {
			t.Fatalf("handleGetProject returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for not found")
		}
	})
}

// TestHandleGetRecentLogs tests the handleGetRecentLogs tool
func TestHandleGetRecentLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID, project1ID, project2ID, _ := setupTestData(t, db)
	token, _ := createTestToken(t, db, userID, "*")

	server := &MCPServer{
		mcpTokenRepo:    models.NewMCPTokenRepository(db),
		mcpActivityRepo: models.NewMCPActivityLogRepository(db),
		logRepo:         models.NewLogRepository(db),
		projectRepo:     models.NewProjectRepository(db),
		userRepo:        models.NewUserRepository(db),
	}

	// Test with no filters (all projects)
	t.Run("AllProjects", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{})

		result, err := server.handleGetRecentLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleGetRecentLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test with project filter
	t.Run("SpecificProject", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"project_ids": []interface{}{project1ID},
			"limit":       float64(10),
		})

		result, err := server.handleGetRecentLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleGetRecentLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test limit enforcement
	t.Run("LimitEnforcement", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"limit": float64(1000), // Above max of 500
		})

		result, err := server.handleGetRecentLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleGetRecentLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test access denied
	t.Run("AccessDenied", func(t *testing.T) {
		token2, _ := createTestToken(t, db, userID, `["test-project-1"]`)
		ctx := context.WithValue(context.Background(), "mcp_token", token2)

		request := createMockRequest(map[string]interface{}{
			"project_ids": []interface{}{project2ID}, // Token doesn't have access
		})

		result, err := server.handleGetRecentLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleGetRecentLogs returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for access denied")
		}
	})
}

// TestHandleQueryLogs tests the handleQueryLogs tool
func TestHandleQueryLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID, project1ID, _, _ := setupTestData(t, db)
	token, _ := createTestToken(t, db, userID, "*")

	server := &MCPServer{
		mcpTokenRepo:    models.NewMCPTokenRepository(db),
		mcpActivityRepo: models.NewMCPActivityLogRepository(db),
		logRepo:         models.NewLogRepository(db),
		projectRepo:     models.NewProjectRepository(db),
		userRepo:        models.NewUserRepository(db),
	}

	// Test basic query
	t.Run("BasicQuery", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"limit": float64(10),
		})

		result, err := server.handleQueryLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleQueryLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test with filters
	t.Run("WithFilters", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"project_ids": []interface{}{project1ID},
			"levels":      []interface{}{"info", "error"},
			"source":      "test-source",
			"limit":       float64(10),
			"offset":      float64(0),
		})

		result, err := server.handleQueryLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleQueryLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test with search
	t.Run("WithSearch", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"search": "error",
			"limit":  float64(10),
		})

		result, err := server.handleQueryLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleQueryLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test limit enforcement
	t.Run("LimitEnforcement", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"limit": float64(2000), // Above max of 1000
		})

		result, err := server.handleQueryLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleQueryLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})
}

// TestHandleSearchLogs tests the handleSearchLogs tool
func TestHandleSearchLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID, project1ID, _, _ := setupTestData(t, db)
	token, _ := createTestToken(t, db, userID, "*")

	server := &MCPServer{
		mcpTokenRepo:    models.NewMCPTokenRepository(db),
		mcpActivityRepo: models.NewMCPActivityLogRepository(db),
		logRepo:         models.NewLogRepository(db),
		projectRepo:     models.NewProjectRepository(db),
		userRepo:        models.NewUserRepository(db),
	}

	// Test successful search
	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"query": "test",
		})

		result, err := server.handleSearchLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleSearchLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test with filters
	t.Run("WithFilters", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"query":       "error",
			"project_ids": []interface{}{project1ID},
			"levels":      []interface{}{"error"},
			"limit":       float64(10),
		})

		result, err := server.handleSearchLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleSearchLogs returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test missing query parameter
	t.Run("MissingQuery", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{})

		result, err := server.handleSearchLogs(ctx, request)
		if err != nil {
			t.Fatalf("handleSearchLogs returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for missing query")
		}
	})
}

// TestHandleGetStats tests the handleGetStats tool
func TestHandleGetStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID, project1ID, _, _ := setupTestData(t, db)
	token, _ := createTestToken(t, db, userID, "*")

	server := &MCPServer{
		mcpTokenRepo:    models.NewMCPTokenRepository(db),
		mcpActivityRepo: models.NewMCPActivityLogRepository(db),
		logRepo:         models.NewLogRepository(db),
		projectRepo:     models.NewProjectRepository(db),
		userRepo:        models.NewUserRepository(db),
	}

	// Test overview stats
	t.Run("Overview", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"scope": "overview",
		})

		result, err := server.handleGetStats(ctx, request)
		if err != nil {
			t.Fatalf("handleGetStats returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test project stats
	t.Run("ProjectStats", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"scope":      "project",
			"project_id": project1ID,
		})

		result, err := server.handleGetStats(ctx, request)
		if err != nil {
			t.Fatalf("handleGetStats returned error: %v", err)
		}

		if result.IsError {
			t.Errorf("Expected success, got error result")
		}
	})

	// Test invalid scope
	t.Run("InvalidScope", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"scope": "invalid",
		})

		result, err := server.handleGetStats(ctx, request)
		if err != nil {
			t.Fatalf("handleGetStats returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for invalid scope")
		}
	})

	// Test project scope without project_id
	t.Run("MissingProjectID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "mcp_token", token)
		request := createMockRequest(map[string]interface{}{
			"scope": "project",
		})

		result, err := server.handleGetStats(ctx, request)
		if err != nil {
			t.Fatalf("handleGetStats returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for missing project_id")
		}
	})

	// Test access denied for project stats
	t.Run("AccessDenied", func(t *testing.T) {
		token2, _ := createTestToken(t, db, userID, `["test-project-2"]`)
		ctx := context.WithValue(context.Background(), "mcp_token", token2)

		request := createMockRequest(map[string]interface{}{
			"scope":      "project",
			"project_id": project1ID, // Token doesn't have access
		})

		result, err := server.handleGetStats(ctx, request)
		if err != nil {
			t.Fatalf("handleGetStats returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for access denied")
		}
	})
}
