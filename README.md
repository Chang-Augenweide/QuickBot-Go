# QuickBot-Go 🐹

> **Go 语言版本的轻量级个人 AI 助手框架**  
> Python 版本: [QuickBot-Python](https://github.com/Chang-Augenweide/QuickBot-Python)

<div align="center">

**一个轻量级、模块化、可扩展的个人 AI 助理框架**

[![Go](https://img.shields.io/badge/Go-1.22+-cyan.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## ✨ 特性

QuickBot-Go 是 QuickBot 框架的原生 Go 实现，特点是高性能和低资源占用。

### 🎯 核心功能

- **🤖 多 AI 提供商** - 支持 OpenAI、Anthropic、Ollama（本地模型）
- **📱 Telegram 平台** - 完整的 Telegram Bot 集成
- **💾 内存管理** - 会话记忆 + 长期记忆（SQLite）
- **⏰ 任务调度** - 支持一次性任务和周期性任务（Cron 表达式）
- **🔧 工具系统** - 内置文件、Shell、计算工具，支持自定义扩展
- **🚀 高性能** - 基于 Go 的高并发、低内存占用设计

---

## 📊 性能对比

| 指标 | Go 版本 | Python 版本 |
|------|---------|-------------|
| **内存占用** | ~20 MB | ~50 MB |
| **响应时间** | < 0.5s | < 1s |
| **并发能力** | 500+ 会话 | 100+ 会话 |

---

## 🚀 快速开始

### 系统要求

- Go 1.22+
- SQLite 3

### 安装与运行

```bash
# 1. 克隆仓库
git clone https://github.com/Chang-Augenweide/QuickBot-Go.git
cd QuickBot-Go

# 2. 下载依赖
go mod download

# 3. 配置启动
go run cmd/quickbot/main.py --cmd init
nano config.yaml

# 4. 运行 QuickBot
go run cmd/quickbot/main.py --cmd run
```

### 🎉 构建可执行文件

```bash
# 编译
go build -o quickbot cmd/quickbot/main.py

# 运行
./quickbot
```

---

## 📖 配置示例

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

# Telegram 平台
platforms:
  telegram:
    enabled: true
    token: your_telegram_bot_token

# 内存管理
memory:
  enabled: true
  max_messages: 1000
  storage: memory.db

# 任务调度
scheduler:
  enabled: true
  storage: scheduler_db
```

---

## 📚 命令

| 命令 | 说明 |
|------|------|
| `--cmd run` | 运行机器人 |
| `--cmd init` | 初始化配置文件 |
| `--cmd test` | 运行所有模块测试 |
| `--cmd version` | 显示版本信息 |

---

## 🛠️ 自定义工具

创建自定义工具：

```go
package tools

import "quickbot/pkg/types"

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
```

---

## 🔄 相关项目

- **QuickBot-Python**: Python 版本实现
- **QuickBot**: 总体项目主页

---

## 📄 许可证

MIT License - see [LICENSE](LICENSE)

---

## 🙏 致谢

感谢以下开源项目：

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [robfig/cron](https://github.com/robfig/cron)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

<div align="center">

**⭐ 如果这个项目对你有帮助，请给个 Star！**

Made with ❤️ by [Chang-Augenweide](https://github.com/Chang-Augenweide)

</div>
