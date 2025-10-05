package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
	"github.com/tenzoki/agen/omni/public/omnistore"
)

// AnonymizerAgent performs pseudonymization using persistent mappings
type AnonymizerAgent struct {
	agent.DefaultAgentRunner
	omniStore omnistore.OmniStore
	config    *AnonymizerConfig
}

// AnonymizerConfig holds configuration for anonymizer
type AnonymizerConfig struct {
	DataPath        string `yaml:"data_path"`
	PipelineVersion string `yaml:"pipeline_version"`
	EnableDebug     bool   `yaml:"enable_debug"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
}

// Entity represents a named entity from NER
type Entity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Confidence float32 `json:"confidence"`
}

// AnonymizerRequest represents an anonymization request
type AnonymizerRequest struct {
	Text      string   `json:"text"`
	Entities  []Entity `json:"entities"`
	ProjectID string   `json:"project_id"`
}

// AnonymizerResponse represents an anonymization response
type AnonymizerResponse struct {
	OriginalText    string            `json:"original_text"`
	AnonymizedText  string            `json:"anonymized_text"`
	EntityCount     int               `json:"entity_count"`
	Mappings        map[string]string `json:"mappings"` // original → pseudonym
	ProcessedAt     string            `json:"processed_at"`
}

// Init initializes the anonymizer agent
func (a *AnonymizerAgent) Init(base *agent.BaseAgent) error {
	// Load configuration
	config := &AnonymizerConfig{
		DataPath:        base.GetConfigString("data_path", "./data/anonymizer"),
		PipelineVersion: base.GetConfigString("pipeline_version", "v1.0"),
		EnableDebug:     base.GetConfigBool("enable_debug", false),
		TimeoutSeconds:  base.GetConfigInt("timeout_seconds", 30),
	}
	a.config = config

	// Initialize omnistore
	store, err := omnistore.NewOmniStoreWithDefaults(config.DataPath)
	if err != nil {
		return fmt.Errorf("failed to initialize omnistore: %w", err)
	}
	a.omniStore = store

	base.LogInfo("Anonymizer initialized")
	base.LogInfo("Data path: %s", config.DataPath)
	base.LogInfo("Pipeline version: %s", config.PipelineVersion)

	return nil
}

// ProcessMessage processes anonymization requests
func (a *AnonymizerAgent) ProcessMessage(
	msg *client.BrokerMessage,
	base *agent.BaseAgent,
) (*client.BrokerMessage, error) {
	// Parse request
	var req AnonymizerRequest
	payload, ok := msg.Payload.([]byte)
	if !ok {
		return a.errorResponse("invalid payload type"), nil
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return a.errorResponse("invalid request format"), nil
	}

	if req.Text == "" {
		return a.errorResponse("empty text"), nil
	}

	if a.config.EnableDebug {
		base.LogInfo("Anonymizing text with %d entities", len(req.Entities))
	}

	// Get project ID from message metadata or config
	projectID := req.ProjectID
	if projectID == "" {
		projectID = "default"
	}

	// Get or create pseudonyms for each entity
	mappings := make(map[string]string)
	for _, entity := range req.Entities {
		pseudonym, err := a.getOrCreatePseudonym(entity, projectID, base)
		if err != nil {
			base.LogError("Failed to get pseudonym for %s: %v", entity.Text, err)
			// Continue with other entities
			continue
		}
		mappings[entity.Text] = pseudonym
	}

	// Replace entities in text
	anonymizedText := a.replaceEntities(req.Text, req.Entities, mappings)

	// Create response
	response := AnonymizerResponse{
		OriginalText:   req.Text,
		AnonymizedText: anonymizedText,
		EntityCount:    len(req.Entities),
		Mappings:       mappings,
		ProcessedAt:    time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(response)
	if err != nil {
		return a.errorResponse(fmt.Sprintf("failed to serialize response: %v", err)), nil
	}

	if a.config.EnableDebug {
		base.LogInfo("Anonymization complete: %d entities replaced", len(mappings))
	}

	return &client.BrokerMessage{
		Payload: payload,
	}, nil
}

// getOrCreatePseudonym gets existing pseudonym or creates a new one
func (a *AnonymizerAgent) getOrCreatePseudonym(
	entity Entity,
	projectID string,
	base *agent.BaseAgent,
) (string, error) {
	// Build storage key
	forwardKey := fmt.Sprintf("anon:forward:%s:%s", projectID, entity.Text)

	// Try to lookup existing pseudonym
	result, err := a.omniStore.KV().Get(forwardKey)
	if err == nil && result != nil {
		// Found existing mapping - unmarshal JSON
		var mappingData map[string]interface{}
		if err := json.Unmarshal(result, &mappingData); err == nil {
			if pseudonym, ok := mappingData["pseudonym"].(string); ok {
				if a.config.EnableDebug {
					base.LogInfo("Reusing pseudonym for %s: %s", entity.Text, pseudonym)
				}
				return pseudonym, nil
			}
		}
	}

	// Generate new pseudonym
	pseudonym := GeneratePseudonym(entity.Type, entity.Text)

	if a.config.EnableDebug {
		base.LogInfo("Generated new pseudonym for %s: %s", entity.Text, pseudonym)
	}

	// Store forward mapping (original → pseudonym)
	forwardValue := map[string]interface{}{
		"pseudonym":        pseudonym,
		"canonical":        entity.Text,
		"entity_type":      entity.Type,
		"created_at":       time.Now().Format(time.RFC3339),
		"pipeline_version": a.config.PipelineVersion,
		"confidence":       entity.Confidence,
	}

	// Marshal forward value to JSON
	forwardBytes, err := json.Marshal(forwardValue)
	if err != nil {
		return "", fmt.Errorf("failed to marshal forward mapping: %w", err)
	}

	if err := a.omniStore.KV().Set(forwardKey, forwardBytes); err != nil {
		return "", fmt.Errorf("failed to store forward mapping: %w", err)
	}

	// Store reverse mapping (pseudonym → original)
	reverseKey := fmt.Sprintf("anon:reverse:%s:%s", projectID, pseudonym)
	reverseValue := map[string]interface{}{
		"original":    entity.Text,
		"canonical":   entity.Text,
		"entity_type": entity.Type,
	}

	// Marshal reverse value to JSON
	reverseBytes, err := json.Marshal(reverseValue)
	if err != nil {
		base.LogInfo("Failed to marshal reverse mapping: %v", err)
		return pseudonym, nil // Not critical
	}

	if err := a.omniStore.KV().Set(reverseKey, reverseBytes); err != nil {
		base.LogInfo("Failed to store reverse mapping: %v", err)
		// Not critical - forward mapping is stored
	}

	return pseudonym, nil
}

// replaceEntities replaces entities in text with pseudonyms
func (a *AnonymizerAgent) replaceEntities(
	text string,
	entities []Entity,
	mappings map[string]string,
) string {
	// Sort entities by position (reverse order to preserve positions)
	sortedEntities := make([]Entity, len(entities))
	copy(sortedEntities, entities)
	sort.Slice(sortedEntities, func(i, j int) bool {
		return sortedEntities[i].Start > sortedEntities[j].Start
	})

	// Build result using string builder
	result := []rune(text)

	for _, entity := range sortedEntities {
		pseudonym, ok := mappings[entity.Text]
		if !ok {
			continue // Skip if no mapping
		}

		// Replace entity text with pseudonym
		before := result[:entity.Start]
		after := result[entity.End:]
		result = append(append(before, []rune(pseudonym)...), after...)
	}

	return string(result)
}

// GeneratePseudonym generates a deterministic pseudonym for an entity
func GeneratePseudonym(entityType, text string) string {
	// Normalize text (lowercase, trim)
	normalized := strings.TrimSpace(strings.ToLower(text))

	// Generate deterministic hash-based ID
	h := sha256.Sum256([]byte(normalized))
	id := binary.BigEndian.Uint64(h[:8]) % 1000000

	// Format pseudonym with entity type prefix
	return fmt.Sprintf("%s_%06d", entityType, id)
}

// errorResponse creates an error response
func (a *AnonymizerAgent) errorResponse(errorMsg string) *client.BrokerMessage {
	resp := map[string]interface{}{
		"error":            errorMsg,
		"anonymized_text":  "",
		"entity_count":     0,
	}
	payload, _ := json.Marshal(resp)
	return &client.BrokerMessage{
		Payload: payload,
	}
}

// Cleanup releases resources
func (a *AnonymizerAgent) Cleanup(base *agent.BaseAgent) {
	if a.omniStore != nil {
		if err := a.omniStore.Close(); err != nil {
			base.LogInfo("Error closing omnistore: %v", err)
		}
	}
	base.LogInfo("Anonymizer agent cleanup complete")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	agent.Run(&AnonymizerAgent{}, "anonymizer")
}
