package responsestore

import (
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
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

func TestRedisStore_SetResponseID(t *testing.T) {
	mockClient := new(MockCacheClient)
	mockClient.On("Set", "user1", "resp_abc123").Return(nil)

	store := NewRedisStore(mockClient)
	err := store.SetResponseID("user1", "resp_abc123")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRedisStore_GetResponseID(t *testing.T) {
	mockClient := new(MockCacheClient)
	mockClient.On("Get", "user1").Return("resp_abc123", nil)

	store := NewRedisStore(mockClient)
	responseID, err := store.GetResponseID("user1")

	assert.NoError(t, err)
	assert.Equal(t, "resp_abc123", responseID)
	mockClient.AssertExpectations(t)
}

func TestRedisStore_GetResponseID_MissingKey(t *testing.T) {
	mockClient := new(MockCacheClient)
	mockClient.On("Get", "unknown").Return("", redis.Nil)

	store := NewRedisStore(mockClient)
	responseID, err := store.GetResponseID("unknown")

	assert.NoError(t, err)
	assert.Equal(t, "", responseID)
	mockClient.AssertExpectations(t)
}

func TestRedisStore_GetResponseID_BackendError(t *testing.T) {
	mockClient := new(MockCacheClient)
	mockClient.On("Get", "user1").Return("", fmt.Errorf("connection refused"))

	store := NewRedisStore(mockClient)
	responseID, err := store.GetResponseID("user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Equal(t, "", responseID)
	mockClient.AssertExpectations(t)
}
