# Document Summary Synthesis

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Document summarization pipeline using Framework-compliant agents

## Intent

Generates comprehensive document summaries with keyword extraction, topic identification, and section-aware analysis to provide concise overviews for search interfaces, knowledge bases, and document management systems.

## Agents

- **summary-agent** (summary-generator) - Generates structured summaries
  - Ingress: channels:synthesis_request
  - Egress: channels:synthesis_result

## Data Flow

```
Synthesis Request → summary-agent (keyword extraction, topic analysis, summarization)
  → Summary Result (JSON, 750 chars, with sections)
```

## Configuration

Summary Generation:
- Max keywords: 25
- Max topics: 12
- Summary length: 750 characters
- Output format: JSON
- Section analysis enabled

Orchestration:
- Startup timeout: 30s, shutdown: 15s
- Max retries: 2
- Health check interval: 15s
- Execution order: summary-agent

## Usage

```bash
./bin/orchestrator -config=./workbench/config/document-summary-synthesis.yaml
```
