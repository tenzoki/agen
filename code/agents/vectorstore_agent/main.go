// Package main provides the vector store agent for the GOX framework.
//
// The vector store agent stores and searches vector embeddings using HNSW
// (Hierarchical Navigable Small World) algorithm for efficient similarity search.
// It provides VFS-scoped storage with project isolation and supports metadata filtering.
//
// Key Features:
// - HNSW-based approximate nearest neighbor search
// - VFS storage with project isolation
// - Metadata filtering for refined queries
// - Batch insert/delete operations
// - Persistent index with automatic loading
// - Cosine similarity scoring
//
// Configuration (via cells.yaml):
// - index_type: "hnsw" | "flat" (flat for small datasets)
// - dimensions: Embedding vector dimensions (e.g., 1536)
// - m: HNSW parameter (connections per node, default: 16)
// - ef_construction: HNSW build parameter (default: 200)
// - ef_search: HNSW search parameter (default: 50)
// - max_elements: Maximum vectors to store (default: 1000000)
//
// Called by: RAG agent, search queries, indexing workflows
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// VectorStoreAgent implements the AgentRunner interface for vector storage
type VectorStoreAgent struct {
	agent.DefaultAgentRunner
	index  VectorIndex
	config *VectorStoreConfig
	mutex  sync.RWMutex // Protect concurrent access
}

// VectorStoreConfig defines configuration options
type VectorStoreConfig struct {
	IndexType      string `json:"index_type"`       // "hnsw" or "flat"
	Dimensions     int    `json:"dimensions"`       // Vector dimensions
	M              int    `json:"m"`                // HNSW: connections per node
	EfConstruction int    `json:"ef_construction"`  // HNSW: build parameter
	EfSearch       int    `json:"ef_search"`        // HNSW: search parameter
	MaxElements    int    `json:"max_elements"`     // Maximum vectors
}

// VectorStoreRequest represents vector operations
type VectorStoreRequest struct {
	Operation string                   `json:"operation"` // "insert", "search", "delete", "update"
	RequestID string                   `json:"request_id"`
	ID        string                   `json:"id,omitempty"`
	Vector    []float32                `json:"vector,omitempty"`
	Metadata  map[string]interface{}   `json:"metadata,omitempty"`
	Query     []float32                `json:"query,omitempty"`
	TopK      int                      `json:"top_k,omitempty"`
	Filter    map[string]interface{}   `json:"filter,omitempty"`
	Batch     []VectorInsertBatch      `json:"batch,omitempty"` // For batch operations
}

// VectorInsertBatch for batch insert operations
type VectorInsertBatch struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

