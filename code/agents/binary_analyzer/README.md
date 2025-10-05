# Binary Analyzer

Comprehensive binary file analysis including file type detection, entropy calculation, and pattern recognition.

## Intent

Analyzes binary content to extract metadata, detect file types via magic bytes, calculate entropy for compression/encryption detection, and identify structural patterns. Provides detailed binary characterization for downstream processing decisions.

## Usage

Input: `ChunkProcessingRequest` with binary content
Output: `ProcessingResult` with comprehensive binary analysis

Configuration:
- `enable_hashing`: MD5/SHA256 hash calculation (default: true)
- `enable_entropy`: Shannon entropy analysis (default: true)
- `enable_magic_bytes`: File type detection (default: true)
- `max_analysis_size`: Analysis size limit in bytes (default: 10MB)
- `enable_structural`: Structural pattern analysis (default: true)
- `enable_compression`: Compression detection (default: false)

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/binary_analyzer ./code/agents/binary_analyzer
```

## Tests

Test file: `binary_analyzer_test.go`

Tests cover file type detection, entropy calculation, hash generation, and pattern recognition.

## Demo

No demo available
