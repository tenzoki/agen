package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tenzoki/agen/cellorg/public/client"
)

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

func main() {
	fmt.Println("=== Godast Storage Agent Demo ===")
	fmt.Println("This demo shows how to interact with the Godast Storage Agent")
	fmt.Println()

	// In a real implementation, this would connect to the actual message broker
	// For demo purposes, we'll simulate the requests and responses

	fmt.Println("🔧 Step 1: Initialize Storage Connection")
	// This would typically involve connecting to the broker and subscribing to responses
	fmt.Println("✅ Connected to storage broker")
	fmt.Println()

	// Demonstrate KV Operations
	fmt.Println("📦 Step 2: Key-Value Store Operations")
	demonstrateKVOperations()
	fmt.Println()

	// Demonstrate File Operations
	fmt.Println("📁 Step 3: File Storage Operations")
	demonstrateFileOperations()
	fmt.Println()

	// Demonstrate Graph Operations
	fmt.Println("🕸️  Step 4: Graph Database Operations")
	demonstrateGraphOperations()
	fmt.Println()

	// Demonstrate Full-text Search
	fmt.Println("🔍 Step 5: Full-text Search Operations")
	demonstrateSearchOperations()
	fmt.Println()

	// Demonstrate Complex Workflow
	fmt.Println("⚙️  Step 6: Complex Storage Workflow")
	demonstrateComplexWorkflow()
	fmt.Println()

	fmt.Println("🎉 Demo completed successfully!")
	fmt.Println("Check the storage data directory for persisted data.")
}

