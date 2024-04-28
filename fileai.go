package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/h2non/filetype.v1"
)

const (
	openAIURL   = "https://api.openai.com/v1/chat/completions"
	contentType = "application/json"
)

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "faileai -f [file path]",
		Short: "FaileAI analyzes files and summarizes if human-readable",
		Run:   runFaileai,
	}

	rootCmd.PersistentFlags().StringP("file", "f", "", "File to analyze")
	rootCmd.MarkPersistentFlagRequired("file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runFaileai(cmd *cobra.Command, args []string) {
	filename, _ := cmd.Flags().GetString("file")

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	if isHumanReadable(content) {
		description, err := getSummaryFromAI(string(content))
		if err != nil {
			fmt.Printf("Error getting summary from AI: %v\n", err)
			return
		}

		result := map[string]string{
			"filename":    filename,
			"description": description,
		}
		resultJSON, _ := json.Marshal(result)
		fmt.Println(string(resultJSON))
	} else {
		kind, _ := filetype.Match(content)
		fileInfo := fmt.Sprintf("File '%s' is of type '%s'.", filename, kind.MIME.Value)
		fmt.Println(fileInfo)
	}
}

func isHumanReadable(content []byte) bool {
	kind, err := filetype.Match(content)
	if err != nil || kind == filetype.Unknown {
		// Assume text if the filetype is unknown and content is not empty
		return len(content) > 0 && isText(content)
	}
	return kind.MIME.Type == "text" || strings.HasPrefix(kind.MIME.Value, "application/json") || strings.HasPrefix(kind.MIME.Value, "application/xml")
}

func isText(data []byte) bool {
	for _, b := range data {
		if b != 0x09 && b != 0x0A && b != 0x0D && (b < 0x20 || b > 0x7E) {
			return false
		}
	}
	return true
}

func getSummaryFromAI(text string) (string, error) {
	authToken := fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY"))
	if authToken == "Bearer " {
		return "", fmt.Errorf("API key not set in the environment variables")
	}

	requestData := OpenAIRequest{
		Model: "gpt-4",
		Messages: []Message{
			{"system", "You are a helpful assistant."},
			{"user", "Summarize this content"},
			{"assistant", text},
		},
	}
	requestBody, _ := json.Marshal(requestData)

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	// Check if choices exist and are in the correct format
	choices, ok := res["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format or empty choices")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format for first choice")
	}

	messages, ok := firstChoice["messages"].([]interface{})
	if !ok || len(messages) == 0 {
		return "", fmt.Errorf("invalid format for messages or empty messages")
	}

	lastMessage, ok := messages[len(messages)-1].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format for last message")
	}

	description, ok := lastMessage["content"].(string)
	if !ok {
		return "", fmt.Errorf("content missing from last message")
	}

	return description, nil
}
