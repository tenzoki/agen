#!/bin/bash

# GOX Framework Search Indexing Demo
# Demonstrates search indexing capabilities with various backends and configurations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOX_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default configuration
INDEX_TYPE=""
INPUT_DIR="${SCRIPT_DIR}/input"
OUTPUT_DIR="/tmp/gox-search-demo/output"
LOG_DIR="/tmp/gox-search-demo/logs"
DEMO_DIR="/tmp/gox-search-demo"
BACKEND="mock"  # Default to mock backend for demo

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
    echo "  --type=TYPE        Run specific indexing type (basic|advanced|realtime|multi|all)"
    echo "  --backend=BACKEND  Search backend (elasticsearch|solr|mock) [default: mock]"
    echo "  --input=DIR        Input directory (default: ./input)"
    echo "  --output=DIR       Output directory (default: /tmp/gox-search-demo/output)"
    echo "  --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Run all indexing types with mock backend"
    echo "  $0 --type=basic            # Run basic indexing only"
    echo "  $0 --backend=elasticsearch # Use Elasticsearch backend"
    echo "  $0 --input=/path/to/docs   # Use custom input directory"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --type=*)
            INDEX_TYPE="${1#*=}"
            shift
            ;;
        --backend=*)
            BACKEND="${1#*=}"
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

    # Check for search indexer binary
    if [[ ! -f "${GOX_ROOT}/build/search_indexer" ]]; then
        log "WARN" "search_indexer binary not found. Some demos may not work."
    fi

    # Check input directory
    if [[ ! -d "$INPUT_DIR" ]]; then
        log "ERROR" "Input directory not found: $INPUT_DIR"
        exit 1
    fi

    # Check backend availability
    check_backend_availability

    log "INFO" "Prerequisites check completed"
}

# Check backend availability
check_backend_availability() {
    case $BACKEND in
        "elasticsearch")
            if ! curl -s -f http://localhost:9200/_cluster/health >/dev/null 2>&1; then
                log "WARN" "Elasticsearch not available at localhost:9200. Falling back to mock backend."
                BACKEND="mock"
            else
                log "INFO" "Elasticsearch backend available"
            fi
            ;;
        "solr")
            if ! curl -s -f http://localhost:8983/solr/admin/info/system >/dev/null 2>&1; then
                log "WARN" "Solr not available at localhost:8983. Falling back to mock backend."
                BACKEND="mock"
            else
                log "INFO" "Solr backend available"
            fi
            ;;
        "mock")
            log "INFO" "Using mock backend for demonstration"
            ;;
        *)
            log "WARN" "Unknown backend: $BACKEND. Using mock backend."
            BACKEND="mock"
            ;;
    esac
}

