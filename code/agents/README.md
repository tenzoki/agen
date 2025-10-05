# Agents

Agent catalog - 27 processing agents implementing business logic for file processing, text analysis, content transformation, and storage operations.

## Intent

Provides reusable processing units (agents) that implement business logic only. Framework handles all infrastructure (connections, lifecycle, routing). Agents are stateless, composable, and deployment-agnostic. See [/reflect/architecture/agents.md](../../reflect/architecture/agents.md) for detailed architecture.

## Agent Catalog

### File Processing (3 agents)
- **adapter** - Protocol adaptation between agent types
- **file_ingester** - Watch directories, emit file events on creation
- **file_writer** - Write output files with template support

### Text Processing (4 agents)
- **text_extractor_native** - Extract text from PDF/DOCX/XLSX
- **text_transformer** - Transform text with metadata preservation
- **text_chunker** - Intelligent text chunking with overlap
- **text_analyzer** - Sentiment, keywords, language detection

### Content Analysis (4 agents)
- **json_analyzer** - JSON validation and schema extraction
- **xml_analyzer** - XML validation and namespace handling
- **binary_analyzer** - File type detection and hash computation
- **image_analyzer** - Image metadata and dimension extraction

### Storage & Search (4 agents)
- **godast_storage** - OmniStore integration for persistence
- **search_indexer** - Full-text indexing and search
- **metadata_collector** - Metadata extraction and aggregation
- **chunk_writer** - Chunk persistence to storage

### Advanced Processing (6 agents)
- **ner_agent** - Named entity recognition (NER) for PII detection
- **ocr_http_stub** - OCR service HTTP client
- **context_enricher** - Context enhancement with external data
- **summary_generator** - Text summarization
- **embedding_agent** - Vector embeddings generation
- **rag_agent** - Retrieval-augmented generation (RAG)

### Pipeline Utilities (6 agents)
- **strategy_selector** - Dynamic routing based on content
- **report_generator** - Report synthesis from aggregated data
- **dataset_builder** - Dataset construction for ML
- **anonymizer** - PII anonymization
- **anonymization_store** - Anonymization mapping persistence
- **vectorstore_agent** - Vector database integration

## Usage

Implement custom agent (3 lines):
```go
type MyAgent struct { agent.DefaultAgentRunner }

func (a *MyAgent) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
    result := transform(msg.Payload)
    return &client.BrokerMessage{Payload: result}, nil
}

func main() { agent.Run(&MyAgent{}, "my-agent") }
```

Define agent type in pool.yaml:
```yaml
agents:
  - type: "my-agent"
    binary: "./bin/my_agent"
    operator: "spawn"
```

Deploy in cell (cells.yaml):
```yaml
agents:
  - id: "my-agent-001"
    agent_type: "my-agent"
    ingress: "file:input/*.txt"
    egress: "pub:processed"
```

## Setup

Build all agents:
```bash
cd code/agents
for agent in */; do
    go build -o ../../bin/$(basename $agent) ./$agent
done
```

Build specific agent:
```bash
go build -o ../../bin/text_transformer ./text_transformer
```

Dependencies:
- cellorg framework (github.com/tenzoki/agen/cellorg)
- omni storage (optional, for storage-backed agents)

## Tests

Run all agent tests:
```bash
go test ./... -v
```

Test specific agent:
```bash
go test ./text_transformer -v
go test ./file_ingester -v
```

## Demo

Agent-specific demos in individual agent directories:
- `file_ingester/examples/` - File watching demo
- `ner_agent/examples/` - NER pipeline demo
- `text_transformer/examples/` - Text transformation demo

Multi-agent pipeline demos in workbench:
- `/workbench/demos/gox_demo` - Full pipeline
- `/workbench/demos/gox_anonymization` - Anonymization pipeline

See [/reflect/architecture/agents.md](../../reflect/architecture/agents.md) for agent patterns and development workflow.
