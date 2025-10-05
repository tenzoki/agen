package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

type ImageAnalyzer struct {
	agent.DefaultAgentRunner
	config *AnalyzerConfig
}

type AnalyzerConfig struct {
	EnableMetadata     bool `json:"enable_metadata"`
	EnableDimensions   bool `json:"enable_dimensions"`
	EnableColorAnalysis bool `json:"enable_color_analysis"`
	MaxAnalysisSize    int  `json:"max_analysis_size"`
	EnableThumbnail    bool `json:"enable_thumbnail"`
	EnableQuality      bool `json:"enable_quality"`
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

func NewImageAnalyzer() *ImageAnalyzer {
	return &ImageAnalyzer{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &AnalyzerConfig{
			EnableMetadata:      true,
			EnableDimensions:    true,
			EnableColorAnalysis: true,
			MaxAnalysisSize:     10485760, // 10MB
			EnableThumbnail:     false,
			EnableQuality:       true,
		},
	}
}

func (ia *ImageAnalyzer) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
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

	if !strings.HasPrefix(request.ContentType, "image/") {
		return nil, fmt.Errorf("unsupported content type: %s", request.ContentType)
	}

	ia.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &ProcessingResult{
		RequestID:      request.RequestID,
		ChunkID:        request.ChunkID,
		FileID:         request.FileID,
		ChunkIndex:     request.ChunkIndex,
		ProcessedBy:    "image_analyzer",
		ProcessingType: "image_analysis",
		ResultData:     make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		ProcessedAt:    startTime,
	}

	data := []byte(request.Content)
	processedData, analysisData, err := ia.analyzeImage(data, request.ContentType)
	if err != nil {
		result.Error = fmt.Sprintf("image analysis failed: %v", err)
		result.Success = false
		result.ProcessingTime = time.Since(startTime)
		return ia.createResultMessage(result), nil
	}

	result.ResultData["processed_content"] = string(processedData)
	result.ResultData = mergeMapStringInterface(result.ResultData, analysisData)
	result.Metadata["analyzer"] = "image_analysis"
	result.Metadata["original_length"] = len(data)
	result.Metadata["processed_length"] = len(processedData)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return ia.createResultMessage(result), nil
}

func (ia *ImageAnalyzer) loadConfigFromAgent(base *agent.BaseAgent) {
	if maxSize := base.GetConfigInt("max_analysis_size", 10485760); maxSize > 0 {
		ia.config.MaxAnalysisSize = maxSize
	}
	ia.config.EnableMetadata = base.GetConfigBool("enable_metadata", true)
	ia.config.EnableDimensions = base.GetConfigBool("enable_dimensions", true)
	ia.config.EnableColorAnalysis = base.GetConfigBool("enable_color_analysis", true)
	ia.config.EnableThumbnail = base.GetConfigBool("enable_thumbnail", false)
	ia.config.EnableQuality = base.GetConfigBool("enable_quality", true)
}

func (ia *ImageAnalyzer) analyzeImage(data []byte, contentType string) ([]byte, map[string]interface{}, error) {
	analysisData := make(map[string]interface{})

	// Basic image information
	analysisData["file_size"] = len(data)
	analysisData["content_type"] = contentType

	// Detect image format from data
	detectedFormat := ia.detectImageFormat(data)
	analysisData["detected_format"] = detectedFormat

	// Validate format consistency
	expectedFormat := ia.contentTypeToFormat(contentType)
	analysisData["expected_format"] = expectedFormat
	analysisData["format_consistent"] = detectedFormat == expectedFormat

	// Basic image structure analysis
	if ia.config.EnableMetadata {
		metadata := ia.extractBasicMetadata(data, detectedFormat)
		analysisData["metadata"] = metadata
	}

	// Dimension analysis (basic header parsing)
	if ia.config.EnableDimensions {
		dimensions := ia.extractDimensions(data, detectedFormat)
		analysisData["dimensions"] = dimensions
	}

	// Color depth and compression analysis
	colorInfo := ia.analyzeColorInfo(data, detectedFormat)
	analysisData["color_info"] = colorInfo

	// Quality assessment
	if ia.config.EnableQuality {
		quality := ia.assessImageQuality(data, detectedFormat)
		analysisData["quality_assessment"] = quality
	}

	// File structure analysis
	structure := ia.analyzeImageStructure(data, detectedFormat)
	analysisData["structure"] = structure

	// Pattern detection
	patterns := ia.detectImagePatterns(data, detectedFormat)
	analysisData["patterns"] = patterns

	// Processing metadata
	analysisData["processed_by"] = "image_analyzer"
	analysisData["processed_at"] = time.Now().Format(time.RFC3339)
	analysisData["processing_type"] = "image_metadata_and_analysis"

	return data, analysisData, nil
}

