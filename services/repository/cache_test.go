package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheClient) Set(key string, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func TestCacheRepository_SetResponseID(t *testing.T) {
	mockClient := new(MockCacheClient)
	mockClient.On("Set", "user1", "resp_abc123").Return(nil)

	repo := NewCachedRepository(mockClient)
	err := repo.SetResponseID("user1", "resp_abc123")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCacheRepository_GetResponseID(t *testing.T) {
	mockClient := new(MockCacheClient)
	mockClient.On("Get", "user1").Return("resp_abc123", nil)

	repo := NewCachedRepository(mockClient)
	responseID, err := repo.GetResponseID("user1")

	assert.NoError(t, err)
	assert.Equal(t, "resp_abc123", responseID)
	mockClient.AssertExpectations(t)
}

func TestCacheRepository_GetResponseID_NotFound(t *testing.T) {
	mockClient := new(MockCacheClient)
	mockClient.On("Get", "unknown").Return("", fmt.Errorf("key not found"))

	repo := NewCachedRepository(mockClient)
	responseID, err := repo.GetResponseID("unknown")

	assert.Error(t, err)
	assert.Equal(t, "", responseID)
	mockClient.AssertExpectations(t)
}
