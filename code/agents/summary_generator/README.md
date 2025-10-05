# Summary Generator

Automatic text summarization service using extractive and abstractive techniques.

## Intent

Generates concise summaries from document text using extractive (sentence selection) and abstractive (text generation) methods. Provides adjustable summary length, multi-document summarization, and key point extraction.

## Usage

Input: Text content with summary parameters
Output: Generated summary with key points and metadata

Configuration:
- `summarization_method`: extractive/abstractive/hybrid
- `summary_length`: Target summary length (words/sentences)
- `preserve_formatting`: Maintain text structure
- `extract_key_points`: Generate bullet-point summary
- `language`: Summary language

## Setup

Dependencies: No external dependencies (extractive method)

Build:
```bash
go build -o bin/summary_generator ./code/agents/summary_generator
```

## Tests

Test file: `summary_generator_test.go`

Tests cover extractive summarization, sentence scoring, and key point extraction.

## Demo

No demo available
