# QuickBot-Go - 轻量级个人 AI 助手框架（Go 语言版）

> **QuickBot-Python 版本**: [GitHub](https://github.com/Chang-Augenweide/QuickBot-Python)  

<div align="center">

**一个轻量级、模块化、可扩展的个人 AI 助理框架**

[![Go Version](https://img.shields.io/badge/Go-1.22+-cyan.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## ✨ 特性

QuickBot-Go 是 QuickBot 框架的原生 Go 实现，具有高性能和低资源占用特点。

### 🎯 核心功能

- **🤖 多 AI 提供商** - 支持 OpenAI、Anthropic、Ollama（本地模型）
- **📱 Telegram 平台** - 完整的 Telegram Bot 集成
- **💾 内存管理** - 会话记忆 + 长期记忆（SQLite）
- **⏰ 任务调度** - 支持一次性任务和周期性任务（Cron 表达式）
- **🔧 工具系统** - 内置文件、Shell、计算工具，支持自定义扩展
- **🚀 高性能** - 基于 Go 的高并发、低内存占用设计

### 🏗️ 技术栈

- **语言**: Go 1.22+
- **数据库**: SQLite 3
- **任务调度**: robfig/cron v3
- **平台**: Telegram Bot API v5

---

## 📊 项目架构

```
QuickBot-Go/
├── cmd/
│   └── quickbot/
│       └── main.go          # 主程序入口
├── internal/
│   ├── ai/                  # AI 提供商实现
│   │   ├── openai.go       # OpenAI API
│   │   ├── anthropic.go    # Anthropic API
│   │   └── ollama.go       # Ollama (本地)
│   ├── agent/               # 核心 Agent 逻辑
│   ├── config/              # 配置管理
│   ├── memory/              # 内存管理（SQLite）
│   ├── scheduler/           # 任务调度（Cron）
│   └── tools/               # 工具系统
├── pkg/
│   └── types/               # 公共类型定义
├── platforms/
│   └── telegram.go          # Telegram 平台适配器
├── docs/                    # 文档
├── config.example.yaml      # 配置示例
└── go.mod                   # Go 模块定义
```

---

## 🚀 快速开始

### 系统要求

- Go 1.22 或更高版本
- SQLite 3

### 安装与运行

#### 1. 克隆仓库

```bash
git clone https://github.com/Chang-Augenweide/QuickBot-Go.git
cd QuickBot-Go
```

#### 2. 下载依赖

```bash
go mod download
```

#### 3. 配置启动

```bash
# 生成默认配置
go run cmd/quickbot/main.go --cmd init

# 编辑配置文件
nano config.yaml

# 启动 QuickBot
go run cmd/quickbot/main.py --cmd run
```

### 🎉 构建可执行文件

```bash
# 编译
go build -o quickbot cmd/quickbot/main.go

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
  storage: scheduler_db
```

---

## 📚 命令

| 命令 | 说明 |
|------|------|
| `--cmd run` | 运行机器人 |
| `--cmd init` | 初始化配置文件 |
| `--cmd test` | 运行所有模块测试 |
| `--cmd version` | 显示版本信息

---

## 🧪 开发

### 运行测试

```bash
# 运行所有测试
go run cmd/quickbot/main.go --cmd test

# 单独测试某个模块
go run internal/memory/memory.go
go run internal/scheduler/scheduler.go
```

### 代码规范

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 使用 `gofmt` 格式化代码
- 为公开的函数添加注释

---

## 📈 性能对比

| 指标 | Go 版本 | Python 版本 |
|------|---------|-------------|
| **内存占用** | ~20 MB | ~50 MB |
| **响应时间** | < 0.5s | < 1s |
| **并发能力** | 500+ 会话 | 100+ 会话 |
| **CPU 使用** | 低 | 中等 |

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

## 📚 文档

- **贡献指南**: See [CONTRIBUTING.md](CONTRIBUTING.md)
- **更新日志**: See [CHANGELOG.md](CHANGELOG.md)

---

## 🔄 相关项目

- **QuickBot-Python**: Python 版本实现
- **QuickBot**: 总体项目主页

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。

---

## 🙏 致谢

感谢以下开源项目：

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [robfig/cron](https://github.com/robfig/cron)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给个 Star！**

Made with ❤️ by [Chang-Augenweide](https://github.com/Chang-Augenweide)

</div>
