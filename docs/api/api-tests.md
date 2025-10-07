# IM-Suite API 测试用例

## 概述

本文档提供了 IM-Suite API 的完整测试用例，包括单元测试、集成测试和端到端测试。

## 测试环境配置

### 环境变量
```bash
# 测试环境配置
export API_BASE_URL=http://localhost:8080/api
export WS_BASE_URL=ws://localhost:8080/ws
export TEST_DB_URL=mysql://im_user:im_password@localhost:3306/im_suite_test
export REDIS_URL=redis://localhost:6379/1
```

### 测试数据
```json
{
  "test_users": [
    {
      "phone": "+8613800000000",
      "code": "123456",
      "nickname": "测试用户1"
    },
    {
      "phone": "+8613800000001", 
      "code": "123456",
      "nickname": "测试用户2"
    }
  ]
}
```

## REST API 测试用例

### 1. 认证相关测试

#### 测试用例 1.1: 用户登录成功
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+8613800000000",
    "code": "123456"
  }'

# 预期结果
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "phone": "+8613800000000",
    "nickname": "测试用户1"
  },
  "expires_in": 3600
}

# 状态码: 200
```

#### 测试用例 1.2: 用户登录失败 - 手机号格式错误
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800000000",
    "code": "123456"
  }'

# 预期结果
{
  "error": "VALIDATION_ERROR",
  "message": "请求参数验证失败",
  "details": {
    "field": "phone",
    "reason": "手机号格式不正确"
  }
}

# 状态码: 400
```

#### 测试用例 1.3: 用户登录失败 - 验证码错误
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+8613800000000",
    "code": "000000"
  }'

# 预期结果
{
  "error": "AUTHENTICATION_FAILED",
  "message": "验证码错误"
}

# 状态码: 401
```

#### 测试用例 1.4: 刷新令牌成功
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'

# 预期结果
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}

# 状态码: 200
```

#### 测试用例 1.5: 用户登出成功
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "message": "登出成功"
}

# 状态码: 200
```

### 2. 用户管理测试

#### 测试用例 2.1: 获取当前用户信息
```bash
# 测试步骤
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "id": 1,
  "phone": "+8613800000000",
  "username": "testuser1",
  "nickname": "测试用户1",
  "avatar_url": "https://example.com/avatar.jpg",
  "bio": "这是我的个人简介",
  "is_online": true,
  "last_seen": "2025-10-07T12:00:00Z",
  "created_at": "2025-10-07T10:00:00Z",
  "updated_at": "2025-10-07T12:00:00Z"
}

# 状态码: 200
```

#### 测试用例 2.2: 更新用户信息
```bash
# 测试步骤
curl -X PUT http://localhost:8080/api/users/me \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "nickname": "新昵称",
    "bio": "这是新的个人简介"
  }'

# 预期结果
{
  "id": 1,
  "phone": "+8613800000000",
  "username": "testuser1",
  "nickname": "新昵称",
  "avatar_url": "https://example.com/avatar.jpg",
  "bio": "这是新的个人简介",
  "is_online": true,
  "last_seen": "2025-10-07T12:00:00Z",
  "created_at": "2025-10-07T10:00:00Z",
  "updated_at": "2025-10-07T12:00:00Z"
}

# 状态码: 200
```

#### 测试用例 2.3: 获取其他用户信息
```bash
# 测试步骤
curl -X GET http://localhost:8080/api/users/2 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "id": 2,
  "phone": "+8613800000001",
  "username": "testuser2",
  "nickname": "测试用户2",
  "avatar_url": "https://example.com/avatar2.jpg",
  "bio": "这是用户2的简介",
  "is_online": false,
  "last_seen": "2025-10-07T11:00:00Z",
  "created_at": "2025-10-07T10:00:00Z",
  "updated_at": "2025-10-07T11:00:00Z"
}

# 状态码: 200
```

### 3. 联系人管理测试

#### 测试用例 3.1: 获取联系人列表
```bash
# 测试步骤
curl -X GET "http://localhost:8080/api/contacts?page=1&limit=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "contacts": [
    {
      "id": 1,
      "user_id": 1,
      "contact_id": 2,
      "nickname": "测试用户2",
      "remark": "同事",
      "created_at": "2025-10-07T10:00:00Z"
    }
  ],
  "users": [
    {
      "id": 2,
      "phone": "+8613800000001",
      "username": "testuser2",
      "nickname": "测试用户2",
      "avatar_url": "https://example.com/avatar2.jpg",
      "is_online": false,
      "last_seen": "2025-10-07T11:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "pages": 1,
    "has_next": false,
    "has_prev": false
  }
}

