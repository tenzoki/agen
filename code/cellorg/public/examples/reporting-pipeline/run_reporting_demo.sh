#!/bin/bash

# GOX Framework Reporting Pipeline Demo
# Demonstrates report generation, summary creation, and metadata collection

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOX_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default configuration
AGENT=""
INPUT_DIR="${SCRIPT_DIR}/input"
OUTPUT_DIR="/tmp/gox-reporting-demo/output"
LOG_DIR="/tmp/gox-reporting-demo/logs"
DEMO_DIR="/tmp/gox-reporting-demo"

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
    echo "  --agent=TYPE       Run specific agent (report|summary|metadata|all)"
    echo "  --input=DIR        Input directory (default: ./input)"
    echo "  --output=DIR       Output directory (default: /tmp/gox-reporting-demo/output)"
    echo "  --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Run all reporting agents"
    echo "  $0 --agent=report          # Run report generator only"
    echo "  $0 --input=/path/to/data   # Use custom input directory"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --agent=*)
            AGENT="${1#*=}"
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

    # Check for reporting agent binaries
    local agents=("report_generator" "summary_generator" "metadata_collector")
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
    mkdir -p "$DEMO_DIR"/{input,output,logs,config,templates}
    mkdir -p "$OUTPUT_DIR"/{reports,summaries,metadata}

    # Copy input files to demo directory
    cp -r "$INPUT_DIR"/* "$DEMO_DIR/input/"

    # Create sample input data if not exists
    create_sample_data

    # Create cell configuration files
    create_cell_configs

    # Create report templates
    create_report_templates

    log "INFO" "Demo environment setup completed"
}

# Create sample input data
create_sample_data() {
    local input_dir="$DEMO_DIR/input"

    # Create analysis results directory
    mkdir -p "$input_dir/analysis-results"
    mkdir -p "$input_dir/processed-data"
    mkdir -p "$input_dir/raw-metadata"

    # Sample analysis result
    cat > "$input_dir/analysis-results/binary-analysis-001.json" << 'EOF'
{
  "analysis_id": "binary-analysis-001",
  "timestamp": "2024-09-27T10:00:00Z",
  "file_path": "/data/sample.exe",
  "results": {
    "file_type": "PE32 executable",
    "security_analysis": {
      "entropy": 7.2,
      "suspicious_strings": 15,
      "packer_detected": false,
      "imports": ["kernel32.dll", "ntdll.dll"],
      "exports": [],
      "sections": [
        {"name": ".text", "virtual_size": 4096, "entropy": 6.8},
        {"name": ".data", "virtual_size": 1024, "entropy": 4.2}
      ]
    },
    "metadata": {
      "size": 2048576,
      "md5": "abc123def456789...",
      "sha256": "def789abc123456...",
      "created": "2024-09-20T08:30:00Z",
      "modified": "2024-09-20T08:30:00Z"
    }
  },
  "processing_info": {
    "duration": 125.3,
    "agent_id": "binary-analyzer-001",
    "success": true
  }
}
EOF

    # Sample JSON analysis result
    cat > "$input_dir/analysis-results/json-analysis-001.json" << 'EOF'
{
  "analysis_id": "json-analysis-001",
  "timestamp": "2024-09-27T10:05:00Z",
  "file_path": "/data/config.json",
  "results": {
    "syntax_valid": true,
    "schema_valid": true,
    "structure": {
      "total_keys": 24,
      "max_depth": 4,
      "array_count": 3,
      "object_count": 8
    },
    "content_analysis": {
      "contains_credentials": false,
      "contains_urls": true,
      "contains_emails": false,
      "sensitive_patterns": []
    }
  },
  "processing_info": {
    "duration": 45.7,
    "agent_id": "json-analyzer-001",
    "success": true
  }
}
EOF

    # Sample processing statistics
    cat > "$input_dir/processed-data/pipeline-stats.json" << 'EOF'
{
  "pipeline_id": "analysis-pipeline-001",
  "start_time": "2024-09-27T09:00:00Z",
  "end_time": "2024-09-27T10:30:00Z",
  "statistics": {
    "files_processed": 45,
    "success_rate": 98.5,
    "average_processing_time": 125.3,
    "total_processing_time": 5638.5,
    "throughput": 0.83
  },
  "errors": [
    {
      "file": "corrupted.bin",
      "error": "Unable to read file header",
      "timestamp": "2024-09-27T09:15:32Z",
      "agent_id": "binary-analyzer-002"
    }
  ],
  "warnings": [
    {
      "file": "large-file.exe",
      "warning": "File size exceeds recommended limit",
      "timestamp": "2024-09-27T09:45:12Z"
    }
  ]
}
EOF

    # Sample metadata
    cat > "$input_dir/raw-metadata/system-info.json" << 'EOF'
{
  "system_metadata": {
    "gox_version": "3.0.0",
    "processing_node": "node-001",
    "environment": "production",
    "resource_usage": {
      "cpu_avg": 45.2,
      "memory_peak": "2.1GB",
      "disk_io": "150MB/s"
    }
  },
  "data_provenance": {
    "source_systems": ["scanner-001", "analyzer-002"],
    "transformation_chain": [
      "file_ingester -> binary_analyzer -> report_generator"
    ],
    "quality_checks": {
      "data_integrity": "passed",
      "completeness": 98.5,
      "accuracy": 99.2
    }
  }
}
EOF
}

# Create cell configuration files
create_cell_configs() {
    local config_dir="$DEMO_DIR/config"

    # Report generator cell config
    cat > "$config_dir/report_generation_cell.yaml" << EOF
cell:
  id: "reporting:report-generation"
  description: "Report generation cell"

  agents:
    - id: "report-generator-001"
      agent_type: "report_generator"
      ingress: "file:$DEMO_DIR/input/analysis-results/*"
      egress: "file:$OUTPUT_DIR/reports/"
      config:
        report_format: "html"
        template: "standard"
        include_charts: true
        include_statistics: true
        max_items_per_section: 100
        output_filename: "analysis-report-{timestamp}.html"
        custom_css: false
        embed_images: true
EOF

    # Summary generator cell config
    cat > "$config_dir/summary_generation_cell.yaml" << EOF
cell:
  id: "reporting:summary-generation"
  description: "Summary generation cell"

  agents:
    - id: "summary-generator-001"
      agent_type: "summary_generator"
      ingress: "file:$DEMO_DIR/input/**/*.json"
      egress: "file:$OUTPUT_DIR/summaries/"
      config:
        summary_type: "executive"
        max_length: 500
        include_trends: true
        highlight_anomalies: true
        confidence_threshold: 0.8
        key_findings_limit: 5
        output_format: "json"
        include_recommendations: true
