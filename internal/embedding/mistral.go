package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// MistralService is a service that interacts with the Mistral API.
type MistralService struct {
	apiKey string
	client *http.Client
}

// NewMistralService creates a new MistralService.
func NewMistralService() Service {
	return &MistralService{
		apiKey: os.Getenv("MISTRAL_API_KEY"),
		client: &http.Client{},
	}
}

// GetEmbeddings sends a request to the Mistral API to get embeddings for the given text.
func (s *MistralService) GetEmbeddings(text string, embeddingType EmbeddingType) (EmbedResponse, error) {
	// Prepare the request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "mistral-embed",
		"input": []string{text},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Send the request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mistral API error: %s - %s", resp.Status, string(bodyBytes))
	}

	// Decode the response
	var mistralResponse struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mistralResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(mistralResponse.Data) == 0 {
		return nil, fmt.Errorf("no embeddings found in response")
	}

	response := mistralResponse.Data[0].Embedding

	return (EmbedResponse)(response), nil
}
