# GoLinko

GoLinko 是一款基于 Go + React 的现代化即时通讯（IM）系统，支持私聊、群聊、文件/语音/视频消息、好友与群组管理、WebSocket 实时通信等功能，适合自部署、二次开发和学习交流。

## 主要特性

- 用户注册、登录、邮箱验证码认证
- 好友管理、黑名单、申请与审批
- 群组创建、成员管理、群资料编辑
- 单聊/群聊消息收发，支持文本、文件、语音、音视频
- WebSocket 实时消息推送
- 个人资料与头像上传
- 管理员后台
- Kafka 消息队列、MySQL、Redis 支持
- 前后端分离，接口文档齐全

## 技术栈

- **后端**：Go (Gin, GORM, Redis, Kafka, JWT)
- **前端**：React 19, TypeScript, Vite, Zustand, TailwindCSS, React Router
- **数据库**：MySQL、Redis
- **消息队列**：Kafka
- **实时通信**：WebSocket

## 快速启动

### 1. 克隆项目

```bash
git clone https://your.repo.url/GoLinko.git
cd GoLinko
```

### 2. 配置后端

- 修改 `configs/configs.toml`，填写数据库、Redis、Kafka、SMTP 等信息

### 3. 启动后端

```bash
cd cmd
go run main.go
```

### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认端口：5173，后端默认端口：8080

## 目录结构

```
GoLinko/
├── api/                # Gin 路由与控制器
├── cmd/                # 后端入口
├── configs/            # 配置文件
├── docs/               # API 文档与开发文档
├── frontend/           # 前端 React 项目
├── internal/           # 业务逻辑、服务、DAO、DTO
├── middleware/         # 中间件
├── pkg/                # 公共包
├── static/             # 静态资源（已忽略版本控制）
└── ...
```

## API 文档

详见 [docs/api.md](docs/api.md) 及其子文档，涵盖认证、用户、联系人、群组、消息等所有接口说明与示例。

## 贡献指南

1. Fork 本仓库并新建分支
2. 提交代码并发起 Pull Request
3. 遵循项目代码风格和提交规范

## License

MIT
