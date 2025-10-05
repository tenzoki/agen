package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetTestDataPath returns the absolute path to a test data file
func GetTestDataPath(relativePath string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	// Navigate from code/agents/testutil to project root
	agentsDir := filepath.Dir(filepath.Dir(currentFile))  // code/agents
	codeDir := filepath.Dir(agentsDir)                     // code
	projectRoot := filepath.Dir(codeDir)                   // project root
	return filepath.Join(projectRoot, "testdata", relativePath)
}

// LoadTestFile reads a test data file and returns its content
func LoadTestFile(relativePath string) ([]byte, error) {
	path := GetTestDataPath(relativePath)
	return os.ReadFile(path)
}

// LoadTestFileString reads a test data file and returns its content as string
func LoadTestFileString(relativePath string) (string, error) {
	content, err := LoadTestFile(relativePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// TestDataExists checks if a test data file exists
func TestDataExists(relativePath string) bool {
	path := GetTestDataPath(relativePath)
	_, err := os.Stat(path)
	return err == nil
}

// GetTestDataDir returns the absolute path to the test data directory
func GetTestDataDir() string {
	_, currentFile, _, _ := runtime.Caller(0)
	// Navigate from code/agents/testutil to project root
	agentsDir := filepath.Dir(filepath.Dir(currentFile))  // code/agents
	codeDir := filepath.Dir(agentsDir)                     // code
	projectRoot := filepath.Dir(codeDir)                   // project root
	return filepath.Join(projectRoot, "testdata")
}

// CreateTempTestFile creates a temporary test file with the given content
func CreateTempTestFile(content string, suffix string) (string, error) {
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("gox_test_*%s", suffix))
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// CleanupTempFile removes a temporary test file
func CleanupTempFile(path string) error {
	return os.Remove(path)
}

// Common test file paths
const (
	SampleArticlePath     = "documents/sample_article.txt"
	ResearchPaperPath     = "documents/research_paper.md"
	EmployeesCSVPath      = "structured/employees.csv"
	DatasetJSONPath       = "json/dataset_example.json"
	MetadataXMLPath       = "xml/metadata_document.xml"
	SimpleJSONPath        = "json/simple.json"
	ComplexJSONPath       = "json/complex.json"
	SimpleXMLPath         = "xml/simple.xml"
	ComplexXMLPath        = "xml/complex.xml"
	UnicodeTestPath       = "multilingual/unicode_test.txt"
	EmptyFilePath         = "edge-cases/empty.txt"
	LargeTextPath         = "edge-cases/large_text.txt"
	MalformedJSONPath     = "edge-cases/malformed.json"
	MalformedXMLPath      = "edge-cases/malformed.xml"
)

// TestFileInfo represents information about a test file
type TestFileInfo struct {
	Path        string
	Description string
	Type        string
	Size        string
	Language    string
}

// GetAvailableTestFiles returns a list of available test files with their information
func GetAvailableTestFiles() []TestFileInfo {
	return []TestFileInfo{
		{
			Path:        SampleArticlePath,
			Description: "Long-form article about machine learning",
			Type:        "text/plain",
			Size:        "large",
			Language:    "en",
		},
		{
			Path:        ResearchPaperPath,
			Description: "Academic paper in Markdown format",
			Type:        "text/markdown",
			Size:        "large",
			Language:    "en",
		},
		{
			Path:        EmployeesCSVPath,
			Description: "Employee data in CSV format",
			Type:        "text/csv",
			Size:        "small",
			Language:    "en",
		},
		{
			Path:        DatasetJSONPath,
			Description: "Machine learning dataset in JSON format",
			Type:        "application/json",
			Size:        "medium",
			Language:    "en",
		},
		{
			Path:        MetadataXMLPath,
			Description: "Structured document with metadata in XML format",
			Type:        "application/xml",
			Size:        "medium",
			Language:    "multi",
		},
		{
			Path:        UnicodeTestPath,
			Description: "Text file with Unicode characters",
			Type:        "text/plain",
			Size:        "small",
			Language:    "multi",
		},
		{
			Path:        EmptyFilePath,
			Description: "Empty file for edge case testing",
			Type:        "text/plain",
			Size:        "empty",
			Language:    "none",
		},
		{
			Path:        LargeTextPath,
			Description: "Large text file for performance testing",
			Type:        "text/plain",
			Size:        "large",
			Language:    "en",
		},
	}
}