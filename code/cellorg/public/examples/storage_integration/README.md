# Godast Storage Agent Integration

This directory contains comprehensive examples, tests, and documentation for the Godast Storage Agent integration with the Gox pipeline framework.

## Overview

The Godast Storage Agent provides unified storage capabilities including:
- **Key-Value Store**: Fast, persistent key-value operations
- **Graph Database**: Relationship modeling and traversal
- **File Storage**: Content-addressable file storage with deduplication
- **Full-text Search**: Document indexing and search capabilities

## Files Structure

```
examples/storage_integration/
├── README.md                          # This documentation
├── enhanced_file_processor.go         # Example agent using storage
├── demo_storage_usage.go              # Interactive demonstration
├── validate_storage.go                # Comprehensive validation suite
├── test_storage_agent.sh              # Test automation script
├── legacy_storage_pipeline.yaml       # Legacy pipeline pattern (deprecated)
└── input/                             # Sample input files
    ├── sample_document.md
    ├── data_report.txt
    └── config.json
```

## Quick Start

### 1. Build the Storage Agent

```bash
# Build the Godast Storage Agent
go build -o build/godast_storage ./agents/godast_storage

# Verify the build
./build/godast_storage --help
```

### 2. Run Unit Tests

```bash
# Run unit tests for the storage agent
cd agents/godast_storage
go test -v

# Run benchmarks
go test -bench=. -benchtime=5s
```

### 3. Run the Test Suite

```bash
# Execute comprehensive test suite
./examples/storage_integration/test_storage_agent.sh
```

### 4. Run Validation

```bash
# Run storage validation suite
go run examples/storage_integration/validate_storage.go
```

### 5. Run the Demo

```bash
# Interactive storage demonstration
go run examples/storage_integration/demo_storage_usage.go
```

## Usage Examples

### Basic Storage Operations

The storage agent accepts JSON requests via the message broker:

```json
{
  "operation": "kv_set",
  "key": "user:123",
  "value": {"name": "Alice", "role": "developer"},
  "request_id": "req-12345"
}
```

### KV Store Operations

```go
// Set a value
request := StorageRequest{
    Operation: "kv_set",
    Key:       "config:app",
    Value:     map[string]interface{}{"theme": "dark", "lang": "en"},
    RequestID: uuid.New().String(),
}

// Get a value
request := StorageRequest{
    Operation: "kv_get",
    Key:       "config:app",
    RequestID: uuid.New().String(),
}

// Check existence
request := StorageRequest{
    Operation: "kv_exists",
    Key:       "config:app",
    RequestID: uuid.New().String(),
}

// Delete a value
request := StorageRequest{
    Operation: "kv_delete",
    Key:       "config:app",
    RequestID: uuid.New().String(),
}
```

### File Storage Operations

```go
// Store a file
request := StorageRequest{
    Operation: "file_store",
    FileData:  []byte("file content here"),
    Metadata: map[string]interface{}{
        "filename": "document.txt",
        "content_type": "text/plain",
    },
    RequestID: uuid.New().String(),
}

// Retrieve a file
request := StorageRequest{
    Operation: "file_retrieve",
    Key:       "file_hash_here",
    RequestID: uuid.New().String(),
}
```

### Graph Operations

```go
// Create a vertex
request := StorageRequest{
    Operation: "graph_create_vertex",
    Key:       "user:alice",
    Value: map[string]interface{}{
        "name": "Alice",
        "type": "user",
    },
    RequestID: uuid.New().String(),
}

// Create an edge
request := StorageRequest{
    Operation: "graph_create_edge",
    Value: map[string]interface{}{
        "from":  "user:alice",
        "to":    "project:gox",
        "label": "contributes_to",
    },
    RequestID: uuid.New().String(),
}

// Query the graph
request := StorageRequest{
    Operation: "graph_query",
    Query:     "g.V().hasLabel('user')",
    RequestID: uuid.New().String(),
}
```

### Full-text Search Operations

```go
// Index content
request := StorageRequest{
    Operation: "fulltext_index",
    Key:       "doc:readme",
    Value:     "This is the content to be indexed and searched",
    Metadata: map[string]interface{}{
        "title": "README",
        "tags":  []string{"documentation"},
    },
    RequestID: uuid.New().String(),
}

// Search content
request := StorageRequest{
    Operation:   "fulltext_search",
    SearchTerms: "documentation search",
    RequestID:   uuid.New().String(),
}
```

## Pipeline Integration

### Pool Configuration

The storage agent is defined in `pool.yaml`:

```yaml
- agent_type: "godast-storage"
  binary: "agents/godast_storage/main.go"
  operator: "spawn"
  capabilities: ["storage", "kv-store", "graph-database", "file-storage", "full-text-search", "persistence"]
  description: "Unified storage backend using Godast with KV, Graph, File, and Search capabilities"
```

### Modern Storage Controller Pattern

**⚠️ Note**: This directory shows the legacy pipeline pattern. For the **recommended controller pattern**, see `../storage-controller-demo/`

### Legacy Cell Configuration

Example cell with storage integration (legacy pattern):

```yaml
cell:
  id: "cell:file-processing-with-storage"
  description: "File processing cell that uses storage controller service"

  # Reference to storage controller (external service)
  controllers:
    - controller_id: "controller:storage-service"
      required: true

  agents:
    - id: "enhanced-ingester-001"
      agent_type: "file-ingester"
      dependencies: []
      ingress: "file:examples/storage_integration/input/*.{txt,md,json}"
      egress: "pub:raw-files"
      config:
        storage_service_url: "http://localhost:9002/storage/v1"
        # ... other configuration
```

