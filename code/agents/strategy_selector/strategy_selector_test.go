package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockStrategySelectorRunner implements agent.AgentRunner for testing
type MockStrategySelectorRunner struct {
	config map[string]interface{}
}

func (m *MockStrategySelectorRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockStrategySelectorRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockStrategySelectorRunner) Cleanup(base *agent.BaseAgent) {
}

func TestStrategySelectorInitialization(t *testing.T) {
	config := map[string]interface{}{
		"available_strategies": []string{"chunking", "extraction", "transformation"},
		"selection_algorithm": "rule_based",
		"fallback_strategy":   "default",
	}

	runner := &MockStrategySelectorRunner{config: config}
	framework := agent.NewFramework(runner, "strategy-selector")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Strategy selector framework created successfully")
}

func TestStrategySelectorContentBasedSelection(t *testing.T) {
	config := map[string]interface{}{
		"selection_criteria": map[string]interface{}{
			"file_type": "priority",
			"file_size": "threshold",
			"content_complexity": "heuristic",
		},
	}

	runner := &MockStrategySelectorRunner{config: config}

	contentAnalysis := map[string]interface{}{
		"file_type": "pdf",
		"file_size": 5242880, // 5MB
		"content_metadata": map[string]interface{}{
			"page_count":    150,
			"has_images":    true,
			"has_tables":    true,
			"text_density":  0.7,
			"language":      "en",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-content-based-selection",
		Type:   "strategy_selection",
		Target: "strategy-selector",
		Payload: map[string]interface{}{
			"operation":         "select_processing_strategy",
			"content_analysis":  contentAnalysis,
			"available_strategies": []string{
				"ocr_extraction",
				"pdf_text_extraction",
				"hybrid_extraction",
				"table_extraction",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Content-based strategy selection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Content-based strategy selection test completed successfully")
}

func TestStrategySelectorPerformanceBasedSelection(t *testing.T) {
	config := map[string]interface{}{
		"performance_tracking": map[string]interface{}{
			"enabled": true,
			"metrics": []string{"processing_time", "accuracy", "resource_usage"},
			"history_window": "7d",
		},
	}

	runner := &MockStrategySelectorRunner{config: config}

	performanceHistory := []map[string]interface{}{
		{
			"strategy": "chunking_strategy_a",
			"metrics": map[string]interface{}{
				"avg_processing_time": 1250, // ms
				"accuracy_score":      0.92,
				"memory_usage":        256,   // MB
				"success_rate":        0.95,
			},
		},
		{
			"strategy": "chunking_strategy_b",
			"metrics": map[string]interface{}{
				"avg_processing_time": 2100, // ms
				"accuracy_score":      0.98,
				"memory_usage":        512,   // MB
				"success_rate":        0.99,
			},
		},
		{
			"strategy": "chunking_strategy_c",
			"metrics": map[string]interface{}{
				"avg_processing_time": 800, // ms
				"accuracy_score":      0.88,
				"memory_usage":        128,  // MB
				"success_rate":        0.90,
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-performance-based-selection",
		Type:   "strategy_selection",
		Target: "strategy-selector",
		Payload: map[string]interface{}{
			"operation":            "select_optimal_strategy",
			"performance_history":  performanceHistory,
			"selection_criteria": map[string]interface{}{
				"optimize_for":     "balanced", // speed, accuracy, balanced
				"max_memory_mb":    400,
				"max_time_ms":      1500,
				"min_accuracy":     0.90,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Performance-based strategy selection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Performance-based strategy selection test completed successfully")
}

func TestStrategySelectorRuleBasedSelection(t *testing.T) {
	config := map[string]interface{}{
		"rule_engine": map[string]interface{}{
			"enabled": true,
			"rule_priority": "sequential",
		},
	}

	runner := &MockStrategySelectorRunner{config: config}

	rules := []map[string]interface{}{
		{
			"name": "large_pdf_rule",
			"conditions": map[string]interface{}{
				"file_type": "pdf",
				"file_size_mb": map[string]interface{}{
					"greater_than": 10,
				},
			},
			"action": map[string]interface{}{
				"strategy": "chunked_pdf_processing",
				"parameters": map[string]interface{}{
					"chunk_size": "2MB",
					"parallel_processing": true,
				},
			},
		},
		{
			"name": "image_heavy_rule",
			"conditions": map[string]interface{}{
				"content_type": "document",
				"image_ratio": map[string]interface{}{
					"greater_than": 0.3,
				},
			},
			"action": map[string]interface{}{
				"strategy": "ocr_enhanced_extraction",
				"parameters": map[string]interface{}{
					"ocr_engine": "tesseract",
					"image_preprocessing": true,
				},
			},
		},
	}

	inputDocument := map[string]interface{}{
		"file_type":    "pdf",
		"file_size_mb": 15.5,
		"content_analysis": map[string]interface{}{
			"image_ratio":   0.4,
			"text_ratio":    0.6,
			"page_count":    200,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-rule-based-selection",
		Type:   "strategy_selection",
		Target: "strategy-selector",
		Payload: map[string]interface{}{
			"operation":       "apply_rules",
			"rules":           rules,
			"input_document":  inputDocument,
			"rule_config": map[string]interface{}{
				"evaluation_mode": "all_matching",
				"conflict_resolution": "priority_order",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Rule-based strategy selection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Rule-based strategy selection test completed successfully")
}

func TestStrategySelectorMLBasedSelection(t *testing.T) {
	config := map[string]interface{}{
		"ml_selection": map[string]interface{}{
			"enabled": true,
			"model_type": "decision_tree",
			"training_data_size": 1000,
		},
	}

	runner := &MockStrategySelectorRunner{config: config}

	documentFeatures := map[string]interface{}{
		"numerical_features": map[string]interface{}{
			"file_size_mb":      25.8,
			"page_count":        45,
			"word_count":        12500,
			"image_count":       8,
			"table_count":       3,
			"text_density":      0.75,
		},
		"categorical_features": map[string]interface{}{
			"file_extension": "docx",
			"language":       "en",
			"document_type":  "technical_report",
			"source":         "internal",
		},
		"derived_features": map[string]interface{}{
			"complexity_score":   0.82,
			"structure_score":    0.65,
			"readability_score":  0.71,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-ml-based-selection",
		Type:   "strategy_selection",
		Target: "strategy-selector",
		Payload: map[string]interface{}{
			"operation":          "ml_predict_strategy",
			"document_features":  documentFeatures,
			"prediction_config": map[string]interface{}{
				"model_version":     "v2.1",
				"confidence_threshold": 0.8,
				"return_probabilities": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("ML-based strategy selection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("ML-based strategy selection test completed successfully")
}

func TestStrategySelectorHybridSelection(t *testing.T) {
	config := map[string]interface{}{
		"hybrid_selection": map[string]interface{}{
			"enabled": true,
			"selection_methods": []string{"rule_based", "ml_based", "performance_based"},
			"aggregation_method": "weighted_voting",
		},
	}

	runner := &MockStrategySelectorRunner{config: config}

	selectionInput := map[string]interface{}{
		"document": map[string]interface{}{
			"id":            "complex-doc-001",
			"type":          "mixed_content",
			"size_mb":       8.5,
			"complexity":    "high",
		},
		"context": map[string]interface{}{
			"processing_deadline": "2024-09-27T15:00:00Z",
			"quality_requirements": "high",
			"resource_constraints": map[string]interface{}{
				"max_memory_mb": 1024,
				"max_cpu_cores": 4,
			},
		},
		"available_strategies": []string{
			"fast_extraction",
			"accurate_extraction",
			"hybrid_extraction",
			"parallel_extraction",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-hybrid-selection",
		Type:   "strategy_selection",
		Target: "strategy-selector",
		Payload: map[string]interface{}{
			"operation":        "hybrid_selection",
			"selection_input":  selectionInput,
			"hybrid_config": map[string]interface{}{
				"method_weights": map[string]interface{}{
					"rule_based":        0.4,
					"ml_based":          0.4,
					"performance_based": 0.2,
				},
				"consensus_threshold": 0.7,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Hybrid strategy selection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Hybrid strategy selection test completed successfully")
}

func TestStrategySelectorAdaptiveSelection(t *testing.T) {
	config := map[string]interface{}{
		"adaptive_selection": map[string]interface{}{
			"enabled": true,
			"feedback_learning": true,
			"adaptation_window": "24h",
		},
	}

	runner := &MockStrategySelectorRunner{config: config}

	feedbackData := []map[string]interface{}{
		{
			"document_id": "doc-001",
			"selected_strategy": "strategy_a",
			"performance_metrics": map[string]interface{}{
				"processing_time": 1200,
				"accuracy":        0.94,
				"user_satisfaction": 4.2,
			},
			"timestamp": "2024-09-27T09:00:00Z",
		},
		{
			"document_id": "doc-002",
			"selected_strategy": "strategy_b",
			"performance_metrics": map[string]interface{}{
				"processing_time": 800,
				"accuracy":        0.89,
				"user_satisfaction": 3.8,
			},
			"timestamp": "2024-09-27T09:30:00Z",
		},
	}

	newDocument := map[string]interface{}{
		"characteristics": map[string]interface{}{
			"file_type":     "pdf",
			"complexity":    "medium",
			"urgency":       "high",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-adaptive-selection",
		Type:   "strategy_selection",
		Target: "strategy-selector",
		Payload: map[string]interface{}{
			"operation":      "adaptive_selection",
			"feedback_data":  feedbackData,
			"new_document":   newDocument,
			"adaptation_config": map[string]interface{}{
				"learning_rate":      0.1,
				"exploration_rate":   0.15,
				"update_frequency":   "real_time",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Adaptive strategy selection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Adaptive strategy selection test completed successfully")
}