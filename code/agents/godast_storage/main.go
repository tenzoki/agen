package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tenzoki/agen/omni/public/omnistore"
	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// GodastStorageAgent implements the AgentRunner interface for storage operations
type GodastStorageAgent struct {
	agent.DefaultAgentRunner
	omniStore omnistore.OmniStore
	config    *StorageConfig
}

// StorageConfig defines configuration options for the storage agent
type StorageConfig struct {
	DataPath       string `yaml:"data_path"`
	MaxFileSize    int64  `yaml:"max_file_size"`
	EnableKV       bool   `yaml:"enable_kv"`
	EnableGraph    bool   `yaml:"enable_graph"`
	EnableFiles    bool   `yaml:"enable_files"`
	EnableFulltext bool   `yaml:"enable_fulltext"`
}

// StorageRequest represents a storage operation request
type StorageRequest struct {
	Operation   string                 `json:"operation"`
	Key         string                 `json:"key,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Query       string                 `json:"query,omitempty"`
	FileData    []byte                 `json:"file_data,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	SearchTerms string                 `json:"search_terms,omitempty"`
	RequestID   string                 `json:"request_id"`
}

// StorageResponse represents a storage operation response
type StorageResponse struct {
	RequestID string      `json:"request_id"`
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	Count     int         `json:"count,omitempty"`
}

// Init initializes the Godast storage agent with VFS support
func (g *GodastStorageAgent) Init(base *agent.BaseAgent) error {
	// Load configuration from base agent
	config := &StorageConfig{
		MaxFileSize:    104857600, // 100MB
		EnableKV:       true,
		EnableGraph:    true,
		EnableFiles:    true,
		EnableFulltext: true,
	}

	g.config = config

	// Use VFS for storage path if available
	var storagePath string
	if base.VFS != nil {
		// Create storage directory within VFS
		if err := base.MkdirVFS("storage"); err != nil {
			return fmt.Errorf("failed to create storage directory in VFS: %w", err)
		}

		// Get absolute path from VFS
		var err error
		storagePath, err = base.VFSPath("storage")
		if err != nil {
			return fmt.Errorf("failed to get VFS storage path: %w", err)
		}

		base.LogInfo("Using VFS for storage (project: %s)", base.ProjectID)
	} else {
		// Fallback to traditional path if VFS not enabled
		storagePath = "/tmp/gox-storage"
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
		base.LogInfo("VFS not enabled, using traditional path")
	}

	config.DataPath = storagePath

	// Initialize OmniStore with VFS-scoped path
	omniStore, err := omnistore.NewOmniStoreWithDefaults(storagePath)
	if err != nil {
		return fmt.Errorf("failed to initialize OmniStore: %w", err)
	}
	g.omniStore = omniStore

	base.LogInfo("Godast OmniStore initialized at %s", storagePath)
	base.LogInfo("Storage capabilities: KV=%v, Graph=%v, Files=%v, FullText=%v",
		g.config.EnableKV, g.config.EnableGraph, g.config.EnableFiles, g.config.EnableFulltext)

	return nil
}

// ProcessMessage handles storage operation requests
func (g *GodastStorageAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	base.LogDebug("GodastStorageAgent received message %s of type %s", msg.ID, msg.Type)

	var request StorageRequest
	var err error

	// Handle different payload types
	switch payload := msg.Payload.(type) {
	case []byte:
		err = json.Unmarshal(payload, &request)
	case string:
		err = json.Unmarshal([]byte(payload), &request)
	default:
		// Try to marshal and unmarshal to handle map[string]interface{} type
		payloadBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			errorResponse := g.createErrorResponse("", fmt.Sprintf("invalid request format: %v", marshalErr))
			responseBytes, _ := json.Marshal(errorResponse)
			return &client.BrokerMessage{
				ID:      msg.ID + "_error",
				Target:  base.GetEgress(),
				Type:    "storage_response",
				Payload: responseBytes,
				Meta: map[string]interface{}{
					"error":   true,
					"message": marshalErr.Error(),
				},
			}, nil
		}
		err = json.Unmarshal(payloadBytes, &request)
	}

	if err != nil {
		errorResponse := g.createErrorResponse("", fmt.Sprintf("invalid request format: %v", err))
		responseBytes, _ := json.Marshal(errorResponse)
		return &client.BrokerMessage{
			ID:      msg.ID + "_error",
			Target:  base.GetEgress(),
			Type:    "storage_response",
			Payload: responseBytes,
			Meta: map[string]interface{}{
				"error":   true,
				"message": err.Error(),
			},
		}, nil
	}

	base.LogDebug("Processing storage operation: %s (request_id: %s)", request.Operation, request.RequestID)

	var response StorageResponse
	response.RequestID = request.RequestID

	switch request.Operation {
	case "kv_set":
		response = g.handleKVSet(request, base)
	case "kv_get":
		response = g.handleKVGet(request, base)
	case "kv_delete":
		response = g.handleKVDelete(request, base)
	case "kv_exists":
		response = g.handleKVExists(request, base)
	case "graph_create_vertex":
		response = g.handleGraphCreateVertex(request, base)
	case "graph_create_edge":
		response = g.handleGraphCreateEdge(request, base)
	case "graph_query":
		response = g.handleGraphQuery(request, base)
	case "file_store":
		response = g.handleFileStore(request, base)
	case "file_retrieve":
		response = g.handleFileRetrieve(request, base)
	case "fulltext_index":
		response = g.handleFulltextIndex(request, base)
	case "fulltext_search":
		response = g.handleFulltextSearch(request, base)
	default:
		response = g.createErrorResponse(request.RequestID, fmt.Sprintf("unknown operation: %s", request.Operation))
	}

	base.LogDebug("Storage operation completed: %s (success: %v)", request.Operation, response.Success)

	responseBytes, _ := json.Marshal(response)
	return &client.BrokerMessage{
		ID:      msg.ID + "_response",
		Target:  base.GetEgress(),
		Type:    "storage_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"request_id": request.RequestID,
			"operation":  request.Operation,
			"success":    response.Success,
		},
	}, nil
}

