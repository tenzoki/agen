#!/usr/bin/env python3
"""
All-in-One OCR Service for GOX Framework

This HTTP service provides comprehensive OCR capabilities including:
- Image OCR with Tesseract (100+ languages)
- PDF OCR with automatic page splitting
- Image preprocessing with ImageMagick
- Quality assessment and confidence scoring
- Batch processing support
- RESTful API compatible with GOX HTTP OCR clients

API Endpoints:
- POST /ocr - Process single file OCR
- POST /ocr/batch - Process multiple files
- GET /health - Service health check
- GET /languages - List available OCR languages
- GET /info - Service information and capabilities

Compatible with GOX HTTPOCRExtractor implementation.
"""

import os
import json
import tempfile
import logging
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple
import traceback

from flask import Flask, request, jsonify, render_template_string
from flask_cors import CORS
import pytesseract
from PIL import Image, ImageEnhance, ImageFilter
from pdf2image import convert_from_path
import cv2
import numpy as np

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

# Configuration
CONFIG = {
    "service_name": "GOX OCR Service",
    "version": "1.0.0",
    "max_file_size": 50 * 1024 * 1024,  # 50MB
    "supported_formats": [".png", ".jpg", ".jpeg", ".tiff", ".bmp", ".pdf"],
    "default_languages": ["eng"],
    "default_psm": 3,
    "default_oem": 3,
    "temp_dir": "/app/temp",
    "upload_dir": "/app/uploads",
    "output_dir": "/app/output"
}

class OCRProcessor:
    """Comprehensive OCR processing with image enhancement and quality assessment."""

    def __init__(self):
        self.temp_dir = Path(CONFIG["temp_dir"])
        self.temp_dir.mkdir(exist_ok=True)

    def preprocess_image(self, image_path: str, enhance: bool = True) -> str:
        """Apply image preprocessing to improve OCR accuracy."""
        try:
            # Load image
            img = cv2.imread(image_path)
            if img is None:
                # Try with PIL for different formats
                pil_img = Image.open(image_path)
                img = cv2.cvtColor(np.array(pil_img), cv2.COLOR_RGB2BGR)

            # Convert to grayscale
            gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

            if enhance:
                # Apply noise reduction
                denoised = cv2.fastNlMeansDenoising(gray)

                # Apply sharpening
                kernel = np.array([[-1,-1,-1], [-1,9,-1], [-1,-1,-1]])
                sharpened = cv2.filter2D(denoised, -1, kernel)

                # Adaptive thresholding
                processed = cv2.adaptiveThreshold(
                    sharpened, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                    cv2.THRESH_BINARY, 11, 2
                )
            else:
                processed = gray

            # Save processed image
            processed_path = str(self.temp_dir / f"processed_{Path(image_path).name}")
            cv2.imwrite(processed_path, processed)

            return processed_path

        except Exception as e:
            logger.warning(f"Image preprocessing failed: {e}")
            return image_path  # Return original if preprocessing fails

    def extract_text_from_image(self, image_path: str, languages: List[str],
                               psm: int, oem: int, preprocess: bool = True) -> Dict:
        """Extract text from a single image with quality assessment."""
        try:
            # Preprocess image if requested
            if preprocess:
                processed_path = self.preprocess_image(image_path, enhance=True)
            else:
                processed_path = image_path

            # Configure Tesseract
            lang_string = "+".join(languages)
            custom_config = f'--oem {oem} --psm {psm}'

            # Extract text with confidence data
            data = pytesseract.image_to_data(
                processed_path,
                lang=lang_string,
                config=custom_config,
                output_type=pytesseract.Output.DICT
            )

            # Extract plain text
            text = pytesseract.image_to_string(
                processed_path,
                lang=lang_string,
                config=custom_config
            ).strip()

            # Calculate confidence score
            confidences = [int(conf) for conf in data['conf'] if int(conf) > 0]
            avg_confidence = sum(confidences) / len(confidences) if confidences else 0

            # Clean up processed image if different from original
            if processed_path != image_path and os.path.exists(processed_path):
                os.unlink(processed_path)

            return {
                "text": text,
                "confidence": round(avg_confidence, 2),
                "word_count": len(text.split()),
                "char_count": len(text),
                "languages_used": languages,
                "processing_settings": {
                    "psm": psm,
                    "oem": oem,
                    "preprocessed": preprocess
                }
            }

        except Exception as e:
            logger.error(f"OCR extraction failed for {image_path}: {e}")
            return {
                "text": "",
                "confidence": 0.0,
                "error": str(e),
                "word_count": 0,
                "char_count": 0
            }

    def extract_text_from_pdf(self, pdf_path: str, languages: List[str],
                             psm: int, oem: int) -> Dict:
        """Extract text from PDF by converting to images and applying OCR."""
        try:
            # Convert PDF pages to images
            images = convert_from_path(pdf_path, dpi=300)

            all_text = []
            total_confidence = 0
            page_count = 0

            for i, image in enumerate(images):
                # Save page as temporary image
                page_path = str(self.temp_dir / f"page_{i+1}.png")
                image.save(page_path, "PNG")

                # Extract text from page
                page_result = self.extract_text_from_image(
                    page_path, languages, psm, oem, preprocess=True
                )

                if page_result.get("text"):
                    all_text.append(f"=== Page {i+1} ===\n{page_result['text']}")
                    total_confidence += page_result.get("confidence", 0)
                    page_count += 1

                # Clean up page image
                if os.path.exists(page_path):
                    os.unlink(page_path)

            combined_text = "\n\n".join(all_text)
            avg_confidence = total_confidence / page_count if page_count > 0 else 0

            return {
                "text": combined_text,
                "confidence": round(avg_confidence, 2),
                "word_count": len(combined_text.split()),
                "char_count": len(combined_text),
                "page_count": len(images),
                "pages_processed": page_count,
                "languages_used": languages,
                "processing_settings": {
                    "psm": psm,
                    "oem": oem,
                    "dpi": 300
                }
            }

        except Exception as e:
            logger.error(f"PDF OCR extraction failed for {pdf_path}: {e}")
            return {
                "text": "",
                "confidence": 0.0,
                "error": str(e),
                "word_count": 0,
                "char_count": 0,
                "page_count": 0
            }

