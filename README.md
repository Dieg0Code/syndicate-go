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

A Go SDK for building and orchestrating intelligent AI agents that seamlessly connect to LLMs, tools, and workflows without the complexity of direct API management.

## 🚀 Project Status

**Current status:** Beta - Stable API but under active development  
**Version:** v0.2.0  
**Go Version:** 1.24+

## 📦 Installation

```bash
go get github.com/Dieg0Code/syndicate-go
```

## 🔑 Key Features

### 🤖 Agent Management

Create AI entities with distinct personalities, knowledge bases, and toolsets. Agents can work independently or together in pipelines to handle complex workflows.

### 🧠 Prompt Engineering

Create structured, detailed prompts that guide agent behavior with consistent responses. The SDK includes utilities for building and managing sophisticated prompts.

### 🛠️ Tool Integration

Connect agents with external tools and services using automatically generated JSON schemas from Go structures, complete with validation.

### 💾 Memory Management

Implement customizable memory systems to maintain context across conversations, with support for various storage backends from in-memory to databases.

### 🔄 Workflow Orchestration

Build multi-agent pipelines that process information sequentially, enabling complex conversational workflows that mirror real-world processes.

## 🔍 Quick Example

```go
package main

import (
    "context"
    "fmt"

    syndicate "github.com/Dieg0Code/syndicate-go"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    // Initialize OpenAI client
    client := syndicate.NewOpenAIClient("YOUR_API_KEY")

    // Create an order processing agent
    orderAgent, _ := syndicate.NewAgent().
        SetClient(client).
        SetName("OrderAgent").
        SetConfigPrompt("You process customer orders.").
        SetModel(openai.GPT4).
        Build()

    // Create a summary agent
    summaryAgent, _ := syndicate.NewAgent().
        SetClient(client).
        SetName("SummaryAgent").
        SetConfigPrompt("You summarize order details.").
        SetModel(openai.GPT4).
        Build()

    // Create a pipeline with both agents
    system := syndicate.NewSyndicate().
        RecruitAgent(orderAgent).
        RecruitAgent(summaryAgent).
        DefinePipeline([]string{"OrderAgent", "SummaryAgent"}).
        Build()

    // Process user input
    response, _ := system.ExecutePipeline(
        context.Background(),
        "User",
        "I'd like to order two pizzas for delivery to 123 Main St."
    )

    fmt.Println(response)
}
```

For a complete step-by-step guide with tool integration and custom memory implementation, see our [detailed examples](https://github.com/Dieg0Code/syndicate-go/tree/main/examples).

## 🛠️ Advanced Features

<details>
  <summary><b>Tool Integration</b></summary>

Integrate external tools with agents using JSON schemas. The SDK automatically generates schemas from Go structures, allowing for easy validation and integration.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	syndicate "github.com/Dieg0Code/syndicate-go"
)

// 📝 Defining the schema for menu items
type MenuItemSchema struct {
	ItemName string `json:"item_name" description:"Menu item name" required:"true"`
	Quantity int    `json:"quantity" description:"Quantity ordered by the user" required:"true"`
	Price    int    `json:"price" description:"Menu item price" required:"true"`
}

// 📝 Defining the schema for the user's order
type UserOrderFunctionSchema struct {
	MenuItems       []MenuItemSchema `json:"menu_items" description:"List of ordered menu items" required:"true"`
	DeliveryAddress string           `json:"delivery_address" description:"Order delivery address" required:"true"`
	UserName        string           `json:"user_name" description:"User's name placing the order" required:"true"`
	PhoneNumber     string           `json:"phone_number" description:"User's phone number" required:"true"`
	PaymentMethod   string           `json:"payment_method" description:"Payment method (cash or transfer only)" required:"true" enum:"cash,transfer"`
}

