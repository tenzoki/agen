# Anonymizer

Pseudonymization service that replaces named entities with deterministic pseudonyms using persistent mappings.

## Intent

Anonymizes text by replacing named entities (detected by NER agent) with consistent pseudonyms. Maintains bidirectional mappings via anonymization_store for data privacy compliance and de-anonymization when authorized.

## Usage

Input: `AnonymizerRequest` containing text and detected entities
Output: `AnonymizerResponse` with anonymized text and entity mappings

Configuration:
- `storage_agent_id`: ID of anonymization_store agent (default: "anon-store-001")
- `pipeline_version`: Version for audit trail (default: "v1.0")
- `enable_debug`: Debug logging (default: false)
- `timeout_seconds`: Storage operation timeout (default: 30)

## Setup

Dependencies:
- anonymization_store agent (must be running)
- Internal storage client for agent communication

Build:
```bash
go build -o bin/anonymizer ./code/agents/anonymizer
```

## Tests

No tests implemented

## Demo

No demo available
