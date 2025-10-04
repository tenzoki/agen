package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agen/omni/public/omnistore"
	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// StorageServiceAgent implements HTTP service mode for storage operations
type StorageServiceAgent struct {
	agent.DefaultAgentRunner
	omniStore omnistore.OmniStore
	config    *StorageServiceConfig
	server    *http.Server
}

type StorageServiceConfig struct {
	// Storage settings
	DataPath       string `yaml:"data_path"`
	MaxFileSize    int64  `yaml:"max_file_size"`
	EnableKV       bool   `yaml:"enable_kv"`
	EnableGraph    bool   `yaml:"enable_graph"`
	EnableFiles    bool   `yaml:"enable_files"`
	EnableFulltext bool   `yaml:"enable_fulltext"`

	// Service settings
	ListenPort            int    `yaml:"listen_port"`
	EnableCORS            bool   `yaml:"enable_cors"`
	LogRequests           bool   `yaml:"log_requests"`
	AuthRequired          bool   `yaml:"auth_required"`
	MaxConcurrentRequests int    `yaml:"max_concurrent_requests"`
	RequestTimeout        string `yaml:"request_timeout"`
	ConnectionPoolSize    int    `yaml:"connection_pool_size"`
}

// Init initializes the storage service agent
func (s *StorageServiceAgent) Init(base *agent.BaseAgent) error {
	// Load configuration from base agent
	config := &StorageServiceConfig{
		// Storage defaults
		DataPath:       base.GetConfigString("data_path", "/tmp/gox-storage-service"),
		MaxFileSize:    int64(base.GetConfigInt("max_file_size", 104857600)), // 100MB
		EnableKV:       base.GetConfigBool("enable_kv", true),
		EnableGraph:    base.GetConfigBool("enable_graph", true),
		EnableFiles:    base.GetConfigBool("enable_files", true),
		EnableFulltext: base.GetConfigBool("enable_fulltext", true),

		// Service defaults
		ListenPort:            base.GetConfigInt("listen_port", 9002),
		EnableCORS:            base.GetConfigBool("enable_cors", true),
		LogRequests:           base.GetConfigBool("log_requests", true),
		AuthRequired:          base.GetConfigBool("auth_required", false),
		MaxConcurrentRequests: base.GetConfigInt("max_concurrent_requests", 100),
		RequestTimeout:        base.GetConfigString("request_timeout", "30s"),
		ConnectionPoolSize:    base.GetConfigInt("connection_pool_size", 10),
	}

	s.config = config

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(s.config.DataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize OmniStore
	omniStore, err := omnistore.NewOmniStoreWithDefaults(s.config.DataPath)
	if err != nil {
		return fmt.Errorf("failed to initialize OmniStore: %w", err)
	}
	s.omniStore = omniStore

	base.LogInfo("Godast Storage Service initialized at %s", s.config.DataPath)
	base.LogInfo("Storage capabilities: KV=%v, Graph=%v, Files=%v, FullText=%v",
		s.config.EnableKV, s.config.EnableGraph, s.config.EnableFiles, s.config.EnableFulltext)

	// Initialize HTTP server
	s.initHTTPServer(base)

	return nil
}

func (s *StorageServiceAgent) initHTTPServer(base *agent.BaseAgent) {
	mux := http.NewServeMux()

	// Register service endpoints
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/storage/v1/kv", s.handleKVOperations)
	mux.HandleFunc("/storage/v1/graph", s.handleGraphOperations)
	mux.HandleFunc("/storage/v1/files", s.handleFileOperations)
	mux.HandleFunc("/storage/v1/search", s.handleSearchOperations)

	// Add middleware
	handler := s.loggingMiddleware(mux, base)
	if s.config.EnableCORS {
		handler = s.corsMiddleware(handler)
	}

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.ListenPort),
		Handler: handler,
	}

	// Start server in background
	go func() {
		base.LogInfo("Storage service starting on port %d", s.config.ListenPort)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			base.LogError("HTTP server error: %v", err)
		}
	}()
}

