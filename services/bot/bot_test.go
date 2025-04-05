package bot

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Common test constants
const (
	testBotName       = "testbot"
	testUserName      = "testuser"
	testAllowedUser   = "allowed_user"
	testPrivateChatID = int64(11111)
	testGroupChatID1  = int64(12345)
	testGroupChatID2  = int64(67890)
)

// Mock implementations
type MockBotAPI struct {
	mock.Mock
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	args := m.Called(config)
	return args.Get(0).(tgbotapi.UpdatesChannel)
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	return args.Get(0).(tgbotapi.Message), args.Error(1)
}

type MockAssistant struct {
	mock.Mock
}

func (m *MockAssistant) Ask(username string, request string) (string, error) {
	args := m.Called(username, request)
	return args.String(0), args.Error(1)
}

type MockSplitter struct {
	mock.Mock
}

func (m *MockSplitter) Split(text string) ([]string, error) {
	args := m.Called(text)
	return args.Get(0).([]string), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(message string, keyValues ...interface{}) error {
	args := m.Called(message, keyValues)
	return args.Error(0)
}

func (m *MockLogger) Error(message string, keyValues ...interface{}) error {
	args := m.Called(message, keyValues)
	return args.Error(0)
}

func (m *MockLogger) Debug(message string, keyValues ...interface{}) error {
	args := m.Called(message, keyValues)
	return args.Error(0)
}

// Helper functions for tests
func createTestBot(mockBotAPI BotAPI, mockSplitter Splitter, mockLogger Logger) *Bot {
	config := BotConfig{
		Name:        testBotName,
		UserChats:   []string{testAllowedUser, testUserName},
		GroupChats:  []int64{testGroupChatID1, testGroupChatID2},
		UseShowMore: true,
	}
	return NewBot(mockBotAPI, config, mockSplitter, mockLogger)
}

// Helper to create a custom bot with specific config
func createCustomBot(mockBotAPI BotAPI, mockSplitter Splitter, mockLogger Logger, useShowMore bool, userChats []string) *Bot {
	config := BotConfig{
		Name:        testBotName,
		UserChats:   userChats,
		GroupChats:  []int64{testGroupChatID1, testGroupChatID2},
		UseShowMore: useShowMore,
	}
	return NewBot(mockBotAPI, config, mockSplitter, mockLogger)
}

// Helper to create a test update with a message
func createMessageUpdate(chatID int64, username, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: chatID},
			From:      &tgbotapi.User{UserName: username},
			Text:      text,
			MessageID: 123, // Use a consistent ID for simplicity
		},
	}
}

// Helper to set up logger expectations for common message patterns
func setupLoggerExpectations(mockLogger *MockLogger, operation string, chatID int64, username, text string) {
	switch operation {
	case "incoming":
		mockLogger.On("Info", "Incoming message", []interface{}{
			LogKeyChatID, chatID, LogKeyFromUser, username, LogKeyText, text,
		}).Return(nil)
	case "outgoing":
		mockLogger.On("Info", "Outgoing message", []interface{}{
			LogKeyChatID, chatID, LogKeyReplyToMsgID, 123, LogKeyText, text, LogKeyChunksCount, 1,
		}).Return(nil)
	case "parse-error":
		// Use mock.MatchedBy for the error parameter instead of a specific type
		mockLogger.On("Error", "Parse error", mock.MatchedBy(func(args []interface{}) bool {
			if len(args) != 6 {
				return false
			}

			// Check for expected keys and values
			if args[0] != LogKeyChatID || args[2] != LogKeyFromUser || args[4] != LogKeyError {
				return false
			}

			// Verify chat ID and username
			chatID, ok := args[1].(int64)
			if !ok || chatID != testGroupChatID1 {
				return false
			}

			username, ok := args[3].(string)
			if !ok || username != testUserName {
				return false
			}

			// Just check that there's an error, don't worry about its specific type
			_, ok = args[5].(error)
			return ok
		})).Return(nil)
	}
}

// setupMockForErrorResponse creates a matcher for error message responses
func setupMockForErrorResponse(mockBotAPI *MockBotAPI, errorMsgContains string) {
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && strings.Contains(msg.Text, errorMsgContains)
	})).Return(tgbotapi.Message{}, nil).Once()
}

