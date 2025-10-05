# Text Extraction Examples

This directory contains examples and demos for the decomposed text extraction agents.

## New Architecture

The original dual-mode `text-extractor` has been split into two specialized agents:

### 1. Native Text Extractor (`text-extractor-native`)
- Direct Tesseract integration
- No external dependencies
- Suitable for development and simple deployments

### 2. HTTP OCR Stub (`ocr-http-stub`)
- Connects to containerized OCR service
- Uses await pattern for service dependencies
- Production-ready with load balancing

## Demo Files

### Input Samples
```
input/
├── document.pdf           # PDF for text extraction
├── spreadsheet.xlsx       # Excel document
├── text_document.docx     # Word document
├── image_with_text.png    # Image requiring OCR
└── mixed_content.pdf      # PDF with text and images
```

### Expected Outputs
```
output/
├── native/               # Results from native extraction
└── http/                 # Results from HTTP OCR service
```

## Running Examples

### 1. Native Text Extraction
```bash
# Build native agent
go build -o build/text_extractor_native ./agents/text_extractor_native/

# Start GOX with native extraction cell
./build/gox config/cells.yaml

# The cell "extraction:native-text" will process documents
# Input:  sub:native-text-extraction
# Output: pub:extracted-text-native
```

### 2. HTTP OCR Service
```bash
# Step 1: Start OCR service
cd agents/ocr_http_stub
../../scripts/build-ocr-http.sh

# Step 2: Build HTTP stub agent
cd ../../
go build -o build/ocr_http_stub ./agents/ocr_http_stub/

# Step 3: Start GOX with HTTP extraction cell
./build/gox config/cells.yaml

# The cell "extraction:http-ocr" will process documents
# Input:  sub:http-ocr-extraction
# Output: pub:extracted-text-http
```

### 3. Direct Agent Testing (Development)
```bash
# Test native agent directly
GOX_AGENT_ID=text-extractor-native-test ./build/text_extractor_native

# Test HTTP stub (requires service running)
GOX_AGENT_ID=ocr-http-stub-test ./build/ocr_http_stub
```

## Message Format Examples

### Text Extraction Request
```json
{
  "request_id": "extract_001",
  "file_path": "examples/text-extraction/input/document.pdf",
  "options": {
    "enable_ocr": true,
    "ocr_languages": ["eng", "deu"],
    "include_metadata": true,
    "quality_threshold": 0.8
  }
}
```

### OCR Request (HTTP Stub)
```json
{
  "request_id": "ocr_001",
  "file_path": "examples/text-extraction/input/image_with_text.png",
  "options": {
    "languages": ["eng"],
    "psm": 3,
    "oem": 3,
    "preprocess": true
  }
}
```

### Response Format
```json
{
  "request_id": "extract_001",
  "success": true,
  "extracted_text": "Document content here...",
  "processing_time": "2.5s",
  "extractor_used": "text_extractor_native",
  "quality": 0.95,
  "metadata": {
    "format": "PDF",
    "pages": 3,
    "language": "en",
    "confidence": 94.2
  }
}
```

## Performance Comparison

| Aspect | Native Agent | HTTP OCR Stub |
|--------|-------------|---------------|
| **Setup** | Simple (binary only) | Service + agent |
| **Dependencies** | Local Tesseract | Docker service |
| **Latency** | Lower | Network overhead |
| **Scalability** | Process-bound | Service scalable |
| **Production** | Limited | Load-balanced |
| **Languages** | System Tesseract | Container (13+ langs) |

## Cell Integration

### Native Extraction Pipeline
```yaml
# Document processing with native OCR
cell:
  id: "pipeline:native-document-processing"
  agents:
    - id: "text-extractor-native-001"
      agent_type: "text-extractor-native"
      ingress: "sub:documents"
      egress: "pub:extracted-text"

    - id: "text-analyzer-001"
      agent_type: "text-analyzer"
      ingress: "sub:extracted-text"
      egress: "pub:analyzed-text"
```

### HTTP OCR Pipeline
```yaml
# Document processing with containerized OCR
cell:
  id: "pipeline:http-document-processing"
  agents:
    - id: "ocr-http-stub-001"
      agent_type: "ocr-http-stub"
      operator: "await"  # Service dependency
      ingress: "sub:documents"
      egress: "pub:extracted-text"

    - id: "text-analyzer-001"
      agent_type: "text-analyzer"
      ingress: "sub:extracted-text"
      egress: "pub:analyzed-text"
```

## Testing

### Unit Tests
```bash
# Test native agent
go test ./agents/text_extractor_native/ -v

# Test HTTP stub (requires mock service)
go test ./agents/ocr_http_stub/ -v
```

### Integration Tests
```bash
# Test complete extraction cells
./scripts/test.sh --cells extraction:native-text,extraction:http-ocr
```

## Migration Guide

### From Legacy Dual-Mode Agent
```bash
# Old usage (DEPRECATED)
./build/text_extractor --ocr-mode=native --input=document.pdf

# New approach - use cells
./build/gox config/cells.yaml  # extraction:native-text cell
```

### Configuration Migration
```yaml
# Old dual-mode config
- agent_type: "text-extractor"
  config:
    ocr_mode: "http"
    ocr_service_url: "http://localhost:8080/ocr"

# New specialized config
- agent_type: "ocr-http-stub"
  operator: "await"
  config:
    service_url: "http://localhost:8080/ocr"
```

This provides clear separation of concerns and proper service dependency management through the await pattern.