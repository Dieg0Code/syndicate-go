<div align="center">
  <img src="https://i.imgur.com/e608zH3.png" alt="Syndicate SDK Logo"/>

[![Go Report Card](https://goreportcard.com/badge/github.com/Dieg0Code/syndicate-go)](https://goreportcard.com/report/github.com/Dieg0Code/syndicate-go)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/Dieg0Code/syndicate-go/ci.yml?branch=main)](https://github.com/Dieg0Code/syndicate-go/actions)
[![GoDoc](https://godoc.org/github.com/Dieg0Code/syndicate-go?status.svg)](https://pkg.go.dev/github.com/Dieg0Code/syndicate-go)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/Dieg0Code/syndicate-go)](https://github.com/Dieg0Code/syndicate-go/releases)

</div>

Syndicate SDK is a lightweight, flexible, and extensible toolkit for building intelligent conversational agents in Golang. It enables you to create agents, engineer prompts, generate tool schemas, manage memory, and orchestrate complex workflows—making it easy to integrate advanced AI capabilities into your projects. 🚀

## Features

- **Agent Management 🤖:** Easily build and configure agents with custom system prompts, tools, and memory.
- **Prompt Engineering 📝:** Create structured prompts with nested sections for improved clarity.
- **Tool Schemas 🔧:** Generate JSON schemas from Go structures to define tools and validate user inputs.
- **Memory Implementations 🧠:** Use built-in SimpleMemory or implement your own memory storage that adheres to the Memory interface.
- **Orchestrator 🎧:** Manage multiple agents and execute them in a predefined sequence to achieve complex workflows.
- **Extendable 🔐:** The SDK is designed to be unopinionated, allowing seamless integration into your projects.

---

#### ⚡ **For a quick and comprehensive overview of the SDK, check out the 👉[Quick Guide](https://github.com/Dieg0Code/syndicate-go/tree/main/examples/QuickGuide)👈** 📚✨🔥🚀

---

## Installation

To install Syndicate SDK, use Go modules:

```bash
go get github.com/Dieg0Code/syndicate-go
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

	syndicate "github.com/Dieg0Code/syndicate-go"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client with your API key.
	client := syndicate.NewOpenAIClient("YOUR_OPENAI_API_KEY")

	// Create a simple memory instance.
	memory := syndicate.NewSimpleMemory()

	// Build a structured prompt using PromptBuilder to instruct the agent to speak in haiku.
	systemPrompt := syndicate.NewPromptBuilder().
		CreateSection("Introduction").
		AddText("Introduction", "You are an agent that always responds in haiku format.").
		CreateSection("Guidelines").
		AddListItem("Guidelines", "Keep responses to a three-line haiku format (5-7-5 syllables).").
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
	response, err := agent.Process(context.Background(), "Jhon Doe", "What is the weather like today?")
	if err != nil {
		log.Fatalf("Error processing request: %v", err)
	}

	fmt.Println("\nAgent Response:")
	fmt.Println(response)
}

```

</details>

## Dependencies and Their Licenses

This project uses the following third-party libraries:

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) - Licensed under **Apache License 2.0**
- [cohesion-org/deepseek-go](https://github.com/cohesion-org/deepseek-go) - Licensed under **MIT License**

Please refer to their respective repositories for the full license texts.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on [GitHub](https://github.com/Dieg0Code/syndicate-go).
