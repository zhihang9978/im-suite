# 环境变量配置示例

## 📋 使用说明

1. 创建 `.env` 文件
2. 复制下面的内容到 `.env` 文件
3. 修改所有 `your_*` 的值为实际值
4. 确保 `.env` 文件不要提交到Git仓库

---

## ⚙️ 完整环境变量配置

```bash
# =====================================================
# 数据库配置 (MySQL)
# =====================================================
DB_HOST=localhost
DB_PORT=3306
DB_USER=zhihang_messenger
DB_PASSWORD=your_secure_database_password
DB_NAME=zhihang_messenger

# MySQL Root 密码（Docker部署使用）
MYSQL_ROOT_PASSWORD=your_mysql_root_password
MYSQL_DATABASE=zhihang_messenger
MYSQL_USER=zhihang_messenger
MYSQL_PASSWORD=your_secure_database_password

# =====================================================
# Redis配置
# =====================================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# =====================================================
# MinIO对象存储配置
# =====================================================
MINIO_ENDPOINT=localhost:9000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_minio_secure_password
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=your_minio_secure_password
MINIO_BUCKET=im-files
MINIO_USE_SSL=false

# =====================================================
# JWT认证配置
# =====================================================
# 注意：JWT_SECRET必须是强密钥，至少32个字符
JWT_SECRET=your_jwt_secret_key_at_least_32_characters_long
JWT_EXPIRES_IN=24h
REFRESH_TOKEN_EXPIRES_IN=168h  # 7天

# =====================================================
# 服务器配置
# =====================================================
PORT=8080
GIN_MODE=release  # development | release | test

# =====================================================
# 文件上传限制
# =====================================================
MAX_FILE_SIZE=10485760      # 10MB (单个文件)
MAX_UPLOAD_SIZE=52428800    # 50MB (总上传大小)

# =====================================================
# 双因子认证配置
# =====================================================
TWO_FACTOR_ISSUER=ZhihangMessenger

# =====================================================
# CORS跨域配置
# =====================================================
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOW_CREDENTIALS=true

# =====================================================
# 日志配置
# =====================================================
LOG_LEVEL=info  # debug | info | warn | error
LOG_FORMAT=json  # json | text

# =====================================================
# 监控配置
# =====================================================
PROMETHEUS_ENABLED=true
METRICS_PORT=9090

# =====================================================
# 安全配置
# =====================================================
# 限流配置
RATE_LIMIT_REQUESTS_PER_SECOND=10
RATE_LIMIT_BURST=20

# Session配置
SESSION_TIMEOUT=3600  # 1小时

# IP黑名单检查
IP_BLACKLIST_ENABLED=true

# =====================================================
# WebRTC配置
# =====================================================
WEBRTC_STUN_SERVERS=stun:stun.l.google.com:19302
WEBRTC_TURN_SERVERS=

# =====================================================
# 邮件配置（可选）
# =====================================================
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=your_smtp_password
SMTP_FROM=ZhihangMessenger <noreply@example.com>

# =====================================================
# SMS短信配置（可选）
# =====================================================
SMS_PROVIDER=aliyun  # aliyun | tencent
SMS_ACCESS_KEY=your_sms_access_key
SMS_SECRET_KEY=your_sms_secret_key
SMS_SIGN_NAME=志航密信
SMS_TEMPLATE_CODE=SMS_123456789

# =====================================================
# 备份配置
# =====================================================
BACKUP_ENABLED=true
BACKUP_SCHEDULE=0 2 * * *  # 每天凌晨2点
BACKUP_RETENTION_DAYS=30
BACKUP_PATH=/backups
```

---

## 🔐 必须修改的配置

以下配置必须修改为实际值，不能使用示例值：

1. ✅ **MYSQL_ROOT_PASSWORD** - MySQL root密码
2. ✅ **DB_PASSWORD** / **MYSQL_PASSWORD** - 数据库用户密码
3. ✅ **REDIS_PASSWORD** - Redis密码
4. ✅ **MINIO_ROOT_PASSWORD** / **MINIO_SECRET_KEY** - MinIO密码
5. ✅ **JWT_SECRET** - JWT密钥（至少32个字符）

---

## ⚠️ 安全提示

- 所有密码必须使用强密码（至少16位，包含大小写字母、数字、特殊字符）
- JWT_SECRET必须保密，泄露后需要立即更换
- 定期更换密码和密钥
- `.env` 文件已在 `.gitignore` 中，不会被提交到Git

