package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// OpenAI API endpoint
const openAIURL = "https://api.openai.com/v1/engines/davinci-codex/completions"

// OpenAIRequest defines the structure for an API request
type OpenAIRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

// OpenAIResponse defines the structure for an API response
type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
	Error *OpenAIError `json:"error"` // Use a pointer to an OpenAIError struct
}

// OpenAIError defines the structure for an error response from OpenAI
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// SummarizeText sends text to the OpenAI API for summarization
func SummarizeText(text, prompt string) (string, error) {
	return callOpenAI(text, prompt, 150) // Adjust max tokens as necessary
}

// DescribeImage sends an image description request to the OpenAI API
func DescribeImage(imagePath, prompt string) (string, error) {
	return callOpenAI(imagePath, prompt, 150) // Adjust max tokens as necessary
}

// callOpenAI handles the common functionality of calling the OpenAI API
func callOpenAI(input, prompt string, maxTokens int) (string, error) {
	requestData := OpenAIRequest{
		Prompt:    prompt + "\n\n" + input, // Combine prompt with input
		MaxTokens: maxTokens,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %v", err)
	}

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is not set")
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to OpenAI: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var apiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("error unmarshalling response from OpenAI: %v", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("error from OpenAI: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) > 0 {
		return apiResp.Choices[0].Text, nil
	}

	return "", fmt.Errorf("no response from OpenAI")
}