func demonstrateKVOperations() {
	fmt.Println("   Setting user data...")

	// KV Set operation
	userData := map[string]interface{}{
		"name":       "Alice Johnson",
		"email":      "alice@example.com",
		"role":       "developer",
		"created_at": time.Now().Format(time.RFC3339),
		"preferences": map[string]interface{}{
			"theme":    "dark",
			"language": "en",
		},
	}

	setRequest := StorageRequest{
		Operation: "kv_set",
		Key:       "user:alice:profile",
		Value:     userData,
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Storing user profile: %s\n", setRequest.Key)
	simulateStorageRequest(setRequest)

	// KV Get operation
	getRequest := StorageRequest{
		Operation: "kv_get",
		Key:       "user:alice:profile",
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Retrieving user profile: %s\n", getRequest.Key)
	simulateStorageRequest(getRequest)

	// KV Exists operation
	existsRequest := StorageRequest{
		Operation: "kv_exists",
		Key:       "user:alice:profile",
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Checking if key exists: %s\n", existsRequest.Key)
	simulateStorageRequest(existsRequest)

	fmt.Println("   ✅ KV operations completed")
}

func demonstrateFileOperations() {
	fmt.Println("   Storing and retrieving files...")

	// Create sample file content
	documentContent := `# Project Documentation

This is a sample document that demonstrates file storage capabilities.

## Features
- Content-addressable storage
- Automatic deduplication
- Metadata preservation
- Efficient retrieval

## Usage
Files are stored with their content hash as the key, enabling automatic
deduplication and integrity verification.

Created: ` + time.Now().Format(time.RFC3339)

	fileData := []byte(documentContent)
	metadata := map[string]interface{}{
		"filename":     "project_docs.md",
		"content_type": "text/markdown",
		"size":         len(fileData),
		"created_by":   "demo_script",
		"version":      "1.0",
		"tags":         []string{"documentation", "demo", "markdown"},
	}

	// File Store operation
	storeRequest := StorageRequest{
		Operation: "file_store",
		FileData:  fileData,
		Metadata:  metadata,
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Storing file: %s (%d bytes)\n", metadata["filename"], len(fileData))
	response := simulateStorageRequest(storeRequest)

	// Extract file hash from response (in real implementation)
	fileHash := "demo_file_hash_" + uuid.New().String()[:8]
	if response != nil && response.Success {
		fmt.Printf("   → File stored with hash: %s\n", fileHash)
	}

	// File Retrieve operation
	retrieveRequest := StorageRequest{
		Operation: "file_retrieve",
		Key:       fileHash,
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Retrieving file by hash: %s\n", fileHash)
	simulateStorageRequest(retrieveRequest)

	fmt.Println("   ✅ File operations completed")
}

func demonstrateGraphOperations() {
	fmt.Println("   Creating graph relationships...")

	// Create user vertex
	userVertex := map[string]interface{}{
		"name":      "Alice Johnson",
		"type":      "user",
		"role":      "developer",
		"team":      "backend",
		"joined_at": time.Now().Format(time.RFC3339),
	}

	createUserRequest := StorageRequest{
		Operation: "graph_create_vertex",
		Key:       "user:alice",
		Value:     userVertex,
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Creating user vertex: %s\n", createUserRequest.Key)
	simulateStorageRequest(createUserRequest)

	// Create project vertex
	projectVertex := map[string]interface{}{
		"name":        "Storage System",
		"type":        "project",
		"description": "Distributed storage with Godast",
		"status":      "active",
		"created_at":  time.Now().Format(time.RFC3339),
	}

	createProjectRequest := StorageRequest{
		Operation: "graph_create_vertex",
		Key:       "project:storage_system",
		Value:     projectVertex,
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Creating project vertex: %s\n", createProjectRequest.Key)
	simulateStorageRequest(createProjectRequest)

	// Create relationship edge
	relationshipData := map[string]interface{}{
		"from":       "user:alice",
		"to":         "project:storage_system",
		"label":      "contributes_to",
		"role":       "lead_developer",
		"since":      time.Now().Format(time.RFC3339),
		"commitment": "full-time",
	}

	createEdgeRequest := StorageRequest{
		Operation: "graph_create_edge",
		Value:     relationshipData,
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Creating relationship: %s → %s (%s)\n",
		relationshipData["from"], relationshipData["to"], relationshipData["label"])
	simulateStorageRequest(createEdgeRequest)

	// Query graph
	queryRequest := StorageRequest{
		Operation: "graph_query",
		Query:     "g.V().hasLabel('user').out('contributes_to').hasLabel('project')",
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Querying graph: Find projects that users contribute to\n")
	simulateStorageRequest(queryRequest)

	fmt.Println("   ✅ Graph operations completed")
}

func demonstrateSearchOperations() {
	fmt.Println("   Indexing and searching content...")

	// Index document 1
	doc1Content := `
	Godast Storage System Documentation

	Overview: Godast provides a unified storage platform combining key-value,
	graph database, file storage, and full-text search capabilities.

	Features:
	- BadgerDB backend for high performance
	- Content-addressable file storage
	- Graph relationships and queries
	- Full-text indexing and search
	- ACID transactions

	Use cases: Document management, analytics, content discovery, relationship mapping
	`

	indexDoc1Request := StorageRequest{
		Operation: "fulltext_index",
		Key:       "doc:godast_overview",
		Value:     doc1Content,
		Metadata: map[string]interface{}{
			"title":      "Godast Storage System Documentation",
			"category":   "documentation",
			"tags":       []string{"storage", "database", "search"},
			"author":     "demo_script",
			"created_at": time.Now().Format(time.RFC3339),
		},
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Indexing document: %s\n", indexDoc1Request.Key)
	simulateStorageRequest(indexDoc1Request)

	// Index document 2
	doc2Content := `
	Gox Pipeline Framework Integration

	The Gox pipeline framework integrates seamlessly with Godast storage
	to provide persistent, scalable data processing capabilities.

	Benefits:
	- Message persistence across restarts
	- Automatic deduplication of processed files
	- Search and analytics on processed data
	- Graph-based dependency tracking
	- Performance monitoring and statistics

	Applications: ETL pipelines, data processing, content analysis, workflow management
	`

	indexDoc2Request := StorageRequest{
		Operation: "fulltext_index",
		Key:       "doc:gox_integration",
		Value:     doc2Content,
		Metadata: map[string]interface{}{
			"title":      "Gox Pipeline Framework Integration",
			"category":   "integration",
			"tags":       []string{"pipeline", "processing", "integration"},
			"author":     "demo_script",
			"created_at": time.Now().Format(time.RFC3339),
		},
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Indexing document: %s\n", indexDoc2Request.Key)
	simulateStorageRequest(indexDoc2Request)

	// Search for documents
	searchRequest := StorageRequest{
		Operation:   "fulltext_search",
		SearchTerms: "storage database search",
		RequestID:   uuid.New().String(),
	}

	fmt.Printf("   → Searching for: '%s'\n", searchRequest.SearchTerms)
	simulateStorageRequest(searchRequest)

	// Search for specific terms
	searchRequest2 := StorageRequest{
		Operation:   "fulltext_search",
		SearchTerms: "pipeline processing",
		RequestID:   uuid.New().String(),
	}

	fmt.Printf("   → Searching for: '%s'\n", searchRequest2.SearchTerms)
	simulateStorageRequest(searchRequest2)

	fmt.Println("   ✅ Search operations completed")
}

func demonstrateComplexWorkflow() {
	fmt.Println("   Executing complex storage workflow...")

	// Step 1: Create a processing session
	sessionID := uuid.New().String()
	sessionData := map[string]interface{}{
		"session_id":  sessionID,
		"started_at":  time.Now().Format(time.RFC3339),
		"status":      "active",
		"workflow":    "document_processing",
		"total_files": 0,
		"processed":   0,
		"errors":      0,
	}

	fmt.Printf("   → Creating processing session: %s\n", sessionID)
	sessionRequest := StorageRequest{
		Operation: "kv_set",
		Key:       "session:" + sessionID,
		Value:     sessionData,
		RequestID: uuid.New().String(),
	}
	simulateStorageRequest(sessionRequest)

	// Step 2: Process multiple files
	files := []string{"document1.txt", "report.pdf", "data.json"}
	for i, filename := range files {
		fmt.Printf("   → Processing file %d: %s\n", i+1, filename)

		// Store file content
		fileContent := fmt.Sprintf("Content of %s - processed at %s", filename, time.Now().Format(time.RFC3339))
		storeRequest := StorageRequest{
			Operation: "file_store",
			FileData:  []byte(fileContent),
			Metadata: map[string]interface{}{
				"filename":   filename,
				"session_id": sessionID,
				"index":      i,
			},
			RequestID: uuid.New().String(),
		}
		simulateStorageRequest(storeRequest)

		// Index for search
		indexRequest := StorageRequest{
			Operation: "fulltext_index",
			Key:       fmt.Sprintf("file:%s:%d", sessionID, i),
			Value:     fileContent,
			Metadata: map[string]interface{}{
				"filename":   filename,
				"session_id": sessionID,
			},
			RequestID: uuid.New().String(),
		}
		simulateStorageRequest(indexRequest)

		// Create graph relationship
		edgeRequest := StorageRequest{
			Operation: "graph_create_edge",
			Value: map[string]interface{}{
				"from":  "session:" + sessionID,
				"to":    fmt.Sprintf("file:%s:%d", sessionID, i),
				"label": "processed",
				"order": i,
			},
			RequestID: uuid.New().String(),
		}
		simulateStorageRequest(edgeRequest)
	}

	// Step 3: Update session statistics
	sessionData["total_files"] = len(files)
	sessionData["processed"] = len(files)
	sessionData["completed_at"] = time.Now().Format(time.RFC3339)
	sessionData["status"] = "completed"

	updateRequest := StorageRequest{
		Operation: "kv_set",
		Key:       "session:" + sessionID,
		Value:     sessionData,
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Updating session statistics\n")
	simulateStorageRequest(updateRequest)

	// Step 4: Query processing results
	queryRequest := StorageRequest{
		Operation: "graph_query",
		Query:     fmt.Sprintf("g.V('session:%s').out('processed')", sessionID),
		RequestID: uuid.New().String(),
	}

	fmt.Printf("   → Querying processed files for session\n")
	simulateStorageRequest(queryRequest)

	// Step 5: Search across all processed content
	searchRequest := StorageRequest{
		Operation:   "fulltext_search",
		SearchTerms: "processed content",
		RequestID:   uuid.New().String(),
	}

	fmt.Printf("   → Searching across processed content\n")
	simulateStorageRequest(searchRequest)

	fmt.Println("   ✅ Complex workflow completed")
}

func simulateStorageRequest(request StorageRequest) *StorageResponse {
	// In a real implementation, this would send the request to the storage agent
	// via the message broker and wait for a response

	requestBytes, _ := json.Marshal(request)
	msg := &client.BrokerMessage{
		ID:      uuid.New().String(),
		Target:  "pub:storage-requests",
		Type:    "storage_request",
		Payload: requestBytes,
		Meta: map[string]interface{}{
			"sender":     "demo_script",
			"request_id": request.RequestID,
		},
	}

	// Simulate sending message
	fmt.Printf("      📤 Sending %s request: %s (msg ID: %s)\n", request.Operation, request.RequestID[:8], msg.ID[:8])

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	// Simulate successful response
	response := &StorageResponse{
		RequestID: request.RequestID,
		Success:   true,
		Result:    "demo_result_" + request.RequestID[:8],
		Count:     1,
	}

	fmt.Printf("      📥 Received response: %s (success: %v)\n", response.RequestID[:8], response.Success)

	return response
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
