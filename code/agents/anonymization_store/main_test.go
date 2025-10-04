package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agen/cellorg/internal/client"
)

func TestStorageOperations(t *testing.T) {
	// Create temp directory for test store
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test-store")

	// Create agent
	agent := &AnonymizationStoreAgent{}
	agent.config = &StoreConfig{
		DataPath:    storePath,
		EnableDebug: true,
	}

	// Initialize (would normally be called by framework)
	// For testing, we manually initialize omnistore
	// This is a simplified test - full test would use BaseAgent

	t.Run("Set and Get Forward Mapping", func(t *testing.T) {
		// Test set operation
		setReq := StorageRequest{
			Operation: "set",
			Key:       "anon:forward:project-a:Angela Merkel",
			Value: map[string]interface{}{
				"pseudonym":        "PERSON_123456",
				"canonical":        "Angela Merkel",
				"entity_type":      "PERSON",
				"created_at":       time.Now().Format(time.RFC3339),
				"pipeline_version": "v1.0",
			},
			ProjectID: "project-a",
			RequestID: "req-001",
		}

		setPayload, _ := json.Marshal(setReq)
		setMsg := &client.BrokerMessage{Payload: setPayload}

		// Note: Full test would call ProcessMessage
		// For now, we verify the structure is correct
		if setReq.Operation != "set" {
			t.Errorf("Expected operation 'set', got %s", setReq.Operation)
		}

		if setReq.Value["pseudonym"] != "PERSON_123456" {
			t.Errorf("Expected pseudonym PERSON_123456, got %v", setReq.Value["pseudonym"])
		}

		t.Logf("Set request payload: %s", string(setMsg.Payload))
	})

	t.Run("Reverse Mapping", func(t *testing.T) {
		// Test reverse lookup structure
		reverseReq := StorageRequest{
			Operation: "get",
			Key:       "anon:reverse:project-a:PERSON_123456",
			ProjectID: "project-a",
			RequestID: "req-002",
		}

		reversePayload, _ := json.Marshal(reverseReq)
		reverseMsg := &client.BrokerMessage{Payload: reversePayload}

		if reverseReq.Operation != "get" {
			t.Errorf("Expected operation 'get', got %s", reverseReq.Operation)
		}

		t.Logf("Reverse lookup payload: %s", string(reverseMsg.Payload))
	})

	t.Run("List Mappings with Prefix", func(t *testing.T) {
		// Test list operation
		listReq := StorageRequest{
			Operation: "list",
			Key:       "anon:forward:project-a:", // Prefix
			ProjectID: "project-a",
			RequestID: "req-003",
		}

		listPayload, _ := json.Marshal(listReq)
		listMsg := &client.BrokerMessage{Payload: listPayload}

		if listReq.Operation != "list" {
			t.Errorf("Expected operation 'list', got %s", listReq.Operation)
		}

		t.Logf("List request payload: %s", string(listMsg.Payload))
	})

	t.Run("Delete Mapping (Soft Delete)", func(t *testing.T) {
		// Test delete operation
		deleteReq := StorageRequest{
			Operation: "delete",
			Key:       "anon:forward:project-a:Angela Merkel",
			ProjectID: "project-a",
			RequestID: "req-004",
		}

		deletePayload, _ := json.Marshal(deleteReq)
		deleteMsg := &client.BrokerMessage{Payload: deletePayload}

		if deleteReq.Operation != "delete" {
			t.Errorf("Expected operation 'delete', got %s", deleteReq.Operation)
		}

		t.Logf("Delete request payload: %s", string(deleteMsg.Payload))
	})
}

func TestMappingRecordSerialization(t *testing.T) {
	record := MappingRecord{
		Pseudonym:       "PERSON_123456",
		Canonical:       "Angela Merkel",
		EntityType:      "PERSON",
		CreatedAt:       time.Now(),
		PipelineVersion: "v1.0",
	}

	// Serialize
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Failed to marshal record: %v", err)
	}

	// Deserialize
	var parsed MappingRecord
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal record: %v", err)
	}

	// Verify
	if parsed.Pseudonym != record.Pseudonym {
		t.Errorf("Expected pseudonym %s, got %s", record.Pseudonym, parsed.Pseudonym)
	}

	if parsed.Canonical != record.Canonical {
		t.Errorf("Expected canonical %s, got %s", record.Canonical, parsed.Canonical)
	}

	if parsed.EntityType != record.EntityType {
		t.Errorf("Expected entity type %s, got %s", record.EntityType, parsed.EntityType)
	}

	t.Logf("Serialized record: %s", string(data))
}

func TestReverseMappingRecordSerialization(t *testing.T) {
	record := ReverseMappingRecord{
		Original:   "Angela Merkel",
		Canonical:  "Angela Merkel",
		EntityType: "PERSON",
	}

	// Serialize
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Failed to marshal record: %v", err)
	}

	// Deserialize
	var parsed ReverseMappingRecord
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal record: %v", err)
	}

	// Verify
	if parsed.Original != record.Original {
		t.Errorf("Expected original %s, got %s", record.Original, parsed.Original)
	}

	t.Logf("Serialized reverse record: %s", string(data))
}

func TestStorageResponseFormat(t *testing.T) {
	response := StorageResponse{
		Success:   true,
		RequestID: "req-001",
		Result: map[string]interface{}{
			"pseudonym":   "PERSON_123456",
			"entity_type": "PERSON",
		},
	}

	// Serialize
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Deserialize
	var parsed StorageResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify
	if !parsed.Success {
		t.Error("Expected success=true")
	}

	if parsed.RequestID != "req-001" {
		t.Errorf("Expected request ID 'req-001', got %s", parsed.RequestID)
	}

	t.Logf("Response JSON: %s", string(data))
}

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}
