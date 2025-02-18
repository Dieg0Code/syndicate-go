package main

import (
	"context"
	"fmt"
	"log"

	gokamy "github.com/Dieg0code/gokamy-ai"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client with your API key.
	client := openai.NewClient("YOUR_OPENAI_API_KEY")

	// Create a simple memory instance.
	memory := gokamy.NewSimpleMemory()

	// Build a structured prompt using PromptBuilder to instruct the agent to speak in haiku.
	systemPrompt := gokamy.NewPromptBuilder().
		CreateSection("Introduction").
		AddText("Introduction", "You are an agent that always responds in haiku format.").
		CreateSection("Guidelines").
		AddListItem("Guidelines", "Keep responses to a three-line haiku format (5-7-5 syllables).").
		AddListItem("Guidelines", "Be creative and concise.").
		Build()

	fmt.Println("System Prompt:")
	fmt.Println(systemPrompt)

	// Build the agent using AgentBuilder.
	agent, err := gokamy.NewAgentBuilder().
		SetClient(client).
		SetName("HaikuAgent").
		SetSystemPrompt(systemPrompt).
		SetMemory(memory).
		SetModel(openai.GPT4).
		SetMaxRecursion(2). // Prevent infinite loops in tool calls.
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
