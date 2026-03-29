package setup

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"telegrambot-assistant/services/bot"
	"telegrambot-assistant/services/config"
	localai "telegrambot-assistant/services/openai"
	"telegrambot-assistant/services/repository"
	"telegrambot-assistant/services/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/logger"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
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

func InitResponseStore(client repository.CacheClient) *repository.CacheRepository {
	return repository.NewCachedRepository(client)
}

func InitAssistant(cfg config.OpenAIConfig, store repository.ResponseStore) *localai.Assistant {
	instructions := fmt.Sprintf("%s Your name is %s", cfg.Role, cfg.Name)
	client := openai.NewClient(option.WithAPIKey(cfg.ApiKey))

	return localai.NewAssistant(&client.Responses, cfg.Model, instructions, store, cfg.RequestTimeout)
}

// LoggerResources holds the logger and async sender for graceful shutdown
type LoggerResources struct {
	Logger logger.Logger
	Sender *logger.AsyncSender
}

func InitLogger(cfg config.LokiConfig, service string) *LoggerResources {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	lokiClient := client.NewLokiClient(cfg.Url, cfg.Username, cfg.Token, httpClient)
	sender := logger.NewAsyncSender(lokiClient)
	log := logger.New(sender, logger.WithJob(service))

	return &LoggerResources{
		Logger: log,
		Sender: sender,
	}
}

func InitBot(cfg config.TelegramConfig, logger bot.Logger) (*bot.Bot, error) {
	telegramBot, err := newBotAPI(cfg.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %v", err)
	}

	log.Printf("TelegramBot: authorized on account %s", telegramBot.Self.UserName)

	// Updated: Use BasicSplitter implementation
	splitter := bot.NewBasicSplitter(cfg.MessageLimit)

	// Create bot config struct
	botConfig := bot.BotConfig{
		Name:        cfg.BotName,
		UserChats:   cfg.Users,
		GroupChats:  cfg.Chats,
		UseShowMore: cfg.ShowMore,
	}

	return bot.NewBot(telegramBot, botConfig, splitter, logger), nil
}
