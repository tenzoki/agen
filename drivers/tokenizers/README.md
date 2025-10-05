# Native Libraries Directory

This directory contains native libraries required for CGO-based agents.

## Current Libraries

### libtokenizers.a
- **Purpose**: HuggingFace tokenizers for NER agent
- **Source**: https://github.com/daulet/tokenizers
- **Size**: ~39MB
- **Platform**: darwin-arm64 (macOS Apple Silicon)

## Installation

### macOS ARM64
```bash
cd /tmp
curl -L -o libtokenizers.darwin-arm64.tar.gz \
  https://github.com/daulet/tokenizers/releases/latest/download/libtokenizers.darwin-arm64.tar.gz
tar -xzf libtokenizers.darwin-arm64.tar.gz
cp libtokenizers.a /path/to/gox/lib/
```

### Linux x86_64
```bash
cd /tmp
curl -L -o libtokenizers.linux-amd64.tar.gz \
  https://github.com/daulet/tokenizers/releases/latest/download/libtokenizers.linux-amd64.tar.gz
tar -xzf libtokenizers.linux-amd64.tar.gz
cp libtokenizers.a /path/to/gox/lib/
```

### Linux ARM64
```bash
cd /tmp
curl -L -o libtokenizers.linux-arm64.tar.gz \
  https://github.com/daulet/tokenizers/releases/latest/download/libtokenizers.linux-arm64.tar.gz
tar -xzf libtokenizers.linux-arm64.tar.gz
cp libtokenizers.a /path/to/gox/lib/
```

## Usage

The `onnx-exports` file in the project root automatically includes this directory in CGO_LDFLAGS:

```bash
source onnx-exports
go build -o build/ner_agent ./agents/ner_agent/
```

## Note

The actual library files (`.a`) are excluded from git via `.gitignore` due to their size.
Each developer needs to download the appropriate library for their platform.
