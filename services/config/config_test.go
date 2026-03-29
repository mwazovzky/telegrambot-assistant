package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func clearTestEnv() {
	os.Unsetenv("TELEGRAM_BOT_NAME")
	os.Unsetenv("TELEGRAM_API_TOKEN")
	os.Unsetenv("TELEGRAM_USER_CHATS")
	os.Unsetenv("TELEGRAM_GROUP_CHATS")
	os.Unsetenv("TELEGRAM_MESSAGE_LIMIT")
	os.Unsetenv("TELEGRAM_SHOW_MORE")

	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_MODEL")
	os.Unsetenv("OPENAI_ASSISTANT_NAME")
	os.Unsetenv("OPENAI_ASSISTANT_ROLE")

	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_EXPIRATION_TIME")

	os.Unsetenv("LOKI_URL")
	os.Unsetenv("LOKI_USERNAME")
	os.Unsetenv("LOKI_AUTH_TOKEN")
}

func setupTestEnv() {
	clearTestEnv()

	os.Setenv("TELEGRAM_BOT_NAME", "test_bot")
	os.Setenv("TELEGRAM_API_TOKEN", "test_token")
	os.Setenv("TELEGRAM_USER_CHATS", "user1,user2,user3")
	os.Setenv("TELEGRAM_GROUP_CHATS", "12345,67890")
	os.Setenv("TELEGRAM_MESSAGE_LIMIT", "4096")
	os.Setenv("TELEGRAM_SHOW_MORE", "true")

	os.Setenv("OPENAI_API_KEY", "test_api_key")
	os.Setenv("OPENAI_MODEL", "test_model")
	os.Setenv("OPENAI_ASSISTANT_NAME", "test_assistant")
	os.Setenv("OPENAI_ASSISTANT_ROLE", "test_role")

	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "test_password")
	os.Setenv("REDIS_EXPIRATION_TIME", "60s")

	os.Setenv("LOKI_URL", "http://localhost:3100")
	os.Setenv("LOKI_USERNAME", "test_user")
	os.Setenv("LOKI_AUTH_TOKEN", "test_token")
}

func TestLoad(t *testing.T) {
	// Clean environment before test
	clearTestEnv()

	// Set environment variables for testing
	env := map[string]string{
		"TELEGRAM_BOT_NAME":      "test_bot",
		"TELEGRAM_API_TOKEN":     "test_token",
		"TELEGRAM_USER_CHATS":    "user1,user2,user3",
		"TELEGRAM_GROUP_CHATS":   "12345,67890",
		"TELEGRAM_MESSAGE_LIMIT": "4096",
		"OPENAI_API_KEY":         "test_api_key",
		"OPENAI_MODEL":           "test_model",
		"OPENAI_ASSISTANT_NAME":  "test_assistant",
		"OPENAI_ASSISTANT_ROLE":  "test_role",
		"REDIS_HOST":             "localhost",
		"REDIS_PORT":             "6379",
		"REDIS_PASSWORD":         "test_password",
		"REDIS_EXPIRATION_TIME":  "60s",
		"LOKI_URL":               "http://localhost:3100",
		"LOKI_USERNAME":          "test_user",
		"LOKI_AUTH_TOKEN":        "test_token",
	}

	for k, v := range env {
		os.Setenv(k, v)
	}
	defer clearTestEnv()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify configuration values
	assert.Equal(t, "test_bot", cfg.Telegram.BotName)
	assert.Equal(t, "test_token", cfg.Telegram.ApiToken)
	assert.ElementsMatch(t, []string{"user1", "user2", "user3"}, cfg.Telegram.Users)
	assert.Equal(t, []int64{12345, 67890}, cfg.Telegram.Chats)
	assert.Equal(t, 4096, cfg.Telegram.MessageLimit)

	assert.Equal(t, "test_api_key", cfg.OpenAI.ApiKey)
	assert.Equal(t, "test_model", cfg.OpenAI.Model)
	assert.Equal(t, "test_assistant", cfg.OpenAI.Name)
	assert.Equal(t, "test_role", cfg.OpenAI.Role)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, "test_password", cfg.Redis.Password)
	assert.Equal(t, 60*time.Second, cfg.Redis.ExpirationTime)
}

func TestLoadMissingRequired(t *testing.T) {
	clearTestEnv()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadInvalidValues(t *testing.T) {
	setupTestEnv()
	defer clearTestEnv()

	os.Setenv("TELEGRAM_USER_CHATS", "")         // Test empty userChats list
	os.Setenv("TELEGRAM_GROUP_CHATS", "invalid") // Test invalid groupChats

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestTelegramShowMoreDefault(t *testing.T) {
	// Clean environment before test
	clearTestEnv()

	// Set other required environment variables without TELEGRAM_SHOW_MORE
	setupTestEnv()
	os.Unsetenv("TELEGRAM_SHOW_MORE") // Explicitly unset to test default value

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// The default value is true as specified in the struct tag `default:"true"`
	assert.Equal(t, true, cfg.Telegram.ShowMore)
}

func TestTelegramShowMoreExplicit(t *testing.T) {
	// Clean environment before test
	clearTestEnv()

	// Set all required environment variables
	setupTestEnv()

	// Test with explicit false value
	os.Setenv("TELEGRAM_SHOW_MORE", "false")
	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, false, cfg.Telegram.ShowMore)

	// Test with explicit true value
	os.Setenv("TELEGRAM_SHOW_MORE", "true")
	cfg, err = Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, true, cfg.Telegram.ShowMore)
}
