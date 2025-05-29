// Package syndicate provides an SDK for interfacing with OpenAI's API,
// offering agents that process inputs, manage tool execution, and maintain memory.
package syndicate

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

// ChatOption defines a function that configures a chat request.
type ChatOption func(*chatRequest)

// chatRequest holds the parameters for a chat request.
type chatRequest struct {
	userName           string
	input              string
	imageURLs          []string
	additionalMessages [][]Message
	timeout            *time.Duration // Timeout específico para esta llamada
}

// WithUserName sets the user name for the chat request.
func WithUserName(userName string) ChatOption {
	return func(r *chatRequest) {
		r.userName = userName
	}
}

// WithInput sets the input text for the chat request.
func WithInput(input string) ChatOption {
	return func(r *chatRequest) {
		r.input = input
	}
}

// WithImages sets the image URLs for the chat request.
func WithImages(imageURLs ...string) ChatOption {
	return func(r *chatRequest) {
		r.imageURLs = imageURLs
	}
}

// WithAdditionalMessages adds additional messages to the chat request.
func WithAdditionalMessages(messages ...[]Message) ChatOption {
	return func(r *chatRequest) {
		r.additionalMessages = messages
	}
}

// WithChatTimeout sets a specific timeout for this chat request.
// Overrides the agent's default timeout for this request only.
func WithChatTimeout(timeout time.Duration) ChatOption {
	return func(r *chatRequest) {
		if timeout > 0 {
			r.timeout = &timeout
		}
	}
}

// Agent defines the interface for processing inputs and managing tools.
type Agent interface {
	Chat(ctx context.Context, options ...ChatOption) (string, error)
	GetName() string
}

// agent holds the implementation of the Agent interface.
type agent struct {
	client         LLMClient
	name           string
	description    string
	systemPrompt   string
	tools          map[string]Tool
	memory         Memory
	model          string
	mutex          sync.RWMutex
	temperature    float32
	responseFormat *ResponseFormat
	timeout        time.Duration // Timeout configurable para el agente
}

// AgentOption defines a function that configures an Agent.
type AgentOption func(*agent) error

// WithClient sets the LLM client for the agent.
func WithClient(client LLMClient) AgentOption {
	return func(a *agent) error {
		if client == nil {
			return errors.New("client cannot be nil")
		}
		a.client = client
		return nil
	}
}

// WithName sets the name for the agent.
func WithName(name string) AgentOption {
	return func(a *agent) error {
		if name == "" {
			return errors.New("name cannot be empty")
		}
		a.name = name
		return nil
	}
}

// WithDescription sets the description for the agent.
func WithDescription(description string) AgentOption {
	return func(a *agent) error {
		a.description = description
		return nil
	}
}

// WithSystemPrompt sets the system prompt for the agent.
func WithSystemPrompt(prompt string) AgentOption {
	return func(a *agent) error {
		a.systemPrompt = prompt
		return nil
	}
}

// WithMemory sets the memory implementation for the agent.
func WithMemory(memory Memory) AgentOption {
	return func(a *agent) error {
		if memory == nil {
			return errors.New("memory cannot be nil")
		}
		a.memory = memory
		return nil
	}
}

// WithModel sets the model for the agent.
func WithModel(model string) AgentOption {
	return func(a *agent) error {
		if model == "" {
			return errors.New("model cannot be empty")
		}
		a.model = model
		return nil
	}
}

// WithTemperature sets the temperature for the agent.
func WithTemperature(temperature float32) AgentOption {
	return func(a *agent) error {
		if temperature < 0 || temperature > 2 {
			return errors.New("temperature must be between 0 and 2")
		}
		a.temperature = temperature
		return nil
	}
}

// WithTimeout sets the context timeout for API requests.
func WithTimeout(timeout time.Duration) AgentOption {
	return func(a *agent) error {
		if timeout <= 0 {
			return errors.New("timeout must be greater than 0")
		}
		a.timeout = timeout
		return nil
	}
}

// WithTool adds a tool to the agent.
func WithTool(tool Tool) AgentOption {
	return func(a *agent) error {
		if tool == nil {
			return errors.New("tool cannot be nil")
		}
		def := tool.GetDefinition()
		if def.Name == "" {
			return errors.New("tool name cannot be empty")
		}
		a.tools[def.Name] = tool
		return nil
	}
}

// WithTools adds multiple tools to the agent.
func WithTools(tools ...Tool) AgentOption {
	return func(a *agent) error {
		for _, tool := range tools {
			if tool == nil {
				return errors.New("tool cannot be nil")
			}
			def := tool.GetDefinition()
			if def.Name == "" {
				return errors.New("tool name cannot be empty")
			}
			a.tools[def.Name] = tool
		}
		return nil
	}
}

// WithJSONResponseFormat configures the agent to use a JSON schema for response formatting.
func WithJSONResponseFormat(schemaName string, structSchema any) AgentOption {
	return func(a *agent) error {
		if schemaName == "" {
			return errors.New("schema name cannot be empty")
		}

		schema, err := GenerateRawSchema(structSchema)
		if err != nil {
			return fmt.Errorf("error generating schema: %w", err)
		}

		a.responseFormat = &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &JSONSchema{
				Name:   schemaName,
				Schema: schema,
				Strict: true,
			},
		}
		return nil
	}
}

