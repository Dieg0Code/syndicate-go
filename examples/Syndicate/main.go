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
