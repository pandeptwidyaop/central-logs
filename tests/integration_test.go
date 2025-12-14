package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

type TestApp struct {
	App             *fiber.App
	DB              *sql.DB
	UserRepo        *models.UserRepository
	ProjectRepo     *models.ProjectRepository
	UserProjectRepo *models.UserProjectRepository
	LogRepo         *models.LogRepository
	ChannelRepo     *models.ChannelRepository
	JWTManager      *utils.JWTManager
	AdminToken      string
	UserToken       string
	AdminUser       *models.User
	RegularUser     *models.User
}

func setupTestApp(t *testing.T) *TestApp {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create all tables
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
		);

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
		);

		CREATE TABLE IF NOT EXISTS user_projects (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'MEMBER',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (project_id) REFERENCES projects(id)
		);

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
		);

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
		);

		CREATE INDEX IF NOT EXISTS idx_logs_project_id ON logs(project_id);
		CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
		CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Initialize repositories
	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	channelRepo := models.NewChannelRepository(db)

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)

	// Create test users
	adminUser := &models.User{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin User",
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	userRepo.Create(adminUser)

	regularUser := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Regular User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(regularUser)

	// Generate tokens
	adminToken, _ := jwtManager.Generate(adminUser.ID, adminUser.Email, string(adminUser.Role))
	userToken, _ := jwtManager.Generate(regularUser.ID, regularUser.Email, string(regularUser.Role))

	// Initialize middlewares
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)
	rbacMiddleware := middleware.NewRBACMiddleware(userProjectRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	userHandler := handlers.NewUserHandler(userRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)
	memberHandler := handlers.NewMemberHandler(userRepo, userProjectRepo)
	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, nil)
	statsHandler := handlers.NewStatsHandler(logRepo, projectRepo, userProjectRepo, userRepo)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Setup routes
	api := app.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	authProtected := auth.Group("", authMiddleware.RequireAuth())
	authProtected.Get("/me", authHandler.Me)
	authProtected.Put("/change-password", authHandler.ChangePassword)

	// Public log ingestion
	v1 := api.Group("/v1")
	logIngestion := v1.Group("/logs", apiKeyMiddleware.RequireAPIKey())
	logIngestion.Post("", logHandler.CreateLog)
	logIngestion.Post("/batch", logHandler.CreateBatchLogs)

	// Admin API
	admin := api.Group("/admin", authMiddleware.RequireAuth())

	// Users (admin only)
	users := admin.Group("/users", authMiddleware.RequireAdmin())
	users.Get("", userHandler.ListUsers)
	users.Post("", userHandler.CreateUser)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// Projects
	projects := admin.Group("/projects")
	projects.Get("", projectHandler.ListProjects)
	projects.Post("", projectHandler.CreateProject)
	projects.Get("/:id", rbacMiddleware.RequireProjectAccess(), projectHandler.GetProject)
	projects.Put("/:id", rbacMiddleware.RequireOwner(), projectHandler.UpdateProject)
	projects.Delete("/:id", rbacMiddleware.RequireOwner(), projectHandler.DeleteProject)
	projects.Post("/:id/rotate-key", rbacMiddleware.RequireOwner(), projectHandler.RotateAPIKey)

	// Members
	projects.Get("/:id/members", rbacMiddleware.RequireProjectAccess(), memberHandler.ListMembers)
	projects.Post("/:id/members", rbacMiddleware.RequireOwner(), memberHandler.AddMember)
	projects.Put("/:id/members/:uid", rbacMiddleware.RequireOwner(), memberHandler.UpdateMember)
	projects.Delete("/:id/members/:uid", rbacMiddleware.RequireOwner(), memberHandler.RemoveMember)

	// Logs
	logs := admin.Group("/logs")
	logs.Get("", logHandler.ListLogs)
	logs.Get("/:id", logHandler.GetLog)

	// Stats
	stats := admin.Group("/stats")
	stats.Get("/overview", statsHandler.GetOverview)

	return &TestApp{
		App:             app,
		DB:              db,
		UserRepo:        userRepo,
		ProjectRepo:     projectRepo,
		UserProjectRepo: userProjectRepo,
		LogRepo:         logRepo,
		ChannelRepo:     channelRepo,
		JWTManager:      jwtManager,
		AdminToken:      adminToken,
		UserToken:       userToken,
		AdminUser:       adminUser,
		RegularUser:     regularUser,
	}
}

func (ta *TestApp) Close() {
	ta.DB.Close()
}

// ==================== Auth Tests ====================

