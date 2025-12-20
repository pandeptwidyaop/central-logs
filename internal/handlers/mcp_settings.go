package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type MCPSettingsHandler struct {
	mcpEnabled *bool // Pointer to shared MCP enabled state
}

func NewMCPSettingsHandler(mcpEnabled *bool) *MCPSettingsHandler {
	return &MCPSettingsHandler{
		mcpEnabled: mcpEnabled,
	}
}

type MCPStatusResponse struct {
	Enabled bool `json:"enabled"`
}

// GetMCPStatus handles GET /api/admin/mcp/status
func (h *MCPSettingsHandler) GetMCPStatus(c *fiber.Ctx) error {
	return c.JSON(MCPStatusResponse{
		Enabled: *h.mcpEnabled,
	})
}

type ToggleMCPRequest struct {
	Enabled bool `json:"enabled"`
}

// ToggleMCP handles POST /api/admin/mcp/toggle
func (h *MCPSettingsHandler) ToggleMCP(c *fiber.Ctx) error {
	var req ToggleMCPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	*h.mcpEnabled = req.Enabled

	return c.JSON(fiber.Map{
		"message": "MCP server toggled successfully",
		"enabled": *h.mcpEnabled,
	})
}
