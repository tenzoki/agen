# Cellorg

Cell-based orchestration framework - zero-boilerplate infrastructure for distributed processing pipelines.

## Intent

Provides zero-boilerplate infrastructure for building distributed processing pipelines through composable cells. Cells are the primary abstraction - self-contained units combining agents, dependencies, and routing. Framework handles all infrastructure (connections, lifecycle, message routing). See [/reflect/architecture/cellorg.md](../../reflect/architecture/cellorg.md) for detailed architecture.

## Usage

Define cell (declarative YAML):
```yaml
# workbench/config/my-pipeline.yaml
cell:
  id: "text-pipeline"
  agents:
    - id: "ingester-001"
      agent_type: "file-ingester"
      ingress: "file:input/*.txt"
      egress: "pub:raw-data"

    - id: "transformer-001"
      agent_type: "text-transformer"
      dependencies: ["ingester-001"]
      ingress: "sub:raw-data"
      egress: "pipe:output"
```

Run orchestrator:
```bash
cd code/cellorg
go build -o ../../bin/orchestrator ./cmd/orchestrator
../../bin/orchestrator -config=../../workbench/config/cells.yaml
```

Implement custom agent (3 lines):
```go
type MyAgent struct { agent.DefaultAgentRunner }
func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) { ... }
func main() { agent.Run(&MyAgent{}, "my-agent") }
```

## Setup

Dependencies:
- atomic utilities (github.com/tenzoki/agen/atomic)
- BadgerDB (via storage)

Build orchestrator:
```bash
cd code/cellorg
go build -o ../../bin/orchestrator ./cmd/orchestrator
```

Configuration files (in workbench/config/):
- `pool.yaml` - Agent type definitions
- `cells.yaml` - Agent instance deployment
- `cellorg.yaml` - Infrastructure settings

## Tests

Run all tests:
```bash
go test ./... -v
```

Integration tests:
```bash
go test ./public/examples/... -v
```

End-to-end cell tests:
```bash
go test ./internal/orchestrator/... -v
```

## Demo

Cell demos in `public/examples/`:
- `file-transform-pipeline/` - File processing cell
- `search-indexing/` - Content indexing
- `adapter-integration/` - Protocol adaptation
- `chunk-processing/` - Chunking pipeline
- `reporting-pipeline/` - Report generation

Run demo:
```bash
../../bin/orchestrator -config=./public/examples/file-transform-pipeline/cell.yaml
```

Multi-agent workbench demos:
- `/workbench/demos/cell_demo` - Full pipeline
- `/workbench/demos/anonymization_demo` - Anonymization

See [/reflect/architecture/cellorg.md](../../reflect/architecture/cellorg.md) for cell execution flow, communication patterns, and public APIs.
