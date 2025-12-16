package handlers

import (
	"central-logs/internal/config"
	"central-logs/internal/services/telegram"

	"github.com/gofiber/fiber/v2"
)

type TelegramHandler struct {
	helper *telegram.Helper
	config *config.Config
}

func NewTelegramHandler(cfg *config.Config) *TelegramHandler {
	return &TelegramHandler{
		helper: telegram.NewHelper(),
		config: cfg,
	}
}

type GetChatsRequest struct {
	BotToken string `json:"bot_token"`
}

type GetChatsResponse struct {
	Chats []telegram.ChatInfo `json:"chats"`
}

type TestBotTokenRequest struct {
	BotToken string `json:"bot_token"`
}

type TestBotTokenResponse struct {
	Valid   bool   `json:"valid"`
	BotName string `json:"bot_name,omitempty"`
	Error   string `json:"error,omitempty"`
}

// GetRecentChats handles POST /api/admin/telegram/chats
func (h *TelegramHandler) GetRecentChats(c *fiber.Ctx) error {
	var req GetChatsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Use global bot token if not provided
	botToken := req.BotToken
	if botToken == "" {
		botToken = h.config.Telegram.BotToken
		if botToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No bot token provided and no global bot token configured",
			})
		}
	}

	chats, err := h.helper.GetRecentChats(botToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(chats) == 0 {
		return c.JSON(GetChatsResponse{
			Chats: []telegram.ChatInfo{},
		})
	}

	return c.JSON(GetChatsResponse{
		Chats: chats,
	})
}

// TestBotToken handles POST /api/admin/telegram/test
func (h *TelegramHandler) TestBotToken(c *fiber.Ctx) error {
	var req TestBotTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Use global bot token if not provided
	botToken := req.BotToken
	if botToken == "" {
		botToken = h.config.Telegram.BotToken
		if botToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No bot token provided and no global bot token configured",
			})
		}
	}

	valid, botName, err := h.helper.TestBotToken(botToken)
	if err != nil {
		return c.JSON(TestBotTokenResponse{
			Valid: false,
			Error: err.Error(),
		})
	}

	return c.JSON(TestBotTokenResponse{
		Valid:   valid,
		BotName: botName,
	})
}
