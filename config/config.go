package config

import (
	"encoding/json"
	"io"
	"os"
)

// PromptsStructure defines the structure for different types of prompts.
type PromptsStructure struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

// Config represents the application configuration with nested prompts.
type Config struct {
	Prompts map[string]PromptsStructure `json:"prompts"`
}

// AppConfig holds the runtime configuration loaded from JSON.
var AppConfig Config

// LoadConfig loads configuration data from a JSON file.
func LoadConfig() error {
	filePath := "prompts/prompts.json" // Adjust the path as necessary.
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &AppConfig); err != nil {
		return err
	}

	return nil
}
