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

	// Build the first agent (HelloAgent) using functional options.
	agentOne, err := syndicate.NewAgent(
		syndicate.WithClient(client),
		syndicate.WithName("HelloAgent"),
		syndicate.WithSystemPrompt("You are an agent that warmly greets users and encourages further interaction."),
		syndicate.WithMemory(memoryAgentOne),
		syndicate.WithModel(openai.GPT4),
	)
	if err != nil {
		log.Fatalf("Error creating HelloAgent: %v", err)
	}

	// Build the second agent (FinalAgent) using functional options.
	agentTwo, err := syndicate.NewAgent(
		syndicate.WithClient(client),
		syndicate.WithName("FinalAgent"),
		syndicate.WithSystemPrompt("You are an agent that provides a final summary based on the conversation."),
		syndicate.WithMemory(memoryAgentTwo),
		syndicate.WithModel(openai.GPT4),
	)
	if err != nil {
		log.Fatalf("Error creating FinalAgent: %v", err)
	}

	// Create a syndicate using functional options and define the execution pipeline.
	syndicateSystem, err := syndicate.NewSyndicate(
		syndicate.WithAgents(agentOne, agentTwo),
		// Define the processing pipeline: first HelloAgent, then FinalAgent.
		syndicate.WithPipeline("HelloAgent", "FinalAgent"),
	)
	if err != nil {
		log.Fatalf("Error creating syndicate: %v", err)
	}

	// User name for the conversation.
	userName := "User"

	// Provide an input and process the pipeline using functional options.
	input := "Please greet the user and provide a summary."
	response, err := syndicateSystem.ExecutePipeline(context.Background(),
		syndicate.WithPipelineUserName(userName),
		syndicate.WithPipelineInput(input),
	)
	if err != nil {
		log.Fatalf("Error processing pipeline: %v", err)
	}

	fmt.Println("Final Syndicate Response:")
	fmt.Println(response)
}
