# Adapter

Schema mapping and data transformation service for converting between different data formats.

## Intent

The adapter agent provides centralized data format conversion between pipeline agents. It acts as a transformation hub supporting schema mapping, text processing, format conversion (JSON/CSV/XML), and encoding operations.

## Usage

Input: `AdapterRequest` containing source format, target format, and data to transform
Output: `AdapterResponse` with transformed data or error information

Configuration:
- `supported_formats`: Array of supported format combinations
- `max_request_size`: Maximum size for transformation requests

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/adapter ./code/agents/adapter
```

## Tests

Test file: `adapter_test.go`

Tests cover:
- Text transformations (uppercase, lowercase, trim)
- JSON formatting (pretty-print, compact)
- Cross-format conversions (CSV ↔ JSON)
- Base64 encoding/decoding
- Error handling for unsupported formats

## Demo

No demo available
