package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNERAgentInitialization(t *testing.T) {
	// Check if model exists
	modelPath := "../../models/ner/xlm-roberta-ner.onnx"
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("NER model not found - run model conversion first")
	}

	// Check if tokenizer exists
	tokenizerPath := "../../models/ner/tokenizer.json"
	if _, err := os.Stat(tokenizerPath); os.IsNotExist(err) {
		t.Skip("Tokenizer not found")
	}

	t.Log("Model and tokenizer files found")
	t.Log("Model:", modelPath)
	t.Log("Tokenizer:", tokenizerPath)
}

func TestEntityTypeNormalization(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"PER", "PERSON"},
		{"ORG", "ORG"},
		{"LOC", "LOC"},
		{"MISC", "MISC"},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeEntityType(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestModelFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping model file test in short mode (requires external model files)")
	}

	modelDir := "../../models/ner"

	// Skip if model directory doesn't exist (external files not in repo)
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		t.Skip("Skipping model file test: model directory not found (external files)")
		return
	}

	// Check required files
	requiredFiles := []string{
		"xlm-roberta-ner.onnx",
		"tokenizer.json",
		"config.json",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(modelDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Required file not found: %s", path)
		} else {
			info, _ := os.Stat(path)
			t.Logf("Found %s (%.2f MB)", file, float64(info.Size())/(1024*1024))
		}
	}
}

func TestTokenizerOutput(t *testing.T) {
	// Test TokenizerOutput structure
	tokens := &TokenizerOutput{
		InputIDs:      []int32{0, 100, 200, 2},
		AttentionMask: []int32{1, 1, 1, 1},
		Offsets:       [][]int{{0, 0}, {0, 5}, {6, 12}, {12, 12}},
	}

	if len(tokens.InputIDs) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(tokens.InputIDs))
	}

	if len(tokens.Offsets) != len(tokens.InputIDs) {
		t.Errorf("Offsets length mismatch: %d vs %d", len(tokens.Offsets), len(tokens.InputIDs))
	}
}

func TestEntityStructure(t *testing.T) {
	// Test Entity structure
	entity := Entity{
		Text:       "Berlin",
		Type:       "LOC",
		Start:      10,
		End:        16,
		Confidence: 0.95,
	}

	if entity.Text != "Berlin" {
		t.Errorf("Expected text 'Berlin', got '%s'", entity.Text)
	}

	if entity.Type != "LOC" {
		t.Errorf("Expected type 'LOC', got '%s'", entity.Type)
	}

	if entity.Confidence < 0 || entity.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", entity.Confidence)
	}
}

// Benchmark tests
func BenchmarkEntityCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Entity{
			Text:       "Test Entity",
			Type:       "PERSON",
			Start:      0,
			End:        11,
			Confidence: 0.9,
		}
	}
}

func BenchmarkNormalization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = normalizeEntityType("PER")
	}
}
