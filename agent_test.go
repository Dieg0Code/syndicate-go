package syndicate

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
)

// ----- Fakes para Testing -----

// fakeMemory implementa la interfaz de una memoria simple para almacenar mensajes.
type fakeMemory struct {
	messages []Message
	mu       sync.Mutex
}

func (m *fakeMemory) Add(msg Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *fakeMemory) Get() []Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

func (m *fakeMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = nil
}

// fakeLLMClient simula la interfaz LLMClient.
// Permite configurar respuestas predefinidas que se consumen en cada llamada.
type fakeLLMClient struct {
	responses []ChatCompletionResponse
	callCount int
}

func (c *fakeLLMClient) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	if c.callCount >= len(c.responses) {
		return ChatCompletionResponse{}, errors.New("no hay más respuestas configuradas")
	}
	resp := c.responses[c.callCount]
	c.callCount++
	return resp, nil
}

// fakeTool implementa la interfaz Tool, permitiendo definir una función de ejecución personalizada.
type fakeTool struct {
	def      ToolDefinition
	execFunc func(args json.RawMessage) (interface{}, error)
}

func (t *fakeTool) GetDefinition() ToolDefinition {
	return t.def
}

func (t *fakeTool) Execute(args json.RawMessage) (interface{}, error) {
	if t.execFunc != nil {
		return t.execFunc(args)
	}
	return nil, nil
}

// fakeLLMClientWithError simula un error en la llamada al LLM.
type fakeLLMClientWithError struct{}

func (c *fakeLLMClientWithError) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	return ChatCompletionResponse{}, errors.New("simulated LLM error")
}

// ----- Tests Básicos -----

// TestGetSystemRole verifica que getSystemRole retorne el rol adecuado según el modelo.
func TestGetSystemRole(t *testing.T) {
	// Suponiendo que el modelo "o1mini" forma parte de los reasonerModels,
	// se espera que retorne RoleDeveloper.
	role := getSystemRole(openai.O1Mini)
	if role != RoleDeveloper {
		t.Errorf("se esperaba %s, se obtuvo %s", RoleDeveloper, role)
	}

	// Para un modelo no incluido se debe retornar el rol de sistema.
	role = getSystemRole("gpt-3.5-turbo")
	if role != RoleSystem {
		t.Errorf("se esperaba %s, se obtuvo %s", RoleSystem, role)
	}
}

// TestNewAgentWithRequiredFields verifica que NewAgent retorne error cuando faltan campos requeridos.
func TestNewAgentWithRequiredFields(t *testing.T) {
	// Sin client
	_, err := NewAgent(
		WithName("test"),
		WithMemory(&fakeMemory{}),
		WithModel("gpt-3.5-turbo"),
	)
	if err == nil || !strings.Contains(err.Error(), "client is required") {
		t.Errorf("se esperaba error por falta de client, se obtuvo: %v", err)
	}

	// Sin name
	_, err = NewAgent(
		WithClient(&fakeLLMClient{}),
		WithMemory(&fakeMemory{}),
		WithModel("gpt-3.5-turbo"),
	)
	if err == nil || !strings.Contains(err.Error(), "name is required") {
		t.Errorf("se esperaba error por falta de name, se obtuvo: %v", err)
	}

	// Sin memory
	_, err = NewAgent(
		WithClient(&fakeLLMClient{}),
		WithName("test"),
		WithModel("gpt-3.5-turbo"),
	)
	if err == nil || !strings.Contains(err.Error(), "memory is required") {
		t.Errorf("se esperaba error por falta de memory, se obtuvo: %v", err)
	}

	// Sin model
	_, err = NewAgent(
		WithClient(&fakeLLMClient{}),
		WithName("test"),
		WithMemory(&fakeMemory{}),
	)
	if err == nil || !strings.Contains(err.Error(), "model is required") {
		t.Errorf("se esperaba error por falta de model, se obtuvo: %v", err)
	}
}

