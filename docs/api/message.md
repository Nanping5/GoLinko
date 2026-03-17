# GoLinko API 文档 - 会话与消息管理

所有接口均需在 Header 中携带 `Authorization: Bearer <token>`。

## 1. 会话管理

### 1.1 检查是否可以发起会话

**接口描述**: 检查当前用户是否可以与指定用户/群组发起会话（校验拉黑、删除、封禁状态）。

**请求方式**: `GET`

**请求路径**: `/v1/check_open_session_allowed?receive_id=xxx`

---

### 1.2 打开/创建会话

**接口描述**: 打开与指定用户或群组的会话。若会话已存在则直接返回，不会重复创建。若该会话此前被当前用户隐藏，打开后会自动恢复显示。

**请求方式**: `POST`

**请求路径**: `/v1/open_session?receive_id=xxx`

---

### 1.3 获取单聊会话列表

**接口描述**: 获取当前用户的所有单聊会话，按创建时间倒序排列。结果有 Redis 缓存，TTL 60 秒。

**请求方式**: `GET`

**请求路径**: `/v1/session_list`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| session_id | string | 会话 UUID |
| receive_id | string | 对方用户 UUID |
| receive_name | string | 对方昵称 |
| avatar | string | 对方头像 |

---

### 1.4 获取群聊会话列表

**接口描述**: 获取当前用户的所有群聊会话，按创建时间倒序排列。结果有 Redis 缓存，TTL 60 秒。

**请求方式**: `GET`

**请求路径**: `/v1/group_session_list`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| session_id | string | 会话 UUID |
| group_id | string | 群组 UUID |
| group_name | string | 群名称 |
| avatar | string | 群头像 |

---

### 1.5 隐藏会话（原 delete_session）

**接口描述**: 仅对当前用户隐藏指定会话（不会物理删除会话及历史消息），并清理当前用户会话缓存。

**请求方式**: `DELETE`

**请求路径**: `/v1/delete_session?session_id=xxx`

**说明**:
- 该操作只影响当前登录用户。
- 对方用户会话列表不受影响。
- 后续再次调用 `open_session` 打开该会话时，会自动取消隐藏。

---

## 2. 消息获取

### 2.1 获取单聊消息列表

**接口描述**: 获取指定单聊会话的消息列表，按发送时间升序排列。结果有 Redis 缓存，TTL 5 分钟。仅会话参与者可访问。

**请求方式**: `GET`

**请求路径**: `/v1/message_list?session_id=xxx`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| send_id | string | 发送者 UUID |
| send_name | string | 发送者昵称 |
| send_avatar | string | 发送者头像 |
| receive_id | string | 接收者 UUID |
| content | string | 消息内容（文本消息） |
| url | string | 文件/图片 URL |
| type | int8 | 消息类型（0=文本,2=文件,3=音视频信令） |
| file_type | string | 文件类型（MIME） |
| file_name | string | 文件名 |
| file_size | int64 | 文件大小（字节） |
| av_data | string | 音视频信令数据 |

**权限与错误说明**:
- 非会话参与者访问会返回无权限错误（`code=400`，业务文案如“无权访问该会话消息”）。
- 会话不存在会返回业务错误（`code=400`，文案如“会话不存在”）。

---

### 2.2 获取群聊消息列表

**接口描述**: 获取指定群组的消息列表，按发送时间升序排列。结果有 Redis 缓存，TTL 5 分钟（按 group_id 维度缓存，所有群成员共享同一份缓存）。仅群成员可访问。

**请求方式**: `GET`

**请求路径**: `/v1/group_message_list?group_id=xxx`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| send_id | string | 发送者 UUID |
| send_name | string | 发送者昵称 |
| send_avatar | string | 发送者头像 |
| group_id | string | 群组 UUID |
| content | string | 消息内容（文本消息） |
| url | string | 文件/图片 URL |
| type | int8 | 消息类型（0=文本,2=文件,3=音视频信令） |
| file_type | string | 文件类型（MIME） |
| file_name | string | 文件名 |
| file_size | int64 | 文件大小（字节） |
| av_data | string | 音视频信令数据 |

**权限与错误说明**:
- 非群成员访问会返回无权限错误（`code=400`，业务文案如“无权访问该群消息”）。

---

## 3. WebSocket 实时通信

### 3.1 建立 WebSocket 连接

**接口描述**: 建立 WebSocket 长连接，用于发送和接收实时消息。Token 通过 Query 参数传入（不经过 HTTP Header）。

**请求方式**: `GET`（协议升级）

**连接地址**:
```
ws://localhost:8080/v1/ws?token=<JWT_TOKEN>
```

**连接说明**:
- 连接建立后，服务端自动将当前用户注册到在线连接管理器
- 服务端会推送一条欢迎消息: `登录成功,欢迎来到GoLinko`
- 客户端断开连接后自动从管理器注销
- `send_id` 字段由服务端基于鉴权 Token 强制覆盖，客户端传入的值会被忽略

---

### 3.2 发送消息

**接口描述**: 客户端通过 WebSocket 发送 JSON 消息。服务端根据消息类型进行入库并转发给目标客户。

**消息格式 (JSON)**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| session_id | string | 是 | 会话 UUID |
| type | int8 | 是 | 消息类型（0=文本, 2=文件, 3=音视频信令） |
| content | string | 否 | 消息内容（文本消息时必填） |
| url | string | 否 | 文件/图片访问 URL（文件消息时必填） |
| send_name | string | 是 | 发送者昵称 |
| send_avatar | string | 否 | 发送者头像 URL |
| receive_id | string | 是 | 接收者 UUID（单聊为用户 UUID，群聊为群组 UUID） |
| file_name | string | 否 | 文件名称 |
| file_type | string | 否 | 文件 MIME 类型 |
| file_size | string | 否 | 文件大小（字符串形式） |
| av_data | string | 否 | 音视频信令数据（type=3 时使用） |

**文本消息示例**:
```json
{
  "session_id": "uuid-xxx",
  "type": 0,
  "content": "你好啊",
  "receive_id": "uuid-yyy",
  "send_name": "小明",
  "send_avatar": "https://xxx"
}
```

**文件消息示例**:
```json
{
  "session_id": "uuid-xxx",
  "type": 2,
  "url": "https://host/static/files/xxx.pdf",
  "receive_id": "uuid-yyy",
  "send_name": "小明",
  "send_avatar": "https://xxx",
  "file_name": "document.pdf",
  "file_type": "application/pdf",
  "file_size": "204800"
}
```

---

### 3.3 接收消息

**接口描述**: 服务端将消息持久化后将原始 JSON 内容循环推送给目标用户。收到的 JSON 格式与发送格式相同，即 `ChatMessageRequest` 结构体内容，其中 `send_id` 字段已被服务端强制覆盖为真实 UUID。

---

## 4. 文件消息附件上传

### 4.1 上传文件（无需鉴权）

**接口描述**: 上传普通文件（图片、PDF、Office 文档、压缩包、音视频等），用于发送文件消息。

**请求方式**: `POST`

**请求路径**: `/v1/upload_file`

**请求头**:
```
Content-Type: multipart/form-data
```

**请求参数**: 文件字段名任意，支持一次上传多个文件。

**约束**:
- 单次请求总大小上限：32MB
- 允许的文件扩展名：`.jpg` `.jpeg` `.png` `.gif` `.webp` `.pdf` `.doc` `.docx` `.xls` `.xlsx` `.txt` `.zip` `.mp4` `.mp3`
