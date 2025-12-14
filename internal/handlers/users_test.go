package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"central-logs/internal/handlers"
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
)

func setupUserTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

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

func TestUserHandler_ListUsers_Success(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	// Create test users
	for i := 0; i < 3; i++ {
		user := &models.User{
			Username: "user" + string(rune('0'+i)),
			Email:    "user" + string(rune('0'+i)) + "@example.com",
			Password: "password123",
			Name:     "User " + string(rune('0'+i)),
			Role:     models.RoleUser,
			IsActive: true,
		}
		userRepo.Create(user)
	}

	app := fiber.New()
	app.Get("/users", userHandler.ListUsers)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
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

	users := response["users"].([]interface{})
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	app := fiber.New()
	app.Post("/users", userHandler.CreateUser)

	reqBody := map[string]interface{}{
		"username": "newuser",
		"password": "password123",
		"name":     "New User",
		"role":     "USER",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var createdUser models.User
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &createdUser)

	if createdUser.Username != "newuser" {
		t.Errorf("Expected username newuser, got %s", createdUser.Username)
	}

	if createdUser.Name != "New User" {
		t.Errorf("Expected name 'New User', got %s", createdUser.Name)
	}

	if createdUser.Role != models.RoleUser {
		t.Errorf("Expected role USER, got %s", createdUser.Role)
	}
}

func TestUserHandler_CreateUser_InvalidBody(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	app := fiber.New()
	app.Post("/users", userHandler.CreateUser)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandler_CreateUser_MissingFields(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	app := fiber.New()
	app.Post("/users", userHandler.CreateUser)

	tests := []struct {
		name     string
		username string
		password string
		userName string
	}{
		{"no username", "", "password123", "Test User"},
		{"no password", "testuser", "", "Test User"},
		{"no name", "testuser", "password123", ""},
		{"all empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]string{
				"username": tt.username,
				"password": tt.password,
				"name":     tt.userName,
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestUserHandler_CreateUser_PasswordTooShort(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	app := fiber.New()
	app.Post("/users", userHandler.CreateUser)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "short",
		"name":     "Test User",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandler_CreateUser_EmailAlreadyExists(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	// Create existing user
	existingUser := &models.User{
		Username: "existing",
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Existing User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(existingUser)

	app := fiber.New()
	app.Post("/users", userHandler.CreateUser)

	reqBody := map[string]string{
		"username": "existing",
		"password": "password123",
		"name":     "New User",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", resp.StatusCode)
	}
}

func TestUserHandler_CreateUser_DefaultRole(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	app := fiber.New()
	app.Post("/users", userHandler.CreateUser)

	reqBody := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
		"name":     "Test User",
		"role":     "INVALID_ROLE",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var createdUser models.User
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &createdUser)

	if createdUser.Role != models.RoleUser {
		t.Errorf("Expected default role USER, got %s", createdUser.Role)
	}
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Get("/users/:id", userHandler.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var returnedUser models.User
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &returnedUser)

	if returnedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, returnedUser.ID)
	}
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	app := fiber.New()
	app.Get("/users/:id", userHandler.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/users/nonexistent-id", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestUserHandler_UpdateUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Put("/users/:id", userHandler.UpdateUser)

	isActive := false
	reqBody := map[string]interface{}{
		"name":      "Updated Name",
		"role":      "ADMIN",
		"is_active": &isActive,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updatedUser models.User
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &updatedUser)

	if updatedUser.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", updatedUser.Name)
	}

	if updatedUser.Role != models.RoleAdmin {
		t.Errorf("Expected role ADMIN, got %s", updatedUser.Role)
	}

	if updatedUser.IsActive {
		t.Error("Expected is_active to be false")
	}
}

func TestUserHandler_UpdateUser_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	app := fiber.New()
	app.Put("/users/:id", userHandler.UpdateUser)

	reqBody := map[string]string{
		"name": "Updated Name",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/nonexistent-id", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestUserHandler_UpdateUser_InvalidBody(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Put("/users/:id", userHandler.UpdateUser)

	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID, bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandler_UpdateUser_RoleUpdate(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "user1",
		Email:    "user1@example.com",
		Password: "password123",
		Name:     "User 1",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Put("/users/:id", userHandler.UpdateUser)

	reqBody := map[string]string{
		"role": "ADMIN",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updatedUser models.User
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &updatedUser)

	if updatedUser.Role != models.RoleAdmin {
		t.Errorf("Expected role ADMIN, got %s", updatedUser.Role)
	}
}

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Delete("/users/:id", userHandler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+user.ID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	deletedUser, _ := userRepo.GetByID(user.ID)
	if deletedUser != nil {
		t.Error("User should be deleted")
	}
}

func TestUserHandler_ResetPassword_Success(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Put("/users/:id/reset-password", userHandler.ResetPassword)

	reqBody := map[string]string{
		"password": "newpassword123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID+"/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify new password works
	updatedUser, _ := userRepo.GetByID(user.ID)
	if !updatedUser.CheckPassword("newpassword123") {
		t.Error("New password should work")
	}

	if updatedUser.CheckPassword("password123") {
		t.Error("Old password should not work")
	}
}

func TestUserHandler_ResetPassword_InvalidBody(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Put("/users/:id/reset-password", userHandler.ResetPassword)

	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID+"/reset-password", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandler_ResetPassword_PasswordTooShort(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Put("/users/:id/reset-password", userHandler.ResetPassword)

	reqBody := map[string]string{
		"password": "short",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID+"/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}
