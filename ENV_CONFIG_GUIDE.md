# 环境配置指南

## 📋 当前环境：测试模式（localhost）

### ✅ 测试环境特点
- 使用localhost，无需域名
- 可使用HTTP（推荐）或HTTPS（自签名证书）
- Devin可以直接测试
- 无需复杂配置

---

## 🔧 配置文件说明

### 测试环境配置
复制 `ENV_TEMPLATE.md` 中的测试环境配置，创建 `.env` 文件

**关键配置**：
```bash
# 测试环境
SERVER_HOST=localhost
SERVER_PORT=8080
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
SSL_ENABLED=false
```

### 生产环境配置
测试通过后，更新 `.env` 文件

**关键配置**：
```bash
# 生产环境
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
VITE_API_BASE_URL=https://api.yourdomain.com
VITE_WS_URL=wss://api.yourdomain.com/ws
SSL_ENABLED=true
```

---

## 🚀 快速切换

### 当前：测试阶段
```bash
# 直接启动，使用HTTP
docker-compose -f docker-compose.production.yml up -d

# 访问
http://localhost:8080      # 后端API
http://localhost:3002      # Web客户端
http://localhost:3001      # 管理后台
```

### 未来：生产部署
```bash
# 1. 获取域名和SSL证书（宝塔面板一键）
# 2. 修改nginx配置中的server_name
# 3. 修改前端环境变量
# 4. 重新部署
docker-compose -f docker-compose.production.yml up --build -d

# 访问
https://yourdomain.com     # Web客户端
https://api.yourdomain.com # 后端API
https://admin.yourdomain.com # 管理后台
```

---

## 📝 需要修改的文件清单

### 生产环境部署时需要修改：

1. **`config/nginx/nginx.conf`** （1处）
   - 第175行：`server_name _;` → `server_name yourdomain.com;`
   - 第177-178行：证书路径

2. **`telegram-web/.env.production`** （2处）
   - API地址
   - WebSocket地址

3. **`im-admin/.env.production`** （1处）
   - API地址

4. **`docker-compose.production.yml`** （1处）
   - nginx volumes证书路径

5. **`telegram-android/...network_security_config.xml`** （可选）
   - SSL Pinning配置

**总计**: 5个文件，约10分钟完成

---

## ✅ 当前状态

- 🟢 **测试环境**: 已配置，可直接使用
- 🟢 **所有代码**: 已完成，无错误
- 🟢 **编译测试**: 通过
- 🟢 **文档**: 完整

**Devin可以直接开始测试！** 🚀

---

**下一步**: Devin测试通过后，按本文档切换到生产环境

