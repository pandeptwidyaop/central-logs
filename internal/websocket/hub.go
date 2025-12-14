package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Client represents a WebSocket client connection
type Client struct {
	Conn      *websocket.Conn
	UserID    string
	ProjectID string // Empty means subscribed to all projects user has access to
}

// Hub manages WebSocket connections and broadcasting
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *LogMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// LogMessage represents a log entry to broadcast
type LogMessage struct {
	Type      string      `json:"type"` // "log", "ping", etc.
	Data      interface{} `json:"data"`
	ProjectID string      `json:"project_id"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *LogMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client connected: user=%s, project=%s", client.UserID, client.ProjectID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Conn.Close()
			}
			h.mu.Unlock()
			log.Printf("[WS] Client disconnected: user=%s", client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			sentCount := 0
			for client := range h.clients {
				// Send to clients subscribed to this project or all projects
				if client.ProjectID == "" || client.ProjectID == message.ProjectID {
					data, err := json.Marshal(message)
					if err != nil {
						log.Printf("[WS] Error marshaling message: %v", err)
						continue
					}

					if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
						log.Printf("[WS] Error sending message to user %s: %v", client.UserID, err)
						h.mu.RUnlock()
						h.unregister <- client
						h.mu.RLock()
					} else {
						sentCount++
					}
				}
			}
			h.mu.RUnlock()
			if sentCount > 0 {
				log.Printf("[WS] Broadcast log to %d clients", sentCount)
			}
		}
	}
}

// Register adds a new client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// BroadcastLog sends a log entry to all relevant clients
func (h *Hub) BroadcastLog(logData interface{}, projectID string) {
	h.broadcast <- &LogMessage{
		Type:      "log",
		Data:      logData,
		ProjectID: projectID,
	}
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HasActiveConnection checks if a user has an active WebSocket connection
func (h *Hub) HasActiveConnection(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.UserID == userID {
			return true
		}
	}
	return false
}

// GetConnectedUserIDs returns a list of user IDs with active connections
func (h *Hub) GetConnectedUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	userIDs := make(map[string]bool)
	for client := range h.clients {
		userIDs[client.UserID] = true
	}
	result := make([]string, 0, len(userIDs))
	for userID := range userIDs {
		result = append(result, userID)
	}
	return result
}