func TestLogin_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	body := map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["token"] == nil {
		t.Error("Expected token in response")
	}
	if result["user"] == nil {
		t.Error("Expected user in response")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	body := map[string]string{
		"email":    "admin@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestGetProfile_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var user map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&user)

	if user["email"] != "admin@example.com" {
		t.Errorf("Expected email admin@example.com, got %v", user["email"])
	}
}

// ==================== User Management Tests ====================

func TestListUsers_AdminOnly(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	// Admin can list users
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for admin, got %d", resp.StatusCode)
	}

	// Regular user cannot list users
	req = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+ta.UserToken)

	resp, err = ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403 for regular user, got %d", resp.StatusCode)
	}
}

func TestCreateUser_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	body := map[string]string{
		"email":    "newuser@example.com",
		"password": "password123",
		"name":     "New User",
		"role":     "USER",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

// ==================== Project Tests ====================

func TestCreateProject_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	body := map[string]string{
		"name":        "Test Project",
		"description": "Test Description",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/projects", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	project, ok := result["project"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected project object in response")
	}

	if project["name"] != "Test Project" {
		t.Errorf("Expected name 'Test Project', got %v", project["name"])
	}

	if result["api_key"] == nil {
		t.Error("Expected api_key in response")
	}
}

func TestListProjects_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	// Create a project first
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test",
	}
	ta.ProjectRepo.Create(project)

	// Assign to admin user
	ta.UserProjectRepo.Create(&models.UserProject{
		UserID:    ta.AdminUser.ID,
		ProjectID: project.ID,
		Role:      models.ProjectRoleOwner,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/projects", nil)
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ==================== Log Ingestion Tests ====================

func TestCreateLog_WithAPIKey(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	// Create a project
	project := &models.Project{
		Name:     "Log Test Project",
		IsActive: true,
	}
	apiKey, _ := ta.ProjectRepo.Create(project)

	body := map[string]interface{}{
		"level":   "INFO",
		"message": "Test log message",
		"metadata": map[string]interface{}{
			"user_id": 123,
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestCreateLog_InvalidAPIKey(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	body := map[string]string{
		"level":   "INFO",
		"message": "Test log message",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "invalid-key")

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestCreateBatchLogs_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	// Create a project
	project := &models.Project{
		Name:     "Batch Log Test",
		IsActive: true,
	}
	apiKey, _ := ta.ProjectRepo.Create(project)

	body := map[string]interface{}{
		"logs": []map[string]string{
			{"level": "INFO", "message": "Log 1"},
			{"level": "WARN", "message": "Log 2"},
			{"level": "ERROR", "message": "Log 3"},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs/batch", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["received"] != float64(3) {
		t.Errorf("Expected received 3, got %v", result["received"])
	}
}

// ==================== Stats Tests ====================

func TestGetStats_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	// Create a project and some logs
	project := &models.Project{Name: "Stats Test"}
	ta.ProjectRepo.Create(project)
	ta.UserProjectRepo.Create(&models.UserProject{
		UserID:    ta.AdminUser.ID,
		ProjectID: project.ID,
		Role:      models.ProjectRoleOwner,
	})

	for i := 0; i < 5; i++ {
		ta.LogRepo.Create(&models.Log{
			ProjectID: project.ID,
			Level:     models.LogLevelInfo,
			Message:   "Test log",
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats/overview", nil)
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&stats)

	if stats["total_logs"] == nil {
		t.Error("Expected total_logs in stats")
	}
}

// ==================== Authorization Tests ====================

func TestUnauthorizedAccess(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/users"},
		{"GET", "/api/admin/projects"},
		{"GET", "/api/admin/logs"},
		{"GET", "/api/admin/stats/overview"},
		{"GET", "/api/auth/me"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)

		resp, err := ta.App.Test(req)
		if err != nil {
			t.Fatalf("Failed to make request to %s: %v", ep.path, err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for %s %s, got %d", ep.method, ep.path, resp.StatusCode)
		}
	}
}

func TestChangePassword_Success(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	body := map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/auth/change-password", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify new password works
	loginBody := map[string]string{
		"email":    "admin@example.com",
		"password": "newpassword123",
	}
	loginJson, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginJson))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, _ := ta.App.Test(loginReq)
	if loginResp.StatusCode != http.StatusOK {
		t.Error("Should be able to login with new password")
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	ta := setupTestApp(t)
	defer ta.Close()

	body := map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/auth/change-password", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ta.AdminToken)

	resp, err := ta.App.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}
