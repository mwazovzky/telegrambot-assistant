package setup

import (
	"testing"

	"telegrambot-assistant/services/config"
	"telegrambot-assistant/services/repository"
	"telegrambot-assistant/services/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBotAPI struct {
	mock.Mock
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
		ApiToken: "testToken",
		BotName:  "testBot",
		ChatID:   12345,
	}
	bot, err := InitBot(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, bot)
}

func TestInitRepository(t *testing.T) {
	redisService := new(storage.RedisService)
	repo := InitRepository(redisService)
	assert.NotNil(t, repo)
}

func TestInitOpenAiClient(t *testing.T) {
	cfg := config.OpenAIConfig{
		ApiUrl: "https://api.openai.com",
		ApiKey: "testKey",
		Model:  "testModel",
		Name:   "testName",
		Role:   "testRole",
	}
	tr := new(repository.ThreadRepository)
	client := InitOpenAiClient(cfg, tr)
	assert.NotNil(t, client)
}
