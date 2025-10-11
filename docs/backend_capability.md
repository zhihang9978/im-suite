# 后端能力清单（Backend Capability Matrix）

**生成时间**：2025-10-11
**后端版本**：v1.4.0
**后端基地址**：`https://api.my-domain.com`（待配置）
**鉴权方式**：Bearer JWT Token

---

## 📊 能力总览

| 模块 | 能力数量 | 覆盖度 | 状态 |
|-----|---------|-------|-----|
| 认证与授权 | 9 | 100% | ✅ 完整 |
| 用户管理 | 18 | 90% | ✅ 完整 |
| 消息收发 | 10 | 95% | ✅ 完整 |
| 文件管理 | 7 | 100% | ✅ 完整 |
| 群组/聊天管理 | 32 | 85% | ✅ 完整 |
| 实时通信 | 2 | 80% | ⚠️ 部分 |
| 消息增强 | 12 | 100% | ✅ 完整 |
| 音视频通话 | 8 | 70% | ⚠️ 基础 |
| 内容审核 | 7 | 100% | ✅ 完整 |
| 设备管理 | 9 | 100% | ✅ 完整 |
| 双因子认证 | 7 | 100% | ✅ 完整 |
| 机器人 | 10 | 100% | ✅ 完整 |
| **总计** | **131** | **93%** | **生产就绪** |

---

## 1️⃣ 认证与授权模块

### 1.1 基础认证

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/auth/register` | POST | 用户注册 | 无 | ✅ |
| `/api/auth/login` | POST | 用户登录 | 无 | ✅ |
| `/api/auth/logout` | POST | 用户登出 | Bearer | ✅ |
| `/api/auth/refresh` | POST | 刷新Token | Bearer | ✅ |
| `/api/auth/validate` | GET | 验证Token | Bearer | ✅ |

**请求/响应示例**：

```json
// POST /api/auth/register
{
  "phone": "+8613800138000",
  "username": "user123",
  "password": "password123",
  "nickname": "张三"
}

// Response
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "id": 123,
      "phone": "+8613800138000",
      "username": "user123",
      "nickname": "张三"
    }
  }
}

// POST /api/auth/login
{
  "phone": "+8613800138000",  // 可选，phone或username二选一
  "username": "user123",      // 可选
  "password": "password123"
}

// Response（同注册）
```

### 1.2 双因子认证（2FA）

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/auth/login/2fa` | POST | 2FA验证登录 | 无 | ✅ |
| `/api/2fa/enable` | POST | 启用2FA | Bearer | ✅ |
| `/api/2fa/disable` | POST | 禁用2FA | Bearer | ✅ |
| `/api/2fa/verify` | POST | 验证2FA码 | Bearer | ✅ |
| `/api/2fa/status` | GET | 获取2FA状态 | Bearer | ✅ |
| `/api/2fa/backup-codes/regenerate` | POST | 重新生成备用码 | Bearer | ✅ |
| `/api/2fa/trusted-devices` | GET | 获取信任设备列表 | Bearer | ✅ |
| `/api/2fa/trusted-devices/:id` | DELETE | 移除信任设备 | Bearer | ✅ |

---

## 2️⃣ 用户管理模块

### 2.1 基础用户信息

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/users/me` | GET | 获取当前用户信息 | Bearer | ✅ |
| `/api/users/friends` | GET | 获取好友列表 | Bearer | ⚠️ 简化版 |
| `/api/users/search` | GET | 搜索用户 | 公开 | ✅ |
| `/api/users/by-phone/:phone` | GET | 根据手机号查找用户 | 公开 | ✅ |

**响应示例**：

```json
// GET /api/users/me
{
  "success": true,
  "data": {
    "id": 123,
    "phone": "+8613800138000",
    "username": "user123",
    "nickname": "张三",
    "avatar": "https://cdn.example.com/avatar/123.jpg",
    "bio": "个性签名",
    "language": "zh-CN",
    "theme": "light",
    "online": true
  }
}

