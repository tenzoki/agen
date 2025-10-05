# Drivers

Platform-specific libraries, binaries, and setup instructions for AGEN.

## Intent

This directory contains all platform-dependent code, libraries, and configuration required to run AGEN on different operating systems and architectures. By isolating platform-specific components here, we maintain a clean separation between core AGEN code and OS/architecture-specific dependencies.

## Structure

```
drivers/
└── tokenizers/         # HuggingFace tokenizers library (macOS/Darwin)
    ├── libtokenizers.a # Static library for tokenizer functionality
    └── README.md       # Tokenizers-specific documentation
```

## Platform Support

### macOS (Darwin)

**Tokenizers Library** (`tokenizers/`):
- Pre-compiled `libtokenizers.a` for macOS (Apple Silicon and Intel)
- Required for NER agent tokenization functionality
- Built from HuggingFace tokenizers Rust library

### Linux

Platform-specific libraries and setup for Linux will be added here as needed.

### Windows

Platform-specific libraries and setup for Windows will be added here as needed.

## Setup

### Installing Tokenizers Library (macOS)

The tokenizers library is pre-built and included in `drivers/tokenizers/`. No additional installation is required - the NER agent will automatically link against `libtokenizers.a` during build.

**Build Requirements**:
- Go 1.24.3 or later
- macOS 11+ (Big Sur or later)
- Xcode Command Line Tools

**Verification**:
```bash
# Check if library exists
ls -lh drivers/tokenizers/libtokenizers.a

# Build NER agent to verify linkage
cd code/agents/ner_agent
go build -v .
```

### Adding New Platform-Specific Components

When adding platform-specific dependencies:

1. Create a subdirectory with a descriptive name (e.g., `onnxruntime/`, `opencv/`)
2. Include platform-specific binaries/libraries with clear naming (e.g., `lib_darwin_arm64.a`)
3. Add a README.md explaining:
   - What the component provides
   - Which AGEN modules depend on it
   - Build/installation instructions for each platform
   - Version information and update procedures

## Testing

Platform-specific components should be tested on their respective platforms:

```bash
# Test NER agent with tokenizers (macOS)
cd code/agents/ner_agent
go test -v .

# Run integration tests
../../builder/test-agents.sh
```

## Notes

- **Static vs Dynamic Linking**: Prefer static libraries (`.a`) for better portability
- **Version Management**: Document library versions in individual component READMEs
- **CI/CD**: Platform-specific components may require platform-specific build pipelines
- **License Compliance**: Ensure all included libraries comply with AGEN's licensing

## Contributing

When adding platform-specific code:
1. Ensure it's truly platform-specific and can't be abstracted
2. Document setup procedures for all supported platforms
3. Include build scripts where applicable
4. Update this README with new component information
