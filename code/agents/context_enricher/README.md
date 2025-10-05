# Context Enricher

Enriches text chunks with contextual metadata including positional, semantic, structural, and relational information.

## Intent

Adds multi-dimensional context to text chunks for improved RAG, search, and analysis. Provides document position tracking, semantic classification, structural hierarchy, and inter-chunk relationships to enable context-aware processing.

## Usage

Input: `ContextEnrichmentRequest` containing chunks and document metadata
Output: `ContextEnrichmentResponse` with enriched chunks containing full contextual information

Configuration:
- `enable_positional_context`: Add document position metadata (default: true)
- `enable_semantic_context`: Add semantic classification (default: true)
- `enable_structural_context`: Add structural hierarchy (default: true)
- `enable_relational_context`: Add inter-chunk relationships (default: false, more expensive)
- `context_depth`: Analysis depth level (default: 3)

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/context_enricher ./code/agents/context_enricher
```

## Tests

No tests implemented

## Demo

No demo available
