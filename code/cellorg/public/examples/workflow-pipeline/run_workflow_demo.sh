#!/bin/bash

# GOX Framework Workflow Pipeline Demo
# Demonstrates strategy selection, context enrichment, dataset building, and text chunking

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOX_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default configuration
WORKFLOW=""
CONFIG_FILE=""
INPUT_DIR="${SCRIPT_DIR}/input"
OUTPUT_DIR="/tmp/gox-workflow-demo/output"
LOG_DIR="/tmp/gox-workflow-demo/logs"
DEMO_DIR="/tmp/gox-workflow-demo"
DEBUG=false

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
    echo "  --workflow=TYPE        Run specific workflow (strategy|enrichment|dataset|chunking|all)"
    echo "  --config=FILE          Use custom configuration file"
    echo "  --input=DIR            Input directory (default: ./input)"
    echo "  --output=DIR           Output directory (default: /tmp/gox-workflow-demo/output)"
    echo "  --debug                Enable debug logging"
    echo "  --help                 Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                              # Run complete workflow"
    echo "  $0 --workflow=strategy          # Run strategy selection only"
    echo "  $0 --workflow=enrichment        # Run context enrichment only"
    echo "  $0 --config=custom.yaml         # Use custom configuration"
    echo "  $0 --debug                      # Enable debug output"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --workflow=*)
            WORKFLOW="${1#*=}"
            shift
            ;;
        --config=*)
            CONFIG_FILE="${1#*=}"
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
        --debug)
            DEBUG=true
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
            echo -e "${GREEN}[INFO]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/workflow_demo.log"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/workflow_demo.log"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/workflow_demo.log"
            ;;
        "DEBUG")
            if [[ "$DEBUG" == "true" ]]; then
                echo -e "${BLUE}[DEBUG]${NC} ${timestamp} - $message" | tee -a "${LOG_DIR}/workflow_demo.log"
            fi
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

    # Check for workflow agent binaries
    local agents=("strategy_selector" "context_enricher" "dataset_builder" "text_chunker")
    for agent in "${agents[@]}"; do
        if [[ ! -f "${GOX_ROOT}/build/${agent}" ]]; then
            log "WARN" "${agent} binary not found. Some workflows may not work."
        fi
    done

    log "INFO" "Prerequisites check completed"
}

# Setup demo environment
setup_demo_environment() {
    log "INFO" "Setting up workflow demo environment..."

    # Create demo directories
    mkdir -p "$DEMO_DIR"/{input,output,logs,config,monitoring}
    mkdir -p "$OUTPUT_DIR"/{strategy_decisions,enriched_content,built_datasets,chunked_text}

    # Create sample input data
    create_sample_data

    # Create cell configuration files
    create_cell_configs

    log "INFO" "Demo environment setup completed"
}

# Create sample input data
create_sample_data() {
    log "DEBUG" "Creating sample input data..."

    local input_docs="$DEMO_DIR/input/documents"
    mkdir -p "$input_docs"/{research_papers,news_articles,technical_docs,mixed_content}

    # Sample research paper
    cat > "$input_docs/research_papers/ml_advances.txt" << EOF
Machine Learning Advances in 2024

Abstract
This paper presents recent advances in machine learning, focusing on transformer architectures, federated learning, and multimodal AI systems. Our research demonstrates significant improvements in efficiency and accuracy across multiple domains.

Introduction
Machine learning has evolved rapidly in recent years. The development of attention mechanisms and transformer models has revolutionized natural language processing. Federated learning has enabled privacy-preserving distributed training. Multimodal systems now integrate text, image, and audio processing seamlessly.

Methodology
We implemented several novel approaches: 1) Efficient transformer variants with reduced computational complexity, 2) Federated learning protocols with differential privacy, 3) Cross-modal attention mechanisms for unified representation learning.

Results
Our experiments show 40% improvement in training efficiency, 15% better accuracy on benchmark tasks, and successful deployment across diverse federated environments with strong privacy guarantees.

