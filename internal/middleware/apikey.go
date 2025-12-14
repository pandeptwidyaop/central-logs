package middleware

import (
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type APIKeyMiddleware struct {
	projectRepo *models.ProjectRepository
}

func NewAPIKeyMiddleware(projectRepo *models.ProjectRepository) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		projectRepo: projectRepo,
	}
}

// RequireAPIKey validates the X-API-Key header and sets project in context
func (m *APIKeyMiddleware) RequireAPIKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing API key",
			})
		}

		project, err := m.projectRepo.GetByAPIKey(apiKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to validate API key",
			})
		}

		if project == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		if !project.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Project is inactive",
			})
		}

		// Set project in context
		c.Locals("project", project)

		return c.Next()
	}
}

// GetProject returns the project from context (set by APIKey middleware)
func GetProject(c *fiber.Ctx) *models.Project {
	project, ok := c.Locals("project").(*models.Project)
	if !ok {
		return nil
	}
	return project
}
