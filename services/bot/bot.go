package bot

import (
	"fmt"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	// Callback actions
	ShowMoreCallback = "show_more"

	// Error messages
	ErrAssistantResponse = "Sorry, I'm having trouble generating a response right now. Please try again later."
	ErrSplittingResponse = "Sorry, I'm having trouble processing my response. Please try a simpler query."
	ErrSendingResponse   = "Sorry, I'm having trouble sending my response. Please try again later."
	ErrLoadingNextPart   = "Sorry, I couldn't load the next part of the message. Please try again."
	ErrNoMoreContent     = "No more content available"
	ErrFailedLoadMore    = "Failed to load more content"

	// Log keys
	LogKeyChatID       = "chat_id"
	LogKeyFromUser     = "from_user"
	LogKeyText         = "text"
	LogKeyError        = "error"
	LogKeyChunks       = "chunks"
	LogKeyChunksCount  = "chunks_count"
	LogKeyReplyToMsgID = "reply_to_message_id"
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
	botApi       BotAPI
	name         string
	userChats    []string
	groupChats   []int64
	splitter     Splitter
	logger       Logger
	useShowMore  bool
	chunkStorage ChunkStorage
}

// BotConfig holds configuration options for the Bot
type BotConfig struct {
	Name        string
	UserChats   []string
	GroupChats  []int64
	UseShowMore bool
}

func NewBot(botApi BotAPI, config BotConfig, splitter Splitter, logger Logger) *Bot {
	// Create default in-memory storage if not provided
	return NewBotWithChunkStorage(botApi, config, splitter, logger, NewInMemoryChunkStorage())
}

// New constructor that accepts custom chunk storage
func NewBotWithChunkStorage(botApi BotAPI, config BotConfig, splitter Splitter, logger Logger, chunkStorage ChunkStorage) *Bot {
	return &Bot{
		botApi:       botApi,
		name:         config.Name,
		userChats:    config.UserChats,
		groupChats:   config.GroupChats,
		useShowMore:  config.UseShowMore,
		splitter:     splitter,
		logger:       logger,
		chunkStorage: chunkStorage,
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
	if query.Data == ShowMoreCallback {
		chatID := query.Message.Chat.ID
		username := query.From.UserName

		chunk, originalID, hasMore, exists := b.chunkStorage.GetNextChunk(chatID, username)
		if !exists {
			// No more chunks available
			callback := tgbotapi.NewCallback(query.ID, ErrNoMoreContent)
			_, _ = b.botApi.Send(callback)
			return
		}

		// Send next chunk
		msg := tgbotapi.NewMessage(chatID, chunk)
		msg.ReplyToMessageID = originalID

		// Add button if there are more chunks
		if hasMore {
			msg.ReplyMarkup = createShowMoreKeyboard()
		}

		_, err := b.botApi.Send(msg)
		if err != nil {
			b.logger.Error("Failed to send next chunk", LogKeyError, err, LogKeyChatID, chatID)

			// Send error message about the failure
			errorMsg := tgbotapi.NewMessage(chatID, ErrLoadingNextPart)
			_, _ = b.botApi.Send(errorMsg)

			// Still acknowledge the callback to remove the loading indicator
			callback := tgbotapi.NewCallback(query.ID, ErrFailedLoadMore)
			_, _ = b.botApi.Send(callback)
			return
		}

		// Acknowledge the callback query
		callback := tgbotapi.NewCallback(query.ID, "")
		_, _ = b.botApi.Send(callback)

		// Clean up chunk storage after all chunks have been delivered
		if !hasMore {
			b.chunkStorage.Clear(chatID, username)
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update, assistant Assistant) {
	msg := update.Message
	if msg == nil {
		return
	}

	b.logger.Info("Incoming message", LogKeyChatID, msg.Chat.ID, LogKeyFromUser, msg.From.UserName, LogKeyText, msg.Text)

	req, err := b.parse(msg.Chat.ID, msg.From.UserName, msg.Text)
	if err != nil {
		b.logger.Error("Parse error", LogKeyChatID, msg.Chat.ID, LogKeyFromUser, msg.From.UserName, LogKeyError, err)
		// We don't send error feedback for parse errors to avoid responding to messages not intended for the bot
		return
	}

	text, err := assistant.Ask(msg.From.UserName, req)
	if err != nil {
		b.logger.Error("Assistant error", LogKeyChatID, msg.Chat.ID, LogKeyFromUser, msg.From.UserName, LogKeyError, err)
		// Send error feedback to user
		b.sendErrorMessage(msg.Chat.ID, msg.MessageID, ErrAssistantResponse)
		return
	}

	chunks, err := b.splitter.Split(text)
	if err != nil {
		b.logger.Error("Splitter error", LogKeyChatID, msg.Chat.ID, LogKeyFromUser, msg.From.UserName, LogKeyError, err, LogKeyText, text, LogKeyChunks, chunks)
		// Send error feedback to user
		b.sendErrorMessage(msg.Chat.ID, msg.MessageID, ErrSplittingResponse)
		return
	}

	b.logger.Info("Outgoing message", LogKeyChatID, msg.Chat.ID, LogKeyReplyToMsgID, msg.MessageID, LogKeyText, text, LogKeyChunksCount, len(chunks))

	err = b.send(msg.Chat.ID, msg.From.UserName, msg.MessageID, chunks)
	if err != nil {
		b.logger.Error("Send error", LogKeyChatID, msg.Chat.ID, LogKeyFromUser, msg.From.UserName, LogKeyError, err)
		// Send error feedback to user
		b.sendErrorMessage(msg.Chat.ID, msg.MessageID, ErrSendingResponse)
		return
	}
}

// sendErrorMessage sends an error message to the user
func (b *Bot) sendErrorMessage(chatID int64, replyToID int, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyToID

	// Try to send the error message, but don't worry if this also fails
	_, err := b.botApi.Send(msg)
	if err != nil {
		b.logger.Error("Failed to send error message", LogKeyChatID, chatID, LogKeyError, err)
	}
}

func (b *Bot) parse(chatID int64, username string, txt string) (string, error) {
	// handle user chat message
	if b.isAuthorizedUser(username) && !b.isAuthorizedGroup(chatID) {
		return txt, nil
	}

	// handle group chat message
	if b.isAuthorizedGroup(chatID) && strings.HasPrefix(txt, b.name) {
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
			// Ensure no ReplyMarkup is set when not using Show More
			msg.ReplyMarkup = nil
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
	msg.ReplyMarkup = createShowMoreKeyboard()

	_, err := b.botApi.Send(msg)
	if err != nil {
		return err
	}

	// Store the remaining chunks in the storage
	b.chunkStorage.StoreChunks(chatID, username, messageID, chunks)

	return nil
}

// createShowMoreKeyboard creates a keyboard with a "Show More" button
func createShowMoreKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Show More", ShowMoreCallback),
		),
	)
}

// isAuthorizedUser checks if a user is authorized to use the bot in private chats
func (b *Bot) isAuthorizedUser(username string) bool {
	return slices.Contains(b.userChats, username)
}

// isAuthorizedGroup checks if a group is authorized to use the bot
func (b *Bot) isAuthorizedGroup(chatID int64) bool {
	return slices.Contains(b.groupChats, chatID)
}
