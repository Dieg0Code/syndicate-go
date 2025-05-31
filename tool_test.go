package syndicate

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewTool(t *testing.T) {
	type TestSchema struct {
		Name string `json:"name" description:"The name" required:"true"`
		Age  int    `json:"age" description:"The age" required:"true"`
	}

	tests := []struct {
		name    string
		options []ToolOption
		wantErr bool
	}{
		{
			name: "valid tool",
			options: []ToolOption{
				WithToolName("TestTool"),
				WithToolDescription("A test tool"),
				WithToolSchema(TestSchema{}),
				WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
					return "executed", nil
				}),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			options: []ToolOption{
				WithToolDescription("A test tool"),
				WithToolSchema(TestSchema{}),
				WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
					return "executed", nil
				}),
			},
			wantErr: true,
		},
		{
			name: "missing description",
			options: []ToolOption{
				WithToolName("TestTool"),
				WithToolSchema(TestSchema{}),
				WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
					return "executed", nil
				}),
			},
			wantErr: true,
		},
		{
			name: "missing schema",
			options: []ToolOption{
				WithToolName("TestTool"),
				WithToolDescription("A test tool"),
				WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
					return "executed", nil
				}),
			},
			wantErr: true,
		},
		{
			name: "missing execute handler",
			options: []ToolOption{
				WithToolName("TestTool"),
				WithToolDescription("A test tool"),
				WithToolSchema(TestSchema{}),
			},
			wantErr: true,
		},
		{
			name: "invalid options",
			options: []ToolOption{
				WithToolName(""),
				WithToolDescription("A test tool"),
				WithToolSchema(TestSchema{}),
				WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
					return "executed", nil
				}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTool(tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToolExecution(t *testing.T) {
	type TestSchema struct {
		Input string `json:"input" description:"The input" required:"true"`
	}

	executed := false
	expectedInput := "test input"

	tool, err := NewTool(
		WithToolName("TestTool"),
		WithToolDescription("A test tool"),
		WithToolSchema(TestSchema{}),
		WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
			var schema TestSchema
			if err := json.Unmarshal(args, &schema); err != nil {
				return nil, err
			}

			if schema.Input != expectedInput {
				return nil, errors.New("unexpected input")
			}

			executed = true
			return "success", nil
		}),
	)

	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	args, _ := json.Marshal(map[string]string{"input": expectedInput})
	result, err := tool.Execute(args)

	if err != nil {
		t.Errorf("Tool execution failed: %v", err)
	}

	if !executed {
		t.Error("Execute handler was not called")
	}

	if result != "success" {
		t.Errorf("Unexpected result: %v", result)
	}
}
