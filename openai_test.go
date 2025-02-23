package syndicate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// TestMapToOpenAIMessages verifica que se transformen correctamente los mensajes internos a OpenAI.
func TestMapToOpenAIMessages(t *testing.T) {
	messages := []Message{
		{
			Role:    "user",
			Name:    "testUser",
			Content: "Hello, world!",
			ToolID:  "tool1",
		},
	}
	openaiMsgs := mapToOpenAIMessages(messages)
	if len(openaiMsgs) != 1 {
		t.Fatalf("se esperaba 1 mensaje, se obtuvo %d", len(openaiMsgs))
	}
	if openaiMsgs[0].Role != "user" {
		t.Errorf("se esperaba role 'user', se obtuvo '%s'", openaiMsgs[0].Role)
	}
	if openaiMsgs[0].Name != "testUser" {
		t.Errorf("se esperaba name 'testUser', se obtuvo '%s'", openaiMsgs[0].Name)
	}
	if openaiMsgs[0].Content != "Hello, world!" {
		t.Errorf("se esperaba content 'Hello, world!', se obtuvo '%s'", openaiMsgs[0].Content)
	}
	if openaiMsgs[0].ToolCallID != "tool1" {
		t.Errorf("se esperaba ToolCallID 'tool1', se obtuvo '%s'", openaiMsgs[0].ToolCallID)
	}
}

// TestMapToOpenAITools verifica la conversión de las definiciones internas de herramienta a las de OpenAI.
func TestMapToOpenAITools(t *testing.T) {
	tools := []ToolDefinition{
		{
			Name:        "testTool",
			Description: "A test tool",
			Parameters:  json.RawMessage(`{"type": "object"}`),
		},
	}
	openaiTools := mapToOpenAITools(tools)
	if len(openaiTools) != 1 {
		t.Fatalf("se esperaba 1 herramienta, se obtuvo %d", len(openaiTools))
	}
	tool := openaiTools[0]
	if tool.Type != openai.ToolTypeFunction {
		t.Errorf("se esperaba tool type '%s', se obtuvo '%s'", openai.ToolTypeFunction, tool.Type)
	}
	if tool.Function == nil {
		t.Fatal("se esperaba que Function no fuera nil")
	}
	if tool.Function.Name != "testTool" {
		t.Errorf("se esperaba name 'testTool', se obtuvo '%s'", tool.Function.Name)
	}
	if tool.Function.Description != "A test tool" {
		t.Errorf("se esperaba description 'A test tool', se obtuvo '%s'", tool.Function.Description)
	}
	params, ok := tool.Function.Parameters.(json.RawMessage)
	if !ok {
		t.Fatalf("se esperaba json.RawMessage, se obtuvo %T", tool.Function.Parameters)
	}
	if string(params) != `{"type": "object"}` {
		t.Errorf("se esperaba parameters '{\"type\": \"object\"}', se obtuvo '%s'", string(params))
	}
	if !tool.Function.Strict {
		t.Error("se esperaba Strict en true")
	}
}

// TestMapFromOpenAIToolCalls verifica que se mapeen correctamente las llamadas a herramientas de OpenAI.
func TestMapFromOpenAIToolCalls(t *testing.T) {
	openaiToolCalls := []openai.ToolCall{
		{
			ID: "call1",
			Function: openai.FunctionCall{
				Name:      "testTool",
				Arguments: `{"arg":"value"}`,
			},
		},
	}
	toolCalls := mapFromOpenAIToolCalls(openaiToolCalls)
	if len(toolCalls) != 1 {
		t.Fatalf("se esperaba 1 tool call, se obtuvo %d", len(toolCalls))
	}
	tc := toolCalls[0]
	if tc.ID != "call1" {
		t.Errorf("se esperaba ID 'call1', se obtuvo '%s'", tc.ID)
	}
	if tc.Name != "testTool" {
		t.Errorf("se esperaba Name 'testTool', se obtuvo '%s'", tc.Name)
	}
	if string(tc.Args) != `{"arg":"value"}` {
		t.Errorf("se esperaba Args '{\"arg\":\"value\"}', se obtuvo '%s'", string(tc.Args))
	}
}

