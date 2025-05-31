package syndicate

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ToolConfig holds the configuration for creating custom tool implementations
type ToolConfig struct {
	Name        string
	Description string
	Schema      any
	ExecuteFunc func(args json.RawMessage) (interface{}, error)
}

// ToolOption defines a function that configures a Tool implementation
type ToolOption func(*ToolConfig) error

// WithToolName sets the name for the tool
func WithToolName(name string) ToolOption {
	return func(config *ToolConfig) error {
		if name == "" {
			return errors.New("tool name cannot be empty")
		}
		config.Name = name
		return nil
	}
}

// WithToolDescription sets the description for the tool
func WithToolDescription(description string) ToolOption {
	return func(config *ToolConfig) error {
		if description == "" {
			return errors.New("tool description cannot be empty")
		}
		config.Description = description
		return nil
	}
}

// WithToolSchema sets the schema for the tool
func WithToolSchema(schema any) ToolOption {
	return func(config *ToolConfig) error {
		if schema == nil {
			return errors.New("tool schema cannot be nil")
		}
		config.Schema = schema
		return nil
	}
}

// WithToolExecuteHandler sets the execute function for the tool
func WithToolExecuteHandler(executeFunc func(args json.RawMessage) (interface{}, error)) ToolOption {
	return func(config *ToolConfig) error {
		if executeFunc == nil {
			return errors.New("execute function cannot be nil")
		}
		config.ExecuteFunc = executeFunc
		return nil
	}
}

// customTool implements Tool using provided functions
type customTool struct {
	name        string
	description string
	schema      json.RawMessage
	executeFunc func(args json.RawMessage) (interface{}, error)
}

func (t *customTool) GetDefinition() ToolDefinition {
	return ToolDefinition{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.schema,
	}
}

func (t *customTool) Execute(args json.RawMessage) (interface{}, error) {
	return t.executeFunc(args)
}

// NewTool creates a custom Tool implementation using functional options.
// Returns an error if required options are not provided.
//
// Example:
//
//	tool, err := syndicate.NewTool(
//		syndicate.WithToolName("ProcessOrder"),
//		syndicate.WithToolDescription("Process customer orders"),
//		syndicate.WithToolSchema(OrderSchema{}),
//		syndicate.WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
//			var order OrderSchema
//			if err := json.Unmarshal(args, &order); err != nil {
//				return nil, err
//			}
//			// Process the order...
//			return "Order processed successfully", nil
//		}),
//	)
func NewTool(options ...ToolOption) (Tool, error) {
	config := &ToolConfig{}

	for _, option := range options {
		if err := option(config); err != nil {
			return nil, fmt.Errorf("failed to apply tool option: %w", err)
		}
	}

	// Validate that all required fields are provided
	if config.Name == "" {
		return nil, errors.New("tool name is required")
	}
	if config.Description == "" {
		return nil, errors.New("tool description is required")
	}
	if config.Schema == nil {
		return nil, errors.New("tool schema is required")
	}
	if config.ExecuteFunc == nil {
		return nil, errors.New("tool execute function is required")
	}

	schema, err := GenerateRawSchema(config.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema: %w", err)
	}

	return &customTool{
		name:        config.Name,
		description: config.Description,
		schema:      schema,
		executeFunc: config.ExecuteFunc,
	}, nil
}
