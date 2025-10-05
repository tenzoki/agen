# Text Analyzer

Comprehensive text analysis including sentiment, keywords, language detection, and content categorization.

## Intent

Analyzes text chunks to extract sentiment (positive/negative/neutral), keywords (TF-IDF), language identification, content categories, and statistical metrics. Provides NLP-based insights for intelligent text processing and search enhancement.

## Usage

Input: `ChunkProcessingRequest` with text content
Output: `ProcessingResult` with comprehensive text analysis

Configuration:
- `enable_nlp`: Enable NLP processing (default: false, lightweight)
- `enable_sentiment`: Sentiment analysis (default: true)
- `enable_keywords`: Keyword extraction (default: true)
- `max_lines`: Maximum lines to process (default: 10000)
- `max_keywords`: Maximum keywords to extract (default: 20)

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/text_analyzer ./code/agents/text_analyzer
```

## Tests

Test file: `text_analyzer_test.go`

Tests cover sentiment analysis, keyword extraction, and language detection.

## Demo

No demo available
