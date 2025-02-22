package syndicate

import (
	"context"
	"errors"
	"fmt"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

// Embedder is responsible for generating embeddings using the OpenAI API.
type Embedder struct {
	client *openai.Client        // OpenAI client to perform API calls.
	model  openai.EmbeddingModel // Default embedding model to use.
}

// GenerateEmbedding generates an embedding for the provided data string.
// It accepts an optional embedding model; if provided, that model overrides the default.
// Returns a slice of float32 representing the embedding vector or an error if any.
func (e *Embedder) GenerateEmbedding(ctx context.Context, data string, model ...openai.EmbeddingModel) ([]float32, error) {
	// Validate input data.
	if data == "" {
		return nil, errors.New("input data cannot be empty")
	}

	// Ensure context is not nil.
	if ctx == nil {
		ctx = context.Background()
	}

	// Determine the embedding model to use.
	modelName := e.model
	if len(model) > 0 && model[0] != "" {
		modelName = model[0]
	}

	// Build the embedding request.
	req := openai.EmbeddingRequest{
		Input: []string{data},
		Model: modelName,
	}

	// Call the OpenAI API to generate embeddings.
	res, err := e.client.CreateEmbeddings(ctx, req)
	if err != nil {
		log.Printf("failed to create embeddings for model '%s': %v", modelName, err)
		return nil, fmt.Errorf("create embeddings error: %w", err)
	}

	// Validate that embedding data was returned.
	if len(res.Data) == 0 {
		return nil, errors.New("no embedding data returned")
	}

	// Return the generated embedding.
	return res.Data[0].Embedding, nil
}

// EmbedderBuilder provides a fluent API to configure and build an Embedder instance.
type EmbedderBuilder struct {
	client *openai.Client        // OpenAI client to be used by the Embedder.
	model  openai.EmbeddingModel // Embedding model to be used; defaults to a preset model.
}

// NewEmbedderBuilder initializes a new EmbedderBuilder with default settings.
func NewEmbedderBuilder() *EmbedderBuilder {
	return &EmbedderBuilder{
		// Default model is set to openai.LargeEmbedding3.
		model: openai.LargeEmbedding3,
	}
}

// SetClient configures the OpenAI client for the Embedder.
func (b *EmbedderBuilder) SetClient(client *openai.Client) *EmbedderBuilder {
	b.client = client
	return b
}

// SetModel configures the embedding model to be used by the Embedder.
func (b *EmbedderBuilder) SetModel(model openai.EmbeddingModel) *EmbedderBuilder {
	b.model = model
	return b
}

// Build constructs the Embedder instance based on the current configuration.
// Returns an error if the required OpenAI client is not configured.
func (b *EmbedderBuilder) Build() (*Embedder, error) {
	if b.client == nil {
		return nil, errors.New("openai client is not configured")
	}
	return &Embedder{
		client: b.client,
		model:  b.model,
	}, nil
}