# Setup demo environment
setup_demo_environment() {
    log "INFO" "Setting up demo environment..."

    # Create demo directories
    mkdir -p "$DEMO_DIR"/{input,output,logs,config,indices}
    mkdir -p "$OUTPUT_DIR"/{basic,advanced,realtime,multi}

    # Copy input files to demo directory
    cp -r "$INPUT_DIR"/* "$DEMO_DIR/input/"

    # Create sample input data if not exists
    create_sample_data

    # Create cell configuration files
    create_cell_configs

    log "INFO" "Demo environment setup completed"
}

# Create sample input data
create_sample_data() {
    local input_dir="$DEMO_DIR/input"

    # Create directories for different data types
    mkdir -p "$input_dir"/{documents,logs,structured-data}

    # Sample document
    cat > "$input_dir/documents/gox-framework-overview.json" << 'EOF'
{
  "id": "doc-001",
  "title": "GOX Framework Overview",
  "content": "The GOX Framework is a powerful agent-based processing system designed for scalable data processing and analysis. It provides a flexible architecture for building complex data pipelines using autonomous agents that can process, transform, and analyze data in real-time.",
  "metadata": {
    "author": "GOX Team",
    "created_date": "2024-09-27T10:00:00Z",
    "file_type": "documentation",
    "file_size": 1567,
    "tags": ["framework", "agents", "processing", "documentation"],
    "category": "technical_docs",
    "language": "en",
    "version": "3.0"
  },
  "extracted_entities": {
    "organizations": ["GOX Framework"],
    "technologies": ["Go", "YAML", "JSON", "Docker"],
    "concepts": ["agent", "pipeline", "cell", "broker", "message"]
  }
}
EOF

    # Sample technical document
    cat > "$input_dir/documents/agent-architecture.json" << 'EOF'
{
  "id": "doc-002",
  "title": "Agent Architecture Guide",
  "content": "GOX agents are autonomous processing units that implement the AgentRunner interface. Each agent has its own configuration, lifecycle management, and message processing capabilities. The BaseAgent provides common functionality while specific agent types implement specialized processing logic for binary analysis, text processing, data transformation, and more.",
  "metadata": {
    "author": "Development Team",
    "created_date": "2024-09-26T15:30:00Z",
    "file_type": "technical_guide",
    "file_size": 2134,
    "tags": ["architecture", "agents", "interfaces", "development"],
    "category": "developer_docs",
    "language": "en",
    "version": "3.0"
  },
  "extracted_entities": {
    "interfaces": ["AgentRunner", "BaseAgent"],
    "methods": ["ProcessMessage", "Init", "Cleanup"],
    "concepts": ["lifecycle", "configuration", "message processing"]
  }
}
EOF

    # Sample security document
    cat > "$input_dir/documents/security-analysis.json" << 'EOF'
{
  "id": "doc-003",
  "title": "Security Analysis Capabilities",
  "content": "The GOX Framework includes comprehensive security analysis capabilities through specialized agents like binary_analyzer and security_scanner. These agents can detect malware signatures, analyze file entropy, identify suspicious patterns, and perform deep security scans on various file types including executables, documents, and archives.",
  "metadata": {
    "author": "Security Team",
    "created_date": "2024-09-25T11:20:00Z",
    "file_type": "security_guide",
    "file_size": 3421,
    "tags": ["security", "malware", "analysis", "scanning"],
    "category": "security_docs",
    "language": "en",
    "version": "2.1",
    "classification": "internal"
  },
  "extracted_entities": {
    "security_tools": ["binary_analyzer", "security_scanner"],
    "threats": ["malware", "suspicious patterns"],
    "file_types": ["executables", "documents", "archives"]
  }
}
EOF

    # Sample log entries
    cat > "$input_dir/logs/processing-log.json" << 'EOF'
{
  "timestamp": "2024-09-27T10:15:32.123Z",
  "level": "INFO",
  "logger": "gox.agent.binary_analyzer",
  "message": "Successfully analyzed binary file",
  "metadata": {
    "file_path": "/data/sample.exe",
    "processing_time": 1250,
    "file_size": 2048576,
    "entropy": 7.2,
    "agent_id": "binary-analyzer-001"
  },
  "labels": {
    "environment": "production",
    "node": "worker-001",
    "pipeline": "security-analysis"
  },
  "extracted_fields": {
    "ip_addresses": [],
    "error_codes": [],
    "file_extensions": [".exe"]
  }
}
EOF

    cat > "$input_dir/logs/error-log.json" << 'EOF'
{
  "timestamp": "2024-09-27T10:20:15.456Z",
  "level": "ERROR",
  "logger": "gox.agent.file_processor",
  "message": "Failed to process corrupted file",
  "metadata": {
    "file_path": "/data/corrupted.bin",
    "error_code": "FILE_CORRUPTION",
    "processing_time": 50,
    "file_size": 0,
    "agent_id": "file-processor-002"
  },
  "labels": {
    "environment": "production",
    "node": "worker-002",
    "pipeline": "data-processing"
  },
  "extracted_fields": {
    "error_codes": ["FILE_CORRUPTION"],
    "file_extensions": [".bin"]
  },
  "stack_trace": "Error processing file: unexpected EOF"
}
EOF

    # Sample structured data
    cat > "$input_dir/structured-data/analysis-result.json" << 'EOF'
{
  "record_id": "rec-001",
  "data_type": "security_analysis_result",
  "timestamp": "2024-09-27T10:00:00Z",
  "source": {
    "system": "gox-analyzer",
    "agent_id": "binary-analyzer-001",
    "pipeline": "security-pipeline"
  },
  "content": {
    "file_analysis": {
      "file_path": "/data/sample.exe",
      "file_type": "PE32 executable",
      "file_hash": "sha256:abc123def456...",
      "security_findings": [
        {
          "type": "high_entropy_section",
          "section": ".text",
          "entropy": 7.8,
          "severity": "medium",
          "description": "Section has high entropy indicating possible packing"
        },
        {
          "type": "suspicious_imports",
          "imports": ["CreateRemoteThread", "WriteProcessMemory"],
          "severity": "high",
          "description": "Imports commonly used in malicious code"
        }
      ],
      "metadata": {
        "compile_time": "2024-09-20T08:30:00Z",
        "size": 2048576,
        "section_count": 4
      }
    }
  },
  "search_metadata": {
    "keywords": ["security", "binary", "analysis", "PE32", "malware"],
    "boost_factor": 2.0,
    "exclude_from_search": false,
    "access_level": "internal"
  }
}
EOF
}

# Create cell configuration files
create_cell_configs() {
    local config_dir="$DEMO_DIR/config"

    # Get backend endpoint
    local endpoint
    case $BACKEND in
        "elasticsearch")
            endpoint="http://localhost:9200"
            ;;
        "solr")
            endpoint="http://localhost:8983/solr"
            ;;
        "mock")
            endpoint="mock://localhost"
            ;;
    esac

    # Basic indexing cell config
    cat > "$config_dir/basic_indexing_cell.yaml" << EOF
cell:
  id: "search:basic-document-indexing"
  description: "Basic document indexing cell"

  agents:
    - id: "search-indexer-001"
      agent_type: "search_indexer"
      ingress: "file:$DEMO_DIR/input/documents/*"
      egress: "index:documents"
      config:
        index_name: "documents"
        backend: "$BACKEND"
        endpoint: "$endpoint"
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
        output_path: "$OUTPUT_DIR/basic/"
EOF

    # Advanced indexing cell config
    cat > "$config_dir/advanced_indexing_cell.yaml" << EOF
cell:
  id: "search:advanced-indexing"
  description: "Advanced indexing with faceted search"

  agents:
    - id: "search-indexer-002"
      agent_type: "search_indexer"
      ingress: "file:$DEMO_DIR/input/documents/*"
      egress: "index:advanced"
      config:
        index_name: "advanced_search"
        backend: "$BACKEND"
        endpoint: "$endpoint"
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
        output_path: "$OUTPUT_DIR/advanced/"
EOF

    # Real-time indexing cell config
    cat > "$config_dir/real_time_indexing_cell.yaml" << EOF
cell:
  id: "search:real-time-indexing"
  description: "Real-time search indexing cell"

  agents:
    - id: "search-indexer-003"
      agent_type: "search_indexer"
      ingress: "file:$DEMO_DIR/input/logs/*"
      egress: "index:realtime"
      config:
        index_name: "realtime_search"
        backend: "$BACKEND"
        endpoint: "$endpoint"
        indexing_mode: "real_time"
        batch_size: 100
        flush_interval: 1000
        refresh_interval: "1s"
        mapping:
          timestamp:
            type: "date"
          level:
            type: "keyword"
          logger:
            type: "keyword"
          message:
            type: "text"
          metadata:
            type: "object"
        time_based_indices: true
        index_pattern: "logs-{yyyy.MM.dd}"
        output_path: "$OUTPUT_DIR/realtime/"
EOF

    # Multi-source indexing cell config
    cat > "$config_dir/multi_source_indexing_cell.yaml" << EOF
cell:
  id: "search:multi-source-indexing"
  description: "Multi-source indexing pipeline"

  agents:
    - id: "data-router-001"
      agent_type: "strategy_selector"
      ingress: "file:$DEMO_DIR/input/**/*"
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
        backend: "$BACKEND"
        endpoint: "$endpoint"
        mapping_template: "document_template"
        output_path: "$OUTPUT_DIR/multi/documents/"

    - id: "log-indexer-001"
      agent_type: "search_indexer"
      ingress: "sub:logs"
      egress: "index:logs"
      config:
        index_name: "logs"
        backend: "$BACKEND"
        endpoint: "$endpoint"
        mapping_template: "log_template"
        time_based_indices: true
        index_pattern: "logs-{yyyy.MM.dd}"
        output_path: "$OUTPUT_DIR/multi/logs/"

    - id: "structured-indexer-001"
      agent_type: "search_indexer"
      ingress: "sub:structured"
      egress: "index:structured"
      config:
        index_name: "structured_data"
        backend: "$BACKEND"
        endpoint: "$endpoint"
        mapping_template: "structured_template"
        nested_field_support: true
        output_path: "$OUTPUT_DIR/multi/structured/"
EOF
}

