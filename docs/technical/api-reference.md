# 志航密信 API 参考文档

## 📋 目录

- [API 概述](#api-概述)
- [认证](#认证)
- [用户管理](#用户管理)
- [联系人管理](#联系人管理)
- [聊天管理](#聊天管理)
- [消息管理](#消息管理)
- [文件管理](#文件管理)
- [WebSocket 事件](#websocket-事件)
- [错误码](#错误码)

## 🔗 API 概述

志航密信提供 RESTful API 和 WebSocket 实时通讯接口，支持用户认证、消息收发、文件传输等功能。

### 基础信息

- **Base URL**: `https://your-domain.com/api`
- **WebSocket URL**: `wss://your-domain.com/ws`
- **API 版本**: v1
- **数据格式**: JSON
- **字符编码**: UTF-8

### 通用响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### 错误响应格式

```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "详细错误信息",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## 🔐 认证

### 用户登录

**POST** `/auth/login`

用户通过手机号和验证码或密码登录。

**请求参数**:
```json
{
  "phone": "13800138000",
  "code": "123456",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user": {
      "id": 1,
      "phone": "13800138000",
      "username": "testuser",
      "nickname": "测试用户",
      "avatar": "https://example.com/avatar.jpg",
      "online": true,
      "last_seen": "2024-12-19T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

### 用户注册

**POST** `/auth/register`

用户注册新账户。

**请求参数**:
```json
{
  "phone": "13800138000",
  "username": "testuser",
  "password": "password123",
  "nickname": "测试用户"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user": {
      "id": 1,
      "phone": "13800138000",
      "username": "testuser",
      "nickname": "测试用户",
      "avatar": "",
      "online": false,
      "last_seen": "2024-12-19T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

### 刷新令牌

**POST** `/auth/refresh`

使用刷新令牌获取新的访问令牌。

**请求参数**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应**:
```json
{
  "code": 200,
  "message": "令牌刷新成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

### 用户登出

**POST** `/auth/logout`

用户登出，使令牌失效。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "登出成功",
  "data": {}
}
```

## 👤 用户管理

### 获取当前用户信息

**GET** `/users/me`

获取当前登录用户的详细信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "获取用户信息成功",
  "data": {
    "id": 1,
    "phone": "13800138000",
    "username": "testuser",
    "nickname": "测试用户",
    "bio": "这是我的个人简介",
    "avatar": "https://example.com/avatar.jpg",
    "online": true,
    "last_seen": "2024-12-19T10:30:00Z",
    "language": "zh-CN",
    "theme": "auto",
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### 更新用户信息

**PUT** `/users/me`

更新当前用户的信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求参数**:
```json
{
  "nickname": "新昵称",
  "bio": "新的个人简介",
  "avatar": "https://example.com/new-avatar.jpg"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "用户信息更新成功",
  "data": {}
}
```

### 获取用户信息

**GET** `/users/{user_id}`

获取指定用户的公开信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "获取用户信息成功",
  "data": {
    "id": 2,
    "username": "otheruser",
    "nickname": "其他用户",
    "avatar": "https://example.com/other-avatar.jpg",
    "online": false,
    "last_seen": "2024-12-19T09:30:00Z"
  }
}
```

## 👥 联系人管理

### 获取联系人列表

**GET** `/contacts`

获取当前用户的联系人列表。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "获取联系人列表成功",
  "data": {
    "contacts": [
      {
        "id": 1,
        "user_id": 1,
        "contact_id": 2,
        "nickname": "好友昵称",
        "is_blocked": false,
        "is_muted": false,
        "created_at": "2024-12-19T10:30:00Z",
        "contact": {
          "id": 2,
          "username": "friend",
          "nickname": "好友",
          "avatar": "https://example.com/friend-avatar.jpg",
          "online": true
        }
      }
    ],
    "total": 1
  }
}
```

### 添加联系人

**POST** `/contacts`

添加新的联系人。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求参数**:
```json
{
  "phone": "13800138001",
  "nickname": "新好友"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "联系人添加成功",
  "data": {
    "id": 2,
    "user_id": 1,
    "contact_id": 3,
    "nickname": "新好友",
    "is_blocked": false,
    "is_muted": false,
    "created_at": "2024-12-19T10:30:00Z"
  }
}
```

### 删除联系人

**DELETE** `/contacts/{contact_id}`

删除指定联系人。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "联系人删除成功",
  "data": {}
}
```

## 💬 聊天管理

### 获取聊天列表

**GET** `/chats`

获取当前用户的聊天列表。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `limit`: 每页数量 (默认: 50)
- `offset`: 偏移量 (默认: 0)

**响应**:
```json
{
  "code": 200,
  "message": "获取聊天列表成功",
  "data": {
    "chats": [
      {
        "id": 1,
        "name": "群聊名称",
        "description": "群聊描述",
        "avatar": "https://example.com/chat-avatar.jpg",
        "type": "group",
        "is_active": true,
        "is_pinned": false,
        "is_muted": false,
        "members_count": 5,
        "created_at": "2024-12-19T10:30:00Z",
        "updated_at": "2024-12-19T10:30:00Z",
        "last_message": {
          "id": 100,
          "content": "最后一条消息",
          "type": "text",
          "sender_id": 2,
          "created_at": "2024-12-19T10:30:00Z"
        }
      }
    ],
    "total": 1,
    "has_more": false
  }
}
```

### 创建聊天

**POST** `/chats`

创建新的聊天。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求参数**:
```json
{
  "name": "新群聊",
  "description": "群聊描述",
  "type": "group",
  "members": [2, 3, 4]
}
```

**响应**:
```json
{
  "code": 200,
  "message": "聊天创建成功",
  "data": {
    "id": 2,
    "name": "新群聊",
    "description": "群聊描述",
    "avatar": "",
    "type": "group",
    "is_active": true,
    "is_pinned": false,
    "is_muted": false,
    "members_count": 4,
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z",
    "members": [
      {
        "id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "role": "owner"
      }
    ]
  }
}
```

### 获取聊天详情

**GET** `/chats/{chat_id}`

获取指定聊天的详细信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "获取聊天详情成功",
  "data": {
    "id": 1,
    "name": "群聊名称",
    "description": "群聊描述",
    "avatar": "https://example.com/chat-avatar.jpg",
    "type": "group",
    "is_active": true,
    "is_pinned": false,
    "is_muted": false,
    "members_count": 5,
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z",
    "members": [
      {
        "id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "avatar": "https://example.com/avatar.jpg",
        "role": "owner",
        "joined_at": "2024-12-19T10:30:00Z"
      }
    ]
  }
}
```

## 📨 消息管理

### 获取消息列表

**GET** `/chats/{chat_id}/messages`

获取指定聊天的消息列表。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `limit`: 每页数量 (默认: 50)
- `offset`: 偏移量 (默认: 0)
- `before`: 获取此消息之前的消息
- `after`: 获取此消息之后的消息
- `search`: 搜索关键词
- `type`: 消息类型过滤

**响应**:
```json
{
  "code": 200,
  "message": "获取消息列表成功",
  "data": {
    "messages": [
      {
        "id": 100,
        "chat_id": 1,
        "sender_id": 1,
        "content": "这是一条测试消息",
        "type": "text",
        "file_name": "",
        "file_size": 0,
        "file_url": "",
        "thumbnail": "",
        "is_read": true,
        "is_edited": false,
        "is_deleted": false,
        "is_pinned": false,
        "reply_to_id": null,
        "forward_from": null,
        "ttl": 0,
        "send_at": null,
        "is_silent": false,
        "created_at": "2024-12-19T10:30:00Z",
        "updated_at": "2024-12-19T10:30:00Z",
        "sender": {
          "id": 1,
          "username": "testuser",
          "nickname": "测试用户",
          "avatar": "https://example.com/avatar.jpg"
        }
      }
    ],
    "total": 1,
    "has_more": false
  }
}
```

### 发送消息

**POST** `/chats/{chat_id}/messages`

向指定聊天发送消息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求参数**:
```json
{
  "content": "这是一条测试消息",
  "type": "text",
  "file_name": "",
  "file_size": 0,
  "file_url": "",
  "thumbnail": "",
  "reply_to_id": null,
  "forward_from": null,
  "ttl": 0,
  "send_at": null,
  "is_silent": false
}
```

**响应**:
```json
{
  "code": 200,
  "message": "消息发送成功",
  "data": {
    "id": 101,
    "chat_id": 1,
    "sender_id": 1,
    "content": "这是一条测试消息",
    "type": "text",
    "file_name": "",
    "file_size": 0,
    "file_url": "",
    "thumbnail": "",
    "is_read": false,
    "is_edited": false,
    "is_deleted": false,
    "is_pinned": false,
    "reply_to_id": null,
    "forward_from": null,
    "ttl": 0,
    "send_at": null,
    "is_silent": false,
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z",
    "sender": {
      "id": 1,
      "username": "testuser",
      "nickname": "测试用户",
      "avatar": "https://example.com/avatar.jpg"
    }
  }
}
```

### 编辑消息

**PUT** `/messages/{message_id}`

编辑指定消息的内容。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求参数**:
```json
{
  "content": "编辑后的消息内容"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "消息编辑成功",
  "data": {
    "id": 101,
    "chat_id": 1,
    "sender_id": 1,
    "content": "编辑后的消息内容",
    "type": "text",
    "is_read": false,
    "is_edited": true,
    "is_deleted": false,
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:31:00Z"
  }
}
```

### 删除消息

**DELETE** `/messages/{message_id}`

删除指定消息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "消息删除成功",
  "data": {}
}
```

### 标记消息为已读

**POST** `/messages/{message_id}/read`

标记指定消息为已读。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "消息已标记为已读",
  "data": {}
}
```

### 置顶消息

**POST** `/messages/{message_id}/pin`

置顶指定消息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "消息已置顶",
  "data": {}
}
```

### 取消置顶消息

**DELETE** `/messages/{message_id}/pin`

取消置顶指定消息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "消息已取消置顶",
  "data": {}
}
```

## 📁 文件管理

### 上传文件

**POST** `/files/upload`

上传文件到服务器。

**请求头**:
```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**请求参数**:
- `file`: 文件数据
- `type`: 文件类型 (image/video/audio/document)

**响应**:
```json
{
  "code": 200,
  "message": "文件上传成功",
  "data": {
    "id": "file_123456789",
    "name": "example.jpg",
    "size": 1024000,
    "type": "image",
    "url": "https://example.com/files/file_123456789",
    "thumbnail": "https://example.com/files/thumb_file_123456789",
    "created_at": "2024-12-19T10:30:00Z"
  }
}
```

### 下载文件

**GET** `/files/{file_id}`

下载指定文件。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
文件二进制数据

### 删除文件

**DELETE** `/files/{file_id}`

删除指定文件。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "message": "文件删除成功",
  "data": {}
}
```

## 🔌 WebSocket 事件

### 连接建立

连接到 WebSocket 服务器后，会收到连接确认事件。

**事件类型**: `connect`

**数据**:
```json
{
  "type": "connect",
  "data": {
    "status": "connected",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

### 新消息事件

当收到新消息时触发。

**事件类型**: `message.new`

**数据**:
```json
{
  "type": "message.new",
  "data": {
    "id": 101,
    "chat_id": 1,
    "sender_id": 2,
    "content": "这是一条新消息",
    "type": "text",
    "created_at": "2024-12-19T10:30:00Z",
    "sender": {
      "id": 2,
      "username": "otheruser",
      "nickname": "其他用户",
      "avatar": "https://example.com/avatar.jpg"
    }
  }
}
```

### 消息编辑事件

当消息被编辑时触发。

**事件类型**: `message.edit`

**数据**:
```json
{
  "type": "message.edit",
  "data": {
    "id": 101,
    "chat_id": 1,
    "content": "编辑后的消息内容",
    "updated_at": "2024-12-19T10:31:00Z"
  }
}
```

### 消息删除事件

当消息被删除时触发。

**事件类型**: `message.delete`

**数据**:
```json
{
  "type": "message.delete",
  "data": {
    "id": 101,
    "chat_id": 1,
    "deleted_at": "2024-12-19T10:31:00Z"
  }
}
```

### 正在输入事件

当用户正在输入时触发。

**事件类型**: `typing`

**数据**:
```json
{
  "type": "typing",
  "data": {
    "chat_id": 1,
    "user_id": 2,
    "is_typing": true,
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

### 在线状态事件

当用户在线状态改变时触发。

**事件类型**: `presence`

**数据**:
```json
{
  "type": "presence",
  "data": {
    "user_id": 2,
    "is_online": true,
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

### 通话相关事件

#### 通话邀请

**事件类型**: `call.offer`

**数据**:
```json
{
  "type": "call.offer",
  "data": {
    "call_id": "call_123456",
    "from_user_id": 2,
    "to_user_id": 1,
    "call_type": "video",
    "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

#### 通话应答

**事件类型**: `call.answer`

**数据**:
```json
{
  "type": "call.answer",
  "data": {
    "call_id": "call_123456",
    "from_user_id": 1,
    "to_user_id": 2,
    "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n...",
    "timestamp": "2024-12-19T10:30:05Z"
  }
}
```

#### ICE 候选

**事件类型**: `call.ice`

**数据**:
```json
{
  "type": "call.ice",
  "data": {
    "call_id": "call_123456",
    "from_user_id": 2,
    "to_user_id": 1,
    "candidate": "candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

#### 通话结束

**事件类型**: `call.end`

**数据**:
```json
{
  "type": "call.end",
  "data": {
    "call_id": "call_123456",
    "reason": "user_hangup",
    "duration": 120,
    "timestamp": "2024-12-19T10:32:00Z"
  }
}
```

## ❌ 错误码

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 数据验证失败 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

### 业务错误码

| 错误码 | 说明 |
|--------|------|
| 10001 | 用户不存在 |
| 10002 | 密码错误 |
| 10003 | 验证码错误 |
| 10004 | 用户已被禁用 |
| 10005 | 手机号已存在 |
| 10006 | 用户名已存在 |
| 10007 | 令牌已过期 |
| 10008 | 令牌无效 |
| 20001 | 聊天不存在 |
| 20002 | 用户不是聊天成员 |
| 20003 | 消息不存在 |
| 20004 | 无权限操作此消息 |
| 30001 | 文件不存在 |
| 30002 | 文件类型不支持 |
| 30003 | 文件大小超限 |
| 40001 | 请求过于频繁 |
| 50001 | 服务器内部错误 |

### 错误响应示例

```json
{
  "code": 40001,
  "message": "请求过于频繁",
  "error": "您发送消息过于频繁，请稍后再试",
  "timestamp": "2024-12-19T10:30:00Z"
}
```


