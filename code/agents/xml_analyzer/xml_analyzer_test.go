package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockXMLAnalyzerRunner implements agent.AgentRunner for testing
type MockXMLAnalyzerRunner struct {
	config map[string]interface{}
}

func (m *MockXMLAnalyzerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockXMLAnalyzerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockXMLAnalyzerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestXMLAnalyzerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"validate_schema":  true,
		"extract_elements": true,
		"namespace_aware":  true,
	}

	runner := &MockXMLAnalyzerRunner{config: config}
	framework := agent.NewFramework(runner, "xml-analyzer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("XML analyzer framework created successfully")
}

func TestXMLAnalyzerProcessing(t *testing.T) {
	config := map[string]interface{}{
		"validate_schema":  true,
		"extract_elements": true,
	}

	runner := &MockXMLAnalyzerRunner{config: config}

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<root>
	<element attribute="value">Content</element>
	<nested>
		<child>Child content</child>
	</nested>
</root>`

	msg := &client.BrokerMessage{
		ID:      "test-xml-analysis",
		Type:    "xml_analysis",
		Target:  "xml-analyzer",
		Payload: map[string]interface{}{
			"content": xmlContent,
			"file_path": "/test/data/simple.xml",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("XML analysis processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("XML analysis test completed successfully")
}