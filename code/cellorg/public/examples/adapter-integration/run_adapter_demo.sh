#!/bin/bash

# GOX Framework Adapter Integration Demo
# Demonstrates integration with external systems through various adapter types

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOX_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default configuration
ADAPTER_TYPE=""
INPUT_DIR="${SCRIPT_DIR}/input"
OUTPUT_DIR="/tmp/gox-adapter-demo/output"
LOG_DIR="/tmp/gox-adapter-demo/logs"
DEMO_DIR="/tmp/gox-adapter-demo"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  --adapter=TYPE     Run specific adapter (api|database|queue|filesystem|streaming|all)"
    echo "  --input=DIR        Input directory (default: ./input)"
    echo "  --output=DIR       Output directory (default: /tmp/gox-adapter-demo/output)"
    echo "  --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Run all adapter types"
    echo "  $0 --adapter=api           # Run API adapter only"
    echo "  $0 --input=/path/to/data   # Use custom input directory"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --adapter=*)
            ADAPTER_TYPE="${1#*=}"
            shift
            ;;
        --input=*)
            INPUT_DIR="${1#*=}"
            shift
            ;;
        --output=*)
            OUTPUT_DIR="${1#*=}"
            shift
            ;;
        --help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    case $level in
        "INFO")
            echo -e "${GREEN}[INFO]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/demo.log"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/demo.log"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/demo.log"
            ;;
        "DEBUG")
            echo -e "${BLUE}[DEBUG]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/demo.log"
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    log "INFO" "Checking prerequisites..."

    # Check if GOX binary exists
    if [[ ! -f "${GOX_ROOT}/build/gox" ]]; then
        log "ERROR" "GOX binary not found. Please run 'make build' from the root directory."
        exit 1
    fi

    # Check for adapter binary
    if [[ ! -f "${GOX_ROOT}/build/adapter" ]]; then
        log "WARN" "adapter binary not found. Some demos may not work."
    fi

    # Check input directory
    if [[ ! -d "$INPUT_DIR" ]]; then
        log "ERROR" "Input directory not found: $INPUT_DIR"
        exit 1
    fi

    # Check external service availability
    check_external_services

    log "INFO" "Prerequisites check completed"
}

# Check external service availability
check_external_services() {
    log "INFO" "Checking external service availability..."

    # Check for common services (optional)
    services=(
        "http://httpbin.org/status/200:HTTP Test Service"
        "postgresql://localhost:5432:PostgreSQL"
        "amqp://localhost:5672:RabbitMQ"
        "kafka://localhost:9092:Apache Kafka"
    )

    for service in "${services[@]}"; do
        IFS=":" read -r url name <<< "$service"
        case $url in
            http*)
                if curl -s -f "$url" >/dev/null 2>&1; then
                    log "DEBUG" "$name is available"
                else
                    log "DEBUG" "$name is not available (using mock)"
                fi
                ;;
            postgresql*)
                if command -v psql >/dev/null 2>&1; then
                    log "DEBUG" "PostgreSQL client available"
                else
                    log "DEBUG" "PostgreSQL client not found (using mock)"
                fi
                ;;
            amqp*)
                if command -v rabbitmqctl >/dev/null 2>&1; then
                    log "DEBUG" "RabbitMQ tools available"
                else
                    log "DEBUG" "RabbitMQ not found (using mock)"
                fi
                ;;
            *)
                log "DEBUG" "$name availability unknown (using mock)"
                ;;
        esac
    done
}

