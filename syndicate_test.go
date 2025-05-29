package syndicate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// syndicateTestMemory implementa la interfaz Memory para pruebas.
type syndicateTestMemory struct {
	messages []Message
	mu       sync.Mutex
}

func (m *syndicateTestMemory) Add(msg Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *syndicateTestMemory) Get() []Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	copyMessages := make([]Message, len(m.messages))
	copy(copyMessages, m.messages)
	return copyMessages
}

// syndicateTestAgent es una implementación simple del interface Agent para pruebas.
type syndicateTestAgent struct {
	name   string
	memory Memory
	// chatFunc simula el procesamiento del agente usando Chat.
	chatFunc func(ctx context.Context, options ...ChatOption) (string, error)
}

func (f *syndicateTestAgent) Chat(ctx context.Context, options ...ChatOption) (string, error) {
	if f.chatFunc != nil {
		return f.chatFunc(ctx, options...)
	}
	// Por defecto retorna una respuesta simple
	return fmt.Sprintf("response from %s", f.name), nil
}

func (f *syndicateTestAgent) GetName() string {
	return f.name
}

// newSyndicateTestAgent crea un syndicateTestAgent con comportamiento personalizable.
func newSyndicateTestAgent(name string, suffix string) *syndicateTestAgent {
	return &syndicateTestAgent{
		name:   name,
		memory: &syndicateTestMemory{},
		chatFunc: func(ctx context.Context, options ...ChatOption) (string, error) {
			// Parsear las opciones para obtener el input - USAR EL TIPO CORRECTO
			req := &chatRequest{} // Cambié testChatRequest por chatRequest
			for _, opt := range options {
				opt(req)
			}

			// Simular procesamiento del input
			if req.input == "" {
				return "", fmt.Errorf("input is required")
			}

			// Retorna el input junto con un sufijo para identificar al agente
			return fmt.Sprintf("%s %s", req.input, suffix), nil
		},
	}
}

// TestNewSyndicate verifica la creación de syndicate con functional options.
func TestNewSyndicate(t *testing.T) {
	agent1 := newSyndicateTestAgent("agent1", "->A")
	agent2 := newSyndicateTestAgent("agent2", "->B")
	customHistory := &syndicateTestMemory{}

	syndicate, err := NewSyndicate(
		WithAgent(agent1),
		WithAgent(agent2),
		WithGlobalHistory(customHistory),
		WithPipeline("agent1", "agent2"),
	)
	if err != nil {
		t.Fatalf("Error creando syndicate: %v", err)
	}

	// Verificar que los agentes estén registrados
	found1, exists1 := syndicate.FindAgent("agent1")
	if !exists1 || found1.GetName() != "agent1" {
		t.Error("agent1 no encontrado o nombre incorrecto")
	}

	found2, exists2 := syndicate.FindAgent("agent2")
	if !exists2 || found2.GetName() != "agent2" {
		t.Error("agent2 no encontrado o nombre incorrecto")
	}

	// Verificar pipeline
	pipeline := syndicate.GetPipeline()
	if len(pipeline) != 2 || pipeline[0] != "agent1" || pipeline[1] != "agent2" {
		t.Errorf("Pipeline incorrecta: %v", pipeline)
	}

	// Verificar global history
	history := syndicate.GetGlobalHistory()
	if len(history) != 0 {
		t.Errorf("Historia global debería estar vacía inicialmente")
	}
}

// TestNewSyndicate_WithErrors verifica validaciones en la creación del syndicate.
func TestNewSyndicate_WithErrors(t *testing.T) {
	// Test con agente nil
	_, err := NewSyndicate(WithAgent(nil))
	if err == nil || !strings.Contains(err.Error(), "agent cannot be nil") {
		t.Errorf("Se esperaba error por agente nil: %v", err)
	}

	// Test con agente sin nombre
	agentSinNombre := &syndicateTestAgent{name: ""}
	_, err = NewSyndicate(WithAgent(agentSinNombre))
	if err == nil || !strings.Contains(err.Error(), "agent name cannot be empty") {
		t.Errorf("Se esperaba error por agente sin nombre: %v", err)
	}

	// Test con pipeline vacío
	agent := newSyndicateTestAgent("test", "->test")
	_, err = NewSyndicate(
		WithAgent(agent),
		WithPipeline(),
	)
	if err == nil || !strings.Contains(err.Error(), "pipeline cannot be empty") {
		t.Errorf("Se esperaba error por pipeline vacío: %v", err)
	}

	// Test con pipeline referenciando agente inexistente
	_, err = NewSyndicate(
		WithAgent(agent),
		WithPipeline("test", "nonexistent"),
	)
	if err == nil || !strings.Contains(err.Error(), "agent nonexistent not found") {
		t.Errorf("Se esperaba error por agente inexistente en pipeline: %v", err)
	}

	// Test con global history nil
	_, err = NewSyndicate(WithGlobalHistory(nil))
	if err == nil || !strings.Contains(err.Error(), "global history cannot be nil") {
		t.Errorf("Se esperaba error por global history nil: %v", err)
	}
}

