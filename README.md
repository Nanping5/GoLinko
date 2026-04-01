# GoLinko

GoLinko 是一款基于 Go 的高性能即时通讯（IM）后端系统，支持私聊、群聊、文件/语音/视频消息、好友与群组管理、WebSocket 实时通信等功能，**支持分布式多实例部署**和**MySQL 读写分离**，适合自部署、二次开发和学习交流。

## 主要特性

- 用户注册、登录、邮箱验证码认证
- 好友管理、黑名单、申请与审批
- 群组创建、成员管理、群资料编辑
- 单聊/群聊消息收发，支持文本、文件、语音、音视频
- WebSocket 实时消息推送
- 个人资料与头像上传
- 管理员后台
- **分布式架构支持**（WebSocket 集群化、Kafka 跨实例通信）
- **MySQL 读写分离**（主从架构 + 多级数据一致性保障）
- Kafka 消息队列、MySQL、Redis 支持
- MinIO 对象存储（分布式文件存储）

## 技术栈

- **语言**: Go 1.23+
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL（主从架构）
- **缓存**: Redis
- **消息队列**: Kafka
- **实时通信**: WebSocket
- **对象存储**: MinIO
- **认证**: JWT

## 分布式架构

### 架构概览

```
                    ┌─────────────────┐
                    │  Nginx 负载均衡  │
                    └────────┬────────┘
           ┌─────────────────┼─────────────────┐
           ▼                 ▼                 ▼
    ┌───────────┐     ┌───────────┐     ┌───────────┐
    │ 实例 A    │     │ 实例 B    │     │ 实例 C    │
    │(HTTP+WS)  │     │(HTTP+WS)  │     │(HTTP+WS)  │
    └─────┬─────┘     └─────┬─────┘     └─────┬─────┘
          └─────────────────┼─────────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
   ┌───────────┐     ┌───────────┐     ┌───────────┐
   │ MySQL主从 │     │   Redis   │     │   Kafka   │
   │  读写分离  │     └───────────┘     └───────────┘
   └───────────┘
        │
   ┌────┴────┐
   ▼         ▼
主库(写)   从库(读)
```

### 分布式特性

| 功能 | 实现方案 |
|------|----------|
| **WebSocket 集群化** | Kafka 跨实例消息广播 |
| **用户在线状态** | Redis 集中存储 + 心跳续期机制 |
| **Kafka 消费者组** | 实例唯一 ID 动态生成 |
| **文件存储** | MinIO 对象存储 |

### MySQL 读写分离

| 功能 | 实现方案 |
|------|----------|
| **主从架构** | 主库负责写，从库负责读 |
| **读写路由** | `GetWriteDB()` / `GetReadDB()` |
| **请求级一致性** | 同一请求内写后读自动走主库（推荐） |
| **读后写一致性** | 记录写入时间戳，短时间内从主库读 |
| **会话一致性** | SessionContext 追踪会话内写入 |
| **延迟容忍** | 可配置主从同步延迟时间 |

#### 数据一致性策略

```
用户请求
   ↓
┌─────────────────────────────────────┐
│  RequestConsistency 中间件          │
│  注入 consistencyCtx                │
└─────────────────────────────────────┘
   ↓
┌─────────────────────────────────────┐
│  写操作 (WriteAndMark)              │
│  1. 执行写入 → 主库                 │
│  2. ctx = MarkWritten(ctx) ← 标记   │
└─────────────────────────────────────┘
   ↓
┌─────────────────────────────────────┐
│  后续读操作                          │
│  检测 HasWritten(ctx) == true       │
│  → 强制走主库 ✓                     │
└─────────────────────────────────────┘
```

## 快速启动

### 1. 克隆项目

```bash
git clone https://github.com/Nanping5/GoLinko.git
cd GoLinko
```

### 2. 启动基础设施（Docker）

```bash
# 单库模式
docker-compose up -d mysql redis kafka minio

# 主从模式
docker-compose up -d mysql-cludae mysql-slave redis kafka minio
```

### 3. 配置后端

```bash
cp configs/configs.toml.example configs/configs.toml
```

修改 `configs/configs.toml`，填写数据库、Redis、Kafka、SMTP 等信息。

### 4. 启动后端

```bash
go run cmd/main.go
```

后端默认端口：8081

## 分布式测试

### 启动多实例

```bash
# 终端 1：启动实例 1
INSTANCE_ID=instance-1 go run cmd/main.go

# 终端 2：启动实例 2（修改 configs.toml 的端口）
INSTANCE_ID=instance-2 go run cmd/main.go
```

