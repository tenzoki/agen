# Fast Document Processing Pipeline

Simplified document processing pipeline optimized for speed

## Intent

Provides high-throughput document processing by disabling OCR, using direct text chunking without strategy selection, and minimizing metadata overhead to achieve minimal latency for native text documents.

## Agents

- **text-extractor-fast-001** (text-extractor) - Fast text extraction without OCR
  - Ingress: sub:fast-document-files
  - Egress: pub:fast-extracted-text

- **text-chunker-fast-001** (text-chunker) - Direct chunking without strategy
  - Ingress: sub:fast-extracted-text
  - Egress: pub:fast-chunked-text

- **chunk-writer-fast-001** (chunk-writer) - Basic text output
  - Ingress: sub:fast-chunked-text
  - Egress: file:output/fast-processing/{{.request_id}}/chunk_{{.index}}.txt

## Data Flow

```
Documents → text-extractor (no OCR, 30s timeout)
  → text-chunker (1024 byte chunks, 128 overlap)
  → chunk-writer (text format, no metadata)
```

## Configuration

Text Extraction:
- OCR disabled for speed
- Timeout: 30 seconds
- Worker pool: 2

Text Chunking:
- Default chunk size: 1024 bytes
- Max chunk size: 4096 bytes
- Overlap: 128 bytes

Output:
- Format: plain text
- Metadata preservation disabled
- Auto-create directories

Orchestration:
- Startup timeout: 45s, shutdown: 20s
- Max retries: 2, retry delay: 5s
- Health check interval: 15s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/fast-document-processing-pipeline.yaml
```
