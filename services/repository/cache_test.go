package repository

import (
	"testing"

	openai "github.com/mwazovzky/assistant"
	"github.com/stretchr/testify/mock"
)

const success = "\u2713"
const failure = "\u2717"

var getStr = "[{\"role\":\"system\",\"content\":\"Assistant\"}]"
var appendStr = "[{\"role\":\"system\",\"content\":\"Assistant\"},{\"role\":\"user\",\"content\":\"Question?\"}]"

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Set(key string, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func TestCacheRepository_CreateThread(t *testing.T) {
	mockRedis := new(MockRedisClient)
	value, _ := encode([]openai.Message{})
	mockRedis.On("Set", "myKey", value).Return(nil)

	r := NewCachedRepository(mockRedis)
	err := r.CreateThread("myKey")

	if err == nil {
		t.Logf("\t%s It should not return error.", success)
	} else {
		t.Fatalf("\t%s It should not return error, got [%s]", failure, err)
	}

	mockRedis.AssertExpectations(t)
}

func TestCacheRepository_GetMessages(t *testing.T) {
	mockRedis := new(MockRedisClient)
	mockRedis.On("Get", "myKey").Return(getStr, nil)

	r := NewCachedRepository(mockRedis)
	messages, err := r.GetMessages("myKey")

	if err == nil {
		t.Logf("\t%s It should not return error.", success)
	} else {
		t.Fatalf("\t%s It should not return error, got [%s]", failure, err)
	}

	len := len(messages)
	expectedLen := 1
	if len == expectedLen {
		t.Logf("\t%s It should contain %d message.", success, expectedLen)
	} else {
		t.Fatalf("\t%s It should contain %d message, got %d", failure, expectedLen, len)
	}

	msg := messages[0]
	expectedMsg := openai.Message{Role: "system", Content: "Assistant"}
	if compare(msg, expectedMsg) {
		t.Logf("\t%s It should contain message %v.", success, expectedMsg)
	} else {
		t.Fatalf("\t%s It should contain message %v, got %v", failure, expectedMsg, msg)
	}

	mockRedis.AssertExpectations(t)
}

func TestCacheRepository_AppendMessage(t *testing.T) {
	mockRedis := new(MockRedisClient)
	mockRedis.On("Get", "myKey").Return(getStr, nil)
	mockRedis.On("Set", "myKey", appendStr).Return(nil)

	msg := openai.Message{Role: "user", Content: "Question?"}
	r := NewCachedRepository(mockRedis)
	err := r.AppendMessage("myKey", msg)

	if err == nil {
		t.Logf("\t%s It should not return error.", success)
	} else {
		t.Fatalf("\t%s It should not return error, got [%s]", failure, err)
	}

	mockRedis.AssertExpectations(t)
}

func compare(one openai.Message, two openai.Message) bool {
	return one.Role == two.Role && one.Content == two.Content
}
