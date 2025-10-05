# XML Analyzer

Comprehensive XML content analysis including validation, namespace resolution, and format detection.

## Intent

Analyzes XML content to provide well-formedness validation, namespace analysis with prefix resolution, element frequency statistics, format detection (HTML, SOAP, RSS, Maven POM), structure analysis (depth, complexity), and schema inference for intelligent XML processing.

## Usage

Input: `ChunkProcessingRequest` with XML content
Output: `ProcessingResult` with comprehensive XML analysis

Configuration:
- `enable_validation`: XML well-formedness validation (default: true)
- `enable_namespace_analysis`: Namespace resolution (default: true)
- `enable_format_detection`: Detect XML formats (default: true)
- `max_depth`: Maximum nesting depth (default: 50)
- `generate_schema`: Generate XSD from structure (default: false)

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/xml_analyzer ./code/agents/xml_analyzer
```

## Tests

No tests implemented

## Demo

No demo available
