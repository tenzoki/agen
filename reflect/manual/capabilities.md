# AGEN Capabilities Overview

**Generated:** 2025-10-14 10:40:33  

High-level guide to what AGEN can do.

---

## What is AGEN?

AGEN is a cell-based orchestration framework for building distributed processing pipelines with self-modification capabilities.

## Core Capabilities

### Document Processing
- Extract text from PDF, DOCX, XLSX, images
- OCR processing (native Tesseract + HTTP service)
- Multi-format document transformation
- Intelligent chunking and context enrichment

### Content Analysis
- Text analysis (sentiment, keywords, language)
- Structured data analysis (JSON, XML)
- Binary and image analysis
- Academic content processing

### Advanced Processing
- Named Entity Recognition (NER) with multilingual support
- PII detection and anonymization (GDPR-compliant)
- RAG (Retrieval-Augmented Generation) with vector embeddings
- Search indexing and full-text search

### Output Generation
- Document summarization
- Report generation with charts and tables
- Dataset creation and export
- Search-ready index generation

## Available Resources

- **Agents:** 27 specialized processing agents
- **Pipelines:** 8 processing workflows
- **Services:** 6 backend services
- **Analysis:** 6 analysis workflows
- **Synthesis:** 5 output generators

## How to Use

### Run a Cell
```bash
bin/orchestrator -config=workbench/config/cells/pipelines/<cell>.yaml
```

### Via Alfa
```bash
bin/alfa --enable-cellorg
> "Process documents in input/ through anonymization pipeline"
```

## Self-Modification

AGEN can modify its own codebase when enabled:
- Add new agents dynamically
- Modify cell configurations
- Update processing pipelines
- All changes are version-controlled via VCR

