# Workflow Pipeline Examples

This directory contains comprehensive examples for the GOX Framework workflow orchestration agents, demonstrating intelligent processing strategies, content enrichment, dataset building, and smart text chunking.

## Overview

The workflow pipeline showcases four advanced workflow agents:

- **Strategy Selector** (`strategy_selector`) - Intelligent routing and processing strategy selection
- **Context Enricher** (`context_enricher`) - Content enrichment with external data sources
- **Dataset Builder** (`dataset_builder`) - Automated dataset assembly and validation
- **Text Chunker** (`text_chunker`) - Intelligent text segmentation and chunking

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Input Data    │───▶│ Strategy        │───▶│ Workflow        │
│   - Multi-format│    │ Selector        │    │ Orchestrator    │
│   - Varied types│    │ - Route by type │    │ - Coordinate    │
└─────────────────┘    │ - Select method │    │ - Monitor       │
                       └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Context         │    │ Text Chunker    │    │ Dataset Builder │
│ Enricher        │    │ - Smart split   │    │ - Aggregate     │
│ - Add metadata  │    │ - Semantic      │    │ - Validate      │
│ - External APIs │    │ - Contextual    │    │ - Structure     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │ Processed       │    │ Final Dataset   │
                       │ Chunks          │    │ - Validated     │
                       │ - Optimized     │    │ - Structured    │
                       │ - Enriched      │    │ - Ready for use │
                       └─────────────────┘    └─────────────────┘
```

## Features Demonstrated

### 1. Strategy Selection and Routing
- Content-based routing decisions
- Processing strategy optimization
- Load balancing and resource allocation
- Dynamic workflow adaptation
- Performance-based strategy switching

### 2. Context Enrichment
- External data source integration
- Metadata augmentation
- Semantic enhancement
- Cross-reference validation
- API-based enrichment services

### 3. Dataset Building
- Multi-source data aggregation
- Schema validation and enforcement
- Data quality assessment
- Automated dataset structuring
- Version control and lineage tracking

### 4. Intelligent Text Chunking
- Semantic boundary detection
- Context-preserving segmentation
- Size optimization algorithms
- Content-aware splitting strategies
- Hierarchical chunk organization

## Quick Start

### Prerequisites

1. **Build the GOX framework and agents:**
   ```bash
   cd /path/to/gox
   make build
   ```

2. **Ensure required binaries exist:**
   ```bash
   ls build/
   # Should include: strategy_selector, context_enricher, dataset_builder, text_chunker
   ```

### Running Examples

1. **Complete workflow demo:**
   ```bash
   cd examples/workflow-pipeline
   ./run_workflow_demo.sh
   ```

2. **Specific workflow agent:**
   ```bash
   # Strategy selection only
   ./run_workflow_demo.sh --workflow=strategy

   # Context enrichment only
   ./run_workflow_demo.sh --workflow=enrichment

   # Dataset building only
   ./run_workflow_demo.sh --workflow=dataset

   # Text chunking only
   ./run_workflow_demo.sh --workflow=chunking
   ```

3. **Custom configuration:**
   ```bash
   ./run_workflow_demo.sh --config=/path/to/custom/config.yaml
   ```

## Example Files Structure

```
examples/workflow-pipeline/
├── README.md                          # This documentation
├── run_workflow_demo.sh               # Main demo script
├── workflow_orchestrator.go           # Go orchestrator implementation
├── cell_configs/                      # Cell configuration files
│   ├── strategy_selection_cell.yaml
│   ├── context_enrichment_cell.yaml
│   ├── dataset_building_cell.yaml
│   ├── text_chunking_cell.yaml
│   └── complete_workflow_cell.yaml
├── input/                             # Sample input data
│   ├── documents/
│   │   ├── research_papers/
│   │   ├── news_articles/
│   │   ├── technical_docs/
│   │   └── mixed_content/
│   ├── datasets/
│   │   ├── structured/
│   │   ├── semi_structured/
│   │   └── unstructured/
│   └── enrichment_sources/
│       ├── apis/
│       ├── databases/
│       └── external_feeds/
├── output/                            # Workflow results
│   ├── strategy_decisions/
│   ├── enriched_content/
│   ├── built_datasets/
│   └── chunked_text/
├── configs/                           # Workflow configurations
│   ├── strategy_rules.yaml
│   ├── enrichment_mappings.yaml
│   ├── dataset_schemas.yaml
│   └── chunking_policies.yaml
└── monitoring/                        # Workflow monitoring
    ├── metrics.json
    ├── performance_logs/
    └── quality_reports/
