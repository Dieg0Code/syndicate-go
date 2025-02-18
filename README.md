# Gokamy AI

<div align="center">
  <img src="https://i.imgur.com/fKAZo4d.png" alt="gokamy">
</div>

Built on top of the [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) library, Gokamy SDK is a lightweight, flexible, and extensible toolkit for building intelligent conversational agents in Golang. It enables you to create agents, engineer prompts, generate tool schemas, manage memory, and orchestrate complex workflows—making it easy to integrate advanced AI capabilities into your projects.

## Features

- **Agent Management:** Easily build and configure agents with custom system prompts, tools, and memory.
- **Prompt Engineering:** Create structured prompts with nested sections for improved clarity.
- **Tool Schemas:** Generate JSON schemas from Go structures to define tools and validate user inputs.
- **Memory Implementations:** Use built-in SimpleMemory or implement your own memory storage that adheres to the Memory interface.
- **Orchestrator:** Manage multiple agents and execute them in a predefined sequence to achieve complex workflows.
- **Extendable:** The SDK is designed to be unopinionated, allowing seamless integration into your projects.
- **CLI (Planned):** Future CLI support to initialize projects and scaffold agents with simple commands.

## Installation

To install Gokamy SDK, use Go modules:

```bash
go get github.com/Dieg0code/gokamy-ai
```

Ensure that you have Go installed and configured in your development environment.

<details open>
  <summary><strong>Quick Start</strong></summary>

Below is a simple example demonstrating how to create a haiku agent using the SDK:

```go
package main

import (
	"context"
	"fmt"
	"log"

	gokamy "github.com/Dieg0code/gokamy-ai"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client with your API key.
	client := openai.NewClient("YOUR_OPENAI_API_KEY")

	// Create a simple memory instance.
	memory := gokamy.NewSimpleMemory()

	// Build a structured prompt using PromptBuilder to instruct the agent to respond in haiku.
	systemPrompt := gokamy.NewPromptBuilder().
		CreateSection("Introduction").
		AddText("Introduction", "You are an agent that always responds in haiku format.").
		CreateSection("Guidelines").
		AddListItem("Guidelines", "Keep responses in a three-line haiku format (5-7-5 syllables).").
		AddListItem("Guidelines", "Be creative and concise.").
		Build()

	fmt.Println("System Prompt:")
	fmt.Println(systemPrompt)

	// Build the agent using AgentBuilder.
	agent, err := gokamy.NewAgentBuilder().
		SetClient(client).
		SetName("HaikuAgent").
		SetSystemPrompt(systemPrompt).
		SetMemory(memory).
		SetModel(openai.GPT4).
		SetMaxRecursion(2).
		Build()
	if err != nil {
		log.Fatalf("Error building agent: %v", err)
	}

	// Process a sample input with the agent.
	response, err := agent.Process(context.Background(), "What is the weather like today?")
	if err != nil {
		log.Fatalf("Error processing request: %v", err)
	}

	fmt.Println("\nAgent Response:")
	fmt.Println(response)
}
```
</details>

<details>
  <summary><strong>Orchestration Example</strong></summary>

Gokamy SDK also provides an orchestrator to manage a sequence of agents. In the following example, two agents are created and executed in sequence:

```go
package main

import (
	"context"
	"fmt"
	"log"

	gokamy "github.com/Dieg0code/gokamy-ai"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client using your API key.
	client := openai.NewClient("YOUR_API_KEY")

	// Create simple memory instances for each agent.
	memoryAgentOne := gokamy.NewSimpleMemory()
	memoryAgentTwo := gokamy.NewSimpleMemory()

	// Build the first agent (HelloAgent).
	agentOne, err := gokamy.NewAgentBuilder().
		SetClient(client).
		SetName("HelloAgent").
		SetSystemPrompt("You are an agent that warmly greets users and encourages further interaction.").
		SetMemory(memoryAgentOne).
		SetModel(openai.GPT4).
		SetMaxRecursion(2).
		Build()
	if err != nil {
		log.Fatalf("Error building HelloAgent: %v", err)
	}

	// Build the second agent (FinalAgent).
	agentTwo, err := gokamy.NewAgentBuilder().
		SetClient(client).
		SetName("FinalAgent").
		SetSystemPrompt("You are an agent that provides a final summary based on the conversation.").
		SetMemory(memoryAgentTwo).
		SetModel(openai.GPT4).
		SetMaxRecursion(2).
		Build()
	if err != nil {
		log.Fatalf("Error building FinalAgent: %v", err)
	}

	// Create an orchestrator, register both agents, and define the execution sequence.
	orchestrator := gokamy.NewOrchestratorBuilder().
		AddAgent(agentOne).
		AddAgent(agentTwo).
		// Define the processing sequence: first HelloAgent, then FinalAgent.
		SetSequence([]string{"HelloAgent", "FinalAgent"}).
		Build()

	// Provide an input and process the sequence.
	input := "Please greet the user and provide a summary."
	response, err := orchestrator.ProcessSequence(context.Background(), input)
	if err != nil {
		log.Fatalf("Error processing sequence: %v", err)
	}

	fmt.Println("Final Orchestrator Response:")
	fmt.Println(response)
}
```
</details>

