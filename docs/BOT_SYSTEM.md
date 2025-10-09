# 🤖 机器人系统文档

**版本**: v1.5.0  
**状态**: ✅ 已完成  
**更新日期**: 2024-12-19

---

## 🎯 系统概述

### 设计背景

为了满足自动化用户管理的需求，我们设计了一套完整的机器人（Bot）系统。机器人是具有特殊权限的自动化账户，可以通过API执行管理操作，无需人工介入。

### 核心功能

✅ **创建用户** - 批量注册、导入用户  
✅ **封禁用户** - 自动风控、违规处理  
✅ **解封用户** - 批量恢复权限  
✅ **删除用户** - 批量清理、数据整理  

---

## 🏗️ 系统架构

### 技术栈

- **后端**: Go + Gin
- **数据库**: MySQL
- **缓存**: Redis（速率限制）
- **认证**: API Key + API Secret
- **加密**: bcrypt

### 核心组件

```
┌─────────────────────────────────────────────┐
│           机器人系统架构                    │
├─────────────────────────────────────────────┤
│                                             │
│  1. 数据模型层 (Model)                      │
│     ├── Bot                                 │
│     │   ├── 基本信息（名称、描述、类型）   │
│     │   ├── 认证信息（API Key/Secret）     │
│     │   ├── 权限配置（JSON数组）           │
│     │   └── 限制配置（速率、每日）         │
│     └── BotAPILog                           │
│         ├── 调用记录                        │
│         ├── 响应时间                        │
│         └── 错误日志                        │
│                                             │
│  2. 服务层 (Service)                        │
│     ├── BotService                          │
│     │   ├── CreateBot                       │
│     │   ├── ValidateBotAPIKey              │
│     │   ├── CheckBotPermission             │
│     │   ├── CheckRateLimit                 │
│     │   ├── BotCreateUser                  │
│     │   ├── BotBanUser                     │
│     │   ├── BotUnbanUser                   │
│     │   └── BotDeleteUser                  │
│     └── 审计日志                            │
│                                             │
│  3. 控制器层 (Controller)                   │
│     └── BotController                       │
│         ├── 管理API（需要super_admin）      │
│         │   ├── POST /super-admin/bots     │
│         │   ├── GET  /super-admin/bots     │
│         │   ├── PUT  /super-admin/bots/:id │
│         │   └── DELETE /super-admin/bots/:id│
│         └── 机器人API（使用Bot认证）        │
│             ├── POST /bot/users            │
│             ├── POST /bot/users/ban        │
│             ├── POST /bot/users/:id/unban  │
│             └── DELETE /bot/users          │
│                                             │
│  4. 中间件层 (Middleware)                   │
│     ├── BotAuthMiddleware                  │
│     │   ├── 验证API Key/Secret             │
│     │   ├── 检查速率限制                   │
│     │   └── 记录API调用                    │
│     └── BotPermissionMiddleware            │
│         └── 检查权限                        │
│                                             │
└─────────────────────────────────────────────┘
```

---

## 📁 文件结构

```
im-backend/
├── internal/
│   ├── model/
│   │   └── bot.go                      # 机器人数据模型
│   ├── service/
│   │   └── bot_service.go              # 机器人业务逻辑
│   ├── controller/
│   │   └── bot_controller.go           # 机器人API控制器
│   └── middleware/
│       └── bot_auth.go                 # 机器人认证中间件
├── config/
│   └── database.go                     # 数据库配置（包含Bot表迁移）
└── main.go                             # 路由配置

docs/
└── api/
    └── bot-api.md                      # API文档
```

---

## 🗂️ 数据模型

### Bot 表结构

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 机器人名称（唯一） |
| description | string | 描述 |
| type | string | 类型：internal/webhook/plugin |
| api_key | string | API密钥（唯一） |
| api_secret | string | API密钥（bcrypt加密） |
| permissions | text | 权限列表（JSON数组） |
| is_active | bool | 是否激活 |
| last_used_at | timestamp | 最后使用时间 |
| rate_limit | int | 速率限制（次/分钟） |
| daily_limit | int | 每日限制 |
| created_by | uint | 创建者用户ID |
| total_calls | int64 | 总调用次数 |
| success_calls | int64 | 成功次数 |
| failed_calls | int64 | 失败次数 |

### BotAPILog 表结构

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| bot_id | uint | 机器人ID |
| endpoint | string | 调用的端点 |
| method | string | HTTP方法 |
| status_code | int | 响应状态码 |
| ip_address | string | 调用IP |
| user_agent | string | User-Agent |
| duration | int64 | 耗时（毫秒） |
| request_body | text | 请求体 |
| response_body | text | 响应体 |
| error | string | 错误信息 |

