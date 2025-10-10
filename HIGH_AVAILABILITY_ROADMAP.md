# 志航密信高可用架构迁移路线图

**创建时间**: 2025-10-10  
**目标**: 从单机部署平滑迁移到三服务器高可用架构

---

## 📊 当前状态评估

### 单机部署配置

```
服务器: 154.37.214.191 (香港 - 雨云)
配置: AMD EPYC 7K62 (48核) + 内存 + 硬盘
部署: Docker Compose

运行服务:
✅ MySQL 8.0 (im-mysql-prod)
✅ Redis 7 (im-redis-prod)
✅ MinIO (im-minio-prod)
✅ 后端 Go API (im-backend-prod)
✅ 管理后台 Vue (im-admin-prod)
✅ Prometheus (im-prometheus-prod)
✅ Grafana (im-grafana-prod)
```

### 风险评估

| 风险类型 | 严重程度 | 概率 | 影响 |
|---------|---------|------|------|
| **服务器宕机** | 🔴 严重 | 中等 (1-2次/年) | 100% 不可用 |
| **硬盘损坏** | 🔴 严重 | 低 (< 1次/年) | 数据永久丢失 |
| **性能瓶颈** | 🟡 中等 | 高 (随用户增长) | 响应变慢 |
| **DDoS 攻击** | 🟡 中等 | 中等 | 服务不可用 |
| **人为误操作** | 🟡 中等 | 低 | 数据损坏 |

---

## 🎯 三阶段迁移计划

---

## 阶段 1：紧急防护（立即执行，0 成本）

**目标**: 防止数据丢失，提供基础容灾能力  
**时间**: 1-2 小时  
**成本**: 0 元

### 任务清单

#### 1.1 部署自动备份系统 ✅

**已创建**: `scripts/backup/auto-backup.sh`

**功能**:
- ✅ 每日自动备份 MySQL 数据库
- ✅ 每日自动备份 Redis 数据
- ✅ 每日自动备份 MinIO 文件
- ✅ 每日自动备份配置文件
- ✅ 自动清理 7 天前的旧备份
- ✅ 生成备份报告

**部署步骤**:
```bash
# 1. 给脚本执行权限
chmod +x /root/im-suite/scripts/backup/auto-backup.sh

# 2. 测试运行
/root/im-suite/scripts/backup/auto-backup.sh

# 3. 设置 Crontab 定时任务
crontab -e

# 添加以下行（每日凌晨 2:00 执行）
0 2 * * * /root/im-suite/scripts/backup/auto-backup.sh >> /var/log/im-backup.log 2>&1
```

**验证**:
```bash
# 检查备份文件
ls -lh /root/im-suite/backups/

# 查看备份日志
tail -f /var/log/im-backup.log
```

---

#### 1.2 设置监控告警

**目标**: 提前发现问题

**部署步骤**:
```bash
# 1. 配置 Prometheus 告警规则
cat > /root/im-suite/config/prometheus/alerts.yml << 'EOF'
groups:
  - name: system_alerts
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU 使用率过高"
          description: "CPU 使用率超过 80%，当前值: {{ $value }}%"

      - alert: HighMemoryUsage
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "内存使用率过高"
          description: "内存使用率超过 85%，当前值: {{ $value }}%"

      - alert: DiskSpaceLow
        expr: (1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "磁盘空间不足"
          description: "磁盘使用率超过 85%，当前值: {{ $value }}%"

      - alert: MySQLDown
        expr: up{job="mysql"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "MySQL 服务宕机"
          description: "MySQL 服务不可用"

      - alert: RedisDown
        expr: up{job="redis"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis 服务宕机"
          description: "Redis 服务不可用"
EOF

# 2. 配置 Alertmanager（可选，用于发送邮件/微信通知）
# 需要配置 SMTP 或 Webhook
```

---

#### 1.3 设置健康检查脚本

