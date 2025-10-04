package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

type BinaryAnalyzer struct {
	agent.DefaultAgentRunner
	config *AnalyzerConfig
}

type AnalyzerConfig struct {
	EnableHashing      bool `json:"enable_hashing"`
	EnableEntropy      bool `json:"enable_entropy"`
	EnableMagicBytes   bool `json:"enable_magic_bytes"`
	MaxAnalysisSize    int  `json:"max_analysis_size"`
	EnableStructural   bool `json:"enable_structural"`
	EnableCompression  bool `json:"enable_compression"`
}

type ChunkProcessingRequest struct {
	RequestID   string                 `json:"request_id"`
	ChunkID     string                 `json:"chunk_id"`
	FileID      string                 `json:"file_id"`
	ChunkIndex  int                    `json:"chunk_index"`
	ContentType string                 `json:"content_type"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Options     map[string]interface{} `json:"options"`
	CreatedAt   time.Time              `json:"created_at"`
}

type ProcessingResult struct {
	RequestID      string                 `json:"request_id"`
	ChunkID        string                 `json:"chunk_id"`
	FileID         string                 `json:"file_id"`
	ChunkIndex     int                    `json:"chunk_index"`
	ProcessedBy    string                 `json:"processed_by"`
	ProcessingType string                 `json:"processing_type"`
	ResultData     map[string]interface{} `json:"result_data"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessedAt    time.Time              `json:"processed_at"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
}

func NewBinaryAnalyzer() *BinaryAnalyzer {
	return &BinaryAnalyzer{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &AnalyzerConfig{
			EnableHashing:     true,
			EnableEntropy:     true,
			EnableMagicBytes:  true,
			MaxAnalysisSize:   10485760, // 10MB
			EnableStructural:  true,
			EnableCompression: false,
		},
	}
}

func (ba *BinaryAnalyzer) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "chunk_processing_request" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var request ChunkProcessingRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	if request.ContentType != "binary" && request.ContentType != "application/octet-stream" {
		return nil, fmt.Errorf("unsupported content type: %s", request.ContentType)
	}

	ba.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &ProcessingResult{
		RequestID:      request.RequestID,
		ChunkID:        request.ChunkID,
		FileID:         request.FileID,
		ChunkIndex:     request.ChunkIndex,
		ProcessedBy:    "binary_analyzer",
		ProcessingType: "binary_analysis",
		ResultData:     make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		ProcessedAt:    startTime,
	}

	data := []byte(request.Content)
	processedData, analysisData, err := ba.analyzeBinary(data)
	if err != nil {
		result.Error = fmt.Sprintf("binary analysis failed: %v", err)
		result.Success = false
		result.ProcessingTime = time.Since(startTime)
		return ba.createResultMessage(result), nil
	}

	result.ResultData["processed_content"] = string(processedData)
	result.ResultData = mergeMapStringInterface(result.ResultData, analysisData)
	result.Metadata["analyzer"] = "binary_analysis"
	result.Metadata["original_length"] = len(data)
	result.Metadata["processed_length"] = len(processedData)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return ba.createResultMessage(result), nil
}

func (ba *BinaryAnalyzer) loadConfigFromAgent(base *agent.BaseAgent) {
	if maxSize := base.GetConfigInt("max_analysis_size", 10485760); maxSize > 0 {
		ba.config.MaxAnalysisSize = maxSize
	}
	ba.config.EnableHashing = base.GetConfigBool("enable_hashing", true)
	ba.config.EnableEntropy = base.GetConfigBool("enable_entropy", true)
	ba.config.EnableMagicBytes = base.GetConfigBool("enable_magic_bytes", true)
	ba.config.EnableStructural = base.GetConfigBool("enable_structural", true)
	ba.config.EnableCompression = base.GetConfigBool("enable_compression", false)
}

func (ba *BinaryAnalyzer) analyzeBinary(data []byte) ([]byte, map[string]interface{}, error) {
	analysisData := make(map[string]interface{})

	// Limit analysis size for performance
	analysisSize := len(data)
	if analysisSize > ba.config.MaxAnalysisSize {
		analysisSize = ba.config.MaxAnalysisSize
	}
	analysisData["analysis_size"] = analysisSize
	analysisData["total_size"] = len(data)

	analysisBytes := data[:analysisSize]

	// Basic binary statistics
	analysisData["size"] = len(data)

	// Hash analysis
	if ba.config.EnableHashing {
		hashes := ba.calculateHashes(analysisBytes)
		analysisData["hashes"] = hashes
	}

	// Magic bytes detection
	if ba.config.EnableMagicBytes {
		fileType, mimeType := ba.detectFileType(analysisBytes)
		analysisData["detected_file_type"] = fileType
		analysisData["detected_mime_type"] = mimeType
	}

	// Entropy analysis
	if ba.config.EnableEntropy {
		entropy := ba.calculateEntropy(analysisBytes)
		analysisData["entropy"] = entropy
		analysisData["entropy_classification"] = ba.classifyEntropy(entropy)
	}

	// Structural analysis
	if ba.config.EnableStructural {
		structural := ba.analyzeStructure(analysisBytes)
		analysisData["structural_analysis"] = structural
	}

	// Byte frequency analysis
	byteFreq := ba.analyzeByteFrequency(analysisBytes)
	analysisData["byte_frequency"] = byteFreq

	// ASCII content analysis
	asciiAnalysis := ba.analyzeASCIIContent(analysisBytes)
	analysisData["ascii_analysis"] = asciiAnalysis

	// Pattern detection
	patterns := ba.detectPatterns(analysisBytes)
	analysisData["patterns"] = patterns

	// Processing metadata
	analysisData["processed_by"] = "binary_analyzer"
	analysisData["processed_at"] = time.Now().Format(time.RFC3339)
	analysisData["processing_type"] = "binary_analysis_and_classification"

	return data, analysisData, nil
}

func (ba *BinaryAnalyzer) calculateHashes(data []byte) map[string]string {
	hashes := make(map[string]string)

	// MD5
	md5Hash := md5.Sum(data)
	hashes["md5"] = hex.EncodeToString(md5Hash[:])

	// SHA256
	sha256Hash := sha256.Sum256(data)
	hashes["sha256"] = hex.EncodeToString(sha256Hash[:])

	return hashes
}

func (ba *BinaryAnalyzer) detectFileType(data []byte) (string, string) {
	if len(data) < 4 {
		return "unknown", "application/octet-stream"
	}

	// Common file signatures
	signatures := map[string][]byte{
		"PDF":  {0x25, 0x50, 0x44, 0x46},                         // %PDF
		"PNG":  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG
		"JPEG": {0xFF, 0xD8, 0xFF},                               // JPEG
		"GIF":  {0x47, 0x49, 0x46, 0x38},                         // GIF8
		"ZIP":  {0x50, 0x4B, 0x03, 0x04},                         // PK..
		"RAR":  {0x52, 0x61, 0x72, 0x21},                         // Rar!
		"7Z":   {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},            // 7z
		"GZIP": {0x1F, 0x8B},                                     // gzip
		"BZ2":  {0x42, 0x5A, 0x68},                               // BZh
		"ELF":  {0x7F, 0x45, 0x4C, 0x46},                         // ELF
		"PE":   {0x4D, 0x5A},                                     // MZ (PE executable)
		"TAR":  {0x75, 0x73, 0x74, 0x61, 0x72},                   // ustar (at offset 257)
	}

	mimeTypes := map[string]string{
		"PDF":  "application/pdf",
		"PNG":  "image/png",
		"JPEG": "image/jpeg",
		"GIF":  "image/gif",
		"ZIP":  "application/zip",
		"RAR":  "application/x-rar-compressed",
		"7Z":   "application/x-7z-compressed",
		"GZIP": "application/gzip",
		"BZ2":  "application/x-bzip2",
		"ELF":  "application/x-executable",
		"PE":   "application/x-msdownload",
		"TAR":  "application/x-tar",
	}

	// Check for TAR at offset 257
	if len(data) > 262 {
		tarSig := signatures["TAR"]
		if ba.bytesMatch(data[257:262], tarSig) {
			return "TAR", mimeTypes["TAR"]
		}
	}

	// Check other signatures at the beginning
	for fileType, signature := range signatures {
		if fileType == "TAR" {
			continue // Already checked above
		}
		if len(data) >= len(signature) && ba.bytesMatch(data[:len(signature)], signature) {
			return fileType, mimeTypes[fileType]
		}
	}

	return "unknown", "application/octet-stream"
}

func (ba *BinaryAnalyzer) bytesMatch(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (ba *BinaryAnalyzer) calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Count byte frequencies
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	// Calculate entropy
	entropy := 0.0
	length := float64(len(data))

	for _, count := range freq {
		if count > 0 {
			probability := float64(count) / length
			entropy -= probability * ba.log2(probability)
		}
	}

	return entropy
}

func (ba *BinaryAnalyzer) log2(x float64) float64 {
	return ba.log(x) / ba.log(2)
}

func (ba *BinaryAnalyzer) log(x float64) float64 {
	// Simple natural logarithm approximation for small values
	if x <= 0 {
		return 0
	}

	// Use Taylor series approximation for ln(1+x) where x = input-1
	if x > 0.5 && x < 1.5 {
		u := x - 1
		result := u
		term := u
		for i := 2; i <= 20; i++ {
			term *= -u
			result += term / float64(i)
		}
		return result
	}

	// For other values, use a simpler approximation
	// This is not mathematically precise but sufficient for entropy calculation
	return 0.69314718 * ba.simpleLog2(x) // ln(2) * log2(x)
}

func (ba *BinaryAnalyzer) simpleLog2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x == 1 {
		return 0
	}

	// Simple bit-shift based approximation
	exp := 0
	for x >= 2 {
		x /= 2
		exp++
	}
	for x < 1 {
		x *= 2
		exp--
	}

	// x is now between 1 and 2
	// Use linear approximation for fractional part
	frac := (x - 1) * 1.44269504 // 1/ln(2)

	return float64(exp) + frac
}

func (ba *BinaryAnalyzer) classifyEntropy(entropy float64) string {
	if entropy < 1.0 {
		return "very_low_entropy"
	} else if entropy < 3.0 {
		return "low_entropy"
	} else if entropy < 6.0 {
		return "medium_entropy"
	} else if entropy < 7.5 {
		return "high_entropy"
	} else {
		return "very_high_entropy"
	}
}

func (ba *BinaryAnalyzer) analyzeStructure(data []byte) map[string]interface{} {
	structure := make(map[string]interface{})

	if len(data) == 0 {
		return structure
	}

	// Null byte analysis
	nullCount := 0
	for _, b := range data {
		if b == 0 {
			nullCount++
		}
	}
	structure["null_bytes"] = nullCount
	structure["null_byte_ratio"] = float64(nullCount) / float64(len(data))

	// Printable character analysis
	printableCount := 0
	for _, b := range data {
		if b >= 32 && b <= 126 {
			printableCount++
		}
	}
	structure["printable_chars"] = printableCount
	structure["printable_ratio"] = float64(printableCount) / float64(len(data))

	// Consecutive byte patterns
	maxRun := 0
	currentRun := 1
	for i := 1; i < len(data); i++ {
		if data[i] == data[i-1] {
			currentRun++
		} else {
			if currentRun > maxRun {
				maxRun = currentRun
			}
			currentRun = 1
		}
	}
	if currentRun > maxRun {
		maxRun = currentRun
	}
	structure["max_consecutive_bytes"] = maxRun

	return structure
}

func (ba *BinaryAnalyzer) analyzeByteFrequency(data []byte) map[string]interface{} {
	frequency := make(map[byte]int)
	for _, b := range data {
		frequency[b]++
	}

	// Find most and least common bytes
	var mostCommonByte, leastCommonByte byte
	maxCount, minCount := 0, len(data)+1

	for b, count := range frequency {
		if count > maxCount {
			maxCount = count
			mostCommonByte = b
		}
		if count < minCount {
			minCount = count
			leastCommonByte = b
		}
	}

	result := map[string]interface{}{
		"unique_bytes":      len(frequency),
		"most_common_byte":  int(mostCommonByte),
		"most_common_count": maxCount,
		"least_common_byte": int(leastCommonByte),
		"least_common_count": minCount,
	}

	// Calculate distribution uniformity
	expectedCount := float64(len(data)) / 256.0
	deviation := 0.0
	for _, count := range frequency {
		diff := float64(count) - expectedCount
		deviation += diff * diff
	}
	result["distribution_variance"] = deviation / float64(len(frequency))

	return result
}

func (ba *BinaryAnalyzer) analyzeASCIIContent(data []byte) map[string]interface{} {
	analysis := make(map[string]interface{})

	asciiCount := 0
	controlCount := 0
	extendedCount := 0
	whitespaceCount := 0

	for _, b := range data {
		if b < 32 {
			controlCount++
			if b == 9 || b == 10 || b == 13 || b == 32 { // Tab, LF, CR, Space
				whitespaceCount++
			}
		} else if b <= 126 {
			asciiCount++
		} else {
			extendedCount++
		}
	}

	total := float64(len(data))
	analysis["ascii_chars"] = asciiCount
	analysis["control_chars"] = controlCount
	analysis["extended_chars"] = extendedCount
	analysis["whitespace_chars"] = whitespaceCount
	analysis["ascii_ratio"] = float64(asciiCount) / total
	analysis["control_ratio"] = float64(controlCount) / total
	analysis["extended_ratio"] = float64(extendedCount) / total

	// Determine if content looks like text
	textLikeRatio := float64(asciiCount+whitespaceCount) / total
	analysis["text_like_ratio"] = textLikeRatio
	analysis["likely_text"] = textLikeRatio > 0.7

	return analysis
}

func (ba *BinaryAnalyzer) detectPatterns(data []byte) map[string]interface{} {
	patterns := make(map[string]interface{})

	if len(data) < 4 {
		return patterns
	}

	// Look for repeated patterns
	patterns["has_repeated_sequences"] = ba.hasRepeatedSequences(data)
	patterns["has_header_structure"] = ba.hasHeaderStructure(data)
	patterns["has_regular_intervals"] = ba.hasRegularIntervals(data)
	patterns["appears_compressed"] = ba.appearsCompressed(data)
	patterns["appears_encrypted"] = ba.appearsEncrypted(data)

	return patterns
}

func (ba *BinaryAnalyzer) hasRepeatedSequences(data []byte) bool {
	if len(data) < 8 {
		return false
	}

	// Look for 4-byte sequences that repeat
	sequences := make(map[string]int)
	for i := 0; i <= len(data)-4; i++ {
		seq := string(data[i : i+4])
		sequences[seq]++
		if sequences[seq] > 2 {
			return true
		}
	}
	return false
}

func (ba *BinaryAnalyzer) hasHeaderStructure(data []byte) bool {
	if len(data) < 16 {
		return false
	}

	// Look for null-terminated strings in the first 256 bytes
	headerSize := 256
	if len(data) < headerSize {
		headerSize = len(data)
	}

	nullCount := 0
	for i := 0; i < headerSize; i++ {
		if data[i] == 0 {
			nullCount++
		}
	}

	// If more than 10% of header is null bytes, might be structured
	return float64(nullCount)/float64(headerSize) > 0.1
}

func (ba *BinaryAnalyzer) hasRegularIntervals(data []byte) bool {
	if len(data) < 32 {
		return false
	}

	// Look for bytes that appear at regular intervals
	for interval := 4; interval <= 16; interval++ {
		matches := 0
		for i := 0; i+interval*3 < len(data); i += interval {
			if data[i] == data[i+interval] && data[i] == data[i+interval*2] {
				matches++
			}
		}
		if matches > 3 {
			return true
		}
	}
	return false
}

func (ba *BinaryAnalyzer) appearsCompressed(data []byte) bool {
	// High entropy and low repetition suggests compression
	entropy := ba.calculateEntropy(data)
	return entropy > 7.0
}

func (ba *BinaryAnalyzer) appearsEncrypted(data []byte) bool {
	// Very high entropy and uniform distribution suggests encryption
	entropy := ba.calculateEntropy(data)
	if entropy < 7.5 {
		return false
	}

	// Check for uniform byte distribution
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	// Calculate chi-square test for uniformity
	expected := float64(len(data)) / 256.0
	chiSquare := 0.0
	for i := 0; i < 256; i++ {
		observed := float64(freq[byte(i)])
		diff := observed - expected
		chiSquare += (diff * diff) / expected
	}

	// If chi-square is low, distribution is uniform (possibly encrypted)
	return chiSquare < 300.0 // Threshold for "uniform enough"
}

func (ba *BinaryAnalyzer) createResultMessage(result *ProcessingResult) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "chunk_processing_result",
		Target:    "chunk_processing_result",
		Payload:   result,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func mergeMapStringInterface(dst, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func main() {
	binaryAnalyzer := NewBinaryAnalyzer()
	agent.Run(binaryAnalyzer, "binary-analyzer")
}