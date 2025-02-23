package syndicate

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

// dummyClient implements the LLMClient interface for testing.
type dummyClient struct{}

func (d dummyClient) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	message := Message{
		Role:    RoleAssistant,
		Content: "dummy response",
	}
	response := ChatCompletionResponse{
		Choices: []Choice{
			{
				Message:      message,
				FinishReason: FinishReasonStop,
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 8,
			TotalTokens:      18,
		},
	}
	return response, nil
}

func TestDummyLLMClient(t *testing.T) {
	client := dummyClient{}
	req := ChatCompletionRequest{
		Model:       "dummy-model",
		Messages:    []Message{{Role: RoleUser, Content: "Test message"}},
		Temperature: 0.7,
	}
	ctx := context.Background()

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}

	choice := resp.Choices[0]
	if choice.Message.Role != RoleAssistant {
		t.Errorf("expected role %s, got %s", RoleAssistant, choice.Message.Role)
	}
	if choice.Message.Content != "dummy response" {
		t.Errorf("expected content 'dummy response', got %s", choice.Message.Content)
	}
	if choice.FinishReason != FinishReasonStop {
		t.Errorf("expected finish reason %s, got %s", FinishReasonStop, choice.FinishReason)
	}

	if resp.Usage.PromptTokens != 10 || resp.Usage.CompletionTokens != 8 || resp.Usage.TotalTokens != 18 {
		t.Errorf("unexpected usage stats: %+v", resp.Usage)
	}
}

func TestToolCallJSONMarshaling(t *testing.T) {
	// Create a sample ToolCall with Args as JSON.
	sampleArgs := struct {
		Key string `json:"key"`
	}{
		Key: "value",
	}
	argsJSON, err := json.Marshal(sampleArgs)
	if err != nil {
		t.Fatalf("failed to marshal sampleArgs: %v", err)
	}

	toolCall := ToolCall{
		ID:   "123",
		Name: "sampleTool",
		Args: argsJSON,
	}

	// Marshal ToolCall back to JSON.
	marshaled, err := json.Marshal(toolCall)
	if err != nil {
		t.Fatalf("failed to marshal ToolCall: %v", err)
	}

	// Unmarshal into a map to check values.
	var result map[string]interface{}
	if err := json.Unmarshal(marshaled, &result); err != nil {
		t.Fatalf("failed to unmarshal marshaled ToolCall: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("expected id '123', got %v", result["id"])
	}
	if result["name"] != "sampleTool" {
		t.Errorf("expected name 'sampleTool', got %v", result["name"])
	}

	// Verify that the Args is valid JSON encoding the sampleArgs.
	argsResult, ok := result["args"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected args to be a JSON object, got %T", result["args"])
	}
	if !reflect.DeepEqual(argsResult, map[string]interface{}{"key": "value"}) {
		t.Errorf("unexpected args value: %v", argsResult)
	}
}

func TestResponseFormatJSONSchema(t *testing.T) {
	// Create a sample JSONSchema.
	schemaData := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{"field": map[string]interface{}{"type": "string"}},
	}
	schemaJSON, err := json.Marshal(schemaData)
	if err != nil {
		t.Fatalf("failed to marshal schemaData: %v", err)
	}

	jsonSchema := JSONSchema{
		Name:   "TestSchema",
		Schema: schemaJSON,
		Strict: true,
	}

	respFormat := ResponseFormat{
		Type:       "json_schema",
		JSONSchema: &jsonSchema,
	}

	// Verify fields.
	if respFormat.Type != "json_schema" {
		t.Errorf("expected ResponseFormat Type to be 'json_schema', got %s", respFormat.Type)
	}
	if respFormat.JSONSchema == nil {
		t.Fatal("expected JSONSchema to be non-nil")
	}
	if respFormat.JSONSchema.Name != "TestSchema" {
		t.Errorf("expected JSONSchema Name to be 'TestSchema', got %s", respFormat.JSONSchema.Name)
	}
	if !respFormat.JSONSchema.Strict {
		t.Error("expected JSONSchema Strict to be true")
	}
}
