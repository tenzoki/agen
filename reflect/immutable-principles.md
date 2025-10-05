# AGEN: Immutable Architectural Principles

**Last Updated**: 2025-10-05

These principles are fundamental to AGEN's architecture and **must not be violated**. They define the core design that makes AGEN a coherent, composable system.

---

## Core Principles

### 1. Cells-First Architecture
**Cells are the primary abstraction** - not agents, not services, not functions.

- **What**: A cell = agents + dependencies + routing + configuration
- **Why**: AI can reason about and modify high-level workflow structure
- **Immutable**: Cells remain the top-level composable unit
- **Implementation**: YAML cell definitions in `workbench/config/`

### 2. Zero External Dependencies for atomic
**The atomic layer has ZERO external dependencies** - only Go stdlib.

- **What**: VFS and VCR use only standard library
- **Why**: Foundation must be stable, portable, auditable
- **Immutable**: No external packages in `code/atomic/`
- **Exceptions**: None

### 3. Single Public Storage API
**Only `omni/public/omnistore` is public** - all other storage is internal.

- **What**: Agents use OmniStore; internal modules use internal storage
- **Why**: Clean API boundary, prevents coupling to implementation
- **Immutable**:
  - ✅ Agents import: `github.com/tenzoki/agen/omni/public/omnistore`
  - ❌ Agents NEVER import: `cellorg/internal/storage` or `omni/internal/*`
- **Correct pattern**:
  ```go
  store, _ := omnistore.NewOmniStoreWithDefaults(dataPath)
  store.KV().Set(key, []byte)  // JSON marshal if needed
  store.Graph().CreateVertex(...)
  store.Files().Store(...)
  ```

### 4. Zero-Boilerplate Agent Framework
**Agents implement only ProcessMessage()** - framework handles everything else.

- **What**: 3 methods (Init, ProcessMessage, Cleanup) vs 120+ lines of boilerplate
- **Why**: Focus on logic, not infrastructure
- **Immutable**: Agent interface stays minimal
- **Implementation**: `cellorg/public/agent` provides framework

---

## Module Principles

### atomic (VFS, VCR)
**Foundation utilities - zero dependencies**

**Immutable Principles**:
1. **No external dependencies** - stdlib only
2. **Security first** - Path validation, traversal prevention
3. **Deterministic** - Same inputs → same outputs
4. **Stateless** - No global state

**Public API**:
- `vfs.SecurePath()` - Path validation
- `vcr.NewVcr()` - Version control wrapper

**What CANNOT change**:
- Adding external dependencies
- Breaking path security guarantees
- Adding stateful operations

---

### omni (OmniStore)
**Unified storage - single backend for all data types**

**Immutable Principles**:
1. **Single public API** - Only `public/omnistore` is exposed
2. **Unified backend** - One Badger store for KV + Graph + Files
3. **ACID transactions** - Graph operations are transactional
4. **Interface-based** - Implementation can change, interface cannot

**Public API** (`omni/public/omnistore`):
```go
type OmniStore interface {
    KV() kv.KVStore           // Key-value operations
    Graph() graph.GraphStore   // Graph database
    Files() filestore.FileStore // File storage
    Search() SearchStore       // Full-text search
    Begin() Transaction        // ACID transactions
    Close() error
}
```

**What CANNOT change**:
- Making internal packages public
- Breaking the unified backend (no separate stores)
- Removing transaction support for graph operations
- Changing KV interface to use anything other than []byte

**Internal** (DO NOT EXPOSE):
- `omni/internal/kv` - KV implementation
- `omni/internal/graph` - Graph implementation
- `omni/internal/filestore` - File storage implementation
- `omni/internal/storage` - Badger backend

---

### cellorg (Cell Orchestration)
**Cell framework - orchestrates agents into workflows**

**Immutable Principles**:
1. **Cells-first** - Cells are the primary abstraction
2. **Declarative** - YAML defines structure
3. **Zero-boilerplate agents** - Framework handles plumbing
4. **Internal orchestration** - Storage client stays internal

**Public API** (`cellorg/public/*`):
- `agent` - Agent framework (BaseAgent, AgentRunner interface)
- `client` - Broker client for messaging
- `orchestrator` - Cell orchestration API

**Internal** (DO NOT EXPOSE):
- `cellorg/internal/storage` - Storage client for orchestration
- `cellorg/internal/chunks` - Chunk tracking
- `cellorg/internal/broker` - Message broker
- `cellorg/internal/orchestrator` - Orchestration engine

**What CANNOT change**:
- Exposing internal/storage to agents (agents use omnistore!)
- Breaking the 3-method agent interface
- Making cells non-declarative
- Removing dependency resolution from cells

**Critical**: `cellorg/internal/storage` is for orchestrator use only:
```go
// ✅ CORRECT: Orchestrator using internal storage
package chunks
import "github.com/tenzoki/agen/cellorg/internal/storage"

// ❌ WRONG: Agent using internal storage
package my_agent
import "github.com/tenzoki/agen/cellorg/internal/storage" // NO!
import "github.com/tenzoki/agen/omni/public/omnistore"   // YES!
```

---

### agents (Agent Implementations)
**Modular processing units - implement business logic**

