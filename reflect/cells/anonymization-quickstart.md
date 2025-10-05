# Anonymization Pipeline - Quick Start Guide

This guide walks through setting up and using Gox's anonymization pipeline for PII detection and pseudonymization.

## Overview

The anonymization pipeline provides:
- **Multilingual NER**: Detects PERSON, ORG, LOC, MISC entities (100+ languages via XLM-RoBERTa)
- **Persistent Pseudonymization**: Deterministic SHA256-based pseudonyms stored in bbolt
- **VFS Isolation**: Per-project anonymization mappings
- **Bidirectional Lookup**: Forward (original→pseudonym) and reverse (pseudonym→original)
- **GDPR Compliance**: Soft delete with audit trail

---

## Prerequisites

### 1. Install ONNXRuntime

**macOS (Homebrew)**:
```bash
brew install onnxruntime

# Use the onnx-exports file in project root
source onnx-exports
```

The `onnx-exports` file sets up all necessary environment variables for ONNXRuntime and Tokenizers.

**Linux (Ubuntu/Debian)**:
```bash
# Install ONNXRuntime
sudo apt-get install -y libonnxruntime-dev

# Or download prebuilt binaries
wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-x64-1.16.3.tgz
tar -xzf onnxruntime-linux-x64-1.16.3.tgz
sudo cp onnxruntime-linux-x64-1.16.3/lib/* /usr/local/lib/
sudo cp -r onnxruntime-linux-x64-1.16.3/include/* /usr/local/include/
sudo ldconfig

# Use the onnx-exports file (adjust paths for Linux)
source onnx-exports
```

### 2. Download and Convert Models

```bash
cd models/

# Create Python virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Download and convert models (takes ~5 minutes)
python download_and_convert.py

# Expected output:
# models/
#   ner/
#     xlm-roberta-ner.onnx
#     config.json
#     tokenizer.json

# Deactivate venv
deactivate
```

---

## Quick Test

### 1. Build Agents

```bash
# From gox root directory
source onnx-exports

# Build all anonymization agents
make build-anonymization

# Or manually:
go build -o build/anonymization_store ./agents/anonymization_store/
go build -o build/ner_agent ./agents/ner_agent/
go build -o build/anonymizer_agent ./agents/anonymizer_agent/
```

**Note**: The NER agent requires the `onnx-exports` environment variables to be set for proper CGO linking.

### 2. Start Pipeline

**Option A: Using Gox Orchestrator**
```bash
# Start the anonymization cell
gox start config/anonymization_pipeline.yaml
```

**Option B: Standalone Testing (Development)**
```bash
# Terminal 1: Storage agent
./build/anonymization_store --agent-id=anon-store-001

# Terminal 2: NER agent
./build/ner_agent --agent-id=ner-agent-001

# Terminal 3: Anonymizer agent
./build/anonymizer --agent-id=anonymizer-001 --storage-agent-id=anon-store-001
```

### 3. Send Test Request

Create `test_request.json`:
```json
{
  "text": "Angela Merkel visited Microsoft in Berlin.",
  "project_id": "project-alpha"
}
```

Send via broker (using Gox CLI):
```bash
gox publish anonymization-requests test_request.json
```

Expected response on `anonymized-documents` topic:
```json
{
  "original_text": "Angela Merkel visited Microsoft in Berlin.",
  "anonymized_text": "PERSON_742315 visited ORG_128943 in LOC_659871.",
  "entity_count": 3,
  "mappings": {
    "Angela Merkel": "PERSON_742315",
    "Microsoft": "ORG_128943",
    "Berlin": "LOC_659871"
  },
  "processed_at": "2025-10-03T14:32:15Z"
}
```

---

## Integration Example

### Python Client

```python
import json
import redis

# Connect to Gox broker (Redis-based)
r = redis.Redis(host='localhost', port=6379)

# Subscribe to results
pubsub = r.pubsub()
pubsub.subscribe('anonymized-documents')

# Send anonymization request
request = {
    "text": "Dr. Sarah Chen works at OpenAI in San Francisco.",
    "project_id": "project-alpha"
}
r.publish('anonymization-requests', json.dumps(request))

# Wait for response
for message in pubsub.listen():
    if message['type'] == 'message':
        result = json.loads(message['data'])
        print("Anonymized:", result['anonymized_text'])
        print("Mappings:", result['mappings'])
        break
```

### Go Client

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/tenzoki/gox/internal/client"
)

