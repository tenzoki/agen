# GOX Framework - Chunk Processing Examples

This directory contains examples demonstrating the chunk processing capabilities of the GOX Framework, focusing on the `chunk_writer` agent and related chunk manipulation operations.

## Overview

The chunk processing pipeline showcases how to handle chunked data efficiently, including writing, combining, transforming, and optimizing data chunks for various use cases such as large file processing, streaming data handling, and distributed processing.

## Agent Covered

### Chunk Writer (`chunk_writer`)
Handles the writing and management of data chunks with various output strategies and optimization techniques.

**Key Features:**
- Multiple output formats (JSON, binary, compressed)
- Chunk size optimization
- Parallel chunk writing
- Chunk validation and integrity checking
- Compression and decompression
- Metadata preservation
- Error recovery and retry mechanisms

**Use Cases:**
- Large file processing
- Streaming data persistence
- Distributed data storage
- Backup and archival systems
- Data pipeline checkpointing
- Memory-efficient processing

## Directory Structure

```
chunk-processing/
├── README.md                    # This file
├── run_chunk_demo.sh            # Demo execution script
├── input/                       # Sample input data
│   ├── large-files/             # Large files for chunking
│   ├── stream-data/             # Streaming data samples
│   └── chunk-configs/           # Chunk configuration files
├── config/                      # Cell configurations
│   ├── basic_chunk_writer_cell.yaml
│   ├── compressed_chunk_writer_cell.yaml
│   ├── parallel_chunk_writer_cell.yaml
│   ├── streaming_chunk_writer_cell.yaml
│   └── complete_chunk_pipeline_cell.yaml
└── schemas/                     # Chunk processing schemas
    ├── chunk-metadata-schema.json
    ├── chunk-config-schema.json
    └── chunk-writer-schema.json
```

## Quick Start

### Run All Chunk Processing Examples
```bash
./run_chunk_demo.sh
```

### Run Specific Chunk Processing Type
```bash
# Basic chunk writing
./run_chunk_demo.sh --type=basic

# Compressed chunk writing
./run_chunk_demo.sh --type=compressed

# Parallel chunk writing
./run_chunk_demo.sh --type=parallel

# Streaming chunk processing
./run_chunk_demo.sh --type=streaming
```

### Custom Input Directory
```bash
./run_chunk_demo.sh --input=/path/to/large/files
```

## Example Configurations

### Basic Chunk Writer
```yaml
cell:
  id: "processing:basic-chunk-writer"
  description: "Basic chunk writing cell"

  agents:
    - id: "chunk-writer-001"
      agent_type: "chunk_writer"
      ingress: "file:input/large-files/*"
      egress: "file:output/chunks/"
      config:
        chunk_size: "10MB"
        output_format: "json"
        naming_pattern: "chunk_{index:04d}.json"
        preserve_metadata: true
        validation:
          enable_checksums: true
          verify_integrity: true
        error_handling:
          retry_attempts: 3
          retry_delay: "1s"
```

### Compressed Chunk Writer
```yaml
cell:
  id: "processing:compressed-chunk-writer"
  description: "Compressed chunk writing cell"

  agents:
    - id: "chunk-writer-002"
      agent_type: "chunk_writer"
      ingress: "file:input/large-files/*"
      egress: "file:output/compressed-chunks/"
      config:
        chunk_size: "50MB"
        output_format: "binary"
        compression:
          algorithm: "gzip"
          level: 6
          enable: true
        naming_pattern: "chunk_{timestamp}_{index:06d}.gz"
        metadata:
          include_original_size: true
          include_compression_ratio: true
          include_checksums: true
        optimization:
          buffer_size: "1MB"
          io_concurrency: 4
```

### Parallel Chunk Writer
```yaml
cell:
  id: "processing:parallel-chunk-writer"
  description: "Parallel chunk writing cell"

  agents:
    - id: "chunk-writer-003"
      agent_type: "chunk_writer"
      ingress: "file:input/large-files/*"
      egress: "file:output/parallel-chunks/"
      config:
        chunk_size: "25MB"
        output_format: "json"
        parallel_processing:
          enabled: true
          worker_count: 8
          queue_size: 100
          load_balancing: "round_robin"
        naming_pattern: "worker_{worker_id}/chunk_{index:08d}.json"
        synchronization:
          enable_barriers: true
          checkpoint_interval: 100
        performance:
          memory_limit: "2GB"
          disk_buffer_size: "100MB"
```

### Streaming Chunk Writer
```yaml
cell:
  id: "processing:streaming-chunk-writer"
  description: "Streaming chunk writing cell"

  agents:
    - id: "chunk-writer-004"
      agent_type: "chunk_writer"
      ingress: "stream:live-data"
      egress: "file:output/stream-chunks/"
      config:
        chunk_strategy: "time_based"
        time_window: "5m"
        max_chunk_size: "100MB"
        output_format: "jsonl"
        streaming:
          buffer_size: "10MB"
          flush_interval: "30s"
          max_buffered_chunks: 10
        naming_pattern: "stream_{timestamp}.jsonl"
        rotation:
          enable: true
          max_file_size: "1GB"
          max_file_age: "1h"
        real_time_processing: true
```

