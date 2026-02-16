# QuickBot-Go 🐹

> **轻量级个人 AI 助手框架（高性能）** | [Python 版本](https://github.com/Chang-Augenweide/QuickBot-Python)

<div align="center">

**高性能、低资源占用的个人 AI 助理框架**

[![Go](https://img.shields.io/badge/Go-1.22+-cyan.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Open Issues](https://img.shields.io/github/issues-raw/Chang-Augenweide/QuickBot-Go)](https://github.com/Chang-Augenweide/QuickBot-Go/issues)
[![Repository Size](https://img.shields.io/github/repo-size/Chang-Augenweide/QuickBot-Go)](https://github.com/Chang-Augenweide/QuickBot-Go)

</div>

---

## ✨ 特性

QuickBot-Go 是 QuickBot 框架的原生 Go 实现，专为高并发和低资源占用设计。

### 🎯 核心优势

- **🚀 高性能** - 高并发处理能力，毫秒级响应
- **💾 低内存占用** - 约 20MB 内存占用（对比 Python 版本 50MB）
- **⚡ 零依赖编译** - 单个可执行文件，无需运行时环境
- **📱 Telegram 完整支持** - 实时消息处理、命令路由、Markdown 格式

### 🛠️ 功能完整

- **🤖 多 AI 提供商** - OpenAI、Anthropic、Ollama（原生 API 调用）
- **💾 智能内存管理** - SQLite 持久化，支持会话记忆
- **⏰ 任务调度系统** - Cron 表达式，精确到秒的定时任务
- **🔧 工具系统** - 模块化设计，易于扩展

---

## 🏗️ 系统架构

```
QuickBot-Go 架构
        │
        ├─ 核心层 (internal/)
        │   ├─ agent/           - 核心代理逻辑
        │   ├─ memory/          - 内存管理
        │   ├─ scheduler/       - 任务调度
        │   ├─ tools/           - 工具系统
        │   └─ config/          - 配置管理
        │
        ├─ AI 层 (internal/ai/)
        │   ├─ openai.go        - OpenAI Provider
        │   ├─ anthropic.go     - Anthropic Provider
        │   └─ ollama.go        - Ollama Provider
        │
        └─ 平台层 (platforms/)
            └─ telegram.go      - Telegram 适配器
```

---

## 📊 性能对比

| 指标 | QuickBot-Go | QuickBot-Python |
|------|------------|-----------------|
| **内存占用** | ~20 MB | ~50 MB |
| **响应时间** | < 300ms | < 500ms |
| **并发能力** | 500+ 会话 | 100+ 会话 |
| **CPU 使用** | 低 | 中等 |
| **部署复杂度** | 单文件 | 需要 Python 环境 |
| **启动速度** | < 50ms | ~1s |

---

## 🚀 快速开始

### 环境要求

- Go 1.22 或更高版本
- SQLite 3

### 安装步骤

```bash
# 1. 克隆仓库
git clone https://github.com/Chang-Augenweide/QuickBot-Go.git
cd QuickBot-Go

# 2. 下载依赖
go mod download

# 3. 初始化配置
go run cmd/quickbot/main.go --cmd init
nano config.yaml

# 4. 运行 QuickBot
go run cmd/quickbot/main.go
```

### 构建可执行文件

```bash
# 编译
go build -o quickbot cmd/quickbot/main.go

# 运行
./quickbot
```

跨平台编译：

```bash
# macOS
GOOS=darwin GOARCH=amd64 go build -o quickbot-mac cmd/quickbot/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o quickbot.exe cmd/quickbot/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o quickbot-linux cmd/quickbot/main.go
```

---

## 📚 配置示例

```yaml
# Bot 基本信息
bot:
  name: QuickBot-Go
  debug: false
  timezone: Asia/Shanghai

# AI 提供商配置
ai:
  provider: openai
  api_key: your_api_key_here
  model: gpt-4o
  base_url: https://api.openai.com/v1
  max_tokens: 2000
  temperature: 0.7

# Telegram 平台
platforms:
  telegram:
    enabled: true
    token: your_telegram_bot_token
    allowed_users: []  # 为空则允许所有用户

# 内存管理
memory:
  enabled: true
  max_messages: 1000
  storage: memory.db

# 任务调度
scheduler:
  enabled: true
  storage: scheduler.db
```

---

## 🔧 可用的命令

| 命令 | 说明 |
|------|------|
| `--cmd run` | 运行机器人 |
| `--cmd init` | 初始化配置文件 |
| `--cmd test` | 运行所有模块测试 |
| `--cmd version` | 显示版本信息 |

---

## 🛠️ 自定义工具

创建自定义工具非常简单：

```go
package main

import (
    "quickbot/pkg/types"
)

type CustomTool struct{}

func (t *CustomTool) Name() string {
    return "custom"
}

func (t *CustomTool) Description() string {
    return "自定义工具描述"
}

func (t *CustomTool) Permission() string {
    return "allow_all"
}

func (t *CustomTool) Execute(args map[string]string) (string, error) {
    // 实现你的逻辑
    return "执行结果", nil
}

func (t *CustomTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type": "string",
                "description": "输入参数"
            }
        },
        "required": []string{"input"}
    }
}
```

---

## 📖 API 参考

### Agent 结构体

```go
type Agent struct {
    config    *config.Config
    memory    *memory.Memory
    scheduler *scheduler.Scheduler
    aiProvider aiProvider
    toolRegistry *tools.ToolRegistry
}
```

### AI Provider 接口

```go
type AIProvider interface {
    ChatCompletion(ctx context.Context, messages []Message) (*Response, error)
    StreamChatCompletion(ctx context.Context, messages []Message) (<-chan string, error)
}
```

### Tool 接口

```go
type Tool interface {
    Name() string
    Description() string
    Permission() string
    Execute(args map[string]string) (string, error)
    Parameters() map[string]interface{}
}
```

---

## 🚀 生产部署

### Systemd 服务配置

创建 `/etc/systemd/system/quickbot.service`：

```ini
[Unit]
Description=QuickBot-Go Service
After=network.target

[Service]
Type=simple
User=your_user
WorkingDirectory=/opt/quickbot-go
ExecStart=/opt/quickbot-go/quickbot --cmd run
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable quickbot
sudo systemctl start quickbot
sudo systemctl status quickbot
```

### Docker 部署

**Dockerfile:**

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o quickbot cmd/quickbot/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/quickbot .
COPY --from=builder /app/config.yaml .
CMD ["./quickbot", "--cmd", "run"]
```

构建和运行：

```bash
docker build -t quickbot-go .
docker run -d --name quickbot -v $(pwd)/config.yaml:/root/config.yaml quickbot-go
```

---

## 🔧 故障排除

### 问题：Go 版本不兼容

```bash
# 检查 Go 版本
go version

# 安装最新版 Go
# Linux
wget https://go.dev/dl/go1.22.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 问题：依赖下载失败

```bash
# 配置 Go 代理（中国大陆用户）
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

### 问题：数据库锁定

```bash
# 备份数据库
cp memory.db memory.db.bak

# 删除数据库文件（会清空内存）
rm memory.db

# 重启 QuickBot
./quickbot --cmd run
```

---

## 📊 监控和日志

QuickBot-Go 提供详细的日志输出：

```log
2026/02/16 12:00:00 ========================================
2026/02/16 12:00:00        QuickBot v1.0.0 (Go Edition)
2026/02/16 12:00:00 ========================================
2026/02/16 12:00:00
2026/02/16 12:00:00 Bot Name: QuickBot-Go
2026/02/16 12:00:00 AI Provider: openai
2026/02/16 12:00:00 AI Model: gpt-4o
2026/02/16 12:00:00
2026/02/16 12:00:00 ✓ Memory system initialized (memory.db)
2026/02/16 12:00:00 ✓ Scheduler initialized (scheduler.db)
2026/02/16 12:00:00 ✓ Agent initialized
2026/02/16 12:00:00   AI: openai (gpt-4o)
2026/02/16 12:00:00   Tools: 15
2026/02/16 12:00:00
2026/02/16 12:00:00 ✓ Telegram platform started
2026/02/16 12:00:00
2026/02/16 12:00:00 ========================================
2026/02/16 12:00:00 QuickBot is running!
2026/02/16 12:00:00 Press Ctrl+C to stop
2026/02/16 12:00:00 ========================================
```

---

## 📚 文档

- **[API 文档](docs/API.md)** - 详细的 API 参考
- **[部署指南](docs/DEPLOYMENT.md)** - 生产环境部署
- **[贡献指南](CONTRIBUTING.md)** - 如何贡献代码

---

## 🤝 贡献

欢迎贡献代码！

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码规范

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 使用 `gofmt` 格式化代码
- 为公开的函数添加注释

---

## 📝 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解详细的版本更新。

---

## 🔗 相关项目

- **[QuickBot-Python](https://github.com/Chang-Augenweide/QuickBot-Python)** - Python 语言版本实现
- **QuickBot 原版** - 已归档，请使用新仓库

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。

---

## 🙏 致谢

感谢以下开源项目：

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [robfig/cron](https://github.com/robfig/cron)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

---

<div align="center">

**🚀 体验高性能的 AI 助手！**

如果这个项目对你有帮助，请给个 ⭐ Star！

Made with ❤️ by [Chang-Augenweide](https://github.com/Chang-Augenweide)

</div>
