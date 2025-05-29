package syndicate

import (
	"context"
	"fmt"
	"sync"
)

// Syndicate defines the interface for managing multiple agents and pipelines.
type Syndicate interface {
	ExecuteAgent(ctx context.Context, agentName string, options ...ExecuteAgentOption) (string, error)
	ExecutePipeline(ctx context.Context, options ...PipelineOption) (string, error)
	FindAgent(name string) (Agent, bool)
	GetGlobalHistory() []Message
	GetAgentNames() []string
	GetPipeline() []string
}

// syndicate is the private implementation of the Syndicate interface.
type syndicate struct {
	agents        map[string]Agent // Registered agents identified by their names.
	globalHistory Memory           // Global conversation history shared across agents.
	pipeline      []string         // Ordered pipeline of agent names for sequential processing.
	mutex         sync.RWMutex     // RWMutex to ensure thread-safe access to the syndicate.
}

// SyndicateOption defines a function that configures a syndicate.
type SyndicateOption func(*syndicate) error

// WithAgent registers an agent with the syndicate using the agent's name as the key.
func WithAgent(agent Agent) SyndicateOption {
	return func(s *syndicate) error {
		if agent == nil {
			return fmt.Errorf("agent cannot be nil")
		}
		name := agent.GetName()
		if name == "" {
			return fmt.Errorf("agent name cannot be empty")
		}
		s.agents[name] = agent
		return nil
	}
}

// WithAgents registers multiple agents with the syndicate.
func WithAgents(agents ...Agent) SyndicateOption {
	return func(s *syndicate) error {
		for _, agent := range agents {
			if agent == nil {
				return fmt.Errorf("agent cannot be nil")
			}
			name := agent.GetName()
			if name == "" {
				return fmt.Errorf("agent name cannot be empty")
			}
			s.agents[name] = agent
		}
		return nil
	}
}

// WithPipeline defines the execution order (pipeline) of agents.
// For example: []string{"agent1", "agent2", "agent3"}.
func WithPipeline(pipeline ...string) SyndicateOption {
	return func(s *syndicate) error {
		if len(pipeline) == 0 {
			return fmt.Errorf("pipeline cannot be empty")
		}
		// Validate that all agents in pipeline exist
		for _, agentName := range pipeline {
			if _, exists := s.agents[agentName]; !exists {
				return fmt.Errorf("agent %s not found in syndicate", agentName)
			}
		}
		s.pipeline = pipeline
		return nil
	}
}

// WithGlobalHistory sets a custom global conversation history for the syndicate.
func WithGlobalHistory(history Memory) SyndicateOption {
	return func(s *syndicate) error {
		if history == nil {
			return fmt.Errorf("global history cannot be nil")
		}
		s.globalHistory = history
		return nil
	}
}

// NewSyndicate creates a new Syndicate with the provided options.
func NewSyndicate(options ...SyndicateOption) (Syndicate, error) {
	s := &syndicate{
		agents:        make(map[string]Agent),
		globalHistory: NewSimpleMemory(),
		pipeline:      []string{},
	}

	// Apply all options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, fmt.Errorf("failed to apply syndicate option: %w", err)
		}
	}

	return s, nil
}

// ExecuteAgentOption defines options for executing a single agent.
type ExecuteAgentOption func(*executeAgentRequest)

// executeAgentRequest holds parameters for executing an agent.
type executeAgentRequest struct {
	userName           string
	input              string
	imageURLs          []string
	additionalMessages [][]Message
	useGlobalHistory   bool
}

// WithExecuteUserName sets the user name for agent execution.
func WithExecuteUserName(userName string) ExecuteAgentOption {
	return func(r *executeAgentRequest) {
		r.userName = userName
	}
}

// WithExecuteInput sets the input for agent execution.
func WithExecuteInput(input string) ExecuteAgentOption {
	return func(r *executeAgentRequest) {
		r.input = input
	}
}

// WithExecuteImages sets image URLs for agent execution.
func WithExecuteImages(imageURLs ...string) ExecuteAgentOption {
	return func(r *executeAgentRequest) {
		r.imageURLs = imageURLs
	}
}

// WithExecuteAdditionalMessages adds additional messages for agent execution.
func WithExecuteAdditionalMessages(messages ...[]Message) ExecuteAgentOption {
	return func(r *executeAgentRequest) {
		r.additionalMessages = messages
	}
}

// WithGlobalHistoryContext includes global history in the agent execution.
func WithGlobalHistoryContext() ExecuteAgentOption {
	return func(r *executeAgentRequest) {
		r.useGlobalHistory = true
	}
}