func main() {
	// 🏗️ Generate the JSON schema
	schema, err := syndicate.GenerateRawSchema(UserOrderFunctionSchema{})
	if err != nil {
		log.Fatal(err)
	}

	// 🎨 Pretty-print the schema
	pretty, err := json.MarshalIndent(json.RawMessage(schema), "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// 📜 Display the generated schema
	fmt.Println("UserOrderFunction schema:")

fmt.Println(string(pretty))
}
```

---

### 🏗️ What does `GenerateRawSchema` do?

The function `GenerateRawSchema` returns a value of type `json.RawMessage`, which is just an alias for `[]byte`. This contains the **JSON schema** we need to define our **Tool**. 🛠️🔧

This structure generates the following JSON schema: 🎯

```json
{
  "type": "object",
  "properties": {
    "delivery_address": {
      "type": "string",
      "description": "Order delivery address"
    },
    "menu_items": {
      "type": "array",
      "description": "List of ordered menu items",
      "items": {
        "type": "object",
        "properties": {
          "item_name": {
            "type": "string",
            "description": "Menu item name"
          },
          "price": {
            "type": "integer",
            "description": "Menu item price"
          },
          "quantity": {
            "type": "integer",
            "description": "Quantity ordered by the user"
          }
        },
        "required": ["item_name", "quantity", "price"],
        "additionalProperties": false
      }
    },
    "payment_method": {
      "type": "string",
      "description": "Payment method (cash or transfer only)",
      "enum": ["cash", "transfer"]
    },
    "phone_number": {
      "type": "string",
      "description": "User's phone number"
    },
    "user_name": {
      "type": "string",
      "description": "User's name placing the order"
    }
  },
  "required": [
    "menu_items",
    "delivery_address",
    "user_name",
    "phone_number",
    "payment_method"
  ],
  "additionalProperties": false
}
```

### 🔄 Deserializing the Response

We can use the same Go structure to capture the response and deserialize it into a Go object. 🧑‍💻📦 This makes it easier to handle the data in your application.

#### Definition of Jsonschemas and Their Handlers 🚀

Now that we know how to create Tools for the LLM, the question arises: **How do we tell the LLM what to do with that information?** 🤔 To do that, we need to define a **`Handler`** for each `Tool`.

Manually creating the logic to distinguish between when the LLM responds with a normal message or with a call to a `Tool` can be tedious and error-prone. 😅 That's why `Syndicate` offers a way to define Handlers for each `Tool`, which are responsible for processing the information the LLM receives.

To achieve this, we have the **`Tool`** interface:

```go
type Tool interface {
	GetDefinition() ToolDefinition   // Returns the definition of the tool (name, description, parameters, etc.)
	Execute(args json.RawMessage) (interface{}, error)  // Executes the tool with the given arguments
}
```

The SDK requires you to implement this interface in order to associate tools with an agent. The interface has two methods:

- **`GetDefinition`**: Returns the definition of the tool, which includes the name, description, parameters, and whether it's strict or not. 📜
- **`Execute`**: This is the method called when the LLM makes a call to the tool. It receives the arguments for the call and returns an object that can be anything, but it must be something that can be converted to a string, since the result of calling the tool will later be passed back to the LLM for further processing. 🔄

Here's an example of what a `Handler` for the `SaveOrder` tool might look like: 🎯

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    syndicate "github.com/Dieg0Code/syndicate-go"
)

type MenuItemSchema struct {
	ItemName string `json:"item_name" description:"Menu item name" required:"true"`
	Quantity int    `json:"quantity" description:"Quantity ordered by the user" required:"true"`
	Price    int    `json:"price" description:"Menu item price" required:"true"`
}

type UserOrderFunctionSchema struct {
	MenuItems       []MenuItemSchema `json:"menu_items" description:"List of ordered menu items" required:"true"`
	DeliveryAddress string           `json:"delivery_address" description:"Order delivery address" required:"true"`
	UserName        string           `json:"user_name" description:"User's name placing the order" required:"true"`
	PhoneNumber     string           `json:"phone_number" description:"User's phone number" required:"true"`
	PaymentMethod   string           `json:"payment_method" description:"Payment method (cash or transfer only)" required:"true" enum:"cash,transfer"`
}

type SaveOrderTool struct {
    // Here you can add any necessary fields to process the call
}

func NewSaveOrderTool() syndicate.Tool {
    return &SaveOrderTool{}
}

func (s *SaveOrderTool) GetDefinition() syndicate.ToolDefinition {
    schema, err := syndicate.GenerateRawSchema(UserOrderFunctionSchema{})
    if err != nil {
        log.Fatal(err)
    }

    return syndicate.ToolDefinition{
        Name:        "SaveOrder",
        Description: "Retrieves the user's order. The user must provide the requested menu items, delivery address, name, phone number, and payment method. The payment method can only be cash or bank transfer.",
        Parameters:  schema,
        Strict:      true,
    }
}

func (s *SaveOrderTool) Execute(args json.RawMessage) (interface{}, error) {
    var order UserOrderFunctionSchema
    if err := json.Unmarshal(args, &order); err != nil {
        return nil, err
    }

    // You can do whatever you want with the order information here
    // Save it to a database, send it to an external service, etc.
    // It's up to you.
    // Usually, you'll want to inject a repo dependency into the SaveOrderTool struct and constructor
    // and use it here to store the information.
    fmt.Printf("Order received: %+v\n", order)

    return "Order received successfully", nil
}

func main() {
    // Create a new instance of the tool
    saveOrderTool := NewSaveOrderTool()

    // Create a new instance of an agent
    agent, err := syndicate.NewAgent().
        SetClient(client).
        SetName("HelloAgent").
        SetConfigPrompt("<YOUR_PROMPT>").
        SetMemory(memoryAgentOne).
        SetModel(openai.GPT4).
        EquipTool(saveOrderTool). // Equip the tool to the agent 🧰
        Build()
    if err != nil {
        fmt.Printf("Error creating agent: %v\n", err)
    }

    // Process a sample input with the agent 🧠
    response, err := agent.Process(context.Background(), "Jhon Doe", "What is on the menu?")
    if err != nil {
        fmt.Printf("Error processing input: %v\n", err)
    }

    fmt.Println("\nAgent Response:")
    fmt.Println(response)
}
```

### Key Points 💡

- **`GetDefinition`** returns the definition of the tool, including its name, description, and parameters that the LLM should send when it calls the tool. 📝
- **`Execute`** processes the arguments passed by the LLM, allowing you to perform actions like storing the order in a database or making API calls. 🔄

In the `main` function, we create an agent, equip it with the `SaveOrderTool`, and process a sample input. The LLM will be able to call the tool and execute it with the provided arguments, and you can customize what happens inside the `Execute` method. 🚀

By simply implementing the `Tool` interface and adding the tool to the agent, you can process calls to the tool and do whatever you want with the information the LLM sends you. 🔧🤖 `Syndicate` internally handles detecting when the LLM uses a tool and uses the corresponding `Handler` to process it. 🛠️✨

</details>

<details>
  <summary><b>Memory Management</b></summary>

Implement long-term memory for the agent using the `Memory` interface. This allows the agent to retain context across conversations and improve its responses over time.

```go
package main


/* Import necessary packages */

type ChatMemory struct {

  /* These fields are important for the memory system */

  Name string // Name of the sender
  Role string // Role of the sender (user or assistant)
  Content string // Content of the message
  ToolCallID string // ID of the tool call (if applicable)
  ToolCalls datatype.JSON // Tool calls made during the conversation

  /* Add any other fields you need for your memory system */

}

type MyMemory struct {
  /* Define your necessary dependencies here */
}

func NewMyMemory(/* Dependencies */) syndicate.Memory {
  return &MyMemory{
    /* Initialize your dependencies here */
  }
}

func (m *MyMemory) Get() []syndicate.Message {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	ctx := context.Background()

	/* Implement the logic to retrieve messages from your memory system */

	return messages
}

func func (m *MyMemory) Add(message syndicate.Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	ctx := context.Background()

  /* Implement the logic to add messages to your memory system */

}
```

</details>

<details>
  <summary><b>Config Prompt Builder</b></summary>

The Config Prompt Builder helps create structured agent configuration prompts using a fluent API:

```go
configPrompt := syndicate.NewPromptBuilder().
  CreateSection("Introduction").
  AddText("Introduction", "You are a customer service agent.").
  CreateSection("Capabilities").
  AddListItem("Capabilities", "Answer product questions.").
  AddListItem("Capabilities", "Handle order inquiries.").
  Build()
```

</details>

## 📦 Dependencies

- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) - Apache License 2.0
- [cohesion-org/deepseek-go](https://github.com/cohesion-org/deepseek-go) - MIT License

## 🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on [GitHub](https://github.com/Dieg0Code/syndicate-go).

## 📜 License

This project is licensed under Apache License 2.0 - See the LICENSE file for details.