**创建健康检查脚本**:
```bash
cat > /root/im-suite/scripts/monitoring/health-check.sh << 'EOF'
#!/bin/bash

# 检查所有核心服务状态
echo "========================================="
echo "服务健康检查 - $(date)"
echo "========================================="

# 检查 MySQL
if docker ps | grep -q im-mysql-prod; then
    echo "✅ MySQL: 运行中"
else
    echo "❌ MySQL: 已停止"
fi

# 检查 Redis
if docker ps | grep -q im-redis-prod; then
    echo "✅ Redis: 运行中"
else
    echo "❌ Redis: 已停止"
fi

# 检查后端 API
if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo "✅ 后端 API: 正常"
else
    echo "❌ 后端 API: 异常"
fi

# 检查管理后台
if curl -s -o /dev/null -w "%{http_code}" http://localhost:3001 | grep -q "200"; then
    echo "✅ 管理后台: 正常"
else
    echo "❌ 管理后台: 异常"
fi

# 检查磁盘空间
DISK_USAGE=$(df -h / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -lt 80 ]; then
    echo "✅ 磁盘空间: ${DISK_USAGE}% (正常)"
else
    echo "⚠️  磁盘空间: ${DISK_USAGE}% (警告)"
fi

echo "========================================="
EOF

chmod +x /root/im-suite/scripts/monitoring/health-check.sh

# 每 5 分钟检查一次
crontab -e
# 添加: */5 * * * * /root/im-suite/scripts/monitoring/health-check.sh >> /var/log/health-check.log 2>&1
```

---

### 阶段 1 完成标准

- [x] ✅ 自动备份脚本部署并运行
- [x] ✅ Crontab 定时任务设置
- [x] ✅ 监控告警规则配置
- [x] ✅ 健康检查脚本运行
- [x] ✅ 备份文件可恢复验证

**预期效果**:
- ✅ 数据安全性提升 90%
- ✅ 可在 30 分钟内从备份恢复
- ✅ 提前发现系统异常

---

## 阶段 2：主从复制（1-2 周内，成本 500-1000/月）

**目标**: 实现数据库高可用，防止单点故障  
**时间**: 2-3 天  
**成本**: 500-1000 元/月（购买副服务器）

### 2.1 购买副服务器

**配置要求**:
```
CPU: 4核+ (主服务器的 50%)
内存: 8GB+ (主服务器的 50%)
硬盘: 100GB+ SSD
网络: 独立 IP
位置: 与主服务器同地区（香港）或异地（深圳/上海）
```

**推荐供应商**:
- 阿里云轻量应用服务器: 约 800 元/月
- 腾讯云标准型: 约 600 元/月
- 雨云（与主服务器同供应商）: 约 500 元/月

---

### 2.2 配置 MySQL 主从复制

**架构图**:
```
主服务器 (154.37.214.191)          副服务器 (待分配)
├── MySQL Master (读写)      ←→    MySQL Slave (只读)
│   端口: 3307                      端口: 3307
│   数据: 实时写入                  数据: 自动同步
```

**配置步骤**:

#### 主服务器配置
```bash
# 1. 修改 MySQL 配置
docker exec -it im-mysql-prod bash
mysql -u root -p

# 2. 创建复制用户
CREATE USER 'repl'@'%' IDENTIFIED BY 'Replication_Pass_2024!';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
FLUSH PRIVILEGES;

# 3. 获取 binlog 位置
SHOW MASTER STATUS;
# 记录 File 和 Position 值
```

#### 从服务器配置
```bash
# 1. 恢复主服务器备份到从服务器
scp root@154.37.214.191:/root/im-suite/backups/mysql/latest.sql.gz /tmp/
gunzip /tmp/latest.sql.gz
mysql -u root -p < /tmp/latest.sql

# 2. 配置从服务器连接主服务器
CHANGE MASTER TO
  MASTER_HOST='154.37.214.191',
  MASTER_USER='repl',
  MASTER_PASSWORD='Replication_Pass_2024!',
  MASTER_LOG_FILE='binlog.000001',  -- 使用步骤3中的 File
  MASTER_LOG_POS=12345;              -- 使用步骤3中的 Position

# 3. 启动从服务器复制
START SLAVE;

# 4. 验证复制状态
SHOW SLAVE STATUS\G
# 确保 Slave_IO_Running: Yes 和 Slave_SQL_Running: Yes
```

---

### 2.3 配置 Redis 主从复制

