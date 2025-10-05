# GOX Framework - Reporting Pipeline Examples

This directory contains examples demonstrating the reporting and aggregation capabilities of the GOX Framework, focusing on report generation, summary creation, and metadata collection.

## Overview

The reporting pipeline showcases three key agent types:
- **report_generator**: Creates comprehensive analysis reports
- **summary_generator**: Generates concise summaries from processed data
- **metadata_collector**: Collects and aggregates metadata from various sources

## Agents Covered

### 1. Report Generator (`report_generator`)
Generates structured reports from processed data streams. Supports multiple output formats and customizable report templates.

**Key Features:**
- Multiple report formats (JSON, HTML, PDF, XML)
- Template-based report generation
- Data aggregation and statistics
- Cross-reference analysis
- Executive summary generation

**Use Cases:**
- Analysis result compilation
- Performance reporting
- Audit trail generation
- Business intelligence reports

### 2. Summary Generator (`summary_generator`)
Creates concise summaries from large datasets or complex analysis results.

**Key Features:**
- Intelligent content extraction
- Key finding identification
- Statistical summarization
- Trend analysis
- Risk assessment summaries

**Use Cases:**
- Executive dashboards
- Alert notifications
- Status updates
- Progress reports

### 3. Metadata Collector (`metadata_collector`)
Aggregates metadata from various processing stages and data sources.

**Key Features:**
- Cross-source metadata aggregation
- Temporal metadata tracking
- Provenance information
- Quality metrics collection
- Processing statistics

**Use Cases:**
- Data lineage tracking
- Quality assurance
- Compliance reporting
- Performance monitoring

## Directory Structure

```
reporting-pipeline/
├── README.md                    # This file
├── run_reporting_demo.sh        # Demo execution script
├── input/                       # Sample input data
│   ├── analysis-results/        # Analysis outputs for reporting
│   ├── processed-data/          # Processed datasets
│   └── raw-metadata/           # Raw metadata files
├── config/                      # Cell configurations
│   ├── report_generation_cell.yaml
│   ├── summary_generation_cell.yaml
│   ├── metadata_collection_cell.yaml
│   └── complete_reporting_cell.yaml
└── schemas/                     # Report and data schemas
    ├── report-schema.json
    ├── summary-schema.json
    └── metadata-schema.json
```

## Quick Start

### Run All Reporting Examples
```bash
./run_reporting_demo.sh
```

### Run Specific Agent
```bash
# Report generation only
./run_reporting_demo.sh --agent=report

# Summary generation only
./run_reporting_demo.sh --agent=summary

# Metadata collection only
./run_reporting_demo.sh --agent=metadata
```

### Custom Input Directory
```bash
./run_reporting_demo.sh --input=/path/to/your/data
```

## Example Configurations

### Basic Report Generation
```yaml
cell:
  id: "reporting:basic-reports"
  description: "Basic report generation cell"

  agents:
    - id: "report-generator-001"
      agent_type: "report_generator"
      ingress: "file:input/analysis-results/*"
      egress: "file:output/reports/"
      config:
        report_format: "html"
        template: "standard"
        include_charts: true
        include_statistics: true
        max_items_per_section: 100
```

### Advanced Summary Generation
```yaml
cell:
  id: "reporting:intelligent-summaries"
  description: "Intelligent summary generation cell"

  agents:
    - id: "summary-generator-001"
      agent_type: "summary_generator"
      ingress: "sub:analysis-results,processing-stats"
      egress: "pub:summaries"
      config:
        summary_type: "executive"
        max_length: 500
        include_trends: true
        highlight_anomalies: true
        confidence_threshold: 0.8
        key_findings_limit: 5
```

### Comprehensive Metadata Collection
```yaml
cell:
  id: "reporting:metadata-aggregation"
  description: "Comprehensive metadata collection cell"

  agents:
    - id: "metadata-collector-001"
      agent_type: "metadata_collector"
      ingress: "route:metadata-sources"
      egress: "file:output/metadata-report.json"
      config:
        collection_scope: "comprehensive"
        include_provenance: true
        include_quality_metrics: true
        include_processing_stats: true
        temporal_tracking: true
        cross_reference_analysis: true
```

