# 志航密信部署指南

## 概述

志航密信是一个基于 Telegram 前端改造的私有通讯系统，支持 Web 端和 Android 端。本文档提供详细的部署指南。

## 系统要求

### 硬件要求
- **CPU**: 2 核心以上
- **内存**: 4GB 以上
- **存储**: 20GB 以上可用空间
- **网络**: 稳定的互联网连接

### 软件要求
- **操作系统**: Linux (Ubuntu 20.04+), macOS, Windows 10+
- **Docker**: 20.10+
- **Docker Compose**: 2.0+

## 快速部署

### 1. 克隆项目
```bash
git clone https://github.com/your-org/zhihang-messenger.git
cd zhihang-messenger
```

### 2. 一键部署
```bash
# 给脚本执行权限
chmod +x scripts/deploy.sh

# 执行部署
./scripts/deploy.sh
```

### 3. 访问系统
- **Web 端**: https://localhost
- **API 文档**: https://localhost/api/ping
- **MinIO 控制台**: http://localhost:9001

## 详细部署步骤

### 1. 环境准备

#### 安装 Docker
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 验证安装
```bash
docker --version
docker-compose --version
```

### 2. 配置系统

#### 环境变量配置
创建 `.env` 文件：
```bash
# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_NAME=zhihang_messenger
DB_USER=zhihang_messenger
DB_PASSWORD=zhihang_messenger_pass

# Redis 配置
REDIS_HOST=redis
REDIS_PORT=6379

# MinIO 配置
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=zhihang_messenger
MINIO_SECRET_KEY=zhihang_messenger_pass

# JWT 配置
JWT_SECRET=zhihang_messenger_secret_key_2024
JWT_EXPIRE_HOURS=24
```

### 3. 启动服务

#### 启动所有服务
```bash
docker-compose up -d
```

#### 检查服务状态
```bash
docker-compose ps
```

#### 查看日志
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f web
```

### 4. 验证部署

#### 检查后端服务
```bash
curl http://localhost:8080/api/ping
```

#### 检查 Web 服务
```bash
curl http://localhost:3000
```

#### 检查数据库连接
```bash
docker-compose exec mysql mysql -u zhihang_messenger -pzhihang_messenger_pass -e "SHOW DATABASES;"
```

## 服务管理

### 启动服务
```bash
docker-compose up -d
```

### 停止服务
```bash
# 停止服务
docker-compose down

# 停止服务并删除数据卷（危险操作）
docker-compose down -v
```

### 重启服务
```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart backend
```

### 更新服务
```bash
# 拉取最新镜像
docker-compose pull

# 重新构建并启动
docker-compose up -d --build
```

## 数据管理

### 备份数据
```bash
# 备份数据库
docker-compose exec mysql mysqldump -u zhihang_messenger -pzhihang_messenger_pass zhihang_messenger > backup.sql

# 备份 Redis 数据
docker-compose exec redis redis-cli BGSAVE
docker cp $(docker-compose ps -q redis):/data/dump.rdb ./redis_backup.rdb

# 备份 MinIO 数据
docker cp $(docker-compose ps -q minio):/data ./minio_backup
```

### 恢复数据
```bash
# 恢复数据库
docker-compose exec -T mysql mysql -u zhihang_messenger -pzhihang_messenger_pass zhihang_messenger < backup.sql

# 恢复 Redis 数据
docker cp ./redis_backup.rdb $(docker-compose ps -q redis):/data/dump.rdb
docker-compose restart redis

# 恢复 MinIO 数据
docker cp ./minio_backup/. $(docker-compose ps -q minio):/data/
```

## 监控和维护

### 查看服务状态
```bash
docker-compose ps
```

### 查看资源使用情况
```bash
docker stats
```

### 查看日志
```bash
# 实时查看日志
docker-compose logs -f

# 查看特定服务的日志
docker-compose logs -f backend
```

### 清理资源
```bash
# 清理未使用的镜像
docker image prune -f

# 清理未使用的网络
docker network prune -f

# 清理所有未使用的资源
docker system prune -f
```

## 故障排除

### 常见问题

#### 1. 端口冲突
```bash
# 检查端口占用
netstat -tulpn | grep :8080
netstat -tulpn | grep :3306

# 修改端口配置
# 编辑 docker-compose.yml 文件
```

#### 2. 数据库连接失败
```bash
# 检查数据库服务状态
docker-compose logs mysql

# 检查数据库连接
docker-compose exec mysql mysql -u root -p
```

#### 3. 内存不足
```bash
# 检查系统内存
free -h

# 检查 Docker 内存使用
docker stats
```

#### 4. 磁盘空间不足
```bash
# 检查磁盘使用情况
df -h

# 清理 Docker 资源
docker system prune -f
```

### 日志分析

#### 后端日志
```bash
# 查看后端错误日志
docker-compose logs backend | grep ERROR

# 查看后端访问日志
docker-compose logs backend | grep "GET\|POST"
```

#### 数据库日志
```bash
# 查看数据库日志
docker-compose logs mysql

# 查看慢查询日志
docker-compose exec mysql mysql -u root -p -e "SHOW VARIABLES LIKE 'slow_query_log';"
```

## 安全配置

### SSL 证书配置
```bash
# 生成自签名证书
openssl req -x509 -newkey rsa:4096 -keyout scripts/ssl/key.pem -out scripts/ssl/cert.pem -days 365 -nodes

# 配置 Nginx SSL
# 编辑 scripts/nginx/conf.d/zhihang-messenger.conf
```

### 防火墙配置
```bash
# 开放必要端口
sudo ufw allow 80
sudo ufw allow 443
sudo ufw allow 22

# 启用防火墙
sudo ufw enable
```

### 数据库安全
```bash
# 修改默认密码
# 编辑 docker-compose.yml 文件中的数据库密码

# 限制数据库访问
# 编辑 scripts/nginx/conf.d/zhihang-messenger.conf
```

## 性能优化

### 数据库优化
```bash
# 调整 MySQL 配置
# 编辑 docker-compose.yml 中的 MySQL 环境变量
```

### Redis 优化
```bash
# 调整 Redis 配置
# 编辑 docker-compose.yml 中的 Redis 配置
```

### Nginx 优化
```bash
# 调整 Nginx 配置
# 编辑 scripts/nginx/nginx.conf
```

## 扩展部署

### 水平扩展
```bash
# 扩展后端服务
docker-compose up -d --scale backend=3

# 配置负载均衡
# 编辑 scripts/nginx/conf.d/zhihang-messenger.conf
```

### 集群部署
```bash
# 使用 Docker Swarm
docker swarm init
docker stack deploy -c docker-compose.yml zhihang-messenger
```

## 联系支持

如果您在部署过程中遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查项目的 Issues 页面
3. 联系技术支持团队

---

**注意**: 本部署指南基于默认配置，生产环境部署时请根据实际需求调整配置参数。