# Setup demo environment
setup_demo_environment() {
    log "INFO" "Setting up demo environment..."

    # Create demo directories
    mkdir -p "$DEMO_DIR"/{input,output,logs,config,mock-services}
    mkdir -p "$OUTPUT_DIR"/{api,database,queue,filesystem,streaming,multi}

    # Copy input files to demo directory
    cp -r "$INPUT_DIR"/* "$DEMO_DIR/input/"

    # Create sample input data if not exists
    create_sample_data

    # Create cell configuration files
    create_cell_configs

    # Start mock services
    start_mock_services

    log "INFO" "Demo environment setup completed"
}

# Create sample input data
create_sample_data() {
    local input_dir="$DEMO_DIR/input"

    # Create directories for different adapter types
    mkdir -p "$input_dir"/{api-configs,database-schemas,message-samples,file-samples}

    # Sample API request
    cat > "$input_dir/api-configs/user-creation.json" << 'EOF'
{
  "request_id": "req-001",
  "method": "POST",
  "endpoint": "/users",
  "headers": {
    "Content-Type": "application/json",
    "Accept": "application/json"
  },
  "body": {
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "department": "engineering"
  },
  "timeout": 30,
  "retries": 3
}
EOF

    # Sample database query
    cat > "$input_dir/database-schemas/user-query.json" << 'EOF'
{
  "query_id": "q-001",
  "operation": "select",
  "table": "users",
  "query": "SELECT id, name, email, created_at FROM users WHERE department = $1 ORDER BY created_at DESC LIMIT $2",
  "parameters": ["engineering", 50],
  "timeout": 30,
  "fetch_size": 1000
}
EOF

    # Sample message
    cat > "$input_dir/message-samples/user-event.json" << 'EOF'
{
  "message_id": "msg-001",
  "timestamp": "2024-09-27T10:00:00Z",
  "type": "user_event",
  "payload": {
    "user_id": "user-123",
    "action": "login",
    "success": true,
    "metadata": {
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
      "session_id": "sess-789",
      "location": "New York, US"
    }
  },
  "routing_key": "events.user.login",
  "priority": "normal"
}
EOF

    # Sample file for processing
    cat > "$input_dir/file-samples/data-export.csv" << 'EOF'
id,name,email,department,created_at
1,Alice Smith,alice@example.com,engineering,2024-09-25T08:30:00Z
2,Bob Johnson,bob@example.com,marketing,2024-09-25T09:15:00Z
3,Carol Williams,carol@example.com,engineering,2024-09-25T10:45:00Z
4,David Brown,david@example.com,sales,2024-09-25T11:20:00Z
5,Eve Davis,eve@example.com,engineering,2024-09-25T14:00:00Z
EOF

    # Sample streaming data
    cat > "$input_dir/file-samples/stream-events.jsonl" << 'EOF'
{"timestamp":"2024-09-27T10:00:00Z","event_type":"page_view","user_id":"user-001","page":"/home","duration":2500}
{"timestamp":"2024-09-27T10:00:01Z","event_type":"click","user_id":"user-001","element":"nav-menu","page":"/home"}
{"timestamp":"2024-09-27T10:00:02Z","event_type":"page_view","user_id":"user-002","page":"/products","duration":1800}
{"timestamp":"2024-09-27T10:00:03Z","event_type":"search","user_id":"user-002","query":"laptop","results":25}
{"timestamp":"2024-09-27T10:00:04Z","event_type":"click","user_id":"user-002","element":"product-link","page":"/products"}
EOF
}

# Start mock services
start_mock_services() {
    local mock_dir="$DEMO_DIR/mock-services"

    # Create mock HTTP service
    cat > "$mock_dir/http-mock.py" << 'EOF'
#!/usr/bin/env python3
import http.server
import socketserver
import json
import datetime
from urllib.parse import urlparse, parse_qs

class MockHTTPHandler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/users':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            request_data = json.loads(post_data.decode('utf-8'))

            response = {
                "id": 456,
                "name": request_data.get("name", "Unknown"),
                "email": request_data.get("email", "unknown@example.com"),
                "department": request_data.get("department", "unknown"),
                "created_at": datetime.datetime.now().isoformat() + "Z"
            }

            self.send_response(201)
            self.send_header('Content-Type', 'application/json')
            self.send_header('Location', f'/users/{response["id"]}')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode('utf-8'))
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        pass  # Suppress default logging

if __name__ == "__main__":
    PORT = 8080
    with socketserver.TCPServer(("", PORT), MockHTTPHandler) as httpd:
        print(f"Mock HTTP service running on port {PORT}")
        httpd.serve_forever()
EOF

    chmod +x "$mock_dir/http-mock.py"

    # Start mock HTTP service in background
    if command -v python3 >/dev/null 2>&1; then
        python3 "$mock_dir/http-mock.py" >/dev/null 2>&1 &
        echo $! > "$mock_dir/http-mock.pid"
        log "DEBUG" "Started mock HTTP service on port 8080"
    fi
}

# Create cell configuration files
create_cell_configs() {
    local config_dir="$DEMO_DIR/config"

    # API adapter cell config
    cat > "$config_dir/api_adapter_cell.yaml" << EOF
cell:
  id: "integration:api-adapter"
  description: "External API integration cell"

  agents:
    - id: "api-adapter-001"
      agent_type: "adapter"
      adapter_type: "http_api"
      ingress: "file:$DEMO_DIR/input/api-configs/*"
      egress: "file:$OUTPUT_DIR/api/"
      config:
        endpoint: "http://localhost:8080"
        authentication:
          type: "none"
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
        mock_mode: true
EOF

    # Database adapter cell config
    cat > "$config_dir/database_adapter_cell.yaml" << EOF
cell:
  id: "integration:database-adapter"
  description: "Database integration cell"

  agents:
    - id: "db-adapter-001"
      agent_type: "adapter"
      adapter_type: "database"
      ingress: "file:$DEMO_DIR/input/database-schemas/*"
      egress: "file:$OUTPUT_DIR/database/"
      config:
        connection:
          driver: "mock"
          database: "gox_demo"
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
        mock_mode: true
        mock_data:
          users:
            - id: 1
              name: "Alice Smith"
              email: "alice@example.com"
              department: "engineering"
              created_at: "2024-09-25T08:30:00Z"
            - id: 2
              name: "Bob Johnson"
              email: "bob@example.com"
              department: "engineering"
              created_at: "2024-09-25T09:15:00Z"
EOF

    # Queue adapter cell config
    cat > "$config_dir/queue_adapter_cell.yaml" << EOF
cell:
  id: "integration:queue-adapter"
  description: "Message queue integration cell"

  agents:
    - id: "queue-adapter-001"
      agent_type: "adapter"
      adapter_type: "message_queue"
      ingress: "file:$DEMO_DIR/input/message-samples/*"
      egress: "file:$OUTPUT_DIR/queue/"
      config:
        broker:
          type: "mock"
          url: "mock://localhost"
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
        mock_mode: true
EOF

    # File system adapter cell config
    cat > "$config_dir/filesystem_adapter_cell.yaml" << EOF
cell:
  id: "integration:filesystem-adapter"
  description: "File system integration cell"

  agents:
    - id: "fs-adapter-001"
      agent_type: "adapter"
      adapter_type: "filesystem"
      ingress: "file:$DEMO_DIR/input/file-samples/*"
      egress: "file:$OUTPUT_DIR/filesystem/"
      config:
        source:
          type: "local"
          path: "$DEMO_DIR/input/file-samples"
          patterns: ["*.csv", "*.json", "*.jsonl"]
          recursive: false
        destination:
          type: "local"
          path: "$OUTPUT_DIR/filesystem/processed"
        processing:
          batch_size: 100
          compression: false
          move_processed: false
          archive_path: "$OUTPUT_DIR/filesystem/archive"
        mock_mode: true
EOF

    # Streaming adapter cell config
    cat > "$config_dir/streaming_adapter_cell.yaml" << EOF
cell:
  id: "integration:streaming-adapter"
  description: "Real-time streaming integration cell"

  agents:
    - id: "stream-adapter-001"
      agent_type: "adapter"
      adapter_type: "streaming"
      ingress: "file:$DEMO_DIR/input/file-samples/stream-events.jsonl"
      egress: "file:$OUTPUT_DIR/streaming/"
      config:
        source:
          type: "file"
          format: "jsonl"
          simulate_stream: true
          events_per_second: 2
        destination:
          type: "file"
          format: "json"
        processing:
          window_size: "5s"
          window_type: "tumbling"
          watermark_delay: "1s"
          parallelism: 2
        stream_config:
          buffer_size: 100
          flush_interval: "1s"
          checkpointing: false
        mock_mode: true
EOF

    # Multi-adapter pipeline config
    cat > "$config_dir/multi_adapter_cell.yaml" << EOF
cell:
  id: "integration:multi-adapter-pipeline"
  description: "Multi-system integration pipeline"

  agents:
    - id: "data-router-001"
      agent_type: "strategy_selector"
      ingress: "file:$DEMO_DIR/input/**/*"
      egress: "route:adapter-type"
      config:
        routing_strategy: "file_path"
        routes:
          api-configs: "pub:api-requests"
          database-schemas: "pub:db-queries"
          message-samples: "pub:messages"
          file-samples: "pub:files"

    - id: "api-processor-001"
      agent_type: "adapter"
      adapter_type: "http_api"
      ingress: "sub:api-requests"
      egress: "pub:api-results"
      config:
        endpoint: "http://localhost:8080"
        mock_mode: true

    - id: "db-processor-001"
      agent_type: "adapter"
      adapter_type: "database"
      ingress: "sub:db-queries"
      egress: "pub:db-results"
      config:
        connection:
          driver: "mock"
        mock_mode: true

    - id: "results-aggregator-001"
      agent_type: "report_generator"
      ingress: "sub:api-results,db-results"
      egress: "file:$OUTPUT_DIR/multi/integration-report.json"
      config:
        report_format: "json"
        include_statistics: true
EOF
}

