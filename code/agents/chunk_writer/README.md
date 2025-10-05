# Chunk Writer

Multi-format file writer for saving enriched text chunks with flexible output formatting and organization.

## Intent

Writes enriched text chunks to various output formats (JSON, text, markdown, CSV, XML) with metadata preservation, configurable file naming, and organization schemes. Provides the final output stage in text processing pipelines.

## Usage

Input: `ChunkWriteRequest` containing enriched chunks and output specifications
Output: `ChunkWriteResponse` with written file paths and statistics

Configuration:
- `default_output_format`: Output format (json/text/markdown/csv/xml, default: "json")
- `output_directory`: Base output directory (default: "/tmp/gox-chunk-writer")
- `create_directories`: Auto-create directories (default: true)
- `preserve_metadata`: Include metadata in output (default: true)
- `naming_scheme`: File naming (chunk_XXXX/hash/timestamp/custom, default: "chunk_XXXX")
- `max_file_size`: Maximum file size (default: 10MB)

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/chunk_writer ./code/agents/chunk_writer
```

## Tests

Test file: `chunk_writer_test.go`

Tests cover multiple output formats, file naming schemes, metadata preservation, and directory organization.

## Demo

No demo available
