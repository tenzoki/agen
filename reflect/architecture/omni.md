# Omni

**Target Audience**: AI/LLM
**Purpose**: Unified storage system specification

Unified storage backend (OmniStore) - KV + Graph + FileStore + Query + Transactions in single API.

**Key Principle**: Single public API (omni/public/omnistore) - See `guidelines/references/architecture.md`

## Intent

Eliminate data silos by providing unified storage for all data types. Single API for key-value, graph relationships, file blobs, queries, and ACID transactions - no external dependencies required.

## Core Concept

### Unified Storage

**Single backend for all persistence needs:**

```
OmniStore = KV + Graph + FileStore + Query + Transactions
```

**Benefits:**
- **Single API** - One interface, multiple storage modes
- **Cross-domain queries** - Join KV data with graph relationships
- **Content deduplication** - Hash-based file storage
- **Transaction consistency** - ACID across all stores
- **Zero external deps** - BadgerDB backend embedded

**Anti-Pattern (Data Silos):**
```
Config → PostgreSQL
Relationships → Neo4j
Files → S3
Metadata → Redis
(Multiple APIs, inconsistent transactions, deployment complexity)
```

**OmniStore Pattern:**
```go
store := omnistore.Open("/data")

// Single transaction across all stores
tx := store.Begin()
tx.KV().Set("config:agent-001", config)
tx.Graph().CreateEdge(source, target, "processes")
tx.FileStore().Store(content, metadata)
tx.Commit()
```

## Components

### Key-Value Store

**Fast lookups for configuration, state, and metadata.**

```go
type KVStore interface {
    Set(key string, value interface{}) error
    Get(key string) (interface{}, error)
    Delete(key string) error
    List(prefix string) ([]string, error)
    Exists(key string) bool
}
```

**Use Cases:**
- Agent configuration and state
- Feature flags and settings
- Performance metrics
- Temporary data storage

**Access Pattern:**
```go
kv := store.KV()
kv.Set("agent:transformer-001:state", "running")
kv.Set("config:pipeline:timeout", 30)

state, _ := kv.Get("agent:transformer-001:state")
```

### Graph Database

**Relationships, dependencies, and traversals.**

```go
type GraphStore interface {
    CreateVertex(label string, properties map[string]interface{}) (*Vertex, error)
    CreateEdge(from, to *Vertex, relation string) (*Edge, error)
    GetVertex(id string) (*Vertex, error)
    Traverse(start *Vertex, pattern TraversalPattern) ([]*Vertex, error)
    Query(query string) ([]map[string]interface{}, error)
}
```

**Use Cases:**
- Agent dependency graphs
- Document relationship networks
- Knowledge graphs
- Workflow topology

**Access Pattern:**
```go
graph := store.Graph()

// Create vertices
agent1 := graph.CreateVertex("agent", map[string]interface{}{
    "id": "transformer-001",
    "type": "text-transformer",
})
agent2 := graph.CreateVertex("agent", map[string]interface{}{
    "id": "writer-001",
    "type": "file-writer",
})

// Create relationship
graph.CreateEdge(agent1, agent2, "sends_to")

// Traverse
dependencies := graph.Traverse(agent2, TraversalPattern{
    Direction: "incoming",
    Relation: "sends_to",
})
```

### FileStore

**Content-addressable blob storage with deduplication.**

```go
type FileStore interface {
    Store(data []byte, metadata map[string]interface{}) (hash string, error)
    Retrieve(hash string) (data []byte, metadata map[string]interface{}, error)
    Delete(hash string) error
    List(pattern string) ([]string, error)
    Metadata(hash string) (map[string]interface{}, error)
}
```

**Features:**
- **Content-addressed** - SHA256 hash as key
- **Automatic deduplication** - Same content = same hash
- **Metadata preservation** - Store alongside content
- **Efficient retrieval** - Direct hash lookup

**Access Pattern:**
```go
files := store.FileStore()

// Store file (auto-deduplicated)
hash, _ := files.Store(fileContent, map[string]interface{}{
    "filename": "document.pdf",
    "size": len(fileContent),
    "type": "application/pdf",
})

// Retrieve by hash
content, meta, _ := files.Retrieve(hash)
```

### Query Language

**Cross-store queries and aggregations.**

```go
type QueryEngine interface {
    Execute(query string) ([]map[string]interface{}, error)
    Aggregate(collection string, pipeline []AggregateOp) (interface{}, error)
}
```

**Query Examples:**
```go
query := store.Query()

// Find all running agents
results := query.Execute(`
    SELECT * FROM kv
    WHERE key LIKE 'agent:%:state'
    AND value = 'running'
`)

// Join KV with Graph
results := query.Execute(`
    SELECT a.id, a.state, COUNT(e.relation) as dependencies
    FROM kv a
    JOIN graph e ON e.to = a.id
    WHERE e.relation = 'depends_on'
    GROUP BY a.id
`)
```

### Transaction System

**ACID operations across all stores.**

```go
type Transaction interface {
    KV() KVStore
    Graph() GraphStore
    FileStore() FileStore
    Query() QueryEngine

    Commit() error
    Rollback() error
}
```

**Isolation Levels:**
- **Serializable** - Full ACID guarantees
- **Snapshot** - Read-committed with versioning
- **Read Uncommitted** - Maximum performance