```

## Agent Demonstrations

### Strategy Selector Example

**Purpose:** Intelligent routing and processing strategy selection based on content type, size, complexity, and available resources.

**Input:** Mixed content requiring different processing approaches
**Process:**
- Content analysis and classification
- Resource availability assessment
- Strategy selection based on rules engine
- Dynamic load balancing

**Sample Configuration:**
```yaml
- id: "strategy-selector-001"
  agent_type: "strategy_selector"
  ingress: "file:input/documents/**/*"
  egress: "route:processing-strategy"
  config:
    routing_strategy: "content_analysis"
    strategies:
      - name: "fast_processing"
        conditions:
          - file_size: "<1MB"
          - content_type: "text"
        route_to: "fast-processor"
      - name: "deep_analysis"
        conditions:
          - file_size: ">10MB"
          - content_type: "pdf"
        route_to: "deep-analyzer"
      - name: "parallel_processing"
        conditions:
          - file_count: ">100"
        route_to: "parallel-processor"
    load_balancing:
      enabled: true
      algorithm: "round_robin"
      health_check: true
    performance_monitoring:
      enabled: true
      metrics: ["throughput", "latency", "resource_usage"]
```

**Sample Output:**
```json
{
  "strategy_decision": {
    "selected_strategy": "deep_analysis",
    "reasoning": [
      "File size 15MB exceeds threshold for fast processing",
      "PDF format requires OCR capabilities",
      "Deep analyzer has available capacity"
    ],
    "route_destination": "deep-analyzer-pool",
    "estimated_processing_time": "45s",
    "resource_allocation": {
      "cpu_cores": 2,
      "memory_mb": 512,
      "priority": "high"
    }
  },
  "alternatives": [
    {
      "strategy": "parallel_processing",
      "score": 0.7,
      "trade_offs": "Faster but higher resource usage"
    }
  ],
  "monitoring": {
    "decision_time_ms": 15,
    "confidence_score": 0.92,
    "fallback_available": true
  }
}
```

### Context Enricher Example

**Purpose:** Enhance content with additional context from external sources, metadata, and cross-references.

**Input:** Documents requiring enrichment with external context
**Process:**
- Content analysis for enrichment opportunities
- External API queries and data retrieval
- Metadata augmentation and validation
- Context integration and quality assessment

**Sample Configuration:**
```yaml
- id: "context-enricher-001"
  agent_type: "context_enricher"
  ingress: "sub:content-for-enrichment"
  egress: "pub:enriched-content"
  config:
    enrichment_sources:
      - type: "api"
        name: "knowledge_graph"
        url: "https://api.knowledge.com/v1/"
        fields: ["entities", "relationships", "concepts"]
        rate_limit: 100
      - type: "database"
        name: "reference_db"
        connection: "postgresql://localhost/refs"
        queries:
          - table: "authors"
            match_field: "name"
          - table: "citations"
            match_field: "doi"
      - type: "external_feed"
        name: "news_api"
        url: "https://newsapi.org/v2/"
        keywords_extraction: true
    enrichment_rules:
      - trigger: "person_name_detected"
        action: "lookup_biography"
        sources: ["knowledge_graph", "reference_db"]
      - trigger: "location_mentioned"
        action: "add_geographic_context"
        sources: ["geographic_api"]
      - trigger: "technical_term_found"
        action: "add_definition"
        sources: ["technical_glossary"]
    quality_control:
      confidence_threshold: 0.8
      max_enrichment_ratio: 0.3
      validate_sources: true
