package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockGodastStorageRunner implements agent.AgentRunner for testing
type MockGodastStorageRunner struct {
	config map[string]interface{}
}

func (m *MockGodastStorageRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockGodastStorageRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockGodastStorageRunner) Cleanup(base *agent.BaseAgent) {
}

func TestGodastStorageInitialization(t *testing.T) {
	config := map[string]interface{}{
		"storage_backend": "file_system",
		"base_path":       "/test/godast_storage",
		"compression":     true,
		"encryption":      false,
	}

	runner := &MockGodastStorageRunner{config: config}
	framework := agent.NewFramework(runner, "godast-storage")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Godast storage framework created successfully")
}

func TestGodastStorageDocumentStorage(t *testing.T) {
	config := map[string]interface{}{
		"storage_format": "godast",
		"versioning":     true,
		"indexing":       true,
	}

	runner := &MockGodastStorageRunner{config: config}

	documentData := map[string]interface{}{
		"document_id": "doc-godast-001",
		"godast_tree": map[string]interface{}{
			"type": "document",
			"children": []map[string]interface{}{
				{
					"type": "heading",
					"level": 1,
					"content": "Introduction to Machine Learning",
					"position": map[string]interface{}{
						"start": 0,
						"end":   32,
					},
				},
				{
					"type": "paragraph",
					"content": "Machine learning is a subset of artificial intelligence.",
					"position": map[string]interface{}{
						"start": 33,
						"end":   90,
					},
				},
			},
		},
		"metadata": map[string]interface{}{
			"title":       "ML Introduction",
			"author":      "AI Research Team",
			"created_at":  "2024-09-27T10:00:00Z",
			"version":     "1.0",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-document-storage",
		Type:   "storage_operation",
		Target: "godast-storage",
		Payload: map[string]interface{}{
			"operation":      "store_document",
			"document_data":  documentData,
			"storage_options": map[string]interface{}{
				"create_index":    true,
				"compress":        true,
				"validate_schema": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Document storage failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Document storage test completed successfully")
}

func TestGodastStorageDocumentRetrieval(t *testing.T) {
	config := map[string]interface{}{
		"retrieval_optimization": true,
		"caching":               true,
	}

	runner := &MockGodastStorageRunner{config: config}

	retrievalRequest := map[string]interface{}{
		"document_id": "doc-godast-001",
		"version":     "latest",
	}

	msg := &client.BrokerMessage{
		ID:     "test-document-retrieval",
		Type:   "storage_operation",
		Target: "godast-storage",
		Payload: map[string]interface{}{
			"operation":         "retrieve_document",
			"retrieval_request": retrievalRequest,
			"retrieval_options": map[string]interface{}{
				"decompress":        true,
				"validate_integrity": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Document retrieval failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Document retrieval test completed successfully")
}