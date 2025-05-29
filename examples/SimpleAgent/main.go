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

	// Create the agent using functional options instead of builder pattern.
	agent, err := syndicate.NewAgent(
		syndicate.WithClient(client),
		syndicate.WithName("HaikuAgent"),
		syndicate.WithSystemPrompt(systemPrompt),
		syndicate.WithMemory(memory),
		syndicate.WithModel(openai.GPT4),
	)
	if err != nil {
		log.Fatalf("Error creating agent: %v", err)
	}

	// Process a sample input with the agent using Chat method instead of Process.
	ctx := context.Background()
	response, err := agent.Chat(ctx,
		syndicate.WithUserName("John Doe"),
		syndicate.WithInput("What is the weather like today?"),
	)
	if err != nil {
		log.Fatalf("Error processing request: %v", err)
	}

	fmt.Println("\nAgent Response:")
	fmt.Println(response)
}
