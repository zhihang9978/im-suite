# 环境变量配置说明

## 📋 目录

- [应用基础配置](#应用基础配置)
- [数据库配置](#数据库配置)
- [Redis配置](#redis配置)
- [JWT认证配置](#jwt认证配置)
- [对象存储配置](#对象存储配置)
- [管理员配置](#管理员配置)
- [邮件服务配置](#邮件服务配置)
- [WebRTC音视频配置](#webrtc音视频配置)
- [推送通知配置](#推送通知配置)
- [安全配置](#安全配置)
- [监控追踪配置](#监控追踪配置)
- [合规配置](#合规配置)

---

## 应用基础配置

### `APP_BASE_URL` [必填]
- **用途**: 应用的基础URL，用于生成链接和回调
- **范围**: 完整的URL（包含协议）
- **示例**: `https://im.example.com`
- **默认值**: 无

### `APP_NAME`
- **用途**: 应用显示名称
- **示例**: `志航密信`
- **默认值**: `志航密信`

### `APP_ENV`
- **用途**: 环境标识，影响日志级别和调试模式
- **范围**: `development` | `staging` | `production`
- **示例**: `production`
- **默认值**: `production`

### `BACKEND_PORT`
- **用途**: 后端API服务监听端口
- **范围**: 1-65535
- **示例**: `8080`
- **默认值**: `8080`

### `ADMIN_PORT`
- **用途**: 管理后台前端服务端口
- **范围**: 1-65535
- **示例**: `3001`
- **默认值**: `3001`

---

## 数据库配置

### `MYSQL_HOST` [必填]
- **用途**: MySQL服务器地址
- **示例**: `mysql` (Docker网络) 或 `192.168.1.100`
- **默认值**: `mysql`

### `MYSQL_PORT`
- **用途**: MySQL服务端口
- **范围**: 1-65535
- **示例**: `3306`
- **默认值**: `3306`

### `MYSQL_DATABASE` [必填]
- **用途**: 数据库名称
- **示例**: `im_suite`
- **默认值**: `im_suite`

### `MYSQL_USER` [必填]
- **用途**: 数据库用户名
- **示例**: `im_user`
- **默认值**: `im_user`

### `MYSQL_PASSWORD` [必填]
- **用途**: 数据库密码
- **范围**: 至少32字符，包含大小写字母、数字、特殊字符
- **示例**: `Kj8#mP2@qW9$nR5&xT7*vB3!zA4%cF6`
- **生成方法**: `openssl rand -base64 32`

### `MYSQL_ROOT_PASSWORD` [必填]
- **用途**: MySQL root用户密码
- **范围**: 至少32字符
- **示例**: 同上
- **生成方法**: `openssl rand -base64 32`

### `MYSQL_MAX_OPEN_CONNS`
- **用途**: 最大打开连接数
- **范围**: 10-1000
- **示例**: `100`
- **默认值**: `100`
- **调优建议**: 根据并发量调整，公式：`并发用户数 / 10`

### `MYSQL_MAX_IDLE_CONNS`
- **用途**: 最大空闲连接数
- **范围**: 5-100
- **示例**: `10`
- **默认值**: `10`

### `MYSQL_CONN_MAX_LIFETIME`
- **用途**: 连接最大生命周期（秒）
- **范围**: 60-3600
- **示例**: `1800` (30分钟)
- **默认值**: `1800`

---

## Redis配置

### `REDIS_HOST` [必填]
- **用途**: Redis服务器地址
- **示例**: `redis` (Docker网络) 或 `192.168.1.100`
- **默认值**: `redis`

### `REDIS_PORT`
- **用途**: Redis服务端口
- **范围**: 1-65535
- **示例**: `6379`
- **默认值**: `6379`

### `REDIS_PASSWORD` [必填]
- **用途**: Redis密码
- **范围**: 至少32字符
- **示例**: `aB3$cD5#eF7@gH9!iJ1&kL2*mN4%oP6`
- **生成方法**: `openssl rand -base64 32`

### `REDIS_DB`
- **用途**: Redis数据库编号
- **范围**: 0-15
- **示例**: `0`
- **默认值**: `0`

### `REDIS_POOL_SIZE`
- **用途**: 连接池大小
- **范围**: 5-50
- **示例**: `10`
- **默认值**: `10`

---

## JWT认证配置

### `JWT_SECRET` [必填]
- **用途**: JWT签名密钥
- **范围**: 至少32字符，建议48字符以上
- **示例**: `RtY8#uI9@oP0!aS1&dF2*gH3%jK4^lZ5$xC6&vB7*nM8#qW9`
- **生成方法**: `openssl rand -base64 48`
- **⚠️ 警告**: 泄露将导致所有token失效，定期轮换（建议90天）

### `JWT_EXPIRE_TIME`
- **用途**: Access Token过期时间（秒）
- **范围**: 300-86400 (5分钟-24小时)
- **示例**: `86400` (24小时)
- **默认值**: `86400`
- **建议**: 移动端24小时，Web端8小时

### `JWT_REFRESH_EXPIRE_TIME`
- **用途**: Refresh Token过期时间（秒）
- **范围**: 86400-2592000 (1天-30天)
- **示例**: `604800` (7天)
- **默认值**: `604800`

---

## 对象存储配置

### MinIO配置

#### `MINIO_ENDPOINT` [必填]
- **用途**: MinIO服务地址
- **示例**: `minio:9000` (Docker网络) 或 `minio.example.com`
- **默认值**: `minio:9000`

#### `MINIO_ROOT_USER` [必填]
- **用途**: MinIO管理员用户名
- **示例**: `admin`
- **默认值**: `admin`

#### `MINIO_ROOT_PASSWORD` [必填]
- **用途**: MinIO管理员密码
- **范围**: 至少32字符
- **示例**: `qW3#eR5@tY7!uI9&oP1*aS2$dF4%gH6`
- **生成方法**: `openssl rand -base64 32`

#### `MINIO_BUCKET` [必填]
- **用途**: 默认存储桶名称
- **示例**: `im-files`
- **默认值**: `im-files`

#### `MINIO_USE_SSL`
- **用途**: 是否使用HTTPS连接
- **范围**: `true` | `false`
- **示例**: `true` (生产环境)
- **默认值**: `false`

### 阿里云OSS配置（可选）

#### `OSS_ENABLED`
- **用途**: 是否启用阿里云OSS（替代MinIO）
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `false`

#### `OSS_ENDPOINT`
- **用途**: OSS区域端点
- **示例**: `oss-cn-hangzhou.aliyuncs.com`

#### `OSS_ACCESS_KEY_ID`
- **用途**: OSS访问密钥ID
- **示例**: `LTAI5txxxxxxxxxx`

#### `OSS_ACCESS_KEY_SECRET`
- **用途**: OSS访问密钥Secret
- **示例**: `xxxxxxxxxxxxxxxxxxxxx`

#### `OSS_BUCKET`
- **用途**: OSS存储桶名称
- **示例**: `im-suite-prod`

---

## WebRTC音视频配置

### `STUN_SERVER`
- **用途**: STUN服务器地址（用于NAT穿透）
- **示例**: `stun:stun.l.google.com:19302`
- **默认值**: `stun:stun.l.google.com:19302`

### `TURN_ENABLED` [必填]
- **用途**: 是否启用TURN服务器
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`
- **说明**: 生产环境强烈建议启用

### `TURN_SERVER` [必填]
- **用途**: TURN服务器地址
- **格式**: `turn:host:port`
- **示例**: `turn:turn.example.com:3478`

### `TURN_USERNAME` [必填]
- **用途**: TURN服务器用户名
- **示例**: `turn_user`
- **默认值**: `turn_user`

### `TURN_PASSWORD` [必填]
- **用途**: TURN服务器密码
- **范围**: 至少16字符
- **示例**: `TurnP@ss2024!Secure`
- **生成方法**: `openssl rand -base64 24`

### `TURN_REALM` [必填]
- **用途**: TURN服务器域
- **示例**: `example.com`

### `TURN_PUBLIC_IPS` [必填]
- **用途**: TURN服务器公网IP列表（逗号分隔）
- **示例**: `1.2.3.4,5.6.7.8`
- **说明**: 多个IP可提高连通率

### `TURN_MIN_PORT`
- **用途**: TURN中继端口范围起始
- **范围**: 1024-49151
- **示例**: `49152`
- **默认值**: `49152`

### `TURN_MAX_PORT`
- **用途**: TURN中继端口范围结束
- **范围**: 49152-65535
- **示例**: `65535`
- **默认值**: `65535`

### SFU配置（可选）

#### `SFU_ENABLED`
- **用途**: 是否启用SFU服务器（多方通话）
- **范围**: `true` | `false`
- **示例**: `false`
- **默认值**: `false`

#### `SFU_TYPE`
- **用途**: SFU服务器类型
- **范围**: `ion-sfu` | `mediasoup` | `janus`
- **示例**: `ion-sfu`
- **默认值**: `ion-sfu`

#### `SFU_PUBLIC_IP`
- **用途**: SFU服务器公网IP
- **示例**: `1.2.3.4`

#### `SFU_WS_URL`
- **用途**: SFU WebSocket连接地址
- **示例**: `wss://sfu.example.com:7000`

---

## 安全配置

### `CORS_ALLOWED_ORIGINS` [必填]
- **用途**: CORS允许的源列表（逗号分隔）
- **示例**: `https://im.example.com,https://www.example.com`
- **⚠️ 警告**: 不要使用 `*`，必须明确列出所有允许的域名

### `RATE_LIMIT_ENABLED`
- **用途**: 是否启用速率限制
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `RATE_LIMIT_REQUESTS`
- **用途**: 速率限制请求数
- **范围**: 10-1000
- **示例**: `60` (每分钟60次)
- **默认值**: `60`

### `RATE_LIMIT_WINDOW`
- **用途**: 速率限制时间窗口（秒）
- **范围**: 1-3600
- **示例**: `60` (1分钟)
- **默认值**: `60`

### `PASSWORD_MIN_LENGTH`
- **用途**: 密码最小长度
- **范围**: 8-32
- **示例**: `8`
- **默认值**: `8`

### `PASSWORD_REQUIRE_UPPERCASE`
- **用途**: 密码必须包含大写字母
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `PASSWORD_REQUIRE_LOWERCASE`
- **用途**: 密码必须包含小写字母
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `PASSWORD_REQUIRE_NUMBER`
- **用途**: 密码必须包含数字
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `PASSWORD_REQUIRE_SPECIAL`
- **用途**: 密码必须包含特殊字符
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `CAPTCHA_ENABLED`
- **用途**: 是否启用验证码
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `ENCRYPTION_KEY` [必填]
- **用途**: 数据加密密钥
- **范围**: 32字符（256位）
- **示例**: `YourBase64EncodedEncryptionKeyHere==`
- **生成方法**: `openssl rand -base64 32`

---

## 监控追踪配置

### `METRICS_ENABLED`
- **用途**: 是否启用Prometheus指标导出
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `METRICS_PORT`
- **用途**: Prometheus指标导出端口
- **范围**: 1-65535
- **示例**: `9090`
- **默认值**: `9090`

### `METRICS_PATH`
- **用途**: Prometheus指标路径
- **示例**: `/metrics`
- **默认值**: `/metrics`

### `SENTRY_ENABLED`
- **用途**: 是否启用Sentry异常追踪
- **范围**: `true` | `false`
- **示例**: `false`
- **默认值**: `false`

### `SENTRY_DSN`
- **用途**: Sentry项目DSN
- **示例**: `https://xxxxx@o123456.ingest.sentry.io/7654321`

### `SENTRY_ENVIRONMENT`
- **用途**: Sentry环境标识
- **示例**: `production`
- **默认值**: `production`

### `SENTRY_TRACES_SAMPLE_RATE`
- **用途**: Sentry性能追踪采样率
- **范围**: 0.0-1.0 (0-100%)
- **示例**: `0.1` (10%)
- **默认值**: `0.1`

---

## 日志配置

### `LOG_LEVEL`
- **用途**: 日志级别
- **范围**: `debug` | `info` | `warn` | `error`
- **示例**: `info` (生产环境)
- **默认值**: `info`

### `LOG_FORMAT`
- **用途**: 日志格式
- **范围**: `json` | `text`
- **示例**: `json` (生产环境推荐)
- **默认值**: `json`

### `LOG_FILE_PATH`
- **用途**: 日志文件路径
- **示例**: `/var/log/im-suite/backend.log`
- **默认值**: `/var/log/im-suite/backend.log`

### `LOG_MAX_SIZE`
- **用途**: 单个日志文件最大大小（MB）
- **范围**: 10-500
- **示例**: `100`
- **默认值**: `100`

### `LOG_MAX_AGE`
- **用途**: 日志保留天数
- **范围**: 1-90
- **示例**: `30`
- **默认值**: `30`

---

## 合规配置

### `REAL_NAME_VERIFICATION_ENABLED`
- **用途**: 是否启用实名认证
- **范围**: `true` | `false`
- **示例**: `false`
- **默认值**: `false`

### `CONTENT_MODERATION_ENABLED`
- **用途**: 是否启用内容审核
- **范围**: `true` | `false`
- **示例**: `true`
- **默认值**: `true`

### `CONTENT_MODERATION_PROVIDER`
- **用途**: 内容审核服务提供商
- **范围**: `aliyun` | `tencent`
- **示例**: `aliyun`
- **默认值**: `aliyun`

### `LOG_RETENTION_DAYS`
- **用途**: 日志留存天数（合规要求）
- **范围**: 30-180
- **示例**: `90` (建议90天)
- **默认值**: `90`

---

## 🔐 安全最佳实践

### 1. 密码管理
```bash
# 生成JWT密钥（48字节）
openssl rand -base64 48

# 生成数据库密码（32字节）
openssl rand -base64 32

# 生成加密密钥（32字节）
openssl rand -base64 32
```

### 2. 文件权限
```bash
# 设置.env文件权限（仅所有者可读写）
chmod 600 .env
chown root:root .env
```

### 3. 环境变量验证
```bash
# 启动前检查必需变量
bash ops/validate-env.sh
```

### 4. 定期轮换
- JWT_SECRET: 每90天轮换一次
- 数据库密码: 每180天轮换一次
- API密钥: 每90天轮换一次

### 5. 生产环境检查清单
- [ ] 所有`[必填]`变量都已配置
- [ ] 所有密码至少32字符
- [ ] DEBUG=false
- [ ] SWAGGER_ENABLED=false
- [ ] CORS_ALLOWED_ORIGINS不使用 `*`
- [ ] 启用HTTPS（MINIO_USE_SSL=true）
- [ ] 启用速率限制（RATE_LIMIT_ENABLED=true）
- [ ] 启用验证码（CAPTCHA_ENABLED=true）
- [ ] 配置TURN服务器
- [ ] 配置监控（METRICS_ENABLED=true）
- [ ] 配置备份（BACKUP_ENABLED=true）

---

## 📚 相关文档

- [生产部署手册](./production/生产部署手册.md)
- [运维手册](./production/运维手册.md)
- [安全配置指南](../docs/security/transport-security.md)
- [备份恢复指南](./production/备份恢复指南.md)

---

**更新时间**: 2025-10-11  
**维护者**: zhihang9978

