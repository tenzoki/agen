# GOX Framework - Adapter Integration Examples

This directory contains examples demonstrating the adapter integration capabilities of the GOX Framework, showcasing how to integrate external systems and services with the GOX agent ecosystem.

## Overview

The adapter integration examples show how to connect GOX agents with external systems, databases, APIs, and services. Adapters serve as bridges between the GOX Framework and the outside world, enabling seamless data flow and system integration.

## Adapter Types Covered

### 1. External API Adapter
Connects to REST APIs, GraphQL endpoints, and other web services.

**Key Features:**
- HTTP/HTTPS client integration
- Authentication handling (OAuth, API keys, tokens)
- Request/response transformation
- Rate limiting and retry logic
- Error handling and circuit breaker patterns

**Use Cases:**
- Third-party service integration
- Data enrichment from external sources
- Webhook processing
- API gateway integration

### 2. Database Adapter
Integrates with various database systems for data persistence and retrieval.

**Key Features:**
- Multi-database support (PostgreSQL, MySQL, MongoDB, etc.)
- Connection pooling
- Transaction management
- Query optimization
- Schema migration support

**Use Cases:**
- Data persistence
- Analytics and reporting
- Legacy system integration
- Data warehousing

### 3. Message Queue Adapter
Connects to message brokers and queue systems.

**Key Features:**
- Multiple broker support (RabbitMQ, Apache Kafka, AWS SQS, etc.)
- Message routing and transformation
- Dead letter queue handling
- Batch processing
- Consumer group management

**Use Cases:**
- Event-driven architectures
- Asynchronous processing
- System decoupling
- Scalable message processing

### 4. File System Adapter
Integrates with local and remote file systems.

**Key Features:**
- Local file system access
- Remote storage (S3, Azure Blob, GCS)
- FTP/SFTP integration
- File watching and monitoring
- Batch file operations

**Use Cases:**
- File processing pipelines
- Data import/export
- Backup and archival
- Legacy file integration

### 5. Streaming Adapter
Handles real-time data streams and event processing.

**Key Features:**
- Stream processing
- Real-time analytics
- Event sourcing
- Windowing operations
- Backpressure handling

**Use Cases:**
- Real-time monitoring
- Stream analytics
- Event processing
- IoT data handling

## Directory Structure

```
adapter-integration/
├── README.md                    # This file
├── run_adapter_demo.sh          # Demo execution script
├── input/                       # Sample input data
│   ├── api-configs/             # API configuration files
│   ├── database-schemas/        # Database schema definitions
│   ├── message-samples/         # Sample messages
│   └── file-samples/            # Sample files for processing
├── config/                      # Cell configurations
│   ├── api_adapter_cell.yaml
│   ├── database_adapter_cell.yaml
│   ├── queue_adapter_cell.yaml
│   ├── filesystem_adapter_cell.yaml
│   ├── streaming_adapter_cell.yaml
│   └── multi_adapter_cell.yaml
└── schemas/                     # Adapter schemas
    ├── api-adapter-schema.json
    ├── database-adapter-schema.json
    ├── queue-adapter-schema.json
    └── adapter-config-schema.json
```

## Quick Start

### Run All Adapter Examples
```bash
./run_adapter_demo.sh
```

### Run Specific Adapter Type
```bash
# API adapter only
./run_adapter_demo.sh --adapter=api

# Database adapter only
./run_adapter_demo.sh --adapter=database

# Queue adapter only
./run_adapter_demo.sh --adapter=queue

# File system adapter only
./run_adapter_demo.sh --adapter=filesystem

# Streaming adapter only
./run_adapter_demo.sh --adapter=streaming
```

### Custom Configuration
```bash
./run_adapter_demo.sh --input=/path/to/configs --adapter=api
```

## Example Configurations

### API Adapter Configuration
```yaml
cell:
  id: "integration:api-adapter"
  description: "External API integration cell"

  agents:
    - id: "api-adapter-001"
      agent_type: "adapter"
      adapter_type: "http_api"
      ingress: "file:input/api-requests/*"
      egress: "file:output/api-responses/"
      config:
        endpoint: "https://api.example.com/v1"
        authentication:
          type: "bearer_token"
          token: "${API_TOKEN}"
        request_config:
          timeout: 30
          max_retries: 3
          retry_delay: 1000
          headers:
            Content-Type: "application/json"
            User-Agent: "GOX-Framework/3.0"
        response_config:
          include_headers: true
          include_status: true
          format: "json"
        rate_limiting:
          requests_per_second: 10
          burst_capacity: 20
        circuit_breaker:
          failure_threshold: 5
          recovery_timeout: 60
```

