# 超级管理后台 API 文档

## 概述

超级管理后台 API 提供了全面的用户管理、系统监控、内容审核等功能，专为系统管理员设计。

## 基础 URL

```
/api/super-admin
```

## 认证

所有 API 都需要超级管理员权限，请在请求头中包含有效的管理员 JWT token：

```
Authorization: Bearer <admin-jwt-token>
```

## 系统统计

### 获取系统统计信息

获取系统的实时统计数据。

```http
GET /api/super-admin/stats
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "total_users": 10523,
    "online_users": 1234,
    "total_messages": 1543289,
    "today_messages": 15432,
    "total_groups": 523,
    "active_groups": 342,
    "total_files": 8932,
    "storage_used": 10737418240,
    "bandwidth_used": 5368709120,
    "server_load": 45.6,
    "memory_usage": 67.8,
    "cpu_usage": 34.2,
    "database_size": 2147483648,
    "redis_memory_usage": 536870912,
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

### 获取服务器健康状态

检查服务器各组件的健康状态。

```http
GET /api/super-admin/health
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "database": "healthy",
    "redis": "healthy",
    "status": "running",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

## 在线用户管理

### 获取在线用户列表

获取当前在线的所有用户信息。

```http
GET /api/super-admin/online-users?page=1&page_size=50
```

**查询参数：**
- `page`: 页码（默认：1）
- `page_size`: 每页大小（默认：50，最大：100）

**响应示例：**

```json
{
  "success": true,
  "data": {
    "users": [
      {
        "user_id": 123,
        "username": "user123",
        "nickname": "测试用户",
        "avatar": "avatar.jpg",
        "online_status": "online",
        "ip_address": "192.168.1.100",
        "device": "Chrome/Windows",
        "location": "Beijing, China",
        "login_time": "2024-12-19T08:00:00Z",
        "last_activity": "2024-12-19T10:25:00Z",
        "session_count": 2
      }
    ],
    "total": 1234,
    "page": 1,
    "page_size": 50
  }
}
```

## 用户管理

### 获取用户活动记录

获取用户的详细活动历史。

```http
GET /api/super-admin/users/{user_id}/activity?limit=100
```

**查询参数：**
- `limit`: 返回记录数量（默认：100）

**响应示例：**

```json
{
  "success": true,
  "data": [
    {
      "user_id": 123,
      "username": "user123",
      "activity_type": "send_message",
      "ip_address": "192.168.1.100",
      "device": "Chrome/Windows",
      "location": "Beijing, China",
      "details": "发送消息到群组456",
      "timestamp": "2024-12-19T10:25:00Z"
    }
  ]
}
```

### 获取用户行为分析

获取用户的行为分析和风险评估。

```http
GET /api/super-admin/users/{user_id}/analysis
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "user_id": 123,
    "username": "user123",
    "message_count": 15432,
    "group_count": 23,
    "file_upload_count": 145,
    "online_time": 432000,
    "last_login_time": "2024-12-19T08:00:00Z",
    "login_frequency": 5,
    "average_session_time": 7200,
    "risk_score": 25.5,
    "violation_count": 2,
    "reported_count": 1,
    "is_blacklisted": false,
    "is_suspicious": false
  }
}
```

### 强制用户下线

强制用户退出所有会话。

```http
POST /api/super-admin/users/{user_id}/force-logout
```

**请求体：**

```json
{
  "reason": "异常行为检测"
}
```

**响应示例：**

```json
{
  "success": true,
  "message": "用户已强制下线"
}
```

### 封禁用户

封禁用户账号一定时间。

```http
POST /api/super-admin/users/{user_id}/ban
```

**请求体：**

```json
{
  "duration": 168,
  "reason": "发送违规内容"
}
```

**参数说明：**
- `duration`: 封禁时长（小时）
- `reason`: 封禁原因

**响应示例：**

```json
{
  "success": true,
  "message": "用户已封禁"
}
```

### 解封用户

解除用户的封禁状态。

```http
POST /api/super-admin/users/{user_id}/unban
```

**响应示例：**

```json
{
  "success": true,
  "message": "用户已解封"
}
```

### 删除用户账号

永久删除用户账号及相关数据。

```http
DELETE /api/super-admin/users/{user_id}
```

**请求体：**

```json
{
  "reason": "用户申请注销"
}
```

**响应示例：**

```json
{
  "success": true,
  "message": "用户账号已删除"
}
```

## 内容审核