// Tests for message parsing
func TestBot_parse(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	tests := []struct {
		name     string
		chatID   int64
		username string
		txt      string
		want     string
		err      bool
	}{
		{"allowed user in private chat", 11111, "allowed_user", "test message", "test message", false},                    // Private chat, allowed user
		{"allowed user in group chat without prefix", 67890, "allowed_user", "test message", "", true},                    // Group chat but no prefix
		{"allowed user in group chat with prefix", 67890, "allowed_user", "testbot hello", "hello", false},                // Group chat with prefix
		{"non-allowed user in group chat with prefix", 67890, "other_user", "testbot hello", "hello", false},              // Non-allowed user but in group with prefix
		{"non-allowed user in group chat with prefix and symbols", 67890, "other_user", "testbot! hello", "hello", false}, // Same with symbols
		{"non-allowed user in group chat without prefix", 67890, "other_user", "test message", "", true},                  // No prefix in group chat
		{"non-allowed user in private chat", 11111, "other_user", "test message", "", true},                               // Not allowed user in private chat
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bot.parse(tt.chatID, tt.username, tt.txt)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// Tests for sending messages
func TestSend_WithShowMore(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Expect only the first message to be sent (with button)
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "message part 1" && msg.ReplyMarkup != nil
	})).Return(tgbotapi.Message{}, nil).Once()
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	err := bot.send(12345, "testuser", 1, []string{"message part 1", "message part 2"})
	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	// Check that the chunks were stored
	bot.chunksMutex.RLock()
	queue, exists := bot.pendingChunks["12345:testuser"]
	bot.chunksMutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, 1, queue.Position)
	assert.Equal(t, []string{"message part 1", "message part 2"}, queue.Chunks)
}

func TestSend_WithoutShowMore(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Expect both messages to be sent (without buttons)
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "message part 1" && msg.ReplyMarkup == nil
	})).Return(tgbotapi.Message{}, nil).Once()
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "message part 2" && msg.ReplyMarkup == nil
	})).Return(tgbotapi.Message{}, nil).Once()
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: false,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	err := bot.send(12345, "testuser", 1, []string{"message part 1", "message part 2"})
	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	// Check that no chunks were stored when not using Show More
	bot.chunksMutex.RLock()
	_, exists := bot.pendingChunks["12345:testuser"]
	bot.chunksMutex.RUnlock()
	assert.False(t, exists)
}

func TestSend_SingleChunk(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Expect a single message without button
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "single message" && msg.ReplyMarkup == nil
	})).Return(tgbotapi.Message{}, nil).Once()
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true, // Even with show more enabled, single chunks have no buttons
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	err := bot.send(12345, "testuser", 1, []string{"single message"})
	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
}

func TestSend_Error(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	// Fix: Add username parameter to match the updated function signature
	err := bot.send(12345, "testuser", 1, []string{"response message"})
	assert.Error(t, err)
	mockBotAPI.AssertExpectations(t)
}

// Tests for handling user messages
func TestBot_handleUpdate_Success(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testPrivateChatID, testUserName, "test message")
	setupLoggerExpectations(mockLogger, "outgoing", testPrivateChatID, testUserName, "response message")
	mockAssistant.On("Ask", testUserName, "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)
	// Create bot and test update
	bot := createTestBot(mockBotAPI, mockSplitter, mockLogger)
	update := createMessageUpdate(testPrivateChatID, testUserName, "test message")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_Success_GroupChat(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testGroupChatID2, testUserName, "testbot hello")
	setupLoggerExpectations(mockLogger, "outgoing", testGroupChatID2, testUserName, "response message")
	mockAssistant.On("Ask", testUserName, "hello").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)
	// Create bot and test update
	bot := createTestBot(mockBotAPI, mockSplitter, mockLogger)
	update := createMessageUpdate(testGroupChatID2, testUserName, "testbot hello")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

