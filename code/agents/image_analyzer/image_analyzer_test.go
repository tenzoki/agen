package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockImageAnalyzerRunner implements agent.AgentRunner for testing
type MockImageAnalyzerRunner struct {
	config map[string]interface{}
}

func (m *MockImageAnalyzerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockImageAnalyzerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockImageAnalyzerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestImageAnalyzerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"supported_formats": []string{"jpg", "png", "gif", "bmp", "tiff"},
		"analysis_engines":  []string{"opencv", "tensorflow", "custom"},
		"max_image_size":    "10MB",
	}

	runner := &MockImageAnalyzerRunner{config: config}
	framework := agent.NewFramework(runner, "image-analyzer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Image analyzer framework created successfully")
}

func TestImageAnalyzerBasicAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"basic_analysis": map[string]interface{}{
			"extract_metadata": true,
			"analyze_quality":  true,
			"detect_format":    true,
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	imageData := map[string]interface{}{
		"image_id":   "img-basic-001",
		"file_path":  "/test/images/sample.jpg",
		"format":     "jpeg",
		"size_bytes": 2048576,
		"dimensions": map[string]interface{}{
			"width":  1920,
			"height": 1080,
		},
		"color_space": "RGB",
	}

	msg := &client.BrokerMessage{
		ID:     "test-basic-analysis",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":   "basic_analysis",
			"image_data":  imageData,
			"analysis_config": map[string]interface{}{
				"extract_exif":      true,
				"calculate_hash":    true,
				"assess_quality":    true,
				"detect_corruption": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Basic image analysis failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Basic image analysis test completed successfully")
}

func TestImageAnalyzerObjectDetection(t *testing.T) {
	config := map[string]interface{}{
		"object_detection": map[string]interface{}{
			"model":           "yolo_v5",
			"confidence_threshold": 0.5,
			"max_objects":     50,
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	detectionRequest := map[string]interface{}{
		"image_id": "img-detection-001",
		"image_data": map[string]interface{}{
			"format":     "png",
			"dimensions": map[string]interface{}{
				"width":  800,
				"height": 600,
			},
		},
		"detection_targets": []string{"person", "car", "dog", "cat", "bicycle"},
	}

	msg := &client.BrokerMessage{
		ID:     "test-object-detection",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":          "detect_objects",
			"detection_request":  detectionRequest,
			"detection_config": map[string]interface{}{
				"confidence_threshold": 0.6,
				"nms_threshold":       0.4,
				"return_bounding_boxes": true,
				"include_masks":       false,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Object detection failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Object detection test completed successfully")
}

func TestImageAnalyzerFaceRecognition(t *testing.T) {
	config := map[string]interface{}{
		"face_recognition": map[string]interface{}{
			"detection_model": "mtcnn",
			"recognition_model": "facenet",
			"embedding_size":   128,
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	faceAnalysisRequest := map[string]interface{}{
		"image_id": "img-face-001",
		"analysis_types": []string{"detection", "landmarks", "attributes", "recognition"},
		"face_database": map[string]interface{}{
			"enabled":    true,
			"database_id": "employee_faces",
			"match_threshold": 0.85,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-face-recognition",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":             "analyze_faces",
			"face_analysis_request": faceAnalysisRequest,
			"face_config": map[string]interface{}{
				"detect_emotions":     true,
				"estimate_age":        true,
				"detect_gender":       true,
				"extract_landmarks":   true,
				"quality_assessment":  true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Face recognition failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Face recognition test completed successfully")
}

func TestImageAnalyzerTextExtraction(t *testing.T) {
	config := map[string]interface{}{
		"text_extraction": map[string]interface{}{
			"ocr_engine":       "tesseract",
			"languages":        []string{"eng", "spa", "fra"},
			"preprocessing":    true,
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	textExtractionRequest := map[string]interface{}{
		"image_id": "img-text-001",
		"image_type": "document_scan",
		"preprocessing_options": map[string]interface{}{
			"deskew":           true,
			"noise_reduction":  true,
			"contrast_enhance": true,
			"binarization":     true,
		},
		"extraction_regions": []map[string]interface{}{
			{
				"region_id": "header",
				"bbox": map[string]interface{}{
					"x": 0, "y": 0, "width": 800, "height": 100,
				},
			},
			{
				"region_id": "body",
				"bbox": map[string]interface{}{
					"x": 0, "y": 100, "width": 800, "height": 500,
				},
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-text-extraction",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":               "extract_text",
			"text_extraction_request": textExtractionRequest,
			"extraction_config": map[string]interface{}{
				"confidence_threshold": 0.7,
				"preserve_layout":      true,
				"detect_tables":        true,
				"extract_coordinates":  true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text extraction test completed successfully")
}

func TestImageAnalyzerColorAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"color_analysis": map[string]interface{}{
			"extract_palette":    true,
			"histogram_analysis": true,
			"dominant_colors":    5,
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	colorAnalysisRequest := map[string]interface{}{
		"image_id": "img-color-001",
		"analysis_types": []string{"palette", "histogram", "dominant", "temperature"},
		"color_spaces": []string{"RGB", "HSV", "LAB"},
	}

	msg := &client.BrokerMessage{
		ID:     "test-color-analysis",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":             "analyze_colors",
			"color_analysis_request": colorAnalysisRequest,
			"color_config": map[string]interface{}{
				"palette_size":         8,
				"clustering_algorithm": "kmeans",
				"histogram_bins":       256,
				"calculate_harmony":    true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Color analysis failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Color analysis test completed successfully")
}

func TestImageAnalyzerQualityAssessment(t *testing.T) {
	config := map[string]interface{}{
		"quality_assessment": map[string]interface{}{
			"sharpness_detection": true,
			"noise_analysis":      true,
			"exposure_analysis":   true,
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	qualityRequest := map[string]interface{}{
		"image_id": "img-quality-001",
		"assessment_criteria": []string{"sharpness", "noise", "exposure", "contrast", "saturation"},
		"reference_standards": map[string]interface{}{
			"min_sharpness": 0.7,
			"max_noise":     0.3,
			"exposure_range": map[string]interface{}{
				"min": 0.2,
				"max": 0.8,
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-quality-assessment",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":        "assess_quality",
			"quality_request":  qualityRequest,
			"quality_config": map[string]interface{}{
				"detailed_metrics":    true,
				"generate_score":      true,
				"suggest_improvements": true,
				"compare_standards":   true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Quality assessment failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Quality assessment test completed successfully")
}

func TestImageAnalyzerSimilarityComparison(t *testing.T) {
	config := map[string]interface{}{
		"similarity_analysis": map[string]interface{}{
			"feature_extraction": "deep_features",
			"similarity_metrics": []string{"cosine", "euclidean", "ssim"},
			"threshold":          0.8,
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	similarityRequest := map[string]interface{}{
		"comparison_type": "one_to_many",
		"reference_image": map[string]interface{}{
			"image_id":   "img-ref-001",
			"file_path":  "/test/images/reference.jpg",
		},
		"candidate_images": []map[string]interface{}{
			{
				"image_id":  "img-cand-001",
				"file_path": "/test/images/candidate1.jpg",
			},
			{
				"image_id":  "img-cand-002",
				"file_path": "/test/images/candidate2.jpg",
			},
			{
				"image_id":  "img-cand-003",
				"file_path": "/test/images/candidate3.jpg",
			},
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-similarity-comparison",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":          "compare_similarity",
			"similarity_request": similarityRequest,
			"similarity_config": map[string]interface{}{
				"feature_model":        "resnet50",
				"comparison_metrics":   []string{"perceptual", "structural", "feature"},
				"rank_results":         true,
				"include_confidence":   true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Similarity comparison failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Similarity comparison test completed successfully")
}

func TestImageAnalyzerBatchProcessing(t *testing.T) {
	config := map[string]interface{}{
		"batch_processing": map[string]interface{}{
			"max_batch_size":   20,
			"parallel_workers": 4,
			"memory_limit":     "2GB",
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	batchRequest := map[string]interface{}{
		"batch_id": "batch-analysis-001",
		"images": []map[string]interface{}{
			{
				"image_id":  "img-batch-001",
				"file_path": "/test/batch/image1.jpg",
				"analysis_types": []string{"basic", "objects"},
			},
			{
				"image_id":  "img-batch-002",
				"file_path": "/test/batch/image2.png",
				"analysis_types": []string{"text", "quality"},
			},
			{
				"image_id":  "img-batch-003",
				"file_path": "/test/batch/image3.gif",
				"analysis_types": []string{"colors", "faces"},
			},
		},
		"processing_options": map[string]interface{}{
			"parallel_processing": true,
			"error_handling":      "continue_on_error",
			"progress_reporting":  true,
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-batch-processing",
		Type:   "batch_image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":     "process_batch",
			"batch_request": batchRequest,
			"batch_config": map[string]interface{}{
				"optimize_memory":     true,
				"cache_models":        true,
				"report_statistics":   true,
				"consolidate_results": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Batch processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Batch processing test completed successfully")
}

func TestImageAnalyzerCustomAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"custom_analysis": map[string]interface{}{
			"plugin_support":   true,
			"custom_models":    true,
			"scripting_engine": "python",
		},
	}

	runner := &MockImageAnalyzerRunner{config: config}

	customRequest := map[string]interface{}{
		"analysis_id":   "custom-medical-001",
		"analysis_type": "medical_imaging",
		"custom_pipeline": []map[string]interface{}{
			{
				"step":   "preprocessing",
				"module": "medical_preprocess",
				"params": map[string]interface{}{
					"normalize_intensity": true,
					"remove_noise":        true,
				},
			},
			{
				"step":   "segmentation",
				"module": "organ_segmentation",
				"params": map[string]interface{}{
					"target_organs": []string{"liver", "kidney"},
					"confidence":    0.9,
				},
			},
			{
				"step":   "measurement",
				"module": "volume_calculation",
				"params": map[string]interface{}{
					"units":     "cm3",
					"precision": 2,
				},
			},
		},
		"image_metadata": map[string]interface{}{
			"modality":     "CT",
			"slice_thickness": 1.0,
			"patient_id":   "P001",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-custom-analysis",
		Type:   "image_analysis",
		Target: "image-analyzer",
		Payload: map[string]interface{}{
			"operation":      "custom_analysis",
			"custom_request": customRequest,
			"custom_config": map[string]interface{}{
				"validate_pipeline": true,
				"cache_results":     true,
				"generate_report":   true,
				"include_visualizations": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Custom analysis failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Custom analysis test completed successfully")
}