# AGEN - Unified Build System
# ============================
# Unified Makefile for all AGEN modules
#
# Usage:
#   make list          - List all available targets
#   make build         - Build all modules
#   make test          - Run all tests
#   make test-quick    - Run quick tests
#   make test-full     - Run full tests with coverage
#   make clean         - Clean all build artifacts

# Project configuration
PROJECT_NAME := AGEN
PROJECT_ROOT := $(shell pwd)
BUILD_DIR := $(PROJECT_ROOT)/bin
SUPPORT_DIR := $(PROJECT_ROOT)/support
REFLECT_DIR := $(PROJECT_ROOT)/reflect
REPORT_DIR := $(REFLECT_DIR)/test-reports
BUILDER_DIR := $(PROJECT_ROOT)/builder

# Go configuration
GO := /opt/homebrew/bin/go
GOPATH := $(shell $(GO) env GOPATH)
GO_VERSION := $(shell $(GO) version | cut -d ' ' -f 3)

# Build metadata
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0-dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# Build flags
BUILD_FLAGS := -v

# CGO Configuration (for agents requiring native libraries like ner_agent)
export MACOSX_DEPLOYMENT_TARGET := 15.5
export CGO_LDFLAGS := -L$(PROJECT_ROOT)/drivers/tokenizers -ltokenizers -ldl -lm -lstdc++

# Module paths
CODE_DIR := $(PROJECT_ROOT)/code
ATOMIC_DIR := $(CODE_DIR)/atomic
OMNI_DIR := $(CODE_DIR)/omni
CELLORG_DIR := $(CODE_DIR)/cellorg
AGENTS_DIR := $(CODE_DIR)/agents
ALFA_DIR := $(CODE_DIR)/alfa

