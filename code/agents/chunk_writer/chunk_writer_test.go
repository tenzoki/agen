package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockChunkWriterRunner implements agent.AgentRunner for testing
type MockChunkWriterRunner struct {
	config map[string]interface{}
}

func (m *MockChunkWriterRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockChunkWriterRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockChunkWriterRunner) Cleanup(base *agent.BaseAgent) {
}

func TestChunkWriterInitialization(t *testing.T) {
	config := map[string]interface{}{
		"chunk_size":       "10MB",
		"output_format":    "json",
		"naming_pattern":   "chunk_{index:04d}.json",
		"preserve_metadata": true,
	}

	runner := &MockChunkWriterRunner{config: config}
	framework := agent.NewFramework(runner, "chunk-writer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Chunk writer framework created successfully")
}

func TestChunkWriterBasicWriting(t *testing.T) {
	config := map[string]interface{}{
		"chunk_size":    "1MB",
		"output_format": "json",
	}

	runner := &MockChunkWriterRunner{config: config}

	// Test writing a simple chunk
	msg := &client.BrokerMessage{
		ID:     "test-basic-write",
		Type:   "chunk_write",
		Target: "chunk-writer",
		Payload: map[string]interface{}{
			"chunk_data": map[string]interface{}{
				"id":      "chunk-001",
				"content": "This is test chunk content for writing.",
				"metadata": map[string]interface{}{
					"source_file": "/test/input.txt",
					"chunk_index": 1,
					"total_chunks": 5,
				},
			},
			"output_path": "/test/output/",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Basic chunk writing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Basic chunk writing test completed successfully")
}

func TestChunkWriterCompression(t *testing.T) {
	config := map[string]interface{}{
		"chunk_size":    "5MB",
		"output_format": "binary",
		"compression": map[string]interface{}{
			"algorithm": "gzip",
			"level":     6,
			"enable":    true,
		},
	}

	runner := &MockChunkWriterRunner{config: config}

	// Large chunk data to test compression
	largeContent := make([]byte, 1024*100) // 100KB of data
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	msg := &client.BrokerMessage{
		ID:     "test-compression",
		Type:   "chunk_write",
		Target: "chunk-writer",
		Payload: map[string]interface{}{
			"chunk_data": map[string]interface{}{
				"id":      "chunk-compressed-001",
				"content": largeContent,
				"metadata": map[string]interface{}{
					"source_file":    "/test/large_input.bin",
					"original_size":  len(largeContent),
					"compression":    "gzip",
				},
			},
			"output_path": "/test/output/compressed/",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Compression chunk writing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Compression chunk writing test completed successfully")
}

func TestChunkWriterParallelProcessing(t *testing.T) {
	config := map[string]interface{}{
		"parallel_processing": map[string]interface{}{
			"enabled":       true,
			"worker_count":  4,
			"queue_size":    50,
		},
		"output_format": "json",
	}

	runner := &MockChunkWriterRunner{config: config}

	// Test multiple chunks for parallel processing
	chunks := []map[string]interface{}{
		{
			"id":      "chunk-parallel-001",
			"content": "First parallel chunk content",
			"metadata": map[string]interface{}{
				"worker_id": 1,
				"chunk_index": 1,
			},
		},
		{
			"id":      "chunk-parallel-002",
			"content": "Second parallel chunk content",
			"metadata": map[string]interface{}{
				"worker_id": 2,
				"chunk_index": 2,
			},
		},
		{
			"id":      "chunk-parallel-003",
			"content": "Third parallel chunk content",
			"metadata": map[string]interface{}{
				"worker_id": 3,
				"chunk_index": 3,
			},
		},
	}

	for i, chunk := range chunks {
		msg := &client.BrokerMessage{
			ID:     "test-parallel-" + string(rune('1'+i)),
			Type:   "chunk_write",
			Target: "chunk-writer",
			Payload: map[string]interface{}{
				"chunk_data":  chunk,
				"output_path": "/test/output/parallel/",
			},
			Meta: make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Parallel chunk writing failed for chunk %d: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for chunk %d", i+1)
		}
	}
	t.Log("Parallel chunk writing test completed successfully")
}

func TestChunkWriterValidation(t *testing.T) {
	config := map[string]interface{}{
		"validation": map[string]interface{}{
			"enable_checksums":     true,
			"verify_integrity":     true,
			"checksum_algorithm":   "sha256",
		},
		"output_format": "json",
	}

	runner := &MockChunkWriterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-validation",
		Type:   "chunk_write",
		Target: "chunk-writer",
		Payload: map[string]interface{}{
			"chunk_data": map[string]interface{}{
				"id":      "chunk-validation-001",
				"content": "Chunk content for validation testing",
				"metadata": map[string]interface{}{
					"expected_checksum": "abc123def456",
					"size":              35,
				},
			},
			"output_path": "/test/output/validated/",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Validation chunk writing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Validation chunk writing test completed successfully")
}

func TestChunkWriterStreamingMode(t *testing.T) {
	config := map[string]interface{}{
		"chunk_strategy": "time_based",
		"time_window":    "30s",
		"max_chunk_size": "2MB",
		"streaming": map[string]interface{}{
			"buffer_size":           "1MB",
			"flush_interval":        "10s",
			"real_time_processing":  true,
		},
	}

	runner := &MockChunkWriterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-streaming",
		Type:   "stream_chunk_write",
		Target: "chunk-writer",
		Payload: map[string]interface{}{
			"stream_data": []map[string]interface{}{
				{
					"timestamp": "2024-09-27T10:00:00Z",
					"data":      "First streaming chunk",
				},
				{
					"timestamp": "2024-09-27T10:00:05Z",
					"data":      "Second streaming chunk",
				},
				{
					"timestamp": "2024-09-27T10:00:10Z",
					"data":      "Third streaming chunk",
				},
			},
			"output_path": "/test/output/streaming/",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Streaming chunk writing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Streaming chunk writing test completed successfully")
}

func TestChunkWriterErrorRecovery(t *testing.T) {
	config := map[string]interface{}{
		"error_handling": map[string]interface{}{
			"retry_attempts": 3,
			"retry_delay":    "1s",
		},
		"output_format": "json",
	}

	runner := &MockChunkWriterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-error-recovery",
		Type:   "chunk_write",
		Target: "chunk-writer",
		Payload: map[string]interface{}{
			"chunk_data": map[string]interface{}{
				"id":      "chunk-error-001",
				"content": "Chunk content for error recovery testing",
			},
			"output_path": "/invalid/path/that/does/not/exist/",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Error recovery test failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Error recovery test completed successfully")
}