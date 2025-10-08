package envelope

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ChunkEnvelope splits a large envelope into manageable chunks based on token budget
// Returns slice of envelopes, each fitting within the target model's token limits
func ChunkEnvelope(env *Envelope, budget *EnvelopeBudget) ([]*Envelope, error) {
	if !budget.NeedsSplitting {
		return []*Envelope{env}, nil
	}

	// Determine payload type and split accordingly
	var chunks [][]byte
	var err error

	if isJSONArray(env.Payload) {
		chunks, err = splitJSONArray(env.Payload, budget.SuggestedChunks)
	} else {
		chunks, err = splitTextPayload(env.Payload, budget.SuggestedChunks)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to split payload: %w", err)
	}

	// Create chunk envelopes
	envelopes := make([]*Envelope, len(chunks))
	chunkID := uuid.New().String() // Group ID for all chunks

	for i, chunk := range chunks {
		envelopes[i] = &Envelope{
			ID:            uuid.New().String(),
			CorrelationID: env.ID, // Link to original envelope
			TraceID:       env.TraceID,
			SpanID:        uuid.New().String(), // New span for each chunk
			Source:        env.Source,
			Destination:   env.Destination,
			MessageType:   env.MessageType,
			Timestamp:     env.Timestamp,
			Payload:       chunk,
			Headers:       copyHeaders(env.Headers),
			Properties:    copyProperties(env.Properties),
			Route:         copyRoute(env.Route),
			TTL:           env.TTL,
			Sequence:      env.Sequence,
			HopCount:      env.HopCount,
		}

		// Add chunk metadata headers
		envelopes[i].Headers["X-Chunk-ID"] = chunkID
		envelopes[i].Headers["X-Chunk-Index"] = strconv.Itoa(i)
		envelopes[i].Headers["X-Chunk-Total"] = strconv.Itoa(len(chunks))
		envelopes[i].Headers["X-Original-ID"] = env.ID
	}

	return envelopes, nil
}

// MergeChunks combines chunked envelopes back into a single envelope
// Used by agents to reassemble chunked messages
func MergeChunks(chunks []*Envelope) (*Envelope, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("cannot merge empty chunk list")
	}

	if len(chunks) == 1 {
		// Single chunk, check if it's actually a chunk or standalone
		if chunks[0].Headers["X-Chunk-ID"] == "" {
			return chunks[0], nil
		}
	}

	// Validate all chunks belong to same group
	chunkID := chunks[0].Headers["X-Chunk-ID"]
	if chunkID == "" {
		return nil, fmt.Errorf("first chunk missing X-Chunk-ID header")
	}

	for i, chunk := range chunks {
		if chunk.Headers["X-Chunk-ID"] != chunkID {
			return nil, fmt.Errorf("chunk %d has different chunk ID: %s vs %s",
				i, chunk.Headers["X-Chunk-ID"], chunkID)
		}
	}

	// Sort chunks by index
	sortedChunks := make([]*Envelope, len(chunks))
	copy(sortedChunks, chunks)

	for i := 0; i < len(sortedChunks); i++ {
		for j := i + 1; j < len(sortedChunks); j++ {
			idxI, _ := strconv.Atoi(sortedChunks[i].Headers["X-Chunk-Index"])
			idxJ, _ := strconv.Atoi(sortedChunks[j].Headers["X-Chunk-Index"])
			if idxI > idxJ {
				sortedChunks[i], sortedChunks[j] = sortedChunks[j], sortedChunks[i]
			}
		}
	}

	// Verify we have all chunks
	expectedTotal, _ := strconv.Atoi(sortedChunks[0].Headers["X-Chunk-Total"])
	if len(sortedChunks) != expectedTotal {
		return nil, fmt.Errorf("missing chunks: have %d, expected %d",
			len(sortedChunks), expectedTotal)
	}

	// Merge payloads
	merged := mergePayloads(sortedChunks)

	// Create merged envelope (use first chunk as template)
	result := &Envelope{
		ID:            sortedChunks[0].Headers["X-Original-ID"],
		CorrelationID: sortedChunks[0].CorrelationID,
		TraceID:       sortedChunks[0].TraceID,
		SpanID:        uuid.New().String(), // New span for merged envelope
		Source:        sortedChunks[0].Source,
		Destination:   sortedChunks[0].Destination,
		MessageType:   sortedChunks[0].MessageType,
		Timestamp:     sortedChunks[0].Timestamp,
		Payload:       merged,
		Headers:       copyHeaders(sortedChunks[0].Headers),
		Properties:    copyProperties(sortedChunks[0].Properties),
		Route:         copyRoute(sortedChunks[0].Route),
		TTL:           sortedChunks[0].TTL,
		Sequence:      sortedChunks[0].Sequence,
		HopCount:      sortedChunks[0].HopCount,
	}

	// Remove chunk headers from merged envelope
	delete(result.Headers, "X-Chunk-ID")
	delete(result.Headers, "X-Chunk-Index")
	delete(result.Headers, "X-Chunk-Total")
	delete(result.Headers, "X-Original-ID")

	return result, nil
}

