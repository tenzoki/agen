# Gox Integration Guide

**Status**: ✅ Integrated (Placeholder Implementation)
**Date**: 2025-10-03
**Version**: Alpha Integration

---

## Overview

Alfa has been integrated with Gox to support advanced multi-agent workflows using cells. This integration follows the cell-based architecture pattern from the Gox framework, where cells are functional units composed of networks of agents.

**Current Status**: The integration is complete with a **placeholder implementation**. The cell management API is fully functional, but actual agent deployment will occur when `github.com/tenzoki/gox/pkg/orchestrator` is published.

---

## Architecture

### Integration Pattern

```
┌─────────────────────────────────────────────┐
│  Alfa Process                               │
│                                             │
│  ┌───────────────────────────────────────┐ │
│  │  Orchestrator                         │ │
│  │  ├─ AI Layer (Claude/OpenAI)          │ │
│  │  ├─ Tool Dispatcher                   │ │
│  │  ├─ VFS/VCR/Context                   │ │
│  │  └─ Gox Manager                       │ │
│  └───────────┬───────────────────────────┘ │
│              │                               │
│  ┌───────────▼───────────────────────────┐ │
│  │  Gox Manager (internal/gox)           │ │
│  │  ├─ Cell Tracking                     │ │
│  │  ├─ Event Subscriptions               │ │
│  │  └─ Placeholder Orchestrator          │ │
│  └───────────────────────────────────────┘ │
│                                             │
│  Future: Will integrate pkg/orchestrator   │
│  when published by Gox team                │
└─────────────────────────────────────────────┘
```

### Key Components

1. **internal/gox/gox.go** - Gox manager wrapper
   - Cell lifecycle management
   - Event pub/sub
   - Placeholder for actual orchestrator

2. **internal/tools/tools.go** - Extended with cell actions
   - `start_cell` - Start a Gox cell
   - `stop_cell` - Stop a running cell
   - `list_cells` - List active cells
   - `query_cell` - Query a cell (synchronous)

