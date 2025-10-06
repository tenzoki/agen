# Native Text Extraction

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Text extraction using native Tesseract OCR (no external dependencies)

## Intent

Provides embedded OCR processing using native Tesseract library with image preprocessing (deskew, noise removal, contrast enhancement) to eliminate external service dependencies while supporting multilingual text extraction.

## Agents

- **text-extractor-native-001** (text-extractor-native) - Native Tesseract OCR
  - Ingress: sub:native-text-extraction
  - Egress: pub:extracted-text-native

## Data Flow

```
Documents (images/PDFs) → text-extractor-native
  → Preprocessing (deskew, noise removal, contrast enhancement)
  → Tesseract OCR (multilingual)
  → Extracted Text
```

## Configuration

OCR Processing:
- OCR enabled with native Tesseract
- Languages: eng, deu, fra (multilingual)
- Timeout: 60 seconds
- Worker pool: 4

Image Preprocessing:
- Deskew enabled
- Noise removal enabled
- Contrast enhancement enabled

Orchestration:
- Startup timeout: 30s, shutdown: 15s
- Max retries: 2, retry delay: 10s
- Health check interval: 30s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/native-text-extraction.yaml
```

Requires Tesseract OCR library installed locally.
