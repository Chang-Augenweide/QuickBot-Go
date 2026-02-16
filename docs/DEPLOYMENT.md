# QuickBot部署指南

本文档提供QuickBot在各种环境下的详细部署指南。

## 目录

1. [本地开发环境](#本地开发环境)
2. [生产环境部署](#生产环境部署)
3. [Docker部署](#docker部署)
4. [Kubernetes部署](#kubernetes部署)
5. [监控和维护](#监控和维护)
6. [性能优化](#性能优化)

---

## 本地开发环境

### 系统要求

- Python 3.8+
- Go 1.22+
- SQLite 3
- 4GB+ RAM
- 10GB+ 磁盘空间

### 安装步骤

#### 1. 克隆代码仓库

```bash
git clone https://github.com/Chang-Augenweide/QuickBot.git
cd QuickBot
```

#### 2. 安装Python依赖

```bash
python3 -m pip install -r requirements.txt
python3 -m pip install -r requirements-dev.txt  # 可选：开发工具
```

#### 3. 配置QuickBot

```bash
cp config.example.yaml config.yaml
nano config.yaml  # 编辑配置文件
```

#### 4. 初始化数据库

数据库会在第一次运行时自动创建。

#### 5. 运行QuickBot

```bash
# Python版本
python main.py

# Go版本
go run main.go --cmd run

# 或使用Makefile
make run
```

---

## 生产环境部署

### 系统要求

- Python 3.8+ 或 Go 1.22+
- 8GB+ RAM
- 50GB+ SSD磁盘空间
- 稳定的网络连接
- API密钥（OpenAI/Anthropic等）

### 部署步骤

#### 1. 环境准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装系统依赖
sudo apt install -y python3 python3-pip gcc git sqlite3

# 安装Python依赖
pip install -r requirements.txt
```

#### 2. 配置文件

创建`config.yaml`：

```yaml
bot:
  name: ProductionQuickBot
  debug: false
  timezone: Asia/Shanghai

ai:
  provider: openai
  api_key: YOUR_API_KEY_HERE
  model: gpt-4o

platforms:
  telegram:
    enabled: true
    token: YOUR_TELEGRAM_BOT_TOKEN

memory:
  enabled: true
  max_messages: 10000
  storage: /var/lib/quickbot/memory.db

scheduler:
  enabled: true
  storage: /var/lib/quickbot/scheduler.db

logging:
  level: INFO
  file: /var/log/quickbot/quickbot.log
```

#### 3. 设置系统服务

创建systemd服务文件 `/etc/systemd/system/quickbot.service`：

```ini
[Unit]
Description=QuickBot Personal AI Assistant
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quickbot
ExecStart=/usr/bin/python3 /opt/quickbot/main.py
Restart=always
RestartSec=10
Environment=PYTHONUNBUFFERED=1

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

#### 4. 配置反向代理（可选）

使用Nginx：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Docker部署

### 使用Docker Compose

#### 1. 准备环境变量

```bash
cp .env.example .env
nano .env  # 填写实际值
```

#### 2. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f quickbot

# 查看状态
docker-compose ps
```

#### 3. 服务管理

```bash
# 停止服务
docker-compose stop

# 重启服务
docker-compose restart quickbot

# 停止并删除容器
docker-compose down

# 重新构建镜像
docker-compose build --no-cache
docker-compose up -d
```

### 单独容器部署

```bash
# 构建镜像
docker build -t quickbot:latest .

# 运行容器
docker run -d \
  --name quickbot \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v quickbot_data:/app/data \
  quickbot:latest

# 查看日志
docker logs -f quickbot
```

---

## Kubernetes部署

### 部署清单（Deployment.yaml）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quickbot
  labels:
    app: quickbot
spec:
  replicas: 2
  selector:
    matchLabels:
      app: quickbot
  template:
    metadata:
      labels:
        app: quickbot
    spec:
      containers:
      - name: quickbot
        image: quickbot:latest
        ports:
        - containerPort: 8080
        env:
        - name: QUICKBOT_AI_API_KEY
          valueFrom:
            secretKeyRef:
              name: quickbot-secrets
              key: ai-api-key
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
        - name: data
          mountPath: /app/data
      volumes:
      - name: config
        configMap:
          name: quickbot-config
      - name: data
        persistentVolumeClaim:
          claimName: quickbot-data
---
apiVersion: v1
kind: Service
metadata:
  name: quickbot-service
spec:
  selector:
    app: quickbot
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 部署命令

```bash
# 应用配置
kubectl apply -f deployment.yaml

# 查看Pod状态
kubectl get pods

# 查看日志
kubectl logs -f deployment/quickbot

# 扩容
kubectl scale deployment quickbot --replicas=3
```

---

## 监控和维护

### 1. 日志管理

查看实时日志：

```bash
# systemd
sudo journalctl -u quickbot -f

# Docker
docker logs -f quickbot

# Kubernetes
kubectl logs -f deployment/quickbot
```

### 2. 健康检查

运行健康检查脚本：

```bash
python health_check.py

# 定期检查（cron）
0 * * * * /usr/bin/python3 /opt/quickbot/health_check.py >> /var/log/quickbot/health.log
```

### 3. 性能监控

使用Prometheus + Grafana：

```bash
# 启动监控栈
docker-compose up -d prometheus grafana

# 访问Grafana
http://localhost:3000
```

### 4. 数据库维护

备份数据库：

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
cp /var/lib/quickbot/memory.db /backup/memory_${DATE}.db
cp /var/lib/quickbot/scheduler.db /backup/scheduler_${DATE}.db
# 保留最近7天的备份
find /backup -name "*.db" -mtime +7 -delete
```

### 5. 更新升级

```bash
# 停止服务
sudo systemctl stop quickbot

# 备份数据
./backup.sh

# 拉取最新代码
git pull origin main

# 更新依赖
pip install -r requirements.txt --upgrade

# 更新Go模块
go mod download

# 启动服务
sudo systemctl start quickbot
```

---

## 性能优化

### 1. 内存优化

- 减少 `max_messages` 配置值
- 定期清理旧消息
- 使用连接池

### 2. 数据库优化

```sql
-- 创建索引
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_sessions_updated ON sessions(updated_at);

-- 定期VACUUM
VACUUM;
```

### 3. API缓存

使用Redis缓存常用查询：

```python
import redis

r = redis.Redis(host='localhost', port=6379, decode_responses=True)

# 缓存内存查询
def get_cached_memory(key):
    cached = r.get(f"memory:{key}")
    if cached:
        return cached
    value = memory.get_long_term(key)
    r.setex(f"memory:{key}", 3600, value)  # 缓存1小时
    return value
```

### 4. 异步处理

使用Celery处理耗时任务：

```python
from celery import Celery

app = Celery('quickbot', broker='redis://localhost:6379/0')

@app.task
def async_message_process(session_id, message):
    # 异步处理消息
    pass
```

---

## 故障排除

### 问题：服务无法启动

检查配置文件语法：
```bash
python -c "import yaml; yaml.safe_load(open('config.yaml'))"
```

### 问题：内存占用过高

1. 减少 `memory.max_messages`
2. 定期清理会话
3. 重启服务

### 问题：API响应慢

1. 检查网络连接
2. 增加AI请求超时时间
3. 使用Redis缓存

### 问题：数据库损坏

```bash
# 尝试恢复
sqlite3 memory.db ".recover" | sqlite3 memory_recovered.db

# 从备份恢复
cp /backup/memory_YYYYMMDD_HHMMSS.db /var/lib/quickbot/memory.db
```

---

## 安全建议

1. **API密钥安全**
   - 使用环境变量存储密钥
   - 定期轮换密钥
   - 使用密钥管理服务

2. **访问控制**
   - 限制allowed_users
   - 使用HTTPS
   - 配置防火墙

3. **日志审计**
   - 记录所有操作
   - 定期审查日志
   - 实施日志轮转

4. **数据加密**
   - 加密敏感数据
   - 使用TLS连接
   - 定期备份

---

## 支持

- 项目主页: https://github.com/Chang-Augenweide/QuickBot
- 问题反馈: https://github.com/Chang-Augenweide/QuickBot/issues
- 文档: https://docs.quickbot.ai

---

**Happy Deployment! 🚀**
