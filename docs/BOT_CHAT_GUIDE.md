# 📱 聊天机器人使用指南 v1.6.0

**功能**: 通过聊天界面与机器人交互，创建和删除用户

---

## 🎯 功能概述

授权用户可以通过**发送消息**给机器人来管理用户，无需调用API。机器人会自动：
1. 解析命令
2. 执行操作
3. 返回结果

---

## 🚀 快速开始

### 步骤1: 超级管理员创建机器人用户

首先需要为机器人在系统中创建一个用户账号：

```bash
POST /api/super-admin/bot-users
Authorization: Bearer {super_admin_token}

{
  "bot_id": 1,
  "username": "userbot",
  "nickname": "用户管理机器人",
  "avatar": "https://example.com/bot-avatar.png"
}
```

**响应**：
```json
{
  "success": true,
  "data": {
    "id": 1,
    "bot_id": 1,
    "user_id": 100,
    "is_active": true
  },
  "message": "机器人用户创建成功，现在可以在聊天中与机器人交互"
}
```

### 步骤2: 授权用户使用机器人

管理员授权用户可以使用机器人：

```bash
POST /api/admin/bot-permissions
Authorization: Bearer {admin_token}

{
  "user_id": 10,
  "bot_id": 1,
  "expires_at": null
}
```

**响应**：
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 10,
    "bot_id": 1,
    "is_active": true
  },
  "message": "用户授权成功，现在可以使用机器人"
}
```

### 步骤3: 用户与机器人聊天

用户在聊天界面中：
1. 搜索机器人用户名：`userbot`
2. 开始对话
3. 发送命令

---

## 💬 可用命令

### 1. /help - 帮助

查看所有可用命令

**使用**：
```
/help
```

**机器人回复**：
```
🤖 **机器人命令帮助**

**用户管理命令：**

📝 **/create** - 创建新用户
格式: /create phone=手机号 username=用户名 password=密码 [nickname=昵称]
示例: /create phone=13800138000 username=testuser password=Pass123! nickname=测试用户

🗑️ **/delete** - 删除用户
格式: /delete user_id=用户ID reason=删除原因
示例: /delete user_id=123 reason=测试完成

📋 **/list** - 查看已创建的用户
格式: /list [limit=数量]
示例: /list limit=10

ℹ️ **/info** - 查看用户信息
格式: /info user_id=用户ID
示例: /info user_id=123

❓ **/help** - 显示此帮助信息
```

---

### 2. /create - 创建用户

创建新的普通用户

**格式**：
```
/create phone=手机号 username=用户名 password=密码 [nickname=昵称]
```

**示例**：
```
/create phone=13800138000 username=testuser password=Pass123! nickname=测试用户
```

**机器人回复**：
```
✅ **用户创建成功！**

👤 **用户信息：**
- ID: 123
- 用户名: testuser
- 手机号: 13800138000
- 昵称: 测试用户
- 角色: user
- 状态: 激活
- 创建时间: 2024-12-19 10:30:00

⚠️ 请妥善保管用户的登录凭证
```

**错误示例**：
```
/create phone=138001380 username=test password=123
```

**机器人回复**：
```
❌ 手机号格式错误，应为11位数字，以1开头
```

---

### 3. /delete - 删除用户

删除机器人创建的用户

**格式**：
```
/delete user_id=用户ID reason=删除原因
```

**示例**：
```
/delete user_id=123 reason=测试完成
```

**机器人回复**：
```
✅ **用户删除成功！**

👤 **已删除用户：**
- ID: 123
- 用户名: testuser
- 手机号: 13800138000
- 删除原因: 测试完成
- 删除时间: 2024-12-19 10:35:00
```

**限制**：
- ⚠️ 只能删除**本机器人创建**的用户
- ⚠️ 不能删除其他方式创建的用户

---

### 4. /list - 查看用户列表

查看机器人创建的所有用户

**格式**：
```
/list [limit=数量]
```

**示例**：
```
/list limit=5
```

**机器人回复**：
```
📋 **用户列表** (共 3 个)

1. **testuser1** (ID:101)
   - 手机: 13800138001
   - 昵称: 测试用户1
   - 状态: ✅ 激活
   - 创建: 2024-12-19 10:00

2. **testuser2** (ID:102)
   - 手机: 13800138002
   - 昵称: 测试用户2
   - 状态: ✅ 激活
   - 创建: 2024-12-19 10:10

3. **testuser3** (ID:103)
   - 手机: 13800138003
   - 昵称: 测试用户3
   - 状态: ✅ 激活
   - 创建: 2024-12-19 10:20

💡 使用 /info user_id=ID 查看详细信息
```

---

### 5. /info - 查看用户详情

查看指定用户的详细信息

**格式**：
```
/info user_id=用户ID
```

**示例**：
```
/info user_id=123
```

**机器人回复**：
```
ℹ️ **用户详细信息**

**基本信息：**
- ID: 123
- 用户名: testuser
- 手机号: 13800138000
- 昵称: 测试用户
- 角色: user

**状态信息：**
- 账户状态: ✅ 激活
- 在线状态: 🟢 在线
- 最后在线: 2024-12-19 10:30:00

**管理信息：**
- 由本机器人创建: ✅ 是
- 可被机器人管理: ✅ 是
- 注册时间: 2024-12-19 09:00:00

