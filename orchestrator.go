package gokamy

import (
	"context"
	"fmt"
	"sync"
)

// Orchestrator manages multiple agents, maintains a global conversation history,
// and optionally defines an execution sequence (pipeline) for agents.
type Orchestrator struct {
	agents        map[string]Agent // Registered agents identified by their names.
	globalHistory Memory           // Global conversation history shared across agents.
	sequence      []string         // Ordered sequence of agent names for pipelined processing.
	mutex         sync.RWMutex     // RWMutex to ensure thread-safe access to the orchestrator.
}

// NewOrchestrator creates and returns a new Orchestrator with default settings.
// It initializes the agents map and the global history with a simple in-memory implementation.
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		agents:        make(map[string]Agent),
		globalHistory: NewSimpleMemory(),
	}
}

// Process executes a specific agent by combining the global history with the agent's internal memory.
// It retrieves the target agent, merges global messages with the agent's own messages (if applicable),
// processes the input, and then updates the global history with both the user input and the agent's response.
func (o *Orchestrator) Process(ctx context.Context, agentName, userName, input string) (string, error) {
	// Retrieve the agent by its name.
	agent, exists := o.GetAgent(agentName)
	if !exists {
		return "", fmt.Errorf("agent not found: %s", agentName)
	}

	var messages []Message
	// If the agent is of type BaseAgent, combine the global and agent-specific messages.
	if baseAgent, ok := agent.(*BaseAgent); ok {
		globalMessages := o.globalHistory.Get()
		agentMessages := baseAgent.memory.Get()
		messages = append(globalMessages, agentMessages...)
	}

	// Process the input using the agent, passing the combined message history.
	response, err := agent.Process(ctx, userName, input, messages)
	if err != nil {
		return "", err
	}

	// Update the global history with the user's input.
	o.globalHistory.Add(Message{
		Role:    RoleUser,
		Content: input,
		Name:    userName,
	})
	// Prefix the agent's response with its name for clarity.
	prefixedResponse := fmt.Sprintf("[%s]: %s", agentName, response)
	// Update the global history with the agent's response.
	o.globalHistory.Add(Message{
		Role:    RoleAssistant,
		Content: prefixedResponse,
		Name:    agentName,
	})

	return response, nil
}

// GetAgent retrieves a registered agent by its name in a thread-safe manner.
func (o *Orchestrator) GetAgent(name string) (Agent, bool) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	agent, exists := o.agents[name]
	return agent, exists
}

// ProcessSequence executes a pipeline of agents as defined in the orchestrator's sequence.
// The output of each agent is used as the input for the next agent in the sequence.
// It returns the final output from the last agent or an error if processing fails.
func (o *Orchestrator) ProcessSequence(ctx context.Context, userName, input string) (string, error) {
	if len(o.sequence) == 0 {
		return "", fmt.Errorf("no sequence defined in orchestrator")
	}

	currentInput := input
	// Iterate over each agent in the defined sequence.
	for _, agentName := range o.sequence {
		resp, err := o.Process(ctx, agentName, userName, currentInput)
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

// OrchestratorBuilder provides a fluent interface for constructing an Orchestrator.
// It allows developers to configure agents, set a custom global history, and define an execution sequence.
type OrchestratorBuilder struct {
	agents        map[string]Agent // Agents to be registered.
	sequence      []string         // Ordered sequence of agent names for pipelined processing.
	globalHistory Memory           // Global conversation history.
}

// NewOrchestratorBuilder initializes a new OrchestratorBuilder with default values.
// It creates an empty agents map, an empty sequence, and a default global history.
func NewOrchestratorBuilder() *OrchestratorBuilder {
	return &OrchestratorBuilder{
		agents:        make(map[string]Agent),
		sequence:      []string{},
		globalHistory: NewSimpleMemory(),
	}
}

// SetGlobalHistory sets a custom global conversation history for the orchestrator.
func (b *OrchestratorBuilder) SetGlobalHistory(history Memory) *OrchestratorBuilder {
	b.globalHistory = history
	return b
}

// AddAgent registers an agent with the orchestrator using the agent's name as the key.
func (b *OrchestratorBuilder) AddAgent(agent Agent) *OrchestratorBuilder {
	b.agents[agent.GetName()] = agent
	return b
}

// SetSequence defines the execution order (pipeline) of agents.
// For example: []string{"agent1", "agent2", "agent3"}.
func (b *OrchestratorBuilder) SetSequence(seq []string) *OrchestratorBuilder {
	b.sequence = seq
	return b
}

// Build constructs and returns an Orchestrator instance based on the current configuration.
func (b *OrchestratorBuilder) Build() *Orchestrator {
	return &Orchestrator{
		agents:        b.agents,
		globalHistory: b.globalHistory,
		sequence:      b.sequence,
	}
}
