package analysis

import (
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
		// Call the appropriate function without passing the prompt directly
		return api.AnalyzeText(textContent)
	} else if file.IsImage(filePath) {
		// Assume content is base64-encoded image data for this example; adjust as needed
		base64Image, err := file.EncodeToBase64(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to encode image: %v", err)
		}
		// Call the appropriate function without a prompt as the second argument
		return api.DescribeImage(base64Image)
	}

	return "", fmt.Errorf("unsupported file type")
}
