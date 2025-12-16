package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// EnvMapping defines the mapping between environment variable names and config paths
type EnvMapping struct {
	EnvKey     string // Environment variable key (e.g., "SERVER_PORT")
	ConfigPath string // Path in config (e.g., "server.port")
	Type       string // Type: "string", "int", "bool"
}

// envMappings defines all supported environment variable mappings
// Supports both formats: SERVER_PORT and CL_SERVER_PORT
var envMappings = []EnvMapping{
	// Server Config
	{"SERVER_PORT", "server.port", "int"},
	{"SERVER_ENV", "server.env", "string"},

	// Database Config
	{"DATABASE_PATH", "database.path", "string"},

	// Redis Config
	{"REDIS_URL", "redis.url", "string"},

	// JWT Config
	{"JWT_SECRET", "jwt.secret", "string"},
	{"JWT_EXPIRY", "jwt.expiry", "string"},

	// VAPID Config
	{"VAPID_PUBLIC_KEY", "vapid.public_key", "string"},
	{"VAPID_PRIVATE_KEY", "vapid.private_key", "string"},
	{"VAPID_SUBJECT", "vapid.subject", "string"},

	// Telegram Config
	{"TELEGRAM_BOT_TOKEN", "telegram.bot_token", "string"},
	{"TELEGRAM_ENABLED", "telegram.enabled", "bool"},

	// Admin Config
	{"ADMIN_USERNAME", "admin.username", "string"},
	{"ADMIN_PASSWORD", "admin.password", "string"},

	// Rate Limit Config
	{"RATE_LIMIT_API_REQUESTS_PER_MINUTE", "rate_limit.api.requests_per_minute", "int"},
	{"RATE_LIMIT_TELEGRAM_MESSAGES_PER_MINUTE", "rate_limit.channels.telegram.messages_per_minute", "int"},
	{"RATE_LIMIT_DISCORD_MESSAGES_PER_MINUTE", "rate_limit.channels.discord.messages_per_minute", "int"},
	{"RATE_LIMIT_PUSH_MESSAGES_PER_MINUTE", "rate_limit.channels.push.messages_per_minute", "int"},

	// WebSocket Config
	{"WEBSOCKET_ENABLED", "websocket.enabled", "bool"},
	{"WEBSOCKET_PING_INTERVAL", "websocket.ping_interval", "string"},
	{"WEBSOCKET_PONG_TIMEOUT", "websocket.pong_timeout", "string"},
	{"WEBSOCKET_MAX_MESSAGE_SIZE", "websocket.max_message_size", "int"},
	{"WEBSOCKET_READ_BUFFER_SIZE", "websocket.read_buffer_size", "int"},
	{"WEBSOCKET_WRITE_BUFFER_SIZE", "websocket.write_buffer_size", "int"},

	// Retention Config
	{"RETENTION_ENABLED", "retention.enabled", "bool"},
	{"RETENTION_DEFAULT_MAX_AGE", "retention.default.max_age", "string"},
	{"RETENTION_DEFAULT_MAX_COUNT", "retention.default.max_count", "int"},
	{"RETENTION_CLEANUP_ENABLED", "retention.cleanup.enabled", "bool"},
	{"RETENTION_CLEANUP_SCHEDULE", "retention.cleanup.schedule", "string"},
	{"RETENTION_CLEANUP_BATCH_SIZE", "retention.cleanup.batch_size", "int"},
}

// getEnvValue gets environment variable value with fallback to CL_ prefix
func getEnvValue(key string) string {
	// Try without prefix first (e.g., SERVER_PORT)
	if v := os.Getenv(key); v != "" {
		return v
	}

	// Try with CL_ prefix (e.g., CL_SERVER_PORT)
	if v := os.Getenv("CL_" + key); v != "" {
		return v
	}

	return ""
}

// loadFromEnvNew loads config from environment variables using the mapping table
// This replaces the old manual loadFromEnv method
func (c *Config) loadFromEnvNew() {
	for _, mapping := range envMappings {
		value := getEnvValue(mapping.EnvKey)
		if value == "" {
			continue
		}

		// Apply the value based on the config path
		if err := c.setConfigValue(mapping.ConfigPath, value, mapping.Type); err != nil {
			// Log error but continue with other mappings
			fmt.Printf("Warning: Failed to set %s from env: %v\n", mapping.ConfigPath, err)
		}
	}
}

// setConfigValue sets a config value using dot-notation path
func (c *Config) setConfigValue(path, value, valueType string) error {
	parts := strings.Split(path, ".")

	// Handle based on path structure
	switch parts[0] {
	case "server":
		return c.setServerValue(parts[1:], value, valueType)
	case "database":
		return c.setDatabaseValue(parts[1:], value, valueType)
	case "redis":
		return c.setRedisValue(parts[1:], value, valueType)
	case "jwt":
		return c.setJWTValue(parts[1:], value, valueType)
	case "vapid":
		return c.setVAPIDValue(parts[1:], value, valueType)
	case "telegram":
		return c.setTelegramValue(parts[1:], value, valueType)
	case "admin":
		return c.setAdminValue(parts[1:], value, valueType)
	case "rate_limit":
		return c.setRateLimitValue(parts[1:], value, valueType)
	case "websocket":
		return c.setWebSocketValue(parts[1:], value, valueType)
	case "retention":
		return c.setRetentionValue(parts[1:], value, valueType)
	default:
		return fmt.Errorf("unknown config section: %s", parts[0])
	}
}