### Complete Chunk Processing Pipeline
```yaml
cell:
  id: "processing:complete-chunk-pipeline"
  description: "Complete chunk processing pipeline"

  agents:
    - id: "file-splitter-001"
      agent_type: "file_splitter"
      ingress: "file:input/large-files/*"
      egress: "pub:raw-chunks"
      config:
        chunk_size: "20MB"
        overlap_size: "1MB"
        split_strategy: "content_aware"

    - id: "chunk-processor-001"
      agent_type: "chunk_processor"
      ingress: "sub:raw-chunks"
      egress: "pub:processed-chunks"
      config:
        processing_type: "transform"
        transformation:
          format_conversion: true
          data_cleaning: true
          validation: true

    - id: "chunk-writer-001"
      agent_type: "chunk_writer"
      ingress: "sub:processed-chunks"
      egress: "file:output/final-chunks/"
      config:
        chunk_size: "30MB"
        output_format: "binary"
        compression:
          algorithm: "lz4"
          enable: true
        validation:
          enable_checksums: true
          cross_chunk_validation: true

    - id: "chunk-synthesizer-001"
      agent_type: "chunk_synthesizer"
      ingress: "file:output/final-chunks/*"
      egress: "file:output/synthesized/"
      config:
        synthesis_strategy: "merge"
        output_format: "original"
        verify_completeness: true
```

## Chunk Configuration Patterns

### Size-Based Chunking
```yaml
chunk_config:
  strategy: "size_based"
  chunk_size: "50MB"
  max_chunk_size: "100MB"
  min_chunk_size: "1MB"
  size_tolerance: "5%"
  overlap_size: "512KB"
```

### Time-Based Chunking
```yaml
chunk_config:
  strategy: "time_based"
  time_window: "10m"
  max_events_per_chunk: 10000
  flush_on_idle: true
  idle_timeout: "2m"
```

### Content-Aware Chunking
```yaml
chunk_config:
  strategy: "content_aware"
  boundary_detection:
    enabled: true
    patterns: ["\n\n", "---", "EOF"]
    max_scan_size: "1MB"
  semantic_boundaries: true
  preserve_structure: true
```

### Memory-Efficient Chunking
```yaml
chunk_config:
  strategy: "memory_efficient"
  memory_limit: "512MB"
  streaming_mode: true
  buffer_size: "10MB"
  disk_spillover:
    enabled: true
    threshold: "80%"
    temp_directory: "/tmp/gox-chunks"
```

## Input Data Formats

### Large File Metadata
```json
{
  "file_id": "large-file-001",
  "file_path": "/data/large-dataset.json",
  "file_size": 1073741824,
  "content_type": "application/json",
  "encoding": "utf-8",
  "checksum": "sha256:abc123...",
  "chunk_preferences": {
    "preferred_chunk_size": "25MB",
    "compression": "gzip",
    "format": "json"
  },
  "processing_hints": {
    "parallelizable": true,
    "content_type": "structured",
    "has_headers": true
  }
}
```

### Streaming Data Configuration
```json
{
  "stream_id": "realtime-events",
  "source": "kafka://localhost:9092/events",
  "format": "json",
  "schema": {
    "timestamp": "datetime",
    "event_type": "string",
    "data": "object"
  },
  "chunk_config": {
    "strategy": "time_based",
    "window_size": "5m",
    "max_events": 5000
  },
  "output_config": {
    "format": "jsonl",
    "compression": "none",
    "naming": "events_{timestamp}.jsonl"
  }
}
```

### Chunk Processing Request
```json
{
  "request_id": "chunk-req-001",
  "operation": "write",
  "source_data": {
    "type": "file",
    "path": "/data/input.json",
    "size": 524288000
  },
  "chunk_config": {
    "size": "10MB",
    "format": "json",
    "compression": "gzip",
    "validation": true
  },
  "output_config": {
    "destination": "/output/chunks/",
    "naming_pattern": "chunk_{index:04d}.json.gz",
    "metadata_file": "chunks.manifest.json"
  }
}
```

## Output Examples

### Chunk Metadata
```json
{
  "chunk_metadata": {
    "chunk_id": "chunk-001",
    "source_file": "/data/large-dataset.json",
    "chunk_index": 1,
    "total_chunks": 42,
    "chunk_size": 10485760,
    "original_size": 10485760,
    "compressed_size": 3145728,
    "compression_ratio": 0.3,
    "checksum": "sha256:def456...",
    "created_at": "2024-09-27T10:00:00Z",
    "format": "json",
    "compression": "gzip"
  },
  "chunk_info": {
    "start_offset": 0,
    "end_offset": 10485759,
    "line_start": 1,
    "line_end": 125432,
    "content_preview": "First 100 characters of chunk content...",
    "content_type": "structured_json",
    "encoding": "utf-8"
  },
  "processing_metadata": {
    "processing_time": 1250,
    "worker_id": "worker-001",
    "validation_passed": true,
    "errors": [],
    "warnings": []
  }
}
```