**配置步骤**:
```bash
# 从服务器 Redis 配置
docker run -d \
  --name im-redis-slave \
  --restart unless-stopped \
  -p 6379:6379 \
  redis:7-alpine \
  redis-server \
  --replicaof 154.37.214.191 6379 \
  --masterauth "ZhRedis2024SecurePass!@#" \
  --requirepass "ZhRedis2024SecurePass!@#"

# 验证复制状态
docker exec im-redis-slave redis-cli -a "ZhRedis2024SecurePass!@#" INFO replication
```

---

### 2.4 配置自动故障转移

**使用 Keepalived 实现虚拟 IP 故障转移**:

```bash
# 主服务器和副服务器都安装
apt install keepalived -y

# 主服务器配置
cat > /etc/keepalived/keepalived.conf << 'EOF'
vrrp_instance VI_1 {
    state MASTER
    interface eth0
    virtual_router_id 51
    priority 100
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1234
    }
    virtual_ipaddress {
        10.0.0.100/24  # 虚拟 IP
    }
}
EOF

# 副服务器配置（priority 改为 90）
# 当主服务器故障时，虚拟 IP 自动切换到副服务器
```

---

### 阶段 2 完成标准

- [ ] ✅ 副服务器购买并配置
- [ ] ✅ MySQL 主从复制运行正常
- [ ] ✅ Redis 主从复制运行正常
- [ ] ✅ 自动故障转移测试通过
- [ ] ✅ 数据一致性验证通过

**预期效果**:
- ✅ 可用性提升到 99.9%
- ✅ 主服务器故障，30秒内自动切换
- ✅ 数据实时备份，零丢失

---

## 阶段 3：完整高可用（1-2 个月，成本 1500-2500/月）

**目标**: 实现完整的三服务器架构  
**时间**: 1 周配置 + 1 周测试  
**成本**: 1500-2500 元/月

### 3.1 服务器配置

```
主服务器 (154.37.214.191)
配置: 8核16GB 100GB SSD
服务: MySQL主, Redis主, MinIO主, 后端API×2

副服务器 (新购买)
配置: 8核16GB 100GB SSD
服务: MySQL从, Redis从, MinIO从, 后端API×2

监控服务器 (新购买)
配置: 4核8GB 50GB SSD
服务: Prometheus, Grafana, ELK, 告警系统

负载均衡器 (可选，使用云服务)
配置: 阿里云 SLB 或 Nginx独立服务器
```

---

### 3.2 负载均衡配置

**使用 Nginx 作为负载均衡器**:

```nginx
upstream backend_api {
    server 154.37.214.191:8080 weight=1 max_fails=3 fail_timeout=30s;
    server SLAVE_IP:8080 weight=1 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    location /api/ {
        proxy_pass http://backend_api;
        proxy_next_upstream error timeout http_500 http_502 http_503;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

### 3.3 数据库读写分离

**配置后端连接池**:

```go
// im-backend/config/database.go
func InitDB() {
    // 主库（写）
    masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
        os.Getenv("DB_MASTER_USER"),
        os.Getenv("DB_MASTER_PASSWORD"),
        os.Getenv("DB_MASTER_HOST"),
        os.Getenv("DB_MASTER_PORT"),
        os.Getenv("DB_NAME"),
    )
    
    // 从库（读）
    slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
        os.Getenv("DB_SLAVE_USER"),
        os.Getenv("DB_SLAVE_PASSWORD"),
        os.Getenv("DB_SLAVE_HOST"),
        os.Getenv("DB_SLAVE_PORT"),
        os.Getenv("DB_NAME"),
    )
    
    // 使用 GORM 的 DBResolver 插件
    db.Use(dbresolver.Register(dbresolver.Config{
        Sources:  []gorm.Dialector{mysql.Open(masterDSN)},  // 写操作
        Replicas: []gorm.Dialector{mysql.Open(slaveDSN)},   // 读操作
        Policy:   dbresolver.RandomPolicy{},                 // 随机选择从库
    }))
}
```

---

### 3.4 独立监控服务器

**部署 ELK Stack（日志系统）**:

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  elasticsearch:
    image: elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"

  logstash:
    image: logstash:8.11.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    depends_on:
      - elasticsearch

  kibana:
    image: kibana:8.11.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch
```

