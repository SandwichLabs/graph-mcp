package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// mockMistralServer sets up a test HTTP server to mock the Mistral API.
// It allows customizing the response handler for different test cases.
func mockMistralServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// After refactor, service uses APIBaseURL (server.URL) + "/chat/completions"
		if r.URL.Path == "/chat/completions" {
			handler(w, r)
		} else {
			http.Error(w, fmt.Sprintf("Not found: Unexpected path %s, expected /chat/completions", r.URL.Path), http.StatusNotFound)
		}
	}))
}

func TestMistralLlmService_GenerateText_Success(t *testing.T) {
	expectedResponseText := "This is a test response."
	server := mockMistralServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": expectedResponseText,
					},
				},
			},
		})
	})
	defer server.Close()

	os.Setenv("MISTRAL_API_KEY", "test_api_key")
	service, err := NewMistralLlmService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	// If not, MistralLlmService needs refactoring for testability.
	service.HTTPClient = server.Client() // Use exported field
	service.APIBaseURL = server.URL      // Point service to the mock server

	actualText, err := service.GenerateText(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("GenerateText failed: %v", err)
	}
	if actualText != expectedResponseText {
		t.Errorf("Expected text '%s', got '%s'", expectedResponseText, actualText)
	}
}

func TestMistralLlmService_GenerateText_APIError(t *testing.T) {
	server := mockMistralServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	os.Setenv("MISTRAL_API_KEY", "test_api_key")
	service, err := NewMistralLlmService()
	if err != nil {
		t.Fatalf("NewMistralLlmService failed: %v", err)
	}
	service.HTTPClient = server.Client() // Use exported field
	service.APIBaseURL = server.URL      // Point service to the mock server

	_, err = service.GenerateText(context.Background(), "test prompt")
	if err == nil {
		t.Fatalf("Expected an error, but got nil")
	}
	// Check for specific parts of the error message
	if !strings.Contains(err.Error(), "mistral API error") || !strings.Contains(err.Error(), "500 Internal Server Error") {
		t.Errorf("Expected error to contain 'mistral API error' and '500 Internal Server Error', got: %v", err)
	}
}

func TestMistralLlmService_GenerateText_MalformedResponse(t *testing.T) {
	server := mockMistralServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices": [{"message": {"content": "test"}}`) // Malformed JSON
	})
	defer server.Close()

	os.Setenv("MISTRAL_API_KEY", "test_api_key")
	service, _ := NewMistralLlmService()
	service.HTTPClient = server.Client()
	service.APIBaseURL = server.URL

	_, err := service.GenerateText(context.Background(), "test prompt")
	if err == nil {
		t.Fatalf("Expected an error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode mistral response") {
		t.Errorf("Expected error to contain 'failed to decode mistral response', got: %v", err)
	}
}

func TestMistralLlmService_GenerateText_EmptyChoices(t *testing.T) {
	server := mockMistralServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{}, // Empty choices
		})
	})
	defer server.Close()

	os.Setenv("MISTRAL_API_KEY", "test_api_key")
	service, _ := NewMistralLlmService()
	service.HTTPClient = server.Client() // Corrected from httpClient
	service.APIBaseURL = server.URL      // Added APIBaseURL setting

	_, err := service.GenerateText(context.Background(), "test prompt")
	if err == nil {
		t.Fatalf("Expected an error for empty choices, got nil")
	}
	if !strings.Contains(err.Error(), "no content found in mistral response") {
		t.Errorf("Expected error to contain 'no content found in mistral response', got: %v", err)
	}
}

func TestMistralLlmService_ExtractTextFromImage_Success(t *testing.T) {
	expectedResponseText := "Wine Name: Test Wine, Region: Test Region, Varietal: Test Varietal"
	server := mockMistralServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, "Bad request body, not JSON", http.StatusBadRequest)
			return
		}

		messages, ok := payload["messages"].([]interface{})
		if !ok || len(messages) == 0 {
			http.Error(w, "Payload missing 'messages' array or it's empty", http.StatusBadRequest)
			return
		}
		firstMessage, ok := messages[0].(map[string]interface{})
		if !ok {
			http.Error(w, "First message is not a map", http.StatusBadRequest)
			return
		}
		contentArray, ok := firstMessage["content"].([]interface{})
		if !ok || len(contentArray) < 2 {
			http.Error(w, "First message 'content' is not an array or has less than 2 elements", http.StatusBadRequest)
			return
		}

		hasText := false
		hasImage := false
		for _, item := range contentArray {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			itemType, ok := itemMap["type"].(string)
			if !ok {
				continue
			}
			if itemType == "text" {
				hasText = true
			} else if itemType == "image_url" {
				hasImage = true
			}
		}
		if !hasText || !hasImage {
			http.Error(w, "Payload does not have text and image_url parts in content", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": expectedResponseText,
					},
				},
			},
		})
	})
	defer server.Close()

	os.Setenv("MISTRAL_API_KEY", "test_api_key")
	service, err := NewMistralLlmService()
	if err != nil {
		t.Fatalf("NewMistralLlmService failed: %v", err)
	}
	service.HTTPClient = server.Client()
	service.APIBaseURL = server.URL

	imageData := []byte("dummyimagedata")
	mimeType := "image/jpeg"
	prompt := "Extract wine info"

	actualText, err := service.ExtractTextFromImage(context.Background(), prompt, imageData, mimeType)
	if err != nil {
		t.Fatalf("ExtractTextFromImage failed: %v", err)
	}
	if actualText != expectedResponseText {
		t.Errorf("Expected text '%s', got '%s'", expectedResponseText, actualText)
	}
}

func TestMistralLlmService_ExtractTextFromImage_EmptyImage(t *testing.T) {
	os.Setenv("MISTRAL_API_KEY", "test_api_key")
	service, err := NewMistralLlmService()
	if err != nil {
		t.Fatalf("NewMistralLlmService failed: %v", err)
	}
	// No server needed as this should be caught before API call by the service.

	_, err = service.ExtractTextFromImage(context.Background(), "prompt", []byte{}, "image/png")
	if err == nil {
		t.Fatalf("Expected an error for empty image data, got nil")
	}
	if !strings.Contains(err.Error(), "image data is empty") {
		t.Errorf("Expected error to contain 'image data is empty', got: %v", err)
	}
}

func TestMistralLlmService_ExtractTextFromImage_APIError(t *testing.T) {
	server := mockMistralServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
	})
	defer server.Close()

	os.Setenv("MISTRAL_API_KEY", "test_api_key")
	service, _ := NewMistralLlmService()
	service.HTTPClient = server.Client()
	service.APIBaseURL = server.URL

	_, err := service.ExtractTextFromImage(context.Background(), "prompt", []byte("dummyData"), "image/jpeg")
	if err == nil {
		t.Fatalf("Expected an API error, got nil")
	}
	if !strings.Contains(err.Error(), "mistral API error (multimodal)") || !strings.Contains(err.Error(), "504 Gateway Timeout") {
		t.Errorf("Expected error to contain 'mistral API error (multimodal)' and '504 Gateway Timeout', got: %v", err)
	}
}
