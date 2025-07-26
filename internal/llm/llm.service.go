package llm

import (
	"context"
	"fmt"
)

// Provider is an enum for the LLM providers.
type Provider string

const (
	ProviderMistral Provider = "mistral"
	// Add other providers like ProviderGemini if needed in the future
)

// LlmService defines the interface for Large Language Model services.
// It includes methods for text generation and extracting text from images.
type LlmService interface {
	// GenerateText generates text based on a given prompt.
	GenerateText(ctx context.Context, prompt string) (responseText string, err error)

	// ExtractTextFromImage extracts relevant text from an image based on a guiding prompt.
	// image is the byte representation of the image.
	// mimeType is the MIME type of the image (e.g., "image/jpeg", "image/png").
	ExtractTextFromImage(ctx context.Context, prompt string, image []byte, mimeType string) (extractedText string, err error)
}

// NewLlmService acts as a factory to create instances of LlmService
// based on the specified provider.
func NewLlmService(provider Provider) (LlmService, error) {
	switch provider {
	case ProviderMistral:
		return NewMistralLlmService()
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", provider)
	}
}
