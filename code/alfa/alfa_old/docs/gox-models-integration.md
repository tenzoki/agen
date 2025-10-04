# Gox NER & Anonymization Models Integration

**Date**: 2025-10-03
**Status**: 📋 Integration Guide
**New Features**: NER Agent, Anonymizer, Coreference Resolution, Multilingual Embeddings

---

## Overview

Gox now includes advanced NLP agents powered by ONNX models for:
1. **Named Entity Recognition (NER)** - Detect PERSON, ORG, LOC, MISC entities
2. **Anonymization** - Replace entities with placeholders or synthetic data
3. **Coreference Resolution** - Link entity mentions that refer to the same entity
4. **Multilingual Embeddings** - Semantic similarity for entity linking

---

## New Agents

### 1. NER Agent
- **Purpose**: Named Entity Recognition using XLM-RoBERTa
- **Model**: `xlm-roberta-large-finetuned-conll03-english`
- **Size**: ~1.4GB (ONNX)
- **Languages**: 100+ via XLM-RoBERTa
- **Entities**: PERSON, ORG, LOC, MISC

### 2. Anonymizer Agent
- **Purpose**: Replace detected entities with placeholders
- **Dependencies**: NER Agent output
- **Strategies**: Placeholder, synthetic, hashing
- **Use Case**: GDPR compliance, privacy protection

### 3. Anonymization Store Agent
- **Purpose**: Store entity-to-placeholder mappings
- **Features**: Reversible anonymization, consistency tracking
- **Storage**: In-memory with optional persistence

---

## Prerequisites

### 1. ONNXRuntime Native Library

The agents require ONNXRuntime C library installed on your system.

#### macOS
```bash
# Using Homebrew
brew install onnxruntime

# Verify installation
ls /opt/homebrew/lib/libonnxruntime.*
```

#### Linux (Ubuntu/Debian)
```bash
# Download ONNXRuntime
wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-x64-1.16.3.tgz
tar -xzf onnxruntime-linux-x64-1.16.3.tgz

# Copy to system directories
sudo cp onnxruntime-linux-x64-1.16.3/lib/* /usr/local/lib/
sudo cp -r onnxruntime-linux-x64-1.16.3/include/* /usr/local/include/

# Update library cache
sudo ldconfig
```

#### Windows
```powershell
# Download from:
https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-win-x64-1.16.3.zip

# Extract to C:\onnxruntime
# Add C:\onnxruntime\lib to PATH
```

### 2. HuggingFace Tokenizers (Native Library)

```bash
# macOS ARM64
cd /tmp
curl -L -o libtokenizers.darwin-arm64.tar.gz \
  https://github.com/daulet/tokenizers/releases/latest/download/libtokenizers.darwin-arm64.tar.gz
tar -xzf libtokenizers.darwin-arm64.tar.gz

# Copy to system library directory
sudo cp libtokenizers.a /opt/homebrew/lib/

# OR copy to gox lib directory
mkdir -p /path/to/gox/lib
cp libtokenizers.a /path/to/gox/lib/
```

### 3. Python Environment (for model conversion)

```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
cd /path/to/gox/models
pip install -r requirements.txt
```

### 4. Go Dependencies

```bash
# Install ONNXRuntime Go bindings
go get github.com/yalue/onnxruntime_go

# Install HuggingFace tokenizers bindings
go get github.com/daulet/tokenizers
```

### 5. Environment Variables

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
# macOS with Homebrew
export CGO_CFLAGS="-I/opt/homebrew/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib -lonnxruntime -ltokenizers"
export DYLD_LIBRARY_PATH="/opt/homebrew/lib:$DYLD_LIBRARY_PATH"

# Linux
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime -ltokenizers"
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"
```

Then reload:
```bash
source ~/.zshrc  # or ~/.bashrc
```

---

## Model Download & Setup

### Step 1: Clone Latest Gox

```bash
cd /tmp
git clone https://github.com/tenzoki/gox.git gox-latest
cd gox-latest/models
```

### Step 2: Download and Convert Models

```bash
# Activate Python virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run conversion script
python download_and_convert.py
```

This will download from HuggingFace and create:
- `models/ner/xlm-roberta-ner.onnx` (~1.4GB)
- `models/coref/spanbert-coref.onnx` (~1.2GB)
- `models/embeddings/multilingual-minilm.onnx` (~470MB)

**Total**: ~3.1GB

### Step 3: Copy Models to Alfa Workbench

```bash
# Create models directory in Alfa workbench
mkdir -p /path/to/alfa/workbench/models/ner
mkdir -p /path/to/alfa/workbench/models/coref
mkdir -p /path/to/alfa/workbench/models/embeddings