func (ia *ImageAnalyzer) detectImageFormat(data []byte) string {
	if len(data) < 8 {
		return "unknown"
	}

	// PNG signature: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "PNG"
	}

	// JPEG signature: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "JPEG"
	}

	// GIF signature: GIF8
	if len(data) >= 4 && string(data[:4]) == "GIF8" {
		return "GIF"
	}

	// BMP signature: BM
	if len(data) >= 2 && data[0] == 0x42 && data[1] == 0x4D {
		return "BMP"
	}

	// WebP signature: RIFF...WEBP
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "WebP"
	}

	// TIFF signatures: II* or MM*
	if len(data) >= 4 {
		if (data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A && data[3] == 0x00) ||
		   (data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00 && data[3] == 0x2A) {
			return "TIFF"
		}
	}

	return "unknown"
}

func (ia *ImageAnalyzer) contentTypeToFormat(contentType string) string {
	switch contentType {
	case "image/png":
		return "PNG"
	case "image/jpeg", "image/jpg":
		return "JPEG"
	case "image/gif":
		return "GIF"
	case "image/bmp":
		return "BMP"
	case "image/webp":
		return "WebP"
	case "image/tiff":
		return "TIFF"
	default:
		return "unknown"
	}
}

func (ia *ImageAnalyzer) extractBasicMetadata(data []byte, format string) map[string]interface{} {
	metadata := make(map[string]interface{})

	switch format {
	case "JPEG":
		metadata = ia.extractJPEGMetadata(data)
	case "PNG":
		metadata = ia.extractPNGMetadata(data)
	case "GIF":
		metadata = ia.extractGIFMetadata(data)
	case "BMP":
		metadata = ia.extractBMPMetadata(data)
	case "WebP":
		metadata = ia.extractWebPMetadata(data)
	case "TIFF":
		metadata = ia.extractTIFFMetadata(data)
	default:
		metadata["format"] = "unsupported"
	}

	return metadata
}

func (ia *ImageAnalyzer) extractJPEGMetadata(data []byte) map[string]interface{} {
	metadata := map[string]interface{}{"format": "JPEG"}

	// Look for EXIF data
	exifOffset := ia.findJPEGSegment(data, 0xE1) // APP1 segment
	if exifOffset != -1 && exifOffset+10 < len(data) {
		// Check for EXIF identifier
		if string(data[exifOffset+4:exifOffset+8]) == "Exif" {
			metadata["has_exif"] = true
		}
	}

	// Look for comment
	commentOffset := ia.findJPEGSegment(data, 0xFE)
	if commentOffset != -1 {
		metadata["has_comment"] = true
	}

	// Check for progressive encoding
	metadata["progressive"] = ia.isProgressiveJPEG(data)

	return metadata
}

func (ia *ImageAnalyzer) findJPEGSegment(data []byte, marker byte) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == marker {
			return i
		}
	}
	return -1
}

func (ia *ImageAnalyzer) isProgressiveJPEG(data []byte) bool {
	// Look for SOF2 marker (0xFFC2) which indicates progressive JPEG
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xC2 {
			return true
		}
	}
	return false
}

