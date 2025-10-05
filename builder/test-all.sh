#!/bin/bash

# AGEN Comprehensive Test Automation
# ====================================
# Tests all modules and generates unified report
# Usage: ./test-all.sh [quick|ci|full]

set -euo pipefail

# Colors for output
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
BLUE='\033[34m'
CYAN='\033[36m'
RESET='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/reflect/test-reports"
COVERAGE_DIR="${REPORT_DIR}/coverage"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${REPORT_DIR}/test-report-${TIMESTAMP}.md"
SUMMARY_FILE="${REPORT_DIR}/latest-summary.txt"

# Test configuration
MODE="${1:-full}"  # quick, ci, full
COVERAGE_THRESHOLD=70
TIMEOUT=300s

# Modules to test
MODULES=("atomic" "omni" "cellorg" "agents" "alfa")

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${RESET}"
}

log_success() {
    echo -e "${GREEN}✅ $1${RESET}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${RESET}"
}

log_error() {
    echo -e "${RED}❌ $1${RESET}"
}

log_section() {
    echo -e "\n${CYAN}═══════════════════════════════════════════════${RESET}"
    echo -e "${CYAN}▶ $1${RESET}"
    echo -e "${CYAN}═══════════════════════════════════════════════${RESET}\n"
}

# Initialize report
init_report() {
    mkdir -p "$REPORT_DIR"
    mkdir -p "$COVERAGE_DIR"

    cat > "$REPORT_FILE" << EOF
# AGEN Test Report
**Generated:** $(date '+%Y-%m-%d %H:%M:%S')
**Mode:** ${MODE}
**Host:** $(hostname)
**Go Version:** $(go version | cut -d' ' -f3)

---

## Summary

EOF
}

# Test a single module
test_module() {
    local module=$1
    local module_path="${PROJECT_ROOT}/code/${module}"

    log_section "Testing ${module}"

    if [ ! -d "$module_path" ]; then
        log_warning "Module ${module} not found, skipping"
        return 0
    fi

    # Check if module has tests
    local test_count=$(find "$module_path" -name "*_test.go" | wc -l | tr -d ' ')
    if [ "$test_count" -eq 0 ]; then
        log_warning "${module}: No tests found"
        echo "### ${module} (No Tests)" >> "$REPORT_FILE"
        echo "No test files found in this module." >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        return 0
    fi

    log_info "${module}: Found ${test_count} test file(s)"

    # Set test flags based on mode
    local test_flags=""
    case "$MODE" in
        quick)
            test_flags="-short -timeout=60s"
            ;;
        ci)
            test_flags="-race -timeout=${TIMEOUT} -v"
            ;;
        full)
            test_flags="-race -timeout=${TIMEOUT} -v -cover -coverprofile=${COVERAGE_DIR}/${module}-coverage.out"
            ;;
    esac

    # Run tests and capture output
    local module_output="${REPORT_DIR}/${module}-${TIMESTAMP}.log"
    local test_result=0

    # Source onnx-exports for agents module (provides CGO_LDFLAGS for libtokenizers)
    if [ "$module" = "agents" ] && [ -f "${PROJECT_ROOT}/onnx-exports" ]; then
        source "${PROJECT_ROOT}/onnx-exports"
    fi

    cd "$module_path"
    if go test $test_flags ./... 2>&1 | tee "$module_output"; then
        log_success "${module}: Tests passed"
        PASSED_TESTS=$((PASSED_TESTS + 1))

        # Generate coverage report if in full mode
        if [ "$MODE" = "full" ] && [ -f "${COVERAGE_DIR}/${module}-coverage.out" ]; then
            local coverage=$(go tool cover -func="${COVERAGE_DIR}/${module}-coverage.out" | grep total | awk '{print $3}' | sed 's/%//')
            log_info "${module}: Coverage: ${coverage}%"

            # Generate HTML coverage report
            go tool cover -html="${COVERAGE_DIR}/${module}-coverage.out" -o "${COVERAGE_DIR}/${module}-coverage.html" 2>/dev/null || true

            # Add to report
            echo "### ${module} ✅" >> "$REPORT_FILE"
            echo "- **Status:** PASSED" >> "$REPORT_FILE"
            echo "- **Coverage:** ${coverage}%" >> "$REPORT_FILE"
            echo "- **Test Files:** ${test_count}" >> "$REPORT_FILE"
            echo "- **Coverage Report:** [${module}-coverage.html](coverage/${module}-coverage.html)" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"

            # Check coverage threshold
            if (( $(echo "$coverage < $COVERAGE_THRESHOLD" | bc -l) )); then
                log_warning "${module}: Coverage below threshold (${COVERAGE_THRESHOLD}%)"
            fi
        else
            echo "### ${module} ✅" >> "$REPORT_FILE"
            echo "- **Status:** PASSED" >> "$REPORT_FILE"
            echo "- **Test Files:** ${test_count}" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"
        fi
    else
        log_error "${module}: Tests failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        test_result=1

        echo "### ${module} ❌" >> "$REPORT_FILE"
        echo "- **Status:** FAILED" >> "$REPORT_FILE"
        echo "- **Test Files:** ${test_count}" >> "$REPORT_FILE"
        echo "- **Log:** [${module}-${TIMESTAMP}.log](${module}-${TIMESTAMP}.log)" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    fi

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    cd "$PROJECT_ROOT"

    return $test_result
}

