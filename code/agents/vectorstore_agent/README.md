# Vectorstore Agent

Vector database service for storing and retrieving embeddings with similarity search.

## Intent

Manages vector storage and similarity search for RAG applications. Stores document embeddings with metadata, performs k-NN search, and supports filtering for efficient semantic retrieval. Integrates with embedding_agent for end-to-end vector search.

## Usage

Input: Store/search operations with vectors and metadata
Output: Storage confirmation or search results with similarity scores

Configuration:
- `storage_backend`: Vector storage backend (in-memory/persistent)
- `index_type`: Similarity index (flat/hnsw/ivf)
- `distance_metric`: Similarity metric (cosine/euclidean/dot)
- `data_path`: Persistent storage location
- `cache_size`: In-memory cache size

## Setup

Dependencies:
- OpenAI API key required for some operations (`OPENAI_API_KEY` environment variable)

Build:
```bash
go build -o bin/vectorstore_agent ./code/agents/vectorstore_agent
```

## Tests

No tests implemented

## Demo

No demo available
