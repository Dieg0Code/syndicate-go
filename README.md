# Syndicate AI

<div align="center">
  <br /><br />
  
  [![Go Report Card](https://goreportcard.com/badge/github.com/Dieg0Code/syndicate)](https://goreportcard.com/report/github.com/Dieg0Code/syndicate)
  [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/Dieg0Code/syndicate/ci.yml?branch=main)](https://github.com/Dieg0Code/syndicate/actions)
  [![GoDoc](https://godoc.org/github.com/Dieg0Code/syndicate?status.svg)](https://pkg.go.dev/github.com/Dieg0Code/syndicate)
  [![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
  [![Release](https://img.shields.io/github/v/release/Dieg0Code/syndicate)](https://github.com/Dieg0Code/syndicate/releases)
</div>

Syndicate SDK is a lightweight, flexible, and extensible toolkit for building intelligent conversational agents in Golang. It enables you to create agents, engineer prompts, generate tool schemas, manage memory, and orchestrate complex workflows—making it easy to integrate advanced AI capabilities into your projects. 🚀

## Features

- **Agent Management 🤖:** Easily build and configure agents with custom system prompts, tools, and memory.
- **Prompt Engineering 📝:** Create structured prompts with nested sections for improved clarity.
- **Tool Schemas 🔧:** Generate JSON schemas from Go structures to define tools and validate user inputs.
- **Memory Implementations 🧠:** Use built-in SimpleMemory or implement your own memory storage that adheres to the Memory interface.
- **Orchestrator 🎛️:** Manage multiple agents and execute them in a predefined sequence to achieve complex workflows.
- **Extendable 🔌:** The SDK is designed to be unopinionated, allowing seamless integration into your projects.
- **CLI (Planned) 🚀:** Future CLI support to initialize projects and scaffold agents with simple commands.

## Installation

To install Syndicate SDK, use Go modules:

```bash
go get github.com/Dieg0Code/syndicate
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

	syndicate "github.com/Dieg0Code/syndicate"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client with your API key.
	client := openai.NewClient("YOUR_OPENAI_API_KEY")

	// Create a simple memory instance.
	memory := syndicate.NewSimpleMemory()

	// Build a structured prompt using PromptBuilder to instruct the agent to respond in haiku.
	systemPrompt := syndicate.NewPromptBuilder(). 
		CreateSection("Introduction"). 
		AddText("Introduction", "You are an agent that always responds in haiku format."). 
		CreateSection("Guidelines"). 
		AddListItem("Guidelines", "Keep responses in a three-line haiku format (5-7-5 syllables)."). 
		AddListItem("Guidelines", "Be creative and concise."). 
		Build()

	fmt.Println("System Prompt:")
	fmt.Println(systemPrompt)

	// Build the agent using AgentBuilder.
	agent, err := syndicate.NewAgentBuilder().
		SetClient(client).
		SetName("HaikuAgent").
		SetConfigPrompt(systemPrompt).
		SetMemory(memory).
		SetModel(openai.GPT4).
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

Syndicate SDK also provides an orchestrator to manage a sequence of agents. In the following example, two agents are created and executed in sequence:

```go
package main

import (
	"context"
	"fmt"
	"log"

	syndicate "github.com/Dieg0Code/syndicate"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client using your API key.
	client := openai.NewClient("YOUR_API_KEY")

	// Create simple memory instances for each agent.
	memoryAgentOne := syndicate.NewSimpleMemory()
	memoryAgentTwo := syndicate.NewSimpleMemory()

	// Build the first agent (HelloAgent).
	agentOne, err := syndicate.NewAgentBuilder().
		SetClient(client).
		SetName("HelloAgent").
		SetConfigPrompt("You are an agent that warmly greets users and encourages further interaction.").
		SetMemory(memoryAgentOne).
		SetModel(openai.GPT4).
		Build()
	if err != nil {
		log.Fatalf("Error building HelloAgent: %v", err)
	}

	// Build the second agent (FinalAgent).
	agentTwo, err := syndicate.NewAgentBuilder().
		SetClient(client).
		SetName("FinalAgent").
		SetConfigPrompt("You are an agent that provides a final summary based on the conversation.").
		SetMemory(memoryAgentTwo).
		SetModel(openai.GPT4).
		Build()
	if err != nil {
		log.Fatalf("Error building FinalAgent: %v", err)
	}

	// Create an orchestrator, register both agents, and define the execution sequence.
	orchestrator := syndicate.NewOrchestratorBuilder().
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

Syndicate SDK includes functionality to automatically generate JSON schemas from Go structures. These generated schemas can be used to define and validate tools for your agents.

For example, consider the following tool definition that generates a JSON schema for a `Product` structure:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	syndicate "github.com/Dieg0Code/syndicate"
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
	schema, err := syndicate.GenerateSchema(Product{})
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

## Dependencies and Their Licenses

This project uses the following third-party libraries:

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) - Licensed under **Apache License 2.0**
- [cohesion-org/deepseek-go](https://github.com/cohesion-org/deepseek-go) - Licensed under **MIT License**

Please refer to their respective repositories for the full license texts.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on [GitHub](https://github.com/Dieg0Code/syndicate).