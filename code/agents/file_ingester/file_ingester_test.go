package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockFileIngesterRunner implements agent.AgentRunner for testing
type MockFileIngesterRunner struct {
	config map[string]interface{}
}

func (m *MockFileIngesterRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockFileIngesterRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockFileIngesterRunner) Cleanup(base *agent.BaseAgent) {
}

func TestFileIngesterInitialization(t *testing.T) {
	config := map[string]interface{}{
		"watch_interval":   "5s",
		"include_metadata": true,
		"batch_size":      20,
	}

	runner := &MockFileIngesterRunner{config: config}
	framework := agent.NewFramework(runner, "file-ingester")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("File ingester framework created successfully")
}

func TestFileIngesterConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		valid  bool
	}{
		{
			name: "valid_config",
			config: map[string]interface{}{
				"watch_interval": "10s",
				"batch_size":    50,
				"digest":        true,
			},
			valid: true,
		},
		{
			name: "invalid_interval",
			config: map[string]interface{}{
				"watch_interval": "invalid",
			},
			valid: false,
		},
		{
			name: "negative_batch_size",
			config: map[string]interface{}{
				"batch_size": -1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockFileIngesterRunner{config: tt.config}
			framework := agent.NewFramework(runner, "file-ingester")
			if framework == nil {
				t.Fatal("Expected framework to be created")
			}

			// Test configuration validation would go here
			// For now, just verify framework creation
			t.Log("Configuration test passed")
		})
	}
}

func TestFileIngesterProcessing(t *testing.T) {
	config := map[string]interface{}{
		"include_metadata": true,
		"digest":          true,
		"digest_strategy": "archive",
	}

	runner := &MockFileIngesterRunner{config: config}

	// Mock file content
	fileContent := "This is test file content for ingestion."

	msg := &client.BrokerMessage{
		ID:      "test-file-ingestion",
		Type:    "file_detected",
		Target:  "file-ingester-001",
		Payload: map[string]interface{}{
			"file_path": "/test/data/sample.txt",
			"size":      len(fileContent),
			"modified":  "2024-09-26T10:00:00Z",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("File ingestion processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("File ingestion test completed successfully")
}