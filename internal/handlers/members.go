package handlers

import (
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type MemberHandler struct {
	userRepo        *models.UserRepository
	userProjectRepo *models.UserProjectRepository
}

func NewMemberHandler(
	userRepo *models.UserRepository,
	userProjectRepo *models.UserProjectRepository,
) *MemberHandler {
	return &MemberHandler{
		userRepo:        userRepo,
		userProjectRepo: userProjectRepo,
	}
}

// ListMembers handles GET /api/admin/projects/:id/members
func (h *MemberHandler) ListMembers(c *fiber.Ctx) error {
	projectID := c.Params("id")

	members, err := h.userProjectRepo.GetProjectMembers(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list members",
		})
	}

	return c.JSON(fiber.Map{
		"members": members,
	})
}

type AddMemberRequest struct {
	UserID   string             `json:"user_id"`
	Username string             `json:"username"`
	Role     models.ProjectRole `json:"role"`
}

// AddMember handles POST /api/admin/projects/:id/members
func (h *MemberHandler) AddMember(c *fiber.Ctx) error {
	projectID := c.Params("id")

	var req AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" && req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID or username is required",
		})
	}

	// Validate role
	if req.Role != models.ProjectRoleOwner && req.Role != models.ProjectRoleMember && req.Role != models.ProjectRoleViewer {
		req.Role = models.ProjectRoleMember
	}

	// Find user by ID or username
	var user *models.User
	var err error
	if req.UserID != "" {
		user, err = h.userRepo.GetByID(req.UserID)
	} else {
		user, err = h.userRepo.GetByUsername(req.Username)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to find user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Check if already a member
	existing, _ := h.userProjectRepo.GetByUserAndProject(user.ID, projectID)
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User is already a member",
		})
	}

	// Add member
	userProject := &models.UserProject{
		UserID:    user.ID,
		ProjectID: projectID,
		Role:      req.Role,
	}

	if err := h.userProjectRepo.Create(userProject); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add member",
		})
	}

	userProject.User = user
	return c.Status(fiber.StatusCreated).JSON(userProject)
}

type UpdateMemberRequest struct {
	Role models.ProjectRole `json:"role"`
}

// UpdateMember handles PUT /api/admin/projects/:id/members/:uid
func (h *MemberHandler) UpdateMember(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Params("uid")

	var req UpdateMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate role
	if req.Role != models.ProjectRoleOwner && req.Role != models.ProjectRoleMember && req.Role != models.ProjectRoleViewer {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role",
		})
	}

	// Check if member exists
	existing, _ := h.userProjectRepo.GetByUserAndProject(userID, projectID)
	if existing == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Member not found",
		})
	}

	if err := h.userProjectRepo.UpdateRole(userID, projectID, req.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update member",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Member updated",
		"role":    req.Role,
	})
}

// RemoveMember handles DELETE /api/admin/projects/:id/members/:uid
func (h *MemberHandler) RemoveMember(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Params("uid")

	// Check if member exists
	existing, _ := h.userProjectRepo.GetByUserAndProject(userID, projectID)
	if existing == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Member not found",
		})
	}

	if err := h.userProjectRepo.Delete(userID, projectID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove member",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Member removed",
	})
}
