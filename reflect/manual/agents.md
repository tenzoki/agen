# AGEN Agent Capabilities

**Generated:** 2025-10-07 07:41:22  
**Agents:** 27  

Complete catalog of available agents and their capabilities.

---

## Adapter

**Type:** `adapter`  
**Operator:** spawn  
**Description:** Provides data transformation services to other agents  

**Intent:**  
The adapter agent provides centralized data format conversion between pipeline agents. It acts as a transformation hub supporting schema mapping, text processing, format conversion (JSON/CSV/XML), and encoding operations.

**Capabilities:**  
- text-transform
- format-conversion
- json-processing
- csv-processing
- base64-encoding

**Usage:**  
Input: `AdapterRequest` containing source format, target format, and data to transform
Output: `AdapterResponse` with transformed data or error information
Configuration:
- `supported_formats`: Array of supported format combinations
- `max_request_size`: Maximum size for transformation requests

---

## Anonymization Store

**Type:** `anonymization-store`  
**Operator:** spawn  
**Description:** Persistent storage for anonymization mappings using godast/omnistore with bbolt backend  

**Intent:**  
Provides persistent key-value storage for anonymization mappings with bidirectional lookup (original → pseudonym, pseudonym → original). Uses bbolt-backed OmniStore for reliable storage with audit trail support.

**Capabilities:**  
- persistent-storage
- key-value-store
- bbolt-backend
- forward-reverse-lookup
- soft-delete
- audit-trail

**Usage:**  
Input: `StorageRequest` with operations: set, get, reverse, list, delete
Output: `StorageResponse` with operation results
Configuration:
- `data_path`: Storage location (default: `/tmp/gox-anonymization-store`)
- `max_file_size`: Maximum database file size (default: 100MB)
...

---

## Anonymizer

**Type:** `anonymizer`  
**Operator:** spawn  
**Description:** Anonymizes entities using deterministic pseudonyms with persistent storage  

**Intent:**  
Anonymizes text by replacing named entities (detected by NER agent) with consistent pseudonyms. Maintains bidirectional mappings via anonymization_store for data privacy compliance and de-anonymization when authorized.

**Capabilities:**  
- pseudonymization
- pii-anonymization
- deterministic-hashing
- persistent-mappings
- bidirectional-lookup
- gdpr-compliance

**Usage:**  
Input: `AnonymizerRequest` containing text and detected entities
Output: `AnonymizerResponse` with anonymized text and entity mappings
Configuration:
- `storage_agent_id`: ID of anonymization_store agent (default: "anon-store-001")
- `pipeline_version`: Version for audit trail (default: "v1.0")
...

---

## Binary Analyzer

**Type:** `binary-analyzer`  
**Operator:** spawn  
**Description:** Analyzes binary content with file type detection, entropy analysis, and pattern recognition  

**Intent:**  
Analyzes binary content to extract metadata, detect file types via magic bytes, calculate entropy for compression/encryption detection, and identify structural patterns. Provides detailed binary characterization for downstream processing decisions.

**Capabilities:**  
- binary-analysis
- file-type-detection
- entropy-analysis
- hash-calculation
- magic-bytes-detection
- compression-detection
- encryption-detection
- structural-analysis

**Usage:**  
Input: `ChunkProcessingRequest` with binary content
Output: `ProcessingResult` with comprehensive binary analysis
Configuration:
- `enable_hashing`: MD5/SHA256 hash calculation (default: true)
- `enable_entropy`: Shannon entropy analysis (default: true)
...

---

## Chunk Writer

**Type:** `chunk-writer`  
**Operator:** spawn  
**Description:** Writes enriched chunks to various output formats and storage systems  

**Intent:**  
Writes enriched text chunks to various output formats (JSON, text, markdown, CSV, XML) with metadata preservation, configurable file naming, and organization schemes. Provides the final output stage in text processing pipelines.

**Capabilities:**  
- chunk-writing
- multi-format-output
- json-formatting
- text-formatting
- markdown-formatting
- csv-formatting
- xml-formatting
- file-organization

