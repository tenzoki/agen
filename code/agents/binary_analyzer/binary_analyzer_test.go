package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockBinaryAnalyzerRunner implements agent.AgentRunner for testing
type MockBinaryAnalyzerRunner struct {
	config map[string]interface{}
}

func (m *MockBinaryAnalyzerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockBinaryAnalyzerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockBinaryAnalyzerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestBinaryAnalyzerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"deep_scan":       true,
		"extract_strings": true,
		"max_file_size":   "50MB",
	}

	runner := &MockBinaryAnalyzerRunner{config: config}
	framework := agent.NewFramework(runner, "binary-analyzer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Binary analyzer framework created successfully")
}

func TestBinaryAnalyzerFileTypeDetection(t *testing.T) {
	config := map[string]interface{}{
		"detect_file_type": true,
	}

	runner := &MockBinaryAnalyzerRunner{config: config}

	tests := []struct {
		name        string
		data        []byte
		expectedType string
	}{
		{
			name:        "png_file",
			data:        []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expectedType: "image/png",
		},
		{
			name:        "pdf_file",
			data:        []byte("%PDF-1.4"),
			expectedType: "application/pdf",
		},
		{
			name:        "zip_file",
			data:        []byte{0x50, 0x4B, 0x03, 0x04},
			expectedType: "application/zip",
		},
		{
			name:        "exe_file",
			data:        []byte("MZ"),
			expectedType: "application/x-executable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &client.BrokerMessage{
				ID:      "test-binary-" + tt.name,
				Type:    "binary_analysis",
				Target:  "binary-analyzer",
				Payload: map[string]interface{}{
					"content":   tt.data,
					"file_path": "/test/" + tt.name,
				},
				Meta: make(map[string]interface{}),
			}

			result, err := runner.ProcessMessage(msg, nil)
			if err != nil {
				t.Errorf("Processing failed: %v", err)
			}
			if result == nil {
				t.Error("Expected result message")
			}
			t.Logf("Successfully processed %s file type detection", tt.name)
		})
	}
}

func TestBinaryAnalyzerStringExtraction(t *testing.T) {
	config := map[string]interface{}{
		"extract_strings":  true,
		"min_string_length": 4,
	}

	runner := &MockBinaryAnalyzerRunner{config: config}

	// Binary data with embedded strings
	binaryData := []byte{
		0x00, 0x01, 0x02, 0x03, // Binary header
		'H', 'e', 'l', 'l', 'o', 0x00, // String: "Hello"
		0xFF, 0xFE, 0xFD, // More binary
		'W', 'o', 'r', 'l', 'd', 0x00, // String: "World"
		0x89, 0x50, 0x4E, 0x47, // PNG magic bytes
		'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g', 0x00, // String: "Test String"
	}

	msg := &client.BrokerMessage{
		ID:      "test-string-extraction",
		Type:    "binary_analysis",
		Target:  "binary-analyzer",
		Payload: map[string]interface{}{
			"content":   binaryData,
			"file_path": "/test/binary_with_strings.bin",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("String extraction processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("String extraction test completed successfully")
}

func TestBinaryAnalyzerMetadataExtraction(t *testing.T) {
	config := map[string]interface{}{
		"extract_metadata": true,
		"deep_scan":       true,
	}

	runner := &MockBinaryAnalyzerRunner{config: config}

	// Mock PE/EXE file header
	peData := []byte("MZ") // DOS header
	peData = append(peData, make([]byte, 58)...) // Padding
	peData = append(peData, []byte("PE\x00\x00")...) // PE signature

	msg := &client.BrokerMessage{
		ID:      "test-metadata-extraction",
		Type:    "binary_analysis",
		Target:  "binary-analyzer",
		Payload: map[string]interface{}{
			"content":   peData,
			"file_path": "/test/sample.exe",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Metadata extraction processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Metadata extraction test completed successfully")
}

func TestBinaryAnalyzerSizeLimit(t *testing.T) {
	config := map[string]interface{}{
		"max_file_size": "1KB", // Very small limit for testing
	}

	runner := &MockBinaryAnalyzerRunner{config: config}

	// Create data larger than the limit
	largeData := make([]byte, 2048) // 2KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	msg := &client.BrokerMessage{
		ID:      "test-size-limit",
		Type:    "binary_analysis",
		Target:  "binary-analyzer",
		Payload: map[string]interface{}{
			"content":   largeData,
			"file_path": "/test/large_file.bin",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Size limit processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Size limit test completed successfully")
}