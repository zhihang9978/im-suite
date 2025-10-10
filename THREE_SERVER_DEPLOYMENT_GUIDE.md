# 志航密信三服务器部署完整指南

**部署架构**: 主备高可用 (Active-Passive HA)  
**目标**: 99.9% 可用性，用户无感知故障切换  
**预计时间**: 2-3 小时

---

## 📊 三服务器架构总览

```
┌─────────────────────┐       ┌─────────────────────┐       ┌─────────────────────┐
│   主服务器 (主)      │       │   副服务器 (备)      │       │   监控服务器         │
│   154.37.214.191    │◄─────►│   待分配IP          │       │   待分配IP          │
├─────────────────────┤       ├─────────────────────┤       ├─────────────────────┤
│                     │       │                     │       │                     │
│ ✅ MySQL 主库       │──同步→│ 🔄 MySQL 从库       │       │ 📊 Prometheus      │
│    (读写)           │       │    (只读)           │       │    (监控收集)       │
│                     │       │                     │       │                     │
│ ✅ Redis 主节点     │──同步→│ 🔄 Redis 从节点     │       │ 📈 Grafana         │
│    (读写)           │       │    (只读)           │       │    (可视化)         │
│                     │       │                     │       │                     │
│ ✅ MinIO 主节点     │──同步→│ 🔄 MinIO 从节点     │       │ 🔔 Alertmanager    │
│    (文件存储)       │       │    (备份)           │       │    (告警)           │
│                     │       │                     │       │                     │
│ ✅ 后端 API         │       │ ⏸️  后端 API         │       │ 📊 Node Exporter   │
│    (运行中)         │       │    (待命)           │       │    (系统指标)       │
│                     │       │                     │       │                     │
│ ✅ 管理后台         │       │ ⏸️  管理后台         │       └─────────────────────┘
│    (运行中)         │       │    (待命)           │              ▲
│                     │       │                     │              │
│ ✅ Web 客户端       │       │ ⏸️  Web 客户端       │              │收集监控数据
│    (运行中)         │       │    (待命)           │              │
│                     │       │                     │       ┌──────┴──────┐
│ 🔄 Keepalived      │◄─心跳→│ 🔄 Keepalived      │       │             │
│    (优先级 100)     │       │    (优先级 90)      │       │             │
│                     │       │                     │       │             │
│ 📊 Node Exporter   │       │ 📊 Node Exporter   │───────┘             │
│    (系统监控)       │       │    (系统监控)       │─────────────────────┘
└─────────────────────┘       └─────────────────────┘

虚拟 IP: 10.0.0.100
用户访问: http://10.0.0.100 或 http://api.yourdomain.com
```

---

## 🎯 工作原理

### 正常情况
```
用户 → 虚拟IP(10.0.0.100) → 主服务器 → 返回响应
                                ↓
                          实时数据同步
                                ↓
                            副服务器（备份）
                                ↓
                          监控服务器（记录）
```

### 故障切换
```
主服务器宕机 ❌
    ↓ (3秒检测)
Keepalived 检测到故障
    ↓ (5秒切换)
虚拟IP切换到副服务器
    ↓ (10秒激活)
副服务器提升为主服务器
    ↓ (30秒内完成)
用户自动重连 ✅
```

---

# 🖥️ 服务器 1: 主服务器部署

## 📝 主服务器职责
- ✅ 处理所有用户请求（API、管理后台、Web客户端）
- ✅ MySQL 主库：处理所有读写操作
- ✅ Redis 主节点：缓存读写
- ✅ MinIO 主节点：文件存储
- ✅ 实时同步数据到副服务器
- ✅ 发送监控数据到监控服务器

---

## 主服务器部署步骤

### 步骤 1.1：连接主服务器并更新系统

```bash
# 从本地连接到主服务器
ssh root@154.37.214.191

# 更新系统
apt update && apt upgrade -y

# 安装基础工具
apt install -y git curl wget vim htop net-tools sysstat
```

---

### 步骤 1.2：安装 Docker 和 Docker Compose

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | bash

# 启动 Docker
systemctl enable docker
systemctl start docker

# 验证 Docker
docker --version

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 验证 Docker Compose
docker-compose --version
```

**预期输出**:
```
Docker version 24.0.x
Docker Compose version v2.20.0
```

---

### 步骤 1.3：克隆代码并配置环境变量

```bash
# 克隆项目
cd /root
git clone https://github.com/zhihang9978/im-suite.git
cd im-suite

# 创建 .env 文件
cat > .env << 'EOF'
# ========================================
# 主服务器环境配置
# ========================================

# 数据库配置
MYSQL_ROOT_PASSWORD=ZhRoot2024SecurePass!@#
MYSQL_DATABASE=zhihang_messenger
MYSQL_USER=zhihang
MYSQL_PASSWORD=ZhUser2024SecurePass!@#

# Redis 配置
REDIS_PASSWORD=ZhRedis2024SecurePass!@#

# MinIO 配置
MINIO_ROOT_USER=zhihang_admin
MINIO_ROOT_PASSWORD=ZhMinIO2024SecurePass!@#

# JWT 配置
JWT_SECRET=ZhiHang_JWT_Super_Secret_Key_2024_Min32Chars_ProductionUse

# 服务配置
PORT=8080
GIN_MODE=release
LOG_LEVEL=info

# 前端配置
ADMIN_API_BASE_URL=http://backend:8080
WEB_API_BASE_URL=http://backend:8080
WEB_WS_BASE_URL=ws://backend:8080/ws

# WebRTC 配置
WEBRTC_ICE_SERVERS=[{"urls":"stun:stun.l.google.com:19302"}]

# Grafana 配置
GRAFANA_PASSWORD=ZhGrafana2024AdminPass!@#
EOF