Conclusion
These advances represent significant progress in making machine learning more efficient, private, and capable of handling diverse data modalities.
EOF

    # Sample news article
    cat > "$input_docs/news_articles/tech_news.txt" << EOF
Tech Industry Report: AI Adoption Accelerates

Breaking: Major technology companies are reporting unprecedented adoption rates of artificial intelligence across enterprise sectors. According to recent surveys, 78% of Fortune 500 companies have implemented some form of AI automation in their operations.

The surge is driven by advances in large language models, computer vision, and robotic process automation. Companies are seeing average productivity gains of 25-30% in areas where AI has been deployed.

Key sectors showing rapid adoption include:
- Healthcare: Diagnostic imaging and drug discovery
- Finance: Risk assessment and fraud detection
- Manufacturing: Predictive maintenance and quality control
- Retail: Personalized recommendations and inventory optimization

Industry experts predict this trend will continue accelerating through 2025, with smaller companies increasingly adopting AI solutions as costs decrease and accessibility improves.
EOF

    # Sample technical documentation
    cat > "$input_docs/technical_docs/api_spec.txt" << EOF
GOX Framework API Specification v3.0

Overview
The GOX Framework provides a comprehensive API for building distributed agent-based processing pipelines. This specification covers all endpoints, message formats, and integration patterns.

Authentication
All API requests require authentication via JWT tokens. Include the token in the Authorization header:
Authorization: Bearer <jwt_token>

Endpoints

POST /api/v3/agents
Creates a new agent instance
Parameters:
- agent_type (string): Type of agent to create
- config (object): Agent configuration parameters
- capabilities (array): List of agent capabilities

GET /api/v3/agents/{id}
Retrieves agent information
Returns: Agent details including status, configuration, and metrics

POST /api/v3/cells
Creates a new processing cell
Parameters:
- cell_definition (object): Complete cell specification
- agents (array): List of agents in the cell
- routing (object): Message routing configuration

Message Formats
All messages use JSON format with the following structure:
{
  "id": "unique_message_id",
  "type": "message_type",
  "payload": {},
  "metadata": {},
  "timestamp": "2024-09-27T10:00:00Z"
}

Error Handling
The API returns standard HTTP status codes with detailed error messages in JSON format.
EOF

    # Sample mixed content
    cat > "$input_docs/mixed_content/dataset_info.txt" << EOF
Customer Analytics Dataset Documentation

Dataset: customer_behavior_2024
Version: 1.2.0
Created: 2024-09-27
Records: 1,250,000
Size: 485MB

Schema:
- customer_id (string): Unique customer identifier
- email (string): Customer email address
- age (integer): Customer age
- registration_date (timestamp): Account creation date
- total_purchases (decimal): Lifetime purchase amount
- last_activity (timestamp): Most recent account activity

Sources:
1. CRM System: Customer profile data
2. E-commerce Platform: Purchase history
3. Mobile App: Activity tracking
4. Support System: Interaction logs

Quality Metrics:
- Completeness: 96%
- Validity: 91%
- Uniqueness: 89%
- Consistency: 94%

Usage Notes:
This dataset is suitable for customer segmentation, churn prediction, and lifetime value modeling. Personal information has been anonymized following GDPR guidelines.
EOF

    log "DEBUG" "Sample input data created"
}

