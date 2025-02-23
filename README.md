<div align="center">
  <img src="https://i.imgur.com/e608zH3.png" alt="Syndicate SDK Logo"/>

[![Go Report Card](https://goreportcard.com/badge/github.com/Dieg0Code/syndicate-go)](https://goreportcard.com/report/github.com/Dieg0Code/syndicate-go)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/Dieg0Code/syndicate-go/ci.yml?branch=main)](https://github.com/Dieg0Code/syndicate-go/actions)
[![GoDoc](https://godoc.org/github.com/Dieg0Code/syndicate-go?status.svg)](https://pkg.go.dev/github.com/Dieg0Code/syndicate-go)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/Dieg0Code/syndicate-go)](https://github.com/Dieg0Code/syndicate-go/releases)

</div>

Syndicate SDK is a powerful, agile, and extensible toolkit for crafting intelligent conversational agents in Golang. Designed with both flexibility and simplicity in mind, it empowers developers to build AI-driven agents capable of sophisticated prompt engineering, seamless tool integration, dynamic memory management, and orchestrating complex workflows. Whether you're prototyping or building production-grade solutions, Syndicate SDK lets you unleash the full potential of conversational AI with a lightweight framework that’s both hacker-friendly and enterprise-ready. 🚀

---

### ⚡ **For a quick and comprehensive overview of the SDK, check out the 👉[Quick Guide](https://github.com/Dieg0Code/syndicate-go/tree/main/examples/QuickGuide)👈** 📚✨🔥🚀

---

## Features

- **Agent Management 🤖:** Easily build and configure agents with custom system prompts, tools, and memory.
- **Prompt Engineering 📝:** Create structured prompts with nested sections for improved clarity.
- **Tool Schemas 🔧:** Generate JSON schemas from Go structures to define tools and validate user inputs.
- **Memory Implementations 🧠:** Use built-in SimpleMemory or implement your own memory storage that adheres to the Memory interface.
- **Syndicate 🦾:** Manage multiple agents and execute them in a predefined sequence to achieve complex workflows.
- **Extendable 🔐:** The SDK is designed to be unopinionated, allowing seamless integration into your projects.

## Installation

To install Syndicate SDK, use Go modules:

```bash
go get github.com/Dieg0Code/syndicate-go
```

Ensure that you have Go installed and configured in your development environment.

<details open>
  <summary><strong>Quick Start</strong></summary>

Below is a simple example demonstrating how to create an agent, define a tool, and execute a pipeline using Syndicate SDK. The example simulates processing a customer order and providing a summary of the conversation.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	syndicate "github.com/Dieg0Code/syndicate-go"
	openai "github.com/sashabaranov/go-openai"
)

// --------------------------
// Order Tool Implementation
// --------------------------

// OrderItem defines a single item in a customer's order.
type OrderItem struct {
	ItemName string `json:"item_name" description:"Name of the menu item" required:"true"`
	Quantity int    `json:"quantity" description:"Number of items ordered" required:"true"`
	Price    int    `json:"price" description:"Price of the item in cents" required:"true"`
}

// OrderSchema defines the structure for a complete order.
type OrderSchema struct {
	Items           []OrderItem `json:"items" description:"List of ordered items" required:"true"`
	DeliveryAddress string      `json:"delivery_address" description:"Delivery address for the order" required:"true"`
	CustomerName    string      `json:"customer_name" description:"Name of the customer" required:"true"`
	PhoneNumber     string      `json:"phone_number" description:"Customer's phone number" required:"true"`
	PaymentMethod   string      `json:"payment_method" description:"Payment method (cash or transfer)" required:"true" enum:"cash,transfer"`
}

// OrderTool implements syndicate.Tool and simulates saving an order.
type OrderTool struct{}

// NewOrderTool returns a new instance of OrderTool.
func NewOrderTool() syndicate.Tool {
	return &OrderTool{}
}

// GetDefinition generates the tool definition using the OrderSchema.
func (ot *OrderTool) GetDefinition() syndicate.ToolDefinition {
	schema, err := syndicate.GenerateRawSchema(OrderSchema{})
	if err != nil {
		log.Fatal(err)
	}

	return syndicate.ToolDefinition{
		Name:        "OrderProcessor",
		Description: "Processes customer orders by saving order details (items, address, customer info, and payment method).",
		Parameters:  schema,
		Strict:      true,
	}
}

// Execute simulates processing the order by printing it and returning a success message.
func (ot *OrderTool) Execute(args json.RawMessage) (interface{}, error) {
	var order OrderSchema
	if err := json.Unmarshal(args, &order); err != nil {
		return nil, err
	}

	// Here you could insert the order into a database or call an external service.
	fmt.Printf("Processing Order: %+v\n", order)
	return "Order processed successfully", nil
}

// --------------------------
// Custom Memory Implementation
// --------------------------

// CustomMemory is an example of a custom memory implementation using SQLite.
type CustomMemory struct {
	db    *gorm.DB
	mutex sync.RWMutex
}

// NewCustomMemory initializes a new CustomMemory instance with the provided Gorm DB.
func NewCustomMemory(db *gorm.DB) syndicate.Memory {
	return &CustomMemory{
		db: db,
	}
}

// Add stores a new message into memory (database).
func (m *CustomMemory) Add(message syndicate.Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// Example: You could save the message to the database here.
	// m.db.Create(&message)
	fmt.Printf("Memory Add: %+v\n", message)
}

// Get retrieves stored messages from memory.
// For simplicity, this example returns nil (in a real implementation, fetch from the DB).
func (m *CustomMemory) Get() []syndicate.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	// Example: Retrieve recent messages from the database.
	return nil
}

// Clear clears all stored messages.
func (m *CustomMemory) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// Example: Delete messages from the database.
	// m.db.Where("1 = 1").Delete(&syndicate.Message{})
	fmt.Println("Memory cleared")
}

// --------------------------
// Main Function & Pipeline
// --------------------------

func main() {
	// Initialize the OpenAI client with your API key.
	client := syndicate.NewOpenAIClient("YOUR_OPENAI_API_KEY")

	// Initialize a Gorm DB connection to SQLite for our custom memory.
	db, err := gorm.Open(sqlite.Open("memory.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	customMemory := NewCustomMemory(db)

	// Create an instance of our OrderTool.
	orderTool := NewOrderTool()

	// Build the first agent: OrderAgent.
	// This agent processes customer orders and can call the OrderTool.
	orderAgent, err := syndicate.NewAgent().
		SetClient(client).
		SetName("OrderAgent").
		SetConfigPrompt("You are an agent that processes customer orders. If the input contains order details, call the OrderProcessor tool to process the order.").
		SetMemory(customMemory).
		SetModel(openai.GPT4).
		EquipTool(orderTool). // Equip the OrderTool.
		Build()
	if err != nil {
		log.Fatalf("Error building OrderAgent: %v", err)
	}

	// Build the second agent: SummaryAgent.
	// This agent provides a final summary of the conversation.
	summaryAgent, err := syndicate.NewAgent().
		SetClient(client).
		SetName("SummaryAgent").
		SetConfigPrompt("You are an agent that provides a final summary confirming that the order was processed and reiterating key details.").
		SetMemory(customMemory).
		SetModel(openai.GPT4).
		Build()
	if err != nil {
		log.Fatalf("Error building SummaryAgent: %v", err)
	}

	// Create a Syndicate system and define the processing pipeline.
	syndicateSystem := syndicate.NewSyndicate().
		RecruitAgent(orderAgent).
		RecruitAgent(summaryAgent).
		DefinePipeline([]string{"OrderAgent", "SummaryAgent"}).
		Build()

	// Simulate a user input written in natural language.
	userName := "Alice"
	input := `Hi, I'd like to place an order. I want two Margherita Pizzas and one Coke. Please deliver them to 123 Main Street in Springfield. My name is Alice, my phone number is 555-1234, and I'll pay with cash.`

	// Execute the pipeline.
	response, err := syndicateSystem.ExecutePipeline(context.Background(), userName, input)
	if err != nil {
		log.Fatalf("Error executing pipeline: %v", err)
	}

	fmt.Println("\nFinal Syndicate Response:")
	fmt.Println(response)
}

```

</details>

<details>
  <summary><strong>Config Prompt Builder</strong></summary>

The Config Prompt Builder is a utility that simplifies the creation of agent configuration prompts. It allows you to define a structured prompt using a fluent API, making it easier to configure agents with specific instructions.

```go
package main

import (
	"fmt"
	"time"

	syndicate "github.com/Dieg0Code/syndicate-go"
)

type MenuItem struct {
	Name        string
	Description string
	Pricing     int
}

// NewConfigPrompt generates the system prompt in XML format
func NewConfigPrompt(name string, additionalContext MenuItem) string {
	configPrompt := syndicate.NewPromptBuilder().
		// Introduction section
		CreateSection("Introduction").
		AddText("Introduction", "You are an agent who provides detailed information about the menu, dishes, and key restaurant data using a semantic search system to enrich responses with relevant context.").

		// Agent identity
		CreateSection("Identity").
		AddText("Identity", "This section defines your name and persona identity.").
		AddSubSection("Name", "Identity").
		AddTextF("Name", name).
		AddSubSection("Persona", "Identity").
		AddText("Persona", "You act as an AI assistant in the restaurant, interacting with customers in a friendly and helpful manner to improve their dining experience.").

		// Capabilities and behavior
		CreateSection("CapabilitiesAndBehavior").
		AddListItem("CapabilitiesAndBehavior", "Respond in a clear and friendly manner, tailoring your answer to the user's query.").
		AddListItem("CapabilitiesAndBehavior", "Provide details about dishes (ingredients, preparation, allergens) and suggest similar options if appropriate.").
		AddListItem("CapabilitiesAndBehavior", "Promote the restaurant positively, emphasizing the quality of dishes and the dining experience.").
		AddListItem("CapabilitiesAndBehavior", "Be cheerful, polite, and respectful at all times; use emojis if appropriate.").
		AddListItem("CapabilitiesAndBehavior", "Register or cancel orders but do not update them; inform the user accordingly.").
		AddListItem("CapabilitiesAndBehavior", "Remember only the last 5 interactions with the user.").

		// Additional context
		CreateSection("AdditionalContext").
		AddText("AdditionalContext", "This section contains additional information about the available dishes used to answer user queries based on semantic similarity.").
		AddSubSection("Menu", "AdditionalContext").
		AddTextF("Menu", additionalContext).
		AddSubSection("CurrentDate", "AdditionalContext").
		AddTextF("CurrentDate", time.Now().Format(time.RFC3339)).
		AddListItem("AdditionalContext", "Select dishes based on similarity without mentioning it explicitly.").
		AddListItem("AdditionalContext", "Use context to enrich responses, but do not reveal it.").
		AddListItem("AdditionalContext", "Offer only dishes available on the menu.").

		// Limitations and directives
		CreateSection("LimitationsAndDirectives").
		AddListItem("LimitationsAndDirectives", "Do not invent data or reveal confidential information.").
		AddListItem("LimitationsAndDirectives", "Redirect unrelated topics to relevant restaurant topics.").
		AddListItem("LimitationsAndDirectives", "Strictly provide information only about the restaurant and its menu.").
		AddListItem("LimitationsAndDirectives", "Offer only available menu items; do not invent dishes.").

		// Response examples
		CreateSection("ResponseExamples").
		AddListItem("ResponseExamples", "If asked about a dish, provide details only if it is on the menu.").
		AddListItem("ResponseExamples", "If asked for recommendations, suggest only from the available menu.").
		AddListItem("ResponseExamples", "If asked for the menu, list only available dishes.").

		// Final considerations
		CreateSection("FinalConsiderations").
		AddText("FinalConsiderations", "**You must follow these directives to ensure an optimal user experience, otherwise you will be dismissed.**").
		Build()

	return configPrompt
}

func main() {
	// Define the additional context for the agent.
	additionalContext := MenuItem{
		Name:        "Spaghetti Carbonara",
		Description: "A classic Italian pasta dish consisting of eggs, cheese, pancetta, and black pepper.",
		Pricing:     15,
	}

	// Generate the system prompt for the agent.
	configPrompt := NewConfigPrompt("Bob", additionalContext)
	fmt.Println(configPrompt)
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