// ExecuteAgent runs a specific agent with the provided options.
func (s *syndicate) ExecuteAgent(ctx context.Context, agentName string, options ...ExecuteAgentOption) (string, error) {
	// Apply default values
	req := &executeAgentRequest{
		useGlobalHistory: true, // Default to using global history
	}

	// Apply all options
	for _, opt := range options {
		opt(req)
	}

	// Validate required fields
	if req.userName == "" {
		return "", fmt.Errorf("user name is required")
	}
	if req.input == "" {
		return "", fmt.Errorf("input is required")
	}

	// Retrieve the agent by its name
	agent, exists := s.FindAgent(agentName)
	if !exists {
		return "", fmt.Errorf("agent not found: %s", agentName)
	}

	// Prepare chat options for the agent
	chatOptions := []ChatOption{
		WithUserName(req.userName),
		WithInput(req.input),
	}

	// Add images if provided
	if len(req.imageURLs) > 0 {
		chatOptions = append(chatOptions, WithImages(req.imageURLs...))
	}

	// Add global history as additional messages if requested
	if req.useGlobalHistory {
		s.mutex.RLock()
		globalMessages := s.globalHistory.Get()
		s.mutex.RUnlock()

		if len(globalMessages) > 0 {
			chatOptions = append(chatOptions, WithAdditionalMessages(globalMessages))
		}
	}

	// Add any additional messages
	for _, msgs := range req.additionalMessages {
		chatOptions = append(chatOptions, WithAdditionalMessages(msgs))
	}

	// Execute the agent
	response, err := agent.Chat(ctx, chatOptions...)
	if err != nil {
		return "", fmt.Errorf("error executing agent %s: %w", agentName, err)
	}

	// Update global history
	s.mutex.Lock()
	s.globalHistory.Add(Message{
		Role:    RoleUser,
		Content: req.input,
		Name:    req.userName,
	})

	// Prefix the agent's response with its name for clarity in global history
	prefixedResponse := fmt.Sprintf("[%s]: %s", agentName, response)
	s.globalHistory.Add(Message{
		Role:    RoleAssistant,
		Content: prefixedResponse,
		Name:    agentName,
	})
	s.mutex.Unlock()

	return response, nil
}

// PipelineOption defines options for pipeline execution.
type PipelineOption func(*pipelineRequest)

// pipelineRequest holds parameters for pipeline execution.
type pipelineRequest struct {
	userName  string
	input     string
	imageURLs []string
}

// WithPipelineUserName sets the user name for pipeline execution.
func WithPipelineUserName(userName string) PipelineOption {
	return func(r *pipelineRequest) {
		r.userName = userName
	}
}

// WithPipelineInput sets the input for pipeline execution.
func WithPipelineInput(input string) PipelineOption {
	return func(r *pipelineRequest) {
		r.input = input
	}
}

// WithPipelineImages sets image URLs for pipeline execution (only for first agent).
func WithPipelineImages(imageURLs ...string) PipelineOption {
	return func(r *pipelineRequest) {
		r.imageURLs = imageURLs
	}
}

// ExecutePipeline runs a sequence of agents as defined in the syndicate's pipeline.
func (s *syndicate) ExecutePipeline(ctx context.Context, options ...PipelineOption) (string, error) {
	if len(s.pipeline) == 0 {
		return "", fmt.Errorf("no pipeline defined in syndicate")
	}

	// Apply default values
	req := &pipelineRequest{}

	// Apply all options
	for _, opt := range options {
		opt(req)
	}

	// Validate required fields
	if req.userName == "" {
		return "", fmt.Errorf("user name is required")
	}
	if req.input == "" {
		return "", fmt.Errorf("input is required")
	}

	currentInput := req.input
	var currentImages []string = req.imageURLs // Images only for first agent

	// Iterate over each agent in the defined pipeline
	for i, agentName := range s.pipeline {
		executeOptions := []ExecuteAgentOption{
			WithExecuteUserName(req.userName),
			WithExecuteInput(currentInput),
			WithGlobalHistoryContext(),
		}

		// Only add images to the first agent in the pipeline
		if i == 0 && len(currentImages) > 0 {
			executeOptions = append(executeOptions, WithExecuteImages(currentImages...))
		}

		resp, err := s.ExecuteAgent(ctx, agentName, executeOptions...)
		if err != nil {
			return "", fmt.Errorf("error in agent %s: %w", agentName, err)
		}
		currentInput = resp
		currentImages = nil // Clear images after first agent
	}

	return currentInput, nil
}

// FindAgent retrieves a registered agent by its name in a thread-safe manner.
func (s *syndicate) FindAgent(name string) (Agent, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	agent, exists := s.agents[name]
	return agent, exists
}

// GetGlobalHistory returns a copy of the global conversation history.
func (s *syndicate) GetGlobalHistory() []Message {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.globalHistory.Get()
}

// GetAgentNames returns a list of all registered agent names.
func (s *syndicate) GetAgentNames() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	names := make([]string, 0, len(s.agents))
	for name := range s.agents {
		names = append(names, name)
	}
	return names
}

// GetPipeline returns a copy of the current pipeline.
func (s *syndicate) GetPipeline() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	pipeline := make([]string, len(s.pipeline))
	copy(pipeline, s.pipeline)
	return pipeline
}
