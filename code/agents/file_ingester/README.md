# File Ingester

File loading and ingestion service for processing documents into the pipeline.

## Intent

Reads files from various sources (filesystem, cloud storage, URLs), performs initial validation, and prepares content for downstream processing. Acts as the entry point for document processing pipelines.

## Usage

Input: File paths or URIs to ingest
Output: File content with metadata for processing pipeline

Configuration:
- `supported_formats`: List of supported file extensions
- `max_file_size`: Maximum file size limit
- `enable_validation`: Validate file integrity
- `temp_directory`: Temporary storage location

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/file_ingester ./code/agents/file_ingester
```

## Tests

Test file: `file_ingester_test.go`

Tests cover file reading, format validation, and metadata extraction.

## Demo

No demo available