# Run specific indexing type
run_indexing_type() {
    local index_type=$1
    local config_file="$DEMO_DIR/config/${index_type}_indexing_cell.yaml"

    log "INFO" "Running $index_type indexing..."

    if [[ ! -f "$config_file" ]]; then
        log "ERROR" "Configuration file not found: $config_file"
        return 1
    fi

    # Start the indexing
    log "DEBUG" "Starting GOX with configuration: $config_file"
    cd "$GOX_ROOT"

    # Run in background and capture PID for cleanup
    ./build/gox cell run "$config_file" > "$LOG_DIR/${index_type}_indexing.log" 2>&1 &
    local gox_pid=$!

    # Give it time to process
    sleep 5

    # Check if process is still running (indicates success)
    if kill -0 $gox_pid 2>/dev/null; then
        log "INFO" "$index_type indexing started successfully (PID: $gox_pid)"

        # Wait for processing to complete (check for output files)
        local timeout=30
        local elapsed=0
        local expected_output="$OUTPUT_DIR/$index_type"

        while [[ $elapsed -lt $timeout ]]; do
            if [[ -d "$expected_output" && $(find "$expected_output" -type f 2>/dev/null | wc -l) -gt 0 ]]; then
                log "INFO" "$index_type indexing completed"
                break
            fi
            sleep 2
            elapsed=$((elapsed + 2))
        done

        # Stop the process
        kill $gox_pid 2>/dev/null || true
        wait $gox_pid 2>/dev/null || true
    else
        log "ERROR" "$index_type indexing failed to start"
        return 1
    fi
}

