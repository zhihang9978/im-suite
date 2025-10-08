# 志航密信生产环境部署指南

## 📋 概述

本指南详细介绍了志航密信IM套件的完整生产环境部署流程，包括Docker容器化部署、监控配置、备份策略等。

## 🏗️ 系统架构

### 核心服务
- **后端服务**: Go + Gin + GORM + MySQL + Redis + MinIO
- **管理后台**: Vue 3 + Element Plus + Vite
- **Web客户端**: AngularJS + Gulp
- **负载均衡**: Nginx
- **监控系统**: Prometheus + Grafana
- **日志收集**: Filebeat + ELK Stack

### 服务端口
- 80/443: Nginx负载均衡器
- 8080: 后端API服务
- 3000: Grafana监控面板
- 3001: 管理后台
- 9090: Prometheus监控
- 3306: MySQL数据库
- 6379: Redis缓存
- 9000/9001: MinIO对象存储

## 🚀 快速部署

### 1. 环境准备

```bash
# 安装Docker和Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo usermod -aG docker $USER

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 重启系统或重新登录
sudo reboot
```

### 2. 克隆项目

```bash
# 克隆项目到生产目录
sudo mkdir -p /opt/im-suite
sudo chown $USER:$USER /opt/im-suite
cd /opt/im-suite
git clone https://github.com/zhihang9978/im-suite.git .
```

### 3. 配置环境变量

```bash
# 复制环境变量示例文件
cp env.production.example .env.production

# 编辑生产环境配置
nano .env.production
```

**重要配置项**:
- `MYSQL_ROOT_PASSWORD`: MySQL root密码
- `MYSQL_PASSWORD`: 应用数据库密码
- `REDIS_PASSWORD`: Redis密码
- `JWT_SECRET`: JWT密钥（必须足够复杂）
- `DOMAIN_NAME`: 你的域名
- `ADMIN_API_BASE_URL`: 管理后台API地址
- `WEB_API_BASE_URL`: Web客户端API地址

### 4. 初始化部署环境

```bash
# 给部署脚本执行权限
chmod +x deploy.sh

# 初始化部署环境
./deploy.sh init
```

### 5. 启动服务

```bash
# 启动所有服务
./deploy.sh start

# 查看服务状态
./deploy.sh status
```

## 🔧 详细配置

### SSL证书配置

```bash
# 将SSL证书放置到ssl目录
sudo mkdir -p ssl
sudo cp your-cert.pem ssl/cert.pem
sudo cp your-key.pem ssl/key.pem
sudo chown -R $USER:$USER ssl/

# 更新Nginx配置启用HTTPS
# 取消注释nginx.conf中的HTTPS server块
```

### 域名配置

```bash
# 配置DNS记录
# A记录: yourdomain.com -> 服务器IP
# A记录: admin.yourdomain.com -> 服务器IP

# 更新环境变量
sed -i 's/yourdomain.com/your-actual-domain.com/g' .env.production
```

### 防火墙配置

```bash
# 开放必要端口
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

## 📊 监控配置

### 访问监控面板

- **Grafana**: http://yourdomain.com:3000
  - 用户名: admin
  - 密码: 在.env.production中配置的GRAFANA_PASSWORD

- **Prometheus**: http://yourdomain.com:9090

## 💾 备份策略

### 自动备份

```bash
# 启用自动备份定时器
sudo cp config/systemd/im-suite-backup.service /etc/systemd/system/
sudo cp config/systemd/im-suite-backup.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable im-suite-backup.timer
sudo systemctl start im-suite-backup.timer

# 查看备份状态
sudo systemctl status im-suite-backup.timer
```

### 手动备份

```bash
# 执行完整备份
./deploy.sh backup

# 备份文件位置
ls -la backups/
```

## 🔄 服务管理

### 使用systemd管理

```bash
# 安装systemd服务
sudo cp config/systemd/im-suite.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable im-suite.service
sudo systemctl start im-suite.service

# 服务管理命令
sudo systemctl start im-suite    # 启动服务
sudo systemctl stop im-suite     # 停止服务
sudo systemctl restart im-suite  # 重启服务
sudo systemctl status im-suite   # 查看状态
```

### 使用部署脚本管理

```bash
./deploy.sh start     # 启动服务
./deploy.sh stop      # 停止服务
./deploy.sh restart   # 重启服务
./deploy.sh status    # 查看状态
./deploy.sh update    # 更新服务
```

## 🛠️ 故障排除

### 查看服务日志

```bash
# 查看所有服务日志
docker-compose -f docker-compose.production.yml logs

# 查看特定服务日志
docker-compose -f docker-compose.production.yml logs backend
docker-compose -f docker-compose.production.yml logs mysql
docker-compose -f docker-compose.production.yml logs nginx
```

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   sudo netstat -tlnp | grep :80
   sudo netstat -tlnp | grep :8080
   ```

2. **内存不足**
   ```bash
   # 检查内存使用
   free -h
   docker stats
   ```

3. **磁盘空间不足**
   ```bash
   # 清理Docker资源
   docker system prune -a
   docker volume prune
   ```

## 📈 性能优化

### 数据库优化

```bash
# 编辑MySQL配置
nano config/mysql/conf.d/mysql.cnf

# 重启MySQL服务
docker-compose -f docker-compose.production.yml restart mysql
```

### Nginx优化

```bash
# 编辑Nginx配置
nano config/nginx/nginx.conf

# 重新加载Nginx配置
docker-compose -f docker-compose.production.yml exec nginx nginx -s reload
```

## 🔒 安全配置

### 防火墙规则

```bash
# 配置iptables规则
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
sudo iptables -A INPUT -j DROP
```

### SSL/TLS配置

```bash
# 使用Let's Encrypt免费证书
sudo apt install certbot
sudo certbot certonly --standalone -d yourdomain.com
sudo certbot certonly --standalone -d admin.yourdomain.com
```

## 📞 技术支持

如果遇到部署问题，请提供以下信息：

1. 操作系统版本: `lsb_release -a`
2. Docker版本: `docker --version`
3. 服务状态: `./deploy.sh status`
4. 错误日志: `docker-compose -f docker-compose.production.yml logs`
5. 系统资源: `free -h && df -h`
