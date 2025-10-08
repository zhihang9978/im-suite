# 消息功能增强 API

## 概述

消息功能增强 API 提供了消息置顶、标记、回复链、状态追踪和分享等功能，让用户能够更好地管理和组织消息。

## 认证

所有消息功能增强 API 都需要通过 JWT Bearer Token 进行认证。在请求头中添加 `Authorization: Bearer <token>`。

## API 列表

### 1. 置顶消息

- **URL**: `/api/messages/pin`
- **方法**: `POST`
- **描述**: 将消息置顶显示
- **请求体**:
  ```json
  {
    "message_id": 123,
    "user_id": 456,
    "reason": "重要通知"
  }
  ```
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "消息置顶成功"
  }
  ```
- **错误响应**:
  - `400 Bad Request`: 消息不存在或已经置顶
  - `401 Unauthorized`: 未授权

### 2. 取消置顶消息

- **URL**: `/api/messages/{message_id}/unpin`
- **方法**: `POST`
- **描述**: 取消消息置顶
- **URL 参数**:
  - `message_id`: 消息ID
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "取消置顶成功"
  }
  ```
- **错误响应**:
  - `400 Bad Request`: 消息未置顶
  - `401 Unauthorized`: 未授权

### 3. 标记消息

- **URL**: `/api/messages/mark`
- **方法**: `POST`
- **描述**: 标记消息为重要、收藏或归档
- **请求体**:
  ```json
  {
    "message_id": 123,
    "user_id": 456,
    "mark_type": "important"
  }
  ```
- **支持的标记类型**:
  - `important`: 重要
  - `favorite`: 收藏
  - `archive`: 归档
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "消息标记成功"
  }
  ```
- **错误响应**:
  - `400 Bad Request`: 消息不存在、无效标记类型或已经标记
  - `401 Unauthorized`: 未授权

### 4. 取消标记消息

- **URL**: `/api/messages/{message_id}/unmark`
- **方法**: `POST`
- **描述**: 取消消息标记
- **URL 参数**:
  - `message_id`: 消息ID
- **查询参数**:
  - `mark_type`: 标记类型 (required)
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "取消标记成功"
  }
  ```
- **错误响应**:
  - `400 Bad Request`: 消息未标记
  - `401 Unauthorized`: 未授权

### 5. 回复消息

- **URL**: `/api/messages/reply`
- **方法**: `POST`
- **描述**: 回复指定消息，支持回复链
- **请求体**:
  ```json
  {
    "message_id": 0,
    "reply_to_id": 123,
    "user_id": 456,
    "content": "回复内容",
    "message_type": "text"
  }
  ```
- **成功响应 (200 OK)**:
  ```json
  {
    "id": 789,
    "sender_id": 456,
    "content": "回复内容",
    "reply_to_id": 123,
    "created_at": "2023-10-27T10:00:00Z",
    "sender": {
      "id": 456,
      "nickname": "用户昵称"
    },
    "reply_to": {
      "id": 123,
      "content": "被回复的消息内容"
    }
  }
  ```
- **错误响应**:
  - `400 Bad Request`: 被回复的消息不存在
  - `401 Unauthorized`: 未授权

### 6. 分享消息

- **URL**: `/api/messages/share`
- **方法**: `POST`
- **描述**: 分享消息给其他用户或群聊
- **请求体**:
  ```json
  {
    "message_id": 123,
    "user_id": 456,
    "shared_to": 789,
    "shared_to_chat_id": 101,
    "share_type": "forward",
    "share_data": "分享备注"
  }
  ```
- **支持的分享类型**:
  - `copy`: 复制链接
  - `forward`: 转发
  - `link`: 生成链接
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "消息分享成功"
  }
  ```
- **错误响应**:
  - `400 Bad Request`: 消息不存在或无效分享类型
  - `401 Unauthorized`: 未授权

### 7. 更新消息状态

- **URL**: `/api/messages/status`
- **方法**: `POST`
- **描述**: 更新消息的发送、送达、已读状态
- **请求体**:
  ```json
  {
    "message_id": 123,
    "user_id": 456,
    "status": "read",
    "device_id": "device123",
    "ip_address": "192.168.1.1"
  }
  ```
- **支持的状态**:
  - `sent`: 已发送
  - `delivered`: 已送达
  - `read`: 已读
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "消息状态更新成功"
  }
  ```
- **错误响应**:
  - `400 Bad Request`: 消息不存在、无效状态或状态已存在
  - `401 Unauthorized`: 未授权

### 8. 获取消息回复链

- **URL**: `/api/messages/{message_id}/reply-chain`
- **方法**: `GET`
- **描述**: 获取消息的完整回复链
- **URL 参数**:
  - `message_id`: 消息ID
- **成功响应 (200 OK)**:
  ```json
  [
    {
      "id": 123,
      "content": "原始消息",
      "sender": {
        "id": 456,
        "nickname": "用户昵称"
      },
      "created_at": "2023-10-27T10:00:00Z"
    },
    {
      "id": 789,
      "content": "回复消息",
      "sender": {
        "id": 101,
        "nickname": "回复者昵称"
      },
      "created_at": "2023-10-27T10:05:00Z"
    }
  ]
  ```
