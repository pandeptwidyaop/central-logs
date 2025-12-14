package middleware_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestAuthMiddleware_RequireAuth_ValidToken(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	// Create a test user
	user := &models.User{
		Username: "testuser",
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

	// Generate token
	token, err := jwtManager.Generate(user.ID, user.Email, string(user.Role))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create Fiber app
	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/protected", func(c *fiber.Ctx) error {
		user := middleware.GetUser(c)
		return c.JSON(fiber.Map{"user_id": user.ID})
	})

	// Make request with valid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_RequireAuth_NoToken(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_RequireAuth_ExpiredToken(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	// Create JWT manager with 0 second expiry (immediate expiry)
	jwtManager := utils.NewJWTManager("test-secret", -1) // Already expired
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	// Create a test user
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	// Generate expired token
	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for expired token, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_RequireAuth_InactiveUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	// Create an inactive user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: false, // Inactive
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for inactive user, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_RequireAdmin_AdminUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	// Create an admin user
	user := &models.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin User",
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Use(authMiddleware.RequireAdmin())
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendString("Admin OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for admin user, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_RequireAdmin_NonAdminUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	// Create a regular user
	user := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Regular User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Use(authMiddleware.RequireAdmin())
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendString("Admin OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-admin user, got %d", resp.StatusCode)
	}
}

func TestGetUser_ReturnsUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userRepo := models.NewUserRepository(db)
	jwtManager := utils.NewJWTManager("test-secret", 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	userRepo.Create(user)

	token, _ := jwtManager.Generate(user.ID, user.Email, string(user.Role))

	var gotUser *models.User

	app := fiber.New()
	app.Use(authMiddleware.RequireAuth())
	app.Get("/test", func(c *fiber.Ctx) error {
		gotUser = middleware.GetUser(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if gotUser == nil {
		t.Fatal("GetUser should return user")
	}

	if gotUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, gotUser.ID)
	}

	if gotUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, gotUser.Email)
	}
}
