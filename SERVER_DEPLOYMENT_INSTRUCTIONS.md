# 志航密信 - 服务器完整部署指令

## 📋 服务器信息
- **服务器IP**: 154.37.214.191
- **用户**: root
- **部署路径**: /opt/im-suite

---

## 🗑️ 第一步：清除现有数据

请在SSH会话中执行以下命令：

```bash
# 1. 停止所有Docker容器
docker stop $(docker ps -aq) 2>/dev/null || true

# 2. 删除所有Docker容器
docker rm $(docker ps -aq) 2>/dev/null || true

# 3. 删除所有Docker镜像
docker rmi $(docker images -q) 2>/dev/null || true

# 4. 清理所有Docker数据卷（会删除数据库数据）
docker volume prune -f

# 5. 清理Docker网络
docker network prune -f

# 6. 清理Docker系统
docker system prune -a -f

# 7. 删除旧的项目目录
rm -rf /opt/im-suite
rm -rf /www/wwwroot/im-suite

# 8. 验证清理结果
echo "=== 清理完成 ==="
docker ps -a
docker images
docker volume ls
ls -la /opt/
```

---

## 🚀 第二步：完整部署

### 方式一：一键自动部署（推荐）

```bash
# 1. 创建部署目录
mkdir -p /opt/im-suite
cd /opt/im-suite

# 2. 下载部署脚本
wget https://raw.githubusercontent.com/zhihang9978/im-suite/main/server-deploy.sh

# 或使用curl
curl -O https://raw.githubusercontent.com/zhihang9978/im-suite/main/server-deploy.sh

# 3. 给予执行权限
chmod +x server-deploy.sh

# 4. 执行自动部署
./server-deploy.sh

# 脚本会自动完成：
# - ✅ 检测并安装Docker
# - ✅ 检测并安装Docker Compose
# - ✅ 克隆项目代码
# - ✅ 配置环境变量
# - ✅ 生成SSL证书
# - ✅ 创建数据目录
# - ✅ 启动所有服务
# - ✅ 显示访问地址
```

### 方式二：手动部署（详细控制）

```bash
# 1. 安装Docker
curl -fsSL https://get.docker.com | bash
systemctl start docker
systemctl enable docker

# 2. 安装Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 3. 验证安装
docker --version
docker-compose --version

# 4. 克隆项目
cd /opt
git clone https://github.com/zhihang9978/im-suite.git
cd im-suite

# 5. 创建环境变量文件
cat > .env.production <<'EOF'
# Docker Compose 配置
MYSQL_ROOT_PASSWORD=zhihang_im_2024_secure_password
MYSQL_DATABASE=zhihang_messenger
MYSQL_USER=zhihang
MYSQL_PASSWORD=zhihang_im_2024_secure_password

REDIS_PASSWORD=zhihang_redis_2024_secure_password

MINIO_ROOT_USER=zhihang_minio_admin
MINIO_ROOT_PASSWORD=zhihang_minio_2024_secure_key

JWT_SECRET=zhihang_jwt_super_secret_key_2024_production

ADMIN_API_BASE_URL=http://backend:8080
WEB_API_BASE_URL=http://backend:8080
WEB_WS_BASE_URL=ws://backend:8080/ws

WEBRTC_ICE_SERVERS=[{"urls":"stun:stun.l.google.com:19302"}]

GRAFANA_PASSWORD=zhihang_grafana_admin_2024

# 后端应用配置
DB_HOST=mysql
DB_PORT=3306
DB_NAME=zhihang_messenger
DB_USER=zhihang

REDIS_HOST=redis
REDIS_PORT=6379

MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=zhihang_minio_admin
MINIO_SECRET_KEY=zhihang_minio_2024_secure_key
MINIO_USE_SSL=false

JWT_EXPIRES_IN=24h
PORT=8080
GIN_MODE=release
LOG_LEVEL=info

MAX_FILE_SIZE=100MB
UPLOAD_PATH=/app/uploads

DOMAIN=154.37.214.191
EOF

# 6. 生成SSL证书（自签名）
mkdir -p ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/key.pem \
  -out ssl/cert.pem \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=ZhiHang/OU=IT/CN=154.37.214.191"

chmod 600 ssl/key.pem
chmod 644 ssl/cert.pem

# 7. 创建数据目录
mkdir -p data/{mysql,redis,minio,prometheus,grafana,logs}
chmod -R 777 data

# 8. 配置防火墙
ufw allow 22/tcp      # SSH
ufw allow 80/tcp      # HTTP
ufw allow 443/tcp     # HTTPS
ufw allow 3000/tcp    # Grafana
ufw allow 3001/tcp    # 管理后台
ufw allow 3002/tcp    # Web客户端
ufw allow 8080/tcp    # 后端API
ufw allow 9000/tcp    # MinIO
ufw allow 9001/tcp    # MinIO控制台
ufw allow 9090/tcp    # Prometheus
ufw --force enable

# 9. 启动所有服务
docker-compose -f docker-compose.production.yml up -d

# 10. 等待服务启动
echo "等待服务启动（约2分钟）..."
sleep 120

# 11. 查看服务状态
docker-compose -f docker-compose.production.yml ps

# 12. 查看服务日志
docker-compose -f docker-compose.production.yml logs -f
```

