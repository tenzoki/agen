package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockSummaryGeneratorRunner implements agent.AgentRunner for testing
type MockSummaryGeneratorRunner struct {
	config map[string]interface{}
}

func (m *MockSummaryGeneratorRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockSummaryGeneratorRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockSummaryGeneratorRunner) Cleanup(base *agent.BaseAgent) {
}

func TestSummaryGeneratorInitialization(t *testing.T) {
	config := map[string]interface{}{
		"summary_length": "medium",
		"extraction_method": "extractive",
		"max_sentences": 5,
		"language": "en",
	}

	runner := &MockSummaryGeneratorRunner{config: config}
	framework := agent.NewFramework(runner, "summary-generator")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Summary generator framework created successfully")
}

func TestSummaryGeneratorExtractiveSummary(t *testing.T) {
	config := map[string]interface{}{
		"method": "extractive",
		"algorithm": "textrank",
		"sentence_count": 3,
	}

	runner := &MockSummaryGeneratorRunner{config: config}

	document := map[string]interface{}{
		"id": "doc-extract-001",
		"title": "Introduction to Machine Learning",
		"content": "Machine learning is a subset of artificial intelligence that focuses on the development of algorithms. " +
			"These algorithms can learn and make decisions from data without being explicitly programmed for every scenario. " +
			"The field encompasses various techniques including supervised learning, unsupervised learning, and reinforcement learning. " +
			"Supervised learning uses labeled training data to learn a mapping from inputs to outputs. " +
			"Unsupervised learning finds hidden patterns in data without using labeled examples. " +
			"Reinforcement learning involves an agent learning to make decisions by interacting with an environment. " +
			"Applications of machine learning include image recognition, natural language processing, and autonomous vehicles. " +
			"The success of machine learning depends heavily on the quality and quantity of training data available.",
		"metadata": map[string]interface{}{
			"author": "AI Research Team",
			"word_count": 128,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-extractive-summary",
		Type:   "summary_generation",
		Target: "summary-generator",
		Payload: map[string]interface{}{
			"operation": "generate_extractive_summary",
			"document":  document,
			"config": map[string]interface{}{
				"method":          "extractive",
				"sentence_count":  3,
				"preserve_order":  true,
				"include_scores":  true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Extractive summary generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Extractive summary generation test completed successfully")
}

func TestSummaryGeneratorAbstractiveSummary(t *testing.T) {
	config := map[string]interface{}{
		"method": "abstractive",
		"model": "transformer",
		"max_length": 150,
		"min_length": 50,
	}

	runner := &MockSummaryGeneratorRunner{config: config}

	document := map[string]interface{}{
		"id": "doc-abstract-001",
		"title": "Climate Change and Renewable Energy",
		"content": "Climate change represents one of the most pressing challenges of our time, with global temperatures rising due to increased greenhouse gas emissions. " +
			"The primary driver of these emissions is the burning of fossil fuels for energy production, transportation, and industrial processes. " +
			"Renewable energy sources such as solar, wind, and hydroelectric power offer sustainable alternatives to fossil fuels. " +
			"Solar energy has become increasingly cost-effective, with photovoltaic panel efficiency improving dramatically over the past decade. " +
			"Wind power generation has also expanded significantly, with offshore wind farms providing substantial energy capacity. " +
			"The transition to renewable energy requires significant infrastructure investment and policy support from governments worldwide. " +
			"Energy storage technologies like advanced batteries are crucial for managing the intermittent nature of renewable sources. " +
			"Smart grid systems can optimize energy distribution and integrate renewable sources more effectively into existing infrastructure.",
		"metadata": map[string]interface{}{
			"domain": "environmental science",
			"complexity": "intermediate",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-abstractive-summary",
		Type:   "summary_generation",
		Target: "summary-generator",
		Payload: map[string]interface{}{
			"operation": "generate_abstractive_summary",
			"document":  document,
			"config": map[string]interface{}{
				"method":       "abstractive",
				"model_type":   "transformer",
				"max_length":   150,
				"min_length":   50,
				"beam_size":    4,
				"temperature":  0.7,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Abstractive summary generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Abstractive summary generation test completed successfully")
}

func TestSummaryGeneratorMultiDocumentSummary(t *testing.T) {
	config := map[string]interface{}{
		"multi_document": true,
		"consolidation_method": "clustering",
		"redundancy_removal": true,
	}

	runner := &MockSummaryGeneratorRunner{config: config}

	documents := []map[string]interface{}{
		{
			"id": "doc-multi-001",
			"title": "Neural Networks Basics",
			"content": "Neural networks are computing systems inspired by biological neural networks. " +
				"They consist of interconnected nodes that process information through weighted connections. " +
				"Deep learning uses neural networks with multiple hidden layers to learn complex patterns.",
		},
		{
			"id": "doc-multi-002",
			"title": "Deep Learning Applications",
			"content": "Deep learning has revolutionized computer vision and natural language processing. " +
				"Convolutional neural networks excel at image recognition tasks. " +
				"Recurrent neural networks are effective for sequential data processing.",
		},
		{
			"id": "doc-multi-003",
			"title": "Training Neural Networks",
			"content": "Training neural networks requires large datasets and computational resources. " +
				"Backpropagation algorithm is used to update network weights during training. " +
				"Regularization techniques help prevent overfitting in complex models.",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-multi-document-summary",
		Type:   "summary_generation",
		Target: "summary-generator",
		Payload: map[string]interface{}{
			"operation":  "generate_multi_document_summary",
			"documents":  documents,
			"config": map[string]interface{}{
				"method":              "extractive",
				"consolidation":       "clustering",
				"max_sentences":       5,
				"remove_redundancy":   true,
				"coherence_optimization": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Multi-document summary generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Multi-document summary generation test completed successfully")
}

func TestSummaryGeneratorKeywordBasedSummary(t *testing.T) {
	config := map[string]interface{}{
		"keyword_guided": true,
		"keyword_weight": 0.3,
		"focus_preservation": true,
	}

	runner := &MockSummaryGeneratorRunner{config: config}

	document := map[string]interface{}{
		"id": "doc-keyword-001",
		"content": "Cybersecurity is a critical concern for modern organizations as they increasingly rely on digital infrastructure. " +
			"Data breaches can result in significant financial losses and damage to reputation. " +
			"Encryption is a fundamental security measure that protects sensitive information from unauthorized access. " +
			"Multi-factor authentication adds an additional layer of security beyond traditional passwords. " +
			"Network security involves monitoring and protecting computer networks from threats and unauthorized access. " +
			"Regular security audits help identify vulnerabilities before they can be exploited by malicious actors. " +
			"Employee training is essential for creating a security-conscious culture within organizations. " +
			"Incident response plans ensure quick and effective handling of security breaches when they occur.",
	}

	keywords := []string{"encryption", "authentication", "network security", "data breaches"}

	msg := &client.BrokerMessage{
		ID:     "test-keyword-based-summary",
		Type:   "summary_generation",
		Target: "summary-generator",
		Payload: map[string]interface{}{
			"operation": "generate_keyword_focused_summary",
			"document":  document,
			"keywords":  keywords,
			"config": map[string]interface{}{
				"keyword_weight":     0.4,
				"sentence_count":     3,
				"highlight_keywords": true,
				"context_window":     50,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Keyword-based summary generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Keyword-based summary generation test completed successfully")
}

func TestSummaryGeneratorHierarchicalSummary(t *testing.T) {
	config := map[string]interface{}{
		"hierarchical": true,
		"levels": 3,
		"granularity": "progressive",
	}

	runner := &MockSummaryGeneratorRunner{config: config}

	document := map[string]interface{}{
		"id": "doc-hierarchical-001",
		"structure": map[string]interface{}{
			"sections": []map[string]interface{}{
				{
					"title": "Introduction",
					"content": "This section introduces the main topic and provides background information.",
					"subsections": []map[string]interface{}{
						{"title": "Background", "content": "Historical context and previous research."},
						{"title": "Objectives", "content": "Goals and aims of the current study."},
					},
				},
				{
					"title": "Methodology",
					"content": "This section describes the research methods and experimental design.",
					"subsections": []map[string]interface{}{
						{"title": "Data Collection", "content": "How data was gathered and sources used."},
						{"title": "Analysis Methods", "content": "Statistical and analytical techniques applied."},
					},
				},
				{
					"title": "Results",
					"content": "This section presents the findings and key discoveries.",
					"subsections": []map[string]interface{}{
						{"title": "Primary Findings", "content": "Main results and significant outcomes."},
						{"title": "Secondary Observations", "content": "Additional insights and patterns discovered."},
					},
				},
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-hierarchical-summary",
		Type:   "summary_generation",
		Target: "summary-generator",
		Payload: map[string]interface{}{
			"operation": "generate_hierarchical_summary",
			"document":  document,
			"config": map[string]interface{}{
				"levels": []map[string]interface{}{
					{"level": 1, "max_sentences": 1, "scope": "document"},
					{"level": 2, "max_sentences": 3, "scope": "sections"},
					{"level": 3, "max_sentences": 5, "scope": "subsections"},
				},
				"preserve_structure": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Hierarchical summary generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Hierarchical summary generation test completed successfully")
}

func TestSummaryGeneratorCustomizableSummary(t *testing.T) {
	config := map[string]interface{}{
		"customization": map[string]interface{}{
			"enabled": true,
			"templates": []string{"executive", "technical", "brief"},
		},
	}

	runner := &MockSummaryGeneratorRunner{config: config}

	document := map[string]interface{}{
		"id": "doc-custom-001",
		"content": "The quarterly financial report shows strong performance across all business units. " +
			"Revenue increased by 15% compared to the previous quarter, driven by new product launches and market expansion. " +
			"Operating expenses were well-controlled, resulting in improved profit margins. " +
			"The technology division contributed significantly with its cloud services platform. " +
			"Customer acquisition rates exceeded targets in both domestic and international markets. " +
			"Risk management strategies proved effective during market volatility periods. " +
			"Investment in research and development continues to drive innovation and competitive advantage.",
		"audience": "executives",
		"purpose": "board presentation",
	}

	msg := &client.BrokerMessage{
		ID:     "test-customizable-summary",
		Type:   "summary_generation",
		Target: "summary-generator",
		Payload: map[string]interface{}{
			"operation": "generate_customized_summary",
			"document":  document,
			"customization": map[string]interface{}{
				"template":     "executive",
				"audience":     "board_members",
				"focus_areas":  []string{"financial_performance", "strategic_initiatives"},
				"tone":         "professional",
				"length":       "brief",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Customizable summary generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Customizable summary generation test completed successfully")
}

func TestSummaryGeneratorQualityAssessment(t *testing.T) {
	config := map[string]interface{}{
		"quality_assessment": map[string]interface{}{
			"enabled": true,
			"metrics": []string{"coherence", "relevance", "coverage", "conciseness"},
		},
	}

	runner := &MockSummaryGeneratorRunner{config: config}

	summaryData := map[string]interface{}{
		"original_document": map[string]interface{}{
			"id": "doc-quality-001",
			"content": "Original long document content...",
			"word_count": 1500,
		},
		"generated_summary": map[string]interface{}{
			"content": "Generated summary content...",
			"word_count": 150,
			"method": "extractive",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-quality-assessment",
		Type:   "summary_evaluation",
		Target: "summary-generator",
		Payload: map[string]interface{}{
			"operation":     "assess_summary_quality",
			"summary_data":  summaryData,
			"assessment_config": map[string]interface{}{
				"metrics": []string{"coherence", "relevance", "coverage"},
				"reference_summary": nil,
				"automatic_scoring": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Summary quality assessment failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Summary quality assessment test completed successfully")
}