# Create cell configuration files
create_cell_configs() {
    local config_dir="$DEMO_DIR/config"

    # Strategy Selection Cell
    cat > "$config_dir/strategy_selection_cell.yaml" << EOF
cell:
  id: "workflow:strategy-selection"
  description: "Intelligent content routing and strategy selection"

  agents:
    - id: "strategy-selector-001"
      agent_type: "strategy_selector"
      ingress: "file:$DEMO_DIR/input/documents/**/*"
      egress: "route:processing-strategy"
      config:
        routing_strategy: "content_analysis"
        strategies:
          - name: "fast_processing"
            conditions:
              - file_size: "<5KB"
              - content_type: "text"
            route_to: "fast-processor"
          - name: "deep_analysis"
            conditions:
              - file_size: ">10KB"
              - content_type: "text"
            route_to: "deep-analyzer"
        load_balancing:
          enabled: true
          algorithm: "round_robin"
        performance_monitoring:
          enabled: true
          metrics: ["throughput", "latency"]
EOF

    # Context Enrichment Cell
    cat > "$config_dir/context_enrichment_cell.yaml" << EOF
cell:
  id: "workflow:context-enrichment"
  description: "Content enrichment with external context"

  agents:
    - id: "context-enricher-001"
      agent_type: "context_enricher"
      ingress: "file:$DEMO_DIR/input/documents/**/*"
      egress: "file:$OUTPUT_DIR/enriched_content/"
      config:
        enrichment_sources:
          - type: "mock_knowledge_graph"
            name: "demo_kg"
            fields: ["entities", "relationships"]
          - type: "mock_database"
            name: "demo_db"
            tables: ["authors", "citations"]
        enrichment_rules:
          - trigger: "person_name_detected"
            action: "lookup_biography"
          - trigger: "location_mentioned"
            action: "add_geographic_context"
          - trigger: "technical_term_found"
            action: "add_definition"
        quality_control:
          confidence_threshold: 0.8
          validate_sources: true
EOF

    # Dataset Building Cell
    cat > "$config_dir/dataset_building_cell.yaml" << EOF
cell:
  id: "workflow:dataset-building"
  description: "Automated dataset assembly and validation"

  agents:
    - id: "dataset-builder-001"
      agent_type: "dataset_builder"
      ingress: "file:$DEMO_DIR/input/documents/**/*"
      egress: "file:$OUTPUT_DIR/built_datasets/"
      config:
        sources:
          - type: "text_files"
            path: "$DEMO_DIR/input/documents/**/*.txt"
            schema_inference: true
        dataset_config:
          output_format: "json"
          compression: "gzip"
          version_control: true
        quality_control:
          duplicate_detection: true
          validation_rules:
            - field: "content"
              rule: "not_empty"
            - field: "metadata"
              rule: "valid_json"
        metadata_generation:
          lineage_tracking: true
          statistics: true
EOF

    # Text Chunking Cell
    cat > "$config_dir/text_chunking_cell.yaml" << EOF
cell:
  id: "workflow:text-chunking"
  description: "Intelligent text segmentation and chunking"

  agents:
    - id: "text-chunker-001"
      agent_type: "text_chunker"
      ingress: "file:$DEMO_DIR/input/documents/**/*"
      egress: "file:$OUTPUT_DIR/chunked_text/"
      config:
        chunking_strategies:
          - name: "semantic_chunking"
            method: "sentence_boundary"
            max_chunk_size: 512
            overlap_size: 50
            preserve_sentences: true
        content_analysis:
          language_detection: true
          complexity_scoring: true
        optimization:
          target_use_case: "rag_pipeline"
          context_window: 4096
        quality_control:
          min_chunk_size: 50
          coherence_check: true
        output_format:
          include_metadata: true
          add_chunk_ids: true
EOF

    # Complete Workflow Cell
    cat > "$config_dir/complete_workflow_cell.yaml" << EOF
cell:
  id: "workflow:complete-processing"
  description: "Complete intelligent content processing workflow"

  agents:
    - id: "strategy-selector-001"
      agent_type: "strategy_selector"
      ingress: "file:$DEMO_DIR/input/documents/**/*"
      egress: "route:processing-strategy"
      config:
        routing_strategy: "content_analysis"

    - id: "text-chunker-001"
      agent_type: "text_chunker"
      ingress: "route:processing-strategy:text"
      egress: "pub:chunked-text"
      config:
        chunking_strategy: "semantic_chunking"

    - id: "context-enricher-001"
      agent_type: "context_enricher"
      ingress: "route:processing-strategy:enrich"
      egress: "pub:enriched-content"
      config:
        enrichment_sources: ["demo_kg"]

    - id: "dataset-builder-001"
      agent_type: "dataset_builder"
      ingress: "sub:chunked-text,enriched-content"
      egress: "file:$OUTPUT_DIR/final-dataset/"
      config:
        output_format: "json"
        quality_control: true

  workflow:
    execution_model: "pipeline"
    error_handling: "continue_with_logging"
    monitoring:
      enabled: true
      metrics: ["throughput", "quality"]
EOF
}

