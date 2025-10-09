# 机器人 API 文档

**版本**: v1.5.0  
**功能**: 内部机器人系统，用于自动化用户管理

---

## 🤖 机器人概述

### 什么是机器人？

机器人（Bot）是具有特殊权限的自动化账户，可以通过API执行管理操作：
- **创建用户** - 批量注册、导入用户
- **封禁用户** - 自动风控、违规处理
- **删除用户** - 批量清理、数据整理

### 机器人类型

- **internal** - 内部管理机器人
- **webhook** - WebHook机器人
- **plugin** - 插件机器人

### 认证方式

机器人使用 **API Key + API Secret** 认证，不同于普通用户的JWT Token。

---

## 🔑 认证

### 认证格式

**请求头**:
```http
X-Bot-Auth: Bot {api_key}:{api_secret}
```

**示例**:
```bash
curl -X POST http://localhost:8080/api/bot/users \
  -H "X-Bot-Auth: Bot bot_abc123...:def456..." \
  -H "Content-Type: application/json" \
  -d '{...}'
```

---

## 📋 机器人管理 API（需要 super_admin 权限）

### 1. 创建机器人

**端点**: `POST /api/super-admin/bots`

**权限**: super_admin

**请求体**:
```json
{
  "name": "用户管理机器人",
  "description": "用于自动化用户管理的内部机器人",
  "type": "internal",
  "permissions": [
    "create_user",
    "ban_user",
    "unban_user",
    "delete_user"
  ]
}
```

**可用权限**:
- `create_user` - 创建用户
- `delete_user` - 删除用户
- `ban_user` - 封禁用户
- `unban_user` - 解封用户
- `update_user` - 更新用户
- `list_users` - 查看用户列表
- `send_message` - 发送消息
- `delete_message` - 删除消息
- `broadcast` - 广播消息
- `create_group` - 创建群组
- `delete_group` - 删除群组
- `manage_group` - 管理群组
- `moderate_content` - 内容审核
- `view_stats` - 查看统计
- `view_logs` - 查看日志

**响应**:
```json
{
  "success": true,
  "data": {
    "bot": {
      "id": 1,
      "name": "用户管理机器人",
      "description": "用于自动化用户管理的内部机器人",
      "type": "internal",
      "api_key": "bot_abc123def456...",
      "permissions": "[\"create_user\",\"ban_user\",\"delete_user\"]",
      "is_active": true,
      "rate_limit": 100,
      "daily_limit": 10000,
      "created_at": "2024-01-01T00:00:00Z"
    },
    "api_key": "bot_abc123def456...",
    "api_secret": "789ghi012jkl..."
  },
  "message": "机器人创建成功，请妥善保管API密钥",
  "warning": "API密钥只显示一次，请立即保存"
}
```

⚠️ **重要**: API Secret 只在创建时返回一次，务必立即保存！

---

### 2. 获取机器人列表

**端点**: `GET /api/super-admin/bots`

