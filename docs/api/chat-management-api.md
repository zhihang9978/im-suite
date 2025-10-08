# 志航密信 - 群组管理 API 文档

## 📋 概述

本文档描述了志航密信群组管理功能的API接口，包括群组权限管理、公告和规则管理、统计分析、备份和恢复等功能。

## 🔐 认证

所有API请求都需要在Header中包含JWT认证令牌：

```
Authorization: Bearer <your-jwt-token>
```

## 🔑 群组权限管理

### 设置群组权限

**POST** `/api/chats/{chat_id}/permissions`

设置群组的成员权限配置。

#### 请求参数

```json
{
  "chat_id": 123,
  "can_send_messages": true,
  "can_send_media": true,
  "can_send_stickers": true,
  "can_send_polls": true,
  "can_change_info": false,
  "can_invite_users": false,
  "can_pin_messages": false,
  "can_delete_messages": false,
  "can_edit_messages": false,
  "can_manage_chat": false,
  "can_manage_voice_chats": false,
  "can_restrict_members": false,
  "can_promote_members": false,
  "can_add_admins": false
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chat_id | uint | 是 | 群组ID |
| can_send_messages | bool | 否 | 是否允许发送消息 |
| can_send_media | bool | 否 | 是否允许发送媒体 |
| can_send_stickers | bool | 否 | 是否允许发送贴纸 |
| can_send_polls | bool | 否 | 是否允许发送投票 |
| can_change_info | bool | 否 | 是否允许修改群组信息 |
| can_invite_users | bool | 否 | 是否允许邀请用户 |
| can_pin_messages | bool | 否 | 是否允许置顶消息 |
| can_delete_messages | bool | 否 | 是否允许删除消息 |
| can_edit_messages | bool | 否 | 是否允许编辑消息 |
| can_manage_chat | bool | 否 | 是否允许管理群组 |
| can_manage_voice_chats | bool | 否 | 是否允许管理语音聊天 |
| can_restrict_members | bool | 否 | 是否允许限制成员 |
| can_promote_members | bool | 否 | 是否允许提升成员 |
| can_add_admins | bool | 否 | 是否允许添加管理员 |

#### 响应

```json
{
  "message": "权限设置成功"
}
```

### 获取群组权限

**GET** `/api/chats/{chat_id}/permissions`

获取群组的权限配置。

#### 响应

```json
{
  "data": {
    "id": 1,
    "chat_id": 123,
    "can_send_messages": true,
    "can_send_media": true,
    "can_send_stickers": true,
    "can_send_polls": true,
    "can_change_info": false,
    "can_invite_users": false,
    "can_pin_messages": false,
    "can_delete_messages": false,
    "can_edit_messages": false,
    "can_manage_chat": false,
    "can_manage_voice_chats": false,
    "can_restrict_members": false,
    "can_promote_members": false,
    "can_add_admins": false,
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### 禁言成员

**POST** `/api/chats/{chat_id}/members/mute`

禁言群组成员。

#### 请求参数

```json
{
  "chat_id": 123,
  "user_id": 456,
  "duration": 60,
  "reason": "发布违规内容"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chat_id | uint | 是 | 群组ID |
| user_id | uint | 是 | 被禁言用户ID |
| duration | int | 是 | 禁言时长（分钟） |
| reason | string | 否 | 禁言原因 |

#### 响应

```json
{
  "message": "成员禁言成功"
}
```

### 解除禁言

**DELETE** `/api/chats/{chat_id}/members/{user_id}/mute`

解除群组成员的禁言。

#### 响应

```json
{
  "message": "解除禁言成功"
}
```

### 踢出成员

**POST** `/api/chats/{chat_id}/members/ban`

踢出群组成员。

#### 请求参数

```json
{
  "chat_id": 123,
  "user_id": 456,
  "reason": "违反群规"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chat_id | uint | 是 | 群组ID |
| user_id | uint | 是 | 被踢出用户ID |
| reason | string | 否 | 踢出原因 |

#### 响应

```json
{
  "message": "成员踢出成功"
}
```

### 解除封禁

**DELETE** `/api/chats/{chat_id}/members/{user_id}/ban`

解除群组成员的封禁。

#### 响应

```json
{
  "message": "解除封禁成功"
}
```

### 提升成员权限

**POST** `/api/chats/{chat_id}/members/promote`

提升群组成员的权限。

#### 请求参数

```json
{
  "chat_id": 123,
  "user_id": 456,
  "role": "admin"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chat_id | uint | 是 | 群组ID |
| user_id | uint | 是 | 用户ID |
| role | string | 是 | 新角色：admin 或 owner |

#### 响应

```json
{
  "message": "成员权限提升成功"
}
```

### 降级成员权限

**POST** `/api/chats/{chat_id}/members/{user_id}/demote`

降级群组成员的权限。

#### 响应

```json
{
  "message": "成员权限降级成功"
}
```

### 获取群组成员列表

**GET** `/api/chats/{chat_id}/members`

获取群组成员列表。

#### 响应

```json
{
  "data": [
    {
      "id": 1,
      "chat_id": 123,
      "user_id": 456,
      "role": "owner",
      "joined_at": "2024-12-19T10:30:00Z",
      "last_seen": "2024-12-19T12:30:00Z",
      "is_active": true,
      "user": {
        "id": 456,
        "username": "user123",
        "nickname": "用户昵称",
        "avatar": "avatar_url"
      }
    }
  ]
}
```

## 📢 公告和规则管理

### 创建群组公告

**POST** `/api/chats/{chat_id}/announcements`

创建群组公告。

#### 请求参数

```json
{
  "chat_id": 123,
  "title": "重要公告",
  "content": "这是群组的重要公告内容",
  "is_pinned": true
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chat_id | uint | 是 | 群组ID |
| title | string | 是 | 公告标题 |
| content | string | 是 | 公告内容 |
| is_pinned | bool | 否 | 是否置顶 |

#### 响应

```json
{
  "data": {
    "id": 1,
    "chat_id": 123,
    "title": "重要公告",
    "content": "这是群组的重要公告内容",
    "author_id": 456,
    "is_pinned": true,
    "is_active": true,
    "created_at": "2024-12-19T10:30:00Z",
    "author": {
      "id": 456,
      "username": "user123",
      "nickname": "用户昵称"
    }
  },
  "message": "公告创建成功"
}
```

### 更新群组公告

**PUT** `/api/chats/announcements/{announcement_id}`

更新群组公告。

#### 请求参数

```json
{
  "announcement_id": 1,
  "title": "更新后的公告标题",
  "content": "更新后的公告内容",
  "is_pinned": false
}
```

#### 响应

```json
{
  "message": "公告更新成功"
}
```

### 删除群组公告

**DELETE** `/api/chats/announcements/{announcement_id}`

删除群组公告。

#### 响应

```json
{
  "message": "公告删除成功"
}
```

### 获取群组公告列表

**GET** `/api/chats/{chat_id}/announcements`

获取群组公告列表。

#### 响应

```json
{
  "data": [
    {
      "id": 1,
      "title": "重要公告",
      "content": "公告内容",
      "is_pinned": true,
      "created_at": "2024-12-19T10:30:00Z",
      "author": {
        "id": 456,
        "username": "user123",
        "nickname": "用户昵称"
      }
    }
  ]
}
```

### 获取置顶公告

**GET** `/api/chats/{chat_id}/announcements/pinned`

获取群组的置顶公告。

#### 响应

```json
{
  "data": {
    "id": 1,
    "title": "置顶公告",
    "content": "这是置顶公告的内容",
    "is_pinned": true,
    "created_at": "2024-12-19T10:30:00Z",
    "author": {
      "id": 456,
      "username": "user123",
      "nickname": "用户昵称"
    }
  }
}
```

### 置顶公告

**POST** `/api/chats/announcements/{announcement_id}/pin`

置顶群组公告。

#### 响应

```json
{
  "message": "公告置顶成功"
}
```

### 取消置顶公告

**DELETE** `/api/chats/announcements/{announcement_id}/pin`

取消置顶群组公告。

#### 响应

```json
{
  "message": "取消置顶成功"
}
```

### 创建群组规则

**POST** `/api/chats/{chat_id}/rules`

创建群组规则。

#### 请求参数

```json
{
  "chat_id": 123,
  "rule_number": 1,
  "title": "第一条规则",
  "content": "禁止发布违法违规内容"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chat_id | uint | 是 | 群组ID |
| rule_number | int | 是 | 规则编号 |
| title | string | 是 | 规则标题 |
| content | string | 是 | 规则内容 |

#### 响应

```json
{
  "data": {
    "id": 1,
    "chat_id": 123,
    "rule_number": 1,
    "title": "第一条规则",
    "content": "禁止发布违法违规内容",
    "author_id": 456,
    "is_active": true,
    "created_at": "2024-12-19T10:30:00Z",
    "author": {
      "id": 456,
      "username": "user123",
      "nickname": "用户昵称"
    }
  },
  "message": "规则创建成功"
}
```

### 更新群组规则

**PUT** `/api/chats/rules/{rule_id}`

更新群组规则。

#### 请求参数

```json
{
  "rule_id": 1,
  "title": "更新后的规则标题",
  "content": "更新后的规则内容",
  "rule_number": 2
}
```

#### 响应

```json
{
  "message": "规则更新成功"
}
```

### 删除群组规则

**DELETE** `/api/chats/rules/{rule_id}`

删除群组规则。

#### 响应

```json
{
  "message": "规则删除成功"
}
```

### 获取群组规则列表

**GET** `/api/chats/{chat_id}/rules`

获取群组规则列表。

#### 响应

```json
{
  "data": [
    {
      "id": 1,
      "rule_number": 1,
      "title": "第一条规则",
      "content": "禁止发布违法违规内容",
      "created_at": "2024-12-19T10:30:00Z",
      "author": {
        "id": 456,
        "username": "user123",
        "nickname": "用户昵称"
      }
    }
  ]
}
```

## 📊 统计分析

### 获取群组统计信息

**GET** `/api/chats/{chat_id}/statistics`

获取群组的统计信息。

#### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| date_from | string | 否 | 开始日期 (YYYY-MM-DD) |
| date_to | string | 否 | 结束日期 (YYYY-MM-DD) |
| group_by | string | 否 | 分组方式：hour, day, week, month |

#### 响应

```json
{
  "data": {
    "chat_id": 123,
    "total_members": 100,
    "active_members": 85,
    "total_messages": 5000,
    "messages_today": 50,
    "messages_this_week": 350,
    "messages_this_month": 1500,
    "total_files": 200,
    "total_images": 800,
    "total_videos": 150,
    "total_audios": 300,
    "total_voice_calls": 25,
    "total_video_calls": 15,
    "average_message_length": 45.5,
    "peak_activity_hour": 14,
    "last_activity_at": "2024-12-19T15:30:00Z",
    "message_trends": [
      {
        "date": "2024-12-19",
        "count": 50,
        "day": 19
      }
    ],
    "member_activity": [
      {
        "user_id": 456,
        "username": "user123",
        "nickname": "用户昵称",
        "message_count": 100,
        "last_active": "2024-12-19 15:30:00",
        "join_date": "2024-12-01"
      }
    ],
    "message_type_distribution": [
      {
        "message_type": "text",
        "count": 3000,
        "percentage": 60.0
      },
      {
        "message_type": "image",
        "count": 800,
        "percentage": 16.0
      }
    ],
    "top_active_members": [
      {
        "user_id": 456,
        "username": "user123",
        "nickname": "用户昵称",
        "message_count": 100,
        "rank": 1
      }
    ]
  }
}
```

## 💾 备份和恢复

### 创建群组备份

**POST** `/api/chats/{chat_id}/backups`

创建群组备份。

#### 请求参数

```json
{
  "chat_id": 123,
  "backup_type": "full",
  "is_encrypted": true,
  "expires_in": 168
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chat_id | uint | 是 | 群组ID |
| backup_type | string | 是 | 备份类型：full, messages, media, settings |
| is_encrypted | bool | 否 | 是否加密备份 |
| expires_in | int | 否 | 过期时间（小时） |

#### 响应

```json
{
  "data": {
    "id": 1,
    "chat_id": 123,
    "backup_type": "full",
    "backup_size": 1024000,
    "created_by": 456,
    "is_encrypted": true,
    "expires_at": "2024-12-26T10:30:00Z",
    "created_at": "2024-12-19T10:30:00Z",
    "creator": {
      "id": 456,
      "username": "user123",
      "nickname": "用户昵称"
    }
  },
  "message": "备份创建成功"
}
```

### 恢复群组备份

**POST** `/api/chats/backups/{backup_id}/restore`

恢复群组备份。

#### 请求参数

```json
{
  "backup_id": 1,
  "chat_id": 123
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| backup_id | uint | 是 | 备份ID |
| chat_id | uint | 是 | 目标群组ID |

#### 响应

```json
{
  "message": "备份恢复成功"
}
```

### 获取备份列表

**GET** `/api/chats/{chat_id}/backups`

获取群组的备份列表。

#### 响应

```json
{
  "data": [
    {
      "id": 1,
      "chat_id": 123,
      "backup_type": "full",
      "backup_size": 1024000,
      "created_by": 456,
      "is_encrypted": true,
      "expires_at": "2024-12-26T10:30:00Z",
      "created_at": "2024-12-19T10:30:00Z",
      "creator": {
        "id": 456,
        "username": "user123",
        "nickname": "用户昵称"
      }
    }
  ]
}
```

### 删除备份

**DELETE** `/api/chats/backups/{backup_id}`

删除群组备份。

#### 响应

```json
{
  "message": "备份删除成功"
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

## 📊 权限说明

### 角色权限

| 角色 | 说明 | 权限 |
|------|------|------|
| owner | 群主 | 所有权限 |
| admin | 管理员 | 部分管理权限 |
| member | 普通成员 | 基础权限 |

### 权限类型

| 权限 | 说明 |
|------|------|
| can_send_messages | 发送消息 |
| can_send_media | 发送媒体文件 |
| can_send_stickers | 发送贴纸 |
| can_send_polls | 发送投票 |
| can_change_info | 修改群组信息 |
| can_invite_users | 邀请用户 |
| can_pin_messages | 置顶消息 |
| can_delete_messages | 删除消息 |
| can_edit_messages | 编辑消息 |
| can_manage_chat | 管理群组 |
| can_manage_voice_chats | 管理语音聊天 |
| can_restrict_members | 限制成员 |
| can_promote_members | 提升成员 |
| can_add_admins | 添加管理员 |

## 🔧 使用示例

### 完整的群组管理流程

```bash
# 1. 设置群组权限
curl -X POST "http://localhost:8080/api/chats/123/permissions" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "can_send_messages": true,
    "can_send_media": true,
    "can_invite_users": false
  }'

# 2. 创建群组公告
curl -X POST "http://localhost:8080/api/chats/123/announcements" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "群组规则",
    "content": "请遵守群组规则，禁止发布违法违规内容",
    "is_pinned": true
  }'

# 3. 创建群组规则
curl -X POST "http://localhost:8080/api/chats/123/rules" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "rule_number": 1,
    "title": "禁止发布违法违规内容",
    "content": "群组内禁止发布违法违规、色情暴力等内容"
  }'

# 4. 禁言违规成员
curl -X POST "http://localhost:8080/api/chats/123/members/mute" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 456,
    "duration": 60,
    "reason": "发布违规内容"
  }'

# 5. 获取群组统计
curl -X GET "http://localhost:8080/api/chats/123/statistics?date_from=2024-12-01&date_to=2024-12-19" \
  -H "Authorization: Bearer your-token"

# 6. 创建群组备份
curl -X POST "http://localhost:8080/api/chats/123/backups" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "backup_type": "full",
    "is_encrypted": true,
    "expires_in": 168
  }'
```

---

**最后更新**: 2024-12-19  
**版本**: v1.1.0  
**维护者**: 志航密信技术团队
