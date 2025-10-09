# 机器人 API 文档（受限版本）

**版本**: v1.5.1  
**功能**: 受限的机器人系统，仅用于创建和管理普通用户

---

## 🔒 权限限制说明

### 机器人只能做什么？

✅ **创建普通用户** - 角色固定为 `user`，不能创建管理员  
✅ **删除自己创建的用户** - 只能删除由本机器人创建的用户  
❌ **不能封禁/解封用户** - 已移除此功能  
❌ **不能创建管理员** - 只能创建普通用户  
❌ **不能删除其他用户** - 只能删除自己创建的用户

---

## 🛡️ 安全机制

### 1. 用户标记系统

每个用户模型包含两个新字段：

```go
type User struct {
    // ...
    CreatedByBotID *uint  // 创建该用户的机器人ID
    BotManageable  bool   // 是否允许机器人管理
}
```

**工作原理**：
- 机器人创建用户时，自动设置 `CreatedByBotID` 为机器人ID
- 自动设置 `BotManageable = true`
- 删除时检查这两个字段，确保只能删除自己创建的用户

### 2. 权限检查流程

```
创建用户：
1. 验证API Key/Secret
2. 检查create_user权限
3. 验证手机号/用户名唯一性
4. 强制role="user"（忽略请求中的role字段）
5. 设置CreatedByBotID和BotManageable
6. 创建用户
7. 记录日志

删除用户：
1. 验证API Key/Secret
2. 检查delete_user权限
3. 查找用户
4. 检查BotManageable = true
5. 检查CreatedByBotID == 当前机器人ID
6. 删除用户
7. 记录日志
```

---

## 📋 可用API端点

### 1. 创建用户

**端点**: `POST /api/bot/users`

**认证**: Bot API Key

**请求体**:
```json
{
  "phone": "13800138000",
  "username": "testuser",
  "password": "Password123!",
  "nickname": "测试用户"
}
```

**注意**：
- ⚠️ 不需要提供 `role` 字段，会自动设置为 `user`
- ⚠️ 即使提供 `role` 字段也会被忽略

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
    "created_by_bot_id": 1,
    "bot_manageable": true,
    "is_active": true,
    "created_at": "2024-12-19T..."
  },
  "message": "用户创建成功（角色：普通用户）"
}
```

---

### 2. 删除用户

**端点**: `DELETE /api/bot/users`

**认证**: Bot API Key

**请求体**:
```json
{
  "user_id": 123,
  "reason": "测试完成，清理账号"
}
```

**限制**：
- ✅ 只能删除 `bot_manageable = true` 的用户
- ✅ 只能删除 `created_by_bot_id` 等于当前机器人ID的用户

**成功响应**:
```json
{
  "success": true,
  "message": "用户已删除"
}
```

**错误响应**:
```json
{
  "error": "该用户不允许被机器人管理"
}
```

或

```json
{
  "error": "只能删除本机器人创建的用户"
}
```

---

## 🚫 已移除的功能

### 封禁用户（已移除）
~~POST /api/bot/users/ban~~

### 解封用户（已移除）
~~POST /api/bot/users/:user_id/unban~~

**原因**：限制机器人权限，防止误操作

---

## 📊 使用示例

### Python 示例（受限版本）

```python
import requests

class RestrictedBotClient:
    def __init__(self, api_key, api_secret, base_url):
        self.api_key = api_key
        self.api_secret = api_secret
        self.base_url = base_url
        self.headers = {
            "X-Bot-Auth": f"Bot {api_key}:{api_secret}",
            "Content-Type": "application/json"
        }
    
    def create_user(self, phone, username, password, nickname=""):
        """创建普通用户（角色自动为user）"""
        url = f"{self.base_url}/api/bot/users"
        data = {
            "phone": phone,
            "username": username,
            "password": password,
            "nickname": nickname
            # 注意：不需要提供role字段
        }
        response = requests.post(url, json=data, headers=self.headers)
        return response.json()
    
    def delete_user(self, user_id, reason):
        """删除用户（仅限自己创建的）"""
        url = f"{self.base_url}/api/bot/users"
        data = {
            "user_id": user_id,
            "reason": reason
        }
        response = requests.delete(url, json=data, headers=self.headers)
        return response.json()
    
    def batch_create_users(self, users_data):
        """批量创建用户"""
        results = []
        for user_data in users_data:
            try:
                result = self.create_user(**user_data)
                if result.get('success'):
                    results.append({
                        'username': user_data['username'],
                        'status': 'success',
                        'user_id': result['data']['id']
                    })
                else:
                    results.append({
                        'username': user_data['username'],
                        'status': 'failed',
                        'error': result.get('error')
                    })
            except Exception as e:
                results.append({
                    'username': user_data['username'],
                    'status': 'error',
                    'error': str(e)
                })
        return results

# 使用示例
bot = RestrictedBotClient(
    api_key="bot_abc123...",
    api_secret="789ghi...",
    base_url="http://localhost:8080"
)

# 批量创建普通用户
users_to_create = [
    {
        "phone": "13800138001",
        "username": "user1",
        "password": "Pass123!",
        "nickname": "用户1"
    },
    {
        "phone": "13800138002",
        "username": "user2",
        "password": "Pass123!",
        "nickname": "用户2"
    }
]

results = bot.batch_create_users(users_to_create)
for result in results:
    print(f"{result['username']}: {result['status']}")

# 删除用户
bot.delete_user(123, "测试完成")
```

### Node.js 示例（受限版本）

```javascript
const axios = require('axios');