# Display results
display_results() {
    log "INFO" "Search Indexing Results Summary:"
    echo ""

    # Check each output directory
    for index_type in basic advanced realtime multi; do
        local output_path="$OUTPUT_DIR/$index_type"
        if [[ -d "$output_path" ]]; then
            local file_count=$(find "$output_path" -type f 2>/dev/null | wc -l)

            if [[ $file_count -gt 0 ]]; then
                echo -e "${GREEN}✅ ${index_type^} Indexing:${NC} $file_count index files created"

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
                echo -e "${YELLOW}⚠️  ${index_type^} Indexing:${NC} No output files generated"
            fi
        else
            echo -e "${RED}❌ ${index_type^} Indexing:${NC} Output directory not found"
        fi
        echo ""
    done

    # Show backend information
    echo -e "${BLUE}🔍 Search Backend:${NC} $BACKEND"
    case $BACKEND in
        "elasticsearch")
            echo "   Endpoint: http://localhost:9200"
            if curl -s -f http://localhost:9200/_cluster/health >/dev/null 2>&1; then
                echo "   Status: Available"
            else
                echo "   Status: Unavailable"
            fi
            ;;
        "solr")
            echo "   Endpoint: http://localhost:8983/solr"
            if curl -s -f http://localhost:8983/solr/admin/info/system >/dev/null 2>&1; then
                echo "   Status: Available"
            else
                echo "   Status: Unavailable"
            fi
            ;;
        "mock")
            echo "   Type: Mock/Demonstration Backend"
            echo "   Status: Active"
            ;;
    esac
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

    # Optionally remove demo directory (uncomment if desired)
    # rm -rf "$DEMO_DIR"

    log "INFO" "Demo completed. Output files available in: $OUTPUT_DIR"
    log "INFO" "Logs available in: $LOG_DIR"
    log "INFO" "Search backend used: $BACKEND"
}

# Main demo execution
main() {
    echo "=== GOX Framework Search Indexing Demo ==="
    echo ""

    # Setup
    check_prerequisites
    setup_demo_environment

    # Trap cleanup on exit
    trap cleanup EXIT

    # Run indexing based on user selection
    case "$INDEX_TYPE" in
        "basic")
            run_indexing_type "basic"
            ;;
        "advanced")
            run_indexing_type "advanced"
            ;;
        "realtime")
            run_indexing_type "real_time"
            ;;
        "multi")
            run_indexing_type "multi_source"
            ;;
        "all"|"")
            log "INFO" "Running complete search indexing pipeline..."

            # Run each indexing type
            for index_type in basic advanced real_time multi_source; do
                if ! run_indexing_type "$index_type"; then
                    log "WARN" "Failed to run $index_type indexing, continuing..."
                fi
            done
            ;;
        *)
            log "ERROR" "Unknown indexing type: $INDEX_TYPE"
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