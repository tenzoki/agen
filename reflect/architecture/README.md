# AGEN Architecture

Multi-purpose AI workbench with cell-based processing, unified storage, and self-modifying capabilities.

## Intent

AGEN is designed to enable AI systems to build, modify, and optimize complex workflows through a composable cell-based pattern. Core principles: **cells-first design**, **unified storage**, **agent modularity**, and **self-modification**.

## Core Design Principles

### 1. Cells-First Architecture

**Cells** are the primary abstraction - self-contained processing units that compose into workflows.

```
Cell = Agents + Dependencies + Routing + Configuration
```

**Key Properties:**
- **Declarative** - YAML defines behavior, not code
- **Composable** - Cells combine to form complex pipelines
- **Isolated** - Each cell operates independently
- **Reusable** - Same cell definition, multiple deployments

**Design Pattern:**
```yaml
cell:
  id: "processing-pipeline"
  agents:
    - id: "source-001"
      agent_type: "file-ingester"
      dependencies: []
      ingress: "file:input/*.txt"
      egress: "pub:raw-data"

    - id: "processor-001"
      agent_type: "text-transformer"
      dependencies: ["source-001"]
      ingress: "sub:raw-data"
      egress: "pipe:processed"
```

**Why Cells-First:**
- AI can reason about and modify high-level workflow structure
- Dependency resolution prevents invalid pipeline configurations
- Declarative nature enables automatic optimization
- Testing and debugging at cell level, not implementation level

### 2. Unified Storage (OmniStore)

**Single storage backend for all data types** - eliminates data silos and enables cross-domain operations.

```
OmniStore = KV + Graph + FileStore + Transactions + Query
```

**Storage Modes:**
- **Key-Value** - Configuration, state, fast lookups
- **Graph** - Relationships, dependencies, knowledge graphs
- **FileStore** - Content-addressable blob storage with deduplication
- **Transactions** - ACID operations across all stores

**Unification Benefits:**
- **Single API** - One interface for all persistence needs
- **Cross-domain queries** - Join KV data with graph relationships
- **Automatic deduplication** - Content-addressed file storage
- **Transaction consistency** - ACID across all data types

**Design Pattern:**
```go
// Single transaction across multiple stores
tx := store.Begin()
tx.KV().Set("config:agent-001", config)
tx.Graph().CreateEdge(source, target, "processes")
tx.FileStore().Store(content, metadata)
tx.Commit()
```

### 3. Agent Modularity

**Agents are stateless processors** - business logic only, infrastructure handled by framework.

```
Agent = ProcessMessage() + Framework
```

**Framework Responsibilities:**
- Connection management (Support, Broker)
- Lifecycle coordination (init, ready, run, shutdown)
- Message routing (ingress/egress)
- Error recovery and retries

**Agent Responsibilities:**
- Business logic only (process input → produce output)
- No infrastructure code required
- Declarative capabilities in pool.yaml

**Design Pattern:**
```go
type MyAgent struct { agent.DefaultAgentRunner }

func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    // Pure business logic - framework handles everything else
    result := processData(msg.Payload)
    return &client.BrokerMessage{Payload: result}, nil
}

func main() { agent.Run(&MyAgent{}, "my-agent") }
```

### 4. Self-Modification

**AGEN can modify its own codebase and workflows** - enabling evolutionary optimization.

**Modification Layers:**
1. **Cell Configuration** - Add/modify agents in cells.yaml (runtime)
2. **Agent Pool** - Add new agent types to pool.yaml (compilation)
3. **Source Code** - Generate/modify agent implementations (development)
4. **Architecture** - Evolve system design patterns (meta-development)

**Self-Modification Flow:**
```
Context Gathering → Knowledge Base → Analysis → Modification → Validation → Deployment
```

**Safety Mechanisms:**
- VFS isolation (workbench vs project scope)
- Git versioning (automatic commits + backups)
- Sandbox execution (Docker isolation optional)
- Validation pipeline (build + test before deploy)

## System Components

AGEN is organized into four core modules. See component-specific docs:

