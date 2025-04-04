package bot

import (
	"fmt"
	"slices"
	"strings"
	"sync"

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

// ChunkQueue maintains pending chunks for a specific conversation
type ChunkQueue struct {
	Chunks     []string
	Position   int
	OriginalID int
}

type Bot struct {
	botApi      BotAPI
	name        string
	userChats   []string
	groupChats  []int64
	splitter    Splitter
	logger      Logger
	useShowMore bool // Renamed from useShowMoreInterface

	// Map to store message chunks: chatID+username → ChunkQueue
	pendingChunks map[string]*ChunkQueue
	chunksMutex   sync.RWMutex
}

// BotConfig holds configuration options for the Bot
type BotConfig struct {
	Name        string
	UserChats   []string
	GroupChats  []int64
	UseShowMore bool // Renamed from UseShowMoreInterface
}

func NewBot(botApi BotAPI, config BotConfig, splitter Splitter, logger Logger) *Bot {
	return &Bot{
		botApi:        botApi,
		name:          config.Name,
		userChats:     config.UserChats,
		groupChats:    config.GroupChats,
		useShowMore:   config.UseShowMore, // Renamed
		splitter:      splitter,
		logger:        logger,
		pendingChunks: make(map[string]*ChunkQueue),
	}
}

func (b *Bot) HandleMessages(assistant Assistant) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.botApi.GetUpdatesChan(u)

	for update := range updates {
		// Handle callback queries (button clicks)
		if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
			continue
		}

		// Handle normal messages
		b.handleUpdate(update, assistant)
	}
}

func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// Parse callback data
	if query.Data == "show_more" {
		chatID := query.Message.Chat.ID
		username := query.From.UserName

		// Get conversation key
		convKey := fmt.Sprintf("%d:%s", chatID, username)

		b.chunksMutex.Lock()
		defer b.chunksMutex.Unlock()

		queue, exists := b.pendingChunks[convKey]
		if !exists || queue.Position >= len(queue.Chunks) {
			// No more chunks available
			callback := tgbotapi.NewCallback(query.ID, "No more content available")
			_, _ = b.botApi.Send(callback)
			return
		}

		// Send next chunk
		msg := tgbotapi.NewMessage(chatID, queue.Chunks[queue.Position])
		msg.ReplyToMessageID = queue.OriginalID

		// Add button if there are more chunks
		if queue.Position < len(queue.Chunks)-1 {
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Show More", "show_more"),
				),
			)
			msg.ReplyMarkup = keyboard
		}

		_, err := b.botApi.Send(msg)
		if err != nil {
			b.logger.Error("Failed to send next chunk", "error", err, "chat_id", chatID)
		}

		// Acknowledge the callback query
		callback := tgbotapi.NewCallback(query.ID, "")
		_, _ = b.botApi.Send(callback)

		// Move to next chunk
		queue.Position++
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

	b.logger.Info("Outgoing message", "chat_id", msg.Chat.ID, "reply_to_message_id", msg.MessageID, "text", text, "chunks_count", len(chunks))

	err = b.send(msg.Chat.ID, msg.From.UserName, msg.MessageID, chunks)
	if err != nil {
		b.logger.Error("Send error", "chat_id", msg.Chat.ID, "from_user", msg.From.UserName, "error", err)
		return
	}
}

func (b *Bot) parse(chatID int64, username string, txt string) (string, error) {
	// handle user chat message
	if slices.Contains(b.userChats, username) && !slices.Contains(b.groupChats, chatID) {
		return txt, nil
	}

	// handle group chat message
	if slices.Contains(b.groupChats, chatID) && strings.HasPrefix(txt, b.name) {
		// Remove the bot's name and any leading symbols
		trimmedSymbols := "!, "
		trimmedText := strings.TrimPrefix(txt, b.name)
		trimmedText = strings.TrimLeft(trimmedText, trimmedSymbols)
		return trimmedText, nil
	}

	return "", fmt.Errorf("cannot process chat message")
}

func (b *Bot) send(chatID int64, username string, messageID int, chunks []string) error {
	if len(chunks) == 0 {
		return nil
	}

	// If not using Show More or only one chunk, send all at once
	if !b.useShowMore || len(chunks) == 1 {
		for _, chunk := range chunks {
			msg := tgbotapi.NewMessage(chatID, chunk)
			msg.ReplyToMessageID = messageID
			_, err := b.botApi.Send(msg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Using Show More with multiple chunks
	// Send the first chunk with a "Show More" button
	msg := tgbotapi.NewMessage(chatID, chunks[0])
	msg.ReplyToMessageID = messageID

	// Add a "Show More" button if there are more chunks
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Show More", "show_more"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := b.botApi.Send(msg)
	if err != nil {
		return err
	}

	// Store the remaining chunks in the queue
	convKey := fmt.Sprintf("%d:%s", chatID, username)

	b.chunksMutex.Lock()
	b.pendingChunks[convKey] = &ChunkQueue{
		Chunks:     chunks,
		Position:   1, // Start from second chunk (index 1)
		OriginalID: messageID,
	}
	b.chunksMutex.Unlock()

	return nil
}