// isJSONArray checks if payload is a JSON array
func isJSONArray(payload []byte) bool {
	var arr []interface{}
	return json.Unmarshal(payload, &arr) == nil
}

// splitJSONArray splits a JSON array into N chunks
func splitJSONArray(payload []byte, numChunks int) ([][]byte, error) {
	var arr []interface{}
	if err := json.Unmarshal(payload, &arr); err != nil {
		return nil, fmt.Errorf("invalid JSON array: %w", err)
	}

	if len(arr) == 0 {
		return [][]byte{payload}, nil
	}

	// Calculate chunk size
	chunkSize := int(math.Ceil(float64(len(arr)) / float64(numChunks)))
	if chunkSize < 1 {
		chunkSize = 1
	}

	chunks := make([][]byte, 0, numChunks)
	for i := 0; i < len(arr); i += chunkSize {
		end := i + chunkSize
		if end > len(arr) {
			end = len(arr)
		}

		chunk := arr[i:end]
		chunkBytes, err := json.Marshal(chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal chunk: %w", err)
		}

		chunks = append(chunks, chunkBytes)
	}

	return chunks, nil
}

// splitTextPayload splits text into N chunks with approximate word boundaries
func splitTextPayload(payload []byte, numChunks int) ([][]byte, error) {
	text := string(payload)
	if len(text) == 0 {
		return [][]byte{payload}, nil
	}

	// Calculate approximate chunk size
	chunkSize := len(text) / numChunks
	if chunkSize < 100 {
		chunkSize = 100 // Minimum chunk size
	}

	chunks := make([][]byte, 0, numChunks)
	start := 0

	for start < len(text) {
		end := start + chunkSize
		if end >= len(text) {
			// Last chunk
			chunks = append(chunks, []byte(text[start:]))
			break
		}

		// Find word boundary near end
		end = findWordBoundary(text, end)
		if end <= start {
			// Couldn't find boundary, force split
			end = start + chunkSize
		}

		chunks = append(chunks, []byte(text[start:end]))
		start = end
	}

	return chunks, nil
}

// findWordBoundary finds the nearest word boundary after pos
// Searches within ±100 chars of pos
func findWordBoundary(text string, pos int) int {
	if pos >= len(text) {
		return len(text)
	}

	// Search forward for whitespace
	for i := pos; i < len(text) && i < pos+100; i++ {
		if isWhitespace(text[i]) {
			return i
		}
	}

	// Search backward for whitespace
	for i := pos; i > 0 && i > pos-100; i-- {
		if isWhitespace(text[i]) {
			return i
		}
	}

	// No boundary found, return original position
	return pos
}

// isWhitespace checks if character is whitespace
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// mergePayloads combines payloads from sorted chunks
func mergePayloads(chunks []*Envelope) []byte {
	if len(chunks) == 0 {
		return []byte("{}")
	}

	// Check if payloads are JSON arrays
	if isJSONArray(chunks[0].Payload) {
		return mergeJSONArrays(chunks)
	}

	// Merge as text
	return mergeTextPayloads(chunks)
}

// mergeJSONArrays combines JSON array chunks
func mergeJSONArrays(chunks []*Envelope) []byte {
	var combined []interface{}

	for _, chunk := range chunks {
		var arr []interface{}
		if err := json.Unmarshal(chunk.Payload, &arr); err != nil {
			// Fallback to text merge on error
			return mergeTextPayloads(chunks)
		}
		combined = append(combined, arr...)
	}

	merged, err := json.Marshal(combined)
	if err != nil {
		// Should never happen, but fallback to text merge
		return mergeTextPayloads(chunks)
	}

	return merged
}

// mergeTextPayloads concatenates text chunks
func mergeTextPayloads(chunks []*Envelope) []byte {
	var builder strings.Builder
	for _, chunk := range chunks {
		builder.Write(chunk.Payload)
	}
	return []byte(builder.String())
}

// Helper functions for copying envelope fields

func copyHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return make(map[string]string)
	}
	copied := make(map[string]string, len(headers))
	for k, v := range headers {
		copied[k] = v
	}
	return copied
}

func copyProperties(props map[string]interface{}) map[string]interface{} {
	if props == nil {
		return make(map[string]interface{})
	}
	copied := make(map[string]interface{}, len(props))
	for k, v := range props {
		copied[k] = v
	}
	return copied
}

func copyRoute(route []string) []string {
	if route == nil {
		return make([]string, 0)
	}
	copied := make([]string, len(route))
	copy(copied, route)
	return copied
}
