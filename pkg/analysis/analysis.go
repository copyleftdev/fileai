package analysis

import (
	"fileai/pkg/file"
	"fmt"
)

// AnalyzeFile determines the type of the file and processes it accordingly.
func AnalyzeFile(filePath string) (string, error) {
	if !file.IsHumanReadable(filePath) {
		return "", fmt.Errorf("unsupported file type")
	}

	// Assuming text analysis is performed here if the file is human-readable.
	content, err := file.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Here you would integrate with your text analysis API or tool.
	// For now, let's just return the content as a string for demonstration.
	return string(content), nil
}
