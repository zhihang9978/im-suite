# 性能优化 API 文档

## 概述

性能优化 API 提供了消息推送优化、大群组性能优化、存储优化和网络优化等功能。

## 基础 URL

```
/api/performance
```

## 认证

所有 API 都需要认证，请在请求头中包含有效的 JWT token：

```
Authorization: Bearer <your-jwt-token>
```

## 消息推送优化

### 获取推送统计

获取消息推送服务的统计信息。

```http
GET /api/performance/push/stats
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "queue_length": 15,
    "recent_batches": 245,
    "worker_count": 5,
    "batch_size": 100
  }
}
```

### 队列推送任务

将消息推送到队列中异步处理。

```http
POST /api/performance/push/queue
```

**请求体：**

```json
{
  "message_id": 12345,
  "chat_id": 67890,
  "sender_id": 11111,
  "content": "Hello World",
  "type": "text",
  "priority": 2
}
```

**参数说明：**
- `message_id`: 消息ID
- `chat_id`: 聊天ID
- `sender_id`: 发送者ID
- `content`: 消息内容
- `type`: 消息类型
- `priority`: 优先级 (1-高优先级, 2-普通, 3-低优先级)

**响应示例：**

```json
{
  "success": true,
  "message": "推送任务已加入队列"
}
```

## 大群组性能优化

### 获取聊天信息（优化版本）

获取聊天信息，使用缓存优化。

```http
GET /api/performance/groups/{id}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "id": 67890,
    "name": "技术讨论群",
    "type": "group",
    "last_message_id": 12345,
    "last_message_at": "2024-01-15T10:30:00Z"
  }
}
```

### 分页获取群成员

分页获取群成员列表，支持缓存。

```http
GET /api/performance/groups/{id}/members?page=1&page_size=50
```

**查询参数：**
- `page`: 页码（默认：1）
- `page_size`: 每页大小（默认：50，最大：100）

**响应示例：**

```json
{
  "success": true,
  "data": {
    "members": [
      {
        "user_id": 11111,
        "username": "user1",
        "nickname": "用户1",
        "avatar": "avatar1.jpg",
        "role": "admin",
        "joined_at": "2024-01-01T00:00:00Z",
        "is_active": true,
        "last_seen": "2024-01-15T10:00:00Z"
      }
    ],
    "total": 150,
    "page": 1,
    "page_size": 50
  }
}
```

### 分页获取消息

分页获取消息列表，支持缓存。

```http
GET /api/performance/groups/{id}/messages?page=1&page_size=50
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": 12345,
        "chat_id": 67890,
        "user_id": 11111,
        "content": "Hello World",
        "type": "text",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 5000,
    "page": 1,
    "page_size": 50
  }
}
```

### 获取聊天统计信息

获取聊天的统计信息。

```http
GET /api/performance/groups/{id}/statistics
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "member_count": 150,
    "message_count": 5000,
    "today_message_count": 50,
    "active_member_count": 120
  }
}
```

### 使缓存失效

清理指定聊天的缓存。

```http
DELETE /api/performance/groups/{id}/cache
```

**响应示例：**

```json
{
  "success": true,
  "message": "缓存已清理"
}
```

### 清理不活跃成员

清理指定天数内不活跃的成员。

```http
POST /api/performance/groups/{id}/cleanup-members?days=30
```

**查询参数：**
- `days`: 不活跃天数（默认：30）

**响应示例：**

```json
{
  "success": true,
  "message": "不活跃成员清理完成"
}
```

## 存储优化

### 获取存储统计

获取数据库存储统计信息。

```http
GET /api/performance/storage/stats
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "total_size_mb": 1024.5,
    "tables": [
      {
        "table_name": "messages",
        "size_mb": 512.3,
        "rows": 100000
      }
    ],
    "recent_compressions": [
      {
        "table_name": "messages",
        "original_size": 600,
        "compressed_size": 512.3,
        "compression_rate": 14.6,
        "record_count": 100000,
        "compressed_at": "2024-01-15T10:00:00Z"
      }
    ]
  }
}
```

### 压缩表数据

压缩指定表的数据。

```http
POST /api/performance/storage/compress/{table}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "table_name": "messages",
    "original_size": 600,
    "compressed_size": 512.3,
    "compression_rate": 14.6,
    "record_count": 100000,
    "compressed_at": "2024-01-15T10:00:00Z"
  }
}
```

### 创建分区

为指定表创建分区。

```http
POST /api/performance/storage/partitions/{table}
```

**响应示例：**

```json
{
  "success": true,
  "message": "分区创建完成"
}
```

### 获取分区信息

获取指定表的分区信息。

```http
GET /api/performance/storage/partitions/{table}
```

**响应示例：**

```json
{
  "success": true,
  "data": [
    {
      "table_name": "messages",
      "partition_name": "p202401",
      "rows": 50000,
      "size": 256.2,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 清理旧消息

调度清理指定天数前的旧消息。

```http
POST /api/performance/storage/cleanup/messages?days=90
```

**查询参数：**
- `days`: 保留天数（默认：90）

**响应示例：**

```json
{
  "success": true,
  "message": "旧消息清理任务已调度"
}
```

### 清理不活跃会话

调度清理指定天数前的不活跃会话。

```http
POST /api/performance/storage/cleanup/sessions?days=30
```

**响应示例：**

```json
{
  "success": true,
  "message": "不活跃会话清理任务已调度"
}
```

### 清理孤立文件

调度清理孤立文件。

```http
POST /api/performance/storage/cleanup/files
```

**响应示例：**

```json
{
  "success": true,
  "message": "孤立文件清理任务已调度"
}
```

## 网络优化

### 获取网络统计

获取网络优化统计信息。

```http
GET /api/performance/network/stats
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "total_requests": 10000,
    "compressed_requests": 8500,
    "cache_hits": 6000,
    "cache_misses": 4000,
    "average_latency": 150.5,
    "compression_ratio": 0.85,
    "bandwidth_saved": 1048576,
    "last_updated": "2024-01-15T10:30:00Z",
    "by_endpoint": {
      "/api/messages": 5000,
      "/api/users": 3000,
      "/api/groups": 2000
    }
  }
}
```

### 获取优化建议

获取基于统计数据的优化建议。

```http
GET /api/performance/network/recommendations
```

**响应示例：**

```json
{
  "success": true,
  "data": [
    "建议启用更多内容的压缩以减少带宽使用",
    "缓存命中率较低，建议优化缓存策略",
    "平均延迟较高，建议优化网络连接和服务器性能"
  ]
}
```

## 数据库优化

### 优化数据库

执行数据库优化操作（创建索引等）。

```http
POST /api/performance/database/optimize
```

**响应示例：**

```json
{
  "success": true,
  "message": "数据库优化完成"
}
```

## 错误响应

所有 API 在出错时都会返回以下格式的错误响应：

```json
{
  "error": "错误描述"
}
```

**常见错误码：**
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未认证或认证失败
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

## 性能监控

建议定期调用以下 API 来监控系统性能：

1. `/api/performance/push/stats` - 监控推送队列状态
2. `/api/performance/storage/stats` - 监控存储使用情况
3. `/api/performance/network/stats` - 监控网络性能
4. `/api/performance/network/recommendations` - 获取优化建议

## 最佳实践

1. **定期清理**：建议定期执行存储清理操作
2. **监控缓存**：关注缓存命中率，必要时清理缓存
3. **数据库优化**：定期执行数据库优化操作
4. **网络压缩**：确保网络压缩功能正常工作
5. **分区管理**：对大数据表使用分区提高查询性能