class RestrictedBotClient {
  constructor(apiKey, apiSecret, baseUrl) {
    this.apiKey = apiKey;
    this.apiSecret = apiSecret;
    this.baseUrl = baseUrl;
    this.headers = {
      'X-Bot-Auth': `Bot ${apiKey}:${apiSecret}`,
      'Content-Type': 'application/json'
    };
  }

  async createUser(phone, username, password, nickname = '') {
    const response = await axios.post(
      `${this.baseUrl}/api/bot/users`,
      { phone, username, password, nickname },
      { headers: this.headers }
    );
    return response.data;
  }

  async deleteUser(userId, reason) {
    const response = await axios.delete(
      `${this.baseUrl}/api/bot/users`,
      {
        data: { user_id: userId, reason },
        headers: this.headers
      }
    );
    return response.data;
  }

  async batchCreateUsers(usersData) {
    const results = [];
    for (const userData of usersData) {
      try {
        const result = await this.createUser(
          userData.phone,
          userData.username,
          userData.password,
          userData.nickname
        );
        results.push({
          username: userData.username,
          status: result.success ? 'success' : 'failed',
          userId: result.data?.id,
          error: result.error
        });
      } catch (error) {
        results.push({
          username: userData.username,
          status: 'error',
          error: error.message
        });
      }
    }
    return results;
  }
}

// 使用示例
(async () => {
  const bot = new RestrictedBotClient(
    'bot_abc123...',
    '789ghi...',
    'http://localhost:8080'
  );

  // 创建用户
  const result = await bot.createUser(
    '13800138001',
    'testuser',
    'Pass123!',
    '测试用户'
  );
  console.log('创建用户:', result);

  // 删除用户
  const deleteResult = await bot.deleteUser(123, '测试完成');
  console.log('删除用户:', deleteResult);
})();
```

---

## 🔍 常见问题

### Q1: 为什么不能创建管理员？
**A**: 为了安全，机器人只能创建普通用户。管理员账号必须由超级管理员手动创建。

### Q2: 为什么不能删除其他用户？
**A**: 每个机器人只能管理自己创建的用户，防止误删其他机器人或管理员创建的用户。

### Q3: 如何查看机器人创建了哪些用户？
**A**: 查询数据库中 `created_by_bot_id` 等于机器人ID的用户：
```sql
SELECT * FROM users WHERE created_by_bot_id = {bot_id} AND deleted_at IS NULL;
```

### Q4: 如果需要封禁用户怎么办？
**A**: 机器人无法封禁用户。如需封禁，请使用超级管理员账号通过管理API操作。

### Q5: 可以修改用户的bot_manageable标记吗？
**A**: 不能通过机器人API修改。如需修改，请联系超级管理员。

---

## 🎯 应用场景

### 场景1: 批量导入测试账号

```python
# 从CSV导入测试用户
import csv

def import_test_users(bot, csv_file):
    with open(csv_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            result = bot.create_user(
                phone=row['phone'],
                username=row['username'],
                password='Test123!',  # 统一密码
                nickname=row['nickname']
            )
            print(f"导入 {row['username']}: {result.get('success', False)}")

import_test_users(bot, 'test_users.csv')
```

### 场景2: 定期清理测试账号

```python
# 清理30天前创建的测试账号
from datetime import datetime, timedelta
import requests

def cleanup_old_test_users(bot, bot_id, days=30):
    # 从数据库查询需要清理的用户
    query = f"""
    SELECT id, username FROM users 
    WHERE created_by_bot_id = {bot_id} 
    AND created_at < DATE_SUB(NOW(), INTERVAL {days} DAY)
    AND deleted_at IS NULL
    """
    
    # 执行查询（需要数据库访问权限）
    users_to_delete = execute_query(query)
    
    for user in users_to_delete:
        result = bot.delete_user(
            user['id'],
            f"自动清理{days}天前的测试账号"
        )
        print(f"清理 {user['username']}: {result}")

# 定期执行
cleanup_old_test_users(bot, bot_id=1, days=30)
```

---

## 📈 监控建议

### 1. 监控创建的用户数量

```python
def get_bot_user_count(bot_id):
    query = f"""
    SELECT COUNT(*) as count 
    FROM users 
    WHERE created_by_bot_id = {bot_id} 
    AND deleted_at IS NULL
    """
    return execute_query(query)[0]['count']
```

### 2. 监控删除操作

```python
def get_bot_delete_logs(bot_id, limit=100):
    # 通过管理API查询日志
    response = requests.get(
        f"http://localhost:8080/api/super-admin/bots/{bot_id}/logs?limit={limit}",
        headers={"Authorization": f"Bearer {admin_token}"}
    )
    logs = response.json()['data']
    
    # 筛选删除操作
    delete_logs = [log for log in logs if 'delete_user' in log['endpoint']]
    return delete_logs
```

---

## 🔧 管理员操作

### 创建受限机器人

```bash
curl -X POST http://localhost:8080/api/super-admin/bots \
  -H "Authorization: Bearer {super_admin_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试用户管理机器人",
    "description": "仅用于创建和删除测试用户",
    "type": "internal",
    "permissions": ["create_user", "delete_user"]
  }'
```

**注意**：
- 只需要 `create_user` 和 `delete_user` 两个权限
- 不要添加 `ban_user`、`unban_user` 等权限（已废弃）

---

## 📝 更新日志

### v1.5.1 (2024-12-19)
- 🔒 限制只能创建普通用户（role固定为user）
- 🔒 限制只能删除自己创建的用户
- ❌ 移除封禁/解封功能
- ✨ 添加用户标记系统（CreatedByBotID、BotManageable）
- 📝 更新API文档

### v1.5.0 (2024-12-19)
- ✨ 初始版本（完整功能）

---

**最后更新**: 2024-12-19  
**作者**: 志航密信开发团队  
**文档版本**: v1.5.1（受限版本）

