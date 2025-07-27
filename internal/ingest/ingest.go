package ingest

import (
	"context"
	"fmt"

	"github.com/kuzudb/go-kuzu"
	"github.com/sandwichlabs/agent-memory-graph/internal/embedding"
	"github.com/sandwichlabs/agent-memory-graph/internal/llm"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

func IngestFile(filePath string) error {
	// Initialize services
	embeddingService, err := embedding.New(embedding.ProviderMistral)
	if err != nil {
		return fmt.Errorf("failed to create embedding service: %w", err)
	}

	llmService, err := llm.NewLlmService(llm.ProviderMistral)
	if err != nil {
		return fmt.Errorf("failed to create llm service: %w", err)
	}

	// Load and chunk document
	loader := documentloaders.NewText(filePath)
	docs, err := loader.Load(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load document: %w", err)
	}

	splitter := textsplitter.NewRecursiveCharacter()
	chunks, err := splitter.SplitDocuments(docs)
	if err != nil {
		return fmt.Errorf("failed to split document: %w", err)
	}

	// Setup KuzuDB
	db, err := kuzu.NewDatabase("amg.db", 0)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer db.Destroy()

	conn, err := kuzu.NewConnection(db)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}
	defer conn.Destroy()

	// Create schema
	_, err = conn.Query("CREATE TABLE Document (content STRING, embedding FLOAT[768], PRIMARY KEY (content))")
	if err != nil {
		fmt.Printf("Error creating table: %v\n", err)
	}

	// Ingest chunks
	for _, chunk := range chunks {
		// Get embedding
		embedding, err := embeddingService.GetEmbeddings(chunk.PageContent, embedding.EmbeddingTypeRetrievalDocument)
		if err != nil {
			return fmt.Errorf("failed to get embedding: %w", err)
		}

		// Ingest into KuzuDB
		query, err := conn.Prepare("CREATE (d:Document {content: $content, embedding: $embedding})")
		if err != nil {
			return fmt.Errorf("failed to prepare query: %w", err)
		}
		defer query.Destroy()

		params := map[string]interface{}{
			"content":   chunk.PageContent,
			"embedding": embedding,
		}

		_, err = conn.Execute(query, params)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		// Extract graph info with LLM
		prompt := fmt.Sprintf("Extract entities and relationships from the following text:\n\n%s", chunk.PageContent)
		graphInfo, err := llmService.GenerateText(context.Background(), prompt)
		if err != nil {
			return fmt.Errorf("failed to extract graph info: %w", err)
		}
		fmt.Println("Graph Info:", graphInfo)
	}

	return nil
}
