# Binary Media Analysis Processing

Specialized analysis for binary data and media files

## Intent

Provides comprehensive analysis of binary files and images through parallel processing of binary structures (magic bytes, entropy, hashing, compression detection) and image characteristics (metadata, dimensions, color analysis, quality assessment).

## Agents

- **binary-analyzer-comprehensive-001** (binary-analyzer) - Comprehensive binary file analysis
  - Ingress: sub:binary-media-binary
  - Egress: pub:comprehensive-analyzed-binary

- **image-analyzer-comprehensive-001** (image-analyzer) - Comprehensive image analysis
  - Ingress: sub:binary-media-image
  - Egress: pub:comprehensive-analyzed-image

## Data Flow

```
Binary Files → binary-analyzer (hashing, entropy, magic bytes, structural, compression)
  → Analysis Results

Image Files → image-analyzer (metadata, dimensions, color, thumbnail, quality)
  → Analysis Results
```

## Configuration

Binary Analysis:
- Hashing, entropy, magic bytes detection enabled
- Structural and compression analysis enabled
- Max analysis size: 50MB

Image Analysis:
- Metadata extraction, dimension detection enabled
- Color analysis and quality assessment enabled
- Thumbnail generation enabled
- Max analysis size: 50MB

Orchestration:
- Startup timeout: 60s, shutdown: 30s
- Max retries: 2, retry delay: 20s
- Health check interval: 45s

## Usage

```bash
./bin/orchestrator -config=./workbench/config/binary-media-analysis-processing.yaml
```
