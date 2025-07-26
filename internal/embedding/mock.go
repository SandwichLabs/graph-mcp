package embedding

type MockService struct{}

// NewMockService creates a new MockService.
func NewMockService() Service {
	return &MockService{}
}

// GetEmbeddings returns a mock embedding response.
func (m *MockService) GetEmbeddings(text string, embeddingType EmbeddingType) (EmbedResponse, error) {
	if text == "" {
		return nil, nil // Return nil for empty text
	}
	// Return a mock embedding response
	mockEmbedding := make(EmbedResponse, 768) // Assuming 768 dimensions for the embedding
	for i := range mockEmbedding {
		mockEmbedding[i] = float32(i) / 1000.0 // Mock values
	}
	return mockEmbedding, nil
}

// GetType returns the type of the embedding service.
func (m *MockService) GetType() Provider {
	return ProviderTestMock
}
