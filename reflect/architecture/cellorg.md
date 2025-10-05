# Cellorg

Cell-based orchestration framework - cells as core units for AI operations, agent framework, and distributed coordination.

## Intent

Provide zero-boilerplate infrastructure for building distributed processing pipelines through composable cells. Cells are the primary abstraction - self-contained units combining agents, dependencies, and routing.

## Core Concepts

### Cell

**Self-contained processing unit** defined declaratively in YAML.

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
      egress: "pipe:output"
```

**Cell Properties:**
- **Declarative** - YAML defines structure, framework handles execution
- **Composable** - Cells combine to form complex workflows
- **Dependency-driven** - Topological ordering ensures correct startup
- **Self-contained** - All configuration in single definition

### Agent Framework

**Zero-boilerplate agent development** - write business logic only, framework provides all infrastructure.

**Traditional Agent (120+ lines):**
```go
// BaseAgent setup, connection management, signal handling,
// message loop, error recovery, lifecycle coordination...
```

**Framework Agent (3 lines):**
```go
type MyAgent struct { agent.DefaultAgentRunner }
func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) { ... }
func main() { agent.Run(&MyAgent{}, "my-agent") }
```

**Framework Responsibilities:**
- BaseAgent initialization and lifecycle
- Support/Broker service connections with retry
- Ingress/egress message routing
- Signal handling and graceful shutdown
- State transitions and error recovery

### Communication Patterns

**File Pattern:** `file:path/*.ext`
- Filesystem watching with glob support
- Triggers on file creation/modification
- Automatic digest tracking (avoid reprocessing)

**Pub/Sub:** `pub:topic` / `sub:topic`
- One-to-many event broadcasting
- Multiple subscribers per topic
- Fire-and-forget delivery

**Pipe:** `pipe:name`
- Point-to-point guaranteed delivery
- Single consumer per pipe
- Message queuing and ordering

**Pattern Usage:**
```yaml
agents:
  - ingress: "file:input/*.json"    # Watch filesystem
    egress: "pub:raw-data"          # Broadcast to subscribers

  - ingress: "sub:raw-data"         # Subscribe to topic
    egress: "pipe:processed"        # Send to specific consumer
```

## Components

### Orchestrator

**Dependency resolution and startup coordination.**

**Responsibilities:**
- Parse cell YAML definitions
- Build dependency graph (topological sort)
- Deploy agents in correct order
- Coordinate lifecycle transitions
- Handle graceful shutdown (reverse order)

**Deployment Strategies:**
- **call** - Embedded execution (same process)
- **spawn** - Isolated process (separate binary)
- **await** - External agent (remote connection)

**API:**
```go
type Orchestrator interface {
    LoadCell(path string) error
    DeployCell(cellID string) error
    StopCell(cellID string) error
    GetStatus(cellID string) (*CellStatus, error)
}
```

### Support Service (Port 9000)

**Agent registry, configuration distribution, and health monitoring.**

**Endpoints:**
- `POST /agents/register` - Agent self-registration
- `POST /agents/ready` - Signal readiness
- `GET /agents/:id/config` - Fetch configuration
- `POST /agents/:id/status` - Update status
- `GET /health` - Service health check

**Responsibilities:**
- Maintain agent registry
- Distribute cell configurations
- Track agent states
- Provide service discovery

**State Transitions:**
```
Installed → Configured → Ready → Running → [Paused|Stopped|Error]
```

### Broker Service (Port 9001)

**Message routing - pub/sub topics and point-to-point pipes.**

**Protocol:** JSON-over-TCP

**Operations:**
- `subscribe:topic` - Subscribe to topic
- `publish:topic:payload` - Publish to topic
- `pipe_send:name:payload` - Send to pipe
- `pipe_receive:name` - Receive from pipe

**Message Envelope:**
```go
type Envelope struct {
    ID, TraceID, CorrelationID string
    Source, Destination string
    Timestamp time.Time
    Payload json.RawMessage
    Route []string      // Full message path
    HopCount int        // Processing depth
}
```

**Responsibilities:**
- Route messages based on pattern
- Maintain topic subscriptions
- Buffer pipe messages
- Automatic connection management

### Deployer

**Agent deployment and process management.**

**Deployment Modes:**
- **Embedded (call)** - `agent.Run()` in same process
- **Process (spawn)** - `exec.Command()` separate process
- **External (await)** - Wait for remote agent

**Lifecycle Management:**
```go
type Deployer interface {
    Deploy(agentConfig AgentConfig) error
    Stop(agentID string) error
    Restart(agentID string) error
    GetStatus(agentID string) (*AgentStatus, error)
}
```

## Public APIs

### Agent API (`public/agent`)

**For implementing custom agents.**

```go
// Agent interface
type AgentRunner interface {
    ProcessMessage(msg *client.BrokerMessage, base *BaseAgent) (*client.BrokerMessage, error)
    OnStart(base *BaseAgent) error  // Optional lifecycle hook
    OnStop(base *BaseAgent)          // Optional cleanup hook
}

// Default implementation
type DefaultAgentRunner struct{}

// Framework entry point
func Run(runner AgentRunner, agentType string) error
```

### Client API (`public/client`)

**For agent-broker communication.**

```go
// Broker client
type BrokerClient interface {
    Subscribe(topic string) (<-chan *BrokerMessage, error)
    Publish(topic string, payload interface{}) error
    SendPipe(pipe string, payload interface{}) error
    ReceivePipe(pipe string) (*BrokerMessage, error)
}

// Message structure
type BrokerMessage struct {
    ID, TraceID string
    Payload json.RawMessage
    Metadata map[string]string
}
```

### Orchestrator API (`public/orchestrator`)

**For cell management.**

```go
// Cell orchestration
type CellOrchestrator interface {
    LoadCell(yamlPath string) (*Cell, error)
    DeployCell(cell *Cell) error
    StopCell(cellID string) error
}

// Event system
type EventBus interface {
    Subscribe(event EventType) <-chan Event
    Publish(event Event)
}
```

## Module Structure

```
code/cellorg/
├── go.mod                    # Module: github.com/tenzoki/agen/cellorg
├── cmd/orchestrator/         # Orchestrator binary
│   └── main.go
├── internal/                 # Internal implementation
│   ├── agent/               # Agent framework
│   ├── broker/              # Broker service
│   ├── client/              # Broker client
│   ├── orchestrator/        # Orchestrator core
│   ├── deployer/            # Agent deployment
│   ├── support/             # Support service
│   ├── envelope/            # Message envelope
│   ├── storage/             # Storage client
│   └── chunks/              # Chunk management
└── public/                   # Public APIs
    ├── agent/               # Agent framework API
    ├── client/              # Broker client API
    ├── orchestrator/        # Orchestrator API
    └── examples/            # Example cells
```

## Configuration Files

### pool.yaml
**Agent type definitions** - reusable templates.

```yaml
agents:
  - type: "text-transformer"
    binary: "./bin/text_transformer"
    operator: "spawn"
    capabilities: ["text-processing", "nlp"]
    description: "Transform text with metadata preservation"
```

### cells.yaml
**Agent instance deployment** - specific configurations.

```yaml
cells:
  - id: "text-pipeline"
    agents:
      - id: "transformer-001"
        agent_type: "text-transformer"
        dependencies: []
        ingress: "file:input/*.txt"
        egress: "pub:processed"
        config:
          custom_key: "value"
```

### gox.yaml
**Infrastructure settings** - service configuration.

```yaml
infrastructure:
  support_port: 9000
  broker_port: 9001
  debug: true
  log_level: "info"
```

## Cell Execution Flow

```
1. Load Configuration
   ├── Parse pool.yaml (agent types)
   ├── Parse cells.yaml (agent instances)
   └── Parse gox.yaml (infrastructure)

2. Dependency Resolution
   ├── Build dependency graph
   ├── Topological sort
   └── Detect cycles/deadlocks

3. Agent Deployment
   ├── Start agents in order
   ├── Wait for readiness signals
   └── Establish connections

4. Message Processing
   ├── Route messages via patterns
   ├── Process through pipeline
   └── Track via envelopes

5. Graceful Shutdown
   ├── Stop in reverse order
   ├── Drain message queues
   └── Cleanup resources
```

## Dependencies

**Internal:**
- `github.com/tenzoki/agen/atomic` - VFS, VCR

**External:**
- Go standard library
- BadgerDB (via storage)

## Setup

```bash
# Build orchestrator
go build -o bin/orchestrator ./code/cellorg/cmd/orchestrator

# Run with config
./bin/orchestrator -config=./workbench/config/cells.yaml
```

## Tests

**Test Structure:**
- Co-located with source files
- Integration tests in `public/examples/`
- End-to-end tests for full cells

**Run Tests:**
```bash
go test ./code/cellorg/... -v
```

## Demo

**Cell Demos** (`public/examples/`):
- `file-transform-pipeline/` - File processing cell
- `text-extraction/` - OCR + text extraction
- `search-indexing/` - Content indexing cell

**Run Demo:**
```bash
# Start orchestrator with example cell
./bin/orchestrator -config=./code/cellorg/public/examples/text-extraction/cell.yaml
```
