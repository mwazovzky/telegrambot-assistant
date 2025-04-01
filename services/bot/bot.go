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
	api           BotAPI
	name          string
	chatID        int64
	assignedChats []int64
	splitter      Splitter
	logger        Logger
}

func NewBot(api BotAPI, name string, chatID int64, assignedChats []int64, splitter Splitter, logger Logger) *Bot {
	return &Bot{
		api:           api,
		name:          name,
		chatID:        chatID,
		assignedChats: assignedChats,
		splitter:      splitter,
		logger:        logger,
	}
}

func (b *Bot) HandleMessages(assistant Assistant) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		b.handleUpdate(update, assistant)
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update, assistant Assistant) {
	msg := update.Message
	if msg == nil {
		return
	}

	b.logger.Debug("Incoming message", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "text", msg.Text)

	req, err := b.parse(msg.Chat.ID, msg.Text)
	if err != nil {
		b.logger.Error("Parse error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err)
		return
	}

	res, err := assistant.Ask(msg.From.UserName, req)
	if err != nil {
		b.logger.Error("Assistant error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err)
		return
	}

	chunks, err := b.splitter.Split(res)
	if err != nil {
		b.logger.Error("Splitter error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err, "text", res)
		return
	}

	err = b.send(msg.Chat.ID, msg.MessageID, chunks)
	if err != nil {
		b.logger.Error("Send error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err)
		return
	}

	b.logger.Debug("Outgoing message", "chat_id", msg.Chat.ID, "reply_to_message_id", msg.MessageID, "text", res)
}

func (b *Bot) parse(chatID int64, txt string) (string, error) {
	if chatID == b.chatID {
		return txt, nil
	}

	if !slices.Contains(b.assignedChats, chatID) {
		return "", fmt.Errorf("cannot process chat")
	}

	if !strings.HasPrefix(txt, b.name) {
		return "", fmt.Errorf("cannot process request")
	}

	trimmedSymbols := "!, "
	return strings.TrimLeft(strings.TrimPrefix(txt, b.name), trimmedSymbols), nil
}

func (b *Bot) send(chatID int64, messageID int, chunks []string) error {
	for _, chunk := range chunks {
		msg := tgbotapi.NewMessage(chatID, chunk)
		msg.ReplyToMessageID = messageID
		_, err := b.api.Send(msg)
		return err
	}

	return nil
}
