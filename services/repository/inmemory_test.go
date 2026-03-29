package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInmemoryRepository(t *testing.T) {
	repo := NewInmemoryRepository()
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.data)
}

func TestInmemoryRepository_SetResponseID(t *testing.T) {
	repo := NewInmemoryRepository()
	err := repo.SetResponseID("user1", "resp_abc123")
	assert.NoError(t, err)
	assert.Equal(t, "resp_abc123", repo.data["user1"])
}

func TestInmemoryRepository_GetResponseID(t *testing.T) {
	repo := NewInmemoryRepository()
	repo.SetResponseID("user1", "resp_abc123")

	responseID, err := repo.GetResponseID("user1")
	assert.NoError(t, err)
	assert.Equal(t, "resp_abc123", responseID)
}

func TestInmemoryRepository_GetResponseID_NotFound(t *testing.T) {
	repo := NewInmemoryRepository()

	responseID, err := repo.GetResponseID("unknown")
	assert.Error(t, err)
	assert.Equal(t, "", responseID)
}

func TestInmemoryRepository_SetResponseID_Overwrite(t *testing.T) {
	repo := NewInmemoryRepository()
	repo.SetResponseID("user1", "resp_first")
	repo.SetResponseID("user1", "resp_second")

	responseID, err := repo.GetResponseID("user1")
	assert.NoError(t, err)
	assert.Equal(t, "resp_second", responseID)
}