// Tests for error handling
func TestBot_handleUpdate_ParseError_GroupChatNoPrefix(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testGroupChatID1, testUserName, "hello without prefix")
	// Alternative approach for parse-error: set up the expectation directly
	mockLogger.On("Error", "Parse error", mock.MatchedBy(func(args []interface{}) bool {
		if len(args) != 6 {
			return false
		}
		// Check that we have the right keys and expected values
		return args[0] == LogKeyChatID &&
			args[1] == testGroupChatID1 &&
			args[2] == LogKeyFromUser &&
			args[3] == testUserName &&
			args[4] == LogKeyError &&
			args[5] != nil // Any error is fine
	})).Return(nil)
	// Create bot and test update
	bot := createTestBot(mockBotAPI, mockSplitter, mockLogger)
	update := createMessageUpdate(testGroupChatID1, testUserName, "hello without prefix")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_ParseError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testGroupChatID2, testUserName, "hello")
	// Alternative approach for parse-error: set up the expectation directly
	mockLogger.On("Error", "Parse error", mock.MatchedBy(func(args []interface{}) bool {
		if len(args) != 6 {
			return false
		}
		// Check that we have the right keys and expected values
		return args[0] == LogKeyChatID &&
			args[1] == testGroupChatID2 &&
			args[2] == LogKeyFromUser &&
			args[3] == testUserName &&
			args[4] == LogKeyError &&
			args[5] != nil // Any error is fine
	})).Return(nil)
	// Create bot and test update
	bot := createCustomBot(mockBotAPI, mockSplitter, mockLogger, true, []string{})
	update := createMessageUpdate(testGroupChatID2, testUserName, "hello")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_AskError_WithErrorMessage(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testPrivateChatID, testUserName, "test message")
	mockLogger.On("Error", "Assistant error", []interface{}{
		LogKeyChatID, testPrivateChatID, LogKeyFromUser, testUserName, LogKeyError, assert.AnError,
	}).Return(nil)
	mockAssistant.On("Ask", testUserName, "test message").Return("", assert.AnError)
	setupMockForErrorResponse(mockBotAPI, ErrAssistantResponse)
	// Create bot and test update
	bot := createTestBot(mockBotAPI, mockSplitter, mockLogger)
	update := createMessageUpdate(testPrivateChatID, testUserName, "test message")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_SplitterError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testPrivateChatID, testUserName, "test message")
	mockLogger.On("Error", "Splitter error", []interface{}{
		LogKeyChatID, testPrivateChatID, LogKeyFromUser, testUserName, LogKeyError, assert.AnError,
		"text", "response message", "chunks", []string{},
	}).Return(nil)
	mockAssistant.On("Ask", testUserName, "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{}, assert.AnError)
	setupMockForErrorResponse(mockBotAPI, ErrSplittingResponse) // Fix: Changed from ErrSplitterResponse to ErrSplittingResponse
	// Create bot and test update
	bot := createTestBot(mockBotAPI, mockSplitter, mockLogger)
	update := createMessageUpdate(testPrivateChatID, testUserName, "test message")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_SendError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testPrivateChatID, testUserName, "test message")
	setupLoggerExpectations(mockLogger, "outgoing", testPrivateChatID, testUserName, "response message")
	mockLogger.On("Error", "Send error", []interface{}{
		LogKeyChatID, testPrivateChatID, LogKeyFromUser, testUserName, LogKeyError, assert.AnError,
	}).Return(nil)
	mockLogger.On("Error", "Failed to send error message", mock.MatchedBy(func(args []interface{}) bool {
		// Just check that we have the right keys but allow any error value
		if len(args) != 4 {
			return false
		}
		if args[0] != "chat_id" || args[2] != "error" {
			return false
		}
		// Check that chat_id is 11111
		chatID, ok := args[1].(int64)
		if !ok || chatID != 11111 {
			return false
		}
		// Don't check the specific error, just make sure there is one
		_, ok = args[3].(error)
		return ok
	})).Return(nil)
	mockAssistant.On("Ask", testUserName, "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)
	// Create bot and test update
	bot := createTestBot(mockBotAPI, mockSplitter, mockLogger)
	update := createMessageUpdate(testPrivateChatID, testUserName, "test message")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

// Tests for callback handling (Show More button)
func TestBot_handleCallbackQuery(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Create a bot with pending chunks
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	// Set up pending chunks for test
	bot.chunksMutex.Lock()
	bot.pendingChunks["12345:testuser"] = &ChunkQueue{
		Chunks:     []string{"chunk 1", "chunk 2", "chunk 3"},
		Position:   1, // Already sent chunk 1
		OriginalID: 100,
	}
	bot.chunksMutex.Unlock()
	// Mock expectations
	// 1. Send the next chunk (chunk 2)
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "chunk 2" && msg.ReplyMarkup != nil
	})).Return(tgbotapi.Message{}, nil).Once()
	// 2. Send callback acknowledgement
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		_, ok := c.(tgbotapi.CallbackConfig)
		return ok
	})).Return(tgbotapi.Message{}, nil).Once()
	// Create callback query
	query := &tgbotapi.CallbackQuery{
		ID: "callback123",
		From: &tgbotapi.User{
			UserName: "testuser",
		},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 12345,
			},
		},
		Data: "show_more",
	}
	// Handle the callback
	bot.handleCallbackQuery(query)
	// Verify expectations
	mockBotAPI.AssertExpectations(t)
	// Check position was incremented
	bot.chunksMutex.RLock()
	position := bot.pendingChunks["12345:testuser"].Position
	bot.chunksMutex.RUnlock()
	assert.Equal(t, 2, position)
}