# Run specific workflow
run_workflow() {
    local workflow_type=$1
    local config_file=""

    case "$workflow_type" in
        "strategy")
            config_file="$DEMO_DIR/config/strategy_selection_cell.yaml"
            ;;
        "enrichment")
            config_file="$DEMO_DIR/config/context_enrichment_cell.yaml"
            ;;
        "dataset")
            config_file="$DEMO_DIR/config/dataset_building_cell.yaml"
            ;;
        "chunking")
            config_file="$DEMO_DIR/config/text_chunking_cell.yaml"
            ;;
        "complete"|"all")
            config_file="$DEMO_DIR/config/complete_workflow_cell.yaml"
            ;;
        *)
            log "ERROR" "Unknown workflow type: $workflow_type"
            return 1
            ;;
    esac

    if [[ -n "$CONFIG_FILE" ]]; then
        config_file="$CONFIG_FILE"
    fi

    log "INFO" "Running $workflow_type workflow..."
    log "DEBUG" "Using configuration: $config_file"

    if [[ ! -f "$config_file" ]]; then
        log "ERROR" "Configuration file not found: $config_file"
        return 1
    fi

    # Start the workflow
    cd "$GOX_ROOT"

    # Set debug environment if requested
    if [[ "$DEBUG" == "true" ]]; then
        export GOX_LOG_LEVEL=debug
        export GOX_WORKFLOW_DEBUG=true
    fi

    # Run in background and capture PID
    ./build/gox cell run "$config_file" > "$LOG_DIR/${workflow_type}_workflow.log" 2>&1 &
    local gox_pid=$!

    log "DEBUG" "Started workflow with PID: $gox_pid"

    # Give it time to process
    sleep 8

    # Check if process is still running
    if kill -0 $gox_pid 2>/dev/null; then
        log "INFO" "$workflow_type workflow started successfully"

        # Wait for processing to complete
        local timeout=60
        local elapsed=0
        while [[ $elapsed -lt $timeout ]]; do
            local output_count=0
            case "$workflow_type" in
                "strategy")
                    output_count=$(find "$OUTPUT_DIR/strategy_decisions" -type f 2>/dev/null | wc -l)
                    ;;
                "enrichment")
                    output_count=$(find "$OUTPUT_DIR/enriched_content" -type f 2>/dev/null | wc -l)
                    ;;
                "dataset")
                    output_count=$(find "$OUTPUT_DIR/built_datasets" -type f 2>/dev/null | wc -l)
                    ;;
                "chunking")
                    output_count=$(find "$OUTPUT_DIR/chunked_text" -type f 2>/dev/null | wc -l)
                    ;;
                "complete"|"all")
                    output_count=$(find "$OUTPUT_DIR" -type f 2>/dev/null | wc -l)
                    ;;
            esac

            if [[ $output_count -gt 0 ]]; then
                log "INFO" "$workflow_type workflow completed ($output_count output files)"
                break
            fi

            sleep 3
            elapsed=$((elapsed + 3))
        done

        # Stop the process
        kill $gox_pid 2>/dev/null || true
        wait $gox_pid 2>/dev/null || true
    else
        log "ERROR" "$workflow_type workflow failed to start"
        return 1
    fi
}

