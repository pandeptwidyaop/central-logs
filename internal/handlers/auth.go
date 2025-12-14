package handlers

import (
	"central-logs/internal/middleware"
	"central-logs/internal/models"
	"central-logs/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	userRepo   *models.UserRepository
	jwtManager *utils.JWTManager
}

func NewAuthHandler(userRepo *models.UserRepository, jwtManager *utils.JWTManager) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token       string       `json:"token,omitempty"`
	User        *models.User `json:"user,omitempty"`
	Requires2FA bool         `json:"requires_2fa,omitempty"`
	TempToken   string       `json:"temp_token,omitempty"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to authenticate",
		})
	}

	if user == nil || !user.CheckPassword(req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	if !user.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Account is deactivated",
		})
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temp token for 2FA verification
		tempToken, err := h.jwtManager.GenerateTempToken(user.ID, user.Username, string(user.Role))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		return c.JSON(LoginResponse{
			Requires2FA: true,
			TempToken:   tempToken,
		})
	}

	token, err := h.jwtManager.Generate(user.ID, user.Username, string(user.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	return c.JSON(user)
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if err := h.userRepo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(user)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Current and new password are required",
		})
	}

	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be at least 8 characters",
		})
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Current password is incorrect",
		})
	}

	if err := h.userRepo.UpdatePassword(user.ID, req.NewPassword); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to change password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}