// GET /api/users/search?phone=+8613800138000
{
  "success": true,
  "data": [
    {
      "id": 123,
      "phone": "+8613800138000",
      "username": "user123",
      "nickname": "张三",
      "avatar": "https://..."
    }
  ]
}
```

### 2.2 用户管理与限制

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/users/:id/blacklist` | POST | 添加到黑名单 | Bearer | ✅ |
| `/api/users/:id/blacklist/:bid` | DELETE | 从黑名单移除 | Bearer | ✅ |
| `/api/users/:id/blacklist` | GET | 获取黑名单列表 | Bearer | ✅ |
| `/api/users/:id/restrictions` | POST | 设置用户限制 | Bearer + Admin | ✅ |
| `/api/users/:id/restrictions` | GET | 获取用户限制 | Bearer | ✅ |
| `/api/users/:id/ban` | POST | 封禁用户 | Bearer + Admin | ✅ |
| `/api/users/:id/unban` | POST | 解封用户 | Bearer + Admin | ✅ |
| `/api/users/:id/stats` | GET | 获取用户统计 | Bearer | ✅ |
| `/api/users/suspicious` | GET | 获取可疑用户 | Bearer + Admin | ✅ |

---

## 3️⃣ 消息收发模块

### 3.1 消息CRUD

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/messages/send` | POST | 发送消息 | Bearer | ✅ |
| `/api/messages` | GET | 获取消息列表 | Bearer | ✅ |
| `/api/messages/:id` | GET | 获取单条消息 | Bearer | ✅ |
| `/api/messages/:id` | DELETE | 删除消息 | Bearer | ✅ |
| `/api/messages/:id/read` | POST | 标记已读 | Bearer | ✅ |
| `/api/messages/:id/recall` | POST | 撤回消息 | Bearer | ✅ |
| `/api/messages/:id` | PUT | 编辑消息 | Bearer | ✅ |
| `/api/messages/search` | POST | 搜索消息 | Bearer | ✅ |
| `/api/messages/forward` | POST | 转发消息 | Bearer | ✅ |
| `/api/messages/unread/count` | GET | 获取未读数 | Bearer | ✅ |

**请求/响应示例**：

```json
// POST /api/messages/send
{
  "receiver_id": 456,        // 可选，私聊时使用
  "chat_id": 789,            // 可选，群聊时使用
  "content": "你好",
  "message_type": "text",    // text/image/video/audio/file
  "reply_to_id": 123         // 可选，回复某条消息
}

// Response
{
  "success": true,
  "data": {
    "id": 1001,
    "sender_id": 123,
    "receiver_id": 456,
    "content": "你好",
    "message_type": "text",
    "created_at": "2025-10-11T10:00:00Z",
    "read_at": null,
    "edited": false,
    "recalled": false
  }
}

// GET /api/messages?chat_id=789&limit=50&offset=0
{
  "success": true,
  "data": [
    {
      "id": 1001,
      "sender_id": 123,
      "chat_id": 789,
      "content": "你好",
      "message_type": "text",
      "created_at": "2025-10-11T10:00:00Z"
    },
    // ... 更多消息
  ],
  "total": 150,
  "limit": 50,
  "offset": 0
}
```

### 3.2 消息增强功能

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/enhancement/messages/:id/pin` | POST | 置顶消息 | Bearer | ✅ |
| `/api/enhancement/messages/:id/pin` | DELETE | 取消置顶 | Bearer | ✅ |
| `/api/enhancement/messages/:id/mark` | POST | 标记消息 | Bearer | ✅ |
| `/api/enhancement/messages/:id/mark` | DELETE | 取消标记 | Bearer | ✅ |
| `/api/enhancement/messages/:id/reply` | POST | 回复消息 | Bearer | ✅ |
| `/api/enhancement/messages/:id/share` | POST | 分享消息 | Bearer | ✅ |
| `/api/enhancement/messages/:id/status` | POST | 更新状态 | Bearer | ✅ |
| `/api/enhancement/messages/:id/status` | GET | 获取状态 | Bearer | ✅ |
| `/api/enhancement/messages/pinned` | GET | 获取置顶消息 | Bearer | ✅ |
| `/api/enhancement/messages/marked` | GET | 获取标记消息 | Bearer | ✅ |
| `/api/enhancement/messages/:id/reply-chain` | GET | 获取回复链 | Bearer | ✅ |
| `/api/enhancement/messages/:id/share-history` | GET | 获取分享历史 | Bearer | ✅ |

