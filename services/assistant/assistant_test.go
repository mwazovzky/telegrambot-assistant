package assistant

import (
	"testing"

	"github.com/mwazovzky/assistant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOpenAiClient is a mock implementation of the OpenAiClient interface
type MockOpenAiClient struct {
	mock.Mock
}

func (m *MockOpenAiClient) GetThread(username string) ([]assistant.Message, error) {
	args := m.Called(username)
	return args.Get(0).([]assistant.Message), args.Error(1)
}

func (m *MockOpenAiClient) CreateThread(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MockOpenAiClient) Post(username, req string) (string, error) {
	args := m.Called(username, req)
	return args.String(0), args.Error(1)
}

func TestNewAssistant(t *testing.T) {
	mockClient := new(MockOpenAiClient)
	openAi := NewAssistant(mockClient)
	assert.NotNil(t, openAi)
	assert.Equal(t, mockClient, openAi.client)
}

func TestAssistant_Ask(t *testing.T) {
	mockClient := new(MockOpenAiClient)
	openAi := NewAssistant(mockClient)

	// Mock the GetThread method
	mockClient.On("GetThread", "testuser").Return([]assistant.Message{}, assert.AnError)

	// Mock the CreateThread method
	mockClient.On("CreateThread", "testuser").Return(nil)

	// Mock the Post method
	mockClient.On("Post", "testuser", "test message").Return("response message", nil)

	res, err := openAi.Ask("test message", "testuser")
	assert.NoError(t, err)
	assert.Equal(t, "response message", res)

	mockClient.AssertExpectations(t)
}

func TestAssistant_Ask_CreateThreadError(t *testing.T) {
	mockClient := new(MockOpenAiClient)
	openAi := NewAssistant(mockClient)

	// Mock the GetThread method
	mockClient.On("GetThread", "testuser").Return([]assistant.Message{}, assert.AnError)

	// Mock the CreateThread method
	mockClient.On("CreateThread", "testuser").Return(assert.AnError)

	res, err := openAi.Ask("test message", "testuser")
	assert.Error(t, err)
	assert.Equal(t, "", res)

	mockClient.AssertExpectations(t)
}

func TestAssistant_Ask_PostError(t *testing.T) {
	mockClient := new(MockOpenAiClient)
	openAi := NewAssistant(mockClient)

	// Mock the GetThread method
	mockClient.On("GetThread", "testuser").Return([]assistant.Message{}, assert.AnError)

	// Mock the CreateThread method
	mockClient.On("CreateThread", "testuser").Return(nil)

	// Mock the Post method
	mockClient.On("Post", "testuser", "test message").Return("", assert.AnError)

	res, err := openAi.Ask("test message", "testuser")
	assert.Error(t, err)
	assert.Equal(t, "", res)

	mockClient.AssertExpectations(t)
}
