package bot

import (
	"fmt"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Assistant interface {
	Ask(username string, request string) (response string, err error)
}

type Splitter interface {
	Split(text string) ([]string, error)
}

type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type Logger interface {
	Info(message string, keyValues ...interface{}) error
	Error(message string, keyValues ...interface{}) error
	Debug(message string, keyValues ...interface{}) error
}

type Bot struct {
	botApi     BotAPI
	name       string
	userChats  []string
	groupChats []int64
	splitter   Splitter
	logger     Logger
}

func NewBot(botApi BotAPI, name string, userChats []string, groupChats []int64, splitter Splitter, logger Logger) *Bot {
	return &Bot{
		botApi:     botApi,
		name:       name,
		userChats:  userChats,
		groupChats: groupChats,
		splitter:   splitter,
		logger:     logger,
	}
}

func (b *Bot) HandleMessages(assistant Assistant) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.botApi.GetUpdatesChan(u)

	for update := range updates {
		b.handleUpdate(update, assistant)
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update, assistant Assistant) {
	msg := update.Message
	if msg == nil {
		return
	}

	b.logger.Info("Incoming message", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "text", msg.Text)

	req, err := b.parse(msg.Chat.ID, msg.From.UserName, msg.Text)
	if err != nil {
		b.logger.Error("Parse error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err)
		return
	}

	text, err := assistant.Ask(msg.From.UserName, req)
	if err != nil {
		b.logger.Error("Assistant error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err)
		return
	}

	chunks, err := b.splitter.Split(text)
	if err != nil {
		b.logger.Error("Splitter error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err, "text", text, "chunks", chunks)
		return
	}

	b.logger.Info("Outgoing message", "chat_id", msg.Chat.ID, "reply_to_message_id", msg.MessageID, "text", text)

	err = b.send(msg.Chat.ID, msg.MessageID, chunks)
	if err != nil {
		b.logger.Error("Send error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err)
		return
	}
}

func (b *Bot) parse(chatID int64, username string, txt string) (string, error) {
	// Check if the username is in the allowed userChats list
	if slices.Contains(b.userChats, username) {
		return txt, nil
	}
	// Check if the chat ID is in the allowed groupChats list
	if !slices.Contains(b.groupChats, chatID) {
		return "", fmt.Errorf("cannot process chat")
	}
	// Check if group chat the message starts with the bot's name
	if !strings.HasPrefix(txt, b.name) {
		return "", fmt.Errorf("cannot process request")
	}
	// Remove the bot's name and any leading symbols
	trimmedSymbols := "!, "
	trimmedText := strings.TrimPrefix(txt, b.name)
	trimmedText = strings.TrimLeft(trimmedText, trimmedSymbols)
	return trimmedText, nil
}

func (b *Bot) send(chatID int64, messageID int, chunks []string) error {
	for _, chunk := range chunks {
		msg := tgbotapi.NewMessage(chatID, chunk)
		msg.ReplyToMessageID = messageID
		_, err := b.botApi.Send(msg)
		return err
	}

	return nil
}