# Run specific adapter
run_adapter() {
    local adapter_type=$1
    local config_file="$DEMO_DIR/config/${adapter_type}_adapter_cell.yaml"

    log "INFO" "Running $adapter_type adapter..."

    if [[ ! -f "$config_file" ]]; then
        log "ERROR" "Configuration file not found: $config_file"
        return 1
    fi

    # Start the adapter
    log "DEBUG" "Starting GOX with configuration: $config_file"
    cd "$GOX_ROOT"

    # Run in background and capture PID for cleanup
    ./build/gox cell run "$config_file" > "$LOG_DIR/${adapter_type}_adapter.log" 2>&1 &
    local gox_pid=$!

    # Give it time to process
    sleep 5

    # Check if process is still running (indicates success)
    if kill -0 $gox_pid 2>/dev/null; then
        log "INFO" "$adapter_type adapter started successfully (PID: $gox_pid)"

        # Wait for processing to complete (check for output files)
        local timeout=30
        local elapsed=0
        local expected_output="$OUTPUT_DIR/$adapter_type"

        while [[ $elapsed -lt $timeout ]]; do
            if [[ -d "$expected_output" && $(find "$expected_output" -type f 2>/dev/null | wc -l) -gt 0 ]]; then
                log "INFO" "$adapter_type adapter processing completed"
                break
            fi
            sleep 2
            elapsed=$((elapsed + 2))
        done

        # Stop the process
        kill $gox_pid 2>/dev/null || true
        wait $gox_pid 2>/dev/null || true
    else
        log "ERROR" "$adapter_type adapter failed to start"
        return 1
    fi
}

