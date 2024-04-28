package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Define a struct for each message within the messages array
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Define the request structure to include an array of messages
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Response and Error handling structures remain the same as your needs
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *OpenAIError `json:"error"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Function to handle text analysis using the Chat API
func AnalyzeText(input string) (string, error) {
	return callOpenAI("gpt-4-turbo", []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: input},
	})
}

// Function to handle image description using the Chat API
func DescribeImage(base64Image string) (string, error) {
	return callOpenAI("gpt-4-turbo", []Message{
		{Role: "user", Content: fmt.Sprintf("[image data:image/jpeg;base64,%s]", base64Image)},
	})
}

// callOpenAI sends requests to the OpenAI API
func callOpenAI(model string, messages []Message) (string, error) {
	requestData := OpenAIRequest{
		Model:    model,
		Messages: messages,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
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
		return apiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from OpenAI")
}
