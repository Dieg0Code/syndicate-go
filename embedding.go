package gomaky

import (
	"context"
	"errors"
	"fmt"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

// Embedder se encarga exclusivamente de generar embeddings.
type Embedder struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

func (e *Embedder) GenerateEmbedding(ctx context.Context, data string, model ...openai.EmbeddingModel) ([]float32, error) {
	if data == "" {
		return nil, errors.New("input data cannot be empty")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	modelName := e.model
	if len(model) > 0 && model[0] != "" {
		modelName = model[0]
	}

	req := openai.EmbeddingRequest{
		Input: []string{data},
		Model: modelName,
	}

	res, err := e.client.CreateEmbeddings(ctx, req)
	if err != nil {
		log.Printf("failed to create embeddings for model '%s': %v", modelName, err)
		return nil, fmt.Errorf("create embeddings error: %w", err)
	}

	if len(res.Data) == 0 {
		return nil, errors.New("no embedding data returned")
	}

	return res.Data[0].Embedding, nil
}

type EmbedderBuilder struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

func NewEmbedderBuilder() *EmbedderBuilder {
	return &EmbedderBuilder{
		model: openai.LargeEmbedding3,
	}
}

func (b *EmbedderBuilder) SetClient(client *openai.Client) *EmbedderBuilder {
	b.client = client
	return b
}

func (b *EmbedderBuilder) SetModel(model openai.EmbeddingModel) *EmbedderBuilder {
	b.model = model
	return b
}

func (b *EmbedderBuilder) Build() (*Embedder, error) {
	if b.client == nil {
		return nil, errors.New("openai client is not configured")
	}
	return &Embedder{
		client: b.client,
		model:  b.model,
	}, nil
}
