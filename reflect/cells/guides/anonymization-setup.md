# Gox Anonymization Pipeline - Setup Guide

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


**Status**: Ready for Implementation
**Documentation**: See `docs/gox_anonymization_concept.md`
**Repository**: https://github.com/tenzoki/gox

---

## Overview

This guide walks you through setting up the anonymization pipeline dependencies **before** implementing the agents.

---

## Step 1: Install ONNXRuntime (Required for Go Agents)

### macOS

```bash
# Install via Homebrew
brew install onnxruntime

# Verify installation
ls -l /opt/homebrew/lib/libonnxruntime.*

# Expected output:
# /opt/homebrew/lib/libonnxruntime.1.16.3.dylib
# /opt/homebrew/lib/libonnxruntime.dylib -> libonnxruntime.1.16.3.dylib
```

### Linux (Ubuntu/Debian)

```bash
# Download ONNXRuntime
cd /tmp
wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-x64-1.16.3.tgz
tar -xzf onnxruntime-linux-x64-1.16.3.tgz

# Install to system directories
sudo cp onnxruntime-linux-x64-1.16.3/lib/* /usr/local/lib/
sudo cp -r onnxruntime-linux-x64-1.16.3/include/* /usr/local/include/

# Update library cache
sudo ldconfig

# Verify
ldconfig -p | grep onnxruntime
```

### Windows

```powershell
# Download from GitHub releases
# https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-win-x64-1.16.3.zip

# Extract to C:\onnxruntime
# Add C:\onnxruntime\lib to system PATH
```

---

## Step 2: Set Up Python Environment (For Model Conversion)

```bash
cd models/

# Create virtual environment
python3 -m venv venv

# Activate it
source venv/bin/activate  # macOS/Linux
# venv\Scripts\activate   # Windows

# Install dependencies
pip install -r requirements.txt

# This installs:
# - PyTorch
# - Transformers (HuggingFace)
# - ONNX conversion tools
# - Optimum library
```

---

## Step 3: Download and Convert Models

```bash
# Still in models/ directory with venv activated

# Option A: Full-size models (production quality)
python download_and_convert.py

# Option B: Smaller models (for development/testing)
python download_and_convert.py --small

# Option C: Skip problematic coref model (recommended for first run)
python download_and_convert.py --skip-coref
```

### What This Does

1. Downloads models from HuggingFace:
   - NER: `xlm-roberta-large-finetuned-conll03-english` (~1.4GB)
   - Embeddings: `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2` (~470MB)
   - Coref: `biu-nlp/f-coref` (optional, ~1.2GB)

2. Converts PyTorch → ONNX format

3. Saves to:
   - `models/ner/xlm-roberta-ner.onnx`
   - `models/embeddings/multilingual-minilm.onnx`
   - `models/coref/spanbert-coref.onnx` (if not skipped)

### Expected Output

```
======================================================================
  Gox Anonymization Model Converter
======================================================================

  Model size: large
  Output directory: /path/to/gox/models

======================================================================
  Converting NER Model: xlm-roberta-large-finetuned-conll03-english
======================================================================

  Downloading model from HuggingFace...
  Converting to ONNX format...
  ✓ Model saved to: models/ner/xlm-roberta-ner.onnx
  Verifying ONNX model: models/ner/xlm-roberta-ner.onnx
  ✓ Model loaded successfully
  ✓ Inputs: ['input_ids', 'attention_mask']
  ✓ Outputs: ['logits']
  ✓ NER model ready!

======================================================================
  Converting Embedding Model: sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2
======================================================================

  Downloading model from HuggingFace...
  Converting to ONNX format...
  ✓ Model saved to: models/embeddings/multilingual-minilm.onnx
  Verifying ONNX model: models/embeddings/multilingual-minilm.onnx
  ✓ Model loaded successfully
  ✓ Inputs: ['input_ids', 'attention_mask']
  ✓ Outputs: ['last_hidden_state']
  ✓ Embedding model ready!

======================================================================
  Conversion Summary
======================================================================

  NER Model:        ✓ Success
  Coref Model:      ⚠ Skipped/Failed
  Embedding Model:  ✓ Success

  ✓ Core models ready! You can start implementing agents.
  ℹ Coref model is optional - can implement later or use rule-based approach
```

