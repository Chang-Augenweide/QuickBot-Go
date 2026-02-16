# QuickBot - Personal AI Assistant Framework

QuickBot是一个功能完备的个人AI助手框架，支持多云AI提供商、多平台、内存管理、任务调度和工具系统。

## 技术栈

- **Python 3**: 完整的服务器端实现，生产就绪
- **Go 1.22**: 高性能模块，提供核心功能的并行实现

## 架构设计

```
QuickBot/
├── agent.go              # Go - 核心Agent逻辑
├── ai_providers.py       # Python - AI提供商集成
├── config.go/config.py   # 配置管理
├── memory.go/memory.py   # 内存管理 (SQLite)
├── scheduler.go/scheduler.py  # 任务调度
├── tools.py/tools.go     # 工具系统
├── main.py               # 主程序入口
├── types.go              # Go公共类型定义
├── config.yaml           # 配置文件
└── platforms/            # 平台适配器
    ├── telegram.py
    └── telegram_platform.go
```

## 核心功能

### 1. AI集成
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- Ollama (本地开源模型)
- 百度文心一言
- 其他OpenAI兼容API

### 2. 多平台支持
- Telegram (完全支持)
- Discord (框架就绪)
- Slack (框架就绪)
- 微信 (规划中)

### 3. 内存管理
- 会话记忆 - 短期对话上下文
- 长期记忆 - 持久化关键信息
- 上下文窗口 - 智能信息检索
- 向量存储(规划中) - 语义检索

### 4. 任务调度
- 一次性任务
- 周期性任务
- 提醒事项
- Cron表达式
- 支持多种任务类型

### 5. 工具系统
- 文件操作 (读/写/列表)
- Shell命令执行
- 计算功能
- 内存查询
- 自定义工具支持

## 项目结构规范

遵循阿里巴巴代码规范：

```
QuickBot/
├── api/                  # API接口定义
├── core/                 # 核心模块
├── internal/             # 内部包
│   ├── ai/              # AI集成
│   ├── memory/          # 内存管理
│   ├── scheduler/       # 任务调度
│   └── tools/           # 工具系统
├── pkg/                  # 公共包
├── platforms/            # 平台适配器
├── configs/              # 配置文件
├── docs/                 # 文档
├── tests/                # 测试
├── examples/             # 示例代码
├── scripts/              # 脚本工具
└── README.md
```

## 快速开始

### Python版本

```bash
# 1. 安装依赖
pip install -r requirements.txt

# 2. 配置设置
cp config.example.yaml config.yaml
nano config.yaml  # 编辑配置

# 3. 运行QuickBot
python main.py
```

### Go版本

```bash
# 1. 下载依赖
go mod download

# 2. 配置设置
cp config.example.yaml config.yaml
nano config.yaml  # 编辑配置

# 3. 运行QuickBot
go build -o quickbot
./quickbot
```

## 配置说明

### AI配置

```yaml
ai:
  provider: openai
  api_key: your_api_key_here
  model: gpt-4o
  base_url: https://api.openai.com/v1
  max_tokens: 2000
  temperature: 0.7
```

### 平台配置

```yaml
platforms:
  telegram:
    enabled: true
    token: your_bot_token
    allowed_users:
      - user1
      - user2
```

### 内存配置

```yaml
memory:
  enabled: true
  max_messages: 1000
  storage: memory.db
```

### 调度器配置

```yaml
scheduler:
  enabled: true
  storage: scheduler.db
```

## 测试

### Python测试

```bash
python -m pytest tests/
```

### Go测试

```bash
go test ./...
```

### 单个模块测试

```bash
# Python
python agent.py

# Go
go run config.go
go run memory.go
go run scheduler_main.go
go run tools_simple.go
go run types.go agent_simple.go
```

## 开发路线图

### Phase 1: 核心 (完成)
- [x] AI集成 (OpenAI, Anthropic, Ollama)
- [x] 多平台框架
- [x] 内存管理
- [x] 任务调度
- [x] 工具系统

### Phase 2: 增强进行中
- [x] Go模块开发
- [x] 配置管理
- [x] 内存管理
- [x] 任务调度
- [x] 工具系统
- [x] Agent逻辑
- [ ] 平台适配器优化
- [ ] 性能优化

### Phase 3: 高级功能（计划）
- [ ] 向量数据库集成
- [ ] 多模态支持 (图像、语音)
- [ ] 工作流编排
- [ ] 插件系统完善
- [ ] Web管理界面
- [ ] 监控和分析

## 性能指标

- 内存占用: < 50MB (Python), < 20MB (Go)
- 响应时间: < 1s (平均)
- 并发处理: 支持100+会话
- 内存容量: > 10,000条消息

## 安全特性

- API密钥加密存储
- 用户验证和授权
- 命令白名单机制
- 路径访问限制
- 日志审计追踪

## 贡献指南

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT License

## 联系方式

- 项目主页: https://github.com/Chang-Augenweide/QuickBot
- 问题反馈: https://github.com/Chang-Augenweide/QuickBot/issues

---

**QuickBot - 打造最强大的个人AI助手框架**
