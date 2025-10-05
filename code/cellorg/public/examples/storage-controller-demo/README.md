# Storage Service Demo

This directory demonstrates the **proper architecture pattern** for integrating storage services with Gox cells using **standard Gox patterns**.

## 🎯 Purpose

This demo showcases how to:
- Define storage as a **service cell** (not a pipeline component)
- Reference external service cells from processing cells
- Integrate storage via HTTP API instead of pub/sub messages
- Follow proper Gox architectural patterns using only standard configuration types

## 📁 Structure

```
storage-controller-demo/
├── README.md           # This documentation
├── cell.yaml          # Demo cell that uses storage service
├── input/             # Sample input files
└── output/            # Processing output directory
```

**Note**: The storage service is defined in the main `config/cells.yaml` as `service:storage`

## 🏗️ Architecture Pattern

### **Before (Anti-pattern):**
```yaml
# ❌ WRONG: Storage embedded in pipeline
agents:
  - id: "storage-backend-001"
    agent_type: "godast-storage"
    ingress: "sub:storage-requests"  # Storage as pipeline step
    egress: "pub:storage-responses"
```

### **After (Correct Gox Pattern):**
```yaml
# ✅ CORRECT: Storage as separate service cell
# In main config/cells.yaml:
cell:
  id: "service:storage"
  agents:
    - id: "storage-service-001"
      agent_type: "godast-storage"  # References pool.yaml
      config:
        storage_mode: "service"
        listen_port: 9002

# In demo cell.yaml:
cell:
  id: "cell:file-processing-with-storage-demo"
  agents:
    - id: "enhanced-ingester-001"
      config:
        storage_service_url: "http://localhost:9002/storage/v1"  # HTTP API
```

## 🚀 Usage

### 1. Start the Storage Service

```bash
# Start the storage service (from main config/cells.yaml)
./build/gox cell run --cell-id "service:storage" config/cells.yaml
```

The storage service will:
- Start HTTP service on port 9002
- Provide REST endpoints for storage operations
- Run independently of processing cells

### 2. Run the Demo Cell

```bash
# Run the demo cell that uses storage
./build/gox cell run examples/storage-controller-demo/cell.yaml
```

The cell will:
- Connect to the external storage service
- Access storage via HTTP API calls
- Process files while persisting data to storage

### 3. Verify Integration

```bash
# Check storage service health
curl http://localhost:9002/health

# Test KV operations
curl -X POST http://localhost:9002/storage/v1/kv \
  -H "Content-Type: application/json" \
  -d '{"key": "test", "value": "demo"}'

curl "http://localhost:9002/storage/v1/kv?key=test"
```

## 🔍 Key Differences

| Aspect | Pipeline Pattern (Wrong) | Service Pattern (Correct) |
|--------|--------------------------|----------------------------|
| **Storage Location** | Inside pipeline as agent | External service cell |
| **Communication** | Pub/Sub messages | HTTP REST API |
| **Lifecycle** | Tied to pipeline | Independent service |
| **Reusability** | Single pipeline only | Multiple cells can use |
| **Scalability** | Scales with pipeline | Scales independently |
| **Dependencies** | Artificial agent dependencies | Service availability check |

## 📋 Configuration Details

### Storage Service (in `config/cells.yaml`)

- **Type**: Standard Gox service cell using `godast-storage` agent type
- **Service Mode**: HTTP REST API on port 9002
- **Endpoints**: `/kv`, `/graph`, `/files`, `/search`, `/health`
- **Storage Backend**: Godast OmniStore
- **Features**: KV store, graph DB, file storage, full-text search

### Demo Cell (`cell.yaml`)

- **Pattern**: Standard Gox cell accessing external HTTP service
- **Communication**: Agents use HTTP client to call storage API
- **Environment**: `GOX_STORAGE_SERVICE_URL` provides service endpoint
- **Dependencies**: Service availability (not agent dependencies)

## 🏆 Benefits of This Pattern

1. **Separation of Concerns**: Data processing vs. storage service
2. **Reusability**: Multiple cells can use the same storage controller
3. **Scalability**: Storage scales independently of processing
4. **Maintainability**: Clear boundaries between components
5. **Testing**: Can test storage and processing separately
6. **Architecture Consistency**: Follows Gox design principles

## 🔄 Extending the Demo

To extend this demo:

1. **Add more storage operations** in agents (graph, files, search)
2. **Create additional cells** that use the same storage service
3. **Implement storage-aware processing** logic
4. **Add monitoring and metrics** for storage operations
5. **Test failover scenarios** when storage is unavailable

## 📚 Related Documentation

- [Gox Architecture Documentation](../../../docs/architecture.md)
- [Storage Integration Guide](../storage_integration/README.md)
- [Service Pattern Documentation](../../../docs/service-pattern.md)

---

**Note**: This demo represents the **canonical implementation** of the service pattern in Gox using standard configuration types. Use this as a template for creating other service cells (authentication, caching, etc.).