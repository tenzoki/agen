package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/daulet/tokenizers"
	ort "github.com/yalue/onnxruntime_go"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// NERAgent performs Named Entity Recognition using ONNXRuntime
type NERAgent struct {
	agent.DefaultAgentRunner
	session   *ort.AdvancedSession
	tokenizer *tokenizers.Tokenizer
	config    *NERConfig
	labelMap  map[int]string
}

// NERConfig holds configuration for NER agent
type NERConfig struct {
	ModelPath      string  `yaml:"model_path"`
	TokenizerPath  string  `yaml:"tokenizer_path"`
	MaxSeqLength   int     `yaml:"max_seq_length"`
	ConfThreshold  float32 `yaml:"confidence_threshold"`
	EnableDebug    bool    `yaml:"enable_debug"`
}

// Entity represents a named entity
type Entity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"`        // PERSON, ORG, LOC, MISC
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Confidence float32 `json:"confidence"`
}

// NERRequest represents an NER extraction request
type NERRequest struct {
	Text      string `json:"text"`
	Language  string `json:"language,omitempty"` // Optional language hint
	ProjectID string `json:"project_id"`
}

// NERResponse represents an NER extraction response
type NERResponse struct {
	Text     string   `json:"text"`
	Entities []Entity `json:"entities"`
	Count    int      `json:"count"`
	Language string   `json:"language,omitempty"`
}

// TokenizerOutput represents tokenized text with offsets
type TokenizerOutput struct {
	InputIDs      []int32
	AttentionMask []int32
	Offsets       [][]int // Character offsets for each token [start, end]
}

// Init initializes the NER agent
func (n *NERAgent) Init(base *agent.BaseAgent) error {
	// Load configuration
	config := &NERConfig{
		ModelPath:     base.GetConfigString("model_path", "models/ner/xlm-roberta-ner.onnx"),
		TokenizerPath: base.GetConfigString("tokenizer_path", "models/ner/"),
		MaxSeqLength:  base.GetConfigInt("max_seq_length", 128),
		ConfThreshold: 0.5, // Default confidence threshold
		EnableDebug:   base.GetConfigBool("enable_debug", false),
	}

	// Override threshold if provided in config
	if thresholdVal, ok := base.Config["confidence_threshold"]; ok {
		if threshold, ok := thresholdVal.(float64); ok {
			config.ConfThreshold = float32(threshold)
		}
	}
	n.config = config

	// Check if model file exists
	if _, err := os.Stat(config.ModelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s (run model conversion script first)", config.ModelPath)
	}

	// Initialize ONNXRuntime environment
	if err := ort.InitializeEnvironment(); err != nil {
		return fmt.Errorf("failed to initialize ONNXRuntime: %w", err)
	}

	// Load ONNX model
	base.LogInfo("Loading NER model from %s", config.ModelPath)
	session, err := ort.NewAdvancedSession(
		config.ModelPath,
		[]string{"input_ids", "attention_mask"},
		[]string{"logits"},
		[]ort.Value{},
		[]ort.Value{},
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to load ONNX model: %w", err)
	}
	n.session = session

	// Load HuggingFace tokenizer
	tokenizerPath := filepath.Join(config.TokenizerPath, "tokenizer.json")
	if _, err := os.Stat(tokenizerPath); os.IsNotExist(err) {
		return fmt.Errorf("tokenizer file not found: %s", tokenizerPath)
	}

	tk, err := tokenizers.FromFile(tokenizerPath)
	if err != nil {
		return fmt.Errorf("failed to load tokenizer: %w", err)
	}
	n.tokenizer = tk

	// Load label mapping from config.json
	configPath := filepath.Join(config.TokenizerPath, "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config.json: %w", err)
	}

	var modelConfig struct {
		ID2Label map[string]string `json:"id2label"`
	}
	if err := json.Unmarshal(configData, &modelConfig); err != nil {
		return fmt.Errorf("failed to parse config.json: %w", err)
	}

	// Convert string keys to int
	n.labelMap = make(map[int]string)
	for idStr, label := range modelConfig.ID2Label {
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		n.labelMap[id] = label
	}

	base.LogInfo("NER agent initialized")
	base.LogInfo("Model: %s", filepath.Base(config.ModelPath))
	base.LogInfo("Tokenizer: %s", filepath.Base(tokenizerPath))
	base.LogInfo("Labels: %d classes", len(n.labelMap))
	base.LogInfo("Max sequence length: %d", config.MaxSeqLength)
	base.LogInfo("Confidence threshold: %.2f", config.ConfThreshold)

	return nil
}

