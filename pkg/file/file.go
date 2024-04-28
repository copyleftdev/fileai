package file

import (
	"os"
	"path/filepath"
	"strings"
)

// ReadFile reads the content of the given file path and returns it as a byte slice.
func ReadFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// IsHumanReadable checks if the file content is likely to be text based on its MIME type.
func IsHumanReadable(content []byte) bool {
	// You might use a third-party library to check the MIME type or use a simple heuristic based on content.
	return isText(content)
}

// IsImage checks if the file extension indicates an image.
func IsImage(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff":
		return true
	default:
		return false
	}
}

// isText attempts to determine if content is text by checking for non-text byte patterns.
func isText(data []byte) bool {
	// Simple heuristic: Check if most bytes are printable characters or common text-related control characters.
	for _, b := range data {
		if b != '\n' && b != '\r' && b != '\t' && (b < 32 || b > 127) {
			return false
		}
	}
	return true
}
