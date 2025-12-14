package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"central-logs/internal/handlers"
	"central-logs/internal/middleware"
	"central-logs/internal/models"
	"central-logs/internal/utils"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
)

func setupLogTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'USER',
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			api_key TEXT UNIQUE NOT NULL,
			api_key_prefix TEXT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			retention_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create projects table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_projects (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'MEMBER',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (project_id) REFERENCES projects(id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create user_projects table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata TEXT,
			source TEXT,
			timestamp DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create logs table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			config TEXT NOT NULL,
			min_level TEXT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create channels table: %v", err)
	}

	return db
}

func TestLogHandler_CreateLog_Success(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	// Create a test project
	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", logHandler.CreateLog)

	reqBody := map[string]interface{}{
		"level":   "ERROR",
		"message": "Test error message",
		"source":  "test-app",
		"metadata": map[string]interface{}{
			"user_id": "123",
			"action":  "login",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response handlers.CreateLogResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	if response.ID == "" {
		t.Error("Expected log ID to be returned")
	}

	if response.Status != "received" {
		t.Errorf("Expected status 'received', got %s", response.Status)
	}

	// Verify log was created
	createdLog, _ := logRepo.GetByID(response.ID)
	if createdLog == nil {
		t.Fatal("Log should be created in database")
	}

	if createdLog.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got %s", createdLog.Message)
	}

	if createdLog.Level != models.LogLevelError {
		t.Errorf("Expected level ERROR, got %s", createdLog.Level)
	}
}

func TestLogHandler_CreateLog_InvalidAPIKey(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", logHandler.CreateLog)

	reqBody := map[string]string{
		"level":   "ERROR",
		"message": "Test error message",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", "invalid-key")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestLogHandler_CreateLog_InvalidBody(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", logHandler.CreateLog)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestLogHandler_CreateLog_MissingMessage(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", logHandler.CreateLog)

	reqBody := map[string]string{
		"level": "ERROR",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestLogHandler_CreateLog_WithTimestamp(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", logHandler.CreateLog)

	customTimestamp := time.Now().Add(-1 * time.Hour)
	reqBody := map[string]interface{}{
		"level":     "INFO",
		"message":   "Test message",
		"timestamp": customTimestamp.Format(time.RFC3339),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response handlers.CreateLogResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	createdLog, _ := logRepo.GetByID(response.ID)
	if createdLog.Timestamp.Unix() != customTimestamp.Unix() {
		t.Errorf("Expected timestamp %v, got %v", customTimestamp, createdLog.Timestamp)
	}
}

func TestLogHandler_CreateBatchLogs_Success(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs/batch", logHandler.CreateBatchLogs)

	reqBody := map[string]interface{}{
		"logs": []map[string]interface{}{
			{
				"level":   "ERROR",
				"message": "Error 1",
			},
			{
				"level":   "WARN",
				"message": "Warning 1",
			},
			{
				"level":   "INFO",
				"message": "Info 1",
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs/batch", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response handlers.BatchLogResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	if response.Received != 3 {
		t.Errorf("Expected 3 logs received, got %d", response.Received)
	}

	if len(response.IDs) != 3 {
		t.Errorf("Expected 3 IDs, got %d", len(response.IDs))
	}
}

func TestLogHandler_CreateBatchLogs_EmptyArray(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs/batch", logHandler.CreateBatchLogs)

	reqBody := map[string]interface{}{
		"logs": []map[string]interface{}{},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs/batch", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestLogHandler_CreateBatchLogs_TooManyLogs(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs/batch", logHandler.CreateBatchLogs)

	// Create 101 logs (exceeds the 100 limit)
	logs := make([]map[string]interface{}, 101)
	for i := 0; i < 101; i++ {
		logs[i] = map[string]interface{}{
			"level":   "INFO",
			"message": "Test message",
		}
	}

	reqBody := map[string]interface{}{
		"logs": logs,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs/batch", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestLogHandler_CreateBatchLogs_SkipsInvalidEntries(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	apiKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs/batch", logHandler.CreateBatchLogs)

	reqBody := map[string]interface{}{
		"logs": []map[string]interface{}{
			{
				"level":   "ERROR",
				"message": "Valid log",
			},
			{
				"level": "INFO",
				// Missing message - should be skipped
			},
			{
				"level":   "WARN",
				"message": "Another valid log",
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/logs/batch", bytes.NewReader(bodyBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response handlers.BatchLogResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	// Should only receive 2 valid logs
	if response.Received != 2 {
		t.Errorf("Expected 2 logs received, got %d", response.Received)
	}
}

func TestLogHandler_ListLogs_Admin(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	admin := &models.User{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin User",
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	userRepo.Create(admin)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	projectRepo.Create(project)

	// Create test logs
	for i := 0; i < 5; i++ {
		log := &models.Log{
			ProjectID: project.ID,
			Level:     models.LogLevelInfo,
			Message:   "Test message",
			Timestamp: time.Now(),
		}
		logRepo.Create(log)
	}

	token, _ := jwtManager.Generate(admin.ID, admin.Email, string(admin.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/logs", logHandler.ListLogs)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	logs := response["logs"].([]interface{})
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(logs))
	}
}

func TestLogHandler_ListLogs_RegularUser(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	user := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Regular User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	// Create two projects
	project1 := &models.Project{
		Name:     "Project 1",
		IsActive: true,
	}
	projectRepo.Create(project1)

	project2 := &models.Project{
		Name:     "Project 2",
		IsActive: true,
	}
	projectRepo.Create(project2)

	// Assign only project1 to user
	userProject := &models.UserProject{
		UserID:    user.ID,
		ProjectID: project1.ID,
		Role:      models.ProjectRoleMember,
	}
	userProjectRepo.Create(userProject)

	// Create logs for both projects
	for i := 0; i < 3; i++ {
		log := &models.Log{
			ProjectID: project1.ID,
			Level:     models.LogLevelInfo,
			Message:   "Project 1 log",
			Timestamp: time.Now(),
		}
		logRepo.Create(log)
	}

	for i := 0; i < 2; i++ {
		log := &models.Log{
			ProjectID: project2.ID,
			Level:     models.LogLevelInfo,
			Message:   "Project 2 log",
			Timestamp: time.Now(),
		}
		logRepo.Create(log)
	}

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/logs", logHandler.ListLogs)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	logs := response["logs"].([]interface{})
	// User should only see logs from project1
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

func TestLogHandler_ListLogs_WithFilters(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	admin := &models.User{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin User",
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	userRepo.Create(admin)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	projectRepo.Create(project)

	// Create logs with different levels
	levels := []models.LogLevel{
		models.LogLevelError,
		models.LogLevelError,
		models.LogLevelWarn,
		models.LogLevelInfo,
		models.LogLevelDebug,
	}

	for _, level := range levels {
		log := &models.Log{
			ProjectID: project.ID,
			Level:     level,
			Message:   "Test message",
			Timestamp: time.Now(),
		}
		logRepo.Create(log)
	}

	token, _ := jwtManager.Generate(admin.ID, admin.Email, string(admin.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/logs", logHandler.ListLogs)

	// Filter by ERROR level
	req := httptest.NewRequest(http.MethodGet, "/logs?levels=ERROR", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	logs := response["logs"].([]interface{})
	// Should only return ERROR logs
	if len(logs) != 2 {
		t.Errorf("Expected 2 ERROR logs, got %d", len(logs))
	}
}

func TestLogHandler_GetLog_Success(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	user := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	projectRepo.Create(project)

	userProject := &models.UserProject{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      models.ProjectRoleMember,
	}
	userProjectRepo.Create(userProject)

	log := &models.Log{
		ProjectID: project.ID,
		Level:     models.LogLevelError,
		Message:   "Test error",
		Timestamp: time.Now(),
	}
	logRepo.Create(log)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/logs/:id", logHandler.GetLog)

	req := httptest.NewRequest(http.MethodGet, "/logs/"+log.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var returnedLog models.Log
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &returnedLog)

	if returnedLog.ID != log.ID {
		t.Errorf("Expected log ID %s, got %s", log.ID, returnedLog.ID)
	}

	if returnedLog.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got %s", returnedLog.Message)
	}
}

func TestLogHandler_GetLog_NotFound(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	user := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/logs/:id", logHandler.GetLog)

	req := httptest.NewRequest(http.MethodGet, "/logs/nonexistent-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestLogHandler_GetLog_AccessDenied(t *testing.T) {
	db := setupLogTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil, nil, nil)

	user := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	projectRepo.Create(project)
	// Note: user is NOT assigned to this project

	log := &models.Log{
		ProjectID: project.ID,
		Level:     models.LogLevelError,
		Message:   "Test error",
		Timestamp: time.Now(),
	}
	logRepo.Create(log)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/logs/:id", logHandler.GetLog)

	req := httptest.NewRequest(http.MethodGet, "/logs/"+log.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}
