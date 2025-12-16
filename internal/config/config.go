package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	VAPID     VAPIDConfig     `yaml:"vapid"`
	Telegram  TelegramConfig  `yaml:"telegram"`
	Admin     AdminConfig     `yaml:"admin"`
	Retention RetentionConfig `yaml:"retention"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	WebSocket WebSocketConfig `yaml:"websocket"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Env  string `yaml:"env"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type RedisConfig struct {
	URL string `yaml:"url"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expiry string `yaml:"expiry"`
}

type VAPIDConfig struct {
	PublicKey  string `yaml:"public_key"`
	PrivateKey string `yaml:"private_key"`
	Subject    string `yaml:"subject"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	Enabled  bool   `yaml:"enabled"`
}

type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type RetentionConfig struct {
	Enabled             bool                      `yaml:"enabled"`
	Default             RetentionPolicy           `yaml:"default"`
	Levels              map[string]RetentionPolicy `yaml:"levels"`
	Cleanup             CleanupConfig             `yaml:"cleanup"`
	NotificationHistory RetentionPolicy           `yaml:"notification_history"`
}

type RetentionPolicy struct {
	MaxAge   string `yaml:"max_age"`
	MaxCount int    `yaml:"max_count"`
}

type CleanupConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Schedule  string `yaml:"schedule"`
	BatchSize int    `yaml:"batch_size"`
}

type RateLimitConfig struct {
	API      APIRateLimit      `yaml:"api"`
	Channels ChannelRateLimit  `yaml:"channels"`
}

type APIRateLimit struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
}

type ChannelRateLimit struct {
	Telegram ChannelLimit `yaml:"telegram"`
	Discord  ChannelLimit `yaml:"discord"`
	Push     ChannelLimit `yaml:"push"`
}

type ChannelLimit struct {
	MessagesPerMinute int `yaml:"messages_per_minute"`
}

type WebSocketConfig struct {
	Enabled         bool   `yaml:"enabled"`
	PingInterval    string `yaml:"ping_interval"`
	PongTimeout     string `yaml:"pong_timeout"`
	MaxMessageSize  int    `yaml:"max_message_size"`
	ReadBufferSize  int    `yaml:"read_buffer_size"`
	WriteBufferSize int    `yaml:"write_buffer_size"`
}

func (c *Config) GetJWTExpiry() time.Duration {
	d, err := time.ParseDuration(c.JWT.Expiry)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func (c *Config) GetWebSocketPingInterval() time.Duration {
	d, err := time.ParseDuration(c.WebSocket.PingInterval)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (c *Config) GetWebSocketPongTimeout() time.Duration {
	d, err := time.ParseDuration(c.WebSocket.PongTimeout)
	if err != nil {
		return 10 * time.Second
	}
	return d
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 3000,
			Env:  "development",
		},
		Database: DatabaseConfig{
			Path: "./data/central-logs.db",
		},
		Redis: RedisConfig{
			URL: "redis://localhost:6379",
		},
		JWT: JWTConfig{
			Secret: "change-this-secret-key",
			Expiry: "24h",
		},
		VAPID: VAPIDConfig{
			Subject: "mailto:admin@example.com",
		},
		Telegram: TelegramConfig{
			Enabled: false,
		},
		Admin: AdminConfig{
			Username: "admin",
			Password: "changeme123",
		},
		Retention: RetentionConfig{
			Enabled: true,
			Default: RetentionPolicy{
				MaxAge:   "30d",
				MaxCount: 100000,
			},
			Levels: map[string]RetentionPolicy{
				"debug":    {MaxAge: "7d", MaxCount: 10000},
				"info":     {MaxAge: "14d", MaxCount: 50000},
				"warn":     {MaxAge: "30d", MaxCount: 50000},
				"error":    {MaxAge: "90d", MaxCount: 100000},
				"critical": {MaxAge: "365d", MaxCount: 0},
			},
			Cleanup: CleanupConfig{
				Enabled:   true,
				Schedule:  "0 2 * * *",
				BatchSize: 1000,
			},
			NotificationHistory: RetentionPolicy{
				MaxAge: "7d",
			},
		},
		RateLimit: RateLimitConfig{
			API: APIRateLimit{
				RequestsPerMinute: 1000,
			},
			Channels: ChannelRateLimit{
				Telegram: ChannelLimit{MessagesPerMinute: 20},
				Discord:  ChannelLimit{MessagesPerMinute: 30},
				Push:     ChannelLimit{MessagesPerMinute: 60},
			},
		},
		WebSocket: WebSocketConfig{
			Enabled:         true,
			PingInterval:    "30s",
			PongTimeout:     "10s",
			MaxMessageSize:  512,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config file
			if err := Save(path, cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Override with environment variables (supports both SERVER_PORT and CL_SERVER_PORT formats)
	cfg.loadFromEnvNew()

	return cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(".", 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// loadFromEnv is deprecated, use loadFromEnvNew instead
// Kept for backward compatibility
func (c *Config) loadFromEnv() {
	c.loadFromEnvNew()
}
