# Agents

**Target Audience**: AI/LLM
**Purpose**: Agent catalog and implementation patterns

Agent concept, base components, and catalog of processing capabilities - business logic without infrastructure code.

**Quick Start**: See `guidelines/agent-development-guide.md` for practical examples

## Intent

Provide reusable processing units that implement business logic only. Framework handles all infrastructure (connections, lifecycle, routing). Agents are stateless, composable, and deployment-agnostic.

## Agent Concept

### Core Principle

**Agents implement one function: ProcessMessage()** - everything else handled by framework.

```go
type Agent interface {
    ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error)
}
```

**Framework provides:**
- BaseAgent initialization
- Support/Broker connections
- Message ingress/egress
- Lifecycle management
- Error recovery

**Agent provides:**
- Business logic transformation
- Input → Output processing
- Optional lifecycle hooks

### Agent vs Agent Type vs Agent Instance

**Agent Type (pool.yaml):**
- Reusable template
- Binary path and capabilities
- Deployment strategy (call/spawn/await)

**Agent Instance (cells.yaml):**
- Specific deployment
- Unique ID and configuration
- Ingress/egress routing
- Dependencies

**Agent Implementation:**
- Go code implementing ProcessMessage()
- Business logic only

**Example:**
```yaml
# pool.yaml - Agent Type
agents:
  - type: "text-transformer"
    binary: "./bin/text_transformer"
    operator: "spawn"

# cells.yaml - Agent Instance
agents:
  - id: "transformer-001"
    agent_type: "text-transformer"
    ingress: "sub:raw"
    egress: "pub:processed"
```

## Base Components

### AgentRunner Interface

**Minimal interface for agent implementation.**

```go
type AgentRunner interface {
    // Required: Process single message
    ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error)

    // Optional: Lifecycle hooks
    OnStart(base *BaseAgent) error  // Initialization
    OnStop(base *BaseAgent)          // Cleanup
}
```

### DefaultAgentRunner

**Base implementation providing default lifecycle behavior.**

```go
type DefaultAgentRunner struct{}

func (d *DefaultAgentRunner) ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error) {
    return msg, nil  // Pass-through by default
}

func (d *DefaultAgentRunner) OnStart(base *BaseAgent) error { return nil }
func (d *DefaultAgentRunner) OnStop(base *BaseAgent) {}
```

**Usage:**
```go
type MyAgent struct { agent.DefaultAgentRunner }

func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error) {
    // Override only ProcessMessage - lifecycle handled by default
    result := transform(msg.Payload)
    return &client.BrokerMessage{Payload: result}, nil
}
```

### BaseAgent

**Provided by framework - connection and state management.**

```go
type BaseAgent struct {
    ID          string
    Type        string
    Config      map[string]interface{}

    // Service clients
    SupportClient *support.Client
    BrokerClient  *broker.Client

    // State
    State       AgentState
    Context     context.Context
}

// Methods
func (b *BaseAgent) UpdateState(state AgentState) error
func (b *BaseAgent) GetConfig(key string) interface{}
func (b *BaseAgent) Log(level, message string)
```

### Framework Entry Point

**Single function to run any agent.**

```go
func Run(runner AgentRunner, agentType string) error {
    // Framework handles:
    // 1. Parse command-line flags (agent-id, support-url, broker-url)
    // 2. Create BaseAgent and connect to services
    // 3. Register with Support Service
    // 4. Fetch configuration
    // 5. Call runner.OnStart()
    // 6. Enter message processing loop
    // 7. Handle shutdown signals
    // 8. Call runner.OnStop()

    return nil
}
```

## Agent Patterns

### Stateless Processor

**Transform input to output without state.**

```go
type TransformAgent struct { agent.DefaultAgentRunner }

func (a *TransformAgent) ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error) {
    result := transform(msg.Payload)
    return &client.BrokerMessage{Payload: result}, nil
}
```

**Characteristics:**
- No persistent state
- Scales horizontally
- Idempotent operations
- Fast processing

### Stateful Accumulator

**Maintain state across messages.**

```go
type AggregatorAgent struct {
    agent.DefaultAgentRunner
    buffer []Message
}

func (a *AggregatorAgent) ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error) {
    a.buffer = append(a.buffer, msg)

    if len(a.buffer) >= threshold {
        result := aggregate(a.buffer)
        a.buffer = nil  // Reset
        return &client.BrokerMessage{Payload: result}, nil
    }

    return nil, nil  // No output yet
}
```

**Characteristics:**
- In-memory state
- Batching/aggregation
- Single-instance deployment
- Requires state management

### Storage-Backed Agent

**Persist state to OmniStore.**

