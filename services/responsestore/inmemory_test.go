package responsestore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInmemoryStore(t *testing.T) {
	store := NewInmemoryStore()
	assert.NotNil(t, store)
	assert.NotNil(t, store.data)
}

func TestInmemoryStore_SetResponseID(t *testing.T) {
	store := NewInmemoryStore()
	err := store.SetResponseID("user1", "resp_abc123")
	assert.NoError(t, err)
	assert.Equal(t, "resp_abc123", store.data["user1"])
}

func TestInmemoryStore_GetResponseID(t *testing.T) {
	store := NewInmemoryStore()
	store.SetResponseID("user1", "resp_abc123")

	responseID, err := store.GetResponseID("user1")
	assert.NoError(t, err)
	assert.Equal(t, "resp_abc123", responseID)
}

func TestInmemoryStore_GetResponseID_NotFound(t *testing.T) {
	store := NewInmemoryStore()

	responseID, err := store.GetResponseID("unknown")
	assert.Error(t, err)
	assert.Equal(t, "", responseID)
}

func TestInmemoryStore_SetResponseID_Overwrite(t *testing.T) {
	store := NewInmemoryStore()
	store.SetResponseID("user1", "resp_first")
	store.SetResponseID("user1", "resp_second")

	responseID, err := store.GetResponseID("user1")
	assert.NoError(t, err)
	assert.Equal(t, "resp_second", responseID)
}
