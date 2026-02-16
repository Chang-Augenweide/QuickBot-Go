package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicRequest represents Anthropic API request
type AnthropicRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	Messages  []AnthropicMessage  `json:"messages"`
	Stream    bool                `json:"stream,omitempty"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents Anthropic API response
type AnthropicResponse struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Message AnthropicResultMessage  `json:"message"`
	Error   *AnthropicError        `json:"error,omitempty"`
}

type AnthropicResultMessage struct {
	ID      string   `json:"id"`
	Role    string   `json:"role"`
	Content []AnthropicContent `json:"content"`
	Type    string   `json:"type"`
}

type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// AnthropicProvider represents Anthropic API provider
type AnthropicProvider struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	httpClient  *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string) AIProvider {
	return &AnthropicProvider{
		apiKey:    apiKey,
		baseURL:   "https://api.anthropic.com/v1/messages",
		model:     model,
		maxTokens: 4096,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *AnthropicProvider) ProviderName() string {
	return "anthropic"
}

// ChatCompletion sends a chat completion request to Anthropic API
func (p *AnthropicProvider) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	// Convert messages to Anthropic format
	anthropicMessages := make([]AnthropicMessage, 0, len(messages))
	for i, msg := range messages {
		// Skip system messages in the conversion (handle separately if needed)
		if msg.Role == "system" {
			continue
		}

		// Anthropic doesn't support 'assistant' role in messages array
		// Convert 'assistant' to 'user' for the API
		role := msg.Role
		if role == "assistant" {
			// Skip for now, or handle differently
			continue
		}

		anthropicMessages = append(anthropicMessages, AnthropicMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Prepare request
	reqBody := AnthropicRequest{
		Model:     p.model,
		MaxTokens: p.maxTokens,
		Messages:  anthropicMessages,
		Stream:    false,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-dangerous-direct-browser-access", "false")

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
		var errorResp AnthropicResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error != nil {
			return "", fmt.Errorf("Anthropic API error: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var response AnthropicResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract text from content
	if len(response.Message.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	var result strings.Builder
	for _, content := range response.Message.Content {
		if content.Type == "text" {
			result.WriteString(content.Text)
		}
	}

	return result.String(), nil
}

// SetMaxTokens sets the maximum tokens for completion
func (p *AnthropicProvider) SetMaxTokens(maxTokens int) {
	p.maxTokens = maxTokens
}