# Copy models
cp /tmp/gox-latest/models/ner/*.onnx /path/to/alfa/workbench/models/ner/
cp /tmp/gox-latest/models/coref/*.onnx /path/to/alfa/workbench/models/coref/
cp /tmp/gox-latest/models/embeddings/*.onnx /path/to/alfa/workbench/models/embeddings/

# Copy tokenizer files
cp /tmp/gox-latest/models/ner/tokenizer.json /path/to/alfa/workbench/models/ner/
```

### Step 4: Verify Models

```bash
ls -lh /path/to/alfa/workbench/models/ner/*.onnx
ls -lh /path/to/alfa/workbench/models/coref/*.onnx
ls -lh /path/to/alfa/workbench/models/embeddings/*.onnx

# Models should be 400MB-1.5GB each
```

---

## Integration TODOs

### Phase 1: Model Setup ✅ (User Action Required)

- [ ] **Install ONNXRuntime** (`brew install onnxruntime` on macOS)
- [ ] **Install Tokenizers Library** (download from GitHub releases)
- [ ] **Set Environment Variables** (CGO_CFLAGS, CGO_LDFLAGS, DYLD_LIBRARY_PATH)
- [ ] **Download Models** (run `download_and_convert.py`)
- [ ] **Copy Models to Workbench** (see Step 3 above)
- [ ] **Verify Models Exist** (`ls` commands above)

**Estimated Time**: 30-45 minutes
**Disk Space Required**: ~3.5GB (models + temporary files)
**RAM Required**: 4GB+ available for agent operations

### Phase 2: Gox Cell Configuration (In Alfa)

- [ ] **Update pool.yaml** with NER agent type
- [ ] **Update pool.yaml** with anonymizer agent type
- [ ] **Create anonymization cell** in cells.yaml
- [ ] **Configure model paths** (point to workbench models directory)
- [ ] **Test cell startup** with `--enable-gox`

### Phase 3: AI Integration (In Alfa)

- [ ] **Add anonymization actions** to tools.go
- [ ] **Update system prompt** with anonymization capabilities
- [ ] **Add model path configuration** to gox.yaml
- [ ] **Test AI can start anonymization cell**
- [ ] **Test AI can query NER agent**

### Phase 4: Testing & Documentation

- [ ] **Create anonymization demo** in demo/gox_anonymization/
- [ ] **Add integration tests** for NER agent
- [ ] **Update README.md** with anonymization features
- [ ] **Document model requirements** in docs/
- [ ] **Add troubleshooting guide** for model loading

---

## Configuration Examples

### pool.yaml (NER Agent)

```yaml
pool:
  agent_types:
    - agent_type: "ner-agent"
      binary: "agents/ner_agent/main.go"
      operator: "spawn"
      capabilities: ["named-entity-recognition", "multilingual-ner"]
      config:
        model_path: "models/ner/xlm-roberta-ner.onnx"
        tokenizer_path: "models/ner/tokenizer.json"
        max_length: 512
        batch_size: 1
```

### pool.yaml (Anonymizer Agent)

```yaml
pool:
  agent_types:
    - agent_type: "anonymizer"
      binary: "agents/anonymizer/main.go"
      operator: "spawn"
      capabilities: ["text-anonymization", "privacy-protection"]
      config:
        strategy: "placeholder"  # or "synthetic", "hash"
        preserve_structure: true
        entity_types: ["PERSON", "ORG", "LOC"]
```

### cells.yaml (Anonymization Cell)

```yaml
cell:
  id: "privacy:anonymization-pipeline"
  description: "NER + Anonymization for privacy protection"

  agents:
    # NER agent detects entities
    - id: "ner-001"
      agent_type: "ner-agent"
      ingress: "sub:text-for-analysis"
      egress: "pub:entities-detected"
      config:
        model_path: "models/ner/xlm-roberta-ner.onnx"
        tokenizer_path: "models/ner/tokenizer.json"

    # Anonymizer replaces entities
    - id: "anonymizer-001"
      agent_type: "anonymizer"
      ingress: "sub:entities-detected"
      egress: "pub:anonymized-text"
      config:
        strategy: "placeholder"
        preserve_structure: true
```

---

## Usage in Alfa

### AI Actions

Once configured, AI can use anonymization:

```json
{
  "action": "start_cell",
  "cell_id": "privacy:anonymization-pipeline",
  "project_id": "my-project"
}
```

```json
{
  "action": "query_cell",
  "project_id": "my-project",
  "query": "Anonymize this text: John Smith works at OpenAI in San Francisco.",
  "params": {
    "strategy": "placeholder",
    "preserve_structure": true
  },
  "timeout": 10
}
```

### Expected Output

```json
{
  "anonymized_text": "[PERSON_1] works at [ORG_1] in [LOC_1].",
  "entities": [
    {"text": "John Smith", "type": "PERSON", "placeholder": "PERSON_1"},
    {"text": "OpenAI", "type": "ORG", "placeholder": "ORG_1"},
    {"text": "San Francisco", "type": "LOC", "placeholder": "LOC_1"}
  ],
  "reversible": true
}
```

---

## Model Details

### 1. NER Model
- **Source**: `xlm-roberta-large-finetuned-conll03-english`
- **HuggingFace**: https://huggingface.co/xlm-roberta-large-finetuned-conll03-english
- **Size**: ~1.4GB (ONNX)
- **Entities**: PERSON, ORG, LOC, MISC
- **Languages**: Multilingual (100+ via XLM-RoBERTa)
- **Performance**: ~50-200ms per document
- **Memory**: ~1.8GB RAM

### 2. Coreference Model
- **Source**: `allenai/coref-spanbert-large`
- **HuggingFace**: https://huggingface.co/allenai/coref-spanbert-large
- **Size**: ~1.2GB (ONNX)
- **Purpose**: Link entity mentions to same entity
- **Languages**: English (primary)
- **Performance**: ~100-500ms per document
- **Memory**: ~1.5GB RAM

### 3. Embedding Model
- **Source**: `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`
- **HuggingFace**: https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2
- **Size**: ~470MB (ONNX)
- **Purpose**: Semantic similarity for entity linking
- **Languages**: 50+
- **Performance**: ~20-100ms per entity
- **Memory**: ~600MB RAM

---

## Performance Considerations

### Model Loading Times
- NER: ~2-5 seconds (first load)
- Coref: ~2-4 seconds
- Embeddings: ~1-2 seconds

### Inference Times (per document)
- NER: ~50-200ms (depends on length)
- Coref: ~100-500ms
- Embeddings: ~20-100ms per entity

### Memory Usage
- **NER agent**: ~1.8GB RAM
- **Coref agent**: ~1.5GB RAM
- **Embeddings agent**: ~600MB RAM
- **Total (all 3 agents)**: ~4GB RAM

**Recommendation**: Run agents on machine with at least **6GB RAM** available

---

## Alternative: Smaller Models

For development or limited resources:

```python
# Edit download_and_convert.py:
NER_MODEL = "Davlan/xlm-roberta-base-ner-hrl"  # ~560MB instead of 1.4GB
COREF_MODEL = "allenai/coref-spanbert-base"    # ~440MB instead of 1.2GB
EMBEDDING_MODEL = "sentence-transformers/paraphrase-multilingual-mpnet-base-v2"
```

**Total with smaller models**: ~1.4GB (instead of 3.1GB)

---

## Troubleshooting

### Model Loading Errors

**Error**: `cannot load ONNX model`
```bash
# Check model exists and has correct permissions
ls -l models/ner/xlm-roberta-ner.onnx

# Re-run conversion if corrupted
python download_and_convert.py
```

### Python Conversion Errors

**Error**: `No module named 'transformers'`
```bash
pip install -r requirements.txt
```

**Error**: `ONNX conversion failed`
```bash
pip install onnx onnxruntime optimum[exporters]
```

### Go Build Errors

**Error**: `could not determine kind of name for C.OrtGetApiBase`
```bash
# ONNXRuntime library not found
# Verify installation:
ls /opt/homebrew/lib/libonnxruntime.*  # macOS
ls /usr/local/lib/libonnxruntime.*     # Linux

# Set CGO flags (see Prerequisites section)
export CGO_LDFLAGS="-L/opt/homebrew/lib -lonnxruntime -ltokenizers"
```

**Error**: `cgo: C compiler not available`
```bash
# Install C compiler
# macOS: xcode-select --install
# Linux: sudo apt-get install build-essential
```

### Runtime Errors

**Error**: `library not loaded: libonnxruntime.dylib`
```bash
# Set DYLD_LIBRARY_PATH (macOS)
export DYLD_LIBRARY_PATH="/opt/homebrew/lib:$DYLD_LIBRARY_PATH"

# Or LD_LIBRARY_PATH (Linux)
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"
```

---

## Use Cases

### 1. GDPR Compliance
Anonymize personal data before processing:
```
Raw: "Email john.smith@company.com about the contract."
Anonymized: "Email [EMAIL_1] about the contract."
```

### 2. Privacy-Preserving RAG
Anonymize code/docs before embedding:
```
Raw: "Author: Jane Doe, Email: jane@acme.com"
Anonymized: "Author: [PERSON_1], Email: [EMAIL_1]"
```

### 3. Data Sharing
Safe data export with reversible anonymization:
```
1. Detect entities (NER)
2. Store mappings (anonymization-store)
3. Replace with placeholders (anonymizer)
4. Export anonymized data
5. Later: reverse using stored mappings
```

---

## License Notes

Models have different licenses:
- **XLM-RoBERTa**: MIT License
- **SpanBERT**: Apache 2.0
- **MiniLM**: Apache 2.0

Check HuggingFace model cards for full license details:
- https://huggingface.co/xlm-roberta-large-finetuned-conll03-english
- https://huggingface.co/allenai/coref-spanbert-large
- https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2

---

## Next Steps

1. ✅ **Read this guide** (you are here!)
2. 📋 **Follow Phase 1 TODOs** (install dependencies, download models)
3. 🔧 **Configure Gox cells** (pool.yaml, cells.yaml)
4. 🧪 **Test with demo** (run anonymization pipeline)
5. 🚀 **Enable in Alfa** (start using in projects)

---

**Status**: Ready for Integration
**Complexity**: Moderate (requires native libraries + large models)
**Benefits**: Advanced NLP capabilities (NER, anonymization, multilingual)
**Recommendation**: Start with smaller models for development/testing

---

**Last Updated**: 2025-10-03
**Gox Version**: Latest (with NER + Anonymization agents)
**Alfa Integration**: Pending Phase 1-4 completion