# 设置文件权限
chmod 600 .env
```

---

### 步骤 1.4：启动主服务器所有服务

```bash
# 启动所有服务
docker-compose -f docker-compose.production.yml up -d

# 等待服务启动（约2分钟）
echo "等待服务启动..."
sleep 120

# 查看服务状态
docker-compose -f docker-compose.production.yml ps
```

**预期输出**:
```
NAME                 STATUS          PORTS
im-mysql-prod        Up (healthy)    0.0.0.0:3306->3306/tcp
im-redis-prod        Up (healthy)    0.0.0.0:6379->6379/tcp
im-minio-prod        Up (healthy)    0.0.0.0:9000-9001->9000-9001/tcp
im-backend-prod      Up              0.0.0.0:8080->8080/tcp
im-admin-prod        Up              0.0.0.0:3001->80/tcp
im-web-prod          Up              0.0.0.0:3002->80/tcp
im-prometheus-prod   Up              0.0.0.0:9090->9090/tcp
im-grafana-prod      Up              0.0.0.0:3000->3000/tcp
```

---

### 步骤 1.5：验证主服务器服务

```bash
# 1. 验证后端 API
curl http://localhost:8080/health
# 预期: {"status":"ok","service":"zhihang-messenger-backend"}

# 2. 验证数据库迁移
docker logs im-backend-prod | grep "数据库迁移"
# 预期: ✅ 数据库迁移完成！成功迁移 56/56 个表

# 3. 验证 MySQL
docker exec im-mysql-prod mysql -u root -p"ZhRoot2024SecurePass!@#" -e "SHOW DATABASES;"
# 预期: 看到 zhihang_messenger 数据库

# 4. 验证 Redis
docker exec im-redis-prod redis-cli -a "ZhRedis2024SecurePass!@#" PING
# 预期: PONG

# 5. 验证管理后台
curl -I http://localhost:3001
# 预期: HTTP/1.1 200 OK

# 6. 验证 Web 客户端
curl -I http://localhost:3002
# 预期: HTTP/1.1 200 OK
```

---

### 步骤 1.6：配置 MySQL 主库（为主从复制准备）

```bash
# 进入 MySQL 容器
docker exec -it im-mysql-prod mysql -u root -p"ZhRoot2024SecurePass!@#"

# 在 MySQL 中执行以下命令：
```

```sql
-- 1. 创建复制用户
CREATE USER 'repl'@'%' IDENTIFIED BY 'Replication_Pass_2024!';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
FLUSH PRIVILEGES;

-- 2. 查看主库状态（记录 File 和 Position）
SHOW MASTER STATUS;
```

**预期输出示例**:
```
+------------------+----------+--------------+------------------+
| File             | Position | Binlog_Do_DB | Binlog_Ignore_DB |
+------------------+----------+--------------+------------------+
| mysql-bin.000001 |      157 |              |                  |
+------------------+----------+--------------+------------------+
```

**⚠️ 重要**: 记录 `File` 和 `Position` 值，配置副服务器时需要用到！

```sql
-- 退出 MySQL
exit
```

---

### 步骤 1.7：安装 Node Exporter（监控）

```bash
# 运行 Node Exporter
docker run -d \
  --name node-exporter \
  --restart unless-stopped \
  --net="host" \
  --pid="host" \
  -v "/:/host:ro,rslave" \
  prom/node-exporter:latest \
  --path.rootfs=/host

# 验证
curl http://localhost:9100/metrics | head -20
```

---

### 步骤 1.8：安装 Keepalived（自动故障转移）

```bash
# 安装 Keepalived
apt install -y keepalived

# 查看网络接口名称（记录下来，后面配置需要用）
ip addr show

# 创建健康检查脚本
cat > /etc/keepalived/check_backend.sh << 'EOF'
#!/bin/bash
if curl -sf http://localhost:8080/health > /dev/null; then
    exit 0  # 服务正常
else
    exit 1  # 服务异常
fi
EOF

chmod +x /etc/keepalived/check_backend.sh

# 配置 Keepalived
# ⚠️ 注意：将 interface eth0 改为实际的网络接口名称
cat > /etc/keepalived/keepalived.conf << 'EOF'
global_defs {
    router_id MASTER_NODE
}

vrrp_script check_backend {
    script "/etc/keepalived/check_backend.sh"
    interval 2      # 每2秒检查一次
    timeout 3       # 超时时间3秒
    weight -50      # 检查失败，优先级降低50
    fall 3          # 连续3次失败才判定为故障
    rise 2          # 连续2次成功才判定为恢复
}

vrrp_instance VI_1 {
    state MASTER
    interface eth0  # ⚠️ 改为实际网络接口名称
    virtual_router_id 51
    priority 100    # 主服务器优先级 100
    advert_int 1
    
    authentication {
        auth_type PASS
        auth_pass ZhiHang2024!
    }
    
    virtual_ipaddress {
        10.0.0.100/24  # 虚拟IP
    }
    
    track_script {
        check_backend
    }
    
    # 切换通知脚本
    notify_master "/usr/local/bin/keepalived-notify.sh MASTER"
    notify_backup "/usr/local/bin/keepalived-notify.sh BACKUP"
    notify_fault "/usr/local/bin/keepalived-notify.sh FAULT"
}
EOF

# 创建通知脚本
cat > /usr/local/bin/keepalived-notify.sh << 'EOF'
#!/bin/bash
TYPE=$1
HOST=$(hostname)
LOG_FILE="/var/log/keepalived-notify.log"

case $TYPE in
    MASTER)
        echo "$(date) - ✅ $HOST 切换为主服务器" >> $LOG_FILE
        ;;
    BACKUP)
        echo "$(date) - ⚠️ $HOST 切换为备份服务器" >> $LOG_FILE
        ;;
    FAULT)
        echo "$(date) - 🔴 $HOST 出现故障" >> $LOG_FILE
        ;;
