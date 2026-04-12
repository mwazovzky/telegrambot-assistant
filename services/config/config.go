package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Telegram TelegramConfig
	OpenAI   OpenAIConfig
	Redis    RedisConfig
}

type TelegramConfig struct {
	BotName      string   `env:"TELEGRAM_BOT_NAME,required"`
	ApiToken     string   `env:"TELEGRAM_API_TOKEN,required"`
	Users        []string `env:"TELEGRAM_USER_CHATS,required"`
	Chats        []int64  `env:"TELEGRAM_GROUP_CHATS,required"`
	MessageLimit int      `env:"TELEGRAM_MESSAGE_LIMIT,required"`
	ShowMore     bool     `env:"TELEGRAM_SHOW_MORE" envDefault:"true"`
}

type OpenAIConfig struct {
	ApiKey         string        `env:"OPENAI_API_KEY,required"`
	Model          string        `env:"OPENAI_MODEL,required"`
	Name           string        `env:"OPENAI_ASSISTANT_NAME,required"`
	Role           string        `env:"OPENAI_ASSISTANT_ROLE,required"`
	RequestTimeout time.Duration `env:"OPENAI_REQUEST_TIMEOUT" envDefault:"30s"`
}

type RedisConfig struct {
	Host           string        `env:"REDIS_HOST,required"`
	Port           string        `env:"REDIS_PORT,required"`
	Password       string        `env:"REDIS_PASSWORD,required"`
	ExpirationTime time.Duration `env:"REDIS_EXPIRATION_TIME,required"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