// ProcessMessage processes incoming text and extracts entities
func (n *NERAgent) ProcessMessage(
	msg *client.BrokerMessage,
	base *agent.BaseAgent,
) (*client.BrokerMessage, error) {
	// Parse request
	var req NERRequest
	payloadBytes, ok := msg.Payload.([]byte)
	if !ok {
		return n.errorResponse("invalid payload type"), nil
	}

	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		// Try plain text
		req.Text = string(payloadBytes)
	}

	if req.Text == "" {
		return n.errorResponse("empty text"), nil
	}

	if n.config.EnableDebug {
		base.LogInfo("Processing text (%d chars)", len(req.Text))
	}

	// Extract entities
	entities, err := n.extractEntities(req.Text, base)
	if err != nil {
		base.LogError("Entity extraction failed: %v", err)
		return n.errorResponse(fmt.Sprintf("extraction failed: %v", err)), nil
	}

	// Create response
	response := NERResponse{
		Text:     req.Text,
		Entities: entities,
		Count:    len(entities),
		Language: req.Language,
	}

	payload, err := json.Marshal(response)
	if err != nil {
		return n.errorResponse(fmt.Sprintf("failed to serialize response: %v", err)), nil
	}

	if n.config.EnableDebug {
		base.LogInfo("Extracted %d entities", len(entities))
	}

	return &client.BrokerMessage{
		Payload: payload,
	}, nil
}

