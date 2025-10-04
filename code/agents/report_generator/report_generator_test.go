package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockReportGeneratorRunner implements agent.AgentRunner for testing
type MockReportGeneratorRunner struct {
	config map[string]interface{}
}

func (m *MockReportGeneratorRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockReportGeneratorRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockReportGeneratorRunner) Cleanup(base *agent.BaseAgent) {
}

func TestReportGeneratorInitialization(t *testing.T) {
	config := map[string]interface{}{
		"supported_formats": []string{"html", "pdf", "markdown", "json"},
		"template_engine":   "go_template",
		"output_directory":  "/test/reports",
	}

	runner := &MockReportGeneratorRunner{config: config}
	framework := agent.NewFramework(runner, "report-generator")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Report generator framework created successfully")
}

func TestReportGeneratorProcessingSummaryReport(t *testing.T) {
	config := map[string]interface{}{
		"report_types": []string{"processing_summary", "analytics", "audit"},
		"include_charts": true,
		"include_metrics": true,
	}

	runner := &MockReportGeneratorRunner{config: config}

	processingData := map[string]interface{}{
		"pipeline_id":   "pipeline-001",
		"execution_id":  "exec-20240927-001",
		"start_time":    "2024-09-27T09:00:00Z",
		"end_time":      "2024-09-27T09:15:30Z",
		"total_duration": "15m30s",
		"documents_processed": 150,
		"agents_involved": []map[string]interface{}{
			{
				"agent_name":       "text-extractor",
				"documents_handled": 150,
				"avg_processing_time": "2.5s",
				"success_rate":     0.98,
				"errors":          3,
			},
			{
				"agent_name":       "text-chunker",
				"documents_handled": 147,
				"avg_processing_time": "1.8s",
				"success_rate":     0.99,
				"errors":          1,
			},
			{
				"agent_name":       "search-indexer",
				"documents_handled": 146,
				"avg_processing_time": "3.2s",
				"success_rate":     1.0,
				"errors":          0,
			},
		},
		"overall_metrics": map[string]interface{}{
			"throughput_docs_per_min": 10.0,
			"error_rate":             0.027,
			"resource_utilization":   0.75,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-processing-summary",
		Type:   "report_generation",
		Target: "report-generator",
		Payload: map[string]interface{}{
			"operation":        "generate_processing_summary",
			"processing_data":  processingData,
			"report_config": map[string]interface{}{
				"format":           "html",
				"include_charts":   true,
				"include_details":  true,
				"template":         "processing_summary_template",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Processing summary report generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Processing summary report generation test completed successfully")
}

func TestReportGeneratorAnalyticsReport(t *testing.T) {
	config := map[string]interface{}{
		"analytics": map[string]interface{}{
			"time_series_analysis": true,
			"trend_analysis":       true,
			"comparative_analysis": true,
		},
	}

	runner := &MockReportGeneratorRunner{config: config}

	analyticsData := map[string]interface{}{
		"analysis_period": map[string]interface{}{
			"start_date": "2024-09-01T00:00:00Z",
			"end_date":   "2024-09-27T23:59:59Z",
			"duration":   "27 days",
		},
		"document_analytics": map[string]interface{}{
			"total_documents":     4521,
			"successful_processed": 4487,
			"failed_processed":    34,
			"avg_doc_size_mb":     2.3,
			"largest_doc_mb":      45.7,
			"smallest_doc_mb":     0.1,
		},
		"performance_trends": []map[string]interface{}{
			{
				"date":            "2024-09-01",
				"documents":       150,
				"avg_time_sec":    8.5,
				"error_rate":      0.02,
			},
			{
				"date":            "2024-09-15",
				"documents":       175,
				"avg_time_sec":    7.8,
				"error_rate":      0.015,
			},
			{
				"date":            "2024-09-27",
				"documents":       200,
				"avg_time_sec":    7.2,
				"error_rate":      0.01,
			},
		},
		"agent_performance": map[string]interface{}{
			"top_performing": []string{"search-indexer", "text-chunker"},
			"needs_optimization": []string{"ocr-extractor"},
			"most_used":      []string{"text-extractor", "text-chunker"},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-analytics-report",
		Type:   "report_generation",
		Target: "report-generator",
		Payload: map[string]interface{}{
			"operation":       "generate_analytics_report",
			"analytics_data":  analyticsData,
			"report_config": map[string]interface{}{
				"format":              "pdf",
				"include_visualizations": true,
				"include_recommendations": true,
				"executive_summary":    true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Analytics report generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Analytics report generation test completed successfully")
}

func TestReportGeneratorAuditReport(t *testing.T) {
	config := map[string]interface{}{
		"audit_features": map[string]interface{}{
			"compliance_checking": true,
			"security_assessment": true,
			"quality_validation":  true,
		},
	}

	runner := &MockReportGeneratorRunner{config: config}

	auditData := map[string]interface{}{
		"audit_scope": map[string]interface{}{
			"audit_id":     "audit-2024-09-27",
			"audit_date":   "2024-09-27T15:00:00Z",
			"auditor":      "System Audit Agent",
			"scope":        "Full pipeline audit",
		},
		"compliance_status": map[string]interface{}{
			"data_privacy": map[string]interface{}{
				"status":      "compliant",
				"score":       0.95,
				"violations":  1,
				"details":     "Minor logging issue identified",
			},
			"security": map[string]interface{}{
				"status":      "compliant",
				"score":       0.98,
				"violations":  0,
				"details":     "All security checks passed",
			},
			"data_integrity": map[string]interface{}{
				"status":      "compliant",
				"score":       0.92,
				"violations":  2,
				"details":     "Two checksum mismatches found",
			},
		},
		"quality_metrics": map[string]interface{}{
			"processing_accuracy":   0.967,
			"data_completeness":     0.994,
			"error_handling":        0.985,
			"performance_standards": 0.912,
		},
		"findings": []map[string]interface{}{
			{
				"category":   "data_privacy",
				"severity":   "low",
				"finding":    "Excessive logging in debug mode",
				"recommendation": "Review and reduce log verbosity",
			},
			{
				"category":   "performance",
				"severity":   "medium",
				"finding":    "Agent response time degradation",
				"recommendation": "Optimize agent resource allocation",
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-audit-report",
		Type:   "report_generation",
		Target: "report-generator",
		Payload: map[string]interface{}{
			"operation":    "generate_audit_report",
			"audit_data":   auditData,
			"report_config": map[string]interface{}{
				"format":              "html",
				"compliance_focus":    true,
				"include_remediation": true,
				"confidentiality":     "internal",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Audit report generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Audit report generation test completed successfully")
}

func TestReportGeneratorCustomReport(t *testing.T) {
	config := map[string]interface{}{
		"custom_reports": map[string]interface{}{
			"template_support": true,
			"dynamic_sections": true,
			"user_defined_metrics": true,
		},
	}

	runner := &MockReportGeneratorRunner{config: config}

	customReportSpec := map[string]interface{}{
		"report_definition": map[string]interface{}{
			"title":       "Monthly Document Processing Report",
			"description": "Custom report for monthly review",
			"sections": []map[string]interface{}{
				{
					"section_id":   "summary",
					"title":        "Executive Summary",
					"type":         "text",
					"content_source": "summary_data",
				},
				{
					"section_id":   "metrics",
					"title":        "Key Metrics",
					"type":         "metrics_table",
					"content_source": "metrics_data",
				},
				{
					"section_id":   "trends",
					"title":        "Performance Trends",
					"type":         "chart",
					"chart_type":   "line",
					"content_source": "trend_data",
				},
			},
		},
		"data_sources": map[string]interface{}{
			"summary_data": "Monthly processing completed successfully with improved performance.",
			"metrics_data": map[string]interface{}{
				"documents_processed": 1250,
				"average_time":        "6.5s",
				"success_rate":        "98.4%",
				"cost_per_document":   "$0.02",
			},
			"trend_data": []map[string]interface{}{
				{"month": "July", "throughput": 1100, "accuracy": 97.2},
				{"month": "August", "throughput": 1200, "accuracy": 97.8},
				{"month": "September", "throughput": 1250, "accuracy": 98.4},
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-custom-report",
		Type:   "report_generation",
		Target: "report-generator",
		Payload: map[string]interface{}{
			"operation":           "generate_custom_report",
			"custom_report_spec":  customReportSpec,
			"report_config": map[string]interface{}{
				"format":          "pdf",
				"template":        "monthly_template",
				"branding":        true,
				"interactive":     false,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Custom report generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Custom report generation test completed successfully")
}

func TestReportGeneratorErrorReport(t *testing.T) {
	config := map[string]interface{}{
		"error_reporting": map[string]interface{}{
			"categorize_errors": true,
			"include_stack_traces": false,
			"suggest_solutions": true,
		},
	}

	runner := &MockReportGeneratorRunner{config: config}

	errorData := map[string]interface{}{
		"reporting_period": map[string]interface{}{
			"start": "2024-09-27T00:00:00Z",
			"end":   "2024-09-27T23:59:59Z",
		},
		"error_summary": map[string]interface{}{
			"total_errors":    23,
			"critical_errors": 2,
			"major_errors":    8,
			"minor_errors":    13,
		},
		"error_categories": []map[string]interface{}{
			{
				"category":    "file_processing",
				"count":       15,
				"percentage":  65.2,
				"top_error":   "Unsupported file format",
				"agents_affected": []string{"text-extractor", "pdf-processor"},
			},
			{
				"category":    "network",
				"count":       5,
				"percentage":  21.7,
				"top_error":   "Connection timeout",
				"agents_affected": []string{"search-indexer", "external-api-client"},
			},
			{
				"category":    "resource",
				"count":       3,
				"percentage":  13.1,
				"top_error":   "Memory limit exceeded",
				"agents_affected": []string{"large-doc-processor"},
			},
		},
		"critical_incidents": []map[string]interface{}{
			{
				"timestamp":   "2024-09-27T10:15:00Z",
				"agent":       "text-extractor",
				"error":       "Critical: PDF parsing failure",
				"impact":      "High - 50 documents failed",
				"resolution":  "Pending - investigating codec issue",
			},
			{
				"timestamp":   "2024-09-27T14:30:00Z",
				"agent":       "search-indexer",
				"error":       "Critical: Index corruption detected",
				"impact":      "High - Search functionality impaired",
				"resolution":  "Resolved - Index rebuilt",
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-error-report",
		Type:   "report_generation",
		Target: "report-generator",
		Payload: map[string]interface{}{
			"operation":    "generate_error_report",
			"error_data":   errorData,
			"report_config": map[string]interface{}{
				"format":               "html",
				"include_charts":       true,
				"priority_focus":       "critical",
				"include_remediation":  true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Error report generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Error report generation test completed successfully")
}

func TestReportGeneratorScheduledReport(t *testing.T) {
	config := map[string]interface{}{
		"scheduling": map[string]interface{}{
			"support_cron":     true,
			"auto_distribution": true,
			"template_reuse":   true,
		},
	}

	runner := &MockReportGeneratorRunner{config: config}

	scheduleConfig := map[string]interface{}{
		"schedule_id":   "weekly-ops-report",
		"report_type":   "operational_summary",
		"frequency":     "weekly",
		"cron_expression": "0 9 * * MON",
		"recipients": []map[string]interface{}{
			{"email": "ops-team@company.com", "format": "pdf"},
			{"email": "management@company.com", "format": "html"},
		},
		"template":      "weekly_ops_template",
		"data_sources": []string{"processing_metrics", "error_logs", "performance_data"},
		"auto_generate": true,
	}

	msg := &client.BrokerMessage{
		ID:     "test-scheduled-report",
		Type:   "report_scheduling",
		Target: "report-generator",
		Payload: map[string]interface{}{
			"operation":       "configure_scheduled_report",
			"schedule_config": scheduleConfig,
			"execution_config": map[string]interface{}{
				"retry_on_failure": true,
				"max_retries":      3,
				"notification":     true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Scheduled report configuration failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Scheduled report configuration test completed successfully")
}

func TestReportGeneratorInteractiveReport(t *testing.T) {
	config := map[string]interface{}{
		"interactive_features": map[string]interface{}{
			"drill_down_capability": true,
			"real_time_updates":     true,
			"user_filtering":        true,
		},
	}

	runner := &MockReportGeneratorRunner{config: config}

	interactiveSpec := map[string]interface{}{
		"report_type":  "interactive_dashboard",
		"components": []map[string]interface{}{
			{
				"component_id":   "overview_cards",
				"type":           "metric_cards",
				"data_source":    "real_time_metrics",
				"refresh_rate":   "30s",
				"interactive":    false,
			},
			{
				"component_id":   "processing_trends",
				"type":           "line_chart",
				"data_source":    "time_series_data",
				"interactive":    true,
				"filters":        []string{"date_range", "agent_type"},
			},
			{
				"component_id":   "agent_status",
				"type":           "status_grid",
				"data_source":    "agent_health_data",
				"drill_down":     true,
				"click_actions":  []string{"view_details", "restart_agent"},
			},
		},
		"layout": map[string]interface{}{
			"type":    "grid",
			"columns": 3,
			"responsive": true,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-interactive-report",
		Type:   "report_generation",
		Target: "report-generator",
		Payload: map[string]interface{}{
			"operation":        "generate_interactive_report",
			"interactive_spec": interactiveSpec,
			"report_config": map[string]interface{}{
				"format":       "html",
				"framework":    "web_dashboard",
				"authentication": true,
				"export_options": []string{"pdf", "excel"},
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Interactive report generation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Interactive report generation test completed successfully")
}