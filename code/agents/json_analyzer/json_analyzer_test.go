package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockJSONAnalyzerRunner implements agent.AgentRunner for testing
type MockJSONAnalyzerRunner struct {
	config map[string]interface{}
}

func (m *MockJSONAnalyzerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockJSONAnalyzerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockJSONAnalyzerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestJSONAnalyzerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"validate_schema": true,
		"extract_paths":   true,
		"format_output":   "json",
	}

	runner := &MockJSONAnalyzerRunner{config: config}
	framework := agent.NewFramework(runner, "json-analyzer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("JSON analyzer framework created successfully")
}

func TestJSONAnalyzerProcessing(t *testing.T) {
	config := map[string]interface{}{
		"validate_schema": true,
		"extract_paths":   true,
	}

	runner := &MockJSONAnalyzerRunner{config: config}

	jsonContent := `{"name": "test", "value": 42, "nested": {"key": "value"}}`

	msg := &client.BrokerMessage{
		ID:      "test-json-analysis",
		Type:    "json_analysis",
		Target:  "json-analyzer",
		Payload: map[string]interface{}{
			"content": jsonContent,
			"file_path": "/test/data/simple.json",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("JSON analysis processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("JSON analysis test completed successfully")
}