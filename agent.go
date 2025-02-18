package gomaky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Agent interface defines the methods for processing inputs and managing tools.
type Agent interface {
	Process(ctx context.Context, input string, additionalMessages ...[]openai.ChatCompletionMessage) (string, error)
	AddTool(tool Tool)
	SetSystemPrompt(prompt string)
	GetName() string
}

// BaseAgent holds the common implementation of the Agent interface.
type BaseAgent struct {
	client         *openai.Client
	name           string
	systemPrompt   string
	tools          map[string]Tool
	memory         Memory
	model          string
	mutex          sync.RWMutex
	maxRecursion   int
	temperature    float32
	responseFormat *openai.ChatCompletionResponseFormat
	buildError     error
}

// AddTool adds a tool to the agent.
func (b *BaseAgent) AddTool(tool Tool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	def := tool.GetDefinition()
	b.tools[def.Name] = tool
}

// GetName returns the name of the agent.
func (b *BaseAgent) GetName() string {
	return b.name
}

// Process processes input with additional messages and tools.
func (b *BaseAgent) Process(ctx context.Context, input string, additionalMessages ...[]openai.ChatCompletionMessage) (string, error) {
	if b.buildError != nil {
		return "", fmt.Errorf("agent build error: %w", b.buildError)
	}

	return b.processWithDepth(ctx, input, 0, additionalMessages...)
}

// SetSystemPrompt sets the system prompt for the agent.
func (b *BaseAgent) SetSystemPrompt(prompt string) {
	b.systemPrompt = prompt
}

// processWithDepth handles recursive processing and tool execution.
func (b *BaseAgent) processWithDepth(ctx context.Context, input string, depth int, additionalMessages ...[]openai.ChatCompletionMessage) (string, error) {
	if depth > b.maxRecursion {
		return "", errors.New("recursion limit reached")
	}

	b.mutex.Lock()
	// Actualizado para usar el nuevo método Add
	b.memory.Add(openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input,
	})
	messages := b.prepareMessages()
	for _, additional := range additionalMessages {
		messages = append(messages, additional...)
	}
	tools := b.prepareTools()
	b.mutex.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Creamos la solicitud incluyendo temperatura y formato de respuesta si están configurados
	request := openai.ChatCompletionRequest{
		Model:    b.model,
		Messages: messages,
		Tools:    tools,
	}

	// Añadir temperatura si está configurada
	if b.temperature > 0 {
		request.Temperature = b.temperature
	}

	// Añadir formato de respuesta si está configurado
	if b.responseFormat != nil {
		request.ResponseFormat = b.responseFormat
	}

	resp, err := b.client.CreateChatCompletion(ctx, request)
	if err != nil {
		log.Printf("error in chat completion: %v", err)
		return "", fmt.Errorf("error in chat completion: %w", err)
	}

	if len(resp.Choices) > 0 && resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		return b.handleToolCalls(ctx, resp.Choices[0].Message.ToolCalls, depth)
	}

	response := resp.Choices[0].Message.Content
	b.mutex.Lock()
	// Actualizado para usar el nuevo método Add
	b.memory.Add(openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
	})
	b.mutex.Unlock()

	return response, nil
}

// handleToolCalls handles tool execution during the agent's processing.
func (b *BaseAgent) handleToolCalls(ctx context.Context, toolCalls []openai.ToolCall, depth int) (string, error) {
	var wg sync.WaitGroup
	type toolResult struct {
		CallID  string
		Name    string
		Content string
		Error   error
	}
	results := make([]toolResult, len(toolCalls))

	// Procesar todas las llamadas a herramientas en paralelo
	for i, call := range toolCalls {
		wg.Add(1)
		go func(i int, call openai.ToolCall) {
			defer wg.Done()

			results[i].CallID = call.ID
			results[i].Name = call.Function.Name

			tool, exists := b.tools[call.Function.Name]
			if !exists {
				results[i].Error = fmt.Errorf("tool %s not found", call.Function.Name)
				return
			}

			result, err := tool.Execute([]byte(call.Function.Arguments))
			if err != nil {
				results[i].Error = fmt.Errorf("error executing tool %s: %w", call.Function.Name, err)
				return
			}

			// Convertir el resultado a string
			resultBytes, err := json.Marshal(result)
			if err != nil {
				results[i].Error = fmt.Errorf("error marshalling tool result: %w", err)
				return
			}

			results[i].Content = string(resultBytes)
		}(i, call)
	}

	wg.Wait()

	// Procesar resultados y detectar errores
	for _, r := range results {
		if r.Error != nil {
			return "", r.Error
		}

		b.mutex.Lock()
		// Actualizado para usar el nuevo método Add
		b.memory.Add(openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    r.Content,
			Name:       r.Name,
			ToolCallID: r.CallID,
		})
		b.mutex.Unlock()
	}

	return b.processWithDepth(ctx, "Synthesize and analyze the results obtained from the tools", depth+1)
}