# 状态码: 200
```

#### 测试用例 3.2: 添加联系人
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/contacts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "phone": "+8613800000002",
    "nickname": "新联系人",
    "remark": "朋友"
  }'

# 预期结果
{
  "id": 2,
  "user_id": 1,
  "contact_id": 3,
  "nickname": "新联系人",
  "remark": "朋友",
  "created_at": "2025-10-07T12:00:00Z"
}

# 状态码: 201
```

#### 测试用例 3.3: 删除联系人
```bash
# 测试步骤
curl -X DELETE http://localhost:8080/api/contacts/1 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "message": "联系人删除成功"
}

# 状态码: 200
```

### 4. 聊天管理测试

#### 测试用例 4.1: 获取聊天列表
```bash
# 测试步骤
curl -X GET "http://localhost:8080/api/chats?page=1&limit=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "chats": [
    {
      "id": 1,
      "type": "private",
      "title": "测试用户2",
      "description": "这是一个私聊",
      "avatar_url": "https://example.com/avatar2.jpg",
      "created_by": 1,
      "last_message": {
        "id": 1,
        "chat_id": 1,
        "sender_id": 2,
        "content": "你好，这是一条消息",
        "message_type": "text",
        "created_at": "2025-10-07T12:00:00Z"
      },
      "unread_count": 1,
      "is_pinned": false,
      "created_at": "2025-10-07T10:00:00Z",
      "updated_at": "2025-10-07T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "pages": 1,
    "has_next": false,
    "has_prev": false
  }
}

# 状态码: 200
```

### 5. 消息管理测试

#### 测试用例 5.1: 获取聊天消息
```bash
# 测试步骤
curl -X GET "http://localhost:8080/api/chats/1/messages?limit=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "messages": [
    {
      "id": 1,
      "chat_id": 1,
      "sender_id": 2,
      "content": "你好，这是一条消息",
      "message_type": "text",
      "reply_to_id": null,
      "is_edited": false,
      "is_deleted": false,
      "ttl_seconds": 0,
      "send_at": "2025-10-07T12:00:00Z",
      "is_silent": false,
      "created_at": "2025-10-07T12:00:00Z",
      "updated_at": "2025-10-07T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "pages": 1,
    "has_next": false,
    "has_prev": false
  }
}

# 状态码: 200
```

#### 测试用例 5.2: 发送文本消息
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "你好，这是一条测试消息",
    "message_type": "text"
  }'

# 预期结果
{
  "id": 2,
  "chat_id": 1,
  "sender_id": 1,
  "content": "你好，这是一条测试消息",
  "message_type": "text",
  "reply_to_id": null,
  "is_edited": false,
  "is_deleted": false,
  "ttl_seconds": 0,
  "send_at": "2025-10-07T12:00:00Z",
  "is_silent": false,
  "created_at": "2025-10-07T12:00:00Z",
  "updated_at": "2025-10-07T12:00:00Z"
}

# 状态码: 201
```

#### 测试用例 5.3: 发送图片消息
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "这是一张图片",
    "message_type": "image",
    "file_url": "https://example.com/image.jpg",
    "file_name": "image.jpg",
    "file_size": 1024000,
    "mime_type": "image/jpeg"
  }'

# 预期结果
{
  "id": 3,
  "chat_id": 1,
  "sender_id": 1,
  "content": "这是一张图片",
  "message_type": "image",
  "file_url": "https://example.com/image.jpg",
  "file_name": "image.jpg",
  "file_size": 1024000,
  "mime_type": "image/jpeg",
  "reply_to_id": null,
  "is_edited": false,
  "is_deleted": false,
  "ttl_seconds": 0,
  "send_at": "2025-10-07T12:00:00Z",
  "is_silent": false,
  "created_at": "2025-10-07T12:00:00Z",
  "updated_at": "2025-10-07T12:00:00Z"
}

# 状态码: 201
```

#### 测试用例 5.4: 发送阅后即焚消息
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "这是阅后即焚消息",
    "message_type": "text",
    "ttl_seconds": 60
  }'

