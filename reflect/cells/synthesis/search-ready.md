# Search-Ready Synthesis

**Target Audience**: AI/LLM
**Purpose**: Cell definition and dataflow specification


Optimized pipeline for search and discovery use cases

## Intent

Generates search-optimized outputs through comprehensive keyword/topic extraction (40 keywords, 20 topics) and extensive term indexing (20K terms, low threshold) to support full-text search engines, discovery interfaces, and recommendation systems.

## Agents

- **summary-agent** (summary-generator) - Generates search-optimized summaries
  - Ingress: channels:synthesis_request
  - Egress: channels:summary_for_search

- **indexer-agent** (search-indexer) - Builds comprehensive search indices
  - Ingress: channels:synthesis_request
  - Egress: channels:search_index_ready

## Data Flow

```
Synthesis Request → summary-agent (40 keywords, 20 topics, 500 char summary)
                 → indexer-agent (20K terms, positions, low threshold 0.01)
                   → Search-Ready Index
```

## Configuration

Summary Generation:
- Max keywords: 40 (high for search)
- Max topics: 20
- Summary length: 500 characters
- Sections disabled (simplified for search)

Search Indexing:
- Max terms: 20,000
- Min term frequency: 1 (comprehensive coverage)
- Calculate positions: true
- Index format: JSON
- Score threshold: 0.01 (low for comprehensive indexing)

Orchestration:
- Startup timeout: 30s, shutdown: 20s
- Max retries: 2
- Health check interval: 15s
- Execution order: summary-agent → indexer-agent
- Search optimization enabled

## Usage

```bash
./bin/orchestrator -config=./workbench/config/search-ready-synthesis.yaml
```
