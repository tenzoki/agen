package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tenzoki/agen/alfa/internal/gox"
	"github.com/tenzoki/agen/alfa/internal/tools"
	"github.com/tenzoki/agen/alfa/internal/vfs"
)

func main() {
	fmt.Println("🔒 Gox Anonymization & NER Demo")
	fmt.Println("================================")
	fmt.Println()

	// Initialize VFS
	projectDir := "/tmp/demo-anonymization-project"
	projectVFS := vfs.New(projectDir)

	// Initialize Gox Manager
	fmt.Println("📦 Initializing Gox Manager...")
	goxMgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "config/gox",
		DefaultDataRoot: "/tmp/gox-demo",
		Debug:           true,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Gox: %v", err)
	}
	defer goxMgr.Close()

	// Initialize tool dispatcher
	dispatcher := tools.NewDispatcher(projectVFS)
	dispatcher.SetGoxManager(goxMgr)

	fmt.Println("✓ Gox Manager initialized")
	fmt.Println()

	// Demo 1: Extract Named Entities
	fmt.Println("=== Demo 1: Named Entity Recognition ===")
	demonstrateNER(dispatcher)
	fmt.Println()

	// Demo 2: Anonymize Text
	fmt.Println("=== Demo 2: Text Anonymization ===")
	demonstrateAnonymization(dispatcher)
	fmt.Println()

	// Demo 3: Deanonymize Text
	fmt.Println("=== Demo 3: Text Deanonymization ===")
	demonstrateDeanonymization(dispatcher)
	fmt.Println()

	fmt.Println("✓ Demo completed successfully")
}

// demonstrateNER shows named entity recognition
func demonstrateNER(dispatcher *tools.Dispatcher) {
	sampleText := "Angela Merkel met with Emmanuel Macron in Berlin to discuss the European Union's climate policy. They visited the headquarters of Siemens AG."

	fmt.Printf("📄 Input Text:\n%s\n\n", sampleText)

	action := tools.Action{
		Type: "extract_entities",
		Params: map[string]interface{}{
			"text":       sampleText,
			"project_id": "demo-project",
			"timeout":    30.0,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("🔍 Extracting entities...")
	result := dispatcher.Execute(ctx, action)

	if !result.Success {
		fmt.Printf("❌ Error: %s\n", result.Message)
		return
	}

	// Display results
	if output, ok := result.Output.(map[string]interface{}); ok {
		entities, _ := output["entities"].([]interface{})
		fmt.Printf("✓ Found %d entities:\n\n", len(entities))

		for i, entity := range entities {
			if e, ok := entity.(map[string]interface{}); ok {
				fmt.Printf("  %d. %s (%s) - confidence: %.2f\n",
					i+1,
					e["text"],
					e["type"],
					e["confidence"],
				)
			}
		}
	}
}

// demonstrateAnonymization shows text anonymization
func demonstrateAnonymization(dispatcher *tools.Dispatcher) {
	sampleText := "John Smith works at OpenAI in San Francisco. His colleague Jane Doe is based in London."

	fmt.Printf("📄 Original Text:\n%s\n\n", sampleText)

	action := tools.Action{
		Type: "anonymize_text",
		Params: map[string]interface{}{
			"text":       sampleText,
			"project_id": "demo-project",
			"timeout":    45.0,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	fmt.Println("🔒 Anonymizing text...")
	result := dispatcher.Execute(ctx, action)

	if !result.Success {
		fmt.Printf("❌ Error: %s\n", result.Message)
		return
	}

	// Display results
	if output, ok := result.Output.(map[string]interface{}); ok {
		anonymizedText, _ := output["anonymized_text"].(string)
		mappings, _ := output["mappings"].(map[string]interface{})
		entityCount, _ := output["entity_count"].(int)

		fmt.Printf("✓ Anonymization complete\n\n")
		fmt.Printf("📝 Anonymized Text:\n%s\n\n", anonymizedText)
		fmt.Printf("🔑 Entity Mappings (%d):\n", entityCount)

		for original, pseudonym := range mappings {
			fmt.Printf("  • %s → %s\n", original, pseudonym)
		}
	}
}

// demonstrateDeanonymization shows text restoration
func demonstrateDeanonymization(dispatcher *tools.Dispatcher) {
	// Sample anonymized data (from previous anonymization)
	anonymizedText := "PERSON_123456 works at ORG_789012 in LOC_345678. His colleague PERSON_234567 is based in LOC_456789."
	mappings := map[string]interface{}{
		"John Smith":     "PERSON_123456",
		"OpenAI":         "ORG_789012",
		"San Francisco":  "LOC_345678",
		"Jane Doe":       "PERSON_234567",
		"London":         "LOC_456789",
	}

	fmt.Printf("📄 Anonymized Text:\n%s\n\n", anonymizedText)

	action := tools.Action{
		Type: "deanonymize_text",
		Params: map[string]interface{}{
			"anonymized_text": anonymizedText,
			"mappings":        mappings,
			"project_id":      "demo-project",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("🔓 Deanonymizing text...")
	result := dispatcher.Execute(ctx, action)

	if !result.Success {
		fmt.Printf("❌ Error: %s\n", result.Message)
		return
	}

	// Display results
	if output, ok := result.Output.(map[string]interface{}); ok {
		restoredText, _ := output["restored_text"].(string)
		replacements, _ := output["replacements"].(int)

		fmt.Printf("✓ Deanonymization complete\n\n")
		fmt.Printf("📝 Restored Text:\n%s\n\n", restoredText)
		fmt.Printf("🔑 Replacements: %d\n", replacements)
	}
}
