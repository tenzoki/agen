# GOX Framework - Search Indexing Examples

This directory contains examples demonstrating the search and indexing capabilities of the GOX Framework, focusing on the `search_indexer` agent.

## Overview

The search indexing pipeline showcases how to create searchable indexes from processed data, enabling fast and efficient retrieval of information from large datasets.

## Agent Covered

### Search Indexer (`search_indexer`)
Creates and maintains searchable indexes from various data sources. Supports multiple indexing strategies and search backends.

**Key Features:**
- Full-text search indexing
- Metadata indexing
- Real-time index updates
- Multiple search backends (Elasticsearch, Solr, custom)
- Faceted search support
- Relevance scoring
- Index optimization
- Incremental indexing

**Use Cases:**
- Document search systems
- Log analysis and search
- Content discovery platforms
- Data lake indexing
- Compliance and audit searches
- Business intelligence queries

## Directory Structure

```
search-indexing/
├── README.md                    # This file
├── run_indexing_demo.sh         # Demo execution script
├── input/                       # Sample input data
│   ├── documents/               # Documents to index
│   ├── logs/                    # Log files for indexing
│   └── structured-data/         # Structured data files
├── config/                      # Cell configurations
│   ├── basic_indexing_cell.yaml
│   ├── advanced_indexing_cell.yaml
│   ├── real_time_indexing_cell.yaml
│   └── multi_source_indexing_cell.yaml
└── schemas/                     # Index and search schemas
    ├── document-index-schema.json
    ├── log-index-schema.json
    └── search-config-schema.json
```

## Quick Start

### Run All Indexing Examples
```bash
./run_indexing_demo.sh
```

### Run Specific Indexing Type
```bash
# Basic document indexing
./run_indexing_demo.sh --type=basic

# Advanced indexing with facets
./run_indexing_demo.sh --type=advanced

# Real-time indexing
./run_indexing_demo.sh --type=realtime

# Multi-source indexing
./run_indexing_demo.sh --type=multi
```

### Custom Input Directory
```bash
./run_indexing_demo.sh --input=/path/to/your/documents
```

## Example Configurations

### Basic Document Indexing
```yaml
cell:
  id: "search:basic-document-indexing"
  description: "Basic document indexing cell"

  agents:
    - id: "search-indexer-001"
      agent_type: "search_indexer"
      ingress: "file:input/documents/*"
      egress: "index:documents"
      config:
        index_name: "documents"
        backend: "elasticsearch"
        endpoint: "http://localhost:9200"
        mapping:
          title:
            type: "text"
            analyzer: "standard"
          content:
            type: "text"
            analyzer: "standard"
          metadata:
            type: "object"
        settings:
          number_of_shards: 1
          number_of_replicas: 0
```

### Advanced Indexing with Facets
```yaml
cell:
  id: "search:advanced-indexing"
  description: "Advanced indexing with faceted search"

  agents:
    - id: "search-indexer-002"
      agent_type: "search_indexer"
      ingress: "file:input/**/*"
      egress: "index:advanced"
      config:
        index_name: "advanced_search"
        backend: "elasticsearch"
        endpoint: "http://localhost:9200"
        features:
          faceted_search: true
          auto_complete: true
          spell_correction: true
          relevance_tuning: true
        mapping:
          title:
            type: "text"
            analyzer: "standard"
            fields:
              raw:
                type: "keyword"
          content:
            type: "text"
            analyzer: "content_analyzer"
          tags:
            type: "keyword"
          category:
            type: "keyword"
          created_date:
            type: "date"
          file_size:
            type: "long"
        facets:
          - field: "category"
            type: "terms"
            size: 20
          - field: "tags"
            type: "terms"
            size: 50
          - field: "created_date"
            type: "date_histogram"
            interval: "month"
        analyzers:
          content_analyzer:
            tokenizer: "standard"
            filters:
              - "lowercase"
              - "stop"
              - "snowball"
```