EOF

    # Metadata collector cell config
    cat > "$config_dir/metadata_collection_cell.yaml" << EOF
cell:
  id: "reporting:metadata-collection"
  description: "Metadata collection cell"

  agents:
    - id: "metadata-collector-001"
      agent_type: "metadata_collector"
      ingress: "file:$DEMO_DIR/input/raw-metadata/*"
      egress: "file:$OUTPUT_DIR/metadata/"
      config:
        collection_scope: "comprehensive"
        include_provenance: true
        include_quality_metrics: true
        include_processing_stats: true
        temporal_tracking: true
        cross_reference_analysis: true
        output_format: "json"
        aggregation_level: "detailed"
EOF

    # Complete reporting pipeline config
    cat > "$config_dir/complete_reporting_cell.yaml" << EOF
cell:
  id: "reporting:complete-pipeline"
  description: "Complete reporting pipeline"

  agents:
    - id: "data-router-001"
      agent_type: "strategy_selector"
      ingress: "file:$DEMO_DIR/input/**/*"
      egress: "route:data-type"
      config:
        routing_strategy: "content_type"

    - id: "metadata-collector-001"
      agent_type: "metadata_collector"
      ingress: "route:data-type:all"
      egress: "pub:metadata"
      config:
        collection_scope: "comprehensive"
        include_provenance: true

    - id: "summary-generator-001"
      agent_type: "summary_generator"
      ingress: "sub:metadata,analysis-results"
      egress: "pub:summaries"
      config:
        summary_type: "technical"
        max_length: 1000

    - id: "report-generator-001"
      agent_type: "report_generator"
      ingress: "sub:metadata,summaries"
      egress: "file:$OUTPUT_DIR/final-report.html"
      config:
        report_format: "html"
        template: "comprehensive"
        include_charts: true
EOF
}

