package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tenzoki/agen/omni/public/omnistore"
	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// AnonymizationStoreAgent provides persistent storage for anonymization mappings
type AnonymizationStoreAgent struct {
	agent.DefaultAgentRunner
	omniStore omnistore.OmniStore
	config    *StoreConfig
}

// StoreConfig holds configuration for the anonymization store
type StoreConfig struct {
	DataPath    string `yaml:"data_path"`
	MaxFileSize int64  `yaml:"max_file_size"`
	EnableDebug bool   `yaml:"enable_debug"`
}

// StorageRequest represents a storage operation request
type StorageRequest struct {
	Operation string                 `json:"operation"` // set, get, reverse, list, delete
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value,omitempty"`
	ProjectID string                 `json:"project_id"`
	RequestID string                 `json:"request_id"`
}

// StorageResponse represents a storage operation response
type StorageResponse struct {
	Success   bool                   `json:"success"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	RequestID string                 `json:"request_id"`
}

// MappingRecord represents an anonymization mapping
type MappingRecord struct {
	Pseudonym       string    `json:"pseudonym"`
	Canonical       string    `json:"canonical"`
	EntityType      string    `json:"entity_type"`
	CreatedAt       time.Time `json:"created_at"`
	PipelineVersion string    `json:"pipeline_version"`
}

// ReverseMappingRecord represents a reverse lookup record
type ReverseMappingRecord struct {
	Original   string `json:"original"`
	Canonical  string `json:"canonical"`
	EntityType string `json:"entity_type"`
}

// Init initializes the anonymization store agent
func (a *AnonymizationStoreAgent) Init(base *agent.BaseAgent) error {
	// Load configuration
	config := &StoreConfig{
		DataPath:    base.GetConfigString("data_path", "/tmp/gox-anonymization-store"),
		MaxFileSize: int64(base.GetConfigInt("max_file_size", 104857600)), // 100MB
		EnableDebug: base.GetConfigBool("enable_debug", false),
	}
	a.config = config

	// Initialize OmniStore with bbolt backend
	omniStore, err := omnistore.NewOmniStoreWithDefaults(config.DataPath)
	if err != nil {
		return fmt.Errorf("failed to initialize OmniStore: %w", err)
	}
	a.omniStore = omniStore

	base.LogInfo("Anonymization store initialized at %s", config.DataPath)
	base.LogInfo("Storage capabilities: KV store with bbolt backend")

	return nil
}

// ProcessMessage handles storage operation requests
func (a *AnonymizationStoreAgent) ProcessMessage(
	msg *client.BrokerMessage,
	base *agent.BaseAgent,
) (*client.BrokerMessage, error) {
	// Parse storage request
	var req StorageRequest
	payloadBytes, ok := msg.Payload.([]byte)
	if !ok {
		return a.errorResponse("", "invalid payload type"), nil
	}
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return a.errorResponse("", fmt.Sprintf("invalid request format: %v", err)), nil
	}

	if a.config.EnableDebug {
		base.LogInfo("Storage operation: %s for key: %s", req.Operation, req.Key)
	}

	// Route to appropriate handler
	var response StorageResponse
	var err error

	switch req.Operation {
	case "set":
		response, err = a.handleSet(req, base)
	case "get":
		response, err = a.handleGet(req, base)
	case "reverse":
		response, err = a.handleReverse(req, base)
	case "list":
		response, err = a.handleList(req, base)
	case "delete":
		response, err = a.handleDelete(req, base)
	default:
		response = StorageResponse{
			Success:   false,
			Error:     fmt.Sprintf("unknown operation: %s", req.Operation),
			RequestID: req.RequestID,
		}
	}

	if err != nil {
		base.LogError("Storage operation failed: %v", err)
		response.Success = false
		response.Error = err.Error()
	}

	response.RequestID = req.RequestID

	// Serialize response
	payload, err := json.Marshal(response)
	if err != nil {
		return a.errorResponse(req.RequestID, fmt.Sprintf("failed to serialize response: %v", err)), nil
	}

	return &client.BrokerMessage{
		Payload: payload,
	}, nil
}

// handleSet stores a forward mapping (original → pseudonym)
func (a *AnonymizationStoreAgent) handleSet(req StorageRequest, base *agent.BaseAgent) (StorageResponse, error) {
	if req.Key == "" {
		return StorageResponse{Success: false, Error: "key is required"}, nil
	}

	if req.Value == nil {
		return StorageResponse{Success: false, Error: "value is required"}, nil
	}

	// Serialize value
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		return StorageResponse{Success: false, Error: fmt.Sprintf("failed to serialize value: %v", err)}, nil
	}

	// Store in KV
	if err := a.omniStore.KV().Set(req.Key, valueJSON); err != nil {
		return StorageResponse{Success: false, Error: fmt.Sprintf("failed to store: %v", err)}, nil
	}

	if a.config.EnableDebug {
		base.LogInfo("Stored mapping: %s", req.Key)
	}

	return StorageResponse{
		Success: true,
		Result: map[string]interface{}{
			"key": req.Key,
		},
	}, nil
}

// handleGet retrieves a forward mapping (original → pseudonym)
func (a *AnonymizationStoreAgent) handleGet(req StorageRequest, base *agent.BaseAgent) (StorageResponse, error) {
	if req.Key == "" {
		return StorageResponse{Success: false, Error: "key is required"}, nil
	}

	// Retrieve from KV
	valueBytes, err := a.omniStore.KV().Get(req.Key)
	if err != nil {
		// Not found is not an error - return empty result
		return StorageResponse{
			Success: false,
			Error:   "not found",
		}, nil
	}

	// Parse value JSON
	var result map[string]interface{}
	if err := json.Unmarshal(valueBytes, &result); err != nil {
		return StorageResponse{Success: false, Error: fmt.Sprintf("failed to parse stored value: %v", err)}, nil
	}

	if a.config.EnableDebug {
		base.LogInfo("Retrieved mapping: %s", req.Key)
	}

	return StorageResponse{
		Success: true,
		Result:  result,
	}, nil
}

// handleReverse performs reverse lookup (pseudonym → original)
func (a *AnonymizationStoreAgent) handleReverse(req StorageRequest, base *agent.BaseAgent) (StorageResponse, error) {
	if req.Key == "" {
		return StorageResponse{Success: false, Error: "key is required"}, nil
	}

	// Retrieve from KV
	valueBytes, err := a.omniStore.KV().Get(req.Key)
	if err != nil {
		return StorageResponse{
			Success: false,
			Error:   "not found",
		}, nil
	}

	// Parse value JSON
	var result map[string]interface{}
	if err := json.Unmarshal(valueBytes, &result); err != nil {
		return StorageResponse{Success: false, Error: fmt.Sprintf("failed to parse stored value: %v", err)}, nil
	}

	if a.config.EnableDebug {
		base.LogInfo("Reverse lookup: %s", req.Key)
	}

	return StorageResponse{
		Success: true,
		Result:  result,
	}, nil
}

// handleList lists all mappings for a project (prefix scan)
func (a *AnonymizationStoreAgent) handleList(req StorageRequest, base *agent.BaseAgent) (StorageResponse, error) {
	prefix := req.Key // Use key as prefix

	if prefix == "" {
		return StorageResponse{Success: false, Error: "prefix (key) is required for list operation"}, nil
	}

	// Get all keys with prefix using new ListKVKeys method
	// Limit to 1000 keys to prevent excessive memory usage
	keys, err := a.omniStore.ListKVKeys(prefix, 1000)
	if err != nil {
		return StorageResponse{Success: false, Error: fmt.Sprintf("failed to list keys: %v", err)}, nil
	}

	// Retrieve all values
	mappings := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		valueBytes, err := a.omniStore.KV().Get(key)
		if err != nil {
			continue // Skip errors
		}

		var mapping map[string]interface{}
		if err := json.Unmarshal(valueBytes, &mapping); err != nil {
			continue // Skip parse errors
		}

		mapping["_key"] = key // Add key to result
		mappings = append(mappings, mapping)
	}

	if a.config.EnableDebug {
		base.LogInfo("Listed %d mappings with prefix: %s", len(mappings), prefix)
	}

	return StorageResponse{
		Success: true,
		Result: map[string]interface{}{
			"count":    len(mappings),
			"mappings": mappings,
		},
	}, nil
}

// handleDelete removes a mapping (soft delete - preserves audit trail)
func (a *AnonymizationStoreAgent) handleDelete(req StorageRequest, base *agent.BaseAgent) (StorageResponse, error) {
	if req.Key == "" {
		return StorageResponse{Success: false, Error: "key is required"}, nil
	}

	// For audit purposes, we mark as deleted rather than removing
	// Get existing value
	valueBytes, err := a.omniStore.KV().Get(req.Key)
	if err != nil {
		return StorageResponse{Success: false, Error: "not found"}, nil
	}

	// Parse and add deleted flag
	var record map[string]interface{}
	if err := json.Unmarshal(valueBytes, &record); err != nil {
		return StorageResponse{Success: false, Error: fmt.Sprintf("failed to parse stored value: %v", err)}, nil
	}

	record["deleted_at"] = time.Now().Format(time.RFC3339)
	record["deleted"] = true

	// Store updated record
	updatedJSON, _ := json.Marshal(record)
	if err := a.omniStore.KV().Set(req.Key, updatedJSON); err != nil {
		return StorageResponse{Success: false, Error: fmt.Sprintf("failed to mark as deleted: %v", err)}, nil
	}

	if a.config.EnableDebug {
		base.LogInfo("Marked as deleted: %s", req.Key)
	}

	return StorageResponse{
		Success: true,
		Result: map[string]interface{}{
			"key":     req.Key,
			"deleted": true,
		},
	}, nil
}

// errorResponse creates an error response message
func (a *AnonymizationStoreAgent) errorResponse(requestID, errorMsg string) *client.BrokerMessage {
	resp := StorageResponse{
		Success:   false,
		Error:     errorMsg,
		RequestID: requestID,
	}
	payload, _ := json.Marshal(resp)
	return &client.BrokerMessage{
		Payload: payload,
	}
}

// Cleanup releases resources
func (a *AnonymizationStoreAgent) Cleanup(base *agent.BaseAgent) {
	if a.omniStore != nil {
		a.omniStore.Close()
		base.LogInfo("Anonymization store closed")
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	agent.Run(&AnonymizationStoreAgent{}, "anonymization-store")
}
