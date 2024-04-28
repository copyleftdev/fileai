package analysis

import (
	"fileai/config"
	"fileai/pkg/api"
	"fileai/pkg/file"
	"fmt"
)

// AnalyzeFile determines the type of the file and processes it accordingly.
func AnalyzeFile(filePath string) (string, error) {
	content, err := file.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	if file.IsHumanReadable(content) {
		textContent := string(content)
		prompt := config.AppConfig.Prompts["text"].Summary
		return api.SummarizeText(textContent, prompt)
	} else if file.IsImage(filePath) {
		prompt := config.AppConfig.Prompts["image"].Description
		return api.DescribeImage(filePath, prompt)
	}

	return "", fmt.Errorf("unsupported file type")
}
