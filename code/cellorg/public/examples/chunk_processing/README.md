# Parallel Chunk Processing Demo

This example demonstrates the complete parallel chunk processing workflow using the Gox framework with the integrated chunk tracking system and Godast storage backend.

## Overview

The demo showcases:

- **File Splitting**: Automatic splitting of files into manageable chunks
- **Batch Registration**: Efficient registration of chunk hierarchies in the graph store
- **Parallel Processing**: Concurrent processing of chunks with dependency tracking
- **Progress Monitoring**: Real-time monitoring of processing progress and metrics
- **Chunk Validation**: Integrity validation and file reassembly
- **Performance Analytics**: Comprehensive metrics and performance reporting

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   File Input    │───▶│  Chunk Tracker  │───▶│ Parallel Procs  │
│   - Split files │    │  - Graph store  │    │ - 4 workers     │
│   - Store chunks│    │  - Batch ops    │    │ - Async capable │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  File Output    │◀───│   Monitoring    │───▶│    Metrics      │
│  - Reassemble   │    │  - Progress     │    │ - Throughput    │
│  - Validate     │    │  - Alerts       │    │ - Efficiency    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Features Demonstrated

### 1. Chunk Tracking System
- Hierarchical graph structure for tracking file chunks and their relationships
- Bi-directional ordering edges for maintaining chunk sequence
- Status tracking throughout the processing lifecycle
- Dependency management for complex processing workflows

### 2. Batch Operations
- Batch registration of multiple file splits in single transactions
- Batch status updates for improved performance
- Parallel query execution for retrieving chunks across multiple files
- Transaction support for atomic operations

### 3. Parallel Processing
- Independent chunk identification for async processing
- Dependency-aware scheduling for complex relationships
- Load balancing across multiple processing agents
- Coordinated parallel execution with configurable concurrency

### 4. Monitoring and Metrics
- Real-time progress tracking with throughput calculations
- System-wide metrics including failure rates and efficiency
- Individual processor performance analytics
- Alerting based on configurable thresholds

### 5. Data Integrity
- Chunk integrity validation using cryptographic hashes
- File reassembly with verification
- Error handling and recovery mechanisms
- Comprehensive audit trail through event tracking

## Quick Start

### Prerequisites

1. **Build the Gox framework:**
   ```bash
   cd /path/to/gox
   make build
   ```

2. **Ensure required binaries exist:**
   - `build/gox` - Main Gox orchestrator
   - `build/godast_storage` - Storage agent

### Running the Demo

1. **Simple execution:**
   ```bash
   cd examples/chunk_processing
   ./run_demo.sh
   ```

2. **With custom broker address:**
   ```bash
   ./run_demo.sh localhost:9090
   ```

3. **Manual execution:**
   ```bash
   # Build demo
   go build -o parallel_demo parallel_demo.go

   # Run demo
   ./parallel_demo localhost:8080
   ```

## Demo Workflow

### Step 1: Environment Setup
- Creates temporary directories for input, output, storage, and logs
- Validates required binaries are available
- Initializes logging and monitoring infrastructure

### Step 2: Service Startup
- Starts Godast storage agent with demo configuration
- Establishes broker connections and message channels
- Initializes chunk tracking and storage systems

### Step 3: File Preparation
- Creates sample files of different types (text, JSON, Go code)
- Demonstrates various chunk splitting strategies
- Shows content-aware processing optimization

### Step 4: Chunk Processing
- Splits files into chunks using configurable sizes
- Registers chunk hierarchies in the graph store using batch operations
- Stores chunk data in the content-addressable storage system

### Step 5: Parallel Execution
- Identifies chunks that can be processed independently
- Distributes chunks across multiple parallel workers
- Coordinates processing with dependency awareness
- Demonstrates different processing types (transformation, compression, encryption)

### Step 6: Progress Monitoring
- Real-time progress reporting with configurable intervals
- Throughput calculation and efficiency metrics
- System-wide and per-file statistics
- Performance alerting based on thresholds

### Step 7: Validation and Reassembly
- Validates chunk integrity using cryptographic verification
- Demonstrates file reassembly from processed chunks
- Compares reassembled files with originals
- Reports any discrepancies or issues

## Configuration

### Chunk Processing Configuration
```yaml
chunk_processing:
  chunk_size: 1024          # Default chunk size in bytes
  chunk_method: "byte_size" # Chunking strategy
  parallel_workers: 4       # Number of parallel processors
  batch_size: 10           # Batch size for operations
  timeout: "30s"           # Processing timeout per chunk
```

### Monitoring Configuration
```yaml
monitoring:
  metrics_interval: "5s"           # Progress reporting interval
  progress_reporting: true         # Enable progress reporting
  alert_thresholds:
    max_failure_rate: 5.0         # Maximum acceptable failure rate (%)
    min_throughput: 10.0          # Minimum chunks per second
    max_processing_time: "2s"     # Maximum time per chunk
    stall_detection_time: "30s"   # Stall detection timeout
```

### Resource Limits
```yaml
resources:
  max_concurrent_chunks: 50   # Maximum chunks processed simultaneously
  memory_limit: "512MB"       # Memory limit for processing
  cpu_limit: "2.0"           # CPU limit (cores)
  storage_limit: "1GB"       # Storage limit for chunks
```

## Sample Output

