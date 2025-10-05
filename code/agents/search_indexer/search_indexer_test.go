package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockSearchIndexerRunner implements agent.AgentRunner for testing
type MockSearchIndexerRunner struct {
	config map[string]interface{}
}

func (m *MockSearchIndexerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockSearchIndexerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockSearchIndexerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestSearchIndexerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"index_name": "documents",
		"backend":    "elasticsearch",
		"endpoint":   "http://localhost:9200",
	}

	runner := &MockSearchIndexerRunner{config: config}
	framework := agent.NewFramework(runner, "search-indexer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Search indexer framework created successfully")
}

func TestSearchIndexerDocumentIndexing(t *testing.T) {
	config := map[string]interface{}{
		"index_name": "test_documents",
		"backend":    "elasticsearch",
	}

	runner := &MockSearchIndexerRunner{config: config}

	document := map[string]interface{}{
		"id":      "doc-001",
		"title":   "Test Document",
		"content": "This is a test document for search indexing functionality.",
		"metadata": map[string]interface{}{
			"author":      "Test Author",
			"created_at":  "2024-09-27T10:00:00Z",
			"category":    "test",
			"tags":        []string{"test", "search", "indexing"},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-document-indexing",
		Type:   "document_index",
		Target: "search-indexer",
		Payload: map[string]interface{}{
			"operation": "index",
			"document":  document,
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Document indexing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Document indexing test completed successfully")
}

func TestSearchIndexerBulkIndexing(t *testing.T) {
	config := map[string]interface{}{
		"index_name":  "bulk_documents",
		"backend":     "elasticsearch",
		"batch_size":  100,
		"bulk_mode":   true,
	}

	runner := &MockSearchIndexerRunner{config: config}

	documents := []map[string]interface{}{
		{
			"id":      "bulk-001",
			"title":   "First Bulk Document",
			"content": "Content of the first bulk document.",
		},
		{
			"id":      "bulk-002",
			"title":   "Second Bulk Document",
			"content": "Content of the second bulk document.",
		},
		{
			"id":      "bulk-003",
			"title":   "Third Bulk Document",
			"content": "Content of the third bulk document.",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-bulk-indexing",
		Type:   "bulk_index",
		Target: "search-indexer",
		Payload: map[string]interface{}{
			"operation":  "bulk_index",
			"documents":  documents,
			"batch_size": 100,
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Bulk indexing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Bulk indexing test completed successfully")
}

func TestSearchIndexerFacetedSearch(t *testing.T) {
	config := map[string]interface{}{
		"index_name": "faceted_documents",
		"backend":    "elasticsearch",
		"features": map[string]interface{}{
			"faceted_search": true,
			"auto_complete":  true,
		},
	}

	runner := &MockSearchIndexerRunner{config: config}

	document := map[string]interface{}{
		"id":      "facet-001",
		"title":   "Faceted Search Document",
		"content": "Document for testing faceted search capabilities.",
		"facets": map[string]interface{}{
			"category":    "research",
			"tags":        []string{"facet", "search", "elasticsearch"},
			"author":      "Research Team",
			"year":        2024,
			"department":  "engineering",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-faceted-search",
		Type:   "faceted_index",
		Target: "search-indexer",
		Payload: map[string]interface{}{
			"operation": "index_with_facets",
			"document":  document,
			"facet_config": map[string]interface{}{
				"category":   map[string]interface{}{"type": "terms", "size": 20},
				"tags":       map[string]interface{}{"type": "terms", "size": 50},
				"year":       map[string]interface{}{"type": "range", "ranges": []map[string]int{{"from": 2020, "to": 2025}}},
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Faceted search indexing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Faceted search indexing test completed successfully")
}

func TestSearchIndexerLogIndexing(t *testing.T) {
	config := map[string]interface{}{
		"index_name": "application_logs",
		"backend":    "elasticsearch",
		"time_based_indices": true,
		"index_pattern": "logs-{yyyy.MM.dd}",
	}

	runner := &MockSearchIndexerRunner{config: config}

	logEntry := map[string]interface{}{
		"timestamp": "2024-09-27T10:15:32.123Z",
		"level":     "INFO",
		"logger":    "gox.agent.search_indexer",
		"message":   "Successfully indexed document",
		"metadata": map[string]interface{}{
			"document_id":      "doc-123",
			"processing_time":  125,
			"index_name":       "test_index",
		},
		"labels": map[string]interface{}{
			"environment": "production",
			"service":     "gox-framework",
			"component":   "search-indexer",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-log-indexing",
		Type:   "log_index",
		Target: "search-indexer",
		Payload: map[string]interface{}{
			"operation":  "index_log",
			"log_entry":  logEntry,
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Log indexing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Log indexing test completed successfully")
}

func TestSearchIndexerRealTimeUpdates(t *testing.T) {
	config := map[string]interface{}{
		"index_name":    "realtime_documents",
		"backend":       "elasticsearch",
		"indexing_mode": "real_time",
		"refresh_interval": "1s",
	}

	runner := &MockSearchIndexerRunner{config: config}

	updates := []map[string]interface{}{
		{
			"operation": "create",
			"document": map[string]interface{}{
				"id":      "rt-001",
				"title":   "Real-time Document",
				"content": "Initial content",
			},
		},
		{
			"operation": "update",
			"document_id": "rt-001",
			"updates": map[string]interface{}{
				"content": "Updated content for real-time document",
				"last_modified": "2024-09-27T10:20:00Z",
			},
		},
		{
			"operation": "delete",
			"document_id": "rt-001",
		},
	}

	for i, update := range updates {
		msg := &client.BrokerMessage{
			ID:     "test-realtime-" + string(rune('1'+i)),
			Type:   "realtime_update",
			Target: "search-indexer",
			Payload: update,
			Meta:    make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Real-time update %d failed: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for update %d", i+1)
		}
	}
	t.Log("Real-time updates test completed successfully")
}

func TestSearchIndexerSearchQuery(t *testing.T) {
	config := map[string]interface{}{
		"index_name": "searchable_documents",
		"backend":    "elasticsearch",
	}

	runner := &MockSearchIndexerRunner{config: config}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "test search framework",
				"fields": []string{"title^2", "content", "tags"},
			},
		},
		"size": 10,
		"from": 0,
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":   map[string]interface{}{},
				"content": map[string]interface{}{},
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-search-query",
		Type:   "search_query",
		Target: "search-indexer",
		Payload: map[string]interface{}{
			"operation": "search",
			"query":     searchQuery,
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Search query failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Search query test completed successfully")
}

func TestSearchIndexerIndexManagement(t *testing.T) {
	config := map[string]interface{}{
		"index_name": "managed_index",
		"backend":    "elasticsearch",
	}

	runner := &MockSearchIndexerRunner{config: config}

	managementOps := []map[string]interface{}{
		{
			"operation": "create_index",
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 0,
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
						"analyzer": "standard",
					},
					"content": map[string]interface{}{
						"type": "text",
						"analyzer": "standard",
					},
				},
			},
		},
		{
			"operation": "get_index_stats",
		},
		{
			"operation": "refresh_index",
		},
	}

	for i, op := range managementOps {
		msg := &client.BrokerMessage{
			ID:     "test-index-mgmt-" + string(rune('1'+i)),
			Type:   "index_management",
			Target: "search-indexer",
			Payload: op,
			Meta:    make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Index management operation %d failed: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for operation %d", i+1)
		}
	}
	t.Log("Index management test completed successfully")
}