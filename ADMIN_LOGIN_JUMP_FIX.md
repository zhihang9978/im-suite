# 管理后台登录跳转问题修复

**问题**: 登录成功但不跳转到仪表盘  
**原因**: 后端返回字段名与前端期望不匹配  
**修复日期**: 2025-10-10

---

## 🔍 问题分析

### 现象
- ✅ 登录API返回成功（HTTP 200）
- ✅ 后端返回了 token 和用户信息
- ❌ 前端显示"登录成功"但停留在登录页
- ❌ 不跳转到仪表盘

### 根本原因

**后端返回的数据结构**:
```json
{
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    ...
  },
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "expires_in": 86400,
  "requires_2fa": false
}
```

**前端期望的数据结构**:
```javascript
// im-admin/src/stores/user.js
const response = await login(credentials)
token.value = response.token  // ❌ 期望 token，但后端返回 access_token
user.value = response.user    // ✅ 这个是对的
```

**结果**:
- `token.value = undefined` (因为 response.token 不存在)
- `localStorage` 中存储了 `undefined`
- `isLoggedIn` 计算为 `false`
- 路由守卫检测到未登录，不允许跳转

---

## 🔧 修复方案

### 修复文件: `im-admin/src/stores/user.js`

#### 修改 1: loginUser 方法

**修复前**:
```javascript
const loginUser = async (credentials) => {
  try {
    const response = await login(credentials)
    token.value = response.token  // ❌ undefined
    user.value = response.user
    localStorage.setItem('admin_token', response.token)  // ❌ 存储 undefined
    return response
  } catch (error) {
    throw error
  }
}
```

**修复后**:
```javascript
const loginUser = async (credentials) => {
  try {
    const response = await login(credentials)
    // 后端返回的是 access_token 和 refresh_token，不是 token
    const accessToken = response.access_token || response.token
    token.value = accessToken
    user.value = response.user
    localStorage.setItem('admin_token', accessToken)
    // 也保存 refresh_token
    if (response.refresh_token) {
      localStorage.setItem('refresh_token', response.refresh_token)
    }
    return response
  } catch (error) {
    throw error
  }
}
```

**关键改进**:
- ✅ 正确提取 `access_token`
- ✅ 兼容两种字段名（access_token 或 token）
- ✅ 保存 refresh_token 用于令牌刷新

---

#### 修改 2: initUser 方法

**修复前**:
```javascript
const initUser = async () => {
  if (token.value) {
    try {
      const userInfo = await getCurrentUser()
      user.value = userInfo  // ❌ 可能不正确
    } catch (error) {
      console.error('获取用户信息失败:', error)
      logout()  // ❌ logout 未定义
    }
  }
}
```

**修复后**:
```javascript
const initUser = async () => {
  if (token.value) {
    try {
      const response = await getCurrentUser()
      // 后端返回的数据结构可能是 { user: {...} } 或直接是用户对象
      user.value = response.user || response
    } catch (error) {
      console.error('获取用户信息失败:', error)
      logoutUser()  // ✅ 使用正确的方法名
    }
  }
}
```

**关键改进**:
- ✅ 兼容两种返回格式
- ✅ 修正方法名（logoutUser）

---

## ✅ 验证

### 数据流验证

**1. 登录流程**:
```
用户点击登录
   ↓
Login.vue: handleLogin()
   ↓
userStore.loginUser({ username, password })
   ↓
API: POST /api/auth/login
   ↓
后端返回: { user: {...}, access_token: "...", ... }
   ↓
提取 access_token: response.access_token ✅
   ↓
保存到 localStorage: admin_token = "eyJhbGci..." ✅
   ↓
设置 user.value = response.user ✅
   ↓
isLoggedIn = true ✅
   ↓
Login.vue: router.push('/') ✅
   ↓
路由守卫: isLoggedIn = true, 允许访问 ✅
   ↓
跳转到 Dashboard ✅
```

**2. localStorage 内容**:
```javascript
localStorage.getItem('admin_token')
// 应该是: "eyJhbGci..."
// 不应该是: undefined 或 "undefined"
```

**3. 路由守卫检查**:
```javascript
isLoggedIn = computed(() => !!token.value)
// token.value = "eyJhbGci..." → !!token.value = true ✅
```

---

## 📋 完整修复文件

### im-admin/src/stores/user.js（完整内容）

```javascript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, logout, getCurrentUser } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('admin_token') || '')
  
  const isLoggedIn = computed(() => !!token.value)
  
  const initUser = async () => {
    if (token.value) {
      try {
        const response = await getCurrentUser()
        // 后端返回的数据结构可能是 { user: {...} } 或直接是用户对象
        user.value = response.user || response
      } catch (error) {
        console.error('获取用户信息失败:', error)
        logoutUser()
      }
    }
  }
  
  const loginUser = async (credentials) => {
    try {
      const response = await login(credentials)
      // 后端返回的是 access_token 和 refresh_token，不是 token
      const accessToken = response.access_token || response.token
      token.value = accessToken
      user.value = response.user
      localStorage.setItem('admin_token', accessToken)
      // 也保存 refresh_token
      if (response.refresh_token) {
        localStorage.setItem('refresh_token', response.refresh_token)
      }
      return response
    } catch (error) {
      throw error
    }
  }
  
  const logoutUser = async () => {
    try {
      await logout()
    } catch (error) {
      console.error('登出失败:', error)
    } finally {
      token.value = ''
      user.value = null
      localStorage.removeItem('admin_token')
      localStorage.removeItem('refresh_token')
    }
  }
  
  return {
    user,
    token,
    isLoggedIn,
    initUser,
    loginUser,
    logoutUser
  }
})
```