# Initialize OCR processor
ocr_processor = OCRProcessor()

@app.route('/health', methods=['GET'])
def health_check():
    """Service health check endpoint."""
    try:
        # Test Tesseract availability
        languages = pytesseract.get_languages()

        return jsonify({
            "status": "healthy",
            "service": CONFIG["service_name"],
            "version": CONFIG["version"],
            "timestamp": datetime.now().isoformat(),
            "tesseract_available": True,
            "language_count": len(languages),
            "capabilities": [
                "image_ocr",
                "pdf_ocr",
                "image_preprocessing",
                "batch_processing",
                "confidence_scoring"
            ]
        })
    except Exception as e:
        return jsonify({
            "status": "unhealthy",
            "error": str(e),
            "timestamp": datetime.now().isoformat()
        }), 500

@app.route('/languages', methods=['GET'])
def get_languages():
    """Get list of available OCR languages."""
    try:
        languages = pytesseract.get_languages()
        return jsonify({
            "languages": sorted(languages),
            "count": len(languages),
            "default": CONFIG["default_languages"]
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/info', methods=['GET'])
def get_info():
    """Get service information and capabilities."""
    return jsonify({
        "service": CONFIG["service_name"],
        "version": CONFIG["version"],
        "max_file_size": CONFIG["max_file_size"],
        "supported_formats": CONFIG["supported_formats"],
        "default_settings": {
            "languages": CONFIG["default_languages"],
            "psm": CONFIG["default_psm"],
            "oem": CONFIG["default_oem"]
        },
        "endpoints": {
            "/ocr": "POST - Single file OCR processing",
            "/ocr/batch": "POST - Batch file OCR processing",
            "/health": "GET - Service health check",
            "/languages": "GET - Available OCR languages",
            "/info": "GET - Service information"
        }
    })

@app.route('/ocr', methods=['POST'])
def process_ocr():
    """Process single file OCR request."""
    try:
        # Check if file is present
        if 'file' not in request.files and 'image' not in request.files:
            return jsonify({
                "error": "No file provided. Use 'file' or 'image' field."
            }), 400

        # Get file (support both 'file' and 'image' field names)
        file = request.files.get('file') or request.files.get('image')

        if file.filename == '':
            return jsonify({"error": "No file selected"}), 400

        # Check file size
        file.seek(0, os.SEEK_END)
        file_size = file.tell()
        file.seek(0)

        if file_size > CONFIG["max_file_size"]:
            return jsonify({
                "error": f"File too large. Max size: {CONFIG['max_file_size']} bytes"
            }), 400

        # Get parameters
        languages = request.form.get('languages', 'eng').split('+')
        psm = int(request.form.get('psm', CONFIG["default_psm"]))
        oem = int(request.form.get('oem', CONFIG["default_oem"]))
        preprocess = request.form.get('preprocess', 'true').lower() == 'true'

        # Save uploaded file
        file_ext = Path(file.filename).suffix.lower()
        if file_ext not in CONFIG["supported_formats"]:
            return jsonify({
                "error": f"Unsupported format: {file_ext}. Supported: {CONFIG['supported_formats']}"
            }), 400

        temp_path = str(ocr_processor.temp_dir / f"upload_{datetime.now().strftime('%Y%m%d_%H%M%S')}{file_ext}")
        file.save(temp_path)

        try:
            # Process based on file type
            if file_ext == '.pdf':
                result = ocr_processor.extract_text_from_pdf(temp_path, languages, psm, oem)
            else:
                result = ocr_processor.extract_text_from_image(temp_path, languages, psm, oem, preprocess)

            # Add metadata
            result.update({
                "filename": file.filename,
                "file_size": file_size,
                "processing_time": datetime.now().isoformat(),
                "service": CONFIG["service_name"]
            })

            return jsonify(result)

        finally:
            # Clean up uploaded file
            if os.path.exists(temp_path):
                os.unlink(temp_path)

    except Exception as e:
        logger.error(f"OCR processing error: {e}")
        logger.error(traceback.format_exc())
        return jsonify({
            "error": "Internal server error",
            "details": str(e)
        }), 500

@app.route('/ocr/batch', methods=['POST'])
def process_batch_ocr():
    """Process multiple files in batch."""
    try:
        files = request.files.getlist('files')
        if not files:
            return jsonify({"error": "No files provided"}), 400

        # Get parameters
        languages = request.form.get('languages', 'eng').split('+')
        psm = int(request.form.get('psm', CONFIG["default_psm"]))
        oem = int(request.form.get('oem', CONFIG["default_oem"]))

        results = []

        for file in files:
            if file.filename == '':
                continue

            try:
                # Process individual file (simplified version)
                file_ext = Path(file.filename).suffix.lower()
                if file_ext not in CONFIG["supported_formats"]:
                    results.append({
                        "filename": file.filename,
                        "error": f"Unsupported format: {file_ext}"
                    })
                    continue

                temp_path = str(ocr_processor.temp_dir / f"batch_{datetime.now().strftime('%Y%m%d_%H%M%S')}_{file.filename}")
                file.save(temp_path)

                if file_ext == '.pdf':
                    result = ocr_processor.extract_text_from_pdf(temp_path, languages, psm, oem)
                else:
                    result = ocr_processor.extract_text_from_image(temp_path, languages, psm, oem)

                result["filename"] = file.filename
                results.append(result)

                # Clean up
                if os.path.exists(temp_path):
                    os.unlink(temp_path)

            except Exception as e:
                results.append({
                    "filename": file.filename,
                    "error": str(e)
                })

        return jsonify({
            "batch_results": results,
            "total_files": len(files),
            "processed_files": len([r for r in results if "error" not in r]),
            "failed_files": len([r for r in results if "error" in r])
        })

    except Exception as e:
        logger.error(f"Batch OCR processing error: {e}")
        return jsonify({
            "error": "Batch processing failed",
            "details": str(e)
        }), 500

@app.route('/', methods=['GET'])
def index():
    """Simple web interface for testing."""
    html = """
    <!DOCTYPE html>
    <html>
    <head>
        <title>GOX OCR Service</title>
        <style>
            body { font-family: Arial, sans-serif; margin: 40px; }
            .container { max-width: 800px; margin: 0 auto; }
            .upload-form { border: 1px solid #ccc; padding: 20px; margin: 20px 0; }
            .result { background: #f5f5f5; padding: 15px; margin: 10px 0; }
            button { background: #007cba; color: white; padding: 10px 20px; border: none; cursor: pointer; }
            button:hover { background: #005a87; }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>GOX OCR Service</h1>
            <p>All-in-One OCR Service for the GOX Framework</p>

            <div class="upload-form">
                <h3>Test OCR Processing</h3>
                <form action="/ocr" method="post" enctype="multipart/form-data">
                    <p>
                        <label>File:</label><br>
                        <input type="file" name="image" accept=".png,.jpg,.jpeg,.pdf,.tiff,.bmp" required>
                    </p>
                    <p>
                        <label>Languages:</label><br>
                        <input type="text" name="languages" value="eng" placeholder="eng+deu+fra">
                    </p>
                    <p>
                        <label>PSM Mode:</label><br>
                        <select name="psm">
                            <option value="3" selected>3 - Fully automatic page segmentation</option>
                            <option value="6">6 - Single uniform block of text</option>
                            <option value="8">8 - Single word</option>
                            <option value="13">13 - Raw line. Treat image as single text line</option>
                        </select>
                    </p>
                    <p>
                        <button type="submit">Process OCR</button>
                    </p>
                </form>
            </div>

            <div class="result">
                <h3>API Endpoints</h3>
                <ul>
                    <li><strong>POST /ocr</strong> - Process single file</li>
                    <li><strong>POST /ocr/batch</strong> - Process multiple files</li>
                    <li><strong>GET /health</strong> - Service health check</li>
                    <li><strong>GET /languages</strong> - Available languages</li>
                    <li><strong>GET /info</strong> - Service information</li>
                </ul>
            </div>
        </div>
    </body>
    </html>
    """
    return html

if __name__ == '__main__':
    # Ensure directories exist
    for dir_path in [CONFIG["temp_dir"], CONFIG["upload_dir"], CONFIG["output_dir"]]:
        os.makedirs(dir_path, exist_ok=True)

    # Start the service
    logger.info(f"Starting {CONFIG['service_name']} v{CONFIG['version']}")
    logger.info(f"Available languages: {len(pytesseract.get_languages())}")

    app.run(
        host='0.0.0.0',
        port=8080,
        debug=False,
        threaded=True
    )