package main

import (
	"context"
	"fmt"
	"log"

	gomaky "github.com/Dieg0code/gokamy-ai"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client using your API key.
	client := openai.NewClient("YOUR_API_KEY")

	// Create simple memory instances for each agent.
	memoryAgentOne := gomaky.NewSimpleMemory()
	memoryAgentTwo := gomaky.NewSimpleMemory()

	// Build the first agent (HelloAgent).
	agentOne, err := gomaky.NewAgentBuilder().
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
	agentTwo, err := gomaky.NewAgentBuilder().
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
	orchestrator := gomaky.NewOrchestratorBuilder().
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
