<div align="center">
  <img src="https://i.imgur.com/e608zH3.png" alt="Syndicate SDK Logo"/>
  
[![Go Report Card](https://goreportcard.com/badge/github.com/Dieg0Code/syndicate-go)](https://goreportcard.com/report/github.com/Dieg0Code/syndicate-go)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/Dieg0Code/syndicate-go/ci.yml?branch=main)](https://github.com/Dieg0Code/syndicate-go/actions)
[![codecov](https://codecov.io/github/Dieg0Code/syndicate-go/graph/badge.svg?token=FXYY1S9EP4)](https://codecov.io/github/Dieg0Code/syndicate-go)
[![GoDoc](https://godoc.org/github.com/Dieg0Code/syndicate-go?status.svg)](https://pkg.go.dev/github.com/Dieg0Code/syndicate-go)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/Dieg0Code/syndicate-go)](https://github.com/Dieg0Code/syndicate-go/releases)
</div>

Syndicate SDK is a powerful, agile, and extensible toolkit for crafting intelligent conversational agents in Golang. Designed with both flexibility and simplicity in mind, it empowers developers to build AI-driven agents capable of sophisticated prompt engineering, seamless tool integration, dynamic memory management, and orchestrating complex workflows. Whether you're prototyping or building production-grade solutions, Syndicate SDK lets you unleash the full potential of conversational AI with a lightweight framework that’s both hacker-friendly and enterprise-ready. 🚀

---

### ⚡ **Quick Guide Overview**

For a complete summary of the SDK features, head over to our 👉 [Quick Guide](https://github.com/Dieg0Code/syndicate-go/tree/main/examples/QuickGuide) 👈 📚✨

---

## Installation

To install Syndicate SDK, use Go modules:

```bash
go get github.com/Dieg0Code/syndicate-go
```

Ensure that you have Go installed and properly configured.

---

## Conceptual Overview & Features

### **Agent Management 🤖**

_What it is:_  
Agents are the core of the Syndicate SDK. They are self-contained AI entities, each with its own personality, knowledge, and optional tools.

_Why it's useful:_

- **Customization:** Tailor each agent's behavior and configuration using custom system prompts.
- **Scalability:** Easily add or modify agents based on evolving requirements.
- **Modularity:** Agents can be combined to form complex workflows (pipelines) that mirror real-world processes.

_Example:_ Use one agent to process an order, and another to provide a summary—each handling different tasks for a comprehensive experience! 😄

---

### **Prompt Engineering 📝**

_What it is:_  
Prompt engineering involves crafting detailed and structured instructions that govern how agents respond.

_Why it's useful:_

- **Clarity:** Creating nested sections and structured prompts ensures the agent clearly understands its role.
- **Consistency:** Maintain a standard set of behaviors and responses, which is crucial for reliability in production scenarios.
- **Adaptability:** Easily update prompts to adapt to new requirements or contextual changes.

_Fun fact:_ This helps guide the AI to focus on specific details, ensuring conversations stay on point! ✨

---

### **Tool Schemas 🔧**

_What it is:_  
Tool schemas are generated JSON definitions from Go structures. They describe how external tools interact with the agents.

_Why it's useful:_

- **Validation:** Automatically validates inputs, ensuring only correct and safe data is passed to tools.
- **Integration:** Seamlessly integrates with external services or databases to perform specialized tasks.
- **Automation:** Reduces manual configuration, making development faster and less error-prone.

_Imagine:_ Your agent can now reliably call external services with a clearly defined contract! 💡

---

### **Memory Implementations 🧠**

_What it is:_  
Memory components allow agents to record, retrieve, and clear conversational context. This can be as simple as an in-memory store or an implementation using databases like SQLite.

_Why it's useful:_

- **Context Preservation:** Maintain important details across multiple interactions, helping agents provide more informed answers.
- **Customization:** Developers can implement custom memory solutions to match performance or persistence requirements.
- **Scalability:** Enhances the conversational flow by keeping track of conversation history.

_For example:_ Using a custom memory lets you record previous orders or context for follow-up questions! 📚

---

### **Syndicate (Orchestrator) 🦾**

_What it is:_  
Syndicate is the engine that recruits agents and establishes pipelines—dictating the order in which agents perform tasks.

_Why it's useful:_

- **Workflow Management:** Orchestrate complex processes by chaining multiple agents together.
- **Simplicity:** Simplifies the process of switching between agents or forming multi-step conversations.
- **Flexibility:** Adapt to different workflows without reworking the core logic each time.

_In practice:_ Easily create pipelines where one agent processes the order and another provides a summary. Coordination made simple! 🤝

---

### **Extendable 🔐**

_What it is:_  
The SDK is designed to be unopinionated, meaning it can be easily integrated into a variety of projects without forcing rigid architectural decisions.

_Why it's useful:_

- **Flexibility:** Choose your own tools, memory implementations, and configuration strategies.
- **Integration:** Seamlessly integrate into existing projects without major rework.
- **Future-Proofing:** Adapt the SDK to new requirements, features, or third-party integrations.

_The outcome:_ A robust toolkit that adapts to your needs whether you’re building a prototype or a production-grade system! 🔄

## Quick Start Example: Step-by-Step Instructions 😊

For a hands-on demo, follow these steps to create agents, define a tool, implement custom memory, and run the processing pipeline.

### **Step 1: Order Tool Implementation 🍕**

1. **Define the Order Data Structures:**  
   Create the structs `OrderItem` and `OrderSchema` to represent a customer's order.

2. **Implement the Order Tool:**  
   Create a new tool that processes the order:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	syndicate "github.com/Dieg0Code/syndicate-go"
)

// OrderItem defines a single item in a customer's order.
type OrderItem struct {
	ItemName string `json:"item_name" description:"Name of the menu item" required:"true"`
	Quantity int    `json:"quantity" description:"Number of items ordered" required:"true"`
	Price    int    `json:"price" description:"Price in cents" required:"true"`
}

// OrderSchema defines the complete order structure.
type OrderSchema struct {
	Items           []OrderItem `json:"items" description:"List of ordered items" required:"true"`
	DeliveryAddress string      `json:"delivery_address" description:"Delivery address" required:"true"`
	CustomerName    string      `json:"customer_name" description:"Customer name" required:"true"`
	PhoneNumber     string      `json:"phone_number" description:"Customer phone number" required:"true"`
	PaymentMethod   string      `json:"payment_method" description:"Payment method (cash or transfer)" required:"true" enum:"cash,transfer"`
}

// OrderTool simulates an order processing tool.
type OrderTool struct{}

// NewOrderTool returns a new OrderTool instance.
func NewOrderTool() syndicate.Tool {
	return &OrderTool{}
}

// GetDefinition generates the tool definition.
func (ot *OrderTool) GetDefinition() syndicate.ToolDefinition {
	schema, err := syndicate.GenerateRawSchema(OrderSchema{})
	if err != nil {
		log.Fatal(err)
	}
	return syndicate.ToolDefinition{
		Name:        "OrderProcessor",
		Description: "Processes orders by saving details (items, address, customer info, and payment method).",
		Parameters:  schema,
		Strict:      true,
	}
}

// Execute processes the order.
func (ot *OrderTool) Execute(args json.RawMessage) (interface{}, error) {
	var order OrderSchema
	if err := json.Unmarshal(args, &order); err != nil {
		return nil, err
	}
	fmt.Printf("Processing Order: %+v\n", order)
	return "Order processed successfully", nil
}
```

---

### **Step 2: Custom Memory Implementation 🧠**

Implement a simple custom memory using SQLite for persisting messages:

```go
package main

import (
	"fmt"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	syndicate "github.com/Dieg0Code/syndicate-go"
)

// CustomMemory stores messages using SQLite.
type CustomMemory struct {
	db    *gorm.DB
	mutex sync.RWMutex
}

// NewCustomMemory initializes CustomMemory with a Gorm DB.
func NewCustomMemory(db *gorm.DB) syndicate.Memory {
	return &CustomMemory{
		db: db,
	}
}

// Add saves a message.
func (m *CustomMemory) Add(message syndicate.Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	fmt.Printf("Memory Add: %+v\n", message)
	// Optionally, save to db here.
}

// Get retrieves stored messages (returns nil for simplicity).
func (m *CustomMemory) Get() []syndicate.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return nil
}

// Clear erases all messages.
func (m *CustomMemory) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	fmt.Println("Memory cleared")
}
```

---

### **Step 3: Building the Agents & Pipeline 🚀**

Combine everything by creating two agents (OrderAgent and SummaryAgent) and a pipeline:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	syndicate "github.com/Dieg0Code/syndicate-go"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize OpenAI client.
	client := syndicate.NewOpenAIClient("YOUR_OPENAI_API_KEY")

	// Set up SQLite connection.
	db, err := gorm.Open(sqlite.Open("memory.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	customMemory := NewCustomMemory(db)

	// Step 3A: Create the Order Tool.
	orderTool := NewOrderTool()

	// Step 3B: Build the OrderAgent.
	orderAgent, err := syndicate.NewAgent().
		SetClient(client).
		SetName("OrderAgent").
		SetConfigPrompt("You are an agent processing customer orders. If the input contains order details, call the OrderProcessor tool.").
		SetMemory(customMemory).
		SetModel(openai.GPT4).
		EquipTool(orderTool). // Equip OrderTool
		Build()
	if err != nil {
		log.Fatalf("Error building OrderAgent: %v", err)
	}

	// Step 3C: Build the SummaryAgent.
	summaryAgent, err := syndicate.NewAgent().
		SetClient(client).
		SetName("SummaryAgent").
		SetConfigPrompt("You are an agent providing the final summary of the order.").
		SetMemory(customMemory).
		SetModel(openai.GPT4).
		Build()
	if err != nil {
		log.Fatalf("Error building SummaryAgent: %v", err)
	}

	// Step 3D: Define the pipeline with both agents.
	syndicateSystem := syndicate.NewSyndicate().
		RecruitAgent(orderAgent).
		RecruitAgent(summaryAgent).
		DefinePipeline([]string{"OrderAgent", "SummaryAgent"}).
		Build()

	// Simulate user input.
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

---

### **Step 4: Run & Enjoy 🎉**

1. Replace `"YOUR_OPENAI_API_KEY"` with your actual API key.
2. Build and run your application.
3. Watch as the agents process and summarize the order step-by-step!

---

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

// NewConfigPrompt generates the system prompt in XML format.
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

---

## Dependencies and Their Licenses

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) - Licensed under **Apache License 2.0**
- [cohesion-org/deepseek-go](https://github.com/cohesion-org/deepseek-go) - Licensed under **MIT License**

Refer to their repositories for full license texts.

---

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on [GitHub](https://github.com/Dieg0Code/syndicate-go).

Happy coding! 👩‍💻👨‍💻