### Docker 集群部署

```bash
./scripts/start_cluster.sh
```

## 目录结构

```
GoLinko/
├── api/                        # Gin 路由与控制器
├── cmd/                        # 入口
├── configs/                    # 配置文件
├── docs/                       # API 文档与开发文档
├── internal/
│   ├── config/                 # 配置管理
│   ├── dao/
│   │   ├── gorm.go             # 数据访问层（读写分离）
│   │   └── consistency.go      # 数据一致性管理
│   ├── dto/                    # 数据传输对象
│   ├── model/                  # 数据模型
│   └── service/
│       ├── chat/               # WebSocket 聊天服务
│       │   ├── server.go       # 服务端（分布式支持）
│       │   ├── client.go       # 客户端连接
│       │   └── kafka_broker.go # Kafka 消息广播器
│       ├── gorms/              # 业务逻辑服务
│       ├── kafka/              # Kafka 服务
│       ├── redis/
│       │   ├── redis_service.go# Redis 基础服务
│       │   |── online.go       # 用户在线状态管理
│       ├── storage/
│       │   └── minio.go        # MinIO 对象存储
│       └── sms/                # 邮件验证码服务
├── middleware/                 # 中间件（JWT 认证）
├── nginx/nginx.conf            # Nginx 负载均衡配置
├── pkg/                        # 公共包
│   ├── const/                  # 常量
│   ├── enum/                   # 枚举
│   ├── utils/                  # 工具函数
│   └── zlog/                   # 日志
├── docker-compose.yml          # Docker 编排
└── Dockerfile                  # Docker 构建文件
```

## API 文档

详见 [docs/api.md](docs/api.md) 及其子文档，涵盖认证、用户、联系人、群组、消息等所有接口说明与示例。

## 后端技术详解

详见 [resume/backend-description.md](resume/backend-description.md)，包含：
- 分布式 WebSocket 架构设计
- MySQL 读写分离与数据一致性方案
- Kafka 消息广播机制
- Redis 缓存策略
- 高并发优化实践

## 配置说明

```toml
# configs/configs.toml

[main_config]
app_name = "GoLinko"
host = "0.0.0.0"
port = 8081

[mysql_config]
# 单库配置（兼容）
host = "127.0.0.1"
port = 3306
user = "root"
password = "your_password"
db_name = "golinko"

# 读写分离配置
write_host = "127.0.0.1"
write_port = 3306
read_hosts = ["127.0.0.1", "127.0.0.1"]
read_ports = [3306, 3307]

# 主从同步延迟容忍时间（秒）
replication_lag = 3

[redis_config]
host = "127.0.0.1"
port = 6379
password = ""
db = 0

[kafka_config]
hostport = "localhost:9092"
message_mode = "channel"  # channel 或 kafka
chat_topic = "chat-messages"
partition = 6

[minio_config]
endpoint = "localhost:9000"
access_key = "minioadmin"
secret_key = "minioadmin"
bucket = "golinko"
use_ssl = false

[distributed]
enabled = true
instance_id = ""  # 为空则自动生成
```

## 服务端口

| 服务 | 端口 |
|------|------|
| HTTP API | 8081 |
| MinIO API | 9000 |
| MinIO Console | 9001 |
| MySQL 主库 | 13306 |
| MySQL 从库 | 13307 |
| Redis | 6379 |
| Kafka | 19092 |

## 性能测试

| 场景 | 指标 | 实测值 |
|------|------|--------|
| WebSocket 连接 | 并发连接 | 46,350 |
| 连接速率 | conn/s | 612.75 |
| 连接延迟 | P50 | 8.71ms |
| 连接延迟 | P99 | 241.32ms |
| Redis OPS | ops/s | 146,527 |

## 常见问题 FAQ

1. 启动报数据库/Redis/Kafka 连接失败？请检查 configs.toml 配置与服务状态。
2. 静态资源上传失败？请检查 static 目录权限与 .gitignore 设置。
3. 邮箱验证码收不到？请检查 SMTP 配置。
4. 分布式消息推送失败？请检查 Kafka 连接和 Topic 配置。
5. 主从同步延迟导致数据不一致？调整 `replication_lag` 配置或使用会话一致性。

如有其他问题，欢迎提 Issue。

## 贡献指南

1. Fork 本仓库并新建分支
2. 提交代码并发起 Pull Request
3. 遵循项目代码风格和提交规范
4. 请勿将敏感配置上传至仓库

## License

MIT
