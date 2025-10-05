#!/bin/bash

# GOX Framework Chunk Processing Demo
# Demonstrates chunk writing and processing capabilities

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOX_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default configuration
CHUNK_TYPE=""
INPUT_DIR="${SCRIPT_DIR}/input"
OUTPUT_DIR="/tmp/gox-chunk-demo/output"
LOG_DIR="/tmp/gox-chunk-demo/logs"
DEMO_DIR="/tmp/gox-chunk-demo"

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
    echo "  --type=TYPE        Run specific chunk processing (basic|compressed|parallel|streaming|all)"
    echo "  --input=DIR        Input directory (default: ./input)"
    echo "  --output=DIR       Output directory (default: /tmp/gox-chunk-demo/output)"
    echo "  --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Run all chunk processing types"
    echo "  $0 --type=basic            # Run basic chunk writing only"
    echo "  $0 --input=/path/to/files  # Use custom input directory"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --type=*)
            CHUNK_TYPE="${1#*=}"
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

    # Check for chunk processing binaries
    local agents=("chunk_writer" "file_splitter" "chunk_processor" "chunk_synthesizer")
    for agent in "${agents[@]}"; do
        if [[ ! -f "${GOX_ROOT}/build/${agent}" ]]; then
            log "WARN" "${agent} binary not found. Some demos may not work."
        fi
    done

    # Check input directory
    if [[ ! -d "$INPUT_DIR" ]]; then
        log "ERROR" "Input directory not found: $INPUT_DIR"
        exit 1
    fi

    log "INFO" "Prerequisites check completed"
}

# Setup demo environment
setup_demo_environment() {
    log "INFO" "Setting up demo environment..."

    # Create demo directories
    mkdir -p "$DEMO_DIR"/{input,output,logs,config,temp}
    mkdir -p "$OUTPUT_DIR"/{basic,compressed,parallel,streaming,complete}

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
    mkdir -p "$input_dir"/{large-files,stream-data,chunk-configs}

    # Create a large sample JSON file
    log "DEBUG" "Creating large sample files..."
    cat > "$input_dir/large-files/sample-dataset.json" << 'EOF'
{
  "metadata": {
    "dataset_name": "GOX Framework Sample Dataset",
    "version": "1.0",
    "created_at": "2024-09-27T10:00:00Z",
    "total_records": 10000,
    "description": "Sample dataset for chunk processing demonstration"
  },
  "records": [
EOF

    # Generate sample records
    for i in {1..1000}; do
        if [[ $i -eq 1000 ]]; then
            # Last record without comma
            cat >> "$input_dir/large-files/sample-dataset.json" << EOF
    {
      "id": $i,
      "name": "Record $i",
      "timestamp": "2024-09-27T$(printf "%02d" $((10 + (i % 14)))):$(printf "%02d" $((i % 60))):$(printf "%02d" $((i % 60)))Z",
      "data": {
        "value": $((i * 42)),
        "category": "category_$((i % 10))",
        "tags": ["tag_$((i % 5))", "tag_$((i % 7))", "tag_$((i % 3))"],
        "metadata": {
          "processed": $((i % 2 == 0)),
          "score": $((i % 100)),
          "description": "This is a sample record number $i with some additional data to make the file larger for chunk processing demonstration purposes."
        }
      }
    }
EOF
        else
            cat >> "$input_dir/large-files/sample-dataset.json" << EOF
    {
      "id": $i,
      "name": "Record $i",
      "timestamp": "2024-09-27T$(printf "%02d" $((10 + (i % 14)))):$(printf "%02d" $((i % 60))):$(printf "%02d" $((i % 60)))Z",
      "data": {
        "value": $((i * 42)),
        "category": "category_$((i % 10))",
        "tags": ["tag_$((i % 5))", "tag_$((i % 7))", "tag_$((i % 3))"],
        "metadata": {
          "processed": $((i % 2 == 0)),
          "score": $((i % 100)),
          "description": "This is a sample record number $i with some additional data to make the file larger for chunk processing demonstration purposes."
        }
      }
    },
EOF
        fi
    done

    cat >> "$input_dir/large-files/sample-dataset.json" << 'EOF'
  ]
}
EOF

    # Create a large CSV file
    cat > "$input_dir/large-files/sample-data.csv" << 'EOF'
id,name,email,department,salary,hire_date,status
EOF

    for i in {1..2000}; do
        echo "$i,Employee $i,employee$i@company.com,dept_$((i % 10)),$((30000 + (i * 500))),2024-$(printf "%02d" $((1 + (i % 12))))-$(printf "%02d" $((1 + (i % 28)))),active" >> "$input_dir/large-files/sample-data.csv"
    done

    # Create streaming data samples
    cat > "$input_dir/stream-data/events.jsonl" << 'EOF'
{"timestamp":"2024-09-27T10:00:00Z","event_type":"user_action","user_id":"user001","action":"login","metadata":{"ip":"192.168.1.100","user_agent":"Mozilla/5.0"}}
{"timestamp":"2024-09-27T10:00:01Z","event_type":"page_view","user_id":"user001","page":"/dashboard","duration":2500}
{"timestamp":"2024-09-27T10:00:02Z","event_type":"user_action","user_id":"user002","action":"login","metadata":{"ip":"192.168.1.101","user_agent":"Chrome/91.0"}}
{"timestamp":"2024-09-27T10:00:03Z","event_type":"api_call","user_id":"user001","endpoint":"/api/data","response_time":150}
{"timestamp":"2024-09-27T10:00:04Z","event_type":"error","user_id":"user002","error_code":"AUTH_FAILED","message":"Invalid credentials"}
EOF

    # Duplicate events to make it larger
    for i in {1..100}; do
        sed "s/10:00:0/10:$(printf "%02d"):$(printf "%02d")/g" "$input_dir/stream-data/events.jsonl" >> "$input_dir/stream-data/events.jsonl.tmp"
    done
    mv "$input_dir/stream-data/events.jsonl.tmp" "$input_dir/stream-data/events.jsonl"

    # Create chunk configuration
    cat > "$input_dir/chunk-configs/large-file-config.json" << 'EOF'
{
  "file_id": "large-dataset-001",
  "file_path": "sample-dataset.json",
  "chunk_preferences": {
    "preferred_chunk_size": "1MB",
    "compression": "gzip",
    "format": "json"
  },
  "processing_hints": {
    "parallelizable": true,
    "content_type": "structured",
    "preserve_structure": true
  },
  "output_preferences": {
    "naming_pattern": "dataset_chunk_{index:04d}.json",
    "include_metadata": true,
    "verify_integrity": true
  }
}
EOF

    log "DEBUG" "Sample data creation completed"
}