- **错误响应**:
  - `400 Bad Request`: 消息不是回复消息
  - `401 Unauthorized`: 未授权

### 9. 获取置顶消息列表

- **URL**: `/api/messages/pinned`
- **方法**: `GET`
- **描述**: 获取聊天中的置顶消息
- **查询参数**:
  - `chat_id`: 聊天ID (required)
  - `limit`: 限制数量 (default: 20)
  - `offset`: 偏移量 (default: 0)
- **成功响应 (200 OK)**:
  ```json
  [
    {
      "id": 123,
      "content": "置顶消息内容",
      "is_pinned": true,
      "pin_time": "2023-10-27T10:00:00Z",
      "sender": {
        "id": 456,
        "nickname": "用户昵称"
      }
    }
  ]
  ```
- **错误响应**:
  - `400 Bad Request`: 聊天ID格式错误
  - `401 Unauthorized`: 未授权

### 10. 获取标记消息列表

- **URL**: `/api/messages/marked`
- **方法**: `GET`
- **描述**: 获取用户的标记消息
- **查询参数**:
  - `mark_type`: 标记类型 (required)
  - `limit`: 限制数量 (default: 20)
  - `offset`: 偏移量 (default: 0)
- **成功响应 (200 OK)**:
  ```json
  [
    {
      "id": 123,
      "content": "标记消息内容",
      "is_marked": true,
      "mark_type": "important",
      "mark_time": "2023-10-27T10:00:00Z",
      "sender": {
        "id": 456,
        "nickname": "用户昵称"
      }
    }
  ]
  ```
- **错误响应**:
  - `400 Bad Request`: 标记类型不能为空
  - `401 Unauthorized`: 未授权

### 11. 获取消息状态

- **URL**: `/api/messages/{message_id}/status`
- **方法**: `GET`
- **描述**: 获取消息的状态追踪信息
- **URL 参数**:
  - `message_id`: 消息ID
- **成功响应 (200 OK)**:
  ```json
  [
    {
      "id": 1,
      "message_id": 123,
      "user_id": 456,
      "status": "sent",
      "status_time": "2023-10-27T10:00:00Z",
      "device_id": "device123",
      "ip_address": "192.168.1.1",
      "user": {
        "id": 456,
        "nickname": "用户昵称"
      }
    },
    {
      "id": 2,
      "message_id": 123,
      "user_id": 789,
      "status": "read",
      "status_time": "2023-10-27T10:05:00Z",
      "device_id": "device456",
      "ip_address": "192.168.1.2",
      "user": {
        "id": 789,
        "nickname": "接收者昵称"
      }
    }
  ]
  ```
- **错误响应**:
  - `400 Bad Request`: 消息ID格式错误
  - `401 Unauthorized`: 未授权

### 12. 获取消息分享历史

- **URL**: `/api/messages/{message_id}/share-history`
- **方法**: `GET`
- **描述**: 获取消息的分享历史记录
- **URL 参数**:
  - `message_id`: 消息ID
- **查询参数**:
  - `limit`: 限制数量 (default: 20)
  - `offset`: 偏移量 (default: 0)
- **成功响应 (200 OK)**:
  ```json
  [
    {
      "id": 1,
      "message_id": 123,
      "share_user": {
        "id": 456,
        "nickname": "分享者昵称"
      },
      "shared_to_user": {
        "id": 789,
        "nickname": "接收者昵称"
      },
      "share_type": "forward",
      "share_time": "2023-10-27T10:00:00Z",
      "share_data": "分享备注"
    }
  ]
  ```
- **错误响应**:
  - `400 Bad Request`: 消息ID格式错误
  - `401 Unauthorized`: 未授权

## 使用示例

### JavaScript/TypeScript 示例

```typescript
// 置顶消息
const pinMessage = async (messageId: number) => {
  const response = await fetch('/api/messages/pin', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      message_id: messageId,
      user_id: userId,
      reason: '重要通知'
    })
  });
  return response.json();
};

// 标记消息为重要
const markAsImportant = async (messageId: number) => {
  const response = await fetch('/api/messages/mark', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      message_id: messageId,
      user_id: userId,
      mark_type: 'important'
    })
  });
  return response.json();
};

// 回复消息
const replyToMessage = async (replyToId: number, content: string) => {
  const response = await fetch('/api/messages/reply', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      message_id: 0,
      reply_to_id: replyToId,
      user_id: userId,
      content: content,
      message_type: 'text'
    })
  });
  return response.json();
};

// 获取置顶消息列表
const getPinnedMessages = async (chatId: number) => {
  const response = await fetch(`/api/messages/pinned?chat_id=${chatId}`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return response.json();
};
```

## 注意事项

1. **权限控制**: 用户只能操作自己发送的消息或所在群聊的消息
2. **状态追踪**: 消息状态更新是幂等的，重复更新相同状态不会创建新记录
3. **回复链**: 回复链会自动计算层级深度，支持多级回复
4. **分享限制**: 分享功能需要根据消息的可见性进行权限控制
5. **性能优化**: 大量消息的置顶和标记操作建议使用分页加载
