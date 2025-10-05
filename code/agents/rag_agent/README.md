# RAG Agent

Retrieval-Augmented Generation service using OpenAI for context-aware text generation.

## Intent

Implements RAG pattern by retrieving relevant context from vector store and generating answers using LLM. Combines vector search, context assembly, and LLM generation for accurate, source-grounded responses.

## Usage

Input: Query with optional context retrieval parameters
Output: Generated answer with source references and confidence

Configuration:
- `openai_model`: LLM model (default: "gpt-4o-mini")
- `max_tokens`: Maximum response tokens
- `temperature`: Generation temperature
- `top_k`: Number of contexts to retrieve
- `enable_citations`: Include source citations

## Setup

Dependencies:
- OpenAI API key required (`OPENAI_API_KEY` environment variable)
- vectorstore_agent (for context retrieval)
- Internet connection for API access

Build:
```bash
go build -o bin/rag_agent ./code/agents/rag_agent
```

## Tests

No tests implemented

## Demo

No demo available