---

## 🔍 第三步：验证部署

```bash
# 1. 检查所有容器状态
docker ps

# 应该看到以下容器运行：
# - im-mysql-prod
# - im-redis-prod
# - im-minio-prod
# - im-backend-prod
# - im-admin-prod
# - im-web-prod
# - im-nginx-prod
# - im-prometheus-prod
# - im-grafana-prod
# - im-filebeat-prod

# 2. 测试后端健康检查
curl http://localhost:8080/health

# 预期输出：
# {"status":"ok","timestamp":...,"service":"zhihang-messenger-backend","version":"1.3.1"}

# 3. 测试Web客户端
curl http://localhost:3002

# 4. 测试管理后台
curl http://localhost:3001

# 5. 测试Grafana
curl http://localhost:3000

# 6. 检查数据库
docker-compose -f docker-compose.production.yml exec mysql mysql -uroot -pzhihang_im_2024_secure_password -e "SHOW DATABASES;"

# 7. 检查Redis
docker-compose -f docker-compose.production.yml exec redis redis-cli -a zhihang_redis_2024_secure_password PING
```

---

## 📊 访问地址

部署成功后，可以通过以下地址访问：

| 服务 | 地址 | 说明 |
|------|------|------|
| **后端API** | http://154.37.214.191:8080 | REST API服务 |
| **健康检查** | http://154.37.214.191:8080/health | 服务状态 |
| **Web客户端** | http://154.37.214.191:3002 | IM聊天界面 |
| **管理后台** | http://154.37.214.191:3001 | 系统管理 |
| **Grafana监控** | http://154.37.214.191:3000 | 监控面板 |
| **Prometheus** | http://154.37.214.191:9090 | 监控数据 |
| **MinIO控制台** | http://154.37.214.191:9001 | 对象存储 |
| **Nginx** | http://154.37.214.191:80 | 负载均衡 |

---

## 🔧 常用管理命令

### 查看服务状态
```bash
cd /opt/im-suite
docker-compose -f docker-compose.production.yml ps
```

### 查看服务日志
```bash
# 查看所有服务日志
docker-compose -f docker-compose.production.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.production.yml logs -f backend
docker-compose -f docker-compose.production.yml logs -f mysql
docker-compose -f docker-compose.production.yml logs -f redis
```

### 重启服务
```bash
# 重启所有服务
docker-compose -f docker-compose.production.yml restart

# 重启特定服务
docker-compose -f docker-compose.production.yml restart backend
```

### 停止服务
```bash
docker-compose -f docker-compose.production.yml down
```

### 更新代码
```bash
cd /opt/im-suite
git pull origin main
docker-compose -f docker-compose.production.yml build
docker-compose -f docker-compose.production.yml up -d
```

---

## 🐛 故障排查

### 服务无法启动
```bash
# 查看详细日志
docker-compose -f docker-compose.production.yml logs backend

# 检查端口占用
netstat -tlnp | grep :8080

# 重新构建镜像
docker-compose -f docker-compose.production.yml build --no-cache backend
docker-compose -f docker-compose.production.yml up -d
```

### 无法访问服务
```bash
# 检查防火墙
ufw status

# 检查容器状态
docker ps

# 检查容器日志
docker logs im-backend-prod
```

### 数据库连接失败
```bash
# 进入MySQL容器
docker-compose -f docker-compose.production.yml exec mysql bash

# 连接数据库
mysql -uroot -pzhihang_im_2024_secure_password

# 查看数据库
SHOW DATABASES;
USE zhihang_messenger;
SHOW TABLES;
```

---

## 📞 需要告诉新对话的完整内容

### 复制以下内容发送给新对话：

```
我需要在我的Linux服务器上完整部署志航密信IM系统。

服务器信息：
- IP: 154.37.214.191
- 用户: root
- 操作系统: Linux

请帮我执行以下操作：

1. 清除服务器上的所有旧Docker数据：
   - 停止并删除所有容器
   - 删除所有镜像
   - 清理所有数据卷
   - 删除 /opt/im-suite 目录

2. 完整部署志航密信系统：
   - 检查并安装Docker和Docker Compose
   - 克隆项目：https://github.com/zhihang9978/im-suite.git
   - 进入 /opt/im-suite 目录
   - 执行 ./server-deploy.sh 一键部署
   - 或者手动执行docker-compose -f docker-compose.production.yml up -d

3. 验证部署：
   - 检查所有容器是否运行
   - 测试 http://154.37.214.191:8080/health
   - 确认所有服务正常

4. 开放必要的防火墙端口：
   - 80, 443 (HTTP/HTTPS)
   - 3000 (Grafana)
   - 3001 (管理后台)
   - 3002 (Web客户端)
   - 8080 (后端API)
   - 9000, 9001 (MinIO)
   - 9090 (Prometheus)

请开始执行，并告诉我每一步的执行结果。
```

