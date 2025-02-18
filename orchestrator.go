package gomaky

import (
	"context"
	"fmt"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

// Orchestrator is responsible for managing agents, global history,
// and optionally, the execution sequence of agents.
type Orchestrator struct {
	agents        map[string]Agent
	globalHistory Memory
	sequence      []string
	mutex         sync.RWMutex
}

// NewOrchestrator creates an orchestrator with minimal configuration.
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		agents:        make(map[string]Agent),
		globalHistory: NewSimpleMemory(),
	}
}

// Process executes a specific agent by combining the global history with the agent's internal memory.
func (o *Orchestrator) Process(ctx context.Context, agentName string, input string) (string, error) {
	agent, exists := o.GetAgent(agentName)
	if !exists {
		return "", fmt.Errorf("agent not found: %s", agentName)
	}

	// Combine globalHistory and the agent's internal memory (if it's a BaseAgent)
	var messages []openai.ChatCompletionMessage
	if baseAgent, ok := agent.(*BaseAgent); ok {
		globalMessages := o.globalHistory.Get()
		agentMessages := baseAgent.memory.Get()
		messages = append(globalMessages, agentMessages...)
	}

	response, err := agent.Process(ctx, input, messages)
	if err != nil {
		return "", err
	}

	// Update the global history
	o.globalHistory.Add(openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input,
	})
	prefixedResponse := fmt.Sprintf("[%s]: %s", agentName, response)
	o.globalHistory.Add(openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: prefixedResponse,
	})

	return response, nil
}

// GetAgent returns a registered agent by its name.
func (o *Orchestrator) GetAgent(name string) (Agent, bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	agent, exists := o.agents[name]
	return agent, exists
}

// ProcessSequence defines an execution pipeline among agents.
// The output of one agent is passed as input to the next.
func (o *Orchestrator) ProcessSequence(ctx context.Context, input string) (string, error) {
	if len(o.sequence) == 0 {
		return "", fmt.Errorf("no sequence defined in orchestrator")
	}

	currentInput := input
	for _, agentName := range o.sequence {
		resp, err := o.Process(ctx, agentName, currentInput)
		if err != nil {
			return "", fmt.Errorf("error in agent %s: %w", agentName, err)
		}
		currentInput = resp
	}
	return currentInput, nil
}

///////////////////////////////////////////////////////////
// Orchestrator Builder
///////////////////////////////////////////////////////////

// OrchestratorBuilder allows for the fluent construction of an orchestrator.
type OrchestratorBuilder struct {
	agents        map[string]Agent
	sequence      []string
	globalHistory Memory
}

// NewOrchestratorBuilder initializes a builder with default values.
func NewOrchestratorBuilder() *OrchestratorBuilder {
	return &OrchestratorBuilder{
		agents:        make(map[string]Agent),
		sequence:      []string{},
		globalHistory: NewSimpleMemory(),
	}
}

// SetGlobalHistory allows you to define a custom global history.
func (b *OrchestratorBuilder) SetGlobalHistory(history Memory) *OrchestratorBuilder {
	b.globalHistory = history
	return b
}

// AddAgent registers an agent within the orchestrator.
func (b *OrchestratorBuilder) AddAgent(agent Agent) *OrchestratorBuilder {
	b.agents[agent.GetName()] = agent
	return b
}

// SetSequence defines the execution order (pipeline) of agents.
// For example: []string{"agent1", "agent2", "agent3"}
func (b *OrchestratorBuilder) SetSequence(seq []string) *OrchestratorBuilder {
	b.sequence = seq
	return b
}

// Build constructs the orchestrator with the specified configuration.
func (b *OrchestratorBuilder) Build() *Orchestrator {
	return &Orchestrator{
		agents:        b.agents,
		globalHistory: b.globalHistory,
		sequence:      b.sequence,
	}
}