### Database Adapter Configuration
```yaml
cell:
  id: "integration:database-adapter"
  description: "Database integration cell"

  agents:
    - id: "db-adapter-001"
      agent_type: "adapter"
      adapter_type: "database"
      ingress: "file:input/sql-queries/*"
      egress: "file:output/query-results/"
      config:
        connection:
          driver: "postgresql"
          host: "localhost"
          port: 5432
          database: "gox_data"
          username: "${DB_USER}"
          password: "${DB_PASSWORD}"
          ssl_mode: "require"
        pool_config:
          max_connections: 10
          min_connections: 2
          max_idle_time: "30m"
          max_lifetime: "1h"
        query_config:
          timeout: 30
          max_rows: 10000
          prepared_statements: true
        transaction_config:
          isolation_level: "read_committed"
          auto_commit: false
```

### Message Queue Adapter Configuration
```yaml
cell:
  id: "integration:queue-adapter"
  description: "Message queue integration cell"

  agents:
    - id: "queue-adapter-001"
      agent_type: "adapter"
      adapter_type: "message_queue"
      ingress: "sub:input-messages"
      egress: "pub:processed-messages"
      config:
        broker:
          type: "rabbitmq"
          url: "amqp://localhost:5672"
          username: "${QUEUE_USER}"
          password: "${QUEUE_PASSWORD}"
        producer:
          exchange: "gox.events"
          routing_key: "processed"
          delivery_mode: "persistent"
          confirm_delivery: true
        consumer:
          queue: "gox.input"
          prefetch_count: 10
          auto_ack: false
          dead_letter_exchange: "gox.dlx"
        retry_policy:
          max_attempts: 3
          backoff_factor: 2
          max_backoff: 30
```

### File System Adapter Configuration
```yaml
cell:
  id: "integration:filesystem-adapter"
  description: "File system integration cell"

  agents:
    - id: "fs-adapter-001"
      agent_type: "adapter"
      adapter_type: "filesystem"
      ingress: "watch:/data/input/"
      egress: "file:/data/output/"
      config:
        source:
          type: "local"
          path: "/data/input"
          watch_patterns: ["*.json", "*.xml", "*.csv"]
          watch_subdirs: true
        destination:
          type: "s3"
          bucket: "gox-processed-data"
          prefix: "processed/"
          region: "us-west-2"
          credentials:
            access_key: "${AWS_ACCESS_KEY}"
            secret_key: "${AWS_SECRET_KEY}"
        processing:
          batch_size: 100
          compression: "gzip"
          encryption: "AES256"
          move_processed: true
          archive_path: "/data/archive"
```

### Streaming Adapter Configuration
```yaml
cell:
  id: "integration:streaming-adapter"
  description: "Real-time streaming integration cell"

  agents:
    - id: "stream-adapter-001"
      agent_type: "adapter"
      adapter_type: "streaming"
      ingress: "stream:kafka://localhost:9092/events"
      egress: "stream:processed-events"
      config:
        source:
          type: "kafka"
          brokers: ["localhost:9092"]
          topic: "raw-events"
          consumer_group: "gox-processors"
          offset: "earliest"
        destination:
          type: "kafka"
          brokers: ["localhost:9092"]
          topic: "processed-events"
        processing:
          window_size: "5m"
          window_type: "tumbling"
          watermark_delay: "1m"
          parallelism: 4
        stream_config:
          buffer_size: 1000
          flush_interval: "1s"
          checkpointing: true
          checkpoint_interval: "10s"
```

## Complete Multi-Adapter Pipeline

The complete pipeline demonstrates integration with multiple external systems:

```yaml
cell:
  id: "integration:multi-adapter-pipeline"
  description: "Multi-system integration pipeline"

  agents:
    - id: "data-router-001"
      agent_type: "strategy_selector"
      ingress: "file:input/**/*"
      egress: "route:data-source"
      config:
        routing_strategy: "source_type"

    - id: "api-enricher-001"
      agent_type: "adapter"
      adapter_type: "http_api"
      ingress: "route:data-source:api"
      egress: "pub:enriched-data"
      config:
        endpoint: "https://enrichment-api.example.com"
        operation: "enrich"

    - id: "db-persister-001"
      agent_type: "adapter"
      adapter_type: "database"
      ingress: "sub:enriched-data"
      egress: "pub:persisted-data"
      config:
        operation: "insert"
        table: "processed_data"

    - id: "queue-notifier-001"
      agent_type: "adapter"
      adapter_type: "message_queue"
      ingress: "sub:persisted-data"
      egress: "queue:notifications"
      config:
        operation: "publish"
        routing_key: "data.processed"

    - id: "fs-archiver-001"
      agent_type: "adapter"
      adapter_type: "filesystem"
      ingress: "sub:persisted-data"
      egress: "file:archive/"
      config:
        operation: "archive"
        compression: true
```

