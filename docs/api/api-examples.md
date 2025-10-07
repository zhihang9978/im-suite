# IM-Suite API 使用示例

## 概述

本文档提供了 IM-Suite API 的详细使用示例，包括 REST API 调用和 WebSocket 事件处理。

## 基础配置

### 环境变量
```bash
# 开发环境
API_BASE_URL=http://localhost:8080/api
WS_BASE_URL=ws://localhost:8080/ws

# 生产环境
API_BASE_URL=https://api.im-suite.com/api
WS_BASE_URL=wss://api.im-suite.com/ws
```

### 认证头
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

## REST API 示例

### 1. 用户认证

#### 用户登录
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+8613800000000",
    "code": "123456"
  }'
```

响应：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "phone": "+8613800000000",
    "username": "zhangsan",
    "nickname": "张三",
    "avatar_url": "https://example.com/avatar.jpg",
    "bio": "这是我的个人简介",
    "is_online": true,
    "last_seen": "2025-10-07T12:00:00Z",
    "created_at": "2025-10-07T10:00:00Z",
    "updated_at": "2025-10-07T12:00:00Z"
  },
  "expires_in": 3600
}
```

#### 刷新令牌
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

#### 用户登出
```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 2. 用户管理

#### 获取当前用户信息
```bash
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 更新用户信息
```bash
curl -X PUT http://localhost:8080/api/users/me \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "nickname": "张三",
    "bio": "这是我的新简介",
    "avatar_url": "https://example.com/new_avatar.jpg"
  }'
```

#### 获取其他用户信息
```bash
curl -X GET http://localhost:8080/api/users/2 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 3. 联系人管理

#### 获取联系人列表
```bash
curl -X GET "http://localhost:8080/api/contacts?page=1&limit=20&search=张" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

响应：
```json
{
  "contacts": [
    {
      "id": 1,
      "user_id": 1,
      "contact_id": 2,
      "nickname": "李四",
      "remark": "同事",
      "created_at": "2025-10-07T10:00:00Z"
    }
  ],
  "users": [
    {
      "id": 2,
      "phone": "+8613800000001",
      "username": "lisi",
      "nickname": "李四",
      "avatar_url": "https://example.com/avatar2.jpg",
      "is_online": true,
      "last_seen": "2025-10-07T12:00:00Z"
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
```

#### 添加联系人
```bash
curl -X POST http://localhost:8080/api/contacts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "phone": "+8613800000001",
    "nickname": "李四",
    "remark": "同事"
  }'
```

#### 删除联系人
```bash
curl -X DELETE http://localhost:8080/api/contacts/1 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 4. 聊天管理

#### 获取聊天列表
```bash
curl -X GET "http://localhost:8080/api/chats?page=1&limit=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

响应：
```json
{
  "chats": [
    {
      "id": 1,
      "type": "private",
      "title": "李四",
      "description": "这是一个私聊",
      "avatar_url": "https://example.com/avatar2.jpg",
      "created_by": 1,
      "last_message": {
        "id": 123,
        "chat_id": 1,
        "sender_id": 2,
        "content": "你好，这是一条消息",
        "message_type": "text",
        "created_at": "2025-10-07T12:00:00Z"
      },
      "unread_count": 5,
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
```

### 5. 消息管理

#### 获取聊天消息
```bash
curl -X GET "http://localhost:8080/api/chats/1/messages?limit=20&before=123" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

响应：
```json
{
  "messages": [
    {
      "id": 123,
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
```

#### 发送文本消息
```bash
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "你好，这是一条测试消息",
    "message_type": "text"
  }'
```

#### 发送图片消息
```bash
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
```

#### 发送阅后即焚消息
```bash
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "这是阅后即焚消息",
    "message_type": "text",
    "ttl_seconds": 60
  }'
```

#### 发送定时消息
```bash
curl -X POST http://localhost:8080/api/chats/1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "这是定时消息",
    "message_type": "text",
    "send_at": "2025-10-07T15:00:00Z"
  }'
```

#### 编辑消息
```bash
curl -X PUT http://localhost:8080/api/messages/123 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "content": "这是编辑后的消息"
  }'
```

#### 删除消息
```bash
curl -X DELETE http://localhost:8080/api/messages/123 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 标记消息已读
```bash
curl -X POST http://localhost:8080/api/messages/123/read \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 置顶聊天
```bash
curl -X POST http://localhost:8080/api/chats/1/pin \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 取消置顶聊天
```bash
curl -X DELETE http://localhost:8080/api/chats/1/pin \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## WebSocket 示例

### 1. 连接 WebSocket

#### JavaScript
```javascript
const ws = new WebSocket('ws://localhost:8080/ws', [], {
  headers: {
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
  }
});

ws.onopen = function(event) {
  console.log('WebSocket 连接已建立');
};

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
  
  switch(data.type) {
    case 'message.new':
      handleNewMessage(data.payload);
      break;
    case 'message.edit':
      handleMessageEdit(data.payload);
      break;
    case 'message.delete':
      handleMessageDelete(data.payload);
      break;
    case 'typing':
      handleTyping(data.payload);
      break;
    case 'presence':
      handlePresence(data.payload);
      break;
    case 'call.offer':
      handleCallOffer(data.payload);
      break;
    case 'call.answer':
      handleCallAnswer(data.payload);
      break;
    case 'call.ice':
      handleCallIce(data.payload);
      break;
    case 'call.end':
      handleCallEnd(data.payload);
      break;
  }
};

