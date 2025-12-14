package middleware

import (
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type RBACMiddleware struct {
	userProjectRepo *models.UserProjectRepository
}

func NewRBACMiddleware(userProjectRepo *models.UserProjectRepository) *RBACMiddleware {
	return &RBACMiddleware{
		userProjectRepo: userProjectRepo,
	}
}

// RequireProjectAccess checks if user has any access to the project
func (m *RBACMiddleware) RequireProjectAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Admins have access to all projects
		if user.IsAdmin() {
			return c.Next()
		}

		projectID := c.Params("id")
		if projectID == "" {
			projectID = c.Params("projectId")
		}
		if projectID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID required",
			})
		}

		hasAccess, err := m.userProjectRepo.HasAccess(user.ID, projectID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check access",
			})
		}

		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied to this project",
			})
		}

		return c.Next()
	}
}

// RequireProjectRole checks if user has specific role(s) in the project
func (m *RBACMiddleware) RequireProjectRole(roles ...models.ProjectRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Admins have all roles
		if user.IsAdmin() {
			return c.Next()
		}

		projectID := c.Params("id")
		if projectID == "" {
			projectID = c.Params("projectId")
		}
		if projectID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID required",
			})
		}

		hasRole, err := m.userProjectRepo.HasRole(user.ID, projectID, roles...)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check role",
			})
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireOwner is a shortcut for requiring OWNER role
func (m *RBACMiddleware) RequireOwner() fiber.Handler {
	return m.RequireProjectRole(models.ProjectRoleOwner)
}

// RequireOwnerOrMember requires OWNER or MEMBER role
func (m *RBACMiddleware) RequireOwnerOrMember() fiber.Handler {
	return m.RequireProjectRole(models.ProjectRoleOwner, models.ProjectRoleMember)
}