---

## 🚀 部署到服务器

### Devin 执行步骤

```bash
# 1. 连接到主服务器
ssh root@154.37.214.191

# 2. 进入项目目录
cd /root/im-suite

# 3. 拉取最新代码（包含修复）
git pull origin main

# 4. 修改 im-admin/src/stores/user.js
# (使用上面的完整内容替换)

# 或者直接应用补丁：
cat > /tmp/user.js.patch << 'EOF'
--- a/im-admin/src/stores/user.js
+++ b/im-admin/src/stores/user.js
@@ -11,7 +11,8 @@
   const initUser = async () => {
     if (token.value) {
       try {
-        const userInfo = await getCurrentUser()
-        user.value = userInfo
+        const response = await getCurrentUser()
+        // 后端返回的数据结构可能是 { user: {...} } 或直接是用户对象
+        user.value = response.user || response
       } catch (error) {
         console.error('获取用户信息失败:', error)
-        logout()
+        logoutUser()
       }
     }
   }
@@ -23,8 +24,13 @@
   const loginUser = async (credentials) => {
     try {
       const response = await login(credentials)
-      token.value = response.token
+      // 后端返回的是 access_token 和 refresh_token，不是 token
+      const accessToken = response.access_token || response.token
+      token.value = accessToken
       user.value = response.user
-      localStorage.setItem('admin_token', response.token)
+      localStorage.setItem('admin_token', accessToken)
+      // 也保存 refresh_token
+      if (response.refresh_token) {
+        localStorage.setItem('refresh_token', response.refresh_token)
+      }
       return response
     } catch (error) {
@@ -41,6 +47,7 @@
       token.value = ''
       user.value = null
       localStorage.removeItem('admin_token')
+      localStorage.removeItem('refresh_token')
     }
   }
EOF

# 5. 重新构建管理后台
docker-compose -f docker-compose.partial.yml build --no-cache admin

# 6. 重启管理后台
docker-compose -f docker-compose.partial.yml restart admin

# 7. 等待启动
sleep 15

# 8. 验证
docker ps | grep admin
curl -I http://154.37.214.191:3001

# 9. 浏览器测试
# 打开 http://154.37.214.191:3001
# 登录 admin / Admin123!
# 应该成功跳转到仪表盘
```

---

## 🧪 测试清单

### 浏览器测试
- [ ] 打开 http://154.37.214.191:3001
- [ ] 输入账号: admin
- [ ] 输入密码: Admin123!
- [ ] 点击登录
- [ ] 等待响应（应该< 1秒）
- [ ] 看到"登录成功"提示 ✅
- [ ] 自动跳转到仪表盘 ✅
- [ ] 仪表盘显示统计数据 ✅

### 开发者工具验证
```javascript
// F12 打开控制台
// Application → Local Storage → http://154.37.214.191:3001

应该看到:
admin_token: "eyJhbGci..." ✅ (不是 undefined)
refresh_token: "eyJhbGci..." ✅
```

### Console 日志验证
```javascript
// 登录成功后，Console 应该没有错误
// 不应该看到:
❌ "获取用户信息失败"
❌ "路由守卫阻止"
❌ Uncaught TypeError

// 应该看到:
✅ POST /api/auth/login 200
✅ 登录成功
✅ 跳转到 /
```

---

## 🎯 关键点总结

### 问题
1. ❌ `response.token` 不存在（后端返回 `access_token`）
2. ❌ localStorage 存储了 `undefined`
3. ❌ `isLoggedIn` 计算为 `false`
4. ❌ 路由守卫阻止跳转

### 修复
1. ✅ 使用 `response.access_token`
2. ✅ 正确存储 token 到 localStorage
3. ✅ `isLoggedIn` 正确计算为 `true`
4. ✅ 路由守卫允许跳转

---

## 📊 前后端字段映射

| 后端字段 | 前端字段 | 说明 |
|---------|---------|------|
| `access_token` | `token` | 访问令牌 |
| `refresh_token` | - | 刷新令牌（新增存储） |
| `user` | `user` | 用户信息 |
| `expires_in` | - | 过期时间 |
| `requires_2fa` | - | 是否需要2FA |

---

## ✅ 修复后的完整流程

```
1. 用户点击登录
   ↓
2. Login.vue 调用 userStore.loginUser()
   ↓
3. API 请求 POST /api/auth/login
   ↓
4. 后端返回:
   {
     user: {..., role: "admin"},
     access_token: "eyJ...",
     refresh_token: "eyJ...",
     ...
   }
   ↓
5. 前端提取 access_token:
   token.value = response.access_token ✅
   user.value = response.user ✅
   ↓
6. 保存到 localStorage:
   admin_token = "eyJ..." ✅
   refresh_token = "eyJ..." ✅
   ↓
7. isLoggedIn = true ✅
   ↓
8. Login.vue: router.push('/') ✅
   ↓
9. 路由守卫检查:
   isLoggedIn = true → 允许访问 ✅
   user.role = "admin" → 管理员权限 ✅
   ↓
10. 跳转到 Dashboard ✅
```

---

## 🎉 修复完成

修复后，登录流程应该正常工作：
- ✅ 显示"登录成功"
- ✅ 自动跳转到仪表盘
- ✅ 显示用户信息和统计数据

如果还有问题，请检查：
1. 浏览器 Console 是否有错误
2. localStorage 中 admin_token 是否正确保存
3. 网络请求是否返回完整数据

