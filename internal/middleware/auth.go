package middleware

import (
	"strings"

	"central-logs/internal/models"
	"central-logs/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	jwtManager *utils.JWTManager
	userRepo   *models.UserRepository
}

func NewAuthMiddleware(jwtManager *utils.JWTManager, userRepo *models.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		userRepo:   userRepo,
	}
}

// RequireAuth validates JWT token and sets user in context
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token := parts[1]
		claims, err := m.jwtManager.Validate(token)
		if err != nil {
			if err == utils.ErrExpiredToken {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token has expired",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Get user from database
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get user",
			})
		}
		if user == nil || !user.IsActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found or inactive",
			})
		}

		// Set user in context
		c.Locals("user", user)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// RequireAdmin requires user to be an admin
func (m *AuthMiddleware) RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		if !user.IsAdmin() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}
		return c.Next()
	}
}

// GetUser returns the authenticated user from context
func GetUser(c *fiber.Ctx) *models.User {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

// GetClaims returns the JWT claims from context
func GetClaims(c *fiber.Ctx) *utils.JWTClaims {
	claims, ok := c.Locals("claims").(*utils.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