// KV Operations
func (g *GodastStorageAgent) handleKVSet(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableKV || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "KV operations are disabled")
	}

	// Convert value to bytes for storage
	valueBytes, err := json.Marshal(request.Value)
	if err != nil {
		base.LogError("KV Set failed to marshal value for key %s: %v", request.Key, err)
		return g.createErrorResponse(request.RequestID, fmt.Sprintf("failed to marshal value: %v", err))
	}

	err = g.omniStore.KV().Set(request.Key, valueBytes)
	if err != nil {
		base.LogError("KV Set failed for key %s: %v", request.Key, err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	base.LogDebug("KV Set successful: %s", request.Key)
	return StorageResponse{RequestID: request.RequestID, Success: true}
}

func (g *GodastStorageAgent) handleKVGet(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableKV || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "KV operations are disabled")
	}

	valueBytes, err := g.omniStore.KV().Get(request.Key)
	if err != nil {
		base.LogError("KV Get failed for key %s: %v", request.Key, err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	// Unmarshal the stored value
	var value interface{}
	if err := json.Unmarshal(valueBytes, &value); err != nil {
		base.LogError("KV Get failed to unmarshal value for key %s: %v", request.Key, err)
		return g.createErrorResponse(request.RequestID, fmt.Sprintf("failed to unmarshal value: %v", err))
	}

	base.LogDebug("KV Get successful: %s", request.Key)
	return StorageResponse{RequestID: request.RequestID, Success: true, Result: value}
}

func (g *GodastStorageAgent) handleKVDelete(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableKV || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "KV operations are disabled")
	}

	err := g.omniStore.KV().Delete(request.Key)
	if err != nil {
		base.LogError("KV Delete failed for key %s: %v", request.Key, err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	base.LogDebug("KV Delete successful: %s", request.Key)
	return StorageResponse{RequestID: request.RequestID, Success: true}
}

func (g *GodastStorageAgent) handleKVExists(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableKV || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "KV operations are disabled")
	}

	_, err := g.omniStore.KV().Get(request.Key)
	exists := err == nil
	base.LogDebug("KV Exists check: %s = %v", request.Key, exists)
	return StorageResponse{RequestID: request.RequestID, Success: true, Result: exists}
}

// Graph Operations
func (g *GodastStorageAgent) handleGraphCreateVertex(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableGraph || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "Graph operations are disabled")
	}

	properties, ok := request.Value.(map[string]interface{})
	if !ok {
		return g.createErrorResponse(request.RequestID, "invalid vertex properties")
	}

	// For now, create a simple vertex entry - will need to check actual Graph interface
	// This is a placeholder until we can verify the exact Graph API
	vertexData := map[string]interface{}{
		"id":         request.Key,
		"properties": properties,
		"type":       "vertex",
	}

	// Store vertex data in KV store for now
	vertexKey := fmt.Sprintf("graph:vertex:%s", request.Key)
	vertexBytes, _ := json.Marshal(vertexData)
	err := g.omniStore.KV().Set(vertexKey, vertexBytes)
	if err != nil {
		base.LogError("Graph CreateVertex failed: %v", err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	base.LogDebug("Graph vertex created: %s", request.Key)
	return StorageResponse{RequestID: request.RequestID, Success: true, Result: request.Key}
}

func (g *GodastStorageAgent) handleGraphCreateEdge(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableGraph || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "Graph operations are disabled")
	}

	// Extract edge parameters from request.Value
	edgeData, ok := request.Value.(map[string]interface{})
	if !ok {
		return g.createErrorResponse(request.RequestID, "invalid edge data")
	}

	from, ok := edgeData["from"].(string)
	if !ok {
		return g.createErrorResponse(request.RequestID, "missing or invalid 'from' vertex")
	}

	to, ok := edgeData["to"].(string)
	if !ok {
		return g.createErrorResponse(request.RequestID, "missing or invalid 'to' vertex")
	}

	label, ok := edgeData["label"].(string)
	if !ok {
		return g.createErrorResponse(request.RequestID, "missing or invalid edge 'label'")
	}

	// Store edge data in KV store for now
	edgeKey := fmt.Sprintf("graph:edge:%s:%s:%s", from, to, label)
	edgeBytes, _ := json.Marshal(edgeData)
	err := g.omniStore.KV().Set(edgeKey, edgeBytes)
	if err != nil {
		base.LogError("Graph CreateEdge failed: %v", err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	base.LogDebug("Graph edge created: %s -> %s (%s)", from, to, label)
	return StorageResponse{RequestID: request.RequestID, Success: true}
}

func (g *GodastStorageAgent) handleGraphQuery(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableGraph || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "Graph operations are disabled")
	}

	// For now, return a simple query result - real implementation would use g.omniStore.Query()
	results := []interface{}{
		map[string]interface{}{
			"message": "Graph query functionality not fully implemented",
			"query":   request.Query,
		},
	}

	base.LogDebug("Graph query executed, returned %d results", len(results))
	return StorageResponse{
		RequestID: request.RequestID,
		Success:   true,
		Result:    results,
		Count:     len(results),
	}
}