// TestChatBasic verifica el flujo básico de Chat, simulando una respuesta simple del LLM.
func TestChatBasic(t *testing.T) {
	fakeClient := &fakeLLMClient{
		responses: []ChatCompletionResponse{
			{
				Choices: []Choice{
					{
						Message: Message{
							Content: "hello world",
						},
					},
				},
				Usage: Usage{},
			},
		},
	}
	mem := &fakeMemory{}
	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("agent1"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTemperature(0.5),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	result, err := agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Test input"),
	)
	if err != nil {
		t.Errorf("error inesperado: %v", err)
	}
	if result != "hello world" {
		t.Errorf("se esperaba 'hello world', se obtuvo '%s'", result)
	}

	// Se espera que en la memoria se hayan agregado al menos el mensaje del usuario y la respuesta del asistente.
	messages := mem.Get()
	if len(messages) < 2 {
		t.Errorf("se esperaban al menos 2 mensajes en la memoria, se obtuvieron %d", len(messages))
	}
}

// TestChatWithToolCall simula una situación en la que la primera respuesta solicita la ejecución de una herramienta.
func TestChatWithToolCall(t *testing.T) {
	// Configuramos una respuesta inicial que indique una llamada a herramienta.
	toolCall := ToolCall{
		ID:   "call1",
		Name: "fakeTool",
		Args: json.RawMessage(`"argumento"`),
	}
	responses := []ChatCompletionResponse{
		{
			Choices: []Choice{
				{
					Message: Message{
						ToolCalls: []ToolCall{toolCall},
					},
					FinishReason: FinishReasonToolCalls,
				},
			},
			Usage: Usage{},
		},
		{
			Choices: []Choice{
				{
					Message: Message{
						Content: "final result",
					},
				},
			},
			Usage: Usage{},
		},
	}
	fakeClient := &fakeLLMClient{
		responses: responses,
	}
	mem := &fakeMemory{}

	// Se registra una herramienta fake que retorna un resultado.
	testTool := &fakeTool{
		def: ToolDefinition{Name: "fakeTool", Description: "herramienta de prueba"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			return map[string]string{"result": "tool executed"}, nil
		},
	}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("agentWithTool"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTemperature(0.7),
		WithTool(testTool),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	result, err := agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Test tool call"),
	)
	if err != nil {
		t.Errorf("error inesperado: %v", err)
	}
	if result != "final result" {
		t.Errorf("se esperaba 'final result', se obtuvo '%s'", result)
	}

	// Verificamos que en la memoria se haya agregado el mensaje del resultado de la herramienta.
	foundToolMsg := false
	for _, msg := range mem.Get() {
		if msg.Role == RoleTool && msg.Name == "fakeTool" {
			if !strings.Contains(msg.Content, "tool executed") {
				t.Errorf("contenido inesperado en el resultado de la herramienta: %s", msg.Content)
			}
			foundToolMsg = true
		}
	}
	if !foundToolMsg {
		t.Error("no se encontró el mensaje del resultado de la herramienta en la memoria")
	}
}

// TestAgentWithFunctionalOptions verifica que los functional options configuren correctamente un agente.
func TestAgentWithFunctionalOptions(t *testing.T) {
	fakeClient := &fakeLLMClient{
		responses: []ChatCompletionResponse{
			{
				Choices: []Choice{
					{
						Message: Message{Content: "response"},
					},
				},
				Usage: Usage{},
			},
		},
	}
	mem := &fakeMemory{}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("functionalAgent"),
		WithDescription("test description"),
		WithSystemPrompt("functional prompt"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTemperature(0.9),
		WithTimeout(45*time.Second),
	)
	if err != nil {
		t.Fatalf("error inesperado al construir el agente: %v", err)
	}

	if agent.GetName() != "functionalAgent" {
		t.Errorf("se esperaba el nombre 'functionalAgent', se obtuvo '%s'", agent.GetName())
	}
}

// ----- Tests Adicionales para Robustez -----

// TestChatLLMClientError simula un escenario en el que el cliente LLM retorna error.
func TestChatLLMClientError(t *testing.T) {
	fakeClient := &fakeLLMClientWithError{}
	mem := &fakeMemory{}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("errAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTemperature(0.5),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	_, err = agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Test error handling"),
	)
	if err == nil || !strings.Contains(err.Error(), "simulated LLM error") {
		t.Errorf("se esperaba error de LLM, se obtuvo: %v", err)
	}
}