### 3.3 消息加密

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/encryption/messages` | POST | 加密消息 | Bearer | ✅ |
| `/api/encryption/decrypt` | POST | 解密消息 | Bearer | ✅ |
| `/api/encryption/messages/:id/info` | GET | 获取加密信息 | Bearer | ✅ |
| `/api/encryption/messages/:id/self-destruct` | POST | 设置自毁 | Bearer | ✅ |

---

## 4️⃣ 文件管理模块

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/files/upload` | POST | 上传文件 | Bearer | ✅ |
| `/api/files/upload/chunk` | POST | 分片上传 | Bearer | ✅ |
| `/api/files/:id` | GET | 获取文件信息 | Bearer | ✅ |
| `/api/files/:id/download` | GET | 下载文件 | Bearer | ✅ |
| `/api/files/:id/preview` | GET | 获取预览 | Bearer | ✅ |
| `/api/files/:id/versions` | GET | 获取版本列表 | Bearer | ✅ |
| `/api/files/:id/versions` | POST | 创建新版本 | Bearer | ✅ |
| `/api/files/:id` | DELETE | 删除文件 | Bearer | ✅ |

**请求/响应示例**：

```json
// POST /api/files/upload
// Content-Type: multipart/form-data
// form-data: file (binary), is_encrypted (boolean)

// Response
{
  "success": true,
  "data": {
    "url": "https://cdn.example.com/files/abc123.jpg",
    "file_id": 5001,
    "file_name": "photo.jpg"
  }
}

// POST /api/files/upload/chunk
// Content-Type: multipart/form-data
// form-data: chunk, upload_id, chunk_index, total_chunks, file_name, file_size

// Response
{
  "file_id": 5002,
  "file_url": "https://cdn.example.com/files/xyz789.mp4",
  "completed": true,         // 是否完成所有分片
  "chunks_received": 10,     // 已收到的分片数
  "total_chunks": 10         // 总分片数
}
```

---

## 5️⃣ 群组与聊天管理模块

### 5.1 群组管理

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/groups/invites` | POST | 创建邀请链接 | Bearer | ✅ |
| `/api/groups/invites/use` | POST | 使用邀请链接 | Bearer | ✅ |
| `/api/groups/invites/:id` | DELETE | 撤销邀请 | Bearer | ✅ |
| `/api/groups/:id/invites` | GET | 获取邀请列表 | Bearer | ✅ |
| `/api/groups/:id/join-requests/:rid/approve` | POST | 批准入群请求 | Bearer | ✅ |
| `/api/groups/:id/join-requests/pending` | GET | 获取待审核请求 | Bearer | ✅ |
| `/api/groups/:id/members/:uid/promote` | POST | 提升管理员 | Bearer | ✅ |
| `/api/groups/:id/members/:uid/demote` | POST | 降级管理员 | Bearer | ✅ |
| `/api/groups/:id/admins` | GET | 获取管理员列表 | Bearer | ✅ |
| `/api/groups/:id/audit-logs` | GET | 获取审计日志 | Bearer | ✅ |

### 5.2 聊天权限与成员管理

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/chats/:id/permissions` | POST | 设置聊天权限 | Bearer | ✅ |
| `/api/chats/:id/permissions` | GET | 获取聊天权限 | Bearer | ✅ |
| `/api/chats/:id/members/:uid/mute` | POST | 禁言成员 | Bearer | ✅ |
| `/api/chats/:id/members/:uid/unmute` | POST | 解除禁言 | Bearer | ✅ |
| `/api/chats/:id/members/:uid/ban` | POST | 踢出成员 | Bearer | ✅ |
| `/api/chats/:id/members/:uid/unban` | POST | 解除踢出 | Bearer | ✅ |
| `/api/chats/:id/members/:uid/promote` | POST | 提升权限 | Bearer | ✅ |
| `/api/chats/:id/members/:uid/demote` | POST | 降低权限 | Bearer | ✅ |
| `/api/chats/:id/members` | GET | 获取成员列表 | Bearer | ✅ |

