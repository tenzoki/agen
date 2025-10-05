# Knowledge Backend RAG

RAG pipeline providing semantic code search and context retrieval for Alfa AI assistant

## Intent

Provides Retrieval-Augmented Generation backend for AI assistants through semantic code search using OpenAI embeddings, vector similarity search, and integrated storage to deliver contextually relevant code snippets with surrounding lines for enhanced AI responses.

## Agents

- **rag-agent-001** (rag-agent) - RAG orchestrator coordinating embedding, search, and context
  - Ingress: sub:{project_id}:rag-queries
  - Egress: pub:{project_id}:rag-results

- **embedding-agent-001** (embedding-agent) - OpenAI embedding generator
  - Ingress: sub:embedding-requests
  - Egress: pub:embeddings

- **vectorstore-agent-001** (vectorstore-agent) - Vector similarity search
  - Ingress: sub:vector-operations
  - Egress: pub:vector-results

- **godast-storage-001** (godast-storage) - Content storage and retrieval
  - Ingress: sub:storage-operations
  - Egress: pub:storage-results

## Data Flow

```
RAG Query → rag-agent → embedding-agent (OpenAI text-embedding-3-small)
  → vectorstore-agent (similarity search, top 5)
  → godast-storage (content retrieval)
  → rag-agent (context assembly + reranking + surrounding lines)
  → RAG Results (max 4000 tokens)
```

## Configuration

RAG Agent:
- Top-k: 5 results
- Reranking enabled
- Max context tokens: 4000
- Include surrounding lines: 3
- Score threshold: 0.5

Embedding Agent:
- Provider: OpenAI
- Model: text-embedding-3-small
- Dimensions: 1536
- Batch size: 100
- Cache enabled, 30s timeout

Vector Store:
- Index type: flat
- Dimensions: 1536
- Max elements: 1,000,000

Godast Storage:
- Storage path: /var/lib/gox/projects/default/storage
- KV, graph, file, search enabled

Environment:
- Requires OPENAI_API_KEY
- Project ID: default
- Debug mode enabled

Orchestration:
- Startup timeout: 60s, shutdown: 30s
- Max retries: 3, retry delay: 5s
- Health check interval: 15s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/knowledge-backend-rag.yaml
```

Requires OpenAI API key in environment.
