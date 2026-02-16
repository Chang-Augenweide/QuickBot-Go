package types

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Session represents a conversation session
type Session struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Platform  string                 `json:"platform"`
	UserID    string                 `json:"user_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Tool represents a tool interface
type Tool interface {
	Name() string
	Description() string
	Permission() string
	Execute(args map[string]string) (string, error)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}
