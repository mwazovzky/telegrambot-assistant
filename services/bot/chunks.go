package bot

import (
	"fmt"
	"sync"
)

// ChunkQueue maintains pending chunks for a specific conversation
type ChunkQueue struct {
	Chunks     []string
	Position   int
	OriginalID int
}

// ChunkManager handles the storage and retrieval of message chunks
type ChunkManager struct {
	pendingChunks map[string]*ChunkQueue
	mutex         sync.RWMutex
}

// NewChunkManager creates a new ChunkManager
func NewChunkManager() *ChunkManager {
	return &ChunkManager{
		pendingChunks: make(map[string]*ChunkQueue),
	}
}

// StoreChunks stores chunks for a conversation
func (cm *ChunkManager) StoreChunks(chatID int64, username string, messageID int, chunks []string) {
	convKey := fmt.Sprintf("%d:%s", chatID, username)

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.pendingChunks[convKey] = &ChunkQueue{
		Chunks:     chunks,
		Position:   1, // Start from second chunk (index 1)
		OriginalID: messageID,
	}
}

// GetNextChunk retrieves the next chunk for a conversation and advances the position
// Returns the chunk, originalMessageID, hasMore, exists
func (cm *ChunkManager) GetNextChunk(chatID int64, username string) (string, int, bool, bool) {
	convKey := fmt.Sprintf("%d:%s", chatID, username)

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	queue, exists := cm.pendingChunks[convKey]
	if !exists || queue.Position >= len(queue.Chunks) {
		return "", 0, false, false
	}

	chunk := queue.Chunks[queue.Position]
	originalID := queue.OriginalID
	hasMore := queue.Position < len(queue.Chunks)-1

	// Move to next chunk
	queue.Position++

	return chunk, originalID, hasMore, true
}

// HasChunks checks if there are more chunks for a conversation
func (cm *ChunkManager) HasChunks(chatID int64, username string) bool {
	convKey := fmt.Sprintf("%d:%s", chatID, username)

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	queue, exists := cm.pendingChunks[convKey]
	return exists && queue.Position < len(queue.Chunks)
}