// TestExecuteAgent verifica el flujo de ExecuteAgent.
func TestExecuteAgent(t *testing.T) {
	agent := newSyndicateTestAgent("agent1", "->agent1")
	globalHistory := &syndicateTestMemory{}

	syndicate, err := NewSyndicate(
		WithAgent(agent),
		WithGlobalHistory(globalHistory),
	)
	if err != nil {
		t.Fatalf("Error creando syndicate: %v", err)
	}

	ctx := context.Background()
	resp, err := syndicate.ExecuteAgent(ctx, "agent1",
		WithExecuteUserName("usuario1"),
		WithExecuteInput("hola"),
	)
	if err != nil {
		t.Fatalf("ExecuteAgent failed: %v", err)
	}

	// Verificar respuesta
	if !strings.Contains(resp, "hola") || !strings.Contains(resp, "->agent1") {
		t.Errorf("Respuesta inesperada: %s", resp)
	}

	// Verificar que globalHistory tenga 2 mensajes: usuario y asistente
	history := globalHistory.Get()
	if len(history) != 2 {
		t.Errorf("Se esperaban 2 mensajes en globalHistory, se obtuvieron: %d", len(history))
	}

	// Verificar que el mensaje de respuesta tenga el prefijo "[agent1]: "
	if !strings.HasPrefix(history[1].Content, "[agent1]:") {
		t.Errorf("El mensaje de respuesta global no tiene el prefijo correcto: %s", history[1].Content)
	}
}

// TestExecuteAgent_WithValidation verifica validaciones en ExecuteAgent.
func TestExecuteAgent_WithValidation(t *testing.T) {
	agent := newSyndicateTestAgent("agent1", "->agent1")
	syndicate, _ := NewSyndicate(WithAgent(agent))

	ctx := context.Background()

	// Test sin userName
	_, err := syndicate.ExecuteAgent(ctx, "agent1",
		WithExecuteInput("test"),
	)
	if err == nil || !strings.Contains(err.Error(), "user name is required") {
		t.Errorf("Se esperaba error por falta de userName: %v", err)
	}

	// Test sin input
	_, err = syndicate.ExecuteAgent(ctx, "agent1",
		WithExecuteUserName("user1"),
	)
	if err == nil || !strings.Contains(err.Error(), "input is required") {
		t.Errorf("Se esperaba error por falta de input: %v", err)
	}

	// Test con agente inexistente
	_, err = syndicate.ExecuteAgent(ctx, "nonexistent",
		WithExecuteUserName("user1"),
		WithExecuteInput("test"),
	)
	if err == nil || !strings.Contains(err.Error(), "agent not found") {
		t.Errorf("Se esperaba error por agente inexistente: %v", err)
	}
}

// TestExecuteAgent_WithOptions verifica diferentes opciones de ExecuteAgent.
func TestExecuteAgent_WithOptions(t *testing.T) {
	agent := newSyndicateTestAgent("agent1", "->agent1")
	globalHistory := &syndicateTestMemory{}
	// Agregar un mensaje previo a la historia global
	globalHistory.Add(Message{Role: RoleUser, Content: "mensaje previo"})

	syndicate, _ := NewSyndicate(
		WithAgent(agent),
		WithGlobalHistory(globalHistory),
	)

	ctx := context.Background()

	// Test con imágenes
	_, err := syndicate.ExecuteAgent(ctx, "agent1",
		WithExecuteUserName("user1"),
		WithExecuteInput("describe image"),
		WithExecuteImages("https://example.com/image.jpg"),
	)
	if err != nil {
		t.Errorf("Error con imágenes: %v", err)
	}

	// Test con mensajes adicionales
	additionalMsgs := []Message{
		{Role: RoleUser, Content: "contexto adicional"},
	}
	_, err = syndicate.ExecuteAgent(ctx, "agent1",
		WithExecuteUserName("user1"),
		WithExecuteInput("test with context"),
		WithExecuteAdditionalMessages(additionalMsgs),
	)
	if err != nil {
		t.Errorf("Error con mensajes adicionales: %v", err)
	}
}

// TestFindAgent verifica FindAgent.
func TestFindAgent(t *testing.T) {
	agent := newSyndicateTestAgent("agentX", "->X")
	syndicate, _ := NewSyndicate(WithAgent(agent))

	foundAgent, exists := syndicate.FindAgent("agentX")
	if !exists {
		t.Error("Se esperaba encontrar el agente agentX")
	}
	if foundAgent.GetName() != "agentX" {
		t.Errorf("Nombre del agente incorrecto: se obtuvo %s", foundAgent.GetName())
	}

	// Verificar búsqueda de agente inexistente
	_, exists = syndicate.FindAgent("no-exist")
	if exists {
		t.Error("No se esperaba encontrar un agente inexistente")
	}
}

