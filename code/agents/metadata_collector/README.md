# Metadata Collector

Aggregates and enriches metadata from multiple processing stages into unified metadata records.

## Intent

Collects metadata from various pipeline agents (extractors, analyzers, enrichers) and creates comprehensive metadata records. Enables metadata-driven search, filtering, and analysis across processed documents.

## Usage

Input: Processing results with metadata from multiple agents
Output: Unified metadata records with aggregated information

Configuration:
- `metadata_fields`: Fields to collect and aggregate
- `enable_deduplication`: Remove duplicate metadata entries
- `enable_validation`: Validate metadata integrity
- `storage_format`: Metadata storage format

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/metadata_collector ./code/agents/metadata_collector
```

## Tests

Test file: `metadata_collector_test.go`

Tests cover metadata aggregation, deduplication, and validation.

## Demo

No demo available
