#!/bin/bash
# HTTP OCR Stub Demo
#
# This demo shows how to use the ocr-http-stub agent with
# containerized OCR service using the await pattern.

set -e

echo "🌐 HTTP OCR Service Demo"
echo "========================"

# Check if GOX is built
if [ ! -f "../../build/gox" ]; then
    echo "❌ GOX orchestrator not found. Please run: ./scripts/build.sh"
    exit 1
fi

if [ ! -f "../../build/ocr_http_stub" ]; then
    echo "❌ OCR HTTP stub not found. Please run: ./scripts/build.sh"
    exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker not found. Please install Docker to run this demo."
    exit 1
fi

# Create test documents
echo "📄 Setting up test documents..."
mkdir -p input output/http

if command -v magick >/dev/null 2>&1; then
    if [ ! -f "input/ocr_test_image.png" ]; then
        echo "📸 Creating OCR test image..."
        cat > input/sample_text.txt << 'EOF'
GOX HTTP OCR Demo
Multi-language support
Containerized processing
Production ready
EOF
        magick -pointsize 20 -size 500x200 -gravity center label:@input/sample_text.txt input/ocr_test_image.png
        rm input/sample_text.txt
    fi
else
    echo "⚠️  ImageMagick not available. Using existing test images."
fi

# Create multi-language test image
if command -v magick >/dev/null 2>&1; then
    if [ ! -f "input/multilang_test.png" ]; then
        echo "🌍 Creating multi-language test image..."
        echo "English: Hello World
Deutsch: Hallo Welt
Français: Bonjour le monde" | magick -pointsize 18 -size 500x150 -gravity center label:@- input/multilang_test.png
    fi
fi

echo "✅ Test files ready"

# Demo 1: Start OCR service
echo ""
echo "🚀 Demo 1: Starting Containerized OCR Service"
echo "--------------------------------------------"
echo "The HTTP OCR stub requires a running OCR service."
echo "Starting containerized OCR service..."

cd ../../agents/ocr_http_stub

echo "📦 Building and starting OCR service..."
echo "Command: ./scripts/build-ocr-http.sh"
echo ""
echo "This will:"
echo "- Build OCR service Docker image"
echo "- Start service on http://localhost:8080"
echo "- Perform health checks"
echo "- Test OCR functionality"
echo ""

# Start the service
cd ../../
./scripts/build-ocr-http.sh

echo "✅ OCR service is running!"
echo ""

# Demo 2: Service health and capabilities
echo "🔍 Demo 2: OCR Service Capabilities"
echo "----------------------------------"

echo "Service health check:"
curl -s http://localhost:8080/health | jq '.' || echo "Health check response received"
echo ""

echo "Available OCR languages:"
curl -s http://localhost:8080/languages | jq '.languages[]' || echo "Languages retrieved"
echo ""

echo "Service information:"
curl -s http://localhost:8080/info | jq '.capabilities[]' || echo "Capabilities listed"
echo ""

# Demo 3: Direct OCR testing
echo "🧪 Demo 3: Direct OCR Service Testing"
echo "------------------------------------"

cd ../../examples/text-extraction

if [ -f "input/ocr_test_image.png" ]; then
    echo "Testing single image OCR:"
    echo "curl -F \"image=@input/ocr_test_image.png\" -F \"languages=eng\" http://localhost:8080/ocr"
    curl -s -F "image=@input/ocr_test_image.png" -F "languages=eng" http://localhost:8080/ocr | jq '.text' || echo "OCR response received"
    echo ""
fi

if [ -f "input/multilang_test.png" ]; then
    echo "Testing multi-language OCR:"
    echo "curl -F \"image=@input/multilang_test.png\" -F \"languages=eng+deu+fra\" http://localhost:8080/ocr"
    curl -s -F "image=@input/multilang_test.png" -F "languages=eng+deu+fra" http://localhost:8080/ocr | jq '.text' || echo "Multi-language OCR response received"
    echo ""
fi

# Demo 4: GOX HTTP Stub Agent
echo "🤖 Demo 4: GOX HTTP OCR Stub Agent"
echo "---------------------------------"
echo "Starting GOX with HTTP OCR extraction cell..."
echo ""
echo "The await pattern ensures:"
echo "- ✅ OCR service must be healthy before cell starts"
echo "- ✅ Agent health tied to service availability"
echo "- ✅ Automatic retry and error handling"
echo ""

