package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// MistralLlmService implements the LlmService interface using the Mistral API.
type MistralLlmService struct {
	apiKey          string
	HTTPClient      *http.Client // Exported for testing
	chatModel       string
	multimodalModel string
	APIBaseURL      string // Added for testing and flexibility
}

// NewMistralLlmService creates a new instance of MistralLlmService.
// It requires the API key to be set in the MISTRAL_API_KEY environment variable.
func NewMistralLlmService() (*MistralLlmService, error) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY environment variable not set")
	}

	return &MistralLlmService{
		apiKey:          apiKey,
		HTTPClient:      &http.Client{},
		chatModel:       "mistral-small-latest",
		multimodalModel: "mistral-medium-latest",
		APIBaseURL:      "https://api.mistral.ai/v1", // Default API base URL
	}, nil
}

// GenerateText generates text using the Mistral chat completions API.
func (s *MistralLlmService) GenerateText(ctx context.Context, prompt string) (string, error) {
	slog.InfoContext(ctx, "MistralLlmService: GenerateText called", "model", s.chatModel, "prompt_length", len(prompt))

	requestPayload := map[string]interface{}{
		"model": s.chatModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		// Optional parameters - good to have some defaults
		"temperature": 0.7,
		"max_tokens":  500,
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to marshal request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := s.APIBaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to create HTTP request", "error", err, "url", url)
		return "", fmt.Errorf("failed to create request to %s: %w", url, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to send request to Mistral API", "error", err, "url", url)
		return "", fmt.Errorf("failed to send request to Mistral API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.ErrorContext(ctx, "MistralLlmService: Mistral API error", "status_code", resp.StatusCode, "response_body", string(bodyBytes))
		return "", fmt.Errorf("mistral API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var mistralResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mistralResponse); err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to decode Mistral API response", "error", err)
		return "", fmt.Errorf("failed to decode mistral response: %w", err)
	}

	if len(mistralResponse.Choices) == 0 || mistralResponse.Choices[0].Message.Content == "" {
		slog.WarnContext(ctx, "MistralLlmService: No content found in Mistral API response", "response", mistralResponse)
		return "", fmt.Errorf("no content found in mistral response")
	}

	slog.InfoContext(ctx, "MistralLlmService: Text generated successfully", "response_length", len(mistralResponse.Choices[0].Message.Content))
	return mistralResponse.Choices[0].Message.Content, nil
}

// ExtractTextFromImage extracts text from an image using a Mistral multimodal model
// by encoding the image as base64 and sending it with a text prompt.
func (s *MistralLlmService) ExtractTextFromImage(ctx context.Context, prompt string, image []byte, mimeType string) (string, error) {
	slog.InfoContext(ctx, "MistralLlmService: ExtractTextFromImage called",
		"model", s.multimodalModel,
		"prompt_length", len(prompt),
		"image_size", len(image),
		"mime_type", mimeType)

	if len(image) == 0 {
		slog.ErrorContext(ctx, "MistralLlmService: Image data is empty")
		return "", fmt.Errorf("image data is empty")
	}

	// Validate or default MIME type if necessary.
	// For "data:<mimeType>;base64,...", the mimeType needs to be accurate.
	// Example: "image/jpeg", "image/png"
	if mimeType == "" {
		// Attempt to detect, or default, or return error
		// For simplicity, we'll require it for now or default to a common one if appropriate for the API
		slog.WarnContext(ctx, "MistralLlmService: MimeType is empty, defaulting to image/jpeg. Accurate MimeType is preferred.")
		mimeType = "image/jpeg" // Or handle more robustly
	}

	base64Image := base64.StdEncoding.EncodeToString(image)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	requestPayload := map[string]interface{}{
		"model": s.multimodalModel,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
				},
			},
		},
		"temperature": 0.2, // Lower temperature for more factual extraction
		"max_tokens":  300, // Max tokens for the extracted information
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to marshal multimodal request body", "error", err)
		return "", fmt.Errorf("failed to marshal multimodal request body: %w", err)
	}

	url := s.APIBaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to create multimodal HTTP request", "error", err, "url", url)
		return "", fmt.Errorf("failed to create multimodal request to %s: %w", url, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to send multimodal request to Mistral API", "error", err, "url", url)
		return "", fmt.Errorf("failed to send multimodal request to Mistral API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.ErrorContext(ctx, "MistralLlmService: Mistral API error on multimodal request", "status_code", resp.StatusCode, "response_body", string(bodyBytes))
		return "", fmt.Errorf("mistral API error (multimodal): %s - %s", resp.Status, string(bodyBytes))
	}

	var mistralResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mistralResponse); err != nil {
		slog.ErrorContext(ctx, "MistralLlmService: Failed to decode Mistral API multimodal response", "error", err)
		return "", fmt.Errorf("failed to decode mistral multimodal response: %w", err)
	}

	if len(mistralResponse.Choices) == 0 || mistralResponse.Choices[0].Message.Content == "" {
		slog.WarnContext(ctx, "MistralLlmService: No content found in Mistral API multimodal response", "response", mistralResponse)
		return "", fmt.Errorf("no content found in mistral multimodal response")
	}

	slog.InfoContext(ctx, "MistralLlmService: Text extracted from image successfully", "response_length", len(mistralResponse.Choices[0].Message.Content))
	return mistralResponse.Choices[0].Message.Content, nil
}

// Ensure NewMistralLlmService is correctly defined and callable from other packages.
// For now, the actual API call logic is deferred.