```go
type IndexerAgent struct {
    agent.DefaultAgentRunner
    storage *omnistore.Store
}

func (a *IndexerAgent) OnStart(base *BaseAgent) error {
    a.storage = omnistore.Open(base.GetConfig("storage_path").(string))
    return nil
}

func (a *IndexerAgent) ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error) {
    // Extract and index
    data := extract(msg.Payload)
    a.storage.KV().Set(data.ID, data)
    a.storage.Graph().CreateVertex("document", data)

    return &client.BrokerMessage{Payload: data.ID}, nil
}

func (a *IndexerAgent) OnStop(base *BaseAgent) {
    a.storage.Close()
}
```

**Characteristics:**
- Durable state
- Queryable storage
- Relationship tracking
- Fault-tolerant

### File Watcher Agent

**React to filesystem changes.**

```go
type FileAgent struct { agent.DefaultAgentRunner }

func (a *FileAgent) ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error) {
    // msg.Payload contains file path from ingress: "file:*.txt"
    content, _ := os.ReadFile(msg.Payload.(string))

    return &client.BrokerMessage{
        Payload: map[string]interface{}{
            "path": msg.Payload,
            "content": string(content),
        },
    }, nil
}
```

**Characteristics:**
- Ingress: `file:path/*.ext`
- Triggered by file creation
- Digest tracking (avoid reprocessing)
- Sequential processing

## Agent Categories

### File Processing
- **file-ingester** - Watch directories, emit file events
- **file-writer** - Write output files with templates
- **file-splitter** - Split files into chunks

### Text Processing
- **text-extractor** - Extract text from PDF/DOCX/XLSX
- **text-transformer** - Transform text with metadata
- **text-chunker** - Intelligent text chunking
- **text-analyzer** - Sentiment, keywords, language

### Content Analysis
- **json-analyzer** - JSON validation and schema
- **xml-analyzer** - XML validation and namespace
- **binary-analyzer** - File type and hash
- **image-analyzer** - Metadata and dimensions

### Storage & Search
- **godast-storage** - OmniStore integration
- **search-indexer** - Full-text indexing
- **metadata-collector** - Metadata extraction
- **chunk-writer** - Chunk persistence

### Advanced Processing
- **ner-agent** - Named entity recognition
- **ocr-http-stub** - OCR service client
- **context-enricher** - Context enhancement
- **summary-generator** - Text summarization

### Pipeline Utilities
- **adapter** - Protocol adaptation
- **strategy-selector** - Dynamic routing
- **report-generator** - Report synthesis
- **dataset-builder** - Dataset construction

## Module Structure

```
code/agents/
├── go.mod                    # Module: github.com/tenzoki/agen/agents
├── file_ingester/
│   ├── main.go              # Agent implementation
│   ├── file_ingester_test.go
│   └── README.md            # Agent documentation
├── text_transformer/
│   ├── main.go
│   ├── text_transformer_test.go
│   └── README.md
├── ... (27 total agents)
└── testutil/                # Shared test utilities
    └── helpers.go
```

## Development Workflow

### 1. Define Agent Type

```yaml
# pool.yaml
agents:
  - type: "my-agent"
    binary: "./bin/my_agent"
    operator: "spawn"
    capabilities: ["custom-processing"]
```

### 2. Implement Agent

```go
// code/agents/my_agent/main.go
package main

import "github.com/tenzoki/agen/cellorg/public/agent"
import "github.com/tenzoki/agen/cellorg/public/client"

type MyAgent struct { agent.DefaultAgentRunner }

func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    // Business logic here
    return &client.BrokerMessage{Payload: result}, nil
}

func main() {
    agent.Run(&MyAgent{}, "my-agent")
}
```

### 3. Deploy in Cell

```yaml
# cells.yaml
cells:
  - id: "my-pipeline"
    agents:
      - id: "my-agent-001"
        agent_type: "my-agent"
        ingress: "file:input/*.txt"
        egress: "pub:processed"
```

### 4. Build and Run

```bash
go build -o bin/my_agent ./code/agents/my_agent
./bin/orchestrator -config=./workbench/config/cells.yaml
```

## Dependencies

**Internal:**
- `github.com/tenzoki/agen/cellorg` - Agent framework
- `github.com/tenzoki/agen/omni` - Storage (optional)

**External:**
- Varies by agent (OCR, NLP libraries, etc.)

## Setup

```bash
# Build all agents
for agent in code/agents/*/; do
    go build -o bin/$(basename $agent) $agent
done

# Build specific agent
go build -o bin/text_transformer ./code/agents/text_transformer
```

## Tests

**Agent Tests:**
- Business logic validation
- Message transformation correctness
- Error handling

**Integration Tests:**
- End-to-end with framework
- Multi-agent pipelines

**Run Tests:**
```bash
go test ./code/agents/... -v
```

## Demo

Agent-specific demos in individual README.md files:
- `code/agents/file_ingester/README.md`
- `code/agents/text_transformer/README.md`
- etc.

Multi-agent pipeline demos in `workbench/demos/`:
- `gox_demo/` - Full pipeline
- `text_extraction/` - OCR + extraction