# Generate summary
generate_summary() {
    log_section "Generating Summary"

    # Calculate success rate
    local success_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$(echo "scale=2; ($PASSED_TESTS / $TOTAL_TESTS) * 100" | bc)
    fi

    # Update report with summary
    sed -i.bak "/^## Summary/a\\
\\
| Metric | Value |\\
|--------|-------|\\
| **Total Modules** | ${TOTAL_TESTS} |\\
| **Passed** | ${PASSED_TESTS} ✅ |\\
| **Failed** | ${FAILED_TESTS} ❌ |\\
| **Success Rate** | ${success_rate}% |\\
| **Mode** | ${MODE} |\\
" "$REPORT_FILE" && rm "${REPORT_FILE}.bak"

    # Create summary file
    cat > "$SUMMARY_FILE" << EOF
AGEN Test Summary - $(date '+%Y-%m-%d %H:%M:%S')
========================================
Mode:         ${MODE}
Total:        ${TOTAL_TESTS}
Passed:       ${PASSED_TESTS}
Failed:       ${FAILED_TESTS}
Success Rate: ${success_rate}%
Report:       test-report-${TIMESTAMP}.md
EOF

    # Create latest symlink
    ln -sf "test-report-${TIMESTAMP}.md" "${REPORT_DIR}/latest-report.md"

    # Print summary
    log_section "Test Summary"
    echo -e "${CYAN}Total Modules:${RESET}  ${TOTAL_TESTS}"
    echo -e "${GREEN}Passed:${RESET}         ${PASSED_TESTS}"
    echo -e "${RED}Failed:${RESET}         ${FAILED_TESTS}"
    echo -e "${BLUE}Success Rate:${RESET}   ${success_rate}%"
    echo ""
    echo -e "${CYAN}Report:${RESET}         ${REPORT_DIR}/test-report-${TIMESTAMP}.md"

    if [ "$MODE" = "full" ]; then
        echo -e "${CYAN}Coverage:${RESET}       ${COVERAGE_DIR}/"
    fi
}

# Main execution
main() {
    log_section "AGEN Comprehensive Test Suite"
    log_info "Mode: ${MODE}"
    log_info "Project: ${PROJECT_ROOT}"

    # Initialize
    init_report

    # Run tests for each module
    local failed_modules=()
    for module in "${MODULES[@]}"; do
        if ! test_module "$module"; then
            failed_modules+=("$module")
        fi
    done

    # Generate summary
    generate_summary

    # Print final status
    echo ""
    if [ ${#failed_modules[@]} -eq 0 ]; then
        log_success "All tests passed!"
        exit 0
    else
        log_error "Tests failed in modules: ${failed_modules[*]}"
        exit 1
    fi
}

# Run main
main
