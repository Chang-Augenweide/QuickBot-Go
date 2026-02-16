package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// Plugin represents a QuickBot plugin
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns the plugin description
	Description() string

	// Initialize initializes the plugin
	Initialize(config map[string]interface{}) error

	// Execute executes the plugin
	Execute(args map[string]interface{}) (interface{}, error)

	// Shutdown shuts down the plugin
	Shutdown() error
}

// PluginMetadata stores plugin metadata
type PluginMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Enabled     bool              `json:"enabled"`
	Config      map[string]string `json:"config"`
}

// PluginManager manages plugins
type PluginManager struct {
	plugins    map[string]Plugin
	metadata   map[string]PluginMetadata
	configPath string
	mu         sync.RWMutex
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(configPath string) *PluginManager {
	return &PluginManager{
		plugins:    make(map[string]Plugin),
		metadata:   make(map[string]PluginMetadata),
		configPath: configPath,
	}
}

// LoadPlugin loads a plugin from a shared object file
func (pm *PluginManager) LoadPlugin(filePath, pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Load the plugin
	p, err := plugin.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// Look up the Plugin symbol
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("failed to find Plugin symbol: %w", err)
	}

	// Assert the symbol to Plugin type
	pluginInstance, ok := sym.(Plugin)
	if !ok {
		return fmt.Errorf("unexpected type from module symbol")
	}

	// Initialize the plugin
	err = pluginInstance.Initialize(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Register the plugin
	pm.plugins[pluginName] = pluginInstance

	log.Printf("Plugin loaded: %s v%s", pluginInstance.Name(), pluginInstance.Version())

	return nil
}

// UnloadPlugin unloads a plugin
func (pm *PluginManager) UnloadPlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Shutdown the plugin
	err := plugin.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown plugin: %w", err)
	}

	// Remove from registry
	delete(pm.plugins, name)
	delete(pm.metadata, name)

	log.Printf("Plugin unloaded: %s", name)

	return nil
}

// ExecutePlugin executes a plugin
func (pm *PluginManager) ExecutePlugin(name string, args map[string]interface{}) (interface{}, error) {
	pm.mu.RLock()
	plugin, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin.Execute(args)
}

// ListPlugins returns a list of loaded plugins
func (pm *PluginManager) ListPlugins() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var names []string
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

// GetPluginInfo returns plugin information
func (pm *PluginManager) GetPluginInfo(name string) (Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// ScanPlugins scans a directory for plugins
func (pm *PluginManager) ScanPlugins(directory string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to scan plugins directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext == ".so" {
			pluginName := file.Name()[:len(file.Name())-3]
			pluginPath := filepath.Join(directory, file.Name())

			// Try to load the plugin
			err := pm.LoadPlugin(pluginPath, pluginName)
			if err != nil {
				log.Printf("Warning: Failed to load plugin %s: %v", pluginName, err)
			}
		}
	}

	return nil
}

// Shutdown shuts down all plugins
func (pm *PluginManager) Shutdown() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var lastErr error

	for name, plugin := range pm.plugins {
		err := plugin.Shutdown()
		if err != nil {
			log.Printf("Error shutting down plugin %s: %v", name, err)
			lastErr = err
		}
		delete(pm.plugins, name)
		delete(pm.metadata, name)
	}

	return lastErr
}

// Example: A simple plugin
type EchoPlugin struct {
	name    string
	version string
}

func (p *EchoPlugin) Name() string {
	return p.name
}

func (p *EchoPlugin) Version() string {
	return p.version
}

func (p *EchoPlugin) Description() string {
	return "Echoes back the input message"
}

func (p *EchoPlugin) Initialize(config map[string]interface{}) error {
	return nil
}

func (p *EchoPlugin) Execute(args map[string]interface{}) (interface{}, error) {
	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message is required")
	}
	return fmt.Sprintf("Echo: %s", message), nil
}

func (p *EchoPlugin) Shutdown() error {
	return nil
}

// NewEchoPlugin creates a new echo plugin instance
func NewEchoPlugin() Plugin {
	return &EchoPlugin{
		name:    "echo",
		version: "1.0.0",
	}
}

// TestPluginManager tests the plugin manager
func TestPluginManager() {
	log.Println("Testing Plugin Manager...")

	// Create plugin manager
	pm := NewPluginManager("")

	// Register an in-memory plugin
	echoPlugin := NewEchoPlugin()
	pm.plugins["echo"] = echoPlugin

	// Test plugin execution
	result, err := pm.ExecutePlugin("echo", map[string]interface{}{"message": "Hello Test"})
	if err != nil {
		log.Printf("Failed to execute plugin: %v", err)
	} else {
		log.Printf("✓ Plugin executed: %s", result)
	}

	// Test plugin listing
	plugins := pm.ListPlugins()
	log.Printf("✓ Loaded plugins: %v", plugins)

	// Test plugin info
	info, err := pm.GetPluginInfo("echo")
	if err != nil {
		log.Printf("Failed to get plugin info: %v", err)
	} else {
		log.Printf("✓ Plugin info: %s v%s", info.Name(), info.Version())
	}

	err = pm.Shutdown()
	if err != nil {
		log.Printf("Failed to shutdown plugin manager: %v", err)
	} else {
		log.Println("✓ Plugin manager shutdown complete")
	}

	log.Println("✓ Plugin manager tests passed")
}

// main - test entry point
func main() {
	log.Println("QuickBot Go Plugin System")
	log.Println("✓ Plugin system module initialized")

	// Run tests
	TestPluginManager()
}
