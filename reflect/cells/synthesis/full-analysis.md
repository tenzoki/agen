# Full Analysis Synthesis

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Comprehensive synthesis pipeline with all available analyzers

## Intent

Executes complete document analysis through parallel synthesis agents (summary, indexer, metadata, report, dataset) to generate comprehensive outputs for search systems, analytics platforms, business intelligence, and data integration workflows.

## Agents

- **summary-agent** (summary-generator) - Generates document summaries
  - Ingress: channels:synthesis_request
  - Egress: channels:summary_complete

- **indexer-agent** (search-indexer) - Builds search indices
  - Ingress: channels:synthesis_request
  - Egress: channels:index_complete

- **metadata-agent** (metadata-collector) - Collects comprehensive metadata
  - Ingress: channels:synthesis_request
  - Egress: channels:metadata_complete

- **report-agent** (report-generator) - Generates business reports
  - Ingress: channels:synthesis_request
  - Egress: channels:report_complete

- **dataset-agent** (dataset-builder) - Builds validated datasets
  - Ingress: channels:synthesis_request
  - Egress: channels:dataset_complete

## Data Flow

```
Synthesis Request → Parallel Processing:
  ├→ summary-agent (keywords, topics, 1000 char summary)
  ├→ indexer-agent (15K terms, positions, scoring)
  ├→ metadata-agent (chunk + file metadata, schema, max 20MB)
  ├→ report-agent (charts, tables, recommendations)
  └→ dataset-agent (validated dataset, 250K records)
    → Complete Analysis Results
```

## Configuration

Summary: 30 keywords, 15 topics, 1000 chars, sections enabled
Indexer: 15,000 terms, min freq 2, positions enabled, score threshold 0.05
Metadata: Chunk + file metadata, schema generation, max 20MB
Report: Charts + tables + recommendations (max 8)
Dataset: Metadata + schema + validation, max 250,000 records

Orchestration:
- Startup timeout: 45s, shutdown: 30s
- Max retries: 3, health check: 20s
- Execution order: summary → indexer → metadata → report → dataset
- Parallel processing enabled

## Usage

```bash
./bin/orchestrator -config=./workbench/config/full-analysis-synthesis.yaml
```
