package chunks

import (
	"fmt"
	"sync"
	"time"
)

// GetProcessingMetrics provides real-time processing statistics
func (ct *ChunkTracker) GetProcessingMetrics(fileHash string) (*ChunkProcessingMetrics, error) {
	queries := []string{
		fmt.Sprintf(`g.V().has('id', 'file:original:%s').out('HAS_CHUNK').count()`, fileHash),
		fmt.Sprintf(`g.V().has('id', 'file:original:%s').out('HAS_CHUNK').has('status', 'completed').count()`, fileHash),
		fmt.Sprintf(`g.V().has('id', 'file:original:%s').out('HAS_CHUNK').has('status', 'failed').count()`, fileHash),
		fmt.Sprintf(`g.V().has('id', 'file:original:%s').out('HAS_CHUNK').has('status', 'processing').values('processed_by').dedup().count()`, fileHash),
	}

	results, err := ct.storageClient.ParallelGraphQuery(queries)
	if err != nil {
		return nil, err
	}

	metrics := &ChunkProcessingMetrics{
		TotalChunks:        int(results[0][0].(float64)),
		ProcessedChunks:    int(results[1][0].(float64)),
		FailedChunks:       int(results[2][0].(float64)),
		ParallelProcessors: int(results[3][0].(float64)),
	}

	// Calculate derived metrics
	if metrics.TotalChunks > 0 {
		metrics.BatchEfficiency = float64(metrics.ProcessedChunks) / float64(metrics.TotalChunks)
	}

	// Calculate average processing time
	avgTime, err := ct.calculateAverageProcessingTime(fileHash)
	if err == nil {
		metrics.AverageProcessingTime = avgTime
	}

	// Calculate throughput
	metrics.ThroughputPerSecond = ct.calculateThroughput(fileHash)

	return metrics, nil
}

// GetBatchProcessingMetrics provides metrics for multiple files processed in batch
func (ct *ChunkTracker) GetBatchProcessingMetrics(fileHashes []string) (map[string]*ChunkProcessingMetrics, error) {
	results := make(map[string]*ChunkProcessingMetrics)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make(chan error, len(fileHashes))

	for _, fileHash := range fileHashes {
		wg.Add(1)
		go func(hash string) {
			defer wg.Done()

			metrics, err := ct.GetProcessingMetrics(hash)
			if err != nil {
				errors <- err
				return
			}

			mu.Lock()
			results[hash] = metrics
			mu.Unlock()
		}(fileHash)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// calculateAverageProcessingTime calculates the average time to process a chunk
func (ct *ChunkTracker) calculateAverageProcessingTime(fileHash string) (time.Duration, error) {
	query := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.out('HAS_CHUNK')
		.has('status', 'completed')
		.out('HAS_EVENT')
		.has('to_state', 'completed')
		.values('timestamp')
		.limit(20)
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil || len(results) < 2 {
		return 0, fmt.Errorf("insufficient data for average calculation")
	}

	// Find corresponding start events for each completion
	var durations []time.Duration

	for _, result := range results {
		_, err := time.Parse(time.RFC3339, result.(string))
		if err != nil {
			continue
		}

		// Find the corresponding start event (simplified - would need chunk correlation)
		// For now, estimate based on average throughput
		estimatedDuration := time.Minute // Placeholder estimation
		durations = append(durations, estimatedDuration)
	}

	if len(durations) == 0 {
		return 0, fmt.Errorf("no valid processing durations found")
	}

	// Calculate average
	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations)), nil
}

// MonitorProcessingProgress continuously monitors and reports processing progress
func (ct *ChunkTracker) MonitorProcessingProgress(fileHash string, interval time.Duration, stopChan <-chan struct{}) <-chan *ProcessingProgress {
	progressChan := make(chan *ProcessingProgress, 10)

	go func() {
		defer close(progressChan)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				progress, err := ct.GetProcessingProgress(fileHash)
				if err != nil {
					ct.logger.Error("Failed to get processing progress", "error", err)
					continue
				}

				select {
				case progressChan <- progress:
				case <-stopChan:
					return
				}

				// Stop monitoring if processing is complete
				if progress.ProgressPercent >= 100.0 {
					return
				}

			case <-stopChan:
				return
			}
		}
	}()

	return progressChan
}