**操作提示：**
💡 您可以删除此用户: /delete user_id=123 reason=删除原因
```

---

## 📱 完整使用流程示例

### 场景1: 批量创建测试用户

```
用户: /create phone=13800138001 username=test1 password=Test123! nickname=测试1
机器人: ✅ 用户创建成功！...

用户: /create phone=13800138002 username=test2 password=Test123! nickname=测试2
机器人: ✅ 用户创建成功！...

用户: /create phone=13800138003 username=test3 password=Test123! nickname=测试3
机器人: ✅ 用户创建成功！...

用户: /list
机器人: 📋 用户列表 (共 3 个)...
```

### 场景2: 查看和删除用户

```
用户: /list limit=5
机器人: 📋 用户列表 (共 3 个)...

用户: /info user_id=101
机器人: ℹ️ 用户详细信息...

用户: /delete user_id=101 reason=测试完成
机器人: ✅ 用户删除成功！...

用户: /list
机器人: 📋 用户列表 (共 2 个)...
```

### 场景3: 权限限制示例

```
# 未授权用户尝试使用机器人
未授权用户: /help
机器人: ❌ 您没有权限使用机器人功能。请联系管理员授权。

# 尝试删除其他用户
用户: /delete user_id=999 reason=测试
机器人: ❌ 删除用户失败: 只能删除本机器人创建的用户
```

---

## 🔒 权限和限制

### 1. 使用权限
- ✅ 必须由管理员**授权**才能使用
- ✅ 可设置**过期时间**
- ❌ 未授权用户无法使用任何命令

### 2. 操作限制
- ✅ 只能创建**普通用户**（role=user）
- ✅ 只能删除**本机器人创建**的用户
- ❌ 不能创建管理员
- ❌ 不能删除其他用户
- ❌ 不能封禁用户

### 3. 参数验证
- ✅ 手机号必须是11位，以1开头
- ✅ 密码至少6位
- ✅ 用户名和手机号必须唯一

---

## 🔧 管理员功能

### 查看机器人的授权用户

```bash
GET /api/super-admin/bot-users/{bot_id}/permissions
Authorization: Bearer {admin_token}
```

### 撤销用户权限

```bash
DELETE /api/admin/bot-permissions/{user_id}/{bot_id}
Authorization: Bearer {admin_token}
```

### 用户查看自己的权限

```bash
GET /api/bot-permissions
Authorization: Bearer {user_token}
```

**响应**：
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "user_id": 10,
      "bot_id": 1,
      "bot": {
        "id": 1,
        "name": "用户管理机器人",
        "description": "..."
      },
      "granted_by": 1,
      "granted_by_user": {
        "username": "admin"
      },
      "is_active": true,
      "expires_at": null,
      "created_at": "2024-12-19T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

## 📊 数据库设计

### BotUser表 - 机器人用户关联

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| bot_id | uint | 机器人ID |
| user_id | uint | 系统用户ID |
| is_active | bool | 是否激活 |

### BotUserPermission表 - 用户使用权限

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 被授权用户ID |
| bot_id | uint | 机器人ID |
| granted_by | uint | 授权者ID |
| is_active | bool | 是否激活 |
| expires_at | timestamp | 过期时间 |

---

## 🛡️ 安全特性

### 1. 权限检查
- 每次命令执行前检查用户权限
- 过期权限自动失效
- 可随时撤销权限

### 2. 操作审计
- 所有操作记录到Bot日志
- 包含操作者、时间、结果
- 便于追踪和审查

### 3. 异步处理
- 消息处理异步执行
- 不影响正常聊天性能
- 自动错误恢复

---

## ❓ 常见问题

### Q1: 如何授权多个用户？
**A**: 管理员可以批量授权：
```bash
# 授权用户1
POST /api/admin/bot-permissions
{"user_id": 10, "bot_id": 1}

# 授权用户2
POST /api/admin/bot-permissions
{"user_id": 11, "bot_id": 1}
```

### Q2: 机器人不回复怎么办？
**A**: 检查：
1. 机器人用户是否创建成功
2. 用户是否已授权
3. 消息是否发送到正确的机器人账号

### Q3: 如何查看机器人创建了多少用户？
**A**: 发送 `/list` 命令查看完整列表

### Q4: 可以设置临时权限吗？
**A**: 可以，授权时设置过期时间：
```json
{
  "user_id": 10,
  "bot_id": 1,
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### Q5: 命令参数可以有空格吗？
**A**: 当前版本不支持参数值包含空格。如需设置昵称为"测试 用户"，请使用下划线：`nickname=测试_用户`

---

## 🚀 未来扩展

### 计划中的功能

- 📋 支持群聊中使用机器人（@mention）
- 📋 支持批量操作命令
- 📋 支持参数值包含空格
- 📋 支持更多管理命令
- 📋 支持自定义命令
- 📋 支持命令历史查询

---

## 📝 版本历史

### v1.6.0 (2024-12-19)
- ✨ 新增聊天机器人功能
- ✨ 支持通过聊天界面创建/删除用户
- ✨ 支持命令解析和执行
- ✨ 支持用户权限管理
- ✨ 支持操作审计

---

**最后更新**: 2024-12-19  
**作者**: 志航密信开发团队  
**文档版本**: v1.6.0

