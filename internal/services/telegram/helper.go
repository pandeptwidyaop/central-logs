package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelegramUpdate represents an update from Telegram API
type TelegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		MessageID int    `json:"message_id"`
		From      *User  `json:"from"`
		Chat      *Chat  `json:"chat"`
		Date      int64  `json:"date"`
		Text      string `json:"text"`
	} `json:"message,omitempty"`
}

// User represents a Telegram user
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"` // "private", "group", "supergroup", "channel"
	Title    string `json:"title,omitempty"`
	Username string `json:"username,omitempty"`
	// For private chats
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// ChatInfo is a simplified chat info for the frontend
type ChatInfo struct {
	ChatID      string `json:"chat_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Username    string `json:"username,omitempty"`
	LastMessage string `json:"last_message"`
	LastDate    string `json:"last_date"`
}

// Helper provides helper functions for Telegram integration
type Helper struct {
	client *http.Client
}

// NewHelper creates a new Telegram helper
func NewHelper() *Helper {
	return &Helper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetRecentChats fetches recent chats that interacted with the bot
func (h *Helper) GetRecentChats(botToken string) ([]ChatInfo, error) {
	// Validate bot token format (basic check)
	if len(botToken) < 20 {
		return nil, fmt.Errorf("invalid bot token format")
	}

	// Get updates from Telegram
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", botToken)
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Telegram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Ok          bool             `json:"ok"`
		Description string           `json:"description,omitempty"`
		Result      []TelegramUpdate `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse Telegram response: %w", err)
	}

	if !result.Ok {
		return nil, fmt.Errorf("telegram API error: %s", result.Description)
	}

	// Extract unique chats
	chatsMap := make(map[int64]*ChatInfo)
	for _, update := range result.Result {
		if update.Message == nil || update.Message.Chat == nil {
			continue
		}

		chat := update.Message.Chat
		chatID := chat.ID

		// Build chat name
		name := chat.Title
		if name == "" {
			// Private chat
			name = chat.FirstName
			if chat.LastName != "" {
				name += " " + chat.LastName
			}
		}

		// Only keep the most recent message per chat
		if existing, ok := chatsMap[chatID]; ok {
			// Update if this message is newer
			if update.Message.Date > parseRFC3339(existing.LastDate).Unix() {
				existing.LastMessage = truncateString(update.Message.Text, 100)
				existing.LastDate = time.Unix(update.Message.Date, 0).Format(time.RFC3339)
			}
		} else {
			// New chat
			chatsMap[chatID] = &ChatInfo{
				ChatID:      fmt.Sprintf("%d", chatID),
				Type:        chat.Type,
				Name:        name,
				Username:    chat.Username,
				LastMessage: truncateString(update.Message.Text, 100),
				LastDate:    time.Unix(update.Message.Date, 0).Format(time.RFC3339),
			}
		}
	}

	// Convert map to slice
	chats := make([]ChatInfo, 0, len(chatsMap))
	for _, chat := range chatsMap {
		chats = append(chats, *chat)
	}

	// Sort by last date (newest first)
	sortChatsByDate(chats)

	return chats, nil
}

// TestBotToken validates a bot token by calling getMe
func (h *Helper) TestBotToken(botToken string) (bool, string, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)
	resp, err := h.client.Get(url)
	if err != nil {
		return false, "", fmt.Errorf("failed to connect to Telegram API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description,omitempty"`
		Result      *struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"result,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Ok {
		return false, "", fmt.Errorf("invalid bot token: %s", result.Description)
	}

	if result.Result == nil {
		return false, "", fmt.Errorf("invalid response from Telegram")
	}

	botName := result.Result.Username
	if botName == "" {
		botName = result.Result.FirstName
	}

	return true, botName, nil
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func parseRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func sortChatsByDate(chats []ChatInfo) {
	// Simple bubble sort (good enough for small lists)
	n := len(chats)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			date1 := parseRFC3339(chats[j].LastDate)
			date2 := parseRFC3339(chats[j+1].LastDate)
			if date1.Before(date2) {
				chats[j], chats[j+1] = chats[j+1], chats[j]
			}
		}
	}
}