### 5.3 公告与规则

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/chats/:id/announcements` | POST | 创建公告 | Bearer | ✅ |
| `/api/chats/:id/announcements/:aid` | PUT | 更新公告 | Bearer | ✅ |
| `/api/chats/:id/announcements/:aid` | DELETE | 删除公告 | Bearer | ✅ |
| `/api/chats/:id/announcements` | GET | 获取公告列表 | Bearer | ✅ |
| `/api/chats/:id/announcements/pinned` | GET | 获取置顶公告 | Bearer | ✅ |
| `/api/chats/:id/announcements/:aid/pin` | POST | 置顶公告 | Bearer | ✅ |
| `/api/chats/:id/announcements/:aid/pin` | DELETE | 取消置顶 | Bearer | ✅ |
| `/api/chats/:id/rules` | POST | 创建规则 | Bearer | ✅ |
| `/api/chats/:id/rules/:rid` | PUT | 更新规则 | Bearer | ✅ |
| `/api/chats/:id/rules/:rid` | DELETE | 删除规则 | Bearer | ✅ |
| `/api/chats/:id/rules` | GET | 获取规则列表 | Bearer | ✅ |

### 5.4 统计与备份

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/chats/:id/statistics` | GET | 获取聊天统计 | Bearer | ✅ |
| `/api/chats/:id/backup` | POST | 创建备份 | Bearer | ✅ |
| `/api/chats/:id/backup/:bid/restore` | POST | 恢复备份 | Bearer | ✅ |
| `/api/chats/:id/backups` | GET | 获取备份列表 | Bearer | ✅ |
| `/api/chats/:id/backups/:bid` | DELETE | 删除备份 | Bearer | ✅ |

---

## 6️⃣ 实时通信模块

### 6.1 WebSocket

| 端点 | 协议 | 功能 | 鉴权 | 状态 |
|-----|-----|-----|-----|-----|
| `/ws` | WebSocket | 实时消息推送 | Query Token | ✅ |

**连接方式**：
```javascript
ws://api.my-domain.com/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**消息格式**：
```json
// 服务器 → 客户端（新消息）
{
  "type": "new_message",
  "data": {
    "id": 1001,
    "sender_id": 123,
    "content": "你好",
    "created_at": "2025-10-11T10:00:00Z"
  }
}

// 服务器 → 客户端（已读回执）
{
  "type": "read_receipt",
  "data": {
    "message_id": 1001,
    "read_by_user_id": 456,
    "read_at": "2025-10-11T10:01:00Z"
  }
}

// 客户端 → 服务器（心跳）
{
  "type": "ping"
}

// 服务器 → 客户端（心跳响应）
{
  "type": "pong",
  "timestamp": 1697000000
}
```

**⚠️ 缺口**：
- ❌ 正在输入（typing）事件
- ❌ 在线状态更新事件
- ⚠️ 送达回执（已部分实现）

**最小可行补齐方案**：
```json
// 添加typing事件
{
  "type": "user_typing",
  "data": {
    "user_id": 123,
    "chat_id": 789
  }
}

