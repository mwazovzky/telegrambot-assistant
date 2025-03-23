package repository

import (
	"testing"

	openai "github.com/mwazovzky/assistant"
	"github.com/stretchr/testify/assert"
)

func TestNewInmemoryRepository(t *testing.T) {
	repo := NewInmemoryRepository()
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.data)
}

func TestInmemoryRepository_CreateThread(t *testing.T) {
	repo := NewInmemoryRepository()
	err := repo.CreateThread("testThread")
	assert.NoError(t, err)
	assert.Contains(t, repo.data, "testThread")
}

func TestInmemoryRepository_AppendMessage(t *testing.T) {
	repo := NewInmemoryRepository()
	repo.CreateThread("testThread")

	msg := openai.Message{Role: "user", Content: "test message"}
	err := repo.AppendMessage("testThread", msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(repo.data["testThread"]))
	assert.Equal(t, msg, repo.data["testThread"][0])
}

func TestInmemoryRepository_GetMessages(t *testing.T) {
	repo := NewInmemoryRepository()
	repo.CreateThread("testThread")

	msg := openai.Message{Role: "user", Content: "test message"}
	repo.AppendMessage("testThread", msg)

	messages, err := repo.GetMessages("testThread")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, msg, messages[0])
}