```
=== Parallel Chunk Processing Demo ===
✅ Step 1: Creating sample files
✅ Step 2: Splitting files into chunks
✅ Step 3: Batch registering file splits
✅ Step 4: Starting parallel chunk processing
✅ Step 5: Monitoring processing progress

=== Processing Progress Report ===
File progress: sample_text.txt    progress: 100.0%  completed: 6   total: 6   failed: 0   throughput: 12.50 chunks/sec
File progress: sample_data.json   progress: 75.0%   completed: 3   total: 4   failed: 0   throughput: 8.33 chunks/sec
File progress: sample_code.go     progress: 88.9%   completed: 8   total: 9   failed: 0   throughput: 15.67 chunks/sec

=== Final Processing Report ===
System Metrics: total_files: 3  total_chunks: 19  completed_chunks: 17  failed_chunks: 0  overall_progress: 89.5%

File Metrics:
  sample_text.txt:    efficiency: 100.0%  avg_processing_time: 45ms    throughput: 12.50 chunks/sec
  sample_data.json:   efficiency: 75.0%   avg_processing_time: 52ms    throughput: 8.33 chunks/sec
  sample_code.go:     efficiency: 88.9%   avg_processing_time: 38ms    throughput: 15.67 chunks/sec

Processor Performance:
  chunk-processor-001: completed_chunks: 6   bytes_processed: 12,288   throughput: 256.0 bytes/sec
  chunk-processor-002: completed_chunks: 5   bytes_processed: 10,240   throughput: 204.8 bytes/sec
  chunk-processor-003: completed_chunks: 6   bytes_processed: 12,288   throughput: 256.0 bytes/sec

✅ Step 6: Generating final metrics
✅ Step 7: Validating and reassembling files
✅ All chunks validated successfully
✅ File reassembled successfully

Demo completed successfully!
```

## Performance Characteristics

### Scalability
- **Chunk Processing**: Linear scaling with number of parallel workers
- **Batch Operations**: Logarithmic improvement with batch size
- **Storage Operations**: Consistent performance with graph-based indexing

### Throughput
- **Small chunks** (< 1KB): ~50-100 chunks/second per worker
- **Medium chunks** (1-10KB): ~20-50 chunks/second per worker
- **Large chunks** (> 10KB): ~5-20 chunks/second per worker

### Memory Usage
- **Base overhead**: ~50MB for framework and storage
- **Per chunk**: ~2-4KB overhead for tracking and metadata
- **Batch operations**: ~10-20% memory savings compared to individual operations

## Customization

### Adding Custom Processors
```go
// Implement custom chunk processor
func customProcessor(chunk *chunks.ChunkInfo) ([]byte, error) {
    data, err := chunkOps.GetChunkData(chunk)
    if err != nil {
        return nil, err
    }

    // Custom processing logic
    processedData := customTransformation(data)

    return processedData, nil
}

// Use with coordinator
coordinator.ProcessFileInParallel(fileHash, customProcessor)
```

### Custom Chunking Strategies
```go
// Implement content-aware chunking
func intelligentChunking(filePath string) ([]*chunks.ChunkInfo, error) {
    // Analyze file content
    fileType := detectFileType(filePath)

    switch fileType {
    case "code":
        return chunkByFunctions(filePath)
    case "text":
        return chunkByParagraphs(filePath)
    default:
        return chunkBySize(filePath, defaultChunkSize)
    }
}
```

### Custom Metrics
```go
// Add custom metrics collection
type CustomMetrics struct {
    ProcessingLatency time.Duration
    CompressionRatio  float64
    ErrorCount        int
}

func collectCustomMetrics(chunk *chunks.ChunkInfo) *CustomMetrics {
    // Custom metrics collection logic
    return &CustomMetrics{
        ProcessingLatency: measureLatency(),
        CompressionRatio:  calculateCompression(),
        ErrorCount:        countErrors(),
    }
}
```

## Troubleshooting

### Common Issues

1. **Storage agent fails to start**
   - Check if required ports are available
   - Verify storage directory permissions
   - Check logs in temp directory

2. **Chunk processing stalls**
   - Increase processing timeout
   - Reduce batch size
   - Check for memory constraints

3. **High failure rate**
   - Check chunk integrity
   - Verify storage connectivity
   - Review error logs for patterns

### Debug Mode
```bash
# Run with debug logging
GOX_LOG_LEVEL=debug ./run_demo.sh

# Check detailed logs
tail -f /tmp/gox-chunk-demo/logs/storage.log
```

### Performance Tuning
```bash
# Increase worker count
export GOX_PARALLEL_WORKERS=8

# Adjust chunk size
export GOX_CHUNK_SIZE=2048

# Enable batch optimizations
export GOX_BATCH_SIZE=20
```

## Integration

This demo can be integrated with:

- **CI/CD pipelines** for large file processing
- **Data processing workflows** for parallel transformation
- **Content management systems** for efficient storage
- **Analytics platforms** for performance monitoring

## Next Steps

1. **Explore the codebase**: Review implementation details in `internal/chunks/`
2. **Run tests**: Execute `go test ./internal/chunks/...` for comprehensive testing
3. **Customize processing**: Implement domain-specific chunk processors
4. **Scale deployment**: Deploy on multiple nodes for distributed processing
5. **Integrate monitoring**: Connect to external monitoring systems

For more information, see the [main Gox documentation](../../README.md) and [storage integration examples](../storage_integration/).