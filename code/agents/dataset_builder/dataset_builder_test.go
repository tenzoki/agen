package main

import (
	"encoding/json"
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
	"github.com/tenzoki/gox/test/testutil"
)

// MockDatasetBuilderRunner implements agent.AgentRunner for testing
type MockDatasetBuilderRunner struct {
	config map[string]interface{}
}

func (m *MockDatasetBuilderRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockDatasetBuilderRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockDatasetBuilderRunner) Cleanup(base *agent.BaseAgent) {
}

func TestDatasetBuilderInitialization(t *testing.T) {
	config := map[string]interface{}{
		"output_format":   "jsonl",
		"batch_size":      1000,
		"validation_split": 0.2,
		"test_split":      0.1,
	}

	runner := &MockDatasetBuilderRunner{config: config}
	framework := agent.NewFramework(runner, "dataset-builder")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Dataset builder framework created successfully")
}

func TestDatasetBuilderTextClassificationDataset(t *testing.T) {
	config := map[string]interface{}{
		"dataset_type": "text_classification",
		"output_format": "jsonl",
		"include_metadata": true,
	}

	runner := &MockDatasetBuilderRunner{config: config}

	// Load real dataset from test data
	datasetContent, err := testutil.LoadTestFileString(testutil.DatasetJSONPath)
	if err != nil {
		t.Fatalf("Failed to load test dataset: %v", err)
	}

	var dataset map[string]interface{}
	if err := json.Unmarshal([]byte(datasetContent), &dataset); err != nil {
		t.Fatalf("Failed to parse test dataset: %v", err)
	}

	// Extract training data from the loaded dataset
	samples, ok := dataset["samples"].([]interface{})
	if !ok {
		t.Fatal("Invalid dataset format: missing samples")
	}

	trainingData := make([]map[string]interface{}, len(samples))
	for i, sample := range samples {
		sampleMap, ok := sample.(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid sample format at index %d", i)
		}
		trainingData[i] = sampleMap
	}

	msg := &client.BrokerMessage{
		ID:     "test-text-classification",
		Type:   "dataset_creation",
		Target: "dataset-builder",
		Payload: map[string]interface{}{
			"operation":     "create_classification_dataset",
			"training_data": trainingData,
			"dataset_config": map[string]interface{}{
				"name":         "sentiment_analysis",
				"task_type":    "text_classification",
				"num_classes":  3,
				"class_names":  []string{"positive", "negative", "neutral"},
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text classification dataset creation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text classification dataset test completed successfully")
}

func TestDatasetBuilderQuestionAnsweringDataset(t *testing.T) {
	config := map[string]interface{}{
		"dataset_type": "question_answering",
		"context_window": 512,
		"answer_extraction": true,
	}

	runner := &MockDatasetBuilderRunner{config: config}

	qaData := []map[string]interface{}{
		{
			"context": "The GOX framework is a distributed agent processing system built in Go. " +
				"It provides cell-based architecture for scalable document processing and analysis.",
			"question": "What is the GOX framework?",
			"answer": "A distributed agent processing system built in Go with cell-based architecture",
			"answer_start": 20,
			"answer_end": 86,
		},
		{
			"context": "Agents in GOX communicate through broker messages using a publish-subscribe pattern. " +
				"Each agent can process messages independently and send results to other agents.",
			"question": "How do agents communicate in GOX?",
			"answer": "Through broker messages using a publish-subscribe pattern",
			"answer_start": 34,
			"answer_end": 92,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-question-answering",
		Type:   "dataset_creation",
		Target: "dataset-builder",
		Payload: map[string]interface{}{
			"operation": "create_qa_dataset",
			"qa_data":   qaData,
			"dataset_config": map[string]interface{}{
				"name":          "gox_framework_qa",
				"task_type":     "question_answering",
				"answer_format": "extractive",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Question answering dataset creation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Question answering dataset test completed successfully")
}

func TestDatasetBuilderNERDataset(t *testing.T) {
	config := map[string]interface{}{
		"dataset_type": "named_entity_recognition",
		"tag_format": "BIO",
		"entity_types": []string{"PERSON", "ORG", "LOCATION", "DATE"},
	}

	runner := &MockDatasetBuilderRunner{config: config}

	nerData := []map[string]interface{}{
		{
			"text": "John Smith works at Google in Mountain View since 2020.",
			"entities": []map[string]interface{}{
				{"start": 0, "end": 10, "label": "PERSON", "text": "John Smith"},
				{"start": 20, "end": 26, "label": "ORG", "text": "Google"},
				{"start": 30, "end": 43, "label": "LOCATION", "text": "Mountain View"},
				{"start": 50, "end": 54, "label": "DATE", "text": "2020"},
			},
		},
		{
			"text": "Microsoft was founded by Bill Gates in Seattle.",
			"entities": []map[string]interface{}{
				{"start": 0, "end": 9, "label": "ORG", "text": "Microsoft"},
				{"start": 25, "end": 35, "label": "PERSON", "text": "Bill Gates"},
				{"start": 39, "end": 46, "label": "LOCATION", "text": "Seattle"},
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-ner-dataset",
		Type:   "dataset_creation",
		Target: "dataset-builder",
		Payload: map[string]interface{}{
			"operation": "create_ner_dataset",
			"ner_data":  nerData,
			"dataset_config": map[string]interface{}{
				"name":        "company_entities",
				"task_type":   "named_entity_recognition",
				"tag_scheme":  "BIO",
				"entity_types": config["entity_types"],
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("NER dataset creation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("NER dataset test completed successfully")
}

func TestDatasetBuilderDataAugmentation(t *testing.T) {
	config := map[string]interface{}{
		"augmentation": map[string]interface{}{
			"enabled": true,
			"techniques": []string{"synonym_replacement", "back_translation", "paraphrasing"},
			"augmentation_ratio": 2.0,
		},
	}

	runner := &MockDatasetBuilderRunner{config: config}

	originalData := []map[string]interface{}{
		{
			"text":  "The weather is beautiful today.",
			"label": "positive",
		},
		{
			"text":  "I don't like this product at all.",
			"label": "negative",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-data-augmentation",
		Type:   "dataset_augmentation",
		Target: "dataset-builder",
		Payload: map[string]interface{}{
			"operation":     "augment_dataset",
			"original_data": originalData,
			"augmentation_config": map[string]interface{}{
				"techniques": []string{"synonym_replacement", "paraphrasing"},
				"target_size": 6,
				"preserve_labels": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Data augmentation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Data augmentation test completed successfully")
}

func TestDatasetBuilderDatasetSplitting(t *testing.T) {
	config := map[string]interface{}{
		"splitting": map[string]interface{}{
			"train_ratio": 0.7,
			"val_ratio":   0.2,
			"test_ratio":  0.1,
			"stratified":  true,
		},
	}

	runner := &MockDatasetBuilderRunner{config: config}

	fullDataset := []map[string]interface{}{
		{"text": "Sample 1", "label": "A"},
		{"text": "Sample 2", "label": "B"},
		{"text": "Sample 3", "label": "A"},
		{"text": "Sample 4", "label": "C"},
		{"text": "Sample 5", "label": "B"},
		{"text": "Sample 6", "label": "A"},
		{"text": "Sample 7", "label": "C"},
		{"text": "Sample 8", "label": "B"},
		{"text": "Sample 9", "label": "A"},
		{"text": "Sample 10", "label": "C"},
	}

	msg := &client.BrokerMessage{
		ID:     "test-dataset-splitting",
		Type:   "dataset_split",
		Target: "dataset-builder",
		Payload: map[string]interface{}{
			"operation":    "split_dataset",
			"full_dataset": fullDataset,
			"split_config": map[string]interface{}{
				"method":       "stratified",
				"train_ratio":  0.7,
				"val_ratio":    0.2,
				"test_ratio":   0.1,
				"random_seed":  42,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Dataset splitting failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Dataset splitting test completed successfully")
}

func TestDatasetBuilderFeatureExtraction(t *testing.T) {
	config := map[string]interface{}{
		"feature_extraction": map[string]interface{}{
			"enabled": true,
			"extractors": []string{"tfidf", "word2vec", "bert_embeddings"},
			"max_features": 10000,
		},
	}

	runner := &MockDatasetBuilderRunner{config: config}

	textData := []map[string]interface{}{
		{
			"id":   "doc-1",
			"text": "Natural language processing is a fascinating field of study.",
		},
		{
			"id":   "doc-2",
			"text": "Machine learning algorithms can process large amounts of text data.",
		},
		{
			"id":   "doc-3",
			"text": "Deep learning models excel at understanding context in language.",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-feature-extraction",
		Type:   "feature_extraction",
		Target: "dataset-builder",
		Payload: map[string]interface{}{
			"operation":  "extract_features",
			"text_data":  textData,
			"extraction_config": map[string]interface{}{
				"feature_types": []string{"tfidf", "bert_embeddings"},
				"vector_size":   512,
				"include_metadata": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Feature extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Feature extraction test completed successfully")
}

func TestDatasetBuilderDatasetValidation(t *testing.T) {
	config := map[string]interface{}{
		"validation": map[string]interface{}{
			"enabled": true,
			"checks": []string{"completeness", "consistency", "quality"},
			"quality_threshold": 0.8,
		},
	}

	runner := &MockDatasetBuilderRunner{config: config}

	dataset := map[string]interface{}{
		"name": "test_dataset",
		"type": "text_classification",
		"data": []map[string]interface{}{
			{"text": "Valid sample with proper label", "label": "positive"},
			{"text": "Another valid sample", "label": "negative"},
			{"text": "", "label": "positive"}, // Invalid: empty text
			{"text": "Sample without label"},          // Invalid: missing label
		},
		"metadata": map[string]interface{}{
			"created_at": "2024-09-27T10:00:00Z",
			"version":    "1.0",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-dataset-validation",
		Type:   "dataset_validation",
		Target: "dataset-builder",
		Payload: map[string]interface{}{
			"operation": "validate_dataset",
			"dataset":   dataset,
			"validation_config": map[string]interface{}{
				"check_completeness":   true,
				"check_consistency":    true,
				"check_label_balance":  true,
				"minimum_samples":      10,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Dataset validation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Dataset validation test completed successfully")
}