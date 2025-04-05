package bot

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryChunkStorage_StoreAndGetChunks(t *testing.T) {
	storage := NewInMemoryChunkStorage()

	// Test data
	chatID := int64(12345)
	username := "testuser"
	messageID := 100
	chunks := []string{"chunk1", "chunk2", "chunk3"}

	// Store chunks
	storage.StoreChunks(chatID, username, messageID, chunks)

	// Get next chunk
	chunk, originalID, hasMore, exists := storage.GetNextChunk(chatID, username)

	// Assert results
	assert.Equal(t, "chunk2", chunk, "Should return the second chunk")
	assert.Equal(t, messageID, originalID, "Original message ID should match")
	assert.True(t, hasMore, "Should have more chunks")
	assert.True(t, exists, "Chunks should exist")

	// Get next chunk again
	chunk, originalID, hasMore, exists = storage.GetNextChunk(chatID, username)

	// Assert results for the last chunk
	assert.Equal(t, "chunk3", chunk, "Should return the third chunk")
	assert.Equal(t, messageID, originalID, "Original message ID should match")
	assert.False(t, hasMore, "Should not have more chunks after the third one")
	assert.True(t, exists, "Chunks should exist")

	// Try to get another chunk when we've reached the end
	chunk, originalID, hasMore, exists = storage.GetNextChunk(chatID, username)

	// Assert that we can't get more chunks
	assert.Equal(t, "", chunk, "Should return empty string when no more chunks")
	assert.Equal(t, 0, originalID, "Original message ID should be 0 when no more chunks")
	assert.False(t, hasMore, "Should not have more chunks")
	assert.False(t, exists, "Should indicate no more chunks exist")
}

func TestInMemoryChunkStorage_HasChunks(t *testing.T) {
	storage := NewInMemoryChunkStorage()

	// Test data
	chatID := int64(12345)
	username := "testuser"
	messageID := 100
	chunks := []string{"chunk1", "chunk2"}

	// Check before storing
	hasChunks := storage.HasChunks(chatID, username)
	assert.False(t, hasChunks, "Should not have chunks before storing")

	// Store chunks
	storage.StoreChunks(chatID, username, messageID, chunks)

	// Check after storing
	hasChunks = storage.HasChunks(chatID, username)
	assert.True(t, hasChunks, "Should have chunks after storing")

	// Retrieve all chunks
	storage.GetNextChunk(chatID, username) // Get chunk2
	storage.GetNextChunk(chatID, username) // No more chunks after this

	// Check after retrieving all
	hasChunks = storage.HasChunks(chatID, username)
	assert.False(t, hasChunks, "Should not have chunks after retrieving all")
}

func TestInMemoryChunkStorage_MultipleConcurrentConversations(t *testing.T) {
	storage := NewInMemoryChunkStorage()

	// Store chunks for conversation 1
	storage.StoreChunks(1001, "user1", 101, []string{"u1c1", "u1c2"})

	// Store chunks for conversation 2
	storage.StoreChunks(1002, "user2", 102, []string{"u2c1", "u2c2"})

	// Check conversation 1
	chunk1, msgID1, hasMore1, exists1 := storage.GetNextChunk(1001, "user1")
	assert.Equal(t, "u1c2", chunk1)
	assert.Equal(t, 101, msgID1)
	assert.False(t, hasMore1)
	assert.True(t, exists1)

	// Check conversation 2
	chunk2, msgID2, hasMore2, exists2 := storage.GetNextChunk(1002, "user2")
	assert.Equal(t, "u2c2", chunk2)
	assert.Equal(t, 102, msgID2)
	assert.False(t, hasMore2)
	assert.True(t, exists2)

	// Verify they're independent
	stillHas1 := storage.HasChunks(1001, "user1")
	stillHas2 := storage.HasChunks(1002, "user2")
	assert.False(t, stillHas1, "Conversation 1 should have no more chunks")
	assert.False(t, stillHas2, "Conversation 2 should have no more chunks")
}

func TestInMemoryChunkStorage_ThreadSafety(t *testing.T) {
	storage := NewInMemoryChunkStorage()

	// Test data
	chatIDs := []int64{1001, 1002, 1003, 1004, 1005}
	usernames := []string{"user1", "user2", "user3", "user4", "user5"}

	// Create a wait group
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < len(chatIDs); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			chatID := chatIDs[idx]
			username := usernames[idx]
			chunks := []string{
				"chunk-" + username + "-1",
				"chunk-" + username + "-2",
			}
			storage.StoreChunks(chatID, username, idx+100, chunks)
		}(i)
	}

	// Wait for all writes to complete
	wg.Wait()

	// Verify all data was stored correctly
	for i := 0; i < len(chatIDs); i++ {
		chatID := chatIDs[i]
		username := usernames[i]

		// Check if chunks exist
		hasChunks := storage.HasChunks(chatID, username)
		assert.True(t, hasChunks, "Should have chunks for conversation %d", i)

		// Get next chunk
		chunk, origID, _, exists := storage.GetNextChunk(chatID, username)
		expectedChunk := "chunk-" + username + "-2"

		assert.Equal(t, expectedChunk, chunk, "Should get correct chunk for conversation %d", i)
		assert.Equal(t, i+100, origID, "Should get correct original ID for conversation %d", i)
		assert.True(t, exists, "Chunks should exist for conversation %d", i)
	}
}

func TestInMemoryChunkStorage_EmptyChunksArray(t *testing.T) {
	storage := NewInMemoryChunkStorage()

	// Store empty chunks array
	storage.StoreChunks(12345, "user", 100, []string{})

	// Get next chunk
	_, _, _, exists := storage.GetNextChunk(12345, "user")
	assert.False(t, exists, "Should not exist when empty chunks array is stored")

	// Check has chunks
	hasChunks := storage.HasChunks(12345, "user")
	assert.False(t, hasChunks, "Should not have chunks when empty chunks array is stored")
}

func TestInMemoryChunkStorage_NonExistentConversation(t *testing.T) {
	storage := NewInMemoryChunkStorage()

	// Try to get chunks for a conversation that doesn't exist
	chunk, origID, hasMore, exists := storage.GetNextChunk(99999, "nonexistent")

	assert.Equal(t, "", chunk, "Should return empty string for non-existent conversation")
	assert.Equal(t, 0, origID, "Should return 0 origID for non-existent conversation")
	assert.False(t, hasMore, "Should return false hasMore for non-existent conversation")
	assert.False(t, exists, "Should return false exists for non-existent conversation")

	// Check has chunks for non-existent conversation
	hasChunks := storage.HasChunks(99999, "nonexistent")
	assert.False(t, hasChunks, "Should return false for non-existent conversation")
}

func TestConversationKey(t *testing.T) {
	key := conversationKey(12345, "testuser")
	assert.Equal(t, "12345:testuser", key, "Should generate correct conversation key")
}