---

## 🔐 权限系统

### 可用权限列表

#### 用户管理
- ✅ `create_user` - 创建用户
- ✅ `delete_user` - 删除用户
- ✅ `ban_user` - 封禁用户
- ✅ `unban_user` - 解封用户
- ✅ `update_user` - 更新用户信息
- ✅ `list_users` - 查看用户列表

#### 消息管理
- 📋 `send_message` - 发送消息
- 📋 `delete_message` - 删除消息
- 📋 `broadcast` - 广播消息

#### 群组管理
- 📋 `create_group` - 创建群组
- 📋 `delete_group` - 删除群组
- 📋 `manage_group` - 管理群组

#### 内容审核
- 📋 `moderate_content` - 内容审核

#### 系统管理
- 📋 `view_stats` - 查看统计
- 📋 `view_logs` - 查看日志

> **当前版本（v1.5.0）已实现**: 用户管理的4个权限  
> **未来版本**: 将逐步实现消息、群组、审核等权限

---

## 🔑 认证流程

### 1. 创建机器人（super_admin）

```bash
POST /api/super-admin/bots
Authorization: Bearer {super_admin_token}

{
  "name": "用户管理机器人",
  "description": "用于自动化用户管理",
  "type": "internal",
  "permissions": ["create_user", "ban_user", "delete_user"]
}

# 响应（只显示一次！）
{
  "api_key": "bot_abc123def456...",
  "api_secret": "789ghi012jkl..."
}
```

### 2. 使用机器人API

```bash
POST /api/bot/users
X-Bot-Auth: Bot bot_abc123def456...:789ghi012jkl...

{
  "phone": "13800138000",
  "username": "testuser",
  "password": "Password123!"
}
```

### 3. 验证流程

```
1. 提取认证头中的 API Key 和 Secret
2. 从数据库查找对应的机器人
3. 验证 API Secret（bcrypt比对）
4. 检查机器人状态（is_active）
5. 检查速率限制（Redis）
6. 检查权限
7. 执行操作
8. 记录日志
```

---

## 🛡️ 安全机制

### 1. API密钥安全

- ✅ **API Secret加密存储** - 使用bcrypt加密
- ✅ **API Key唯一性** - 数据库唯一索引
- ✅ **只显示一次** - 创建时返回，之后无法查看
- ✅ **支持重新生成** - 旧密钥立即失效

### 2. 速率限制

```go
// 分钟级限制
Key: bot:ratelimit:minute:{bot_id}
默认: 100次/分钟
TTL: 1分钟

// 每日限制
Key: bot:ratelimit:day:{bot_id}
默认: 10,000次/天
TTL: 24小时
```

### 3. 权限控制

- ✅ 每个操作都需要相应权限
- ✅ 不能创建super_admin用户
- ✅ 不能封禁/删除super_admin
- ✅ 所有操作都有审计日志

### 4. 操作审计

```go
type BotAPILog struct {
    BotID      uint      // 哪个机器人
    Endpoint   string    // 调用什么接口
    Method     string    // HTTP方法
    StatusCode int       // 结果
    Duration   int64     // 耗时
    IPAddress  string    // 来源IP
    Error      string    // 错误信息
    CreatedAt  time.Time // 时间
}
```

---

## 📊 使用示例

### Python SDK

```python
import requests

class BotClient:
    def __init__(self, api_key, api_secret, base_url):
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = base_url
        self.headers = {
            "X-Bot-Auth": f"Bot {api_key}:{api_secret}",
            "Content-Type": "application/json"
        }
    
    def create_user(self, phone, username, password, role="user"):
        url = f"{self.base_url}/api/bot/users"
        data = {
            "phone": phone,
            "username": username,
            "password": password,
            "role": role
        }
        response = requests.post(url, json=data, headers=self.headers)
        return response.json()
    
    def ban_user(self, user_id, duration, reason):
        url = f"{self.base_url}/api/bot/users/ban"
        data = {
            "user_id": user_id,
            "duration": duration,
            "reason": reason
        }
        response = requests.post(url, json=data, headers=self.headers)
        return response.json()
    
    def delete_user(self, user_id, reason):
        url = f"{self.base_url}/api/bot/users"
        data = {
            "user_id": user_id,
            "reason": reason
        }
        response = requests.delete(url, json=data, headers=self.headers)
        return response.json()

# 使用示例
bot = BotClient(
    api_key="bot_abc123...",
    api_secret="789ghi...",
    base_url="http://localhost:8080"
)

# 批量创建用户
users_to_create = [
    {"phone": "13800138001", "username": "user1", "password": "Pass123!"},
    {"phone": "13800138002", "username": "user2", "password": "Pass123!"},
    {"phone": "13800138003", "username": "user3", "password": "Pass123!"},
]

for user in users_to_create:
    result = bot.create_user(**user)
    print(f"创建用户 {user['username']}: {result}")
```

