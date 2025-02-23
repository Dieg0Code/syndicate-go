# Syndicate Quick Guide 🌟

The main idea behind `Syndicate` is to simplify the process of building AI agents 🤖 that integrate multiple AI services, such as OpenAI, DeepSeek, and more. `Syndicate` provides a unified interface to interact with these services and offers tools to eliminate repetitive code and improve development efficiency.

## A Use Case for a Customer Service Agent 💬

Let's suppose we have a food delivery service 🍕🚀. We want to automate customer support by integrating an OpenAI LLM into our system, enabling it to chat with customers via WhatsApp 📱, inform them about the available menu, and take or cancel orders. Creating such a system involves several key concepts:

- **Agents 🤖**: An agent is an entity that interacts with users. Its core engine is an LLM responsible for generating responses based on a given context.
- **RAG 🔍**: RAG, short for "Retrieval-Augmented Generation," is a technique that allows the LLM to access additional information that wasn’t included in its training data. While an LLM knows almost everything, there is private information it cannot access, such as our menu prices.
- **Memory 🧠**: Memory is a component that enables agents to remember information from past conversations. For example, if a customer orders a pepperoni pizza 🍕, the agent should remember this and avoid asking the same question again.
- **Tool Calls 🛠️**: The agent must be able to store customer orders in a database using predefined data schemas. This is achieved through tool calls defined in [JSON Schema](https://json-schema.org/).
- **System Prompt 🎭**: The system prompt is a special instruction given to the LLM to guide its behavior. Through this prompt, we can configure its personality, language style, additional knowledge, and more.
- **Semantic Search 🔎**: Semantic search allows retrieving information from a database based on meaning rather than exact keyword matches. For an LLM, this is crucial to providing information based on user intent rather than specific words.

Implementing such a system in practice is complex and requires multiple services, logic layers, and configurations that can quickly turn into spaghetti code 🍝. `Syndicate` aims to simplify this process, allowing developers to focus on business logic rather than AI service integration.

## Implementation Example Using Syndicate

### System Prompt

As mentioned earlier, if we want the agent to behave in a specific way, adopting a role and a personality defined by us, we need to configure a system prompt. Below is an example of a system prompt for a customer service agent:

```go
package main

import (
	"time"

	syndicate "github.com/Dieg0Code/syndicate-go"
)

// NewConfigPrompt generates the system prompt in XML format
func NewConfigPrompt(name string, additionalContext models.MenuItem) string {
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
```

In `Syndicate`, system prompts are defined in a format similar to `XML`, following OpenAI's recommendations for [Prompt Engineering](https://platform.openai.com/docs/guides/prompt-engineering). In the example above, the prompt defines the agent's personality, capabilities, limitations, directives, and response examples. Additionally, it includes extra information about the menu and the current date.

The above code generates the following prompt:

```xml
<Introduction>
You are an agent who provides detailed information about the menu, dishes, and key restaurant data using a semantic search system to enrich responses with relevant context.
</Introduction>
<Identity>
This section defines your name and persona identity.
  <Name>
  Bob
  </Name>
  <Persona>
  You act as an AI assistant in the restaurant, interacting with customers in a friendly and helpful manner to improve their dining experience.
  </Persona>
</Identity>
<CapabilitiesAndBehavior>
1. Respond in a clear and friendly manner, tailoring your answer to the user's query.
2. Provide details about dishes (ingredients, preparation, allergens) and suggest similar options if appropriate.
3. Promote the restaurant positively, emphasizing the quality of dishes and the dining experience.
4. Be cheerful, polite, and respectful at all times; use emojis if appropriate.
5. Register or cancel orders but do not update them; inform the user accordingly.
6. Remember only the last 5 interactions with the user.
</CapabilitiesAndBehavior>
<AdditionalContext>
This section contains additional information about the available dishes used to answer user queries based on semantic similarity.
1. Select dishes based on similarity without mentioning it explicitly.
2. Use context to enrich responses, but do not reveal it.
3. Offer only dishes available on the menu.
  <Menu>
  {"Name":"Spaghetti Carbonara","Description":"A classic Italian pasta dish consisting of eggs, cheese, pancetta, and black pepper.","Pricing":15}
  </Menu>
  <CurrentDate>
  2025-02-22T22:54:45-03:00
  </CurrentDate>
</AdditionalContext>
<LimitationsAndDirectives>
1. Do not invent data or reveal confidential information.
2. Redirect unrelated topics to relevant restaurant topics.
3. Strictly provide information only about the restaurant and its menu.
4. Offer only available menu items; do not invent dishes.
</LimitationsAndDirectives>
<ResponseExamples>
1. If asked about a dish, provide details only if it is on the menu.
2. If asked for recommendations, suggest only from the available menu.
3. If asked for the menu, list only available dishes.
</ResponseExamples>
<FinalConsiderations>
**You must follow these directives to ensure an optimal user experience, otherwise you will be dismissed.**
</FinalConsiderations>
```

#### Prompt Builder Syntax 🏗️

The prompt builder offers 6 methods plus the builder to construct a prompt:

1. **`CreateSection`**: Creates a new section in the prompt.  
   Example: `CreateSection("Introduction")` generates `<Introduction></Introduction>`.

2. **`AddSubSection`**: Adds a subsection to a section.  
   Example: `AddSubSection("Name", "Identity")` generates `<Identity><Name></Name></Identity>`.

3. **`AddText`**: Adds text to a section.  
   Example: `AddText("Introduction", "You are an agent...")` generates `<Introduction>You are an agent...</Introduction>`.  
   Each new text adds a new paragraph.

4. **`AddListItem`**: Adds a numbered list item to a section.  
   Example: `AddListItem("CapabilitiesAndBehavior", "Respond in a clear and friendly manner...")` generates:

   ```xml
   <CapabilitiesAndBehavior>
   1. Respond in a clear and friendly manner...
   </CapabilitiesAndBehavior>
   ```

   Each new item adds a new number to the list.

5. **`AddTextF`**: Adds formatted text to a section. Useful for passing variables or dynamic data.  
   Example: `AddTextF("CurrentDate", time.Now().Format(time.RFC3339))` generates:

   ```xml
   <CurrentDate>2025-02-22T22:54:45-03:00</CurrentDate>
   ```

6. **`AddListItemF`**: Adds a formatted list item to a section. Useful for passing variables or dynamic data.  
   Example: `AddListItemF("Menu", additionalContext)` generates:

   ```xml
   <Menu>{"Name":"Spaghetti Carbonara","Description":"A classic Italian pasta dish consisting of eggs, cheese, pancetta, and black pepper.","Pricing":15}</Menu>
   ```

7. **`Build`**: Constructs the prompt in XML-like format (as a string).

---

### Agents 🤖

When creating an agent, several aspects must be considered:

- The LLM model to use.
- The AI service provider.
- The system prompt.
- The agent's memory.
- The tools the agent can call.
- **Temperature**, to configure the agent's creativity.

`Syndicate` abstracts all these concepts, allowing agents to be created easily. Below is an example of a customer service agent:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    syndicate "github.com/Dieg0Code/syndicate-go"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    // Initialize the OpenAI client with your API key 🔑
    client := syndicate.NewOpenAIClient("YOUR_OPENAI_API_KEY")

    // Create a simple memory instance 🧠
    memory := syndicate.NewSimpleMemory()

    // Create a new agent 🏗️
    customerServiceAgent, err := syndicate.NewAgent().
        SetClient(client).
        SetName("HelloAgent").
        SetConfigPrompt("<YOUR_PROMPT>").
        SetMemory(memory).
        SetModel(openai.GPT4).
        Build()
    if err != nil {
        fmt.Printf("❌ Error creating agent: %v\n", err)
    }

    // Process a sample input with the agent 🗣️
    response, err := customerServiceAgent.Process(context.Background(), "Jhon Doe", "What is on the menu?")
    if err != nil {
        fmt.Printf("❌ Error processing input: %v\n", err)
    }

    fmt.Println("\n🤖 Agent Response:")
    fmt.Println(response)
}
```

In this basic example, we configure the provider, the model to use, the system prompt, and the agent's memory. Then, we process a sample input and display the agent’s response.

Currently, `Syndicate` only supports models from **OpenAI** and **Deepseek R1**, but future updates will include support for more AI providers. 🚀

### Tool Calls ⚒️🤖

`Tool Calls`, previously known as `Function Calling`, is a method that allows you to define data schemas and pass them to an LLM. This enables the model to:

1. Retrieve information from an external source.
2. Perform actions within the system.

### Structure of a Tool 🏗️

A `Tool` consists of **four main parts**:

- **`Name`** 🏷️: The name of the tool. Example: `"SaveOrder"`.
- **`Description`** 📝: A brief explanation of what the tool does and when it should be used.
  - Example:
    ```plaintext
    Retrieves the user's order. The user must provide the requested menu items,
    delivery address, name, phone number, and payment method.
    The payment method can only be cash or bank transfer.
    ```
- **`Parameters`** 📜: The `jsonschema` that defines the tool’s structure.
- **`Strict`** ✅: A boolean that determines whether the `jsonschema` must be followed exactly (`true`) or if additional fields are allowed (`false`).
  - Typically, this should **always** be `true`.

---

### Understanding JSON Schema 📋

Before diving into the JSON schema, let’s first define the **Go struct** that represents the required data. This way, we can understand how the structure is defined before seeing its JSON equivalent.

---

### Defining the Order System 🛒📦

In our example, we need our agent to **schedule customer orders**. So, what information do we need for an order? 🤔

- **Customer Name**: A `string`.
- **Delivery Address**: A `string`.
- **Phone Number**: A `string`.
- **Payment Method**: A `string`. (We only accept **cash** or **bank transfer**).
- **Menu Items Ordered**: An `array` of objects, where each object contains:
  - **Item Name** 🏷️: A `string`.
  - **Quantity** 🔢: An `integer`.
  - **Price** 💰: An `integer`.

Now that we have a clear structure, let’s define the corresponding **Go struct** that represents it! 🚀

### Defining the Order Structure in Go ⚡🛠️

Alright! Now, let's start by defining the **basic structure** for each menu item:

```go
// Defining the structure of a menu item 🍽️
type MenuItem struct {
    ItemName string `json:"item_name"`  // Name of the menu item
    Quantity int    `json:"quantity"`   // Quantity ordered
    Price    int    `json:"price"`      // Price per unit
}