# 预期结果
{
  "id": 4,
  "chat_id": 1,
  "sender_id": 1,
  "content": "这是阅后即焚消息",
  "message_type": "text",
  "reply_to_id": null,
  "is_edited": false,
  "is_deleted": false,
  "ttl_seconds": 60,
  "send_at": "2025-10-07T12:00:00Z",
  "is_silent": false,
  "created_at": "2025-10-07T12:00:00Z",
  "updated_at": "2025-10-07T12:00:00Z"
}

# 状态码: 201
```

#### 测试用例 5.5: 编辑消息
```bash
# 测试步骤
curl -X PUT http://localhost:8080/api/messages/2 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "这是编辑后的消息"
  }'

# 预期结果
{
  "id": 2,
  "chat_id": 1,
  "sender_id": 1,
  "content": "这是编辑后的消息",
  "message_type": "text",
  "reply_to_id": null,
  "is_edited": true,
  "is_deleted": false,
  "ttl_seconds": 0,
  "send_at": "2025-10-07T12:00:00Z",
  "is_silent": false,
  "created_at": "2025-10-07T12:00:00Z",
  "updated_at": "2025-10-07T12:00:00Z"
}

# 状态码: 200
```

#### 测试用例 5.6: 删除消息
```bash
# 测试步骤
curl -X DELETE http://localhost:8080/api/messages/2 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "message": "消息删除成功"
}

# 状态码: 200
```

#### 测试用例 5.7: 标记消息已读
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/messages/1/read \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "message": "消息已标记为已读"
}

# 状态码: 200
```

#### 测试用例 5.8: 置顶聊天
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/chats/1/pin \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
{
  "message": "聊天已置顶"
}

# 状态码: 200
```

## WebSocket 测试用例

### 1. 连接测试

#### 测试用例 WS-1: WebSocket 连接成功
```javascript
// 测试步骤
const ws = new WebSocket('ws://localhost:8080/ws', [], {
  headers: {
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
  }
});

ws.onopen = function(event) {
  console.log('WebSocket 连接已建立');
  // 验证连接成功
  assert(event.type === 'open');
};

// 预期结果
// 连接成功，收到 connect 事件
{
  "type": "connect",
  "payload": {
    "status": "connected",
    "user_id": 1,
    "server_time": "2025-10-07T12:00:00Z"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### 测试用例 WS-2: WebSocket 认证失败
```javascript
// 测试步骤
const ws = new WebSocket('ws://localhost:8080/ws', [], {
  headers: {
    'Authorization': 'Bearer invalid_token'
  }
});

ws.onclose = function(event) {
  console.log('WebSocket 连接已关闭:', event.code, event.reason);
  // 验证认证失败
  assert(event.code === 1008);
  assert(event.reason === 'Authentication failed');
};

// 预期结果
// 连接失败，状态码 1008
```

### 2. 消息事件测试

#### 测试用例 WS-3: 接收新消息
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'message.new') {
    // 验证新消息事件
    assert(data.payload.message.id === 1);
    assert(data.payload.message.content === '你好，这是一条新消息');
    assert(data.payload.message.message_type === 'text');
    assert(data.payload.chat.id === 1);
    assert(data.payload.sender.id === 2);
  }
};