**权限**: super_admin

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "用户管理机器人",
      "type": "internal",
      "is_active": true,
      "total_calls": 1250,
      "success_calls": 1200,
      "failed_calls": 50,
      "last_used_at": "2024-01-01T12:00:00Z",
      "creator": {
        "id": 1,
        "username": "admin"
      }
    }
  ],
  "total": 1
}
```

---

### 3. 获取机器人详情

**端点**: `GET /api/super-admin/bots/:id`

**权限**: super_admin

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "用户管理机器人",
    "description": "用于自动化用户管理的内部机器人",
    "type": "internal",
    "api_key": "bot_abc123...",
    "permissions": "[\"create_user\",\"ban_user\",\"delete_user\"]",
    "is_active": true,
    "rate_limit": 100,
    "daily_limit": 10000,
    "total_calls": 1250,
    "success_calls": 1200,
    "failed_calls": 50,
    "last_used_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 4. 更新机器人权限

**端点**: `PUT /api/super-admin/bots/:id/permissions`

**权限**: super_admin

**请求体**:
```json
{
  "permissions": [
    "create_user",
    "ban_user",
    "unban_user",
    "delete_user",
    "list_users"
  ]
}
```

**响应**:
```json
{
  "success": true,
  "message": "权限已更新"
}
```

---

### 5. 切换机器人状态

**端点**: `PUT /api/super-admin/bots/:id/status`

**权限**: super_admin

**请求体**:
```json
{
  "is_active": false
}
```

**响应**:
```json
{
  "success": true,
  "message": "状态已更新"
}
```

---

### 6. 删除机器人

**端点**: `DELETE /api/super-admin/bots/:id`

**权限**: super_admin

**响应**:
```json
{
  "success": true,
  "message": "机器人已删除"
}
```

---

### 7. 获取机器人调用日志

**端点**: `GET /api/super-admin/bots/:id/logs?limit=100`

**权限**: super_admin

**参数**:
- `limit` - 限制返回数量（默认100）

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "bot_id": 1,
      "endpoint": "/api/bot/users",
      "method": "POST",
      "status_code": 201,
      "ip_address": "192.168.1.1",
      "duration": 125,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

---

### 8. 获取机器人统计

**端点**: `GET /api/super-admin/bots/:id/stats`

**权限**: super_admin

**响应**:
```json
{
  "success": true,
  "data": {
    "total_calls": 1250,
    "success_calls": 1200,
    "failed_calls": 50,
    "success_rate": 96.0,
    "today_calls": 150,
    "last_used_at": "2024-01-01T12:00:00Z",
    "is_active": true
  }
}
```

---

### 9. 重新生成API密钥

**端点**: `POST /api/super-admin/bots/:id/regenerate-secret`

**权限**: super_admin

**响应**:
```json
{
  "success": true,
  "api_secret": "new_secret_789ghi012jkl...",
  "message": "API密钥已重新生成，请立即保存",
  "warning": "旧密钥将立即失效"
}
```

⚠️ **重要**: 旧密钥立即失效！请更新所有使用该机器人的应用。

---

## 🤖 机器人操作 API（使用 Bot 认证）

### 1. 创建用户

**端点**: `POST /api/bot/users`

**认证**: Bot API Key

**权限**: `create_user`

**请求体**:
```json
{
  "phone": "13800138000",
  "username": "testuser",
  "password": "Password123!",
  "nickname": "测试用户",
  "role": "user"
}
```

**字段说明**:
- `phone` (必填) - 手机号
- `username` (必填) - 用户名
- `password` (必填) - 密码
- `nickname` (可选) - 昵称
- `role` (可选) - 角色，可选: `user`, `admin`（不能创建`super_admin`）

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 123,
    "phone": "13800138000",
    "username": "testuser",
    "nickname": "测试用户",
    "role": "user",
    "is_active": true,
    "created_at": "2024-01-01T12:00:00Z"
  },
  "message": "用户创建成功"
}
```

**错误响应**:
```json
{
  "error": "手机号已存在"
}
```

---

### 2. 封禁用户

**端点**: `POST /api/bot/users/ban`

**认证**: Bot API Key

**权限**: `ban_user`

**请求体**:
```json
{
  "user_id": 123,
  "duration": 86400,
  "reason": "发送违规内容"
}
```

**字段说明**:
- `user_id` (必填) - 用户ID
- `duration` (必填) - 封禁时长（秒），0表示永久封禁
- `reason` (必填) - 封禁原因

**响应**:
```json
{
  "success": true,
  "message": "用户已封禁"
}
```

**错误响应**:
```json
{
  "error": "不能封禁超级管理员"
}
```

---

### 3. 解封用户

**端点**: `POST /api/bot/users/:user_id/unban`

**认证**: Bot API Key

**权限**: `unban_user`

**响应**:
```json
{
  "success": true,
  "message": "用户已解封"
}
```

---

### 4. 删除用户

**端点**: `DELETE /api/bot/users`

**认证**: Bot API Key

**权限**: `delete_user`

**请求体**:
```json
{
  "user_id": 123,
  "reason": "批量清理测试账号"
}
```

**字段说明**:
- `user_id` (必填) - 用户ID
- `reason` (必填) - 删除原因

**响应**:
```json
{
  "success": true,
  "message": "用户已删除"
}
```

**错误响应**:
```json
{
  "error": "不能删除超级管理员"
}
```

---

## 🛡️ 安全特性

### 速率限制

机器人有两级速率限制：

1. **分钟级限制**: 默认 100 次/分钟
2. **每日限制**: 默认 10,000 次/天

超过限制时返回 `429 Too Many Requests`：
```json
{
  "error": "速率限制：超过100次/分钟"
}
```

### 权限控制

每个机器人只能执行被授予的权限：

```json
{
  "error": "权限不足",
  "required_permission": "create_user",
  "message": "机器人没有执行此操作的权限"
}
```

### 操作审计

所有机器人操作都被记录：
- 操作类型
- 操作时间
- 操作对象
- 操作结果
- IP地址
- 响应时间