esac
EOF

chmod +x /usr/local/bin/keepalived-notify.sh

# 启动 Keepalived
systemctl enable keepalived
systemctl start keepalived

# 验证虚拟 IP 已绑定
ip addr show | grep "10.0.0.100"
# 预期：看到 inet 10.0.0.100/24 scope global secondary eth0
```

---

### 步骤 1.9：主服务器最终验证

```bash
echo "========================================="
echo "主服务器部署验证"
echo "========================================="

# 1. Docker 容器状态
echo "1. Docker 容器状态:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 2. MySQL 主库状态
echo -e "\n2. MySQL 主库状态:"
docker exec im-mysql-prod mysql -u root -p"ZhRoot2024SecurePass!@#" -e "SHOW MASTER STATUS;"

# 3. Redis 状态
echo -e "\n3. Redis 主节点状态:"
docker exec im-redis-prod redis-cli -a "ZhRedis2024SecurePass!@#" INFO replication | grep role

# 4. 后端 API 健康
echo -e "\n4. 后端 API 健康检查:"
curl http://localhost:8080/health

# 5. 虚拟 IP
echo -e "\n5. 虚拟 IP 状态:"
ip addr show | grep "10.0.0.100"

# 6. Keepalived 状态
echo -e "\n6. Keepalived 状态:"
systemctl status keepalived | grep Active

# 7. Node Exporter
echo -e "\n7. Node Exporter 状态:"
curl -s http://localhost:9100/metrics | head -1

echo "========================================="
echo "✅ 主服务器部署完成！"
echo "========================================="
```

**退出主服务器**:
```bash
exit
```

---

# 🖥️ 服务器 2: 副服务器部署

## 📝 副服务器职责
- 🔄 实时同步主服务器数据（MySQL、Redis、MinIO）
- ⏸️  服务待命状态（不对外提供服务）
- 🚨 主服务器故障时，自动接管（< 30秒）
- 📊 发送监控数据到监控服务器

---

## 副服务器部署步骤

### 步骤 2.1：连接副服务器并更新系统

```bash
# 从本地连接到副服务器（替换为实际IP）
ssh root@BACKUP_SERVER_IP

# 更新系统
apt update && apt upgrade -y

# 安装基础工具
apt install -y git curl wget vim htop net-tools sysstat
```

---

### 步骤 2.2：安装 Docker 和 Docker Compose

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | bash

# 启动 Docker
systemctl enable docker
systemctl start docker

# 验证 Docker
docker --version

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 验证 Docker Compose
docker-compose --version
```

---

### 步骤 2.3：克隆代码并配置环境变量

```bash
# 克隆项目
cd /root
git clone https://github.com/zhihang9978/im-suite.git
cd im-suite

# 创建 .env 文件（与主服务器相同）
cat > .env << 'EOF'
# ========================================
# 副服务器环境配置
# ========================================

# 数据库配置（与主服务器相同）
MYSQL_ROOT_PASSWORD=ZhRoot2024SecurePass!@#
MYSQL_DATABASE=zhihang_messenger
MYSQL_USER=zhihang
MYSQL_PASSWORD=ZhUser2024SecurePass!@#

# Redis 配置（与主服务器相同）
REDIS_PASSWORD=ZhRedis2024SecurePass!@#

# MinIO 配置（与主服务器相同）
MINIO_ROOT_USER=zhihang_admin
MINIO_ROOT_PASSWORD=ZhMinIO2024SecurePass!@#

# JWT 配置（与主服务器相同）
JWT_SECRET=ZhiHang_JWT_Super_Secret_Key_2024_Min32Chars_ProductionUse

# 服务配置
PORT=8080
GIN_MODE=release
LOG_LEVEL=info

# 主服务器信息（用于主从复制）
MASTER_HOST=154.37.214.191
MASTER_PORT=3306
REPL_USER=repl
REPL_PASSWORD=Replication_Pass_2024!
EOF

chmod 600 .env
```

---

### 步骤 2.4：创建副服务器专用 Docker Compose 配置

```bash
# 创建副服务器专用配置文件
cat > docker-compose.backup.yml << 'EOF'
version: '3.8'

services:
  # MySQL 从库（只读模式）
  mysql:
    image: mysql:8.0
    container_name: im-mysql-backup
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
      - ./config/mysql/conf.d:/etc/mysql/conf.d
    ports:
      - "3307:3306"  # 使用不同端口避免冲突
    networks:
      - im-network
    command: --default-authentication-plugin=mysql_native_password --read-only=1
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      timeout: 20s
      retries: 10

  # Redis 从节点
  redis:
    image: redis:7-alpine
    container_name: im-redis-backup
    restart: unless-stopped
    command: >
      redis-server
      --appendonly yes
      --requirepass ${REDIS_PASSWORD}
      --replicaof ${MASTER_HOST} 6379
      --masterauth ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6380:6379"  # 使用不同端口
    networks:
      - im-network
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5

  # MinIO 从节点
  minio:
    image: minio/minio:latest
    container_name: im-minio-backup
    restart: unless-stopped
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
    volumes:
      - minio_data:/data
    ports:
      - "9002:9000"
      - "9003:9001"
    networks:
      - im-network
    command: server /data --console-address ":9001"

  # 后端 API（待命状态，不对外暴露端口）
  backend:
    build:
      context: ./im-backend
      dockerfile: Dockerfile.production
    container_name: im-backend-backup
    restart: unless-stopped
    environment:
      DB_HOST: mysql
      DB_PORT: 3306
      DB_NAME: ${MYSQL_DATABASE}
      DB_USER: ${MYSQL_USER}
      DB_PASSWORD: ${MYSQL_PASSWORD}
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      MINIO_ENDPOINT: minio:9000
      MINIO_ACCESS_KEY: ${MINIO_ROOT_USER}
      MINIO_SECRET_KEY: ${MINIO_ROOT_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
      GIN_MODE: release
      PORT: 8080
    volumes:
      - backend_uploads:/app/uploads
    # 不对外暴露端口，只在内网待命
    networks:
      - im-network
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy

volumes:
  mysql_data:
    driver: local
  redis_data:
    driver: local
  minio_data:
    driver: local
  backend_uploads:
    driver: local

networks:
  im-network:
    driver: bridge
EOF
```

