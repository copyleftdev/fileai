package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/h2non/filetype.v1"
)

const (
	openAIURL   = "https://api.openai.com/v1/chat/completions"
	contentType = "application/json"
	maxTokens   = 128000 // Max tokens for your selected model
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

func encodeImage(imagePath string) (string, error) {
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(imageBytes), nil
}

func runFaileai(cmd *cobra.Command, args []string) {
	filename, _ := cmd.Flags().GetString("file")
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	if isHumanReadable(content) && AccurateTokenCount(string(content)) <= maxTokens {
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
	} else if isImage(filepath.Ext(filename)) {
		base64Image, err := encodeImage(filename)
		if err != nil {
			fmt.Printf("Error encoding image: %v\n", err)
			return
		}
		description, err := getDescriptionOfImage(base64Image)
		if err != nil {
			fmt.Printf("Error getting image description: %v\n", err)
			return
		}
		fmt.Printf("Image description: %s\n", description)
	} else {
		kind, _ := filetype.Match(content)
		fileInfo := fmt.Sprintf("File '%s' is of type '%s'.", filename, kind.MIME.Value)
		fmt.Println(fileInfo)
	}
}

func getDescriptionOfImage(base64Image string) (string, error) {
	authToken := fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY"))
	payload := map[string]interface{}{
		"model": "gpt-4-vision-preview",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "What’s in this image?",
					},
					map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]string{
							"url": "data:image/jpeg;base64," + base64Image,
						},
					},
				},
			},
		},
		"max_tokens": 300,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %v", err)
	}

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	fmt.Println("DEBUG - Raw Response:", string(bodyBytes))

	var res map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	choices, ok := res["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format or empty choices")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format for first choice")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format for message")
	}

	description, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("content missing from message")
	}

	return description, nil
}

func isImage(ext string) bool {
	// Check if the file extension indicates an image
	return strings.ToLower(ext) == ".jpg" || strings.ToLower(ext) == ".jpeg" || strings.ToLower(ext) == ".png" || strings.ToLower(ext) == ".gif" || strings.ToLower(ext) == ".bmp" || strings.ToLower(ext) == ".tiff"
}

func isHumanReadable(content []byte) bool {
	kind, err := filetype.Match(content)
	if err == nil && kind != filetype.Unknown {
		return kind.MIME.Type == "text" || strings.HasPrefix(kind.MIME.Value, "application/json") || strings.HasPrefix(kind.MIME.Value, "application/xml")
	}
	return isLikelyText(content)
}

func isLikelyText(data []byte) bool {
	const sampleSize = 512 // Check the first 512 bytes, or the full content if smaller
	limit := len(data)
	if limit > sampleSize {
		limit = sampleSize
	}
	textCount := 0
	for _, b := range data[:limit] {
		if b == '\n' || b == '\r' || b == '\t' || b >= ' ' && b <= '~' { // printable characters and common whitespace
			textCount++
		}
	}
	return textCount >= int(0.9*float64(limit))
}
func getSummaryFromAI(text string) (string, error) {
	authToken := fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY"))
	if authToken == "Bearer " {
		return "", fmt.Errorf("API key not set in the environment variables")
	}

	requestData := OpenAIRequest{
		Model: "gpt-4-1106-preview",
		Messages: []Message{
			{"system", "You are a helpful assistant."},
			{"user", "Summarize this content"},
			{"assistant", text},
		},
	}
	requestBody, _ := json.Marshal(requestData)

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	choices, ok := res["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format or empty choices")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format for first choice")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format for message")
	}

	description, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("content missing from message")
	}

	return description, nil
}

func AccurateTokenCount(text string) int {
	// More accurate token estimation considering spaces and some punctuation
	return strings.Count(text, " ") + strings.Count(text, ".") + strings.Count(text, ",") + strings.Count(text, ";") + strings.Count(text, ":") + 1
}
