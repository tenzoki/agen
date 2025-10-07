# Intelligent Document Processing Pipeline

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Complete intelligent document processing with strategy selection and context enrichment

## Intent

Provides advanced document processing through OCR-enabled text extraction, content-driven strategy selection, adaptive text chunking, and multi-dimensional context enrichment (positional, semantic, structural, relational) to generate knowledge-graph-ready document chunks.

## Agents

- **text-extractor-idp-001** (text-extractor) - Multilingual OCR text extraction
  - Ingress: sub:document-files
  - Egress: pub:extracted-text

- **strategy-selector-idp-001** (strategy-selector) - Content-aware strategy selection
  - Ingress: sub:extracted-text
  - Egress: pub:selected-strategies

- **text-chunker-idp-001** (text-chunker) - Adaptive text chunking
  - Ingress: sub:selected-strategies
  - Egress: pub:chunked-text

- **context-enricher-idp-001** (context-enricher) - Multi-dimensional context enrichment
  - Ingress: sub:chunked-text
  - Egress: pub:enriched-chunks

- **chunk-writer-idp-001** (chunk-writer) - Writes enriched chunk outputs
  - Ingress: sub:enriched-chunks
  - Egress: file:output/intelligent-processing/{{.request_id}}/{{.format}}/chunk_{{.index}}.{{.ext}}

## Data Flow

```
Documents → text-extractor (OCR, multilingual, 4 workers, 120s timeout)
  → strategy-selector (content analysis, document-aware default)
  → text-chunker (adaptive 2048-8192 bytes, 256 overlap)
  → context-enricher (positional + semantic + structural + relational, depth 3)
  → chunk-writer (JSON with metadata, custom naming)
```

## Configuration

Text Extraction:
- OCR enabled (eng, deu, fra)
- Timeout: 120 seconds
- Worker pool: 4

Strategy Selection:
- Default: document_aware
- Content analysis enabled

Text Chunking:
- Default size: 2048 bytes
- Max size: 8192 bytes
- Overlap: 256 bytes

Context Enrichment:
- Positional, semantic, structural, relational context enabled
- Context depth: 3

Output:
- Format: JSON with metadata preservation
- Custom naming scheme (chunk_XXXX)
- Auto-create directories by request_id and format

Orchestration:
- Startup timeout: 90s, shutdown: 30s
- Max retries: 3, retry delay: 10s
- Health check interval: 30s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/intelligent-document-processing-pipeline.yaml
```
