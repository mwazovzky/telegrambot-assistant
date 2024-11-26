package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
	botName := os.Getenv("BOT_NAME")
	bot := initBot()
	ai := initAssistant(botName)
	p := initParser()

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

func initAssistant(name string) *openai.Assistant {
	role := fmt.Sprintf("You are assistant. Your name is %s", name)

	url := "https://api.openai.com/v1/chat/completions"
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_MODEL")
	client := openaiclient.NewOpenAiClient(url, apiKey)

	r := initRedis()
	s := initStorage(r)
	tr := repository.NewCachedRepository(s)

	return openai.NewAssistant(model, role, client, tr)
}

func initBot() *tgbotapi.BotAPI {
	botToken := os.Getenv("TELEGRAM_HTTP_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true
	log.Printf("TelegramBot: authorized on account %s", bot.Self.UserName)

	return bot
}

func initStorage(r *redis.Client) *storage.RedisService {
	ets := os.Getenv("REDIS_EXPIRATION_TIME")
	et, err := strconv.Atoi(ets)
	if err != nil {
		log.Fatal("can not load config, error", err)
	}
	ttl := time.Duration(et) * time.Second
	return storage.NewRedisService(r, ttl)
}

func initRedis() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	pwd := os.Getenv("REDIS_PASSWORD")
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

func initParser() *parser.Parser {
	botName := os.Getenv("BOT_NAME")
	if botName == "" {
		log.Fatal("config err, bot name empty")
	}

	botChatID := os.Getenv("BOT_CHAT_ID")
	botChat, err := strconv.ParseInt(botChatID, 10, 64)
	if err != nil {
		log.Fatal("config err", err)
	}

	chats := os.Getenv("ASSIGNED_CHATS")
	assignedChats := []int64{}
	for _, value := range strings.Split(chats, ",") {
		chat, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatal("config err", err)
		}
		assignedChats = append(assignedChats, chat)
	}

	return parser.NewParser(botName, botChat, assignedChats)
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