3. **config/gox/** - Gox configuration
   - `gox.yaml` - Main configuration
   - `pool.yaml` - Agent types pool
   - `cells.yaml` - Cell definitions

---

## Usage

### Command Line

Enable Gox features with the `--enable-gox` flag:

```bash
# Start Alfa with Gox enabled
./alfa --enable-gox --project myproject

# With custom Gox config path
./alfa --enable-gox --gox-config /path/to/gox/config --project myproject
```

### AI Actions

Once enabled, the AI can use cell management actions:

#### Start a Cell

```json
{
  "action": "start_cell",
  "cell_id": "rag:knowledge-backend",
  "project_id": "my-project",
  "environment": {
    "OPENAI_API_KEY": "sk-..."
  }
}
```

#### Stop a Cell

```json
{
  "action": "stop_cell",
  "cell_id": "rag:knowledge-backend",
  "project_id": "my-project"
}
```

#### List Running Cells

```json
{
  "action": "list_cells"
}
```

#### Query a Cell

```json
{
  "action": "query_cell",
  "project_id": "my-project",
  "query": "find authentication code",
  "params": {
    "top_k": 5
  },
  "timeout": 10
}
```

---

## Configuration

### gox.yaml

```yaml
app_name: "alfa-gox"
debug: true
support:
  port: ":9000"
broker:
  port: ":9001"
  protocol: "tcp"
  codec: "json"
basedir:
  - "config/gox"
pool:
  - "pool.yaml"
cells:
  - "cells.yaml"
```

### pool.yaml

Define available agent types:

```yaml
pool:
  agent_types:
    - agent_type: "embedding-agent"
      binary: "build/embedding_agent"
      operator: "spawn"
      capabilities: ["embeddings"]
    - agent_type: "vectorstore-agent"
      binary: "build/vectorstore_agent"
      operator: "spawn"
      capabilities: ["vector-search"]
```

### cells.yaml

Define cells (agent networks):

```yaml
cell:
  id: "rag:knowledge-backend"
  description: "RAG cell for code search"
  agents:
    - id: "embedding-001"
      agent_type: "embedding-agent"
      ingress: "sub:embeddings:requests"
      egress: "pub:embeddings:results"
    - id: "vectorstore-001"
      agent_type: "vectorstore-agent"
      ingress: "sub:vectorstore:requests"
      egress: "pub:vectorstore:results"
```

---

## Cell Patterns

### RAG (Retrieval Augmented Generation)

```
Cell: "rag:knowledge-backend"
Purpose: Semantic code search and retrieval

Workflow:
1. Start cell for project
2. Query with natural language
3. Receive relevant code context
4. Use context for AI operations
5. Stop cell when done
```

### Document Processing

```
Cell: "pipeline:document-processing"
Purpose: Multi-step document analysis

Workflow:
1. Ingest files from VFS
2. Extract and chunk text
3. Generate embeddings
4. Store in vector database
5. Enable semantic search
```

---

## Placeholder Implementation

**Current State**: The Gox integration is implemented as a placeholder until `github.com/tenzoki/gox/pkg/orchestrator` is published.

### What Works Now

✅ Cell management API
✅ Cell tracking and lifecycle
✅ Event subscription API
✅ Configuration loading
✅ Integration with tool dispatcher
✅ AI can request cell operations
✅ Tests passing
✅ Demo application working

### What's Placeholder

⏳ Actual agent deployment
⏳ Agent process spawning
⏳ Broker message routing
⏳ Cell-to-cell communication
⏳ Synchronous queries (PublishAndWait)

### Migration Path

When `pkg/orchestrator` is published:

1. Update `internal/gox/gox.go` to import actual package
2. Replace placeholder implementation with real orchestrator calls
3. No changes needed to:
   - Tool dispatcher actions
   - AI system prompt
   - Configuration files
   - User-facing API

The migration will be **transparent** to users and the AI.

---

## Testing

### Run Tests

```bash
# Run Gox integration tests
go test ./test/gox -v

# Run all tests
go test ./... -v
```

### Run Demo

```bash
# Run interactive demo
go run demo/gox_demo/main.go
```

Expected output demonstrates:
- Cell lifecycle management
- Event subscription
- Cell information queries
- Graceful shutdown

---

## Examples

### Example: RAG Cell Usage

```go
// In Alfa's code or via AI action
err := goxManager.StartCell(
    "rag:knowledge-backend",
    "my-project",
    "/path/to/project",
    map[string]string{
        "OPENAI_API_KEY": apiKey,
    },
)

// Query for code context
result, err := goxManager.PublishAndWait(
    "my-project:rag-queries",
    "my-project:rag-results",
    map[string]interface{}{
        "query": "authentication implementation",
        "top_k": 5,
    },
    10*time.Second,
)

// Use result.Data for context
```

### Example: AI Workflow

1. User: "Start RAG for this project"
2. AI executes: `start_cell` action
3. User: "Find authentication code"
4. AI executes: `query_cell` action
5. AI uses results to provide informed answer
6. AI executes: `stop_cell` when done

### Example: Named Entity Recognition

```go
// Extract entities from text
action := tools.Action{
    Type: "extract_entities",
    Params: map[string]interface{}{
        "text":       "Angela Merkel met with Emmanuel Macron in Berlin",
        "project_id": "my-project",
        "language":   "en", // optional
    },
}

result := dispatcher.Execute(ctx, action)
// Result contains: {"entities": [...], "count": 3}
```

### Example: Text Anonymization

```go
// Anonymize sensitive text
action := tools.Action{
    Type: "anonymize_text",
    Params: map[string]interface{}{
        "text":       "John Smith works at OpenAI in San Francisco",
        "project_id": "my-project",
    },
}

result := dispatcher.Execute(ctx, action)
// Result contains:
// {
//   "anonymized_text": "PERSON_123456 works at ORG_789012 in LOC_345678",
//   "mappings": {"John Smith": "PERSON_123456", ...},
//   "entity_count": 3
// }

// Later: Deanonymize using mappings
deanonAction := tools.Action{
    Type: "deanonymize_text",
    Params: map[string]interface{}{
        "anonymized_text": result.Output["anonymized_text"],
        "mappings":        result.Output["mappings"],
        "project_id":      "my-project",
    },
}

restored := dispatcher.Execute(ctx, deanonAction)
// Restored text: "John Smith works at OpenAI in San Francisco"
```

---

## Troubleshooting

### Gox Not Available

If you see warnings about Gox not being available:
- Ensure `--enable-gox` flag is set
- Check configuration path is correct
- Verify config files exist

### Placeholder Messages

Messages like "Placeholder: Cell will be started..." are expected in the current implementation. They indicate the API is working correctly and waiting for the actual orchestrator package.

### Configuration Errors

Common issues:
- Missing `gox.yaml` - create in `config/gox/`
- Invalid YAML syntax - validate with a YAML linter
- Port conflicts - ensure `:9000` and `:9001` are available

---

## Future Enhancements

### Phase 1: Core Integration (✅ Complete)
- Cell management API
- Tool dispatcher actions
- Configuration loading
- Tests and demos

### Phase 2: Orchestrator Integration (⏳ Pending)
- Replace placeholder with actual `pkg/orchestrator`
- Enable agent deployment
- Activate broker routing
- Enable synchronous queries

### Phase 3: Advanced Features (📋 Planned)
- Auto-start cells on project open
- Cell health monitoring
- Multi-cell coordination
- Custom cell templates
- RAG cell integration for code search

---

## References

- **Integration Docs**: `integration_docs/alfa-gox-cell-integration.md`
- **Gox Repository**: https://github.com/tenzoki/gox
- **Phase 3 Status**: `integration_docs/PHASE3-COMPLETE.md`
- **Quick Start**: `integration_docs/ALFA-QUICKSTART.md`

---

**Integrated**: 2025-10-03
**Status**: Alpha - Placeholder Implementation
**Next**: Awaiting `pkg/orchestrator` publication
