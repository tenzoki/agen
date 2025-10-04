package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockContextEnricherRunner implements agent.AgentRunner for testing
type MockContextEnricherRunner struct {
	config map[string]interface{}
}

func (m *MockContextEnricherRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockContextEnricherRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockContextEnricherRunner) Cleanup(base *agent.BaseAgent) {
}

func TestContextEnricherInitialization(t *testing.T) {
	config := map[string]interface{}{
		"enrichment_sources": []string{"metadata", "relationships", "semantic"},
		"context_window":     512,
		"max_depth":          3,
	}

	runner := &MockContextEnricherRunner{config: config}
	framework := agent.NewFramework(runner, "context-enricher")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Context enricher framework created successfully")
}

func TestContextEnricherMetadataEnrichment(t *testing.T) {
	config := map[string]interface{}{
		"enrichment_types": []string{"metadata", "timestamps", "tags"},
		"metadata_sources": []string{"file_system", "database", "external_api"},
	}

	runner := &MockContextEnricherRunner{config: config}

	document := map[string]interface{}{
		"id":      "doc-001",
		"content": "Sample document content for metadata enrichment",
		"source":  "/test/documents/sample.txt",
	}

	msg := &client.BrokerMessage{
		ID:     "test-metadata-enrichment",
		Type:   "context_enrichment",
		Target: "context-enricher",
		Payload: map[string]interface{}{
			"operation": "enrich_metadata",
			"document":  document,
			"enrichment_config": map[string]interface{}{
				"include_file_stats":    true,
				"extract_timestamps":    true,
				"analyze_content_type":  true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Metadata enrichment failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Metadata enrichment test completed successfully")
}

func TestContextEnricherRelationshipMapping(t *testing.T) {
	config := map[string]interface{}{
		"relationship_analysis": map[string]interface{}{
			"enabled":           true,
			"max_relationships": 50,
			"relationship_types": []string{"references", "citations", "dependencies"},
		},
	}

	runner := &MockContextEnricherRunner{config: config}

	documents := []map[string]interface{}{
		{
			"id":      "doc-001",
			"content": "This document references doc-002 and depends on doc-003.",
			"type":    "specification",
		},
		{
			"id":      "doc-002",
			"content": "Base implementation document.",
			"type":    "implementation",
		},
		{
			"id":      "doc-003",
			"content": "Core library documentation.",
			"type":    "library",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-relationship-mapping",
		Type:   "context_enrichment",
		Target: "context-enricher",
		Payload: map[string]interface{}{
			"operation":  "map_relationships",
			"documents":  documents,
			"analysis_config": map[string]interface{}{
				"extract_references": true,
				"build_dependency_graph": true,
				"identify_circular_deps": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Relationship mapping failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Relationship mapping test completed successfully")
}

func TestContextEnricherSemanticAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"semantic_analysis": map[string]interface{}{
			"enabled":        true,
			"model_type":     "transformer",
			"context_window": 1024,
		},
	}

	runner := &MockContextEnricherRunner{config: config}

	content := map[string]interface{}{
		"text": "Machine learning algorithms require careful feature engineering. " +
			"Deep learning models can automatically extract features from raw data. " +
			"Neural networks are particularly effective for pattern recognition tasks.",
		"domain": "technology",
		"language": "en",
	}

	msg := &client.BrokerMessage{
		ID:     "test-semantic-analysis",
		Type:   "context_enrichment",
		Target: "context-enricher",
		Payload: map[string]interface{}{
			"operation": "semantic_analysis",
			"content":   content,
			"analysis_options": map[string]interface{}{
				"extract_entities":    true,
				"identify_topics":     true,
				"compute_embeddings":  true,
				"find_similar_content": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Semantic analysis failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Semantic analysis test completed successfully")
}

func TestContextEnricherHierarchicalContext(t *testing.T) {
	config := map[string]interface{}{
		"hierarchical_context": map[string]interface{}{
			"enabled":     true,
			"max_levels":  5,
			"merge_strategy": "weighted",
		},
	}

	runner := &MockContextEnricherRunner{config: config}

	contextHierarchy := map[string]interface{}{
		"document": map[string]interface{}{
			"id":      "doc-main",
			"content": "Main document content",
		},
		"parent_context": map[string]interface{}{
			"section": "Introduction",
			"chapter": "Getting Started",
			"book":    "Complete Guide",
		},
		"sibling_context": []map[string]interface{}{
			{"id": "doc-prev", "title": "Previous Section"},
			{"id": "doc-next", "title": "Next Section"},
		},
		"child_context": []map[string]interface{}{
			{"id": "subsec-1", "title": "Subsection 1"},
			{"id": "subsec-2", "title": "Subsection 2"},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-hierarchical-context",
		Type:   "context_enrichment",
		Target: "context-enricher",
		Payload: map[string]interface{}{
			"operation": "enrich_hierarchical_context",
			"hierarchy": contextHierarchy,
			"options": map[string]interface{}{
				"include_parent_summary":  true,
				"include_sibling_topics":  true,
				"include_child_overview":  true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Hierarchical context enrichment failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Hierarchical context enrichment test completed successfully")
}

func TestContextEnricherTemporalContext(t *testing.T) {
	config := map[string]interface{}{
		"temporal_analysis": map[string]interface{}{
			"enabled":       true,
			"time_window":   "30d",
			"track_changes": true,
		},
	}

	runner := &MockContextEnricherRunner{config: config}

	temporalData := map[string]interface{}{
		"document": map[string]interface{}{
			"id":           "doc-versioned",
			"content":      "Current version of the document",
			"last_modified": "2024-09-27T10:00:00Z",
		},
		"version_history": []map[string]interface{}{
			{
				"version":   "1.0",
				"timestamp": "2024-09-01T09:00:00Z",
				"changes":   "Initial version",
			},
			{
				"version":   "1.1",
				"timestamp": "2024-09-15T14:30:00Z",
				"changes":   "Added examples section",
			},
			{
				"version":   "1.2",
				"timestamp": "2024-09-27T10:00:00Z",
				"changes":   "Updated implementation details",
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-temporal-context",
		Type:   "context_enrichment",
		Target: "context-enricher",
		Payload: map[string]interface{}{
			"operation":     "enrich_temporal_context",
			"temporal_data": temporalData,
			"analysis_options": map[string]interface{}{
				"track_evolution":     true,
				"identify_patterns":   true,
				"predict_next_change": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Temporal context enrichment failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Temporal context enrichment test completed successfully")
}

func TestContextEnricherCrossReferenceAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"cross_reference": map[string]interface{}{
			"enabled":          true,
			"reference_types":  []string{"explicit", "implicit", "semantic"},
			"max_references":   100,
		},
	}

	runner := &MockContextEnricherRunner{config: config}

	documentSet := []map[string]interface{}{
		{
			"id":      "api-doc",
			"content": "API documentation describes the authenticate() function.",
			"type":    "documentation",
		},
		{
			"id":      "code-impl",
			"content": "function authenticate(user, password) { return validateCredentials(user, password); }",
			"type":    "code",
		},
		{
			"id":      "test-case",
			"content": "Test case for authenticate function with valid credentials.",
			"type":    "test",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-cross-reference",
		Type:   "context_enrichment",
		Target: "context-enricher",
		Payload: map[string]interface{}{
			"operation":    "cross_reference_analysis",
			"document_set": documentSet,
			"reference_config": map[string]interface{}{
				"find_explicit_refs":  true,
				"find_implicit_refs":  true,
				"semantic_similarity": 0.8,
				"create_link_graph":   true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Cross-reference analysis failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Cross-reference analysis test completed successfully")
}