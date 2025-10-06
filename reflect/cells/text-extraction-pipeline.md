# Text Extraction Pipeline

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Document text extraction pipeline with multi-format support and OCR

## Intent

Extracts text from multi-format documents (PDF, DOCX, XLSX, TXT, images) through OCR-enabled processing with multilingual support, quality thresholds, and JSON output to enable downstream text analysis, indexing, and archival workflows.

## Agents

- **file-ingester-text-001** (file-ingester) - Monitors directory for documents
  - Ingress: file:examples/text-extraction/input/*.{pdf,docx,xlsx,txt,png,jpg}
  - Egress: pub:document-files

- **text-extractor-001** (text-extractor) - Multi-format OCR text extraction
  - Ingress: sub:document-files
  - Egress: pub:extracted-text

- **file-writer-text-001** (file-writer) - Writes JSON extraction outputs
  - Ingress: sub:extracted-text
  - Egress: file:examples/text-extraction/output/extracted_{{.filename}}_{{.timestamp}}.json

## Data Flow

```
Documents (PDF/DOCX/XLSX/TXT/PNG/JPG) → file-ingester
  → text-extractor (OCR: eng/deu/fra, quality 0.7, 60s timeout)
  → file-writer (JSON with metadata)
```

## Configuration

File Ingestion:
- Formats: PDF, DOCX, XLSX, TXT, PNG, JPG, JPEG
- Watch interval: 3 seconds
- No digest (continuous processing)

Text Extraction:
- OCR enabled (multilingual: eng, deu, fra)
- Quality threshold: 0.7
- Timeout: 60 seconds
- Output format: JSON

Output:
- Format: JSON with metadata preservation
- Timestamped filenames
- Auto-create directories

Orchestration:
- Startup timeout: 45s, shutdown: 20s
- Max retries: 3, retry delay: 5s
- Health check interval: 15s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/text-extraction-pipeline.yaml
```
