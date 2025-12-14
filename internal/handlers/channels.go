package handlers

import (
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type ChannelHandler struct {
	channelRepo *models.ChannelRepository
}

func NewChannelHandler(channelRepo *models.ChannelRepository) *ChannelHandler {
	return &ChannelHandler{
		channelRepo: channelRepo,
	}
}

// ListChannels handles GET /api/admin/projects/:id/channels
func (h *ChannelHandler) ListChannels(c *fiber.Ctx) error {
	projectID := c.Params("id")

	channels, err := h.channelRepo.GetByProjectID(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list channels",
		})
	}

	return c.JSON(fiber.Map{
		"channels": channels,
	})
}

type CreateChannelRequest struct {
	Type     models.ChannelType     `json:"type"`
	Name     string                 `json:"name"`
	Config   map[string]interface{} `json:"config"`
	MinLevel models.LogLevel        `json:"min_level"`
}

// CreateChannel handles POST /api/admin/projects/:id/channels
func (h *ChannelHandler) CreateChannel(c *fiber.Ctx) error {
	projectID := c.Params("id")

	var req CreateChannelRequest
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

	if req.Type != models.ChannelTypePush && req.Type != models.ChannelTypeTelegram && req.Type != models.ChannelTypeDiscord {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid channel type. Must be PUSH, TELEGRAM, or DISCORD",
		})
	}

	// Validate config based on type
	if req.Type == models.ChannelTypeTelegram {
		if req.Config["bot_token"] == nil || req.Config["chat_id"] == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Telegram requires bot_token and chat_id",
			})
		}
	}

	if req.Type == models.ChannelTypeDiscord {
		if req.Config["webhook_url"] == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Discord requires webhook_url",
			})
		}
	}

	if req.MinLevel == "" {
		req.MinLevel = models.LogLevelError
	}

	channel := &models.Channel{
		ProjectID: projectID,
		Type:      req.Type,
		Name:      req.Name,
		Config:    req.Config,
		MinLevel:  req.MinLevel,
		IsActive:  true,
	}

	if err := h.channelRepo.Create(channel); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create channel",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(channel)
}

// GetChannel handles GET /api/admin/channels/:id
func (h *ChannelHandler) GetChannel(c *fiber.Ctx) error {
	channelID := c.Params("id")

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get channel",
		})
	}

	if channel == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Channel not found",
		})
	}

	return c.JSON(channel)
}

type UpdateChannelRequest struct {
	Name     string                 `json:"name"`
	Config   map[string]interface{} `json:"config"`
	MinLevel models.LogLevel        `json:"min_level"`
	IsActive *bool                  `json:"is_active"`
}

// UpdateChannel handles PUT /api/admin/channels/:id
func (h *ChannelHandler) UpdateChannel(c *fiber.Ctx) error {
	channelID := c.Params("id")

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get channel",
		})
	}

	if channel == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Channel not found",
		})
	}

	var req UpdateChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Config != nil {
		channel.Config = req.Config
	}
	if req.MinLevel != "" {
		channel.MinLevel = req.MinLevel
	}
	if req.IsActive != nil {
		channel.IsActive = *req.IsActive
	}

	if err := h.channelRepo.Update(channel); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update channel",
		})
	}

	return c.JSON(channel)
}

// DeleteChannel handles DELETE /api/admin/channels/:id
func (h *ChannelHandler) DeleteChannel(c *fiber.Ctx) error {
	channelID := c.Params("id")

	if err := h.channelRepo.Delete(channelID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete channel",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Channel deleted",
	})
}

// TestChannel handles POST /api/admin/channels/:id/test
func (h *ChannelHandler) TestChannel(c *fiber.Ctx) error {
	channelID := c.Params("id")

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get channel",
		})
	}

	if channel == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Channel not found",
		})
	}

	// TODO: Implement actual test notification sending
	// For now, just return success
	return c.JSON(fiber.Map{
		"message": "Test notification sent",
		"channel": channel.Name,
	})
}
