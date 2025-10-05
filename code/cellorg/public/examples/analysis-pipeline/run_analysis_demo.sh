#!/bin/bash

# GOX Framework Analysis Pipeline Demo
# Demonstrates binary, JSON, XML, and image analysis capabilities

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOX_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default configuration
ANALYZER=""
INPUT_DIR="${SCRIPT_DIR}/input"
OUTPUT_DIR="/tmp/gox-analysis-demo/output"
LOG_DIR="/tmp/gox-analysis-demo/logs"
DEMO_DIR="/tmp/gox-analysis-demo"

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
    echo "  --analyzer=TYPE    Run specific analyzer (binary|json|xml|image|all)"
    echo "  --input=DIR        Input directory (default: ./input)"
    echo "  --output=DIR       Output directory (default: /tmp/gox-analysis-demo/output)"
    echo "  --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Run all analyzers"
    echo "  $0 --analyzer=json         # Run JSON analyzer only"
    echo "  $0 --input=/path/to/files  # Use custom input directory"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --analyzer=*)
            ANALYZER="${1#*=}"
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

    # Check for analyzer binaries
    local analyzers=("binary_analyzer" "json_analyzer" "xml_analyzer" "image_analyzer")
    for analyzer in "${analyzers[@]}"; do
        if [[ ! -f "${GOX_ROOT}/build/${analyzer}" ]]; then
            log "WARN" "${analyzer} binary not found. Some demos may not work."
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
    mkdir -p "$DEMO_DIR"/{input,output,logs,config}
    mkdir -p "$OUTPUT_DIR"/{binary,json,xml,image}

    # Copy input files to demo directory
    cp -r "$INPUT_DIR"/* "$DEMO_DIR/input/"

    # Create cell configuration files
    create_cell_configs

    log "INFO" "Demo environment setup completed"
}

# Create cell configuration files
create_cell_configs() {
    local config_dir="$DEMO_DIR/config"

    # Binary analyzer cell config
    cat > "$config_dir/binary_analysis_cell.yaml" << EOF
cell:
  id: "analysis:binary-files"
  description: "Binary file analysis cell"

  agents:
    - id: "binary-analyzer-001"
      agent_type: "binary_analyzer"
      ingress: "file:$DEMO_DIR/input/binary/*"
      egress: "file:$OUTPUT_DIR/binary/"
      config:
        deep_scan: true
        extract_strings: true
        min_string_length: 4
        max_file_size: "50MB"
        security_analysis: true
        detect_packing: true
EOF

    # JSON analyzer cell config
    cat > "$config_dir/json_analysis_cell.yaml" << EOF
cell:
  id: "analysis:json-files"
  description: "JSON file analysis cell"

  agents:
    - id: "json-analyzer-001"
      agent_type: "json_analyzer"
      ingress: "file:$DEMO_DIR/input/json/*"
      egress: "file:$OUTPUT_DIR/json/"
      config:
        validate_syntax: true
        validate_schema: true
        schema_path: "$SCRIPT_DIR/schemas/"
        extract_paths: true
        pretty_print: true
        max_depth: 10
EOF

    # XML analyzer cell config
    cat > "$config_dir/xml_analysis_cell.yaml" << EOF
cell:
  id: "analysis:xml-files"
  description: "XML file analysis cell"

  agents:
    - id: "xml-analyzer-001"
      agent_type: "xml_analyzer"
      ingress: "file:$DEMO_DIR/input/xml/*"
      egress: "file:$OUTPUT_DIR/xml/"
      config:
        validate_wellformed: true
        validate_schema: true
        schema_path: "$SCRIPT_DIR/schemas/"
        extract_elements: true
        namespace_aware: true
        preserve_whitespace: false
EOF

    # Image analyzer cell config
    cat > "$config_dir/image_analysis_cell.yaml" << EOF
cell:
  id: "analysis:image-files"
  description: "Image file analysis cell"

  agents:
    - id: "image-analyzer-001"
      agent_type: "image_analyzer"
      ingress: "file:$DEMO_DIR/input/images/*"
      egress: "file:$OUTPUT_DIR/image/"
      config:
        extract_metadata: true
        extract_exif: true
        analyze_colors: true
        generate_thumbnail: true
        max_image_size: "100MB"
        thumbnail_size: "200x200"
EOF

    # Complete analysis cell config
    cat > "$config_dir/complete_analysis_cell.yaml" << EOF
cell:
  id: "analysis:complete-pipeline"
  description: "Complete file analysis pipeline"

  agents:
    - id: "file-router-001"
      agent_type: "strategy_selector"
      ingress: "file:$DEMO_DIR/input/**/*"
      egress: "route:analysis"
      config:
        routing_strategy: "file_type"

    - id: "binary-analyzer-001"
      agent_type: "binary_analyzer"
      ingress: "route:analysis:binary"
      egress: "pub:binary-results"
      config:
        deep_scan: true
        extract_strings: true
        security_analysis: true

    - id: "json-analyzer-001"
      agent_type: "json_analyzer"
      ingress: "route:analysis:json"
      egress: "pub:json-results"
      config:
        validate_syntax: true
        validate_schema: true

    - id: "xml-analyzer-001"
      agent_type: "xml_analyzer"
      ingress: "route:analysis:xml"
      egress: "pub:xml-results"
      config:
        validate_wellformed: true
        validate_schema: true

    - id: "image-analyzer-001"
      agent_type: "image_analyzer"
      ingress: "route:analysis:image"
      egress: "pub:image-results"
      config:
        extract_metadata: true
        analyze_colors: true

    - id: "results-aggregator-001"
      agent_type: "report_generator"
      ingress: "sub:binary-results,json-results,xml-results,image-results"
      egress: "file:$OUTPUT_DIR/analysis-report.json"
      config:
        report_format: "comprehensive"
        include_statistics: true
