# IM-Suite WebSocket 事件规范

## 概述

IM-Suite 使用 WebSocket 提供实时通讯功能，包括消息推送、状态更新、通话信令等。所有 WebSocket 事件都使用 JSON 格式进行数据交换。

## 连接信息

- **开发环境**: `ws://localhost:8080/ws`
- **生产环境**: `wss://api.im-suite.com/ws`
- **认证**: 连接时需要在 URL 中传递 JWT 令牌，或通过 HTTP 头传递

## 事件格式

所有 WebSocket 事件都遵循以下格式：

```json
{
  "type": "事件类型",
  "payload": {
    "事件数据"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

## 事件类型

### 1. 连接事件

#### `connect`
客户端连接成功时触发。

```json
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

#### `disconnect`
客户端断开连接时触发。

```json
{
  "type": "disconnect",
  "payload": {
    "status": "disconnected",
    "reason": "client_close"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

### 2. 消息事件

#### `message.new`
新消息到达时触发。

```json
{
  "type": "message.new",
  "payload": {
    "message": {
      "id": 123,
      "chat_id": 1,
      "sender_id": 2,
      "content": "你好，这是一条新消息",
      "message_type": "text",
      "created_at": "2025-10-07T12:00:00Z"
    },
    "chat": {
      "id": 1,
      "type": "private",
      "title": "张三"
    },
    "sender": {
      "id": 2,
      "nickname": "张三",
      "avatar_url": "https://example.com/avatar.jpg"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### `message.edit`
消息被编辑时触发。

```json
{
  "type": "message.edit",
  "payload": {
    "message": {
      "id": 123,
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
      "title": "张三"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### `message.delete`
消息被删除时触发。

```json
{
  "type": "message.delete",
  "payload": {
    "message_id": 123,
    "chat_id": 1,
    "deleted_by": 2,
    "deleted_at": "2025-10-07T12:00:00Z"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### `message.read`
消息被标记为已读时触发。

```json
{
  "type": "message.read",
  "payload": {
    "message_id": 123,
    "chat_id": 1,
    "read_by": 2,
    "read_at": "2025-10-07T12:00:00Z"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

### 3. 状态事件

#### `typing`
用户正在输入时触发。

```json
{
  "type": "typing",
  "payload": {
    "chat_id": 1,
    "user_id": 2,
    "is_typing": true,
    "user": {
      "id": 2,
      "nickname": "张三",
      "avatar_url": "https://example.com/avatar.jpg"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### `presence`
用户在线状态变化时触发。

```json
{
  "type": "presence",
  "payload": {
    "user_id": 2,
    "is_online": true,
    "last_seen": "2025-10-07T12:00:00Z",
    "user": {
      "id": 2,
      "nickname": "张三",
      "avatar_url": "https://example.com/avatar.jpg"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

### 4. 通话事件

#### `call.offer`
通话邀请时触发。

```json
{
  "type": "call.offer",
  "payload": {
    "call_id": "call_123456",
    "from_user": 1,
    "to_user": 2,
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

#### `call.answer`
通话应答时触发。

```json
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

#### `call.ice`
ICE 候选交换时触发。

```json
{
  "type": "call.ice",
  "payload": {
    "call_id": "call_123456",
    "from_user": 1,
    "to_user": 2,
    "ice_candidate": {
      "candidate": "candidate:2 1 UDP 1694498814 203.0.113.1 54402 typ srflx",
      "sdpMLineIndex": 0,
      "sdpMid": "0"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### `call.end`
通话结束时触发。

```json
{
  "type": "call.end",
  "payload": {
    "call_id": "call_123456",
    "from_user": 1,
    "to_user": 2,
    "reason": "user_hangup",
    "duration": 120,
    "ended_at": "2025-10-07T12:02:00Z"
  },
  "timestamp": "2025-10-07T12:02:00Z"
}
```

### 5. 系统事件

#### `error`
系统错误时触发。

```json
{
  "type": "error",
  "payload": {
    "error_code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": {
      "field": "phone",
      "reason": "手机号格式不正确"
    }
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

#### `notification`
系统通知时触发。

```json
{
  "type": "notification",
  "payload": {
    "title": "系统通知",
    "message": "服务器将在 10 分钟后进行维护",
    "type": "info",
    "action_url": "https://example.com/maintenance"
  },
  "timestamp": "2025-10-07T12:00:00Z"
}
```

## 客户端发送事件

### 1. 消息相关

#### 发送消息
```json
{
  "type": "message.send",
  "payload": {
    "chat_id": 1,
    "content": "你好，这是一条消息",
    "message_type": "text",
    "reply_to_id": 123
  }
}
```

#### 编辑消息
```json
{
  "type": "message.edit",
  "payload": {
    "message_id": 123,
    "content": "这是编辑后的消息"
  }
}
```

#### 删除消息
```json
{
  "type": "message.delete",
  "payload": {
    "message_id": 123
  }
}
```

#### 标记已读
```json
{
  "type": "message.read",
  "payload": {
    "message_id": 123
  }
}
```

### 2. 状态相关

#### 正在输入
```json
{
  "type": "typing",
  "payload": {
    "chat_id": 1,
    "is_typing": true
  }
}
```

#### 在线状态
```json
{
  "type": "presence",
  "payload": {
    "is_online": true
  }
}
```

### 3. 通话相关

#### 发起通话
```json
{
  "type": "call.offer",
  "payload": {
    "to_user": 2,
    "call_type": "video",
    "sdp": "v=0\r\no=- 1234567890 1234567890 IN IP4 127.0.0.1\r\n...",
    "ice_candidates": [...]
  }
}
```

#### 应答通话
```json
{
  "type": "call.answer",
  "payload": {
    "call_id": "call_123456",
    "sdp": "v=0\r\no=- 1234567890 1234567890 IN IP4 127.0.0.1\r\n...",
    "ice_candidates": [...]
  }
}
```

#### ICE 候选交换
```json
{
  "type": "call.ice",
  "payload": {
    "call_id": "call_123456",
    "ice_candidate": {
      "candidate": "candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host",
      "sdpMLineIndex": 0,
      "sdpMid": "0"
    }
  }
}
```

#### 结束通话
```json
{
  "type": "call.end",
  "payload": {
    "call_id": "call_123456",
    "reason": "user_hangup"
  }
}
```

## 错误处理

### 连接错误
- **1000**: 正常关闭
- **1001**: 端点离开
- **1002**: 协议错误
- **1003**: 不支持的数据类型
- **1006**: 异常关闭
- **1007**: 数据格式错误
- **1008**: 策略违规
- **1009**: 消息过大
- **1010**: 扩展协商失败
- **1011**: 意外错误

### 业务错误
```json
{
  "type": "error",
  "payload": {
    "error_code": "AUTHENTICATION_FAILED",
    "message": "认证失败",
    "details": {
      "reason": "令牌已过期"
    }
  }
}
```

## 心跳机制

客户端应定期发送心跳消息以保持连接：

```json
{
  "type": "ping",
  "payload": {
    "timestamp": "2025-10-07T12:00:00Z"
  }
}
```

服务器响应：

```json
{
  "type": "pong",
  "payload": {
    "timestamp": "2025-10-07T12:00:00Z"
  }
}
```

## 重连机制

客户端应实现自动重连机制：

1. 连接断开时，等待 1 秒后重连
2. 重连失败时，等待时间翻倍（最大 30 秒）
3. 重连次数超过 5 次时，停止重连
4. 重连成功后，重新发送认证信息

## 安全考虑

1. **认证**: 所有连接都需要有效的 JWT 令牌
2. **加密**: 生产环境必须使用 WSS (WebSocket Secure)
3. **速率限制**: 防止消息洪水攻击
4. **输入验证**: 所有消息内容都需要验证
5. **权限检查**: 确保用户只能访问有权限的资源
