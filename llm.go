package syndicate

import (
	"context"
	"encoding/json"
)

// Role constants define standard message roles across different providers
const (
	RoleSystem    = "system"
	RoleDeveloper = "developer"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// FinishReason constants define standard reasons for completion.
const (
	FinishReasonStop      = "stop"
	FinishReasonLength    = "length"
	FinishReasonToolCalls = "tool_calls"
)

// LLMClient defines the interface for interacting with LLM providers.
type LLMClient interface {
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error)
}

// ChatCompletionRequest represents a unified chat completion request.
type ChatCompletionRequest struct {
	Model          string           `json:"model"`
	Messages       []Message        `json:"messages"`
	Tools          []ToolDefinition `json:"tools,omitempty"`
	Temperature    float32          `json:"temperature"`
	ResponseFormat *ResponseFormat  `json:"response_format,omitempty"`
}

// Message represents a chat message with standardized fields.
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	Name      string     `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolID    string     `json:"tool_id,omitempty"`
}

// ToolCall represents a tool invocation request.
type ToolCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

// ToolDefinition describes a tool's capabilities.
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// Tool defines the interface for executable tools.
type Tool interface {
	GetDefinition() ToolDefinition
	Execute(args json.RawMessage) (interface{}, error)
}

// ResponseFormat specifies how the LLM should format its response.
type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema defines the structure for responses in JSON.
type JSONSchema struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
	Strict bool            `json:"strict"`
}

// ChatCompletionResponse represents a unified response structure from the LLM.
type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a single completion option.
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage provides token usage statistics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
