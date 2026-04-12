package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"telegrambot-assistant/services/config"
	"telegrambot-assistant/services/logger"
	"telegrambot-assistant/services/setup"
)

func main() {
	appLogger := logger.New()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		appLogger.Error(ctx, "Failed to load configuration", "error", err)
		os.Exit(1)
	}

	bot, err := setup.InitBot(cfg.Telegram, appLogger)
	if err != nil {
		appLogger.Error(ctx, "Failed to initialize Telegram bot", "error", err)
		os.Exit(1)
	}

	redisClient, err := setup.InitRedis(cfg.Redis)
	if err != nil {
		appLogger.Error(ctx, "Failed to initialize Redis", "error", err)
		os.Exit(1)
	}

	redisStorage := setup.InitStorage(redisClient, cfg.Redis.ExpirationTime)
	responseStore := setup.InitResponseStore(redisStorage)
	openAiAssistant := setup.InitAssistant(cfg.OpenAI, responseStore, appLogger)

	go bot.HandleMessages(openAiAssistant)

	appLogger.Info(ctx, "Assistant service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	appLogger.Info(ctx, "Shutting down assistant service")
}
