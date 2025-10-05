# Text Chunker

Intelligent text segmentation service with multiple chunking strategies.

## Intent

Splits text into manageable chunks for processing, embedding, and analysis. Supports multiple strategies (paragraph-based, section-based, size-based, boundary-based, semantic) with overlap configuration for context preservation.

## Usage

Input: Text content with chunking parameters
Output: Text chunks with metadata and position information

Configuration:
- `chunk_size`: Target chunk size in characters/tokens
- `overlap_size`: Overlap between consecutive chunks
- `strategy`: Chunking strategy (paragraph_based/section_based/size_based/boundary_based/semantic)
- `preserve_sentences`: Avoid breaking mid-sentence
- `preserve_paragraphs`: Avoid breaking mid-paragraph

## Setup

Dependencies: No external dependencies

Build:
```bash
go build -o bin/text_chunker ./code/agents/text_chunker
```

## Tests

No tests implemented

## Demo

No demo available
