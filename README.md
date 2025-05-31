<div align="center">
  <img src="https://i.imgur.com/e608zH3.png" alt="Syndicate SDK Logo"/>
  
[![Go Report Card](https://goreportcard.com/badge/github.com/Dieg0Code/syndicate-go)](https://goreportcard.com/report/github.com/Dieg0Code/syndicate-go)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/Dieg0Code/syndicate-go/ci.yml?branch=main)](https://github.com/Dieg0Code/syndicate-go/actions)
[![codecov](https://codecov.io/github/Dieg0Code/syndicate-go/graph/badge.svg?token=FXYY1S9EP4)](https://codecov.io/github/Dieg0Code/syndicate-go)
[![GoDoc](https://godoc.org/github.com/Dieg0Code/syndicate-go?status.svg)](https://pkg.go.dev/github.com/Dieg0Code/syndicate-go)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/Dieg0Code/syndicate-go)](https://github.com/Dieg0Code/syndicate-go/releases)
</div>

# Syndicate

**A clean, simple Go SDK for building AI agent applications with LLMs, tools, and workflows.**

Eliminate the complexity of managing LLM APIs directly. Perfect for prototypes, MVPs, and applications that need straightforward AI agent integration.

## 🚀 Quick Start

```bash
go get github.com/Dieg0Code/syndicate-go
```

```go
package main

import (
    "context"
    "fmt"
    syndicate "github.com/Dieg0Code/syndicate-go"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    client := syndicate.NewOpenAIClient("YOUR_API_KEY")

    agent, _ := syndicate.NewAgent(
        syndicate.WithClient(client),
        syndicate.WithName("Assistant"),
        syndicate.WithSystemPrompt("You are a helpful AI assistant."),
        syndicate.WithModel(openai.GPT4),
        syndicate.WithMemory(syndicate.NewSimpleMemory()),
    )

    response, _ := agent.Chat(context.Background(),
        syndicate.WithUserName("User"),
        syndicate.WithInput("Hello! What can you help me with?"),
    )

    fmt.Println(response)
}
```

## ✨ Key Features

- **🤖 Agent Orchestration**: Create agents that work independently or in simple sequential pipelines
- **🛠️ Tool Integration**: Connect agents to external APIs with automatic JSON schema generation
- **💾 Flexible Memory**: From simple in-memory to custom database backends
- **🔄 Sequential Workflows**: Chain agents for multi-step processing
- **📝 Structured Prompts**: Build consistent, maintainable agent instructions
- **⚡ Clean API**: Functional options pattern for readable, maintainable code

## 🎯 Ideal For

- **✅ Prototypes & MVPs** - Get AI features running quickly
- **✅ Small to medium applications** - Clean integration without overhead
- **✅ Learning AI development** - Simple, well-documented patterns
- **✅ Custom tool integration** - Easy to extend with your APIs
- **✅ Sequential workflows** - Chain agents for multi-step tasks

**Not ideal for:** Complex branching workflows, high-scale production systems requiring advanced observability, or enterprise-grade orchestration needs.

## 📋 Examples

### Single Agent with Tools

```go
// Define your tool schema
type OrderSchema struct {
    Items   []string `json:"items" description:"Items to order" required:"true"`
    Address string   `json:"address" description:"Delivery address" required:"true"`
}

// Create a tool with functional options
tool, _ := syndicate.NewTool(
    syndicate.WithToolName("ProcessOrder"),
    syndicate.WithToolDescription("Process customer orders"),
    syndicate.WithToolSchema(OrderSchema{}),
    syndicate.WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
        var order OrderSchema
        if err := json.Unmarshal(args, &order); err != nil {
            return nil, err
        }
        // Process the order...
        return "Order processed successfully", nil
    }),
)

// Create agent with tool
agent, _ := syndicate.NewAgent(
    syndicate.WithClient(client),
    syndicate.WithName("OrderAgent"),
    syndicate.WithSystemPrompt("You process customer orders."),
    syndicate.WithTools(tool),
    syndicate.WithMemory(syndicate.NewSimpleMemory()),
)
```

### Sequential Multi-Agent Pipeline

```go
// Create specialized agents
orderAgent, _ := syndicate.NewAgent(
    syndicate.WithClient(client),
    syndicate.WithName("OrderProcessor"),
    syndicate.WithSystemPrompt("You validate and process orders."),
    syndicate.WithMemory(syndicate.NewSimpleMemory()),
)

summaryAgent, _ := syndicate.NewAgent(
    syndicate.WithClient(client),
    syndicate.WithName("OrderSummarizer"),
    syndicate.WithSystemPrompt("You create order summaries."),
    syndicate.WithMemory(syndicate.NewSimpleMemory()),
)

// Create sequential pipeline
pipeline, _ := syndicate.NewSyndicate(
    syndicate.WithAgents(orderAgent, summaryAgent),
    syndicate.WithPipeline("OrderProcessor", "OrderSummarizer"),
)

// Execute pipeline
result, _ := pipeline.ExecutePipeline(context.Background(),
    syndicate.WithPipelineUserName("Customer"),
    syndicate.WithPipelineInput("I want 2 pizzas delivered to 123 Main St"),
)
```

### Custom Memory Backend

```go
// Create database-backed memory
func NewDatabaseMemory(db *sql.DB, agentID string) (syndicate.Memory, error) {
    return syndicate.NewMemory(
        syndicate.WithAddHandler(func(msg syndicate.Message) {
            data, _ := json.Marshal(msg)
            db.Exec("INSERT INTO messages (agent_id, data) VALUES (?, ?)", agentID, data)
        }),
        syndicate.WithGetHandler(func() []syndicate.Message {
            rows, _ := db.Query("SELECT data FROM messages WHERE agent_id = ?", agentID)
            var messages []syndicate.Message
            // Parse rows into messages...
            return messages
        }),
    )
}

// Use custom memory
dbMemory, _ := NewDatabaseMemory(db, "agent-123")
agent, _ := syndicate.NewAgent(
    syndicate.WithClient(client),
    syndicate.WithName("PersistentAgent"),
    syndicate.WithMemory(dbMemory),
    // ... other options
)
```

## 🏗️ Architecture

**Agent**: Individual AI entity with specific capabilities and memory  
**Tool**: External function/API that agents can call  
**Memory**: Conversation storage (in-memory, database, Redis, etc.)  
**Syndicate**: Orchestrator that manages sequential multi-agent workflows  
**Pipeline**: Sequential execution of multiple agents

## 📚 Advanced Usage

<details>
<summary><b>Tool Integration</b></summary>

Tools allow agents to interact with external systems. You can create tools easily using the functional options pattern:

```go
tool, err := syndicate.NewTool(
    syndicate.WithToolName("ToolName"),
    syndicate.WithToolDescription("Tool description"),
    syndicate.WithToolSchema(YourSchema{}),
    syndicate.WithToolExecuteHandler(func(args json.RawMessage) (interface{}, error) {
        // Your implementation here
        return result, nil
    }),
)
```

Alternatively, you can implement the `Tool` interface directly:

```go
type Tool interface {
    GetDefinition() ToolDefinition
    Execute(args json.RawMessage) (interface{}, error)
}
```

The SDK automatically generates JSON schemas from Go structs using reflection and struct tags.

</details>

<details>
<summary><b>Memory Management</b></summary>

All memory implementations satisfy this interface:

```go
type Memory interface {
    Add(message Message)
    Get() []Message
}
```

- Use `syndicate.NewSimpleMemory()` for development
- Use `syndicate.NewMemory()` with handlers for custom backends

</details>

<details>
<summary><b>Prompt Building</b></summary>

Create structured prompts with the builder, now with comprehensive markdown support:

```go
prompt := syndicate.NewPromptBuilder().
    // Basic sections and text
    CreateSection("Role").
    AddText("Role", "You are a customer service agent.").

    // Formatting options
    CreateSection("Instructions").
    AddHeader("Instructions", "Important Guidelines", 2).
    AddBoldText("Instructions", "Follow these rules carefully:").
    AddBulletItem("Instructions", "Be helpful and professional").
    AddBulletItem("Instructions", "Use clear, concise language").
    AddListItem("Instructions", "Verify customer information first").
    AddListItem("Instructions", "Solve the customer's problem").
    AddBlockquote("Instructions", "Customer satisfaction is our priority").

    // Code examples
    CreateSection("Examples").
    AddText("Examples", "Here's how to greet a customer:").
    AddCodeBlock("Examples", `function greet(name) {
    return "Hello " + name + ", how can I help you today?";
}`, "javascript").

    // Tables and links
    CreateSection("Resources").
    AddLink("Resources", "Customer Knowledge Base", "https://example.com/kb").
    AddHorizontalRule("Resources").
    AddTable("Resources",
        []string{"Resource Type", "URL", "Description"},
        [][]string{
            {"FAQ", "https://example.com/faq", "Frequently asked questions"},
            {"Policy", "https://example.com/policy", "Company policies"},
        }).

    Build()
```

The PromptBuilder combines XML-style hierarchical structure with markdown formatting for optimal LLM prompting.

Basic table example:

```go
// Basic table example
pb := syndicate.NewPromptBuilder().
    CreateSection("Tables").
    AddText("Tables", "Here's a simple table:").
    AddTable("Tables",
        []string{"Name", "Age", "Role"},  // Headers
        [][]string{                        // Rows
            {"John", "30", "Developer"},
            {"Jane", "28", "Designer"},
            {"Bob", "35", "Manager"},
        })
```

This produces a markdown table like:

```
<Tables>
Here's a simple table:
| Name | Age | Role |
| --- | --- | --- |
| John | 30 | Developer |
| Jane | 28 | Designer |
| Bob | 35 | Manager |
</Tables>
```

</details>

## 🔧 Configuration

**Supported LLM Providers**: OpenAI, DeepSeek  
**Go Version**: 1.24+  
**Architecture**: Sequential pipelines, simple agent orchestration  
**Dependencies**: Minimal external dependencies

## 📖 Documentation

- [API Reference](https://pkg.go.dev/github.com/Dieg0Code/syndicate-go)
- [Examples](https://github.com/Dieg0Code/syndicate-go/tree/main/examples)
- [Contributing Guide](CONTRIBUTING.md)

## 📦 Dependencies

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) - Apache License 2.0
- [cohesion-org/deepseek-go](https://github.com/cohesion-org/deepseek-go) - MIT License

## 🤝 Contributing

Contributions welcome! Please read our [contributing guidelines](CONTRIBUTING.md) and submit issues or pull requests.

## 📜 License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.