// GetChunkProcessingHistory returns the processing history for a specific chunk
func (ct *ChunkTracker) GetChunkProcessingHistory(fileHash string, chunkIndex int) ([]ChunkProcessingEvent, error) {
	chunkID := fmt.Sprintf("chunk:%s:%d", fileHash, chunkIndex)

	query := fmt.Sprintf(`
		g.V().has('id', '%s')
		.out('HAS_EVENT')
		.order().by('timestamp', asc)
		.valueMap()
	`, chunkID)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunk history: %w", err)
	}

	var events []ChunkProcessingEvent
	for _, result := range results {
		resultMap := result.(map[string]interface{})

		timestamp, err := time.Parse(time.RFC3339, resultMap["timestamp"].(string))
		if err != nil {
			continue
		}

		event := ChunkProcessingEvent{
			ChunkID:   chunkID,
			FromState: resultMap["from_state"].(string),
			ToState:   resultMap["to_state"].(string),
			Timestamp: timestamp,
			AgentID:   resultMap["agent_id"].(string),
		}

		events = append(events, event)
	}

	return events, nil
}

// GetSystemWideMetrics provides overall system processing metrics
func (ct *ChunkTracker) GetSystemWideMetrics() (*SystemProcessingMetrics, error) {
	queries := []string{
		`g.V().hasLabel('file_original').count()`,                             // Total files
		`g.V().hasLabel('file_chunk').count()`,                                // Total chunks
		`g.V().hasLabel('file_chunk').has('status', 'completed').count()`,     // Completed chunks
		`g.V().hasLabel('file_chunk').has('status', 'processing').count()`,    // Processing chunks
		`g.V().hasLabel('file_chunk').has('status', 'failed').count()`,        // Failed chunks
		`g.V().hasLabel('file_chunk').values('processed_by').dedup().count()`, // Active processors
		`g.V().hasLabel('file_chunk').has('can_process_async', true).count()`, // Async-capable chunks
	}

	results, err := ct.storageClient.ParallelGraphQuery(queries)
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	metrics := &SystemProcessingMetrics{
		TotalFiles:         int(results[0][0].(float64)),
		TotalChunks:        int(results[1][0].(float64)),
		CompletedChunks:    int(results[2][0].(float64)),
		ProcessingChunks:   int(results[3][0].(float64)),
		FailedChunks:       int(results[4][0].(float64)),
		ActiveProcessors:   int(results[5][0].(float64)),
		AsyncCapableChunks: int(results[6][0].(float64)),
		LastUpdated:        time.Now(),
	}

	// Calculate derived metrics
	if metrics.TotalChunks > 0 {
		metrics.OverallProgress = float64(metrics.CompletedChunks) / float64(metrics.TotalChunks) * 100
		metrics.FailureRate = float64(metrics.FailedChunks) / float64(metrics.TotalChunks) * 100
		metrics.AsyncUtilization = float64(metrics.AsyncCapableChunks) / float64(metrics.TotalChunks) * 100
	}

	return metrics, nil
}

// GetProcessorPerformance analyzes performance of individual processors
func (ct *ChunkTracker) GetProcessorPerformance() (map[string]*ProcessorMetrics, error) {
	query := `
		g.V().hasLabel('file_chunk')
		.has('status', 'completed')
		.group().by('processed_by')
		.by(fold())
	`

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get processor performance: %w", err)
	}

	processors := make(map[string]*ProcessorMetrics)

	if len(results) > 0 {
		for processor, chunks := range results[0].(map[string]interface{}) {
			chunkList := chunks.([]interface{})

			metrics := &ProcessorMetrics{
				ProcessorID:     processor,
				CompletedChunks: len(chunkList),
				LastActive:      time.Now(), // Would be calculated from actual data
			}

			// Calculate total size processed
			for _, chunk := range chunkList {
				chunkMap := chunk.(map[string]interface{})
				if size, ok := chunkMap["size"].(float64); ok {
					metrics.TotalBytesProcessed += int64(size)
				}
			}

			// Calculate average chunk size
			if metrics.CompletedChunks > 0 {
				metrics.AverageChunkSize = metrics.TotalBytesProcessed / int64(metrics.CompletedChunks)
			}

			processors[processor] = metrics
		}
	}

	return processors, nil
}

