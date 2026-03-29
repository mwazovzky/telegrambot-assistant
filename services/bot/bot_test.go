// Package bot contains the core Telegram bot functionality and tests.
// bot_test contains tests for the Telegram bot functionality including message handling,
// chunk storage, and user interaction flows.
package bot

import (
	"context"
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
	testUserID        = int64(99999)
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

func (m *MockLogger) Info(_ context.Context, message string, keyValues ...interface{}) error {
	args := m.Called(message, keyValues)
	return args.Error(0)
}

func (m *MockLogger) Error(_ context.Context, message string, keyValues ...interface{}) error {
	args := m.Called(message, keyValues)
	return args.Error(0)
}

func (m *MockLogger) Debug(_ context.Context, message string, keyValues ...interface{}) error {
	args := m.Called(message, keyValues)
	return args.Error(0)
}

// Add a MockChunkStorage implementation
type MockChunkStorage struct {
	mock.Mock
}

func (m *MockChunkStorage) StoreChunks(chatID int64, username string, messageID int, chunks []string) {
	m.Called(chatID, username, messageID, chunks)
}

func (m *MockChunkStorage) GetNextChunk(chatID int64, username string) (chunk string, originalID int, hasMore bool, exists bool) {
	args := m.Called(chatID, username)
	return args.String(0), args.Int(1), args.Bool(2), args.Bool(3)
}

func (m *MockChunkStorage) HasChunks(chatID int64, username string) bool {
	args := m.Called(chatID, username)
	return args.Bool(0)
}

func (m *MockChunkStorage) Clear(chatID int64, username string) {
	m.Called(chatID, username)
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

// Helper to create a test update with a message
func createMessageUpdate(chatID int64, username, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: chatID},
			From:      &tgbotapi.User{ID: testUserID, UserName: username},
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
			LogKeyChatID, chatID, LogKeyFromUser, username,
		}).Return(nil)
	case "outgoing":
		mockLogger.On("Info", "Outgoing message", []interface{}{
			LogKeyChatID, chatID, LogKeyReplyToMsgID, 123, LogKeyChunksCount, 1,
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

// Consider using a test helper function to avoid config duplication
func createTestConfig() BotConfig {
	return BotConfig{
		Name:        testBotName,
		UserChats:   []string{testAllowedUser, testUserName},
		GroupChats:  []int64{testGroupChatID1, testGroupChatID2},
		UseShowMore: true,
	}
}

// Tests for message parsing
func TestBot_parse(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	config := createTestConfig()
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
	mockChunkStorage := new(MockChunkStorage)

	// Expect only the first message to be sent (with button)
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "message part 1" && msg.ReplyMarkup != nil
	})).Return(tgbotapi.Message{}, nil).Once()

	// Expect chunk storage to be called
	mockChunkStorage.On("StoreChunks", int64(12345), "testuser", 1,
		[]string{"message part 1", "message part 2"}).Return()

	config := createTestConfig()

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
	err := bot.send(12345, "testuser", 1, []string{"message part 1", "message part 2"})

	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	mockChunkStorage.AssertExpectations(t)
}

func TestSend_WithoutShowMore(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockChunkStorage := new(MockChunkStorage)

	// Create a more precise matcher for messages without buttons
	// Check that ReplyMarkup is explicitly nil
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "message part 1" && msg.ReplyMarkup == nil
	})).Return(tgbotapi.Message{}, nil).Once()

	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "message part 2" && msg.ReplyMarkup == nil
	})).Return(tgbotapi.Message{}, nil).Once()

	// We should not store chunks when ShowMore is disabled
	// No expectation for mockChunkStorage.StoreChunks()

	// Create a config with ShowMore explicitly set to false
	config := BotConfig{
		Name:        testBotName,
		UserChats:   []string{testAllowedUser},
		GroupChats:  []int64{testGroupChatID1, testGroupChatID2},
		UseShowMore: false, // Make sure this is false
	}

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
	err := bot.send(12345, "testuser", 1, []string{"message part 1", "message part 2"})

	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
	mockChunkStorage.AssertExpectations(t)
}

