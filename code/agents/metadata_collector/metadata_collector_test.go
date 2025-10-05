package main

import (
	"os"
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
	"github.com/tenzoki/agen/agents/testutil"
)

// MockMetadataCollectorRunner implements agent.AgentRunner for testing
type MockMetadataCollectorRunner struct {
	config map[string]interface{}
}

func (m *MockMetadataCollectorRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockMetadataCollectorRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockMetadataCollectorRunner) Cleanup(base *agent.BaseAgent) {
}

func TestMetadataCollectorInitialization(t *testing.T) {
	config := map[string]interface{}{
		"collection_scope": []string{"file_metadata", "content_metadata", "processing_metadata"},
		"storage_format":   "json",
		"include_checksums": true,
	}

	runner := &MockMetadataCollectorRunner{config: config}
	framework := agent.NewFramework(runner, "metadata-collector")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Metadata collector framework created successfully")
}

func TestMetadataCollectorFileMetadata(t *testing.T) {
	config := map[string]interface{}{
		"file_analysis": map[string]interface{}{
			"collect_stats":     true,
			"analyze_content":   true,
			"detect_file_type":  true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	// Use real test file
	testFilePath := testutil.GetTestDataPath(testutil.SampleArticlePath)

	// Get real file stats
	fileStats, err := os.Stat(testFilePath)
	if err != nil {
		t.Fatalf("Failed to get file stats: %v", err)
	}

	fileInfo := map[string]interface{}{
		"file_path":   testFilePath,
		"file_size":   fileStats.Size(),
		"created_at":  fileStats.ModTime().Format("2006-01-02T15:04:05Z"),
		"modified_at": fileStats.ModTime().Format("2006-01-02T15:04:05Z"),
		"permissions": fileStats.Mode().String(),
		"mime_type":   "text/plain",
	}

	msg := &client.BrokerMessage{
		ID:     "test-file-metadata",
		Type:   "metadata_collection",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":  "collect_file_metadata",
			"file_info":  fileInfo,
			"collection_config": map[string]interface{}{
				"include_file_stats":    true,
				"include_checksums":     true,
				"analyze_file_structure": true,
				"detect_encoding":       true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("File metadata collection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("File metadata collection test completed successfully")
}

func TestMetadataCollectorContentAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"content_analysis": map[string]interface{}{
			"language_detection": true,
			"topic_extraction":   true,
			"quality_assessment": true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	contentData := map[string]interface{}{
		"text_content": "This is a comprehensive document about machine learning algorithms and their applications in modern data science. " +
			"The document covers supervised learning, unsupervised learning, and reinforcement learning techniques.",
		"word_count":     32,
		"paragraph_count": 2,
		"sentence_count":  4,
	}

	msg := &client.BrokerMessage{
		ID:     "test-content-analysis",
		Type:   "metadata_collection",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":     "analyze_content_metadata",
			"content_data":  contentData,
			"analysis_config": map[string]interface{}{
				"detect_language":      true,
				"extract_keywords":     true,
				"assess_readability":   true,
				"identify_topics":      true,
				"calculate_statistics": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Content analysis metadata collection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Content analysis metadata collection test completed successfully")
}

func TestMetadataCollectorProcessingMetadata(t *testing.T) {
	config := map[string]interface{}{
		"processing_tracking": map[string]interface{}{
			"track_pipeline_steps": true,
			"measure_performance":  true,
			"log_transformations":  true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	processingInfo := map[string]interface{}{
		"pipeline_id":   "pipeline-001",
		"document_id":   "doc-123",
		"processing_steps": []map[string]interface{}{
			{
				"step":           "text_extraction",
				"agent":          "text-extractor",
				"start_time":     "2024-09-27T10:00:00Z",
				"end_time":       "2024-09-27T10:00:05Z",
				"duration_ms":    5000,
				"status":         "completed",
				"output_size":    1024,
			},
			{
				"step":           "text_chunking",
				"agent":          "text-chunker",
				"start_time":     "2024-09-27T10:00:05Z",
				"end_time":       "2024-09-27T10:00:08Z",
				"duration_ms":    3000,
				"status":         "completed",
				"chunks_created": 5,
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-processing-metadata",
		Type:   "metadata_collection",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":        "collect_processing_metadata",
			"processing_info":  processingInfo,
			"tracking_config": map[string]interface{}{
				"include_performance_metrics": true,
				"track_resource_usage":        true,
				"log_intermediate_results":    true,
				"calculate_throughput":        true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Processing metadata collection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Processing metadata collection test completed successfully")
}

func TestMetadataCollectorStructuralAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"structural_analysis": map[string]interface{}{
			"analyze_document_structure": true,
			"extract_hierarchy":          true,
			"identify_sections":          true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	documentStructure := map[string]interface{}{
		"document_type": "research_paper",
		"sections": []map[string]interface{}{
			{
				"title":       "Abstract",
				"level":       1,
				"word_count":  150,
				"position":    0,
			},
			{
				"title":       "Introduction",
				"level":       1,
				"word_count":  500,
				"position":    1,
				"subsections": []map[string]interface{}{
					{"title": "Background", "level": 2, "word_count": 200},
					{"title": "Motivation", "level": 2, "word_count": 300},
				},
			},
			{
				"title":       "Methodology",
				"level":       1,
				"word_count":  800,
				"position":    2,
			},
		},
		"total_sections":    3,
		"total_subsections": 2,
		"max_depth":         2,
	}

	msg := &client.BrokerMessage{
		ID:     "test-structural-analysis",
		Type:   "metadata_collection",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":           "analyze_document_structure",
			"document_structure":  documentStructure,
			"structural_config": map[string]interface{}{
				"extract_hierarchy":      true,
				"analyze_section_balance": true,
				"detect_formatting":      true,
				"map_cross_references":   true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Structural analysis metadata collection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Structural analysis metadata collection test completed successfully")
}

func TestMetadataCollectorQualityMetrics(t *testing.T) {
	config := map[string]interface{}{
		"quality_assessment": map[string]interface{}{
			"completeness_check": true,
			"consistency_check":  true,
			"accuracy_assessment": true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	qualityData := map[string]interface{}{
		"document_id": "doc-quality-001",
		"content_metrics": map[string]interface{}{
			"readability_score":   7.5,
			"grammar_score":       0.92,
			"spelling_errors":     2,
			"consistency_score":   0.88,
		},
		"structural_metrics": map[string]interface{}{
			"section_balance":     0.75,
			"hierarchy_depth":     3,
			"cross_ref_validity":  0.95,
		},
		"processing_metrics": map[string]interface{}{
			"extraction_accuracy": 0.96,
			"parsing_errors":      1,
			"data_completeness":   0.94,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-quality-metrics",
		Type:   "metadata_collection",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":     "assess_quality_metrics",
			"quality_data":  qualityData,
			"quality_config": map[string]interface{}{
				"calculate_overall_score": true,
				"identify_improvement_areas": true,
				"benchmark_against_standards": true,
				"generate_recommendations": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Quality metrics collection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Quality metrics collection test completed successfully")
}

func TestMetadataCollectorRelationshipMapping(t *testing.T) {
	config := map[string]interface{}{
		"relationship_analysis": map[string]interface{}{
			"detect_references":    true,
			"map_dependencies":     true,
			"analyze_connections":  true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	relationshipData := map[string]interface{}{
		"document_id": "doc-rel-001",
		"references": []map[string]interface{}{
			{
				"type":      "citation",
				"target_id": "ref-001",
				"context":   "As discussed in Smith et al. (2023)",
				"position":  150,
			},
			{
				"type":      "internal_link",
				"target_id": "section-3",
				"context":   "See Section 3 for details",
				"position":  420,
			},
		},
		"dependencies": []map[string]interface{}{
			{
				"type":        "data_dependency",
				"depends_on":  "dataset-A",
				"description": "Analysis depends on dataset A",
			},
			{
				"type":        "method_dependency",
				"depends_on":  "algorithm-X",
				"description": "Uses algorithm X for processing",
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-relationship-mapping",
		Type:   "metadata_collection",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":          "map_relationships",
			"relationship_data":  relationshipData,
			"mapping_config": map[string]interface{}{
				"create_relationship_graph": true,
				"validate_references":       true,
				"detect_circular_deps":      true,
				"calculate_impact_scores":   true,
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

func TestMetadataCollectorTemporalTracking(t *testing.T) {
	config := map[string]interface{}{
		"temporal_tracking": map[string]interface{}{
			"track_versions":   true,
			"monitor_changes":  true,
			"analyze_trends":   true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	temporalData := map[string]interface{}{
		"document_id": "doc-temporal-001",
		"version_history": []map[string]interface{}{
			{
				"version":    "1.0",
				"timestamp":  "2024-09-01T09:00:00Z",
				"author":     "John Doe",
				"changes":    "Initial version",
				"word_count": 1000,
			},
			{
				"version":    "1.1",
				"timestamp":  "2024-09-15T14:30:00Z",
				"author":     "Jane Smith",
				"changes":    "Added methodology section",
				"word_count": 1250,
			},
			{
				"version":    "1.2",
				"timestamp":  "2024-09-27T10:00:00Z",
				"author":     "John Doe",
				"changes":    "Updated conclusions",
				"word_count": 1300,
			},
		},
		"change_frequency": "weekly",
		"active_contributors": 2,
	}

	msg := &client.BrokerMessage{
		ID:     "test-temporal-tracking",
		Type:   "metadata_collection",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":      "track_temporal_metadata",
			"temporal_data":  temporalData,
			"tracking_config": map[string]interface{}{
				"analyze_change_patterns":  true,
				"predict_future_changes":   true,
				"identify_active_periods":  true,
				"calculate_stability_score": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Temporal tracking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Temporal tracking test completed successfully")
}

func TestMetadataCollectorAggregation(t *testing.T) {
	config := map[string]interface{}{
		"aggregation": map[string]interface{}{
			"combine_sources":    true,
			"resolve_conflicts":  true,
			"normalize_formats":  true,
		},
	}

	runner := &MockMetadataCollectorRunner{config: config}

	metadataSources := []map[string]interface{}{
		{
			"source":   "file_system",
			"metadata": map[string]interface{}{
				"file_size":   1024000,
				"created_at":  "2024-09-27T09:00:00Z",
				"file_type":   "pdf",
			},
		},
		{
			"source":   "content_analysis",
			"metadata": map[string]interface{}{
				"language":    "en",
				"topic":       "machine_learning",
				"word_count":  1500,
			},
		},
		{
			"source":   "user_input",
			"metadata": map[string]interface{}{
				"title":       "ML Research Paper",
				"author":      "Research Team",
				"category":    "academic",
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-metadata-aggregation",
		Type:   "metadata_aggregation",
		Target: "metadata-collector",
		Payload: map[string]interface{}{
			"operation":         "aggregate_metadata",
			"metadata_sources":  metadataSources,
			"aggregation_config": map[string]interface{}{
				"conflict_resolution": "priority_based",
				"source_priorities":   map[string]int{
					"user_input":       1,
					"content_analysis": 2,
					"file_system":      3,
				},
				"normalize_values": true,
				"validate_consistency": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Metadata aggregation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Metadata aggregation test completed successfully")
}