### Chunk Manifest
```json
{
  "manifest": {
    "manifest_id": "manifest-001",
    "source_file": "/data/large-dataset.json",
    "total_chunks": 42,
    "total_size": 440401920,
    "total_compressed_size": 132120576,
    "created_at": "2024-09-27T10:00:00Z",
    "chunk_format": "json",
    "compression": "gzip"
  },
  "chunks": [
    {
      "chunk_id": "chunk-001",
      "file_name": "chunk_0001.json.gz",
      "size": 3145728,
      "checksum": "sha256:def456...",
      "start_offset": 0,
      "end_offset": 10485759
    },
    {
      "chunk_id": "chunk-002",
      "file_name": "chunk_0002.json.gz",
      "size": 3142144,
      "checksum": "sha256:ghi789...",
      "start_offset": 10485760,
      "end_offset": 20971519
    }
  ],
  "integrity": {
    "total_checksum": "sha256:master123...",
    "verification_method": "sequential",
    "verified_at": "2024-09-27T10:05:00Z"
  }
}
```

### Processing Statistics
```json
{
  "processing_stats": {
    "session_id": "proc-session-001",
    "start_time": "2024-09-27T10:00:00Z",
    "end_time": "2024-09-27T10:15:32Z",
    "total_duration": 932000,
    "files_processed": 5,
    "chunks_created": 210,
    "total_input_size": 2097152000,
    "total_output_size": 629145600,
    "compression_ratio": 0.3,
    "throughput": {
      "mb_per_second": 35.2,
      "chunks_per_second": 13.5
    }
  },
  "performance_metrics": {
    "cpu_usage": {
      "average": 45.2,
      "peak": 78.5
    },
    "memory_usage": {
      "peak": "1.2GB",
      "average": "756MB"
    },
    "disk_io": {
      "read_throughput": "125MB/s",
      "write_throughput": "89MB/s"
    }
  },
  "error_summary": {
    "total_errors": 0,
    "total_warnings": 3,
    "retries_performed": 1,
    "chunks_skipped": 0
  }
}
```

## Advanced Features

### Chunk Validation
```yaml
validation:
  enable_checksums: true
  checksum_algorithm: "sha256"
  cross_chunk_validation: true
  content_validation:
    schema_validation: true
    format_validation: true
    encoding_validation: true
  integrity_checks:
    size_verification: true
    boundary_verification: true
    sequence_verification: true
```

### Chunk Optimization
```yaml
optimization:
  compression:
    algorithm: "lz4"  # or "gzip", "snappy", "zstd"
    level: 6
    adaptive: true
  memory_management:
    buffer_pooling: true
    garbage_collection: "aggressive"
    memory_mapping: true
  io_optimization:
    read_ahead: true
    write_behind: true
    io_concurrency: 4
    batch_operations: true
```

### Error Recovery
```yaml
error_recovery:
  retry_policy:
    max_attempts: 5
    backoff_strategy: "exponential"
    base_delay: "1s"
    max_delay: "30s"
  corruption_handling:
    detect_corruption: true
    auto_repair: false
    quarantine_corrupted: true
  partial_failure_handling:
    continue_on_error: true
    skip_corrupted_chunks: true
    generate_error_report: true
```

## Performance Tuning

### High-Throughput Configuration
```yaml
performance:
  parallel_processing:
    enabled: true
    worker_count: 16
    queue_size: 1000
  memory_optimization:
    buffer_size: "100MB"
    streaming_mode: true
    memory_limit: "4GB"
  disk_optimization:
    use_direct_io: true
    sync_frequency: "never"
    write_cache_size: "256MB"
```

### Low-Memory Configuration
```yaml
performance:
  memory_conservation:
    streaming_only: true
    small_buffers: true
    disk_spillover: true
  chunk_size_adaptation:
    dynamic_sizing: true
    memory_pressure_aware: true
    min_chunk_size: "1MB"
    max_chunk_size: "10MB"
```

## Requirements

- GOX Framework v3+
- Built chunk processing agents:
  - `build/chunk_writer`
  - `build/file_splitter` (optional)
  - `build/chunk_processor` (optional)
  - `build/chunk_synthesizer` (optional)

## Building Required Agents

```bash
cd ../../
make build-chunk  # or individual builds:
go build -o build/chunk_writer ./agents/chunk_writer
go build -o build/file_splitter ./agents/file_splitter
go build -o build/chunk_processor ./agents/chunk_processor
go build -o build/chunk_synthesizer ./agents/chunk_synthesizer
```

## Use Case Examples

### Large File Processing
Process multi-gigabyte files by splitting them into manageable chunks for parallel processing.

### Streaming Data Persistence
Write continuous data streams to disk in time-based or size-based chunks.

### Distributed Storage
Split large datasets across multiple storage nodes with proper metadata tracking.

### Backup and Archival
Create compressed, verified chunks for efficient backup and long-term storage.

### Memory-Efficient ETL
Process large datasets without loading entire files into memory.