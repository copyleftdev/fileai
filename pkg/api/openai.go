package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

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
	return callOpenAI("gpt-4", []Message{
		{Role: "system", Content: "You are a helpful assistant. Return a JSON object with a 'summary' and 'tags'."},
		{Role: "user", Content: input},
	})
}

// Function to handle image description using the Chat API
func DescribeImage(base64Image string) (string, error) {
	return callOpenAI("gpt-4", []Message{
		{Role: "user", Content: fmt.Sprintf("Describe this image and return a JSON object with 'description' and 'tags': [image data:image/jpeg;base64,%s]", base64Image)},
	})
}

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
