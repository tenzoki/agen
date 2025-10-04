# GOX OCR Container - Quick Start Guide

## 🚀 Quick Start (30 seconds)

```bash
# From the ocr_http_stub agent directory
../../../scripts/build-ocr-http.sh
```

**That's it!** Your OCR service is now running at:
- **🌐 Web Interface**: http://localhost:8080
- **🔗 API Endpoint**: http://localhost:8080/ocr

## 🧪 Test It Now

```bash
# Create test image
echo "Hello OCR World!" | magick -pointsize 24 -size 400x80 -gravity center label:@- test.png

# Test OCR (or use any image from examples/data/images/)
curl -F "image=@examples/data/images/test_sample.png" http://localhost:8080/ocr

# Expected result:
# {"text":"Hello OCR World!","confidence":95.2,"word_count":3,...}
```

## 🎯 GOX Framework Integration

```bash
# Build GOX orchestrator and OCR HTTP stub
cd /path/to/gox && make build-all

# Start HTTP OCR service first
./scripts/build-ocr-http.sh

# Use containerized OCR via GOX cells
./build/gox config/cells.yaml  # Uses extraction:http-ocr cell
```

## 📋 Available Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/` | GET | Web interface for testing |
| `/ocr` | POST | Single file OCR processing |
| `/ocr/batch` | POST | Multiple file processing |
| `/health` | GET | Service health check |
| `/languages` | GET | Available OCR languages |
| `/info` | GET | Service capabilities |

## 🌍 Multi-Language Examples

```bash
# English + German + French
curl -F "image=@document.png" -F "languages=eng+deu+fra" http://localhost:8080/ocr

# Chinese Simplified
curl -F "image=@chinese.png" -F "languages=chi_sim" http://localhost:8080/ocr

# Japanese
curl -F "image=@japanese.png" -F "languages=jpn" http://localhost:8080/ocr

# Available languages: eng, deu, fra, spa, ita, por, rus, ara, chi_sim, chi_tra, jpn, kor
```

## 📄 PDF Processing

```bash
# OCR a PDF document
curl -F "image=@scanned_document.pdf" -F "languages=eng" http://localhost:8080/ocr

# The service automatically:
# 1. Converts PDF pages to images
# 2. Applies OCR to each page
# 3. Combines results with page markers
```

## ⚙️ Advanced Settings

```bash
# Custom OCR parameters
curl -F "image=@document.png" \
     -F "languages=eng" \
     -F "psm=6" \
     -F "oem=3" \
     -F "preprocess=true" \
     http://localhost:8080/ocr

# PSM (Page Segmentation Mode):
#   3 = Fully automatic page segmentation (default)
#   6 = Single uniform block of text
#   8 = Single word
#   13 = Raw line, single text line

# OEM (OCR Engine Mode):
#   3 = Default, based on what is available (default)
#   1 = Original Tesseract only
#   2 = Neural nets LSTM only
```

## 🔄 Batch Processing

```bash
# Process multiple files at once
curl -F "files=@doc1.png" \
     -F "files=@doc2.pdf" \
     -F "files=@doc3.jpg" \
     -F "languages=eng" \
     http://localhost:8080/ocr/batch

# Returns results for all files with success/failure status
```

## 🛠 Management Commands

```bash
# View service logs
docker compose -f docker-compose.simple.yml logs -f

# Check service status
curl -s http://localhost:8080/health | jq

# Stop the service
docker compose -f docker-compose.simple.yml down

# Restart the service
docker compose -f docker-compose.simple.yml restart

# Start production setup (load balanced)
.../../../scripts/build-ocr-http.sh --production
```

## 📊 Production Setup

For production deployments with load balancing:

```bash
# Start production mode
.../../../scripts/build-ocr-http.sh --production

# This provides:
# - 2 OCR service instances
# - Nginx load balancer on port 8000
# - Redis caching
# - Health monitoring
# - Automatic failover

# Access via load balancer
curl -F "image=@document.png" http://localhost:8000/ocr
```

## 🧬 Service Health Monitoring

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed service info
curl http://localhost:8080/info

# Available languages
curl http://localhost:8080/languages

# Example health response:
# {
#   "status": "healthy",
#   "service": "GOX OCR Service",
#   "version": "1.0.0",
#   "tesseract_available": true,
#   "language_count": 13,
#   "capabilities": ["image_ocr", "pdf_ocr", "batch_processing"]
# }
```

## 🔧 Configuration Options

### **Environment Variables**
```bash
# Customize service behavior
docker run -e SERVICE_NAME="Custom OCR" \
           -e MAX_FILE_SIZE=104857600 \
           -p 8080:8080 \
           gox-ocr-service
```

### **Volume Mounts**
```bash
# Persist temporary files and uploads
docker run -v /host/ocr-temp:/app/temp \
           -v /host/ocr-uploads:/app/uploads \
           -p 8080:8080 \
           gox-ocr-service
```

## ❓ Troubleshooting

### **Service Won't Start**
```bash
# Check container logs
docker logs gox-ocr-simple

# Check if port is already in use
lsof -i :8080

# Remove conflicting containers
docker rm -f $(docker ps -aq --filter ancestor=gox-ocr-service)
```

### **OCR Quality Issues**
```bash
# Try different PSM modes
curl -F "image=@document.png" -F "psm=6" http://localhost:8080/ocr    # Single block
curl -F "image=@document.png" -F "psm=8" http://localhost:8080/ocr    # Single word
curl -F "image=@document.png" -F "psm=13" http://localhost:8080/ocr   # Single line

# Enable preprocessing for poor quality images
curl -F "image=@blurry.png" -F "preprocess=true" http://localhost:8080/ocr
```

### **File Upload Issues**
```bash
# Check file size (max 50MB by default)
ls -lh document.pdf

# Check file format support
curl http://localhost:8080/info | jq '.supported_formats'
# Supported: .png, .jpg, .jpeg, .tiff, .bmp, .pdf
```

## 🎉 Success Indicators

You'll know everything is working when:

✅ **Service Health**: `curl http://localhost:8080/health` returns `"status":"healthy"`
✅ **OCR Processing**: Text extraction returns confidence scores >80%
✅ **GOX Integration**: `--ocr-mode=http` shows "OCR service URL" in logs
✅ **Web Interface**: http://localhost:8080 shows upload form
✅ **Multi-Language**: Multiple languages detected and processed

## 📚 Next Steps

- **Integrate with GOX pipelines**: Use `--ocr-mode=http` in your text extraction workflows
- **Scale for production**: Use `../../../scripts/build-ocr-http.sh --production` for load balancing
- **Custom languages**: Add more Tesseract language packs to the container
- **API integration**: Build applications using the REST API endpoints
- **Monitoring**: Set up log aggregation and metrics collection

**🚀 You're ready to process documents at scale with containerized OCR!**