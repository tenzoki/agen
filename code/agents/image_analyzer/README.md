# Image Analyzer

Image content analysis service for extracting metadata, detecting features, and classifying image types.

## Intent

Analyzes image files to extract metadata (dimensions, format, color space), detect visual features, identify image types, and assess quality. Provides comprehensive image characterization for downstream processing decisions.

## Usage

Input: `ChunkProcessingRequest` with image content
Output: `ProcessingResult` with image analysis metadata

Configuration:
- `enable_metadata`: Extract EXIF and image metadata
- `enable_feature_detection`: Detect visual features
- `enable_classification`: Classify image types
- `max_image_size`: Maximum image size for analysis
- `supported_formats`: Image formats to process

## Setup

Dependencies: No external dependencies (uses Go standard library image package)

Build:
```bash
go build -o bin/image_analyzer ./code/agents/image_analyzer
```

## Tests

Test file: `image_analyzer_test.go`

Tests cover metadata extraction, format detection, and feature analysis.

## Demo

No demo available