## Complete Reporting Pipeline

The complete pipeline demonstrates end-to-end reporting workflow:

1. **Data Collection**: Aggregates processed results from multiple sources
2. **Metadata Extraction**: Collects comprehensive metadata about processing
3. **Summary Generation**: Creates intelligent summaries of findings
4. **Report Compilation**: Generates formatted reports with visualizations
5. **Distribution**: Outputs reports in multiple formats for different audiences

```yaml
cell:
  id: "reporting:complete-pipeline"
  description: "End-to-end reporting pipeline"

  agents:
    - id: "data-router-001"
      agent_type: "strategy_selector"
      ingress: "file:input/**/*"
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
      ingress: "sub:metadata,processing-results"
      egress: "pub:summaries"
      config:
        summary_type: "technical"
        max_length: 1000

    - id: "report-generator-001"
      agent_type: "report_generator"
      ingress: "sub:metadata,summaries"
      egress: "file:output/final-report.html"
      config:
        report_format: "html"
        template: "comprehensive"
        include_charts: true
```

## Input Data Format

### Analysis Results
Expected format for analysis results input:
```json
{
  "analysis_id": "binary-analysis-001",
  "timestamp": "2024-09-27T10:00:00Z",
  "file_path": "/data/sample.exe",
  "results": {
    "file_type": "PE32 executable",
    "security_analysis": {
      "entropy": 7.2,
      "suspicious_strings": 15,
      "packer_detected": false
    },
    "metadata": {
      "size": 2048576,
      "md5": "abc123...",
      "created": "2024-09-20T08:30:00Z"
    }
  }
}
```

### Processing Statistics
```json
{
  "pipeline_id": "analysis-pipeline-001",
  "start_time": "2024-09-27T09:00:00Z",
  "end_time": "2024-09-27T10:30:00Z",
  "files_processed": 45,
  "success_rate": 98.5,
  "average_processing_time": 125.3,
  "errors": [
    {
      "file": "corrupted.bin",
      "error": "Unable to read file header",
      "timestamp": "2024-09-27T09:15:32Z"
    }
  ]
}
```

## Output Examples

### Generated Report Structure
```html
<!DOCTYPE html>
<html>
<head>
    <title>GOX Analysis Report - 2024-09-27</title>
</head>
<body>
    <h1>Analysis Report Summary</h1>

    <section id="executive-summary">
        <h2>Executive Summary</h2>
        <p>Processed 45 files with 98.5% success rate...</p>
    </section>

    <section id="detailed-findings">
        <h2>Detailed Findings</h2>
        <!-- Charts and detailed analysis -->
    </section>

    <section id="metadata">
        <h2>Processing Metadata</h2>
        <!-- Provenance and quality metrics -->
    </section>
</body>
</html>
```

### Summary Output
```json
{
  "summary": {
    "overview": "Analysis of 45 binary files completed successfully",
    "key_findings": [
      "2 files detected with potential security concerns",
      "Average entropy score: 6.8/10",
      "No packed executables found"
    ],
    "statistics": {
      "total_files": 45,
      "success_rate": "98.5%",
      "processing_time": "1h 30m"
    },
    "recommendations": [
      "Review files with entropy > 7.5",
      "Consider additional string analysis for suspicious files"
    ]
  }
}
```

## Requirements

- GOX Framework v3+
- Built reporting agents:
  - `build/report_generator`
  - `build/summary_generator`
  - `build/metadata_collector`

## Building Required Agents

```bash
cd ../../
make build-reporting  # or individual builds:
go build -o build/report_generator ./agents/report_generator
go build -o build/summary_generator ./agents/summary_generator
go build -o build/metadata_collector ./agents/metadata_collector
```

## Advanced Usage

### Custom Report Templates
Create custom report templates in the templates directory and reference them in your cell configuration.

### Multi-format Output
Configure multiple report generators with different output formats:
- HTML for web viewing
- PDF for formal documents
- JSON for programmatic access
- XML for system integration

### Real-time Reporting
Use pub/sub messaging for real-time report updates and live dashboards.

### Integration with External Systems
Export reports to external business intelligence tools, notification systems, or data warehouses.