---

## 📝 补充说明

### 如果新对话需要更详细的信息，告诉它：

```
项目仓库: https://github.com/zhihang9978/im-suite
分支: main
最新提交: 已包含所有功能，100%完整

项目包含：
- Go后端服务（21个服务，108个API）
- Vue3管理后台
- AngularJS Web客户端
- MySQL + Redis + MinIO
- Prometheus + Grafana监控
- Nginx负载均衡

部署方式：
推荐使用项目根目录的 server-deploy.sh 脚本一键部署。

该脚本会自动：
1. 检测并安装Docker和Docker Compose
2. 克隆代码到/opt/im-suite
3. 生成环境变量文件(.env.production)
4. 生成自签名SSL证书
5. 创建数据目录
6. 启动所有Docker服务
7. 显示访问地址

如果自动脚本失败，可以手动使用：
docker-compose -f docker-compose.production.yml up -d
```

---

## ⚠️ 重要提醒

### 告诉新对话注意事项：

```
重要提醒：

1. 清除数据会删除所有现有的数据库、Redis缓存、文件存储
   确认无重要数据后再执行

2. 部署后需要修改的配置：
   - 编辑 /opt/im-suite/.env.production
   - 修改所有默认密码
   - 配置真实域名（如果有）

3. 服务启动需要约2-3分钟
   - MySQL需要初始化数据库
   - 后端需要编译Go代码
   - 前端需要准备静态文件

4. 默认端口分配：
   - 80/443: Nginx
   - 3000: Grafana
   - 3001: 管理后台
   - 3002: Web客户端
   - 8080: 后端API
   - 9000/9001: MinIO
   - 9090: Prometheus

5. 首次访问：
   - 管理后台: http://154.37.214.191:3001
   - Web客户端: http://154.37.214.191:3002
   - 后端API: http://154.37.214.191:8080/health
```

---

## 🎯 执行顺序

告诉新对话按以下顺序执行：

```
第1步：清理环境（5分钟）
→ 停止所有Docker容器
→ 删除所有Docker数据
→ 删除旧项目目录
→ 验证清理完成

第2步：检查环境（2分钟）
→ 检查Docker是否安装
→ 检查Docker Compose是否安装
→ 检查网络连接

第3步：克隆项目（3分钟）
→ 创建/opt/im-suite目录
→ git clone项目
→ 进入项目目录

第4步：执行部署（10分钟）
→ 运行server-deploy.sh
→ 或手动docker-compose up -d
→ 等待服务启动

第5步：验证部署（5分钟）
→ 检查容器状态
→ 测试各个服务
→ 查看日志
→ 确认正常运行

总耗时：约25分钟
```

---

## 📋 验证清单

部署完成后，让新对话帮您验证：

```bash
# 1. 容器状态检查
docker ps | grep -E "im-mysql|im-redis|im-backend|im-admin|im-web|im-nginx"

# 应该看到所有容器都是 "Up" 状态

# 2. 健康检查
curl http://154.37.214.191:8080/health
# 应该返回: {"status":"ok",...}

# 3. 数据库检查
docker exec im-mysql-prod mysql -uroot -pzhihang_im_2024_secure_password -e "SHOW DATABASES;"
# 应该看到: zhihang_messenger

# 4. Redis检查
docker exec im-redis-prod redis-cli -a zhihang_redis_2024_secure_password PING
# 应该返回: PONG

# 5. 服务日志检查
docker-compose -f docker-compose.production.yml logs --tail=50
# 不应该有ERROR级别的日志

# 6. 访问测试
curl -I http://154.37.214.191:3001  # 管理后台
curl -I http://154.37.214.191:3002  # Web客户端
# 都应该返回 200 OK
```

---

## 🎉 完成标志

当新对话告诉您看到以下信息时，部署就成功了：

```
✅ 所有Docker容器都在运行
✅ curl http://154.37.214.191:8080/health 返回正常
✅ 可以访问 http://154.37.214.191:3001（管理后台）
✅ 可以访问 http://154.37.214.191:3002（Web客户端）
✅ 没有错误日志

显示类似信息：
服务访问地址:
  - 后端API: http://154.37.214.191:8080
  - Web客户端: http://154.37.214.191:3002
  - 管理后台: http://154.37.214.191:3001
  - Grafana监控: http://154.37.214.191:3000
  - Prometheus: http://154.37.214.191:9090
  - Nginx负载均衡: http://154.37.214.191:80
```

---

**准备好后，直接复制上面"需要告诉新对话的完整内容"发送给新的SSH会话！** 🚀