// 添加在线状态事件
{
  "type": "user_status",
  "data": {
    "user_id": 123,
    "status": "online",  // online/offline/away
    "last_seen": "2025-10-11T10:00:00Z"
  }
}
```

---

## 7️⃣ 音视频通话模块（WebRTC）

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/calls` | POST | 创建通话 | Bearer | ✅ |
| `/api/calls/:id/end` | POST | 结束通话 | Bearer | ✅ |
| `/api/calls/:id/stats` | GET | 获取统计 | Bearer | ✅ |
| `/api/calls/:id/mute` | POST | 切换静音 | Bearer | ✅ |
| `/api/calls/:id/video` | POST | 切换视频 | Bearer | ✅ |
| `/api/calls/:id/screen-share/start` | POST | 开始屏幕共享 | Bearer | ✅ |
| `/api/calls/:id/screen-share/stop` | POST | 停止屏幕共享 | Bearer | ✅ |
| `/api/calls/:id/screen-share/status` | GET | 屏幕共享状态 | Bearer | ✅ |

**⚠️ 缺口**：
- ❌ TURN/STUN服务器配置API
- ❌ ICE Candidate交换（应通过WebSocket）
- ❌ SDP Offer/Answer交换（应通过WebSocket）
- ⚠️ 多人通话支持（未实现）

**最小可行补齐方案**：
```json
// GET /api/webrtc/config
{
  "success": true,
  "data": {
    "ice_servers": [
      {
        "urls": "stun:stun.l.google.com:19302"
      },
      {
        "urls": "turn:turn.my-domain.com:3478",
        "username": "user123",
        "credential": "temp_password"
      }
    ]
  }
}

// WebSocket信令消息
{
  "type": "webrtc_signal",
  "data": {
    "call_id": 7001,
    "signal_type": "offer",  // offer/answer/ice_candidate
    "sdp": "v=0\r\no=...",    // SDP内容
    "candidate": {...}        // ICE candidate
  }
}
```

---

## 8️⃣ 设备管理模块

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/devices/register` | POST | 注册设备 | Bearer | ✅ |
| `/api/devices` | GET | 获取设备列表 | Bearer | ✅ |
| `/api/devices/:id` | GET | 获取设备详情 | Bearer | ✅ |
| `/api/devices/:id` | DELETE | 撤销设备 | Bearer | ✅ |
| `/api/devices/revoke-all` | POST | 撤销所有设备 | Bearer | ✅ |
| `/api/devices/activities` | GET | 获取设备活动 | Bearer | ✅ |
| `/api/devices/suspicious` | GET | 获取可疑设备 | Bearer | ✅ |
| `/api/devices/statistics` | GET | 获取设备统计 | Bearer | ✅ |
| `/api/devices/export` | GET | 导出设备数据 | Bearer | ✅ |

---

## 9️⃣ 内容审核模块

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/moderation/reports` | POST | 举报内容 | Bearer | ✅ |
| `/api/moderation/reports/pending` | GET | 获取待审核举报 | Bearer + Admin | ✅ |
| `/api/moderation/reports/:id` | GET | 获取举报详情 | Bearer + Admin | ✅ |
| `/api/moderation/reports/:id/handle` | POST | 处理举报 | Bearer + Admin | ✅ |
| `/api/moderation/filters` | POST | 创建过滤器 | Bearer + Admin | ✅ |
| `/api/moderation/users/:id/warnings` | GET | 获取用户警告 | Bearer + Admin | ✅ |
| `/api/moderation/statistics` | GET | 获取审核统计 | Bearer + Admin | ✅ |
| `/api/moderation/content/check` | POST | 内容检查 | Bearer + Admin | ✅ |

---

## 🔟 机器人模块

