package setup

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"telegrambot-assistant/services/bot"
	"telegrambot-assistant/services/config"
	"telegrambot-assistant/services/repository"
	"telegrambot-assistant/services/storage"
	"telegrambot-assistant/services/textsplitter"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mwazovzky/assistant"
	openaiclient "github.com/mwazovzky/assistant/http/client"
	"github.com/mwazovzky/cloudlog"
	"github.com/redis/go-redis/v9"
)

var newBotAPI = tgbotapi.NewBotAPI

func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return client, nil
}

func InitStorage(r *redis.Client, ttl time.Duration) *storage.RedisService {
	return storage.NewRedisService(r, ttl)
}

func InitRepository(client repository.CacheClient) *repository.CacheRepository {
	return repository.NewCachedRepository(client)
}

func InitAssistant(cfg config.OpenAIConfig, tr assistant.ThreadRepository) *assistant.Assistant {
	role := fmt.Sprintf("%s Your name is %s", cfg.Role, cfg.Name)
	client := openaiclient.NewOpenAiClient(cfg.ApiUrl, cfg.ApiKey)

	return assistant.NewAssistant(cfg.Model, role, client, tr)
}

func InitLogger(cfg config.LokiConfig, service string) *cloudlog.Logger {
	httpClient := &http.Client{
		Timeout: 10 * time.Second, // Added timeout for HTTP client
	}
	client := cloudlog.NewClient(cfg.Url, cfg.Username, cfg.Token, httpClient)
	return cloudlog.NewLogger(client, service)
}

func InitBot(cfg config.TelegramConfig, logger bot.Logger) (*bot.Bot, error) {
	telegramBot, err := newBotAPI(cfg.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %v", err)
	}

	log.Printf("TelegramBot: authorized on account %s", telegramBot.Self.UserName)

	splitter := textsplitter.NewTextSplitter(cfg.MessageLimit)

	return bot.NewBot(telegramBot, cfg.BotName, cfg.Users, cfg.Chats, splitter, logger), nil
}