# Display results
display_results() {
    log "INFO" "Adapter Integration Results Summary:"
    echo ""

    # Check each output directory
    for adapter_type in api database queue filesystem streaming multi; do
        local output_path="$OUTPUT_DIR/$adapter_type"
        if [[ -d "$output_path" ]]; then
            local file_count=$(find "$output_path" -type f 2>/dev/null | wc -l)

            if [[ $file_count -gt 0 ]]; then
                echo -e "${GREEN}✅ ${adapter_type^} Adapter:${NC} $file_count output files created"

                # Show sample results
                local sample_files=($(find "$output_path" -type f | head -3))
                for sample_file in "${sample_files[@]}"; do
                    echo "   Created: $(basename "$sample_file")"

                    # Show preview for JSON files
                    if [[ "$sample_file" == *.json ]] && command -v jq >/dev/null 2>&1; then
                        echo "   Preview:"
                        jq -C '.' < "$sample_file" 2>/dev/null | head -5 | sed 's/^/     /'
                        echo "     ..."
                    fi
                done
            else
                echo -e "${YELLOW}⚠️  ${adapter_type^} Adapter:${NC} No output files generated"
            fi
        else
            echo -e "${RED}❌ ${adapter_type^} Adapter:${NC} Output directory not found"
        fi
        echo ""
    done

    # Show integration statistics
    echo -e "${BLUE}📊 Integration Statistics:${NC}"
    local total_files=$(find "$OUTPUT_DIR" -type f 2>/dev/null | wc -l)
    echo "   Total output files: $total_files"
    echo "   Demo directory: $DEMO_DIR"
    echo "   Mock services used: HTTP API (port 8080)"
    echo ""

    # Show log summary
    echo -e "${BLUE}📋 Processing Logs:${NC}"
    if [[ -f "$LOG_DIR/demo.log" ]]; then
        echo "   Main log: $LOG_DIR/demo.log"
        local error_count=$(grep -c "ERROR" "$LOG_DIR/demo.log" 2>/dev/null || echo "0")
        local warn_count=$(grep -c "WARN" "$LOG_DIR/demo.log" 2>/dev/null || echo "0")
        echo "   Errors: $error_count, Warnings: $warn_count"
    fi
}

