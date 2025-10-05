# Academic Document Analysis Pipeline

Specialized pipeline for academic document analysis and research paper processing

## Intent

Processes academic papers and research documents through OCR-enabled text extraction, structured document processing with academic paper strategies, chunk-level analysis, and comprehensive summarization to produce structured analysis outputs suitable for research knowledge bases.

## Agents

- **file-ingester-academic-001** (file-ingester) - Monitors directory for academic PDFs
  - Ingress: file:examples/academic-pipeline/papers/*.pdf
  - Egress: pub:research-papers

- **text-extractor-academic-001** (text-extractor) - Extracts text with OCR support
  - Ingress: sub:research-papers
  - Egress: pub:paper-text

- **document-processor-academic-001** (document-processor) - Chunks text using academic paper strategy
  - Ingress: sub:paper-text
  - Egress: pub:paper-sections

- **chunk-processor-academic-001** (chunk-processor) - Analyzes chunks with keyword extraction
  - Ingress: sub:paper-sections
  - Egress: pub:analyzed-sections

- **chunk-synthesizer-academic-001** (chunk-synthesizer) - Generates document summaries with sections
  - Ingress: sub:analyzed-sections
  - Egress: pub:paper-summary

- **file-writer-academic-001** (file-writer) - Writes JSON analysis outputs
  - Ingress: sub:paper-summary
  - Egress: file:examples/academic-pipeline/summaries/{{.paper_title}}_analysis_{{.timestamp}}.json

## Data Flow

```
PDF Files → file-ingester → text-extractor (OCR) → document-processor (academic strategy)
  → chunk-processor (keyword extraction) → chunk-synthesizer (summary generation)
  → file-writer (JSON output)
```

## Configuration

- OCR enabled with 85% quality threshold, 120s timeout
- Academic paper chunking strategy with 1024 byte chunks, 128 byte overlap
- Preserve document structure and metadata throughout pipeline
- Keyword extraction enabled (max 25 keywords)
- Summary length: 500 characters with section inclusion
- Startup timeout: 60s, max retries: 2

## Usage

```bash
./bin/orchestrator -config=./workbench/config/academic-analysis-pipeline.yaml
```
