# File Chunking Pipeline

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Complete file chunking pipeline with splitting, processing, and synthesis

## Intent

Processes large files through semantic-aware splitting, parallel chunk analysis with keyword and sentiment extraction, and comprehensive synthesis to generate unified document summaries with charts and statistics.

## Agents

- **file-ingester-chunking-001** (file-ingester) - Monitors directory for files
  - Ingress: file:examples/chunking-pipeline/input/*.*
  - Egress: pub:source-files

- **file-splitter-001** (file-splitter) - Semantic file splitting
  - Ingress: sub:source-files
  - Egress: pub:file-chunks

- **chunk-processor-001** (chunk-processor) - Analyzes chunks with keywords and sentiment
  - Ingress: sub:file-chunks
  - Egress: pub:processed-chunks

- **chunk-synthesizer-001** (chunk-synthesizer) - Synthesizes document summaries
  - Ingress: sub:processed-chunks
  - Egress: pub:synthesis-results

- **file-writer-synthesis-001** (file-writer) - Writes JSON synthesis outputs
  - Ingress: sub:synthesis-results
  - Egress: file:examples/chunking-pipeline/output/synthesis_{{.document_hash}}_{{.timestamp}}.json

## Data Flow

```
Files → file-ingester → file-splitter (semantic, 10MB chunks, 512 byte overlap)
  → chunk-processor (keywords + sentiment, 4 workers, 5m timeout)
  → chunk-synthesizer (summary + keywords + charts)
  → file-writer (JSON output)
```

## Configuration

File Splitting:
- Strategy: semantic-aware chunking
- Chunk size: 10MB, overlap: 512 bytes
- Temp directory: /tmp/gox-chunks

Chunk Processing:
- Type: text_processor
- Keywords and sentiment enabled
- 4 workers, 5 minute job timeout

Synthesis:
- Type: document_summary
- Max keywords: 50
- Charts included, JSON output

Orchestration:
- Startup timeout: 60s, shutdown: 30s
- Max retries: 3, retry delay: 10s
- Health check interval: 20s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/file-chunking-pipeline.yaml
```
