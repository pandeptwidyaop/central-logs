package worker

import (
	"bytes"
	"central-logs/internal/config"
	"central-logs/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Notifier handles sending notifications to various channels
type Notifier struct {
	channelRepo *models.ChannelRepository
	client      *http.Client
	config      *config.Config
}

// NewNotifier creates a new notification worker
func NewNotifier(channelRepo *models.ChannelRepository, cfg *config.Config) *Notifier {
	return &Notifier{
		channelRepo: channelRepo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		config: cfg,
	}
}

// ProcessLog processes a log entry and sends notifications if needed
func (n *Notifier) ProcessLog(logEntry *models.Log) {
	// Get all active channels for this project
	channels, err := n.channelRepo.GetByProjectID(logEntry.ProjectID)
	if err != nil {
		log.Printf("Failed to get channels for project %s: %v", logEntry.ProjectID, err)
		return
	}

	// Process each active channel
	for _, channel := range channels {
		if !channel.IsActive {
			continue
		}

		// Check if log level meets minimum level
		if !meetsMinLevel(logEntry.Level, channel.MinLevel) {
			continue
		}

		// Send notification based on channel type
		switch channel.Type {
		case models.ChannelTypeTelegram:
			go n.sendTelegram(channel, logEntry)
		case models.ChannelTypeDiscord:
			go n.sendDiscord(channel, logEntry)
		case models.ChannelTypePush:
			go n.sendPush(channel, logEntry)
		}
	}
}

// sendTelegram sends a notification to Telegram
func (n *Notifier) sendTelegram(channel *models.Channel, logEntry *models.Log) {
	// Get bot token - use channel's token or fallback to global config
	botToken, ok := channel.Config["bot_token"].(string)
	if !ok || botToken == "" {
		// Use global bot token from config
		botToken = n.config.Telegram.BotToken
		if botToken == "" {
			log.Printf("No bot_token configured for channel %s and no global bot_token in config", channel.ID)
			return
		}
	}

	chatID, ok := channel.Config["chat_id"].(string)
	if !ok || chatID == "" {
		log.Printf("Invalid chat_id for channel %s", channel.ID)
		return
	}

	// Format message with emoji based on level
	emoji := getLogEmoji(logEntry.Level)
	message := fmt.Sprintf("%s *%s*\n\n*Message:* %s\n*Source:* %s\n*Time:* %s",
		emoji,
		logEntry.Level,
		escapeMarkdown(logEntry.Message),
		escapeMarkdown(logEntry.Source),
		logEntry.Timestamp.Format("2006-01-02 15:04:05"),
	)

	// Add metadata if present
	if len(logEntry.Metadata) > 0 {
		metadataStr, _ := json.MarshalIndent(logEntry.Metadata, "", "  ")
		message += fmt.Sprintf("\n\n*Metadata:*\n```\n%s\n```", string(metadataStr))
	}

	// Prepare Telegram API request
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal Telegram payload: %v", err)
		return
	}

	resp, err := n.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send Telegram notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram API returned status %d for channel %s", resp.StatusCode, channel.ID)
		return
	}

	log.Printf("Sent Telegram notification for log %s to channel %s", logEntry.ID, channel.Name)
}

// sendDiscord sends a notification to Discord (placeholder)
func (n *Notifier) sendDiscord(channel *models.Channel, logEntry *models.Log) {
	// TODO: Implement Discord webhook notification
	log.Printf("Discord notifications not yet implemented")
}

// sendPush sends a push notification (placeholder)
func (n *Notifier) sendPush(channel *models.Channel, logEntry *models.Log) {
	// TODO: Implement Web Push notification
	log.Printf("Push notifications not yet implemented")
}

// Helper functions

func meetsMinLevel(logLevel, minLevel models.LogLevel) bool {
	levels := map[models.LogLevel]int{
		models.LogLevelDebug:    0,
		models.LogLevelInfo:     1,
		models.LogLevelWarn:     2,
		models.LogLevelError:    3,
		models.LogLevelCritical: 4,
	}

	logLevelInt, ok1 := levels[logLevel]
	minLevelInt, ok2 := levels[minLevel]

	if !ok1 || !ok2 {
		return false
	}

	return logLevelInt >= minLevelInt
}

func getLogEmoji(level models.LogLevel) string {
	switch level {
	case models.LogLevelDebug:
		return "🔍"
	case models.LogLevelInfo:
		return "ℹ️"
	case models.LogLevelWarn:
		return "⚠️"
	case models.LogLevelError:
		return "❌"
	case models.LogLevelCritical:
		return "🚨"
	default:
		return "📝"
	}
}

func escapeMarkdown(s string) string {
	// Escape special Markdown characters for Telegram
	replacer := []string{
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	}

	for i := 0; i < len(replacer); i += 2 {
		s = replaceAll(s, replacer[i], replacer[i+1])
	}

	return s
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}