**Usage:**  
Input: `ChunkWriteRequest` containing enriched chunks and output specifications
Output: `ChunkWriteResponse` with written file paths and statistics
Configuration:
- `default_output_format`: Output format (json/text/markdown/csv/xml, default: "json")
- `output_directory`: Base output directory (default: "/tmp/gox-chunk-writer")
...

---

## Context Enricher

**Type:** `context-enricher`  
**Operator:** spawn  
**Description:** Enriches text chunks with contextual metadata and relationships  

**Intent:**  
Adds multi-dimensional context to text chunks for improved RAG, search, and analysis. Provides document position tracking, semantic classification, structural hierarchy, and inter-chunk relationships to enable context-aware processing.

**Capabilities:**  
- context-enrichment
- semantic-classification
- positional-analysis
- structural-analysis
- relational-analysis
- metadata-extraction
- keyword-extraction

**Usage:**  
Input: `ContextEnrichmentRequest` containing chunks and document metadata
Output: `ContextEnrichmentResponse` with enriched chunks containing full contextual information
Configuration:
- `enable_positional_context`: Add document position metadata (default: true)
- `enable_semantic_context`: Add semantic classification (default: true)
...

---

## Dataset Builder

**Type:** `dataset-builder`  
**Operator:** spawn  
**Description:** Converts chunk processing results into structured datasets  

**Intent:**  
Synthesizes processed chunks into structured datasets for ML training, analysis, or export. Generates schemas, validates records, calculates statistics, and produces export-ready datasets with comprehensive metadata.

**Capabilities:**  
- dataset-creation
- data-structuring
- schema-generation
- data-validation
- statistical-analysis
- record-formatting
- data-export

**Usage:**  
Input: `SynthesisRequest` containing chunk IDs and output specifications
Output: `SynthesisResult` with structured dataset including schema and statistics
Configuration:
- `output_format`: Dataset format (default: "json")
- `include_metadata`: Include chunk metadata (default: true)
...

---

## Embedding Agent

**Type:** `embedding-agent`  
**Operator:** spawn  
**Description:** Generates vector embeddings for text/code chunks using OpenAI with VFS-based caching  

**Intent:**  
Generates vector embeddings for text/code chunks using embedding providers (OpenAI, HuggingFace, local models). Provides caching for efficiency, batch processing for API optimization, and project-isolated storage via VFS.

**Capabilities:**  
- embedding-generation
- openai-embeddings
- vector-generation
- caching
- batch-processing

