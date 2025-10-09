# 权限系统配置文档

## 📋 权限级别说明

### 三级权限体系

| 角色 | 权限级别 | 说明 | 可访问功能 |
|------|---------|------|----------|
| **user** | 普通用户 | 默认角色 | 基础通讯功能 |
| **admin** | 管理员 | 中级权限 | 用户管理、内容审核 |
| **super_admin** | 超级管理员 | 最高权限 | 全部功能 + 系统管理 |

---

## 🔐 当前权限配置

### 1. 公开路由（无需登录）
```go
/api/auth/login           - 用户登录
/api/auth/register        - 用户注册
/api/auth/login/2fa       - 2FA验证登录
/api/auth/2fa/validate    - 2FA验证
/health                   - 健康检查
```

**权限**: ✅ 无需认证

---

### 2. 受保护路由（需要登录）
```go
/api/messages/*           - 消息管理
/api/users/*              - 用户个人信息
/api/files/*              - 文件管理
/api/2fa/*                - 2FA个人设置
/api/devices/*            - 设备个人管理
/api/themes/*             - 主题设置
/api/moderation/reports   - 内容举报
```

**权限**: ✅ 需要JWT认证（任何已登录用户）  
**中间件**: `AuthMiddleware()`

**说明**: 这些是用户个人功能，每个用户都可以管理自己的设置。

---

### 3. 超级管理员路由（需要super_admin）
```go
/api/super-admin/stats                    - 系统统计
/api/super-admin/users/online             - 在线用户
/api/super-admin/users/:id/logout         - 强制下线
/api/super-admin/users/:id/ban            - 封禁用户
/api/super-admin/users/:id/unban          - 解封用户
/api/super-admin/users/:id                - 删除用户
/api/super-admin/users/:id/analysis       - 用户分析
/api/super-admin/alerts                   - 系统告警
/api/super-admin/logs                     - 管理日志
/api/super-admin/broadcast                - 系统广播
```

**权限**: ✅ 需要super_admin角色  
**中间件**: `AuthMiddleware()` + `SuperAdmin()`

**说明**: 这些是系统管理功能，只有超级管理员可以访问。

---

## 🎯 管理后台权限要求

### im-admin管理后台访问

#### 推荐配置：Admin或Super_Admin ⭐

**方式1：前端路由守卫（推荐）**

编辑 `im-admin/src/router/index.js`:

```javascript
// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && userStore.isLoggedIn) {
    // 检查用户角色
    if (userStore.user.role === 'user') {
      // 普通用户不允许访问管理后台
      alert('需要管理员权限才能访问')
      next(false)
      return
    }
    next('/')
  } else {
    // 检查管理员权限
    if (to.meta.requiresAdmin && userStore.user.role === 'user') {
      alert('需要管理员权限')
      next('/')
      return
    }
    next()
  }
})
```

**方式2：创建Admin中间件**

创建 `im-backend/internal/middleware/admin.go`:

```go
package middleware

import (
	"net/http"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
	"github.com/gin-gonic/gin"
)

// Admin 管理员权限中间件（admin或super_admin）
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 查询用户角色
		var user model.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查是否为管理员或超级管理员
		if user.Role != "admin" && user.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

---

## ✅ v1.4.0 权限配置状态

### 2FA设置 - 用户级别 ✅ 正确

**配置**: `protected.Group("/2fa")`  
**权限**: AuthMiddleware（登录用户）  
**原因**: 每个用户都应该能管理自己的2FA

**这是正确的！** ✅

---

### 设备管理 - 用户级别 ✅ 正确

**配置**: `protected.Group("/devices")`  
**权限**: AuthMiddleware（登录用户）  
**原因**: 每个用户都应该能管理自己的设备

**这是正确的！** ✅

---

### 超级管理员功能 - Super_Admin级别 ✅ 正确

**配置**: `superAdmin.Group("/super-admin")`  
**权限**: AuthMiddleware + SuperAdmin  
**原因**: 系统管理功能，只有超级管理员可访问

**这是正确的！** ✅

---

## 🎯 建议的权限优化（可选）

### 1. 管理后台访问控制（推荐）

**场景**: 管理后台（im-admin）应该只允许admin和super_admin访问

**实现方式**：

#### 方式A：前端控制（简单）
在 `im-admin/src/router/index.js` 添加角色检查

#### 方式B：后端控制（安全）
创建Admin中间件，保护管理相关API

#### 方式C：综合方案（推荐）
前端+后端双重验证

---

### 2. 敏感功能的管理员查看（可选）

**场景**: 管理员可以查看所有用户的设备和2FA状态

**新增功能**：
```go
// 管理员专用路由
adminRoutes := api.Group("/admin")
adminRoutes.Use(middleware.AuthMiddleware())
adminRoutes.Use(middleware.Admin()) // 需要admin或super_admin

{
    // 查看所有用户的2FA状态
    adminRoutes.GET("/users/:id/2fa/status", ...)
    
    // 查看所有用户的设备
    adminRoutes.GET("/users/:id/devices", ...)
    
    // 强制重置用户2FA
    adminRoutes.POST("/users/:id/2fa/reset", ...)
}
```

---

## 📊 当前权限矩阵

| 功能 | user | admin | super_admin |
|------|------|-------|-------------|
| 登录/注册 | ✅ | ✅ | ✅ |
| 发送消息 | ✅ | ✅ | ✅ |
| 管理自己的2FA | ✅ | ✅ | ✅ |
| 管理自己的设备 | ✅ | ✅ | ✅ |
| 查看自己的信息 | ✅ | ✅ | ✅ |
| **访问管理后台** | ❓ | ✅ 建议 | ✅ |
| **内容审核** | ❌ | ✅ | ✅ |
| **用户管理** | ❌ | ✅ | ✅ |
| **系统统计** | ❌ | ❌ | ✅ |
| **强制下线** | ❌ | ❌ | ✅ |
| **封禁用户** | ❌ | ❌ | ✅ |
| **删除用户** | ❌ | ❌ | ✅ |
| **系统广播** | ❌ | ❌ | ✅ |

---

## ✅ 结论

### 当前配置：正确且合理 ✅

**现有权限配置是正确的**：
- ✅ 用户个人功能（2FA、设备）- 任何用户可访问
- ✅ 系统管理功能 - 仅super_admin可访问
- ✅ 权限中间件配置正确

### 建议优化（可选）

**如果需要限制管理后台访问**：
1. 前端添加角色检查（10分钟）
2. 或创建Admin中间件（15分钟）

**如果需要管理员查看所有用户状态**：
1. 新增管理员专用API（30分钟）
2. 添加Admin中间件保护（已准备好代码）

---

## 🤔 需要您确认

**问题**: 您说"后台拥有最高级别权限"，是指：

**A. 管理后台（im-admin）应该只允许admin和super_admin访问？**
- 需要添加前端角色检查
- 或创建Admin中间件

**B. 超级管理员应该能查看所有用户的2FA和设备状态？**
- 需要新增管理员查看API
- 添加Admin中间件

**C. 当前配置已经正确？**
- 用户管理自己的2FA和设备 ✅
- 超级管理员有系统管理权限 ✅
- 无需修改

请告诉我您的需求，我立即实现！🚀

