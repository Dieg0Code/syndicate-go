package syndicate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// TestMapToDeepseekMessages verifica que se conviertan correctamente los mensajes internos a deepseek.ChatCompletionMessage.
// Se debe cambiar el role "system" por "user" y mantener los demás sin modificación.
func TestMapToDeepseekMessages(t *testing.T) {
	messages := []Message{
		{
			Role:    RoleSystem, // se espera que se convierta a "user"
			Content: "Mensaje de sistema",
		},
		{
			Role:    "user",
			Content: "Mensaje de usuario",
		},
		{
			Role:    "assistant",
			Content: "Mensaje del asistente",
		},
	}

	dsMsgs := mapToDeepseekMessages(messages)
	if len(dsMsgs) != len(messages) {
		t.Fatalf("se esperaban %d mensajes, se obtuvo %d", len(messages), len(dsMsgs))
	}

	if !strings.EqualFold(dsMsgs[0].Role, RoleUser) {
		t.Errorf("se esperaba que el role '%s' se convirtiera a '%s', se obtuvo '%s'", RoleSystem, RoleUser, dsMsgs[0].Role)
	}
	if dsMsgs[1].Role != "user" {
		t.Errorf("se esperaba role 'user', se obtuvo '%s'", dsMsgs[1].Role)
	}
	if dsMsgs[2].Role != "assistant" {
		t.Errorf("se esperaba role 'assistant', se obtuvo '%s'", dsMsgs[2].Role)
	}
}

// TestMapFromDeepseekResponse verifica que se mapee correctamente una respuesta de Deepseek a la estructura interna.
func TestMapFromDeepseekResponse(t *testing.T) {
	// Creamos una respuesta simulada de Deepseek.
	fakeResp := &deepseek.ChatCompletionResponse{
		Choices: []deepseek.Choice{
			{
				Message: deepseek.Message{
					Role:    "assistant",
					Content: "Respuesta de prueba",
				},
			},
		},
		Usage: deepseek.Usage{
			PromptTokens:     15,
			CompletionTokens: 25,
			TotalTokens:      40,
		},
	}

	internalResp := mapFromDeepseekResponse(fakeResp)
	if len(internalResp.Choices) != 1 {
		t.Fatalf("se esperaba 1 choice, se obtuvo %d", len(internalResp.Choices))
	}
	choice := internalResp.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("se esperaba role 'assistant', se obtuvo '%s'", choice.Message.Role)
	}
	if choice.Message.Content != "Respuesta de prueba" {
		t.Errorf("se esperaba content 'Respuesta de prueba', se obtuvo '%s'", choice.Message.Content)
	}
	// finishReason siempre se fija a FinishReasonStop
	if choice.FinishReason != FinishReasonStop {
		t.Errorf("se esperaba finish reason '%s', se obtuvo '%s'", FinishReasonStop, choice.FinishReason)
	}
	if internalResp.Usage.PromptTokens != 15 ||
		internalResp.Usage.CompletionTokens != 25 ||
		internalResp.Usage.TotalTokens != 40 {
		t.Errorf("uso inesperado: %+v", internalResp.Usage)
	}
}

// TestCreateChatCompletion simula la llamada a la API de DeepseekR1 mediante un servidor HTTP fake.
// Se valida que el método CreateChatCompletion convierta correctamente el request y mapee la respuesta.
func TestDeepSeekCreateChatCompletion(t *testing.T) {
	// Definimos un servidor HTTP fake para simular la API de DeepseekR1.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Opcional: podemos inspeccionar r.Body para validar el request recibido.
		// Respuesta simulada en formato JSON que se ajusta a deepseek.ChatCompletionResponse.
		response := map[string]interface{}{
			"id": "dummy-id",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "test response",
					},
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil {
			t.Fatalf("error al escribir la respuesta: %v", err)
		}
	}))
	defer server.Close()

	// Creamos un cliente DeepseekR1 usando el servidor fake.
	apiKey := "dummy-api-key"
	clientInterface := NewDeepseekR1Client(apiKey, server.URL+"/")
	dsClient, ok := clientInterface.(*DeepseekR1Client)
	if !ok {
		t.Fatal("no se pudo convertir el cliente a *DeepseekR1Client")
	}

	// Preparamos un ChatCompletionRequest.
	req := ChatCompletionRequest{
		Model: "deepseek-model-v1",
		Messages: []Message{
			{
				Role:    RoleSystem, // se convertirá a "user" en el mapping
				Content: "Mensaje de prueba",
			},
		},
		// Tools y ResponseFormat se ignoran para DeepseekR1.
	}

	// Ejecutamos la llamada a CreateChatCompletion.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := dsClient.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion retornó error: %v", err)
	}

	// Verificamos que la respuesta se haya mapeado correctamente.
	if len(resp.Choices) != 1 {
		t.Fatalf("se esperaba 1 choice, se obtuvo %d", len(resp.Choices))
	}
	choice := resp.Choices[0]
	if !strings.EqualFold(choice.Message.Content, "test response") {
		t.Errorf("se esperaba content 'test response', se obtuvo '%s'", choice.Message.Content)
	}
	if resp.Usage.PromptTokens != 10 || resp.Usage.CompletionTokens != 20 || resp.Usage.TotalTokens != 30 {
		t.Errorf("uso inesperado: %+v", resp.Usage)
	}
}
