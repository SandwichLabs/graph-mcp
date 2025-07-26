package embedding

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/genai"
)

type EmbeddingType string

const (
	EmbeddingTypeRetrievalDocument EmbeddingType = "RETRIEVAL_DOCUMENT"
	EmbeddintTypeRetrievalQuery    EmbeddingType = "RETRIEVAL_QUERY"
)

type EmbedResponse = []float32

// Service represents a service that interacts with the embedding client.
type Service interface {
	GetEmbeddings(text string, embeddingType EmbeddingType) (EmbedResponse, error)
}

// Provider is an enum for the embedding providers.
type Provider string

const (
	ProviderGemini   Provider = "gemini"
	ProviderMistral  Provider = "mistral"
	ProviderTestMock Provider = "testing" // For testing purposes
)

type service struct {
	client *genai.Client
}

var (
	apiKey          = os.Getenv("GEMINI_API_KEY")
	serviceInstance *service
)

// New creates a new embedding service based on the specified provider.
func New(provider Provider) (Service, error) {
	switch provider {
	case ProviderGemini:
		return newGeminiService(), nil
	case ProviderMistral:
		return NewMistralService(), nil
	case ProviderTestMock:
		// For testing purposes, we can return a mock service.
		return NewMockService(), nil
	default:
		return nil, fmt.Errorf("unknown embedding provider: %s", provider)
	}
}

// geminiService is a service that interacts with the Gemini API.
type geminiService struct {
	client *genai.Client
}

// newGeminiService creates a new geminiService.
func newGeminiService() Service {
	apiKey := os.Getenv("GEMINI_API_KEY")
	ctx := context.Background()
	clientInstance, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		slog.Error("failed to create genai client", "error", err)
		panic("failed to create genai client: " + err.Error())
	}
	slog.Info("genai client created successfully")
	return &geminiService{
		client: clientInstance,
	}
}

func extractEmbeddingVector(embeddings []*genai.ContentEmbedding) EmbedResponse {
	if len(embeddings) == 0 {
		return nil
	}
	return embeddings[0].Values
}

// GetEmbeddings sends a request to the Gemini API to get embeddings for the given text.
func (s *geminiService) GetEmbeddings(text string, embeddingType EmbeddingType) (EmbedResponse, error) {
	ctx := context.Background()
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}
	slog.Info("Requesting embeddings", "text", text, "embeddingType", string(embeddingType))
	result, err := s.client.Models.EmbedContent(ctx,
		"gemini-embedding-exp-03-07",
		contents,
		&genai.EmbedContentConfig{
			TaskType: string(embeddingType),
		},
	)
	if err != nil {
		slog.Error("failed to get embeddings", "error", err)
		return nil, err
	}

	embedResponse := extractEmbeddingVector(result.Embeddings)

	return embedResponse, nil
}
