package handlers

import (
	"encoding/json"
	"time"

	"central-logs/internal/middleware"
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type MCPTokenHandler struct {
	mcpTokenRepo     *models.MCPTokenRepository
	mcpActivityRepo  *models.MCPActivityLogRepository
	projectRepo      *models.ProjectRepository
}

func NewMCPTokenHandler(
	mcpTokenRepo *models.MCPTokenRepository,
	mcpActivityRepo *models.MCPActivityLogRepository,
	projectRepo *models.ProjectRepository,
) *MCPTokenHandler {
	return &MCPTokenHandler{
		mcpTokenRepo:    mcpTokenRepo,
		mcpActivityRepo: mcpActivityRepo,
		projectRepo:     projectRepo,
	}
}

// ListTokens handles GET /api/admin/mcp/tokens
func (h *MCPTokenHandler) ListTokens(c *fiber.Ctx) error {
	tokens, err := h.mcpTokenRepo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list MCP tokens",
		})
	}

	return c.JSON(fiber.Map{
		"tokens": tokens,
	})
}

type CreateTokenRequest struct {
	Name            string   `json:"name"`
	GrantedProjects []string `json:"granted_projects"` // Array of project IDs or ["*"] for all
	ExpiresInDays   *int     `json:"expires_in_days"`  // Optional, null for permanent
}

type CreateTokenResponse struct {
	Token     string         `json:"token"`      // Full token (shown only once!)
	TokenInfo *models.MCPToken `json:"token_info"` // Token metadata
}

// CreateToken handles POST /api/admin/mcp/tokens
func (h *MCPTokenHandler) CreateToken(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req CreateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate name
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token name is required",
		})
	}

	// Validate granted projects
	if len(req.GrantedProjects) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one project must be granted or use ['*'] for all projects",
		})
	}

	// Build granted_projects JSON
	var grantedProjectsJSON string
	if len(req.GrantedProjects) == 1 && req.GrantedProjects[0] == "*" {
		grantedProjectsJSON = "*"
	} else {
		// Validate that all project IDs exist
		for _, projectID := range req.GrantedProjects {
			project, err := h.projectRepo.GetByID(projectID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to validate projects",
				})
			}
			if project == nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid project ID: " + projectID,
				})
			}
		}

		// Convert to JSON array
		jsonBytes, err := json.Marshal(req.GrantedProjects)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to serialize granted projects",
			})
		}
		grantedProjectsJSON = string(jsonBytes)
	}

	// Calculate expiry if specified
	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		expiry := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &expiry
	}

	// Create token
	token := &models.MCPToken{
		Name:            req.Name,
		GrantedProjects: grantedProjectsJSON,
		ExpiresAt:       expiresAt,
		IsActive:        true,
		CreatedBy:       user.ID,
	}

	rawToken, err := h.mcpTokenRepo.Create(token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create MCP token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateTokenResponse{
		Token:     rawToken,
		TokenInfo: token,
	})
}

// GetToken handles GET /api/admin/mcp/tokens/:id
func (h *MCPTokenHandler) GetToken(c *fiber.Ctx) error {
	tokenID := c.Params("id")

	token, err := h.mcpTokenRepo.GetByID(tokenID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get MCP token",
		})
	}

	if token == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Token not found",
		})
	}

	return c.JSON(token)
}

type UpdateTokenRequest struct {
	Name            *string  `json:"name"`
	GrantedProjects []string `json:"granted_projects"`
	ExpiresInDays   *int     `json:"expires_in_days"` // null to keep current, 0 for permanent, >0 for days from now
	IsActive        *bool    `json:"is_active"`
}

// UpdateToken handles PUT /api/admin/mcp/tokens/:id
func (h *MCPTokenHandler) UpdateToken(c *fiber.Ctx) error {
	tokenID := c.Params("id")

	var req UpdateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing token
	token, err := h.mcpTokenRepo.GetByID(tokenID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get MCP token",
		})
	}
	if token == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Token not found",
		})
	}

	// Update fields if provided
	if req.Name != nil {
		if *req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Token name cannot be empty",
			})
		}
		token.Name = *req.Name
	}

	if len(req.GrantedProjects) > 0 {
		var grantedProjectsJSON string
		if len(req.GrantedProjects) == 1 && req.GrantedProjects[0] == "*" {
			grantedProjectsJSON = "*"
		} else {
			// Validate projects
			for _, projectID := range req.GrantedProjects {
				project, err := h.projectRepo.GetByID(projectID)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to validate projects",
					})
				}
				if project == nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Invalid project ID: " + projectID,
					})
				}
			}

			jsonBytes, err := json.Marshal(req.GrantedProjects)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to serialize granted projects",
				})
			}
			grantedProjectsJSON = string(jsonBytes)
		}
		token.GrantedProjects = grantedProjectsJSON
	}

	if req.ExpiresInDays != nil {
		if *req.ExpiresInDays == 0 {
			// 0 means permanent (null)
			token.ExpiresAt = nil
		} else if *req.ExpiresInDays > 0 {
			// Positive means days from now
			expiry := time.Now().AddDate(0, 0, *req.ExpiresInDays)
			token.ExpiresAt = &expiry
		}
		// Negative values ignored (keep current)
	}

	if req.IsActive != nil {
		token.IsActive = *req.IsActive
	}

	if err := h.mcpTokenRepo.Update(token); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update MCP token",
		})
	}

	return c.JSON(token)
}

// DeleteToken handles DELETE /api/admin/mcp/tokens/:id
func (h *MCPTokenHandler) DeleteToken(c *fiber.Ctx) error {
	tokenID := c.Params("id")

	if err := h.mcpTokenRepo.Delete(tokenID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete MCP token",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Token deleted successfully",
	})
}

type GetTokenActivityResponse struct {
	Logs   []*models.MCPActivityLog `json:"logs"`
	Total  int                      `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}

// GetTokenActivity handles GET /api/admin/mcp/tokens/:id/activity
func (h *MCPTokenHandler) GetTokenActivity(c *fiber.Ctx) error {
	tokenID := c.Params("id")

	// Parse pagination params
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	if limit > 500 {
		limit = 500
	}

	logs, total, err := h.mcpActivityRepo.GetByTokenID(tokenID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get token activity",
		})
	}

	return c.JSON(GetTokenActivityResponse{
		Logs:   logs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