// TestNoResponseChoices simula el escenario en el que la respuesta del LLM no tiene choices.
func TestNoResponseChoices(t *testing.T) {
	fakeClient := &fakeLLMClient{
		responses: []ChatCompletionResponse{
			{
				Choices: []Choice{},
				Usage:   Usage{},
			},
		},
	}
	mem := &fakeMemory{}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("noChoiceAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTemperature(0.5),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	_, err = agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Test no choices"),
	)
	if err == nil || !strings.Contains(err.Error(), "no response choices available") {
		t.Errorf("se esperaba error por falta de choices, se obtuvo: %v", err)
	}
}

// TestToolExecutionError simula un error durante la ejecución de una herramienta.
func TestToolExecutionError(t *testing.T) {
	toolCall := ToolCall{
		ID:   "errorCall",
		Name: "errorTool",
		Args: json.RawMessage(`{}`),
	}
	responses := []ChatCompletionResponse{
		{
			Choices: []Choice{
				{
					Message: Message{
						ToolCalls: []ToolCall{toolCall},
					},
					FinishReason: FinishReasonToolCalls,
				},
			},
			Usage: Usage{},
		},
	}
	fakeClient := &fakeLLMClient{
		responses: responses,
	}
	mem := &fakeMemory{}

	// Registrar una herramienta que falla en la ejecución.
	errorTool := &fakeTool{
		def: ToolDefinition{Name: "errorTool", Description: "falla en ejecución"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			return nil, errors.New("tool execution failure")
		},
	}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("errorToolAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTemperature(0.5),
		WithTool(errorTool),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	_, err = agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Trigger tool error"),
	)
	if err == nil || !strings.Contains(err.Error(), "tool execution failure") {
		t.Errorf("se esperaba error en la ejecución de la herramienta, se obtuvo: %v", err)
	}
}

// TestMultipleToolCalls simula múltiples llamadas a herramientas concurrentes.
func TestMultipleToolCalls(t *testing.T) {
	toolCall1 := ToolCall{
		ID:   "call1",
		Name: "toolA",
		Args: json.RawMessage(`{}`),
	}
	toolCall2 := ToolCall{
		ID:   "call2",
		Name: "toolB",
		Args: json.RawMessage(`{}`),
	}
	responses := []ChatCompletionResponse{
		{
			Choices: []Choice{
				{
					Message: Message{
						ToolCalls: []ToolCall{toolCall1, toolCall2},
					},
					FinishReason: FinishReasonToolCalls,
				},
			},
			Usage: Usage{},
		},
		{
			Choices: []Choice{
				{
					Message: Message{
						Content: "multi tool final",
					},
					FinishReason: FinishReasonStop,
				},
			},
			Usage: Usage{},
		},
	}
	fakeClient := &fakeLLMClient{
		responses: responses,
	}
	mem := &fakeMemory{}

	toolA := &fakeTool{
		def: ToolDefinition{Name: "toolA", Description: "herramienta A"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return map[string]string{"result": "A executed"}, nil
		},
	}
	toolB := &fakeTool{
		def: ToolDefinition{Name: "toolB", Description: "herramienta B"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			time.Sleep(30 * time.Millisecond)
			return map[string]string{"result": "B executed"}, nil
		},
	}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("multiToolAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTemperature(0.5),
		WithTools(toolA, toolB),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	result, err := agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Test multiple tools"),
	)
	if err != nil {
		t.Errorf("error inesperado: %v", err)
	}
	if result != "multi tool final" {
		t.Errorf("se esperaba 'multi tool final', se obtuvo '%s'", result)
	}

	// Verificar que en la memoria se hayan agregado mensajes de ambas herramientas.
	foundA, foundB := false, false
	for _, msg := range mem.Get() {
		if msg.Role == RoleTool {
			if msg.Name == "toolA" && strings.Contains(msg.Content, "A executed") {
				foundA = true
			}
			if msg.Name == "toolB" && strings.Contains(msg.Content, "B executed") {
				foundB = true
			}
		}
	}
	if !foundA || !foundB {
		t.Errorf("no se encontraron ambas respuestas de herramienta, toolA: %v, toolB: %v", foundA, foundB)
	}
}

