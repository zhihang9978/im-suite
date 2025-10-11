# 🚀 志航密信 - 快速开始指南

**适用人员**: Devin或其他运维人员  
**前置条件**: Linux服务器 + root权限  
**预计时间**: 30-60分钟

---

## ⚡ 5分钟快速部署

```bash
# 1. 克隆项目
cd /opt
git clone https://github.com/zhihang9978/im-suite.git
cd im-suite

# 2. 系统初始化
sudo bash ops/bootstrap.sh

# 3. 配置环境变量
cp .env.example .env
vim .env  # 填写数据库密码、JWT密钥等

# 4. 配置TURN服务器
sudo bash ops/setup-turn.sh

# 5. 配置SSL
sudo bash ops/setup-ssl.sh

# 6. 部署应用
bash ops/deploy.sh

# 7. 验证部署
bash ops/smoke.sh
```

**完成！服务已启动** ✅

---

## 📋 必填配置项

编辑`.env`文件，至少填写以下变量：

```bash
# 数据库（必填）
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=<生成强密码>
DB_NAME=zhihang_messenger

# Redis（必填）
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=<生成强密码>

# JWT（必填）
JWT_SECRET=<生成48字符密钥>

# Docker Compose（必填）
MYSQL_ROOT_PASSWORD=<生成强密码>
REDIS_PASSWORD=<生成强密码>
MINIO_ROOT_PASSWORD=<生成强密码>
```

**生成密码命令**:
```bash
# JWT密钥（48字符）
openssl rand -base64 48

# 其他密码（32字符）
openssl rand -base64 32
```

---

## 🧪 验证部署

### 1. 健康检查
```bash
curl http://localhost:8080/health
# 预期：{"status":"ok",...}
```

### 2. 服务状态
```bash
docker-compose -f docker-compose.production.yml ps
# 预期：所有服务 Up (healthy)
```

### 3. 运行冒烟测试
```bash
bash ops/smoke.sh
# 预期：所有检查通过
```

### 4. 查看日志
```bash
docker logs im-backend-prod --tail=50
# 预期：无ERROR日志
```

---

## 🔧 常用运维命令

### 查看服务
```bash
docker-compose -f docker-compose.production.yml ps
docker-compose -f docker-compose.production.yml logs -f backend
```

### 重启服务
```bash
docker-compose -f docker-compose.production.yml restart backend
```

### 停止服务
```bash
docker-compose -f docker-compose.production.yml down
```

### 备份数据
```bash
bash ops/backup_restore.sh backup
```

### 恢复数据
```bash
bash ops/backup_restore.sh restore 20251011-150000
```

### 回滚版本
```bash
bash ops/rollback.sh 20251011-150000
```

---

## 📊 监控访问

### Grafana
- URL: http://server-ip:3000
- 用户名: admin
- 密码: admin（首次登录后修改）

### Prometheus
- URL: http://server-ip:9090

---

## ⚠️ 常见问题

### Q1: 服务启动失败
```bash
# 查看日志
docker logs im-backend-prod

# 常见原因：
# 1. 环境变量未配置
# 2. 端口被占用
# 3. 数据库连接失败
```

### Q2: 数据库连接失败
```bash
# 检查MySQL状态
docker exec im-mysql-prod mysql -uroot -p

# 检查.env配置
cat .env | grep DB_
```

### Q3: 前端无法访问
```bash
# 检查Nginx状态
systemctl status nginx

# 查看Nginx配置
nginx -t

# 重启Nginx
systemctl restart nginx
```

---

## 📞 获取帮助

- **文档**: `docs/` 目录
- **脚本**: `ops/` 目录
- **问题**: GitHub Issues

---

**快速开始指南 v1.0**  
**更新时间**: 2025-10-11

