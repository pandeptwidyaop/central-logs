package websocket

import (
	"log"
	"strings"

	"central-logs/internal/models"
	"central-logs/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Handler handles WebSocket connections
type Handler struct {
	hub        *Hub
	jwtManager *utils.JWTManager
	userRepo   *models.UserRepository
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, jwtManager *utils.JWTManager, userRepo *models.UserRepository) *Handler {
	return &Handler{
		hub:        hub,
		jwtManager: jwtManager,
		userRepo:   userRepo,
	}
}

// AuthMiddleware validates JWT token from WebSocket upgrade request headers
func (h *Handler) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// Try Sec-WebSocket-Protocol for compatibility
			protocols := c.Get("Sec-WebSocket-Protocol")
			if protocols != "" {
				// Format: "token, <actual-token-value>"
				parts := strings.Split(protocols, ", ")
				if len(parts) == 2 && parts[0] == "token" {
					authHeader = "Bearer " + parts[1]
				}
			}
		}

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token := parts[1]
		claims, err := h.jwtManager.Validate(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Get user from database
		user, err := h.userRepo.GetByID(claims.UserID)
		if err != nil || user == nil || !user.IsActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found or inactive",
			})
		}

		// Store user in locals for WebSocket handler
		c.Locals("user", user)
		c.Locals("user_id", user.ID)

		// Set Sec-WebSocket-Protocol response header if client sent protocols
		// This is required by the WebSocket spec when using subprotocols
		if c.Get("Sec-WebSocket-Protocol") != "" {
			c.Set("Sec-WebSocket-Protocol", "token")
		}

		return c.Next()
	}
}

// Upgrade middleware to check if request is a WebSocket upgrade
func (h *Handler) Upgrade() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleLogs handles WebSocket connections for log streaming
func (h *Handler) HandleLogs() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Get user from Locals (set by AuthMiddleware)
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			log.Println("[WS] Connection rejected: no authenticated user")
			c.Close()
			return
		}

		user, ok := c.Locals("user").(*models.User)
		if !ok {
			log.Println("[WS] Connection rejected: invalid user data")
			c.Close()
			return
		}

		// Optional: Get project filter from query params (still allowed for filtering)
		projectID := c.Query("project_id")

		log.Printf("[WS] Client connected: user=%s, project=%s", user.Username, projectID)

		client := &Client{
			Conn:      c,
			UserID:    userID,
			ProjectID: projectID,
		}

		h.hub.Register(client)
		defer func() {
			h.hub.Unregister(client)
			log.Printf("[WS] Client disconnected: user=%s", user.Username)
		}()

		// Keep connection alive and listen for client messages
		for {
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[WS] Client closed connection: %v", err)
				} else {
					log.Printf("[WS] Read error: %v", err)
				}
				break
			}

			// Handle ping/pong or other client messages
			if messageType == websocket.TextMessage {
				// Echo back pings
				if string(msg) == "ping" {
					c.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
				}
			}
		}
	})
}

// GetHub returns the hub instance
func (h *Handler) GetHub() *Hub {
	return h.hub
}
