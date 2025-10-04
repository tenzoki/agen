package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

type TextAnalyzer struct {
	agent.DefaultAgentRunner
	config *AnalyzerConfig
}

type AnalyzerConfig struct {
	EnableNLP       bool `json:"enable_nlp"`
	EnableSentiment bool `json:"enable_sentiment"`
	EnableKeywords  bool `json:"enable_keywords"`
	MaxLines        int  `json:"max_lines"`
	MaxKeywords     int  `json:"max_keywords"`
}

type ChunkProcessingRequest struct {
	RequestID   string                 `json:"request_id"`
	ChunkID     string                 `json:"chunk_id"`
	FileID      string                 `json:"file_id"`
	ChunkIndex  int                    `json:"chunk_index"`
	ContentType string                 `json:"content_type"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Options     map[string]interface{} `json:"options"`
	CreatedAt   time.Time              `json:"created_at"`
}

type ProcessingResult struct {
	RequestID      string                 `json:"request_id"`
	ChunkID        string                 `json:"chunk_id"`
	FileID         string                 `json:"file_id"`
	ChunkIndex     int                    `json:"chunk_index"`
	ProcessedBy    string                 `json:"processed_by"`
	ProcessingType string                 `json:"processing_type"`
	ResultData     map[string]interface{} `json:"result_data"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessedAt    time.Time              `json:"processed_at"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
}

func NewTextAnalyzer() *TextAnalyzer {
	return &TextAnalyzer{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &AnalyzerConfig{
			EnableNLP:       false,
			EnableSentiment: true,
			EnableKeywords:  true,
			MaxLines:        10000,
			MaxKeywords:     20,
		},
	}
}

func (ta *TextAnalyzer) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "chunk_processing_request" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var request ChunkProcessingRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	if request.ContentType != "text" && request.ContentType != "text/plain" {
		return nil, fmt.Errorf("unsupported content type: %s", request.ContentType)
	}

	ta.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &ProcessingResult{
		RequestID:      request.RequestID,
		ChunkID:        request.ChunkID,
		FileID:         request.FileID,
		ChunkIndex:     request.ChunkIndex,
		ProcessedBy:    "text_analyzer",
		ProcessingType: "text_analysis",
		ResultData:     make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		ProcessedAt:    startTime,
	}

	processedText, analysisData, err := ta.analyzeText(request.Content)
	if err != nil {
		result.Error = fmt.Sprintf("text analysis failed: %v", err)
		result.Success = false
		result.ProcessingTime = time.Since(startTime)
		return ta.createResultMessage(result), nil
	}

	result.ResultData["processed_content"] = processedText
	result.ResultData = mergeMapStringInterface(result.ResultData, analysisData)
	result.Metadata["analyzer"] = "text_analysis"
	result.Metadata["original_length"] = len(request.Content)
	result.Metadata["processed_length"] = len(processedText)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return ta.createResultMessage(result), nil
}

func (ta *TextAnalyzer) loadConfigFromAgent(base *agent.BaseAgent) {
	if maxLines := base.GetConfigInt("max_lines", 10000); maxLines > 0 {
		ta.config.MaxLines = maxLines
	}
	if maxKeywords := base.GetConfigInt("max_keywords", 20); maxKeywords > 0 {
		ta.config.MaxKeywords = maxKeywords
	}
	ta.config.EnableNLP = base.GetConfigBool("enable_nlp", false)
	ta.config.EnableSentiment = base.GetConfigBool("enable_sentiment", true)
	ta.config.EnableKeywords = base.GetConfigBool("enable_keywords", true)
}

func (ta *TextAnalyzer) analyzeText(text string) (string, map[string]interface{}, error) {
	analysisData := make(map[string]interface{})

	// Basic text statistics
	lines := strings.Split(text, "\n")
	words := strings.Fields(text)
	chars := len(text)

	analysisData["line_count"] = len(lines)
	analysisData["word_count"] = len(words)
	analysisData["char_count"] = chars

	// Text normalization
	processedText := strings.TrimSpace(text)
	processedText = regexp.MustCompile(`\s+`).ReplaceAllString(processedText, " ")

	// Truncate if too many lines
	if len(lines) > ta.config.MaxLines {
		lines = lines[:ta.config.MaxLines]
		processedText = strings.Join(lines, "\n")
		analysisData["truncated"] = true
		analysisData["truncated_at_lines"] = ta.config.MaxLines
	}

	// Extract keywords if enabled
	if ta.config.EnableKeywords {
		keywords := ta.extractKeywords(text)
		analysisData["keywords"] = keywords
	}

	// Sentiment analysis if enabled
	if ta.config.EnableSentiment {
		sentiment, sentimentScore := ta.analyzeSentiment(text)
		analysisData["sentiment"] = sentiment
		analysisData["sentiment_score"] = sentimentScore
	}

	// Language detection (simple heuristic)
	language := ta.detectLanguage(text)
	analysisData["language"] = language

	// Text quality metrics
	analysisData["avg_word_length"] = ta.calculateAverageWordLength(words)
	analysisData["sentence_count"] = ta.countSentences(text)
	analysisData["reading_level"] = ta.estimateReadingLevel(text)

	// Processing metadata
	analysisData["processed_by"] = "text_analyzer"
	analysisData["processed_at"] = time.Now().Format(time.RFC3339)
	analysisData["processing_type"] = "text_normalization_and_analysis"

	return processedText, analysisData, nil
}

func (ta *TextAnalyzer) extractKeywords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// Common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "should": true, "could": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true, "me": true, "him": true,
		"her": true, "us": true, "them": true, "my": true, "your": true, "his": true, "our": true,
		"their": true, "can": true, "may": true, "might": true, "must": true, "shall": true,
	}

	for _, word := range words {
		// Remove punctuation and filter
		cleanWord := regexp.MustCompile(`[^\w]`).ReplaceAllString(word, "")
		if len(cleanWord) > 2 && !stopWords[cleanWord] {
			wordCount[cleanWord]++
		}
	}

	// Convert to sorted list by frequency
	type wordFreq struct {
		word string
		freq int
	}

	var sortedWords []wordFreq
	for word, freq := range wordCount {
		if freq >= 2 { // Word appears at least twice
			sortedWords = append(sortedWords, wordFreq{word, freq})
		}
	}

	// Sort by frequency (descending)
	for i := 0; i < len(sortedWords)-1; i++ {
		for j := i + 1; j < len(sortedWords); j++ {
			if sortedWords[i].freq < sortedWords[j].freq {
				sortedWords[i], sortedWords[j] = sortedWords[j], sortedWords[i]
			}
		}
	}

	// Extract top keywords
	maxKeywords := ta.config.MaxKeywords
	if len(sortedWords) < maxKeywords {
		maxKeywords = len(sortedWords)
	}

	keywords := make([]string, maxKeywords)
	for i := 0; i < maxKeywords; i++ {
		keywords[i] = sortedWords[i].word
	}

	return keywords
}

func (ta *TextAnalyzer) analyzeSentiment(text string) (string, float64) {
	positiveWords := []string{
		"good", "great", "excellent", "amazing", "wonderful", "fantastic", "positive",
		"happy", "success", "win", "love", "like", "best", "perfect", "awesome",
		"brilliant", "outstanding", "superb", "magnificent", "beautiful", "joy",
	}

	negativeWords := []string{
		"bad", "terrible", "awful", "horrible", "negative", "sad", "fail", "failure",
		"lose", "problem", "hate", "dislike", "worst", "horrible", "disgusting",
		"annoying", "frustrating", "disappointing", "angry", "upset", "worried",
	}

	lowerText := strings.ToLower(text)
	words := strings.Fields(lowerText)
	totalWords := len(words)

	if totalWords == 0 {
		return "neutral", 0.0
	}

	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(lowerText, word)
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(lowerText, word)
	}

	// Calculate sentiment score (-1 to +1)
	sentimentScore := float64(positiveCount-negativeCount) / float64(totalWords) * 100

	// Classify sentiment
	if sentimentScore > 0.5 {
		return "positive", sentimentScore
	} else if sentimentScore < -0.5 {
		return "negative", sentimentScore
	}
	return "neutral", sentimentScore
}

func (ta *TextAnalyzer) detectLanguage(text string) string {
	// Simple language detection based on common words
	lowerText := strings.ToLower(text)

	englishIndicators := []string{"the", "and", "of", "to", "a", "in", "for", "is", "on", "that"}
	germanIndicators := []string{"der", "die", "das", "und", "in", "den", "von", "zu", "mit", "ist"}
	frenchIndicators := []string{"le", "de", "et", "à", "un", "il", "être", "et", "en", "avoir"}
	spanishIndicators := []string{"el", "la", "de", "que", "y", "a", "en", "un", "ser", "se"}

	englishScore := ta.countLanguageIndicators(lowerText, englishIndicators)
	germanScore := ta.countLanguageIndicators(lowerText, germanIndicators)
	frenchScore := ta.countLanguageIndicators(lowerText, frenchIndicators)
	spanishScore := ta.countLanguageIndicators(lowerText, spanishIndicators)

	maxScore := englishScore
	language := "en"

	if germanScore > maxScore {
		maxScore = germanScore
		language = "de"
	}
	if frenchScore > maxScore {
		maxScore = frenchScore
		language = "fr"
	}
	if spanishScore > maxScore {
		language = "es"
	}

	// If no clear language detected
	if maxScore < 3 {
		return "unknown"
	}

	return language
}

func (ta *TextAnalyzer) countLanguageIndicators(text string, indicators []string) int {
	count := 0
	for _, indicator := range indicators {
		count += strings.Count(text, " "+indicator+" ")
		count += strings.Count(text, indicator+" ") // Start of text
		count += strings.Count(text, " "+indicator) // End of text
	}
	return count
}

func (ta *TextAnalyzer) calculateAverageWordLength(words []string) float64 {
	if len(words) == 0 {
		return 0.0
	}

	totalLength := 0
	for _, word := range words {
		totalLength += len(word)
	}

	return float64(totalLength) / float64(len(words))
}

func (ta *TextAnalyzer) countSentences(text string) int {
	// Simple sentence counting based on punctuation
	sentenceEnders := []string{".", "!", "?"}
	count := 0

	for _, ender := range sentenceEnders {
		count += strings.Count(text, ender)
	}

	// If no sentence enders found but text exists, count as 1 sentence
	if count == 0 && strings.TrimSpace(text) != "" {
		count = 1
	}

	return count
}

func (ta *TextAnalyzer) estimateReadingLevel(text string) string {
	words := strings.Fields(text)
	sentences := ta.countSentences(text)

	if len(words) == 0 || sentences == 0 {
		return "unknown"
	}

	// Simple Flesch Reading Ease approximation
	avgWordsPerSentence := float64(len(words)) / float64(sentences)
	avgWordLength := ta.calculateAverageWordLength(words)

	// Simplified scoring
	if avgWordsPerSentence < 10 && avgWordLength < 4 {
		return "elementary"
	} else if avgWordsPerSentence < 15 && avgWordLength < 5 {
		return "middle_school"
	} else if avgWordsPerSentence < 20 && avgWordLength < 6 {
		return "high_school"
	} else {
		return "college"
	}
}

func (ta *TextAnalyzer) createResultMessage(result *ProcessingResult) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "chunk_processing_result",
		Target:    "chunk_processing_result",
		Payload:   result,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func mergeMapStringInterface(dst, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func main() {
	textAnalyzer := NewTextAnalyzer()
	agent.Run(textAnalyzer, "text-analyzer")
}