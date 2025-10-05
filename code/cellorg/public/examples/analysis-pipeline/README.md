# Analysis Pipeline Examples

This directory contains comprehensive examples for the GOX Framework analysis agents, demonstrating file format analysis, content validation, and metadata extraction capabilities.

## Overview

The analysis pipeline showcases four specialized analyzer agents:

- **Binary Analyzer** (`binary_analyzer`) - Binary file analysis and string extraction
- **JSON Analyzer** (`json_analyzer`) - JSON parsing, validation, and schema checking
- **XML Analyzer** (`xml_analyzer`) - XML processing, validation, and namespace handling
- **Image Analyzer** (`image_analyzer`) - Image analysis, metadata extraction, and format detection

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   File Input    │───▶│  Format Router  │───▶│ Analyzer Agent  │
│   - Multiple    │    │  - Auto-detect  │    │ - Format-specific│
│   - Formats     │    │  - Route to     │    │ - Deep analysis │
└─────────────────┘    │    appropriate  │    └─────────────────┘
                       │    analyzer     │
                       └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Analysis       │◀───│   Results       │───▶│    Metadata     │
│  Report         │    │   Aggregator    │    │   Collector     │
│  - Structured   │    │   - Combine     │    │   - Extract     │
│  - Detailed     │    │   - Validate    │    │   - Enrich      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Features Demonstrated

### 1. Binary File Analysis
- Magic byte detection for file type identification
- String extraction from binary content
- Executable analysis (PE, ELF headers)
- Security analysis (entropy, suspicious patterns)
- Size limit enforcement and validation

### 2. JSON Data Analysis
- JSON syntax validation and parsing
- Schema validation against JSON Schema
- Path extraction and value analysis
- Nested structure analysis
- Format standardization and pretty-printing

### 3. XML Document Analysis
- XML syntax validation and well-formedness checking
- DTD and XSD schema validation
- Namespace processing and resolution
- Element and attribute extraction
- Content structure analysis

### 4. Image Content Analysis
- Image format detection (JPEG, PNG, GIF, etc.)
- Metadata extraction (EXIF, IPTC, XMP)
- Dimension and color space analysis
- Quality assessment
- Thumbnail generation

## Quick Start

### Prerequisites

1. **Build the GOX framework and agents:**
   ```bash
   cd /path/to/gox
   make build
   ```

2. **Ensure required binaries exist:**
   ```bash
   ls build/
   # Should include: binary_analyzer, json_analyzer, xml_analyzer, image_analyzer
   ```

### Running Examples

1. **All analyzers demo:**
   ```bash
   cd examples/analysis-pipeline
   ./run_analysis_demo.sh
   ```

2. **Specific analyzer:**
   ```bash
   # Binary analysis only
   ./run_analysis_demo.sh --analyzer=binary

   # JSON analysis only
   ./run_analysis_demo.sh --analyzer=json

   # XML analysis only
   ./run_analysis_demo.sh --analyzer=xml

   # Image analysis only
   ./run_analysis_demo.sh --analyzer=image
   ```

3. **Custom input directory:**
   ```bash
   ./run_analysis_demo.sh --input=/path/to/your/files
   ```

## Example Files Structure

```
examples/analysis-pipeline/
├── README.md                     # This documentation
├── run_analysis_demo.sh          # Main demo script
├── analysis_demo.go              # Go demo implementation
├── cell_configs/                 # Cell configuration files
│   ├── binary_analysis_cell.yaml
│   ├── json_analysis_cell.yaml
│   ├── xml_analysis_cell.yaml
│   ├── image_analysis_cell.yaml
│   └── complete_analysis_cell.yaml
├── input/                        # Sample input files
│   ├── binary/
│   │   ├── sample.exe
│   │   ├── library.so
│   │   ├── document.pdf
│   │   └── archive.zip
│   ├── json/
│   │   ├── simple.json
│   │   ├── complex.json
│   │   ├── schema.json
│   │   └── malformed.json
│   ├── xml/
│   │   ├── simple.xml
│   │   ├── namespaced.xml
│   │   ├── schema.xsd
│   │   └── malformed.xml
│   └── images/
│       ├── photo.jpg
│       ├── diagram.png
│       ├── icon.gif
│       └── corrupted.bmp
├── output/                       # Analysis results
│   ├── binary_analysis/
│   ├── json_analysis/
│   ├── xml_analysis/
│   └── image_analysis/
└── schemas/                      # Validation schemas
    ├── config_schema.json
    ├── report_schema.xsd
    └── metadata_schema.json
```

