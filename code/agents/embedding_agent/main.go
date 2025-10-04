// Package main provides the embedding agent for the GOX framework.
//
// The embedding agent generates vector embeddings for text/code chunks using
// various providers (OpenAI, HuggingFace, local models). It caches embeddings
// in VFS for efficiency and supports batch processing.
//
// Key Features:
// - Multiple embedding providers (OpenAI, HuggingFace, local ONNX)
// - Batch processing for efficiency
// - VFS-based caching with project isolation
// - Configurable models and dimensions
// - Error handling with partial success reporting
//
// Configuration (via cells.yaml):
// - provider: "openai" | "huggingface" | "local"
// - model: Model identifier (e.g., "text-embedding-3-small")
// - batch_size: Number of texts to process per API call
// - cache_enabled: Enable/disable embedding cache
// - timeout: API request timeout
//
// Called by: RAG agent, chunking pipeline, indexing workflows
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// EmbeddingAgent implements the AgentRunner interface for embedding generation
type EmbeddingAgent struct {
	agent.DefaultAgentRunner
	provider EmbeddingProvider
	config   *EmbeddingConfig
	cache    *EmbeddingCache
}

// EmbeddingConfig defines configuration options for the embedding agent
type EmbeddingConfig struct {
	Provider     string        `json:"provider"`      // "openai", "huggingface", "local"
	Model        string        `json:"model"`         // Model identifier
	APIKey       string        `json:"-"`             // API key (from env, not logged)
	BatchSize    int           `json:"batch_size"`    // Texts per API call
	CacheEnabled bool          `json:"cache_enabled"` // Enable embedding cache
	Timeout      time.Duration `json:"timeout"`       // API request timeout
	Dimensions   int           `json:"dimensions"`    // Expected embedding dimensions
}

// EmbeddingRequest represents an embedding generation request
type EmbeddingRequest struct {
	RequestID string   `json:"request_id"`
	Texts     []string `json:"texts"`
	Provider  string   `json:"provider,omitempty"`  // Override default
	Model     string   `json:"model,omitempty"`     // Override default
	ProjectID string   `json:"project_id,omitempty"` // For cache isolation
}

// EmbeddingResponse represents an embedding generation response
type EmbeddingResponse struct {
	RequestID    string      `json:"request_id"`
	Embeddings   [][]float32 `json:"embeddings"`
	Dimensions   int         `json:"dimensions"`
	Model        string      `json:"model"`
	Provider     string      `json:"provider"`
	CachedCount  int         `json:"cached_count"`
	GeneratedCount int       `json:"generated_count"`
	Error        string      `json:"error,omitempty"`
}

// EmbeddingProvider interface for different embedding backends
type EmbeddingProvider interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
	Model() string
	Provider() string
}

// OpenAIProvider implements EmbeddingProvider for OpenAI
type OpenAIProvider struct {
	apiKey     string
	model      string
	dimensions int
	timeout    time.Duration
	httpClient *http.Client
}

// OpenAI API structures
type openAIEmbeddingRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

type openAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIProvider creates a new OpenAI embedding provider
func NewOpenAIProvider(apiKey, model string, dimensions int, timeout time.Duration) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:     apiKey,
		model:      model,
		dimensions: dimensions,
		timeout:    timeout,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := openAIEmbeddingRequest{
		Input:          texts,
		Model:          p.model,
		EncodingFormat: "float",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp openAIEmbeddingResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract embeddings in order
	embeddings := make([][]float32, len(texts))
	for _, item := range apiResp.Data {
		if item.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid index %d in response", item.Index)
		}
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

func (p *OpenAIProvider) Dimensions() int {
	return p.dimensions
}

func (p *OpenAIProvider) Model() string {
	return p.model
}

func (p *OpenAIProvider) Provider() string {
	return "openai"
}

// EmbeddingCache provides VFS-based caching for embeddings
type EmbeddingCache struct {
	base *agent.BaseAgent
}

func NewEmbeddingCache(base *agent.BaseAgent) *EmbeddingCache {
	return &EmbeddingCache{base: base}
}

func (c *EmbeddingCache) Get(text string) ([]float32, bool) {
	if c.base.VFS == nil {
		return nil, false
	}

	hash := c.hashText(text)
	data, err := c.base.ReadFile("embeddings", "cache", hash+".json")
	if err != nil {
		return nil, false
	}

	var embedding []float32
	if err := json.Unmarshal(data, &embedding); err != nil {
		return nil, false
	}

	return embedding, true
}

func (c *EmbeddingCache) Set(text string, embedding []float32) error {
	if c.base.VFS == nil {
		return nil // Caching disabled
	}

	hash := c.hashText(text)
	data, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	c.base.MkdirVFS("embeddings", "cache")
	return c.base.WriteFile(data, "embeddings", "cache", hash+".json")
}

func (c *EmbeddingCache) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash[:16]) // Use first 16 bytes for filename
}