```

**Sample Output:**
```json
{
  "original_content": {
    "text": "Dr. Smith published research on machine learning in Berlin.",
    "metadata": {"type": "academic_text", "length": 58}
  },
  "enrichments": [
    {
      "type": "person_enrichment",
      "entity": "Dr. Smith",
      "added_context": {
        "full_name": "Dr. John Smith",
        "affiliation": "Technical University of Berlin",
        "research_areas": ["Machine Learning", "Computer Vision"],
        "h_index": 42,
        "recent_publications": 15
      },
      "source": "knowledge_graph",
      "confidence": 0.94
    },
    {
      "type": "location_enrichment",
      "entity": "Berlin",
      "added_context": {
        "country": "Germany",
        "coordinates": [52.5200, 13.4050],
        "universities": ["TU Berlin", "Humboldt University"],
        "tech_hub_score": 8.7
      },
      "source": "geographic_api",
      "confidence": 0.99
    },
    {
      "type": "concept_enrichment",
      "entity": "machine learning",
      "added_context": {
        "definition": "Field of AI that enables computers to learn...",
        "subcategories": ["supervised", "unsupervised", "reinforcement"],
        "trending_topics": ["transformers", "federated_learning"],
        "market_size": "$15.3B (2024)"
      },
      "source": "technical_glossary",
      "confidence": 0.91
    }
  ],
  "enrichment_summary": {
    "total_enrichments": 3,
    "processing_time": "1.2s",
    "external_api_calls": 5,
    "content_expansion": "250%",
    "quality_score": 0.95
  }
}
```

### Dataset Builder Example

**Purpose:** Automatically aggregate, validate, and structure datasets from multiple sources with quality control.

**Input:** Multiple data sources requiring aggregation into a cohesive dataset
**Process:**
- Multi-source data ingestion
- Schema inference and validation
- Data quality assessment
- Automated structuring and organization

**Sample Configuration:**
```yaml
- id: "dataset-builder-001"
  agent_type: "dataset_builder"
  ingress: "sub:data-sources"
  egress: "file:output/datasets/"
  config:
    sources:
      - type: "json_files"
        path: "input/datasets/structured/*.json"
        schema_inference: true
      - type: "csv_files"
        path: "input/datasets/structured/*.csv"
        header_detection: true
        delimiter_auto_detect: true
      - type: "api_endpoint"
        url: "https://api.data-source.com/v1/datasets"
        pagination: true
        rate_limit: 50
      - type: "database_table"
        connection: "postgresql://localhost/data"
        tables: ["users", "transactions", "products"]
    dataset_config:
      output_format: "parquet"
      compression: "snappy"
      partitioning:
        field: "date"
        strategy: "daily"
      schema_evolution: true
      version_control: true
    quality_control:
      duplicate_detection: true
      null_value_threshold: 0.05
      outlier_detection: true
      data_profiling: true
      validation_rules:
        - field: "email"
          rule: "valid_email_format"
        - field: "age"
          rule: "range(0, 150)"
        - field: "amount"
          rule: "positive_number"
    metadata_generation:
      lineage_tracking: true
      statistics: true
      documentation: "auto"
      tags: ["processed", "validated", "production"]
