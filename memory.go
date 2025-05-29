package syndicate

import (
	"errors"
	"fmt"
	"sync"
)

// Memory defines the interface for managing a history of chat messages.
type Memory interface {
	// Add appends a message to the memory.
	Add(message Message)
	// Get returns all stored chat messages.
	Get() []Message
}

// MemoryConfig holds the configuration for creating custom memory implementations
type MemoryConfig struct {
	addFunc func(message Message)
	getFunc func() []Message
}

// MemoryOption defines a function that configures a Memory implementation
type MemoryOption func(*MemoryConfig) error

// WithAddHandler sets the function to handle adding messages
func WithAddHandler(addFunc func(message Message)) MemoryOption {
	return func(config *MemoryConfig) error {
		if addFunc == nil {
			return errors.New("addFunc cannot be nil")
		}
		config.addFunc = addFunc
		return nil
	}
}

// WithGetHandler sets the function to handle retrieving messages
func WithGetHandler(getFunc func() []Message) MemoryOption {
	return func(config *MemoryConfig) error {
		if getFunc == nil {
			return errors.New("getFunc cannot be nil")
		}
		config.getFunc = getFunc
		return nil
	}
}

// customMemory implements Memory using provided functions
type customMemory struct {
	addFunc func(message Message)
	getFunc func() []Message
}

func (c *customMemory) Add(message Message) {
	if c.addFunc != nil {
		c.addFunc(message)
	}
}

func (c *customMemory) Get() []Message {
	if c.getFunc != nil {
		return c.getFunc()
	}
	return []Message{}
}

// NewMemory creates a custom Memory implementation using functional options.
// Returns an error if required handlers are not provided.
//
// Example:
//
//	memory, err := syndicate.NewMemory(
//	    syndicate.WithAddHandler(func(msg syndicate.Message) {
//	        // Your custom add logic here
//	    }),
//	    syndicate.WithGetHandler(func() []syndicate.Message {
//	        // Your custom get logic here
//	        return messages
//	    }),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewMemory(options ...MemoryOption) (Memory, error) {
	config := &MemoryConfig{}

	for _, option := range options {
		if err := option(config); err != nil {
			return nil, fmt.Errorf("failed to apply memory option: %w", err)
		}
	}

	// Validate that both handlers are provided
	if config.addFunc == nil {
		return nil, errors.New("WithAddHandler is required when creating custom memory")
	}
	if config.getFunc == nil {
		return nil, errors.New("WithGetHandler is required when creating custom memory")
	}

	return &customMemory{
		addFunc: config.addFunc,
		getFunc: config.getFunc,
	}, nil
}

// NewSimpleMemory creates a basic in-memory storage for chat messages.
// This function cannot fail, so it doesn't return an error.
func NewSimpleMemory() Memory {
	messages := make([]Message, 0)
	var mutex sync.RWMutex

	// We know these handlers are valid, so we can ignore the error
	memory, _ := NewMemory(
		WithAddHandler(func(message Message) {
			mutex.Lock()
			defer mutex.Unlock()
			messages = append(messages, message)
		}),
		WithGetHandler(func() []Message {
			mutex.RLock()
			defer mutex.RUnlock()
			copyMessages := make([]Message, len(messages))
			copy(copyMessages, messages)
			return copyMessages
		}),
	)

	return memory
}
