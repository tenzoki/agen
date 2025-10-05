# Builder

Build infrastructure for AGEN modules with automated testing and CI/CD support.

## Intent

Provides unified build system for all AGEN modules (omni, cellorg, agents, alfa, atomic) with comprehensive test automation, coverage reporting, and build orchestration. Standardized Makefiles and test scripts ensure consistent builds across modules.

## Usage

Build entire system:
```bash
cd builder
make -f Makefile.omni all      # Build omni module
./test-omni.sh                  # Run omni tests with coverage
```

Build specific module:
```bash
cd ../code/omni
go build -o ../../bin/omni ./cmd/...
```

Run tests with coverage:
```bash
cd builder
./test-omni.sh                  # Comprehensive test suite
./test-omni.sh --coverage       # Generate coverage report
```

## Setup

Dependencies:
- Go 1.24.3+
- git (for version tagging)

No additional setup required - Makefiles are self-contained.

## Tests

Test automation script (`test-omni.sh`):
- Unit tests with coverage threshold (75%)
- Integration tests
- Benchmark tests
- Race condition detection
- Timeout enforcement (300s)

Run full test suite:
```bash
./test-omni.sh
```

Generate coverage report:
```bash
./test-omni.sh --coverage
open coverage.html
```

## Demo

Build and test workflow:
```bash
# Build all binaries
make -f Makefile.omni build

# Run tests
./test-omni.sh

# Clean build artifacts
make -f Makefile.omni clean
```

See `Makefile.omni` for additional targets: `help`, `build`, `test`, `clean`, `demos`.
