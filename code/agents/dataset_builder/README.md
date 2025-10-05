# Dataset Builder

Constructs structured datasets from chunk processing results with schema generation and validation.

## Intent

Synthesizes processed chunks into structured datasets for ML training, analysis, or export. Generates schemas, validates records, calculates statistics, and produces export-ready datasets with comprehensive metadata.

## Usage

Input: `SynthesisRequest` containing chunk IDs and output specifications
Output: `SynthesisResult` with structured dataset including schema and statistics

Configuration:
- `output_format`: Dataset format (default: "json")
- `include_metadata`: Include chunk metadata (default: true)
- `generate_schema`: Auto-generate dataset schema (default: true)
- `naming_scheme`: Record naming scheme (default: "chunk_XXXX")
- `max_records`: Maximum records per dataset (default: 100000)
- `enable_validation`: Validate dataset structure (default: true)

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/dataset_builder ./code/agents/dataset_builder
```

## Tests

Test file: `dataset_builder_test.go`

Tests cover schema generation, record validation, statistics calculation, and dataset export.

## Demo

No demo available