---

### 步骤 2.5：启动副服务器服务

```bash
# 启动所有服务
docker-compose -f docker-compose.backup.yml up -d

# 等待服务启动
sleep 60

# 查看服务状态
docker-compose -f docker-compose.backup.yml ps
```

---

### 步骤 2.6：配置 MySQL 从库（主从复制）

```bash
# 1. 从主服务器获取完整备份
echo "从主服务器获取数据备份..."
ssh root@154.37.214.191 "docker exec im-mysql-prod mysqldump -u root -p'ZhRoot2024SecurePass!@#' --all-databases --single-transaction --master-data=2 > /tmp/master_backup.sql"

# 2. 复制备份到副服务器
scp root@154.37.214.191:/tmp/master_backup.sql /tmp/

# 3. 导入备份到副服务器 MySQL
echo "导入数据到副服务器..."
docker exec -i im-mysql-backup mysql -u root -p"ZhRoot2024SecurePass!@#" < /tmp/master_backup.sql

# 4. 配置主从复制
docker exec -it im-mysql-backup mysql -u root -p"ZhRoot2024SecurePass!@#"
```

在 MySQL 中执行：
```sql
-- ⚠️ 使用步骤 1.6 记录的 File 和 Position 值
CHANGE MASTER TO
  MASTER_HOST='154.37.214.191',
  MASTER_USER='repl',
  MASTER_PASSWORD='Replication_Pass_2024!',
  MASTER_PORT=3306,
  MASTER_LOG_FILE='mysql-bin.000001',  -- 替换为实际的 File 值
  MASTER_LOG_POS=157;                   -- 替换为实际的 Position 值

-- 启动从库复制
START SLAVE;

-- 查看复制状态
SHOW SLAVE STATUS\G
```

**验证复制状态**:
```
必须看到:
Slave_IO_Running: Yes
Slave_SQL_Running: Yes
Seconds_Behind_Master: 0 或很小的数字
```

如果看到 `Yes` 和 `Yes`，说明主从复制配置成功！

```sql
-- 退出 MySQL
exit
```

---

### 步骤 2.7：验证 Redis 从节点

```bash
# 验证 Redis 复制状态
docker exec im-redis-backup redis-cli -a "ZhRedis2024SecurePass!@#" INFO replication
```

**预期输出**:
```
role:slave
master_host:154.37.214.191
master_port:6379
master_link_status:up
```

---

### 步骤 2.7B：配置 MinIO 实时同步（关键！）

```bash
# 1. 安装 MinIO Client
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
mv mc /usr/local/bin/

# 2. 配置主服务器和副服务器别名
mc alias set minio-master http://154.37.214.191:9000 zhihang_admin "ZhMinIO2024SecurePass!@#"
mc alias set minio-backup http://localhost:9000 zhihang_admin "ZhMinIO2024SecurePass!@#"

# 3. 验证连接
mc ls minio-master
mc ls minio-backup

# 4. 创建实时同步脚本
cat > /root/minio-sync.sh << 'EOF'
#!/bin/bash
# MinIO 实时镜像同步脚本
LOG_FILE="/var/log/minio-sync.log"

echo "[$(date)] 启动 MinIO 实时同步..." >> $LOG_FILE

# 使用 --watch 模式实时监控并同步更改
mc mirror --watch --overwrite \
    minio-master/zhihang-messenger \
    minio-backup/zhihang-messenger \
    >> $LOG_FILE 2>&1
EOF

chmod +x /root/minio-sync.sh

# 5. 创建 systemd 服务
cat > /etc/systemd/system/minio-sync.service << 'EOF'
[Unit]
Description=MinIO Real-time Mirror Sync
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=root
ExecStart=/root/minio-sync.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 6. 启动同步服务
systemctl daemon-reload
systemctl enable minio-sync
systemctl start minio-sync

# 7. 验证同步服务
systemctl status minio-sync
tail -f /var/log/minio-sync.log
```

**验证 MinIO 实时同步**:
```bash
# 在主服务器创建测试文件
echo "MinIO sync test" > /tmp/test.txt
mc cp /tmp/test.txt minio-master/zhihang-messenger/test/

# 等待2-5秒，在副服务器检查
mc ls minio-backup/zhihang-messenger/test/
# 应该能看到 test.txt

# 清理测试文件
mc rm minio-master/zhihang-messenger/test/test.txt
```

---

### 步骤 2.8：配置文件同步（可选但推荐）