# Cleanup function
cleanup() {
    log "INFO" "Cleaning up demo environment..."

    # Kill any remaining GOX processes
    pkill -f "gox.*cell.*run" 2>/dev/null || true

    # Stop mock services
    if [[ -f "$DEMO_DIR/mock-services/http-mock.pid" ]]; then
        local http_pid=$(cat "$DEMO_DIR/mock-services/http-mock.pid")
        kill $http_pid 2>/dev/null || true
        rm -f "$DEMO_DIR/mock-services/http-mock.pid"
        log "DEBUG" "Stopped mock HTTP service"
    fi

    # Optionally remove demo directory (uncomment if desired)
    # rm -rf "$DEMO_DIR"

    log "INFO" "Demo completed. Output files available in: $OUTPUT_DIR"
    log "INFO" "Logs available in: $LOG_DIR"
}

# Main demo execution
main() {
    echo "=== GOX Framework Adapter Integration Demo ==="
    echo ""

    # Setup
    check_prerequisites
    setup_demo_environment

    # Trap cleanup on exit
    trap cleanup EXIT

    # Run adapters based on user selection
    case "$ADAPTER_TYPE" in
        "api")
            run_adapter "api"
            ;;
        "database")
            run_adapter "database"
            ;;
        "queue")
            run_adapter "queue"
            ;;
        "filesystem")
            run_adapter "filesystem"
            ;;
        "streaming")
            run_adapter "streaming"
            ;;
        "all"|"")
            log "INFO" "Running complete adapter integration pipeline..."

            # Run each adapter type
            for adapter in api database queue filesystem streaming; do
                if ! run_adapter "$adapter"; then
                    log "WARN" "Failed to run $adapter adapter, continuing..."
                fi
            done

            # Run multi-adapter pipeline
            if ! run_adapter "multi"; then
                log "WARN" "Failed to run multi-adapter pipeline"
            fi
            ;;
        *)
            log "ERROR" "Unknown adapter type: $ADAPTER_TYPE"
            usage
            ;;
    esac

    # Display results
    echo ""
    display_results

    echo ""
    echo "=== Demo Completed Successfully ==="
}

# Run the demo
main "$@"