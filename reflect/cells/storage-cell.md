# Storage Cell

File processing cell that uses external storage service

## Intent

Demonstrates integration pattern for external storage service through HTTP API, enabling agents to leverage centralized storage (indexing, relationships, backups) while maintaining clean separation between data processing and storage concerns.

## Agents

- **enhanced-ingester-001** (file-ingester) - File ingestion with storage logging
  - Ingress: file:examples/data/samples/storage-controller-demo/input/*.{txt,md,json}
  - Egress: pub:raw-files

- **enhanced-transformer-001** (text-transformer) - Text transformation with storage integration
  - Ingress: sub:raw-files
  - Egress: pub:processed-content

- **enhanced-writer-001** (file-writer) - File writing with storage metadata
  - Ingress: sub:processed-content
  - Egress: file:examples/data/samples/storage-controller-demo/output/{{.type}}_{{.timestamp}}.txt

## Data Flow

```
Files (txt/md/json) → enhanced-ingester (digest + archive + storage logging)
  → enhanced-transformer (enhance + timestamps + auto-index + store intermediate)
  → enhanced-writer (backup to storage + track relationships)
    ↕
  Storage Service HTTP API (localhost:9002)
```

## Configuration

File Ingester:
- Digest enabled with archive strategy
- Watch interval: 5 seconds
- Storage service URL: http://localhost:9002/storage/v1
- Storage logging enabled

Text Transformer:
- Transformation: enhance with timestamps
- Auto-index content enabled
- Store intermediate results
- Storage service integration

File Writer:
- Auto-create directories
- Backup to storage enabled
- Track file relationships

Storage Integration:
- Service URL: http://localhost:9002/storage/v1
- Health check: http://localhost:9002/health
- Health check interval: 30s, timeout: 5s

Monitoring:
- Metrics: files_processed_total, processing_duration, storage_api_calls_total
- Agent health checks every 15s
- External service health check every 30s

Orchestration:
- Startup timeout: 45s, shutdown: 20s
- Max retries: 3
- Health check interval: 10s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/storage-cell.yaml
```

Requires storage-service running on localhost:9002.