```

**Sample Output:**
```json
{
  "dataset_info": {
    "id": "customer_analytics_v1.2",
    "name": "Customer Analytics Dataset",
    "version": "1.2.0",
    "created": "2024-09-27T10:00:00Z",
    "total_records": 1250000,
    "file_size": "485MB",
    "format": "parquet"
  },
  "sources": [
    {
      "name": "customer_json",
      "type": "json_files",
      "records_contributed": 850000,
      "last_updated": "2024-09-26T18:30:00Z",
      "quality_score": 0.94
    },
    {
      "name": "transaction_csv",
      "type": "csv_files",
      "records_contributed": 300000,
      "last_updated": "2024-09-27T08:15:00Z",
      "quality_score": 0.89
    },
    {
      "name": "product_api",
      "type": "api_endpoint",
      "records_contributed": 100000,
      "last_updated": "2024-09-27T09:45:00Z",
      "quality_score": 0.97
    }
  ],
  "schema": {
    "fields": [
      {"name": "customer_id", "type": "string", "nullable": false},
      {"name": "email", "type": "string", "nullable": false},
      {"name": "age", "type": "integer", "nullable": true},
      {"name": "registration_date", "type": "timestamp", "nullable": false},
      {"name": "total_purchases", "type": "decimal", "nullable": false},
      {"name": "last_activity", "type": "timestamp", "nullable": true}
    ],
    "primary_key": ["customer_id"],
    "indexes": ["email", "registration_date"],
    "partitions": ["date(registration_date)"]
  },
  "quality_report": {
    "overall_score": 0.92,
    "completeness": 0.96,
    "validity": 0.91,
    "uniqueness": 0.89,
    "consistency": 0.94,
    "issues_found": [
      {"type": "duplicate_records", "count": 1250, "percentage": 0.1},
      {"type": "invalid_email", "count": 89, "percentage": 0.007},
      {"type": "outlier_age", "count": 23, "percentage": 0.002}
    ],
    "recommendations": [
      "Implement email validation at source",
      "Review age data entry constraints",
      "Add duplicate detection to ingestion pipeline"
    ]
  },
  "lineage": {
    "processing_steps": [
      "source_ingestion",
      "schema_inference",
      "quality_validation",
      "deduplication",
      "formatting",
      "partitioning"
    ],
    "transformations": [
      "email_normalization",
      "date_standardization",
      "currency_conversion"
    ],
    "dependencies": [
      "customer_master_data_v2.1",
      "product_catalog_v1.8"
    ]
  }
}
```

### Text Chunker Example

**Purpose:** Intelligently segment text into optimal chunks while preserving semantic meaning and context.

**Input:** Large text documents requiring intelligent segmentation
**Process:**
- Semantic boundary detection
- Context-preserving chunk creation
- Size optimization based on target use case
- Hierarchical organization of chunks

**Sample Configuration:**
```yaml
- id: "text-chunker-001"
  agent_type: "text_chunker"
  ingress: "sub:large-text-documents"
  egress: "pub:text-chunks"
  config:
    chunking_strategies:
      - name: "semantic_chunking"
        method: "sentence_boundary"
        max_chunk_size: 512
        overlap_size: 50
        preserve_sentences: true
      - name: "paragraph_chunking"
        method: "paragraph_boundary"
        max_chunk_size: 1024
        overlap_size: 100
        preserve_paragraphs: true
      - name: "section_chunking"
        method: "heading_boundary"
        max_chunk_size: 2048
        overlap_size: 200
        preserve_sections: true
    content_analysis:
      language_detection: true
      topic_modeling: true
      entity_recognition: true
      complexity_scoring: true
    optimization:
      target_use_case: "rag_pipeline"
      embedding_model: "sentence-transformers"
      similarity_threshold: 0.85
      context_window: 4096
    quality_control:
      min_chunk_size: 50
      coherence_check: true
      information_density: true
      readability_score: true
    output_format:
      include_metadata: true
      preserve_formatting: false
      add_chunk_ids: true
      cross_references: true
