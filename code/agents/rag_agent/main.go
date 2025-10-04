// Package main provides the RAG (Retrieval-Augmented Generation) agent for GOX.
//
// The RAG agent orchestrates the retrieval workflow: query embedding generation,
// vector similarity search, content fetching, and context assembly for LLM consumption.
// It coordinates between embedding-agent, vectorstore-agent, and storage agents.
//
// Key Features:
// - Query embedding generation via embedding-agent
// - Vector similarity search via vectorstore-agent
// - Content fetching from storage
// - Optional reranking with TF-IDF
// - Context assembly with token limits
// - Metadata enrichment
//
// Configuration (via cells.yaml):
// - top_k: Number of results to retrieve (default: 5)
// - rerank: Enable TF-IDF reranking (default: false)
// - max_context_tokens: Maximum tokens in context (default: 4000)
// - include_surrounding_lines: Lines of context around chunks (default: 3)
// - score_threshold: Minimum similarity score (default: 0.0)
//
// Called by: API Gateway
// Calls: embedding-agent, vectorstore-agent, godast-storage
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// RAGAgent implements the AgentRunner interface
type RAGAgent struct {
	agent.DefaultAgentRunner
	config *RAGConfig
}

// RAGConfig defines configuration options
type RAGConfig struct {
	TopK                   int     `json:"top_k"`
	Rerank                 bool    `json:"rerank"`
	MaxContextTokens       int     `json:"max_context_tokens"`
	IncludeSurroundingLines int     `json:"include_surrounding_lines"`
	ScoreThreshold         float32 `json:"score_threshold"`
}

// RAGRequest represents a RAG query request
type RAGRequest struct {
	RequestID  string                 `json:"request_id"`
	Query      string                 `json:"query"`
	ProjectID  string                 `json:"project_id"`
	TopK       int                    `json:"top_k,omitempty"`
	Filter     map[string]interface{} `json:"filter,omitempty"`
	Rerank     bool                   `json:"rerank,omitempty"`
	IncludeCode bool                  `json:"include_code,omitempty"`
}

// RAGResponse represents the RAG query response
type RAGResponse struct {
	RequestID string        `json:"request_id"`
	Query     string        `json:"query"`
	Chunks    []ChunkResult `json:"chunks"`
	Context   string        `json:"context"`
	Metadata  RAGMetadata   `json:"metadata"`
	Error     string        `json:"error,omitempty"`
}

// ChunkResult represents a single retrieved chunk
type ChunkResult struct {
	ChunkID  string                 `json:"chunk_id"`
	Content  string                 `json:"content"`
	Score    float32                `json:"score"`
	File     string                 `json:"file"`
	Lines    [2]int                 `json:"lines"` // [start, end]
	Metadata map[string]interface{} `json:"metadata"`
}

// RAGMetadata contains metadata about the retrieval process
type RAGMetadata struct {
	TotalChunks     int   `json:"total_chunks"`
	SearchTimeMs    int64 `json:"search_time_ms"`
	EmbeddingTimeMs int64 `json:"embedding_time_ms"`
	FetchTimeMs     int64 `json:"fetch_time_ms"`
	RerankApplied   bool  `json:"rerank_applied"`
}

// Init initializes the RAG agent
func (r *RAGAgent) Init(base *agent.BaseAgent) error {
	// Load configuration
	config := &RAGConfig{
		TopK:                    5,
		Rerank:                  false,
		MaxContextTokens:        4000,
		IncludeSurroundingLines: 3,
		ScoreThreshold:          0.0,
	}

	// Override from base agent config
	if topK, ok := base.Config["top_k"].(float64); ok {
		config.TopK = int(topK)
	}
	if rerank, ok := base.Config["rerank"].(bool); ok {
		config.Rerank = rerank
	}
	if maxTokens, ok := base.Config["max_context_tokens"].(float64); ok {
		config.MaxContextTokens = int(maxTokens)
	}
	if surrounding, ok := base.Config["include_surrounding_lines"].(float64); ok {
		config.IncludeSurroundingLines = int(surrounding)
	}
	if threshold, ok := base.Config["score_threshold"].(float64); ok {
		config.ScoreThreshold = float32(threshold)
	}

	r.config = config

	base.LogInfo("RAG agent initialized: top_k=%d, rerank=%v, max_tokens=%d",
		config.TopK, config.Rerank, config.MaxContextTokens)

	return nil
}