## Adapter Configuration Patterns

### Authentication Patterns
```yaml
# API Key Authentication
authentication:
  type: "api_key"
  header: "X-API-Key"
  value: "${API_KEY}"

# OAuth2 Authentication
authentication:
  type: "oauth2"
  client_id: "${OAUTH_CLIENT_ID}"
  client_secret: "${OAUTH_CLIENT_SECRET}"
  token_url: "https://auth.example.com/token"
  scope: "read write"

# Certificate Authentication
authentication:
  type: "certificate"
  cert_file: "/path/to/client.crt"
  key_file: "/path/to/client.key"
  ca_file: "/path/to/ca.crt"
```

### Error Handling Patterns
```yaml
error_handling:
  strategy: "retry_with_backoff"
  max_retries: 3
  initial_delay: 1000
  max_delay: 30000
  backoff_factor: 2
  retry_conditions:
    - "network_error"
    - "timeout"
    - "rate_limit"
  circuit_breaker:
    failure_threshold: 5
    recovery_timeout: 60
    half_open_max_calls: 3
```

### Data Transformation Patterns
```yaml
transformation:
  input_format: "json"
  output_format: "avro"
  schema_registry: "http://localhost:8081"
  mappings:
    - from: "timestamp"
      to: "event_time"
      type: "datetime"
      format: "ISO8601"
    - from: "user.id"
      to: "user_id"
      type: "string"
    - from: "metadata"
      to: "extra_data"
      type: "object"
      flatten: true
```

## Input Data Formats

### API Request Format
```json
{
  "request_id": "req-001",
  "method": "POST",
  "endpoint": "/users",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user"
  },
  "timeout": 30,
  "retries": 3
}
```

### Database Query Format
```json
{
  "query_id": "q-001",
  "operation": "select",
  "query": "SELECT * FROM users WHERE created_at > $1",
  "parameters": ["2024-09-01T00:00:00Z"],
  "timeout": 30,
  "fetch_size": 1000
}
```

### Message Format
```json
{
  "message_id": "msg-001",
  "timestamp": "2024-09-27T10:00:00Z",
  "type": "user_event",
  "payload": {
    "user_id": "user-123",
    "action": "login",
    "metadata": {
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0..."
    }
  },
  "routing_key": "events.user.login"
}
```

## Output Examples

### API Response
```json
{
  "response_id": "resp-001",
  "request_id": "req-001",
  "status_code": 201,
  "headers": {
    "Content-Type": "application/json",
    "Location": "/users/456"
  },
  "body": {
    "id": 456,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-09-27T10:00:00Z"
  },
  "timing": {
    "total_time": 245,
    "dns_time": 12,
    "connect_time": 35,
    "response_time": 198
  }
}
```

### Database Result
```json
{
  "query_id": "q-001",
  "rows_affected": 25,
  "execution_time": 156,
  "results": [
    {
      "id": 1,
      "name": "Alice Smith",
      "email": "alice@example.com",
      "created_at": "2024-09-25T08:30:00Z"
    }
  ],
  "metadata": {
    "columns": ["id", "name", "email", "created_at"],
    "row_count": 25,
    "has_more": false
  }
}
```

## Requirements

- GOX Framework v3+
- Adapter binary: `build/adapter`
- External systems (APIs, databases, message queues) for full functionality

## Building Required Components

```bash
cd ../../
make build-adapter  # or individual build:
go build -o build/adapter ./cmd/adapter
```

## Advanced Integration Patterns

### Data Pipeline with External Systems
```yaml
# Data flows from API → Database → Queue → File System
api_source → enrichment → validation → persistence → notification → archival
```

### Event-Driven Integration
```yaml
# Real-time event processing with multiple destinations
stream_input → transformation → [database, queue, file_system]
```

### Batch Processing Integration
```yaml
# Scheduled batch processing with external systems
file_watcher → batch_processor → [api_upload, database_bulk_insert]
```

### Microservices Integration
```yaml
# Service mesh integration with circuit breakers and retries
service_a → [adapter] → service_b → [adapter] → service_c
```

## Security Considerations

- **Credential Management**: Use environment variables and secure credential stores
- **Network Security**: Implement TLS/SSL for all external communications
- **Access Control**: Apply principle of least privilege for external system access
- **Audit Logging**: Log all external system interactions for compliance
- **Data Encryption**: Encrypt sensitive data in transit and at rest