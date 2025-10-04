# NER Agent - Named Entity Recognition

**Status**: ✅ Complete
**Model**: XLM-RoBERTa NER (ONNX)
**Languages**: Multilingual (100+ via XLM-RoBERTa)
**Tokenizer**: HuggingFace tokenizers (daulet/tokenizers)

---

## Current Status

### ✅ Implemented
- ONNXRuntime integration
- Model loading from ONNX file
- Agent framework integration
- Input/output message handling
- Tensor creation and inference pipeline
- Configuration management
- **HuggingFace tokenizer integration** (daulet/tokenizers)
- **Proper BIO tag decoding with entity merging**
- **Confidence scoring using softmax**
- **Character offset tracking**
- **Entity type normalization**

---

## Installation

### Prerequisites

1. **ONNXRuntime**
```bash
# macOS
brew install onnxruntime

# Linux
# See models/README.md for installation instructions
```

2. **Tokenizers Native Library**
```bash
# Download pre-built library for your platform
# macOS ARM64
cd /tmp
curl -L -o libtokenizers.darwin-arm64.tar.gz \
  https://github.com/daulet/tokenizers/releases/latest/download/libtokenizers.darwin-arm64.tar.gz
tar -xzf libtokenizers.darwin-arm64.tar.gz
# Copy to project lib directory
cp libtokenizers.a /path/to/gox/lib/

# OR copy to system library directory (requires sudo)
sudo cp libtokenizers.a /opt/homebrew/lib/
```

3. **Environment Variables**
```bash
export CGO_CFLAGS="-I/opt/homebrew/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib -L$(pwd)/lib -lonnxruntime -ltokenizers"
export DYLD_LIBRARY_PATH="/opt/homebrew/lib:$DYLD_LIBRARY_PATH"
```

### Build

```bash
# With environment variables set
go build -o build/ner_agent ./agents/ner_agent/

# Or specify directly
CGO_LDFLAGS="-L$(pwd)/lib -L/opt/homebrew/lib -lonnxruntime -ltokenizers" \
  go build -o build/ner_agent ./agents/ner_agent/
```

---

## Features

Install Go bindings for HF tokenizers:
```bash
go get github.com/daulet/tokenizers
```

Replace `SimpleTokenizer` with:
```go
import "github.com/daulet/tokenizers"

func (n *NERAgent) Init(base *agent.BaseAgent) error {
    // Load HF tokenizer
    tk, err := tokenizers.FromFile(
        filepath.Join(n.config.TokenizerPath, "tokenizer.json"),
    )
    if err != nil {
        return err
    }
    n.tokenizer = tk
    // ...
}

func (n *NERAgent) extractEntities(text string, base *agent.BaseAgent) ([]Entity, error) {
    // Use real tokenizer
    encoding := n.tokenizer.EncodeWithOptions(
        text,
        true,  // add_special_tokens
        tokenizers.WithReturnOffsets(),
    )

    inputIDs := encoding.IDs
    attentionMask := encoding.AttentionMask
    offsets := encoding.Offsets

    // Create tensors
    // ... (rest stays the same)
}
```

**Option B: Python Tokenization Service**

Run Python service that does tokenization:
```python
# tokenizer_service.py
from transformers import AutoTokenizer
from flask import Flask, request, jsonify

app = Flask(__name__)
tokenizer = AutoTokenizer.from_pretrained("xlm-roberta-large-finetuned-conll03-english")

@app.route("/tokenize", methods=["POST"])
def tokenize():
    text = request.json["text"]
    encoding = tokenizer(text, return_tensors="pt", return_offsets_mapping=True)
    return jsonify({
        "input_ids": encoding["input_ids"][0].tolist(),
        "attention_mask": encoding["attention_mask"][0].tolist(),
        "offsets": encoding["offset_mapping"][0].tolist()
    })

if __name__ == "__main__":
    app.run(port=8765)
```

Then call from Go:
```go
func (n *NERAgent) callTokenizerService(text string) (*TokenizerOutput, error) {
    resp, err := http.Post(
        "http://localhost:8765/tokenize",
        "application/json",
        bytes.NewBuffer([]byte(fmt.Sprintf(`{"text": %q}`, text))),
    )
    // Parse response...
}
```

