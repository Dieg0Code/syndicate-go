## Example With Persistent Memory

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    gokamy "github.com/Dieg0code/gokamy"
    openai "github.com/sashabaranov/go-openai"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Message represents the schema for storing chat messages.
type Message struct {
    ID        uint      `gorm:"primaryKey"`
    CreatedAt time.Time
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
        CreatedAt: time.Now(),
    }
    if err := m.db.Create(&msg).Error; err != nil {
        log.Printf("failed to add message: %v", err)
    }
}

// Get retrieves all stored messages ordered by creation time.
func (m *ORMMemory) Get() []openai.ChatCompletionMessage {
    var msgs []Message
    if err := m.db.Order("created_at asc").Find(&msgs).Error; err != nil {
        log.Printf("failed to retrieve messages: %v", err)
    }
    result := make([]openai.ChatCompletionMessage, len(msgs))
    for i, msg := range msgs {
        result[i] = openai.ChatCompletionMessage{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }
    return result
}

// Clear removes all messages from the persistent memory.
func (m *ORMMemory) Clear() {
    if err := m.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Message{}).Error; err != nil {
        log.Printf("failed to clear messages: %v", err)
    }
}

// NewORMMemory returns a Memory interface backed by ORMMemory.
// It auto-migrates the Message table using GORM.
func NewORMMemory(db *gorm.DB) gokamy.Memory {
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