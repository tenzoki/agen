# Fast Analysis Processing

Fast content analysis with reduced features for high-throughput scenarios

## Intent

Provides optimized high-speed content analysis for text, JSON, and binary data by disabling expensive features (NLP, schema generation, entropy analysis) to achieve minimal latency in high-throughput production environments.

## Agents

- **text-analyzer-fast-001** (text-analyzer) - Fast text analysis (keywords only)
  - Ingress: sub:fast-analysis-text
  - Egress: pub:fast-analyzed-text

- **json-analyzer-fast-001** (json-analyzer) - Fast JSON validation
  - Ingress: sub:fast-analysis-json
  - Egress: pub:fast-analyzed-json

- **binary-analyzer-fast-001** (binary-analyzer) - Fast binary analysis
  - Ingress: sub:fast-analysis-binary
  - Egress: pub:fast-analyzed-binary

## Data Flow

```
Input Files → Router (by type)
  ├→ text-analyzer (keywords only, max 1K lines)
  ├→ json-analyzer (validation + minification, max depth 10)
  └→ binary-analyzer (hashing + magic bytes, max 1MB)
    → Fast Analysis Results
```

## Configuration

Text: NLP and sentiment disabled, keywords only, max 1,000 lines, 10 keywords
JSON: Validation only, schema/key analysis disabled, minification enabled, depth 10
Binary: Hashing + magic bytes only, no entropy/structural/compression, max 1MB

Orchestration:
- Startup timeout: 15s, shutdown: 10s
- Max retries: 1, retry delay: 5s
- Health check interval: 20s
- Debug disabled for performance

## Usage

```bash
./bin/orchestrator -config=./workbench/config/fast-analysis-processing.yaml
```