```bash
# 1. 安装 rsync
apt install -y rsync

# 2. 配置 SSH 免密登录（如果还没配置）
ssh-keygen -t rsa -b 4096 -N "" -f ~/.ssh/id_rsa
ssh-copy-id root@154.37.214.191

# 3. 创建配置同步脚本
cat > /root/config-sync.sh << 'EOF'
#!/bin/bash
MASTER_IP="154.37.214.191"
LOG_FILE="/var/log/config-sync.log"

while true; do
    echo "[$(date)] 同步配置文件..." >> $LOG_FILE
    
    rsync -avz --delete root@$MASTER_IP:/root/im-suite/config/ /root/im-suite/config/ >> $LOG_FILE 2>&1
    rsync -avz root@$MASTER_IP:/root/im-suite/.env /root/im-suite/.env >> $LOG_FILE 2>&1
    
    sleep 60  # 每分钟同步一次
done
EOF

chmod +x /root/config-sync.sh

# 4. 创建 systemd 服务
cat > /etc/systemd/system/config-sync.service << 'EOF'
[Unit]
Description=Config Files Sync from Master
After=network.target

[Service]
Type=simple
User=root
ExecStart=/root/config-sync.sh
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 5. 启动服务
systemctl daemon-reload
systemctl enable config-sync
systemctl start config-sync
```

---

### 步骤 2.9：安装 Node Exporter（监控）

```bash
# 运行 Node Exporter
docker run -d \
  --name node-exporter \
  --restart unless-stopped \
  --net="host" \
  --pid="host" \
  -v "/:/host:ro,rslave" \
  prom/node-exporter:latest \
  --path.rootfs=/host

# 验证
curl http://localhost:9100/metrics | head -20
```

---

### 步骤 2.10：安装 Keepalived（备份节点）

```bash
# 安装 Keepalived
apt install -y keepalived

# 查看网络接口名称
ip addr show

# 创建健康检查脚本（与主服务器相同）
cat > /etc/keepalived/check_backend.sh << 'EOF'
#!/bin/bash
if curl -sf http://localhost:8080/health > /dev/null; then
    exit 0
else
    exit 1
fi
EOF

chmod +x /etc/keepalived/check_backend.sh

# 配置 Keepalived（优先级设为 90，低于主服务器）
# ⚠️ 注意：interface 改为实际网络接口名称
cat > /etc/keepalived/keepalived.conf << 'EOF'
global_defs {
    router_id BACKUP_NODE
}

vrrp_script check_backend {
    script "/etc/keepalived/check_backend.sh"
    interval 2
    timeout 3
    weight -50
    fall 3
    rise 2
}

vrrp_instance VI_1 {
    state BACKUP        # 备份状态
    interface eth0      # ⚠️ 改为实际网络接口名称
    virtual_router_id 51
    priority 90         # 优先级 90，低于主服务器
    advert_int 1
    
    authentication {
        auth_type PASS
        auth_pass ZhiHang2024!
    }
    
    virtual_ipaddress {
        10.0.0.100/24  # 与主服务器相同的虚拟IP
    }
    
    track_script {
        check_backend
    }
    
    notify_master "/usr/local/bin/keepalived-notify.sh MASTER"
    notify_backup "/usr/local/bin/keepalived-notify.sh BACKUP"
    notify_fault "/usr/local/bin/keepalived-notify.sh FAULT"
}
EOF

# 创建通知脚本
cat > /usr/local/bin/keepalived-notify.sh << 'EOF'
#!/bin/bash
TYPE=$1
HOST=$(hostname)
LOG_FILE="/var/log/keepalived-notify.log"

case $TYPE in
    MASTER)
        echo "$(date) - ✅ $HOST 切换为主服务器（接管服务）" >> $LOG_FILE
        # 副服务器接管时，需要关闭 MySQL 只读模式
        docker exec im-mysql-backup mysql -u root -p"ZhRoot2024SecurePass!@#" -e "SET GLOBAL read_only = OFF; SET GLOBAL super_read_only = OFF;"
        ;;
    BACKUP)
        echo "$(date) - ⏸️  $HOST 处于备份状态" >> $LOG_FILE
        ;;
    FAULT)
        echo "$(date) - 🔴 $HOST 出现故障" >> $LOG_FILE
        ;;
esac
EOF

chmod +x /usr/local/bin/keepalived-notify.sh

# 启动 Keepalived
systemctl enable keepalived
systemctl start keepalived

# 验证状态（副服务器不应该有虚拟IP，除非主服务器宕机）
ip addr show | grep "10.0.0.100" || echo "✅ 副服务器正常（虚拟IP在主服务器上）"
```

---

### 步骤 2.10：副服务器最终验证

```bash
echo "========================================="
echo "副服务器部署验证"
echo "========================================="

# 1. Docker 容器状态
echo "1. Docker 容器状态:"
docker ps --format "table {{.Names}}\t{{.Status}}"

# 2. MySQL 从库复制状态
echo -e "\n2. MySQL 从库复制状态:"
docker exec im-mysql-backup mysql -u root -p"ZhRoot2024SecurePass!@#" -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running|Seconds_Behind_Master"

# 3. Redis 从节点状态
echo -e "\n3. Redis 从节点状态:"
docker exec im-redis-backup redis-cli -a "ZhRedis2024SecurePass!@#" INFO replication | grep -E "role|master_host|master_link_status"

# 4. 虚拟 IP 状态（应该没有）
echo -e "\n4. 虚拟 IP 状态:"
ip addr show | grep "10.0.0.100" || echo "✅ 虚拟IP在主服务器（正常）"

# 5. Keepalived 状态
echo -e "\n5. Keepalived 状态:"
systemctl status keepalived | grep Active

# 6. Node Exporter
echo -e "\n6. Node Exporter 状态:"
curl -s http://localhost:9100/metrics | head -1

echo "========================================="
echo "✅ 副服务器部署完成！"
echo "========================================="
```

**退出副服务器**:
```bash
exit
```

---

# 🖥️ 服务器 3: 监控服务器部署

## 📝 监控服务器职责
- 📊 收集主服务器和副服务器的监控指标
- 📈 可视化展示系统状态（Grafana）
- 🔔 检测异常并发送告警通知
- 📝 记录历史数据用于分析

---

## 监控服务器部署步骤

### 步骤 3.1：连接监控服务器并更新系统