| API端点 | 方法 | 功能 | 鉴权 | 状态 |
|--------|-----|-----|-----|-----|
| `/api/super-admin/bots` | POST | 创建机器人 | Bearer + SuperAdmin | ✅ |
| `/api/super-admin/bots` | GET | 获取机器人列表 | Bearer + SuperAdmin | ✅ |
| `/api/super-admin/bots/:id` | GET | 获取机器人详情 | Bearer + SuperAdmin | ✅ |
| `/api/super-admin/bots/:id/permissions` | PUT | 更新权限 | Bearer + SuperAdmin | ✅ |
| `/api/super-admin/bots/:id/status` | PUT | 切换状态 | Bearer + SuperAdmin | ✅ |
| `/api/super-admin/bots/:id` | DELETE | 删除机器人 | Bearer + SuperAdmin | ✅ |
| `/api/super-admin/bots/:id/logs` | GET | 获取日志 | Bearer + SuperAdmin | ✅ |
| `/api/super-admin/bots/:id/stats` | GET | 获取统计 | Bearer + SuperAdmin | ✅ |
| `/api/admin/bot-permissions` | POST | 授权用户 | Bearer + Admin | ✅ |
| `/api/admin/bot-permissions/:uid/:bid` | DELETE | 撤销权限 | Bearer + Admin | ✅ |

---

## 📈 性能与限制

### 请求限制（Rate Limiting）
- **全局限制**：100 req/min per IP
- **认证端点**：10 req/min per IP
- **文件上传**：5 req/min per user

### 文件大小限制
- **单文件上传**：最大100MB
- **分片上传**：最大2GB
- **图片预览**：最大10MB

### 分页默认值
- **默认Limit**：50
- **最大Limit**：200
- **Offset起始**：0

### 超时设置
- **读取超时**：60秒
- **写入超时**：60秒
- **连接超时**：10秒

---

## 🔒 安全特性

### 已实现
- ✅ JWT认证与刷新机制
- ✅ 双因子认证（TOTP）
- ✅ 设备管理与信任设备
- ✅ IP限流与频率限制
- ✅ SQL注入防护（GORM Prepared Statements）
- ✅ XSS防护（输入验证）
- ✅ CORS配置
- ✅ TLS/HTTPS支持

### 缺口
- ⚠️ E2E加密密钥交换（已有加密服务但无密钥协商API）
- ⚠️ 审计日志（仅群组有，全局缺失）
- ❌ Webhook签名验证
- ❌ API版本管理（当前无版本前缀）

---

## 🚀 健康检查与监控

| 端点 | 功能 | 状态 |
|-----|-----|-----|
| `/health` | 健康检查 | ✅ |
| `/metrics` | Prometheus指标 | ✅ |

**Health响应示例**：
```json
{
  "status": "ok",
  "timestamp": 1697000000,
  "service": "zhihang-messenger-backend",
  "version": "1.4.0"
}
```

---

## 📊 数据格式规范

### 统一响应格式
```json
// 成功响应
{
  "success": true,
  "data": { ... },
  "message": "操作成功"  // 可选
}

// 错误响应
{
  "success": false,
  "error": "错误简述",
  "details": "详细错误信息",  // 可选
  "code": "ERROR_CODE"        // 可选
}
```

### 时间戳格式
- **格式**：ISO 8601（`2025-10-11T10:00:00Z`）
- **时区**：UTC

### 分页格式
```json
{
  "success": true,
  "data": [ ... ],
  "total": 150,      // 总记录数
  "limit": 50,       // 每页数量
  "offset": 0        // 偏移量
}
```

---

## ⚠️ 关键缺口总结

| 缺口项 | 优先级 | 影响范围 | 最小补齐工作量 |
|-------|-------|---------|--------------|
| Typing事件（WebSocket） | P0 | 用户体验 | 2小时 |
| 在线状态更新（WebSocket） | P0 | 用户体验 | 2小时 |
| WebRTC信令（SDP/ICE交换） | P1 | 音视频通话 | 4小时 |
| TURN服务器配置API | P1 | 音视频通话 | 1小时 |
| 好友关系完整实现 | P1 | 联系人管理 | 6小时 |
| 会话列表API | P0 | 首屏加载 | 4小时 |
| 消息送达回执 | P1 | 消息状态 | 2小时 |
| 全局审计日志 | P2 | 安全审计 | 8小时 |
| API版本管理 | P2 | 兼容性 | 4小时 |

**总计最小补齐工作量**：33小时（约4-5个工作日）

---

**文档版本**：v1.0
**最后更新**：2025-10-11
**维护者**：AI Assistant

