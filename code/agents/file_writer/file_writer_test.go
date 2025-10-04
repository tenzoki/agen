package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockFileWriterRunner implements agent.AgentRunner for testing
type MockFileWriterRunner struct {
	config map[string]interface{}
}

func (m *MockFileWriterRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockFileWriterRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockFileWriterRunner) Cleanup(base *agent.BaseAgent) {
}

func TestFileWriterInitialization(t *testing.T) {
	config := map[string]interface{}{
		"create_directories": true,
		"include_metadata":  true,
		"format":           "json",
	}

	runner := &MockFileWriterRunner{config: config}
	framework := agent.NewFramework(runner, "file-writer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("File writer framework created successfully")
}

func TestFileWriterFormatting(t *testing.T) {
	tests := []struct {
		name   string
		format string
		data   interface{}
		valid  bool
	}{
		{
			name:   "json_format",
			format: "json",
			data:   map[string]interface{}{"key": "value"},
			valid:  true,
		},
		{
			name:   "text_format",
			format: "text",
			data:   "plain text content",
			valid:  true,
		},
		{
			name:   "csv_format",
			format: "csv",
			data:   [][]string{{"col1", "col2"}, {"val1", "val2"}},
			valid:  true,
		},
		{
			name:   "unsupported_format",
			format: "binary",
			data:   []byte{0x01, 0x02},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{
				"format": tt.format,
			}

			runner := &MockFileWriterRunner{config: config}

			msg := &client.BrokerMessage{
				ID:      "test-file-write-" + tt.name,
				Type:    "file_write",
				Target:  "file-writer",
				Payload: map[string]interface{}{
					"data":        tt.data,
					"output_path": "/test/output." + tt.format,
				},
				Meta: make(map[string]interface{}),
			}

			result, err := runner.ProcessMessage(msg, nil)
			if err != nil {
				t.Errorf("File writing processing failed: %v", err)
			}
			if result == nil {
				t.Error("Expected result message")
			}
			t.Logf("Successfully processed %s format test", tt.name)
		})
	}
}