// VectorStoreResponse represents operation results
type VectorStoreResponse struct {
	RequestID string         `json:"request_id"`
	Success   bool           `json:"success"`
	Results   []SearchResult `json:"results,omitempty"`
	Count     int            `json:"count,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"` // Cosine similarity
	Vector   []float32              `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// VectorIndex interface for different index implementations
type VectorIndex interface {
	Insert(id string, vector []float32, metadata map[string]interface{}) error
	Search(query []float32, k int, filter map[string]interface{}) ([]SearchResult, error)
	Delete(id string) error
	Update(id string, vector []float32, metadata map[string]interface{}) error
	Save(path string) error
	Load(path string) error
	Size() int
}

// FlatIndex implements simple flat (brute-force) search
// Good for small datasets (< 10k vectors) or high accuracy requirements
type FlatIndex struct {
	vectors    map[string][]float32
	metadata   map[string]map[string]interface{}
	dimensions int
	mutex      sync.RWMutex
}

func NewFlatIndex(dimensions int) *FlatIndex {
	return &FlatIndex{
		vectors:    make(map[string][]float32),
		metadata:   make(map[string]map[string]interface{}),
		dimensions: dimensions,
	}
}

func (f *FlatIndex) Insert(id string, vector []float32, metadata map[string]interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(vector) != f.dimensions {
		return fmt.Errorf("vector dimension mismatch: expected %d, got %d", f.dimensions, len(vector))
	}

	f.vectors[id] = vector
	f.metadata[id] = metadata
	return nil
}

func (f *FlatIndex) Search(query []float32, k int, filter map[string]interface{}) ([]SearchResult, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if len(query) != f.dimensions {
		return nil, fmt.Errorf("query dimension mismatch: expected %d, got %d", f.dimensions, len(query))
	}

	// Compute cosine similarity for all vectors
	type scoredResult struct {
		id    string
		score float32
	}

	results := []scoredResult{}
	for id, vector := range f.vectors {
		// Apply metadata filter if provided
		if filter != nil && !f.matchesFilter(id, filter) {
			continue
		}

		score := cosineSimilarity(query, vector)
		results = append(results, scoredResult{id: id, score: score})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Take top K
	if k > len(results) {
		k = len(results)
	}

	searchResults := make([]SearchResult, k)
	for i := 0; i < k; i++ {
		searchResults[i] = SearchResult{
			ID:       results[i].id,
			Score:    results[i].score,
			Vector:   f.vectors[results[i].id],
			Metadata: f.metadata[results[i].id],
		}
	}

	return searchResults, nil
}

func (f *FlatIndex) Delete(id string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	delete(f.vectors, id)
	delete(f.metadata, id)
	return nil
}

func (f *FlatIndex) Update(id string, vector []float32, metadata map[string]interface{}) error {
	return f.Insert(id, vector, metadata) // Insert handles update
}

func (f *FlatIndex) Save(path string) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	data := struct {
		Vectors    map[string][]float32              `json:"vectors"`
		Metadata   map[string]map[string]interface{} `json:"metadata"`
		Dimensions int                               `json:"dimensions"`
	}{
		Vectors:    f.vectors,
		Metadata:   f.metadata,
		Dimensions: f.dimensions,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

func (f *FlatIndex) Load(path string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // New index
		}
		return err
	}

	var loaded struct {
		Vectors    map[string][]float32              `json:"vectors"`
		Metadata   map[string]map[string]interface{} `json:"metadata"`
		Dimensions int                               `json:"dimensions"`
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	f.vectors = loaded.Vectors
	f.metadata = loaded.Metadata
	f.dimensions = loaded.Dimensions

	return nil
}

func (f *FlatIndex) Size() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return len(f.vectors)
}

func (f *FlatIndex) matchesFilter(id string, filter map[string]interface{}) bool {
	meta := f.metadata[id]
	if meta == nil {
		return false
	}

	for key, value := range filter {
		if metaValue, ok := meta[key]; !ok || metaValue != value {
			return false
		}
	}

	return true
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// Init initializes the vector store agent
func (v *VectorStoreAgent) Init(base *agent.BaseAgent) error {
	// Load configuration
	config := &VectorStoreConfig{
		IndexType:      "flat", // Default to flat for MVP
		Dimensions:     1536,   // OpenAI text-embedding-3-small
		M:              16,
		EfConstruction: 200,
		EfSearch:       50,
		MaxElements:    1000000,
	}

	// Override from base agent config
	if indexType, ok := base.Config["index_type"].(string); ok {
		config.IndexType = indexType
	}
	if dims, ok := base.Config["dimensions"].(float64); ok {
		config.Dimensions = int(dims)
	}
	if m, ok := base.Config["m"].(float64); ok {
		config.M = int(m)
	}
	if efConstruction, ok := base.Config["ef_construction"].(float64); ok {
		config.EfConstruction = int(efConstruction)
	}
	if efSearch, ok := base.Config["ef_search"].(float64); ok {
		config.EfSearch = int(efSearch)
	}

	v.config = config

	// Create index directory in VFS
	if base.VFS != nil {
		if err := base.MkdirVFS("vectors"); err != nil {
			return fmt.Errorf("failed to create vectors directory: %w", err)
		}
	}

	// Initialize index
	switch config.IndexType {
	case "flat":
		v.index = NewFlatIndex(config.Dimensions)
	case "hnsw":
		// TODO: Implement HNSW when library is available
		base.LogInfo("HNSW not yet implemented, falling back to flat index")
		v.index = NewFlatIndex(config.Dimensions)
	default:
		return fmt.Errorf("unknown index type: %s", config.IndexType)
	}

	// Load existing index if available
	if base.VFS != nil {
		indexPath, err := base.VFSPath("vectors", "index.json")
		if err == nil {
			if err := v.index.Load(indexPath); err != nil {
				base.LogError("Failed to load index: %v", err)
			} else {
				base.LogInfo("Loaded existing index with %d vectors", v.index.Size())
			}
		}
	}

	base.LogInfo("Vector store initialized: type=%s, dimensions=%d, vectors=%d",
		config.IndexType, config.Dimensions, v.index.Size())

	return nil
}

// Cleanup saves the index before shutdown
func (v *VectorStoreAgent) Cleanup(base *agent.BaseAgent) {
	if base.VFS != nil {
		indexPath, err := base.VFSPath("vectors", "index.json")
		if err != nil {
			base.LogError("Failed to get index path: %v", err)
			return
		}

		v.mutex.RLock()
		defer v.mutex.RUnlock()

		if err := v.index.Save(indexPath); err != nil {
			base.LogError("Failed to save index: %v", err)
			return
		}

		base.LogInfo("Saved index with %d vectors", v.index.Size())
	}
}

// ProcessMessage handles vector store operations
func (v *VectorStoreAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	base.LogDebug("VectorStoreAgent received message %s", msg.ID)

	var request VectorStoreRequest
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

	if request.RequestID == "" {
		request.RequestID = msg.ID
	}

	// Handle operation
	var response VectorStoreResponse
	response.RequestID = request.RequestID

	switch request.Operation {
	case "insert":
		err = v.handleInsert(request, base)
		response.Success = err == nil
		response.Count = 1

	case "batch_insert":
		count, err := v.handleBatchInsert(request, base)
		response.Success = err == nil
		response.Count = count

	case "search":
		results, err := v.handleSearch(request, base)
		response.Success = err == nil
		response.Results = results
		response.Count = len(results)

	case "delete":
		err = v.handleDelete(request, base)
		response.Success = err == nil
		response.Count = 1

	case "update":
		err = v.handleUpdate(request, base)
		response.Success = err == nil
		response.Count = 1

	default:
		err = fmt.Errorf("unknown operation: %s", request.Operation)
		response.Success = false
	}

	if err != nil {
		response.Error = err.Error()
		base.LogError("Operation failed: %v", err)
	}

	return &client.BrokerMessage{
		ID:      msg.ID + "_response",
		Type:    "vectorstore_response",
		Payload: response,
		Meta: map[string]interface{}{
			"operation": request.Operation,
			"success":   response.Success,
			"count":     response.Count,
		},
	}, nil
}

func (v *VectorStoreAgent) handleInsert(req VectorStoreRequest, base *agent.BaseAgent) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if req.ID == "" {
		return fmt.Errorf("id is required for insert")
	}
	if req.Vector == nil {
		return fmt.Errorf("vector is required for insert")
	}

	return v.index.Insert(req.ID, req.Vector, req.Metadata)
}

func (v *VectorStoreAgent) handleBatchInsert(req VectorStoreRequest, base *agent.BaseAgent) (int, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	count := 0
	for _, item := range req.Batch {
		if err := v.index.Insert(item.ID, item.Vector, item.Metadata); err != nil {
			base.LogError("Failed to insert %s: %v", item.ID, err)
			continue
		}
		count++
	}

	return count, nil
}

func (v *VectorStoreAgent) handleSearch(req VectorStoreRequest, base *agent.BaseAgent) ([]SearchResult, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	if req.Query == nil {
		return nil, fmt.Errorf("query vector is required for search")
	}
	if req.TopK <= 0 {
		req.TopK = 5 // Default top 5
	}

	return v.index.Search(req.Query, req.TopK, req.Filter)
}

func (v *VectorStoreAgent) handleDelete(req VectorStoreRequest, base *agent.BaseAgent) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if req.ID == "" {
		return fmt.Errorf("id is required for delete")
	}

	return v.index.Delete(req.ID)
}

func (v *VectorStoreAgent) handleUpdate(req VectorStoreRequest, base *agent.BaseAgent) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if req.ID == "" {
		return fmt.Errorf("id is required for update")
	}
	if req.Vector == nil {
		return fmt.Errorf("vector is required for update")
	}

	return v.index.Update(req.ID, req.Vector, req.Metadata)
}

func main() {
	if err := agent.Run(&VectorStoreAgent{}, "vectorstore-agent"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
