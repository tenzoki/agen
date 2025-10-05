# Test Reports

Auto-generated test reports and coverage analysis for AGEN.

## Intent

Central repository for all test execution reports, coverage analysis, and quality metrics. Generated automatically by the test automation system.

## Structure

```
test-reports/
├── test-report-YYYYMMDD_HHMMSS.md   # Timestamped test reports
├── latest-report.md                  # Symlink to most recent report
├── latest-summary.txt                # Quick status summary
├── coverage/                         # HTML coverage reports
│   ├── atomic-coverage.html
│   ├── omni-coverage.html
│   ├── cellorg-coverage.html
│   ├── agents-coverage.html
│   └── alfa-coverage.html
└── *.log                             # Individual module test logs
```

## Usage

### Generate Test Report

```bash
# Quick tests (no coverage)
make test-quick

# Standard tests with report
make test

# Full tests with coverage
make test-full
```

### View Latest Report

```bash
# View summary
cat reflect/test-reports/latest-summary.txt

# View full report
cat reflect/test-reports/latest-report.md

# Open coverage report in browser
open reflect/test-reports/coverage/omni-coverage.html
```

### Direct Script Usage

```bash
# Quick mode (fast, no coverage)
./builder/test-all.sh quick

# CI mode (race detection, verbose)
./builder/test-all.sh ci

# Full mode (coverage reports)
./builder/test-all.sh full
```

## Report Contents

Each test report includes:
- **Summary:** Total modules, pass/fail counts, success rate
- **Per-Module Results:** Status, test count, coverage percentage
- **Coverage Links:** HTML coverage reports (full mode only)
- **Failure Logs:** Links to detailed logs for failed modules
- **Metadata:** Timestamp, mode, Go version, host

## Test Modes

**quick:**
- Runs `-short` flag tests only
- No race detection
- No coverage
- Fast (< 30 seconds)

**ci:**
- All tests
- Race detection enabled
- Verbose output
- Moderate (1-3 minutes)

**full:**
- All tests
- Race detection
- Coverage analysis
- HTML coverage reports
- Slowest (2-5 minutes)

## Coverage Thresholds

Minimum coverage requirements:
- **omni:** 70%
- **cellorg:** 70%
- **agents:** 60% (varies by agent)
- **atomic:** 70%
- **alfa:** 60%

## Setup

No setup required - directory and reports created automatically by test automation.

## Tests

Test automation itself is tested through:
- Successful execution across all modes
- Report generation validation
- Coverage threshold verification

## Demo

```bash
# Run full test suite
make test-full

# Check results
cat reflect/test-reports/latest-summary.txt
```