// Additional metric types

type ChunkProcessingEvent struct {
	ChunkID   string    `json:"chunk_id"`
	FromState string    `json:"from_state"`
	ToState   string    `json:"to_state"`
	Timestamp time.Time `json:"timestamp"`
	AgentID   string    `json:"agent_id"`
}

type SystemProcessingMetrics struct {
	TotalFiles         int       `json:"total_files"`
	TotalChunks        int       `json:"total_chunks"`
	CompletedChunks    int       `json:"completed_chunks"`
	ProcessingChunks   int       `json:"processing_chunks"`
	FailedChunks       int       `json:"failed_chunks"`
	ActiveProcessors   int       `json:"active_processors"`
	AsyncCapableChunks int       `json:"async_capable_chunks"`
	OverallProgress    float64   `json:"overall_progress"`
	FailureRate        float64   `json:"failure_rate"`
	AsyncUtilization   float64   `json:"async_utilization"`
	LastUpdated        time.Time `json:"last_updated"`
}

type ProcessorMetrics struct {
	ProcessorID           string    `json:"processor_id"`
	CompletedChunks       int       `json:"completed_chunks"`
	TotalBytesProcessed   int64     `json:"total_bytes_processed"`
	AverageChunkSize      int64     `json:"average_chunk_size"`
	LastActive            time.Time `json:"last_active"`
	ThroughputBytesPerSec float64   `json:"throughput_bytes_per_sec"`
}

// AlertingThresholds defines thresholds for performance alerts
type AlertingThresholds struct {
	MaxFailureRate     float64       `json:"max_failure_rate"`     // Maximum acceptable failure rate (%)
	MinThroughput      float64       `json:"min_throughput"`       // Minimum chunks per second
	MaxProcessingTime  time.Duration `json:"max_processing_time"`  // Maximum time per chunk
	StallDetectionTime time.Duration `json:"stall_detection_time"` // Time to detect processing stall
}

// CheckAlerts checks if any performance thresholds are exceeded
func (ct *ChunkTracker) CheckAlerts(fileHash string, thresholds *AlertingThresholds) ([]string, error) {
	var alerts []string

	metrics, err := ct.GetProcessingMetrics(fileHash)
	if err != nil {
		return nil, err
	}

	progress, err := ct.GetProcessingProgress(fileHash)
	if err != nil {
		return nil, err
	}

	// Check failure rate
	if metrics.TotalChunks > 0 {
		failureRate := float64(metrics.FailedChunks) / float64(metrics.TotalChunks) * 100
		if failureRate > thresholds.MaxFailureRate {
			alerts = append(alerts, fmt.Sprintf("High failure rate: %.2f%% (threshold: %.2f%%)", failureRate, thresholds.MaxFailureRate))
		}
	}

	// Check throughput
	if progress.Throughput < thresholds.MinThroughput {
		alerts = append(alerts, fmt.Sprintf("Low throughput: %.2f chunks/sec (threshold: %.2f)", progress.Throughput, thresholds.MinThroughput))
	}

	// Check processing time
	if metrics.AverageProcessingTime > thresholds.MaxProcessingTime {
		alerts = append(alerts, fmt.Sprintf("Slow processing: %v per chunk (threshold: %v)", metrics.AverageProcessingTime, thresholds.MaxProcessingTime))
	}

	return alerts, nil
}
