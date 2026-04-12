package setup

import (
	"context"
	"testing"

	"telegrambot-assistant/services/config"
	"telegrambot-assistant/services/responsestore"
	"telegrambot-assistant/services/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBotAPI struct {
	mock.Mock
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

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	return args.Get(0).(tgbotapi.Message), args.Error(1)
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	args := m.Called(config)
	return args.Get(0).(tgbotapi.UpdatesChannel)
}

func TestInitBot(t *testing.T) {
	// Override the newBotAPI function
	originalNewBotAPI := newBotAPI
	defer func() { newBotAPI = originalNewBotAPI }()
	newBotAPI = func(token string) (*tgbotapi.BotAPI, error) {
		return &tgbotapi.BotAPI{}, nil
	}

	cfg := config.TelegramConfig{
		ApiToken:     "testToken",
		BotName:      "testBot",
		Users:        []string{"user1", "user2"}, // Added Users field
		Chats:        []int64{12345, 67890},      // Updated Chats field
		MessageLimit: 4096,                       // Added MessageLimit
	}
	mockLogger := new(MockLogger)
	mockLogger.On("Info", "TelegramBot: authorized on account", []interface{}{"username", ""}).Return(nil)
	bot, err := InitBot(cfg, mockLogger)
	assert.NoError(t, err)
	assert.NotNil(t, bot)
	mockLogger.AssertExpectations(t)
}

func TestInitResponseStore(t *testing.T) {
	redisService := new(storage.RedisService)
	store := InitResponseStore(redisService)
	assert.NotNil(t, store)
}

func TestInitAssistant(t *testing.T) {
	cfg := config.OpenAIConfig{
		ApiKey: "testKey",
		Model:  "testModel",
		Name:   "testName",
		Role:   "testRole",
	}
	store := responsestore.NewInmemoryStore()
	mockLogger := new(MockLogger)
	assistant := InitAssistant(cfg, store, mockLogger)
	assert.NotNil(t, assistant)
}

