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
		{"allowed user can send message in any chat", 11111, "allowed_user", "test message", "test message", false}, // Even in disallowed chat
		{"allowed user with bot prefix", 67890, "allowed_user", "testbot hello", "testbot hello", false},
		{"non-allowed user in allowed chat", 67890, "other_user", "test message", "", true},
		{"non-allowed user with prefix", 67890, "other_user", "testbot hello", "hello", false},
		{"non-allowed user in disallowed chat", 11111, "other_user", "test message", "", true},
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
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	bot := NewBot(mockBotAPI, "testbot", []string{"allowed_user"}, []int64{12345, 67890}, mockSplitter, mockLogger)
	err := bot.send(12345, 1, []string{"response message"})

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
		"chat_id", int64(12345), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(12345), "reply_to_message_id", 0, "text", "response message",
	}).Return(nil)

	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
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

func TestBot_handleUpdate_ParseError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	mockSplitter := new(MockSplitter)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", "Incoming message", []interface{}{
		"chat_id", int64(67890), "from_user", "testuser", "text", "hello",
	}).Return(nil)
	mockLogger.On("Error", "Parse error", []interface{}{
		"chat_id", int64(67890), "from_user", "testuser", "error", fmt.Errorf("cannot process request"),
	}).Return(nil)

	// Note: We DON'T set up any expectations for mockAssistant.Ask
	// because it should not be called after a parse error

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
		"chat_id", int64(12345), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Error", "Assistant error", []interface{}{
		"chat_id", int64(12345), "from_user", "testuser", "error", assert.AnError,
	}).Return(nil)

	mockAssistant.On("Ask", "testuser", "test message").Return("", assert.AnError)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
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
		"chat_id", int64(12345), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(12345), "reply_to_message_id", 0, "text", "response message",
	}).Return(nil)
	mockLogger.On("Error", "Send error", []interface{}{
		"chat_id", int64(12345), "from_user", "testuser", "error", assert.AnError,
	}).Return(nil)

	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)
	mockSplitter.On("Split", "response message").Return([]string{"response message"}, nil)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
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
		"chat_id", int64(12345), "from_user", "testuser", "text", "test message",
	}).Return(nil)
	mockLogger.On("Info", "Outgoing message", []interface{}{
		"chat_id", int64(12345), "reply_to_message_id", 0, "text", "response message",
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

	// Send a mock update
	mockUpdates <- tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
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