# Create cell configuration files
create_cell_configs() {
    local config_dir="$DEMO_DIR/config"

    # Basic chunk writer cell config
    cat > "$config_dir/basic_chunk_writer_cell.yaml" << EOF
cell:
  id: "processing:basic-chunk-writer"
  description: "Basic chunk writing cell"

  agents:
    - id: "chunk-writer-001"
      agent_type: "chunk_writer"
      ingress: "file:$DEMO_DIR/input/large-files/*"
      egress: "file:$OUTPUT_DIR/basic/"
      config:
        chunk_size: "512KB"
        output_format: "json"
        naming_pattern: "chunk_{index:04d}.json"
        preserve_metadata: true
        validation:
          enable_checksums: true
          verify_integrity: true
        error_handling:
          retry_attempts: 3
          retry_delay: "1s"
        output_config:
          include_manifest: true
          manifest_file: "chunks.manifest.json"
EOF

    # Compressed chunk writer cell config
    cat > "$config_dir/compressed_chunk_writer_cell.yaml" << EOF
cell:
  id: "processing:compressed-chunk-writer"
  description: "Compressed chunk writing cell"

  agents:
    - id: "chunk-writer-002"
      agent_type: "chunk_writer"
      ingress: "file:$DEMO_DIR/input/large-files/*"
      egress: "file:$OUTPUT_DIR/compressed/"
      config:
        chunk_size: "1MB"
        output_format: "binary"
        compression:
          algorithm: "gzip"
          level: 6
          enable: true
        naming_pattern: "chunk_{timestamp}_{index:06d}.gz"
        metadata:
          include_original_size: true
          include_compression_ratio: true
          include_checksums: true
        optimization:
          buffer_size: "256KB"
          io_concurrency: 2
        output_config:
          include_manifest: true
          manifest_file: "compressed_chunks.manifest.json"
EOF

    # Parallel chunk writer cell config
    cat > "$config_dir/parallel_chunk_writer_cell.yaml" << EOF
cell:
  id: "processing:parallel-chunk-writer"
  description: "Parallel chunk writing cell"

  agents:
    - id: "chunk-writer-003"
      agent_type: "chunk_writer"
      ingress: "file:$DEMO_DIR/input/large-files/*"
      egress: "file:$OUTPUT_DIR/parallel/"
      config:
        chunk_size: "256KB"
        output_format: "json"
        parallel_processing:
          enabled: true
          worker_count: 4
          queue_size: 50
          load_balancing: "round_robin"
        naming_pattern: "worker_{worker_id}/chunk_{index:08d}.json"
        synchronization:
          enable_barriers: true
          checkpoint_interval: 50
        performance:
          memory_limit: "512MB"
          disk_buffer_size: "50MB"
        output_config:
          include_manifest: true
          manifest_file: "parallel_chunks.manifest.json"
EOF

    # Streaming chunk writer cell config
    cat > "$config_dir/streaming_chunk_writer_cell.yaml" << EOF
cell:
  id: "processing:streaming-chunk-writer"
  description: "Streaming chunk writing cell"

  agents:
    - id: "chunk-writer-004"
      agent_type: "chunk_writer"
      ingress: "file:$DEMO_DIR/input/stream-data/*"
      egress: "file:$OUTPUT_DIR/streaming/"
      config:
        chunk_strategy: "time_based"
        time_window: "30s"
        max_chunk_size: "2MB"
        output_format: "jsonl"
        streaming:
          buffer_size: "1MB"
          flush_interval: "10s"
          max_buffered_chunks: 5
        naming_pattern: "stream_{timestamp}.jsonl"
        rotation:
          enable: true
          max_file_size: "10MB"
          max_file_age: "5m"
        real_time_processing: true
        output_config:
          include_manifest: true
          manifest_file: "stream_chunks.manifest.json"
EOF

    # Complete chunk processing pipeline config
    cat > "$config_dir/complete_chunk_pipeline_cell.yaml" << EOF
cell:
  id: "processing:complete-chunk-pipeline"
  description: "Complete chunk processing pipeline"

  agents:
    - id: "file-splitter-001"
      agent_type: "file_splitter"
      ingress: "file:$DEMO_DIR/input/large-files/*"
      egress: "pub:raw-chunks"
      config:
        chunk_size: "512KB"
        overlap_size: "64KB"
        split_strategy: "content_aware"
        output_format: "json"

    - id: "chunk-processor-001"
      agent_type: "chunk_processor"
      ingress: "sub:raw-chunks"
      egress: "pub:processed-chunks"
      config:
        processing_type: "transform"
        transformation:
          format_conversion: true
          data_cleaning: true
          validation: true
        batch_size: 10

    - id: "chunk-writer-001"
      agent_type: "chunk_writer"
      ingress: "sub:processed-chunks"
      egress: "file:$OUTPUT_DIR/complete/final-chunks/"
      config:
        chunk_size: "1MB"
        output_format: "binary"
        compression:
          algorithm: "lz4"
          enable: true
        validation:
          enable_checksums: true
          cross_chunk_validation: true
        output_config:
          include_manifest: true
          manifest_file: "complete_chunks.manifest.json"

    - id: "chunk-synthesizer-001"
      agent_type: "chunk_synthesizer"
      ingress: "file:$OUTPUT_DIR/complete/final-chunks/*"
      egress: "file:$OUTPUT_DIR/complete/synthesized/"
      config:
        synthesis_strategy: "merge"
        output_format: "original"
        verify_completeness: true
        output_config:
          generate_report: true
          report_file: "synthesis_report.json"
EOF
}

