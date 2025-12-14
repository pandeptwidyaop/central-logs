package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"central-logs/internal/config"
	"central-logs/internal/database"
	"central-logs/internal/database/migrations"
	"central-logs/internal/handlers"
	"central-logs/internal/middleware"
	"central-logs/internal/models"
	"central-logs/internal/queue"
	"central-logs/internal/services/notification"
	"central-logs/internal/utils"
	"central-logs/internal/websocket"
	"central-logs/web"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Build-time variables (injected via ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations with Laravel-style tracking
	if err := db.MigrateWithRegistry(migrations.GetAll()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	redisClient, err := queue.NewRedisClient(cfg.Redis.URL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Printf("Realtime features and rate limiting will be disabled")
		redisClient = nil
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Initialize repositories
	userRepo := models.NewUserRepository(db.DB)
	projectRepo := models.NewProjectRepository(db.DB)
	userProjectRepo := models.NewUserProjectRepository(db.DB)
	logRepo := models.NewLogRepository(db.DB)
	channelRepo := models.NewChannelRepository(db.DB)
	subscriptionRepo := models.NewPushSubscriptionRepository(db.DB)

	// Create initial admin user if no users exist
	if err := createInitialAdmin(userRepo, cfg); err != nil {
		log.Printf("Warning: Failed to create initial admin: %v", err)
	}

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(cfg.JWT.Secret, cfg.GetJWTExpiry())

	// Initialize middlewares
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(projectRepo)
	rbacMiddleware := middleware.NewRBACMiddleware(userProjectRepo)

	var rateLimitMiddleware *middleware.RateLimitMiddleware
	if redisClient != nil {
		rateLimiter := queue.NewRateLimiter(redisClient.Client())
		rateLimitMiddleware = middleware.NewRateLimitMiddleware(rateLimiter, cfg.RateLimit.API.RequestsPerMinute)
	}

	// Initialize services
	pushService := notification.NewPushService(subscriptionRepo, channelRepo, cfg)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	wsHandler := websocket.NewHandler(wsHub)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	twoFactorHandler := handlers.NewTwoFactorHandler(userRepo, jwtManager, "Central Logs")
	userHandler := handlers.NewUserHandler(userRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo, userProjectRepo, logRepo)
	memberHandler := handlers.NewMemberHandler(userRepo, userProjectRepo)
	logHandler := handlers.NewLogHandler(logRepo, channelRepo, userProjectRepo, redisClient, pushService, wsHub)
	channelHandler := handlers.NewChannelHandler(channelRepo)
	statsHandler := handlers.NewStatsHandler(logRepo, projectRepo, userProjectRepo, userRepo)
	pushHandler := handlers.NewPushHandler(subscriptionRepo, cfg)
	versionHandler := handlers.NewVersionHandler(Version)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Global middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key",
		AllowCredentials: false,
	}))

	// API routes
	api := app.Group("/api")

	// Version endpoint (public)
	api.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":    Version,
			"build_time": BuildTime,
			"git_commit": GitCommit,
		})
	})

	// Check for updates endpoint (public)
	api.Get("/version/check", versionHandler.CheckUpdate)

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/2fa/verify", twoFactorHandler.VerifyLogin) // Verify 2FA during login

	// Auth routes (protected)
	authProtected := auth.Group("", authMiddleware.RequireAuth())
	authProtected.Get("/me", authHandler.Me)
	authProtected.Put("/me", authHandler.UpdateProfile)
	authProtected.Put("/change-password", authHandler.ChangePassword)

	// Public log ingestion API (API key auth)
	v1 := api.Group("/v1")
	logIngestion := v1.Group("/logs", apiKeyMiddleware.RequireAPIKey())
	if rateLimitMiddleware != nil {
		logIngestion.Use(rateLimitMiddleware.RateLimitByProject())
	}
	logIngestion.Post("", logHandler.CreateLog)
	logIngestion.Post("/batch", logHandler.CreateBatchLogs)

	// Admin API (JWT auth)
	admin := api.Group("/admin", authMiddleware.RequireAuth())

	// 2FA routes (protected, all authenticated users)
	twoFactor := admin.Group("/2fa")
	twoFactor.Get("/status", twoFactorHandler.GetStatus)
	twoFactor.Post("/setup", twoFactorHandler.Setup)
	twoFactor.Post("/verify", twoFactorHandler.Verify)
	twoFactor.Post("/disable", twoFactorHandler.Disable)
	twoFactor.Post("/backup-codes", twoFactorHandler.RegenerateBackupCodes)

	// Users (admin only)
	users := admin.Group("/users", authMiddleware.RequireAdmin())
	users.Get("", userHandler.ListUsers)
	users.Post("", userHandler.CreateUser)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)
	users.Put("/:id/reset-password", userHandler.ResetPassword)

	// Projects
	projects := admin.Group("/projects")
	projects.Get("", projectHandler.ListProjects)
	projects.Post("", projectHandler.CreateProject)
	projects.Get("/:id", rbacMiddleware.RequireProjectAccess(), projectHandler.GetProject)
	projects.Put("/:id", rbacMiddleware.RequireOwner(), projectHandler.UpdateProject)
	projects.Delete("/:id", rbacMiddleware.RequireOwner(), projectHandler.DeleteProject)
	projects.Post("/:id/rotate-key", rbacMiddleware.RequireOwner(), projectHandler.RotateAPIKey)

	// Project members
	projects.Get("/:id/members", rbacMiddleware.RequireProjectAccess(), memberHandler.ListMembers)
	projects.Post("/:id/members", rbacMiddleware.RequireOwner(), memberHandler.AddMember)
	projects.Put("/:id/members/:uid", rbacMiddleware.RequireOwner(), memberHandler.UpdateMember)
	projects.Delete("/:id/members/:uid", rbacMiddleware.RequireOwner(), memberHandler.RemoveMember)

	// Project channels
	projects.Get("/:id/channels", rbacMiddleware.RequireProjectAccess(), channelHandler.ListChannels)
	projects.Post("/:id/channels", rbacMiddleware.RequireOwnerOrMember(), channelHandler.CreateChannel)

	// Channels
	channels := admin.Group("/channels")
	channels.Get("/:id", channelHandler.GetChannel)
	channels.Put("/:id", channelHandler.UpdateChannel)
	channels.Delete("/:id", channelHandler.DeleteChannel)
	channels.Post("/:id/test", channelHandler.TestChannel)

	// Logs
	logs := admin.Group("/logs")
	logs.Get("", logHandler.ListLogs)
	logs.Get("/:id", logHandler.GetLog)

	// Stats
	stats := admin.Group("/stats")
	stats.Get("/overview", statsHandler.GetOverview)
	stats.Get("/projects/:id", statsHandler.GetProjectStats)

	// Push notification routes
	push := api.Group("/push")
	push.Get("/vapid-key", pushHandler.GetVAPIDPublicKey) // Public - get VAPID key
	pushProtected := push.Group("", authMiddleware.RequireAuth())
	pushProtected.Post("/subscribe", pushHandler.Subscribe)
	pushProtected.Post("/unsubscribe", pushHandler.Unsubscribe)
	pushProtected.Get("/subscriptions", pushHandler.ListSubscriptions)
	pushProtected.Post("/test", pushHandler.TestNotification)

	// WebSocket routes for real-time logs
	app.Use("/ws", wsHandler.Upgrade())
	app.Get("/ws/logs", wsHandler.HandleLogs())

	// Serve embedded frontend (SPA)
	frontendFS, err := web.GetFileSystem()
	if err == nil {
		app.Use("/", filesystem.New(filesystem.Config{
			Root:         http.FS(frontendFS),
			Browse:       false,
			Index:        "index.html",
			NotFoundFile: "index.html",
		}))
	} else {
		log.Printf("Frontend not embedded, serving API only")
		app.Get("/", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"name":    "Central Logs API",
				"version": "1.0.0",
			})
		})
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		app.Shutdown()
	}()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting server on %s (env: %s)", addr, cfg.Server.Env)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createInitialAdmin(userRepo *models.UserRepository, cfg *config.Config) error {
	count, err := userRepo.Count()
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Users already exist
	}

	admin := &models.User{
		Username: cfg.Admin.Username,
		Password: cfg.Admin.Password,
		Name:     "Admin",
		Role:     models.RoleAdmin,
		IsActive: true,
	}

	if err := userRepo.Create(admin); err != nil {
		return err
	}

	log.Printf("Created initial admin user: %s", cfg.Admin.Username)
	return nil
}