// ProcessMessage handles RAG query requests
func (r *RAGAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	base.LogDebug("RAGAgent received message %s", msg.ID)

	var request RAGRequest
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
	if request.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if request.RequestID == "" {
		request.RequestID = msg.ID
	}

	// Override config with request parameters
	topK := r.config.TopK
	if request.TopK > 0 {
		topK = request.TopK
	}

	base.LogInfo("Processing RAG query: '%s' (top_k=%d)", request.Query, topK)

	// Initialize response
	response := RAGResponse{
		RequestID: request.RequestID,
		Query:     request.Query,
		Metadata:  RAGMetadata{},
	}

	// Step 1: Generate query embedding
	embeddingStart := time.Now()
	queryEmbedding, err := r.generateQueryEmbedding(request.Query, base)
	if err != nil {
		response.Error = fmt.Sprintf("embedding generation failed: %v", err)
		return r.buildErrorResponse(msg, response), nil
	}
	response.Metadata.EmbeddingTimeMs = time.Since(embeddingStart).Milliseconds()
	base.LogDebug("Query embedding generated in %dms", response.Metadata.EmbeddingTimeMs)

	// Step 2: Search vector store
	searchStart := time.Now()
	searchResults, err := r.searchVectors(queryEmbedding, topK, request.Filter, base)
	if err != nil {
		response.Error = fmt.Sprintf("vector search failed: %v", err)
		return r.buildErrorResponse(msg, response), nil
	}
	response.Metadata.SearchTimeMs = time.Since(searchStart).Milliseconds()
	base.LogDebug("Vector search completed in %dms, found %d results",
		response.Metadata.SearchTimeMs, len(searchResults))

	// Step 3: Fetch chunk content
	fetchStart := time.Now()
	chunks, err := r.fetchChunkContent(searchResults, base)
	if err != nil {
		response.Error = fmt.Sprintf("content fetch failed: %v", err)
		return r.buildErrorResponse(msg, response), nil
	}
	response.Metadata.FetchTimeMs = time.Since(fetchStart).Milliseconds()
	base.LogDebug("Content fetched in %dms", response.Metadata.FetchTimeMs)

	// Step 4: Apply score threshold
	filteredChunks := []ChunkResult{}
	for _, chunk := range chunks {
		if chunk.Score >= r.config.ScoreThreshold {
			filteredChunks = append(filteredChunks, chunk)
		}
	}
	chunks = filteredChunks

	// Step 5: Optional reranking
	if r.config.Rerank || request.Rerank {
		base.LogDebug("Applying reranking")
		chunks = r.rerankChunks(request.Query, chunks)
		response.Metadata.RerankApplied = true
	}

	// Step 6: Assemble context
	context := r.assembleContext(chunks, r.config.MaxContextTokens)

	response.Chunks = chunks
	response.Context = context
	response.Metadata.TotalChunks = len(chunks)

	base.LogInfo("RAG query completed: %d chunks retrieved, %d chars context",
		len(chunks), len(context))

	return &client.BrokerMessage{
		ID:      msg.ID + "_rag_response",
		Type:    "rag_response",
		Payload: response,
		Meta: map[string]interface{}{
			"query":        request.Query,
			"total_chunks": len(chunks),
			"project_id":   request.ProjectID,
		},
	}, nil
}

// generateQueryEmbedding generates embedding for the query
func (r *RAGAgent) generateQueryEmbedding(query string, base *agent.BaseAgent) ([]float32, error) {
	// In a real implementation, this would publish to embedding-agent topic
	// and wait for response. For MVP, we'll simulate or use direct call.

	// Create embedding request
	embeddingReq := map[string]interface{}{
		"request_id": fmt.Sprintf("embed_%d", time.Now().UnixNano()),
		"texts":      []string{query},
	}

	reqJSON, err := json.Marshal(embeddingReq)
	if err != nil {
		return nil, err
	}

	base.LogDebug("Sending embedding request for query")

	// For MVP: Return mock embedding (in production, use broker communication)
	// TODO: Implement broker-based communication with embedding-agent

	// Mock embedding for testing
	mockEmbedding := make([]float32, 1536)
	for i := range mockEmbedding {
		mockEmbedding[i] = 0.01 // Placeholder
	}

	base.LogDebug("Using mock embedding (TODO: implement broker communication)")
	_ = reqJSON // Will be used when broker communication is implemented

	return mockEmbedding, nil
}