// extractEntities performs NER using the ONNX model
func (n *NERAgent) extractEntities(text string, base *agent.BaseAgent) ([]Entity, error) {
	// Tokenize text using HuggingFace tokenizer
	encoding := n.tokenizer.EncodeWithOptions(text, true, tokenizers.WithReturnOffsets())

	// Get token IDs and attention mask
	inputIDs := encoding.IDs
	attentionMask := encoding.AttentionMask
	offsets := encoding.Offsets

	// Truncate or pad to max sequence length
	maxLen := n.config.MaxSeqLength
	if len(inputIDs) > maxLen {
		inputIDs = inputIDs[:maxLen]
		attentionMask = attentionMask[:maxLen]
		offsets = offsets[:maxLen]
	} else if len(inputIDs) < maxLen {
		// Pad with pad token (1 for XLM-RoBERTa)
		for len(inputIDs) < maxLen {
			inputIDs = append(inputIDs, 1)
			attentionMask = append(attentionMask, 0)
			offsets = append(offsets, tokenizers.Offset{0, 0})
		}
	}

	// Convert to TokenizerOutput for later use
	tokens := &TokenizerOutput{
		InputIDs:      make([]int32, len(inputIDs)),
		AttentionMask: make([]int32, len(attentionMask)),
		Offsets:       make([][]int, len(offsets)),
	}

	for i := range inputIDs {
		tokens.InputIDs[i] = int32(inputIDs[i])
		tokens.AttentionMask[i] = int32(attentionMask[i])
		tokens.Offsets[i] = []int{int(offsets[i][0]), int(offsets[i][1])}
	}

	// Prepare input tensors for ONNX (int64 for input)
	inputIDsInt64 := make([]int64, maxLen)
	attentionMaskInt64 := make([]int64, maxLen)
	for i := range inputIDs {
		inputIDsInt64[i] = int64(inputIDs[i])
		attentionMaskInt64[i] = int64(attentionMask[i])
	}

	// Create input tensors
	inputIDsTensor, err := ort.NewTensor(
		ort.NewShape(1, int64(maxLen)),
		inputIDsInt64,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Destroy()

	attentionMaskTensor, err := ort.NewTensor(
		ort.NewShape(1, int64(maxLen)),
		attentionMaskInt64,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Destroy()

	// Prepare output tensor (8 labels for this model)
	numLabels := len(n.labelMap)
	outputData := make([]float32, maxLen*numLabels)
	outputTensor, err := ort.NewTensor(
		ort.NewShape(1, int64(maxLen), int64(numLabels)),
		outputData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	// Create temporary session for this inference
	// (AdvancedSession requires inputs/outputs at construction time)
	tempSession, err := ort.NewAdvancedSession(
		n.config.ModelPath,
		[]string{"input_ids", "attention_mask"},
		[]string{"logits"},
		[]ort.Value{inputIDsTensor, attentionMaskTensor},
		[]ort.Value{outputTensor},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create inference session: %w", err)
	}
	defer tempSession.Destroy()

	// Run inference (updates outputData slice in-place)
	if err = tempSession.Run(); err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Decode predictions using the updated outputData
	entities := n.decodePredictionsFromData(outputData, maxLen, numLabels, tokens, text, base)

	return entities, nil
}

// decodePredictionsFromData converts model output data to entities using BIO tag decoding
func (n *NERAgent) decodePredictionsFromData(
	logitsData []float32,
	seqLen int,
	numLabels int,
	tokens *TokenizerOutput,
	text string,
	base *agent.BaseAgent,
) []Entity {
	if n.config.EnableDebug {
		base.LogInfo("Decoding predictions: seq_len=%d, num_labels=%d", seqLen, numLabels)
	}

	// Find argmax for each token (predicted label ID)
	predictions := make([]int, seqLen)
	confidences := make([]float32, seqLen)

	for i := 0; i < seqLen; i++ {
		maxIdx := 0
		maxVal := logitsData[i*numLabels+0]

		// Calculate softmax for confidence (simplified - just use max logit value)
		expSum := float32(0.0)
		for j := 0; j < numLabels; j++ {
			val := logitsData[i*numLabels+j]
			expSum += float32(math.Exp(float64(val)))
			if val > maxVal {
				maxVal = val
				maxIdx = j
			}
		}

		predictions[i] = maxIdx
		// Softmax probability for the max class
		confidences[i] = float32(math.Exp(float64(maxVal))) / expSum
	}

	// Decode BIO tags and merge into entities
	entities := []Entity{}
	var currentEntity *Entity

	for i := 0; i < seqLen; i++ {
		// Skip special tokens (first and last are <s> and </s> for XLM-RoBERTa)
		if i == 0 || i >= seqLen-1 {
			continue
		}

		// Skip padding tokens
		if tokens.AttentionMask[i] == 0 {
			continue
		}

		predLabel := n.labelMap[predictions[i]]
		confidence := confidences[i]

		// Skip if below confidence threshold
		if confidence < n.config.ConfThreshold {
			if currentEntity != nil {
				entities = append(entities, *currentEntity)
				currentEntity = nil
			}
			continue
		}

		// Parse BIO tag
		if predLabel == "O" {
			// End current entity if any
			if currentEntity != nil {
				entities = append(entities, *currentEntity)
				currentEntity = nil
			}
			continue
		}

		// Extract position (B or I) and entity type
		parts := strings.Split(predLabel, "-")
		if len(parts) != 2 {
			continue
		}

		position := parts[0] // B or I
		entityType := parts[1]

		// Normalize entity type to standard names
		normalizedType := normalizeEntityType(entityType)

		// Get character offsets
		charStart := tokens.Offsets[i][0]
		charEnd := tokens.Offsets[i][1]

		// Skip tokens with zero-length spans (subword tokens)
		if charStart >= charEnd {
			continue
		}

		if position == "B" {
			// Begin new entity
			if currentEntity != nil {
				entities = append(entities, *currentEntity)
			}

			// Extract entity text from original text
			entityText := text[charStart:charEnd]

			currentEntity = &Entity{
				Text:       entityText,
				Type:       normalizedType,
				Start:      charStart,
				End:        charEnd,
				Confidence: confidence,
			}
		} else if position == "I" && currentEntity != nil {
			// Continue current entity
			if currentEntity.Type == normalizedType {
				currentEntity.End = charEnd
				currentEntity.Text = text[currentEntity.Start:charEnd]
				// Update confidence (use average)
				currentEntity.Confidence = (currentEntity.Confidence + confidence) / 2.0
			} else {
				// Type mismatch - start new entity
				entities = append(entities, *currentEntity)
				entityText := text[charStart:charEnd]
				currentEntity = &Entity{
					Text:       entityText,
					Type:       normalizedType,
					Start:      charStart,
					End:        charEnd,
					Confidence: confidence,
				}
			}
		}
	}

	// Add last entity if any
	if currentEntity != nil {
		entities = append(entities, *currentEntity)
	}

	if n.config.EnableDebug {
		base.LogInfo("Decoded %d entities", len(entities))
		for _, e := range entities {
			base.LogInfo("  - %s: %s (%.2f)", e.Type, e.Text, e.Confidence)
		}
	}

	return entities
}

// normalizeEntityType converts model-specific entity types to standard names
func normalizeEntityType(entityType string) string {
	switch entityType {
	case "PER":
		return "PERSON"
	case "ORG":
		return "ORG"
	case "LOC":
		return "LOC"
	case "MISC":
		return "MISC"
	default:
		return entityType
	}
}

// errorResponse creates an error response
func (n *NERAgent) errorResponse(errorMsg string) *client.BrokerMessage {
	resp := map[string]interface{}{
		"error":    errorMsg,
		"entities": []Entity{},
		"count":    0,
	}
	payload, _ := json.Marshal(resp)
	return &client.BrokerMessage{
		Payload: payload,
	}
}

// Cleanup releases resources
func (n *NERAgent) Cleanup(base *agent.BaseAgent) {
	if n.session != nil {
		n.session.Destroy()
		base.LogInfo("NER model session destroyed")
	}
	ort.DestroyEnvironment()
	base.LogInfo("ONNXRuntime environment destroyed")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	agent.Run(&NERAgent{}, "ner-agent")
}
