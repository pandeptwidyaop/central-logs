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

func setupAuthTestDB(t *testing.T) *sql.DB {
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

	return db
}

func TestAuthHandler_Login_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	app := fiber.New()
	app.Post("/login", authHandler.Login)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response handlers.LoginResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	if response.Token == "" {
		t.Error("Expected token to be returned")
	}

	if response.User == nil {
		t.Error("Expected user to be returned")
	}

	if response.User.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", response.User.Email)
	}
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	app := fiber.New()
	app.Post("/login", authHandler.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	app := fiber.New()
	app.Post("/login", authHandler.Login)

	tests := []struct {
		name  string
		email string
		pass  string
	}{
		{"no email", "", "password123"},
		{"no password", "test@example.com", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]string{
				"email":    tt.email,
				"password": tt.pass,
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
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

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Post("/login", authHandler.Login)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	app := fiber.New()
	app.Post("/login", authHandler.Login)

	reqBody := map[string]string{
		"email":    "notfound@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_Login_InactiveUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	// Create an inactive user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: false,
	}
	userRepo.Create(user)

	app := fiber.New()
	app.Post("/login", authHandler.Login)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_Me_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	// Manually set user in context (simulating authenticated middleware)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", user)
		return c.Next()
	})
	app.Get("/me", authHandler.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)

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

	if returnedUser.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", returnedUser.Email)
	}
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	app := fiber.New()
	app.Get("/me", authHandler.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_UpdateProfile_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	// Manually set user in context (simulating authenticated middleware)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", user)
		return c.Next()
	})
	app.Put("/profile", authHandler.UpdateProfile)

	reqBody := map[string]string{
		"name":  "Updated Name",
		"email": "updated@example.com",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewReader(bodyBytes))
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

	if updatedUser.Email != "updated@example.com" {
		t.Errorf("Expected email 'updated@example.com', got %s", updatedUser.Email)
	}
}

func TestAuthHandler_UpdateProfile_InvalidBody(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	app := fiber.New()
	// Manually set user in context (simulating authenticated middleware)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", user)
		return c.Next()
	})
	app.Put("/profile", authHandler.UpdateProfile)

	req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_UpdateProfile_EmailAlreadyInUse(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	// Create first user
	user1 := &models.User{
		Email:    "user1@example.com",
		Password: "password123",
		Name:     "User 1",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user1)

	// Create second user
	user2 := &models.User{
		Email:    "user2@example.com",
		Password: "password123",
		Name:     "User 2",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user2)

	app := fiber.New()
	// Manually set user in context (simulating authenticated middleware)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", user1)
		return c.Next()
	})
	app.Put("/profile", authHandler.UpdateProfile)

	// Try to update user1's email to user2's email
	reqBody := map[string]string{
		"email": "user2@example.com",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Put("/change-password", authHandler.ChangePassword)

	reqBody := map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
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

func TestAuthHandler_ChangePassword_InvalidBody(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Put("/change-password", authHandler.ChangePassword)

	req := httptest.NewRequest(http.MethodPut, "/change-password", bytes.NewReader([]byte("invalid json")))
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

func TestAuthHandler_ChangePassword_MissingFields(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Put("/change-password", authHandler.ChangePassword)

	tests := []struct {
		name     string
		current  string
		newPass  string
		expected int
	}{
		{"no current password", "", "newpassword123", http.StatusBadRequest},
		{"no new password", "password123", "", http.StatusBadRequest},
		{"both empty", "", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]string{
				"current_password": tt.current,
				"new_password":     tt.newPass,
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPut, "/change-password", bytes.NewReader(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			if resp.StatusCode != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, resp.StatusCode)
			}
		})
	}
}

func TestAuthHandler_ChangePassword_PasswordTooShort(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Put("/change-password", authHandler.ChangePassword)

	reqBody := map[string]string{
		"current_password": "password123",
		"new_password":     "short",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/change-password", bytes.NewReader(bodyBytes))
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

func TestAuthHandler_ChangePassword_WrongCurrentPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Put("/change-password", authHandler.ChangePassword)

	reqBody := map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}