// TestMapFromOpenAIResponse verifica que se mapee correctamente la respuesta de OpenAI a la estructura interna.
func TestMapFromOpenAIResponse(t *testing.T) {
	fakeResp := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Name:    "assistant",
					Content: "test response",
					ToolCalls: []openai.ToolCall{
						{
							ID: "call1",
							Function: openai.FunctionCall{
								Name:      "testTool",
								Arguments: `{"arg":"value"}`,
							},
						},
					},
				},
				FinishReason: openai.FinishReasonStop,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}
	internalResp := mapFromOpenAIResponse(fakeResp)
	if len(internalResp.Choices) != 1 {
		t.Fatalf("se esperaba 1 choice, se obtuvo %d", len(internalResp.Choices))
	}
	choice := internalResp.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("se esperaba role 'assistant', se obtuvo '%s'", choice.Message.Role)
	}
	if choice.Message.Name != "assistant" {
		t.Errorf("se esperaba name 'assistant', se obtuvo '%s'", choice.Message.Name)
	}
	if choice.Message.Content != "test response" {
		t.Errorf("se esperaba content 'test response', se obtuvo '%s'", choice.Message.Content)
	}
	if choice.FinishReason != "stop" {
		t.Errorf("se esperaba finish reason 'stop', se obtuvo '%s'", choice.FinishReason)
	}
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("se esperaba 1 tool call, se obtuvo %d", len(choice.Message.ToolCalls))
	}
	tc := choice.Message.ToolCalls[0]
	if tc.ID != "call1" {
		t.Errorf("se esperaba tool call ID 'call1', se obtuvo '%s'", tc.ID)
	}
	if tc.Name != "testTool" {
		t.Errorf("se esperaba tool call Name 'testTool', se obtuvo '%s'", tc.Name)
	}
	if string(tc.Args) != `{"arg":"value"}` {
		t.Errorf("se esperaba tool call Args '{\"arg\":\"value\"}', se obtuvo '%s'", string(tc.Args))
	}
	if internalResp.Usage.PromptTokens != 10 || internalResp.Usage.CompletionTokens != 20 || internalResp.Usage.TotalTokens != 30 {
		t.Errorf("usage inesperado: %+v", internalResp.Usage)
	}
}

// TestCreateChatCompletion simula una llamada a la API de OpenAI mediante un servidor HTTP fake.
func TestCreateChatCompletion(t *testing.T) {
	// Servidor fake para simular la API de OpenAI.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Se puede inspeccionar r.Body si se desea verificar el request.
		response := `{
			"choices": [{
				"message": {
					"role": "assistant",
					"name": "assistant",
					"content": "test response",
					"tool_calls": []
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 20,
				"total_tokens": 30
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Configuramos el cliente OpenAI para usar el servidor fake.
	apiKey := "test-api-key"
	config := openai.DefaultAzureConfig(apiKey, server.URL)
	fakeClient := openai.NewClientWithConfig(config)

	// Creamos la instancia de OpenAIClient con el cliente fake.
	client := &OpenAIClient{
		client: fakeClient,
	}

	// Preparamos el request.
	req := ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Name:    "tester",
				Content: "Hello",
			},
		},
		Temperature: 0.7,
		Tools: []ToolDefinition{
			{
				Name:        "testTool",
				Description: "A test tool",
				Parameters:  json.RawMessage(`{"type": "object"}`),
			},
		},
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &JSONSchema{
				Name:   "TestSchema",
				Schema: json.RawMessage(`{"type": "object"}`),
				Strict: true,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion retornó error: %v", err)
	}
	// Verificamos que se mapee correctamente la respuesta.
	if len(resp.Choices) != 1 {
		t.Fatalf("se esperaba 1 choice, se obtuvo %d", len(resp.Choices))
	}
	choice := resp.Choices[0]
	if !strings.EqualFold(choice.Message.Content, "test response") {
		t.Errorf("se esperaba message content 'test response', se obtuvo '%s'", choice.Message.Content)
	}
	if resp.Usage.PromptTokens != 10 || resp.Usage.CompletionTokens != 20 || resp.Usage.TotalTokens != 30 {
		t.Errorf("usage inesperado: %+v", resp.Usage)
	}
}
