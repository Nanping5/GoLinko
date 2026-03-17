# GoLinko API 文档 - 系统与认证

## 1. 系统接口

### 1.1 健康检查

**接口描述**: 测试服务是否正常运行

**请求方式**: `GET`

**请求路径**: `/v1/ping`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 200,
  "message": "pong"
}
```

---

## 2. 用户认证

### 2.1 发送邮箱验证码

**接口描述**: 发送6位数字验证码到指定邮箱，有效期3分钟

**请求方式**: `POST`

**请求路径**: `/v1/send_email_code`

**请求头**:
```
Content-Type: application/json
```

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| email  | string | 是 | 有效的邮箱地址 |

**请求示例**:
```json
{
  "email": "5738645529@163.com"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "验证码发送成功"
}
```

---

### 2.2 用户注册

**接口描述**: 用户注册接口

**请求方式**: `POST`

**请求路径**: `/v1/register`

**请求头**:
```
Content-Type: application/json
```

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nickname  | string | 是 | 用户昵称，2-20个字符 |
| telephone | string | 是 | 手机号，11位数字 |
| email     | string | 是 | 有效邮箱地址 |
| password  | string | 是 | 密码，6-20个字符 |
| code      | string | 是 | 6位验证码 |

**请求示例**:
```json
{
  "nickname": "测试用户",
  "telephone": "13800138000",
  "email": "a15738645529@163.com",
  "password": "liuxu123123",
  "code": "123456"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": "uuid-xxx",
    "nickname": "测试用户",
    "telephone": "13800138000",
    "email": "test@example.com",
    "avatar": "https://xxx",
    "is_admin": 0,
    "status": 0,
    "created_at": "2026-03-12"
  }
}
```

---

### 2.3 用户登录（密码）

**接口描述**: 使用邮箱和密码登录

**请求方式**: `POST`

**请求路径**: `/v1/login`

**请求头**:
```
Content-Type: application/json
```

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| email    | string | 是 | 有效邮箱地址 |
| password | string | 是 | 密码，6-20个字符 |

**请求示例**:
```json
{
  "email": "5738645529@163.com",
  "password": "liuxu123123"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user_id": "uuid-xxx",
    "nickname": "测试用户",
    "telephone": "13800138000",
    "email": "test@example.com",
    "avatar": "https://xxx",
    "gender": 0,
    "signature": "",
    "birthday": "",
    "is_admin": 0,
    "status": 0,
    "token": "eyJhbGci..."
  }
}
```

---

### 2.4 用户登录（验证码）

**接口描述**: 使用邮箱验证码登录

**请求方式**: `POST`

**请求路径**: `/v1/login_by_code`

**请求头**:
```
Content-Type: application/json
```

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| email | string | 是 | 有效邮箱地址 |
| code  | string | 是 | 6位验证码 |

**请求示例**:
```json
{
  "email": "5738645529@163.com",
  "code": "123456"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user_id": "uuid-xxx",
    "nickname": "测试用户",
    "telephone": "13800138000",
    "email": "test@example.com",
    "avatar": "https://xxx",
    "gender": 0,
    "signature": "",
    "birthday": "",
    "is_admin": 0,
    "status": 0,
    "token": "eyJhbGci..."
  }
}
```
