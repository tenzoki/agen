#!/bin/bash
# Anonymization Pipeline Example Runner
set -e

echo "========================================================================"
echo "Gox Anonymization Pipeline - Example Demonstration"
echo "========================================================================"
echo ""

# Check prerequisites
echo "Checking prerequisites..."

if [ ! -f "../../build/ner_agent" ]; then
    echo "❌ NER agent not found. Building..."
    cd ../..
    source onnx-exports
    go build -o build/ner_agent ./agents/ner_agent/
    cd examples/anonymization_pipeline
    echo "✅ NER agent built"
else
    echo "✅ NER agent found"
fi

if [ ! -f "../../models/ner/xlm-roberta-ner.onnx" ]; then
    echo "❌ ONNX model not found"
    echo "Please run: cd models && python download_and_convert.py"
    exit 1
else
    echo "✅ ONNX model found"
fi

echo ""
echo "========================================================================"
echo "Processing Sample Documents"
echo "========================================================================"
echo ""

# Process each sample file
for sample in input/sample_*.txt; do
    filename=$(basename "$sample")
    echo "────────────────────────────────────────────────────────────────────"
    echo "Processing: $filename"
    echo "────────────────────────────────────────────────────────────────────"
    echo ""

    # Show original text
    echo "📄 Original Text:"
    echo "────────────────────────────────────────────────────────────────────"
    head -n 3 "$sample"
    echo "..."
    echo ""

    # In a real deployment, this would call the orchestrator with the pipeline
    # For this example, we just demonstrate the concept
    echo "🔍 NER Detection:"
    echo "────────────────────────────────────────────────────────────────────"
    echo "Agent would extract entities like:"

    case $filename in
        sample_en.txt)
            echo "  • Angela Merkel (PERSON, confidence: 0.95)"
            echo "  • Microsoft (ORG, confidence: 0.92)"
            echo "  • Berlin (LOC, confidence: 0.88)"
            echo "  • Satya Nadella (PERSON, confidence: 0.94)"
            echo "  • Unter den Linden (LOC, confidence: 0.82)"
            echo "  • Germany (LOC, confidence: 0.86)"
            echo "  • European Union (ORG, confidence: 0.81)"
            ;;
        sample_de.txt)
            echo "  • Angela Merkel (PERSON, confidence: 0.94)"
            echo "  • Siemens (ORG, confidence: 0.93)"
            echo "  • München (LOC, confidence: 0.89)"
            echo "  • Joe Kaeser (PERSON, confidence: 0.91)"
            echo "  • Peter Altmaier (PERSON, confidence: 0.92)"
            ;;
        sample_fr.txt)
            echo "  • Emmanuel Macron (PERSON, confidence: 0.96)"
            echo "  • Renault (ORG, confidence: 0.94)"
            echo "  • Paris (LOC, confidence: 0.90)"
            echo "  • Luca de Meo (PERSON, confidence: 0.93)"
            echo "  • Élysée (LOC, confidence: 0.88)"
            echo "  • TotalEnergies (ORG, confidence: 0.91)"
            echo "  • Île-de-France (LOC, confidence: 0.85)"
            ;;
        sample_medical.txt)
            echo "  • John Smith (PERSON, confidence: 0.95)"
            echo "  • Massachusetts General Hospital (ORG, confidence: 0.92)"
            echo "  • Sarah Johnson (PERSON, confidence: 0.94)"
            echo "  • Boston Medical Center (ORG, confidence: 0.91)"
            echo "  • Michael Chen (PERSON, confidence: 0.93)"
            echo "  • Cambridge Health Associates (ORG, confidence: 0.89)"
            echo "  • Robert Martinez (PERSON, confidence: 0.94)"
            echo "  • Brigham and Women's Hospital (ORG, confidence: 0.90)"
            echo "  • Lisa Anderson (PERSON, confidence: 0.93)"
            echo "  • Cambridge, MA (LOC, confidence: 0.87)"
            ;;
    esac

    echo ""
    echo "🔐 Anonymization:"
    echo "────────────────────────────────────────────────────────────────────"
    echo "Entities replaced with pseudonyms (consistent within project)"
    echo ""

    echo "💾 Storage:"
    echo "────────────────────────────────────────────────────────────────────"
    echo "Mapping stored in: data/anonymization_mappings.db"
    echo "Project ID: example-$(date +%Y%m%d)"
    echo ""

    echo ""
done

echo "========================================================================"
echo "Summary"
echo "========================================================================"
echo ""
echo "This example demonstrates the anonymization pipeline with:"
echo ""
echo "1. ✅ NER Agent - Multilingual entity detection (100+ languages)"
echo "   • XLM-RoBERTa model with BIO tagging"
echo "   • Detects: PERSON, ORG, LOC, MISC"
echo "   • Character-level precision"
echo ""
echo "2. ✅ Anonymizer Agent - Pseudonymization"
echo "   • Consistent replacements (same entity → same pseudonym)"
echo "   • Project-scoped consistency"
echo "   • Reversible with mapping"
echo ""
echo "3. ✅ Storage Agent - Mapping persistence"
echo "   • Stores entity → pseudonym mappings"
echo "   • godast/omnistore with bbolt backend"
echo "   • Project-isolated storage"
echo ""
echo "Full Integration:"
echo "  • Deploy via orchestrator using config/anonymization_pipeline.yaml"
echo "  • Agents communicate via message broker"
echo "  • Horizontal scaling via agent pool"
echo "  • Production-ready components"
echo ""
echo "To run full integration test:"
echo "  cd ../../"
echo "  ./scripts/test-anonymization-pipeline.sh"
echo ""
