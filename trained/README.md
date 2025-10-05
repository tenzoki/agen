# Gox Anonymization Models

This directory contains ONNX models for the anonymization pipeline. Models are **not checked into git** due to size.

---

## Directory Structure

```
models/
├── README.md                    # This file
├── download_and_convert.py      # Model conversion script
├── requirements.txt             # Python dependencies
├── ner/                         # Named Entity Recognition models
│   └── xlm-roberta-ner.onnx
├── coref/                       # Coreference Resolution models
│   └── spanbert-coref.onnx
└── embeddings/                  # Embedding models
    └── multilingual-minilm.onnx
```

---

## Prerequisites

### 1. Python Environment (for model conversion)

```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
```

### 2. ONNXRuntime Native Library (for Go agents)

The Go agents require ONNXRuntime C library installed on your system.

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
# Download from: https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-win-x64-1.16.3.zip
# Extract to C:\onnxruntime
# Add C:\onnxruntime\lib to PATH
```

### 3. Go Dependencies

```bash
# Install ONNXRuntime Go bindings
go get github.com/yalue/onnxruntime_go

# Set CGO flags and library paths
# Add these to your ~/.zshrc or ~/.bashrc for persistence

# macOS with Homebrew
export CGO_CFLAGS="-I/opt/homebrew/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib -lonnxruntime"
export DYLD_LIBRARY_PATH="/opt/homebrew/lib:$DYLD_LIBRARY_PATH"

# Linux
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime"
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"

# Reload shell config
source ~/.zshrc  # or ~/.bashrc
```

---

## Model Download & Conversion

### Step 1: Download and Convert Models

```bash
# Activate Python virtual environment
source venv/bin/activate

# Run conversion script (downloads from HuggingFace and converts to ONNX)
python download_and_convert.py

# This will create:
# - models/ner/xlm-roberta-ner.onnx
# - models/coref/spanbert-coref.onnx
# - models/embeddings/multilingual-minilm.onnx
```

The script will:
1. Download PyTorch models from HuggingFace
2. Convert to ONNX format
3. Optimize for inference
4. Validate output shapes
5. Save to respective directories

### Step 2: Verify Models

```bash
# Check that models exist
ls -lh models/ner/*.onnx
ls -lh models/coref/*.onnx
ls -lh models/embeddings/*.onnx

# Models should be 400MB-1.5GB each
```

---

## Model Details

### 1. NER Model
- **Source**: `xlm-roberta-large-finetuned-conll03-english`
- **HuggingFace**: https://huggingface.co/xlm-roberta-large-finetuned-conll03-english
- **Size**: ~1.4GB (ONNX)
- **Entities**: PERSON, ORG, LOC, MISC
- **Languages**: Multilingual (100+ languages via XLM-RoBERTa)

### 2. Coreference Model
- **Source**: `allenai/coref-spanbert-large`
- **HuggingFace**: https://huggingface.co/allenai/coref-spanbert-large
- **Size**: ~1.2GB (ONNX)
- **Purpose**: Cluster entity mentions that refer to the same entity
- **Languages**: English (primary), decent multilingual performance

### 3. Embedding Model
- **Source**: `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`
- **HuggingFace**: https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2
- **Size**: ~470MB (ONNX)
- **Purpose**: Semantic similarity for entity linking
- **Languages**: 50+ languages

---

## Alternative: Smaller Models (for development)

If disk space or RAM is limited, use these smaller models:

```python
# Edit download_and_convert.py and change models to:

NER_MODEL = "Davlan/xlm-roberta-base-ner-hrl"  # ~560MB instead of 1.4GB
COREF_MODEL = "allenai/coref-spanbert-base"    # ~440MB instead of 1.2GB
EMBEDDING_MODEL = "sentence-transformers/paraphrase-multilingual-mpnet-base-v2"  # ~420MB
```

---

## Troubleshooting

### Python Conversion Errors

**Error**: `No module named 'transformers'`
```bash
pip install -r requirements.txt
```

**Error**: `ONNX conversion failed`
```bash
# Install ONNX tools
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
```

**Error**: `cgo: C compiler not available`
```bash
# Install C compiler
# macOS: xcode-select --install
# Linux: sudo apt-get install build-essential
```

### Runtime Errors

**Error**: `cannot load ONNX model`
```bash
# Check model exists and has correct permissions
ls -l models/ner/xlm-roberta-ner.onnx

# Re-run conversion if corrupted
python download_and_convert.py
```

---

## Testing Models

```bash
# Test NER agent
GOX_AGENT_ID=ner-test-001 ./build/ner_agent

# Send test message (from another terminal)
echo '{"text": "Angela Merkel visited Berlin."}' | nc localhost 9001
```

---

## Performance Notes

### Model Loading Times
- NER: ~2-5 seconds (first load)
- Coref: ~2-4 seconds
- Embeddings: ~1-2 seconds

### Inference Times (per document)
- NER: ~50-200ms (depends on text length)
- Coref: ~100-500ms
- Embeddings: ~20-100ms per entity

### Memory Usage
- NER agent: ~1.8GB RAM
- Coref agent: ~1.5GB RAM
- Embeddings agent: ~600MB RAM

**Recommendation**: Run agents on machine with at least 4GB available RAM

---

## Model Updates

To update models to newer versions:

1. Edit `download_and_convert.py` with new model IDs
2. Run conversion script
3. Test with `go test ./agents/ner_agent/` etc.
4. Update version in `pool.yaml` if needed

---

## License Notes

Models have different licenses:
- XLM-RoBERTa: MIT License
- SpanBERT: Apache 2.0
- MiniLM: Apache 2.0

Check HuggingFace model cards for full license details.

---

**Last Updated**: 2025-10-02