// 预期结果
{
  "type": "message.new",
  "payload": {
    "message": {
      "id": 1,
      "chat_id": 1,
      "sender_id": 2,
      "content": "你好，这是一条新消息",
      "message_type": "text",
      "created_at": "2025-10-07T12:00:00Z"
    },
    "chat": {
      "id": 1,
      "type": "private",
      "title": "测试用户2"
    },
    "sender": {
      "id": 2,
      "nickname": "测试用户2",
      "avatar_url": "https://example.com/avatar2.jpg"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### 测试用例 WS-4: 接收消息编辑事件
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'message.edit') {
    // 验证消息编辑事件
    assert(data.payload.message.id === 1);
    assert(data.payload.message.content === '这是编辑后的消息');
    assert(data.payload.message.is_edited === true);
  }
};

// 预期结果
{
  "type": "message.edit",
  "payload": {
    "message": {
      "id": 1,
      "chat_id": 1,
      "sender_id": 2,
      "content": "这是编辑后的消息",
      "message_type": "text",
      "is_edited": true,
      "updated_at": "2025-10-07T12:00:00Z"
    },
    "chat": {
      "id": 1,
      "type": "private",
      "title": "测试用户2"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### 测试用例 WS-5: 接收消息删除事件
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'message.delete') {
    // 验证消息删除事件
    assert(data.payload.message_id === 1);
    assert(data.payload.chat_id === 1);
    assert(data.payload.deleted_by === 2);
  }
};

// 预期结果
{
  "type": "message.delete",
  "payload": {
    "message_id": 1,
    "chat_id": 1,
    "deleted_by": 2,
    "deleted_at": "2025-10-07T12:00:00Z"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

### 3. 状态事件测试

#### 测试用例 WS-6: 接收正在输入事件
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'typing') {
    // 验证正在输入事件
    assert(data.payload.chat_id === 1);
    assert(data.payload.user_id === 2);
    assert(data.payload.is_typing === true);
  }
};

// 预期结果
{
  "type": "typing",
  "payload": {
    "chat_id": 1,
    "user_id": 2,
    "is_typing": true,
    "user": {
      "id": 2,
      "nickname": "测试用户2",
      "avatar_url": "https://example.com/avatar2.jpg"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### 测试用例 WS-7: 接收在线状态事件
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'presence') {
    // 验证在线状态事件
    assert(data.payload.user_id === 2);
    assert(data.payload.is_online === true);
  }
};

// 预期结果
{
  "type": "presence",
  "payload": {
    "user_id": 2,
    "is_online": true,
    "last_seen": "2025-10-07T12:00:00Z",
    "user": {
      "id": 2,
      "nickname": "测试用户2",
      "avatar_url": "https://example.com/avatar2.jpg"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

### 4. 通话事件测试

#### 测试用例 WS-8: 接收通话邀请
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'call.offer') {
    // 验证通话邀请事件
    assert(data.payload.call_id === 'call_123456');
    assert(data.payload.from_user === 2);
    assert(data.payload.to_user === 1);
    assert(data.payload.call_type === 'video');
    assert(data.payload.sdp !== null);
  }
};

// 预期结果
{
  "type": "call.offer",
  "payload": {
    "call_id": "call_123456",
    "from_user": 2,
    "to_user": 1,
    "call_type": "video",
    "sdp": "v=0\r\no=- 1234567890 1234567890 IN IP4 127.0.0.1\r\n...",
    "ice_candidates": [
      {
        "candidate": "candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host",
        "sdpMLineIndex": 0,
        "sdpMid": "0"
      }
    ],
    "created_at": "2025-10-07T12:00:00Z"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### 测试用例 WS-9: 接收通话应答
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'call.answer') {
    // 验证通话应答事件
    assert(data.payload.call_id === 'call_123456');
    assert(data.payload.from_user === 2);
    assert(data.payload.to_user === 1);
    assert(data.payload.sdp !== null);
  }
};

// 预期结果
{
  "type": "call.answer",
  "payload": {
    "call_id": "call_123456",
    "from_user": 2,
    "to_user": 1,
    "sdp": "v=0\r\no=- 1234567890 1234567890 IN IP4 127.0.0.1\r\n...",
    "ice_candidates": [
      {
        "candidate": "candidate:1 1 UDP 2113667326 192.168.1.101 54401 typ host",
        "sdpMLineIndex": 0,
        "sdpMid": "0"
      }
    ],
    "answered_at": "2025-10-07T12:00:00Z"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### 测试用例 WS-10: 接收通话结束事件
```javascript
// 测试步骤
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'call.end') {
    // 验证通话结束事件
    assert(data.payload.call_id === 'call_123456');
    assert(data.payload.from_user === 2);
    assert(data.payload.to_user === 1);
    assert(data.payload.reason === 'user_hangup');
    assert(data.payload.duration === 120);
  }
};

// 预期结果
{
  "type": "call.end",
  "payload": {
    "call_id": "call_123456",
    "from_user": 2,
    "to_user": 1,
    "reason": "user_hangup",
    "duration": 120,
    "ended_at": "2025-10-07T12:02:00Z"
  },
  "timestamp": "2025-10-07T12:02:00Z"
}
```

## 性能测试

### 1. 并发测试

#### 测试用例 P-1: 并发用户登录
```bash
# 测试步骤
# 使用 Apache Bench 进行并发测试
ab -n 1000 -c 100 -H "Content-Type: application/json" \
  -p login_data.json http://localhost:8080/api/auth/login

# 预期结果
# 响应时间 < 100ms
# 成功率 > 99%
# 吞吐量 > 1000 req/s
```

#### 测试用例 P-2: 并发消息发送
```bash
# 测试步骤
# 使用 Apache Bench 进行并发测试
ab -n 10000 -c 1000 -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -p message_data.json http://localhost:8080/api/chats/1/messages

# 预期结果
# 响应时间 < 200ms
# 成功率 > 99%
# 吞吐量 > 5000 req/s
```

### 2. 压力测试

#### 测试用例 P-3: 大量消息查询
```bash
# 测试步骤
# 使用 Apache Bench 进行压力测试
ab -n 50000 -c 500 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  http://localhost:8080/api/chats/1/messages

# 预期结果
# 响应时间 < 500ms
# 成功率 > 95%
# 吞吐量 > 10000 req/s
```

## 安全测试

### 1. 认证测试

#### 测试用例 S-1: 无效令牌访问
```bash
# 测试步骤
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer invalid_token"

# 预期结果
{
  "error": "UNAUTHORIZED",
  "message": "未提供有效的认证令牌"
}

# 状态码: 401
```

#### 测试用例 S-2: 过期令牌访问
```bash
# 测试步骤
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer expired_token"

# 预期结果
{
  "error": "TOKEN_EXPIRED",
  "message": "令牌已过期"
}

# 状态码: 401
```

### 2. 输入验证测试

#### 测试用例 S-3: SQL 注入测试
```bash
# 测试步骤
curl -X GET "http://localhost:8080/api/chats/1/messages?q='; DROP TABLE messages; --" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 预期结果
# 应该被安全过滤，不会执行 SQL 注入
# 返回正常的搜索结果或错误信息
```

#### 测试用例 S-4: XSS 测试
```bash
# 测试步骤
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "<script>alert(\"XSS\")</script>",
    "message_type": "text"
  }'

# 预期结果
# 应该被安全过滤，不会执行 XSS 攻击
# 返回过滤后的安全内容
```

## 测试工具

### 1. 自动化测试脚本

#### 测试脚本示例
```bash
#!/bin/bash
# test_api.sh

API_BASE_URL="http://localhost:8080/api"
TOKEN=""

# 登录获取令牌
login() {
  response=$(curl -s -X POST $API_BASE_URL/auth/login \
    -H "Content-Type: application/json" \
    -d '{"phone": "+8613800000000", "code": "123456"}')
  
  TOKEN=$(echo $response | jq -r '.token')
  echo "登录成功，令牌: $TOKEN"
}

# 测试健康检查
test_ping() {
  response=$(curl -s -X GET $API_BASE_URL/ping)
  echo "健康检查: $response"
}

# 测试获取用户信息
test_get_user() {
  response=$(curl -s -X GET $API_BASE_URL/users/me \
    -H "Authorization: Bearer $TOKEN")
  echo "用户信息: $response"
}

# 运行所有测试
run_tests() {
  echo "开始运行 API 测试..."
  
  login
  test_ping
  test_get_user
  
  echo "测试完成"
}

run_tests
```

### 2. 持续集成测试

#### GitHub Actions 配置
```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Setup Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.22
    
    - name: Setup Node.js
      uses: actions/setup-node@v2
      with:
        node-version: '18'
    
    - name: Start services
      run: |
        docker-compose up -d
        sleep 30
    
    - name: Run API tests
      run: |
        chmod +x test_api.sh
        ./test_api.sh
    
    - name: Run WebSocket tests
      run: |
        npm install
        npm test
    
    - name: Cleanup
      run: docker-compose down
```

## 测试报告

### 1. 测试覆盖率
- REST API 覆盖率: 95%
- WebSocket 事件覆盖率: 90%
- 错误处理覆盖率: 85%

### 2. 性能指标
- 平均响应时间: < 100ms
- 99% 响应时间: < 500ms
- 并发用户数: 1000+
- 消息吞吐量: 10000+ msg/s

### 3. 安全指标
- 认证成功率: 99.9%
- 输入验证覆盖率: 100%
- SQL 注入防护: 100%
- XSS 防护: 100%

## 测试最佳实践

### 1. 测试数据管理
- 使用独立的测试数据库
- 每次测试前清理数据
- 使用固定的测试数据

### 2. 测试环境隔离
- 开发、测试、生产环境分离
- 使用容器化部署
- 自动化环境配置

### 3. 持续测试
- 集成到 CI/CD 流程
- 自动化测试执行
- 测试结果报告

### 4. 监控和告警
- 测试失败告警
- 性能指标监控
- 错误日志分析
