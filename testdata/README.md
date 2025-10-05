# GOX Framework Test Data

This directory contains comprehensive test data, sample files, configurations, and examples for the GOX framework testing.

## 📁 Directory Structure

### **🖼️ images/**
Sample images for OCR and image processing testing:
- `test_sample.png` - Basic test image with text
- `test_ocr_image.png` - OCR-specific test image
- `complex_test.png` - Complex text layout testing
- `multilang_test.png` - Multi-language OCR testing
- `simple_test.png` - Simple text extraction testing

### **📄 documents/**
Sample documents for text extraction and processing:
- `sample.txt` - Basic text document with various content types
- `data_report.txt` - Data analysis report sample
- `roster.txt` - Simple roster/list document
- `test_text.txt` - Simple text for PDF testing
- `scanned_document.pdf` - Sample PDF document

### **📊 json/**
JSON test files for json-analyzer agent:
- `simple.json` - Basic JSON structure with common data types
- `complex.json` - Advanced JSON with nested structures and configurations

### **🏷️ xml/**
XML test files for xml-analyzer agent:
- `simple.xml` - Basic XML document with standard elements
- `complex.xml` - Advanced XML with namespaces, CDATA, and complex structures

### **📈 structured/**
Structured data files for data processing tests:
- `sample.csv` - CSV file with document processing metrics
- `metrics.tsv` - Tab-separated values with agent performance data

### **🌍 multilingual/**
Multilingual and Unicode test files:
- `unicode_test.txt` - Comprehensive multilingual text with special characters

### **⚙️ config/**
Configuration files for testing framework components:
- `pool_test.yaml` - Agent pool configuration for testing
- `cell_test.yaml` - Comprehensive cell configuration with all agent types

### **🔧 binary/**
Binary files for binary-analyzer testing:
- `test_binary.bin` - Mixed binary and text content
- `executable_header.bin` - Mock executable with binary header

### **⚠️ edge-cases/**
Edge case and error testing files:
- `empty.txt` - Empty file for boundary testing
- `whitespace_only.txt` - File containing only whitespace
- `single_line.txt` - Single line text without line breaks
- `large_text.txt` - Large file for chunking and splitting tests
- `malformed.json` - Intentionally malformed JSON for error handling tests
- `malformed.xml` - Intentionally malformed XML for error handling tests

## 🚀 Quick Usage

### **Test OCR Functionality**
```bash
# Run OCR tests with sample images
../scripts/test_ocr.sh

# Test with cell-based extraction
./build/gox config/cells.yaml  # Uses extraction:native-text cell
```

### **Test Docker OCR Service**
```bash
# Start OCR container and test
./scripts/build-ocr-http.sh

# Test with sample image
curl -F "image=@test/data/images/test_sample.png" http://localhost:8080/ocr
```

### **Test JSON Analysis**
```bash
# Test JSON analyzer with simple file
./build/gox test/data/config/cell_test.yaml

# Test with complex JSON structure (cell-based processing)
./build/gox test/data/config/json_analysis_cell.yaml
```

### **Test XML Analysis**
```bash
# Test XML parsing and analysis (cell-based processing)
./build/gox test/data/config/xml_analysis_cell.yaml

# Test complex XML with namespaces (cell-based processing)
./build/gox test/data/config/xml_analysis_cell.yaml
```

### **Test Multilingual Processing**
```bash
# Test Unicode and multilingual text processing (cell-based processing)
./build/gox test/data/config/text_analysis_cell.yaml
```

### **Test File Chunking**
```bash
# Test file chunking with large file (cell-based processing)
./build/gox test/data/config/chunking_cell.yaml

# Test chunking configuration
./build/gox test/data/config/cell_test.yaml
```

### **Test Error Handling**
```bash
# Test malformed JSON handling (cell-based processing)
./build/gox test/data/config/json_error_handling_cell.yaml

# Test malformed XML handling (cell-based processing)
./build/gox test/data/config/xml_error_handling_cell.yaml
```

## 📝 File Descriptions

### **Images**
- **test_sample.png**: "GOX Framework Test Image" - basic functionality test
- **test_ocr_image.png**: "This is a test for OCR processing in the GOX framework" - OCR testing
- **complex_test.png**: "Testing All-in-One OCR Container" - container testing
- **multilang_test.png**: "Hello World - Hallo Welt - Bonjour Monde" - multi-language testing