**Transaction Pattern:**
```go
// Atomic operation across stores
tx := store.Begin()

// Update configuration
tx.KV().Set("config:pipeline:v2", newConfig)

// Update graph topology
tx.Graph().CreateEdge(sourceAgent, targetAgent, "routes_to")

// Store processing result
tx.FileStore().Store(result, metadata)

// All-or-nothing commit
if err := tx.Commit(); err != nil {
    tx.Rollback()
}
```

## Public API

### OmniStore Interface

**Single entry point for all storage operations.**

```go
// Open store
func Open(path string) (*Store, error)
func OpenWithOptions(path string, opts Options) (*Store, error)

// Store interface
type Store interface {
    // Direct access
    KV() KVStore
    Graph() GraphStore
    FileStore() FileStore
    Query() QueryEngine

    // Transactions
    Begin() Transaction
    BeginWithLevel(level IsolationLevel) Transaction

    // Lifecycle
    Close() error
    Backup(path string) error
    Restore(path string) error
}
```

### Options and Configuration

```go
type Options struct {
    ReadOnly      bool
    SyncWrites    bool
    CacheSize     int64
    MaxBatchSize  int
    ValueLogSize  int64
}

// Create optimized store
store := omnistore.OpenWithOptions("/data", omnistore.Options{
    SyncWrites: true,        // Durability
    CacheSize: 100 << 20,    // 100MB cache
})
```

## Integration Patterns

### Agent State Persistence

```go
type StorageAgent struct {
    agent.DefaultAgentRunner
    store *omnistore.Store
}

func (a *StorageAgent) OnStart(base *agent.BaseAgent) error {
    a.store = omnistore.Open(base.GetConfig("storage_path").(string))
    return nil
}

func (a *StorageAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    // Store message in KV
    a.store.KV().Set(msg.ID, msg)

    // Create graph relationship
    source := a.store.Graph().GetVertex(msg.Source)
    target := a.store.Graph().CreateVertex("message", msg)
    a.store.Graph().CreateEdge(source, target, "produced")

    return msg, nil
}
```

### Large Payload Management

```go
func (a *Agent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    if len(msg.Payload) > threshold {
        // Store large payload in FileStore
        hash, _ := a.store.FileStore().Store(msg.Payload, map[string]interface{}{
            "message_id": msg.ID,
            "timestamp": time.Now(),
        })

        // Replace with reference
        msg.Payload = map[string]string{
            "type": "file_reference",
            "hash": hash,
        }
    }

    return msg, nil
}
```

### Knowledge Graph Construction

```go
// Build document knowledge graph
tx := store.Begin()

// Create document vertex
doc := tx.Graph().CreateVertex("document", map[string]interface{}{
    "id": docID,
    "title": title,
})

// Create entity vertices and relationships
for _, entity := range entities {
    ent := tx.Graph().CreateVertex("entity", entity)
    tx.Graph().CreateEdge(doc, ent, "mentions")
}

// Store document content
tx.FileStore().Store(content, metadata)

tx.Commit()
```

## Module Structure

```
code/omni/
├── go.mod                    # Module: github.com/tenzoki/agen/omni
├── public/omnistore/         # Public API
│   ├── store.go             # Store interface
│   ├── kv.go                # KV interface
│   ├── graph.go             # Graph interface
│   ├── filestore.go         # FileStore interface
│   └── query.go             # Query interface
└── internal/                 # Internal implementation
    ├── common/              # Shared types
    ├── kv/                  # KV implementation
    ├── graph/               # Graph implementation
    ├── filestore/           # FileStore implementation
    ├── query/               # Query engine
    ├── transaction/         # Transaction manager
    └── storage/             # BadgerDB backend
```

## Backend Storage

**BadgerDB** - Embedded LSM key-value store

**Benefits:**
- No external database required
- Embedded in-process
- High-performance LSM tree
- ACID transactions
- Compression and encryption

**Storage Layout:**
```
/data/
├── kv/          # Key-value data
├── graph/       # Graph vertices and edges
├── files/       # Content-addressed blobs
└── indexes/     # Query indexes
```

## Dependencies

**Internal:**
- `github.com/tenzoki/agen/atomic` - VFS utilities

**External:**
- `github.com/dgraph-io/badger/v4` - Storage backend

## Setup

```bash
# Build omni module
go build -v ./code/omni/...

# Run tests
go test ./code/omni/... -v
```

## Tests

**Component Tests:**
- KV operations and queries
- Graph traversals and relationships
- FileStore deduplication
- Query engine
- Transaction isolation

**Integration Tests:**
- Cross-store transactions
- Concurrent access
- Data consistency
- Performance benchmarks

**Run Tests:**
```bash
go test ./code/omni/public/omnistore -v
go test ./code/omni/internal/... -v
```

## Demo

**Component Demos** (in `code/omni/internal/`):
- `kv/kv_demo.go` - Key-value operations
- `graph/graph_demo.go` - Graph traversals
- `filestore/filestore_demo.go` - File storage
- `query/query_demo.go` - Query examples

**Integration Demos:**
- `internal/unified_demo.go` - Cross-store operations
- `internal/transaction_demo.go` - ACID transactions

**Run Demo:**
```bash
# KV demo
go run ./code/omni/internal/kv/kv_demo.go

# Full integration demo
go run ./code/omni/internal/unified_demo.go
```
