# Search Indexer

Full-text search indexing service for enabling fast document retrieval.

## Intent

Creates and maintains search indexes from processed documents enabling fast full-text search, filtering, and ranking. Supports inverted indexes, keyword extraction, and relevance scoring for efficient document discovery.

## Usage

Input: Documents and chunks to index
Output: Search index with retrieval capabilities

Configuration:
- `index_path`: Search index storage location
- `analyzer`: Text analyzer (standard/keyword/n-gram)
- `enable_stemming`: Enable word stemming
- `enable_stopwords`: Filter common stopwords
- `boost_fields`: Field-specific relevance boosting

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/search_indexer ./code/agents/search_indexer
```

## Tests

Test file: `search_indexer_test.go`

Tests cover indexing, search queries, and ranking algorithms.

## Demo

No demo available
