package main

import (
	"context"
	"fmt"
	"log"

	gokamy "github.com/Dieg0code/gokamy-ai"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Create an OpenAI client using your API key.
	apiKey := "YOUR_API_KEY" // Replace with your OpenAI API key.
	client := openai.NewClient(apiKey)

	// Build the Embedder using the builder.
	embedder, err := gokamy.NewEmbedderBuilder().
		SetClient(client).
		// Optionally, set a different model by uncommenting the next line:
		// SetModel(openai.SmallEmbedding3).
		Build()
	if err != nil {
		log.Fatalf("error building embedder: %v", err)
	}

	// Text to generate the embedding from.
	data := "Example text to generate embedding"

	// Generate the embedding.
	embedding, err := embedder.GenerateEmbedding(context.Background(), data)
	if err != nil {
		log.Fatalf("error generating embedding: %v", err)
	}

	// Print the generated embedding.
	fmt.Printf("Generated embedding: %v\n", embedding)
}
