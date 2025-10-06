# Reporting Synthesis

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Comprehensive reporting pipeline for business intelligence

## Intent

Generates business intelligence reports with charts, tables, and recommendations by combining metadata collection and report generation to support executive dashboards, analytics platforms, and decision-making workflows.

## Agents

- **metadata-agent** (metadata-collector) - Collects metadata for reporting
  - Ingress: channels:synthesis_request
  - Egress: channels:metadata_for_reporting

- **report-agent** (report-generator) - Generates comprehensive reports
  - Ingress: channels:synthesis_request
  - Egress: channels:report_generated

## Data Flow

```
Synthesis Request → metadata-agent (chunk + file metadata, no schema)
                 → report-agent (charts + tables + recommendations)
                   → Report Generated (JSON format)
```

## Configuration

Metadata Collection:
- Chunk and file metadata included
- Schema generation disabled (optimized for reporting)

Report Generation:
- Charts and tables included
- Recommendations enabled (max 10)
- Report format: JSON

Orchestration:
- Startup timeout: 30s, shutdown: 20s
- Max retries: 2
- Health check interval: 15s
- Execution order: metadata-agent → report-agent
- Reporting mode: comprehensive

## Usage

```bash
./bin/orchestrator -config=./workbench/config/reporting-synthesis.yaml
```
