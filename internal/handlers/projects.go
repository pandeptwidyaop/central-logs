package handlers

import (
	"central-logs/internal/middleware"
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type ProjectHandler struct {
	projectRepo     *models.ProjectRepository
	userProjectRepo *models.UserProjectRepository
	logRepo         *models.LogRepository
}

func NewProjectHandler(
	projectRepo *models.ProjectRepository,
	userProjectRepo *models.UserProjectRepository,
	logRepo *models.LogRepository,
) *ProjectHandler {
	return &ProjectHandler{
		projectRepo:     projectRepo,
		userProjectRepo: userProjectRepo,
		logRepo:         logRepo,
	}
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateProjectResponse struct {
	Project *models.Project `json:"project"`
	APIKey  string          `json:"api_key"`
}

// ListProjects handles GET /api/admin/projects
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var projects []*models.Project
	var err error

	if user.IsAdmin() {
		projects, err = h.projectRepo.GetAll()
	} else {
		projects, err = h.projectRepo.GetByUserID(user.ID)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list projects",
		})
	}

	return c.JSON(fiber.Map{
		"projects": projects,
	})
}

// CreateProject handles POST /api/admin/projects
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	apiKey, err := h.projectRepo.Create(project)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create project",
		})
	}

	// Add creator as owner
	userProject := &models.UserProject{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      models.ProjectRoleOwner,
	}
	if err := h.userProjectRepo.Create(userProject); err != nil {
		// Rollback project creation
		h.projectRepo.Delete(project.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to assign owner",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateProjectResponse{
		Project: project,
		APIKey:  apiKey,
	})
}

// GetProject handles GET /api/admin/projects/:id
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	projectID := c.Params("id")
	project, err := h.projectRepo.GetByID(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get project",
		})
	}

	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check access (unless admin)
	if !user.IsAdmin() {
		hasAccess, _ := h.userProjectRepo.HasAccess(user.ID, projectID)
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	// Get log stats
	stats, _ := h.logRepo.GetProjectStats(projectID)

	return c.JSON(fiber.Map{
		"project": project,
		"stats":   stats,
	})
}

type UpdateProjectRequest struct {
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	IsActive        *bool                   `json:"is_active"`
	RetentionConfig *models.RetentionConfig `json:"retention_config"`
}

// UpdateProject handles PUT /api/admin/projects/:id
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	projectID := c.Params("id")
	project, err := h.projectRepo.GetByID(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get project",
		})
	}

	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var req UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.IsActive != nil {
		project.IsActive = *req.IsActive
	}
	if req.RetentionConfig != nil {
		project.RetentionConfig = req.RetentionConfig
	}

	if err := h.projectRepo.Update(project); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update project",
		})
	}

	return c.JSON(project)
}

// DeleteProject handles DELETE /api/admin/projects/:id
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	projectID := c.Params("id")

	if err := h.projectRepo.Delete(projectID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete project",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Project deleted",
	})
}

// RotateAPIKey handles POST /api/admin/projects/:id/rotate-key
func (h *ProjectHandler) RotateAPIKey(c *fiber.Ctx) error {
	projectID := c.Params("id")

	project, err := h.projectRepo.GetByID(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get project",
		})
	}

	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	newAPIKey, err := h.projectRepo.RotateAPIKey(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to rotate API key",
		})
	}

	// Get updated project
	project, _ = h.projectRepo.GetByID(projectID)

	return c.JSON(fiber.Map{
		"api_key":        newAPIKey,
		"api_key_prefix": project.APIKeyPrefix,
	})
}
