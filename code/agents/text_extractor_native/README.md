# Text Extractor Native

Native text extraction using local Tesseract OCR for images and plain text files.

## Intent

Provides simple, fast text extraction without external dependencies. Uses local Tesseract installation for OCR on images (.png, .jpg, .tiff) and direct reading for plain text files. Low latency alternative to HTTP OCR service for smaller workloads.

## Usage

Input: File paths for text/image files
Output: Extracted text with quality metrics and metadata

Configuration:
- `enable_ocr`: Enable OCR processing (default: true)
- `ocr_languages`: Tesseract language packs (default: ["eng"])
- `include_metadata`: Include extraction metadata (default: true)
- `quality_threshold`: Minimum OCR quality threshold (default: 0.8)

## Setup

Dependencies:
- Tesseract OCR (local installation required)
  - macOS: `brew install tesseract`
  - Ubuntu: `apt-get install tesseract-ocr`
- Additional language packs as needed

Verify installation:
```bash
tesseract --version
tesseract --list-langs
```

Build:
```bash
go build -o bin/text_extractor_native ./code/agents/text_extractor_native
```

## Tests

Test file: `text_extractor_native_test.go`

Tests cover OCR extraction, text file reading, and metadata generation.

## Demo

No demo available