- **[Atomic](atomic.md)** - Foundational utilities (VFS, VCR)
- **[Cellorg](cellorg.md)** - Cell orchestration framework
- **[Agents](agents.md)** - Processing agent catalog
- **[Omni](omni.md)** - Unified storage backend

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                        ALFA                              │
│                   (AI Workbench)                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Speech Interface │ AI Layer │ Project Manager   │   │
│  └──────────────────────────────────────────────────┘   │
└────────────┬────────────────────────────┬───────────────┘
             │                            │
             ▼                            ▼
┌─────────────────────────┐    ┌──────────────────────────┐
│      CELLORG            │    │       OMNISTORE          │
│  (Cell Framework)       │◄───┤   (Unified Storage)      │
│                         │    │                          │
│  ┌─────────────────┐    │    │  ┌───────────────────┐  │
│  │  Orchestrator   │    │    │  │ KV │ Graph │ File │  │
│  │  Support/Broker │    │    │  │ Query │ Transaction│  │
│  │  Deployer       │    │    │  └───────────────────┘  │
│  └─────────────────┘    │    └──────────────────────────┘
└────────────┬────────────┘                 ▲
             │                              │
             ▼                              │
┌─────────────────────────────────────────┐│
│           AGENTS                         ││
│  ┌─────────────────────────────────┐    ││
│  │ File Processing │ Text Analysis │    ││
│  │ Content Indexing │ Search       │────┘│
│  │ NER │ OCR │ Transformers │ ...  │     │
│  └─────────────────────────────────┘     │
└───────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│           ATOMIC                         │
│  ┌─────────────────────────────────┐    │
│  │ VFS (Virtual Filesystem)        │    │
│  │ VCR (Version Control)           │    │
│  └─────────────────────────────────┘    │
└───────────────────────────────────────────┘
```

## Data Flow Patterns

### 1. Cell Execution Flow
```
Config Load → Dependency Resolution → Agent Deployment → Message Routing → Processing → Storage
```

### 2. Message Flow
```
Producer → Broker → Routing → Consumer(s) → Processing → Egress → Next Stage
```

### 3. Storage Flow
```
Agent → OmniStore → Transaction → Backend → Persistence
```

### 4. Self-Modification Flow
```
Context → Knowledge → Analysis → Code Generation → Validation → Deployment
```

## Module Dependencies

```
alfa → cellorg, atomic
cellorg → atomic, omni
agents → cellorg, omni
omni → atomic
atomic → (no dependencies)
```

**Dependency Rules:**
- Atomic has zero dependencies (foundational utilities)
- Omni depends only on atomic (storage layer)
- Cellorg orchestrates agents using atomic utilities
- Agents use cellorg framework and omni storage
- Alfa provides AI interface to entire system

## Usage

### Define Cell
```yaml
# workbench/config/my-pipeline.yaml
cell:
  id: "my-pipeline"
  agents:
    - id: "ingester"
      agent_type: "file-ingester"
      ingress: "file:input/*.json"
      egress: "pub:raw-data"
```

### Implement Agent
```go
// code/agents/my_agent/main.go
type MyAgent struct { agent.DefaultAgentRunner }
func (a *MyAgent) ProcessMessage(...) { /* logic */ }
func main() { agent.Run(&MyAgent{}, "my-agent") }
```

### Execute via Alfa
```bash
alfa --workdir=./workbench
> "Process all files in input/ through my-pipeline"
```

## Setup

**Build:**
```bash
# Build all modules
make -C builder all

# Build specific binary
go build -o bin/alfa ./code/alfa/cmd/alfa
go build -o bin/orchestrator ./code/cellorg/cmd/orchestrator
```

**Dependencies:**
- Go 1.24.3+
- BadgerDB (via omni)
- go-git (via atomic/vcr)
- Optional: Docker (for sandboxed execution)
- Optional: Tesseract, sox (for OCR, audio)

## Tests

Tests co-located with source code:
```bash
# Run all tests
go test ./...

# Test specific module
go test ./code/cellorg/...
go test ./code/agents/...
```

## Demo

Multi-module demos in `workbench/demos/`:
- `gox_demo/` - Cell orchestration
- `speech_ai/` - Voice + AI integration
- `anonymization/` - Privacy pipeline

See component-specific docs for detailed demos.
