package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("TELEGRAM_BOT_NAME", "test_bot")
	os.Setenv("TELEGRAM_API_TOKEN", "test_token")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	os.Setenv("TELEGRAM_ASSIGNED_CHATS", "12345,67890")
	os.Setenv("TELEGRAM_MESSAGE_LIMIT", "4096")

	os.Setenv("OPENAI_API_URL", "https://api.openai.com")
	os.Setenv("OPENAI_API_KEY", "test_api_key")
	os.Setenv("OPENAI_MODEL", "test_model")
	os.Setenv("OPENAI_ASSISTANT_NAME", "test_assistant")
	os.Setenv("OPENAI_ASSISTANT_ROLE", "test_role")

	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "test_password")
	os.Setenv("REDIS_EXPIRATION_TIME", "60")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "test_bot", cfg.Telegram.BotName)
	assert.Equal(t, "test_token", cfg.Telegram.ApiToken)
	assert.Equal(t, int64(12345), cfg.Telegram.ChatID)
	assert.Equal(t, []int64{12345, 67890}, cfg.Telegram.AssignedChats)
	assert.Equal(t, 4096, cfg.Telegram.MessageLimit)

	assert.Equal(t, "https://api.openai.com", cfg.OpenAI.ApiUrl)
	assert.Equal(t, "test_api_key", cfg.OpenAI.ApiKey)
	assert.Equal(t, "test_model", cfg.OpenAI.Model)
	assert.Equal(t, "test_assistant", cfg.OpenAI.Name)
	assert.Equal(t, "test_role", cfg.OpenAI.Role)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, "test_password", cfg.Redis.Password)
	assert.Equal(t, time.Duration(60)*time.Second, cfg.Redis.ExpirationTime)
}

func TestLoadMissingRequired(t *testing.T) {
	// Unset environment variables to test missing required fields
	os.Unsetenv("TELEGRAM_BOT_NAME")
	os.Unsetenv("TELEGRAM_API_TOKEN")
	os.Unsetenv("TELEGRAM_CHAT_ID")
	os.Unsetenv("TELEGRAM_ASSIGNED_CHATS")
	os.Unsetenv("TELEGRAM_MESSAGE_LIMIT")

	os.Unsetenv("OPENAI_API_URL")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_MODEL")
	os.Unsetenv("OPENAI_ASSISTANT_NAME")
	os.Unsetenv("OPENAI_ASSISTANT_ROLE")

	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_EXPIRATION_TIME")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadInvalidValues(t *testing.T) {
	// Set invalid environment variables for testing
	os.Setenv("TELEGRAM_BOT_NAME", "test_bot")
	os.Setenv("TELEGRAM_API_TOKEN", "test_token")
	os.Setenv("TELEGRAM_CHAT_ID", "invalid")
	os.Setenv("TELEGRAM_ASSIGNED_CHATS", "invalid")
	os.Setenv("TELEGRAM_MESSAGE_LIMIT", "invalid")

	os.Setenv("OPENAI_API_URL", "https://api.openai.com")
	os.Setenv("OPENAI_API_KEY", "test_api_key")
	os.Setenv("OPENAI_MODEL", "test_model")
	os.Setenv("OPENAI_ASSISTANT_NAME", "test_assistant")
	os.Setenv("OPENAI_ASSISTANT_ROLE", "test_role")

	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "test_password")
	os.Setenv("REDIS_EXPIRATION_TIME", "invalid")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