func TestSend_SingleChunk(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockChunkStorage := new(MockChunkStorage)
	// Expect a single message without button
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "single message" && msg.ReplyMarkup == nil
	})).Return(tgbotapi.Message{}, nil).Once()
	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
	err := bot.send(12345, "testuser", 1, []string{"single message"})
	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
}

func TestSend_Error(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockChunkStorage := new(MockChunkStorage)
	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
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
	mockChunkStorage := new(MockChunkStorage)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testPrivateChatID, testUserName, "test message")
	setupLoggerExpectations(mockLogger, "outgoing", testPrivateChatID, testUserName, "response message")
	mockAssistant.On("Ask", fmt.Sprintf("%d", testUserID), "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	// Fix: Define the config variable here
	config := createTestConfig()

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
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
	mockChunkStorage := new(MockChunkStorage)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testGroupChatID2, testUserName, "testbot hello")
	setupLoggerExpectations(mockLogger, "outgoing", testGroupChatID2, testUserName, "response message")
	mockAssistant.On("Ask", fmt.Sprintf("%d", testUserID), "hello").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	// Fix: Define the config variable here
	config := createTestConfig()

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
	update := createMessageUpdate(testGroupChatID2, testUserName, "testbot hello")
	// Handle the update
	bot.handleUpdate(update, mockAssistant)
	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

// Fix the remaining test functions that use the undeclared 'config' variable
func TestBot_handleUpdate_ParseError_GroupChatNoPrefix(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockChunkStorage := new(MockChunkStorage)
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

	// Fix: Define the config variable here
	config := createTestConfig()

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
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
	mockChunkStorage := new(MockChunkStorage)
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

	// Fix: Define the config variable here
	config := BotConfig{
		Name:        testBotName,
		UserChats:   []string{}, // Empty user list to test parse error
		GroupChats:  []int64{testGroupChatID1, testGroupChatID2},
		UseShowMore: true,
	}

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
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
	mockChunkStorage := new(MockChunkStorage)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testPrivateChatID, testUserName, "test message")
	mockLogger.On("Error", "Assistant error", []interface{}{
		LogKeyChatID, testPrivateChatID, LogKeyFromUser, testUserName, LogKeyError, assert.AnError,
	}).Return(nil)
	mockAssistant.On("Ask", fmt.Sprintf("%d", testUserID), "test message").Return("", assert.AnError)
	setupMockForErrorResponse(mockBotAPI, ErrAssistantResponse)
	// Create bot and test update

	// Fix: Define the config variable here
	config := createTestConfig()

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
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
	mockChunkStorage := new(MockChunkStorage)
	// Set up expectations
	setupLoggerExpectations(mockLogger, "incoming", testPrivateChatID, testUserName, "test message")
	mockLogger.On("Error", "Splitter error", []interface{}{
		LogKeyChatID, testPrivateChatID, LogKeyFromUser, testUserName, LogKeyError, assert.AnError,
		"text", "response message", "chunks", []string{},
	}).Return(nil)
	mockAssistant.On("Ask", fmt.Sprintf("%d", testUserID), "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{}, assert.AnError)
	setupMockForErrorResponse(mockBotAPI, ErrSplittingResponse) // Fix: Changed from ErrSplitterResponse to ErrSplittingResponse
	// Create bot and test update

	// Fix: Define the config variable here
	config := createTestConfig()

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
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
	mockChunkStorage := new(MockChunkStorage)
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
	mockAssistant.On("Ask", fmt.Sprintf("%d", testUserID), "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)
	// Create bot and test update

	// Fix: Define the config variable here
	config := createTestConfig()

	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
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
	mockChunkStorage := new(MockChunkStorage)

	// Mock the chunk storage to return the next chunk
	mockChunkStorage.On("GetNextChunk", int64(12345), "testuser").
		Return("chunk 2", 100, true, true)

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

	// Create bot and callback query
	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)

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

	// Verify expectations - Clear should NOT be called when hasMore=true
	mockBotAPI.AssertExpectations(t)
	mockChunkStorage.AssertExpectations(t)
	mockChunkStorage.AssertNotCalled(t, "Clear", mock.Anything, mock.Anything)
}

