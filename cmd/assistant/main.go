package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"telegrambot-assistant/services/config"
	"telegrambot-assistant/services/setup"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger := setup.InitLogger(cfg.Loki, "telegram-assistant")

	bot, err := setup.InitBot(cfg.Telegram, logger)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	redisClient, err := setup.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	redisStorage := setup.InitStorage(redisClient, cfg.Redis.ExpirationTime)
	threadRepo := setup.InitRepository(redisStorage)
	openAiAssistant := setup.InitAssistant(cfg.OpenAI, threadRepo)

	go bot.HandleMessages(openAiAssistant)

	log.Println("Assistant service started...")

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Shutting down assistant service...")
}
