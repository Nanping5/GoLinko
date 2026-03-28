# GoLinko API 文档 - 联系人管理

所有接口均需在 Header 中携带 `Authorization: Bearer <token>`。

## 1. 好友列表与详情

### 1.1 获取联系人列表 (好友)

**接口描述**: 获取当前用户的个人好友列表（已排除拉黑、删除、群组）

**请求方式**: `GET`

**请求路径**: `/v1/contact_user_list`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_id | string | 用户UUID |
| nickname | string | 用户昵称 |
| avatar | string | 头像路径 |

---

### 1.2 获取已加入的群组

**接口描述**: 获取当前用户加入的群组列表（排除自己创建的群），已退群/被踢出群组不包含在内。

**请求方式**: `GET`

**请求路径**: `/v1/load_my_joined_groups`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| group_id | string | 群组UUID |
| group_name | string | 群组名称 |
| avatar | string | 群头像 |

---

### 1.3 获取联系人详情

**接口描述**: 获取用户或群组的详细资料。如果 `contact_id` 对应一个用户，返回用户资料；如果对应一个群组，返回群组资料。

**请求方式**: `GET`

**请求路径**: `/v1/contact_info?contact_id=xxx`

**响应参数 (data）**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| contact_id | string | UUID |
| contact_name | string | 姓名/群名 |
| contact_avatar | string | 头像 |
| contact_phone | string | 手机号（仅用户） |
| contact_email | string | 邮箱（仅用户） |
| contact_gender | int8 | 性别（0=男,1=女，仅用户） |
| contact_signature | string | 个性签名（仅用户） |
| contact_birthday | string | 生日（仅用户） |
| contact_notice | string | 群公告（仅群组） |
| contact_add_mode | int8 | 入群方式（仅群组） |
| contact_owner_id | string | 群主 UUID（仅群组） |
| contact_member_cnt | int | 群成员数（仅群组） |

---

## 2. 关系维护

### 2.1 拉黑联系人 (仅个人)

**接口描述**: 将好友拉入黑名单。此操作是双向的：你将对方拉黑，对方状态变为被拉黑。暂不支持拉黑群组。

**请求方式**: `POST`

**请求路径**: `/v1/black_contact?contact_id=xxx`

---

### 2.2 解除拉黑

**接口描述**: 将好友从黑名单移出，恢复为正常联系人。

**请求方式**: `POST`

**请求路径**: `/v1/unblack_contact?contact_id=xxx`

---

### 2.3 删除联系人

**接口描述**: 删除好友关系。此操作会同步删除双向关系、申请记录及会话记录（物理删除）。

**请求方式**: `DELETE`

**请求路径**: `/v1/delete_contact?contact_id=xxx`

---

## 3. 申请管理

### 3.1 申请添加联系人/加入群聊

**接口描述**: 向用户发起好友申请，或向群组发起入群申请。

**请求方式**: `POST`

**请求路径**: `/v1/apply_add_contact`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| contact_id | string | 是 | 对方用户UUID 或 群组UUID |
| message | string | 否 | 申请备注信息 |

**响应示例**:
```json
{
  "code": 200,
  "message": "申请已提交，请等待对方审核"
}
```

---

### 3.2 获取申请列表 (待处理)

**接口描述**: 获取自己收到的好友申请，以及作为群主收到的入群申请。

**请求方式**: `GET`

**请求路径**: `/v1/contact_apply_list`

**响应参数 (data 数组项)**:

| 参数名 | 类型 | 说明 |
|--------|------|------|
| apply_id | string | 申请记录ID |
| applicant_id | string | 申请人ID |
| applicant | string | 申请人昵称 |
| avatar | string | 申请人头像 |
| message | string | 申请备注 |
| contact_type | int8 | 0: 申请加好友, 1: 申请入群 |
| target_name | string | 如果是入群申请，则显示群名；好友申请显示"个人" |
| created_at | string | 申请时间 |

---

### 3.3 同意申请

**接口描述**: 同意好友申请或入群申请。如果是好友，会自动建立双向关系并创建会话；如果是群组，会自动更新成员列表。

**请求方式**: `POST`

**请求路径**: `/v1/accept_contact_apply?apply_id=xxx`

---

### 3.4 拒绝申请

**接口描述**: 拒绝好友或入群申请。

**请求方式**: `POST`

**请求路径**: `/v1/reject_contact_apply?apply_id=xxx`