## Agent Demonstrations

### Binary Analyzer Example

**Input:** Mixed binary files (executables, libraries, documents)
**Process:**
- File type detection using magic bytes
- String extraction with configurable minimum length
- Executable header analysis
- Security assessment (entropy, packing detection)

**Sample Configuration:**
```yaml
- id: "binary-analyzer-001"
  agent_type: "binary_analyzer"
  ingress: "file:examples/analysis-pipeline/input/binary/*"
  egress: "pub:binary-analysis-results"
  config:
    deep_scan: true
    extract_strings: true
    min_string_length: 4
    max_file_size: "50MB"
    security_analysis: true
    detect_packing: true
```

**Sample Output:**
```json
{
  "file_info": {
    "path": "/input/binary/sample.exe",
    "size": 1048576,
    "mime_type": "application/x-executable",
    "format": "PE32"
  },
  "strings": {
    "extracted": ["kernel32.dll", "CreateFile", "ReadFile"],
    "count": 156,
    "encoding": "ascii"
  },
  "security": {
    "entropy": 6.8,
    "packed": false,
    "suspicious_strings": [],
    "risk_level": "low"
  },
  "metadata": {
    "architecture": "x86",
    "subsystem": "console",
    "compilation_timestamp": "2024-09-15T10:30:00Z"
  }
}
```

### JSON Analyzer Example

**Input:** JSON files with various structures and validation requirements
**Process:**
- Syntax validation and error reporting
- Schema validation against JSON Schema
- Path extraction and value type analysis
- Structure optimization suggestions

**Sample Configuration:**
```yaml
- id: "json-analyzer-001"
  agent_type: "json_analyzer"
  ingress: "file:examples/analysis-pipeline/input/json/*"
  egress: "pub:json-analysis-results"
  config:
    validate_syntax: true
    validate_schema: true
    schema_path: "examples/analysis-pipeline/schemas/"
    extract_paths: true
    pretty_print: true
    max_depth: 10
```

**Sample Output:**
```json
{
  "validation": {
    "syntax_valid": true,
    "schema_valid": true,
    "schema_used": "config_schema.json",
    "errors": []
  },
  "structure": {
    "total_keys": 45,
    "max_depth": 4,
    "arrays": 3,
    "objects": 8,
    "primitives": 34
  },
  "paths": [
    "$.config.database.host",
    "$.config.database.port",
    "$.config.features[*].enabled"
  ],
  "analysis": {
    "complexity": "moderate",
    "recommendations": ["Consider flattening nested config objects"]
  }
}
```

### XML Analyzer Example

**Input:** XML documents with various schemas and namespaces
**Process:**
- Well-formedness validation
- DTD/XSD schema validation
- Namespace resolution
- Element and attribute extraction

**Sample Configuration:**
```yaml
- id: "xml-analyzer-001"
  agent_type: "xml_analyzer"
  ingress: "file:examples/analysis-pipeline/input/xml/*"
  egress: "pub:xml-analysis-results"
  config:
    validate_wellformed: true
    validate_schema: true
    schema_path: "examples/analysis-pipeline/schemas/"
    extract_elements: true
    namespace_aware: true
    preserve_whitespace: false
```

