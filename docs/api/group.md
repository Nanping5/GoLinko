# GoLinko API 文档 - 群组管理

所有接口均需在 Header 中携带 `Authorization: Bearer <token>`。

## 1. 群组创建与基本信息

### 1.1 创建群组

**接口描述**: 创建一个新的群组

**请求方式**: `POST`

**请求路径**: `/v1/create_group`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| group_name | string | 是 | 群名称，2-20字符 |
| notice     | string | 否 | 群公告，最大100字符 |
| add_mode   | int    | 是 | 加群方式：0直接加入，1需验证，2禁止加入 |
| avatar     | string | 否 | 群头像URL |

**响应示例**:
```json
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "group_id": "uuid-xxx"
  }
}
```

---

### 1.2 获取群信息

**接口描述**: 获取群组详细资料

**请求方式**: `GET`

**请求路径**: `/v1/get_group_info?group_id=xxx`

**响应参数 (data)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| group_id | string | 群组 UUID |
| group_name | string | 群名称 |
| avatar | string | 群头像 |
| notice | string | 群公告 |
| add_mode | int8 | 入群方式（0=直接,1=验证,2=禁止） |
| owner_id | string | 群主 UUID |
| status | int8 | 群状态（0=正常,1=已解散） |
| member_cnt | int | 群成员数 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

---

### 1.3 更新群信息

**接口描述**: 群主更新群组资料

**请求方式**: `PUT`

**请求路径**: `/v1/update_group_info`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| group_id   | string | 是 | 群组 ID |
| group_name | string | 否 | 新群名称 |
| avatar     | string | 否 | 新头像 URL |
| notice     | string | 否 | 新公告内容 |
| add_mode   | int    | 否 | 新入群模式 |

---

## 2. 我的群组

### 2.1 加载我的群组

**接口描述**: 获取当前用户加入的所有群组列表（含自己创建的和加入的）

**请求方式**: `GET`

**请求路径**: `/v1/load_my_groups`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| group_id | string | 群组 UUID |
| group_name | string | 群名称 |
| avatar | string | 群头像 |
| notice | string | 群公告 |
| add_mode | int8 | 入群方式 |
| owner_id | string | 群主 UUID |
| created_at | string | 创建时间 |

---

### 2.2 获取已加入的群组

**接口描述**: 获取当前用户加入的所有群组列表（排除自己创建的群）。

**请求方式**: `GET`

**请求路径**: `/v1/load_my_joined_groups`

---

## 3. 群成员管理

### 3.1 获取群成员

**接口描述**: 获取群组内所有成员列表

**请求方式**: `GET`

**请求路径**: `/v1/get_group_members?group_id=xxx`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_id | string | 用户 UUID |
| nickname | string | 用户昵称 |
| avatar | string | 用户头像 |

---

### 3.2 移除群成员

**接口描述**: 群主将指定成员移出群组

**请求方式**: `DELETE`

**请求路径**: `/v1/remove_group_member`

**请求头**:
```
Content-Type: application/json
```

**请求参数（JSON 请求体）**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| group_id | string | 是 | 群组 UUID |
| user_id  | string | 是 | 被移除成员的 UUID |

---

## 4. 入群与退群

### 4.1 检查加群方式

**接口描述**: 查询指定群组的入群审批模式

**请求方式**: `GET`

**请求路径**: `/v1/check_group_add_mode?group_id=xxx`

---

### 4.2 直接入群

**接口描述**: 针对 add_mode 为 0 的群组直接加入

**请求方式**: `POST`

**请求路径**: `/v1/enter_group_directly?group_id=xxx`

---

### 4.3 退出群组

**接口描述**: 退出指定的群组（群主不可退群）

**请求方式**: `POST`

**请求路径**: `/v1/leave_group`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| group_id | string | 是 | 群组 ID |

---

### 4.4 解散群组

**接口描述**: 群主解散指定的群组

**请求方式**: `POST`

**请求路径**: `/v1/dismiss_group`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| group_id | string | 是 | 群组 ID |
