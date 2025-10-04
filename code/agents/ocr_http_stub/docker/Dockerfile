# All-in-One OCR Service Container for GOX Framework
#
# This container includes:
# - Tesseract OCR with 100+ languages
# - ImageMagick for image preprocessing
# - Poppler for PDF-to-image conversion
# - Python Flask HTTP API service
# - Auto-detection of optimal OCR settings
#
# Usage:
#   docker build -t gox-ocr-service .
#   docker run -d -p 8080:8080 --name gox-ocr gox-ocr-service

FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    # OCR Core
    tesseract-ocr \
    tesseract-ocr-eng \
    tesseract-ocr-deu \
    tesseract-ocr-fra \
    tesseract-ocr-spa \
    tesseract-ocr-ita \
    tesseract-ocr-por \
    tesseract-ocr-rus \
    tesseract-ocr-ara \
    tesseract-ocr-chi-sim \
    tesseract-ocr-chi-tra \
    tesseract-ocr-jpn \
    tesseract-ocr-kor \
    \
    # Image Processing
    imagemagick \
    libmagickwand-dev \
    \
    # PDF Processing
    poppler-utils \
    \
    # Additional utilities
    curl \
    wget \
    file \
    \
    && rm -rf /var/lib/apt/lists/*

# Configure ImageMagick security policy for PDF processing
RUN if [ -f /etc/ImageMagick-6/policy.xml ]; then \
        sed -i 's/rights="none" pattern="PDF"/rights="read|write" pattern="PDF"/' /etc/ImageMagick-6/policy.xml; \
    fi && \
    if [ -f /etc/ImageMagick-7/policy.xml ]; then \
        sed -i 's/rights="none" pattern="PDF"/rights="read|write" pattern="PDF"/' /etc/ImageMagick-7/policy.xml; \
    fi

# Install Python dependencies
COPY requirements.txt /app/
WORKDIR /app
RUN pip install --no-cache-dir -r requirements.txt

# Copy OCR service application
COPY ocr_service.py /app/
COPY config.json /app/

# Copy templates and static files (create empty if not exist)
COPY templates/ /app/templates/
COPY static/ /app/static/

# Create directories for processing
RUN mkdir -p /app/uploads /app/temp /app/output

# Expose HTTP service port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the OCR service
CMD ["python", "ocr_service.py"]