package syndicate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// fakeAgent es una implementación simple del interface Agent para pruebas.
type fakeAgent struct {
	name   string
	memory Memory
	// processFunc simula el procesamiento del agente.
	// Recibe el input y retorna una respuesta personalizada.
	processFunc func(ctx context.Context, userName, input string, additionalMessages ...[]Message) (string, error)
}

func (f *fakeAgent) Process(ctx context.Context, userName, input string, additionalMessages ...[]Message) (string, error) {
	if f.processFunc != nil {
		return f.processFunc(ctx, userName, input, additionalMessages...)
	}
	// Por defecto retorna el input con un sufijo identificador.
	return fmt.Sprintf("response from %s", f.name), nil
}

func (f *fakeAgent) AddTool(tool Tool) {
	// No es necesario para estas pruebas.
}

func (f *fakeAgent) SetConfigPrompt(prompt string) {
	// No es necesario para estas pruebas.
}

func (f *fakeAgent) GetName() string {
	return f.name
}

// newFakeAgentBuilder crea un fakeAgent con memoria simple (fakeMemory).
func newFakeAgentBuilder(name string, suffix string) *fakeAgent {
	mem := &fakeMemory{}
	return &fakeAgent{
		name:   name,
		memory: mem,
		processFunc: func(ctx context.Context, userName, input string, additionalMessages ...[]Message) (string, error) {
			// Simula la utilización de los mensajes adicionales concatenándolos
			var combined string
			for _, msgs := range additionalMessages {
				for _, m := range msgs {
					combined += m.Content + " "
				}
			}
			// Retorna el input junto con un sufijo para identificar al agente.
			return fmt.Sprintf("%s %s%s", input, suffix, combined), nil
		},
	}
}

// TestExecuteAgent verifica el flujo de ExecuteAgent y el manejo de la historia global.
func TestExecuteAgent(t *testing.T) {
	// Crear un fake agente tipo BaseAgent (para que se pueda combinar history)
	agent := newFakeAgentBuilder("agent1", "->agent1")
	// Agregamos un mensaje interno en el agente para simular historia previa.
	agent.memory.Add(Message{
		Role:    RoleUser,
		Content: "internal agent message",
		Name:    agent.name,
	})
	// Crear una historia global simple.
	globalHistory := &fakeMemory{}
	// Registrar el agente en el syndicate.
	s := &Syndicate{
		agents:        map[string]Agent{"agent1": agent},
		globalHistory: globalHistory,
	}
	// Ejecutar el agente.
	input := "hola"
	ctx := context.Background()
	resp, err := s.ExecuteAgent(ctx, "agent1", "usuario1", input)
	if err != nil {
		t.Fatalf("ExecuteAgent failed: %v", err)
	}
	// Verificar que la respuesta no incluya los prefijos de globalHistory
	if !strings.Contains(resp, "agent1") && !strings.Contains(resp, "->agent1") {
		t.Errorf("respuesta inesperada: %s", resp)
	}
	// Verificar que globalHistory tenga 2 nuevos mensajes: input del usuario y respuesta del agente.
	history := globalHistory.Get()
	if len(history) != 2 {
		t.Errorf("se esperaban 2 mensajes en globalHistory, se obtuvieron: %d", len(history))
	}
	// Verificar que el mensaje de respuesta tenga el prefijo "[agent1]: "
	if !strings.HasPrefix(history[1].Content, "[agent1]:") {
		t.Errorf("el mensaje de respuesta global no tiene el prefijo correcto: %s", history[1].Content)
	}
}

// TestFindAgent verifica FindAgent en un Syndicate.
func TestFindAgent(t *testing.T) {
	agent := newFakeAgentBuilder("agentX", "")
	s := &Syndicate{
		agents: map[string]Agent{"agentX": agent},
	}
	foundAgent, exists := s.FindAgent("agentX")
	if !exists {
		t.Error("se esperaba encontrar el agente agentX")
	}
	if foundAgent.GetName() != "agentX" {
		t.Errorf("nombre del agente incorrecto: se obtuvo %s", foundAgent.GetName())
	}
	// Verificar búsqueda de agente inexistente.
	_, exists = s.FindAgent("no-exist")
	if exists {
		t.Error("no se esperaba encontrar un agente inexistente")
	}
}

