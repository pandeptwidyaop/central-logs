package notification

import (
	"encoding/json"
	"log"

	"central-logs/internal/config"
	"central-logs/internal/models"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type PushService struct {
	subscriptionRepo *models.PushSubscriptionRepository
	channelRepo      *models.ChannelRepository
	config           *config.Config
}

func NewPushService(
	subscriptionRepo *models.PushSubscriptionRepository,
	channelRepo *models.ChannelRepository,
	cfg *config.Config,
) *PushService {
	return &PushService{
		subscriptionRepo: subscriptionRepo,
		channelRepo:      channelRepo,
		config:           cfg,
	}
}

type PushPayload struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Icon      string `json:"icon,omitempty"`
	Badge     string `json:"badge,omitempty"`
	Tag       string `json:"tag,omitempty"`
	URL       string `json:"url,omitempty"`
	LogID     string `json:"log_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Level     string `json:"level,omitempty"`
}

// SendLogNotification sends push notifications for a new log entry
func (s *PushService) SendLogNotification(logEntry *models.Log, projectName string) error {
	if s.config.VAPID.PublicKey == "" || s.config.VAPID.PrivateKey == "" {
		log.Println("[Push] VAPID keys not configured, skipping push notifications")
		return nil
	}

	// Get active push channels for this project
	channels, err := s.channelRepo.GetActiveByProjectID(logEntry.ProjectID)
	if err != nil {
		return err
	}

	// Find push channel and check if we should notify for this level
	var pushChannel *models.Channel
	for _, ch := range channels {
		if ch.Type == models.ChannelTypePush {
			pushChannel = ch
			break
		}
	}

	if pushChannel == nil {
		// No push channel configured for this project
		return nil
	}

	// Check if log level meets minimum threshold
	if !pushChannel.ShouldNotify(logEntry.Level) {
		return nil
	}

	// Get all push subscriptions for this project
	subscriptions, err := s.subscriptionRepo.GetByProjectID(logEntry.ProjectID)
	if err != nil {
		return err
	}

	if len(subscriptions) == 0 {
		return nil
	}

	// Create push payload
	payload := PushPayload{
		Title:     getLevelEmoji(logEntry.Level) + " " + string(logEntry.Level) + " - " + projectName,
		Body:      truncateString(logEntry.Message, 200),
		Icon:      "/icons/image.png",
		Badge:     "/icons/image.png",
		Tag:       "log-" + string(logEntry.Level) + "-" + logEntry.ID,
		URL:       "/logs?id=" + logEntry.ID,
		LogID:     logEntry.ID,
		ProjectID: logEntry.ProjectID,
		Level:     string(logEntry.Level),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send to all subscriptions
	// Service worker will check visibility and skip if page is visible
	for _, sub := range subscriptions {
		go s.sendPush(sub, payloadBytes)
	}

	return nil
}

func (s *PushService) sendPush(sub *models.PushSubscription, payload []byte) {
	log.Printf("[Push] Sending payload: %s", string(payload))

	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
		Subscriber:      s.config.VAPID.Subject,
		VAPIDPublicKey:  s.config.VAPID.PublicKey,
		VAPIDPrivateKey: s.config.VAPID.PrivateKey,
		TTL:             60,
	})

	if err != nil {
		log.Printf("[Push] Failed to send notification to %s: %v", sub.Endpoint[:50], err)
		// If subscription is invalid (410 Gone), remove it
		if resp != nil && resp.StatusCode == 410 {
			log.Printf("[Push] Subscription expired, removing: %s", sub.Endpoint[:50])
			s.subscriptionRepo.DeleteByEndpoint(sub.Endpoint)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("[Push] Notification sent successfully to %s", sub.Endpoint[:50])
	} else {
		log.Printf("[Push] Unexpected status %d for %s", resp.StatusCode, sub.Endpoint[:50])
	}
}

func getLevelEmoji(level models.LogLevel) string {
	switch level {
	case models.LogLevelDebug:
		return "🐛"
	case models.LogLevelInfo:
		return "ℹ️"
	case models.LogLevelWarn:
		return "⚠️"
	case models.LogLevelError:
		return "❌"
	case models.LogLevelCritical:
		return "💀"
	default:
		return "📋"
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
