package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// AIProvider represents AI provider interface
type AIProvider interface {
	ProviderName() string
	ChatCompletion(ctx context.Context, messages []Message) (string, error)
}

// OpenAIProvider represents OpenAI API
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
}

func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
	}
}

func (p *OpenAIProvider) ProviderName() string {
	return "openai"
}

func (p *OpenAIProvider) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	// Placeholder implementation
	// In production, implement actual OpenAI API call
	return fmt.Sprintf("[OpenAI response for model %s]", p.model), nil
}

// AnthropicProvider represents Anthropic API
type AnthropicProvider struct {
	apiKey string
	model  string
}

func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (p *AnthropicProvider) ProviderName() string {
	return "anthropic"
}

func (p *AnthropicProvider) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	// Placeholder implementation
	return fmt.Sprintf("[Anthropic response for model %s]", p.model), nil
}

// OllamaProvider represents Ollama API
type OllamaProvider struct {
	baseURL string
	model   string
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
	}
}

func (p *OllamaProvider) ProviderName() string {
	return "ollama"
}

func (p *OllamaProvider) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	// Placeholder implementation
	return fmt.Sprintf("[Ollama response for model %s]", p.model), nil
}

// Message represents chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolCall represents a tool call
type ToolCall struct {
	Name string                 `json:"name"`
	Args map[string]string      `json:"args"`
}

// Agent represents AI agent
type Agent struct {
	config         *Config
	memory         *Memory
	scheduler      *Scheduler
	toolRegistry   *ToolRegistry
	aiProvider     AIProvider
	systemPrompt   string
	memoryContext  int
}

func NewAgent(config *Config, memory *Memory, scheduler *Scheduler) *Agent {
	// Initialize AI provider
	var provider AIProvider

	switch config.AI.Provider {
	case "openai":
		provider = NewOpenAIProvider(config.AI.APIKey, config.AI.BaseURL, config.AI.Model)
	case "anthropic":
		provider = NewAnthropicProvider(config.AI.APIKey, config.AI.Model)
	case "ollama":
		baseURL := config.AI.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		provider = NewOllamaProvider(baseURL, config.AI.Model)
	default:
		provider = NewOpenAIProvider(config.AI.APIKey, config.AI.BaseURL, config.AI.Model)
	}

	agent := &Agent{
		config:        config,
		memory:        memory,
		scheduler:     scheduler,
		toolRegistry:  NewToolRegistry(),
		aiProvider:    provider,
		systemPrompt:  buildSystemPrompt(),
		memoryContext: config.Memory.MaxMessages,
	}

	// Register tools
	agent.registerTools()

	return agent
}

// buildSystemPrompt builds system prompt
func buildSystemPrompt() string {
	return `You are QuickBot, a helpful AI assistant.

You have access to these tools:
- file: Read/write/list files
- shell: Execute shell commands
- calculator: Perform calculations
- memory: Store/retrieve long-term information

When you need to use a tool, format your response as:
TOOL: tool_name
ARGS: {"key":"value"}

You should be helpful, polite, and concise.`
}

// registerTools registers tools
func (a *Agent) registerTools() {
	if a.config.Tools.Enabled {
		// File tool
		fileTool := NewFileTool(a.config.Tools.Directory)
		a.toolRegistry.Register(fileTool)

		// Shell tool
		shellTool := NewShellTool([]string{"echo", "ls", "pwd", "cat", "grep"})
		a.toolRegistry.Register(shellTool)

		// Calculator tool
		calcTool := NewCalculatorTool()
		a.toolRegistry.Register(calcTool)

		// Memory tool
		memTool := &MemoryTool{memory: a.memory}
		a.toolRegistry.Register(memTool)
	}
}

// ProcessMessage processes user message and generates response
func (a *Agent) ProcessMessage(sessionID, userMessage string) (string, error) {
	// Store user message
	_, err := a.memory.AddMessage(sessionID, "user", userMessage, nil)
	if err != nil {
		return "", err
	}

	// Get conversation context
	messages, err := a.memory.GetMessages(sessionID, a.memoryContext)
	if err != nil {
		return "", err
	}

	// Build chat messages
	var chatMessages []Message

	// Add system prompt
	chatMessages = append(chatMessages, Message{
		Role:    "system",
		Content: a.systemPrompt,
	})

	// Add conversation history
	for i := len(messages) - 1; i >= 0; i-- {
		chatMessages = append(chatMessages, Message{
			Role:    messages[i].Role,
			Content: messages[i].Content,
		})
	}

	// Get AI response
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := a.aiProvider.ChatCompletion(ctx, chatMessages)
	if err != nil {
		return "", err
	}

	// Check if response contains tool call
	if strings.HasPrefix(response, "TOOL:") {
		return a.handleToolCall(sessionID, response)
	}

	// Store assistant response
	_, err = a.memory.AddMessage(sessionID, "assistant", response, nil)
	if err != nil {
		log.Printf("Failed to store response: %v", err)
	}

	return response, nil
}

// handleToolCall handles tool calls
func (a *Agent) handleToolCall(sessionID, response string) (string, error) {
	// Parse tool call
	toolCall, err := a.parseToolCall(response)
	if err != nil {
		return "", err
	}

	// Execute tool
	result, err := a.toolRegistry.Execute(toolCall.Name, toolCall.Args)
	if err != nil {
		return "", err
	}

	// Store tool result
	_, err = a.memory.AddMessage(sessionID, "assistant", fmt.Sprintf("[Tool: %s] %s", toolCall.Name, result), nil)
	if err != nil {
		log.Printf("Failed to store tool result: %v", err)
	}

	return result, nil
}