### Real-time Indexing
```yaml
cell:
  id: "search:real-time-indexing"
  description: "Real-time search indexing cell"

  agents:
    - id: "search-indexer-003"
      agent_type: "search_indexer"
      ingress: "stream:live-data"
      egress: "index:realtime"
      config:
        index_name: "realtime_search"
        backend: "elasticsearch"
        endpoint: "http://localhost:9200"
        indexing_mode: "real_time"
        batch_size: 100
        flush_interval: 1000  # milliseconds
        refresh_interval: "1s"
        pipeline_processing: true
        ingest_pipelines:
          - name: "timestamp_enrichment"
            processors:
              - set:
                  field: "indexed_at"
                  value: "{{_ingest.timestamp}}"
          - name: "content_extraction"
            processors:
              - attachment:
                  field: "data"
                  target_field: "attachment"
```

### Multi-source Indexing Pipeline
```yaml
cell:
  id: "search:multi-source-indexing"
  description: "Multi-source indexing pipeline"

  agents:
    - id: "data-router-001"
      agent_type: "strategy_selector"
      ingress: "file:input/**/*"
      egress: "route:content-type"
      config:
        routing_strategy: "content_type"
        routes:
          documents: "pub:documents"
          logs: "pub:logs"
          structured: "pub:structured"

    - id: "document-indexer-001"
      agent_type: "search_indexer"
      ingress: "sub:documents"
      egress: "index:documents"
      config:
        index_name: "documents"
        backend: "elasticsearch"
        mapping_template: "document_template"

    - id: "log-indexer-001"
      agent_type: "search_indexer"
      ingress: "sub:logs"
      egress: "index:logs"
      config:
        index_name: "logs"
        backend: "elasticsearch"
        mapping_template: "log_template"
        time_based_indices: true
        index_pattern: "logs-{yyyy.MM.dd}"

    - id: "structured-indexer-001"
      agent_type: "search_indexer"
      ingress: "sub:structured"
      egress: "index:structured"
      config:
        index_name: "structured_data"
        backend: "elasticsearch"
        mapping_template: "structured_template"
        nested_field_support: true
```

## Input Data Formats

### Document Format
Expected format for document indexing:
```json
{
  "id": "doc-001",
  "title": "GOX Framework Documentation",
  "content": "The GOX Framework is a powerful agent-based processing system...",
  "metadata": {
    "author": "GOX Team",
    "created_date": "2024-09-27T10:00:00Z",
    "file_type": "markdown",
    "file_size": 15672,
    "tags": ["documentation", "framework", "agents"],
    "category": "technical_docs",
    "language": "en",
    "version": "1.0"
  },
  "extracted_entities": {
    "organizations": ["GOX Framework"],
    "technologies": ["Go", "YAML", "JSON"],
    "concepts": ["agent", "pipeline", "cell"]
  }
}
```

### Log Entry Format
```json
{
  "timestamp": "2024-09-27T10:15:32.123Z",
  "level": "INFO",
  "logger": "gox.agent.binary_analyzer",
  "message": "Successfully analyzed binary file",
  "metadata": {
    "file_path": "/data/sample.exe",
    "processing_time": 1250,
    "file_size": 2048576,
    "entropy": 7.2
  },
  "labels": {
    "environment": "production",
    "node": "worker-001",
    "pipeline": "security-analysis"
  },
  "extracted_fields": {
    "ip_addresses": ["192.168.1.100"],
    "error_codes": [],
    "user_agents": []
  }
}
```

### Structured Data Format
```json
{
  "record_id": "rec-001",
  "data_type": "analysis_result",
  "timestamp": "2024-09-27T10:00:00Z",
  "source": {
    "system": "gox-analyzer",
    "agent_id": "binary-analyzer-001"
  },
  "content": {
    "file_analysis": {
      "file_path": "/data/sample.exe",
      "file_type": "PE32 executable",
      "security_findings": [
        {
          "type": "high_entropy_section",
          "section": ".text",
          "entropy": 7.8,
          "severity": "medium"
        }
      ]
    }
  },
  "search_metadata": {
    "keywords": ["security", "binary", "analysis", "PE32"],
    "boost_factor": 1.5,
    "exclude_from_search": false
  }
}
```

## Search Query Examples

