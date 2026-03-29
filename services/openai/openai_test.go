package openai

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResponseClient mocks the OpenAI Responses API client.
type MockResponseClient struct {
	mock.Mock
}

func (m *MockResponseClient) New(ctx context.Context, params responses.ResponseNewParams, opts ...option.RequestOption) (*responses.Response, error) {
	args := m.Called(ctx, params)
	resp, _ := args.Get(0).(*responses.Response)
	return resp, args.Error(1)
}

// MockResponseStore mocks the ResponseStore interface.
type MockResponseStore struct {
	mock.Mock
}

func (m *MockResponseStore) GetResponseID(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockResponseStore) SetResponseID(key string, responseID string) error {
	args := m.Called(key, responseID)
	return args.Error(0)
}

// MockLogger mocks the Logger interface.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(_ context.Context, message string, keyValues ...interface{}) error {
	args := m.Called(message, keyValues)
	return args.Error(0)
}

func TestAssistant_Ask_NewConversation(t *testing.T) {
	mockClient := new(MockResponseClient)
	mockStore := new(MockResponseStore)

	// No previous response ID — new conversation
	mockStore.On("GetResponseID", "user1").Return("", fmt.Errorf("not found"))

	// Expect API call without PreviousResponseID
	mockClient.On("New", mock.Anything, mock.MatchedBy(func(p responses.ResponseNewParams) bool {
		return !p.PreviousResponseID.Valid()
	})).Return(&responses.Response{
		ID: "resp_abc123",
	}, nil)

	// Expect response ID stored
	mockStore.On("SetResponseID", "user1", "resp_abc123").Return(nil)

	mockLogger := new(MockLogger)
	assistant := NewAssistant(mockClient, "gpt-4o-mini", "You are helpful.", mockStore, mockLogger, 30*time.Second)
	result, err := assistant.Ask("user1", "Hello")

	assert.NoError(t, err)
	assert.Equal(t, "", result) // OutputText() returns empty for response without output items
	mockClient.AssertExpectations(t)
	mockStore.AssertExpectations(t)
}

func TestAssistant_Ask_ContinuedConversation(t *testing.T) {
	mockClient := new(MockResponseClient)
	mockStore := new(MockResponseStore)

	// Previous response ID exists
	mockStore.On("GetResponseID", "user1").Return("resp_previous", nil)

	// Expect API call with PreviousResponseID set to the stored value
	mockClient.On("New", mock.Anything, mock.MatchedBy(func(p responses.ResponseNewParams) bool {
		return p.PreviousResponseID.Valid() && p.PreviousResponseID.Value == "resp_previous"
	})).Return(&responses.Response{
		ID: "resp_new",
	}, nil)

	mockStore.On("SetResponseID", "user1", "resp_new").Return(nil)

	mockLogger := new(MockLogger)
	assistant := NewAssistant(mockClient, "gpt-4o-mini", "You are helpful.", mockStore, mockLogger, 30*time.Second)
	_, err := assistant.Ask("user1", "Follow up question")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
	mockStore.AssertExpectations(t)
}

func TestAssistant_Ask_APIError(t *testing.T) {
	mockClient := new(MockResponseClient)
	mockStore := new(MockResponseStore)

	mockStore.On("GetResponseID", "user1").Return("", fmt.Errorf("not found"))

	mockClient.On("New", mock.Anything, mock.Anything).
		Return((*responses.Response)(nil), fmt.Errorf("API rate limit exceeded"))

	mockLogger := new(MockLogger)
	assistant := NewAssistant(mockClient, "gpt-4o-mini", "You are helpful.", mockStore, mockLogger, 30*time.Second)
	result, err := assistant.Ask("user1", "Hello")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openai response error")
	assert.Equal(t, "", result)
	mockStore.AssertNotCalled(t, "SetResponseID", mock.Anything, mock.Anything)
}

func TestAssistant_Ask_StoreError(t *testing.T) {
	mockClient := new(MockResponseClient)
	mockStore := new(MockResponseStore)

	mockStore.On("GetResponseID", "user1").Return("", fmt.Errorf("not found"))

	mockClient.On("New", mock.Anything, mock.Anything).
		Return(&responses.Response{ID: "resp_abc"}, nil)

	mockStore.On("SetResponseID", "user1", "resp_abc").
		Return(fmt.Errorf("redis connection failed"))

	mockLogger := new(MockLogger)
	mockLogger.On("Error", "Failed to store response ID", mock.Anything).Return(nil)

	assistant := NewAssistant(mockClient, "gpt-4o-mini", "You are helpful.", mockStore, mockLogger, 30*time.Second)
	result, err := assistant.Ask("user1", "Hello")

	// Store failure is non-fatal — response is still returned
	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
	assert.Equal(t, "", result) // OutputText() returns empty for response without output items
}
