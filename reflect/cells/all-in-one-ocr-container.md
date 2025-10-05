# All-in-One OCR Container for GOX Framework

## Overview

The GOX framework now supports a comprehensive **All-in-One OCR Container** that provides scalable, containerized OCR services. This container includes everything needed for production OCR processing:

- **Tesseract OCR** with 100+ languages
- **ImageMagick** for image preprocessing
- **Poppler** for PDF-to-image conversion
- **Python Flask HTTP API** with RESTful endpoints
- **Advanced image enhancement** and quality assessment
- **Load balancing** and monitoring support

## 🐳 What's Included

### **Single Container, All Tools**
```dockerfile
# The container includes:
FROM python:3.11-slim

# OCR Core + 100+ Languages
tesseract-ocr with eng, deu, fra, spa, ita, por, rus, ara,
chi-sim, chi-tra, jpn, kor, and many more

# Image Processing
imagemagick + libmagickwand-dev

# PDF Processing
poppler-utils

# Python OCR Service
flask + pytesseract + pillow + opencv + pdf2image
```

### **HTTP API Service**
- **POST /ocr** - Process single file OCR
- **POST /ocr/batch** - Process multiple files
- **GET /health** - Service health check
- **GET /languages** - List available OCR languages
- **GET /info** - Service capabilities and settings
- **GET /** - Web interface for testing

## 🚀 Quick Start

### **1. Build and Deploy**
```bash
# Build the all-in-one OCR service
./scripts/build-ocr-http.sh

# This automatically:
# - Builds the OCR container
# - Starts load-balanced services
# - Performs health checks
# - Shows service endpoints
```

### **2. Single Container (Simple)**
```bash
# Build the container
docker build -t gox-ocr-service ./docker/ocr-service/

# Run single instance
docker run -d --name gox-ocr \
  -p 8080:8080 \
  gox-ocr-service

# Test the service
curl -F "image=@test.png" http://localhost:8080/ocr
```

### **3. Production Setup (Load Balanced)**
```bash
# Use Docker Compose for production
docker-compose -f docker/docker-compose.ocr.yml up -d

# This provides:
# - 2x OCR service instances
# - Nginx load balancer (port 8000)
# - Redis caching
# - Health monitoring
# - Automatic failover
```

## 🔧 Usage with GOX

### **Native vs HTTP OCR Modes**

#### **Native Mode**
```bash
# Uses local tesseract binaries via GOX cell
./build/gox config/cells.yaml  # Uses extraction:native-text cell
```

#### **HTTP Mode (Containerized)**
```bash
# Uses containerized OCR service via GOX cell
./scripts/build-ocr-http.sh  # Start HTTP OCR service
./build/gox config/cells.yaml  # Uses extraction:http-ocr cell
```

### **Configuration Files**

#### **Native OCR Configuration**
```json
{
  "enable_ocr": true,
  "ocr_mode": "native",
  "ocr_config": {
    "languages": ["eng", "deu", "fra"],
    "psm": 3,
    "oem": 3,
    "enable_preprocessing": true
  }
}
```

#### **HTTP OCR Configuration**
```json
{
  "enable_ocr": true,
  "ocr_mode": "http",
  "ocr_service_url": "http://localhost:8000/ocr",
  "timeout": "120s",
  "ocr_config": {
    "languages": ["eng", "deu", "fra", "spa", "ita"],
    "psm": 3,
    "oem": 3,
    "enable_preprocessing": true
  }
}
```

## 📋 Container Features

### **Advanced OCR Processing**
- **Image Preprocessing**: Noise removal, sharpening, adaptive thresholding
- **Quality Assessment**: Confidence scoring and text validation
- **Multi-language Support**: 100+ languages including CJK scripts
- **Format Support**: PNG, JPG, PDF, TIFF, BMP
- **Batch Processing**: Multiple files in single request

### **Production Ready**
- **Load Balancing**: Nginx with health checks and failover
- **Monitoring**: Health endpoints and metrics
- **Scalability**: Horizontal scaling with multiple instances
- **Resource Management**: Configurable memory limits and timeouts
- **Error Handling**: Graceful degradation and retry logic

### **API Examples**

#### **Single File OCR**
```bash
# Basic OCR
curl -F "image=@document.png" http://localhost:8000/ocr

# With custom settings
curl -F "image=@document.pdf" \
     -F "languages=eng+deu" \
     -F "psm=6" \
     -F "preprocess=true" \
     http://localhost:8000/ocr
```

#### **Batch Processing**
```bash
# Multiple files
curl -F "files=@doc1.png" \
     -F "files=@doc2.pdf" \
     -F "files=@doc3.jpg" \
     -F "languages=eng" \
     http://localhost:8000/ocr/batch
```

#### **Service Information**
```bash
# Available languages
curl http://localhost:8000/languages

# Service capabilities
curl http://localhost:8000/info

# Health check
curl http://localhost:8000/health
```

## 🔍 Comparison: Native vs Container

| Aspect | Native Binaries | All-in-One Container |
|--------|-----------------|---------------------|
| **Setup** | `brew install tesseract imagemagick` | `docker-compose up -d` |
| **Startup** | ~100ms | ~2-5s |
| **Memory** | 10-50MB | 200-500MB |
| **Scalability** | Single process | Horizontal scaling |
| **Isolation** | Process-level | Container-level |
| **Languages** | Manual installation | 100+ pre-installed |
| **Updates** | System package manager | Container rebuild |
| **Monitoring** | Mixed with app logs | Separate OCR metrics |
| **Load Balancing** | Not available | Nginx + health checks |
| **Fault Tolerance** | Single point of failure | Service redundancy |

## 🎯 When to Use Each Approach

### **Use Native Binaries When:**
- ✅ Development and testing
- ✅ Single-machine deployments
- ✅ Minimal resource usage required
- ✅ Simple setup preferred
- ✅ OCR is not critical path

### **Use All-in-One Container When:**
- ✅ Production deployments
- ✅ High OCR throughput required
- ✅ Multiple language support needed
- ✅ Fault tolerance and scaling required
- ✅ Separate OCR infrastructure desired
- ✅ Team collaboration (consistent environment)

## 🛠 Management Commands

### **Service Control**
```bash
# Start services
docker-compose -f docker/docker-compose.ocr.yml up -d

# View logs
docker-compose -f docker/docker-compose.ocr.yml logs -f

# Stop services
docker-compose -f docker/docker-compose.ocr.yml down

# Restart services
docker-compose -f docker/docker-compose.ocr.yml restart

# Scale services
docker-compose -f docker/docker-compose.ocr.yml up -d --scale gox-ocr-1=3
```

### **Monitoring**
```bash
# Service status
curl http://localhost:8000/health

# Nginx status
curl http://localhost:8000/nginx_status

# Container stats
docker stats gox-ocr-1 gox-ocr-2

# View service logs
docker logs gox-ocr-1 --tail=50 -f
```

## 📊 Performance Tuning

### **Container Configuration**
```yaml
# docker-compose.ocr.yml
services:
  gox-ocr-1:
    environment:
      - WORKER_PROCESSES=4      # CPU cores
      - MAX_FILE_SIZE=104857600 # 100MB
      - OCR_TIMEOUT=300         # 5 minutes
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '2.0'
```

### **Nginx Load Balancing**
```nginx
# nginx-ocr.conf
upstream ocr_backend {
    least_conn;
    server gox-ocr-1:8080 weight=2;
    server gox-ocr-2:8080 weight=1;
    server gox-ocr-3:8080 weight=1;
}
```

## 🎉 Benefits Summary

The **All-in-One OCR Container** provides:

1. **Complete Package**: Everything needed for OCR in one container
2. **Production Ready**: Load balancing, monitoring, and fault tolerance
3. **Language Rich**: 100+ pre-installed languages
4. **API Driven**: RESTful endpoints compatible with GOX HTTP clients
5. **Scalable**: Horizontal scaling with Docker Compose
6. **Maintainable**: Containerized updates and dependency management
7. **Consistent**: Same environment across development and production

This approach gives you the **best of both worlds**: the simplicity of native binaries for development and the scalability of containerized services for production, all while maintaining full compatibility with the GOX framework's existing OCR integration.