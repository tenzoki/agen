package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockAdapterRunner implements agent.AgentRunner for testing
type MockAdapterRunner struct {
	config map[string]interface{}
}

func (m *MockAdapterRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockAdapterRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockAdapterRunner) Cleanup(base *agent.BaseAgent) {
}

func TestAdapterInitialization(t *testing.T) {
	config := map[string]interface{}{
		"adapter_type": "http_api",
		"endpoint":     "http://localhost:8080",
		"timeout":      30,
	}

	runner := &MockAdapterRunner{config: config}
	framework := agent.NewFramework(runner, "adapter")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Adapter framework created successfully")
}

func TestAdapterHTTPAPIIntegration(t *testing.T) {
	config := map[string]interface{}{
		"adapter_type": "http_api",
		"endpoint":     "https://api.example.com",
		"authentication": map[string]interface{}{
			"type":  "bearer_token",
			"token": "test_token",
		},
	}

	runner := &MockAdapterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-http-api",
		Type:   "api_request",
		Target: "adapter",
		Payload: map[string]interface{}{
			"method": "GET",
			"path":   "/users",
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("HTTP API adapter processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("HTTP API adapter test completed successfully")
}

func TestAdapterDatabaseIntegration(t *testing.T) {
	config := map[string]interface{}{
		"adapter_type": "database",
		"connection": map[string]interface{}{
			"driver":   "postgresql",
			"host":     "localhost",
			"port":     5432,
			"database": "test_db",
		},
	}

	runner := &MockAdapterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-database",
		Type:   "database_query",
		Target: "adapter",
		Payload: map[string]interface{}{
			"operation": "select",
			"query":     "SELECT * FROM users WHERE id = $1",
			"parameters": []interface{}{1},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Database adapter processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Database adapter test completed successfully")
}

func TestAdapterMessageQueueIntegration(t *testing.T) {
	config := map[string]interface{}{
		"adapter_type": "message_queue",
		"broker": map[string]interface{}{
			"type": "rabbitmq",
			"url":  "amqp://localhost:5672",
		},
	}

	runner := &MockAdapterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-message-queue",
		Type:   "queue_message",
		Target: "adapter",
		Payload: map[string]interface{}{
			"operation":    "publish",
			"exchange":     "test_exchange",
			"routing_key":  "test.message",
			"message_body": "Hello, World!",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Message queue adapter processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Message queue adapter test completed successfully")
}

func TestAdapterFileSystemIntegration(t *testing.T) {
	config := map[string]interface{}{
		"adapter_type": "filesystem",
		"source": map[string]interface{}{
			"type": "local",
			"path": "/test/input",
		},
		"destination": map[string]interface{}{
			"type":   "s3",
			"bucket": "test-bucket",
		},
	}

	runner := &MockAdapterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-filesystem",
		Type:   "file_operation",
		Target: "adapter",
		Payload: map[string]interface{}{
			"operation": "copy",
			"source":    "/test/input/file.txt",
			"destination": "s3://test-bucket/processed/file.txt",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Filesystem adapter processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Filesystem adapter test completed successfully")
}

func TestAdapterErrorHandling(t *testing.T) {
	config := map[string]interface{}{
		"adapter_type": "http_api",
		"endpoint":     "http://invalid-endpoint",
		"error_handling": map[string]interface{}{
			"max_retries":   3,
			"retry_delay":   "1s",
			"circuit_breaker": true,
		},
	}

	runner := &MockAdapterRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-error-handling",
		Type:   "api_request",
		Target: "adapter",
		Payload: map[string]interface{}{
			"method": "GET",
			"path":   "/nonexistent",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Error handling test failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Error handling test completed successfully")
}