# File Chunking Pipeline Examples

This directory contains practical examples of using the file chunking pipeline with the three core operators:

1. **File Splitter** - Splits large files into manageable chunks
2. **Chunk Processor** - Processes chunks with specialized analysis
3. **Chunk Synthesizer** - Aggregates results into meaningful outputs

## Quick Start

```bash
# Build all operators
make build

# Run a simple text file processing example
./examples/file_chunking/basic_text_processing.sh

# Run a comprehensive document analysis example
./examples/file_chunking/document_analysis.sh

# Run a data processing pipeline example
./examples/file_chunking/data_pipeline.sh
```

## Examples Overview

### 1. Basic Text Processing (`basic_text_processing.sh`)
- Split a text document into line-based chunks
- Extract keywords and sentiment from each chunk
- Create a document summary with aggregated insights

### 2. Large Document Analysis (`document_analysis.sh`)
- Process a large PDF or document
- Perform semantic splitting for better context preservation
- Generate comprehensive analysis report with charts and tables

### 3. Data Processing Pipeline (`data_pipeline.sh`)
- Process large JSON/CSV datasets
- Split data into logical chunks
- Create structured datasets and search indices

### 4. Multi-Format Content (`multi_format.sh`)
- Handle mixed content types (text, JSON, XML, images)
- Demonstrate different processing strategies per file type
- Generate unified metadata collection

### 5. Performance Testing (`performance_test.sh`)
- Process large files to demonstrate scalability
- Show parallel processing capabilities
- Generate performance metrics

## Architecture Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│File Splitter│───▶│Chunk Processor│───▶│Chunk Synthesizer│
└─────────────┘    └─────────────┘    └─────────────┘
       │                    │                    │
       ▼                    ▼                    ▼
   chunk files         processed data        final outputs
   - chunk_0001.txt    - keywords           - document.json
   - chunk_0002.txt    - sentiment          - search_index.json
   - chunk_0003.txt    - metadata           - report.json
                                            - dataset.json
```

## Configuration Examples

Each example includes configuration files showing how to:
- Customize chunk sizes and splitting strategies
- Configure processing options per content type
- Set up synthesis parameters for different output formats
- Optimize performance settings

## Output Examples

See the `outputs/` directory for sample results from each example, including:
- Document summaries with extracted insights
- Search indices for content discovery
- Analysis reports with visualizations
- Structured datasets for further processing

## Integration Patterns

Examples demonstrate common integration patterns:
- **Batch Processing** - Process multiple files sequentially
- **Stream Processing** - Handle continuous file streams
- **Error Handling** - Robust error recovery and retry logic
- **Monitoring** - Progress tracking and metrics collection