package syndicate

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the LLMClient interface using the OpenAI SDK.
// It wraps the official OpenAI client and provides a consistent interface for making chat completion requests.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIAzureClient creates an LLMClient for Azure using the provided API key and endpoint.
// It configures the client with Azure-specific settings.
func NewOpenAIAzureClient(apiKey, endpoint string) LLMClient {
	config := openai.DefaultAzureConfig(apiKey, endpoint)
	return &OpenAIClient{
		client: openai.NewClientWithConfig(config),
	}
}

// NewOpenAIClient creates a new LLMClient using the provided API key with the standard OpenAI endpoint.
func NewOpenAIClient(apiKey string) LLMClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

// mapToOpenAIMessages converts a slice of internal Message structs into the format required by the OpenAI ChatCompletion API.
func mapToOpenAIMessages(messages []Message) []openai.ChatCompletionMessage {
	var msgs []openai.ChatCompletionMessage
	for _, m := range messages {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:       m.Role,
			Name:       m.Name,
			Content:    m.Content,
			ToolCallID: m.ToolID,
		})
	}
	return msgs
}

// mapToOpenAITools converts a slice of internal ToolDefinition structs into OpenAI Tools.
// These definitions are used to enable function calls in the API.
func mapToOpenAITools(tools []ToolDefinition) []openai.Tool {
	var result []openai.Tool
	for _, t := range tools {
		result = append(result, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
				Strict:      true,
			},
		})
	}
	return result
}

// mapFromOpenAIToolCalls converts a slice of OpenAI ToolCall objects into the internal ToolCall structure.
// This enables the SDK to process tool calls in a provider-agnostic manner.
func mapFromOpenAIToolCalls(calls []openai.ToolCall) []ToolCall {
	var result []ToolCall
	for _, call := range calls {
		result = append(result, ToolCall{
			ID:   call.ID,
			Name: call.Function.Name,
			Args: json.RawMessage(call.Function.Arguments),
		})
	}
	return result
}

// mapFromOpenAIResponse converts an OpenAI ChatCompletionResponse into the internal ChatCompletionResponse format.
// It maps messages, token usage, and tool calls.
func mapFromOpenAIResponse(resp openai.ChatCompletionResponse) ChatCompletionResponse {
	var choices []Choice
	for _, c := range resp.Choices {
		choices = append(choices, Choice{
			Message: Message{
				Role:      c.Message.Role,
				Name:      c.Message.Name,
				Content:   c.Message.Content,
				ToolCalls: mapFromOpenAIToolCalls(c.Message.ToolCalls), // Map tool calls from OpenAI to internal representation.
			},
			FinishReason: string(c.FinishReason),
		})
	}
	usage := Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	return ChatCompletionResponse{
		Choices: choices,
		Usage:   usage,
	}
}

// CreateChatCompletion sends a chat completion request to the OpenAI API using the provided request parameters.
// It converts internal messages and tool definitions to OpenAI formats, sends the request,
// and maps the response back into the SDK's unified structure.
func (o *OpenAIClient) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	openaiReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    mapToOpenAIMessages(req.Messages),
		Temperature: req.Temperature,
		Tools:       mapToOpenAITools(req.Tools),
	}

	// Map the ResponseFormat if it is configured.
	if req.ResponseFormat != nil {
		openaiReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatType(req.ResponseFormat.Type),
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Schema: req.ResponseFormat.JSONSchema.Schema,
				Strict: req.ResponseFormat.JSONSchema.Strict,
			},
		}
	}

	// Send the request to the OpenAI API.
	resp, err := o.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return ChatCompletionResponse{}, fmt.Errorf("openai error: %w", err)
	}

	// Map the OpenAI response into our internal unified format.
	return mapFromOpenAIResponse(resp), nil
}
