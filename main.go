package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"telegrambot-assistant/services/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/mwazovzky/assistant"
	openaiclient "github.com/mwazovzky/assistant/http/client"
)

func main() {
	botName := os.Getenv("BOT_NAME")
	bot := initBot()
	ai := initAssistant(botName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			msg := update.Message
			log.Printf("Incoming message: chat_id: %d, from_user: %s, text: %s\n", msg.Chat.ID, msg.From.UserName, msg.Text)

			if !isValidChat(msg.Chat.ID) {
				continue
			}

			req, err := parseRequest(msg.Text, botName)
			if err != nil {
				log.Println("error", err)
				continue
			}

			res, err := getResponse(ai, req, msg.From.UserName)
			if err != nil {
				log.Println("error", err)
				continue
			}

			sendResponse(bot, msg.Chat.ID, msg.MessageID, res)
			log.Printf("Outgoing message: chat_id: %d, reply_to_message_id: %d, text: %s", msg.Chat.ID, msg.MessageID, res)
		}
	}
}

func initAssistant(name string) *openai.Assistant {
	url := "https://api.openai.com/v1/chat/completions"
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openaiclient.NewOpenAiClient(url, apiKey)
	tr := repository.NewThreadRepository()
	role := fmt.Sprintf("You are assistant. Your name is %s", name)

	return openai.NewAssistant(role, client, tr)
}

func initBot() *tgbotapi.BotAPI {
	botToken := os.Getenv("TELEGRAM_HTTP_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return bot
}

func parseRequest(txt string, botName string) (string, error) {
	if !strings.HasPrefix(txt, botName) {
		return "", fmt.Errorf("can not process request")
	}

	trimmedSymbols := "!, "

	return strings.TrimLeft(strings.TrimPrefix(txt, botName), trimmedSymbols), nil
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

func sendResponse(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = messageID
	bot.Send(msg)
}

func isValidChat(chatID int64) bool {
	return slices.Contains(getChats(), strconv.FormatInt(chatID, 10))
}

func getChats() []string {
	chats := os.Getenv("ALLOWED_CHATS")
	return strings.Split(chats, ",")
}