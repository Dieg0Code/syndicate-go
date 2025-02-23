package syndicate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

// testEmbeddingRequest es una estructura auxiliar para decodificar el request JSON.
type testEmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// fakeEmbeddingResponse genera una respuesta JSON simulada de la API de embeddings.
func fakeEmbeddingResponse(embedding []float32, extraData bool) string {
	// Si extraData es true, se devuelve un JSON con data; si es false, se devuelve data vacío.
	dataField := "[]"
	if extraData {
		// Se devuelve un slice con un único objeto que contiene la embedding.
		// Se incluyen campos mínimos que se esperan, como "embedding" y "index".
		dataField = fmt.Sprintf(`[{"object": "embedding", "embedding": %s, "index": 0}]`, toJSON(embedding))
	}
	response := fmt.Sprintf(`{
		"object": "list",
		"data": %s,
		"model": "test-model",
		"usage": {
			"prompt_tokens": 5,
			"completion_tokens": 0,
			"total_tokens": 5
		}
	}`, dataField)
	return response
}

// toJSON convierte un slice de float32 a su representación JSON.
func toJSON(f []float32) string {
	b, _ := json.Marshal(f)
	return string(b)
}

// TestGenerateEmbeddingEmptyData verifica que se retorne error si se suministra data vacía.
func TestGenerateEmbeddingEmptyData(t *testing.T) {
	// Aunque el cliente no se llegue a usar, se crea un fake client.
	client := openai.NewClient("dummy-key")
	embedder := &Embedder{client: client, model: openai.LargeEmbedding3}

	_, err := embedder.GenerateEmbedding(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "input data cannot be empty") {
		t.Errorf("se esperaba error por input vacío, se obtuvo: %v", err)
	}
}

// TestGenerateEmbeddingSuccess verifica el flujo exitoso, comprobando además que se use el modelo override.
func TestGenerateEmbeddingSuccess(t *testing.T) {
	// Variable para capturar el modelo enviado en el request.
	var requestedModel string

	// Creamos un servidor HTTP fake que simula la respuesta de la API de embeddings.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Leemos y decodificamos el request JSON para verificar el campo Model.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error leyendo body: %v", err)
		}
		var req testEmbeddingRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("error decodificando request: %v", err)
		}
		if len(req.Input) != 1 || req.Input[0] != "prueba de embedding" {
			t.Errorf("input inesperado: %v", req.Input)
		}
		requestedModel = req.Model

		// Se responde con una embedding simulada.
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fakeEmbeddingResponse([]float32{0.1, 0.2, 0.3}, true))
	}))
	defer server.Close()

	// Creamos un cliente de OpenAI configurado para usar el servidor fake.
	config := openai.DefaultConfig("dummy-key")
	config.BaseURL = server.URL
	client := openai.NewClientWithConfig(config)

	embedder := &Embedder{client: client, model: openai.LargeEmbedding3}

	// Llamamos a GenerateEmbedding y pasamos un override de modelo ("custom-model").
	embedding, err := embedder.GenerateEmbedding(context.Background(), "prueba de embedding", "custom-model")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(embedding) != 3 {
		t.Fatalf("se esperaba embedding de 3 elementos, se obtuvo %d", len(embedding))
	}
	expected := []float32{0.1, 0.2, 0.3}
	for i, v := range expected {
		if embedding[i] != v {
			t.Errorf("en la posición %d se esperaba %v, se obtuvo %v", i, v, embedding[i])
		}
	}
	// Verificamos que se haya usado el modelo override.
	if requestedModel != "custom-model" {
		t.Errorf("se esperaba modelo 'custom-model', se recibió '%s'", requestedModel)
	}
}

// TestGenerateEmbeddingNoData simula una respuesta sin datos de embedding y verifica el error.
func TestGenerateEmbeddingNoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Se responde con una respuesta válida pero sin datos en "data".
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fakeEmbeddingResponse(nil, false))
	}))
	defer server.Close()

	config := openai.DefaultConfig("dummy-key")
	config.BaseURL = server.URL
	client := openai.NewClientWithConfig(config)

	embedder := &Embedder{client: client, model: openai.LargeEmbedding3}
	_, err := embedder.GenerateEmbedding(context.Background(), "algún texto")
	if err == nil || !strings.Contains(err.Error(), "no embedding data returned") {
		t.Errorf("se esperaba error por falta de data, se obtuvo: %v", err)
	}
}

// TestGenerateEmbeddingClientError simula un error en la llamada a CreateEmbeddings.
func TestGenerateEmbeddingClientError(t *testing.T) {
	// Servidor fake que devuelve un error HTTP.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "simulated error", http.StatusInternalServerError)
	}))
	defer server.Close()

	config := openai.DefaultConfig("dummy-key")
	config.BaseURL = server.URL
	client := openai.NewClientWithConfig(config)

	embedder := &Embedder{client: client, model: openai.LargeEmbedding3}
	_, err := embedder.GenerateEmbedding(context.Background(), "texto de prueba")
	if err == nil || !strings.Contains(err.Error(), "create embeddings error:") {
		t.Errorf("se esperaba error de CreateEmbeddings, se obtuvo: %v", err)
	}
}

// TestEmbedderBuilderMissingClient verifica que el builder falle si no se configura el cliente.
func TestEmbedderBuilderMissingClient(t *testing.T) {
	builder := NewEmbedderBuilder()
	_, err := builder.Build()
	if err == nil || !strings.Contains(err.Error(), "openai client is not configured") {
		t.Errorf("se esperaba error por falta de cliente, se obtuvo: %v", err)
	}
}

// TestEmbedderBuilderSuccess verifica que el builder construya correctamente un Embedder.
func TestEmbedderBuilderSuccess(t *testing.T) {
	// Para este test, no necesitamos un servidor fake ya que no se realizará una llamada.
	client := openai.NewClient("dummy-key")
	builder := NewEmbedderBuilder().SetClient(client).SetModel("custom-model")
	embedder, err := builder.Build()
	if err != nil {
		t.Fatalf("error al construir el embedder: %v", err)
	}
	if embedder.client != client {
		t.Error("el cliente configurado no coincide")
	}
	if embedder.model != "custom-model" {
		t.Errorf("se esperaba el modelo 'custom-model', se obtuvo '%s'", embedder.model)
	}
}

// TestGenerateEmbeddingNilContext verifica que si se pasa un contexto nil se use context.Background.
func TestGenerateEmbeddingNilContext(t *testing.T) {
	// Este test es similar al de éxito; se envía nil como contexto.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No es necesario validar el contenido, solo responder.
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fakeEmbeddingResponse([]float32{0.4, 0.5, 0.6}, true))
	}))
	defer server.Close()

	config := openai.DefaultConfig("dummy-key")
	config.BaseURL = server.URL
	client := openai.NewClientWithConfig(config)

	embedder := &Embedder{client: client, model: openai.LargeEmbedding3}
	embedding, err := embedder.GenerateEmbedding(context.TODO(), "texto con nil context")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	expected := []float32{0.4, 0.5, 0.6}
	if len(embedding) != len(expected) {
		t.Fatalf("se esperaba embedding de %d elementos, se obtuvo %d", len(expected), len(embedding))
	}
	for i, v := range expected {
		if embedding[i] != v {
			t.Errorf("en la posición %d se esperaba %v, se obtuvo %v", i, v, embedding[i])
		}
	}
}
