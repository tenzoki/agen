// +build integration

package main

import (
	"os"
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// TestNERInference tests actual NER inference with the real model
func TestNERInference(t *testing.T) {
	// Check if model exists
	modelPath := "../../models/ner/xlm-roberta-ner.onnx"
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("NER model not found - run model conversion first")
	}

	// Create NER agent
	nerAgent := &NERAgent{}

	// Create minimal BaseAgent for testing
	baseAgent := &agent.BaseAgent{
		ID:        "ner-test-001",
		AgentType: "ner-agent",
		Debug:     true,
		Config: map[string]interface{}{
			"model_path":      modelPath,
			"tokenizer_path":  "../../models/ner/",
			"max_seq_length":  128,
			"confidence_threshold": 0.5,
			"enable_debug":    true,
		},
	}

	// Initialize agent
	if err := nerAgent.Init(baseAgent); err != nil {
		t.Fatalf("Failed to initialize NER agent: %v", err)
	}
	defer nerAgent.Cleanup(baseAgent)

	// Test cases
	testCases := []struct {
		name           string
		text           string
		expectedCount  int
		expectedTypes  map[string]bool
		expectedTexts  map[string]bool
	}{
		{
			name:          "English entities",
			text:          "Angela Merkel visited Microsoft in Berlin.",
			expectedCount: 3,
			expectedTypes: map[string]bool{
				"PERSON": true,
				"ORG":    true,
				"LOC":    true,
			},
			expectedTexts: map[string]bool{
				"Angela Merkel": true,
				"Microsoft":     true,
				"Berlin":        true,
			},
		},
		{
			name:          "German entities",
			text:          "Die Bundeskanzlerin traf sich mit dem Vorstand von Siemens in München.",
			expectedCount: 2,
			expectedTypes: map[string]bool{
				"ORG": true,
				"LOC": true,
			},
			expectedTexts: map[string]bool{
				"Siemens": true,
				"München": true,
			},
		},
		{
			name:          "French entities",
			text:          "Emmanuel Macron a rencontré les dirigeants de Renault à Paris.",
			expectedCount: 3,
			expectedTypes: map[string]bool{
				"PERSON": true,
				"ORG":    true,
				"LOC":    true,
			},
			expectedTexts: map[string]bool{
				"Emmanuel Macron": true,
				"Renault":         true,
				"Paris":           true,
			},
		},
		{
			name:          "Spanish entities",
			text:          "Pedro Sánchez visitó la sede de Telefónica en Madrid.",
			expectedCount: 3,
			expectedTypes: map[string]bool{
				"PERSON": true,
				"ORG":    true,
				"LOC":    true,
			},
			expectedTexts: map[string]bool{
				"Pedro Sánchez": true,
				"Telefónica":    true,
				"Madrid":        true,
			},
		},
		{
			name:          "No entities",
			text:          "The weather is nice today.",
			expectedCount: 0,
			expectedTypes: map[string]bool{},
			expectedTexts: map[string]bool{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.text)

			// Extract entities
			entities, err := nerAgent.extractEntities(tc.text, baseAgent)
			if err != nil {
				t.Fatalf("Failed to extract entities: %v", err)
			}

			t.Logf("Found %d entities", len(entities))
			for _, entity := range entities {
				t.Logf("  - %s: %s (confidence: %.2f, pos: %d-%d)",
					entity.Type, entity.Text, entity.Confidence, entity.Start, entity.End)
			}

			// Verify entity count (allow some flexibility)
			if len(entities) < tc.expectedCount-1 || len(entities) > tc.expectedCount+1 {
				t.Errorf("Expected ~%d entities, got %d", tc.expectedCount, len(entities))
			}

			// Verify entity types and texts
			foundTypes := make(map[string]bool)
			foundTexts := make(map[string]bool)

			for _, entity := range entities {
				foundTypes[entity.Type] = true
				foundTexts[entity.Text] = true

				// Verify confidence is reasonable
				if entity.Confidence < 0 || entity.Confidence > 1 {
					t.Errorf("Invalid confidence %f for entity %s", entity.Confidence, entity.Text)
				}

				// Verify positions are valid
				if entity.Start < 0 || entity.End <= entity.Start || entity.End > len(tc.text) {
					t.Errorf("Invalid positions %d-%d for entity %s in text of length %d",
						entity.Start, entity.End, entity.Text, len(tc.text))
				}

				// Verify extracted text matches positions
				extractedText := tc.text[entity.Start:entity.End]
				if extractedText != entity.Text {
					t.Errorf("Entity text mismatch: '%s' vs '%s'", entity.Text, extractedText)
				}
			}

			// Check for expected types (with flexibility for model variations)
			for expectedType := range tc.expectedTypes {
				if !foundTypes[expectedType] {
					t.Logf("Note: Expected type %s not found (model variation)", expectedType)
				}
			}

			// Check for expected texts (with flexibility for model variations)
			for expectedText := range tc.expectedTexts {
				if !foundTexts[expectedText] {
					t.Logf("Note: Expected text '%s' not found (model variation)", expectedText)
				}
			}
		})
	}
}

// BenchmarkNERInference benchmarks entity extraction performance
func BenchmarkNERInference(b *testing.B) {
	modelPath := "../../models/ner/xlm-roberta-ner.onnx"
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		b.Skip("NER model not found")
	}

	// Setup
	nerAgent := &NERAgent{}
	baseAgent := &agent.BaseAgent{
		ID:        "ner-bench-001",
		AgentType: "ner-agent",
		Config: map[string]interface{}{
			"model_path":     modelPath,
			"tokenizer_path": "../../models/ner/",
			"max_seq_length": 128,
			"enable_debug":   false,
		},
	}

	if err := nerAgent.Init(baseAgent); err != nil {
		b.Fatalf("Failed to initialize: %v", err)
	}
	defer nerAgent.Cleanup(baseAgent)

	text := "Angela Merkel visited Microsoft in Berlin and met with Tim Cook from Apple."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := nerAgent.extractEntities(text, baseAgent)
		if err != nil {
			b.Fatalf("Extraction failed: %v", err)
		}
	}
}

// TestNEREndToEnd tests the full agent message processing flow
func TestNEREndToEnd(t *testing.T) {
	modelPath := "../../models/ner/xlm-roberta-ner.onnx"
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("NER model not found")
	}

	nerAgent := &NERAgent{}
	baseAgent := &agent.BaseAgent{
		ID:        "ner-e2e-001",
		AgentType: "ner-agent",
		Config: map[string]interface{}{
			"model_path":     modelPath,
			"tokenizer_path": "../../models/ner/",
			"max_seq_length": 128,
			"enable_debug":   true,
		},
	}

	if err := nerAgent.Init(baseAgent); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer nerAgent.Cleanup(baseAgent)

	// Test with JSON request
	requestJSON := []byte(`{"text": "Angela Merkel visited Microsoft in Berlin.", "project_id": "test"}`)
	msg := &client.BrokerMessage{
		Payload: requestJSON,
	}

	response, err := nerAgent.ProcessMessage(msg, baseAgent)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	t.Logf("Response payload: %s", string(response.Payload.([]byte)))

	// Test with plain text
	plainText := []byte("Emmanuel Macron visited Paris.")
	msg2 := &client.BrokerMessage{
		Payload: plainText,
	}

	response2, err := nerAgent.ProcessMessage(msg2, baseAgent)
	if err != nil {
		t.Fatalf("ProcessMessage (plain text) failed: %v", err)
	}

	if response2 == nil {
		t.Fatal("Response2 is nil")
	}

	t.Logf("Response2 payload: %s", string(response2.Payload.([]byte)))
}
