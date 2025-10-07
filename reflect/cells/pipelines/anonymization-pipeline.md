# Anonymization Pipeline

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


PII anonymization pipeline with NER, persistent mappings, and VFS isolation

## Intent

Provides GDPR-compliant document anonymization through multilingual Named Entity Recognition (XLM-RoBERTa), deterministic pseudonymization with SHA256-based mapping, and persistent bidirectional storage for entity mappings with project-level isolation.

## Agents

- **anonymization-store-001** (anonymization-store) - Persistent mapping storage using bbolt
  - Ingress: sub:storage-requests
  - Egress: pub:storage-responses

- **ner-agent-001** (ner-agent) - Multilingual NER using ONNX Runtime
  - Ingress: sub:ner-requests
  - Egress: pub:detected-entities

- **anonymizer-001** (anonymizer) - Pseudonymization with persistent storage
  - Ingress: sub:anonymization-requests
  - Egress: pub:anonymized-documents

## Data Flow

```
Text Input → anonymizer-001 → ner-agent-001 (entity detection)
  → anonymizer-001 (pseudonym generation) ↔ anonymization-store-001 (persistent mappings)
  → Anonymized Output
```

## Configuration

- NER Model: XLM-RoBERTa ONNX (multilingual support)
- Entity Types: PERSON, ORG, LOC, MISC
- Max sequence length: 128, confidence threshold: 0.5
- Storage: /tmp/gox-anonymization-store with 100MB max file size
- Deterministic pseudonyms using SHA256
- Project-level isolation for mappings
- Bidirectional mapping (forward and reverse lookups)
- GDPR-compliant soft delete with audit trail
- Startup timeout: 90s (for NER model loading)

## Usage

```bash
./bin/orchestrator -config=./workbench/config/anonymization_pipeline.yaml
```

Send request to "anonymization-requests" topic:
```json
{
  "text": "Angela Merkel visited Berlin.",
  "entities": [],
  "project_id": "project-a"
}
```

Requires ONNX Runtime installation and model conversion (see models/README.md).
