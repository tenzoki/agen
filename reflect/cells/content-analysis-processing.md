# Content Analysis Processing

Comprehensive content analysis for text, JSON, XML, binary, and image data

## Intent

Provides parallel multi-format content analysis through specialized analyzers for text (sentiment, keywords), JSON (validation, schema), XML (namespace, validation), binary (hashing, entropy), and images (metadata, dimensions) to support heterogeneous document processing workflows.

## Agents

- **text-analyzer-001** (text-analyzer) - Text content analysis with sentiment and keywords
  - Ingress: sub:content-analysis-text
  - Egress: pub:analyzed-text

- **json-analyzer-001** (json-analyzer) - JSON validation and schema generation
  - Ingress: sub:content-analysis-json
  - Egress: pub:analyzed-json

- **xml-analyzer-001** (xml-analyzer) - XML validation and namespace analysis
  - Ingress: sub:content-analysis-xml
  - Egress: pub:analyzed-xml

- **binary-analyzer-001** (binary-analyzer) - Binary file analysis
  - Ingress: sub:content-analysis-binary
  - Egress: pub:analyzed-binary

- **image-analyzer-001** (image-analyzer) - Image metadata and quality analysis
  - Ingress: sub:content-analysis-image
  - Egress: pub:analyzed-image

## Data Flow

```
Input Files → Router (by type)
  ├→ text-analyzer (sentiment, keywords, max 10K lines)
  ├→ json-analyzer (validation, schema, max depth 20)
  ├→ xml-analyzer (validation, namespaces, max depth 20)
  ├→ binary-analyzer (hashing, entropy, magic bytes, max 10MB)
  └→ image-analyzer (metadata, dimensions, quality, max 10MB)
    → Analyzed Outputs
```

## Configuration

Text: NLP disabled, sentiment + keywords enabled, max 10,000 lines, 20 keywords
JSON: Validation + schema generation + key analysis, max depth 20, 1000 keys
XML: Validation + schema + namespace analysis, max depth 20, 1000 elements
Binary: Hashing + entropy + magic bytes + structural, max 10MB
Image: Metadata + dimensions + color + quality, max 10MB, thumbnail disabled

Orchestration: 30s startup, 15s shutdown, 2 retries, 30s health checks

## Usage

```bash
./bin/orchestrator -config=./workbench/config/content-analysis-processing.yaml
```
