package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"quickbot/internal/types"
)

// OpenAIRequest represents OpenAI API request
type OpenAIRequest struct {
	Model          string         `json:"model"`
	Messages       []types.Message `json:"messages"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
	Temperature    float64       `json:"temperature,omitempty"`
	Stream         bool          `json:"stream,omitempty"`
}

// OpenAIResponse represents OpenAI API response
type OpenAIResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Choices []OpenAIChoice       `json:"choices"`
	Error   *OpenAIError         `json:"error,omitempty"`
}

type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      types.Message  `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// OpenAIProvider represents OpenAI API provider
type OpenAIProvider struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float64
	httpClient  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, baseURL, model string) AIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		maxTokens: 2000,
		temperature: 0.7,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *OpenAIProvider) ProviderName() string {
	return "openai"
}

// ChatCompletion sends a chat completion request to OpenAI API
func (p *OpenAIProvider) ChatCompletion(ctx context.Context, messages []types.Message) (string, error) {
	// Convert to OpenAI format
	openAIMessages := make([]types.Message, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = types.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Prepare request
	reqBody := OpenAIRequest{
		Model:       p.model,
		Messages:    openAIMessages,
		MaxTokens:   p.maxTokens,
		Temperature: p.temperature,
		Stream:      false,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResp OpenAIResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error != nil {
			return "", fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var response OpenAIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// SetMaxTokens sets the maximum tokens for completion
func (p *OpenAIProvider) SetMaxTokens(maxTokens int) {
	p.maxTokens = maxTokens
}

// SetTemperature sets the temperature for completion
func (p *OpenAIProvider) SetTemperature(temperature float64) {
	p.temperature = temperature
}