func TestBot_handleCallbackQuery_SendError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Create a bot with pending chunks
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	// Set up pending chunks for test
	bot.chunksMutex.Lock()
	bot.pendingChunks["12345:testuser"] = &ChunkQueue{
		Chunks:     []string{"chunk 1", "chunk 2", "chunk 3"},
		Position:   1, // Already sent chunk 1
		OriginalID: 100,
	}
	bot.chunksMutex.Unlock()
	// Expect error when sending chunk
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "chunk 2"
	})).Return(tgbotapi.Message{}, fmt.Errorf("send error")).Once()
	// Expect error log - Fix: Use a matcher function instead of AnythingOfType
	mockLogger.On("Error", "Failed to send next chunk", mock.MatchedBy(func(args []interface{}) bool {
		// Just check that we have the right keys
		if len(args) != 4 {
			return false
		}
		if args[0] != "error" || args[2] != "chat_id" {
			return false
		}
		// Check that chat_id is 12345
		chatID, ok := args[3].(int64)
		if !ok || chatID != 12345 {
			return false
		}
		// Don't check the specific error, just make sure there is one
		_, ok = args[1].(error)
		return ok
	})).Return(nil)
	// Expect error message to user
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && strings.Contains(msg.Text, "Sorry, I couldn't load the next part")
	})).Return(tgbotapi.Message{}, nil).Once()
	// Expect callback response
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		cb, ok := c.(tgbotapi.CallbackConfig)
		return ok && strings.Contains(cb.Text, "Failed to load more content")
	})).Return(tgbotapi.Message{}, nil).Once()
	// Create callback query
	query := &tgbotapi.CallbackQuery{
		ID: "callback123",
		From: &tgbotapi.User{
			UserName: "testuser",
		},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 12345,
			},
		},
		Data: "show_more",
	}
	// Handle the callback
	bot.handleCallbackQuery(query)
	// Verify expectations
	mockBotAPI.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Tests for error message handling
func TestBot_sendErrorMessage(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Expect an error message to be sent
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "Test error message" && msg.ReplyToMessageID == 123
	})).Return(tgbotapi.Message{}, nil).Once()
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	bot.sendErrorMessage(12345, 123, "Test error message")
	mockBotAPI.AssertExpectations(t)
}

func TestBot_sendErrorMessage_Failure(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	// Expect an error when sending the message
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, fmt.Errorf("failed to send message")).Once()
	// Fix: Use a custom matcher for the error parameter
	mockLogger.On("Error", "Failed to send error message", mock.MatchedBy(func(args []interface{}) bool {
		// Just check that we have the right keys but allow any error value
		if len(args) != 4 {
			return false
		}
		if args[0] != "chat_id" || args[2] != "error" {
			return false
		}
		// Check that chat_id is 12345
		chatID, ok := args[1].(int64)
		if !ok || chatID != 12345 {
			return false
		}
		// Don't check the specific error, just make sure there is one
		_, ok = args[3].(error)
		return ok
	})).Return(nil)
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	bot.sendErrorMessage(12345, 123, "Test error message")
	mockBotAPI.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Group message tests
func TestBot_HandleMessages(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockUpdates := make(chan tgbotapi.Update)
	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser", "text", "test message",
	}).Return(nil)

	// Fix: Use reply_to_message_id of 0 to match actual code behavior
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(11111), "reply_to_message_id", 0, "text", "response message", "chunks_count", 1,
	}).Return(nil)

	mockBotAPI.On("GetUpdatesChan", mock.Anything).Return((tgbotapi.UpdatesChannel)(mockUpdates))
	mockAssistant.On("Ask", testUserName, "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"testuser"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bot.HandleMessages(mockAssistant)
	}()

	// Send a mock update with a non-group chat ID
	// Note: We're creating an Update with a Message that has MessageID 0 (default value)
	mockUpdates <- tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 11111},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "test message",
			// MessageID is not explicitly set, so it defaults to 0
		},
	}

	// Close the mock updates channel to stop the goroutine
	close(mockUpdates)
	wg.Wait()

	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_NilMessage(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	update := tgbotapi.Update{
		Message: nil,
	}
	config := BotConfig{
		Name:        "testbot",
		UserChats:   []string{"allowed_user"},
		GroupChats:  []int64{12345, 67890},
		UseShowMore: true,
	}
	bot := NewBot(mockBotAPI, config, mockSplitter, mockLogger)
	bot.handleUpdate(update, mockAssistant)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}
