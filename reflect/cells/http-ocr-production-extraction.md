# HTTP OCR Production Extraction

Production text extraction using load-balanced containerized OCR service

## Intent

Provides production-grade OCR processing through load-balanced HTTP service with extended timeouts, increased retry logic, and disabled debug mode to support large-scale document processing with high availability requirements.

## Agents

- **ocr-http-stub-prod-001** (ocr-http-stub) - Production OCR HTTP client
  - Ingress: sub:production-ocr-extraction
  - Egress: pub:extracted-text-production

## Data Flow

```
Documents (images/PDFs) → ocr-http-stub-prod
  → HTTP POST to Load-Balanced OCR Service (localhost:8000/ocr)
  → Service Processing (await pattern, 10 min timeout)
  → Extracted Text
```

## Configuration

OCR Service:
- Service URL: http://localhost:8000/ocr (load-balanced)
- Health check: http://localhost:8000/health
- Request timeout: 10 minutes (600s) for large documents
- Max retries: 5, retry delay: 10 seconds
- Supported formats: PNG, JPG, JPEG, TIFF, BMP, PDF

Orchestration:
- Startup timeout: 90s
- Shutdown timeout: 30s
- Max retries: 5, retry delay: 20s
- Health check interval: 60s
- Debug disabled (production mode)

## Usage

```bash
./bin/orchestrator -config=./workbench/config/http-ocr-production-extraction.yaml
```

Requires external load-balanced OCR service on localhost:8000.
