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

func setupProjectTestDB(t *testing.T) *sql.DB {
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

	return db
}

func TestProjectHandler_ListProjects_Admin(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

	// Create admin user
	admin := &models.User{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin User",
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	userRepo.Create(admin)

	// Create projects
	for i := 0; i < 3; i++ {
		project := &models.Project{
			Name:     "Project " + string(rune('0'+i)),
			IsActive: true,
		}
		projectRepo.Create(project)
	}

	token, _ := jwtManager.Generate(admin.ID, admin.Email, string(admin.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/projects", projectHandler.ListProjects)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
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

	projects := response["projects"].([]interface{})
	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}
}

func TestProjectHandler_ListProjects_RegularUser(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

	// Create regular user
	user := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Regular User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	// Create projects and assign one to user
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

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/projects", projectHandler.ListProjects)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
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

	projects := response["projects"].([]interface{})
	// User should only see their assigned project
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
}

func TestProjectHandler_CreateProject_Success(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

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
	app.Post("/projects", projectHandler.CreateProject)

	reqBody := map[string]string{
		"name":        "New Project",
		"description": "Test project description",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response handlers.CreateProjectResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	if response.Project == nil {
		t.Fatal("Expected project to be returned")
	}

	if response.Project.Name != "New Project" {
		t.Errorf("Expected name 'New Project', got %s", response.Project.Name)
	}

	if response.APIKey == "" {
		t.Error("Expected API key to be returned")
	}

	// Verify user is assigned as owner
	userProject, _ := userProjectRepo.GetByUserAndProject(user.ID, response.Project.ID)
	if userProject == nil {
		t.Fatal("Expected user to be assigned to project")
	}

	if userProject.Role != models.ProjectRoleOwner {
		t.Errorf("Expected role OWNER, got %s", userProject.Role)
	}
}

func TestProjectHandler_CreateProject_InvalidBody(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

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
	app.Post("/projects", projectHandler.CreateProject)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestProjectHandler_CreateProject_MissingName(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

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
	app.Post("/projects", projectHandler.CreateProject)

	reqBody := map[string]string{
		"description": "Test description",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestProjectHandler_GetProject_Success(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

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

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/projects/:id", projectHandler.GetProject)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID, nil)
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

	projectData := response["project"].(map[string]interface{})
	if projectData["name"] != "Test Project" {
		t.Errorf("Expected name 'Test Project', got %s", projectData["name"])
	}
}

func TestProjectHandler_GetProject_NotFound(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

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
	app.Get("/projects/:id", projectHandler.GetProject)

	req := httptest.NewRequest(http.MethodGet, "/projects/nonexistent-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestProjectHandler_GetProject_AccessDenied(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

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

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/projects/:id", projectHandler.GetProject)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestProjectHandler_UpdateProject_Success(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

	project := &models.Project{
		Name:        "Test Project",
		Description: "Old description",
		IsActive:    true,
	}
	projectRepo.Create(project)

	app := fiber.New()
	app.Put("/projects/:id", projectHandler.UpdateProject)

	isActive := false
	reqBody := map[string]interface{}{
		"name":        "Updated Project",
		"description": "New description",
		"is_active":   &isActive,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updatedProject models.Project
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &updatedProject)

	if updatedProject.Name != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got %s", updatedProject.Name)
	}

	if updatedProject.Description != "New description" {
		t.Errorf("Expected description 'New description', got %s", updatedProject.Description)
	}

	if updatedProject.IsActive {
		t.Error("Expected is_active to be false")
	}
}

func TestProjectHandler_UpdateProject_NotFound(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

	app := fiber.New()
	app.Put("/projects/:id", projectHandler.UpdateProject)

	reqBody := map[string]string{
		"name": "Updated Project",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/projects/nonexistent-id", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestProjectHandler_DeleteProject_Success(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	projectRepo.Create(project)

	app := fiber.New()
	app.Delete("/projects/:id", projectHandler.DeleteProject)

	req := httptest.NewRequest(http.MethodDelete, "/projects/"+project.ID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	deletedProject, _ := projectRepo.GetByID(project.ID)
	if deletedProject != nil {
		t.Error("Project should be deleted")
	}
}

func TestProjectHandler_RotateAPIKey_Success(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

	project := &models.Project{
		Name:     "Test Project",
		IsActive: true,
	}
	oldAPIKey, _ := projectRepo.Create(project)

	app := fiber.New()
	app.Post("/projects/:id/rotate-key", projectHandler.RotateAPIKey)

	req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID+"/rotate-key", nil)
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

	newAPIKey := response["api_key"].(string)
	if newAPIKey == "" {
		t.Error("Expected new API key to be returned")
	}

	if newAPIKey == oldAPIKey {
		t.Error("New API key should be different from old API key")
	}

	// Verify old API key no longer works
	oldProject, _ := projectRepo.GetByAPIKey(oldAPIKey)
	if oldProject != nil {
		t.Error("Old API key should not work")
	}

	// Verify new API key works
	newProject, _ := projectRepo.GetByAPIKey(newAPIKey)
	if newProject == nil {
		t.Error("New API key should work")
	}
}

func TestProjectHandler_RotateAPIKey_NotFound(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	userProjectRepo := models.NewUserProjectRepository(db)
	logRepo := models.NewLogRepository(db)

	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)

	app := fiber.New()
	app.Post("/projects/:id/rotate-key", projectHandler.RotateAPIKey)

	req := httptest.NewRequest(http.MethodPost, "/projects/nonexistent-id/rotate-key", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}
