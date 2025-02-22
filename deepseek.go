package syndicate

import (
	"context"
	"fmt"
	"strings"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// DeepseekR1Client implementa LLMClient usando el SDK de DeepseekR1.
type DeepseekR1Client struct {
	client *deepseek.Client
}

// NewDeepseekR1Client crea un nuevo cliente para DeepseekR1.
// Recibe la API key y el baseURL (por ejemplo, "https://models.inference.ai.azure.com/").
func NewDeepseekR1Client(apiKey, baseURL string) LLMClient {
	return &DeepseekR1Client{
		client: deepseek.NewClient(apiKey, baseURL),
	}
}

// mapToDeepseekMessages convierte nuestro []Message a []deepseek.ChatCompletionMessage,
// cambiando el role "system" por "user" para evitar problemas.
func mapToDeepseekMessages(messages []Message) []deepseek.ChatCompletionMessage {
	msgs := make([]deepseek.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		role := m.Role
		if strings.EqualFold(m.Role, RoleSystem) {
			role = RoleUser
		}
		msgs[i] = deepseek.ChatCompletionMessage{
			Role:    role,
			Content: m.Content,
		}
	}
	return msgs
}

// mapFromDeepseekResponse convierte la respuesta de Deepseek en ChatCompletionResponse.
func mapFromDeepseekResponse(resp *deepseek.ChatCompletionResponse) ChatCompletionResponse {
	var choices []Choice
	for _, c := range resp.Choices {
		choices = append(choices, Choice{
			Message: Message{
				Role:    c.Message.Role,
				Content: c.Message.Content,
			},
			FinishReason: FinishReasonStop,
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

// CreateChatCompletion envía la solicitud de chat a DeepseekR1.
// Se ignoran tools y ResponseFormat, ya que DeepseekR1 no los soporta.
func (d *DeepseekR1Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	deepseekReq := &deepseek.ChatCompletionRequest{
		Model:    req.Model,
		Messages: mapToDeepseekMessages(req.Messages),
		// DeepseekR1 no soporta tools ni parámetros como Temperature o ResponseFormat.
	}

	resp, err := d.client.CreateChatCompletion(ctx, deepseekReq)
	if err != nil {
		return ChatCompletionResponse{}, fmt.Errorf("deepseek error: %w", err)
	}

	return mapFromDeepseekResponse(resp), nil
}
