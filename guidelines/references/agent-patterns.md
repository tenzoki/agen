# Agent Development Guide

**Audience**: AI/LLM
**Purpose**: Practical guide for creating new AGEN agents
**Last Updated**: 2025-10-06

---

## Quick Start

**3-method pattern** - Agents implement only:
```go
type MyAgent struct {
    agent.DefaultAgentRunner
    // Your fields here
}

func (a *MyAgent) Init(base *agent.BaseAgent) error { }
func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) { }
func (a *MyAgent) Cleanup(base *agent.BaseAgent) error { }
```

**Main function** - Single line:
```go
func main() {
    agent.Run(&MyAgent{}, "my-agent")
}
```

Done. Framework handles all infrastructure.

---

## File Structure

```
code/agents/my_agent/
├── main.go              # Main entry point
├── agent.go             # Agent implementation
├── main_test.go         # Tests
├── go.mod               # Module: github.com/tenzoki/agen/agents
└── README.md            # Documentation
```

---

## Agent Implementation Pattern

### 1. Imports (Public APIs Only)

```go
import (
    "github.com/tenzoki/agen/cellorg/public/agent"   // ✅ Agent framework
    "github.com/tenzoki/agen/cellorg/public/client"  // ✅ Messaging
    "github.com/tenzoki/agen/omni/public/omnistore"  // ✅ Storage
    "github.com/tenzoki/agen/atomic/vfs"             // ✅ File operations
)

// ❌ NEVER import internal packages:
// github.com/tenzoki/agen/cellorg/internal/*
// github.com/tenzoki/agen/omni/internal/*
```

### 2. Agent Struct

```go
type MyAgent struct {
    agent.DefaultAgentRunner  // Embed default runner

    // Optional: Add storage if needed
    omniStore omnistore.OmniStore

    // Your custom fields
    customField string
}
```

### 3. Init Method

**Purpose**: Initialize resources, connect to storage

```go
func (a *MyAgent) Init(base *agent.BaseAgent) error {
    // Get config values (from file or support service)
    dataPath := base.GetConfigString("data_path", "./data/myagent")
    a.customField = base.GetConfigString("custom_field", "default")

    // Initialize storage (if needed)
    store, err := omnistore.NewOmniStoreWithDefaults(dataPath)
    if err != nil {
        return err
    }
    a.omniStore = store

    base.LogInfo("MyAgent initialized with data_path=%s", dataPath)
    return nil
}
```

### 4. ProcessMessage Method

**Purpose**: Core business logic

**Pattern 1: Transform and forward**
```go
func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    // Extract payload
    data, ok := msg.Payload.([]byte)
    if !ok {
        return nil, fmt.Errorf("invalid payload type")
    }

    // Process data
    result := a.doWork(data)

    // Return transformed message
    return &client.BrokerMessage{
        ID:      msg.ID,
        Payload: result,
        Source:  base.ID,
    }, nil
}
```

**Pattern 2: Side effects only** (file writers, storage)
```go
func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    // Process and store
    data := msg.Payload.([]byte)

    // Save to storage
    key := fmt.Sprintf("item:%s", msg.ID)
    json.Marshal(data) // Marshal to []byte
    a.omniStore.KV().Set(key, marshaled)

    // No output message
    return nil, nil
}
```

### 5. Cleanup Method

**Purpose**: Release resources

```go
func (a *MyAgent) Cleanup(base *agent.BaseAgent) error {
    if a.omniStore != nil {
        a.omniStore.Close()
    }
    base.LogInfo("MyAgent cleanup complete")
    return nil
}
```

### 6. Main Function

```go
func main() {
    agent.Run(&MyAgent{}, "my-agent")
}
```

---

## Storage Patterns

### KV Store (JSON marshaling)

```go
// Write
type MyData struct { Field string }
data := MyData{Field: "value"}
bytes, _ := json.Marshal(data)
a.omniStore.KV().Set("key", bytes)

// Read
bytes, _ := a.omniStore.KV().Get("key")
var data MyData
json.Unmarshal(bytes, &data)
```

