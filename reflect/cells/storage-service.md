# Storage Service

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Storage service providing KV, graph, files, and full-text capabilities

## Intent

Provides centralized storage service with HTTP API supporting key-value operations, graph relationships, file storage, and full-text search to enable multiple cells to share persistent data without local storage dependencies.

## Agents

- **storage-service-001** (godast-storage) - HTTP storage service
  - Ingress: HTTP API (localhost:9002)
  - Egress: HTTP Responses

## Data Flow

```
HTTP Requests (localhost:9002) → storage-service-001
  ↕
Godast Storage Backend (/tmp/gox-storage-service)
  ├→ KV Store (key-value operations)
  ├→ Graph Store (relationship management)
  ├→ File Store (file operations, max 100MB)
  └→ Full-Text Search (indexing and queries)
```

## Configuration

Storage Backend:
- Data path: /tmp/gox-storage-service
- Max file size: 100MB
- KV, graph, files, full-text enabled

Service Mode:
- Listen port: 9002
- CORS enabled
- Request logging enabled
- No authentication required
- Max concurrent requests: 100
- Request timeout: 30s

Orchestration:
- Startup timeout: 60s, shutdown: 30s
- Max retries: 3
- Health check interval: 30s
- Metrics enabled

## Usage

```bash
./bin/orchestrator -config=./workbench/config/storage-service.yaml
```

Service endpoints available at http://localhost:9002/storage/v1/.
Health check: http://localhost:9002/health
