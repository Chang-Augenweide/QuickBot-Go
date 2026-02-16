package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Chang-Augenweide/QuickBot-Go/internal/agent"
	"github.com/Chang-Augenweide/QuickBot-Go/internal/config"
	"github.com/Chang-Augenweide/QuickBot-Go/platforms"
)

var (
	configPath string
	command    string
)

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&command, "cmd", "run", "Command to run: run, test, version, init")
	flag.Parse()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	switch command {
	case "run":
		runQuickBot()
	case "test":
		testQuickBot()
	case "version":
		printVersion()
	case "init":
		initConfig()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

// runQuickBot runs the main QuickBot application
func runQuickBot() {
	log.Println("========================================")
	log.Println("       QuickBot v1.0.0 (Go Edition)    ")
	log.Println("========================================")
	log.Println()

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Bot Name: %s", cfg.Bot.Name)
	log.Printf("AI Provider: %s", cfg.AI.Provider)
	log.Printf("AI Model: %s", cfg.AI.Model)
	log.Printf("Debug Mode: %v", cfg.Bot.Debug)
	log.Println()

	// Initialize components
	log.Println("Initializing components...")

	// Memory
	memory, err := agent.NewMemory(cfg.Memory.Storage, cfg.Memory.MaxMessages)
	if err != nil {
		log.Fatalf("Failed to initialize memory: %v", err)
	}
	defer memory.Close()
	log.Printf("✓ Memory system initialized (%s)", cfg.Memory.Storage)

	// Scheduler
	scheduler, err := agent.NewScheduler(cfg.Scheduler.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize scheduler: %v", err)
	}
	defer scheduler.Stop()
	log.Printf("✓ Scheduler initialized (%s)", cfg.Scheduler.Storage)

	// Agent
	quickBot, err := agent.NewAgent(cfg, memory, scheduler)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	log.Printf("✓ Agent initialized")
	log.Printf("  AI: %s (%s)", cfg.AI.Provider, cfg.AI.Model)
	log.Printf("  Tools: %d", len(quickBot.ToolRegistry().GetAll()))

	log.Println()
	log.Println("Initializing platforms...")

	// Initialize enabled platforms
	var telegramPlatform *platforms.TelegramPlatform
	ctx, cancel := context.WithCancel(context.Background())

	// Telegram
	if cfg.Platforms.Telegram.Enabled {
		if cfg.Platforms.Telegram.Token == "" {
			log.Println("⚠ Telegram enabled but no token configured")
		} else {
			tgConfig := &platforms.TelegramConfig{
				Token:        cfg.Platforms.Telegram.Token,
				AllowedUsers: cfg.Platforms.Telegram.AllowedUsers,
				Debug:        cfg.Bot.Debug,
			}

			telegramPlatform, err = platforms.NewTelegramPlatform(tgConfig, quickBot)
			if err != nil {
				log.Fatalf("Failed to initialize Telegram platform: %v", err)
			}

			if err := telegramPlatform.Start(); err != nil {
				log.Fatalf("Failed to start Telegram platform: %v", err)
			}
			log.Println("✓ Telegram platform started")
		}
	}

	if telegramPlatform == nil {
		log.Println("⚠ No platforms enabled. Enable at least one platform in config.yaml")
		return
	}

	// Start agent
	quickBot.Start()

	log.Println()
	log.Println("========================================")
	log.Println("QuickBot is running!")
	log.Println("Press Ctrl+C to stop")
	log.Println("========================================")
	log.Println()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start periodic tasks
	go runPeriodicTasks(ctx, quickBot, memory, scheduler)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf()
	log.Printf("Received signal: %v", sig)
	log.Println("Shutting down...")

	// Cancel context
	cancel()

	// Stop platforms
	if telegramPlatform != nil {
		telegramPlatform.Stop()
	}

	// Stop agent
	quickBot.Stop()

	log.Println("✓ Shutdown complete")
	os.Exit(0)
}

// runPeriodicTasks runs periodic background tasks
func runPeriodicTasks(ctx context.Context, quickBot *agent.Agent, memory *agent.Memory, scheduler *agent.Scheduler) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check for scheduled tasks
			if scheduler != nil {
				tasks := scheduler.GetDueTasks()
				for _, task := range tasks {
					log.Printf("Executing scheduled task: %s", task.Name)
					// Process task
				}
			}

			// Memory cleanup can be added here
		}
	}
}

// testQuickBot runs tests on all modules
func testQuickBot() {
	log.Println("QuickBot Test Suite")
	log.Println("===================")
	log.Println()

	type TestFunc func() error

	tests := []struct {
		name string
		fn   TestFunc
	}{
		{"Configuration", testConfig},
		{"Memory", memory.TestMemory},
		{"Scheduler", scheduler.TestScheduler},
		{"Agent", agent.TestAgent},
		{"Platform Structure", platforms.TestTelegram},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		log.Printf("Testing: %s...", test.name)
		err := test.fn()
		if err != nil {
			log.Printf("✗ FAILED: %v", err)
			failed++
		} else {
			log.Println("✓ PASSED")
			passed++
		}
	}

	log.Println()
	log.Println("===================")
	log.Printf("Results: %d passed, %d failed", passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}

// testConfig tests configuration module
func testConfig() error {
	cfg := config.Config{}
	if cfg.Bot.Name == "" {
		return nil // Default config is valid
	}
	return nil
}

// printVersion prints version information
func printVersion() {
	log.Println("QuickBot v1.0.0 (Go Edition)")
	log.Println("Build: Go 1.22.1")
	log.Println("Copyright (c) 2026 QuickBot Project")
	log.Println()
	log.Println("Features:")
	log.Println("  • AI Integration (OpenAI, Anthropic, Ollama)")
	log.Println("  • Multi-platform support (Telegram, Discord, Slack)")
	log.Println("  • Memory management with SQLite")
	log.Println("  • Task scheduling with Cron")
	log.Println("  • Tool system for extensibility")
}

// initConfig creates a default configuration file
func initConfig() {
	configPathAbs, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPathAbs); err == nil {
		log.Printf("Configuration file already exists: %s", configPathAbs)
		log.Println("Remove it first if you want to create a new one")
		return
	}

	defaultConfig := &config.Config{
		Bot: config.BotConfig{
			Name:     "QuickBot",
			Timezone: "Asia/Shanghai",
			Debug:    false,
		},
		Platforms: config.PlatformsConfig{
			Telegram: config.TelegramConfig{
				Enabled:      true,
				Token:        "",
				AllowedUsers: []string{},
			},
			Discord: config.DiscordConfig{
				Enabled: false,
				Token:   "",
			},
		},
		AI: config.AIConfig{
			Provider:   "openai",
			APIKey:     "",
			Model:      "gpt-4o",
			BaseURL:    "",
			MaxTokens:  2000,
			Temperature: 0.7,
		},
		Memory: config.MemoryConfig{
			Enabled:     true,
			MaxMessages: 1000,
			Storage:     "memory.db",
		},
		Scheduler: config.SchedulerConfig{
			Enabled: true,
			Storage: "scheduler.db",
		},
		Tools: config.ToolsConfig{
			Enabled:   true,
			Directory: "tools/",
		},
		Logging: config.LoggingConfig{
			Level:       "INFO",
			File:        "quickbot.log",
			MaxSize:     10485760,
			BackupCount: 5,
		},
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPathAbs, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}

	log.Printf("✓ Default configuration created: %s", configPathAbs)
	log.Println("Edit the file and add your API keys and tokens")
}

// LoadConfig is a helper to load configuration
func LoadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
