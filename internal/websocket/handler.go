package websocket

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
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
		// Get user ID from query params (passed after JWT validation)
		userID := c.Query("user_id")
		projectID := c.Query("project_id") // Optional: filter by project

		if userID == "" {
			log.Println("[WS] Connection rejected: no user_id")
			c.Close()
			return
		}

		client := &Client{
			Conn:      c,
			UserID:    userID,
			ProjectID: projectID,
		}

		h.hub.Register(client)
		defer h.hub.Unregister(client)

		// Keep connection alive and listen for client messages
		for {
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[WS] Client closed connection: %v", err)
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
