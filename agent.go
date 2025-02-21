// Package gokamy provides an SDK for interfacing with OpenAI's API,
// offering agents that process inputs, manage tool execution, and maintain memory.
package gokamy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// getSystemRole determines the system role for the prompt based on the model being used.
// For specific reasoner models, it returns a custom "developer" role.
func getSystemRole(model string) string {
	reasonerModels := []string{
		openai.O1Mini,
		openai.O1Mini20240912,
		openai.O1Preview,
		openai.O1Preview20240912,
		openai.O1,
		openai.O120241217,
		openai.O3Mini,
		openai.O3Mini20250131,
	}

	for _, m := range reasonerModels {
		if strings.EqualFold(model, m) {
			return RoleDeveloper
		}
	}
	return openai.ChatMessageRoleSystem
}

// Agent defines the interface for processing inputs and managing tools.
// Implementations of Agent should support processing messages, adding tools,
// configuring prompts, and providing a name identifier.
type Agent interface {
	Process(ctx context.Context, userName string, input string, additionalMessages ...[]Message) (string, error)
	AddTool(tool Tool)
	SetConfigPrompt(prompt string)
	GetName() string
}

// BaseAgent holds the common implementation of the Agent interface, including
// the OpenAI client, system prompt, tools, memory, model configuration, and concurrency control.
type BaseAgent struct {
	client         LLMClient
	name           string
	systemPrompt   string
	tools          map[string]Tool
	memory         Memory
	model          string
	mutex          sync.RWMutex
	temperature    float32
	responseFormat *ResponseFormat
	buildError     error
}

// AddTool adds a tool to the agent.
// AddTool adds a tool to the agent.
func (b *BaseAgent) AddTool(tool Tool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	def := tool.GetDefinition()
	b.tools[def.Name] = tool
}

// GetName returns the name identifier of the agent.
func (b *BaseAgent) GetName() string {
	return b.name
}

// Process takes the user input along with optional additional messages,
// adds the initial user message to memory, and initiates processing with available tools.
func (b *BaseAgent) Process(ctx context.Context, userName, input string, additionalMessages ...[]Message) (string, error) {
	if b.buildError != nil {
		return "", fmt.Errorf("agent build error: %w", b.buildError)
	}

	b.mutex.Lock()
	// Add the user's message to memory.
	b.memory.Add(Message{
		Role:    RoleUser,
		Name:    userName,
		Content: input,
	})
	// Prepare messages: include the system prompt and the conversation memory.
	messages := b.prepareMessages()
	for _, additional := range additionalMessages {
		messages = append(messages, additional...)
	}
	// Prepare tool definitions to be used.
	tools := b.prepareTools()
	b.mutex.Unlock()

	return b.processWithTools(ctx, messages, tools)
}

// SetConfigPrompt sets the system prompt for the agent, which can be used to configure behavior.
func (b *BaseAgent) SetConfigPrompt(prompt string) {
	b.systemPrompt = prompt
}

// processWithTools handles the API request to OpenAI, including executing tool calls if required.
// It manages context timeout, request setup, and response processing.
func (b *BaseAgent) processWithTools(ctx context.Context, messages []Message, tools []ToolDefinition) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req := ChatCompletionRequest{
		Model:          b.model,
		Messages:       messages,
		Tools:          tools,
		Temperature:    b.temperature,
		ResponseFormat: b.responseFormat,
	}

	resp, err := b.client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("error in chat completion: %v", err)
		return "", fmt.Errorf("error in chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response choices available")
	}

	choice := resp.Choices[0]

	// If the response indicates that tool calls are required, execute them.
	if choice.FinishReason == FinishReasonToolCalls {
		if err := b.handleToolCalls(choice.Message.ToolCalls); err != nil {
			return "", err
		}
		b.mutex.Lock()
		newMessages := b.prepareMessages()
		b.mutex.Unlock()
		return b.processWithTools(ctx, newMessages, tools)
	}

	// Store the assistant's response in memory and return it.
	response := choice.Message.Content
	b.mutex.Lock()
	b.memory.Add(Message{
		Role:    RoleAssistant,
		Content: response,
		Name:    b.name,
	})
	b.mutex.Unlock()

	return response, nil
}

