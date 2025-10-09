# 🚀 Devin部署指南 v1.6.0

**版本**: v1.6.0  
**服务器**: 154.37.214.191  
**状态**: 🔴 网络问题待修复  
**更新日期**: 2024-12-19

---

## 🚨 当前状态

### 已知问题

❌ **服务器网络连接问题**:
- DNS解析失败
- Docker Hub无法访问
- 镜像拉取全部失败

### 需要先解决

⚠️ **必须先修复网络问题才能部署**

详细修复方案见: `NETWORK_TROUBLESHOOTING_GUIDE.md`

---

## 📋 部署前准备

### 1. 修复网络问题（必须）

**快速修复**（15分钟）:

```bash
# SSH连接到服务器
ssh root@154.37.214.191

# 执行一键修复
curl -O https://raw.githubusercontent.com/zhihang9978/im-suite/main/scripts/fix-docker-network.sh
chmod +x fix-docker-network.sh
sudo ./fix-docker-network.sh

# 验证修复
docker pull alpine:latest
```

**如果失败**，必须：
1. 检查云服务器安全组（出站规则）
2. 确保允许: HTTPS(443), HTTP(80), DNS(53)

---

## 🎯 部署方案选择

根据网络修复结果，选择合适的部署方案：

### 方案A: Docker Compose部署（推荐，需要网络修复）

**前提**: 网络问题已修复，能拉取Docker镜像

**时间**: 15-30分钟

**步骤**: 见下方"Docker部署流程"

---

### 方案B: 手动上传镜像部署（备用，网络问题未解决）

**前提**: 网络无法修复，但可以SSH上传文件

**时间**: 60-90分钟

**步骤**:

#### B.1 在本地准备镜像（本地机器执行）

```bash
# 克隆项目
git clone https://github.com/zhihang9978/im-suite.git
cd im-suite

# 拉取基础镜像
docker pull mysql:8.0
docker pull redis:7-alpine
docker pull nginx:alpine
docker pull minio/minio:latest
docker pull golang:1.21-alpine
docker pull node:18-alpine

# 构建项目镜像
cd im-backend
docker build -t zhihang-backend:v1.6.0 -f Dockerfile.production .

cd ../im-admin
docker build -t zhihang-admin:v1.6.0 -f Dockerfile.production .

cd ../telegram-web
docker build -t zhihang-web:v1.6.0 -f Dockerfile.production .

# 保存所有镜像
cd ..
mkdir docker-images
docker save mysql:8.0 -o docker-images/mysql.tar
docker save redis:7-alpine -o docker-images/redis.tar
docker save nginx:alpine -o docker-images/nginx.tar
docker save minio/minio:latest -o docker-images/minio.tar
docker save zhihang-backend:v1.6.0 -o docker-images/backend.tar
docker save zhihang-admin:v1.6.0 -o docker-images/admin.tar
docker save zhihang-web:v1.6.0 -o docker-images/web.tar

# 打包
tar czf docker-images-v1.6.0.tar.gz docker-images/

# 查看大小
ls -lh docker-images-v1.6.0.tar.gz
# 预计大小: 2-3GB
```

#### B.2 上传到服务器

```bash
# 使用scp上传（可能需要较长时间）
scp docker-images-v1.6.0.tar.gz root@154.37.214.191:/tmp/

# 或使用rsync（支持断点续传）
rsync -avz --progress docker-images-v1.6.0.tar.gz root@154.37.214.191:/tmp/
```

#### B.3 在服务器加载镜像

```bash
# SSH连接
ssh root@154.37.214.191

# 解压
cd /tmp
tar xzf docker-images-v1.6.0.tar.gz

# 加载所有镜像
cd docker-images
for img in *.tar; do
    echo "加载 $img..."
    docker load -i $img
done

# 验证
docker images

# 清理
cd /tmp
rm -rf docker-images docker-images-v1.6.0.tar.gz
```

#### B.4 部署服务

```bash
# 克隆项目代码（如果还没有）
git clone https://github.com/zhihang9978/im-suite.git
cd im-suite

# 配置环境变量
cp .env.example .env
vi .env
# 修改必要的配置

# 启动服务
docker-compose up -d
```

---

### 方案C: 二进制部署（最后手段，网络和Docker都有问题）

**前提**: 无法使用Docker

**步骤**: 

#### C.1 安装依赖

```bash
# 安装MySQL
sudo apt-get update
sudo apt-get install -y mysql-server

# 安装Redis
sudo apt-get install -y redis-server

# 安装Nginx
sudo apt-get install -y nginx

# 安装Go（用于编译后端）
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装Node.js（用于编译前端）
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

#### C.2 部署后端

```bash
cd im-suite/im-backend

