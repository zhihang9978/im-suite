# 环境变量配置模板

> **重要**: 请复制以下内容到 `.env` 文件并修改所有密码和密钥！

---

## 📋 完整环境变量配置

### 方式一: 复制到 `.env` 文件

```bash
# ========================================
# 志航密信 - 生产环境配置
# ========================================

# ========================================
# 数据库配置 (MySQL)
# ========================================
DB_HOST=mysql
DB_PORT=3306
DB_NAME=zhihang_messenger
DB_USER=zhihang
DB_PASSWORD=ZhiHang_MySQL_P@ssw0rd_2024_CHANGE_ME

# Docker Compose 专用
MYSQL_ROOT_PASSWORD=ZhiHang_Root_P@ssw0rd_2024_CHANGE_ME
MYSQL_DATABASE=zhihang_messenger
MYSQL_USER=zhihang
MYSQL_PASSWORD=ZhiHang_MySQL_P@ssw0rd_2024_CHANGE_ME

# ========================================
# Redis 配置
# ========================================
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=ZhiHang_Redis_P@ssw0rd_2024_CHANGE_ME

# ========================================
# MinIO 对象存储配置
# ========================================
MINIO_ENDPOINT=minio:9000
MINIO_ROOT_USER=zhihang_minio_admin
MINIO_ROOT_PASSWORD=ZhiHang_MinIO_SecretKey_2024_CHANGE_ME
MINIO_ACCESS_KEY=zhihang_minio_admin
MINIO_SECRET_KEY=ZhiHang_MinIO_SecretKey_2024_CHANGE_ME
MINIO_USE_SSL=false

# ========================================
# JWT 认证配置
# ========================================
# 注意：必须至少 32 个字符！使用随机字符串！
JWT_SECRET=ZhiHang_JWT_Super_Secret_Key_Min_32_Chars_CHANGE_ME_To_Random_String
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# ========================================
# 服务配置
# ========================================
PORT=8080
GIN_MODE=release
LOG_LEVEL=info

# ========================================
# 前端配置
# ========================================
ADMIN_API_BASE_URL=http://backend:8080
WEB_API_BASE_URL=http://backend:8080
WEB_WS_BASE_URL=ws://backend:8080/ws

# ========================================
# 文件上传配置
# ========================================
MAX_FILE_SIZE=100MB
UPLOAD_PATH=/app/uploads

# ========================================
# WebRTC 音视频通话配置
# ========================================
WEBRTC_ICE_SERVERS=[{"urls":"stun:stun.l.google.com:19302"},{"urls":"stun:stun1.l.google.com:19302"}]

# ========================================
# 监控配置
# ========================================
GRAFANA_PASSWORD=ZhiHang_Grafana_Admin_2024_CHANGE_ME

# ========================================
# 域名配置（可选，使用域名时配置）
# ========================================
DOMAIN=your-domain.com
API_BASE_URL=https://api.your-domain.com
WEB_BASE_URL=https://your-domain.com

# ========================================
# HTTPS 配置
# ========================================
SSL_ENABLED=true
SSL_CERT_PATH=/etc/nginx/ssl/cert.pem
SSL_KEY_PATH=/etc/nginx/ssl/key.pem

# ========================================
# 安全配置
# ========================================
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=1m

# ========================================
# 性能优化配置
# ========================================
MESSAGE_QUEUE_SIZE=1000
MESSAGE_WORKERS=5
CACHE_TTL=3600
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
```

---

## 🔐 安全提示

### 必须修改的配置项

1. **DB_PASSWORD** - 数据库密码
2. **MYSQL_ROOT_PASSWORD** - MySQL root 密码
3. **REDIS_PASSWORD** - Redis 密码
4. **MINIO_ROOT_PASSWORD** - MinIO 密码
5. **JWT_SECRET** - JWT 密钥（至少 32 字符）
6. **GRAFANA_PASSWORD** - Grafana 管理员密码

### 密码要求

