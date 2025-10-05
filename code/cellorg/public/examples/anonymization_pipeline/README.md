# Anonymization Pipeline Example

This example demonstrates the complete anonymization pipeline with NER (Named Entity Recognition), anonymization, and storage.

## Architecture

```
Input Text
    ↓
NER Agent (extracts entities)
    ↓ entities: [{text, type, start, end, confidence}]
Anonymizer Agent (replaces entities with pseudonyms)
    ↓ anonymized text + mapping
Storage Agent (stores mapping)
    ↓ confirmation
Output (anonymized text)
```

## Agents Involved

1. **NER Agent** (`agents/ner_agent`)
   - Detects named entities using XLM-RoBERTa model
   - Supports 100+ languages
   - Returns: PERSON, ORG, LOC, MISC entities

2. **Anonymizer Agent** (`agents/anonymizer_agent`)
   - Replaces entities with pseudonyms
   - Consistent mapping (same entity → same pseudonym)
   - Reversible (can de-anonymize with mapping)

3. **Storage Agent** (`agents/anonymization_store`)
   - Stores anonymization mappings using godast/omnistore
   - Key-value storage with bbolt backend
   - Project-scoped isolation

## Quick Start

### 1. Build All Agents

```bash
# From project root
source onnx-exports

# Build NER agent
go build -o build/ner_agent ./agents/ner_agent/

# Build Anonymizer agent
go build -o build/anonymizer_agent ./agents/anonymizer_agent/

# Build Storage agent
go build -o build/anonymization_store ./agents/anonymization_store/
```

### 2. Run the Example

```bash
cd examples/anonymization_pipeline
./run_example.sh
```

This will:
1. Process sample documents through the pipeline
2. Extract named entities
3. Anonymize the text
4. Store the mapping
5. Show before/after comparison

## Example Input/Output

### Input (English)
```
Angela Merkel visited Microsoft headquarters in Berlin to discuss
digital transformation with CEO Satya Nadella.
```

### NER Output
```json
{
  "entities": [
    {"text": "Angela Merkel", "type": "PERSON", "start": 0, "end": 13, "confidence": 0.95},
    {"text": "Microsoft", "type": "ORG", "start": 22, "end": 31, "confidence": 0.92},
    {"text": "Berlin", "type": "LOC", "start": 49, "end": 55, "confidence": 0.88},
    {"text": "Satya Nadella", "type": "PERSON", "start": 97, "end": 110, "confidence": 0.94}
  ]
}
```

### Anonymized Output
```
PERSON-001 visited ORG-001 headquarters in LOC-001 to discuss
digital transformation with CEO PERSON-002.
```

### Mapping Stored
```json
{
  "project_id": "example-001",
  "mappings": {
    "Angela Merkel": "PERSON-001",
    "Microsoft": "ORG-001",
    "Berlin": "LOC-001",
    "Satya Nadella": "PERSON-002"
  }
}
```

## Multilingual Support

The pipeline works with 100+ languages:

### German
```
Input:  "Die Bundeskanzlerin traf sich mit dem Vorstand von Siemens in München."
Output: "Die PERSON-001 traf sich mit dem Vorstand von ORG-001 in LOC-001."
```

### French
```
Input:  "Emmanuel Macron a rencontré les dirigeants de Renault à Paris."
Output: "PERSON-001 a rencontré les dirigeants de ORG-001 à LOC-001."
```

### Spanish
```
Input:  "Pedro Sánchez visitó la sede de Telefónica en Madrid."
Output: "PERSON-001 visitó la sede de ORG-001 en LOC-001."
```

## Configuration

### NER Agent Config (`config/pool.yaml`)
```yaml
agents:
  - id: ner-agent-001
    type: ner-agent
    config:
      model_path: models/ner/xlm-roberta-ner.onnx
      tokenizer_path: models/ner/
      max_seq_length: 128
      confidence_threshold: 0.5
```

### Anonymizer Agent Config
```yaml
agents:
  - id: anonymizer-001
    type: anonymizer-agent
    config:
      strategy: pseudonymize
      consistency: project-scoped
```

### Storage Agent Config
```yaml
agents:
  - id: anon-store-001
    type: anonymization-store
    config:
      storage_path: data/anonymization_mappings.db
```

## Use Cases

1. **Medical Records Anonymization**
   - Remove patient names, doctor names, hospital names
   - Keep text structure for analysis
   - Store mapping for authorized de-anonymization

2. **Legal Document Processing**
   - Anonymize party names, law firms, locations
   - Maintain document coherence
   - Enable redaction workflows

3. **Research Data Preparation**
   - Anonymize datasets for privacy compliance (GDPR, HIPAA)
   - Consistent pseudonymization across documents
   - Reversible for authorized researchers

4. **Customer Support Analytics**
   - Anonymize customer names, company names
   - Analyze support interactions
   - Protect PII while enabling insights

## Testing

Run tests for each agent:

```bash
# NER Agent tests
source onnx-exports
cd agents/ner_agent
go test -v

# Anonymizer Agent tests
cd ../anonymizer_agent
go test -v

# Storage Agent tests
cd ../anonymization_store
go test -v
```

## Performance

**Expected throughput** (single agent instance):
- NER: 5-20 documents/second (~128 tokens each)
- Anonymizer: 100+ documents/second
- Storage: 1000+ operations/second

**Scalability**:
- Run multiple agent instances for higher throughput
- Orchestrator handles load balancing
- Horizontal scaling via agent pool

## Troubleshooting

### NER Agent Issues

**Error: "library 'tokenizers' not found"**
```bash
# Ensure you sourced onnx-exports
source onnx-exports
go build -o build/ner_agent ./agents/ner_agent/
```

**Error: "ONNX model not found"**
```bash
# Convert models
cd models
python download_and_convert.py
```

### Anonymizer Agent Issues

**Inconsistent pseudonyms**
- Check that `project_id` is consistent across requests
- Verify storage agent is running and accessible

### Storage Agent Issues

**Database lock errors**
- Only one process can write to bbolt database at a time
- Use agent pool for concurrency

## Next Steps

1. **Customize Pipeline**: Add agents for your specific use case
2. **Integrate with Orchestrator**: Deploy via `config/anonymization_pipeline.yaml`
3. **Scale Up**: Run multiple instances of each agent
4. **Monitor**: Track metrics (throughput, latency, errors)

## Resources

- [NER Agent Implementation](../../agents/ner_agent/README.md)
- [Anonymizer Agent](../../agents/anonymizer_agent/README.md)
- [Storage Agent](../../agents/anonymization_store/README.md)
- [Gox Orchestrator](../../docs/gox-overview.md)
- [Pipeline Configuration](../../config/anonymization_pipeline.yaml)