### Running the Cell

```bash
# Build the project
make build

# For the modern controller pattern, see:
cd ../storage-controller-demo
./build/gox controller run controller.yaml  # Start storage controller
./build/gox cell run cell.yaml              # Run cell

# Monitor storage operations
tail -f /tmp/gox-storage-demo/logs/*.log
```

## Storage Client Library

Other agents can use the storage client library to interact with the storage agent:

```go
import "github.com/tenzoki/gox/internal/storage"

// Initialize storage client
storageClient := storage.NewStorageClient(brokerClient, agentID)

// KV operations
err := storageClient.KVSet("key", value)
value, err := storageClient.KVGet("key")
exists, err := storageClient.KVExists("key")

// File operations
hash, err := storageClient.StoreFile(data, metadata)
data, err := storageClient.RetrieveFile(hash)

// Graph operations
vertexID, err := storageClient.CreateVertex("label", properties)
err = storageClient.CreateEdge(from, to, label)
results, err := storageClient.GraphQuery(query)

// Search operations
err = storageClient.IndexContent(id, content, metadata)
results, err := storageClient.SearchContent(searchTerms)
```

## Configuration Options

### Storage Agent Configuration

```yaml
config:
  data_path: "/path/to/storage"      # Storage directory
  max_file_size: 104857600           # Maximum file size (100MB)
  enable_kv: true                    # Enable KV store
  enable_graph: true                 # Enable graph database
  enable_files: true                 # Enable file storage
  enable_fulltext: true              # Enable full-text search
```

### Environment Variables

```bash
GOX_LOG_LEVEL=debug                  # Logging level
GOX_STORAGE_ENABLED=true             # Enable storage features
GOX_DEMO_MODE=true                   # Enable demo features
GOX_METRICS_ENABLED=true             # Enable metrics collection
```

## Performance Characteristics

### Benchmarks

Based on test results:

- **KV Operations**: ~10,000 ops/sec for small values
- **File Storage**: ~100 MB/sec for large files
- **Search Operations**: ~1,000 documents/sec indexing
- **Graph Operations**: ~1,000 vertex/edge ops/sec

### Resource Usage

- **Memory**: ~50MB base + data size
- **Disk**: BadgerDB storage with compression
- **CPU**: Low overhead for most operations

## Monitoring and Debugging

### Logging

The storage agent provides comprehensive logging:

```bash
# View storage agent logs
tail -f /tmp/gox-storage-demo/logs/godast-storage.log

# View all pipeline logs
tail -f /tmp/gox-storage-demo/logs/*.log
```

### Metrics

Available metrics:
- `storage_operations_total`: Total operations by type
- `storage_operation_duration`: Operation latency
- `storage_space_used`: Storage utilization
- `files_processed_total`: File processing statistics

### Health Checks

```bash
# Check storage agent health
curl http://localhost:8080/health

# Check storage statistics
curl http://localhost:8080/stats
```

## Testing

### Unit Tests

```bash
cd agents/godast_storage
go test -v                           # Run all tests
go test -run TestKVOperations        # Run specific tests
go test -bench=.                     # Run benchmarks
```

### Integration Tests

```bash
# Run the comprehensive test suite
./examples/storage_integration/test_storage_agent.sh

# Run validation suite
go run examples/storage_integration/validate_storage.go
```

### Load Testing

```bash
# Run performance validation
go test -bench=BenchmarkKVOperations -benchtime=30s
go test -bench=BenchmarkFileOperations -benchtime=10s
```

## Troubleshooting

### Common Issues

1. **Storage initialization fails**
   - Check data directory permissions
   - Verify disk space availability
   - Check BadgerDB compatibility

2. **Performance issues**
   - Monitor disk I/O
   - Check memory usage
   - Review batch operation patterns

3. **Connection issues**
   - Verify broker connectivity
   - Check message routing configuration
   - Review agent dependencies

### Debug Mode

Enable debug logging:

```bash
export GOX_LOG_LEVEL=debug
./build/godast_storage
```

### Data Recovery

BadgerDB provides data recovery tools:

```bash
# Check database integrity
badger info --dir=/tmp/gox-storage-demo

# Backup database
badger backup --dir=/tmp/gox-storage-demo --backup-file=backup.db
```

## Advanced Usage

### Custom Storage Agents

Create custom agents that extend the storage functionality:

```go
type CustomStorageAgent struct {
    GodastStorageAgent
    customFeatures map[string]interface{}
}

func (c *CustomStorageAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    // Add custom logic
    if msg.Type == "custom_operation" {
        return c.handleCustomOperation(msg, base)
    }

    // Delegate to base storage agent
    return c.GodastStorageAgent.ProcessMessage(msg, base)
}
```

### Multi-tenant Storage

Configure separate storage spaces:

```yaml
- id: "storage-tenant-a"
  agent_type: "godast-storage"
  config:
    data_path: "/storage/tenant-a"

- id: "storage-tenant-b"
  agent_type: "godast-storage"
  config:
    data_path: "/storage/tenant-b"
```

## Contributing

To contribute to the storage integration:

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Update documentation
5. Submit pull request

### Development Setup

```bash
# Install dependencies
go mod tidy

# Run tests
make test

# Build project
make build

# Run validation
./examples/storage_integration/test_storage_agent.sh
```

## License

This storage integration is part of the Gox project and follows the same license terms.

## Support

For support and questions:
- Check the troubleshooting section
- Review test outputs and logs
- Consult the Godast documentation
- Submit issues via the project repository