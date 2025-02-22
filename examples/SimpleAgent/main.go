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