---

## 📈 监控和维护

### 1. 查看机器人统计

```bash
GET /api/super-admin/bots/{id}/stats

# 响应
{
  "total_calls": 1250,
  "success_calls": 1200,
  "failed_calls": 50,
  "success_rate": 96.0,
  "today_calls": 150,
  "last_used_at": "2024-01-01T12:00:00Z"
}
```

### 2. 查看调用日志

```bash
GET /api/super-admin/bots/{id}/logs?limit=100

# 响应
[
  {
    "id": 1,
    "endpoint": "/api/bot/users",
    "method": "POST",
    "status_code": 201,
    "duration": 125,
    "created_at": "2024-01-01T12:00:00Z"
  }
]
```

### 3. 管理机器人状态

```bash
# 停用机器人
PUT /api/super-admin/bots/{id}/status
{"is_active": false}

# 重新生成密钥
POST /api/super-admin/bots/{id}/regenerate-secret
```

---

## 🎯 应用场景

### 场景1: 批量导入用户

```python
# 从CSV导入用户
import csv

def import_users_from_csv(bot, csv_file):
    with open(csv_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            result = bot.create_user(
                phone=row['phone'],
                username=row['username'],
                password=row['password']
            )
            if result['success']:
                print(f"✅ {row['username']} 导入成功")
            else:
                print(f"❌ {row['username']} 导入失败: {result['error']}")

import_users_from_csv(bot, 'users.csv')
```

### 场景2: 自动风控

```python
# 检测并封禁违规用户
def auto_ban_violators(bot, violation_list):
    for violation in violation_list:
        user_id = violation['user_id']
        reason = violation['reason']
        duration = 86400  # 24小时
        
        result = bot.ban_user(user_id, duration, reason)
        if result['success']:
            print(f"✅ 用户 {user_id} 已封禁: {reason}")
            # 发送通知
            notify_user(user_id, f"您因{reason}被封禁24小时")
```

### 场景3: 定期清理

```python
# 定期清理测试账号
def cleanup_test_accounts(bot):
    test_users = get_test_users()  # 获取测试账号列表
    
    for user_id in test_users:
        result = bot.delete_user(user_id, "定期清理测试账号")
        if result['success']:
            print(f"✅ 测试账号 {user_id} 已清理")

# 每周执行一次
schedule.every().week.do(cleanup_test_accounts, bot)
```

---

## 🔧 配置参数

### 机器人默认配置

```go
// 创建机器人时的默认值
Bot {
    IsActive:    true,
    RateLimit:   100,    // 100次/分钟
    DailyLimit:  10000,  // 10,000次/天
    TotalCalls:  0,
    SuccessCalls: 0,
    FailedCalls: 0,
}
```

### 环境变量

```env
# Redis配置（用于速率限制）
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=zhihang_messenger
```

---

## 🚀 部署建议

### 1. 生产环境

- ✅ 使用环境变量存储API密钥
- ✅ 启用HTTPS
- ✅ 配置防火墙规则
- ✅ 定期检查日志
- ✅ 监控API调用量

### 2. 高可用

- ✅ Redis主从复制
- ✅ 数据库读写分离
- ✅ 负载均衡
- ✅ 定期备份

### 3. 安全加固

- ✅ IP白名单
- ✅ 请求签名验证
- ✅ 异常行为检测
- ✅ 自动告警

---

## 📝 版本历史

### v1.5.0 (2024-12-19)
- ✨ 新增机器人系统
- ✨ 支持创建、封禁、删除用户
- ✨ API Key认证
- ✨ 速率限制
- ✨ 操作审计
- ✨ 完整的管理API

---

## 🔜 未来规划

### v1.6.0 - 消息管理权限
- 📋 机器人发送消息
- 📋 机器人删除消息
- 📋 系统广播

### v1.7.0 - 群组管理权限
- 📋 机器人创建群组
- 📋 机器人管理群组
- 📋 批量群组操作

### v1.8.0 - 内容审核权限
- 📋 自动内容审核
- 📋 违规内容处理
- 📋 审核报告

---

## 📚 相关文档

- [机器人API文档](./api/bot-api.md)
- [超级管理员功能清单](./SUPER_ADMIN_FEATURES.md)
- [安全最佳实践](./security/transport-security.md)

---

**最后更新**: 2024-12-19  
**维护者**: 志航密信开发团队  
**文档版本**: v1.5.0