func (ia *ImageAnalyzer) extractPNGMetadata(data []byte) map[string]interface{} {
	metadata := map[string]interface{}{"format": "PNG"}

	if len(data) < 33 {
		return metadata
	}

	// PNG IHDR chunk starts at byte 8
	ihdr := data[8:33]
	if string(ihdr[:4]) == "IHDR" {
		metadata["has_ihdr"] = true
	}

	// Look for text chunks
	textChunks := []string{"tEXt", "zTXt", "iTXt"}
	for _, chunk := range textChunks {
		if strings.Contains(string(data), chunk) {
			metadata["has_text"] = true
			break
		}
	}

	// Check for transparency
	if strings.Contains(string(data), "tRNS") {
		metadata["has_transparency"] = true
	}

	// Check for color profile
	if strings.Contains(string(data), "iCCP") {
		metadata["has_color_profile"] = true
	}

	return metadata
}

func (ia *ImageAnalyzer) extractGIFMetadata(data []byte) map[string]interface{} {
	metadata := map[string]interface{}{"format": "GIF"}

	if len(data) < 6 {
		return metadata
	}

	// Check GIF version
	if len(data) >= 6 {
		version := string(data[:6])
		metadata["version"] = version
		metadata["animated"] = version == "GIF89a" // GIF89a supports animation
	}

	// Look for application extension (for animation)
	if strings.Contains(string(data), "NETSCAPE") {
		metadata["has_netscape_extension"] = true
		metadata["likely_animated"] = true
	}

	return metadata
}

func (ia *ImageAnalyzer) extractBMPMetadata(data []byte) map[string]interface{} {
	metadata := map[string]interface{}{"format": "BMP"}

	if len(data) < 54 {
		return metadata
	}

	// BMP header information is in the first 54 bytes
	// File size is at offset 2-5
	if len(data) >= 6 {
		fileSize := uint32(data[2]) | uint32(data[3])<<8 | uint32(data[4])<<16 | uint32(data[5])<<24
		metadata["header_file_size"] = fileSize
	}

	// DIB header size at offset 14-17
	if len(data) >= 18 {
		dibHeaderSize := uint32(data[14]) | uint32(data[15])<<8 | uint32(data[16])<<16 | uint32(data[17])<<24
		metadata["dib_header_size"] = dibHeaderSize
	}

	return metadata
}

func (ia *ImageAnalyzer) extractWebPMetadata(data []byte) map[string]interface{} {
	metadata := map[string]interface{}{"format": "WebP"}

	if len(data) < 12 {
		return metadata
	}

	// Check WebP format variant
	if len(data) >= 16 {
		fourcc := string(data[12:16])
		metadata["webp_format"] = fourcc

		switch fourcc {
		case "VP8 ":
			metadata["webp_type"] = "lossy"
		case "VP8L":
			metadata["webp_type"] = "lossless"
		case "VP8X":
			metadata["webp_type"] = "extended"
			// VP8X supports animation and other features
			if len(data) >= 20 {
				flags := data[20]
				metadata["has_animation"] = (flags & 0x02) != 0
				metadata["has_exif"] = (flags & 0x08) != 0
				metadata["has_xmp"] = (flags & 0x04) != 0
			}
		}
	}

	return metadata
}

func (ia *ImageAnalyzer) extractTIFFMetadata(data []byte) map[string]interface{} {
	metadata := map[string]interface{}{"format": "TIFF"}

	if len(data) < 8 {
		return metadata
	}

	// Check byte order
	if data[0] == 0x49 && data[1] == 0x49 {
		metadata["byte_order"] = "little_endian"
	} else if data[0] == 0x4D && data[1] == 0x4D {
		metadata["byte_order"] = "big_endian"
	}

	// TIFF magic number at offset 2-3
	if len(data) >= 4 {
		if (data[0] == 0x49 && data[2] == 0x2A && data[3] == 0x00) ||
		   (data[0] == 0x4D && data[2] == 0x00 && data[3] == 0x2A) {
			metadata["valid_tiff_magic"] = true
		}
	}

	return metadata
}