# Create report templates
create_report_templates() {
    local templates_dir="$DEMO_DIR/templates"

    # HTML report template
    cat > "$templates_dir/standard-report.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GOX Analysis Report - {{.Timestamp}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .finding { background: #fff; border-left: 4px solid #007cba; padding: 10px; margin: 10px 0; }
        .warning { border-left-color: #ffa500; }
        .error { border-left-color: #ff0000; }
        .chart { text-align: center; margin: 20px 0; }
        table { width: 100%; border-collapse: collapse; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>GOX Framework Analysis Report</h1>
        <p>Generated: {{.Timestamp}}</p>
        <p>Report ID: {{.ReportID}}</p>
    </div>

    <div class="section">
        <h2>Executive Summary</h2>
        <p>{{.ExecutiveSummary}}</p>
    </div>

    <div class="section">
        <h2>Key Findings</h2>
        {{range .KeyFindings}}
        <div class="finding {{.Severity}}">
            <strong>{{.Title}}</strong>: {{.Description}}
        </div>
        {{end}}
    </div>

    <div class="section">
        <h2>Processing Statistics</h2>
        <table>
            <tr><th>Metric</th><th>Value</th></tr>
            {{range .Statistics}}
            <tr><td>{{.Name}}</td><td>{{.Value}}</td></tr>
            {{end}}
        </table>
    </div>

    <div class="section">
        <h2>Detailed Results</h2>
        {{.DetailedResults}}
    </div>
</body>
</html>
EOF
}

# Run specific agent
run_agent() {
    local agent_type=$1
    local config_file="$DEMO_DIR/config/${agent_type}_cell.yaml"

    log "INFO" "Running $agent_type agent..."

    if [[ ! -f "$config_file" ]]; then
        log "ERROR" "Configuration file not found: $config_file"
        return 1
    fi

    # Start the agent
    log "DEBUG" "Starting GOX with configuration: $config_file"
    cd "$GOX_ROOT"

    # Run in background and capture PID for cleanup
    ./build/gox cell run "$config_file" > "$LOG_DIR/${agent_type}.log" 2>&1 &
    local gox_pid=$!

    # Give it time to process
    sleep 5

    # Check if process is still running (indicates success)
    if kill -0 $gox_pid 2>/dev/null; then
        log "INFO" "$agent_type agent started successfully (PID: $gox_pid)"

        # Wait for processing to complete (check for output files)
        local timeout=30
        local elapsed=0
        local expected_output=""

        case $agent_type in
            "report_generation")
                expected_output="$OUTPUT_DIR/reports"
                ;;
            "summary_generation")
                expected_output="$OUTPUT_DIR/summaries"
                ;;
            "metadata_collection")
                expected_output="$OUTPUT_DIR/metadata"
                ;;
        esac

        while [[ $elapsed -lt $timeout ]]; do
            if [[ -n "$expected_output" && $(find "$expected_output" -type f 2>/dev/null | wc -l) -gt 0 ]]; then
                log "INFO" "$agent_type processing completed"
                break
            fi
            sleep 2
            elapsed=$((elapsed + 2))
        done

        # Stop the process
        kill $gox_pid 2>/dev/null || true
        wait $gox_pid 2>/dev/null || true
    else
        log "ERROR" "$agent_type agent failed to start"
        return 1
    fi
}

# Display results
display_results() {
    log "INFO" "Reporting Results Summary:"
    echo ""

    # Check each output directory
    for agent_type in reports summaries metadata; do
        local output_path="$OUTPUT_DIR/$agent_type"
        if [[ -d "$output_path" ]]; then
            local file_count=$(find "$output_path" -type f 2>/dev/null | wc -l)

            if [[ $file_count -gt 0 ]]; then
                echo -e "${GREEN}✅ ${agent_type^} Generation:${NC} $file_count files created"

                # Show sample results
                local sample_files=($(find "$output_path" -type f | head -3))
                for sample_file in "${sample_files[@]}"; do
                    echo "   Generated: $(basename "$sample_file")"

                    # Show preview for JSON files
                    if [[ "$sample_file" == *.json ]] && command -v jq >/dev/null 2>&1; then
                        echo "   Preview:"
                        jq -C '.' < "$sample_file" 2>/dev/null | head -5 | sed 's/^/     /'
                        echo "     ..."
                    fi
                done
            else
                echo -e "${YELLOW}⚠️  ${agent_type^} Generation:${NC} No output files generated"
            fi
        else
            echo -e "${RED}❌ ${agent_type^} Generation:${NC} Output directory not found"
        fi
        echo ""
    done

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
    echo "=== GOX Framework Reporting Pipeline Demo ==="
    echo ""

    # Setup
    check_prerequisites
    setup_demo_environment

    # Trap cleanup on exit
    trap cleanup EXIT

    # Run agents based on user selection
    case "$AGENT" in
        "report")
            run_agent "report_generation"
            ;;
        "summary")
            run_agent "summary_generation"
            ;;
        "metadata")
            run_agent "metadata_collection"
            ;;
        "all"|"")
            log "INFO" "Running complete reporting pipeline..."

            # Run each agent
            for agent in report_generation summary_generation metadata_collection; do
                if ! run_agent "$agent"; then
                    log "WARN" "Failed to run $agent agent, continuing..."
                fi
            done
            ;;
        *)
            log "ERROR" "Unknown agent type: $AGENT"
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