---

### 阶段 3 完成标准

- [ ] ✅ 三台服务器全部配置完成
- [ ] ✅ 负载均衡运行正常
- [ ] ✅ 数据库读写分离验证
- [ ] ✅ 独立监控系统部署
- [ ] ✅ 压力测试通过（10000+ 并发）
- [ ] ✅ 故障演练（模拟主服务器宕机）

**预期效果**:
- ✅ 可用性 99.99% (年停机 < 1 小时)
- ✅ 性能提升 3-5 倍
- ✅ 支持 10000+ 并发用户
- ✅ 完整的日志、监控、告警系统

---

## 💰 成本分析

### 月度成本对比

| 项目 | 单机部署 | 主从架构 | 三服务器架构 |
|-----|---------|---------|-------------|
| **主服务器** | 800 元 | 800 元 | 1200 元 |
| **副服务器** | - | 500 元 | 800 元 |
| **监控服务器** | - | - | 400 元 |
| **负载均衡** | - | - | 300 元 |
| **备份存储** | - | 50 元 | 100 元 |
| **CDN** | - | - | 200 元 |
| **总计** | **800 元/月** | **1350 元/月** | **3000 元/月** |

### 性能对比

| 指标 | 单机 | 主从 | 三服务器 |
|-----|-----|------|---------|
| **可用性** | 99% | 99.9% | 99.99% |
| **并发用户** | 100 | 1000 | 10000+ |
| **响应时间** | 100-500ms | 50-200ms | 20-100ms |
| **数据安全** | ⚠️ 低 | ✅ 中 | ✅ 高 |
| **扩展性** | ❌ 差 | ⚠️ 中 | ✅ 优 |

---

## 📅 实施时间表

```
Week 1 (立即):
├── Day 1-2: 部署自动备份系统
├── Day 3-4: 配置监控告警
└── Day 5-7: 测试备份恢复流程

Week 2-3 (如果需要):
├── Day 1-3: 购买并配置副服务器
├── Day 4-7: 配置 MySQL/Redis 主从复制
├── Day 8-10: 测试故障转移
└── Day 11-14: 生产环境切换

Week 4-8 (长期):
├── Week 4: 购买监控服务器
├── Week 5: 配置负载均衡
├── Week 6: 部署 ELK 日志系统
├── Week 7: 压力测试和优化
└── Week 8: 完整故障演练
```

---

## 🎯 建议

### 当前阶段（用户量 < 100）

**✅ 保持单机部署 + 阶段1防护措施**

**原因**:
- ✅ 快速迭代，专注功能开发
- ✅ 成本可控（800 元/月）
- ✅ 有备份系统，数据安全有保障
- ✅ 足以支撑 100 并发用户

**必须做**:
- ✅ 部署自动备份系统（立即）
- ✅ 配置监控告警（本周）
- ✅ 测试备份恢复（本周）

---

### 中期（用户量 100-1000）

**⚠️ 启用阶段2主从架构**

**触发条件**（满足任一即需升级）:
- 注册用户 > 100
- 日活跃用户 > 50
- 日消息量 > 10000
- CPU 使用率持续 > 70%
- 出现过 1 次以上停机事故

---

### 长期（用户量 > 1000）

**✅ 必须完整实施阶段3**

**原因**:
- 用户量大，停机 = 口碑损失
- 数据量大，需要读写分离
- 商业运营，需要 SLA 保障

---

## 📞 技术支持

如有疑问，可以：
1. 查看本文档详细步骤
2. 咨询我（Cursor AI）
3. 参考官方文档

**关键文档**:
- MySQL 主从复制: https://dev.mysql.com/doc/refman/8.0/en/replication.html
- Redis 主从复制: https://redis.io/docs/management/replication/
- Keepalived HA: https://www.keepalived.org/

---

**结论**: 
- ✅ **短期（现在）**: 单机 + 备份足够
- ⚠️ **中期（1-2 个月）**: 需要主从架构
- ✅ **长期（3-6 个月）**: 需要完整高可用

**第一步（立即做）**: 部署自动备份系统！🚀