---

## 📊 使用示例

### Python 示例

```python
import requests

# 机器人配置
API_KEY = "bot_abc123def456..."
API_SECRET = "789ghi012jkl..."
BASE_URL = "http://localhost:8080"

# 认证头
headers = {
    "X-Bot-Auth": f"Bot {API_KEY}:{API_SECRET}",
    "Content-Type": "application/json"
}

# 创建用户
def create_user(phone, username, password):
    url = f"{BASE_URL}/api/bot/users"
    data = {
        "phone": phone,
        "username": username,
        "password": password,
        "role": "user"
    }
    
    response = requests.post(url, json=data, headers=headers)
    return response.json()

# 封禁用户
def ban_user(user_id, duration, reason):
    url = f"{BASE_URL}/api/bot/users/ban"
    data = {
        "user_id": user_id,
        "duration": duration,
        "reason": reason
    }
    
    response = requests.post(url, json=data, headers=headers)
    return response.json()

# 删除用户
def delete_user(user_id, reason):
    url = f"{BASE_URL}/api/bot/users"
    data = {
        "user_id": user_id,
        "reason": reason
    }
    
    response = requests.delete(url, json=data, headers=headers)
    return response.json()

# 使用示例
if __name__ == "__main__":
    # 创建用户
    result = create_user("13800138000", "testuser", "Password123!")
    print("创建用户:", result)
    
    # 封禁用户24小时
    result = ban_user(123, 86400, "违规发送广告")
    print("封禁用户:", result)
    
    # 删除用户
    result = delete_user(123, "测试完成")
    print("删除用户:", result)
```

### Node.js 示例

```javascript
const axios = require('axios');

// 机器人配置
const API_KEY = 'bot_abc123def456...';
const API_SECRET = '789ghi012jkl...';
const BASE_URL = 'http://localhost:8080';

// 认证头
const headers = {
  'X-Bot-Auth': `Bot ${API_KEY}:${API_SECRET}`,
  'Content-Type': 'application/json'
};

// 创建用户
async function createUser(phone, username, password) {
  const response = await axios.post(
    `${BASE_URL}/api/bot/users`,
    { phone, username, password, role: 'user' },
    { headers }
  );
  return response.data;
}

// 封禁用户
async function banUser(userId, duration, reason) {
  const response = await axios.post(
    `${BASE_URL}/api/bot/users/ban`,
    { user_id: userId, duration, reason },
    { headers }
  );
  return response.data;
}

// 使用示例
(async () => {
  try {
    // 创建用户
    const result = await createUser('13800138000', 'testuser', 'Password123!');
    console.log('创建用户:', result);
    
    // 封禁用户
    const banResult = await banUser(123, 86400, '违规发送广告');
    console.log('封禁用户:', banResult);
  } catch (error) {
    console.error('错误:', error.response?.data || error.message);
  }
})();
```

---

## 🔧 最佳实践

### 1. 密钥管理
- ✅ 将API密钥存储在环境变量中
- ✅ 不要将密钥提交到版本控制
- ✅ 定期轮换API密钥
- ✅ 为不同环境使用不同的机器人

### 2. 错误处理
- ✅ 检查响应状态码
- ✅ 处理速率限制错误（429）
- ✅ 实现重试机制
- ✅ 记录错误日志

### 3. 批量操作
- ✅ 遵守速率限制
- ✅ 使用适当的延迟
- ✅ 处理部分失败
- ✅ 记录操作结果

### 4. 监控
- ✅ 定期检查机器人状态
- ✅ 监控调用次数
- ✅ 查看错误率
- ✅ 审查操作日志

---

## ❌ 错误码

| 状态码 | 说明 | 示例 |
|--------|------|------|
| 400 | 请求参数错误 | 缺少必填字段 |
| 401 | 认证失败 | 无效的API密钥 |
| 403 | 权限不足 | 机器人没有该权限 |
| 404 | 资源不存在 | 用户不存在 |
| 429 | 速率限制 | 超过调用次数 |
| 500 | 服务器错误 | 内部错误 |

---

## 📝 更新日志

### v1.5.0 (2024-12-19)
- ✨ 新增机器人系统
- ✨ 支持创建、封禁、删除用户
- ✨ API Key认证
- ✨ 速率限制
- ✨ 操作审计

---

**最后更新**: 2024-12-19  
**作者**: 志航密信开发团队  
**文档版本**: v1.5.0