```

**Sample Output:**
```json
{
  "document_info": {
    "id": "research_paper_001",
    "title": "Advances in Machine Learning",
    "total_length": 45000,
    "language": "en",
    "complexity_score": 0.78
  },
  "chunking_summary": {
    "total_chunks": 87,
    "average_chunk_size": 517,
    "chunking_strategy": "semantic_chunking",
    "processing_time": "2.3s",
    "quality_score": 0.91
  },
  "chunks": [
    {
      "chunk_id": "chunk_001",
      "sequence_number": 1,
      "text": "Machine learning has revolutionized the field of artificial intelligence by enabling computers to learn patterns from data without explicit programming. This paradigm shift has led to breakthroughs in various domains including computer vision, natural language processing, and robotics.",
      "metadata": {
        "section": "introduction",
        "word_count": 45,
        "sentence_count": 2,
        "start_position": 0,
        "end_position": 287,
        "topics": ["machine_learning", "artificial_intelligence"],
        "entities": ["computer vision", "natural language processing", "robotics"],
        "readability_score": 0.72,
        "information_density": 0.84
      },
      "relationships": {
        "previous_chunk": null,
        "next_chunk": "chunk_002",
        "parent_section": "introduction",
        "related_chunks": ["chunk_015", "chunk_032"]
      }
    },
    {
      "chunk_id": "chunk_002",
      "sequence_number": 2,
      "text": "The foundation of modern machine learning rests on statistical learning theory and computational algorithms that can automatically improve their performance through experience. Key algorithms include supervised learning methods such as support vector machines and neural networks, unsupervised learning techniques like clustering and dimensionality reduction, and reinforcement learning approaches that learn through interaction with environments.",
      "metadata": {
        "section": "introduction",
        "word_count": 64,
        "sentence_count": 2,
        "start_position": 238,
        "end_position": 723,
        "topics": ["statistical_learning", "algorithms", "supervised_learning"],
        "entities": ["support vector machines", "neural networks", "clustering"],
        "readability_score": 0.68,
        "information_density": 0.91
      },
      "relationships": {
        "previous_chunk": "chunk_001",
        "next_chunk": "chunk_003",
        "parent_section": "introduction",
        "related_chunks": ["chunk_018", "chunk_025", "chunk_041"]
      }
    }
  ],
  "section_hierarchy": [
    {
      "section": "introduction",
      "chunks": ["chunk_001", "chunk_002", "chunk_003"],
      "summary": "Overview of machine learning fundamentals"
    },
    {
      "section": "methodology",
      "chunks": ["chunk_004", "chunk_005", "chunk_006", "chunk_007"],
      "summary": "Research methods and experimental design"
    },
    {
      "section": "results",
      "chunks": ["chunk_008", "chunk_009", "chunk_010"],
      "summary": "Experimental results and findings"
    }
  ],
  "optimization_report": {
    "target_achieved": true,
    "average_similarity": 0.87,
    "context_preservation": 0.93,
    "chunk_coherence": 0.89,
    "recommendations": [
      "Consider slightly larger chunks for dense technical sections",
      "Increase overlap for complex mathematical content"
    ]
  }
}
```

## Complete Workflow Integration

### Multi-Agent Workflow Cell Configuration

```yaml
cell:
  id: "workflow:intelligent-content-processing"
  description: "Complete intelligent content processing workflow"

  agents:
    # Stage 1: Strategy Selection
    - id: "strategy-selector-001"
      agent_type: "strategy_selector"
      ingress: "file:input/documents/**/*"
      egress: "route:processing-strategy"
      config:
        routing_strategy: "content_analysis"
        performance_optimization: true

    # Stage 2: Content Processing (parallel paths)
    - id: "text-chunker-001"
      agent_type: "text_chunker"
      ingress: "route:processing-strategy:text"
      egress: "pub:chunked-text"
      config:
        chunking_strategy: "semantic_chunking"
        target_use_case: "rag_pipeline"

    - id: "context-enricher-001"
      agent_type: "context_enricher"
      ingress: "route:processing-strategy:enrich"
      egress: "pub:enriched-content"
      config:
        enrichment_sources: ["knowledge_graph", "reference_db"]
        quality_control: true

    # Stage 3: Dataset Building
    - id: "dataset-builder-001"
      agent_type: "dataset_builder"
      ingress: "sub:chunked-text,enriched-content"
      egress: "file:output/final-dataset/"
      config:
        output_format: "parquet"
        quality_control: true
        version_control: true

    # Stage 4: Quality Monitoring
    - id: "quality-monitor-001"
      agent_type: "metadata_collector"
      ingress: "sub:processing-metrics"
      egress: "file:output/quality-reports/"
      config:
        collect_performance_metrics: true
        generate_quality_reports: true

  workflow:
    execution_model: "pipeline"
    error_handling: "continue_with_logging"
    monitoring:
      enabled: true
      metrics: ["throughput", "quality", "resource_usage"]
      alerting: true
