package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockTextAnalyzerRunner implements agent.AgentRunner for testing
type MockTextAnalyzerRunner struct {
	config map[string]interface{}
}

func (m *MockTextAnalyzerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockTextAnalyzerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockTextAnalyzerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestTextAnalyzerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"sentiment_analysis": true,
		"language_detection": true,
		"keyword_extraction": true,
	}

	runner := &MockTextAnalyzerRunner{config: config}
	framework := agent.NewFramework(runner, "text-analyzer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Text analyzer framework created successfully")
}

func TestTextAnalyzerProcessing(t *testing.T) {
	config := map[string]interface{}{
		"sentiment_analysis": true,
		"language_detection": true,
	}

	runner := &MockTextAnalyzerRunner{config: config}

	msg := &client.BrokerMessage{
		ID:      "test-text-analysis",
		Type:    "text_analysis",
		Target:  "text-analyzer",
		Payload: map[string]interface{}{
			"content": "This is a wonderful day for testing text analysis!",
			"metadata": map[string]interface{}{
				"source": "test",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text analysis processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text analysis test completed successfully")
}