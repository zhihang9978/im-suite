# 志航密信 - 高级消息功能 API 文档

## 📋 概述

本文档描述了志航密信高级消息功能的API接口，包括消息撤回、编辑、转发、搜索、定时发送、加密和自毁等功能。

## 🔐 认证

所有API请求都需要在Header中包含JWT认证令牌：

```
Authorization: Bearer <your-jwt-token>
```

## 📨 消息撤回

### 撤回消息

**POST** `/api/messages/recall`

撤回已发送的消息。

#### 请求参数

```json
{
  "message_id": 123,
  "reason": "发错了"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | uint | 是 | 要撤回的消息ID |
| reason | string | 否 | 撤回原因 |

#### 响应

```json
{
  "message": "消息撤回成功"
}
```

#### 限制

- 只有消息发送者或群管理员可以撤回消息
- 消息发送超过24小时无法撤回
- 已撤回的消息无法再次撤回

## ✏️ 消息编辑

### 编辑消息

**POST** `/api/messages/edit`

编辑已发送的消息。

#### 请求参数

```json
{
  "message_id": 123,
  "content": "修改后的内容",
  "reason": "修正错误"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | uint | 是 | 要编辑的消息ID |
| content | string | 是 | 新的消息内容 |
| reason | string | 否 | 编辑原因 |

#### 响应

```json
{
  "message": "消息编辑成功"
}
```

#### 限制

- 只有消息发送者可以编辑消息
- 消息发送超过48小时无法编辑
- 每条消息最多编辑5次
- 已撤回的消息无法编辑

### 获取编辑历史

**GET** `/api/messages/{message_id}/edit-history`

获取消息的编辑历史记录。

#### 响应

```json
{
  "data": [
    {
      "id": 1,
      "message_id": 123,
      "old_content": "原内容",
      "new_content": "修改后内容",
      "edit_time": "2024-12-19T10:30:00Z",
      "edit_reason": "修正错误"
    }
  ]
}
```

## 📤 消息转发

### 转发消息

**POST** `/api/messages/forward`

转发消息到其他聊天。

#### 请求参数

```json
{
  "message_id": 123,
  "target_chat_id": 456,
  "comment": "转发说明"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | uint | 是 | 要转发的消息ID |
| target_chat_id | uint | 否 | 目标群聊ID |
| target_user_id | uint | 否 | 目标用户ID（单聊） |
| comment | string | 否 | 转发时的评论 |

#### 响应

```json
{
  "message": "消息转发成功"
}
```

#### 限制

- 不能转发已撤回的消息
- 必须指定转发目标（群聊或用户）
- target_chat_id和target_user_id不能同时为空

## 🔍 消息搜索

### 搜索消息

**GET** `/api/messages/search`

搜索消息内容。

#### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| query | string | 是 | 搜索关键词 |
| chat_id | uint | 否 | 限制搜索的群聊ID |
| user_id | uint | 否 | 限制搜索的用户ID |
| message_type | string | 否 | 消息类型过滤 |
| date_from | string | 否 | 开始日期 |
| date_to | string | 否 | 结束日期 |
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20，最大100 |

#### 响应

```json
{
  "data": [
    {
      "id": 123,
      "content": "搜索结果",
      "sender": {
        "id": 1,
        "nickname": "用户1"
      },
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "pages": 5
  }
}
```

## ⏰ 定时消息

### 设置定时发送

**POST** `/api/messages/schedule`

设置消息定时发送。

#### 请求参数

```json
{
  "content": "定时消息内容",
  "message_type": "text",
  "target_chat_id": 456,
  "scheduled_time": "2024-12-20T10:00:00Z",
  "is_silent": false
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 消息内容 |
| message_type | string | 是 | 消息类型 |
| target_chat_id | uint | 否 | 目标群聊ID |
| target_user_id | uint | 否 | 目标用户ID |
| scheduled_time | string | 是 | 定时发送时间 |
| is_silent | bool | 否 | 是否静默发送 |

#### 响应

```json
{
  "message": "定时消息设置成功"
}
```

### 取消定时消息

**DELETE** `/api/messages/{message_id}/schedule`

取消已设置的定时消息。

#### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| reason | string | 否 | 取消原因 |

#### 响应

```json
{
  "message": "定时消息取消成功"
}
```

### 获取定时消息列表

**GET** `/api/messages/scheduled`

获取用户的定时消息列表。

#### 响应

```json
{
  "data": [
    {
      "id": 123,
      "content": "定时消息内容",
      "scheduled_time": "2024-12-20T10:00:00Z",
      "is_executed": false,
      "is_cancelled": false
    }
  ]
}
```

## 🔒 消息加密

### 加密消息

**POST** `/api/encryption/encrypt`

对消息进行加密。

#### 请求参数

```json
{
  "message_id": 123,
  "encryption_type": "simple",
  "password": "123456",
  "self_destruct_time": 3600
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | uint | 是 | 要加密的消息ID |
| encryption_type | string | 是 | 加密类型：simple（简单加密）或e2e（端到端加密） |
| password | string | 否 | 加密密码（简单加密时必填） |
| self_destruct_time | int | 否 | 自毁时间（秒） |

#### 响应

```json
{
  "message": "消息加密成功"
}
```

### 解密消息

**POST** `/api/encryption/decrypt`

解密消息内容。

#### 请求参数

```json
{
  "message_id": 123,
  "password": "123456"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | uint | 是 | 要解密的消息ID |
| password | string | 否 | 解密密码（简单加密时必填） |

#### 响应

```json
{
  "content": "解密后的消息内容",
  "message": "消息解密成功"
}
```

### 获取加密消息信息

**GET** `/api/encryption/{message_id}/info`

获取消息的加密信息。

#### 响应

```json
{
  "data": {
    "is_encrypted": true,
    "is_self_destruct": true,
    "message_type": "text",
    "created_at": "2024-12-19T10:30:00Z",
    "self_destruct_time": "2024-12-19T11:30:00Z",
    "time_remaining": 3600
  }
}
```

### 设置消息自毁

**POST** `/api/encryption/self-destruct`

设置消息自毁时间。

#### 请求参数

```json
{
  "message_id": 123,
  "destruct_time": 3600
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | uint | 是 | 消息ID |
| destruct_time | int | 是 | 自毁时间（秒），最小1秒 |

#### 响应

```json
{
  "message": "消息自毁时间设置成功"
}
```

## 📝 消息回复

### 回复消息

**POST** `/api/messages/reply`

回复指定消息。

#### 请求参数

```json
{
  "reply_to_id": 123,
  "content": "回复内容",
  "message_type": "text",
  "chat_id": 456
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| reply_to_id | uint | 是 | 要回复的消息ID |
| content | string | 是 | 回复内容 |
| message_type | string | 是 | 消息类型 |
| chat_id | uint | 否 | 群聊ID |
| user_id | uint | 否 | 用户ID（单聊） |

#### 响应

```json
{
  "message": "回复消息发送成功"
}
```

## 🚨 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权或认证失败 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 📊 状态码

### 消息状态

| 状态 | 说明 |
|------|------|
| sent | 已发送 |
| delivered | 已送达 |
| read | 已读 |
| recalled | 已撤回 |
| edited | 已编辑 |
| scheduled | 定时发送 |
| cancelled | 已取消 |
| destroyed | 已自毁 |

### 消息类型

| 类型 | 说明 |
|------|------|
| text | 文本消息 |
| image | 图片消息 |
| video | 视频消息 |
| audio | 音频消息 |
| file | 文件消息 |
| sticker | 贴纸消息 |
| gif | GIF消息 |

### 加密类型

| 类型 | 说明 |
|------|------|
| simple | 简单加密（基于密码） |
| e2e | 端到端加密（基于密钥） |

## 🔧 使用示例

### 完整的消息操作流程

```bash
# 1. 发送消息
curl -X POST "http://localhost:8080/api/messages/send" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "这是一条测试消息",
    "message_type": "text",
    "chat_id": 1
  }'

# 2. 编辑消息
curl -X POST "http://localhost:8080/api/messages/edit" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": 123,
    "content": "修改后的消息内容"
  }'

# 3. 加密消息
curl -X POST "http://localhost:8080/api/encryption/encrypt" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": 123,
    "encryption_type": "simple",
    "password": "123456",
    "self_destruct_time": 3600
  }'

# 4. 搜索消息
curl -X GET "http://localhost:8080/api/messages/search?query=测试&page=1&page_size=20" \
  -H "Authorization: Bearer your-token"

# 5. 撤回消息
curl -X POST "http://localhost:8080/api/messages/recall" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": 123,
    "reason": "发错了"
  }'
```

---

**最后更新**: 2024-12-19  
**版本**: v1.1.0  
**维护者**: 志航密信技术团队