<details>
  <summary><strong>Tools</strong></summary>

Gokamy SDK includes functionality to automatically generate JSON schemas from Go structures. These generated schemas can be used to define and validate tools for your agents.

For example, consider the following tool definition that generates a JSON schema for a `Product` structure:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	gokamy "github.com/Dieg0code/gokamy-ai"
)

// Product represents a product with various attributes.
type Product struct {
	ID        int     `json:"id" description:"Unique product identifier" required:"true"`
	Name      string  `json:"name" description:"Product name" required:"true"`
	Category  string  `json:"category" description:"Category of the product" enum:"Electronic,Furniture,Clothing"`
	Price     float64 `json:"price" description:"Price of the product"`
	Available bool    `json:"available" description:"Product availability" required:"true"`
}

func main() {
	// Generate the JSON schema for the Product struct.
	schema, err := gokamy.GenerateSchema(Product{})
	if err != nil {
		log.Fatal(err)
	}
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
```

The schema generation leverages reflection along with custom struct tags (e.g., description, required, enum) to produce a JSON Schema that describes the tool's expected input. This schema can then be used to interface with language models or validate user-provided data.
</details>

<details>
  <summary><strong>Custom Memory Implementation</strong></summary>

In addition to the built-in SimpleMemory (an in-memory slice), Gokamy SDK allows you to create your own memory implementations. Simply ensure your implementation satisfies the `Memory` interface.

Example custom memory implementation:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    gomaky "github.com/Dieg0code/gokamy"
    openai "github.com/sashabaranov/go-openai"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Message represents the schema for storing chat messages.
type Message struct {
    gorm.Model
    Role      string
    Content   string
}

// ORMMemory is a custom Memory implementation that persists messages with GORM.
type ORMMemory struct {
    db *gorm.DB
}

// Add stores a new message in the database.
func (m *ORMMemory) Add(message openai.ChatCompletionMessage) {
    msg := Message{
        Role:      message.Role,
        Content:   message.Content,
    }
    if err := m.db.Create(&msg).Error; err != nil {
        log.Printf("failed to add message: %v", err)
    }
}

// Get retrieves all stored messages ordered by creation time.
func (m *ORMMemory) Get() []openai.ChatCompletionMessage {
    var messages []Message
	if err := m.db.Order("created_at").Find(&messages).Error; err != nil {
		log.Printf("failed to get messages: %v", err)
		return nil
	}

	var chatMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
}

// Clear removes all messages from the persistent memory.
func (m *ORMMemory) Clear() {
    // optional: implement clear functionality
}

// NewORMMemory returns a Memory interface backed by ORMMemory.
// It auto-migrates the Message table using GORM.
func NewORMMemory(db *gorm.DB) gomaky.Memory {
    if err := db.AutoMigrate(&Message{}); err != nil {
        log.Fatalf("AutoMigrate failed: %v", err)
    }
    return &ORMMemory{db: db}
}

func main() {
    // Set up the PostgreSQL DSN. Replace with your PostgreSQL credentials.
    dsn := "host=localhost user=postgres password=YOUR_PASSWORD dbname=your_db port=5432 sslmode=disable TimeZone=UTC"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }

    // Create ORM-based memory instances for each agent.
    memoryAgentOne := NewORMMemory(db)
    memoryAgentTwo := NewORMMemory(db)

    // Create an ORM-based memory instance for orchestrator global history.
    globalHistory := NewORMMemory(db)

    // Initialize the OpenAI client using your API key.
    client := openai.NewClient("YOUR_API_KEY")

    // Build the first agent (HelloAgent).
    agentOne, err := gomaky.NewAgentBuilder().
        SetClient(client).
        SetName("HelloAgent").
        SetSystemPrompt("You are an agent that warmly greets users and encourages further interaction.").
        SetMemory(memoryAgentOne).
        SetModel(openai.GPT4).
        SetMaxRecursion(2).
        Build()
    if err != nil {
        log.Fatalf("Error building HelloAgent: %v", err)
    }

    // Build the second agent (FinalAgent).
    agentTwo, err := gomaky.NewAgentBuilder().
        SetClient(client).
        SetName("FinalAgent").
        SetSystemPrompt("You are an agent that provides a final summary based on the conversation.").
        SetMemory(memoryAgentTwo).
        SetModel(openai.GPT4).
        SetMaxRecursion(2).
        Build()
    if err != nil {
        log.Fatalf("Error building FinalAgent: %v", err)
    }

    // Create an orchestrator, register both agents, and define the execution sequence.
    orchestrator := gomaky.NewOrchestratorBuilder().
        SetGlobalHistory(globalHistory).
        AddAgent(agentOne).
        AddAgent(agentTwo).
        // Define the processing sequence: first HelloAgent, then FinalAgent.
        SetSequence([]string{"HelloAgent", "FinalAgent"}).
        Build()

    // Provide an input and process the sequence.
    input := "Please greet the user and provide a summary."
    response, err := orchestrator.ProcessSequence(context.Background(), input)
    if err != nil {
        log.Fatalf("Error processing sequence: %v", err)
    }

    fmt.Println("Final Orchestrator Response:")
    fmt.Println(response)
}
```
</details>

<details>
  <summary><strong>Advanced Configuration: Temperature and JSON Response Format</strong></summary>


You can configure the behavior of your agent by configuring parameters such as temperature and JSON response format. The following example demonstrates how to set these options using the AgentBuilder:

```go
package main

import (
	"context"
	"fmt"
	"log"

	gokamy "github.com/Dieg0code/gokamy-ai"
	openai "github.com/sashabaranov/go-openai"
)

// MyResponse defines the expected JSON structure of the response.
type MyResponse struct {
	Message string   `json:"message" description:"The response message from the agent" required:"true"`
	Code    int      `json:"code" description:"The status code of the response" required:"true"`
	Status  string   `json:"status" description:"The status of the operation" enum:"success,failure" required:"true"`
	Details []string `json:"details" description:"Optional additional details about the response" required:"false"`
}

func main() {
	// Initialize the OpenAI client with your API key.
	client := openai.NewClient("YOUR_OPENAI_API_KEY")

	// Create a simple memory instance.
	memory := gokamy.NewSimpleMemory()

	// Set a basic system prompt.
	systemPrompt := "You are an advanced agent configured with custom settings. Please provide a JSON response following the expected format."

	// Build the agent with custom temperature and JSON response format.
	agent, err := gokamy.NewAgentBuilder().
		SetClient(client).
		SetName("AdvancedAgent").
		SetSystemPrompt(systemPrompt).
		SetMemory(memory).
		SetModel(openai.GPT4).
		SetMaxRecursion(2).
		SetTemperature(0.7).                // Set the temperature to influence randomness.
		SetJSONResponseFormat(MyResponse{}). // Set the expected JSON response format using tags.
		Build()
	if err != nil {
		log.Fatalf("Error building agent: %v", err)
	}

	// Process a sample input with the agent.
	response, err := agent.Process(context.Background(), "Provide a response in JSON format.")
	if err != nil {
		log.Fatalf("Error processing request: %v", err)
	}

	// Parse the JSON response into MyResponse struct.
	var parsedResponse MyResponse
	if err := json.Unmarshal([]byte(response), &parsedResponse); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	// Print the parsed response.
	fmt.Println("Agent Response:")
	fmt.Printf("Message: %s\n", parsedResponse.Message)
	fmt.Printf("Code: %d\n", parsedResponse.Code)
}

```
</details>

<details>
  <summary><strong>Future CLI Support</strong></summary>

The project is also planning a CLI tool to streamline project setup. The planned commands include:

- **`gokamy init`**: Initializes the project structure by creating an `agents` directory.
- **`gokamy new AgentName`**: Creates a new agent scaffold with placeholder files (e.g., AgentName.go, AgentNameTool.go, prompt.go).

Stay tuned for further updates!
</details>

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on [GitHub](https://github.com/Dieg0code/gokamy-ai).

## License

This project is licensed under the MIT License.