package config

import (
	"fmt"
	"time"

	"github.com/mwazovzky/config"
)

type Config struct {
	Telegram TelegramConfig
	OpenAI   OpenAIConfig
	Redis    RedisConfig
	Loki     LokiConfig
}

type TelegramConfig struct {
	BotName      string   `env:"TELEGRAM_BOT_NAME" required:"true"`
	ApiToken     string   `env:"TELEGRAM_API_TOKEN" required:"true"`
	Users        []string `env:"TELEGRAM_USER_CHATS" required:"true"`
	Chats        []int64  `env:"TELEGRAM_GROUP_CHATS" required:"true"`
	MessageLimit int      `env:"TELEGRAM_MESSAGE_LIMIT" required:"true"`
	ShowMore     bool     `env:"TELEGRAM_SHOW_MORE" default:"true"`
}

type OpenAIConfig struct {
	ApiKey         string        `env:"OPENAI_API_KEY" required:"true"`
	Model          string        `env:"OPENAI_MODEL" required:"true"`
	Name           string        `env:"OPENAI_ASSISTANT_NAME" required:"true"`
	Role           string        `env:"OPENAI_ASSISTANT_ROLE" required:"true"`
	RequestTimeout time.Duration `env:"OPENAI_REQUEST_TIMEOUT" default:"30"`
}

type RedisConfig struct {
	Host           string        `env:"REDIS_HOST" required:"true"`
	Port           string        `env:"REDIS_PORT" required:"true"`
	Password       string        `env:"REDIS_PASSWORD" required:"true"`
	ExpirationTime time.Duration `env:"REDIS_EXPIRATION_TIME" required:"true"`
}

type LokiConfig struct {
	Url      string `env:"LOKI_URL" required:"true"`
	Username string `env:"LOKI_USERNAME" required:"true"`
	Token    string `env:"LOKI_AUTH_TOKEN" required:"true"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := config.LoadConfig(&cfg.Telegram); err != nil {
		return nil, fmt.Errorf("loading telegram config: %w", err)
	}
	if err := config.LoadConfig(&cfg.OpenAI); err != nil {
		return nil, fmt.Errorf("loading openai config: %w", err)
	}
	if err := config.LoadConfig(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("loading redis config: %w", err)
	}
	if err := config.LoadConfig(&cfg.Loki); err != nil {
		return nil, fmt.Errorf("loading loki config: %w", err)
	}

	return cfg, nil
}
