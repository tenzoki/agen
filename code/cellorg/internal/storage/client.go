package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/agen/cellorg/internal/client"
)

// HTTPClient provides HTTP client for storage service
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// KVSetRequest represents a KV set request
type KVSetRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// KVResponse represents a KV response
type KVResponse struct {
	Success bool        `json:"success,omitempty"`
	Key     string      `json:"key,omitempty"`
	Value   interface{} `json:"value,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Client provides broker-based storage client
type Client struct {
	agentID      string
	brokerClient *client.BrokerClient
	timeout      time.Duration
	responses    map[string]chan StorageResponse
	mu           sync.RWMutex
}

// StorageRequest represents a storage operation request
type StorageRequest struct {
	Operation   string                 `json:"operation"`
	Key         string                 `json:"key,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Query       string                 `json:"query,omitempty"`
	SearchTerms string                 `json:"search_terms,omitempty"`
	FileData    []byte                 `json:"file_data,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	RequestID   string                 `json:"request_id"`
}

// StorageResponse represents a storage operation response
type StorageResponse struct {
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id"`
}

// NewClient creates a new broker-based storage client
func NewClient(agentID string, brokerClient *client.BrokerClient) *Client {
	return &Client{
		agentID:      agentID,
		brokerClient: brokerClient,
		timeout:      30 * time.Second,
		responses:    make(map[string]chan StorageResponse),
	}
}

// NewHTTPClient creates a new HTTP storage client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTimeout sets the timeout for storage operations
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// KV Operations