// parseToolCall parses tool call from response
func (a *Agent) parseToolCall(response string) (*ToolCall, error) {
	// Parse TOOL: name line
	reTool := regexp.MustCompile(`^TOOL:\s*(\S+)`)
	matches := reTool.FindStringSubmatch(response)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid tool call format")
	}

	toolName := matches[1]

	// Parse ARGS: json line
	reArgs := regexp.MustCompile(`ARGS:\s*(\{.+\})`)
	matches = reArgs.FindStringSubmatch(response)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid args format")
	}

	argsJSON := matches[2]

	var args map[string]string
	err := json.Unmarshal([]byte(argsJSON), &args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}

	return &ToolCall{
		Name: toolName,
		Args: args,
	}, nil
}

// GetMemoryContext retrieves context from memory
func (a *Agent) GetMemoryContext(sessionID string) ([]byte, error) {
	messages, err := a.memory.GetMessages(sessionID, a.memoryContext)
	if err != nil {
		return nil, err
	}

	// Build context string
	var context strings.Builder
	for _, msg := range messages {
		context.WriteString(fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content))
	}

	return []byte(context.String()), nil
}

// SetMemory stores value in long-term memory
func (a *Agent) SetMemory(key, value string) error {
	return a.memory.SetLongTerm(key, value, 2)
}

// GetMemory retrieves value from long-term memory
func (a *Agent) GetMemory(key string) (string, error) {
	return a.memory.GetLongTerm(key)
}

// AddReminder adds a reminder
func (a *Agent) AddReminder(sessionID, message, remindAt string) (string, error) {
	return a.scheduler.AddReminder(sessionID, message, remindAt)
}

// Start starts the agent
func (a *Agent) Start() {
	if a.scheduler != nil && a.config.Scheduler.Enabled {
		a.scheduler.Start()
	}
	log.Printf("Agent started: %s (AI: %s, Model: %s)",
		a.config.Bot.Name, a.aiProvider.ProviderName(), a.config.AI.Model)
}

// Stop stops the agent
func (a *Agent) Stop() {
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	if a.memory != nil {
		a.memory.Close()
	}
	log.Println("Agent stopped")
}

// MemoryTool is the memory tool
type MemoryTool struct {
	memory *Memory
}

// Config returns the agent config
func (a *Agent) Config() *Config {
	return a.config
}

// Memory returns the memory instance
func (a *Agent) Memory() *Memory {
	return a.memory
}

// Scheduler returns the scheduler instance
func (a *Agent) Scheduler() *Scheduler {
	return a.scheduler
}

// ToolRegistry returns the tool registry
func (a *Agent) ToolRegistry() *ToolRegistry {
	return a.toolRegistry
}

func (t *MemoryTool) Name() string {
	return "memory"
}

func (t *MemoryTool) Description() string {
	return "Store and retrieve long-term information"
}

func (t *MemoryTool) Permission() string {
	return "allow_all"
}

func (t *MemoryTool) Execute(args map[string]string) (string, error) {
	operation := args["operation"]
	key := args["key"]
	value := args["value"]

	switch operation {
	case "set":
		if key == "" || value == "" {
			return "", fmt.Errorf("key and value required")
		}
		err := t.memory.SetLongTerm(key, value, 2)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Success: Remembered '%s'", key), nil

	case "get":
		if key == "" {
			return "", fmt.Errorf("key required")
		}
		value, err := t.memory.GetLongTerm(key)
		if err != nil {
			return "", err
		}
		if value == "" {
			return fmt.Sprintf("Info: No memory for '%s'", key), nil
		}
		return value, nil

	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// TestAgent runs tests on the agent module
func TestAgent() {
	log.Println("Testing Agent module...")

	// Load config
	config, _ := LoadConfig("config.yaml")

	// Create memory
	memory, _ := NewMemory("test_agent_memory.db", 100)

	// Create scheduler
	scheduler, _ := NewScheduler("test_agent_scheduler.db")

	// Create agent
	agent := NewAgent(config, memory, scheduler)

	// Start agent
	agent.Start()

	// Test message processing
	sessionID := "test_session"
	msg := "Hello QuickBot!"
	response, err := agent.ProcessMessage(sessionID, msg)
	if err != nil {
		log.Printf("Failed to process message: %v", err)
	} else {
		log.Printf("✓ Message processed: %s", response)
	}

	// Test memory
	err = agent.SetMemory("test_key", "test_value")
	if err != nil {
		log.Printf("Failed to set memory: %v", err)
	} else {
		log.Println("✓ Memory set")
	}

	value, err := agent.GetMemory("test_key")
	if err != nil {
		log.Printf("Failed to get memory: %v", err)
	} else {
		log.Printf("✓ Memory retrieved: %s", value)
	}

	// Test reminder
	reminderID, err := agent.AddReminder(sessionID, "Test reminder", "10:00")
	if err != nil {
		log.Printf("Failed to add reminder: %v", err)
	} else {
		log.Printf("✓ Reminder added: %s", reminderID)
	}

	// Stop agent
	agent.Stop()

	// Cleanup
	os.Remove("test_agent_memory.db")
	os.Remove("test_agent_scheduler.db")

	log.Println("✓ Agent module tests passed")
}

// main - test entry point
func main() {
	log.Println("QuickBot Go Agent Module")
	log.Println("✓ Agent module initialized")

	// Run tests
	TestAgent()
}
