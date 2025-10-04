# Omni

OmniStore - unified multi-model storage system providing KV, Graph, File, and Search capabilities.

## Intent

Omni provides a single unified interface for multiple storage paradigms, eliminating the need to integrate separate databases. It combines key-value, graph, file, and full-text search operations with ACID transactions and efficient storage through Badger DB.

## Usage

```go
import "agen/code/omni/public/omnistore"

// Create store
store, err := omnistore.NewOmniStoreWithDefaults("/path/to/data")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// Use KV operations
store.KV().Set("key", []byte("value"))
value, _ := store.KV().Get("key")

// Use graph operations
vertex := &common.Vertex{ID: "v1", Type: "user", Properties: props}
store.Graph().CreateVertex(vertex)

// Use file operations
hash, _ := store.Files().Store(fileData, metadata)
data, _ := store.Files().Retrieve(hash)

// List keys with prefix
keys, _ := store.ListKVKeys("prefix:", 1000)
```

## Setup

Dependencies:
- Badger DB v4
- MessagePack v5
- UUID

Install:
```bash
go get agen/code/omni
```

## Tests

Run all tests:
```bash
cd agen/code/omni
go test ./...
```

Test coverage:
- `internal/common` - Data types and validation
- `internal/graph` - Graph operations
- `internal/kv` - Key-value operations
- `internal/query` - Query language
- `internal/transaction` - Transaction management (1 failing test)
- `public/omnistore` - Integration tests

## Demo

See `godast_old/demo/` for working examples:
- `unified-omnistore-demo.go` - Complete workflow
- `filestore-standalone-demo.go` - File operations
- `filestore-working-demo.go` - Content-addressable storage
