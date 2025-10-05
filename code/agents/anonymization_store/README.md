# Anonymization Store

Persistent storage service for anonymization mappings using OmniStore backend.

## Intent

Provides persistent key-value storage for anonymization mappings with bidirectional lookup (original → pseudonym, pseudonym → original). Uses bbolt-backed OmniStore for reliable storage with audit trail support.

## Usage

Input: `StorageRequest` with operations: set, get, reverse, list, delete
Output: `StorageResponse` with operation results

Configuration:
- `data_path`: Storage location (default: `/tmp/gox-anonymization-store`)
- `max_file_size`: Maximum database file size (default: 100MB)
- `enable_debug`: Debug logging (default: false)

## Setup

Dependencies:
- OmniStore (github.com/tenzoki/agen/omni/public/omnistore)
- bbolt (key-value database backend)

Storage path configuration required for production use.

Build:
```bash
go build -o bin/anonymization_store ./code/agents/anonymization_store
```

## Tests

Test file: `main_test.go`

Tests cover storage operations, mapping persistence, and reverse lookups.

## Demo

No demo available