func (ia *ImageAnalyzer) extractDimensions(data []byte, format string) map[string]interface{} {
	dimensions := make(map[string]interface{})

	switch format {
	case "PNG":
		if len(data) >= 24 {
			// PNG dimensions are in IHDR chunk (bytes 16-23)
			width := uint32(data[16])<<24 | uint32(data[17])<<16 | uint32(data[18])<<8 | uint32(data[19])
			height := uint32(data[20])<<24 | uint32(data[21])<<16 | uint32(data[22])<<8 | uint32(data[23])
			dimensions["width"] = width
			dimensions["height"] = height
		}
	case "GIF":
		if len(data) >= 10 {
			// GIF dimensions at bytes 6-9 (little endian)
			width := uint16(data[6]) | uint16(data[7])<<8
			height := uint16(data[8]) | uint16(data[9])<<8
			dimensions["width"] = width
			dimensions["height"] = height
		}
	case "BMP":
		if len(data) >= 26 {
			// BMP dimensions at bytes 18-25 (little endian)
			width := uint32(data[18]) | uint32(data[19])<<8 | uint32(data[20])<<16 | uint32(data[21])<<24
			height := uint32(data[22]) | uint32(data[23])<<8 | uint32(data[24])<<16 | uint32(data[25])<<24
			dimensions["width"] = width
			dimensions["height"] = height
		}
	default:
		dimensions["extraction_supported"] = false
	}

	return dimensions
}

func (ia *ImageAnalyzer) analyzeColorInfo(data []byte, format string) map[string]interface{} {
	colorInfo := make(map[string]interface{})

	switch format {
	case "PNG":
		if len(data) >= 25 {
			// PNG color type is at byte 25 in IHDR
			colorType := data[25]
			colorInfo["color_type"] = colorType

			switch colorType {
			case 0:
				colorInfo["color_description"] = "grayscale"
			case 2:
				colorInfo["color_description"] = "rgb"
			case 3:
				colorInfo["color_description"] = "palette"
			case 4:
				colorInfo["color_description"] = "grayscale_alpha"
			case 6:
				colorInfo["color_description"] = "rgb_alpha"
			}

			// Bit depth at byte 24
			bitDepth := data[24]
			colorInfo["bit_depth"] = bitDepth
		}
	case "GIF":
		if len(data) >= 11 {
			// GIF packed field at byte 10
			packed := data[10]
			colorInfo["has_global_color_table"] = (packed & 0x80) != 0
			colorInfo["color_resolution"] = ((packed & 0x70) >> 4) + 1
			colorInfo["global_color_table_size"] = 2 << (packed & 0x07)
		}
	case "BMP":
		if len(data) >= 30 {
			// BMP bit count at bytes 28-29
			bitCount := uint16(data[28]) | uint16(data[29])<<8
			colorInfo["bits_per_pixel"] = bitCount
		}
	}

	return colorInfo
}

func (ia *ImageAnalyzer) assessImageQuality(data []byte, format string) map[string]interface{} {
	quality := make(map[string]interface{})

	// File size based assessment
	fileSize := len(data)
	quality["file_size"] = fileSize

	if fileSize < 1024 {
		quality["size_category"] = "very_small"
	} else if fileSize < 10240 {
		quality["size_category"] = "small"
	} else if fileSize < 102400 {
		quality["size_category"] = "medium"
	} else if fileSize < 1048576 {
		quality["size_category"] = "large"
	} else {
		quality["size_category"] = "very_large"
	}

	// Format-specific quality indicators
	switch format {
	case "JPEG":
		quality["compression_artifacts_likely"] = fileSize < 50000 // Rough heuristic
	case "PNG":
		quality["lossless"] = true
	case "GIF":
		quality["limited_colors"] = true
		quality["lossless"] = true
	}

	return quality
}

