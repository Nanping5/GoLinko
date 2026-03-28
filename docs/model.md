# GoLinko 数据库表结构设计

## 目录

- [1. user_info - 用户信息表](#1-user_info---用户信息表)
- [2. user_contact - 用户联系人关系表](#2-user_contact---用户联系人关系表)
- [3. contact_apply - 好友/群组申请表](#3-contact_apply---好友群组申请表)
- [4. session - 会话表](#4-session---会话表)
- [5. message - 消息表](#5-message---消息表)
- [6. group_info - 群组信息表](#6-group_info---群组信息表)
- [表关系图](#表关系图)

---

## 1. user_info - 用户信息表

存储用户基本信息，用于登录认证和用户档案管理。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| `id` | bigint | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| `uuid` | char(20) | UNIQUE, NOT NULL | 用户唯一ID |
| `nickname` | varchar(20) | NOT NULL | 昵称 |
| `telephone` | char(11) | INDEX, NOT NULL | 手机号（登录用） |
| `email` | varchar(30) | | 邮箱 |
| `avatar` | varchar(255) | NOT NULL, DEFAULT | 头像URL |
| `gender` | tinyint | DEFAULT 0 | 性别：0=男，1=女 |
| `signature` | varchar(100) | | 个性签名 |
| `password` | char(64) | NOT NULL | 加密密码（bcrypt等） |
| `birthday` | char(8) | | 生日（格式：20030101） |
| `is_admin` | tinyint | NOT NULL, DEFAULT 0 | 是否管理员：0=否，1=是 |
| `status` | tinyint | NOT NULL, DEFAULT 0 | 状态：0=正常，1=禁用 |
| `created_at` | datetime | | 创建时间 |
| `updated_at` | datetime | | 更新时间 |
| `deleted_at` | datetime | | 软删除时间 |

**常用查询：**
```sql
-- 手机号登录
SELECT * FROM user_info WHERE telephone = '13800138000' AND status = 0;

-- 获取用户信息
SELECT uuid, nickname, avatar, signature FROM user_info WHERE uuid = 'xxx';
```

---

## 2. user_contact - 用户联系人关系表

存储用户的好友关系和群组关系。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| `id` | bigint | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| `user_id` | char(20) | INDEX, NOT NULL | 用户ID |
| `contact_id` | char(20) | INDEX, NOT NULL | 联系人ID（好友ID或群ID） |
| `contact_type` | tinyint | NOT NULL | 类型：0=好友，1=群聊 |
| `status` | tinyint | NOT NULL, DEFAULT 0 | 状态：0=正常，1=拉黑，2=被拉黑，3=删除，4=被删除，5=禁言，6=退出群聊，7=被移出群聊 |
| `created_at` | datetime | | 创建时间 |
| `updated_at` | datetime | | 更新时间 |
| `deleted_at` | datetime | | 软删除时间 |

**状态说明：**
- `status = 0`：正常关系
- `status = 1`：我拉黑对方
- `status = 2`：对方拉黑我
- `status = 3`：我删除对方
- `status = 4`：对方删除我
- `status = 5`：被禁言（群聊场景）
- `status = 6`：退出群聊
- `status = 7`：被移出群聊

**常用查询：**
```sql
-- 查询用户的所有好友
SELECT * FROM user_contact
WHERE user_id = 'xxx' AND contact_type = 0 AND status = 0;

-- 查询用户加入的所有群
SELECT * FROM user_contact
WHERE user_id = 'xxx' AND contact_type = 1 AND status = 0;
```

---

## 3. contact_apply - 好友/群组申请表

处理好友申请和加群申请流程。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| `id` | bigint | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| `uuid` | char(20) | UNIQUE, NOT NULL | 申请唯一ID |
| `user_id` | char(20) | INDEX, NOT NULL | 申请人ID |
| `contact_id` | char(20) | INDEX, NOT NULL | 被申请ID（用户ID或群ID） |
| `contact_type` | tinyint | NOT NULL | 类型：0=好友申请，1=加群申请 |
| `message` | varchar(255) | | 申请消息/验证消息 |
| `status` | tinyint | NOT NULL, DEFAULT 0 | 状态：0=待处理，1=同意，2=拒绝，3=拉黑 |
| `last_apply_at` | datetime | NOT NULL | 最后申请时间 |
| `created_at` | datetime | | 创建时间 |
| `updated_at` | datetime | | 更新时间 |
| `deleted_at` | datetime | | 软删除时间 |

**常用查询：**
```sql
-- 查询我收到的好友申请
SELECT * FROM contact_apply
WHERE contact_id = '我的uuid' AND contact_type = 0 AND status = 0;

-- 查询我收到的加群申请（我是群主）
SELECT * FROM contact_apply
WHERE contact_id = '群uuid' AND contact_type = 1 AND status = 0;
```

---

## 4. session - 会话表

存储用户的会话信息，用于会话列表展示。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| `id` | bigint | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| `uuid` | char(20) | UNIQUE, NOT NULL | 会话唯一ID |
| `send_id` | char(20) | INDEX, NOT NULL | 发送者ID（当前用户） |
| `receive_id` | char(20) | INDEX, NOT NULL | 接收者ID（对方用户ID或群ID） |
| `receive_name` | varchar(20) | | 会话名称（对方昵称或群名） |
| `avatar` | varchar(255) | | 会话头像 |
| `created_at` | datetime | | 创建时间 |
| `updated_at` | datetime | | 更新时间 |
| `deleted_at` | datetime | | 软删除时间 |

**⚠️ 建议补充字段：**
| 字段名 | 类型 | 说明 |
|--------|------|------|
| `last_message` | varchar(500) | 最新消息预览 |
| `last_message_time` | datetime | 最新消息时间 |
| `unread_count` | int | 未读消息数 |
| `session_type` | tinyint | 会话类型：0=单聊，1=群聊 |

---

## 5. message - 消息表

存储所有聊天消息，支持文本、文件、通话三种类型。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| `id` | bigint | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| `uuid` | char(20) | UNIQUE, NOT NULL | 消息唯一ID |
| `session_id` | char(20) | INDEX, NOT NULL | 会话ID |
| `type` | tinyint | NOT NULL | 消息类型：0=文本，1=文件，2=通话 |
| `content` | text | | 消息内容（文本消息） |
| `url` | varchar(255) | | 消息URL（文件链接等） |
| `send_id` | char(20) | INDEX, NOT NULL | 发送者ID |
| `send_name` | varchar(20) | | 发送者名称（快照，冗余设计） |
| `send_avatar` | varchar(255) | | 发送者头像（快照，冗余设计） |
| `receive_id` | char(20) | INDEX, NOT NULL | 接收者ID |
| `file_type` | varchar(50) | | 文件类型（MIME） |
| `file_name` | varchar(255) | | 文件名称 |
| `file_size` | bigint | | 文件大小（字节） |
| `status` | tinyint | NOT NULL, DEFAULT 0 | 消息状态：0=未发送，1=已发送 |
| `av_data` | varchar(255) | | 通话数据 |
| `created_at` | datetime | | 发送时间 |
| `updated_at` | datetime | | 更新时间 |
| `deleted_at` | datetime | | 软删除时间 |

**消息类型说明：**
- `type = 0`：文本消息，使用 `content` 字段
- `type = 1`：文件消息，使用 `url`、`file_type`、`file_name`、`file_size` 字段
- `type = 2`：通话消息，使用 `av_data` 字段

**冗余设计说明：**
- `send_name` 和 `send_avatar` 采用快照设计，即使用户修改昵称/头像，历史消息仍显示发送时的信息

**常用查询：**
```sql
-- 查询会话的消息列表（分页）
SELECT * FROM message
WHERE session_id = 'xxx'
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;

-- 查询某会话的最新消息
SELECT * FROM message
WHERE session_id = 'xxx'
ORDER BY created_at DESC
LIMIT 1;
```

---

## 6. group_info - 群组信息表

存储群组基本信息。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| `id` | bigint | PRIMARY KEY, AUTO_INCREMENT | 主键 |
| `uuid` | char(20) | UNIQUE | 群组唯一ID |
| `name` | varchar(20) | | 群名称 |
| `members` | json | | 群成员列表（JSON数组） |
| `notice` | varchar(500) | | 群公告 |
| `member_cnt` | int | DEFAULT 1 | 群成员数量 |
| `owner_id` | char(20) | | 群主UUID |
| `add_mode` | tinyint | DEFAULT 0 | 加群方式：0=公开，1=需验证 |
| `avatar` | varchar(255) | | 群头像 |
| `status` | tinyint | | 群状态：0=正常，1=解散 |
| `created_at` | datetime | | 创建时间 |
| `updated_at` | datetime | | 更新时间 |
| `deleted_at` | datetime | | 软删除时间 |

**⚠️ members 字段说明：**
- 当前使用 JSON 存储成员列表，适合小型应用
- 大规模应用建议建立独立的 `group_member` 关联表，支持高效查询

---

## 表关系图

```
                    ┌─────────────────────────────────────┐
                    │           user_info (用户)          │
                    │  ─────────────────────────────────  │
                    │  uuid, nickname, telephone, avatar │
                    └──────────────┬──────────────────────┘
                                   │
                ┌──────────────────┼──────────────────┐
                │                  │                  │
                ▼                  ▼                  ▼
    ┌───────────────┐    ┌───────────────┐    ┌───────────────┐
    │ user_contact  │    │   session     │    │contact_apply  │
    │  (联系人)     │    │   (会话)      │    │   (申请)      │
    │──────────────│    │──────────────│    │──────────────│
    │ user_id       │    │ send_id       │    │ user_id       │
    │ contact_id    │    │ receive_id    │    │ contact_id    │
    │ contact_type  │    │ receive_name  │    │ contact_type  │
    │ status        │    │ avatar        │    │ status        │
    └───────┬───────┘    └───────┬───────┘    └───────────────┘
            │                    │
            │ contact_type:      │ session_id
            │ 0=好友, 1=群聊     │
            ▼                    ▼
    ┌───────────────┐    ┌───────────────┐
    │  group_info   │    │   message     │
    │   (群组)      │    │   (消息)      │
    │──────────────│    │──────────────│
    │ uuid          │    │ session_id    │
    │ name          │    │ send_id       │
    │ members       │───▶│ send_name     │
    │ owner_id      │    │ send_avatar   │
    │ add_mode      │    │ content       │
    └───────────────┘    │ type          │
                         │ status        │
                         └───────────────┘
```

**关系说明：**
- `user_info` ↔ `user_contact`：一个用户可以有多条联系人记录
- `user_contact` ↔ `group_info`：通过 `contact_id` 关联群组
- `user_info` ↔ `session`：用户拥有多个会话
- `session` ↔ `message`：一个会话包含多条消息
- `message.send_id` ↔ `user_info.uuid`：消息发送者

---

## 设计说明

### 命名规范
- 表名使用**单数形式**：`user_info`、`message`（而非 `messages`）
- 字段名使用**蛇形命名**：`send_id`、`receive_name`
- 主键统一使用 `id`（自增 bigint）
- 业务唯一标识使用 `uuid`（char(20)）

### 冗余设计
以下字段采用快照冗余设计，优先考虑查询性能：
- `message.send_name`、`message.send_avatar`：发送者信息快照
- `session.receive_name`、`session.avatar`：会话显示信息快照

**优点：**
- 减少表 JOIN，提升查询性能
- 历史消息保留发送时的上下文信息

**缺点：**
- 用户修改信息后，历史消息仍显示旧数据
- 需要额外存储空间

### 索引策略
- `uuid`：唯一索引，支持业务ID查询
- `user_id`、`send_id`、`receive_id`：普通索引，支持高频关联查询
- `session_id`：普通索引，消息列表查询必备

### 待优化项
1. **Session 表**：建议添加 `last_message`、`unread_count`、`session_type` 字段
2. **群成员管理**：大规模应用建议建立独立的 `group_member` 表
3. **消息分区**：消息量大时建议按时间或用户ID进行表分区
