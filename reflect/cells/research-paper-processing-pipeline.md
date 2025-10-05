# Research Paper Processing Pipeline

Specialized pipeline for academic research paper processing

## Intent

Processes academic research papers through high-quality OCR extraction, academic-specific chunking strategies with large overlap, and deep context enrichment (depth 5) to preserve citation relationships, section boundaries, and semantic connections for research knowledge bases.

## Agents

- **text-extractor-research-001** (text-extractor) - High-quality OCR extraction
  - Ingress: sub:research-papers
  - Egress: pub:research-text

- **strategy-selector-research-001** (strategy-selector) - Academic strategy selection
  - Ingress: sub:research-text
  - Egress: pub:research-strategies

- **text-chunker-research-001** (text-chunker) - Section-aware chunking
  - Ingress: sub:research-strategies
  - Egress: pub:research-chunks

- **context-enricher-research-001** (context-enricher) - Deep context enrichment
  - Ingress: sub:research-chunks
  - Egress: pub:research-enriched

- **chunk-writer-research-001** (chunk-writer) - Multi-format output
  - Ingress: sub:research-enriched
  - Egress: file:output/research-papers/{{.request_id}}/{{.content_type}}/chunk_{{.index}}.json

## Data Flow

```
Research Papers → text-extractor (OCR, 6 workers, 180s timeout)
  → strategy-selector (academic_paper strategy, content analysis)
  → text-chunker (1024-3072 bytes, 512 overlap for citations)
  → context-enricher (positional + semantic + structural + relational, depth 5)
  → chunk-writer (JSON with metadata, organized by content_type)
```

## Configuration

Text Extraction:
- OCR enabled (English)
- Timeout: 180 seconds (for complex academic PDFs)
- Worker pool: 6

Strategy Selection:
- Default: academic_paper
- Content analysis enabled

Text Chunking:
- Default size: 1024 bytes
- Max size: 3072 bytes
- Overlap: 512 bytes (larger for academic content)

Context Enrichment:
- All context types enabled
- Context depth: 5 (deeper for academic relationships)

Output:
- Format: JSON with full metadata preservation
- Custom naming scheme
- Organized by request_id and content_type

Orchestration:
- Startup timeout: 120s, shutdown: 40s
- Max retries: 3, retry delay: 15s
- Health check interval: 45s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/research-paper-processing-pipeline.yaml
```
