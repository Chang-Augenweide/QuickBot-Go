package ai

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

// OllamaRequest represents Ollama API request
type OllamaRequest struct {
	Model    string              `json:"model"`
	Messages []OllamaMessage     `json:"messages"`
	Stream   bool                `json:"stream"`
	Options  OllamaOptions       `json:"options,omitempty"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaOptions struct {
	NumPredict int     `json:"num_predict,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// OllamaResponse represents Ollama API response
type OllamaResponse struct {
	Model     string          `json:"model"`
	CreatedAt string         `json:"created_at"`
	Message   types.Message  `json:"message"`
	Done     bool           `json:"done"`
	Error     string         `json:"error,omitempty"`
}

// OllamaProvider represents Ollama API provider
type OllamaProvider struct {
	baseURL     string
	model       string
	numPredict  int
	temperature float64
	httpClient  *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL, model string) AIProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		numPredict: 2000,
		temperature: 0.7,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Ollama can be slow
		},
	}
}

func (p *OllamaProvider) ProviderName() string {
	return "ollama"
}

// ChatCompletion sends a chat completion request to Ollama API
func (p *OllamaProvider) ChatCompletion(ctx context.Context, messages []types.Message) (string, error) {
	// Convert messages to Ollama format
	ollamaMessages := make([]OllamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Prepare request
	reqBody := OllamaRequest{
		Model:    p.model,
		Messages: ollamaMessages,
		Stream:   false,
		Options: OllamaOptions{
			NumPredict:   p.numPredict,
			Temperature:  p.temperature,
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

	// Parse response
	var response OllamaResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for errors
	if response.Error != "" {
		return "", fmt.Errorf("Ollama API error: %s", response.Error)
	}

	return response.Message.Content, nil
}

// SetNumPredict sets the maximum number of tokens to predict
func (p *OllamaProvider) SetNumPredict(numPredict int) {
	p.numPredict = numPredict
}

// SetTemperature sets the temperature for completion
func (p *OllamaProvider) SetTemperature(temperature float64) {
	p.temperature = temperature
}
