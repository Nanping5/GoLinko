# GoLinko frpc 开发联调说明

本文用于在不同设备上访问本机开发环境（前端 Vite + 后端 Go），便于测试消息、文件与音视频通话。

## 1. 准备 frpc 配置

1. 复制模板文件：

```bash
cp configs/frpc.dev.ini.example configs/frpc.dev.ini
```

2. 编辑 `configs/frpc.dev.ini`，替换以下字段：
- `server_addr`
- `server_port`
- `token`
- `custom_domains`

建议使用两个域名：
- 前端：`web-GoLinko.xxx`
- 后端：`api-GoLinko.xxx`

## 2. 配置前端 frpc 环境

1. 复制前端环境模板：

```bash
cp frontend/.env.frpc.example frontend/.env.frpc
```

2. 编辑 `frontend/.env.frpc`：

```env
# 推荐留空，使用前端同域代理
# VITE_API_URL=
```

注意：
- 推荐不设置 `VITE_API_URL`，前端会走当前站点同域名，请求 `/v1` 由 Vite 代理到本机后端 `127.0.0.1:8080`。
- 如你要直连后端隧道，`VITE_API_URL` 必须指向浏览器可信任证书的 HTTPS 域名。
- 前端会基于该地址自动推导 WebSocket 地址（`ws/wss`），用于实时消息与通话信令。

## 3. 启动本地服务

在项目根目录启动后端：

```bash
cd /Users/liuxu/Code/proj/GoLinko
./coco_backend
```

在前端目录启动 frpc 模式：

```bash
cd /Users/liuxu/Code/proj/GoLinko/frontend
npm run dev:frpc
```

## 4. 启动 frpc

在项目根目录执行：

```bash
frpc -c configs/frpc.dev.ini
```

## 5. 跨设备访问

其他设备打开：
- 前端地址：`https://web-GoLinko.xxx`

前端会调用：
- `https://api-GoLinko.xxx`（HTTP API）
- `wss://api-GoLinko.xxx/v1/ws`（WebSocket）

## 6. 常见问题

1. 页面能开但接口 401/网络错误：
- 若使用“同域代理方案”，确认 Vite 正在运行 `npm run dev:frpc`，且后端在 `127.0.0.1:8080`。
- 若直连后端隧道，检查 `frontend/.env.frpc` 的 `VITE_API_URL` 是否正确，且证书被浏览器信任。

2. 页面能开但后端请求失败并提示证书错误：
- 这是后端隧道 HTTPS 证书不被浏览器信任导致。
- 处理方式：改为“同域代理方案”（推荐），或换成可被浏览器信任证书的后端域名。

3. 通话建立失败：
- 先确认双方都能正常收发文本消息（说明 WebSocket 通道可用）。
- 若同一局域网可通、跨网络无画面/无声音，通常是 NAT 穿透失败，需要 TURN 中继。
- 如果使用 Cloudflare TURN，后端必须先配置环境变量（不要把密钥写入代码）：

```bash
export CF_TURN_KEY_ID=your_cloudflare_turn_key_id
export CF_TURN_API_TOKEN=your_cloudflare_turn_api_token
```

- 前端支持以下 TURN 配置（任选其一）：

```env
# 方案 A：JSON 配置全部 ICE（优先级最高）
VITE_WEBRTC_ICE_SERVERS=[{"urls":["stun:stun.l.google.com:19302"]},{"urls":["turn:turn.example.com:3478?transport=udp","turn:turn.example.com:3478?transport=tcp"],"username":"demo","credential":"demo-pass"}]

# 方案 B：拆分变量（会自动叠加默认 STUN）
VITE_TURN_URLS=turn:turn.example.com:3478?transport=udp,turn:turn.example.com:3478?transport=tcp
VITE_TURN_USERNAME=demo
VITE_TURN_CREDENTIAL=demo-pass
```

- 修改 `.env` / `.env.frpc` 后，必须重启前端开发服务。

4. 前端改了 `.env.frpc` 没生效：
- 重启 `npm run dev:frpc`。
