package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegrambot-assistant/services/config"
	"telegrambot-assistant/services/parser"
	"telegrambot-assistant/services/repository"
	"telegrambot-assistant/services/storage"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/mwazovzky/assistant"
	openaiclient "github.com/mwazovzky/assistant/http/client"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	r := initRedis(cfg.Redis)
	s := initStorage(r, cfg.Redis.ExpirationTime)
	tr := repository.NewCachedRepository(s)
	ai := initAssistant(cfg.OpenAI, tr)
	bot := initBot(cfg.Telegram)
	p := initParser(cfg.Telegram)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		msg := update.Message
		if msg == nil {
			continue
		}

		log.Printf("Incoming message: chat_id: %d, from_user: %s, text: %s\n", msg.Chat.ID, msg.From.UserName, msg.Text)

		req, err := p.Parse(msg.Chat.ID, msg.Text)
		if err != nil {
			log.Println("Parse error", err)
			continue
		}

		res, err := getResponse(ai, req, msg.From.UserName)
		if err != nil {
			log.Println("getResponse error", err)
			continue
		}

		err = sendResponse(bot, msg.Chat.ID, msg.MessageID, res)

		if err != nil {
			log.Println("sendResponse error", err)
			continue
		}

		log.Printf("Outgoing message: chat_id: %d, reply_to_message_id: %d, text: %s", msg.Chat.ID, msg.MessageID, res)
	}
}

func initAssistant(cfg config.OpenAIConfig, tr openai.TreadRepository) *openai.Assistant {
	role := fmt.Sprintf("You are assistant. Your name is %s", cfg.Name)
	client := openaiclient.NewOpenAiClient(cfg.ApiUrl, cfg.ApiKey)

	return openai.NewAssistant(cfg.Model, role, client, tr)
}

func initBot(cfg config.TelegramConfig) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(cfg.ApiToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true
	log.Printf("TelegramBot: authorized on account %s", bot.Self.UserName)

	return bot
}

func initStorage(r *redis.Client, ets string) *storage.RedisService {
	et, err := strconv.Atoi(ets)
	if err != nil {
		log.Fatal("can not load config, error", err)
	}
	ttl := time.Duration(et) * time.Second
	return storage.NewRedisService(r, ttl)
}

func initRedis(cfg config.RedisConfig) *redis.Client {
	host := cfg.Host
	port := cfg.Port
	pwd := cfg.Password
	adr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     adr,
		Password: pwd,
		DB:       0,
	})

	ctx := context.Background()
	str, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("ERROR", err)
	}

	log.Println("Connected to redis", str)

	return client
}

func initParser(cfg config.TelegramConfig) *parser.Parser {
	botName := cfg.BotName
	botChatID := cfg.ChatID
	assignedChats := cfg.AssignedChats

	if botName == "" {
		log.Fatal("config err, bot name empty")
	}

	botChat, err := strconv.ParseInt(botChatID, 10, 64)
	if err != nil {
		log.Fatal("config err", err)
	}

	chats := []int64{}
	for _, value := range strings.Split(assignedChats, ",") {
		chat, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatal("config err", err)
		}
		chats = append(chats, chat)
	}

	return parser.NewParser(botName, botChat, chats)
}

func getResponse(ai *openai.Assistant, req string, username string) (string, error) {
	_, err := ai.GetThread(username)

	if err != nil {
		err = ai.CreateThread(username)
		if err != nil {
			return "", fmt.Errorf("can not get or create thread, error: %s", err)
		}
	}

	res, err := ai.Post(username, req)
	if err != nil {
		return "", fmt.Errorf("can not post a question, error: %s", err)
	}

	return res, nil
}

func sendResponse(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = messageID
	_, err := bot.Send(msg)
	return err
}
