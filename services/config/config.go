package config

import (
	"telegrambot-assistant/services/configloader"
	"time"
)

type Config struct {
	Telegram TelegramConfig
	OpenAI   OpenAIConfig
	Redis    RedisConfig
}

type TelegramConfig struct {
	BotName       string  `env:"TELEGRAM_BOT_NAME" required:"true"`
	ApiToken      string  `env:"TELEGRAM_API_TOKEN" required:"true"`
	ChatID        int64   `env:"TELEGRAM_CHAT_ID" required:"true"`
	AssignedChats []int64 `env:"TELEGRAM_ASSIGNED_CHATS" required:"true"`
	MessageLimit  int     `env:"TELEGRAM_MESSAGE_LIMIT" required:"true"`
}

type OpenAIConfig struct {
	ApiUrl string `env:"OPENAI_API_URL" required:"true"`
	ApiKey string `env:"OPENAI_API_KEY" required:"true"`
	Model  string `env:"OPENAI_MODEL" required:"true"`
	Name   string `env:"OPENAI_ASSISTANT_NAME" required:"true"`
	Role   string `env:"OPENAI_ASSISTANT_ROLE" required:"true"`
}

type RedisConfig struct {
	Host           string        `env:"REDIS_HOST" required:"true"`
	Port           string        `env:"REDIS_PORT" required:"true"`
	Password       string        `env:"REDIS_PASSWORD" required:"true"`
	ExpirationTime time.Duration `env:"REDIS_EXPIRATION_TIME" required:"true"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := configloader.LoadConfig(&cfg.Telegram); err != nil {
		return nil, err
	}
	if err := configloader.LoadConfig(&cfg.OpenAI); err != nil {
		return nil, err
	}
	if err := configloader.LoadConfig(&cfg.Redis); err != nil {
		return nil, err
	}
	return cfg, nil
}
