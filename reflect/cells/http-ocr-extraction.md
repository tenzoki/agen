# HTTP OCR Extraction

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Text extraction using containerized HTTP OCR service with await pattern

## Intent

Delegates OCR processing to external containerized HTTP service to decouple compute-intensive OCR workloads from the pipeline, enabling horizontal scaling of OCR capacity and simplified deployment without local Tesseract dependencies.

## Agents

- **ocr-http-stub-001** (ocr-http-stub) - HTTP client for OCR service
  - Ingress: sub:http-ocr-extraction
  - Egress: pub:extracted-text-http

## Data Flow

```
Documents (images/PDFs) → ocr-http-stub
  → HTTP POST to OCR service (localhost:8080/ocr)
  → Service Processing (await pattern, 5 min timeout)
  → Extracted Text
```

## Configuration

OCR Service:
- Service URL: http://localhost:8080/ocr
- Health check: http://localhost:8080/health
- Request timeout: 5 minutes (300s)
- Max retries: 3, retry delay: 5 seconds
- Supported formats: PNG, JPG, JPEG, TIFF, BMP, PDF

Orchestration:
- Startup timeout: 60s (wait for service dependency)
- Shutdown timeout: 20s
- Max retries: 3, retry delay: 15s
- Health check interval: 45s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/http-ocr-extraction.yaml
```

Requires external OCR service running on localhost:8080.
