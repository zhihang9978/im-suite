# 🔍 API路径匹配分析

## 后端路由定义

### SuperAdmin路由 (`/api/super-admin`)
```go
superAdmin := api.Group("/super-admin")
superAdmin.Use(middleware.SuperAdmin())
{
    superAdminController.SetupRoutes(superAdmin)
    // SetupRoutes中定义的路由:
    // GET /stats                    → /api/super-admin/stats
    // GET /stats/system            → /api/super-admin/stats/system  
    // GET /users                   → /api/super-admin/users
    // GET /users/online            → /api/super-admin/users/online
    // POST /users/:id/logout       → /api/super-admin/users/:id/logout
    // POST /users/:id/ban          → /api/super-admin/users/:id/ban
    // POST /users/:id/unban        → /api/super-admin/users/:id/unban
    // DELETE /users/:id            → /api/super-admin/users/:id
    // GET /users/:id/analysis      → /api/super-admin/users/:id/analysis
    // GET /alerts                  → /api/super-admin/alerts
    // GET /logs                    → /api/super-admin/logs
    // POST /broadcast              → /api/super-admin/broadcast
}
```

## 前端API调用

### SuperAdmin.vue
```javascript
request.get('/super-admin/stats')            // ✅ 匹配
request.get('/super-admin/online-users')     // ❌ 错误！应该是 /super-admin/users/online
request.get('/super-admin/moderation/queue') // ❌ 后端未定义
request.get('/super-admin/system/logs')      // ❌ 错误！应该是 /super-admin/logs
request.post('/super-admin/users/:id/force-logout') // ❌ 错误！应该是 /super-admin/users/:id/logout
```

## 🚨 发现的路径错误

### 错误 #1: online-users
- **前端**: `/super-admin/online-users`
- **后端**: `/super-admin/users/online`
- **修复**: 前端改为 `/super-admin/users/online`

### 错误 #2: force-logout  
- **前端**: `/super-admin/users/:id/force-logout`
- **后端**: `/super-admin/users/:id/logout`
- **修复**: 前端改为 `/super-admin/users/:id/logout`

### 错误 #3: system/logs
- **前端**: `/super-admin/system/logs`
- **后端**: `/super-admin/logs`
- **修复**: 前端改为 `/super-admin/logs`

### 错误 #4: moderation/queue
- **前端**: `/super-admin/moderation/queue`
- **后端**: 未定义此路由
- **修复**: 删除前端调用或在后端添加路由

### 错误 #5: system/broadcast
- **前端**: `/super-admin/system/broadcast`
- **后端**: `/super-admin/broadcast`
- **修复**: 前端改为 `/super-admin/broadcast`

### 错误 #6: moderation/:id/moderate
- **前端**: `/super-admin/moderation/:id/moderate`
- **后端**: 未定义此路由
- **修复**: 删除前端调用或在后端添加路由

---

## 需要修复的文件

1. `im-admin/src/views/SuperAdmin.vue` - 修复所有API路径

---

## 修复后的正确路径

| 功能 | 正确路径 |
|------|---------|
| 系统统计 | `/super-admin/stats` |
| 在线用户 | `/super-admin/users/online` |
| 强制下线 | `/super-admin/users/:id/logout` |
| 封禁用户 | `/super-admin/users/:id/ban` |
| 系统日志 | `/super-admin/logs` |
| 广播消息 | `/super-admin/broadcast` |
| 用户分析 | `/super-admin/users/:id/analysis` |

