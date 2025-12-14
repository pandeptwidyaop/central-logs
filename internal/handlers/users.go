package handlers

import (
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userRepo *models.UserRepository
}

func NewUserHandler(userRepo *models.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// ListUsers handles GET /api/admin/users (Admin only)
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.userRepo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
	})
}

type CreateUserRequest struct {
	Username string          `json:"username"`
	Password string          `json:"password"`
	Name     string          `json:"name"`
	Role     models.UserRole `json:"role"`
}

// CreateUser handles POST /api/admin/users (Admin only)
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Username == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username, password, and name are required",
		})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 8 characters",
		})
	}

	// Check if username exists
	existing, _ := h.userRepo.GetByUsername(req.Username)
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username already in use",
		})
	}

	// Validate role
	if req.Role != models.RoleAdmin && req.Role != models.RoleUser {
		req.Role = models.RoleUser
	}

	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		Name:     req.Name,
		Role:     req.Role,
		IsActive: true,
	}

	if err := h.userRepo.Create(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetUser handles GET /api/admin/users/:id (Admin only)
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

type UpdateUserRequest struct {
	Name     string          `json:"name"`
	Role     models.UserRole `json:"role"`
	IsActive *bool           `json:"is_active"`
}

// UpdateUser handles PUT /api/admin/users/:id (Admin only)
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		if req.Role == models.RoleAdmin || req.Role == models.RoleUser {
			user.Role = req.Role
		}
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.userRepo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.JSON(user)
}

// DeleteUser handles DELETE /api/admin/users/:id (Admin only)
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	if err := h.userRepo.Delete(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted",
	})
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

// ResetPassword handles PUT /api/admin/users/:id/reset-password (Admin only)
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 8 characters",
		})
	}

	if err := h.userRepo.UpdatePassword(userID, req.Password); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset successfully",
	})
}
