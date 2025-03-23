package bot

import (
	"fmt"
	"log"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Assistant interface {
	Ask(username string, request string) (response string, err error)
}

type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type Bot struct {
	api           BotAPI
	name          string
	chatID        int64
	assignedChats []int64
}

func NewBot(api BotAPI, name string, chatID int64, assignedChats []int64) *Bot {
	return &Bot{
		api:           api,
		name:          name,
		chatID:        chatID,
		assignedChats: assignedChats,
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

	log.Printf("Incoming message: chat_id: %d, from_user: %s, text: %s\n", msg.Chat.ID, msg.From.UserName, msg.Text)

	req, err := b.parse(msg.Chat.ID, msg.Text)
	if err != nil {
		log.Println("Parse error:", err)
		return
	}

	res, err := assistant.Ask(msg.From.UserName, req)
	if err != nil {
		log.Println("Message handler error:", err)
		return
	}

	err = b.send(msg.Chat.ID, msg.MessageID, res)
	if err != nil {
		log.Println("Send error:", err)
		return
	}

	log.Printf("Outgoing message: chat_id: %d, reply_to_message_id: %d, text: %s", msg.Chat.ID, msg.MessageID, res)
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

func (b *Bot) send(chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = messageID
	_, err := b.api.Send(msg)
	return err
}
