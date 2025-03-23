package bot

import (
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

func TestBot_parse(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})

	tests := []struct {
		chatID int64
		txt    string
		want   string
		err    bool
	}{
		{12345, "test message", "test message", false},
		{67890, "testbot hello", "hello", false},
		{67890, "hello", "", true},
		{11111, "testbot hello", "", true},
	}

	for _, tt := range tests {
		got, err := bot.parse(tt.chatID, tt.txt)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		}
	}
}

func TestSend(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})
	err := bot.send(12345, 1, "response message")
	assert.NoError(t, err)

	mockBotAPI.AssertExpectations(t)
}

func TestSend_Error(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)

	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})
	err := bot.send(12345, 1, "response message")
	assert.Error(t, err)

	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_Success(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})

	// Mock the Ask method
	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)

	// Mock the Send method
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "test message",
		},
	}

	bot.handleUpdate(update, mockAssistant)

	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_ParseError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 67890},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "hello",
		},
	}

	bot.handleUpdate(update, mockAssistant)

	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_HandleError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})

	// Mock the Ask method
	mockAssistant.On("Ask", "testuser", "test message").Return("", assert.AnError)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "test message",
		},
	}

	bot.handleUpdate(update, mockAssistant)

	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_SendError(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})

	// Mock the Ask method
	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)

	// Mock the Send method
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, assert.AnError)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
			From: &tgbotapi.User{UserName: "testuser"},
			Text: "test message",
		},
	}

	bot.handleUpdate(update, mockAssistant)

	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_handleUpdate_NilMessage(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})

	update := tgbotapi.Update{
		Message: nil,
	}

	bot.handleUpdate(update, mockAssistant)

	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}

func TestBot_HandleMessages(t *testing.T) {
	mockBotAPI := new(MockBotAPI)
	mockAssistant := new(MockAssistant)
	bot := NewBot(mockBotAPI, "testbot", 12345, []int64{12345, 67890})

	// Mock the GetUpdatesChan method
	mockUpdates := make(chan tgbotapi.Update)
	mockBotAPI.On("GetUpdatesChan", mock.Anything).Return((tgbotapi.UpdatesChannel)(mockUpdates))

	// Mock the Ask method
	mockAssistant.On("Ask", "testuser", "test message").Return("response message", nil)

	// Mock the Send method
	mockBotAPI.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

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

	mockAssistant.AssertExpectations(t)
	mockBotAPI.AssertExpectations(t)
}
