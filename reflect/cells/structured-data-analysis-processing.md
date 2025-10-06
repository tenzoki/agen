# Structured Data Analysis Processing

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Specialized analysis for structured data formats (JSON, XML)

## Intent

Provides comprehensive analysis of structured data through deep JSON validation (50 levels, 10K keys) and XML namespace analysis (50 levels, 10K elements) with schema generation to support data quality validation, schema discovery, and format conversion workflows.

## Agents

- **json-analyzer-comprehensive-001** (json-analyzer) - Comprehensive JSON analysis
  - Ingress: sub:structured-data-json
  - Egress: pub:comprehensive-analyzed-json

- **xml-analyzer-comprehensive-001** (xml-analyzer) - Comprehensive XML analysis
  - Ingress: sub:structured-data-xml
  - Egress: pub:comprehensive-analyzed-xml

## Data Flow

```
JSON Files → json-analyzer (validation + schema + key analysis, depth 50, 10K keys)
  → Analysis Results

XML Files → xml-analyzer (validation + schema + namespaces, depth 50, 10K elements)
  → Analysis Results
```

## Configuration

JSON Analysis:
- Validation, schema generation, key analysis enabled
- Max depth: 50 levels
- Max keys: 10,000
- Minification disabled (preserve structure)

XML Analysis:
- Validation, schema generation, namespace analysis enabled
- Max depth: 50 levels
- Max elements: 10,000
- Minification disabled (preserve structure)

Orchestration:
- Startup timeout: 30s, shutdown: 15s
- Max retries: 2, retry delay: 10s
- Health check interval: 30s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/structured-data-analysis-processing.yaml
```
