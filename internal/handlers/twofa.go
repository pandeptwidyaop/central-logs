package handlers

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"strings"

	"central-logs/internal/middleware"
	"central-logs/internal/models"
	"central-logs/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type TwoFactorHandler struct {
	userRepo   *models.UserRepository
	jwtManager *utils.JWTManager
	issuer     string
}

func NewTwoFactorHandler(userRepo *models.UserRepository, jwtManager *utils.JWTManager, issuer string) *TwoFactorHandler {
	if issuer == "" {
		issuer = "Central Logs"
	}
	return &TwoFactorHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		issuer:     issuer,
	}
}

type SetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

// Setup generates a new TOTP secret and returns QR code URL
func (h *TwoFactorHandler) Setup(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if user.TwoFactorEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Two-factor authentication is already enabled",
		})
	}

	// Generate a new TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      h.issuer,
		AccountName: user.Username,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate 2FA secret",
		})
	}

	// Store the secret temporarily (not enabled yet)
	if err := h.userRepo.UpdateTwoFactor(user.ID, key.Secret(), false, ""); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save 2FA secret",
		})
	}

	return c.JSON(SetupResponse{
		Secret: key.Secret(),
		QRCode: key.URL(),
	})
}

type VerifyRequest struct {
	Code string `json:"code"`
}

// Verify validates the TOTP code and enables 2FA
func (h *TwoFactorHandler) Verify(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get the latest user data to get the pending secret
	user, err := h.userRepo.GetByID(user.ID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if user.TwoFactorEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Two-factor authentication is already enabled",
		})
	}

	if user.TwoFactorSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No pending 2FA setup found. Please call setup first.",
		})
	}

	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Code is required",
		})
	}

	// Verify the code
	valid := totp.Validate(req.Code, user.TwoFactorSecret)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid verification code",
		})
	}

	// Generate backup codes
	backupCodes, hashedCodes, err := generateBackupCodes(8)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate backup codes",
		})
	}

	// Store hashed backup codes as JSON
	backupCodesJSON, _ := json.Marshal(hashedCodes)

	// Enable 2FA
	if err := h.userRepo.UpdateTwoFactor(user.ID, user.TwoFactorSecret, true, string(backupCodesJSON)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to enable 2FA",
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Two-factor authentication enabled successfully",
		"backup_codes": backupCodes,
	})
}

type DisableRequest struct {
	Code string `json:"code"`
}

// Disable removes 2FA (requires current TOTP code)
func (h *TwoFactorHandler) Disable(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get the latest user data
	user, err := h.userRepo.GetByID(user.ID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if !user.TwoFactorEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Two-factor authentication is not enabled",
		})
	}

	var req DisableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Code is required",
		})
	}

	// Verify the code (TOTP or backup code)
	if !h.verifyCode(user, req.Code) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid verification code",
		})
	}

	// Disable 2FA
	if err := h.userRepo.UpdateTwoFactor(user.ID, "", false, ""); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to disable 2FA",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Two-factor authentication disabled successfully",
	})
}

// RegenerateBackupCodes generates new backup codes (requires current TOTP code)
func (h *TwoFactorHandler) RegenerateBackupCodes(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get the latest user data
	user, err := h.userRepo.GetByID(user.ID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if !user.TwoFactorEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Two-factor authentication is not enabled",
		})
	}

	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify TOTP code (not backup code)
	if req.Code == "" || !totp.Validate(req.Code, user.TwoFactorSecret) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid verification code",
		})
	}

	// Generate new backup codes
	backupCodes, hashedCodes, err := generateBackupCodes(8)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate backup codes",
		})
	}

	// Store hashed backup codes
	backupCodesJSON, _ := json.Marshal(hashedCodes)
	if err := h.userRepo.UpdateBackupCodes(user.ID, string(backupCodesJSON)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save backup codes",
		})
	}

	return c.JSON(fiber.Map{
		"backup_codes": backupCodes,
	})
}

type VerifyLoginRequest struct {
	TempToken string `json:"temp_token"`
	Code      string `json:"code"`
}

// VerifyLogin validates the TOTP code during login and returns full JWT
func (h *TwoFactorHandler) VerifyLogin(c *fiber.Ctx) error {
	var req VerifyLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.TempToken == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Temp token and code are required",
		})
	}

	// Validate temp token and get user ID
	claims, err := h.jwtManager.ValidateTempToken(req.TempToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Get user
	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if !user.TwoFactorEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Two-factor authentication is not enabled for this user",
		})
	}

	// Verify the code (TOTP or backup code)
	if !h.verifyCode(user, req.Code) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid verification code",
		})
	}

	// Generate full JWT token
	token, err := h.jwtManager.Generate(user.ID, user.Username, string(user.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

// verifyCode checks if the code is a valid TOTP or backup code
func (h *TwoFactorHandler) verifyCode(user *models.User, code string) bool {
	// First, try TOTP
	if totp.Validate(code, user.TwoFactorSecret) {
		return true
	}

	// Then, try backup codes
	if user.BackupCodes == "" {
		return false
	}

	var hashedCodes []string
	if err := json.Unmarshal([]byte(user.BackupCodes), &hashedCodes); err != nil {
		return false
	}

	// Clean the input code (remove any dashes or spaces)
	cleanCode := strings.ReplaceAll(strings.ReplaceAll(code, "-", ""), " ", "")
	cleanCode = strings.ToUpper(cleanCode)

	for i, hashedCode := range hashedCodes {
		if bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(cleanCode)) == nil {
			// Remove used backup code
			hashedCodes = append(hashedCodes[:i], hashedCodes[i+1:]...)
			newCodesJSON, _ := json.Marshal(hashedCodes)
			h.userRepo.UpdateBackupCodes(user.ID, string(newCodesJSON))
			return true
		}
	}

	return false
}

// GetStatus returns the current 2FA status
func (h *TwoFactorHandler) GetStatus(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get the latest user data
	user, err := h.userRepo.GetByID(user.ID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Count remaining backup codes
	backupCodesCount := 0
	if user.BackupCodes != "" {
		var hashedCodes []string
		if err := json.Unmarshal([]byte(user.BackupCodes), &hashedCodes); err == nil {
			backupCodesCount = len(hashedCodes)
		}
	}

	return c.JSON(fiber.Map{
		"enabled":            user.TwoFactorEnabled,
		"backup_codes_count": backupCodesCount,
	})
}

// generateBackupCodes creates random backup codes and returns both plain and hashed versions
func generateBackupCodes(count int) ([]string, []string, error) {
	plainCodes := make([]string, count)
	hashedCodes := make([]string, count)

	for i := 0; i < count; i++ {
		// Generate 10 random bytes
		bytes := make([]byte, 5)
		if _, err := rand.Read(bytes); err != nil {
			return nil, nil, err
		}

		// Encode to base32 and format as XXXX-XXXX
		code := base32.StdEncoding.EncodeToString(bytes)[:8]
		formattedCode := code[:4] + "-" + code[4:]
		plainCodes[i] = formattedCode

		// Hash the code (without formatting)
		hashed, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, err
		}
		hashedCodes[i] = string(hashed)
	}

	return plainCodes, hashedCodes, nil
}