# Agent list (dynamically generated, excluding testutil)
AGENT_LIST := $(shell find $(AGENTS_DIR) -mindepth 1 -maxdepth 1 -type d ! -name testutil -exec basename {} \; | sort)

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
CYAN := \033[36m
MAGENTA := \033[35m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := help

.PHONY: help list build test clean

# ==============================================================================
# Help and List Targets
# ==============================================================================

help: ## Show this help message with descriptions
	@echo "$(CYAN)╔═══════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(CYAN)║           $(MAGENTA)AGEN - Unified Build System$(CYAN)                     ║$(RESET)"
	@echo "$(CYAN)╚═══════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(YELLOW)Project:$(RESET)  $(PROJECT_NAME)"
	@echo "$(YELLOW)Go:$(RESET)       $(GO_VERSION)"
	@echo "$(YELLOW)Version:$(RESET)  $(VERSION)"
	@echo ""
	@echo "$(CYAN)Available targets:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)  %-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(CYAN)Usage examples:$(RESET)"
	@echo "  make list          - List all targets grouped by category"
	@echo "  make build         - Build core components (orchestrator + alfa)"
	@echo "  make build-all     - Build all modules (libs + core + agents)"
	@echo "  make build-agents -j4 - Build agents in parallel (4 jobs)"
	@echo "  make test          - Run all tests with report"
	@echo "  make clean-all     - Clean everything"
	@echo ""

list: ## List all available make targets grouped by category
	@echo "$(CYAN)╔═══════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(CYAN)║              $(MAGENTA)AGEN Build Targets$(CYAN)                           ║$(RESET)"
	@echo "$(CYAN)╚═══════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(MAGENTA)▶ Build Targets:$(RESET)"
	@grep -E '^build.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(MAGENTA)▶ Test Targets:$(RESET)"
	@grep -E '^test.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(MAGENTA)▶ Module Targets:$(RESET)"
	@grep -E '^(atomic|omni|cellorg|agents|alfa).*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(MAGENTA)▶ Quality Targets:$(RESET)"
	@grep -E '^(format|vet|lint|check).*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(MAGENTA)▶ Utility Targets:$(RESET)"
	@grep -E '^(clean|deps|install).*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ==============================================================================
# Build Targets
# ==============================================================================

build: build-core ## Build core components (orchestrator + alfa)

build-core: build-cellorg build-alfa ## Build core components (orchestrator + alfa)
	@echo "$(GREEN)✅ Core components built successfully!$(RESET)"

build-all: build-atomic build-omni build-cellorg build-agents build-alfa ## Build all modules (libs + core + agents)
	@echo "$(GREEN)✅ All modules built successfully!$(RESET)"

build-atomic: ## Build atomic module (VFS, VCR)
	@echo "$(BLUE)🔨 Building atomic...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@cd $(ATOMIC_DIR) && $(GO) build $(BUILD_FLAGS) ./...
	@echo "$(GREEN)  ✅ atomic built$(RESET)"

build-omni: ## Build omni module (OmniStore)
	@echo "$(BLUE)🔨 Building omni...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@cd $(OMNI_DIR) && $(GO) build $(BUILD_FLAGS) ./...
	@echo "$(GREEN)  ✅ omni built$(RESET)"

build-cellorg: ## Build cellorg module (Cell framework)
	@echo "$(BLUE)🔨 Building cellorg...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@cd $(CELLORG_DIR) && $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/orchestrator ./cmd/orchestrator
	@echo "$(GREEN)  ✅ cellorg built → bin/orchestrator$(RESET)"

build-agents: $(addprefix build-agent-,$(AGENT_LIST)) ## Build all agents (parallel with -j4)
	@echo "$(GREEN)  ✅ All agents built$(RESET)"

build-agent-%: ## Build individual agent (use: make build-agent-<name>)
	@echo "$(CYAN)  Building $*...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@if [ -f "$(AGENTS_DIR)/$*/main.go" ]; then \
		cd $(AGENTS_DIR)/$* && $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$* . ; \
	else \
		echo "$(YELLOW)  ⚠️  No main.go found for $*, skipping$(RESET)"; \
	fi

build-alfa: ## Build alfa module (AI workbench)
	@echo "$(BLUE)🔨 Building alfa...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@cd $(ALFA_DIR) && $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/alfa ./cmd/alfa
	@echo "$(GREEN)  ✅ alfa built → bin/alfa$(RESET)"

# ==============================================================================
# Test Targets
# ==============================================================================

test: ## Run comprehensive test suite with report generation
	@echo "$(BLUE)🧪 Running comprehensive test suite...$(RESET)"
	@$(BUILDER_DIR)/test-all.sh ci
	@echo "$(GREEN)✅ Test report: $(REPORT_DIR)/latest-report.md$(RESET)"

test-quick: ## Run quick tests (no race detection, no coverage)
	@echo "$(BLUE)🧪 Running quick tests...$(RESET)"
	@$(BUILDER_DIR)/test-all.sh quick

test-full: ## Run full tests with coverage reports
	@echo "$(BLUE)🧪 Running full test suite with coverage...$(RESET)"
	@$(BUILDER_DIR)/test-all.sh full
	@echo "$(GREEN)✅ Coverage reports: $(REPORT_DIR)/coverage/$(RESET)"

test-atomic: ## Run atomic module tests
	@echo "$(BLUE)🧪 Testing atomic...$(RESET)"
	@cd $(ATOMIC_DIR) && $(GO) test -v -race ./...

test-omni: ## Run omni module tests
	@echo "$(BLUE)🧪 Testing omni...$(RESET)"
	@cd $(OMNI_DIR) && $(GO) test -v -race ./...

test-cellorg: ## Run cellorg module tests
	@echo "$(BLUE)🧪 Testing cellorg...$(RESET)"
	@cd $(CELLORG_DIR) && $(GO) test -v -race ./...

test-agents: ## Run agent tests
	@echo "$(BLUE)🧪 Testing agents...$(RESET)"
	@cd $(AGENTS_DIR) && $(GO) test -v -race ./...

test-alfa: ## Run alfa module tests
	@echo "$(BLUE)🧪 Testing alfa...$(RESET)"
	@cd $(ALFA_DIR) && $(GO) test -v -race ./...

benchmark: ## Run benchmarks across all modules
	@echo "$(BLUE)⚡ Running benchmarks...$(RESET)"
	@cd $(OMNI_DIR) && $(GO) test -bench=. -benchmem ./...
	@cd $(CELLORG_DIR) && $(GO) test -bench=. -benchmem ./...

# ==============================================================================
# Module-Specific Targets
# ==============================================================================

atomic: build-atomic test-atomic ## Build and test atomic module

omni: build-omni test-omni ## Build and test omni module

cellorg: build-cellorg test-cellorg ## Build and test cellorg module

agents: build-agents test-agents ## Build and test all agents

alfa: build-alfa test-alfa ## Build and test alfa module

# ==============================================================================
# Code Quality Targets
# ==============================================================================

format: ## Format all Go code
	@echo "$(BLUE)🎨 Formatting code...$(RESET)"
	@cd $(CODE_DIR) && $(GO) fmt ./...
	@echo "$(GREEN)✅ Code formatted$(RESET)"

vet: ## Run go vet on all modules
	@echo "$(BLUE)🔍 Running go vet...$(RESET)"
	@cd $(CODE_DIR) && $(GO) vet ./...
	@echo "$(GREEN)✅ Vet complete$(RESET)"

lint: ## Run linters (requires golangci-lint)
	@echo "$(BLUE)🔍 Running linters...$(RESET)"
	@which golangci-lint > /dev/null || (echo "$(RED)❌ golangci-lint not installed$(RESET)" && exit 1)
	@cd $(CODE_DIR) && golangci-lint run ./...
	@echo "$(GREEN)✅ Linting complete$(RESET)"

check-atomic-deps: ## Verify atomic module has no external dependencies (stdlib only)
	@echo "$(BLUE)🔍 Checking atomic module dependencies...$(RESET)"
	@cd $(ATOMIC_DIR) && \
		deps=$$($(GO) list -m all | grep -v "^github.com/tenzoki/agen/atomic$$" || true) && \
		if [ -n "$$deps" ]; then \
			echo "$(RED)❌ CRITICAL: atomic module has external dependencies!$(RESET)" && \
			echo "$(RED)   Atomic MUST remain stdlib-only. Found:$(RESET)" && \
			echo "$$deps" && \
			echo "" && \
			echo "$(YELLOW)   This violates AGEN architecture principles.$(RESET)" && \
			echo "$(YELLOW)   See: guidelines/references/architecture.md$(RESET)" && \
			exit 1; \
		else \
			echo "$(GREEN)  ✅ atomic has zero external dependencies$(RESET)"; \
		fi

check: format vet check-atomic-deps ## Run format, vet, and dependency checks

# ==============================================================================
# Development Targets
# ==============================================================================

deps: ## Download and tidy all module dependencies
	@echo "$(BLUE)📦 Downloading dependencies...$(RESET)"
	@cd $(ATOMIC_DIR) && $(GO) mod download && $(GO) mod tidy
	@cd $(OMNI_DIR) && $(GO) mod download && $(GO) mod tidy
	@cd $(CELLORG_DIR) && $(GO) mod download && $(GO) mod tidy
	@cd $(AGENTS_DIR) && $(GO) mod download && $(GO) mod tidy
	@cd $(ALFA_DIR) && $(GO) mod download && $(GO) mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(RESET)"

install: build ## Install binaries to GOPATH/bin
	@echo "$(BLUE)📦 Installing binaries...$(RESET)"
	@cp $(BUILD_DIR)/* $(GOPATH)/bin/ 2>/dev/null || true
	@echo "$(GREEN)✅ Binaries installed to $(GOPATH)/bin$(RESET)"

# ==============================================================================
# Cleanup Targets
# ==============================================================================

clean: ## Clean build artifacts
	@echo "$(BLUE)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)/*
	@echo "$(GREEN)✅ Build artifacts cleaned$(RESET)"

clean-test: ## Clean test artifacts and reports
	@echo "$(BLUE)🧹 Cleaning test artifacts...$(RESET)"
	@rm -rf $(REPORT_DIR)/*.log
	@find $(CODE_DIR) -name "*.test" -delete
	@find $(CODE_DIR) -name "*.prof" -delete
	@echo "$(GREEN)✅ Test artifacts cleaned$(RESET)"

clean-reports: ## Clean all test reports in reflect/test-reports/
	@echo "$(BLUE)🧹 Cleaning test reports...$(RESET)"
	@rm -f $(REPORT_DIR)/*.md
	@rm -f $(REPORT_DIR)/*.txt
	@rm -f $(REPORT_DIR)/*.json
	@rm -f $(REPORT_DIR)/*.log
	@rm -rf $(REPORT_DIR)/coverage/
	@echo "$(GREEN)✅ Test reports cleaned$(RESET)"

clean-all: clean clean-test ## Clean everything (build + test artifacts)
	@echo "$(BLUE)🧹 Deep cleaning...$(RESET)"
	@find $(CODE_DIR) -name "demo_*" -type d -exec rm -rf {} + 2>/dev/null || true
	@find $(CODE_DIR) -name "temp_*" -type d -exec rm -rf {} + 2>/dev/null || true
	@find $(CODE_DIR) -name "tmp_*" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "$(GREEN)✅ All artifacts cleaned$(RESET)"

# ==============================================================================
# Info Targets
# ==============================================================================

info: ## Display project information
	@echo "$(CYAN)╔═══════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(CYAN)║                $(MAGENTA)AGEN Project Information$(CYAN)                  ║$(RESET)"
	@echo "$(CYAN)╚═══════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(YELLOW)Project:$(RESET)      $(PROJECT_NAME)"
	@echo "$(YELLOW)Version:$(RESET)      $(VERSION)"
	@echo "$(YELLOW)Commit:$(RESET)       $(COMMIT)"
	@echo "$(YELLOW)Build Time:$(RESET)   $(BUILD_TIME)"
	@echo "$(YELLOW)Go Version:$(RESET)   $(GO_VERSION)"
	@echo ""
	@echo "$(YELLOW)Paths:$(RESET)"
	@echo "  Root:           $(PROJECT_ROOT)"
	@echo "  Build Dir:      $(BUILD_DIR)"
	@echo "  Report Dir:     $(REPORT_DIR)"
	@echo ""
	@echo "$(YELLOW)Modules:$(RESET)"
	@echo "  atomic          $(shell cd $(ATOMIC_DIR) 2>/dev/null && find . -name "*_test.go" | wc -l | tr -d ' ') tests"
	@echo "  omni            $(shell cd $(OMNI_DIR) 2>/dev/null && find . -name "*_test.go" | wc -l | tr -d ' ') tests"
	@echo "  cellorg         $(shell cd $(CELLORG_DIR) 2>/dev/null && find . -name "*_test.go" | wc -l | tr -d ' ') tests"
	@echo "  agents          $(shell cd $(AGENTS_DIR) 2>/dev/null && find . -name "*_test.go" | wc -l | tr -d ' ') tests"
	@echo "  alfa            $(shell cd $(ALFA_DIR) 2>/dev/null && find . -name "*_test.go" | wc -l | tr -d ' ') tests"
	@echo ""
