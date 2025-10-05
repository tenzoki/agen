# Document Processing Pipeline

Intelligent document processing with strategy-based chunking

## Intent

Processes multi-format documents (PDF, DOCX, TXT) through strategy-aware chunking that adapts to document structure, enabling context-preserving segmentation for downstream RAG, analysis, and storage workflows.

## Agents

- **file-ingester-docs-001** (file-ingester) - Monitors directory for documents
  - Ingress: file:examples/document-processing/input/*.{pdf,docx,txt}
  - Egress: pub:raw-documents

- **document-processor-001** (document-processor) - Strategy-based intelligent chunking
  - Ingress: sub:raw-documents
  - Egress: pub:processed-chunks

- **file-writer-chunks-001** (file-writer) - Writes chunk outputs
  - Ingress: sub:processed-chunks
  - Egress: file:examples/document-processing/output/{{.document_id}}/chunk_{{.chunk_index}}.txt

## Data Flow

```
Documents (PDF/DOCX/TXT) → file-ingester → document-processor (document-aware strategy)
  → Chunked Outputs (with metadata)
```

## Configuration

Document Processing:
- Strategy: document_aware (adapts to content structure)
- Chunk size: 2048 bytes, overlap: 256 bytes
- OCR enabled, structure preservation enabled
- Batch processing with 4 workers

File Management:
- Watch interval: 3 seconds, no digest
- Output format: TXT with metadata
- Auto-create output directories by document_id

Orchestration:
- Startup timeout: 45s, shutdown: 20s
- Max retries: 3, retry delay: 5s
- Health check interval: 15s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/document-processing-pipeline.yaml
```