// prepareMessages prepares the messages for the API request.
func (a *BaseAgent) prepareMessages() []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{}
	if a.systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.systemPrompt,
		})
	}
	messages = append(messages, a.memory.Get()...)
	return messages
}

// prepareTools prepares the tools for the API request.
func (a *BaseAgent) prepareTools() []openai.Tool {
	toolsList := []openai.Tool{}
	for _, tool := range a.tools {
		toolsList = append(toolsList, openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: tool.GetDefinition(),
		})
	}
	return toolsList
}

// AgentConfig holds configuration for creating an agent.
type AgentConfig struct {
	Client       *openai.Client
	Name         string
	SystemPrompt string
	Memory       Memory
	Model        string
	MaxRecursion int
}

// AgentBuilder allows for constructing an agent in a fluent and modular way.
type AgentBuilder struct {
	client         *openai.Client
	name           string
	systemPrompt   string
	memory         Memory
	model          string
	maxRecursion   int
	tools          map[string]Tool
	temperature    float32
	responseFormat *openai.ChatCompletionResponseFormat
	buildError     error
}

// NewAgentBuilder initializes a new AgentBuilder.
func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{
		tools: make(map[string]Tool),
	}
}

func (b *AgentBuilder) SetJSONResponseFormat(typeSample any) *AgentBuilder {
	schema, err := GenerateSchema(typeSample)
	if err != nil {
		b.buildError = fmt.Errorf("error generating schema: %w", err)
		return b
	}

	b.responseFormat = &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Schema: schema,
			Strict: true,
		},
	}

	return b
}

// SetClient sets the OpenAI client for the agent.
func (b *AgentBuilder) SetClient(client *openai.Client) *AgentBuilder {
	b.client = client
	return b
}

// SetName sets the name of the agent.
func (b *AgentBuilder) SetName(name string) *AgentBuilder {
	b.name = name
	return b
}

// SetSystemPrompt sets the system prompt for the agent.
func (b *AgentBuilder) SetSystemPrompt(prompt string) *AgentBuilder {
	b.systemPrompt = prompt
	return b
}

// SetMemory sets the memory for the agent.
func (b *AgentBuilder) SetMemory(memory Memory) *AgentBuilder {
	b.memory = memory
	return b
}

// SetModel sets the model to be used by the agent.
func (b *AgentBuilder) SetModel(model string) *AgentBuilder {
	b.model = model
	return b
}

// SetMaxRecursion sets the maximum recursion depth for the agent.
func (b *AgentBuilder) SetMaxRecursion(maxRecursion int) *AgentBuilder {
	b.maxRecursion = maxRecursion
	return b
}

func (b *AgentBuilder) SetTemperature(temperature float32) *AgentBuilder {
	b.temperature = temperature
	return b
}

func (b *AgentBuilder) SetResponseFormat(format *openai.ChatCompletionResponseFormat) *AgentBuilder {
	b.responseFormat = format
	return b
}

// AddTool adds a tool to the agent.
func (b *AgentBuilder) AddTool(tool Tool) *AgentBuilder {
	def := tool.GetDefinition()
	b.tools[def.Name] = tool
	return b
}

// Build constructs the agent from the current configuration.
func (b *AgentBuilder) Build() (Agent, error) {
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
		maxRecursion:   b.maxRecursion,
		temperature:    b.temperature,
		responseFormat: b.responseFormat,
		buildError:     b.buildError,
	}, nil
}