// searchVectors searches the vector store
func (r *RAGAgent) searchVectors(queryVector []float32, topK int, filter map[string]interface{}, base *agent.BaseAgent) ([]VectorSearchResult, error) {
	// Create search request
	searchReq := map[string]interface{}{
		"operation":  "search",
		"request_id": fmt.Sprintf("search_%d", time.Now().UnixNano()),
		"query":      queryVector,
		"top_k":      topK,
		"filter":     filter,
	}

	reqJSON, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}

	base.LogDebug("Sending vector search request (top_k=%d)", topK)

	// For MVP: Return mock results (in production, use broker communication)
	// TODO: Implement broker-based communication with vectorstore-agent

	// Mock search results for testing
	mockResults := []VectorSearchResult{
		{ID: "chunk_001", Score: 0.95},
		{ID: "chunk_002", Score: 0.87},
		{ID: "chunk_003", Score: 0.78},
	}

	base.LogDebug("Using mock search results (TODO: implement broker communication)")
	_ = reqJSON // Will be used when broker communication is implemented

	return mockResults, nil
}

type VectorSearchResult struct {
	ID       string
	Score    float32
	Metadata map[string]interface{}
}

// fetchChunkContent fetches actual content for chunk IDs
func (r *RAGAgent) fetchChunkContent(results []VectorSearchResult, base *agent.BaseAgent) ([]ChunkResult, error) {
	chunks := make([]ChunkResult, 0, len(results))

	for _, result := range results {
		// In production, fetch from VFS or storage agent
		// For MVP, create mock chunks
		chunk := ChunkResult{
			ChunkID: result.ID,
			Content: r.getMockChunkContent(result.ID),
			Score:   result.Score,
			File:    "example.go",
			Lines:   [2]int{10, 25},
			Metadata: map[string]interface{}{
				"language": "go",
				"type":     "function",
			},
		}

		if result.Metadata != nil {
			for k, v := range result.Metadata {
				chunk.Metadata[k] = v
			}
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// getMockChunkContent returns mock content for testing
func (r *RAGAgent) getMockChunkContent(chunkID string) string {
	// Mock content for testing
	return fmt.Sprintf(`// Chunk: %s
func ProcessRequest(req Request) (*Response, error) {
    // Validate request
    if err := validateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %%w", err)
    }

    // Process data
    result := processData(req.Data)

    return &Response{
        Status: "success",
        Data:   result,
    }, nil
}`, chunkID)
}

// rerankChunks applies reranking based on query relevance
func (r *RAGAgent) rerankChunks(query string, chunks []ChunkResult) []ChunkResult {
	// Simple reranking: boost chunks that contain query terms
	queryTerms := strings.Fields(strings.ToLower(query))

	for i := range chunks {
		content := strings.ToLower(chunks[i].Content)
		matchCount := 0
		for _, term := range queryTerms {
			if strings.Contains(content, term) {
				matchCount++
			}
		}

		// Boost score based on term matches
		if matchCount > 0 {
			boost := float32(matchCount) * 0.1
			chunks[i].Score = chunks[i].Score + boost
			if chunks[i].Score > 1.0 {
				chunks[i].Score = 1.0
			}
		}
	}

	// Re-sort by updated scores
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Score > chunks[j].Score
	})

	return chunks
}

// assembleContext assembles chunks into LLM context string
func (r *RAGAgent) assembleContext(chunks []ChunkResult, maxTokens int) string {
	if len(chunks) == 0 {
		return ""
	}

	var builder strings.Builder
	estimatedTokens := 0

	for i, chunk := range chunks {
		// Format chunk
		chunkText := fmt.Sprintf("--- Source: %s (lines %d-%d, score: %.2f) ---\n%s\n\n",
			chunk.File, chunk.Lines[0], chunk.Lines[1], chunk.Score, chunk.Content)

		// Rough token estimation: ~4 chars per token
		chunkTokens := len(chunkText) / 4

		if estimatedTokens+chunkTokens > maxTokens {
			// Would exceed limit, stop here
			if i > 0 {
				builder.WriteString(fmt.Sprintf("... (truncated, showing %d of %d chunks)\n", i, len(chunks)))
			}
			break
		}

		builder.WriteString(chunkText)
		estimatedTokens += chunkTokens
	}

	return builder.String()
}

// buildErrorResponse creates an error response message
func (r *RAGAgent) buildErrorResponse(msg *client.BrokerMessage, response RAGResponse) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:      msg.ID + "_error",
		Type:    "rag_error",
		Payload: response,
		Meta: map[string]interface{}{
			"error": response.Error,
		},
	}
}

func main() {
	if err := agent.Run(&RAGAgent{}, "rag-agent"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