### Basic Text Search
```json
{
  "query": {
    "multi_match": {
      "query": "GOX Framework agent",
      "fields": ["title^2", "content", "metadata.tags"]
    }
  }
}
```

### Faceted Search with Filters
```json
{
  "query": {
    "bool": {
      "must": {
        "match": {
          "content": "security analysis"
        }
      },
      "filter": [
        {
          "term": {
            "category": "technical_docs"
          }
        },
        {
          "range": {
            "created_date": {
              "gte": "2024-01-01"
            }
          }
        }
      ]
    }
  },
  "aggs": {
    "categories": {
      "terms": {
        "field": "category",
        "size": 10
      }
    },
    "tags": {
      "terms": {
        "field": "tags",
        "size": 20
      }
    }
  }
}
```

### Time-based Log Search
```json
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "message": "error"
          }
        },
        {
          "range": {
            "timestamp": {
              "gte": "now-1h"
            }
          }
        }
      ]
    }
  },
  "sort": [
    {
      "timestamp": {
        "order": "desc"
      }
    }
  ]
}
```

## Backend Configuration

### Elasticsearch Backend
```yaml
search_indexer:
  backend: "elasticsearch"
  config:
    endpoint: "http://localhost:9200"
    username: "elastic"  # optional
    password: "password"  # optional
    api_key: "api_key"    # alternative to username/password
    timeout: 30
    max_retries: 3
    index_settings:
      number_of_shards: 1
      number_of_replicas: 0
      refresh_interval: "1s"
    cluster_settings:
      max_result_window: 100000
```

### Solr Backend
```yaml
search_indexer:
  backend: "solr"
  config:
    endpoint: "http://localhost:8983/solr"
    collection: "gox_documents"
    commit_within: 1000
    batch_size: 100
    schema_config:
      unique_key: "id"
      default_search_field: "content"
```

### Custom Backend
```yaml
search_indexer:
  backend: "custom"
  config:
    driver: "custom_search_driver"
    endpoint: "http://custom-search:8080"
    api_version: "v1"
    authentication:
      type: "bearer_token"
      token: "your_api_token"
```

## Performance Tuning

### Bulk Indexing
```yaml
search_indexer:
  config:
    indexing_mode: "bulk"
    batch_size: 1000
    max_batch_size: 10000
    flush_interval: 5000  # milliseconds
    parallel_workers: 4
    memory_limit: "512MB"
```

### Index Optimization
```yaml
search_indexer:
  config:
    optimization:
      force_merge_segments: 1
      delete_by_query_conflicts: "proceed"
      refresh_interval: "30s"
      translog_flush_threshold_size: "200MB"
```

### Search Performance
```yaml
search_indexer:
  config:
    search_optimization:
      enable_caching: true
      cache_size: "100MB"
      enable_profiling: true
      slow_query_threshold: "1s"
      routing_enabled: true
```

## Monitoring and Maintenance

### Index Health Monitoring
```yaml
search_indexer:
  config:
    monitoring:
      health_check_interval: "30s"
      alert_on_errors: true
      metrics_collection: true
      index_size_alerts:
        warning_threshold: "1GB"
        critical_threshold: "5GB"
```

### Maintenance Tasks
```yaml
search_indexer:
  config:
    maintenance:
      auto_cleanup: true
      retention_policy: "30d"
      cleanup_schedule: "0 2 * * *"  # daily at 2 AM
      reindex_schedule: "0 3 * * 0"  # weekly on Sunday at 3 AM
```

## Requirements

- GOX Framework v3+
- Search backend (Elasticsearch, Solr, or custom)
- Built search indexer agent: `build/search_indexer`

## Building Required Agents

```bash
cd ../../
make build-search  # or individual builds:
go build -o build/search_indexer ./agents/search_indexer
```

## Advanced Features

### Custom Analyzers
Define custom text analysis pipelines for specialized content processing.

### Machine Learning Integration
Integrate with ML models for enhanced relevance scoring and content understanding.

### Multi-language Support
Configure language-specific analyzers and stemming for international content.

### Security and Access Control
Implement field-level security and role-based access control for sensitive data.

### High Availability
Configure clustering and replication for production search deployments.