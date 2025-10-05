# Report Generator

Generates formatted reports from processed document data with multiple output formats.

## Intent

Creates comprehensive reports aggregating analysis results, statistics, and insights from document processing pipelines. Supports multiple formats (PDF, HTML, Markdown) with customizable templates and visualizations.

## Usage

Input: Processing results and report configuration
Output: Formatted report in specified format

Configuration:
- `output_format`: Report format (pdf/html/markdown)
- `template_path`: Report template location
- `include_visualizations`: Generate charts and graphs
- `include_statistics`: Include processing statistics
- `branding`: Custom branding configuration

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/report_generator ./code/agents/report_generator
```

## Tests

No tests implemented

## Demo

No demo available