### Graph Store

```go
// Create vertex
vertex := &graph.Vertex{
    ID:   "user:123",
    Type: "user",
    Properties: map[string]interface{}{
        "name": "Alice",
        "age":  30,
    },
}
a.omniStore.Graph().AddVertex(vertex)

// Create edge
edge := &graph.Edge{
    ID:         "follows:1",
    Type:       "follows",
    SourceID:   "user:123",
    TargetID:   "user:456",
    Properties: map[string]interface{}{"since": "2025-01-01"},
}
a.omniStore.Graph().AddEdge(edge)

// Query
vertices := a.omniStore.Graph().GetVerticesByType("user")
```

### File Store

```go
// Store file
metadata := &filestore.FileMetadata{
    ID:          "file:123",
    Filename:    "doc.pdf",
    ContentType: "application/pdf",
    Size:        1024,
}
a.omniStore.Files().Store(metadata, fileBytes)

// Retrieve
metadata, content, _ := a.omniStore.Files().Get("file:123")
```

---

## Configuration

### Agent Config File

**Location**: `workbench/config/agents/my-agent.yaml`

```yaml
data_path: "./data/myagent"
custom_field: "production_value"
debug: false
```

### Resolution Order

1. `--config` flag
2. `AGEN_CONFIG_PATH` env var
3. `AGEN_WORKBENCH_DIR/config/agents/my-agent.yaml`
4. `./config/my-agent.yaml`
5. `./workbench/config/agents/my-agent.yaml`
6. Binary-relative config
7. Embedded defaults

**In code**:
```go
value := base.GetConfigString("key", "default")
intVal := base.GetConfigInt("port", 8080)
boolVal := base.GetConfigBool("enabled", true)
```

---

## Testing

### Test File: `main_test.go`

```go
package main

import (
    "testing"
    "github.com/tenzoki/agen/cellorg/public/agent"
    "github.com/tenzoki/agen/cellorg/public/client"
)

func TestAgentInit(t *testing.T) {
    // Create test base agent
    config := agent.AgentConfig{
        ID:        "test-agent",
        AgentType: "my-agent",
        Debug:     true,
    }

    // Skip support connection in tests
    // Use mock or embedded mode
}

func TestProcessMessage(t *testing.T) {
    myAgent := &MyAgent{}

    msg := &client.BrokerMessage{
        ID:      "test-1",
        Payload: []byte("test data"),
    }

    result, err := myAgent.ProcessMessage(msg, nil)
    if err != nil {
        t.Fatalf("ProcessMessage failed: %v", err)
    }

    // Assert result
}
```

### Test Data

Use centralized: `/testdata/`

```go
import "github.com/tenzoki/agen/agents/testutil"

path := testutil.GetTestDataPath("documents/sample.txt")
```

---

## Cell Integration

### Define Agent in Pool

**File**: `workbench/config/pool.yaml`

```yaml
agent_types:
  - agent_type: my-agent
    binary: bin/my_agent
    capabilities:
      - data-processing
      - custom-capability
    description: "My custom agent"
```

### Use in Cell

**File**: `workbench/config/cells/my-cell.yaml`

```yaml
cells:
  - id: my-processing-cell
    description: "Custom processing pipeline"
    agents:
      - id: my-agent-001
        agent_type: my-agent
        ingress: "file:input/*.txt"
        egress: "pub:processed"
        config:
          data_path: "./data/myagent"
          custom_field: "value"
```

---

## Communication Patterns

### File-based (default)

```yaml
ingress: "file:input/*.txt"
egress: "file:output/"
```

Agent receives file path in message, processes, writes to output.

### Pub/Sub

```yaml
ingress: "sub:events"
egress: "pub:processed"
```

Subscribe to topic, publish results.

### Direct Pipe

```yaml
ingress: "pipe:previous-agent"
egress: "pipe:next-agent"
```

Direct message flow between agents.

---

