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

// TestAddTool verifica que al agregar una herramienta el agente la incluya en su mapa.
func TestAddTool(t *testing.T) {
	agent := &BaseAgent{
		tools:  make(map[string]Tool),
		memory: &fakeMemory{},
	}
	myTool := &fakeTool{
		def: ToolDefinition{Name: "testTool", Description: "herramienta de prueba"},
	}
	agent.AddTool(myTool)
	if _, exists := agent.tools["testTool"]; !exists {
		t.Error("la herramienta no fue agregada al agente")
	}
}

// TestSetConfigPrompt verifica que se asigne correctamente el prompt de configuración.
func TestSetConfigPrompt(t *testing.T) {
	agent := &BaseAgent{}
	prompt := "Este es un prompt de sistema"
	agent.SetConfigPrompt(prompt)
	if agent.systemPrompt != prompt {
		t.Errorf("se esperaba prompt '%s', se obtuvo '%s'", prompt, agent.systemPrompt)
	}
}

// TestProcess verifica el flujo básico de Process, simulando una respuesta simple del LLM.
func TestProcess(t *testing.T) {
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
	agent := &BaseAgent{
		client:      fakeClient,
		name:        "agent1",
		tools:       make(map[string]Tool),
		memory:      mem,
		model:       "gpt-3.5-turbo",
		temperature: 0.5,
	}
	result, err := agent.Process(context.Background(), "user1", "Test input")
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

// TestProcessWithToolCall simula una situación en la que la primera respuesta solicita la ejecución de una herramienta.
func TestProcessWithToolCall(t *testing.T) {
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
	agent := &BaseAgent{
		client:      fakeClient,
		name:        "agentWithTool",
		tools:       make(map[string]Tool),
		memory:      mem,
		model:       "gpt-3.5-turbo",
		temperature: 0.7,
	}
	// Se registra una herramienta fake que retorna un resultado.
	agent.tools["fakeTool"] = &fakeTool{
		def: ToolDefinition{Name: "fakeTool", Description: "herramienta de prueba"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			return map[string]string{"result": "tool executed"}, nil
		},
	}
	result, err := agent.Process(context.Background(), "user1", "Test tool call")
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

// TestAgentBuilder verifica que el AgentBuilder configure correctamente un agente.
func TestAgentBuilder(t *testing.T) {
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
	builder := NewAgent().
		SetClient(fakeClient).
		SetName("builderAgent").
		SetConfigPrompt("builder prompt").
		SetMemory(mem).
		SetModel("gpt-3.5-turbo").
		SetTemperature(0.9)
	agent, err := builder.Build()
	if err != nil {
		t.Fatalf("error inesperado al construir el agente: %v", err)
	}
	if agent.name != "builderAgent" {
		t.Errorf("se esperaba el nombre 'builderAgent', se obtuvo '%s'", agent.name)
	}
	if agent.systemPrompt != "builder prompt" {
		t.Errorf("se esperaba el prompt 'builder prompt', se obtuvo '%s'", agent.systemPrompt)
	}
	if agent.model != "gpt-3.5-turbo" {
		t.Errorf("se esperaba el modelo 'gpt-3.5-turbo', se obtuvo '%s'", agent.model)
	}
	if agent.temperature != 0.9 {
		t.Errorf("se esperaba la temperatura 0.9, se obtuvo %f", agent.temperature)
	}
}

// ----- Tests Adicionales para Robustez -----

// TestProcessLLMClientError simula un escenario en el que el cliente LLM retorna error.
func TestProcessLLMClientError(t *testing.T) {
	fakeClient := &fakeLLMClientWithError{}
	mem := &fakeMemory{}
	agent := &BaseAgent{
		client:      fakeClient,
		name:        "errAgent",
		tools:       make(map[string]Tool),
		memory:      mem,
		model:       "gpt-3.5-turbo",
		temperature: 0.5,
	}
	_, err := agent.Process(context.Background(), "user1", "Test error handling")
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
	agent := &BaseAgent{
		client:      fakeClient,
		name:        "noChoiceAgent",
		tools:       make(map[string]Tool),
		memory:      mem,
		model:       "gpt-3.5-turbo",
		temperature: 0.5,
	}
	_, err := agent.Process(context.Background(), "user1", "Test no choices")
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
	agent := &BaseAgent{
		client:      fakeClient,
		name:        "errorToolAgent",
		tools:       make(map[string]Tool),
		memory:      mem,
		model:       "gpt-3.5-turbo",
		temperature: 0.5,
	}
	// Registrar una herramienta que falla en la ejecución.
	agent.tools["errorTool"] = &fakeTool{
		def: ToolDefinition{Name: "errorTool", Description: "falla en ejecución"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			return nil, errors.New("tool execution failure")
		},
	}
	_, err := agent.Process(context.Background(), "user1", "Trigger tool error")
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
	agent := &BaseAgent{
		client:      fakeClient,
		name:        "multiToolAgent",
		tools:       make(map[string]Tool),
		memory:      mem,
		model:       "gpt-3.5-turbo",
		temperature: 0.5,
	}
	agent.tools["toolA"] = &fakeTool{
		def: ToolDefinition{Name: "toolA", Description: "herramienta A"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return map[string]string{"result": "A executed"}, nil
		},
	}
	agent.tools["toolB"] = &fakeTool{
		def: ToolDefinition{Name: "toolB", Description: "herramienta B"},
		execFunc: func(args json.RawMessage) (interface{}, error) {
			time.Sleep(30 * time.Millisecond)
			return map[string]string{"result": "B executed"}, nil
		},
	}
	result, err := agent.Process(context.Background(), "user1", "Test multiple tools")
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