### Verify Models

```bash
# Check models exist
ls -lh models/ner/*.onnx
ls -lh models/embeddings/*.onnx

# Output should show:
# models/ner/xlm-roberta-ner.onnx           (~1.4GB or ~560MB for small)
# models/embeddings/multilingual-minilm.onnx (~470MB)
```

---

## Step 4: Install Go Dependencies

```bash
# Return to gox root directory
cd ..

# Install ONNXRuntime Go bindings
go get github.com/yalue/onnxruntime_go

# Godast/omnistore is already in go.mod (existing Gox storage)
# No additional storage dependencies needed!
```

### Set CGO Environment Variables

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# macOS (Homebrew)
export CGO_CFLAGS="-I/opt/homebrew/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib -lonnxruntime"

# Linux
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime"
```

Then reload:
```bash
source ~/.bashrc  # or ~/.zshrc
```

---

## Step 5: Test Setup

### Test Python Environment

```bash
cd models/
source venv/bin/activate

python -c "import torch; import transformers; import onnx; print('✓ Python setup OK')"
```

### Test ONNX Model Loading (Python)

```bash
python -c "
import onnxruntime as ort
session = ort.InferenceSession('models/ner/xlm-roberta-ner.onnx')
print('✓ NER model loads successfully')
print('Inputs:', [i.name for i in session.get_inputs()])
print('Outputs:', [o.name for o in session.get_outputs()])
"
```

### Test Go ONNXRuntime Bindings

Create a test file `test_onnx_go.go`:

```go
package main

import (
    "fmt"
    "log"
    ort "github.com/yalue/onnxruntime_go"
)

func main() {
    // Initialize ONNXRuntime
    err := ort.InitializeEnvironment()
    if err != nil {
        log.Fatal("Failed to initialize ONNX Runtime:", err)
    }
    defer ort.DestroyEnvironment()

    fmt.Println("✓ ONNXRuntime Go binding working!")

    // Try loading a model
    session, err := ort.NewAdvancedSession("models/ner/xlm-roberta-ner.onnx",
        []string{"input_ids", "attention_mask"},
        []string{"logits"},
        nil,
    )
    if err != nil {
        log.Fatal("Failed to load model:", err)
    }
    defer session.Destroy()

    fmt.Println("✓ NER model loads in Go!")
}
```

Run it:
```bash
go run test_onnx_go.go
```

Expected output:
```
✓ ONNXRuntime Go binding working!
✓ NER model loads in Go!
```

---

## Troubleshooting

### Issue: "dyld: Library not loaded: libonnxruntime"

**macOS**:
```bash
# Add to ~/.zshrc or ~/.bashrc
export DYLD_LIBRARY_PATH=/opt/homebrew/lib:$DYLD_LIBRARY_PATH

# Reload
source ~/.zshrc
```

### Issue: "cannot find -lonnxruntime"

Check library exists:
```bash
# macOS
ls /opt/homebrew/lib/libonnxruntime.*

# Linux
ls /usr/local/lib/libonnxruntime.*
```

If missing, reinstall ONNXRuntime (Step 1).

### Issue: Python conversion fails with memory error

Use smaller models:
```bash
python download_and_convert.py --small
```

Or increase swap space (Linux).

### Issue: Go build fails with "cgo: C compiler not available"

**macOS**:
```bash
xcode-select --install
```

**Linux**:
```bash
sudo apt-get install build-essential
```

---

## Next Steps

Once setup is complete:

1. ✅ Models converted and in `models/` directory
2. ✅ ONNXRuntime library installed
3. ✅ Go bindings working
4. ✅ Test data in `test/data/multi-lang/`

You're ready to implement the agents:
- `agents/ner_agent/` - Named Entity Recognition
- `agents/coref_agent/` - Coreference Resolution
- `agents/synonym_linker/` - Entity Linking
- `agents/anonymizer/` - Pseudonymization

See `docs/gox_anonymization_concept.md` for pipeline architecture.

---

**Last Updated**: 2025-10-02