// KVSet stores a key-value pair
func (c *Client) KVSet(key string, value interface{}) error {
	request := StorageRequest{
		Operation: "kv_set",
		Key:       key,
		Value:     value,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// KVGet retrieves a value by key
func (c *Client) KVGet(key string) (interface{}, error) {
	request := StorageRequest{
		Operation: "kv_get",
		Key:       key,
		RequestID: uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

// KVDelete removes a key-value pair
func (c *Client) KVDelete(key string) error {
	request := StorageRequest{
		Operation: "kv_delete",
		Key:       key,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// KVExists checks if a key exists
func (c *Client) KVExists(key string) (bool, error) {
	request := StorageRequest{
		Operation: "kv_exists",
		Key:       key,
		RequestID: uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return false, err
	}

	exists, ok := response.Result.(bool)
	if !ok {
		return false, fmt.Errorf("invalid response type for exists check")
	}

	return exists, nil
}

// Graph Operations

// CreateVertex creates a new graph vertex
func (c *Client) CreateVertex(label string, properties map[string]interface{}) (string, error) {
	request := StorageRequest{
		Operation: "graph_create_vertex",
		Key:       label,
		Value:     properties,
		RequestID: uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return "", err
	}

	vertexID, ok := response.Result.(string)
	if !ok {
		return "", fmt.Errorf("invalid response type for vertex creation")
	}

	return vertexID, nil
}

// CreateEdge creates an edge between two vertices
func (c *Client) CreateEdge(from, to, label string) error {
	edgeData := map[string]interface{}{
		"from":  from,
		"to":    to,
		"label": label,
	}

	request := StorageRequest{
		Operation: "graph_create_edge",
		Value:     edgeData,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// GraphQuery executes a graph query
func (c *Client) GraphQuery(query string) ([]interface{}, error) {
	request := StorageRequest{
		Operation: "graph_query",
		Query:     query,
		RequestID: uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return nil, err
	}

	results, ok := response.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response type for graph query")
	}

	return results, nil
}

// File Operations

// StoreFile stores file data and returns a hash
func (c *Client) StoreFile(data []byte, metadata map[string]interface{}) (string, error) {
	request := StorageRequest{
		Operation: "file_store",
		FileData:  data,
		Metadata:  metadata,
		RequestID: uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return "", err
	}

	hash, ok := response.Result.(string)
	if !ok {
		return "", fmt.Errorf("invalid response type for file store")
	}

	return hash, nil
}

// RetrieveFile retrieves file data by hash
func (c *Client) RetrieveFile(hash string) ([]byte, error) {
	request := StorageRequest{
		Operation: "file_retrieve",
		Key:       hash,
		RequestID: uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return nil, err
	}

	// Handle different response formats for file data
	switch data := response.Result.(type) {
	case []byte:
		return data, nil
	case string:
		return []byte(data), nil
	case []interface{}:
		// Convert slice of numbers to bytes
		bytes := make([]byte, len(data))
		for i, v := range data {
			if num, ok := v.(float64); ok {
				bytes[i] = byte(num)
			} else {
				return nil, fmt.Errorf("invalid byte data format")
			}
		}
		return bytes, nil
	default:
		return nil, fmt.Errorf("invalid response type for file retrieve: %T", data)
	}
}

// Full-text Operations

// IndexContent indexes content for full-text search
func (c *Client) IndexContent(id, content string, metadata map[string]interface{}) error {
	request := StorageRequest{
		Operation: "fulltext_index",
		Key:       id,
		Value:     content,
		Metadata:  metadata,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// SearchContent performs full-text search
func (c *Client) SearchContent(searchTerms string) ([]interface{}, error) {
	request := StorageRequest{
		Operation:   "fulltext_search",
		SearchTerms: searchTerms,
		RequestID:   uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return nil, err
	}

	results, ok := response.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response type for full-text search")
	}

	return results, nil
}

// Helper Methods

// HandleResponse processes incoming storage responses
// This should be called by agents when they receive storage_response messages
func (c *Client) HandleResponse(msg *client.BrokerMessage) error {
	if msg.Type != "storage_response" {
		return nil // Not a storage response, ignore
	}

	var response StorageResponse
	var err error

	// Handle different payload types
	switch payload := msg.Payload.(type) {
	case []byte:
		err = json.Unmarshal(payload, &response)
	case string:
		err = json.Unmarshal([]byte(payload), &response)
	default:
		// Try to marshal and unmarshal
		payloadBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal response payload: %w", marshalErr)
		}
		err = json.Unmarshal(payloadBytes, &response)
	}

	if err != nil {
		return fmt.Errorf("failed to parse storage response: %w", err)
	}

	c.mu.RLock()
	responseChan, exists := c.responses[response.RequestID]
	c.mu.RUnlock()

	if exists {
		select {
		case responseChan <- response:
			// Successfully delivered response
		case <-time.After(1 * time.Second):
			// Channel blocked, cleanup
			c.mu.Lock()
			delete(c.responses, response.RequestID)
			close(responseChan)
			c.mu.Unlock()
		}
	}

	return nil
}

// sendStorageRequest sends a storage request without waiting for response
func (c *Client) sendStorageRequest(request StorageRequest) error {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	msg := &client.BrokerMessage{
		ID:      uuid.New().String(),
		Target:  "pub:storage-requests",
		Type:    "storage_request",
		Payload: requestBytes,
		Meta: map[string]interface{}{
			"sender":     c.agentID,
			"request_id": request.RequestID,
		},
	}

	return c.brokerClient.Publish("storage-requests", *msg)
}

// sendStorageRequestWithResponse sends a storage request and waits for response
func (c *Client) sendStorageRequestWithResponse(request StorageRequest) (*StorageResponse, error) {
	// Create response channel
	responseChan := make(chan StorageResponse, 1)
	c.mu.Lock()
	c.responses[request.RequestID] = responseChan
	c.mu.Unlock()

	// Clean up on exit
	defer func() {
		c.mu.Lock()
		delete(c.responses, request.RequestID)
		close(responseChan)
		c.mu.Unlock()
	}()

	// Send request
	err := c.sendStorageRequest(request)
	if err != nil {
		return nil, err
	}

	// Wait for response with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	select {
	case response := <-responseChan:
		if !response.Success {
			return nil, fmt.Errorf("storage operation failed: %s", response.Error)
		}
		return &response, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("storage operation timed out after %v", c.timeout)
	}
}

// Batch Operations for Parallel Processing

// BatchVertex represents a vertex for batch creation
type BatchVertex struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties"`
}

// BatchEdge represents an edge for batch creation
type BatchEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

// EdgeSpec represents an edge specification for deletion
type EdgeSpec struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

// TransactionID represents a transaction identifier
type TransactionID string

// BatchCreateVertices creates multiple vertices in a single operation
func (c *Client) BatchCreateVertices(vertices []BatchVertex) error {
	// Group vertices by label for optimized creation
	labelGroups := make(map[string][]BatchVertex)

	for _, vertex := range vertices {
		labelGroups[vertex.Label] = append(labelGroups[vertex.Label], vertex)
	}

	// Create vertices in batches by label
	for label, verts := range labelGroups {
		request := StorageRequest{
			Operation: "batch_create_vertices",
			Value: map[string]interface{}{
				"label":    label,
				"vertices": verts,
			},
			RequestID: uuid.New().String(),
		}

		if err := c.sendStorageRequest(request); err != nil {
			return fmt.Errorf("failed to batch create vertices for label %s: %w", label, err)
		}
	}

	return nil
}

// BatchCreateEdges creates multiple edges in a single operation
func (c *Client) BatchCreateEdges(edges []BatchEdge) error {
	request := StorageRequest{
		Operation: "batch_create_edges",
		Value:     edges,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// BatchUpdateVertexProperties updates properties for multiple vertices
func (c *Client) BatchUpdateVertexProperties(vertexIDs []string, properties map[string]interface{}) error {
	request := StorageRequest{
		Operation: "batch_update_vertices",
		Value: map[string]interface{}{
			"vertex_ids": vertexIDs,
			"properties": properties,
		},
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// BatchDeleteVertices deletes multiple vertices
func (c *Client) BatchDeleteVertices(vertexIDs []string) error {
	request := StorageRequest{
		Operation: "batch_delete_vertices",
		Value:     vertexIDs,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// BatchDeleteEdges deletes multiple edges
func (c *Client) BatchDeleteEdges(edgeSpecs []EdgeSpec) error {
	request := StorageRequest{
		Operation: "batch_delete_edges",
		Value:     edgeSpecs,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// ParallelGraphQuery executes multiple graph queries concurrently
func (c *Client) ParallelGraphQuery(queries []string) ([][]interface{}, error) {
	results := make([][]interface{}, len(queries))
	errors := make([]error, len(queries))

	var wg sync.WaitGroup
	for i, query := range queries {
		wg.Add(1)
		go func(index int, q string) {
			defer wg.Done()
			result, err := c.GraphQuery(q)
			results[index] = result
			errors[index] = err
		}(i, query)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// BeginTransaction starts a new transaction
func (c *Client) BeginTransaction() (TransactionID, error) {
	request := StorageRequest{
		Operation: "begin_transaction",
		RequestID: uuid.New().String(),
	}

	response, err := c.sendStorageRequestWithResponse(request)
	if err != nil {
		return "", err
	}

	txID, ok := response.Result.(string)
	if !ok {
		return "", fmt.Errorf("invalid response type for transaction begin")
	}

	return TransactionID(txID), nil
}

// CommitTransaction commits a transaction
func (c *Client) CommitTransaction(txID TransactionID) error {
	request := StorageRequest{
		Operation: "commit_transaction",
		Key:       string(txID),
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// RollbackTransaction rolls back a transaction
func (c *Client) RollbackTransaction(txID TransactionID) error {
	request := StorageRequest{
		Operation: "rollback_transaction",
		Key:       string(txID),
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}

// UpdateVertexProperties updates properties of a single vertex
func (c *Client) UpdateVertexProperties(vertexID string, properties map[string]interface{}) error {
	request := StorageRequest{
		Operation: "update_vertex_properties",
		Key:       vertexID,
		Value:     properties,
		RequestID: uuid.New().String(),
	}

	return c.sendStorageRequest(request)
}
