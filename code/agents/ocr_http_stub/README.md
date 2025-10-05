# OCR HTTP Stub

HTTP client agent connecting to containerized OCR service for optical character recognition.

## Intent

Provides OCR capabilities by connecting to a containerized HTTP OCR service (Python Flask + Tesseract). Uses await pattern to ensure external service availability before cell activation. Supports multiple languages, PDF processing, and image preprocessing.

## Usage

Input: `ocr_request` with image/PDF file paths
Output: `ocr_response` with extracted text and confidence metrics

Configuration:
- `service_url`: OCR service URL (default: "http://localhost:8080/ocr")
- `request_timeout`: API timeout (default: 5 minutes)
- `max_retries`: Retry attempts (default: 3)
- `health_check_url`: Service health endpoint
- `supported_formats`: File formats (.png, .jpg, .pdf, etc.)

## Setup

Dependencies:
- Dockerized OCR service (must be running)
- Docker and docker-compose for service deployment

Start OCR service:
```bash
./scripts/build-ocr-http.sh
```

Build agent:
```bash
go build -o bin/ocr_http_stub ./code/agents/ocr_http_stub
```

## Tests

Test file: `ocr_http_stub_test.go`

Tests cover HTTP communication, error handling, and response parsing.

## Demo

No demo available