func TestBot_handleCallbackQuery_LastChunk(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockChunkStorage := new(MockChunkStorage)

	// Mock the chunk storage to return the last chunk (hasMore=false)
	mockChunkStorage.On("GetNextChunk", int64(12345), "testuser").
		Return("chunk 3", 100, false, true)

	// Expect Clear to be called after delivering the last chunk
	mockChunkStorage.On("Clear", int64(12345), "testuser").Return()

	// 1. Send the last chunk (no button)
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "chunk 3" && msg.ReplyMarkup == nil
	})).Return(tgbotapi.Message{}, nil).Once()

	// 2. Send callback acknowledgement
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		_, ok := c.(tgbotapi.CallbackConfig)
		return ok
	})).Return(tgbotapi.Message{}, nil).Once()

	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)

	query := &tgbotapi.CallbackQuery{
		ID:   "callback456",
		From: &tgbotapi.User{UserName: "testuser"},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
		},
		Data: "show_more",
	}

	bot.handleCallbackQuery(query)

	// Verify Clear was called
	mockBotAPI.AssertExpectations(t)
	mockChunkStorage.AssertExpectations(t)
}

func TestBot_handleCallbackQuery_SendError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockChunkStorage := new(MockChunkStorage)

	// Mock the chunk storage to return the next chunk
	mockChunkStorage.On("GetNextChunk", int64(12345), "testuser").
		Return("chunk 2", 100, true, true)

	// Expect error when sending chunk
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "chunk 2"
	})).Return(tgbotapi.Message{}, fmt.Errorf("send error")).Once()

	// Expect error log
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

	// Create bot with mocked dependencies
	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)

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
	mockChunkStorage.AssertExpectations(t)
}

// Tests for error message handling
func TestBot_sendErrorMessage(t *testing.T) {
	// Replace the manual setup with the helper function
	// Fix: Remove unused variables from the return values
	bot, mockBotAPI, _, _, _, _ := setupTestBot(t)

	// Expect an error message to be sent
	mockBotAPI.On("Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "Test error message" && msg.ReplyToMessageID == 123
	})).Return(tgbotapi.Message{}, nil).Once()

	bot.sendErrorMessage(12345, 123, "Test error message")
	mockBotAPI.AssertExpectations(t)
}

func TestBot_sendErrorMessage_Failure(t *testing.T) {
	// Replace the manual setup with the helper function
	// Fix: Remove unused variables from the return values, keep mockLogger
	bot, mockBotAPI, _, _, mockLogger, _ := setupTestBot(t)

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
	mockChunkStorage := new(MockChunkStorage)
	mockUpdates := make(chan tgbotapi.Update)
	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser",
	}).Return(nil)

	// Fix: Use reply_to_message_id of 0 to match actual code behavior
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(11111), "reply_to_message_id", 0, "chunks_count", 1,
	}).Return(nil)

	mockBotAPI.On("GetUpdatesChan", mock.Anything).Return((tgbotapi.UpdatesChannel)(mockUpdates))
	mockAssistant.On("Ask", fmt.Sprintf("%d", testUserID), "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)

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
			From: &tgbotapi.User{ID: testUserID, UserName: "testuser"},
			Text: "test message",
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
	mockChunkStorage := new(MockChunkStorage)
	update := tgbotapi.Update{
		Message: nil,
	}
	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)
	bot.handleUpdate(update, mockAssistant)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

// Additional test for the Clear method in the InMemoryChunkStorage
func TestInMemoryChunkStorage_Clear(t *testing.T) {
	storage := NewInMemoryChunkStorage()

	// Store chunks
	storage.StoreChunks(12345, "testuser", 100, []string{"chunk1", "chunk2"})

	// Verify chunks were stored
	assert.True(t, storage.HasChunks(12345, "testuser"))

	// Clear chunks
	storage.Clear(12345, "testuser")

	// Verify chunks were cleared
	assert.False(t, storage.HasChunks(12345, "testuser"))
}

// Common setup function for tests with similar initialization
func setupTestBot(t *testing.T) (*Bot, *MockBotAPI, *MockAssistant, *MockSplitter, *MockLogger, *MockChunkStorage) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockChunkStorage := new(MockChunkStorage)

	config := createTestConfig()
	bot := NewBotWithChunkStorage(mockBotAPI, config, mockSplitter, mockLogger, mockChunkStorage)

	return bot, mockBotAPI, mockAssistant, mockSplitter, mockLogger, mockChunkStorage
}