ws.onclose = function(event) {
  console.log('WebSocket 连接已关闭:', event.code, event.reason);
};

ws.onerror = function(error) {
  console.error('WebSocket 错误:', error);
};
```

#### Android (Kotlin)
```kotlin
val wsUrl = "ws://localhost:8080/ws"
val request = Request.Builder()
    .url(wsUrl)
    .addHeader("Authorization", "Bearer $token")
    .build()

val webSocket = client.newWebSocket(request, object : WebSocketListener() {
    override fun onOpen(webSocket: WebSocket, response: Response) {
        Log.d("WebSocket", "连接已建立")
    }
    
    override fun onMessage(webSocket: WebSocket, text: String) {
        val data = JSONObject(text)
        Log.d("WebSocket", "收到消息: $data")
        
        when (data.getString("type")) {
            "message.new" -> handleNewMessage(data.getJSONObject("payload"))
            "message.edit" -> handleMessageEdit(data.getJSONObject("payload"))
            "message.delete" -> handleMessageDelete(data.getJSONObject("payload"))
            "typing" -> handleTyping(data.getJSONObject("payload"))
            "presence" -> handlePresence(data.getJSONObject("payload"))
            "call.offer" -> handleCallOffer(data.getJSONObject("payload"))
            "call.answer" -> handleCallAnswer(data.getJSONObject("payload"))
            "call.ice" -> handleCallIce(data.getJSONObject("payload"))
            "call.end" -> handleCallEnd(data.getJSONObject("payload"))
        }
    }
    
    override fun onClose(webSocket: WebSocket, code: Int, reason: String) {
        Log.d("WebSocket", "连接已关闭: $code $reason")
    }
    
    override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
        Log.e("WebSocket", "连接失败", t)
    }
})
```

### 2. 发送消息

#### 发送文本消息
```javascript
const message = {
  type: 'message.send',
  payload: {
    chat_id: 1,
    content: '你好，这是一条消息',
    message_type: 'text'
  }
};

ws.send(JSON.stringify(message));
```

#### 发送正在输入状态
```javascript
const typing = {
  type: 'typing',
  payload: {
    chat_id: 1,
    is_typing: true
  }
};

ws.send(JSON.stringify(typing));
```

#### 发送在线状态
```javascript
const presence = {
  type: 'presence',
  payload: {
    is_online: true
  }
};

ws.send(JSON.stringify(presence));
```

### 3. 通话信令

#### 发起通话
```javascript
const callOffer = {
  type: 'call.offer',
  payload: {
    to_user: 2,
    call_type: 'video',
    sdp: 'v=0\r\no=- 1234567890 1234567890 IN IP4 127.0.0.1\r\n...',
    ice_candidates: [
      {
        candidate: 'candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host',
        sdpMLineIndex: 0,
        sdpMid: '0'
      }
    ]
  }
};

ws.send(JSON.stringify(callOffer));
```

#### 应答通话
```javascript
const callAnswer = {
  type: 'call.answer',
  payload: {
    call_id: 'call_123456',
    sdp: 'v=0\r\no=- 1234567890 1234567890 IN IP4 127.0.0.1\r\n...',
    ice_candidates: [
      {
        candidate: 'candidate:1 1 UDP 2113667326 192.168.1.101 54401 typ host',
        sdpMLineIndex: 0,
        sdpMid: '0'
      }
    ]
  }
};

ws.send(JSON.stringify(callAnswer));
```

#### ICE 候选交换
```javascript
const callIce = {
  type: 'call.ice',
  payload: {
    call_id: 'call_123456',
    ice_candidate: {
      candidate: 'candidate:2 1 UDP 1694498814 203.0.113.1 54402 typ srflx',
      sdpMLineIndex: 0,
      sdpMid: '0'
    }
  }
};

ws.send(JSON.stringify(callIce));
```

#### 结束通话
```javascript
const callEnd = {
  type: 'call.end',
  payload: {
    call_id: 'call_123456',
    reason: 'user_hangup'
  }
};

ws.send(JSON.stringify(callEnd));
```

## 错误处理

### REST API 错误响应
```json
{
  "error": "VALIDATION_ERROR",
  "message": "请求参数验证失败",
  "details": {
    "field": "phone",
    "reason": "手机号格式不正确"
  }
}
```

### WebSocket 错误事件
```json
{
  "type": "error",
  "payload": {
    "error_code": "AUTHENTICATION_FAILED",
    "message": "认证失败",
    "details": {
      "reason": "令牌已过期"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

## 最佳实践

### 1. 认证管理
- 定期检查令牌有效性
- 实现自动刷新机制
- 处理认证失败情况

### 2. 错误处理
- 实现重试机制
- 处理网络错误
- 显示用户友好的错误信息

### 3. 性能优化
- 实现消息分页
- 使用 WebSocket 进行实时更新
- 缓存用户和聊天信息

### 4. 安全考虑
- 验证所有输入数据
- 使用 HTTPS/WSS
- 实现速率限制
- 保护敏感信息

### 5. 用户体验
- 显示连接状态
- 实现离线消息
- 提供消息状态反馈
- 支持消息搜索和过滤