### **Documents**
- **sample.txt**: Multi-paragraph text with special characters, numbers, mixed content
- **data_report.txt**: Sample data analysis report for processing tests
- **roster.txt**: Simple list/roster document for basic text extraction
- **test_text.txt**: Minimal text content for PDF testing
- **scanned_document.pdf**: Sample PDF document for OCR testing

### **JSON Files**
- **simple.json**: Basic JSON with common data types (strings, numbers, arrays, objects)
- **complex.json**: Advanced GOX configuration with nested structures, arrays, and complex objects

### **XML Files**
- **simple.xml**: Standard XML document with elements, attributes, and text content
- **complex.xml**: Advanced XML with namespaces, CDATA sections, DTD references, and complex structures

### **Structured Data**
- **sample.csv**: Document processing metrics in CSV format
- **metrics.tsv**: Agent performance data in tab-separated format

### **Multilingual**
- **unicode_test.txt**: Comprehensive multilingual text in 9 languages with special characters, symbols, and mixed content

### **Configuration Files**
- **pool_test.yaml**: Complete agent pool configuration with all GOX agent types
- **cell_test.yaml**: Comprehensive cell configuration demonstrating all agent interactions

### **Binary Files**
- **test_binary.bin**: Mixed binary and text content for binary analyzer testing
- **executable_header.bin**: Mock executable header for file type detection testing

### **Edge Cases**
- **empty.txt**: Zero-byte file for boundary condition testing
- **whitespace_only.txt**: File containing only whitespace characters
- **single_line.txt**: Single line text without line breaks
- **large_text.txt**: Large text file (3KB+) for chunking and splitting tests
- **malformed.json**: Intentionally broken JSON with syntax errors
- **malformed.xml**: Intentionally broken XML with various syntax errors

## 🔄 Maintenance

### **Adding New Test Data**
1. Place files in appropriate subdirectory based on file type and purpose
2. Update this README with file descriptions
3. Update relevant test scripts to include new files
4. Ensure file paths in documentation reference test/data/ directory

### **Updating Paths**
When moving or renaming files:
1. Update all references in documentation (docs/*.md)
2. Update test scripts in scripts/ directory
3. Update configuration examples
4. Update Docker quickstart guide

### **File Naming Conventions**
- **Images**: `{purpose}_test.png` or `test_{description}.png`
- **Documents**: `{type}.txt` or `{purpose}_{type}.txt`
- **Configs**: `{mode}_test.yaml` or `{purpose}_test.json`
- **Data**: `{type}_{purpose}.{ext}` (e.g., `metrics_sample.csv`)
- **Edge cases**: `{condition}.{ext}` (e.g., `empty.txt`, `malformed.json`)

## ⚠️ Important Notes

- **Do not commit large files** (>10MB) to the repository
- **Keep test data minimal** but representative of real use cases
- **Use relative paths** in scripts: `test/data/images/test.png` not absolute paths
- **Document all test data** with clear descriptions of purpose and expected results
- **Maintain backwards compatibility** when updating file paths
- **Test edge cases** to ensure robust error handling

## 🧪 Testing Guidelines

### **Image Testing**
- Test with different image formats (PNG, JPG)
- Include images with various text layouts (single line, paragraphs, complex)
- Test multi-language content when applicable
- Ensure images are readable and provide good OCR results

### **Document Testing**
- Include documents with different content types (plain text, structured data)
- Test various file sizes (small, medium, large within reason)
- Include edge cases (empty files, special characters, unicode)

### **JSON/XML Testing**
- Validate syntax-correct files parse properly
- Test malformed files trigger appropriate error handling
- Include nested structures and complex data types
- Test with both simple and complex schemas

### **Configuration Testing**
- Validate all YAML/JSON configuration files are syntactically correct
- Test both minimal and comprehensive configuration examples
- Ensure configurations work with current framework version
- Document required vs optional configuration parameters

### **Binary Testing**
- Include various binary file types for format detection
- Test files with mixed binary and text content
- Include files with recognizable headers (e.g., ZIP, EXE)
- Test boundary conditions (very small/large files)

### **Edge Case Testing**
- Test empty files and whitespace-only files
- Test extremely large files (within CI/CD limits)
- Test files with various encoding issues
- Test malformed data for error handling validation

This test data directory provides a centralized, organized collection of all test materials needed for comprehensive GOX framework testing and development.