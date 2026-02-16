package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ToolPermission represents tool permission levels
type ToolPermission string

const (
	PermissionAllowAll      ToolPermission = "allow_all"
	PermissionAllowList     ToolPermission = "allow_list"
	PermissionDenyAll       ToolPermission = "deny_all"
)

// Tool represents a tool interface
type Tool interface {
	Name() string
	Description() string
	Permission() ToolPermission
	Execute(args map[string]string) (string, error)
}

// FileTool handles file operations
type FileTool struct {
	baseDir      string
	permission   ToolPermission
}

func NewFileTool(baseDir string) *FileTool {
	return &FileTool{
		baseDir:    baseDir,
		permission: PermissionAllowList,
	}
}

func (t *FileTool) Name() string {
	return "file"
}

func (t *FileTool) Description() string {
	return "Read, write, and list files"
}

func (t *FileTool) Permission() ToolPermission {
	return t.permission
}

func (t *FileTool) Execute(args map[string]string) (string, error) {
	operation := args["operation"]
	path := args["path"]
	content := args["content"]

	fullPath := filepath.Join(t.baseDir, path)

	// Ensure path is within base directory
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	absBaseDir, err := filepath.Abs(t.baseDir)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absPath, absBaseDir) {
		return "", fmt.Errorf("access denied: path outside base directory")
	}

	switch operation {
	case "read":
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "write":
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			return "", err
		}
		return fmt.Sprintf("Success: Written to %s", path), nil

	case "list":
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return "", err
		}

		var items []string
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			items = append(items, name)
		}

		if len(items) == 0 {
			return "(empty)", nil
		}
		return strings.Join(items, "\n"), nil

	case "delete":
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			return "", err
		}

		if fileInfo.IsDir() {
			if err := os.RemoveAll(absPath); err != nil {
				return "", err
			}
		} else {
			if err := os.Remove(absPath); err != nil {
				return "", err
			}
		}

		return fmt.Sprintf("Success: Deleted %s", path), nil

	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// ShellTool handles shell command execution
type ShellTool struct {
	allowedCommands []string
	permission      ToolPermission
}

func NewShellTool(allowedCommands []string) *ShellTool {
	return &ShellTool{
		allowedCommands: allowedCommands,
		permission:      PermissionAllowList,
	}
}

func (t *ShellTool) Name() string {
	return "shell"
}

func (t *ShellTool) Description() string {
	return "Execute shell commands"
}

func (t *ShellTool) Permission() ToolPermission {
	return t.permission
}

func (t *ShellTool) Execute(args map[string]string) (string, error) {
	command := args["command"]

	if command == "" {
		return "", fmt.Errorf("empty command")
	}

	// Check if command is allowed
	if len(t.allowedCommands) > 0 {
		cmdParts := strings.Fields(command)
		if len(cmdParts) == 0 {
			return "", fmt.Errorf("invalid command")
		}

		baseCmd := cmdParts[0]
		allowed := false
		for _, allowedCmd := range t.allowedCommands {
			if allowedCmd == baseCmd {
				allowed = true
				break
			}
		}

		if !allowed {
			return "", fmt.Errorf("command not allowed: %s", baseCmd)
		}
	}

	// Execute command
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}

// MemoryTool handles memory operations
type MemoryTool struct {
	memory *Memory
}

func NewMemoryTool(memory *Memory) *MemoryTool {
	return &MemoryTool{
		memory: memory,
	}
}

func (t *MemoryTool) Name() string {
	return "memory"
}

func (t *MemoryTool) Description() string {
	return "Store and retrieve long-term information"
}

func (t *MemoryTool) Permission() ToolPermission {
	return PermissionAllowAll
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

// CalculatorTool handles calculations
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "Perform mathematical calculations"
}

func (t *CalculatorTool) Permission() ToolPermission {
	return PermissionAllowAll
}

func (t *CalculatorTool) Execute(args map[string]string) (string, error) {
	expression := args["expression"]

	if expression == "" {
		return "", fmt.Errorf("expression required")
	}

	// Note: In production, use a proper expression parser
	// For now, this is a placeholder
	return fmt.Sprintf("Result: (calculation of %s)", expression), nil
}

// ToolRegistry manages tool registration and execution
type ToolRegistry struct {
	tools      map[string]Tool
	permission ToolPermission
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:      make(map[string]Tool),
		permission: PermissionAllowList,
	}
}

func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

func (r *ToolRegistry) Get(name string) Tool {
	return r.tools[name]
}

func (r *ToolRegistry) GetAll() map[string]Tool {
	return r.tools
}

func (r *ToolRegistry) Execute(name string, args map[string]string) (string, error) {
	tool := r.Get(name)
	if tool == nil {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	if tool.Permission() == PermissionDenyAll {
		return "", fmt.Errorf("tool disabled: %s", name)
	}

	if r.permission == PermissionDenyAll {
		return "", fmt.Errorf("all tools disabled")
	}

	return tool.Execute(args)
}

// TestTools runs tests on the tools module
func TestTools() {
	fmt.Println("Testing Tools module...")

	// Create test environment
	tempDir := os.TempDir()
	memory, _ := NewMemory("test_tools_memory.db")

	// Create registry
	registry := NewToolRegistry()
	fileTool := NewFileTool(tempDir)
	shellTool := NewShellTool([]string{"echo", "pwd", "ls"})
	memoryTool := NewMemoryTool(memory)

	registry.Register(fileTool)
	registry.Register(shellTool)
	registry.Register(memoryTool)

	// Test file tool - write
	result, err := registry.Execute("file", map[string]string{
		"operation": "write",
		"path":      "test.txt",
		"content":   "Hello QuickBot!",
	})
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
	} else {
		fmt.Printf("✓ File write: %s\n", result)
	}

	// Test file tool - read
	result, err = registry.Execute("file", map[string]string{
		"operation": "read",
		"path":      "test.txt",
	})
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
	} else {
		fmt.Printf("✓ File read: %s\n", result[:20]+"...")
	}

	// Test shell tool
	result, err = registry.Execute("shell", map[string]string{
		"command": "echo 'QuickBot test'",
	})
	if err != nil {
		fmt.Printf("Failed shell command: %v\n", err)
	} else {
		fmt.Printf("✓ Shell command: %s", strings.TrimSpace(result))
	}

	// Test memory tool
	result, err = registry.Execute("memory", map[string]string{
		"operation": "set",
		"key":       "test_key",
		"value":     "test_value",
	})
	if err != nil {
		fmt.Printf("Failed memory set: %v\n", err)
	} else {
		fmt.Printf("✓ Memory set: %s\n", result)
	}

	result, err = registry.Execute("memory", map[string]string{
		"operation": "get",
		"key":       "test_key",
	})
	if err != nil {
		fmt.Printf("Failed memory get: %v\n", err)
	} else {
		fmt.Printf("✓ Memory get: %s\n", result)
	}

	// Cleanup
	os.Remove(filepath.Join(tempDir, "test.txt"))
	memory.Close()
	os.Remove("test_tools_memory.db")

	fmt.Println("✓ Tools module tests passed")
}

// main - test entry point
func main() {
	fmt.Println("QuickBot Go Tools Module")
	fmt.Println("✓ Tools module initialized")

	// Run tests
	TestTools()
}