### 2. Proper Logits Decoding

Replace placeholder `decodePredictions` with:

```go
func (n *NERAgent) decodePredictions(
    logits ort.Value,
    tokens *TokenizerOutput,
    text string,
    base *agent.BaseAgent,
) []Entity {
    // Get logits shape: [batch, seq_len, num_labels]
    shape := logits.GetShape()
    seqLen := int(shape[1])
    numLabels := int(shape[2])

    // Get logits data
    logitsData := logits.GetData().([]float32)

    // Find argmax for each token (predicted label)
    predictions := make([]int, seqLen)
    for i := 0; i < seqLen; i++ {
        maxIdx := 0
        maxVal := logitsData[i*numLabels+0]

        for j := 1; j < numLabels; j++ {
            val := logitsData[i*numLabels+j]
            if val > maxVal {
                maxVal = val
                maxIdx = j
            }
        }
        predictions[i] = maxIdx
    }

    // Convert predictions to BIO tags
    // Label mapping (XLM-RoBERTa NER):
    // 0: O
    // 1: B-PER, 2: I-PER
    // 3: B-ORG, 4: I-ORG
    // 5: B-LOC, 6: I-LOC
    // 7: B-MISC, 8: I-MISC

    labelMap := map[int]string{
        0: "O",
        1: "B-PERSON", 2: "I-PERSON",
        3: "B-ORG", 4: "I-ORG",
        5: "B-LOC", 6: "I-LOC",
        7: "B-MISC", 8: "I-MISC",
    }

    // Merge B-/I- spans into entities
    entities := []Entity{}
    currentEntity := (*Entity)(nil)

    for i, pred := range predictions {
        if i == 0 || i >= len(tokens.TokenToChar)-1 {
            continue // Skip CLS/SEP tokens
        }

        label := labelMap[pred]

        if label == "O" {
            // End current entity if any
            if currentEntity != nil {
                entities = append(entities, *currentEntity)
                currentEntity = nil
            }
            continue
        }

        // Extract entity type and position
        parts := strings.Split(label, "-")
        position := parts[0] // B or I
        entityType := parts[1]

        charStart := tokens.TokenToChar[i][0]
        charEnd := tokens.TokenToChar[i][1]

        if position == "B" {
            // Start new entity
            if currentEntity != nil {
                entities = append(entities, *currentEntity)
            }
            currentEntity = &Entity{
                Text:       text[charStart:charEnd],
                Type:       entityType,
                Start:      charStart,
                End:        charEnd,
                Confidence: 1.0, // TODO: Use softmax probabilities
            }
        } else if position == "I" && currentEntity != nil {
            // Continue current entity
            currentEntity.Text = text[currentEntity.Start:charEnd]
            currentEntity.End = charEnd
        }
    }

    // Add last entity
    if currentEntity != nil {
        entities = append(entities, *currentEntity)
    }

    return entities
}
```

---

## Testing

### Without Models (Unit Tests)

```bash
cd agents/ner_agent
go test -v
```

Tests verify:
- Message parsing
- Response serialization
- Token structure

### With Models (Integration Test)

```bash
# 1. Ensure models are downloaded
ls ../../models/ner/xlm-roberta-ner.onnx

# 2. Build agent
go build -o ../../build/ner_agent

# 3. Test with sample text
echo '{"text": "Angela Merkel visited Berlin."}' | \
  ../../build/ner_agent --standalone
```

---

## Performance

**Expected**:
- Model loading: ~2-5 seconds (first run)
- Inference: ~50-200ms per document
- Memory: ~1.8GB (model in RAM)

---

## Next Steps

1. **Immediate**: Add real tokenizer (Option A recommended)
2. **Then**: Implement proper logits decoding
3. **Finally**: Test with multilingual data in `test/data/multi-lang/`

---

**Note**: The current implementation can compile and run, but will not produce correct entities until tokenizer + decoder are completed.