# 编译
go mod download
go build -o im-backend main.go

# 配置
cp .env.example .env
vi .env

# 运行
nohup ./im-backend > backend.log 2>&1 &
```

#### C.3 部署前端

```bash
# 管理后台
cd ../im-admin
npm install
npm run build
sudo cp -r dist/* /var/www/html/admin/

# Web端
cd ../telegram-web
npm install
npm run build
sudo cp -r dist/* /var/www/html/web/
```

#### C.4 配置Nginx

```bash
sudo tee /etc/nginx/sites-available/im-suite > /dev/null <<'EOF'
server {
    listen 80;
    server_name 154.37.214.191;

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /admin {
        alias /var/www/html/admin;
        try_files $uri $uri/ /admin/index.html;
    }

    location / {
        root /var/www/html/web;
        try_files $uri $uri/ /index.html;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/im-suite /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## 🔄 Docker部署流程（网络修复后）

### 完整部署步骤

#### 步骤1: 修复网络（必须）

```bash
ssh root@154.37.214.191

# 执行修复脚本
curl -O https://raw.githubusercontent.com/zhihang9978/im-suite/main/scripts/fix-docker-network.sh
chmod +x fix-docker-network.sh
sudo ./fix-docker-network.sh

# 验证
docker pull alpine:latest
```

#### 步骤2: 克隆项目

```bash
# 如果还没有克隆
git clone https://github.com/zhihang9978/im-suite.git
cd im-suite

# 如果已经克隆，拉取最新代码
cd im-suite
git pull origin main

# 检查版本
git log --oneline -1
# 应该显示: 9520eff docs: add v1.6.0 final summary report
```

#### 步骤3: 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置
vi .env

# 必须修改的配置:
DB_PASSWORD=your_secure_password_here
REDIS_PASSWORD=your_redis_password_here
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_minio_password_here
JWT_SECRET=your_jwt_secret_key_here
```

#### 步骤4: 初始化数据库

```bash
# 检查数据库初始化脚本
cat scripts/init.sql

# 如果需要修改，编辑:
vi scripts/init.sql
```

#### 步骤5: 启动服务

```bash
# 拉取镜像（这一步会比较慢）
docker-compose pull

# 构建自定义镜像
docker-compose build

# 启动所有服务
docker-compose up -d

# 查看启动日志
docker-compose logs -f
```

#### 步骤6: 验证部署

```bash
# 检查服务状态
docker-compose ps

# 应该看到所有服务都是 Up 状态:
# mysql       Up
# redis       Up
# minio       Up
# nginx       Up
# backend     Up
# admin       Up
# web         Up

# 检查后端健康
curl http://localhost:8080/api/health

# 检查前端
curl -I http://localhost:8081
curl -I http://localhost
```

#### 步骤7: 访问测试

```bash
# 在本地浏览器访问:
http://154.37.214.191          # Web端
http://154.37.214.191:8081     # 管理后台
http://154.37.214.191:8080/api/health  # 后端健康检查

# 默认账号测试:
# 超级管理员: admin / Admin123!
# 普通用户: testuser / Test123!
```

---

## 🎯 测试清单

### 基础功能测试

- [ ] 用户注册登录
- [ ] 发送接收消息
- [ ] 文件上传下载
- [ ] 群组创建和管理
- [ ] 管理后台登录

### v1.4.0 功能测试

- [ ] 2FA启用和验证
- [ ] 设备管理功能
- [ ] 超级管理员后台
- [ ] 内容审核功能

### v1.6.0 新功能测试（重点）

#### 机器人系统测试

**后台管理测试**:
- [ ] 登录管理后台
- [ ] 进入系统管理页面
- [ ] 切换到"🤖 机器人管理"标签
- [ ] 创建机器人（保存API密钥）
- [ ] 查看机器人列表和统计
- [ ] 切换机器人状态
- [ ] 查看机器人详情

**机器人用户测试**:
- [ ] 切换到"👤 机器人用户"标签
- [ ] 创建机器人用户（用户名: testbot）
- [ ] 查看机器人用户列表
- [ ] 验证用户创建成功

**用户授权测试**:
- [ ] 切换到"🔑 用户授权"标签
- [ ] 授权testuser使用机器人
- [ ] 查看授权列表
- [ ] 设置过期时间测试

**聊天交互测试**:
- [ ] 使用testuser登录聊天应用
- [ ] 搜索机器人: testbot
- [ ] 开始对话
- [ ] 测试 `/help` 命令
- [ ] 测试 `/create` 命令:
  ```
  /create phone=13800138001 username=demo1 password=Demo123! nickname=演示1
  ```
- [ ] 测试 `/list` 命令
- [ ] 测试 `/info` 命令:
  ```
  /info user_id=101
  ```
- [ ] 测试 `/delete` 命令:
  ```
  /delete user_id=101 reason=测试完成
  ```

**API调用测试**:
- [ ] 使用保存的API密钥
- [ ] 测试创建用户API
- [ ] 测试删除用户API
- [ ] 验证速率限制

**安全测试**:
- [ ] 未授权用户尝试使用机器人（应该失败）
- [ ] 尝试创建管理员（应该自动为user）
- [ ] 尝试删除其他用户（应该失败）
- [ ] 验证权限过期功能

---

## 📦 Docker镜像清单

### 需要拉取的基础镜像

```
mysql:8.0                    # 数据库
redis:7-alpine              # 缓存
nginx:alpine                # Web服务器
minio/minio:latest          # 对象存储
golang:1.21-alpine          # 后端构建
node:18-alpine              # 前端构建
```

### 需要构建的项目镜像

```
zhihang-backend:v1.6.0      # 后端服务
zhihang-admin:v1.6.0        # 管理后台
zhihang-web:v1.6.0          # Web端
```

### 镜像大小预估

| 镜像 | 大小 |
|------|------|
| mysql:8.0 | ~500MB |
| redis:7-alpine | ~30MB |
| nginx:alpine | ~40MB |
| minio/minio | ~200MB |
| zhihang-backend | ~50MB |
| zhihang-admin | ~100MB |
| zhihang-web | ~150MB |
| **总计** | **~1.1GB** |

---

## 🔧 故障排查

### 问题1: Docker Compose启动失败

**检查**:
```bash
# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs backend
docker-compose logs mysql
docker-compose logs redis

# 查看资源使用
docker stats
```

**常见原因**:
- 端口被占用
- 内存不足
- 配置文件错误
- 依赖服务未启动

---

### 问题2: 后端无法连接数据库

**检查**:
```bash
# 测试MySQL连接
docker-compose exec mysql mysql -uroot -p

# 检查数据库是否初始化
docker-compose exec mysql mysql -uroot -p -e "SHOW DATABASES;"

# 检查后端日志
docker-compose logs backend | grep -i mysql
```

**修复**:
```bash
# 重新初始化数据库
docker-compose down
docker volume rm im-suite_mysql-data
docker-compose up -d mysql
# 等待30秒让MySQL初始化完成
sleep 30
docker-compose up -d
```

---

### 问题3: 前端无法访问

**检查**:
```bash
# 检查Nginx配置
docker-compose exec nginx nginx -t

# 查看Nginx日志
docker-compose logs nginx

# 检查端口监听
netstat -tlnp | grep :80
netstat -tlnp | grep :8081
```

**修复**:
```bash
# 重启Nginx
docker-compose restart nginx

# 查看配置文件
cat config/nginx/nginx.conf
```

---

### 问题4: 机器人不响应

**检查**:
```bash
# 1. 检查机器人是否创建
# 管理后台 → 系统管理 → 🤖 机器人管理

# 2. 检查机器人用户是否创建
# 管理后台 → 系统管理 → 👤 机器人用户

# 3. 检查用户是否授权
# 管理后台 → 系统管理 → 🔑 用户授权

# 4. 查看后端日志
docker-compose logs backend | grep -i bot

# 5. 检查数据库
docker-compose exec mysql mysql -uroot -p -e "USE zhihang_messenger; SELECT * FROM bots;"
docker-compose exec mysql mysql -uroot -p -e "USE zhihang_messenger; SELECT * FROM bot_users;"
docker-compose exec mysql mysql -uroot -p -e "USE zhihang_messenger; SELECT * FROM bot_user_permissions WHERE user_id=2;"
```

---

## 📊 部署验证清单

### 基础服务验证

```bash
# MySQL
docker-compose exec mysql mysql -uroot -p -e "SELECT 1;"

# Redis
docker-compose exec redis redis-cli ping

# MinIO
curl -I http://localhost:9000

# Nginx
curl -I http://localhost

# 后端
curl http://localhost:8080/api/health
```

### 功能验证

```bash
# 1. 用户注册
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "username": "testuser2",
    "password": "Test123!"
  }'

# 2. 用户登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123!"
  }'

# 3. 访问前端
curl -I http://localhost:8081  # 管理后台
curl -I http://localhost        # Web端
```

---

## 🎯 机器人功能完整测试流程

### 流程图

```
1. 修复网络问题 ✅
   ↓
2. 部署系统 ✅
   ↓
3. 登录管理后台（admin）
   ↓
4. 系统管理 → 🤖 机器人管理
   ↓
5. 创建机器人（保存API密钥）
   ↓
6. 系统管理 → 👤 机器人用户
   ↓
7. 创建机器人用户（用户名: testbot）
   ↓
8. 系统管理 → 🔑 用户授权
   ↓
9. 授权testuser使用机器人
   ↓
10. 登出，使用testuser登录
   ↓
11. 搜索"testbot"，开始对话
   ↓
12. 测试所有命令:
    - /help
    - /create phone=... username=... password=...
    - /list
    - /info user_id=...
    - /delete user_id=... reason=...
   ↓
13. 验证权限限制:
    - 创建的用户应该是role=user
    - 只能删除自己创建的用户
    - 未授权用户无法使用
   ↓
14. 验证API调用:
    - 使用API密钥调用
    - 测试速率限制
   ↓
15. 查看日志和统计:
    - 管理后台查看调用统计
    - 查看操作日志
   ↓
16. ✅ 测试完成
```

---

## 📝 测试报告模板

```markdown
# v1.6.0 部署测试报告

**测试人员**: Devin
**测试日期**: 2024-12-XX
**服务器**: 154.37.214.191

## 1. 网络修复

### 执行的方案:
- [ ] 方案1: DNS修复
- [ ] 方案2: Docker镜像源
- [ ] 方案3: 安全组配置
- [ ] 方案5: 手动上传镜像

### 结果:
- 网络问题: [ ] 已修复 / [ ] 未修复
- Docker拉取: [ ] 成功 / [ ] 失败

## 2. 系统部署

### 部署方式:
- [ ] Docker Compose
- [ ] 手动上传镜像
- [ ] 二进制部署

### 服务状态:
- MySQL: [ ] 运行中 / [ ] 失败
- Redis: [ ] 运行中 / [ ] 失败
- Nginx: [ ] 运行中 / [ ] 失败
- Backend: [ ] 运行中 / [ ] 失败
- Admin: [ ] 运行中 / [ ] 失败
- Web: [ ] 运行中 / [ ] 失败

## 3. 机器人功能测试

### 后台管理:
- 创建机器人: [ ] 成功 / [ ] 失败
- 创建机器人用户: [ ] 成功 / [ ] 失败
- 授权用户: [ ] 成功 / [ ] 失败

### 聊天交互:
- /help: [ ] 成功 / [ ] 失败
- /create: [ ] 成功 / [ ] 失败
- /list: [ ] 成功 / [ ] 失败
- /info: [ ] 成功 / [ ] 失败
- /delete: [ ] 成功 / [ ] 失败

### API调用:
- 创建用户: [ ] 成功 / [ ] 失败
- 删除用户: [ ] 成功 / [ ] 失败
- 速率限制: [ ] 验证通过 / [ ] 未验证

### 权限测试:
- 只能创建普通用户: [ ] 验证通过 / [ ] 失败
- 只能删除自己创建的: [ ] 验证通过 / [ ] 失败
- 未授权用户无法使用: [ ] 验证通过 / [ ] 失败

## 4. 遇到的问题

### 问题列表:
1. 
2. 
3. 

### 解决方案:
1. 
2. 
3. 

## 5. 性能测试

- 响应时间: 
- 并发测试: 
- 内存使用: 
- CPU使用: 

## 6. 总体评价

- 部署难度: [ ] 简单 / [ ] 中等 / [ ] 困难
- 文档完整性: [ ] 完整 / [ ] 缺失部分
- 功能完整性: [ ] 100% / [ ] 部分完成
- 推荐使用: [ ] 是 / [ ] 否

## 7. 建议和改进

-
-
-
```

---

## 🆘 紧急联系

### 如果遇到无法解决的问题

**GitHub Issues**:
https://github.com/zhihang9978/im-suite/issues

**提供信息**:
1. 服务器IP和配置
2. 错误日志
3. 已尝试的修复方案
4. 系统环境信息

---

## 📚 相关文档

- **网络故障排查**: `NETWORK_TROUBLESHOOTING_GUIDE.md`
- **快速开始**: `QUICK_START_V1.6.0.md`
- **部署说明**: `SERVER_DEPLOYMENT_INSTRUCTIONS.md`
- **机器人测试**: `docs/BOT_CHAT_GUIDE.md`

---

**重要提示**: 

1. ⚠️ **必须先解决网络问题**才能继续部署
2. 📞 **建议先检查云服务器安全组**（最常见原因）
3. 🔧 **使用提供的修复脚本**快速解决
4. 📦 **如果网络无法修复**，使用方案B手动上传镜像

---

**Devin加油！网络问题解决后部署会很顺利！** 💪