EOF
}

# Run specific analyzer
run_analyzer() {
    local analyzer_type=$1
    local config_file="$DEMO_DIR/config/${analyzer_type}_analysis_cell.yaml"

    log "INFO" "Running $analyzer_type analyzer..."

    if [[ ! -f "$config_file" ]]; then
        log "ERROR" "Configuration file not found: $config_file"
        return 1
    fi

    # Start the analysis
    log "DEBUG" "Starting GOX with configuration: $config_file"
    cd "$GOX_ROOT"

    # Run in background and capture PID for cleanup
    ./build/gox cell run "$config_file" > "$LOG_DIR/${analyzer_type}_analysis.log" 2>&1 &
    local gox_pid=$!

    # Give it time to process
    sleep 5

    # Check if process is still running (indicates success)
    if kill -0 $gox_pid 2>/dev/null; then
        log "INFO" "$analyzer_type analysis started successfully (PID: $gox_pid)"

        # Wait for processing to complete (check for output files)
        local timeout=30
        local elapsed=0
        while [[ $elapsed -lt $timeout ]]; do
            if [[ $(find "$OUTPUT_DIR/${analyzer_type}" -type f 2>/dev/null | wc -l) -gt 0 ]]; then
                log "INFO" "$analyzer_type analysis completed"
                break
            fi
            sleep 2
            elapsed=$((elapsed + 2))
        done

        # Stop the process
        kill $gox_pid 2>/dev/null || true
        wait $gox_pid 2>/dev/null || true
    else
        log "ERROR" "$analyzer_type analysis failed to start"
        return 1
    fi
}

# Display results
display_results() {
    log "INFO" "Analysis Results Summary:"
    echo ""

    # Count processed files by type
    for analyzer_type in binary json xml image; do
        local output_path="$OUTPUT_DIR/$analyzer_type"
        if [[ -d "$output_path" ]]; then
            local file_count=$(find "$output_path" -type f 2>/dev/null | wc -l)
            local input_count=$(find "$DEMO_DIR/input/$analyzer_type" -type f 2>/dev/null | wc -l)

            if [[ $file_count -gt 0 ]]; then
                echo -e "${GREEN}✅ $analyzer_type Analysis:${NC} $file_count/$input_count files processed"

                # Show sample results
                local sample_file=$(find "$output_path" -name "*.json" -type f | head -1)
                if [[ -n "$sample_file" ]]; then
                    echo "   Sample result: $(basename "$sample_file")"
                    if command -v jq >/dev/null 2>&1; then
                        echo "   Preview:"
                        jq -C '.' < "$sample_file" 2>/dev/null | head -10 | sed 's/^/     /'
                        echo "     ..."
                    fi
                fi
            else
                echo -e "${YELLOW}⚠️  $analyzer_type Analysis:${NC} No output files generated"
            fi
        else
            echo -e "${RED}❌ $analyzer_type Analysis:${NC} Output directory not found"
        fi
        echo ""
    done
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
    echo "=== GOX Framework Analysis Pipeline Demo ==="
    echo ""

    # Setup
    check_prerequisites
    setup_demo_environment

    # Trap cleanup on exit
    trap cleanup EXIT

    # Run analysis based on user selection
    case "$ANALYZER" in
        "binary")
            run_analyzer "binary"
            ;;
        "json")
            run_analyzer "json"
            ;;
        "xml")
            run_analyzer "xml"
            ;;
        "image")
            run_analyzer "image"
            ;;
        "all"|"")
            log "INFO" "Running complete analysis pipeline..."

            # Run each analyzer
            for analyzer in binary json xml image; do
                if ! run_analyzer "$analyzer"; then
                    log "WARN" "Failed to run $analyzer analyzer, continuing..."
                fi
            done
            ;;
        *)
            log "ERROR" "Unknown analyzer type: $ANALYZER"
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