**Sample Output:**
```json
{
  "validation": {
    "wellformed": true,
    "schema_valid": true,
    "schema_type": "XSD",
    "errors": []
  },
  "structure": {
    "root_element": "configuration",
    "namespaces": {
      "": "http://example.com/config",
      "ext": "http://example.com/extensions"
    },
    "elements": 23,
    "attributes": 15,
    "text_nodes": 18
  },
  "content": {
    "extracted_text": "Configuration values...",
    "cdata_sections": 2,
    "processing_instructions": 1
  }
}
```

### Image Analyzer Example

**Input:** Various image formats with metadata
**Process:**
- Format detection and validation
- Metadata extraction (EXIF, IPTC, XMP)
- Dimension and color analysis
- Quality assessment

**Sample Configuration:**
```yaml
- id: "image-analyzer-001"
  agent_type: "image_analyzer"
  ingress: "file:examples/analysis-pipeline/input/images/*"
  egress: "pub:image-analysis-results"
  config:
    extract_metadata: true
    extract_exif: true
    analyze_colors: true
    generate_thumbnail: true
    max_image_size: "100MB"
    thumbnail_size: "200x200"
```

**Sample Output:**
```json
{
  "image_info": {
    "format": "JPEG",
    "dimensions": {
      "width": 1920,
      "height": 1080,
      "aspect_ratio": 1.78
    },
    "color_space": "RGB",
    "bit_depth": 8,
    "compression": "JPEG"
  },
  "metadata": {
    "exif": {
      "camera_make": "Canon",
      "camera_model": "EOS 5D Mark IV",
      "capture_time": "2024-09-15T14:30:00Z",
      "gps_location": {
        "latitude": 52.5200,
        "longitude": 13.4050
      }
    },
    "file_size": 2458624,
    "created": "2024-09-15T14:30:00Z"
  },
  "analysis": {
    "quality_score": 0.92,
    "dominant_colors": ["#2C3E50", "#ECF0F1", "#E74C3C"],
    "histogram": {
      "red": [125, 200, 180, 95, ...],
      "green": [110, 185, 170, 90, ...],
      "blue": [95, 170, 155, 85, ...]
    }
  }
}
```

## Integration Pipeline Example

### Complete Analysis Cell Configuration

```yaml
cell:
  id: "pipeline:complete-file-analysis"
  description: "Comprehensive file analysis pipeline"

  agents:
    # Router agent (determines file type and routes to appropriate analyzer)
    - id: "file-router-001"
      agent_type: "strategy_selector"
      ingress: "file:input/*"
      egress: "route:analysis"
      config:
        routing_strategy: "file_type"

    # Binary file analyzer
    - id: "binary-analyzer-001"
      agent_type: "binary_analyzer"
      ingress: "route:analysis:binary"
      egress: "pub:binary-results"
      config:
        deep_scan: true
        extract_strings: true
        security_analysis: true

    # JSON analyzer
    - id: "json-analyzer-001"
      agent_type: "json_analyzer"
      ingress: "route:analysis:json"
      egress: "pub:json-results"
      config:
        validate_syntax: true
        validate_schema: true

    # XML analyzer
    - id: "xml-analyzer-001"
      agent_type: "xml_analyzer"
      ingress: "route:analysis:xml"
      egress: "pub:xml-results"
      config:
        validate_wellformed: true
        validate_schema: true

    # Image analyzer
    - id: "image-analyzer-001"
      agent_type: "image_analyzer"
      ingress: "route:analysis:image"
      egress: "pub:image-results"
      config:
        extract_metadata: true
        analyze_colors: true

    # Results aggregator
    - id: "results-aggregator-001"
      agent_type: "report_generator"
      ingress: "sub:binary-results,json-results,xml-results,image-results"
      egress: "file:output/analysis-report.json"
      config:
        report_format: "comprehensive"
        include_statistics: true
```

## Demo Workflow

