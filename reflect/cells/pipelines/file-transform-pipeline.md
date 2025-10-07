# File Transform Pipeline

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


File processing pipeline with ingestion, transformation, and output

## Intent

Demonstrates basic file processing workflow with file monitoring, text transformation (uppercase with metadata), and templated output generation to illustrate fundamental pipeline patterns and agent dependencies.

## Agents

- **file-ingester-demo-001** (file-ingester) - Monitors directory for text files
  - Ingress: file:examples/pipeline-demo/input/*.txt
  - Egress: pub:new-file

- **text-transformer-demo-001** (text-transformer) - Transforms text to uppercase
  - Ingress: sub:new-file
  - Egress: pipe:transform-data

- **data-adapter-demo-001** (adapter) - Adapter service for format conversion
  - Ingress: sub:transform-requests
  - Egress: pub:transform-responses

- **file-writer-demo-001** (file-writer) - Writes transformed output
  - Ingress: pipe:transform-data
  - Egress: file:examples/pipeline-demo/output/processed_{{.timestamp}}.txt

## Data Flow

```
Text Files → file-ingester (digest + delete, 2s watch)
  → text-transformer (uppercase + metadata)
  → file-writer (timestamped output)

Transform Requests → data-adapter (JSON/CSV/text, max 1MB)
  → Transform Responses
```

## Configuration

File Ingestion:
- Digest enabled with delete strategy
- Watch interval: 2 seconds

Text Transformation:
- Transformation: uppercase
- Metadata addition enabled

Data Adapter:
- Formats: JSON, CSV, text
- Max request size: 1MB

Output:
- Format: TXT
- Auto-create directories
- Timestamped filenames

Orchestration:
- Startup timeout: 30s, shutdown: 15s
- Max retries: 3, retry delay: 5s
- Health check interval: 10s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/file-transform-pipeline.yaml
```
