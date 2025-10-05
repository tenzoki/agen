#!/bin/bash
# Native Text Extraction Demo
#
# This demo shows how to use the text-extractor-native agent
# for document processing without external dependencies.

set -e

echo "🔍 Native Text Extraction Demo"
echo "==============================="

# Check if GOX is built
if [ ! -f "../../build/gox" ]; then
    echo "❌ GOX orchestrator not found. Please run: ./scripts/build.sh"
    exit 1
fi

if [ ! -f "../../build/text_extractor_native" ]; then
    echo "❌ Native text extractor not found. Please run: ./scripts/build.sh"
    exit 1
fi

# Create test documents if they don't exist
echo "📄 Setting up test documents..."
mkdir -p input output/native

if [ ! -f "input/test_document.txt" ]; then
    cat > input/test_document.txt << 'EOF'
# Sample Document

This is a test document for the GOX native text extraction demo.

## Features Tested
- Plain text extraction
- Multi-format support
- Metadata extraction

The native text extractor uses local Tesseract OCR for image processing
and supports PDF, DOCX, XLSX, and plain text files.

Processing Time: Fast (no network overhead)
Dependencies: Local Tesseract only
Best For: Development, simple deployments
EOF
fi

if command -v magick >/dev/null 2>&1; then
    if [ ! -f "input/test_image.png" ]; then
        echo "📸 Creating test image with text..."
        echo "GOX Native OCR Test" | magick -pointsize 24 -size 400x100 -gravity center label:@- input/test_image.png
    fi
fi

echo "✅ Test files ready"

# Demo 1: Start native extraction cell
echo ""
echo "🚀 Demo 1: Native Text Extraction Cell"
echo "--------------------------------------"
echo "Starting GOX with native extraction cell..."
echo "Cell ID: extraction:native-text"
echo ""
echo "Press Ctrl+C to stop and continue to next demo"
echo ""

# Note: In a real demo, you would send messages to the cell
# For now, we show how to start it
echo "Command: ./build/gox config/cells.yaml"
echo ""
echo "The cell will listen for messages on: sub:native-text-extraction"
echo "Results will be published to: pub:extracted-text-native"
echo ""
echo "(Simulated - press Enter to continue)"
read -p ""

# Demo 2: Direct agent testing
echo ""
echo "🔧 Demo 2: Direct Agent Testing (Development Mode)"
echo "--------------------------------------------------"
echo "Running native text extractor agent directly..."

# Create a simple test message file
cat > input/test_request.json << 'EOF'
{
  "request_id": "demo_native_001",
  "file_path": "examples/text-extraction/input/test_document.txt",
  "options": {
    "enable_ocr": true,
    "ocr_languages": ["eng"],
    "include_metadata": true,
    "quality_threshold": 0.8
  }
}
EOF

echo "Test request created: input/test_request.json"
echo ""
echo "To test the agent directly:"
echo "GOX_AGENT_ID=demo-native ./build/text_extractor_native"
echo ""
echo "(The agent will connect to GOX broker and wait for messages)"

# Demo 3: Performance comparison
echo ""
echo "📊 Demo 3: Performance Characteristics"
echo "-------------------------------------"
echo "Native Text Extractor Performance:"
echo "- ⚡ Low Latency: No network overhead"
echo "- 🔧 Simple Setup: Binary only, local dependencies"
echo "- 📈 CPU Bound: Performance limited by local resources"
echo "- 🎯 Best For: Development, single-node deployments"
echo ""

# Demo 4: Supported formats
echo "📁 Demo 4: Supported Document Formats"
echo "------------------------------------"
echo "The native text extractor supports:"
echo "- ✅ Plain Text (.txt, .md)"
echo "- ✅ PDF documents (.pdf)"
echo "- ✅ Word documents (.docx)"
echo "- ✅ Excel spreadsheets (.xlsx)"
echo "- ✅ Images with OCR (.png, .jpg, .tiff, .bmp)"
echo ""

if command -v tesseract >/dev/null 2>&1; then
    echo "🔍 Local Tesseract OCR detected:"
    tesseract --version | head -1
    echo ""
    echo "📚 Available OCR languages:"
    tesseract --list-langs | tail -n +2 | head -10
else
    echo "⚠️  Tesseract OCR not found. OCR functionality will be disabled."
    echo "   Install with: apt-get install tesseract-ocr (Ubuntu)"
    echo "   Install with: brew install tesseract (macOS)"
fi

echo ""
echo "🎉 Native Text Extraction Demo Complete!"
echo ""
echo "💡 Next Steps:"
echo "  1. Run: ./build/gox config/cells.yaml"
echo "  2. Use cell: extraction:native-text"
echo "  3. Send documents to: sub:native-text-extraction"
echo "  4. Receive results from: pub:extracted-text-native"
echo ""
echo "🔄 For HTTP OCR demo, run: ./demo_http.sh"