// handleToolCalls executes each tool call concurrently and collects their results.
// It updates the agent's memory with the tool results and handles errors during execution.
func (b *BaseAgent) handleToolCalls(toolCalls []ToolCall) error {
	var wg sync.WaitGroup

	type toolResult struct {
		CallID  string
		Name    string
		Content string
		Error   error
	}

	results := make([]toolResult, len(toolCalls))

	for i, call := range toolCalls {
		wg.Add(1)
		go func(i int, call ToolCall) {
			defer wg.Done()
			results[i].CallID = call.ID
			results[i].Name = call.Name

			tool, exists := b.tools[call.Name]
			if !exists {
				results[i].Error = fmt.Errorf("tool %s not found", call.Name)
				return
			}

			result, err := tool.Execute(context.Background(), call.Args)
			if err != nil {
				results[i].Error = fmt.Errorf("error executing tool %s: %w", call.Name, err)
				return
			}

			resultBytes, err := json.Marshal(result)
			if err != nil {
				results[i].Error = fmt.Errorf("error marshalling tool result: %w", err)
				return
			}

			results[i].Content = string(resultBytes)
		}(i, call)
	}

	wg.Wait()

	for _, r := range results {
		if r.Error != nil {
			return r.Error
		}
		b.mutex.Lock()
		b.memory.Add(Message{
			Role:    RoleTool,
			Content: r.Content,
			Name:    r.Name,
		})
		b.mutex.Unlock()
	}
	return nil
}

// prepareMessages compiles the messages to be sent to the API, including the system prompt and conversation memory.
func (b *BaseAgent) prepareMessages() []Message {
	var msgs []Message
	if b.systemPrompt != "" {
		msgs = append(msgs, Message{
			Role:    getSystemRole(b.model),
			Content: b.systemPrompt,
		})
	}
	msgs = append(msgs, b.memory.Get()...)
	return msgs
}

// prepareTools compiles the list of tools to be included in the API request.
func (b *BaseAgent) prepareTools() []ToolDefinition {
	var defs []ToolDefinition
	for _, tool := range b.tools {
		defs = append(defs, tool.GetDefinition())
	}
	return defs
}

// AgentBuilder provides a fluent, modular way to configure and construct an Agent.
type AgentBuilder struct {
	client         LLMClient
	name           string
	systemPrompt   string
	memory         Memory
	model          string
	tools          map[string]Tool
	temperature    float32
	responseFormat *ResponseFormat
	buildError     error
}

// NewAgentBuilder initializes and returns a new instance of AgentBuilder.
func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{
		tools: make(map[string]Tool),
	}
}

// SetJSONResponseFormat configures the agent to use a JSON schema for response formatting,
// generating the schema from a provided sample type.
func (b *AgentBuilder) SetJSONResponseFormat(schemaName string, structSchema any) *AgentBuilder {
	schemaDef, err := GenerateSchema(structSchema) // Assumes GenerateSchema returns a json.RawMessage
	if err != nil {
		b.buildError = fmt.Errorf("error generating schema: %w", err)
		return b
	}

	// Marshal the tool definition into a json.RawMessage.
	rawSchema, err := json.Marshal(schemaDef)
	if err != nil {
		b.buildError = fmt.Errorf("error marshalling schema: %w", err)
		return b
	}

	b.responseFormat = &ResponseFormat{
		Type: "json_schema",
		JSONSchema: &JSONSchema{
			Name:   schemaName,
			Schema: rawSchema,
			Strict: true,
		},
	}
	return b
}

// SetClient sets the LLMClient to be used by the agent.
func (b *AgentBuilder) SetClient(client LLMClient) *AgentBuilder {
	b.client = client
	return b
}

// SetName sets the name identifier for the agent.
func (b *AgentBuilder) SetName(name string) *AgentBuilder {
	b.name = name
	return b
}

// SetConfigPrompt sets the system prompt that configures the agent's behavior.
func (b *AgentBuilder) SetConfigPrompt(prompt string) *AgentBuilder {
	b.systemPrompt = prompt
	return b
}

// SetMemory sets the memory implementation for the agent.
func (b *AgentBuilder) SetMemory(memory Memory) *AgentBuilder {
	b.memory = memory
	return b
}

// SetModel configures the model to be used by the agent.
func (b *AgentBuilder) SetModel(model string) *AgentBuilder {
	b.model = model
	return b
}

// SetTemperature sets the temperature parameter for the agent's responses.
func (b *AgentBuilder) SetTemperature(temperature float32) *AgentBuilder {
	b.temperature = temperature
	return b
}

// AddTool adds a tool to the agent's configuration, making it available during processing.
func (b *AgentBuilder) AddTool(tool Tool) *AgentBuilder {
	def := tool.GetDefinition()
	b.tools[def.Name] = tool
	return b
}

// Build constructs and returns an Agent based on the current configuration.
// It returns an error if any issues occurred during the builder setup.
func (b *AgentBuilder) Build() (*BaseAgent, error) {
	if b.buildError != nil {
		return nil, b.buildError
	}

	return &BaseAgent{
		client:         b.client,
		name:           b.name,
		systemPrompt:   b.systemPrompt,
		tools:          b.tools,
		memory:         b.memory,
		model:          b.model,
		temperature:    b.temperature,
		responseFormat: b.responseFormat,
		buildError:     b.buildError,
	}, nil
}
