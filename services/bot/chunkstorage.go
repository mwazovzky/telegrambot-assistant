package bot

import (
	"fmt"
	"sync"
)

// ChunkStorage defines the interface for storing and retrieving message chunks
type ChunkStorage interface {
	// StoreChunks stores a sequence of message chunks for a conversation
	StoreChunks(chatID int64, username string, messageID int, chunks []string)

	// GetNextChunk retrieves the next chunk for a conversation and advances the position
	// Returns the chunk content, original message ID, whether more chunks exist, and if the query was successful
	GetNextChunk(chatID int64, username string) (chunk string, originalID int, hasMore bool, exists bool)

	// HasChunks checks if there are more chunks available for a conversation
	HasChunks(chatID int64, username string) bool

	// Clear removes all stored chunks for a conversation
	Clear(chatID int64, username string)
}

// ChunkQueue maintains pending chunks for a specific conversation
type ChunkQueue struct {
	Chunks     []string
	Position   int
	OriginalID int
}

// InMemoryChunkStorage implements the ChunkStorage interface using in-memory storage
type InMemoryChunkStorage struct {
	pendingChunks map[string]*ChunkQueue
	mutex         sync.RWMutex
}

// NewInMemoryChunkStorage creates a new InMemoryChunkStorage
func NewInMemoryChunkStorage() *InMemoryChunkStorage {
	return &InMemoryChunkStorage{
		pendingChunks: make(map[string]*ChunkQueue),
	}
}

// StoreChunks stores chunks for a conversation
func (cs *InMemoryChunkStorage) StoreChunks(chatID int64, username string, messageID int, chunks []string) {
	convKey := conversationKey(chatID, username)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.pendingChunks[convKey] = &ChunkQueue{
		Chunks:     chunks,
		Position:   1, // Start from second chunk (index 1)
		OriginalID: messageID,
	}
}

// GetNextChunk retrieves the next chunk for a conversation and advances the position
func (cs *InMemoryChunkStorage) GetNextChunk(chatID int64, username string) (string, int, bool, bool) {
	convKey := conversationKey(chatID, username)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	queue, exists := cs.pendingChunks[convKey]
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
func (cs *InMemoryChunkStorage) HasChunks(chatID int64, username string) bool {
	convKey := conversationKey(chatID, username)

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	queue, exists := cs.pendingChunks[convKey]
	return exists && queue.Position < len(queue.Chunks)
}

// Consider adding a clear method for testing and cleanup purposes
func (cs *InMemoryChunkStorage) Clear(chatID int64, username string) {
	convKey := conversationKey(chatID, username)

	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.pendingChunks, convKey)
}

// Helper function moved from bot.go to avoid duplication
func conversationKey(chatID int64, username string) string {
	return fmt.Sprintf("%d:%s", chatID, username)
}