func main() {
    // Create broker client
    broker := client.NewBrokerClient("localhost:6379")

    // Create anonymization request
    req := map[string]interface{}{
        "text":       "Dr. Sarah Chen works at OpenAI in San Francisco.",
        "project_id": "project-alpha",
    }

    payload, _ := json.Marshal(req)
    msg := &client.BrokerMessage{Payload: payload}

    // Send request
    broker.Publish("anonymization-requests", msg)

    // Subscribe to results
    results := broker.Subscribe("anonymized-documents")
    result := <-results

    var response map[string]interface{}
    json.Unmarshal(result.Payload, &response)

    fmt.Println("Anonymized:", response["anonymized_text"])
    fmt.Println("Mappings:", response["mappings"])
}
```

---

## Advanced Usage

### Reverse Lookup (Deanonymization)

Query the storage agent directly:
```json
{
  "operation": "reverse",
  "key": "anon:reverse:project-alpha:PERSON_742315",
  "project_id": "project-alpha",
  "request_id": "reverse-001"
}
```

Response:
```json
{
  "success": true,
  "result": {
    "original": "Angela Merkel",
    "canonical": "Angela Merkel",
    "entity_type": "PERSON"
  }
}
```

### List All Mappings for a Project

```json
{
  "operation": "list",
  "key": "anon:forward:project-alpha:",
  "project_id": "project-alpha",
  "request_id": "list-001"
}
```

### Delete Mapping (Soft Delete)

```json
{
  "operation": "delete",
  "key": "anon:forward:project-alpha:Angela Merkel",
  "project_id": "project-alpha",
  "request_id": "delete-001"
}
```

---

## Pipeline Architecture

```
┌─────────────────┐
│  Input Text     │
│  + Project ID   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   NER Agent     │ ◄─── models/ner/xlm-roberta-ner.onnx
│  (ONNXRuntime)  │
└────────┬────────┘
         │ entities: [{text, type, start, end, confidence}]
         ▼
┌─────────────────┐
│  Anonymizer     │
│    Agent        │
└────────┬────────┘
         │
         ├─── Check Storage ───► ┌──────────────────┐
         │                       │ Anonymization    │
         │   ◄─── Get/Set ────   │  Store Agent     │
         │                       │  (bbolt/godast)  │
         ▼                       └──────────────────┘
┌─────────────────┐
│ Anonymized Text │
│  + Mappings     │
└─────────────────┘
```

---

## Troubleshooting

### ONNXRuntime Not Found

**macOS**:
```bash
# Verify installation
ls /opt/homebrew/lib/libonnxruntime*

# If missing, reinstall
brew reinstall onnxruntime

# Add to shell profile (~/.zshrc or ~/.bash_profile)
export DYLD_LIBRARY_PATH="/opt/homebrew/lib:$DYLD_LIBRARY_PATH"
```

**Linux**:
```bash
# Check if library is in path
ldconfig -p | grep onnxruntime

# If missing, copy library
sudo cp /path/to/onnxruntime/lib/* /usr/local/lib/
sudo ldconfig
```

### Model Not Found Error

```bash
# Verify models exist
ls models/ner/xlm-roberta-ner.onnx

# If missing, run conversion
cd models && python download_and_convert.py
```

### NER Agent Returns Empty Entities

**Issue**: Model files not found or tokenizer library not linked.

**Solution**:
1. Verify ONNX model and tokenizer exist:
   ```bash
   ls models/ner/xlm-roberta-ner.onnx
   ls models/ner/tokenizer.json
   ```

2. Ensure you sourced `onnx-exports` before building:
   ```bash
   source onnx-exports
   go build -o build/ner_agent ./agents/ner_agent/
   ```

3. Check that `lib/libtokenizers.a` exists (see `lib/README.md` for installation)

### Storage Connection Failed

**Issue**: Storage agent not running or wrong agent ID.

**Check**:
```bash
# Verify storage agent is running
ps aux | grep anonymization_store

# Check logs for connection errors
tail -f logs/anonymizer.log
```

**Fix**: Ensure `storage_agent_id` in anonymizer config matches the storage agent's ID.

---

## Performance

**Expected throughput** (with ONNX model loaded):
- NER inference: 50-200ms per document (~128 tokens)
- Pseudonym generation: <1ms per entity
- Storage lookup: <5ms per operation
- End-to-end latency: ~100-300ms per document

**Memory usage**:
- NER model: ~1.8GB (loaded in RAM)
- Storage agent: ~50MB (bbolt database)
- Anonymizer agent: ~20MB

**Optimization tips**:
- Batch multiple documents when possible
- Use persistent connections to storage agent
- Cache entity mappings in anonymizer (future enhancement)

---

## Examples

Check out the example in `examples/anonymization_pipeline/`:

```bash
cd examples/anonymization_pipeline
./run_example.sh
```

This demonstrates:
- Multilingual NER (English, German, French, Spanish)
- Medical records anonymization
- Entity extraction with confidence scores
- Pseudonymization mapping

## Next Steps

1. **✅ NER Tokenizer**: Completed with HuggingFace tokenizers
2. **Add Coreference Resolution**: Link entity mentions (e.g., "she" → "Angela Merkel")
3. **Add Synonym Linking**: Normalize variations (e.g., "Microsoft Corp" → "Microsoft")
4. **Integrate with Alfa**: Use anonymization in document indexing pipeline

See `docs/anonymization-implementation-plan.md` for full roadmap.

---

## Testing

Run integration tests:
```bash
go test ./test/integration/anonymization_pipeline_test.go -v
```

Run benchmarks:
```bash
go test ./test/integration/anonymization_pipeline_test.go -bench=. -benchmem
```

---

## Support

- **Documentation**: `docs/gox_anonymization_concept.md`
- **Setup Guide**: `docs/anonymization-setup.md`
- **Implementation Plan**: `docs/anonymization-implementation-plan.md`
- **Model Setup**: `models/README.md`
- **NER Agent Status**: `agents/ner_agent/README.md`
