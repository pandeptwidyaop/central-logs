package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	ChannelTypePush     ChannelType = "PUSH"
	ChannelTypeTelegram ChannelType = "TELEGRAM"
	ChannelTypeDiscord  ChannelType = "DISCORD"
)

type Channel struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"project_id"`
	Type      ChannelType            `json:"type"`
	Name      string                 `json:"name"`
	Config    map[string]interface{} `json:"config"`
	MinLevel  LogLevel               `json:"min_level"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
}

type ChannelRepository struct {
	db *sql.DB
}

func NewChannelRepository(db *sql.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(channel *Channel) error {
	channel.ID = uuid.New().String()
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(channel.Config)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO channels (id, project_id, type, name, config, min_level, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, channel.ID, channel.ProjectID, channel.Type, channel.Name, string(configJSON), channel.MinLevel, channel.IsActive, channel.CreatedAt, channel.UpdatedAt)

	return err
}

func (r *ChannelRepository) GetByID(id string) (*Channel, error) {
	channel := &Channel{}
	var configJSON string

	err := r.db.QueryRow(`
		SELECT id, project_id, type, name, config, min_level, is_active, created_at, updated_at
		FROM channels WHERE id = ?
	`, id).Scan(&channel.ID, &channel.ProjectID, &channel.Type, &channel.Name, &configJSON, &channel.MinLevel, &channel.IsActive, &channel.CreatedAt, &channel.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(configJSON), &channel.Config); err != nil {
		return nil, err
	}

	return channel, nil
}

func (r *ChannelRepository) GetByProjectID(projectID string) ([]*Channel, error) {
	rows, err := r.db.Query(`
		SELECT id, project_id, type, name, config, min_level, is_active, created_at, updated_at
		FROM channels WHERE project_id = ?
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		var configJSON string

		if err := rows.Scan(&channel.ID, &channel.ProjectID, &channel.Type, &channel.Name, &configJSON, &channel.MinLevel, &channel.IsActive, &channel.CreatedAt, &channel.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(configJSON), &channel.Config); err != nil {
			return nil, err
		}

		channels = append(channels, channel)
	}
	return channels, nil
}

func (r *ChannelRepository) GetActiveByProjectID(projectID string) ([]*Channel, error) {
	rows, err := r.db.Query(`
		SELECT id, project_id, type, name, config, min_level, is_active, created_at, updated_at
		FROM channels WHERE project_id = ? AND is_active = 1
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		var configJSON string

		if err := rows.Scan(&channel.ID, &channel.ProjectID, &channel.Type, &channel.Name, &configJSON, &channel.MinLevel, &channel.IsActive, &channel.CreatedAt, &channel.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(configJSON), &channel.Config); err != nil {
			return nil, err
		}

		channels = append(channels, channel)
	}
	return channels, nil
}

func (r *ChannelRepository) Update(channel *Channel) error {
	channel.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(channel.Config)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		UPDATE channels SET name = ?, config = ?, min_level = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`, channel.Name, string(configJSON), channel.MinLevel, channel.IsActive, channel.UpdatedAt, channel.ID)
	return err
}

func (r *ChannelRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM channels WHERE id = ?`, id)
	return err
}

func (c *Channel) ShouldNotify(level LogLevel) bool {
	return level.Priority() >= c.MinLevel.Priority()
}

func (c *Channel) GetTelegramConfig() (*TelegramConfig, error) {
	data, err := json.Marshal(c.Config)
	if err != nil {
		return nil, err
	}
	var config TelegramConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Channel) GetDiscordConfig() (*DiscordConfig, error) {
	data, err := json.Marshal(c.Config)
	if err != nil {
		return nil, err
	}
	var config DiscordConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
