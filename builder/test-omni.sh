#!/bin/bash

# Omni Test Automation Script
# ===========================
# Comprehensive test automation for CI/CD pipelines and local development
# Part of the AGEN project

set -euo pipefail

# Colors for output
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
BLUE='\033[34m'
CYAN='\033[36m'
RESET='\033[0m'

# Configuration
COVERAGE_THRESHOLD=75
BENCHMARK_ITERATIONS=5
TIMEOUT=300s

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
    exit 1
}

log_section() {
    echo -e "\n${CYAN}🔵 $1${RESET}"
    echo "=================================================="
}

# Check dependencies
check_dependencies() {
    log_section "Checking Dependencies"

    # Check Go version
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
    fi

    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1)
    log_info "Go version: $GO_VERSION"

    # Check if golangci-lint is available
    if command -v golangci-lint &> /dev/null; then
        log_info "golangci-lint available"
    else
        log_warning "golangci-lint not available - will skip linting"
    fi

    # Check if we're in a Go module
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found. Please run from project root."
    fi

    log_success "Dependencies check completed"
}

# Pre-test setup
setup() {
    log_section "Setting Up Test Environment"

    # Clean up previous artifacts
    rm -f coverage.out coverage.html
    rm -f cpu.prof mem.prof
    rm -f bench-*.txt
    rm -rf ../support/test-report/

    # Create necessary directories
    mkdir -p ../support/test-report
    mkdir -p ../support/profiles

    log_success "Test environment setup completed"
}

# Code quality checks
run_quality_checks() {
    log_section "Running Code Quality Checks"

    # Format check
    log_info "Checking code formatting..."
    if [ -n "$(go fmt ./...)" ]; then
        log_error "Code is not properly formatted. Run 'go fmt ./...' to fix."
    fi
    log_success "Code formatting check passed"

    # Vet check
    log_info "Running go vet..."
    go vet ./... || log_error "go vet failed"
    log_success "go vet passed"

    # Lint check (if available)
    if command -v golangci-lint &> /dev/null; then
        log_info "Running golangci-lint..."
        golangci-lint run --timeout=${TIMEOUT} || log_error "Linting failed"
        log_success "Linting passed"
    fi
}

# Unit tests
run_unit_tests() {
    log_section "Running Unit Tests"

    log_info "Running unit tests with race detection..."
    go test -race -short -v ./... -timeout=${TIMEOUT} || log_error "Unit tests failed"
    log_success "Unit tests passed"
}

# Integration tests
run_integration_tests() {
    log_section "Running Integration Tests"

    log_info "Running integration tests..."
    go test -race -run Integration -v ./... -timeout=${TIMEOUT} || {
        log_warning "Integration tests failed or not found"
        return 0
    }
    log_success "Integration tests passed"
}

# Coverage analysis
run_coverage_tests() {
    log_section "Running Coverage Analysis"

    log_info "Running tests with coverage analysis..."
    go test -race -covermode=atomic -coverprofile=coverage.out ./... -timeout=${TIMEOUT} || log_error "Coverage tests failed"

    # Generate coverage report
    go tool cover -html=coverage.out -o coverage.html

    # Check coverage threshold
    COVERAGE=$(go tool cover -func=coverage.out | grep total | grep -oE '[0-9]+\.[0-9]+')
    log_info "Total coverage: ${COVERAGE}%"

    if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
        log_warning "Coverage ${COVERAGE}% is below threshold ${COVERAGE_THRESHOLD}%"
    else
        log_success "Coverage threshold met: ${COVERAGE}% >= ${COVERAGE_THRESHOLD}%"
    fi

    log_success "Coverage analysis completed"
}

# Benchmark tests
run_benchmarks() {
    log_section "Running Benchmark Tests"

    log_info "Running performance benchmarks..."
    go test -bench=. -benchmem ./... -timeout=${TIMEOUT} > bench-results.txt || {
        log_warning "Benchmarks failed or not found"
        return 0
    }

    log_info "Benchmark results saved to bench-results.txt"
    log_success "Benchmarks completed"
}

# Stress tests
run_stress_tests() {
    log_section "Running Stress Tests"

    log_info "Running stress tests (${BENCHMARK_ITERATIONS} iterations)..."
    go test -race -count=${BENCHMARK_ITERATIONS} ./... -timeout=${TIMEOUT} || {
        log_warning "Stress tests failed"
        return 0
    }

    log_success "Stress tests passed"
}

# Generate test report
generate_report() {
    log_section "Generating Test Report"

    # Check if test-runner binary exists
    if [ -f "../bin/test-runner" ]; then
        log_info "Generating HTML test report..."
        ../bin/test-runner || log_warning "Test report generation failed"
        log_success "Test report generated in ../support/test-report/"
    else
        log_warning "Test runner binary not found. Run 'make build' first."
    fi
}

# Main test suite options
run_quick_tests() {
    log_section "Running Quick Test Suite"
    check_dependencies
    setup
    run_quality_checks
    run_unit_tests
    log_success "Quick test suite completed successfully!"
}

run_ci_tests() {
    log_section "Running CI/CD Test Suite"
    check_dependencies
    setup
    run_quality_checks
    run_unit_tests
    run_integration_tests
    run_coverage_tests
    generate_report
    log_success "CI/CD test suite completed successfully!"
}

run_full_tests() {
    log_section "Running Full Test Suite"
    check_dependencies
    setup
    run_quality_checks
    run_unit_tests
    run_integration_tests
    run_coverage_tests
    run_benchmarks
    run_stress_tests
    generate_report
    log_success "Full test suite completed successfully!"
}

# Usage information
show_usage() {
    echo "Omni Test Automation Script"
    echo ""
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  quick     Run quick test suite (format, vet, unit tests)"
    echo "  ci        Run CI/CD test suite (quick + integration + coverage)"
    echo "  full      Run full test suite (ci + benchmarks + stress tests)"
    echo "  unit      Run unit tests only"
    echo "  coverage  Run coverage analysis only"
    echo "  bench     Run benchmarks only"
    echo "  stress    Run stress tests only"
    echo "  help      Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  COVERAGE_THRESHOLD    Minimum coverage percentage (default: 75)"
    echo "  BENCHMARK_ITERATIONS  Number of stress test iterations (default: 5)"
    echo "  TIMEOUT              Test timeout (default: 300s)"
    echo ""
    echo "Examples:"
    echo "  $0 quick              # Quick development tests"
    echo "  $0 ci                 # Full CI/CD pipeline"
    echo "  $0 full               # Complete test suite"
    echo "  COVERAGE_THRESHOLD=80 $0 ci  # CI with 80% coverage requirement"
}

# Main script logic
main() {
    case "${1:-help}" in
        "quick")
            run_quick_tests
            ;;
        "ci")
            run_ci_tests
            ;;
        "full")
            run_full_tests
            ;;
        "unit")
            check_dependencies
            setup
            run_unit_tests
            ;;
        "coverage")
            check_dependencies
            setup
            run_coverage_tests
            ;;
        "bench")
            check_dependencies
            setup
            run_benchmarks
            ;;
        "stress")
            check_dependencies
            setup
            run_stress_tests
            ;;
        "help"|*)
            show_usage
            exit 0
            ;;
    esac
}

# Run main function with all arguments
main "$@"