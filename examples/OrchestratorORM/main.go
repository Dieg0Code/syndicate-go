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

	// Create a custom memory implementation using functional options
	customMemory, err := syndicate.NewMemory(
		syndicate.WithAddHandler(func(msg syndicate.Message) {
			fmt.Println("CustomMemory - Adding message:", msg)
		}),
		syndicate.WithGetHandler(func() []syndicate.Message {
			fmt.Println("CustomMemory - Getting messages")
			// In a real implementation, you would return your stored messages here
			return []syndicate.Message{}
		}),
	)
	if err != nil {
		log.Fatalf("Error creating custom memory: %v", err)
	}

	// Create an agent using functional options
	agent, err := syndicate.NewAgent(
		syndicate.WithClient(client),
		syndicate.WithName("CustomMemoryAgent"),
		syndicate.WithSystemPrompt("You are an agent that logs all messages to a custom memory implementation."),
		syndicate.WithMemory(customMemory),
		syndicate.WithModel(openai.GPT4),
	)
	if err != nil {
		log.Fatalf("Error creating agent: %v", err)
	}

	// Process a sample input with the agent using Chat method
	ctx := context.Background()
	response, err := agent.Chat(ctx,
		syndicate.WithUserName("User"),
		syndicate.WithInput("What is your favorite quote?"),
	)
	if err != nil {
		log.Fatalf("Error processing request: %v", err)
	}

	fmt.Println("Agent Response:")
	fmt.Println(response)

	// Display the contents of the custom memory
	fmt.Println("\nCustom Memory Contents:")
	for _, msg := range customMemory.Get() {
		fmt.Printf("Role: %s, Content: %s\n", msg.Role, msg.Content)
	}
}
