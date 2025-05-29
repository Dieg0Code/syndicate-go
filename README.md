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

// Implement the Tool interface
type OrderTool struct{}

func (t *OrderTool) GetDefinition() syndicate.ToolDefinition {
    schema, _ := syndicate.GenerateRawSchema(OrderSchema{})
    return syndicate.ToolDefinition{
        Name:        "ProcessOrder",
        Description: "Process customer orders",
        Parameters:  schema,
    }
}

func (t *OrderTool) Execute(args json.RawMessage) (interface{}, error) {
    var order OrderSchema
    json.Unmarshal(args, &order)
    // Process the order...
    return "Order processed successfully", nil
}

// Create agent with tool
agent, _ := syndicate.NewAgent(
    syndicate.WithClient(client),
    syndicate.WithName("OrderAgent"),
    syndicate.WithSystemPrompt("You process customer orders."),
    syndicate.WithTools(&OrderTool{}),
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

Tools allow agents to interact with external systems. Implement the `Tool` interface:

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

Create structured prompts with the builder:

```go
prompt := syndicate.NewPromptBuilder().
    CreateSection("Role").
    AddText("Role", "You are a customer service agent.").
    CreateSection("Guidelines").
    AddListItem("Guidelines", "Be helpful and professional.").
    AddListItem("Guidelines", "Always ask for clarification.").
    Build()
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
