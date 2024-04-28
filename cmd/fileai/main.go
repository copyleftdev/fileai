package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fileai/config"
	"fileai/pkg/analysis"

	"github.com/spf13/cobra"
)

func main() {
	var filePath string

	// Root command for the FileAI application
	var rootCmd = &cobra.Command{
		Use:   "fileai -f [file path]",
		Short: "FileAI analyzes files and provides summaries or descriptions based on content type.",
		Long: `FileAI is a tool that analyzes files and provides summaries for text content
and descriptions for images. It uses OpenAI's models to process the content dynamically
based on the file type.`,
		Run: func(cmd *cobra.Command, args []string) {
			runFileAI(filePath)
		},
	}

	// Adding a flag to the root command to accept a file path
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the file to be analyzed")
	rootCmd.MarkFlagRequired("file")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runFileAI handles the logic to process the file based on its type
func runFileAI(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute file path: %v\n", err)
		return
	}

	// Load configuration and dynamic prompts
	if err := config.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		return
	}

	// Analyze the file based on its content type
	result, err := analysis.AnalyzeFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing file: %v\n", err)
		return
	}

	// Print the analysis results
	fmt.Println(result)
}
