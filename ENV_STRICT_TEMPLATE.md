# 严格模式环境变量模板

**重要**: 所有标记为 `必须` 的变量都不能为空，否则Docker Compose将拒绝启动！

---

## 数据库配置（必须）

```bash
# MySQL root用户密码（必须，>16字符）
MYSQL_ROOT_PASSWORD=your_secure_root_password_here

# 应用数据库名（必须）
MYSQL_DATABASE=zhihang_messenger

# 应用数据库用户（必须）
MYSQL_USER=zhihang_user

# 应用数据库密码（必须，>16字符）
MYSQL_PASSWORD=your_secure_db_password_here
```

---

## Redis配置（必须）

```bash
# Redis密码（必须，>16字符）
REDIS_PASSWORD=your_secure_redis_password_here
```

---

## MinIO配置（必须）

```bash
# MinIO管理员用户名（必须，>8字符）
MINIO_ROOT_USER=zhihang_admin

# MinIO管理员密码（必须，>16字符）
MINIO_ROOT_PASSWORD=your_secure_minio_password_here
```

---

## JWT配置（必须）

```bash
# JWT密钥（必须，>32字符随机字符串）
JWT_SECRET=your_super_secret_jwt_key_at_least_32_characters_long_random_string
```

---

## 管理后台配置（必须）

```bash
# 管理后台API地址（必须）
ADMIN_API_BASE_URL=http://your-server-ip:8080
```

---

## WebRTC配置（可选）

```bash
# STUN/TURN服务器（可选）
WEBRTC_ICE_SERVERS='[{"urls":"stun:stun.l.google.com:19302"}]'
```

---

## 使用说明

### 1. 创建.env文件

```bash
# 复制模板
cp ENV_STRICT_TEMPLATE.md .env

# 编辑配置
nano .env  # 或使用你喜欢的编辑器
```

### 2. 生成强密码

```bash
# 使用openssl生成32字符随机密码
openssl rand -base64 32

# 或使用pwgen
pwgen -s 32 1
```

### 3. 验证配置

```bash
# 验证环境变量（会显示缺失的变量）
docker-compose -f docker-compose.production.yml config
```

### 4. 启动服务

```bash
# 使用部署脚本（会自动检查）
./scripts/deploy_prod.sh

# 或直接使用docker-compose
docker-compose -f docker-compose.production.yml up -d
```

---

## 严格模式说明

### 环境变量硬失败机制

docker-compose.production.yml 中已配置：

```yaml
x-environment-check: &env-check
  - ${MYSQL_ROOT_PASSWORD:?请在.env中设置MYSQL_ROOT_PASSWORD}
  - ${MYSQL_DATABASE:?请在.env中设置MYSQL_DATABASE}
  # ... 更多必须变量
```

**效果**: 如果任何必须的环境变量未设置，Docker Compose会立即失败并显示错误信息，而不是使用隐式默认值。

---

## 安全默认原则

### 1. 端口暴露最小化

**修改内容**:
- ✅ MySQL: 不对外暴露（仅内部网络）
- ✅ Redis: 不对外暴露（仅内部网络）
- ✅ MinIO: 不对外暴露（仅内部网络）
- ✅ 后端: 保留8080端口（需要外部访问）
- ✅ 管理后台: 保留3001端口（需要外部访问）

**访问方式**:
- 后端API: 通过Nginx反向代理
- 管理后台: 通过Nginx
- 数据库/缓存/存储: 仅服务间内部访问

### 2. 密码强度要求

**最低要求**:
- MySQL密码: >= 16字符
- Redis密码: >= 16字符
- MinIO密码: >= 16字符
- JWT密钥: >= 32字符

**推荐**:
- 使用大小写字母+数字+特殊字符
- 每个服务使用不同密码
- 定期更换密码（每90天）

### 3. .env文件安全

**已配置**:
- `.gitignore` 已包含 `.env`
- `.cursorignore` 已包含 `.env`

**注意**:
- ❌ 不要将.env提交到Git
- ❌ 不要在文档中包含真实密码
- ✅ 仅在服务器上存在.env文件
- ✅ 设置正确的文件权限: `chmod 600 .env`

---

## 故障排查

### 错误: "请在.env中设置XXX"

**原因**: 缺少必要的环境变量

**解决**:
1. 检查.env文件是否存在
2. 确认变量名拼写正确
3. 确认变量有值（不为空）

### 错误: "network error"

**原因**: 服务无法连接

**解决**:
1. 检查healthcheck状态: `docker-compose ps`
2. 查看服务日志: `docker-compose logs [service]`
3. 确认内部网络正常: `docker network ls`

### 错误: "authentication failed"

**原因**: 密码不正确

**解决**:
1. 确认.env中的密码正确
2. 删除数据卷重新初始化: `docker-compose down -v`
3. 重新启动: `docker-compose up -d`

---

## 生产环境checklist

- [ ] 所有密码都已设置为强密码
- [ ] .env文件权限设置为600
- [ ] .env文件已从Git中排除
- [ ] 已测试docker-compose config成功
- [ ] 所有服务healthcheck正常
- [ ] 数据卷已配置备份
- [ ] 日志轮转已配置
- [ ] 监控告警已配置

---

**安全第一！永远不要在生产环境使用默认密码！** 🔒

