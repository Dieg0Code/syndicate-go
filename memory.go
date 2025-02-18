package gokamy

import (
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

// Memory interface define métodos para manejar el historial de mensajes
type Memory interface {
	// Add añade un mensaje ChatCompletionMessage completo
	Add(message openai.ChatCompletionMessage)
	// Get devuelve todos los mensajes
	Get() []openai.ChatCompletionMessage
	// Clear elimina todos los mensajes
	Clear()
}

// SimpleMemory implementa una memoria básica para el agente
type SimpleMemory struct {
	messages []openai.ChatCompletionMessage
	mutex    sync.RWMutex
}

// Add implementa Memory.Add para añadir un mensaje completo
func (s *SimpleMemory) Add(message openai.ChatCompletionMessage) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = append(s.messages, message)
}

// Clear implementa Memory.Clear
func (s *SimpleMemory) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = make([]openai.ChatCompletionMessage, 0)
}

// Get implementa Memory.Get
func (s *SimpleMemory) Get() []openai.ChatCompletionMessage {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	copyMessages := make([]openai.ChatCompletionMessage, len(s.messages))
	copy(copyMessages, s.messages)
	return copyMessages
}

// NewSimpleMemory crea una nueva instancia de SimpleMemory
func NewSimpleMemory() Memory {
	return &SimpleMemory{
		messages: make([]openai.ChatCompletionMessage, 0),
	}
}
