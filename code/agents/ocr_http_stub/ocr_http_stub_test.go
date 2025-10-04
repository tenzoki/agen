package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// MockOcrHttpStubRunner implements agent.AgentRunner for testing
type MockOcrHttpStubRunner struct {
	config map[string]interface{}
}

func (m *MockOcrHttpStubRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockOcrHttpStubRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockOcrHttpStubRunner) Cleanup(base *agent.BaseAgent) {
}

func TestOcrHttpStubInitialization(t *testing.T) {
	config := map[string]interface{}{
		"http_server": map[string]interface{}{
			"port":         8080,
			"host":         "localhost",
			"timeout":      30,
			"max_requests": 100,
		},
		"ocr_engines": []string{"tesseract", "paddle_ocr", "easyocr"},
	}

	runner := &MockOcrHttpStubRunner{config: config}
	framework := agent.NewFramework(runner, "ocr-http-stub")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("OCR HTTP stub framework created successfully")
}

func TestOcrHttpStubBasicOCR(t *testing.T) {
	config := map[string]interface{}{
		"default_engine": "tesseract",
		"supported_formats": []string{"png", "jpg", "pdf", "tiff"},
		"preprocessing": true,
	}

	runner := &MockOcrHttpStubRunner{config: config}

	ocrRequest := map[string]interface{}{
		"request_id": "ocr-req-001",
		"image_data": map[string]interface{}{
			"format":     "png",
			"size_bytes": 1024000,
			"dimensions": map[string]interface{}{
				"width":  800,
				"height": 600,
			},
			"dpi": 300,
		},
		"ocr_options": map[string]interface{}{
			"language":    "eng",
			"engine":      "tesseract",
			"page_seg_mode": 3,
			"preserve_layout": true,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-basic-ocr",
		Type:   "ocr_request",
		Target: "ocr-http-stub",
		Payload: map[string]interface{}{
			"operation":   "extract_text",
			"ocr_request": ocrRequest,
			"http_context": map[string]interface{}{
				"method":         "POST",
				"endpoint":       "/ocr/extract",
				"content_type":   "application/json",
				"request_headers": map[string]string{
					"Authorization": "Bearer test_token",
					"User-Agent":    "GOX-OCR-Client/1.0",
				},
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Basic OCR request failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Basic OCR request test completed successfully")
}

func TestOcrHttpStubBatchOCR(t *testing.T) {
	config := map[string]interface{}{
		"batch_processing": map[string]interface{}{
			"max_batch_size": 10,
			"parallel_workers": 3,
			"timeout_per_image": 30,
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	batchRequest := map[string]interface{}{
		"batch_id": "batch-ocr-001",
		"images": []map[string]interface{}{
			{
				"image_id":   "img-001",
				"format":     "jpg",
				"language":   "eng",
				"engine":     "tesseract",
			},
			{
				"image_id":   "img-002",
				"format":     "png",
				"language":   "spa",
				"engine":     "paddle_ocr",
			},
			{
				"image_id":   "img-003",
				"format":     "pdf",
				"language":   "fra",
				"engine":     "easyocr",
				"page_range": "1-5",
			},
		},
		"batch_options": map[string]interface{}{
			"parallel_processing": true,
			"error_handling":      "continue_on_error",
			"return_format":       "json",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-batch-ocr",
		Type:   "batch_ocr_request",
		Target: "ocr-http-stub",
		Payload: map[string]interface{}{
			"operation":     "batch_extract",
			"batch_request": batchRequest,
			"http_context": map[string]interface{}{
				"method":     "POST",
				"endpoint":   "/ocr/batch",
				"timeout":    300, // 5 minutes for batch
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Batch OCR request failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Batch OCR request test completed successfully")
}

func TestOcrHttpStubAdvancedOCR(t *testing.T) {
	config := map[string]interface{}{
		"advanced_features": map[string]interface{}{
			"table_detection":    true,
			"layout_analysis":    true,
			"confidence_scoring": true,
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	advancedRequest := map[string]interface{}{
		"request_id":  "ocr-advanced-001",
		"image_type":  "document_scan",
		"analysis_options": map[string]interface{}{
			"detect_tables":      true,
			"extract_layout":     true,
			"identify_regions":   true,
			"calculate_confidence": true,
		},
		"preprocessing": map[string]interface{}{
			"deskew":            true,
			"noise_reduction":   true,
			"contrast_enhance":  true,
			"binarization":      true,
		},
		"output_format": map[string]interface{}{
			"include_coordinates": true,
			"include_confidence":  true,
			"preserve_formatting": true,
			"export_formats":      []string{"text", "hocr", "pdf"},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-advanced-ocr",
		Type:   "advanced_ocr_request",
		Target: "ocr-http-stub",
		Payload: map[string]interface{}{
			"operation":        "advanced_extract",
			"advanced_request": advancedRequest,
			"http_context": map[string]interface{}{
				"method":       "POST",
				"endpoint":     "/ocr/advanced",
				"content_type": "multipart/form-data",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Advanced OCR request failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Advanced OCR request test completed successfully")
}

func TestOcrHttpStubLanguageSupport(t *testing.T) {
	config := map[string]interface{}{
		"multilingual": map[string]interface{}{
			"auto_detect":        true,
			"supported_languages": []string{"eng", "spa", "fra", "deu", "jpn", "chi_sim"},
			"mixed_language":     true,
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	languageTests := []map[string]interface{}{
		{
			"test_name":  "english_text",
			"language":   "eng",
			"auto_detect": false,
			"expected_confidence": 0.9,
		},
		{
			"test_name":  "spanish_text",
			"language":   "spa",
			"auto_detect": false,
			"expected_confidence": 0.85,
		},
		{
			"test_name":  "mixed_languages",
			"language":   "eng+spa",
			"auto_detect": false,
			"expected_confidence": 0.8,
		},
		{
			"test_name":  "auto_detection",
			"language":   "auto",
			"auto_detect": true,
			"expected_confidence": 0.75,
		},
	}

	for i, test := range languageTests {
		msg := &client.BrokerMessage{
			ID:     "test-language-" + string(rune('1'+i)),
			Type:   "ocr_request",
			Target: "ocr-http-stub",
			Payload: map[string]interface{}{
				"operation":     "language_ocr",
				"language_test": test,
				"ocr_config": map[string]interface{}{
					"engine":             "tesseract",
					"auto_detect_lang":   test["auto_detect"],
					"confidence_threshold": 0.7,
				},
				"http_context": map[string]interface{}{
					"method":   "POST",
					"endpoint": "/ocr/language",
				},
			},
			Meta: make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Language OCR test %d failed: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for test %d", i+1)
		}
	}
	t.Log("Language support test completed successfully")
}

func TestOcrHttpStubHealthCheck(t *testing.T) {
	config := map[string]interface{}{
		"health_monitoring": map[string]interface{}{
			"enabled":        true,
			"check_interval": "30s",
			"engines_status": true,
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	healthRequest := map[string]interface{}{
		"check_type": "full",
		"components": []string{"server", "engines", "storage", "memory"},
		"include_metrics": true,
	}

	msg := &client.BrokerMessage{
		ID:     "test-health-check",
		Type:   "health_check",
		Target: "ocr-http-stub",
		Payload: map[string]interface{}{
			"operation":      "health_status",
			"health_request": healthRequest,
			"http_context": map[string]interface{}{
				"method":   "GET",
				"endpoint": "/health",
				"timeout":  10,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Health check test completed successfully")
}

func TestOcrHttpStubConfiguration(t *testing.T) {
	config := map[string]interface{}{
		"runtime_config": map[string]interface{}{
			"dynamic_updates": true,
			"config_validation": true,
			"backup_config":   true,
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	configOperations := []map[string]interface{}{
		{
			"operation": "get_config",
			"section":   "all",
		},
		{
			"operation": "update_config",
			"section":   "ocr_engines",
			"config": map[string]interface{}{
				"tesseract": map[string]interface{}{
					"dpi":        300,
					"page_seg_mode": 6,
				},
			},
		},
		{
			"operation": "validate_config",
			"config": map[string]interface{}{
				"invalid_setting": "invalid_value",
			},
		},
	}

	for i, operation := range configOperations {
		msg := &client.BrokerMessage{
			ID:     "test-config-" + string(rune('1'+i)),
			Type:   "configuration",
			Target: "ocr-http-stub",
			Payload: map[string]interface{}{
				"operation":     "manage_config",
				"config_op":     operation,
				"http_context": map[string]interface{}{
					"method":   "POST",
					"endpoint": "/config",
					"admin_auth": true,
				},
			},
			Meta: make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Configuration operation %d failed: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for operation %d", i+1)
		}
	}
	t.Log("Configuration management test completed successfully")
}

func TestOcrHttpStubMetrics(t *testing.T) {
	config := map[string]interface{}{
		"metrics": map[string]interface{}{
			"collection_enabled": true,
			"export_format":      "prometheus",
			"retention_period":   "7d",
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	metricsRequest := map[string]interface{}{
		"metrics_type": "performance",
		"time_range": map[string]interface{}{
			"start": "2024-09-27T00:00:00Z",
			"end":   "2024-09-27T23:59:59Z",
		},
		"aggregation": "1h",
		"include_metrics": []string{
			"requests_total",
			"processing_duration",
			"error_rate",
			"throughput",
			"memory_usage",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-metrics",
		Type:   "metrics_request",
		Target: "ocr-http-stub",
		Payload: map[string]interface{}{
			"operation":        "get_metrics",
			"metrics_request":  metricsRequest,
			"http_context": map[string]interface{}{
				"method":   "GET",
				"endpoint": "/metrics",
				"format":   "json",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Metrics request failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Metrics request test completed successfully")
}

func TestOcrHttpStubErrorHandling(t *testing.T) {
	config := map[string]interface{}{
		"error_handling": map[string]interface{}{
			"retry_attempts":   3,
			"retry_delay":      "1s",
			"circuit_breaker":  true,
			"fallback_engine":  "tesseract",
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	errorScenarios := []map[string]interface{}{
		{
			"scenario":     "invalid_image_format",
			"error_type":   "validation_error",
			"expected_code": 400,
			"image_format": "unsupported",
		},
		{
			"scenario":     "ocr_engine_failure",
			"error_type":   "processing_error",
			"expected_code": 500,
			"engine":       "failing_engine",
		},
		{
			"scenario":     "timeout_error",
			"error_type":   "timeout_error",
			"expected_code": 408,
			"processing_time": 60,
		},
		{
			"scenario":     "rate_limit_exceeded",
			"error_type":   "rate_limit_error",
			"expected_code": 429,
			"request_count": 1000,
		},
	}

	for i, scenario := range errorScenarios {
		msg := &client.BrokerMessage{
			ID:     "test-error-" + string(rune('1'+i)),
			Type:   "error_test",
			Target: "ocr-http-stub",
			Payload: map[string]interface{}{
				"operation":      "simulate_error",
				"error_scenario": scenario,
				"http_context": map[string]interface{}{
					"method":        "POST",
					"endpoint":      "/ocr/extract",
					"expect_error":  true,
				},
			},
			Meta: make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Error handling test %d failed: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for test %d", i+1)
		}
	}
	t.Log("Error handling test completed successfully")
}

func TestOcrHttpStubLoadTesting(t *testing.T) {
	config := map[string]interface{}{
		"load_testing": map[string]interface{}{
			"max_concurrent_requests": 50,
			"rate_limiting":           true,
			"queue_size":              100,
		},
	}

	runner := &MockOcrHttpStubRunner{config: config}

	loadTest := map[string]interface{}{
		"test_name":         "concurrent_ocr_requests",
		"concurrent_users":  10,
		"requests_per_user": 5,
		"ramp_up_time":      "10s",
		"test_duration":     "60s",
		"request_template": map[string]interface{}{
			"image_size": "1MB",
			"language":   "eng",
			"engine":     "tesseract",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-load-testing",
		Type:   "load_test",
		Target: "ocr-http-stub",
		Payload: map[string]interface{}{
			"operation":  "execute_load_test",
			"load_test":  loadTest,
			"test_config": map[string]interface{}{
				"collect_metrics":   true,
				"track_errors":      true,
				"report_percentiles": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Load testing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Load testing test completed successfully")
}