// TestChatWithTimeout verifica que el timeout personalizado funcione correctamente.
func TestChatWithTimeout(t *testing.T) {
	fakeClient := &fakeLLMClient{
		responses: []ChatCompletionResponse{
			{
				Choices: []Choice{
					{
						Message: Message{Content: "timeout test response"},
					},
				},
				Usage: Usage{},
			},
		},
	}
	mem := &fakeMemory{}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("timeoutAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	result, err := agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Test timeout"),
		WithChatTimeout(5*time.Second),
	)
	if err != nil {
		t.Errorf("error inesperado: %v", err)
	}
	if result != "timeout test response" {
		t.Errorf("se esperaba 'timeout test response', se obtuvo '%s'", result)
	}
}

// TestChatValidation verifica que la validación de parámetros funcione correctamente.
func TestChatValidation(t *testing.T) {
	fakeClient := &fakeLLMClient{}
	mem := &fakeMemory{}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("validationAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	// Test sin userName
	_, err = agent.Chat(context.Background(),
		WithInput("Test input"),
	)
	if err == nil || !strings.Contains(err.Error(), "user name is required") {
		t.Errorf("se esperaba error por falta de userName, se obtuvo: %v", err)
	}

	// Test sin input
	_, err = agent.Chat(context.Background(),
		WithUserName("user1"),
	)
	if err == nil || !strings.Contains(err.Error(), "input is required") {
		t.Errorf("se esperaba error por falta de input, se obtuvo: %v", err)
	}
}

// TestChatWithImages verifica que las imágenes se manejen correctamente.
func TestChatWithImages(t *testing.T) {
	fakeClient := &fakeLLMClient{
		responses: []ChatCompletionResponse{
			{
				Choices: []Choice{
					{
						Message: Message{Content: "image processed"},
					},
				},
				Usage: Usage{},
			},
		},
	}
	mem := &fakeMemory{}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("imageAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	result, err := agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Describe this image"),
		WithImages("https://example.com/image.jpg"),
	)
	if err != nil {
		t.Errorf("error inesperado: %v", err)
	}
	if result != "image processed" {
		t.Errorf("se esperaba 'image processed', se obtuvo '%s'", result)
	}

	// Verificar que el mensaje del usuario tenga las URLs de imágenes
	messages := mem.Get()
	found := false
	for _, msg := range messages {
		if msg.Role == RoleUser && len(msg.ImageURLs) > 0 {
			if msg.ImageURLs[0] == "https://example.com/image.jpg" {
				found = true
			}
		}
	}
	if !found {
		t.Error("no se encontró el mensaje del usuario con las URLs de imágenes")
	}
}

// TestAgentWithJSONResponseFormat verifica que el formato de respuesta JSON funcione.
func TestAgentWithJSONResponseFormat(t *testing.T) {
	type ResponseSchema struct {
		Result string `json:"result"`
	}

	fakeClient := &fakeLLMClient{
		responses: []ChatCompletionResponse{
			{
				Choices: []Choice{
					{
						Message: Message{Content: `{"result": "json response"}`},
					},
				},
				Usage: Usage{},
			},
		},
	}
	mem := &fakeMemory{}

	agent, err := NewAgent(
		WithClient(fakeClient),
		WithName("jsonAgent"),
		WithMemory(mem),
		WithModel("gpt-3.5-turbo"),
		WithJSONResponseFormat("test_schema", ResponseSchema{}),
	)
	if err != nil {
		t.Fatalf("error creando agente: %v", err)
	}

	result, err := agent.Chat(context.Background(),
		WithUserName("user1"),
		WithInput("Return JSON response"),
	)
	if err != nil {
		t.Errorf("error inesperado: %v", err)
	}
	if result != `{"result": "json response"}` {
		t.Errorf("se esperaba respuesta JSON, se obtuvo '%s'", result)
	}
}