**Usage:**  
Input: `EmbeddingRequest` containing texts to embed
Output: `EmbeddingResponse` with vector embeddings and cache statistics
Configuration:
- `provider`: Embedding provider ("openai"/hu ggingface"/"local", default: "openai")
- `model`: Model identifier (default: "text-embedding-3-small")
...

---

## File Ingester

**Type:** `file-ingester`  
**Operator:** call  
**Description:** Watches directories and ingests files  

**Intent:**  
Reads files from various sources (filesystem, cloud storage, URLs), performs initial validation, and prepares content for downstream processing. Acts as the entry point for document processing pipelines.

**Capabilities:**  
- file-ingestion
- directory-watching

**Usage:**  
Input: File paths or URIs to ingest
Output: File content with metadata for processing pipeline
Configuration:
- `supported_formats`: List of supported file extensions
- `max_file_size`: Maximum file size limit
...

---

## File Writer

**Type:** `file-writer`  
**Operator:** spawn  
**Description:** Writes processed data to files  

**Intent:**  
Writes processed content to various destinations with path management, directory creation, and atomic write operations. Provides reliable file output with backup and overwrite protection options.

**Capabilities:**  
- file-writing
- data-persistence

**Usage:**  
Input: Content to write with file path and options
Output: Written file confirmation with metadata
Configuration:
- `output_directory`: Default output directory
- `create_directories`: Auto-create parent directories
...

---

## GoDAST Storage

**Type:** `godast-storage`  
**Operator:** spawn  
**Description:** Unified storage backend using Godast with KV, Graph, File, and Search capabilities  

**Intent:**  
Stores and retrieves Go source code AST analysis results with efficient querying capabilities. Uses OmniStore with bbolt backend for reliable persistence of parsed code structures, symbol tables, and dependency graphs.

**Capabilities:**  
- storage
- kv-store
- graph-database
- file-storage
- full-text-search
- persistence

**Usage:**  
Input: Storage operations (set/get/query) for AST data
Output: Operation results with AST data
Configuration:
- `data_path`: Storage location for AST database
- `max_file_size`: Maximum database file size
...

---

## Image Analyzer

**Type:** `image-analyzer`  
**Operator:** spawn  
**Description:** Analyzes image content with metadata extraction, format detection, and quality assessment  

**Intent:**  
Analyzes image files to extract metadata (dimensions, format, color space), detect visual features, identify image types, and assess quality. Provides comprehensive image characterization for downstream processing decisions.

**Capabilities:**  
- image-analysis
- image-metadata-extraction
- dimension-analysis
- color-analysis
- image-format-detection
- quality-assessment
- image-classification
- pattern-detection

**Usage:**  
Input: `ChunkProcessingRequest` with image content
Output: `ProcessingResult` with image analysis metadata
Configuration:
- `enable_metadata`: Extract EXIF and image metadata
- `enable_feature_detection`: Detect visual features
...

---

## JSON Analyzer

**Type:** `json-analyzer`  
**Operator:** spawn  
**Description:** Analyzes JSON content with validation, schema generation, and structural analysis  

**Intent:**  
Analyzes JSON content to provide validation (RFC 7159), automatic schema generation (JSON Schema draft-07), structure detection (depth, complexity), pattern recognition (API responses, config files), and format classification for intelligent JSON processing.

**Capabilities:**  
- json-analysis
- json-validation
- schema-generation
- key-analysis
- json-minification
- structure-analysis
- pattern-detection
- json-classification

**Usage:**  
Input: `ChunkProcessingRequest` with JSON content
Output: `ProcessingResult` with comprehensive JSON analysis
Configuration:
- `enable_validation`: JSON syntax validation (default: true)
- `enable_schema_generation`: Auto-generate JSON schema (default: true)
...

---

## Metadata Collector

**Type:** `metadata-collector`  
**Operator:** spawn  
**Description:** Collects and aggregates metadata from all processed chunks  

**Intent:**  
Collects metadata from various pipeline agents (extractors, analyzers, enrichers) and creates comprehensive metadata records. Enables metadata-driven search, filtering, and analysis across processed documents.

**Capabilities:**  
- metadata-collection
- schema-generation
- metadata-aggregation
- statistics-calculation
- tag-extraction
- property-analysis

**Usage:**  
Input: Processing results with metadata from multiple agents
Output: Unified metadata records with aggregated information
Configuration:
- `metadata_fields`: Fields to collect and aggregate
- `enable_deduplication`: Remove duplicate metadata entries
...

---

## NER Agent - Named Entity Recognition

**Type:** `ner-agent`  
**Operator:** spawn  
**Description:** Named Entity Recognition using ONNXRuntime with multilingual XLM-RoBERTa model  

**Capabilities:**  
- named-entity-recognition
- onnxruntime-inference
- multilingual-ner
- xlm-roberta
- pii-detection
- entity-extraction

---

## OCR HTTP Stub

**Type:** `ocr-http-stub`  
**Operator:** await  
**Description:** HTTP client stub for containerized OCR service with await pattern for service dependencies  

**Intent:**  
Provides OCR capabilities by connecting to a containerized HTTP OCR service (Python Flask + Tesseract). Uses await pattern to ensure external service availability before cell activation. Supports multiple languages, PDF processing, and image preprocessing.

**Capabilities:**  
- ocr-processing
- http-client
- image-processing
- pdf-ocr
- batch-processing
- multi-language-ocr
- containerized-ocr
- service-integration

**Usage:**  
Input: `ocr_request` with image/PDF file paths
Output: `ocr_response` with extracted text and confidence metrics
Configuration:
- `service_url`: OCR service URL (default: "http://localhost:8080/ocr")
- `request_timeout`: API timeout (default: 5 minutes)
...

---

## RAG Agent

**Type:** `rag-agent`  
**Operator:** spawn  
**Description:** Orchestrates RAG workflow: embedding generation, vector search, content retrieval, and context assembly  

**Intent:**  
Implements RAG pattern by retrieving relevant context from vector store and generating answers using LLM. Combines vector search, context assembly, and LLM generation for accurate, source-grounded responses.

**Capabilities:**  
- rag-orchestration
- retrieval
- context-assembly
- reranking
- query-processing

**Usage:**  
Input: Query with optional context retrieval parameters
Output: Generated answer with source references and confidence
Configuration:
- `openai_model`: LLM model (default: "gpt-4o-mini")
- `max_tokens`: Maximum response tokens
...

---

## Report Generator

**Type:** `report-generator`  
**Operator:** spawn  
**Description:** Generates comprehensive analysis reports with charts and recommendations  

**Intent:**  
Creates comprehensive reports aggregating analysis results, statistics, and insights from document processing pipelines. Supports multiple formats (PDF, HTML, Markdown) with customizable templates and visualizations.

**Capabilities:**  
- report-generation
- analysis-reporting
- chart-generation
- table-generation
- recommendation-generation
- quality-assessment
- statistical-analysis

**Usage:**  
Input: Processing results and report configuration
Output: Formatted report in specified format
Configuration:
- `output_format`: Report format (pdf/html/markdown)
- `template_path`: Report template location
...

---

## Search Indexer

**Type:** `search-indexer`  
**Operator:** spawn  
**Description:** Builds searchable indexes from processed chunk content  

**Intent:**  
Creates and maintains search indexes from processed documents enabling fast full-text search, filtering, and ranking. Supports inverted indexes, keyword extraction, and relevance scoring for efficient document discovery.

**Capabilities:**  
- search-indexing
- term-indexing
- document-indexing
- tf-idf-calculation
- index-statistics
- keyword-scoring
- full-text-search

**Usage:**  
Input: Documents and chunks to index
Output: Search index with retrieval capabilities
Configuration:
- `index_path`: Search index storage location
- `analyzer`: Text analyzer (standard/keyword/n-gram)
...

---

## Strategy Selector

**Type:** `strategy-selector`  
**Operator:** spawn  
**Description:** Analyzes documents to select optimal chunking strategies  

**Intent:**  
Analyzes document characteristics (format, size, complexity, language) and selects optimal processing strategies for extraction, chunking, and analysis. Enables adaptive pipeline configuration based on content properties.

**Capabilities:**  
- strategy-selection
- content-analysis
- document-classification
- quality-assessment
- rule-based-selection
- format-detection

**Usage:**  
Input: Document metadata and content sample
Output: Selected processing strategies and configuration
Configuration:
- `strategy_rules`: Rule-based strategy selection
- `enable_ml_selection`: Use ML for strategy selection
...

---

## Summary Generator

**Type:** `summary-generator`  
**Operator:** spawn  
**Description:** Generates comprehensive document summaries from processed chunks  

**Intent:**  
Generates concise summaries from document text using extractive (sentence selection) and abstractive (text generation) methods. Provides adjustable summary length, multi-document summarization, and key point extraction.

**Capabilities:**  
- document-summarization
- content-synthesis
- keyword-extraction
- topic-analysis
- language-detection
- title-generation
- section-analysis

**Usage:**  
Input: Text content with summary parameters
Output: Generated summary with key points and metadata
Configuration:
- `summarization_method`: extractive/abstractive/hybrid
- `summary_length`: Target summary length (words/sentences)
...

---

## Text Analyzer

**Type:** `text-analyzer`  
**Operator:** spawn  
**Description:** Analyzes text content with sentiment, keywords, language detection, and quality metrics  

**Intent:**  
Analyzes text chunks to extract sentiment (positive/negative/neutral), keywords (TF-IDF), language identification, content categories, and statistical metrics. Provides NLP-based insights for intelligent text processing and search enhancement.

**Capabilities:**  
- text-analysis
- sentiment-analysis
- keyword-extraction
- language-detection
- text-normalization
- reading-level-analysis
- text-statistics
- text-classification

**Usage:**  
Input: `ChunkProcessingRequest` with text content
Output: `ProcessingResult` with comprehensive text analysis
Configuration:
- `enable_nlp`: Enable NLP processing (default: false, lightweight)
- `enable_sentiment`: Sentiment analysis (default: true)
...

---

## Text Chunker

**Type:** `text-chunker`  
**Operator:** spawn  
**Description:** Chunks text using various strategies with configurable parameters  

**Intent:**  
Splits text into manageable chunks for processing, embedding, and analysis. Supports multiple strategies (paragraph-based, section-based, size-based, boundary-based, semantic) with overlap configuration for context preservation.

**Capabilities:**  
- text-chunking
- paragraph-based-chunking
- section-based-chunking
- boundary-based-chunking
- size-based-chunking
- overlap-management
- strategy-application

**Usage:**  
Input: Text content with chunking parameters
Output: Text chunks with metadata and position information
Configuration:
- `chunk_size`: Target chunk size in characters/tokens
- `overlap_size`: Overlap between consecutive chunks
...

---

## Text Extractor Native

**Type:** `text-extractor-native`  
**Operator:** spawn  
**Description:** Multi-format text extraction with native Tesseract OCR for PDF, DOCX, XLSX, and image files  

**Intent:**  
Provides simple, fast text extraction without external dependencies. Uses local Tesseract installation for OCR on images (.png, .jpg, .tiff) and direct reading for plain text files. Low latency alternative to HTTP OCR service for smaller workloads.

**Capabilities:**  
- text-extraction
- pdf-processing
- docx-processing
- xlsx-processing
- native-ocr-processing
- image-processing
- multi-format-support
- metadata-extraction
- tesseract-ocr

**Usage:**  
Input: File paths for text/image files
Output: Extracted text with quality metrics and metadata
Configuration:
- `enable_ocr`: Enable OCR processing (default: true)
- `ocr_languages`: Tesseract language packs (default: ["eng"])
...

---

## Text Transformer

**Type:** `text-transformer`  
**Operator:** spawn  
**Description:** Transforms text content  

**Intent:**  
Applies various transformations to text including case conversion, whitespace normalization, special character handling, encoding fixes, and format conversion. Prepares text for downstream NLP and analysis tasks.

**Capabilities:**  
- text-processing
- transformation

**Usage:**  
Input: Text content with transformation specifications
Output: Transformed text with operation metadata
Configuration:
- `transformations`: List of transformations to apply
- `preserve_formatting`: Maintain structural formatting
...

---

## Vectorstore Agent

**Type:** `vectorstore-agent`  
**Operator:** spawn  
**Description:** Stores and searches vector embeddings with cosine similarity search and metadata filtering  

**Intent:**  
Manages vector storage and similarity search for RAG applications. Stores document embeddings with metadata, performs k-NN search, and supports filtering for efficient semantic retrieval. Integrates with embedding_agent for end-to-end vector search.

**Capabilities:**  
- vector-storage
- similarity-search
- cosine-similarity
- flat-index
- metadata-filtering
- persistent-storage

**Usage:**  
Input: Store/search operations with vectors and metadata
Output: Storage confirmation or search results with similarity scores
Configuration:
- `storage_backend`: Vector storage backend (in-memory/persistent)
- `index_type`: Similarity index (flat/hnsw/ivf)
...

---

## XML Analyzer

**Type:** `xml-analyzer`  
**Operator:** spawn  
**Description:** Analyzes XML content with validation, namespace analysis, and structural classification  

**Intent:**  
Analyzes XML content to provide well-formedness validation, namespace analysis with prefix resolution, element frequency statistics, format detection (HTML, SOAP, RSS, Maven POM), structure analysis (depth, complexity), and schema inference for intelligent XML processing.

**Capabilities:**  
- xml-analysis
- xml-validation
- namespace-analysis
- element-analysis
- xml-minification
- schema-detection
- xml-classification
- structure-analysis

**Usage:**  
Input: `ChunkProcessingRequest` with XML content
Output: `ProcessingResult` with comprehensive XML analysis
Configuration:
- `enable_validation`: XML well-formedness validation (default: true)
- `enable_namespace_analysis`: Namespace resolution (default: true)
...

---