### 获取内容审核队列

获取待审核的内容列表。

```http
GET /api/super-admin/moderation/queue?status=pending&page=1&page_size=20
```

**查询参数：**
- `status`: 审核状态（pending, reviewed）
- `page`: 页码
- `page_size`: 每页大小

**响应示例：**

```json
{
  "success": true,
  "data": {
    "records": [
      {
        "id": 456,
        "content_type": "message",
        "content_id": 789,
        "user_id": 123,
        "username": "user123",
        "content": "待审核的内容",
        "moderation_status": "pending",
        "violation_type": "spam",
        "severity": "medium",
        "created_at": "2024-12-19T10:00:00Z"
      }
    ],
    "total": 45,
    "page": 1,
    "page_size": 20
  }
}
```

### 审核内容

对内容进行审核并采取行动。

```http
POST /api/super-admin/moderation/{content_id}/moderate
```

**请求体：**

```json
{
  "action": "delete",
  "reason": "违反社区规则"
}
```

**参数说明：**
- `action`: 操作类型（approve, reject, delete, warn, ban）
- `reason`: 操作原因

**响应示例：**

```json
{
  "success": true,
  "message": "内容审核完成"
}
```

## 系统管理

### 获取系统日志

获取系统操作日志。

```http
GET /api/super-admin/system/logs?type=all&page=1&page_size=50
```

**查询参数：**
- `type`: 日志类型（all, error, warning, info, security）
- `page`: 页码
- `page_size`: 每页大小

**响应示例：**

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "level": "info",
        "message": "用户登录",
        "user_id": 123,
        "ip_address": "192.168.1.100",
        "timestamp": "2024-12-19T10:30:00Z",
        "details": {}
      }
    ],
    "total": 5432,
    "page": 1,
    "page_size": 50
  }
}
```

### 广播系统消息

向用户或群组广播系统消息。

```http
POST /api/super-admin/system/broadcast
```

**请求体：**

```json
{
  "message": "系统将在30分钟后进行维护",
  "target_type": "all",
  "target_ids": []
}
```

**参数说明：**
- `message`: 消息内容
- `target_type`: 目标类型（all, users, groups）
- `target_ids`: 目标ID列表（当target_type不为all时）

**响应示例：**

```json
{
  "success": true,
  "message": "系统消息已广播"
}
```

## 权限说明

### 超级管理员权限

超级管理员拥有以下权限：

1. **用户管理**
   - 查看所有用户信息
   - 强制用户下线
   - 封禁/解封用户
   - 删除用户账号

2. **内容管理**
   - 审核所有内容
   - 删除违规内容
   - 管理敏感词库

3. **系统监控**
   - 查看系统统计
   - 查看服务器状态
   - 查看系统日志

4. **消息管理**
   - 广播系统消息
   - 查看消息记录

## 安全建议

1. **严格控制超级管理员账号**
   - 限制管理员数量
   - 定期审计管理员操作
   - 使用强密码和2FA

2. **操作记录**
   - 所有管理操作都会被记录
   - 保留完整的审计日志

3. **权限分级**
   - 根据需要分配不同级别的管理权限
   - 遵循最小权限原则

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
- `403 Forbidden`: 权限不足（非超级管理员）
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

## 使用示例

### 查看可疑用户并封禁

```javascript
// 1. 获取用户行为分析
const analysis = await fetch('/api/super-admin/users/123/analysis');
const { risk_score } = analysis.data;

// 2. 如果风险分数高，封禁用户
if (risk_score > 80) {
  await fetch('/api/super-admin/users/123/ban', {
    method: 'POST',
    body: JSON.stringify({
      duration: 168, // 7天
      reason: '高风险行为检测'
    })
  });
}
```

### 批量处理审核队列

```javascript
// 1. 获取待审核内容
const queue = await fetch('/api/super-admin/moderation/queue?status=pending');

// 2. 批量审核
for (const item of queue.data.records) {
  await fetch(`/api/super-admin/moderation/${item.id}/moderate`, {
    method: 'POST',
    body: JSON.stringify({
      action: 'delete',
      reason: '违规内容'
    })
  });
}
```

## 最佳实践

1. **定期监控**: 每天查看系统统计和在线用户
2. **及时处理**: 尽快处理审核队列中的内容
3. **谨慎操作**: 封禁和删除操作前仔细核实
4. **保留记录**: 所有管理操作都要有明确的原因
5. **数据备份**: 在删除用户前确认数据已备份