// TestExecutePipeline verifica la ejecución secuencial de agentes mediante la pipeline.
func TestExecutePipeline(t *testing.T) {
	// Crear dos fake agentes que transforman el input.
	agent1 := newFakeAgentBuilder("agent1", "->A")
	agent2 := newFakeAgentBuilder("agent2", "->B")
	// Definir la pipeline.
	pipeline := []string{"agent1", "agent2"}
	// Registrar los agentes.
	s := &Syndicate{
		agents: map[string]Agent{
			"agent1": agent1,
			"agent2": agent2,
		},
		// No se requiere globalHistory en este test; se puede dejar vacío.
		globalHistory: &fakeMemory{},
		pipeline:      pipeline,
	}
	ctx := context.Background()
	// El input del pipeline.
	input := "inicio"
	finalResp, err := s.ExecutePipeline(ctx, "usuario1", input)
	if err != nil {
		t.Fatalf("ExecutePipeline failed: %v", err)
	}
	// Se espera que cada agente agregue su sufijo al input.
	if !strings.Contains(finalResp, "->A") || !strings.Contains(finalResp, "->B") {
		t.Errorf("pipeline no procesada correctamente, salida: %s", finalResp)
	}
}

// TestExecutePipelineSinDefinir verifica que se retorne error al ejecutar pipeline vacío.
func TestExecutePipelineSinDefinir(t *testing.T) {
	s := &Syndicate{
		agents:        map[string]Agent{},
		globalHistory: &fakeMemory{},
		pipeline:      []string{},
	}
	ctx := context.Background()
	_, err := s.ExecutePipeline(ctx, "usuario1", "inicio")
	if err == nil || !strings.Contains(err.Error(), "no pipeline defined") {
		t.Errorf("se esperaba error por pipeline vacío, se obtuvo: %v", err)
	}
}

// TestSyndicateBuilder verifica el constructor fluido del Syndicate.
func TestSyndicateBuilder(t *testing.T) {
	// Crear un fake agente para registrar.
	agent := newFakeAgentBuilder("builderAgent", "->builder")
	// Crear una historia global customizada.
	customHistory := &fakeMemory{}
	// Utilizar el builder.
	builder := NewSyndicate().
		SetGlobalHistory(customHistory).
		RecruitAgent(agent).
		DefinePipeline([]string{"builderAgent"})
	synd := builder.Build()
	// Verificar que el agente se encuentre registrado.
	found, exists := synd.FindAgent("builderAgent")
	if !exists {
		t.Error("se esperaba encontrar el agente builderAgent en el syndicate")
	}
	if found.GetName() != "builderAgent" {
		t.Errorf("nombre del agente incorrecto: se obtuvo %s", found.GetName())
	}
	// Verificar que la pipeline tenga lo definido.
	if len(synd.pipeline) != 1 || synd.pipeline[0] != "builderAgent" {
		t.Errorf("pipeline incorrecta, se obtuvo: %v", synd.pipeline)
	}
	// Verificar que la globalHistory sea la customizada.
	if synd.globalHistory != customHistory {
		t.Error("globalHistory no fue asignada correctamente en el builder")
	}
}

// TestExecuteAgentConContextoTimeout simula un timeout en la ejecución del agente.
func TestExecuteAgentConContextoTimeout(t *testing.T) {
	// Creamos un agente que demora en responder.
	agent := newFakeAgentBuilder("slowAgent", "->slow")
	agent.processFunc = func(ctx context.Context, userName, input string, additionalMessages ...[]Message) (string, error) {
		// Demoramos la respuesta.
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return fmt.Sprintf("slow response from %s", "slowAgent"), nil
		}
	}
	s := &Syndicate{
		agents:        map[string]Agent{"slowAgent": agent},
		globalHistory: &fakeMemory{},
	}
	// Creamos un contexto que expira rápidamente.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := s.ExecuteAgent(ctx, "slowAgent", "usuarioTimeout", "entrada")
	if err == nil {
		t.Error("se esperaba error por timeout, pero no ocurrió")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("se esperaba context.DeadlineExceeded, se obtuvo: %v", err)
	}
}
