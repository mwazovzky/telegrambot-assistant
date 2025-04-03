package bot

import (
	"fmt"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBotAPI is a mock implementation of the BotAPI interface
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

// MockAssistant is a mock implementation of the Assistant interface
type MockAssistant struct {
	mock.Mock
}

func (m *MockAssistant) Ask(req string, username string) (string, error) {
	args := m.Called(req, username)
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

func TestBot_parse(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	bot := NewBot(mockBotAPI, "testbot", []string{"allowed_user"}, []int64{12345, 67890}, mockSplitter, mockLogger)

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

func TestSend(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil).Times(2) // Should be called for each chunk
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	bot := NewBot(mockBotAPI, "testbot", []string{"allowed_user"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	err := bot.send(12345, 1, []string{"message part 1", "message part 2"})

	assert.NoError(t, err)
	mockBotAPI.AssertExpectations(t)
}

func TestSend_Error(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	bot := NewBot(mockBotAPI, "testbot", []string{"allowed_user"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	err := bot.send(12345, 1, []string{"response message"})

	assert.Error(t, err)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_Success(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(11111), "reply_to_message_id", 0, "text", "response message",
	}).Return(nil)

	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	// Use chat ID that's NOT in groupChats to ensure it's treated as a private chat
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 11111},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "test message",
		},
	}

	bot := NewBot(mockBotAPI, "testbot", []string{"testuser"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	bot.handleUpdate(update, mockAssistant)

	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_Success_GroupChat(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(67890), "from_user", "testuser", "text", "testbot hello",
	}).Return(nil)
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(67890), "reply_to_message_id", 0, "text", "response message",
	}).Return(nil)

	mockAssistant.On("Ask", "testuser", "hello").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	// Use a chat ID that is in groupChats and include the bot name prefix
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 67890},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "testbot hello",
		},
	}

	bot := NewBot(mockBotAPI, "testbot", []string{"testuser"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	bot.handleUpdate(update, mockAssistant)

	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_ParseError_GroupChatNoPrefix(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(12345), "from_user", "testuser", "text", "hello without prefix",
	}).Return(nil)
	mockLogger.On("Error", "Parse error", []interface{}{
		"chat_id", int64(12345), "from_user", "testuser", "error", fmt.Errorf("cannot process chat message"),
	}).Return(nil)

	// Message in group chat without bot prefix
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345}, // Group chat
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "hello without prefix",
		},
	}

	bot := NewBot(mockBotAPI, "testbot", []string{"testuser"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	bot.handleUpdate(update, mockAssistant)

	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_ParseError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(67890), "from_user", "testuser", "text", "hello",
	}).Return(nil)
	mockLogger.On("Error", "Parse error", []interface{}{
		"chat_id", int64(67890), "from_user", "testuser", "error", fmt.Errorf("cannot process chat message"),
	}).Return(nil)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 67890},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "hello",
		},
	}

	bot := NewBot(mockBotAPI, "testbot", []string{}, []int64{12345, 67890}, mockSplitter, mockLogger) // Empty userChats list to force parse error
	bot.handleUpdate(update, mockAssistant)

	mockLogger.AssertExpectations(t)
	// Don't assert mockAssistant expectations as it shouldn't be called
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_AskError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Error", "Assistant error", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser", "error", assert.AnError,
	}).Return(nil)

	mockAssistant.On("Ask", "testuser", "test message").Return("", assert.AnError)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 11111}, // Use non-group chat ID
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "test message",
		},
	}

	bot := NewBot(mockBotAPI, "testbot", []string{"testuser"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	bot.handleUpdate(update, mockAssistant)

	mockLogger.AssertExpectations(t)
	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_SendError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(11111), "reply_to_message_id", 0, "text", "response message",
	}).Return(nil)
	mockLogger.On("Error", "Send error", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser", "error", assert.AnError,
	}).Return(nil)

	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 11111}, // Use non-group chat ID
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "test message",
		},
	}

	bot := NewBot(mockBotAPI, "testbot", []string{"testuser"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	bot.handleUpdate(update, mockAssistant)

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

	bot := NewBot(mockBotAPI, "testbot", []string{"allowed_user"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	bot.handleUpdate(update, mockAssistant)

	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_HandleMessages(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)
	mockUpdates := make(chan tgbotapi.Update)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(11111), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(11111), "reply_to_message_id", 0, "text", "response message",
	}).Return(nil)

	mockBotAPI.On("GetUpdatesChan", mock.Anything).Return((tgbotapi.UpdatesChannel)(mockUpdates))
	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	bot := NewBot(mockBotAPI, "testbot", []string{"testuser"}, []int64{12345, 67890}, mockSplitter, mockLogger)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		bot.HandleMessages(mockAssistant)
	}()

	// Send a mock update with a non-group chat ID
	mockUpdates <- tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 11111},
			From: &tgbotapi.User{UserName: "testuser"},
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
