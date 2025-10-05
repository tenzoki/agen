# Embedding Agent

Vector embedding generation service with multiple provider support and VFS-based caching.

## Intent

Generates vector embeddings for text/code chunks using embedding providers (OpenAI, HuggingFace, local models). Provides caching for efficiency, batch processing for API optimization, and project-isolated storage via VFS.

## Usage

Input: `EmbeddingRequest` containing texts to embed
Output: `EmbeddingResponse` with vector embeddings and cache statistics

Configuration:
- `provider`: Embedding provider ("openai"/hu ggingface"/"local", default: "openai")
- `model`: Model identifier (default: "text-embedding-3-small")
- `batch_size`: Texts per API call (default: 100)
- `cache_enabled`: Enable VFS caching (default: true)
- `timeout`: API request timeout (default: 30s)
- `dimensions`: Expected embedding dimensions (default: 1536)

## Setup

Dependencies:
- OpenAI API key required (`OPENAI_API_KEY` environment variable)
- Internet connection for API access

Build:
```bash
go build -o bin/embedding_agent ./code/agents/embedding_agent
```

## Tests

No tests implemented

## Demo

No demo available
