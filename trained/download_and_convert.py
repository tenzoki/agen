#!/usr/bin/env python3
"""
Download and convert HuggingFace models to ONNX format for Gox anonymization pipeline.

Usage:
    python download_and_convert.py [--small]

Options:
    --small    Use smaller model variants (for development/testing)
"""

import os
import sys
import argparse
from pathlib import Path
import torch
from transformers import (
    AutoTokenizer,
    AutoModelForTokenClassification,
    AutoModel,
)
from optimum.onnxruntime import ORTModelForTokenClassification, ORTModelForFeatureExtraction
import onnxruntime as ort


# Model configurations
MODELS = {
    "ner": {
        "large": "xlm-roberta-large-finetuned-conll03-english",
        "small": "Davlan/xlm-roberta-base-ner-hrl",
    },
    "coref": {
        "large": "biu-nlp/f-coref",  # SpanBERT-based coreference
        "small": "biu-nlp/f-coref",  # Same for now, no smaller variant
    },
    "embeddings": {
        "large": "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2",
        "small": "sentence-transformers/paraphrase-multilingual-mpnet-base-v2",
    }
}


def print_header(text):
    """Print formatted section header."""
    print(f"\n{'='*70}")
    print(f"  {text}")
    print(f"{'='*70}\n")


def verify_onnx_model(model_path):
    """Verify ONNX model can be loaded."""
    print(f"  Verifying ONNX model: {model_path}")
    try:
        session = ort.InferenceSession(str(model_path))
        inputs = session.get_inputs()
        outputs = session.get_outputs()

        print(f"  ✓ Model loaded successfully")
        print(f"  ✓ Inputs: {[inp.name for inp in inputs]}")
        print(f"  ✓ Outputs: {[out.name for out in outputs]}")
        return True
    except Exception as e:
        print(f"  ✗ Verification failed: {e}")
        return False


def convert_ner_model(model_name, output_dir):
    """Convert NER model to ONNX."""
    print_header(f"Converting NER Model: {model_name}")

    output_path = Path(output_dir) / "ner"
    output_path.mkdir(exist_ok=True, parents=True)

    print(f"  Downloading model from HuggingFace...")
    tokenizer = AutoTokenizer.from_pretrained(model_name)

    print(f"  Converting to ONNX format...")
    # Use optimum for conversion
    model = ORTModelForTokenClassification.from_pretrained(
        model_name,
        export=True,
    )

    # Save ONNX model
    onnx_path = output_path / "xlm-roberta-ner.onnx"
    model.save_pretrained(output_path)
    tokenizer.save_pretrained(output_path)

    # Move model.onnx to our naming convention
    if (output_path / "model.onnx").exists():
        (output_path / "model.onnx").rename(onnx_path)

    print(f"  ✓ Model saved to: {onnx_path}")

    # Verify
    if verify_onnx_model(onnx_path):
        print(f"  ✓ NER model ready!")
        return True
    return False


def convert_coref_model(model_name, output_dir):
    """Convert Coreference model to ONNX."""
    print_header(f"Converting Coreference Model: {model_name}")

    output_path = Path(output_dir) / "coref"
    output_path.mkdir(exist_ok=True, parents=True)

    print(f"  Downloading model from HuggingFace...")

    # Note: F-Coref may need special handling
    # For now, we'll use a generic approach
    try:
        tokenizer = AutoTokenizer.from_pretrained(model_name)
        model = AutoModel.from_pretrained(model_name)

        print(f"  Converting to ONNX format...")

        # Prepare dummy input
        dummy_text = "This is a test sentence."
        inputs = tokenizer(dummy_text, return_tensors="pt")

        # Export to ONNX
        onnx_path = output_path / "spanbert-coref.onnx"
        torch.onnx.export(
            model,
            tuple(inputs.values()),
            str(onnx_path),
            input_names=['input_ids', 'attention_mask'],
            output_names=['output'],
            dynamic_axes={
                'input_ids': {0: 'batch', 1: 'sequence'},
                'attention_mask': {0: 'batch', 1: 'sequence'},
                'output': {0: 'batch', 1: 'sequence'}
            },
            opset_version=14
        )

        tokenizer.save_pretrained(output_path)

        print(f"  ✓ Model saved to: {onnx_path}")

        if verify_onnx_model(onnx_path):
            print(f"  ✓ Coref model ready!")
            return True
    except Exception as e:
        print(f"  ⚠ Coref model conversion failed: {e}")
        print(f"  ℹ This is expected - coref models need special handling")
        print(f"  ℹ Consider using a simpler span extraction approach for now")

    return False


def convert_embedding_model(model_name, output_dir):
    """Convert sentence embedding model to ONNX."""
    print_header(f"Converting Embedding Model: {model_name}")

    output_path = Path(output_dir) / "embeddings"
    output_path.mkdir(exist_ok=True, parents=True)

    print(f"  Downloading model from HuggingFace...")
    tokenizer = AutoTokenizer.from_pretrained(model_name)

    print(f"  Converting to ONNX format...")
    model = ORTModelForFeatureExtraction.from_pretrained(
        model_name,
        export=True,
    )

    # Save ONNX model
    onnx_path = output_path / "multilingual-minilm.onnx"
    model.save_pretrained(output_path)
    tokenizer.save_pretrained(output_path)

    # Move model.onnx to our naming convention
    if (output_path / "model.onnx").exists():
        (output_path / "model.onnx").rename(onnx_path)

    print(f"  ✓ Model saved to: {onnx_path}")

    if verify_onnx_model(onnx_path):
        print(f"  ✓ Embedding model ready!")
        return True
    return False


def main():
    parser = argparse.ArgumentParser(description='Convert HF models to ONNX for Gox')
    parser.add_argument('--small', action='store_true',
                       help='Use smaller model variants')
    parser.add_argument('--skip-coref', action='store_true',
                       help='Skip coreference model (complex conversion)')
    args = parser.parse_args()

    size = "small" if args.small else "large"
    output_dir = Path(__file__).parent

    print_header("Gox Anonymization Model Converter")
    print(f"  Model size: {size}")
    print(f"  Output directory: {output_dir}")

    results = {}

    # Convert NER model
    results['ner'] = convert_ner_model(
        MODELS['ner'][size],
        output_dir
    )

    # Convert Coref model (optional, often fails)
    if not args.skip_coref:
        results['coref'] = convert_coref_model(
            MODELS['coref'][size],
            output_dir
        )
    else:
        print_header("Skipping Coreference Model")
        results['coref'] = None

    # Convert Embedding model
    results['embeddings'] = convert_embedding_model(
        MODELS['embeddings'][size],
        output_dir
    )

    # Summary
    print_header("Conversion Summary")
    print(f"  NER Model:        {'✓ Success' if results['ner'] else '✗ Failed'}")
    print(f"  Coref Model:      {'✓ Success' if results['coref'] else '⚠ Skipped/Failed'}")
    print(f"  Embedding Model:  {'✓ Success' if results['embeddings'] else '✗ Failed'}")

    if results['ner'] and results['embeddings']:
        print(f"\n  ✓ Core models ready! You can start implementing agents.")
        print(f"  ℹ Coref model is optional - can implement later or use rule-based approach")
        return 0
    else:
        print(f"\n  ✗ Some conversions failed. Check errors above.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
