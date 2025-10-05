# Text Analysis Deep Processing

Deep text analysis with NLP and advanced features

## Intent

Provides comprehensive text analysis with full NLP pipeline, sentiment analysis, and extensive keyword extraction (50 keywords) to support content classification, topic modeling, and semantic understanding workflows for large documents (up to 50K lines).

## Agents

- **text-analyzer-deep-001** (text-analyzer) - Deep text analysis with all features
  - Ingress: sub:deep-text-analysis
  - Egress: pub:deep-analyzed-text

## Data Flow

```
Text Documents → text-analyzer-deep (NLP + sentiment + keywords, max 50K lines)
  → Deep Analysis Results (50 keywords)
```

## Configuration

Text Analysis:
- NLP enabled (full pipeline)
- Sentiment analysis enabled
- Keyword extraction enabled
- Max lines: 50,000
- Max keywords: 50

Orchestration:
- Startup timeout: 45s, shutdown: 20s
- Max retries: 3, retry delay: 15s
- Health check interval: 40s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/text-analysis-deep-processing.yaml
```