// NewAgent creates a new agent with the provided options.
func NewAgent(options ...AgentOption) (Agent, error) {
	a := &agent{
		tools:       make(map[string]Tool),
		temperature: 1.0,              // Default temperature
		timeout:     30 * time.Second, // Default timeout
	}

	for _, option := range options {
		if err := option(a); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate required fields
	if a.client == nil {
		return nil, errors.New("client is required")
	}
	if a.name == "" {
		return nil, errors.New("name is required")
	}
	if a.memory == nil {
		return nil, errors.New("memory is required")
	}
	if a.model == "" {
		return nil, errors.New("model is required")
	}

	return a, nil
}

// GetName returns the agent's name.
func (a *agent) GetName() string {
	return a.name
}

// Chat processes a chat request with the provided options.
func (a *agent) Chat(ctx context.Context, options ...ChatOption) (string, error) {
	// Apply default values
	req := &chatRequest{}

	// Apply all options
	for _, opt := range options {
		opt(req)
	}

	// Validate required fields
	if req.userName == "" {
		return "", errors.New("user name is required")
	}
	if req.input == "" {
		return "", errors.New("input is required")
	}

	a.mutex.Lock()
	// Add the user's message to memory
	message := Message{
		Role:    RoleUser,
		Name:    req.userName,
		Content: req.input,
	}

	// Add images if provided
	if len(req.imageURLs) > 0 {
		message.ImageURLs = req.imageURLs
	}

	a.memory.Add(message)

	// Prepare messages: include the system prompt and the conversation memory
	messages := a.prepareMessages()
	for _, additional := range req.additionalMessages {
		messages = append(messages, additional...)
	}

	// Prepare tool definitions to be used
	tools := a.prepareTools()
	a.mutex.Unlock()

	// Usar timeout específico si se proporciona, sino usar el del agente
	timeout := a.timeout
	if req.timeout != nil {
		timeout = *req.timeout
	}

	return a.processWithTools(ctx, messages, tools, timeout)
}

// processWithTools handles the API request to OpenAI, including executing tool calls if required.
// It manages context timeout, request setup, and response processing.
func (a *agent) processWithTools(ctx context.Context, messages []Message, tools []ToolDefinition, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req := ChatCompletionRequest{
		Model:          a.model,
		Messages:       messages,
		Tools:          tools,
		Temperature:    a.temperature,
		ResponseFormat: a.responseFormat,
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
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
		if err := a.handleToolCalls(choice.Message.ToolCalls); err != nil {
			return "", err
		}
		a.mutex.Lock()
		newMessages := a.prepareMessages()
		a.mutex.Unlock()
		return a.processWithTools(ctx, newMessages, tools, timeout)
	}

	// Store the assistant's response in memory and return it.
	response := choice.Message.Content
	a.mutex.Lock()
	a.memory.Add(Message{
		Role:    RoleAssistant,
		Content: response,
		Name:    a.name,
	})
	a.mutex.Unlock()

	return response, nil
}

// handleToolCalls executes each tool call concurrently and collects their results.
// It updates the agent's memory with the tool results and handles errors during execution.
func (a *agent) handleToolCalls(toolCalls []ToolCall) error {
	a.mutex.Lock()
	a.memory.Add(Message{
		Role:      RoleAssistant,
		ToolCalls: toolCalls,
		Content:   "Executing tool calls...",
		Name:      a.name,
	})
	a.mutex.Unlock()

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

			a.mutex.RLock()
			tool, exists := a.tools[call.Name]
			a.mutex.RUnlock()

			if !exists {
				results[i].Error = fmt.Errorf("tool %s not found", call.Name)
				return
			}

			result, err := tool.Execute(call.Args)
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
		a.mutex.Lock()
		a.memory.Add(Message{
			Role:       RoleTool,
			Content:    r.Content,
			Name:       r.Name,
			ToolCallID: r.CallID,
		})
		a.mutex.Unlock()
	}
	return nil
}

// validateAndFixMessageSequence ensures tool_calls and tool messages are properly paired
// This prevents OpenAI API errors when loading partial conversation history from limited memory windows
func validateAndFixMessageSequence(messages []Message) []Message {
	var validMessages []Message
	var pendingToolCalls map[string]bool // Track pending tool call IDs

	for _, msg := range messages {
		switch msg.Role {
		case RoleAssistant:
			if len(msg.ToolCalls) > 0 {
				// This message has tool calls - track them
				pendingToolCalls = make(map[string]bool)
				for _, toolCall := range msg.ToolCalls {
					pendingToolCalls[toolCall.ID] = true
				}
				validMessages = append(validMessages, msg)
			} else {
				// Regular assistant message - clear any pending tool calls
				pendingToolCalls = nil
				validMessages = append(validMessages, msg)
			}

		case RoleTool:
			// Tool message must have a preceding tool_calls message
			if pendingToolCalls != nil && pendingToolCalls[msg.ToolCallID] {
				// Valid tool response - remove from pending
				delete(pendingToolCalls, msg.ToolCallID)
				validMessages = append(validMessages, msg)
			}
			// If no matching tool call found, skip this message (it's orphaned)

		default:
			// User, system, developer messages are always valid
			pendingToolCalls = nil // Clear any pending tool calls
			validMessages = append(validMessages, msg)
		}
	}

	return validMessages
}

// prepareMessages compiles the messages to be sent to the API, including the system prompt and conversation memory.
func (a *agent) prepareMessages() []Message {
	var msgs []Message
	if a.systemPrompt != "" {
		msgs = append(msgs, Message{
			Role:    getSystemRole(a.model),
			Content: a.systemPrompt,
		})
	}

	// Get memory messages and validate tool call sequences
	memoryMessages := a.memory.Get()
	validatedMessages := validateAndFixMessageSequence(memoryMessages)

	msgs = append(msgs, validatedMessages...)
	return msgs
}

// prepareTools compiles the list of tools to be included in the API request.
func (a *agent) prepareTools() []ToolDefinition {
	var defs []ToolDefinition
	for _, tool := range a.tools {
		defs = append(defs, tool.GetDefinition())
	}
	return defs
}