**Immutable Principles**:
1. **Use only public APIs** - Never import internal packages
2. **Use omnistore for persistence** - Not internal storage clients
3. **Stateless where possible** - State goes in omnistore
4. **Self-contained** - Each agent is independent

**Correct Imports**:
```go
// ✅ ALLOWED
import "github.com/tenzoki/agen/cellorg/public/agent"
import "github.com/tenzoki/agen/cellorg/public/client"
import "github.com/tenzoki/agen/omni/public/omnistore"
import "github.com/tenzoki/agen/atomic/vfs"
import "github.com/tenzoki/agen/atomic/vcr"

// ❌ FORBIDDEN
import "github.com/tenzoki/agen/cellorg/internal/storage"  // NO!
import "github.com/tenzoki/agen/omni/internal/kv"          // NO!
import "github.com/tenzoki/agen/cellorg/internal/broker"   // NO!
```

**Storage Pattern**:
```go
type MyAgent struct {
    agent.DefaultAgentRunner
    omniStore omnistore.OmniStore  // ✅ Correct
}

func (a *MyAgent) Init(base *agent.BaseAgent) error {
    dataPath := base.GetConfigString("data_path", "./data/myagent")
    store, err := omnistore.NewOmniStoreWithDefaults(dataPath)
    a.omniStore = store
    return err
}

// Use KV with JSON marshaling
func (a *MyAgent) saveData(key string, value interface{}) error {
    data, _ := json.Marshal(value)
    return a.omniStore.KV().Set(key, data)
}

func (a *MyAgent) loadData(key string, value interface{}) error {
    data, _ := a.omniStore.KV().Get(key)
    return json.Unmarshal(data, value)
}
```

**What CANNOT change**:
- Agents importing internal packages
- Agents bypassing the agent framework
- Agents using anything other than omnistore for persistence

---

## Dependency Flow (Immutable)

```
atomic (stdlib only)
  ↓
omni/public/omnistore (PUBLIC - agents use this)
  ↓
omni/internal/* (INTERNAL - agents NEVER use this)
  ↓
cellorg/public/* (PUBLIC - agents use this)
  ↓
cellorg/internal/* (INTERNAL - orchestrator uses this)
  ↓
agents/* (Use public APIs only)
  ↓
alfa (Application layer)
```

**Rules**:
1. **Upward imports only** - Lower layers never import higher layers
2. **Public/internal boundary** - Never import internal from outside module
3. **atomic has zero deps** - Only stdlib

---

## Communication Patterns (Immutable)

### 1. File-based Communication
```yaml
routing:
  - source: agent-a
    target: agent-b
    pattern: "file:output/*.txt"
```

### 2. Pub/Sub Communication
```yaml
routing:
  - source: agent-a
    target: pub:events
  - source: sub:events
    target: agent-b
```

### 3. Direct Pipes
```yaml
routing:
  - source: agent-a
    target: pipe:agent-b
```

**What CANNOT change**:
- These three patterns define all communication
- File-based is default (most reliable)
- Pattern matching on file paths

---

## Self-Modification Capability (Immutable)

**AGEN can modify its own codebase** - This is a core feature.

**Requirements**:
1. **VCR tracks all changes** - Automatic git commits
2. **VFS validates paths** - Security boundaries enforced
3. **Cells remain valid** - YAML structure preserved
4. **Tests must pass** - Before changes are committed

**What CANNOT change**:
- Removing self-modification capability
- Bypassing VCR for code changes
- Breaking VFS security model

---

## Testing Principles (Immutable)

1. **All modules must have tests** - No exceptions
2. **Tests use public APIs only** - Internal tests are internal
3. **Test data in `/testdata/`** - Centralized location
4. **Coverage threshold: 70%** - Minimum for core modules

**Build tags for demos**:
```go
//go:build ignore

package main  // Demo files excluded from builds
```

---

## What IS Allowed to Change

These can evolve without breaking principles:

1. **Implementation details** - As long as public API stays stable
2. **Performance optimizations** - Internal refactoring OK
3. **New agents** - Adding agents is encouraged
4. **New cells** - Composing agents into cells is encouraged
5. **Storage backend** - Could replace Badger if needed (via internal swap)

---

## Violation Detection

**Signs you're breaking immutability**:

❌ Agent imports `cellorg/internal/storage`
❌ Agent imports `omni/internal/kv`
❌ Exposing internal packages to public
❌ atomic importing external packages
❌ Breaking the 3-method agent interface
❌ Multiple storage APIs for agents
❌ Bypassing VFS/VCR for file operations

**When in doubt**:
1. Check this document
2. Agents use: `omni/public/omnistore`, `cellorg/public/agent`, `cellorg/public/client`
3. Everything else stays internal

---

## Summary

**The 4 Immutable Pillars**:

1. **atomic**: Zero dependencies, security-first utilities
2. **omni**: Single public storage API (`omni/public/omnistore`)
3. **cellorg**: Cells-first orchestration, internal storage stays internal
4. **agents**: Use public APIs only, omnistore for persistence

**Golden Rule**:
> If an agent needs storage, it uses omnistore.
> If orchestrator needs storage, it uses internal storage.
> Never mix them.

This architecture makes AGEN:
- **Composable** - Cells combine agents flexibly
- **Evolvable** - Internals can change, APIs stay stable
- **Secure** - VFS/VCR enforce boundaries
- **Simple** - Agents are 3 methods, everything else is framework
