package models_test

import (
	"database/sql"
	"testing"

	"central-logs/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func setupProjectTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create projects table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			api_key TEXT NOT NULL,
			api_key_prefix TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			retention_config TEXT,
			icon_type TEXT DEFAULT 'initials',
			icon_value TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create projects table: %v", err)
	}

	return db
}

func TestProjectRepository_Create(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}

	apiKey, err := repo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.ID == "" {
		t.Error("Project ID should be set after creation")
	}

	if apiKey == "" {
		t.Error("API key should be returned")
	}

	if len(apiKey) < 32 {
		t.Error("API key should be at least 32 characters")
	}
}

func TestProjectRepository_GetByID(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	// Create a project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	_, err := repo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Get by ID
	found, err := repo.GetByID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project by ID: %v", err)
	}

	if found == nil {
		t.Fatal("Project should be found")
	}

	if found.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", found.Name)
	}
}

func TestProjectRepository_GetByAPIKey(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	// Create a project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	apiKey, err := repo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Get by API key
	found, err := repo.GetByAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to get project by API key: %v", err)
	}

	if found == nil {
		t.Fatal("Project should be found")
	}

	if found.ID != project.ID {
		t.Errorf("Expected project ID %s, got %s", project.ID, found.ID)
	}
}

func TestProjectRepository_GetByAPIKey_InactiveProject(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	// Create an inactive project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	apiKey, err := repo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Deactivate project
	project.IsActive = false
	err = repo.Update(project)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	// Get by API key should fail for inactive project
	found, err := repo.GetByAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if found != nil {
		t.Error("Inactive project should not be found by API key")
	}
}

func TestProjectRepository_Update(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	// Create a project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	_, err := repo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Update
	project.Name = "Updated Project"
	project.Description = "Updated Description"
	err = repo.Update(project)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	// Verify
	found, err := repo.GetByID(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if found.Name != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got '%s'", found.Name)
	}

	if found.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", found.Description)
	}
}

func TestProjectRepository_RotateAPIKey(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	// Create a project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	oldKey, err := repo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Rotate API key
	newKey, err := repo.RotateAPIKey(project.ID)
	if err != nil {
		t.Fatalf("Failed to rotate API key: %v", err)
	}

	if newKey == oldKey {
		t.Error("New API key should be different from old key")
	}

	if len(newKey) < 32 {
		t.Error("New API key should be at least 32 characters")
	}

	// Verify old key no longer works
	found, err := repo.GetByAPIKey(oldKey)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if found != nil {
		t.Error("Old API key should no longer work")
	}

	// Verify new key works
	found, err = repo.GetByAPIKey(newKey)
	if err != nil {
		t.Fatalf("Failed to get project by new API key: %v", err)
	}
	if found == nil {
		t.Error("New API key should work")
	}
}

func TestProjectRepository_Delete(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	// Create a project
	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		IsActive:    true,
	}
	_, err := repo.Create(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Delete
	err = repo.Delete(project.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify deletion
	found, err := repo.GetByID(project.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if found != nil {
		t.Error("Project should be deleted")
	}
}

func TestProjectRepository_GetAll(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	repo := models.NewProjectRepository(db)

	// Create multiple projects
	for i := 0; i < 3; i++ {
		project := &models.Project{
			Name:        "Test Project " + string(rune('A'+i)),
			Description: "Description " + string(rune('A'+i)),
			IsActive:    true,
		}
		_, err := repo.Create(project)
		if err != nil {
			t.Fatalf("Failed to create project %d: %v", i, err)
		}
	}

	// List
	projects, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}
}

func TestGenerateAPIKey(t *testing.T) {
	key1, hash1, prefix1, err := models.GenerateAPIKey()
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	key2, hash2, prefix2, err := models.GenerateAPIKey()
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if key1 == key2 {
		t.Error("Generated API keys should be unique")
	}

	if hash1 == hash2 {
		t.Error("Generated hashes should be unique")
	}

	if len(key1) < 32 {
		t.Error("API key should be at least 32 characters")
	}

	if prefix1 == "" || prefix2 == "" {
		t.Error("API key prefix should not be empty")
	}
}

func TestHashAPIKey(t *testing.T) {
	key := "cl_test_key_12345678901234567890"
	hash1 := models.HashAPIKey(key)
	hash2 := models.HashAPIKey(key)

	if hash1 != hash2 {
		t.Error("Same key should produce same hash")
	}

	differentKey := "cl_different_key_123456789012345"
	differentHash := models.HashAPIKey(differentKey)

	if hash1 == differentHash {
		t.Error("Different keys should produce different hashes")
	}
}
