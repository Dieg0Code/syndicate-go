package gokamy

import (
	"sync"
)

// Memory defines the interface for managing a history of chat messages.
// It provides methods for adding messages, retrieving the complete history, and clearing the history.
type Memory interface {
	// Add appends a complete ChatCompletionMessage to the memory.
	Add(message Message)
	// Get returns a copy of all stored chat messages.
	Get() []Message
	// Clear removes all stored chat messages from memory.
	Clear()
}

// SimpleMemory implements a basic in-memory storage for chat messages.
// It uses a slice to store messages and a RWMutex for safe concurrent access.
type SimpleMemory struct {
	messages []Message    // Slice holding the chat messages.
	mutex    sync.RWMutex // RWMutex to ensure thread-safe access to messages.
}

// Add appends a complete chat message to the SimpleMemory.
func (s *SimpleMemory) Add(message Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = append(s.messages, message)
}

// Clear removes all stored messages from the memory.
func (s *SimpleMemory) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = make([]Message, 0)
}

// Get returns a copy of all stored chat messages to avoid data races.
// A copy of the messages slice is returned to ensure that external modifications
// do not affect the internal state.
func (s *SimpleMemory) Get() []Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	copyMessages := make([]Message, len(s.messages))
	copy(copyMessages, s.messages)
	return copyMessages
}

// NewSimpleMemory creates and returns a new instance of SimpleMemory.
// It initializes the internal message slice and ensures the memory is ready for use.
func NewSimpleMemory() Memory {
	return &SimpleMemory{
		messages: make([]Message, 0),
	}
}
