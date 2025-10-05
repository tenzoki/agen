# Config

Cell configuration directory - 30+ declarative YAML configurations for pipeline orchestration.

## Intent

Provides production-ready cell configurations for common processing pipelines (document processing, anonymization, text analysis, search indexing, RAG). Configurations are declarative, composable, and deployment-ready. AI can read, modify, and create new configurations for self-modification.

## Usage

Deploy cell configuration:
```bash
../../bin/orchestrator -config=./text-pipeline.yaml
```

Validate configuration:
```bash
../../bin/orchestrator -config=./my-pipeline.yaml -validate
```

List available configurations:
```bash
ls *.yaml
# 30+ configurations organized by use case
```

Create custom configuration (copy and modify):
```bash
cp document-processing-pipeline.yaml my-custom-pipeline.yaml
vim my-custom-pipeline.yaml
../../bin/orchestrator -config=./my-custom-pipeline.yaml
```

## Setup

No dependencies - configurations reference agent binaries in `../../bin/` and agent pool definitions.

Configuration structure:
```yaml
cell:
  id: "pipeline-name"
  agents:
    - id: "agent-instance-id"
      agent_type: "agent-type-from-pool"
      dependencies: ["other-agent-id"]
      ingress: "file:input/*.txt"
      egress: "pub:topic"
      config:
        custom_key: "value"
```

## Tests

Validate all configurations:
```bash
for file in *.yaml; do
    echo "Validating $file"
    ../../bin/orchestrator -config=$file -validate
done
```

Test specific configuration:
```bash
../../bin/orchestrator -config=./anonymization_pipeline.yaml -dry-run
```

## Demo

Available configurations (organized by category):

**Document Processing:**
- `document-processing-pipeline.yaml` - General document processing
- `fast-document-processing-pipeline.yaml` - Optimized processing
- `intelligent-document-processing-pipeline.yaml` - AI-enhanced processing
- `academic-analysis-pipeline.yaml` - Academic document analysis
- `research-paper-processing-pipeline.yaml` - Research paper processing

**Text Processing:**
- `text-extraction-pipeline.yaml` - Text extraction
- `native-text-extraction.yaml` - Native extraction (no OCR)
- `file-transform-pipeline.yaml` - File transformation
- `file-chunking-pipeline.yaml` - Chunking pipeline

**Content Analysis:**
- `content-analysis-processing.yaml` - Content analysis
- `text-analysis-deep-processing.yaml` - Deep text analysis
- `fast-analysis-processing.yaml` - Fast analysis
- `structured-data-analysis-processing.yaml` - Structured data
- `binary-media-analysis-processing.yaml` - Binary/media analysis

**Synthesis & Output:**
- `full-analysis-synthesis.yaml` - Complete analysis synthesis
- `document-summary-synthesis.yaml` - Document summarization
- `data-export-synthesis.yaml` - Data export
- `reporting-synthesis.yaml` - Report generation
- `search-ready-synthesis.yaml` - Search indexing

**Specialized Pipelines:**
- `anonymization_pipeline.yaml` - Privacy/PII anonymization
- `http-ocr-extraction.yaml` - HTTP-based OCR
- `http-ocr-production-extraction.yaml` - Production OCR
- `all-in-one-ocr-container.yaml` - Containerized OCR
- `knowledge-backend-rag.yaml` - RAG implementation
- `storage-service.yaml` - Storage backend
- `storage-cell.yaml` - Storage cell

See [/reflect/cells/](../../reflect/cells/) for detailed cell documentation and [/reflect/architecture/cellorg.md](../../reflect/architecture/cellorg.md) for configuration syntax.