# Run specific chunk processing type
run_chunk_type() {
    local chunk_type=$1
    local config_file="$DEMO_DIR/config/${chunk_type}_cell.yaml"

    log "INFO" "Running $chunk_type chunk processing..."

    if [[ ! -f "$config_file" ]]; then
        log "ERROR" "Configuration file not found: $config_file"
        return 1
    fi

    # Start the chunk processing
    log "DEBUG" "Starting GOX with configuration: $config_file"
    cd "$GOX_ROOT"

    # Run in background and capture PID for cleanup
    ./build/gox cell run "$config_file" > "$LOG_DIR/${chunk_type}_chunk.log" 2>&1 &
    local gox_pid=$!

    # Give it time to process
    sleep 8

    # Check if process is still running (indicates success)
    if kill -0 $gox_pid 2>/dev/null; then
        log "INFO" "$chunk_type chunk processing started successfully (PID: $gox_pid)"

        # Wait for processing to complete (check for output files)
        local timeout=45
        local elapsed=0
        local expected_output="$OUTPUT_DIR/$chunk_type"

        while [[ $elapsed -lt $timeout ]]; do
            if [[ -d "$expected_output" && $(find "$expected_output" -type f 2>/dev/null | wc -l) -gt 0 ]]; then
                log "INFO" "$chunk_type chunk processing completed"
                break
            fi
            sleep 3
            elapsed=$((elapsed + 3))
        done

        # Stop the process
        kill $gox_pid 2>/dev/null || true
        wait $gox_pid 2>/dev/null || true
    else
        log "ERROR" "$chunk_type chunk processing failed to start"
        return 1
    fi
}