func (ia *ImageAnalyzer) analyzeImageStructure(data []byte, format string) map[string]interface{} {
	structure := make(map[string]interface{})

	// Header validation
	structure["has_valid_header"] = ia.hasValidHeader(data, format)

	// Corruption indicators
	structure["truncated"] = ia.appearsIncomplete(data, format)

	// Embedded data
	structure["has_embedded_thumbnail"] = ia.hasEmbeddedThumbnail(data, format)

	// Multiple images (for formats that support it)
	if format == "TIFF" || format == "GIF" {
		structure["potentially_multi_image"] = ia.hasPotentialMultipleImages(data, format)
	}

	return structure
}

func (ia *ImageAnalyzer) hasValidHeader(data []byte, format string) bool {
	switch format {
	case "PNG":
		return len(data) >= 8 && data[0] == 0x89 && string(data[1:4]) == "PNG"
	case "JPEG":
		return len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
	case "GIF":
		return len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a")
	case "BMP":
		return len(data) >= 2 && string(data[:2]) == "BM"
	case "WebP":
		return len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP"
	default:
		return false
	}
}

func (ia *ImageAnalyzer) appearsIncomplete(data []byte, format string) bool {
	switch format {
	case "JPEG":
		// JPEG should end with FFD9
		return len(data) < 2 || data[len(data)-2] != 0xFF || data[len(data)-1] != 0xD9
	case "PNG":
		// PNG should end with IEND chunk
		return len(data) < 12 || string(data[len(data)-8:len(data)-4]) != "IEND"
	case "GIF":
		// GIF should end with 0x3B
		return len(data) < 1 || data[len(data)-1] != 0x3B
	default:
		return false
	}
}

func (ia *ImageAnalyzer) hasEmbeddedThumbnail(data []byte, format string) bool {
	switch format {
	case "JPEG":
		// Look for thumbnail in EXIF data
		return strings.Contains(string(data), "thumbnail") ||
		       ia.findJPEGSegment(data, 0xE1) != -1 // APP1 segment often contains thumbnails
	default:
		return false
	}
}

func (ia *ImageAnalyzer) hasPotentialMultipleImages(data []byte, format string) bool {
	switch format {
	case "GIF":
		// Count image descriptors (0x2C)
		count := 0
		for i := 0; i < len(data); i++ {
			if data[i] == 0x2C {
				count++
			}
		}
		return count > 1
	case "TIFF":
		// TIFF can have multiple IFDs, but this requires more complex parsing
		return strings.Count(string(data), "IFD") > 1 // Very rough heuristic
	default:
		return false
	}
}

func (ia *ImageAnalyzer) detectImagePatterns(data []byte, format string) map[string]interface{} {
	patterns := make(map[string]interface{})

	// Check for common metadata patterns
	patterns["has_copyright"] = ia.containsPattern(data, []string{"copyright", "©", "(c)"})
	patterns["has_camera_info"] = ia.containsPattern(data, []string{"canon", "nikon", "sony", "camera"})
	patterns["has_gps_data"] = ia.containsPattern(data, []string{"gps", "latitude", "longitude"})
	patterns["has_software_info"] = ia.containsPattern(data, []string{"photoshop", "gimp", "software"})

	// Format-specific patterns
	switch format {
	case "JPEG":
		patterns["has_exif"] = strings.Contains(string(data), "Exif")
		patterns["has_jfif"] = strings.Contains(string(data), "JFIF")
	case "PNG":
		patterns["has_text_chunks"] = ia.containsPattern(data, []string{"tEXt", "iTXt", "zTXt"})
	case "GIF":
		patterns["animated"] = strings.Contains(string(data), "NETSCAPE")
	}

	return patterns
}

func (ia *ImageAnalyzer) containsPattern(data []byte, patterns []string) bool {
	dataStr := strings.ToLower(string(data))
	for _, pattern := range patterns {
		if strings.Contains(dataStr, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func (ia *ImageAnalyzer) createResultMessage(result *ProcessingResult) *client.BrokerMessage {
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
	imageAnalyzer := NewImageAnalyzer()
	agent.Run(imageAnalyzer, "image-analyzer")
}