# Create test request
cat > input/ocr_request.json << 'EOF'
{
  "request_id": "demo_http_001",
  "file_path": "examples/text-extraction/input/ocr_test_image.png",
  "options": {
    "languages": ["eng"],
    "psm": 3,
    "oem": 3,
    "preprocess": true
  }
}
EOF

echo "Test request created: input/ocr_request.json"
echo ""
echo "To start the HTTP OCR cell:"
echo "./build/gox config/cells.yaml"
echo ""
echo "Cell configuration:"
echo "- ID: extraction:http-ocr"
echo "- Agent: ocr-http-stub (await pattern)"
echo "- Service: http://localhost:8080/ocr"
echo "- Ingress: sub:http-ocr-extraction"
echo "- Egress: pub:extracted-text-http"
echo ""
echo "(Press Enter to continue)"
read -p ""

# Demo 5: Production setup
echo "🏭 Demo 5: Production Setup (Load Balanced)"
echo "------------------------------------------"
echo "For production deployments, use load-balanced setup:"
echo ""
echo "Command: ./scripts/build-ocr-http.sh --production"
echo ""
echo "This provides:"
echo "- 📊 2 OCR service instances"
echo "- ⚖️  Nginx load balancer (port 8000)"
echo "- 💾 Redis caching layer"
echo "- 📈 Health monitoring"
echo "- 🔄 Automatic failover"
echo ""

# Demo 6: Performance comparison
echo "📊 Demo 6: Performance Characteristics"
echo "-------------------------------------"
echo "HTTP OCR Stub Performance:"
echo "- 🌐 Network Latency: HTTP overhead"
echo "- 🐳 Complex Setup: Docker service required"
echo "- 📈 Horizontally Scalable: Multiple service instances"
echo "- 🏭 Production Ready: Load balancing, monitoring"
echo "- 🌍 Rich Features: 13+ languages, preprocessing"
echo ""

echo "Comparison with Native Agent:"
echo "┌──────────────┬─────────────────┬──────────────────┐"
echo "│ Aspect       │ Native Agent    │ HTTP OCR Stub    │"
echo "├──────────────┼─────────────────┼──────────────────┤"
echo "│ Setup        │ Simple          │ Service + agent  │"
echo "│ Dependencies │ Local Tesseract │ Docker service   │"
echo "│ Latency      │ Lower           │ Network overhead │"
echo "│ Scalability  │ Process-bound   │ Service scalable │"
echo "│ Production   │ Limited         │ Load-balanced    │"
echo "│ Languages    │ System langs    │ 13+ in container │"
echo "└──────────────┴─────────────────┴──────────────────┘"
echo ""

# Demo 7: Await pattern benefits
echo "⏳ Demo 7: Await Pattern Benefits"
echo "--------------------------------"
echo "The await pattern provides:"
echo "- 🔄 Service Dependency Management"
echo "- 💊 Health Check Integration"
echo "- 🎯 Clean Service Separation"
echo "- 🚀 Framework Compliance"
echo ""
echo "Traditional approach (problematic):"
echo "- Agent starts regardless of service state"
echo "- Manual health checking required"
echo "- Errors when service unavailable"
echo ""
echo "Await pattern (GOX solution):"
echo "- Cell waits for service health"
echo "- Agent connects when ready"
echo "- Integrated error handling"
echo ""

echo "🎉 HTTP OCR Service Demo Complete!"
echo ""
echo "💡 Service Management Commands:"
echo "  # View logs"
echo "  docker compose -f agents/ocr_http_stub/docker/docker-compose.simple.yml logs -f"
echo ""
echo "  # Stop service"
echo "  docker compose -f agents/ocr_http_stub/docker/docker-compose.simple.yml down"
echo ""
echo "  # Restart service"
echo "  docker compose -f agents/ocr_http_stub/docker/docker-compose.simple.yml restart"
echo ""
echo "🔄 For native extraction demo, run: ./demo_native.sh"

# Offer to stop the service
echo ""
read -p "Stop OCR service? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🛑 Stopping OCR service..."
    docker compose -f ../../agents/ocr_http_stub/docker/docker-compose.simple.yml down
    echo "✅ Service stopped"
fi