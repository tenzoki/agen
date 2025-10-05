# JSON Analyzer

Comprehensive JSON content analysis including validation, schema generation, and structural assessment.

## Intent

Analyzes JSON content to provide validation (RFC 7159), automatic schema generation (JSON Schema draft-07), structure detection (depth, complexity), pattern recognition (API responses, config files), and format classification for intelligent JSON processing.

## Usage

Input: `ChunkProcessingRequest` with JSON content
Output: `ProcessingResult` with comprehensive JSON analysis

Configuration:
- `enable_validation`: JSON syntax validation (default: true)
- `enable_schema_generation`: Auto-generate JSON schema (default: true)
- `enable_pattern_matching`: Detect common JSON patterns (default: true)
- `max_depth`: Maximum nesting depth to analyze (default: 10)
- `max_size_mb`: Maximum JSON size in MB (default: 50)

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/json_analyzer ./code/agents/json_analyzer
```

## Tests

Test file: `json_analyzer_test.go`

Tests cover validation, schema generation, pattern detection, and error handling.

## Demo

No demo available