// Init initializes the embedding agent
func (e *EmbeddingAgent) Init(base *agent.BaseAgent) error {
	// Load configuration
	config := &EmbeddingConfig{
		Provider:     "openai",
		Model:        "text-embedding-3-small",
		BatchSize:    100,
		CacheEnabled: true,
		Timeout:      30 * time.Second,
		Dimensions:   1536,
	}

	// Override from base agent config
	if provider, ok := base.Config["provider"].(string); ok {
		config.Provider = provider
	}
	if model, ok := base.Config["model"].(string); ok {
		config.Model = model
	}
	if batchSize, ok := base.Config["batch_size"].(float64); ok {
		config.BatchSize = int(batchSize)
	}
	if timeout, ok := base.Config["timeout"].(float64); ok {
		config.Timeout = time.Duration(timeout)
	}
	if dims, ok := base.Config["dimensions"].(float64); ok {
		config.Dimensions = int(dims)
	}

	// Get API key from environment
	config.APIKey = os.Getenv("OPENAI_API_KEY")
	if config.APIKey == "" && config.Provider == "openai" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	e.config = config

	// Initialize provider
	switch config.Provider {
	case "openai":
		e.provider = NewOpenAIProvider(config.APIKey, config.Model, config.Dimensions, config.Timeout)
	case "huggingface":
		return fmt.Errorf("HuggingFace provider not yet implemented")
	case "local":
		return fmt.Errorf("local provider not yet implemented")
	default:
		return fmt.Errorf("unknown provider: %s", config.Provider)
	}

	// Initialize cache if enabled
	if config.CacheEnabled && base.VFS != nil {
		e.cache = NewEmbeddingCache(base)
		base.MkdirVFS("embeddings", "cache")
	}

	base.LogInfo("Embedding agent initialized: provider=%s, model=%s, dimensions=%d, cache=%v",
		config.Provider, config.Model, config.Dimensions, config.CacheEnabled)

	return nil
}

// ProcessMessage handles embedding generation requests
func (e *EmbeddingAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	base.LogDebug("EmbeddingAgent received message %s", msg.ID)

	var request EmbeddingRequest
	var err error

	// Parse request
	switch payload := msg.Payload.(type) {
	case []byte:
		err = json.Unmarshal(payload, &request)
	case string:
		err = json.Unmarshal([]byte(payload), &request)
	default:
		payloadBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", marshalErr)
		}
		err = json.Unmarshal(payloadBytes, &request)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Validate request
	if len(request.Texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	if request.RequestID == "" {
		request.RequestID = msg.ID
	}

	base.LogDebug("Generating embeddings for %d texts", len(request.Texts))

	// Check cache for each text
	embeddings := make([][]float32, len(request.Texts))
	cachedCount := 0
	textsToEmbed := []string{}
	indicesToEmbed := []int{}

	if e.cache != nil {
		for i, text := range request.Texts {
			if cached, found := e.cache.Get(text); found {
				embeddings[i] = cached
				cachedCount++
				base.LogDebug("Cache hit for text %d", i)
			} else {
				textsToEmbed = append(textsToEmbed, text)
				indicesToEmbed = append(indicesToEmbed, i)
			}
		}
	} else {
		textsToEmbed = request.Texts
		for i := range request.Texts {
			indicesToEmbed = append(indicesToEmbed, i)
		}
	}

	// Generate embeddings for uncached texts
	generatedCount := 0
	if len(textsToEmbed) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
		defer cancel()

		// Process in batches
		for i := 0; i < len(textsToEmbed); i += e.config.BatchSize {
			end := i + e.config.BatchSize
			if end > len(textsToEmbed) {
				end = len(textsToEmbed)
			}

			batch := textsToEmbed[i:end]
			batchIndices := indicesToEmbed[i:end]

			base.LogDebug("Processing batch %d-%d", i, end)

			generated, err := e.provider.Embed(ctx, batch)
			if err != nil {
				return nil, fmt.Errorf("failed to generate embeddings: %w", err)
			}

			// Store results and cache
			for j, embedding := range generated {
				idx := batchIndices[j]
				embeddings[idx] = embedding
				generatedCount++

				// Cache if enabled
				if e.cache != nil {
					if err := e.cache.Set(textsToEmbed[i+j], embedding); err != nil {
						base.LogError("Failed to cache embedding: %v", err)
					}
				}
			}
		}
	}

	base.LogInfo("Generated %d embeddings (%d cached, %d new)", len(request.Texts), cachedCount, generatedCount)

	// Build response
	response := EmbeddingResponse{
		RequestID:      request.RequestID,
		Embeddings:     embeddings,
		Dimensions:     e.provider.Dimensions(),
		Model:          e.provider.Model(),
		Provider:       e.provider.Provider(),
		CachedCount:    cachedCount,
		GeneratedCount: generatedCount,
	}

	return &client.BrokerMessage{
		ID:      msg.ID + "_embedded",
		Type:    "embedding_response",
		Payload: response,
		Meta: map[string]interface{}{
			"dimensions":      response.Dimensions,
			"model":           response.Model,
			"provider":        response.Provider,
			"cached_count":    cachedCount,
			"generated_count": generatedCount,
		},
	}, nil
}

func main() {
	if err := agent.Run(&EmbeddingAgent{}, "embedding-agent"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
