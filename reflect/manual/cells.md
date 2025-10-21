# AGEN Cell Patterns

**Generated:** 2025-10-14 10:40:33  
**Cells:** 25  

Available cell configurations organized by purpose.

---

## Processing Pipelines

### Anonymization Pipeline

Provides GDPR-compliant document anonymization through multilingual Named Entity Recognition (XLM-RoBERTa), deterministic pseudonymization with SHA256-based mapping, and persistent bidirectional storage for entity mappings with project-level isolation.

**Config:** `workbench/config/cells/pipelines/anonymization-pipeline.yaml`

### Document Processing Pipeline

Processes multi-format documents (PDF, DOCX, TXT) through strategy-aware chunking that adapts to document structure, enabling context-preserving segmentation for downstream RAG, analysis, and storage workflows.

**Config:** `workbench/config/cells/pipelines/document-processing-pipeline.yaml`

### Fast Document Processing Pipeline

Provides high-throughput document processing by disabling OCR, using direct text chunking without strategy selection, and minimizing metadata overhead to achieve minimal latency for native text documents.

**Config:** `workbench/config/cells/pipelines/fast-document-processing-pipeline.yaml`

### File Chunking Pipeline

Processes large files through semantic-aware splitting, parallel chunk analysis with keyword and sentiment extraction, and comprehensive synthesis to generate unified document summaries with charts and statistics.

**Config:** `workbench/config/cells/pipelines/file-chunking-pipeline.yaml`

### File Transform Pipeline

Demonstrates basic file processing workflow with file monitoring, text transformation (uppercase with metadata), and templated output generation to illustrate fundamental pipeline patterns and agent dependencies.

**Config:** `workbench/config/cells/pipelines/file-transform-pipeline.yaml`

### Intelligent Document Processing Pipeline

Provides advanced document processing through OCR-enabled text extraction, content-driven strategy selection, adaptive text chunking, and multi-dimensional context enrichment (positional, semantic, structural, relational) to generate knowledge-graph-ready document chunks.

**Config:** `workbench/config/cells/pipelines/intelligent-document-processing-pipeline.yaml`

### Research Paper Processing Pipeline

Processes academic research papers through high-quality OCR extraction, academic-specific chunking strategies with large overlap, and deep context enrichment (depth 5) to preserve citation relationships, section boundaries, and semantic connections for research knowledge bases.

**Config:** `workbench/config/cells/pipelines/research-paper-processing-pipeline.yaml`

### Text Extraction Pipeline

Extracts text from multi-format documents (PDF, DOCX, XLSX, TXT, images) through OCR-enabled processing with multilingual support, quality thresholds, and JSON output to enable downstream text analysis, indexing, and archival workflows.

**Config:** `workbench/config/cells/pipelines/text-extraction-pipeline.yaml`

---

## Backend Services

### HTTP OCR Extraction

Delegates OCR processing to external containerized HTTP service to decouple compute-intensive OCR workloads from the pipeline, enabling horizontal scaling of OCR capacity and simplified deployment without local Tesseract dependencies.

**Config:** `workbench/config/cells/services/ocr.yaml`

### HTTP OCR Production Extraction

Provides production-grade OCR processing through load-balanced HTTP service with extended timeouts, increased retry logic, and disabled debug mode to support large-scale document processing with high availability requirements.

**Config:** `workbench/config/cells/services/ocr-production.yaml`

### Knowledge Backend RAG

Provides Retrieval-Augmented Generation backend for AI assistants through semantic code search using OpenAI embeddings, vector similarity search, and integrated storage to deliver contextually relevant code snippets with surrounding lines for enhanced AI responses.

**Config:** `workbench/config/cells/services/rag.yaml`

### Native Text Extraction

Provides embedded OCR processing using native Tesseract library with image preprocessing (deskew, noise removal, contrast enhancement) to eliminate external service dependencies while supporting multilingual text extraction.

