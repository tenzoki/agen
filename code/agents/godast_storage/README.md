# GoDAST Storage

Persistent storage service for Go AST (Abstract Syntax Tree) analysis results using OmniStore.

## Intent

Stores and retrieves Go source code AST analysis results with efficient querying capabilities. Uses OmniStore with bbolt backend for reliable persistence of parsed code structures, symbol tables, and dependency graphs.

## Usage

Input: Storage operations (set/get/query) for AST data
Output: Operation results with AST data

Configuration:
- `data_path`: Storage location for AST database
- `max_file_size`: Maximum database file size
- `enable_indexing`: Enable AST indexing for queries
- `cache_size`: In-memory cache size

## Setup

Dependencies:
- OmniStore (github.com/tenzoki/agen/omni/public/omnistore)
- bbolt (key-value database backend)

Storage path configuration required for production use.

Build:
```bash
go build -o bin/godast_storage ./code/agents/godast_storage
```

## Tests

Test file: `godast_storage_test.go`

Tests cover AST storage, retrieval, indexing, and query operations.

## Demo

No demo available
