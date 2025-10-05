# Data Export Synthesis

Structured data export pipeline for analytics and integration

## Intent

Generates structured datasets and metadata collections from processed documents to support analytics platforms, data integration workflows, and external system consumption through schema-validated JSON exports.

## Agents

- **metadata-agent** (metadata-collector) - Collects chunk and file metadata
  - Ingress: channels:synthesis_request
  - Egress: channels:metadata_collected

- **dataset-agent** (dataset-builder) - Builds validated datasets
  - Ingress: channels:synthesis_request
  - Egress: channels:dataset_exported

## Data Flow

```
Synthesis Request → metadata-agent (metadata + schema generation)
                 → dataset-agent (dataset build + validation)
                   → Exported Dataset (JSON, max 500K records)
```

## Configuration

Metadata Collection:
- Include chunk and file metadata
- Generate schema automatically
- Output format: JSON

Dataset Building:
- Output format: JSON with metadata and schema
- Validation enabled
- Naming scheme: export_XXXX
- Max records: 500,000 (large-scale export optimized)

Orchestration:
- Startup timeout: 30s, shutdown: 20s
- Max retries: 2, health check: 15s
- Execution order: metadata-agent → dataset-agent

## Usage

```bash
./bin/orchestrator -config=./workbench/config/data-export-synthesis.yaml
```
