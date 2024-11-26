package config

import "os"

type Config struct {
	Telegram TelegramConfig
	OpenAI   OpenAIConfig
	Redis    RedisConfig
}

type TelegramConfig struct {
	BotName       string `env:"TELEGRAM_BOT_NAME"`
	ApiToken      string `env:"TELEGRAM_API_TOKEN"`
	ChatID        string `env:"TELEGRAM_CHAT_ID"`
	AssignedChats string `env:"TELEGRAM_ASSIGNED_CHATS"`
}

type OpenAIConfig struct {
	ApiUrl string `env:"OPENAI_API_URL"`
	ApiKey string `env:"OPENAI_API_KEY"`
	Model  string `env:"OPENAI_MODEL"`
	Name   string `env:"OPENAI_ASSISTANT_NAME"`
}

type RedisConfig struct {
	Host           string `env:"REDIS_HOST"`
	Port           string `env:"REDIS_PORT"`
	Password       string `env:"REDIS_PASSWORD"`
	ExpirationTime string `env:"REDIS_EXPIRATION_TIME"`
}

func Load() *Config {
	telegramConfig := TelegramConfig{
		BotName:       os.Getenv("TELEGRAM_BOT_NAME"),
		ApiToken:      os.Getenv("TELEGRAM_API_TOKEN"),
		ChatID:        os.Getenv("TELEGRAM_CHAT_ID"),
		AssignedChats: os.Getenv("TELEGRAM_ASSIGNED_CHATS"),
	}

	openAIConfig := OpenAIConfig{
		ApiUrl: os.Getenv("OPENAI_API_URL"),
		ApiKey: os.Getenv("OPENAI_API_KEY"),
		Model:  os.Getenv("OPENAI_MODEL"),
		Name:   os.Getenv("OPENAI_ASSISTANT_NAME"),
	}

	redisConfig := RedisConfig{
		Host:           os.Getenv("REDIS_HOST"),
		Port:           os.Getenv("REDIS_PORT"),
		Password:       os.Getenv("REDIS_PASSWORD"),
		ExpirationTime: os.Getenv("REDIS_EXPIRATION_TIME"),
	}

	return &Config{
		Telegram: telegramConfig,
		OpenAI:   openAIConfig,
		Redis:    redisConfig,
	}
}
