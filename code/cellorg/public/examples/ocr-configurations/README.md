# OCR Configuration Examples

**UPDATED**: The original dual-mode `text-extractor` has been decomposed into specialized agents. This directory shows configurations for both legacy and new approaches.

## Current Architecture (Recommended)

### Native OCR Agent
- **Agent**: `text-extractor-native`
- **Cell**: `extraction:native-text`
- **Usage**: No external dependencies, direct Tesseract integration

### HTTP OCR Agent
- **Agent**: `ocr-http-stub` (await pattern)
- **Service**: Containerized Python OCR service
- **Cells**: `extraction:http-ocr`, `extraction:http-ocr-production`

## Migration from Legacy

### Old Dual-Mode Agent (DEPRECATED)
```yaml
# Old approach - DEPRECATED
- agent_type: "text-extractor"
  config:
    ocr_mode: "native"  # or "http"
```

### New Specialized Agents
```yaml
# Native OCR
- agent_type: "text-extractor-native"
  operator: "spawn"
  config:
    enable_ocr: true
    ocr_languages: "eng,deu,fra"

# HTTP OCR
- agent_type: "ocr-http-stub"
  operator: "await"  # Waits for service dependency
  config:
    service_url: "http://localhost:8080/ocr"
```

## Quick Start Examples

### 1. Native OCR Cell
```bash
# Start GOX with native OCR cell
./build/gox config/cells.yaml
# Use cell: extraction:native-text
```

### 2. HTTP OCR with Service
```bash
# Step 1: Start containerized OCR service
cd agents/ocr_http_stub
../../scripts/build-ocr-http.sh

# Step 2: Start GOX with HTTP OCR cell
./build/gox config/cells.yaml
# Use cell: extraction:http-ocr
```

### 3. Production HTTP OCR
```bash
# Start production setup (load-balanced)
cd agents/ocr_http_stub
../../scripts/build-ocr-http.sh --production

# Use cell: extraction:http-ocr-production
```

## Configuration Files

- **`native-ocr.json`** - Legacy config for native Tesseract (deprecated)
- **`http-ocr.json`** - Legacy config for HTTP service (deprecated)

## Cell Examples

### Native Text Extraction
```yaml
cell:
  id: "extraction:native-text"
  agents:
    - id: "text-extractor-native-001"
      agent_type: "text-extractor-native"
      config:
        enable_ocr: true
        ocr_languages: "eng,deu,fra"
        enable_preprocessing: true
```

### HTTP OCR with Await Pattern
```yaml
cell:
  id: "extraction:http-ocr"
  agents:
    - id: "ocr-http-stub-001"
      agent_type: "ocr-http-stub"
      operator: "await"  # Service dependency
      config:
        service_url: "http://localhost:8080/ocr"
        max_retries: 3
```

## Benefits of New Architecture

✅ **Single Responsibility**: Each agent has one clear purpose
✅ **Dependency Management**: Await pattern ensures service availability
✅ **Deployment Flexibility**: Native vs. containerized based on needs
✅ **Framework Compliance**: Both agents follow GOX patterns
✅ **Production Ready**: Load balancing and health monitoring for HTTP mode