# Display workflow results
display_results() {
    log "INFO" "Workflow Results Summary:"
    echo ""

    # Strategy Selection Results
    if [[ -d "$OUTPUT_DIR/strategy_decisions" ]]; then
        local strategy_files=$(find "$OUTPUT_DIR/strategy_decisions" -type f 2>/dev/null | wc -l)
        if [[ $strategy_files -gt 0 ]]; then
            echo -e "${GREEN}✅ Strategy Selection:${NC} $strategy_files routing decisions made"
        fi
    fi

    # Context Enrichment Results
    if [[ -d "$OUTPUT_DIR/enriched_content" ]]; then
        local enriched_files=$(find "$OUTPUT_DIR/enriched_content" -type f 2>/dev/null | wc -l)
        if [[ $enriched_files -gt 0 ]]; then
            echo -e "${GREEN}✅ Context Enrichment:${NC} $enriched_files documents enriched"
        fi
    fi

    # Dataset Building Results
    if [[ -d "$OUTPUT_DIR/built_datasets" ]]; then
        local dataset_files=$(find "$OUTPUT_DIR/built_datasets" -type f 2>/dev/null | wc -l)
        if [[ $dataset_files -gt 0 ]]; then
            echo -e "${GREEN}✅ Dataset Building:${NC} $dataset_files datasets created"
        fi
    fi

    # Text Chunking Results
    if [[ -d "$OUTPUT_DIR/chunked_text" ]]; then
        local chunk_files=$(find "$OUTPUT_DIR/chunked_text" -type f 2>/dev/null | wc -l)
        if [[ $chunk_files -gt 0 ]]; then
            echo -e "${GREEN}✅ Text Chunking:${NC} $chunk_files chunk files created"
        fi
    fi

    # Show sample output
    echo ""
    log "INFO" "Sample workflow outputs:"

    for output_type in strategy_decisions enriched_content built_datasets chunked_text; do
        local sample_file=$(find "$OUTPUT_DIR/$output_type" -name "*.json" -type f 2>/dev/null | head -1)
        if [[ -n "$sample_file" ]]; then
            echo ""
            echo -e "${BLUE}Sample $output_type result:${NC}"
            if command -v jq >/dev/null 2>&1; then
                jq -C '.' < "$sample_file" 2>/dev/null | head -15 | sed 's/^/  /'
                echo "  ..."
            else
                head -10 "$sample_file" | sed 's/^/  /'
                echo "  ..."
            fi
        fi
    done
}

# Cleanup function
cleanup() {
    log "INFO" "Cleaning up workflow demo environment..."

    # Kill any remaining GOX processes
    pkill -f "gox.*cell.*run" 2>/dev/null || true

    log "INFO" "Workflow demo completed"
    log "INFO" "Output files available in: $OUTPUT_DIR"
    log "INFO" "Logs available in: $LOG_DIR"
}

# Main demo execution
main() {
    echo "=== GOX Framework Workflow Pipeline Demo ==="
    echo ""

    # Setup
    check_prerequisites
    setup_demo_environment

    # Trap cleanup on exit
    trap cleanup EXIT

    # Run workflow based on user selection
    case "$WORKFLOW" in
        "strategy"|"enrichment"|"dataset"|"chunking")
            run_workflow "$WORKFLOW"
            ;;
        "all"|"complete"|"")
            log "INFO" "Running complete workflow pipeline..."

            # Run individual workflows in sequence
            for workflow in strategy enrichment dataset chunking; do
                if ! run_workflow "$workflow"; then
                    log "WARN" "Failed to run $workflow workflow, continuing..."
                fi
                sleep 2  # Brief pause between workflows
            done
            ;;
        *)
            log "ERROR" "Unknown workflow type: $WORKFLOW"
            usage
            ;;
    esac

    # Display results
    echo ""
    display_results

    echo ""
    echo "=== Workflow Demo Completed Successfully ==="
}

# Run the demo
main "$@"