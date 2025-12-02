# DuDu Proxy

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**DuDu Proxy** 是一个高性能代理服务器，支持 HTTP 和 SOCKS5 协议，具备认证、IP 封禁、限流和熔断等企业级功能。

[English Documentation](README.md)

## ✨ 功能特性

- **多协议支持**
  - HTTP/HTTPS 代理（包括 CONNECT 隧道）
  - 完整的 SOCKS5 代理协议支持

- **安全与认证**
  - 多用户认证（用户名/密码）
  - IP 封禁机制，可配置失败阈值
  - IP 白名单支持

- **流量控制**
  - 基于 Token Bucket 算法的全局和单 IP 限流
  - 带滑动窗口的熔断器，防止突发大量失败请求
  - 自动恢复机制

- **运维支持**
  - 基于 JSON 的配置管理
  - 结构化日志，支持多种格式（JSON/控制台）
  - 优雅关闭
  - Docker 支持

## 📦 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/seakee/dudu-proxy.git
cd dudu-proxy

# 编译
make build

# 或者手动编译
go build -o build/dudu-proxy cmd/dudu-proxy/main.go
```

### 使用 Docker

```bash
# 构建 Docker 镜像
make docker

# 或使用 docker-compose
docker-compose up -d
```

## 🚀 快速开始

1. **配置代理服务器**

   复制示例配置文件并根据需要修改：
   ```bash
   cp configs/config.example.json configs/config.json
   ```

2. **运行代理服务器**

   ```bash
   # 使用 make
   make run

   # 或直接运行
   ./build/dudu-proxy -config configs/config.json
   ```

3. **测试代理**

   ```bash
   # HTTP 代理
   curl -x http://user1:pass1@localhost:8080 http://www.google.com

   # HTTPS 代理
   curl -x http://user1:pass1@localhost:8080 https://www.google.com

   # SOCKS5 代理
   curl --socks5 user1:pass1@localhost:1080 http://www.google.com
   ```

## ⚙️ 配置说明

配置文件使用 JSON 格式。完整示例如下：

```json
{
  "server": {
    "http_port": 8080,       // HTTP 代理端口
    "socks5_port": 1080      // SOCKS5 代理端口
  },
  "auth": {
    "enabled": true,          // 启用认证
    "users": [
      {"username": "user1", "password": "pass1"},
      {"username": "user2", "password": "pass2"}
    ]
  },
  "ip_ban": {
    "enabled": true,          // 启用 IP 封禁
    "max_failures": 3,        // 封禁前最大失败次数
    "ban_duration_seconds": 300,  // 封禁时长（5 分钟）
    "whitelist": ["127.0.0.1"]  // 永不封禁的 IP
  },
  "rate_limit": {
    "enabled": true,
    "global_requests_per_second": 1000,  // 全局限流
    "per_ip_requests_per_second": 10     // 单 IP 限流
  },
  "circuit_breaker": {
    "enabled": true,
    "failure_threshold_percent": 50,   // 触发熔断的失败率
    "window_size_seconds": 60,          // 统计时间窗口
    "min_requests": 20,                  // 最小请求数
    "break_duration_seconds": 30        // 熔断持续时间
  },
  "log": {
    "level": "info",          // debug, info, warn, error
    "format": "json"          // json 或 console
  }
}
```

### 配置项说明

| 模块 | 选项 | 说明 | 默认值 |
|------|------|------|--------|
| `server` | `http_port` | HTTP 代理监听端口 | 8080 |
| `server` | `socks5_port` | SOCKS5 代理监听端口 | 1080 |
| `auth` | `enabled` | 启用用户认证 | false |
| `auth` | `users` | 用户名密码列表 | [] |
| `ip_ban` | `enabled` | 启用 IP 封禁 | false |
| `ip_ban` | `max_failures` | 封禁前失败次数 | 3 |
| `ip_ban` | `ban_duration_seconds` | 封禁时长（秒） | 300 |
| `ip_ban` | `whitelist` | IP 白名单 | [] |
| `rate_limit` | `enabled` | 启用限流 | false |
| `rate_limit` | `global_requests_per_second` | 全局每秒请求数 | 1000 |
| `rate_limit` | `per_ip_requests_per_second` | 单 IP 每秒请求数 | 10 |
| `circuit_breaker` | `enabled` | 启用熔断器 | false |
| `circuit_breaker` | `failure_threshold_percent` | 熔断失败率阈值 | 50 |
| `circuit_breaker` | `window_size_seconds` | 统计窗口大小 | 60 |
| `circuit_breaker` | `min_requests` | 窗口内最小请求数 | 20 |
| `circuit_breaker` | `break_duration_seconds` | 熔断持续时间 | 30 |
| `log` | `level` | 日志级别 | info |
| `log` | `format` | 日志格式 (json/console) | json |

## 🛠️ 开发

### 前置要求

- Go 1.24 或更高版本
- Make（可选，用于使用 Makefile）

### 构建

```bash
# 编译二进制文件
make build

# 运行测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 格式化代码
make fmt

# 清理构建产物
make clean
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行基准测试
go test -bench=. ./...
```

## 📊 架构

```
dudu-proxy/
├── cmd/dudu-proxy/         # 应用程序入口
├── internal/
│   ├── config/             # 配置管理
│   ├── proxy/              # HTTP 和 SOCKS5 代理实现
│   ├── middleware/         # 认证、限流、IP 封禁、熔断
│   ├── manager/            # 状态管理器（IP 封禁、熔断器）
│   └── server/             # 服务器编排
├── pkg/logger/             # 日志工具
└── configs/                # 配置文件
```

## 🐳 Docker 部署

### 使用 Docker Compose

```bash
# 启动代理
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止代理
docker-compose down
```

### 手动运行 Docker

```bash
# 构建镜像
docker build -t dudu-proxy .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -p 1080:1080 \
  -v $(pwd)/configs:/app/configs:ro \
  dudu-proxy
```

## 🔍 监控

DuDu Proxy 以结构化格式记录所有重要事件：

- 认证尝试（成功/失败）
- IP 封禁和解封
- 限流违规
- 熔断器状态变化
- 代理请求和响应

日志示例（JSON 格式）：
```json
{
  "level":"info",
  "ts":"2024-01-01T00:00:00.000Z",
  "msg":"HTTPS tunnel established",
  "client_ip":"10.0.0.1",
  "target":"example.com:443"
}
```

## 📝 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

## ⚠️ 免责声明

此代理服务器仅用于合法用途。用户有责任确保遵守适用的法律法规。

## 📧 联系方式

- GitHub: [@seakee](https://github.com/seakee)
- 项目链接: [https://github.com/seakee/dudu-proxy](https://github.com/seakee/dudu-proxy)