# Display results
display_results() {
    log "INFO" "Chunk Processing Results Summary:"
    echo ""

    # Check each output directory
    for chunk_type in basic compressed parallel streaming complete; do
        local output_path="$OUTPUT_DIR/$chunk_type"
        if [[ -d "$output_path" ]]; then
            local file_count=$(find "$output_path" -type f 2>/dev/null | wc -l)
            local total_size=$(find "$output_path" -type f -exec stat -f%z {} + 2>/dev/null | awk '{sum += $1} END {printf "%.2fKB", sum/1024}')

            if [[ $file_count -gt 0 ]]; then
                echo -e "${GREEN}✅ ${chunk_type^} Chunk Processing:${NC} $file_count files created ($total_size)"

                # Show sample results
                local chunk_files=($(find "$output_path" -name "chunk_*.json" -o -name "chunk_*.gz" | head -3))
                local manifest_files=($(find "$output_path" -name "*.manifest.json"))

                for chunk_file in "${chunk_files[@]}"; do
                    local file_size=$(stat -f%z "$chunk_file" 2>/dev/null || echo "0")
                    echo "   Chunk: $(basename "$chunk_file") ($(echo "scale=2; $file_size/1024" | bc 2>/dev/null || echo "0")KB)"
                done

                for manifest_file in "${manifest_files[@]}"; do
                    echo "   Manifest: $(basename "$manifest_file")"
                    if command -v jq >/dev/null 2>&1; then
                        local chunk_count=$(jq -r '.chunks | length' "$manifest_file" 2>/dev/null || echo "unknown")
                        echo "     Total chunks: $chunk_count"
                    fi
                done
            else
                echo -e "${YELLOW}⚠️  ${chunk_type^} Chunk Processing:${NC} No output files generated"
            fi
        else
            echo -e "${RED}❌ ${chunk_type^} Chunk Processing:${NC} Output directory not found"
        fi
        echo ""
    done

    # Show overall statistics
    echo -e "${BLUE}📊 Processing Statistics:${NC}"
    local total_output_files=$(find "$OUTPUT_DIR" -type f 2>/dev/null | wc -l)
    local total_output_size=$(find "$OUTPUT_DIR" -type f -exec stat -f%z {} + 2>/dev/null | awk '{sum += $1} END {printf "%.2fMB", sum/(1024*1024)}')
    echo "   Total output files: $total_output_files"
    echo "   Total output size: $total_output_size"

    # Show input vs output comparison
    local input_files=$(find "$DEMO_DIR/input" -type f 2>/dev/null | wc -l)
    local input_size=$(find "$DEMO_DIR/input" -type f -exec stat -f%z {} + 2>/dev/null | awk '{sum += $1} END {printf "%.2fMB", sum/(1024*1024)}')
    echo "   Original files: $input_files ($input_size)"
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
}

# Main demo execution
main() {
    echo "=== GOX Framework Chunk Processing Demo ==="
    echo ""

    # Setup
    check_prerequisites
    setup_demo_environment

    # Trap cleanup on exit
    trap cleanup EXIT

    # Run chunk processing based on user selection
    case "$CHUNK_TYPE" in
        "basic")
            run_chunk_type "basic_chunk_writer"
            ;;
        "compressed")
            run_chunk_type "compressed_chunk_writer"
            ;;
        "parallel")
            run_chunk_type "parallel_chunk_writer"
            ;;
        "streaming")
            run_chunk_type "streaming_chunk_writer"
            ;;
        "all"|"")
            log "INFO" "Running complete chunk processing pipeline..."

            # Run each chunk processing type
            for chunk_type in basic_chunk_writer compressed_chunk_writer parallel_chunk_writer streaming_chunk_writer; do
                if ! run_chunk_type "$chunk_type"; then
                    log "WARN" "Failed to run $chunk_type, continuing..."
                fi
            done

            # Run complete pipeline
            if ! run_chunk_type "complete_chunk_pipeline"; then
                log "WARN" "Failed to run complete chunk pipeline"
            fi
            ;;
        *)
            log "ERROR" "Unknown chunk processing type: $CHUNK_TYPE"
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