func (c *Config) setServerValue(path []string, value, valueType string) error {
	switch path[0] {
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.Server.Port = port
	case "env":
		c.Server.Env = value
	default:
		return fmt.Errorf("unknown server field: %s", path[0])
	}
	return nil
}

func (c *Config) setDatabaseValue(path []string, value, valueType string) error {
	switch path[0] {
	case "path":
		c.Database.Path = value
	default:
		return fmt.Errorf("unknown database field: %s", path[0])
	}
	return nil
}

func (c *Config) setRedisValue(path []string, value, valueType string) error {
	switch path[0] {
	case "url":
		c.Redis.URL = value
	default:
		return fmt.Errorf("unknown redis field: %s", path[0])
	}
	return nil
}

func (c *Config) setJWTValue(path []string, value, valueType string) error {
	switch path[0] {
	case "secret":
		c.JWT.Secret = value
	case "expiry":
		c.JWT.Expiry = value
	default:
		return fmt.Errorf("unknown jwt field: %s", path[0])
	}
	return nil
}

func (c *Config) setVAPIDValue(path []string, value, valueType string) error {
	switch path[0] {
	case "public_key":
		c.VAPID.PublicKey = value
	case "private_key":
		c.VAPID.PrivateKey = value
	case "subject":
		c.VAPID.Subject = value
	default:
		return fmt.Errorf("unknown vapid field: %s", path[0])
	}
	return nil
}

func (c *Config) setTelegramValue(path []string, value, valueType string) error {
	switch path[0] {
	case "bot_token":
		c.Telegram.BotToken = value
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.Telegram.Enabled = enabled
	default:
		return fmt.Errorf("unknown telegram field: %s", path[0])
	}
	return nil
}

func (c *Config) setAdminValue(path []string, value, valueType string) error {
	switch path[0] {
	case "username":
		c.Admin.Username = value
	case "password":
		c.Admin.Password = value
	default:
		return fmt.Errorf("unknown admin field: %s", path[0])
	}
	return nil
}

func (c *Config) setRateLimitValue(path []string, value, valueType string) error {
	if len(path) < 2 {
		return fmt.Errorf("invalid rate_limit path: %v", path)
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	switch path[0] {
	case "api":
		if path[1] == "requests_per_minute" {
			c.RateLimit.API.RequestsPerMinute = intVal
		}
	case "channels":
		if len(path) < 3 {
			return fmt.Errorf("invalid channels path: %v", path)
		}
		switch path[1] {
		case "telegram":
			c.RateLimit.Channels.Telegram.MessagesPerMinute = intVal
		case "discord":
			c.RateLimit.Channels.Discord.MessagesPerMinute = intVal
		case "push":
			c.RateLimit.Channels.Push.MessagesPerMinute = intVal
		}
	}
	return nil
}

func (c *Config) setWebSocketValue(path []string, value, valueType string) error {
	switch path[0] {
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.WebSocket.Enabled = enabled
	case "ping_interval":
		c.WebSocket.PingInterval = value
	case "pong_timeout":
		c.WebSocket.PongTimeout = value
	case "max_message_size":
		size, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.WebSocket.MaxMessageSize = size
	case "read_buffer_size":
		size, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.WebSocket.ReadBufferSize = size
	case "write_buffer_size":
		size, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.WebSocket.WriteBufferSize = size
	default:
		return fmt.Errorf("unknown websocket field: %s", path[0])
	}
	return nil
}

func (c *Config) setRetentionValue(path []string, value, valueType string) error {
	switch path[0] {
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.Retention.Enabled = enabled
	case "default":
		if len(path) < 2 {
			return fmt.Errorf("invalid retention.default path: %v", path)
		}
		switch path[1] {
		case "max_age":
			c.Retention.Default.MaxAge = value
		case "max_count":
			count, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			c.Retention.Default.MaxCount = count
		}
	case "cleanup":
		if len(path) < 2 {
			return fmt.Errorf("invalid retention.cleanup path: %v", path)
		}
		switch path[1] {
		case "enabled":
			enabled, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			c.Retention.Cleanup.Enabled = enabled
		case "schedule":
			c.Retention.Cleanup.Schedule = value
		case "batch_size":
			size, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			c.Retention.Cleanup.BatchSize = size
		}
	}
	return nil
}

// PrintEnvHelp prints all supported environment variables
func PrintEnvHelp() {
	fmt.Println("Supported Environment Variables:")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("All variables support both formats:")
	fmt.Println("  1. Direct:     VARIABLE_NAME (e.g., SERVER_PORT)")
	fmt.Println("  2. Prefixed:   CL_VARIABLE_NAME (e.g., CL_SERVER_PORT)")
	fmt.Println()

	currentSection := ""
	for _, mapping := range envMappings {
		section := strings.Split(mapping.ConfigPath, ".")[0]
		if section != currentSection {
			currentSection = section
			fmt.Printf("\n[%s]\n", strings.ToUpper(section))
		}
		fmt.Printf("  %-50s %s (type: %s)\n", mapping.EnvKey, mapping.ConfigPath, mapping.Type)
	}

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  export SERVER_PORT=8080")
	fmt.Println("  export CL_DATABASE_PATH=/var/lib/central-logs/db.sqlite")
	fmt.Println("  export VAPID_PUBLIC_KEY=\"your-public-key\"")
	fmt.Println()
}
