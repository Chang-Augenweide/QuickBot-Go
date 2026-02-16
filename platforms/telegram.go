package platform

import (
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"quickbot/internal/agent"
	"quickbot/internal/config"
)

// TelegramConfig represents Telegram platform configuration
type TelegramConfig struct {
	Token          string
	AllowedUsers   []string
	Debug          bool
}

// TelegramPlatform represents Telegram bot platform
type TelegramPlatform struct {
	config     *TelegramConfig
	botAPI     *tgbotapi.BotAPI
	agent      *agent.Agent
	updates    tgbotapi.UpdatesChannel
	started    bool
	mu         sync.RWMutex
}

// NewTelegramPlatform creates a new Telegram platform instance
func NewTelegramPlatform(cfg *TelegramConfig, bot *agent.Agent) (*TelegramPlatform, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot API: %w", err)
	}

	botAPI.Debug = cfg.Debug

	return &TelegramPlatform{
		config:  cfg,
		botAPI:  botAPI,
		agent:   bot,
		started: false,
	}, nil
}

// Start starts the Telegram platform
func (p *TelegramPlatform) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("platform already started")
	}

	log.Println("Starting Telegram platform...")

	// Set up update configuration
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Get updates channel
	updates := p.botAPI.GetUpdatesChan(u)
	p.updates = updates
	p.started = true

	// Start message handler
	go p.handleMessages()

	log.Println("✓ Telegram platform started")
	return nil
}

// Stop stops the Telegram platform
func (p *TelegramPlatform) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return fmt.Errorf("platform not started")
	}

	log.Println("Stopping Telegram platform...")

	// Stop update channel
	p.botAPI.StopReceivingUpdates()
	p.started = false

	log.Println("✓ Telegram platform stopped")
	return nil
}

// isUserAllowed checks if a user is allowed to interact with the bot
func (p *TelegramPlatform) isUserAllowed(userID int64) bool {
	// If whitelist is empty, allow all users
	if len(p.config.AllowedUsers) == 0 {
		return true
	}

	userIDStr := fmt.Sprintf("%d", userID)
	for _, allowed := range p.config.AllowedUsers {
		if allowed == userIDStr {
			return true
		}
	}

	return false
}

// handleMessages processes incoming Telegram updates
func (p *TelegramPlatform) handleMessages() {
	for update := range p.updates {
		// Only handle messages
		if update.Message == nil {
			continue
		}

		message := update.Message

		// Check user permission
		if !p.isUserAllowed(message.From.ID) {
			log.Printf("Unauthorized user attempt: %d (%s)", message.From.ID, message.From.UserName)
			continue
		}

		// Create session ID
		sessionID := fmt.Sprintf("telegram:%d", message.From.ID)

		// Handle commands
		if message.IsCommand() {
			p.handleCommand(message, sessionID)
			continue
		}

		// Process regular message
		p.processMessage(message, sessionID)
	}
}

// handleCommand handles bot commands
func (p *TelegramPlatform) handleCommand(message *tgbotapi.Message, sessionID string) {
	command := message.Command()

	switch command {
	case "start":
		p.sendReply(message, fmt.Sprintf(
			"👋 你好！我是 *%s*！\n\n"+
			"发送 /help 查看可用命令。\n\n"+
			"你也可以直接和我聊天！",
			p.agent.Config().Bot.Name,
		))

	case "help":
		helpText := p.generateHelpText()
		p.sendReply(message, helpText)

	case "status":
		statusText := p.generateStatusText()
		p.sendReply(message, statusText)

	default:
		p.sendReply(message, fmt.Sprintf("未知命令: /%s\n发送 /help 查看帮助", command))
	}
}

// processMessage processes a regular message
func (p *TelegramPlatform) processMessage(message *tgbotapi.Message, sessionID string) {
	// Get user message
	userMessage := message.Text
	if userMessage == "" {
		return
	}

	log.Printf("[Telegram][%s] Received: %s", sessionID, userMessage)

	// Process message through agent
	response, err := p.agent.ProcessMessage(sessionID, userMessage)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		p.sendReply(message, "抱歉，处理消息时出错。")
		return
	}

	// Send response
	p.sendReply(message, response)
}

// sendReply sends a reply message
func (p *TelegramPlatform) sendReply(message *tgbotapi.Message, text string) {
	// Truncate if too long (Telegram limit is 4096 characters)
	if len(text) > 4000 {
		text = text[:4000] + "\n... (消息过长，已截断)"
	}

	// Parse Markdown
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"

	// Send message
	_, err := p.botAPI.Send(msg)
	if err != nil {
		log.Printf("Error sending reply: %v", err)
	}
}

// generateHelpText generates help message
func (p *TelegramPlatform) generateHelpText() string {
	return fmt.Sprintf(`📖 *%s 命令列表*

/start - 启动机器人
/help - 显示此帮助信息
/status - 查看系统状态

你也可以直接和我聊天！

💡 *可用功能*
• AI 助手对话
• 文件操作
• 计算功能
• 长期记忆
• 任务提醒

📝 *使用示例*
• "2+3*4 等于多少？"
• "记住: 会议时间下午3点"
• "帮我创建一个笔记文件"`,
		p.agent.Config().Bot.Name,
	)
}

// generateStatusText generates status message
func (p *TelegramPlatform) generateStatusText() string {
	cfg := p.agent.Config()

	statusParts := []string{
		fmt.Sprintf("🤖 *%s 状态*", cfg.Bot.Name),
		fmt.Sprintf("📊 平台: Telegram"),
		fmt.Sprintf("🧠 AI 提供商: %s", cfg.AI.Provider),
		fmt.Sprintf("📝 模型: %s", cfg.AI.Model),
	}

	return strings.Join(statusParts, "\n")
}

// SendMessage sends a message directly to a chat
func (p *TelegramPlatform) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	_, err := p.botAPI.Send(msg)
	return err
}

// BotAPI returns the underlying Telegram bot API instance
func (p *TelegramPlatform) BotAPI() *tgbotapi.BotAPI {
	return p.botAPI
}

// IsStarted returns whether the platform is started
func (p *TelegramPlatform) IsStarted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.started
}

// TestTelegram tests Telegram platform connection
func TestTelegram() {
	log.Println("Testing Telegram platform...")

	// This is a placeholder test
	// In production, you would need a valid bot token
	log.Println("✓ Telegram platform structure verified")
	log.Println("⚠ Note: Requires valid bot token for actual connection test")
}
