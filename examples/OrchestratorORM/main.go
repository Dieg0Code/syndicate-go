package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	gokamy "github.com/Dieg0Code/gokamy-ai"
	openai "github.com/sashabaranov/go-openai"
)

// CustomMemory is a custom implementation of the Memory interface.
// It wraps a simple in-memory slice and logs each message addition.
type CustomMemory struct {
	messages []openai.ChatCompletionMessage
	mutex    sync.RWMutex
}

// Add adds a message to the custom memory.
func (m *CustomMemory) Add(message openai.ChatCompletionMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	fmt.Println("CustomMemory - Adding message:", message)
	m.messages = append(m.messages, message)
}

// Get returns all messages stored in the custom memory.
func (m *CustomMemory) Get() []openai.ChatCompletionMessage {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	copied := make([]openai.ChatCompletionMessage, len(m.messages))
	copy(copied, m.messages)
	return copied
}

// Clear removes all messages from the custom memory.
func (m *CustomMemory) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = []openai.ChatCompletionMessage{}
}

// NewCustomMemory returns a new instance of Memory interface backed by CustomMemory.
func NewCustomMemory() gokamy.Memory {
	return &CustomMemory{
		messages: make([]openai.ChatCompletionMessage, 0),
	}
}

func main() {
	// Initialize the OpenAI client using your API key.
	client := openai.NewClient("YOUR_API_KEY")

	// Use the custom memory implementation.
	customMemory := NewCustomMemory()

	// Build an agent that uses CustomMemory.
	agent, err := gokamy.NewAgentBuilder().
		SetClient(client).
		SetName("CustomMemoryAgent").
		SetConfigPrompt("You are an agent that logs all messages to a custom memory implementation.").
		SetMemory(customMemory).
		SetModel(openai.GPT4).
		Build()
	if err != nil {
		log.Fatalf("Error building agent: %v", err)
	}

	// User name for the conversation.
	userName := "User"

	// Process a sample input with the agent.
	input := "What is your favorite quote?"
	response, err := agent.Process(context.Background(), userName, input)
	if err != nil {
		log.Fatalf("Error processing request: %v", err)
	}

	fmt.Println("Agent Response:")
	fmt.Println(response)

	// Display the contents of the custom memory.
	fmt.Println("\nCustom Memory Contents:")
	for _, msg := range customMemory.Get() {
		fmt.Printf("Role: %s, Content: %s\n", msg.Role, msg.Content)
	}
}