```

## Performance Characteristics

### Throughput Benchmarks
- **Strategy Selector**: ~1000 decisions/second
- **Context Enricher**: ~50 documents/second (API-dependent)
- **Dataset Builder**: ~100MB/second processing rate
- **Text Chunker**: ~500 documents/second

### Resource Usage
- **Memory**: ~200MB base + data-dependent
- **CPU**: Moderate to high (varies by workflow complexity)
- **I/O**: High for dataset operations, moderate for others
- **Network**: High for context enrichment (external APIs)

## Testing

### Unit Tests
```bash
# Test all workflow agents
go test ./agents/strategy_selector ./agents/context_enricher ./agents/dataset_builder ./agents/text_chunker -v

# Test specific workflow
go test ./agents/strategy_selector -v -run TestStrategySelection
```

### Integration Tests
```bash
# Run complete workflow pipeline test
./examples/workflow-pipeline/test_workflow_pipeline.sh

# Test specific workflow combinations
./examples/workflow-pipeline/test_workflow_pipeline.sh --workflow=enrichment,dataset
```

## Customization

### Custom Strategy Rules
```yaml
# Custom routing strategies
strategy_rules:
  - name: "pdf_processing"
    conditions:
      - file_extension: ".pdf"
      - file_size: ">5MB"
    actions:
      - route_to: "ocr_pipeline"
      - set_priority: "high"
      - allocate_memory: "1GB"
```

### Custom Enrichment Sources
```go
// Custom enrichment source implementation
type CustomEnrichmentSource struct {
    Name string
    URL  string
}

func (c *CustomEnrichmentSource) Enrich(content string) (map[string]interface{}, error) {
    // Custom enrichment logic
    return enrichedData, nil
}
```

### Custom Dataset Schemas
```yaml
# Custom dataset schema definitions
dataset_schemas:
  customer_data:
    version: "2.0"
    fields:
      - name: "customer_id"
        type: "string"
        constraints: ["not_null", "unique"]
      - name: "preferences"
        type: "json"
        validation: "json_schema:customer_preferences.json"
```

## Troubleshooting

### Common Issues

1. **Strategy selection failures**
   - Check routing rules syntax
   - Verify content type detection
   - Review resource availability

2. **Enrichment timeouts**
   - Increase API timeout settings
   - Check external service availability
   - Implement retry mechanisms

3. **Dataset quality issues**
   - Review validation rules
   - Check source data quality
   - Adjust quality thresholds

4. **Chunking inefficiencies**
   - Tune chunk size parameters
   - Adjust overlap settings
   - Review semantic boundary detection

### Debug Mode
```bash
# Enable debug logging for workflow
export GOX_LOG_LEVEL=debug
export GOX_WORKFLOW_DEBUG=true

# Run specific workflow in debug mode
./run_workflow_demo.sh --debug --workflow=strategy
```

## Next Steps

1. **Extend workflows**: Add custom processing strategies
2. **Integrate external services**: Connect to domain-specific APIs
3. **Optimize performance**: Tune for specific data patterns
4. **Add monitoring**: Implement comprehensive workflow observability
5. **Scale deployment**: Deploy across multiple nodes for distributed processing

For more information, see the [main GOX documentation](../../README.md) and workflow orchestration guides.