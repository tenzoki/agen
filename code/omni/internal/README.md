# Internal

Internal implementation packages for OmniStore.

## Intent

Contains all private implementation details of OmniStore including KV storage, graph operations, file handling, query processing, and transactions. These packages are not meant to be imported directly by external code.

## Usage

Internal packages are accessed only through the public OmniStore interface.

## Setup

No direct setup required - used internally by `public/omnistore`.

## Tests

Each package has co-located tests:
- `common/` - Core types and validation
- `filestore/` - Content-addressable file storage
- `graph/` - Graph database operations
- `kv/` - Key-value store
- `query/` - Query language parser and executor
- `storage/` - Underlying Badger storage layer
- `transaction/` - ACID transaction support

Run all internal tests:
```bash
go test ./internal/...
```

## Demo

Internal packages are not demoed directly. See `public/omnistore` README for usage examples.
