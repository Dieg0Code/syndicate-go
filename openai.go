package gokamy

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the LLMClient interface using the OpenAI SDK.
// It wraps the official OpenAI client and provides a consistent interface
// for making chat completion requests.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIAzureClient creates an LLMClient for Azure using the provided API key and endpoint.
// It uses Azure specific configuration to initialize the underlying OpenAI client.
func NewOpenAIAzureClient(apiKey, endpoint string) LLMClient {
	config := openai.DefaultAzureConfig(apiKey, endpoint)
	return &OpenAIClient{
		client: openai.NewClientWithConfig(config),
	}
}

// NewOpenAIClient creates a new LLMClient using the provided API key.
// This client will use the standard OpenAI endpoint.
func NewOpenAIClient(apiKey string) LLMClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

// mapToOpenAIMessages converts a slice of internal Message structs into the format
// required by the OpenAI ChatCompletion API.
func mapToOpenAIMessages(messages []Message) []openai.ChatCompletionMessage {
	var msgs []openai.ChatCompletionMessage
	for _, m := range messages {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    m.Role,
			Name:    m.Name,
			Content: m.Content,
		})
	}
	return msgs
}

// mapToOpenAITools converts a slice of ToolDefinition structs into OpenAI Tools.
// These are used to pass function definitions to the OpenAI API.
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

// mapFromOpenAIResponse converts an OpenAI ChatCompletionResponse into the internal
// ChatCompletionResponse format used by the SDK.
// This includes mapping messages and token usage.
func mapFromOpenAIResponse(resp openai.ChatCompletionResponse) ChatCompletionResponse {
	var choices []Choice
	for _, c := range resp.Choices {
		choices = append(choices, Choice{
			Message: Message{
				Role:    c.Message.Role,
				Name:    c.Message.Name,
				Content: c.Message.Content,
				// Note: ToolCalls could be mapped here if provided by the API.
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

// CreateChatCompletion sends a chat completion request to the OpenAI API using the
// provided request parameters and returns a formatted ChatCompletionResponse.
// It converts internal message and tool definitions to OpenAI formats,
// sends the request, and maps the response back to the SDK structure.
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
	return mapFromOpenAIResponse(resp), nil
}
