package syndicate

import (
	"context"
	"fmt"
	"sync"
)

// Syndicate manages multiple agents, maintains a global conversation history,
// and optionally defines an execution pipeline for agents.
type Syndicate struct {
	agents        map[string]Agent // Registered agents identified by their names.
	globalHistory Memory           // Global conversation history shared across agents.
	pipeline      []string         // Ordered pipeline of agent names for sequential processing.
	mutex         sync.RWMutex     // RWMutex to ensure thread-safe access to the syndicate.
}

// NewSyndicate creates and returns a new Syndicate with default settings.
// It initializes the agents map and the global history with a simple in-memory implementation.
// func NewSyndicate() *Syndicate {
// 	return &Syndicate{
// 		agents:        make(map[string]Agent),
// 		globalHistory: NewSimpleMemory(),
// 	}
// }

// ExecuteAgent runs a specific agent by combining the global history with the agent's internal memory.
// It retrieves the target agent, merges global messages with the agent's own messages (if applicable),
// processes the input, and then updates the global history with both the user input and the agent's response.
func (s *Syndicate) ExecuteAgent(ctx context.Context, agentName, userName, input string) (string, error) {
	// Retrieve the agent by its name.
	agent, exists := s.FindAgent(agentName)
	if !exists {
		return "", fmt.Errorf("agent not found: %s", agentName)
	}

	var messages []Message
	// If the agent is of type BaseAgent, combine the global and agent-specific messages.
	if baseAgent, ok := agent.(*BaseAgent); ok {
		globalMessages := s.globalHistory.Get()
		agentMessages := baseAgent.memory.Get()
		messages = append(globalMessages, agentMessages...)
	}

	// Process the input using the agent, passing the combined message history.
	response, err := agent.Process(ctx, userName, input, messages)
	if err != nil {
		return "", err
	}

	// Update the global history with the user's input.
	s.globalHistory.Add(Message{
		Role:    RoleUser,
		Content: input,
		Name:    userName,
	})
	// Prefix the agent's response with its name for clarity.
	prefixedResponse := fmt.Sprintf("[%s]: %s", agentName, response)
	// Update the global history with the agent's response.
	s.globalHistory.Add(Message{
		Role:    RoleAssistant,
		Content: prefixedResponse,
		Name:    agentName,
	})

	return response, nil
}

// FindAgent retrieves a registered agent by its name in a thread-safe manner.
func (s *Syndicate) FindAgent(name string) (Agent, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	agent, exists := s.agents[name]
	return agent, exists
}

// ExecutePipeline runs a sequence of agents as defined in the syndicate's pipeline.
// The output of each agent is used as the input for the next agent in the pipeline.
// It returns the final output from the last agent or an error if processing fails.
func (s *Syndicate) ExecutePipeline(ctx context.Context, userName, input string) (string, error) {
	if len(s.pipeline) == 0 {
		return "", fmt.Errorf("no pipeline defined in syndicate")
	}

	currentInput := input
	// Iterate over each agent in the defined pipeline.
	for _, agentName := range s.pipeline {
		resp, err := s.ExecuteAgent(ctx, agentName, userName, currentInput)
		if err != nil {
			return "", fmt.Errorf("error in agent %s: %w", agentName, err)
		}
		currentInput = resp
	}
	return currentInput, nil
}

///////////////////////////////////////////////////////////
// Syndicate Builder
///////////////////////////////////////////////////////////

// SyndicateBuilder provides a fluent interface for constructing a Syndicate.
// It allows developers to configure agents, set a custom global history, and define an execution pipeline.
type SyndicateBuilder struct {
	agents        map[string]Agent // Agents to be registered.
	pipeline      []string         // Ordered pipeline of agent names for sequential processing.
	globalHistory Memory           // Global conversation history.
}

// NewSyndicateBuilder initializes a new SyndicateBuilder with default values and a simple in-memory history.
// It creates an empty agents map, an empty pipeline, and a default global history.
func NewSyndicate() *SyndicateBuilder {
	return &SyndicateBuilder{
		agents:        make(map[string]Agent),
		pipeline:      []string{},
		globalHistory: NewSimpleMemory(),
	}
}

// SetGlobalHistory sets a custom global conversation history for the syndicate.
func (b *SyndicateBuilder) SetGlobalHistory(history Memory) *SyndicateBuilder {
	b.globalHistory = history
	return b
}

// RecruitAgent registers an agent with the syndicate using the agent's name as the key.
func (b *SyndicateBuilder) RecruitAgent(agent Agent) *SyndicateBuilder {
	b.agents[agent.GetName()] = agent
	return b
}

// DefinePipeline defines the execution order (pipeline) of agents.
// For example: []string{"agent1", "agent2", "agent3"}.
func (b *SyndicateBuilder) DefinePipeline(pipeline []string) *SyndicateBuilder {
	b.pipeline = pipeline
	return b
}

// Build constructs and returns a Syndicate instance based on the current configuration.
func (b *SyndicateBuilder) Build() *Syndicate {
	return &Syndicate{
		agents:        b.agents,
		globalHistory: b.globalHistory,
		pipeline:      b.pipeline,
	}
}