### Step 1: Environment Setup
```bash
# Create temporary directories
mkdir -p /tmp/gox-analysis-demo/{input,output,logs}

# Copy sample files
cp -r examples/analysis-pipeline/input/* /tmp/gox-analysis-demo/input/
```

### Step 2: Start Analysis Pipeline
```bash
# Start GOX with analysis cell
./build/gox cell run examples/analysis-pipeline/cell_configs/complete_analysis_cell.yaml
```

### Step 3: Monitor Processing
```bash
# Watch analysis results
tail -f /tmp/gox-analysis-demo/logs/analysis.log

# Check output files
ls -la /tmp/gox-analysis-demo/output/
```

### Step 4: Review Results
```bash
# View comprehensive analysis report
cat /tmp/gox-analysis-demo/output/analysis-report.json | jq '.'

# View individual analyzer results
ls /tmp/gox-analysis-demo/output/*/
```

## Performance Characteristics

### Throughput Benchmarks
- **Binary Analyzer**: ~50 files/second (average 2MB files)
- **JSON Analyzer**: ~200 files/second (average 100KB files)
- **XML Analyzer**: ~150 files/second (average 200KB files)
- **Image Analyzer**: ~25 files/second (average 5MB images)

### Resource Usage
- **Memory**: ~100MB base + 2-10MB per concurrent file
- **CPU**: Moderate (varies by analysis depth)
- **I/O**: High for large binary/image files

## Testing

### Unit Tests
```bash
# Test all analyzers
go test ./agents/binary_analyzer ./agents/json_analyzer ./agents/xml_analyzer ./agents/image_analyzer -v

# Test specific analyzer
go test ./agents/json_analyzer -v -run TestJSONValidation
```

### Integration Tests
```bash
# Run complete analysis pipeline test
./examples/analysis-pipeline/test_analysis_pipeline.sh

# Test with custom input
./examples/analysis-pipeline/test_analysis_pipeline.sh --input=/path/to/test/files
```

## Customization

### Custom File Type Detection
```go
// Add custom file type detection
func detectCustomFormat(data []byte) string {
    // Custom magic byte detection
    if bytes.HasPrefix(data, []byte("CUSTOM")) {
        return "application/x-custom"
    }
    return "unknown"
}
```

### Custom Analysis Rules
```yaml
# Custom validation rules for JSON analyzer
config:
  custom_validators:
    - rule: "required_fields"
      fields: ["id", "name", "version"]
    - rule: "value_constraints"
      constraints:
        version: "^\\d+\\.\\d+\\.\\d+$"
```

### Custom Metadata Extraction
```go
// Custom metadata extractor for images
func extractCustomMetadata(imgPath string) map[string]interface{} {
    // Custom logic for proprietary formats
    return map[string]interface{}{
        "custom_field": extractCustomField(imgPath),
        "proprietary_data": parseProprietaryFormat(imgPath),
    }
}
```

## Troubleshooting

### Common Issues

1. **File type not detected**
   - Check magic byte definitions
   - Verify file integrity
   - Update file type detection rules

2. **Schema validation fails**
   - Check schema file paths
   - Verify schema syntax
   - Update schema definitions

3. **Image analysis errors**
   - Check image file corruption
   - Verify supported formats
   - Check memory limits for large images

4. **Performance issues**
   - Adjust concurrency settings
   - Optimize file size limits
   - Monitor resource usage

### Debug Mode
```bash
# Enable debug logging
export GOX_LOG_LEVEL=debug

# Run specific analyzer in debug mode
GOX_DEBUG=true ./build/json_analyzer --debug
```

## Next Steps

1. **Extend analyzers**: Add support for additional file formats
2. **Custom validators**: Implement domain-specific validation rules
3. **Performance optimization**: Tune for specific file types and sizes
4. **Integration**: Connect to external validation services
5. **Monitoring**: Add metrics collection and alerting

For more information, see the [main GOX documentation](../../README.md) and individual agent documentation in `agents/*/README.md`.