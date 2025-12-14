package middleware_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"central-logs/internal/middleware"
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
)

func setupAPIKeyTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
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

	return db
}

func TestAPIKeyMiddleware_ValidKey(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	// Create a project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	apiKey, err := projectRepo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", func(c *fiber.Ctx) error {
		proj := middleware.GetProject(c)
		return c.JSON(fiber.Map{"project_id": proj.ID})
	})

	req := httptest.NewRequest(http.MethodPost, "/logs", nil)
	req.Header.Set("X-API-Key", apiKey)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAPIKeyMiddleware_NoKey(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/logs", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/logs", nil)
	req.Header.Set("X-API-Key", "invalid-api-key")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestGetProject_ReturnsProject(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	defer db.Close()

	projectRepo := models.NewProjectRepository(db)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)

	// Create a project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	apiKey, err := projectRepo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	var gotProject *models.Project

	app := fiber.New()
	app.Use(apiKeyMiddleware.RequireAPIKey())
	app.Post("/logs", func(c *fiber.Ctx) error {
		gotProject = middleware.GetProject(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/logs", nil)
	req.Header.Set("X-API-Key", apiKey)

	_, err = app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if gotProject == nil {
		t.Fatal("GetProject should return project")
	}

	if gotProject.ID != project.ID {
		t.Errorf("Expected project ID %s, got %s", project.ID, gotProject.ID)
	}

	if gotProject.Name != project.Name {
		t.Errorf("Expected project name %s, got %s", project.Name, gotProject.Name)
	}
}
