# Text Transformer

Text transformation service for cleaning, normalizing, and reformatting text content.

## Intent

Applies various transformations to text including case conversion, whitespace normalization, special character handling, encoding fixes, and format conversion. Prepares text for downstream NLP and analysis tasks.

## Usage

Input: Text content with transformation specifications
Output: Transformed text with operation metadata

Configuration:
- `transformations`: List of transformations to apply
- `preserve_formatting`: Maintain structural formatting
- `encoding`: Target text encoding
- `normalize_unicode`: Unicode normalization
- `remove_special_chars`: Strip special characters

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/text_transformer ./code/agents/text_transformer
```

## Tests

Test file: `text_transformer_test.go`

Tests cover various text transformations and encoding operations.

## Demo

No demo available
