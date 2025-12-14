package handlers

import (
	"encoding/json"
	"log"

	"central-logs/internal/config"
	"central-logs/internal/middleware"
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
	webpush "github.com/SherClockHolmes/webpush-go"
)

type PushHandler struct {
	subscriptionRepo *models.PushSubscriptionRepository
	config           *config.Config
}

func NewPushHandler(subscriptionRepo *models.PushSubscriptionRepository, cfg *config.Config) *PushHandler {
	return &PushHandler{
		subscriptionRepo: subscriptionRepo,
		config:           cfg,
	}
}

// GetVAPIDPublicKey returns the VAPID public key for push subscriptions
func (h *PushHandler) GetVAPIDPublicKey(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"public_key": h.config.VAPID.PublicKey,
	})
}

type SubscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// Subscribe handles push subscription registration
func (h *PushHandler) Subscribe(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID := user.ID

	var req SubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields: endpoint, keys.p256dh, keys.auth",
		})
	}

	// Check if subscription already exists
	existing, err := h.subscriptionRepo.GetByEndpoint(req.Endpoint)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing subscription",
		})
	}

	if existing != nil {
		// Update existing subscription (e.g., if user re-subscribes)
		return c.JSON(fiber.Map{
			"message": "Already subscribed",
			"id":      existing.ID,
		})
	}

	// Create new subscription
	subscription := &models.PushSubscription{
		UserID:    userID,
		Endpoint:  req.Endpoint,
		P256dh:    req.Keys.P256dh,
		Auth:      req.Keys.Auth,
		UserAgent: c.Get("User-Agent"),
	}

	if err := h.subscriptionRepo.Create(subscription); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save subscription",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Subscribed successfully",
		"id":      subscription.ID,
	})
}

type UnsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

// Unsubscribe handles push subscription removal
func (h *PushHandler) Unsubscribe(c *fiber.Ctx) error {
	var req UnsubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Endpoint is required",
		})
	}

	if err := h.subscriptionRepo.DeleteByEndpoint(req.Endpoint); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove subscription",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Unsubscribed successfully",
	})
}

// ListSubscriptions returns all subscriptions for the current user
func (h *PushHandler) ListSubscriptions(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID := user.ID

	subs, err := h.subscriptionRepo.GetByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list subscriptions",
		})
	}

	return c.JSON(fiber.Map{
		"subscriptions": subs,
	})
}

// TestNotification sends a test push notification to the user
func (h *PushHandler) TestNotification(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get user's subscriptions
	subs, err := h.subscriptionRepo.GetByUserID(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subscriptions",
		})
	}

	if len(subs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No push subscriptions found. Please enable browser notifications first.",
		})
	}

	// Create test payload
	payload := map[string]interface{}{
		"title": "🔔 Test Notification",
		"body":  "Push notifications are working correctly!",
		"icon":  "/icons/image.png",
		"badge": "/icons/image.png",
		"tag":   "test-notification",
		"url":   "/",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create notification payload",
		})
	}

	log.Printf("[Push] Test notification payload: %s", string(payloadBytes))

	// Send to all user's subscriptions
	sentCount := 0
	for _, sub := range subs {
		subscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}

		resp, err := webpush.SendNotification(payloadBytes, subscription, &webpush.Options{
			Subscriber:      h.config.VAPID.Subject,
			VAPIDPublicKey:  h.config.VAPID.PublicKey,
			VAPIDPrivateKey: h.config.VAPID.PrivateKey,
			TTL:             60,
		})

		if err != nil {
			log.Printf("[Push] Failed to send test notification: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			sentCount++
		} else if resp.StatusCode == 410 {
			// Subscription expired, remove it
			h.subscriptionRepo.DeleteByEndpoint(sub.Endpoint)
		}
	}

	if sentCount == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send notification to any device",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Test notification sent",
		"sent_to": sentCount,
	})
}