- 长度：至少 16 个字符
- 包含：大写字母、小写字母、数字、特殊字符
- 不要使用：常见单词、生日、公司名称

### 生成强密码示例

```bash
# 使用 OpenSSL 生成随机密码
openssl rand -base64 32

# 使用 pwgen (Linux)
pwgen -s 32 1

# 手动生成建议
# 例如: Zh!H@ng_M3ss3ng3r_2024_Pr0d_S3cur3
```

---

## 📝 配置步骤

### 1. 创建 .env 文件

```bash
cd /opt/im-suite
cp ENV_TEMPLATE.md .env
```

### 2. 修改所有密码

```bash
nano .env
# 或
vim .env
```

### 3. 验证配置

```bash
# 检查必填项
grep "CHANGE_ME" .env
# 如果有输出，说明还有未修改的密码！

# 验证 JWT_SECRET 长度
JWT_SECRET=$(grep JWT_SECRET .env | cut -d= -f2)
echo ${#JWT_SECRET}
# 应该 >= 32
```

### 4. 设置文件权限

```bash
chmod 600 .env
chown root:root .env
```

---

## 🚀 快速配置（一键生成强密码）

```bash
#!/bin/bash
# generate-secure-env.sh

cat > .env << EOF
# 自动生成的安全配置
DB_PASSWORD=$(openssl rand -base64 24)
MYSQL_ROOT_PASSWORD=$(openssl rand -base64 24)
REDIS_PASSWORD=$(openssl rand -base64 24)
MINIO_ROOT_PASSWORD=$(openssl rand -base64 24)
JWT_SECRET=$(openssl rand -base64 48)
GRAFANA_PASSWORD=$(openssl rand -base64 24)

# 其他固定配置
DB_HOST=mysql
DB_PORT=3306
DB_NAME=zhihang_messenger
DB_USER=zhihang
REDIS_HOST=redis
REDIS_PORT=6379
MINIO_ENDPOINT=minio:9000
PORT=8080
GIN_MODE=release
EOF

echo "✅ .env 文件已生成！请查看并确认配置。"
echo "⚠️ 请保存好这些密码，丢失后无法恢复！"
```

---

## 📋 环境变量说明

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `DB_HOST` | ✅ | mysql | 数据库主机 |
| `DB_PORT` | ✅ | 3306 | 数据库端口 |
| `DB_NAME` | ✅ | zhihang_messenger | 数据库名称 |
| `DB_USER` | ✅ | zhihang | 数据库用户 |
| `DB_PASSWORD` | ✅ | - | 数据库密码（必须修改） |
| `REDIS_HOST` | ✅ | redis | Redis 主机 |
| `REDIS_PORT` | ✅ | 6379 | Redis 端口 |
| `REDIS_PASSWORD` | ✅ | - | Redis 密码（必须修改） |
| `JWT_SECRET` | ✅ | - | JWT 密钥（至少 32 字符） |
| `PORT` | ✅ | 8080 | 后端服务端口 |
| `GIN_MODE` | ✅ | release | Gin 运行模式 |
| `MINIO_ENDPOINT` | ✅ | minio:9000 | MinIO 端点 |
| `GRAFANA_PASSWORD` | ✅ | - | Grafana 密码（必须修改） |

---

## ✅ 配置检查清单

部署前请确认：

- [ ] 所有包含 `CHANGE_ME` 的值都已修改
- [ ] JWT_SECRET 长度 >= 32 字符
- [ ] 所有密码都是强密码
- [ ] .env 文件权限设置为 600
- [ ] .env 文件未提交到 Git
- [ ] 已备份 .env 文件到安全位置

---

## 🔗 相关文档

- **部署指南**: `SERVER_DEPLOYMENT_INSTRUCTIONS.md`
- **缺陷报告**: `PROJECT_DEFECTS_AND_ISSUES_REPORT.md`
- **修复记录**: `DEFECT_FIXES_APPLIED.md`


