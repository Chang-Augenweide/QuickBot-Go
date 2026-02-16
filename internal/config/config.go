package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Bot        BotConfig        `yaml:"bot"`
	Platforms  PlatformsConfig  `yaml:"platforms"`
	AI         AIConfig         `yaml:"ai"`
	Memory     MemoryConfig     `yaml:"memory"`
	Scheduler  SchedulerConfig  `yaml:"scheduler"`
	Tools      ToolsConfig      `yaml:"tools"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// BotConfig represents bot-specific configuration
type BotConfig struct {
	Name     string `yaml:"name"`
	Debug    bool   `yaml:"debug"`
	Timezone string `yaml:"timezone"`
}

// PlatformsConfig represents platform integrations
type PlatformsConfig struct {
	Telegram TelegramConfig `yaml:"telegram"`
	Discord  DiscordConfig  `yaml:"discord"`
}

// TelegramConfig represents Telegram bot configuration
type TelegramConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Token        string   `yaml:"token"`
	AllowedUsers []string `yaml:"allowed_users"`
}

// DiscordConfig represents Discord bot configuration
type DiscordConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"`
}

// AIConfig represents AI provider configuration
type AIConfig struct {
	Provider    string  `yaml:"provider"`
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base_url"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

// MemoryConfig represents memory management configuration
type MemoryConfig struct {
	Enabled     bool   `yaml:"enabled"`
	MaxMessages int    `yaml:"max_messages"`
	Storage     string `yaml:"storage"`
}

// SchedulerConfig represents scheduler configuration
type SchedulerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Storage string `yaml:"storage"`
}

// ToolsConfig represents tools configuration
type ToolsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Directory string `yaml:"directory"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level       string `yaml:"level"`
	File        string `yaml:"file"`
	MaxSize     int64  `yaml:"max_size"`
	BackupCount int    `yaml:"backup_count"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults
	config.applyDefaults()

	return &config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(cfg *Config, path string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// applyDefaults applies default values to configuration
func (c *Config) applyDefaults() {
	// Bot defaults
	if c.Bot.Name == "" {
		c.Bot.Name = "QuickBot"
	}
	if c.Bot.Timezone == "" {
		c.Bot.Timezone = "Asia/Shanghai"
	}

	// AI defaults
	if c.AI.Provider == "" {
		c.AI.Provider = "openai"
	}
	if c.AI.Model == "" {
		c.AI.Model = "gpt-4o"
	}
	if c.AI.MaxTokens == 0 {
		c.AI.MaxTokens = 2000
	}
	if c.AI.Temperature == 0 {
		c.AI.Temperature = 0.7
	}
	if c.AI.BaseURL == "" && c.AI.Provider == "openai" {
		c.AI.BaseURL = "https://api.openai.com/v1"
	}

	// Memory defaults
	if c.Memory.MaxMessages == 0 {
		c.Memory.MaxMessages = 1000
	}
	if c.Memory.Storage == "" {
		c.Memory.Storage = "memory.db"
	}

	// Scheduler defaults
	if c.Scheduler.Storage == "" {
		c.Scheduler.Storage = "scheduler.db"
	}

	// Tools defaults
	if c.Tools.Directory == "" {
		c.Tools.Directory = "tools/"
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "INFO"
	}
	if c.Logging.File == "" {
		c.Logging.File = "quickbot.log"
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = 10 * 1024 * 1024 // 10MB
	}
	if c.Logging.BackupCount == 0 {
		c.Logging.BackupCount = 5
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate bot configuration
	if c.Bot.Name == "" {
		return fmt.Errorf("bot name cannot be empty")
	}

	// Validate AI configuration
	if c.AI.Provider == "" {
		return fmt.Errorf("AI provider cannot be empty")
	}

	if c.AI.Provider == "openai" || c.AI.Provider == "anthropic" {
		if c.AI.APIKey == "" {
			return fmt.Errorf("%s provider requires API key", c.AI.Provider)
		}
	}

	// Validate platform configuration
	if c.Platforms.Telegram.Enabled && c.Platforms.Telegram.Token == "" {
		return fmt.Errorf("telegram enabled but token not configured")
	}
	if c.Platforms.Discord.Enabled && c.Platforms.Discord.Token == "" {
		return fmt.Errorf("discord enabled but token not configured")
	}

	return nil
}

// Config示例
func DefaultConfig() *Config {
	return &Config{
		Bot: BotConfig{
			Name:     "QuickBot",
			Timezone: "Asia/Shanghai",
			Debug:    false,
		},
		Platforms: PlatformsConfig{
			Telegram: TelegramConfig{
				Enabled: true,
			},
		},
		AI: AIConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			MaxTokens:   2000,
			Temperature: 0.7,
			BaseURL:     "https://api.openai.com/v1",
		},
		Memory: MemoryConfig{
			Enabled:     true,
			MaxMessages: 1000,
			Storage:     "memory.db",
		},
		Scheduler: SchedulerConfig{
			Enabled: true,
			Storage: "scheduler.db",
		},
		Tools: ToolsConfig{
			Enabled:   true,
			Directory: "tools/",
		},
		Logging: LoggingConfig{
			Level:       "INFO",
			File:        "quickbot.log",
			MaxSize:     10 * 1024 * 1024,
			BackupCount: 5,
		},
	}
}

func main() {
	log.Println("QuickBot Go Configuration Module")
	log.Println("Bot Name: QuickBot")
	log.Println("✓ Go config module initialized")
}
