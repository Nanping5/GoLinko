# GoLinko API 文档 - 用户管理

## 1. 用户资料管理

所有接口均需在 Header 中携带 `Authorization: Bearer <token>`。

### 1.1 获取用户信息

**接口描述**: 获取当前登录用户的信息

**请求方式**: `GET`

**请求路径**: `/v1/get_user_info`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 200,
  "message": "查询用户信息成功",
  "data": {
    "user_id": "uuid-xxx",
    "nickname": "测试用户",
    "telephone": "13800138000",
    "email": "test@example.com",
    "avatar": "https://xxx",
    "gender": 0,
    "signature": "个性签名",
    "birthday": "2000-01-01",
    "is_admin": 0,
    "status": 0,
    "created_at": "2026-03-12-10:00:00"
  }
}
```

---

### 1.2 更新用户信息

**接口描述**: 更新用户信息（支持部分更新）

**请求方式**: `PUT`

**请求路径**: `/v1/user_info`

**请求头**:
```
Content-Type: application/json
```

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nickname  | string | 否 | 用户昵称，2-20个字符 |
| avatar    | string | 否 | 头像URL |
| birthday  | string | 否 | 生日，格式：YYYY-MM-DD |
| gender    | int    | 否 | 性别，0=男，1=女 |
| signature | string | 否 | 个性签名，最多100字符 |
| telephone | string | 否 | 手机号，11位数字 |

**说明**: 只需要传递要更新的字段，未传递的字段不会被更新。用户 ID 会从 Authorization Token 中自动获取。

**请求示例（部分更新）**:
```json
{
  "nickname": "新昵称"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "user_id": "uuid-xxx",
    "nickname": "新昵称",
    "avatar": "https://xxx",
    "birthday": "2000-01-01",
    "gender": 0,
    "signature": "个性签名",
    "telephone": "13800138000"
  }
}
```

> 说明：`data` 中只返回本次更新了的字段，未更新的字段不会包含在响应中。

---

## 2. 头像上传

### 2.1 上传头像

**接口描述**: 上传当前登录用户的头像，上传成功后自动更新数据库中的 avatar 字段。

**请求方式**: `POST`

**请求路径**: `/v1/upload_avatar`

**请求头**:
```
Content-Type: multipart/form-data
```

**请求参数**:

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file   | file | 是   | 头像图片文件，大小 ≤ 10MB，支持 jpg/jpeg/png/gif/webp |

**响应示例**:
```json
{
  "code": 200,
  "message": "头像上传成功"
}
```
