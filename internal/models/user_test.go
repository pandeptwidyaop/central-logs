package models_test

import (
	"database/sql"
	"testing"

	"central-logs/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'USER',
			is_active INTEGER NOT NULL DEFAULT 1,
			two_factor_secret TEXT DEFAULT '',
			two_factor_enabled INTEGER DEFAULT 0,
			backup_codes TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == "" {
		t.Error("User ID should be set after creation")
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	// Create a user first
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get by email
	found, err := repo.GetByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if found == nil {
		t.Fatal("User should be found")
	}

	if found.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", found.Email)
	}

	if found.Name != "Test User" {
		t.Errorf("Expected name Test User, got %s", found.Name)
	}
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	found, err := repo.GetByEmail("notfound@example.com")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if found != nil {
		t.Error("User should not be found")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	// Create user (password is hashed during creation)
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Retrieve user to get hashed password
	found, err := repo.GetByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if !found.CheckPassword("password123") {
		t.Error("CheckPassword should return true for correct password")
	}

	if found.CheckPassword("wrongpassword") {
		t.Error("CheckPassword should return false for incorrect password")
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	// Create a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update user
	user.Name = "Updated Name"
	user.Role = models.RoleAdmin
	err = repo.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	found, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if found.Name != "Updated Name" {
		t.Errorf("Expected name Updated Name, got %s", found.Name)
	}

	if found.Role != models.RoleAdmin {
		t.Errorf("Expected role ADMIN, got %s", found.Role)
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	// Create a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update password
	err = repo.UpdatePassword(user.ID, "newpassword456")
	if err != nil {
		t.Fatalf("Failed to update password: %v", err)
	}

	// Verify new password works
	found, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if !found.CheckPassword("newpassword456") {
		t.Error("New password should work")
	}

	if found.CheckPassword("password123") {
		t.Error("Old password should not work")
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	// Create a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete user
	err = repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify deletion
	found, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if found != nil {
		t.Error("User should be deleted")
	}
}

func TestUserRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	// Create multiple users
	for i := 0; i < 5; i++ {
		user := &models.User{
			Username: "testuser" + string(rune('0'+i)),
			Email:    "test" + string(rune('0'+i)) + "@example.com",
			Password: "password123",
			Name:     "Test User " + string(rune('0'+i)),
			Role:     models.RoleUser,
			IsActive: true,
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	// List users
	users, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
}

func TestUserRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.NewUserRepository(db)

	// Count empty
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 users, got %d", count)
	}

	// Create a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err = repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Count again
	count, err = repo.Count()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}
}

func TestUser_IsAdmin(t *testing.T) {
	adminUser := &models.User{Role: models.RoleAdmin}
	regularUser := &models.User{Role: models.RoleUser}

	if !adminUser.IsAdmin() {
		t.Error("Admin user should return true for IsAdmin")
	}

	if regularUser.IsAdmin() {
		t.Error("Regular user should return false for IsAdmin")
	}
}