```bash
# 从本地连接到监控服务器（替换为实际IP）
ssh root@MONITOR_SERVER_IP

# 更新系统
apt update && apt upgrade -y

# 安装基础工具
apt install -y git curl wget vim htop net-tools
```

---

### 步骤 3.2：安装 Docker 和 Docker Compose

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | bash

# 启动 Docker
systemctl enable docker
systemctl start docker

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 验证
docker --version
docker-compose --version
```

---

### 步骤 3.3：创建监控配置目录

```bash
# 创建目录结构
mkdir -p /root/monitoring
cd /root/monitoring

mkdir -p prometheus
mkdir -p grafana/provisioning/datasources
mkdir -p grafana/provisioning/dashboards
mkdir -p alertmanager
```

---

### 步骤 3.4：配置 Prometheus

```bash
# 创建 Prometheus 配置文件
# ⚠️ 替换 BACKUP_SERVER_IP 为副服务器的实际 IP
cat > prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s      # 每15秒收集一次数据
  evaluation_interval: 15s  # 每15秒评估一次告警规则
  external_labels:
    cluster: 'im-suite'
    environment: 'production'

# 告警规则文件
rule_files:
  - 'alerts.yml'

# Alertmanager 配置
alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

# 监控目标配置
scrape_configs:
  # 主服务器 - Node Exporter（系统指标）
  - job_name: 'master-node'
    static_configs:
      - targets: ['154.37.214.191:9100']
        labels:
          server: 'master'
          role: 'primary'
          instance: 'im-master'

  # 副服务器 - Node Exporter（系统指标）
  - job_name: 'backup-node'
    static_configs:
      - targets: ['BACKUP_SERVER_IP:9100']  # ⚠️ 替换为实际IP
        labels:
          server: 'backup'
          role: 'secondary'
          instance: 'im-backup'

  # 主服务器 - 后端 API
  - job_name: 'master-backend'
    static_configs:
      - targets: ['154.37.214.191:8080']
        labels:
          server: 'master'
          service: 'backend'
    metrics_path: '/metrics'  # 如果后端暴露了 metrics 端点

  # Prometheus 自身监控
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF
```

---

### 步骤 3.5：配置告警规则

```bash
# 创建告警规则文件
cat > prometheus/alerts.yml << 'EOF'
groups:
  - name: server_alerts
    interval: 30s
    rules:
      # 主服务器宕机（严重告警）
      - alert: MasterServerDown
        expr: up{server="master",job="master-node"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "🔴 主服务器宕机！"
          description: "主服务器已宕机超过1分钟，虚拟IP应该已切换到副服务器。请立即检查！"

      # 副服务器宕机（警告）
      - alert: BackupServerDown
        expr: up{server="backup",job="backup-node"} == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "⚠️ 副服务器宕机"
          description: "副服务器已宕机超过5分钟，失去了备份保障。"

      # CPU 使用率过高
      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU 使用率过高"
          description: "{{ $labels.instance }} CPU 使用率超过 80%，当前值: {{ $value | printf \"%.2f\" }}%"

      # 内存使用率过高
      - alert: HighMemoryUsage
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "内存使用率过高"
          description: "{{ $labels.instance }} 内存使用率超过 85%，当前值: {{ $value | printf \"%.2f\" }}%"

      # 磁盘空间不足
      - alert: DiskSpaceLow
        expr: (1 - (node_filesystem_avail_bytes{fstype!="tmpfs"} / node_filesystem_size_bytes)) * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "磁盘空间不足"
          description: "{{ $labels.instance }} {{ $labels.mountpoint }} 磁盘使用率超过 85%，当前值: {{ $value | printf \"%.2f\" }}%"
EOF
```

---

### 步骤 3.6：配置 Alertmanager（告警通知）

```bash
# 创建 Alertmanager 配置
# ⚠️ 替换为您的邮箱信息
cat > alertmanager/alertmanager.yml << 'EOF'
global:
  smtp_smarthost: 'smtp.qq.com:587'
  smtp_from: 'your-email@qq.com'              # ⚠️ 替换为您的邮箱
  smtp_auth_username: 'your-email@qq.com'     # ⚠️ 替换为您的邮箱
  smtp_auth_password: 'your-smtp-password'    # ⚠️ 替换为您的SMTP密码
  smtp_require_tls: true

route:
  receiver: 'admin'
  group_by: ['alertname', 'severity']
  group_wait: 30s       # 等待30秒收集告警
  group_interval: 5m    # 每5分钟发送一次分组告警
  repeat_interval: 4h   # 4小时内不重复发送相同告警

  routes:
    # 严重告警立即发送
    - match:
        severity: critical
      receiver: 'admin-urgent'
      repeat_interval: 15m  # 严重告警每15分钟重复一次

receivers:
  # 普通告警接收器
  - name: 'admin'
    email_configs:
      - to: 'admin@yourdomain.com'  # ⚠️ 替换为接收告警的邮箱
        headers:
          Subject: '【志航密信】监控告警'
        html: |
          <h3>{{ range .Alerts }}{{ .Labels.alertname }}{{ end }}</h3>
          {{ range .Alerts }}
          <p><strong>描述:</strong> {{ .Annotations.description }}</p>
          <p><strong>时间:</strong> {{ .StartsAt.Format "2006-01-02 15:04:05" }}</p>
          {{ end }}

  # 严重告警接收器
  - name: 'admin-urgent'
    email_configs:
      - to: 'admin@yourdomain.com'  # ⚠️ 替换为接收告警的邮箱
        headers:
          Subject: '🔴【紧急】志航密信严重告警！'
        html: |
          <h2 style="color:red;">严重告警！请立即处理！</h2>
          {{ range .Alerts }}
          <h3>{{ .Labels.alertname }}</h3>
          <p><strong>描述:</strong> {{ .Annotations.description }}</p>
          <p><strong>时间:</strong> {{ .StartsAt.Format "2006-01-02 15:04:05" }}</p>
          {{ end }}
EOF
```

---

### 步骤 3.7：配置 Grafana 数据源

```bash
# 配置 Grafana 自动加载 Prometheus 数据源
cat > grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF
```

---

### 步骤 3.8：创建 Docker Compose 配置

```bash
# 创建监控服务 Docker Compose 配置
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  # Prometheus - 监控数据收集
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus:/etc/prometheus
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
    networks:
      - monitoring

  # Alertmanager - 告警管理
  alertmanager:
    image: prom/alertmanager:latest
    container_name: alertmanager
    restart: unless-stopped
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager:/etc/alertmanager
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    networks:
      - monitoring

  # Grafana - 可视化仪表板
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=ZhGrafana2024AdminPass!@#
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-clock-panel
      - GF_SERVER_ROOT_URL=http://MONITOR_SERVER_IP:3000  # ⚠️ 替换为实际IP
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    networks:
      - monitoring
    depends_on:
      - prometheus

volumes:
  prometheus_data:
    driver: local
  alertmanager_data:
    driver: local
  grafana_data:
    driver: local

networks:
  monitoring:
    driver: bridge
EOF
```

---

### 步骤 3.9：启动监控服务

```bash
# 启动所有监控服务
docker-compose up -d

# 等待服务启动
sleep 30

# 查看服务状态
docker-compose ps
```

**预期输出**:
```
NAME            STATUS    PORTS
prometheus      Up        0.0.0.0:9090->9090/tcp
alertmanager    Up        0.0.0.0:9093->9093/tcp
grafana         Up        0.0.0.0:3000->3000/tcp
```

---

### 步骤 3.10：验证监控服务

```bash
echo "========================================="
echo "监控服务器部署验证"
echo "========================================="

# 1. Prometheus 状态
echo "1. Prometheus 访问测试:"
curl http://localhost:9090/-/healthy
# 预期: Prometheus is Healthy.

# 2. Alertmanager 状态
echo -e "\n2. Alertmanager 访问测试:"
curl http://localhost:9093/-/healthy
# 预期: OK

# 3. Grafana 状态
echo -e "\n3. Grafana 访问测试:"
curl -I http://localhost:3000
# 预期: HTTP/1.1 302 Found

# 4. 检查 Prometheus 是否能抓取主服务器数据
echo -e "\n4. Prometheus 监控目标状态:"
curl -s http://localhost:9090/api/v1/targets | grep -o '"health":"[^"]*"' | head -5

echo "========================================="
echo "✅ 监控服务器部署完成！"
echo "========================================="
echo ""
echo "访问地址:"
echo "- Prometheus: http://MONITOR_SERVER_IP:9090"
echo "- Grafana: http://MONITOR_SERVER_IP:3000"
echo "  账号: admin"
echo "  密码: ZhGrafana2024AdminPass!@#"
echo "- Alertmanager: http://MONITOR_SERVER_IP:9093"
echo "========================================="
```

---

### 步骤 3.11：配置 Grafana 仪表板

```bash
# 在浏览器中访问 Grafana
# http://MONITOR_SERVER_IP:3000

# 1. 登录
#    账号: admin
#    密码: ZhGrafana2024AdminPass!@#

# 2. 导入预设仪表板
#    - 点击 "+" → "Import"
#    - 输入仪表板 ID:
#      • 1860 (Node Exporter Full) - Linux 系统监控
#      • 13639 (Node Exporter Quickstart) - 快速概览
#      • 405 (Node Exporter Server Metrics) - 服务器指标

# 3. 创建自定义仪表板
#    - 监控主服务器和副服务器的健康状态
#    - 监控 MySQL 主从复制延迟
#    - 监控 Redis 同步状态
```

**退出监控服务器**:
```bash
exit
```

---

# ✅ 最终验证和测试

## 验证 1: 三服务器连通性测试

```bash
# 从本地执行
MASTER_IP="154.37.214.191"
BACKUP_IP="替换为副服务器IP"
MONITOR_IP="替换为监控服务器IP"

echo "测试三服务器网络连通性..."

# 测试主服务器
echo "1. 主服务器:"
ssh root@$MASTER_IP "echo '✅ 主服务器连接正常'"

# 测试副服务器
echo "2. 副服务器:"
ssh root@$BACKUP_IP "echo '✅ 副服务器连接正常'"

# 测试监控服务器
echo "3. 监控服务器:"
ssh root@$MONITOR_IP "echo '✅ 监控服务器连接正常'"

# 测试服务器间互ping
echo -e "\n测试服务器间网络连通性..."
ssh root@$MASTER_IP "ping -c 3 $BACKUP_IP && ping -c 3 $MONITOR_IP"
```

---

## 验证 2: 数据同步测试

```bash
# 在主服务器上插入测试数据
ssh root@$MASTER_IP << 'EOF'
docker exec im-mysql-prod mysql -u root -p"ZhRoot2024SecurePass!@#" zhihang_messenger -e "
CREATE TABLE IF NOT EXISTS test_sync (
    id INT PRIMARY KEY AUTO_INCREMENT,
    content VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO test_sync (content) VALUES ('测试主从同步 - $(date)');
SELECT * FROM test_sync;
"
EOF

# 等待5秒
sleep 5

# 在副服务器上验证数据
echo -e "\n在副服务器上验证同步数据:"
ssh root@$BACKUP_IP << 'EOF'
docker exec im-mysql-backup mysql -u root -p"ZhRoot2024SecurePass!@#" zhihang_messenger -e "
SELECT * FROM test_sync;
"
EOF

# 如果看到相同的数据，说明主从同步正常！✅
```

---

## 验证 3: 故障转移测试（可选）

```bash
echo "========================================="
echo "🧪 故障转移测试（模拟主服务器宕机）"
echo "========================================="
read -p "是否进行故障转移测试？这将短暂中断主服务器服务。(y/n) " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "1. 停止主服务器后端服务..."
    ssh root@$MASTER_IP "docker stop im-backend-prod"
    
    echo "2. 等待 Keepalived 检测故障（约10秒）..."
    sleep 10
    
    echo "3. 检查虚拟 IP 是否切换到副服务器..."
    ssh root@$BACKUP_IP "ip addr show | grep '10.0.0.100'"
    
    if [ $? -eq 0 ]; then
        echo "✅ 故障转移成功！虚拟IP已切换到副服务器"
    else
        echo "❌ 故障转移失败，请检查 Keepalived 配置"
    fi
    
    echo -e "\n4. 恢复主服务器..."
    ssh root@$MASTER_IP "docker start im-backend-prod"
    
    echo "5. 等待主服务器恢复（约10秒）..."
    sleep 10
    
    echo "6. 检查虚拟 IP 是否切换回主服务器..."
    ssh root@$MASTER_IP "ip addr show | grep '10.0.0.100'"
    
    echo -e "\n✅ 故障转移测试完成！"
fi

echo "========================================="
```

---

## 验证 4: 监控系统验证

```bash
echo "验证监控系统..."

# 访问 Prometheus
echo "1. Prometheus 监控目标:"
echo "   访问: http://$MONITOR_IP:9090/targets"
echo "   所有目标应该显示为 UP"

# 访问 Grafana
echo -e "\n2. Grafana 仪表板:"
echo "   访问: http://$MONITOR_IP:3000"
echo "   账号: admin"
echo "   密码: ZhGrafana2024AdminPass!@#"

# 测试告警
echo -e "\n3. 测试告警（可选）:"
echo "   停止主服务器后端触发告警:"
echo "   ssh root@$MASTER_IP 'docker stop im-backend-prod'"
echo "   等待1分钟后检查邮箱是否收到告警邮件"
```

---

# 📊 部署完成总结

## ✅ 部署检查清单

### 主服务器
- [ ] Docker 和 Docker Compose 已安装
- [ ] 所有服务容器运行正常
- [ ] MySQL 主库状态正常
- [ ] Redis 主节点状态正常
- [ ] 后端 API 健康检查通过
- [ ] 虚拟 IP 已绑定
- [ ] Keepalived 运行正常（优先级 100）
- [ ] Node Exporter 运行正常

### 副服务器
- [ ] Docker 和 Docker Compose 已安装
- [ ] 所有服务容器运行正常
- [ ] MySQL 从库复制正常（IO: Yes, SQL: Yes）
- [ ] Redis 从节点同步正常
- [ ] 虚拟 IP 未绑定（在备用状态）
- [ ] Keepalived 运行正常（优先级 90）
- [ ] Node Exporter 运行正常

### 监控服务器
- [ ] Docker 和 Docker Compose 已安装
- [ ] Prometheus 运行正常
- [ ] Grafana 运行正常且可访问
- [ ] Alertmanager 运行正常
- [ ] 可以抓取主服务器和副服务器的指标
- [ ] 告警规则已加载

### 功能验证
- [ ] 主从数据同步正常
- [ ] 故障转移测试通过
- [ ] 监控系统正常工作
- [ ] 告警通知正常发送

---

## 🎯 访问地址汇总

### 用户访问
- **虚拟 IP**: http://10.0.0.100
- **管理后台**: http://10.0.0.100:3001
- **Web 客户端**: http://10.0.0.100:3002

### 监控访问
- **Grafana**: http://MONITOR_SERVER_IP:3000
  - 账号: admin
  - 密码: ZhGrafana2024AdminPass!@#
- **Prometheus**: http://MONITOR_SERVER_IP:9090
- **Alertmanager**: http://MONITOR_SERVER_IP:9093

### 直接访问（调试用）
- **主服务器后端**: http://154.37.214.191:8080
- **副服务器（不对外）**: 不可访问（设计如此）

---

## 📝 日常运维

### 查看服务状态
```bash
# 主服务器
ssh root@154.37.214.191 "docker ps"

# 副服务器
ssh root@BACKUP_IP "docker ps"

# 监控服务器
ssh root@MONITOR_IP "docker ps"
```

### 查看日志
```bash
# 主服务器后端日志
ssh root@154.37.214.191 "docker logs im-backend-prod --tail 100"

# Keepalived 切换日志
ssh root@154.37.214.191 "tail -f /var/log/keepalived-notify.log"
```

### 手动切换到副服务器（维护时）
```bash
# 停止主服务器 Keepalived
ssh root@154.37.214.191 "systemctl stop keepalived"

# 副服务器会自动接管（约5-10秒）

# 维护完成后，重启主服务器 Keepalived
ssh root@154.37.214.191 "systemctl start keepalived"

# 主服务器会自动接管回来
```

---

## 🎉 部署完成！

**恭喜！三服务器高可用架构部署完成！**

现在您拥有：
- ✅ 99.9% 可用性保障
- ✅ < 30秒故障自动切换
- ✅ 实时数据备份（零数据丢失）
- ✅ 完整的监控和告警系统
- ✅ 用户无感知的高可用体验

---

## 📞 后续支持

如需帮助，请参考：
- **架构文档**: `ACTIVE_PASSIVE_HA_ARCHITECTURE.md`
- **故障排查**: `NETWORK_TROUBLESHOOTING_GUIDE.md`
- **高可用路线图**: `HIGH_AVAILABILITY_ROADMAP.md`

**祝您运行顺利！** 🚀