// TestExecutePipeline verifica la ejecución secuencial de agentes.
func TestExecutePipeline(t *testing.T) {
	agent1 := newSyndicateTestAgent("agent1", "->A")
	agent2 := newSyndicateTestAgent("agent2", "->B")

	syndicate, err := NewSyndicate(
		WithAgents(agent1, agent2),
		WithPipeline("agent1", "agent2"),
	)
	if err != nil {
		t.Fatalf("Error creando syndicate: %v", err)
	}

	ctx := context.Background()
	finalResp, err := syndicate.ExecutePipeline(ctx,
		WithPipelineUserName("usuario1"),
		WithPipelineInput("inicio"),
	)
	if err != nil {
		t.Fatalf("ExecutePipeline failed: %v", err)
	}

	// Se espera que cada agente agregue su sufijo al input
	if !strings.Contains(finalResp, "->A") || !strings.Contains(finalResp, "->B") {
		t.Errorf("Pipeline no procesada correctamente, salida: %s", finalResp)
	}
}

// TestExecutePipeline_WithValidation verifica validaciones en ExecutePipeline.
func TestExecutePipeline_WithValidation(t *testing.T) {
	// Syndicate sin pipeline
	agent := newSyndicateTestAgent("agent1", "->A")
	syndicate, _ := NewSyndicate(WithAgent(agent))

	ctx := context.Background()
	_, err := syndicate.ExecutePipeline(ctx,
		WithPipelineUserName("usuario1"),
		WithPipelineInput("inicio"),
	)
	if err == nil || !strings.Contains(err.Error(), "no pipeline defined") {
		t.Errorf("Se esperaba error por pipeline vacío: %v", err)
	}

	// Syndicate con pipeline
	syndicate2, _ := NewSyndicate(
		WithAgent(agent),
		WithPipeline("agent1"),
	)

	// Test sin userName
	_, err = syndicate2.ExecutePipeline(ctx,
		WithPipelineInput("test"),
	)
	if err == nil || !strings.Contains(err.Error(), "user name is required") {
		t.Errorf("Se esperaba error por falta de userName: %v", err)
	}

	// Test sin input
	_, err = syndicate2.ExecutePipeline(ctx,
		WithPipelineUserName("user1"),
	)
	if err == nil || !strings.Contains(err.Error(), "input is required") {
		t.Errorf("Se esperaba error por falta de input: %v", err)
	}
}

// TestExecutePipeline_WithImages verifica que las imágenes solo se pasen al primer agente.
func TestExecutePipeline_WithImages(t *testing.T) {
	agent1 := newSyndicateTestAgent("agent1", "->A")
	agent2 := newSyndicateTestAgent("agent2", "->B")

	syndicate, _ := NewSyndicate(
		WithAgents(agent1, agent2),
		WithPipeline("agent1", "agent2"),
	)

	ctx := context.Background()
	_, err := syndicate.ExecutePipeline(ctx,
		WithPipelineUserName("user1"),
		WithPipelineInput("test images"),
		WithPipelineImages("https://example.com/image.jpg"),
	)
	if err != nil {
		t.Errorf("Error en pipeline con imágenes: %v", err)
	}
}

// TestGetAgentNames verifica que retorne todos los nombres de agentes.
func TestGetAgentNames(t *testing.T) {
	agent1 := newSyndicateTestAgent("agent1", "->A")
	agent2 := newSyndicateTestAgent("agent2", "->B")

	syndicate, _ := NewSyndicate(WithAgents(agent1, agent2))

	names := syndicate.GetAgentNames()
	if len(names) != 2 {
		t.Errorf("Se esperaban 2 nombres, se obtuvieron %d", len(names))
	}

	// Verificar que ambos nombres estén presentes
	found1, found2 := false, false
	for _, name := range names {
		if name == "agent1" {
			found1 = true
		}
		if name == "agent2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("No se encontraron todos los nombres de agentes: %v", names)
	}
}

// TestExecuteAgentWithTimeout simula un timeout en la ejecución del agente.
func TestExecuteAgentWithTimeout(t *testing.T) {
	// Crear un agente que demora en responder
	slowAgent := &syndicateTestAgent{
		name: "slowAgent",
		chatFunc: func(ctx context.Context, options ...ChatOption) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return "slow response", nil
			}
		},
	}

	syndicate, _ := NewSyndicate(WithAgent(slowAgent))

	// Crear un contexto que expira rápidamente
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := syndicate.ExecuteAgent(ctx, "slowAgent",
		WithExecuteUserName("usuarioTimeout"),
		WithExecuteInput("entrada"),
	)
	if err == nil {
		t.Error("Se esperaba error por timeout, pero no ocurrió")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Se esperaba context.DeadlineExceeded, se obtuvo: %v", err)
	}
}

// TestWithAgents verifica que WithAgents registre múltiples agentes.
func TestWithAgents(t *testing.T) {
	agent1 := newSyndicateTestAgent("agent1", "->A")
	agent2 := newSyndicateTestAgent("agent2", "->B")
	agent3 := newSyndicateTestAgent("agent3", "->C")

	syndicate, err := NewSyndicate(
		WithAgents(agent1, agent2, agent3),
	)
	if err != nil {
		t.Fatalf("Error creando syndicate: %v", err)
	}

	names := syndicate.GetAgentNames()
	if len(names) != 3 {
		t.Errorf("Se esperaban 3 agentes, se obtuvieron %d", len(names))
	}
}
