# Strategy Selector

Intelligent processing strategy selection based on content analysis.

## Intent

Analyzes document characteristics (format, size, complexity, language) and selects optimal processing strategies for extraction, chunking, and analysis. Enables adaptive pipeline configuration based on content properties.

## Usage

Input: Document metadata and content sample
Output: Selected processing strategies and configuration

Configuration:
- `strategy_rules`: Rule-based strategy selection
- `enable_ml_selection`: Use ML for strategy selection
- `confidence_threshold`: Minimum confidence for strategy
- `fallback_strategy`: Default strategy when uncertain

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/strategy_selector ./code/agents/strategy_selector
```

## Tests

Test file: `strategy_selector_test.go`

Tests cover strategy selection logic, rule evaluation, and confidence scoring.

## Demo

No demo available