**Config:** `workbench/config/cells/services/native-text-extraction.yaml`

### Storage Cell

Demonstrates integration pattern for external storage service through HTTP API, enabling agents to leverage centralized storage (indexing, relationships, backups) while maintaining clean separation between data processing and storage concerns.

**Config:** `workbench/config/cells/services/storage.yaml`

### Storage Service

Provides centralized storage service with HTTP API supporting key-value operations, graph relationships, file storage, and full-text search to enable multiple cells to share persistent data without local storage dependencies.

**Config:** `workbench/config/cells/services/storage-service.yaml`

---

## Content Analysis

### Academic Document Analysis Pipeline

Processes academic papers and research documents through OCR-enabled text extraction, structured document processing with academic paper strategies, chunk-level analysis, and comprehensive summarization to produce structured analysis outputs suitable for research knowledge bases.

**Config:** `workbench/config/cells/analysis/academic-analysis.yaml`

### Binary Media Analysis Processing

Provides comprehensive analysis of binary files and images through parallel processing of binary structures (magic bytes, entropy, hashing, compression detection) and image characteristics (metadata, dimensions, color analysis, quality assessment).

**Config:** `workbench/config/cells/analysis/binary-media.yaml`

### Content Analysis Processing

Provides parallel multi-format content analysis through specialized analyzers for text (sentiment, keywords), JSON (validation, schema), XML (namespace, validation), binary (hashing, entropy), and images (metadata, dimensions) to support heterogeneous document processing workflows.

**Config:** `workbench/config/cells/analysis/content-analysis.yaml`

### Fast Analysis Processing

Provides optimized high-speed content analysis for text, JSON, and binary data by disabling expensive features (NLP, schema generation, entropy analysis) to achieve minimal latency in high-throughput production environments.

**Config:** `workbench/config/cells/analysis/fast-analysis.yaml`

### Structured Data Analysis Processing

Provides comprehensive analysis of structured data through deep JSON validation (50 levels, 10K keys) and XML namespace analysis (50 levels, 10K elements) with schema generation to support data quality validation, schema discovery, and format conversion workflows.

**Config:** `workbench/config/cells/analysis/structured-data.yaml`

### Text Analysis Deep Processing

Provides comprehensive text analysis with full NLP pipeline, sentiment analysis, and extensive keyword extraction (50 keywords) to support content classification, topic modeling, and semantic understanding workflows for large documents (up to 50K lines).

**Config:** `workbench/config/cells/analysis/text-deep.yaml`

---

## Output Synthesis

### Data Export Synthesis

Generates structured datasets and metadata collections from processed documents to support analytics platforms, data integration workflows, and external system consumption through schema-validated JSON exports.

**Config:** `workbench/config/cells/synthesis/data-export.yaml`

### Document Summary Synthesis

Generates comprehensive document summaries with keyword extraction, topic identification, and section-aware analysis to provide concise overviews for search interfaces, knowledge bases, and document management systems.

**Config:** `workbench/config/cells/synthesis/document-summary.yaml`

### Full Analysis Synthesis

Executes complete document analysis through parallel synthesis agents (summary, indexer, metadata, report, dataset) to generate comprehensive outputs for search systems, analytics platforms, business intelligence, and data integration workflows.

**Config:** `workbench/config/cells/synthesis/full-analysis.yaml`

### Reporting Synthesis

Generates business intelligence reports with charts, tables, and recommendations by combining metadata collection and report generation to support executive dashboards, analytics platforms, and decision-making workflows.

**Config:** `workbench/config/cells/synthesis/reporting.yaml`

### Search-Ready Synthesis

Generates search-optimized outputs through comprehensive keyword/topic extraction (40 keywords, 20 topics) and extensive term indexing (20K terms, low threshold) to support full-text search engines, discovery interfaces, and recommendation systems.

**Config:** `workbench/config/cells/synthesis/search-ready.yaml`

---

