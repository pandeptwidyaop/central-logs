package config

import (
	"os"
	"testing"
)

func TestEnvOverride(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envKeys := []string{
		"SERVER_PORT", "CL_SERVER_PORT",
		"SERVER_ENV", "CL_SERVER_ENV",
		"DATABASE_PATH", "CL_DATABASE_PATH",
		"REDIS_URL", "CL_REDIS_URL",
		"JWT_SECRET", "CL_JWT_SECRET",
		"VAPID_PUBLIC_KEY", "CL_VAPID_PUBLIC_KEY",
	}
	for _, key := range envKeys {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		// Restore original env vars
		for key, val := range originalEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	tests := []struct {
		name     string
		envKey   string
		envValue string
		check    func(*Config) bool
	}{
		{
			name:     "SERVER_PORT without prefix",
			envKey:   "SERVER_PORT",
			envValue: "8080",
			check:    func(c *Config) bool { return c.Server.Port == 8080 },
		},
		{
			name:     "SERVER_PORT with CL_ prefix",
			envKey:   "CL_SERVER_PORT",
			envValue: "9000",
			check:    func(c *Config) bool { return c.Server.Port == 9000 },
		},
		{
			name:     "SERVER_ENV without prefix",
			envKey:   "SERVER_ENV",
			envValue: "production",
			check:    func(c *Config) bool { return c.Server.Env == "production" },
		},
		{
			name:     "DATABASE_PATH without prefix",
			envKey:   "DATABASE_PATH",
			envValue: "/custom/path/db.sqlite",
			check:    func(c *Config) bool { return c.Database.Path == "/custom/path/db.sqlite" },
		},
		{
			name:     "REDIS_URL with CL_ prefix",
			envKey:   "CL_REDIS_URL",
			envValue: "redis://custom:6379",
			check:    func(c *Config) bool { return c.Redis.URL == "redis://custom:6379" },
		},
		{
			name:     "JWT_SECRET without prefix",
			envKey:   "JWT_SECRET",
			envValue: "super-secret-key",
			check:    func(c *Config) bool { return c.JWT.Secret == "super-secret-key" },
		},
		{
			name:     "VAPID_PUBLIC_KEY without prefix",
			envKey:   "VAPID_PUBLIC_KEY",
			envValue: "test-public-key",
			check:    func(c *Config) bool { return c.VAPID.PublicKey == "test-public-key" },
		},
		{
			name:     "WEBSOCKET_ENABLED bool",
			envKey:   "WEBSOCKET_ENABLED",
			envValue: "false",
			check:    func(c *Config) bool { return c.WebSocket.Enabled == false },
		},
		{
			name:     "WEBSOCKET_MAX_MESSAGE_SIZE int",
			envKey:   "WEBSOCKET_MAX_MESSAGE_SIZE",
			envValue: "1024",
			check:    func(c *Config) bool { return c.WebSocket.MaxMessageSize == 1024 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set the test env var
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			// Create config and load from env
			cfg := DefaultConfig()
			cfg.loadFromEnvNew()

			// Check the result
			if !tt.check(cfg) {
				t.Errorf("Environment variable %s=%s was not applied correctly", tt.envKey, tt.envValue)
			}
		})
	}
}

func TestEnvPriority(t *testing.T) {
	// Test that non-prefixed env vars take priority over prefixed ones
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("CL_SERVER_PORT", "9000")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("CL_SERVER_PORT")
	}()

	cfg := DefaultConfig()
	cfg.loadFromEnvNew()

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected non-prefixed SERVER_PORT (8080) to take priority, got %d", cfg.Server.Port)
	}
}

func TestGetEnvValue(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setValue       string
		setPrefixValue string
		expected       string
	}{
		{
			name:     "only non-prefixed set",
			key:      "TEST_KEY",
			setValue: "value1",
			expected: "value1",
		},
		{
			name:           "only prefixed set",
			key:            "TEST_KEY",
			setPrefixValue: "value2",
			expected:       "value2",
		},
		{
			name:           "both set, non-prefixed takes priority",
			key:            "TEST_KEY",
			setValue:       "value1",
			setPrefixValue: "value2",
			expected:       "value1",
		},
		{
			name:     "neither set",
			key:      "TEST_KEY",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.Unsetenv(tt.key)
			os.Unsetenv("CL_" + tt.key)
			defer func() {
				os.Unsetenv(tt.key)
				os.Unsetenv("CL_" + tt.key)
			}()

			// Set values
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}
			if tt.setPrefixValue != "" {
				os.Setenv("CL_"+tt.key, tt.setPrefixValue)
			}

			// Test
			result := getEnvValue(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestComplexNestedPaths(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		check    func(*Config) bool
	}{
		{
			name:     "Rate limit API",
			envKey:   "RATE_LIMIT_API_REQUESTS_PER_MINUTE",
			envValue: "5000",
			check:    func(c *Config) bool { return c.RateLimit.API.RequestsPerMinute == 5000 },
		},
		{
			name:     "Rate limit Telegram",
			envKey:   "RATE_LIMIT_TELEGRAM_MESSAGES_PER_MINUTE",
			envValue: "100",
			check:    func(c *Config) bool { return c.RateLimit.Channels.Telegram.MessagesPerMinute == 100 },
		},
		{
			name:     "Retention default max age",
			envKey:   "RETENTION_DEFAULT_MAX_AGE",
			envValue: "60d",
			check:    func(c *Config) bool { return c.Retention.Default.MaxAge == "60d" },
		},
		{
			name:     "Retention cleanup enabled",
			envKey:   "RETENTION_CLEANUP_ENABLED",
			envValue: "false",
			check:    func(c *Config) bool { return c.Retention.Cleanup.Enabled == false },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			cfg := DefaultConfig()
			cfg.loadFromEnvNew()

			if !tt.check(cfg) {
				t.Errorf("Environment variable %s=%s was not applied correctly", tt.envKey, tt.envValue)
			}
		})
	}
}