// File Operations
func (g *GodastStorageAgent) handleFileStore(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableFiles || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "File operations are disabled")
	}

	if len(request.FileData) == 0 {
		return g.createErrorResponse(request.RequestID, "no file data provided")
	}

	if g.config.MaxFileSize > 0 && int64(len(request.FileData)) > g.config.MaxFileSize {
		return g.createErrorResponse(request.RequestID, fmt.Sprintf("file size exceeds maximum: %d > %d", len(request.FileData), g.config.MaxFileSize))
	}

	// Convert metadata to string map
	metadataStrings := make(map[string]string)
	if request.Metadata != nil {
		for k, v := range request.Metadata {
			metadataStrings[k] = fmt.Sprintf("%v", v)
		}
	}

	hash, err := g.omniStore.Files().Store(request.FileData, metadataStrings)
	if err != nil {
		base.LogError("File Store failed: %v", err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	base.LogDebug("File stored with hash: %s (size: %d bytes)", hash, len(request.FileData))
	return StorageResponse{RequestID: request.RequestID, Success: true, Result: hash}
}

func (g *GodastStorageAgent) handleFileRetrieve(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableFiles || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "File operations are disabled")
	}

	data, _, err := g.omniStore.Files().Retrieve(request.Key)
	if err != nil {
		base.LogError("File Retrieve failed for hash %s: %v", request.Key, err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	base.LogDebug("File retrieved: %s (size: %d bytes)", request.Key, len(data))
	return StorageResponse{RequestID: request.RequestID, Success: true, Result: data}
}

// Full-text Operations
func (g *GodastStorageAgent) handleFulltextIndex(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableFulltext || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "Full-text operations are disabled")
	}

	content, ok := request.Value.(string)
	if !ok {
		return g.createErrorResponse(request.RequestID, "invalid content for indexing")
	}

	// Create document for indexing
	document := map[string]interface{}{
		"id":      request.Key,
		"content": content,
	}

	// Add metadata to document
	if request.Metadata != nil {
		for k, v := range request.Metadata {
			document[k] = v
		}
	}

	// Convert document to JSON string for indexing
	documentJSON, _ := json.Marshal(document)
	err := g.omniStore.Search().IndexDocument(request.Key, string(documentJSON), request.Metadata)
	if err != nil {
		base.LogError("Full-text Index failed: %v", err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	base.LogDebug("Content indexed: %s", request.Key)
	return StorageResponse{RequestID: request.RequestID, Success: true}
}

func (g *GodastStorageAgent) handleFulltextSearch(request StorageRequest, base *agent.BaseAgent) StorageResponse {
	if !g.config.EnableFulltext || g.omniStore == nil {
		return g.createErrorResponse(request.RequestID, "Full-text operations are disabled")
	}

	// Create search options
	searchOptions := &omnistore.SearchOptions{}

	results, err := g.omniStore.Search().Search(request.SearchTerms, searchOptions)
	if err != nil {
		base.LogError("Full-text Search failed: %v", err)
		return g.createErrorResponse(request.RequestID, err.Error())
	}

	// Extract results from SearchResult
	var resultList []interface{}
	if results != nil {
		// Convert search results to interface slice
		resultList = []interface{}{results}
	}

	base.LogDebug("Full-text search completed, returned %d results", len(resultList))
	return StorageResponse{
		RequestID: request.RequestID,
		Success:   true,
		Result:    resultList,
		Count:     len(resultList),
	}
}

func (g *GodastStorageAgent) createErrorResponse(requestID, errorMsg string) StorageResponse {
	return StorageResponse{
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
}

// Main function is now in main_service.go which handles both service and pipeline modes