// ProcessMessage handles messages in service mode (minimal processing)
func (s *StorageServiceAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// In service mode, most operations are handled via HTTP
	// This can handle control messages or health checks
	base.LogDebug("Service received control message: %s", msg.Type)

	switch msg.Type {
	case "health_check":
		return &client.BrokerMessage{
			ID:      msg.ID + "_response",
			Type:    "health_response",
			Payload: map[string]interface{}{"status": "healthy", "service": "storage"},
		}, nil
	case "shutdown":
		// Graceful shutdown
		base.LogInfo("Received shutdown signal")
		if s.server != nil {
			s.server.Close()
		}
		return nil, nil
	default:
		base.LogDebug("Unknown control message type: %s", msg.Type)
		return nil, nil
	}
}

// HTTP Handlers

func (s *StorageServiceAgent) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "godast-storage",
		"version": "1.0.0",
		"capabilities": map[string]bool{
			"kv":       s.config.EnableKV,
			"graph":    s.config.EnableGraph,
			"files":    s.config.EnableFiles,
			"fulltext": s.config.EnableFulltext,
		},
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *StorageServiceAgent) handleKVOperations(w http.ResponseWriter, r *http.Request) {
	if !s.config.EnableKV {
		http.Error(w, "KV operations are disabled", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case "GET":
		s.handleKVGet(w, r)
	case "POST", "PUT":
		s.handleKVSet(w, r)
	case "DELETE":
		s.handleKVDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *StorageServiceAgent) handleKVGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	valueBytes, err := s.omniStore.KV().Get(key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Key not found: %v", err), http.StatusNotFound)
		return
	}

	// Unmarshal the stored value
	var value interface{}
	if err := json.Unmarshal(valueBytes, &value); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal value: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"key":   key,
		"value": value,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *StorageServiceAgent) handleKVSet(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.Key == "" {
		http.Error(w, "Missing key", http.StatusBadRequest)
		return
	}

	// Marshal value for storage
	valueBytes, err := json.Marshal(request.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal value: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.omniStore.KV().Set(request.Key, valueBytes); err != nil {
		http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"key":     request.Key,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *StorageServiceAgent) handleKVDelete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	if err := s.omniStore.KV().Delete(key); err != nil {
		http.Error(w, fmt.Sprintf("Delete error: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"key":     key,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *StorageServiceAgent) handleGraphOperations(w http.ResponseWriter, r *http.Request) {
	if !s.config.EnableGraph {
		http.Error(w, "Graph operations are disabled", http.StatusServiceUnavailable)
		return
	}

	// Placeholder implementation
	response := map[string]interface{}{
		"message": "Graph operations not yet implemented in service mode",
		"method":  r.Method,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *StorageServiceAgent) handleFileOperations(w http.ResponseWriter, r *http.Request) {
	if !s.config.EnableFiles {
		http.Error(w, "File operations are disabled", http.StatusServiceUnavailable)
		return
	}

	// Placeholder implementation
	response := map[string]interface{}{
		"message": "File operations not yet implemented in service mode",
		"method":  r.Method,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *StorageServiceAgent) handleSearchOperations(w http.ResponseWriter, r *http.Request) {
	if !s.config.EnableFulltext {
		http.Error(w, "Search operations are disabled", http.StatusServiceUnavailable)
		return
	}

	// Placeholder implementation
	response := map[string]interface{}{
		"message": "Search operations not yet implemented in service mode",
		"method":  r.Method,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Middleware

func (s *StorageServiceAgent) loggingMiddleware(next http.Handler, base *agent.BaseAgent) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.config.LogRequests {
			start := time.Now()
			next.ServeHTTP(w, r)
			base.LogDebug("HTTP %s %s - %v", r.Method, r.URL.Path, time.Since(start))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (s *StorageServiceAgent) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper function to detect if running in service mode
func isServiceMode() bool {
	return os.Getenv("GOX_STORAGE_MODE") == "service"
}

func main() {
	if isServiceMode() {
		log.Println("Starting Godast Storage Service")
		agent.Run(&StorageServiceAgent{}, "godast-storage")
	} else {
		log.Println("Starting Godast Storage Agent (pipeline mode)")
		agent.Run(&GodastStorageAgent{}, "godast-storage")
	}
}