// Now, we define the overall order structure 📦
type UserOrder struct {
    UserName        string     `json:"user_name"`        // Name of the customer
    DeliveryAddress string     `json:"delivery_address"` // Delivery location
    PhoneNumber     string     `json:"phone_number"`     // Contact number
    PaymentMethod   string     `json:"payment_method"`   // Payment method (cash/transfer)
    Items           []MenuItem `json:"items"`            // List of ordered items
}
```

---

### Adding Context for JSON Schema 🧠📜

Now, **this structure alone isn’t enough** to generate a proper `jsonschema`.

Why? 🤔

Because we’re working with an **LLM**, and it needs **explicit context** about each field. Simply having the field name and type **isn’t enough**.

So, let’s enhance our structure by adding **descriptions** and **constraints** to make it more useful for the LLM:

```go
// Defining the schema for an individual menu item 🍕
type MenuItemSchema struct {
	ItemName string `json:"item_name" description:"Menu item name" required:"true"`
	Quantity int    `json:"quantity" description:"Quantity ordered by the user" required:"true"`
	Price    int    `json:"price" description:"Menu item price" required:"true"`
}

// Defining the schema for the entire order 📦
type UserOrderFunctionSchema struct {
	MenuItems       []MenuItemSchema `json:"menu_items" description:"List of ordered menu items" required:"true"`
	DeliveryAddress string           `json:"delivery_address" description:"Order delivery address" required:"true"`
	UserName        string           `json:"user_name" description:"User's name placing the order" required:"true"`
	PhoneNumber     string           `json:"phone_number" description:"User's phone number" required:"true"`
	PaymentMethod   string           `json:"payment_method" description:"Payment method (cash or transfer only)" required:"true" enum:"cash,transfer"`
}
```

---

### What's Different? 🤔

🔹 **Added `description` tags**: Now, each field has a brief explanation, so the LLM understands its purpose.  
🔹 **Marked fields as `required`**: This ensures that the LLM knows which fields **must** be included.  
🔹 **Added `enum` for `PaymentMethod`**: This restricts the allowed values to **only** `"cash"` or `"transfer"`.

With these improvements, our `jsonschema` will be much **clearer** and **more useful** for AI-based processing! 🚀

### 🏷️ JSONSchema Tags

`Syndicate` supports the following tags to generate a JSON schema:

- **`description`**: A description of the field. 📝
- **`required`**: A boolean that indicates whether the field is mandatory or optional. This helps the LLM understand if the field **must** be included or if it can be left empty. ✅
- **`enum`**: An array of strings defining the possible values a field can take. For example, in the case of `PaymentMethod`, it can **only** be `"cash"` or `"transfer"`. 💰

---

### 🛠️ Generating the JSON Schema

Now that we have both:  
1️⃣ The **Go structure** that defines our order.  
2️⃣ The **Go structure** that defines the corresponding JSON schema.

Let's generate the JSON schema! 🎉

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

### 🛠️ Tool Call Example

This schema tells the LLM that when calling the `SaveOrder` tool, it must pass an object containing the fields `menu_items`, `delivery_address`, `user_name`, `phone_number`, and `payment_method`, and that these fields are required. 📜💡

For example, when a user completes an order, the LLM could respond with the following:

```json
{
  "menu_items": [
    {
      "item_name": "Spaghetti Carbonara",
      "price": 15,
      "quantity": 1
    },
    {
      "item_name": "Pizza Margherita",
      "price": 12,
      "quantity": 2
    }
  ],
  "delivery_address": "123 Main St",
  "user_name": "Jhon Doe",
  "phone_number": "123-456-7890",
  "payment_method": "cash"
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

### Memory 🧠

Memory is a component that allows agents to remember information from past conversations. For example, if a customer orders a pepperoni pizza, the agent should remember that the customer already ordered a pepperoni pizza, so it doesn't ask again. 🍕

To give memory to an agent, you have two options:

- **SimpleMemory**: A simple memory that stores information in an in-memory map. It’s useful for quick prototypes and testing. 🧩
- **Implement the Memory interface**: You can implement the `Memory` interface to use a database or an external service as memory. 💾

Here’s an example of how to use `SimpleMemory`:

```go
package main

import (
    "context"
    "fmt"

    syndicate "github.com/Dieg0Code/syndicate-go"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    // Initialize the OpenAI client with your API key.
    client := syndicate.NewOpenAIClient("YOUR_OPENAI_API_KEY")

    // Create a new instance of simple memory.
    memory := syndicate.NewSimpleMemory()

    // Create a new instance of an agent.
    agent, err := syndicate.NewAgent().
        SetClient(client).
        SetName("HelloAgent").
        SetConfigPrompt("<YOUR_PROMPT>").
        SetMemory(memory).
        SetModel(openai.GPT4).
        Build()

    if err != nil {
        fmt.Printf("Error creating agent: %v\n", err)
    }

    // Process an example input with the agent.
    response, err := agent.Process(context.Background(), "Jhon Doe", "What is on the menu?")
    if err != nil {
        fmt.Printf("Error processing input: %v\n", err)
    }

    fmt.Println("\nAgent Response:")
    fmt.Println(response)
}
```

Now your agent has a memory 🧠 and can handle repetitive tasks more efficiently!

`SimpleMemory` is basically a slice of objects with `role` and `content` stored in memory. 🧠 However, in production, you generally won’t want to use this. Instead, you can implement the `Memory` interface and use a database or an external service for better storage options. 💾

```go
type Memory interface {
	// Add appends a complete ChatCompletionMessage to the memory.
	Add(message Message)
	// Get returns a copy of all stored chat messages.
	Get() []Message
	// Clear removes all stored chat messages from memory.
	Clear()
}
```

```go
type Message struct {
	Role      string     // One of RoleSystem, RoleUser, RoleAssistant, or RoleTool.
	Content   string     // The textual content of the message.
	Name      string     // Optional identifier for the sender.
	ToolCalls []ToolCall // Optional tool calls made by the assistant.
	ToolID    string     // For tool responses, references the original tool call.
}
```

### Example of Implementing the `Memory` Interface 🔧

```go
package main

import (
    "context"
    "fmt"

    syndicate "github.com/Dieg0Code/syndicate-go"
    openai "github.com/sashabaranov/go-openai"
)

type MyMemory struct {
    db *gorm.DB
    mutex sync.RWMutex
}

func NewMyMemory(db *gorm.DB) syndicate.Memory {
    return &MyMemory{
        db: db,
    }
}

func (m *MyMemory) Add(message syndicate.Message) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    // Save the message to the database
    // m.db.Save(&message)
}

func (m *MyMemory) Get() []syndicate.Message {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    // Get all the messages from the database
    // Or, consider getting the last 10 messages of today
    // Actually, it's recommended because the LLM can't handle too many messages at once.
    // var messages []syndicate.Message
    // m.db.Find(&messages)
    // return messages
    return nil
}

func (m *MyMemory) Clear() {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    // Clear the database
    // m.db.Delete(&syndicate.Message{})
    // Optional in this case
}

func main() {
    // Initialize the OpenAI client with your API key.
    client := syndicate.NewOpenAIClient("YOUR_OPENAI_API_KEY")

    // Create a new instance of our custom memory.
    db, err := gorm.Open(sqlite.Open("memory.db"), &gorm.Config{})
    if err != nil {
        fmt.Printf("Error opening database: %v\n", err)
    }
    myMemory := NewMyMemory(db)

    // Create a new agent instance.
    agent, err := syndicate.NewAgent().
        SetClient(client).
        SetName("HelloAgent").
        SetConfigPrompt("<YOUR_PROMPT>").
        SetMemory(myMemory).
        SetModel(openai.GPT4).
        Build()
    if err != nil {
        fmt.Printf("Error creating agent: %v\n", err)
    }

    // Process an example input with the agent.
    response, err := agent.Process(context.Background(), "Jhon Doe", "What is on the menu?")
    if err != nil {
        fmt.Printf("Error processing input: %v\n", err)
    }

    fmt.Println("\nAgent Response:")
    fmt.Println(response)
}
```

Just like with the `Tool` interface, it's up to you how you want to manage the information in the database. 💡 You can store all messages, just the last 10, or only those from today, etc. `Syndicate` only provides the interface, and you decide how to implement it based on your needs. 🔧

### Semantic Search 🌐

Semantic search is a technique that allows you to search for information in a database semantically, meaning you’re looking for relevant information instead of exact matches. For an LLM (Large Language Model), this is crucial to provide information based on the user's intent, not just keywords. 🤖

To implement this, you need to keep the following concepts in mind:

- **Embeddings**: Embeddings are vector representations of words, phrases, or documents. These vectors are generated by a language model and capture the semantic meaning of the words. In the case of `Syndicate`, we currently support `OpenAI` models:

```go
    AdaEmbeddingV2  EmbeddingModel = "text-embedding-ada-002"
    SmallEmbedding3 EmbeddingModel = "text-embedding-3-small"
    LargeEmbedding3 EmbeddingModel = "text-embedding-3-large"
```

- **Vector Database**: A vector database is one that supports storing and searching vectors. For example, the `PgVector` extension for `PostgreSQL` allows you to store vectors in a `PostgreSQL` database and perform semantic searches.

- **Semantic Search**: Semantic search is the process of searching for information in a vector database using a query vector. The idea is to search for the vectors closest to the query vector and return the information associated with those vectors. 🧠

When you combine these three concepts, you can perform **Retrieval Augmented Generation** (RAG), which is a technique that combines text generation with semantic search to provide relevant information based on user intent. 🔄

The RAG flow is as follows:

1. You generate embeddings for the information you want (e.g., embeddings of each menu item).

An example using `GORM` and `LargeEmbedding3`:

```go
package models

import (
    "github.com/pgvector/pgvector-go"
    "gorm.io/gorm"
)

type Menu struct {
    gorm.Model
    ItemName     string          `gorm:"type:varchar(100);not null"`
    Description  string          `gorm:"type:varchar(255);not null"`
    Price        int             `gorm:"not null"`
    Likes        int             `gorm:"default:0"`
    Embedding    pgvector.Vector `gorm:"type:vector(3072)"` // 3072 is the dimension of LargeEmbedding3
}
```

2. You store the embeddings in the database. 💾
3. You generate a query vector from the user’s message. 📝
4. You perform a semantic search in the vector database based on the embedding of the user’s message.
5. You return the information associated with the closest vectors to the query vector.
6. You pass this information to the LLM to generate a response. ✨

And just like that, your LLM can access domain-specific information and understand what the user wants, providing relevant information. 🎯

Here's a theoretical example of what the code would look like:

- First, we define the Go struct that represents a menu item:

```go
package models

import (
    "github.com/pgvector/pgvector-go"
    "gorm.io/gorm"
)

type Menu struct {
    gorm.Model
    ItemName     string          `gorm:"type:varchar(100);not null"`
    Description  string          `gorm:"type:varchar(255);not null"`
    Price        int             `gorm:"not null"`
    Likes        int             `gorm:"default:0"`
    Embedding    pgvector.Vector `gorm:"type:vector(3072)"` // 3072 is the dimension of LargeEmbedding3
}
```

- Then we generate embeddings for each menu item:

```go
package main

import (
    "context"
    "fmt"
    "log"

    syndicate "github.com/Dieg0Code/syndicate-go"
)

func main() {
    // Initialize the OpenAI client with your API key.
    client := syndicate.NewOpenAIClient("YOUR_OPENAI_API_KEY")

    // Build the Embedder using the builder.
    embedder, err := syndicate.NewEmbedderBuilder().
        SetClient(client).
        // Optionally, set a different model by uncommenting the next line:
        // SetModel(openai.SmallEmbedding3).
        Build()
    if err != nil {
        fmt.Printf("Error creating embedder: %v\n", err)
    }

    // Define a menu item example
    menuItem := dto.MenuItem{
        ItemName: "Spaghetti Carbonara",
        Description: "A classic Italian pasta dish consisting of eggs, cheese, pancetta, and black pepper.",
        Price: 15,
    }

    // Convert the menu item to JSON
    itemJSON, err := json.Marshal(menuItem)
    if err != nil {
        fmt.Printf("Error marshaling menu item: %v\n", err)
    }

    // Generate the embedding of the menu item from the JSON
    embedding, err := embedder.GenerateEmbedding(context.Background(), itemJSON)
    if err != nil {
        fmt.Printf("Error generating embedding: %v\n", err)
    }

    // Save the embedding to the database
    // db.Save(&models.Menu{
    //     ItemName: menuItem.ItemName,
    //     Description: menuItem.Description,
    //     Price: menuItem.Price,
    //     Embedding: embedding, // Usually the embeddings returned by the API are of type []float32
    // })                       // But this depends on the database implementation you're using.
}
```

- Next, we define our method for performing semantic search, for example, using `GORM` and `PostgreSQL`:

```go
// SemanticSearchMenu implements MenuRepository.
func (m *MenuRepositoryImpl) SemanticSearchMenu(queryEmbedding []float32, similarityThreshold float32, matchCount int, restaurantID uint) ([]dto.MenuSearchResponse, error) {
    var results []dto.MenuSearchResponse

    vectorEmbedding := pgvector.NewVector(queryEmbedding)

    result := m.db.Model(&models.Menu{}).
        Select(`
            id,
            item_name,
            price,
            description,
            likes,
            embedding <#> ? AS similarity
        `, vectorEmbedding).
        Where("restaurant_id = ?", restaurantID).
        Where("embedding <#> ? < ?", vectorEmbedding, similarityThreshold).
        Order("similarity").
        Limit(matchCount).
        Scan(&results)

    if result.Error != nil {
        logrus.WithError(result.Error).Error("Error fetching menu")
        return nil, fmt.Errorf("error fetching menu")
    }
    return results, nil
}
```

This part is key:

```go
    vectorEmbedding := pgvector.NewVector(queryEmbedding)

    result := m.db.Model(&models.Menu{}).
        Select(`
            id,
            item_name,
            price,
            description,
            likes,
            embedding <#> ? AS similarity
        `, vectorEmbedding).
        Where("restaurant_id = ?", restaurantID).
        Where("embedding <#> ? < ?", vectorEmbedding, similarityThreshold).
        Order("similarity").
        Limit(matchCount).
        Scan(&results)
```

`PgVector` has several operators; in this case, we use `<#>`. I recommend taking a look at the [documentation](https://github.com/pgvector/pgvector) to see which operators you can use and how to use them.

In the previous function, we’re basically doing the following:

- We receive a query vector, a similarity threshold, a number of results, and a restaurant ID.
- We convert the query vector, which comes as `[]float32`, into a `pgvector.Vector`.
- We perform a semantic search in the vector database based on the query vector.
- We order the results by similarity.
- We limit the number of results.
- We return the results.

This brings back semantically relevant information based on the user's message. After that, we need to pass this information to the LLM to generate a response. ✨

#### Example of Semantic Search Usage

```go
package main

import (
    "context"
    "fmt"

    syndicate "github.com/Dieg0Code/syndicate-go"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    // User's message
    message := "I want a pizza"

    // Create a client for OpenAI
    client := syndicate.NewOpenAIClient("YOUR_OPEN_AI_API_KEY")

    // Create an embedder
    embedder, err := syndicate.NewEmbedderBuilder().
        SetClient(client).
        SetModel(openai.LargeEmbedding3).
        Build()
    if err != nil {
        fmt.Printf("Error creating embedder: %v\n", err)
    }

    // Generate the embedding for the user's message
    embedding, err := embedder.GenerateEmbedding(context.Background(), message)
    if err != nil {
        fmt.Printf("Error generating embedding: %v\n", err)
    }

    // Retrieve relevant information based on the user's message
    similarityThreshold := 0.8 // The higher the threshold, the more precise the search
    matchCount := 5 // Number of results to fetch
    restaurantID := 1 // Restaurant ID, depending on your business logic
    results, err := menuRepository.SemanticSearchMenu(embedding, similarityThreshold, matchCount, restaurantID)
    if err != nil {
        fmt.Printf("Error fetching menu: %v\n", err)
    }

    // Pass the information to the LLM to generate a response
    systemPrompt := syndicate.NewPromptBuilder().
        CreateSection("Introduction").
        AddText("Introduction", "Answer the user's request for a pizza.").
        CreateSection("Personality").
        AddText("Personality", "This section describes how you should respond to the user.").
        AddListItem("Personality", "Be friendly and helpful.").
        AddListItem("Personality", "Offer to the user the pizza in the menu.").
        CreateSection("Menu").
        AddTextF("Menu", results). // Pass the semantic search results
        Build()

    memory := syndicate.NewSimpleMemory()

    // Create an agent
    agent, err := syndicate.NewAgent().
        SetClient(client).
        SetName("PizzaAgent").
        SetConfigPrompt(systemPrompt).
        SetMemory(memory).
        SetModel(openai.GPT4).
        Build()
    if err != nil {
        fmt.Printf("Error creating agent: %v\n", err)
    }

    // Process an example input with the agent
    agentResponse, err := agent.Process(context.Background(), "Client", message)
    if err != nil {
        fmt.Printf("Error processing input: %v\n", err)
    }

    fmt.Println("\nAgent Response:")
    fmt.Println(agentResponse)
}
```

The code above is educational but illustrates the flow of how to use semantic search in `Syndicate`. Basically, it goes like this:

1. Generate embeddings for the information you want.
2. Store the embeddings in the database.
3. Generate a query vector from the user's message.
4. Perform a semantic search in the vector database based on the user's message embedding.
5. Return the information associated with the vectors closest to the query vector.
6. Pass that information to the LLM to generate a response.
7. Show prices and available pizzas from the menu to the user.
8. 🍕🍕🍕

## Syndicate 🚀

The SDK is called `Syndicate` because, in addition to all these amazing and magical abstractions I’ve shown you earlier, it provides a tool to create processing flows between multiple agents, each with their own unique configuration, tools, and individual memory. 🤖✨

Here’s a quick example of how to use it:

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
	// Initialize the OpenAI client using your API key.
	client := syndicate.NewOpenAIClient("YOUR_API_KEY")

	// Create simple memory instances for each agent.
	memoryAgentOne := syndicate.NewSimpleMemory()
	memoryAgentTwo := syndicate.NewSimpleMemory()

	// Build the first agent (HelloAgent).
	agentOne, err := syndicate.NewAgent().
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
	agentTwo, err := syndicate.NewAgent().
		SetClient(client).
		SetName("FinalAgent").
		SetConfigPrompt("You are an agent that provides a final summary based on the conversation.").
		SetMemory(memoryAgentTwo).
		SetModel(openai.GPT4).
		Build()
	if err != nil {
		log.Fatalf("Error building FinalAgent: %v", err)
	}

	// Create a syndicate, recruit both agents, and define the execution pipeline.
	syndicateSystem := syndicate.NewSyndicate().
		RecruitAgent(agentOne).
		RecruitAgent(agentTwo).
		// Define the processing pipeline: first HelloAgent, then FinalAgent.
		DefinePipeline([]string{"HelloAgent", "FinalAgent"}).
		Build()

	// User name for the conversation.
	userName := "User"

	// Provide an input and process the pipeline.
	input := "Please greet the user and provide a summary."
	response, err := syndicateSystem.ExecutePipeline(context.Background(), userName, input)
	if err != nil {
		log.Fatalf("Error processing pipeline: %v", err)
	}

	fmt.Println("Final Syndicate Response:")
	fmt.Println(response)
}
```

In the code above, we create two agents: `HelloAgent` and `FinalAgent`, each with its own configuration, tools, and memory. Then, we create a `Syndicate`:

```go
syndicateSystem := syndicate.NewSyndicate().
    RecruitAgent(agentOne).
    RecruitAgent(agentTwo).
    DefinePipeline([]string{"HelloAgent", "FinalAgent"}).
    Build()
```

We recruit the agents with:

```go
RecruitAgent(agentOne).
RecruitAgent(agentTwo).
```

And define the execution pipeline with:

```go
DefinePipeline([]string{"HelloAgent", "FinalAgent"}).
```

This means the user's query will first be processed by `HelloAgent` and then by `FinalAgent`. The flows, agents, tools, and logic you want to apply depend on your creativity and business logic. 💡🔧

Chaining agents can be super useful to analyze and extract information from different sources, process it, and return a more complete and relevant response for the user. Plus, you can use tools and memory within each agent to make the processing flow richer and more personalized. 🎯

Basically, the response from one agent becomes the input for the next agent, and so on, until the final agent in the chain returns the ultimate response. 🔄

I also recommend passing the original user query in each agent’s system prompt so that each agent doesn’t lose context during the conversation. 🧠

That’s all for now! I hope you enjoyed it and that you’re excited to try `Syndicate` in your next LLM project. 🚀

Big hugs and kisses 🤗🤗🤗