## Build and Run

### Build

```bash
# From module root
go build -o bin/my_agent ./code/agents/my_agent/

# Or use Makefile
make build-agent-my_agent
```

### Run Standalone

```bash
# With config file
bin/my_agent --config=workbench/config/agents/my-agent.yaml

# With environment
AGEN_CONFIG_PATH=config.yaml bin/my_agent

# Auto-generated ID
bin/my_agent

# Specific ID (matches cell config)
bin/my_agent --agent-id=my-agent-001
```

### Run via Orchestrator

```bash
# Orchestrator spawns agent automatically
bin/orchestrator -config=workbench/config/cellorg.yaml
```

---

## Checklist

Agent development completion checklist:

- [ ] Implements 3 methods: Init, ProcessMessage, Cleanup
- [ ] Uses only public APIs (no internal imports)
- [ ] Main function calls `agent.Run()`
- [ ] Config values accessed via `base.GetConfig*()`
- [ ] Storage uses `omnistore` if needed
- [ ] Tests in `main_test.go`
- [ ] README.md documents agent purpose
- [ ] Added to `pool.yaml`
- [ ] Used in at least one cell YAML
- [ ] Builds successfully
- [ ] Tests pass

---

## Common Patterns

### Error Handling

```go
func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    data, ok := msg.Payload.([]byte)
    if !ok {
        base.LogError("Invalid payload type: %T", msg.Payload)
        return nil, fmt.Errorf("invalid payload type")
    }

    result, err := a.process(data)
    if err != nil {
        base.LogError("Processing failed for message %s: %v", msg.ID, err)
        return nil, err
    }

    base.LogInfo("Successfully processed message %s", msg.ID)
    return result, nil
}
```

### Conditional Logging

```go
base.LogInfo("Always visible")
base.LogDebug("Only when debug=true")  // From config or --debug flag
base.LogError("Error occurred: %v", err)
```

### VFS Access

```go
// BaseAgent provides VFS if configured
content, err := base.ReadFile("subdir", "file.txt")
base.WriteFile([]byte("content"), "output", "result.txt")
exists := base.FileExists("data", "input.json")
```

---

## Anti-Patterns (Avoid)

❌ Importing internal packages:
```go
import "github.com/tenzoki/agen/cellorg/internal/storage"  // NO!
```

❌ Manual infrastructure setup:
```go
// Don't do this - framework handles it
func main() {
    supportClient := client.NewSupportClient(...)
    brokerClient := client.NewBrokerClient(...)
    // ... 100+ lines of boilerplate
}
```

❌ Direct broker/support access:
```go
// Don't access BaseAgent internals
msg.BrokerClient.Send(...)  // NO!
```

❌ Global state:
```go
var globalStore omnistore.OmniStore  // NO! Use agent fields
```

✅ **Do**: Keep agents stateless where possible, store state in OmniStore

---

## Reference

- **Immutable Principles**: `guidelines/immutable-principles.md`
- **Configuration Guide**: `guidelines/configuration-conventions.md`
- **Architecture**: `reflect/architecture/`
- **Agent Examples**: `code/agents/*/`
- **Cell Examples**: `reflect/cells/`

---

## Example: Minimal Agent

```go
package main

import (
    "github.com/tenzoki/agen/cellorg/public/agent"
    "github.com/tenzoki/agen/cellorg/public/client"
)

type MinimalAgent struct {
    agent.DefaultAgentRunner
}

func (a *MinimalAgent) Init(base *agent.BaseAgent) error {
    base.LogInfo("Minimal agent initialized")
    return nil
}

func (a *MinimalAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    base.LogInfo("Processing message: %s", msg.ID)
    return msg, nil  // Pass through
}

func (a *MinimalAgent) Cleanup(base *agent.BaseAgent) error {
    base.LogInfo("Cleanup complete")
    return nil
}

func main() {
    agent.Run(&MinimalAgent{}, "minimal-agent")
}
```

**That's it.** 30 lines. Framework does the rest.
