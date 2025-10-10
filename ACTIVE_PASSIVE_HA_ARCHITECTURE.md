# 志航密信主备高可用架构设计方案

**架构类型**: Active-Passive（主备模式）  
**设计日期**: 2025-10-10  
**核心理念**: 主服务器提供服务，副服务器实时备份，故障时自动切换

---

## 🎯 架构目标

### 用户体验
- ✅ **零感知切换**: 主服务器故障时，用户自动连接到副服务器，无需任何操作
- ✅ **服务不中断**: 切换时间 < 30 秒，用户可能只感觉到短暂卡顿
- ✅ **数据不丢失**: 实时同步，RPO（数据丢失时间）< 1 秒

### 管理员体验
- ✅ **可视化监控**: 通过 Grafana 实时查看所有服务器状态
- ✅ **自动告警**: 出现异常自动发送通知（邮件/短信/微信）
- ✅ **一键切换**: 支持手动切换到副服务器（如维护时）

---

## 🏗️ 架构设计

### 整体架构图

```
                          ┌─────────────────────────┐
                          │   用户访问入口           │
                          │   api.yourdomain.com    │
                          │   (虚拟IP/域名)          │
                          └───────────┬─────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │        Keepalived/DNS 切换         │
                    │     (故障检测 + 自动切换)           │
                    └─────────────────┬─────────────────┘
                                      │
                ┌─────────────────────┴─────────────────────┐
                │                                             │
                ▼                                             ▼
    ┌──────────────────────────┐              ┌──────────────────────────┐
    │   主服务器 (Active)       │              │   副服务器 (Passive)      │
    │   154.37.214.191         │──实时同步──→│   待分配IP                │
    ├──────────────────────────┤              ├──────────────────────────┤
    │ ✅ MySQL 主库 (读写)      │─┐          ┌─│ 🔄 MySQL 从库 (同步)      │
    │ ✅ Redis 主节点           │ │          │ │ 🔄 Redis 从节点           │
    │ ✅ MinIO 主节点           │ │ 实时复制 │ │ 🔄 MinIO 同步             │
    │ ✅ 后端 API (运行中)      │ │          │ │ ⏸️  后端 API (待命)        │
    │ ✅ 管理后台 (运行中)      │─┘          └─│ ⏸️  管理后台 (待命)        │
    │                          │              │                          │
    │ 🔍 心跳检测 Agent        │◄────心跳────►│ 🔍 心跳检测 Agent        │
    └──────────────────────────┘              └──────────────────────────┘
                │                                             │
                │                                             │
                └─────────────────┬───────────────────────────┘
                                  │ 监控数据收集
                                  ▼
                    ┌──────────────────────────┐
                    │   监控服务器 (Monitor)    │
                    │   独立IP                  │
                    ├──────────────────────────┤
                    │ 📊 Prometheus            │
                    │    - 收集监控指标        │
                    │    - 评估服务健康        │
                    │                          │
                    │ 📈 Grafana               │
                    │    - 可视化仪表板        │
                    │    - 历史数据查询        │
                    │                          │
                    │ 🔔 Alertmanager          │
                    │    - 告警规则引擎        │
                    │    - 通知发送            │
                    │                          │
                    │ 📝 ELK Stack (可选)      │
                    │    - 日志收集分析        │
                    └──────────────────────────┘
                                  │
                                  ▼
                    ┌──────────────────────────┐
                    │    管理员监控界面         │
                    │    http://monitor:3000   │
                    └──────────────────────────┘
```

---

## 🔄 工作原理

### 正常状态（主服务器工作）

```
1. 用户请求流程:
   用户 → 域名(api.yourdomain.com) → 主服务器IP → 主服务器服务

2. 数据同步流程:
   主服务器写入数据 → 实时复制 → 副服务器保存
   
3. 监控流程:
   主服务器 → 发送指标 → Prometheus → Grafana 展示
   副服务器 → 发送指标 → Prometheus → Grafana 展示
   
4. 心跳检测:
   主服务器 ←→ 每秒互相检测 ←→ 副服务器
```

**副服务器状态**:
- ✅ MySQL 从库：实时同步数据（只读模式）
- ✅ Redis 从节点：实时同步数据
- ✅ MinIO：实时同步文件
- ⏸️  后端 API：容器运行但不对外提供服务
- ⏸️  Nginx：未启动或不监听外部端口

**关键配置**：副服务器数据库设置为只读，防止误写入
```sql
-- 副服务器 MySQL 配置
SET GLOBAL read_only = ON;
SET GLOBAL super_read_only = ON;
```

---

### 故障切换（主服务器宕机）

```
故障发生:
1. 主服务器宕机（硬件故障/网络中断/系统崩溃）
   ↓
2. 心跳检测失败（连续3次，约3秒）
   ↓
3. Keepalived/监控系统确认主服务器不可用
   ↓
4. 自动触发切换流程:
   
   4.1 更新 DNS 或虚拟 IP
       api.yourdomain.com: 主服务器IP → 副服务器IP
   
   4.2 副服务器接管:
       - MySQL 从库 → 提升为主库（关闭只读模式）
       - Redis 从节点 → 提升为主节点
       - 后端 API → 启动并对外提供服务
       - Nginx → 启动并监听 80/443 端口
   
   4.3 发送告警通知:
       → 管理员邮箱/手机短信/微信
       → 告知: "主服务器故障，已自动切换到副服务器"
   ↓
5. 用户自动重新连接到副服务器（无感知）
   - 浏览器 DNS 缓存刷新（1-5分钟）
   - WebSocket 自动重连
   - 移动端自动重连
   ↓
6. 服务恢复正常
```

**时间线**:
```
0秒   - 主服务器宕机
3秒   - 检测到故障
5秒   - 开始切换
10秒  - DNS/IP 更新完成
30秒  - 副服务器完全接管
60秒  - 大部分用户已重连
5分钟 - 所有用户完成切换
```

**用户体验**:
- 🟢 正在聊天的用户：可能看到"连接中断，正在重连..."，3-10秒后恢复
- 🟢 刚发送消息的用户：消息可能显示"发送中"，重连后自动完成
- 🟢 新访问的用户：直接连接到副服务器，无任何异常

---

### 故障恢复（主服务器修复后）

```
主服务器修复完成:
1. 管理员确认主服务器已修复
   ↓
2. 将主服务器降级为"从服务器"
   - MySQL：配置为从库，从当前副服务器同步
   - Redis：配置为从节点
   ↓
3. 等待数据完全同步（1-30分钟，取决于数据量）
   ↓
4. 管理员决定是否切换回主服务器:
   
   选项 A: 继续使用副服务器作为主服务器
   ✅ 优点：不影响用户
   ❌ 缺点：副服务器可能性能较弱
   
   选项 B: 手动切换回主服务器
   ✅ 优点：恢复原有架构
   ❌ 缺点：需要再次切换（30秒中断）
   
   推荐：晚上低峰期手动切换回主服务器
```

---

## 🛠️ 技术实现方案

### 方案 1: 使用 Keepalived + 虚拟 IP（推荐）

**适用场景**: 主服务器和副服务器在同一机房/同一内网  
**优点**: 切换速度快（< 5秒），完全自动化  
**缺点**: 需要同一局域网

#### 架构说明
```
虚拟 IP: 10.0.0.100 (VIP)
主服务器: 10.0.0.10 (优先级 100)
副服务器: 10.0.0.20 (优先级 90)

正常时: VIP → 主服务器 (10.0.0.10)
故障时: VIP → 副服务器 (10.0.0.20)
```

#### 配置步骤

**主服务器配置**:
```bash
# 1. 安装 Keepalived
apt install keepalived -y

# 2. 配置
cat > /etc/keepalived/keepalived.conf << 'EOF'
global_defs {
    router_id MASTER_NODE
    script_user root
    enable_script_security
}

# 健康检查脚本
vrrp_script check_backend {
    script "/etc/keepalived/check_backend.sh"
    interval 2    # 每2秒检查一次
    timeout 3     # 超时时间3秒
    weight -50    # 检查失败，优先级降低50
    fall 3        # 连续3次失败才判定为故障
    rise 2        # 连续2次成功才判定为恢复
}

vrrp_instance VI_1 {
    state MASTER
    interface eth0    # 网络接口，使用 ip addr 查看
    virtual_router_id 51
    priority 100      # 主服务器优先级
    advert_int 1      # 通告间隔1秒
    
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
    
    # 切换到主状态时执行
    notify_master "/etc/keepalived/notify.sh MASTER"
    
    # 切换到备状态时执行
    notify_backup "/etc/keepalived/notify.sh BACKUP"
    
    # 出现故障时执行
    notify_fault "/etc/keepalived/notify.sh FAULT"
}
EOF

# 3. 创建健康检查脚本
cat > /etc/keepalived/check_backend.sh << 'EOF'
#!/bin/bash

# 检查后端 API 是否正常
if curl -sf http://localhost:8080/health > /dev/null; then
    exit 0  # 正常
else
    exit 1  # 异常
fi
EOF

chmod +x /etc/keepalived/check_backend.sh

# 4. 创建通知脚本
cat > /etc/keepalived/notify.sh << 'EOF'
#!/bin/bash

TYPE=$1
HOST=$(hostname)
VIP="10.0.0.100"

case $TYPE in
    MASTER)
        echo "$(date) - $HOST 切换为 MASTER" >> /var/log/keepalived-notify.log
        # 发送通知给管理员
        curl -X POST "https://your-webhook-url" \
          -d "{\"text\":\"✅ $HOST 切换为主服务器，接管虚拟IP $VIP\"}"
        ;;
    BACKUP)
        echo "$(date) - $HOST 切换为 BACKUP" >> /var/log/keepalived-notify.log
        curl -X POST "https://your-webhook-url" \
          -d "{\"text\":\"⚠️ $HOST 切换为备份服务器\"}"
        ;;
    FAULT)
        echo "$(date) - $HOST 出现故障" >> /var/log/keepalived-notify.log
        curl -X POST "https://your-webhook-url" \
          -d "{\"text\":\"🔴 $HOST 出现故障！\"}"
        ;;
esac
EOF

chmod +x /etc/keepalived/notify.sh

# 5. 启动 Keepalived
systemctl enable keepalived
systemctl start keepalived

# 6. 验证
ip addr show  # 应该能看到虚拟 IP
```

**副服务器配置**:
```bash
# 配置几乎相同，只修改以下部分：
cat > /etc/keepalived/keepalived.conf << 'EOF'
global_defs {
    router_id BACKUP_NODE  # 修改为 BACKUP
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
    state BACKUP          # 修改为 BACKUP
    interface eth0
    virtual_router_id 51  # 必须与主服务器相同
    priority 90           # 修改为 90（低于主服务器）
    advert_int 1
    
    authentication {
        auth_type PASS
        auth_pass ZhiHang2024!  # 必须与主服务器相同
    }
    
    virtual_ipaddress {
        10.0.0.100/24  # 必须与主服务器相同
    }
    
    track_script {
        check_backend
    }
    
    notify_master "/etc/keepalived/notify.sh MASTER"
    notify_backup "/etc/keepalived/notify.sh BACKUP"
    notify_fault "/etc/keepalived/notify.sh FAULT"
}
EOF

# 其他脚本与主服务器相同
# 启动 Keepalived
systemctl enable keepalived
systemctl start keepalived
```

---

### 方案 2: 使用 DNS 故障转移（异地部署）

**适用场景**: 主服务器和副服务器在不同机房/不同城市  
**优点**: 支持异地容灾，成本低  
**缺点**: 切换较慢（1-5分钟，DNS 缓存）

#### 配置步骤（使用阿里云DNS）

```bash
# 1. 配置阿里云 DNS 健康检查
登录阿里云控制台 → 云解析 DNS → 健康检查

添加健康检查规则:
- 域名: api.yourdomain.com
- 主服务器IP: 154.37.214.191
- 副服务器IP: 副服务器IP
- 检查频率: 1分钟
- 检查协议: HTTP
- 检查URL: /health
- 预期响应: 200

# 2. 配置故障转移策略
默认解析: 主服务器IP
备用解析: 副服务器IP
切换条件: 主服务器健康检查失败

# 3. 设置 TTL
TTL: 60秒（减少缓存时间，加快切换）
```

---

### 方案 3: 使用云负载均衡（最简单）

**适用场景**: 不想自己配置，愿意付费  
**优点**: 最简单，最可靠，云厂商负责维护  
**缺点**: 有成本（约200-500元/月）

#### 配置步骤（阿里云 SLB）

```bash
# 1. 购买阿里云 SLB（负载均衡）
规格: 标准型
带宽: 按流量计费
地区: 与主服务器相同

# 2. 配置后端服务器池
添加后端服务器:
- 主服务器: 154.37.214.191:8080 (权重 100)
- 副服务器: 副服务器IP:8080 (权重 0)  # 权重0表示不分配流量

# 3. 配置健康检查
协议: HTTP
检查URL: /health
健康阈值: 2次
不健康阈值: 3次
检查间隔: 2秒

# 4. 配置故障转移
当主服务器健康检查失败时:
- 自动将主服务器权重改为 0
- 自动将副服务器权重改为 100
- 所有流量切换到副服务器

# 5. 域名解析
api.yourdomain.com → SLB 公网IP
```

---

## 📊 监控服务器配置

### 目标
- ✅ 实时监控主服务器和副服务器的健康状态
- ✅ 可视化展示系统资源使用情况
- ✅ 自动告警并通知管理员

### 服务器配置

```
监控服务器配置:
CPU: 4核
内存: 8GB
硬盘: 50GB SSD
带宽: 5Mbps

部署服务:
- Prometheus (监控数据收集)
- Grafana (可视化仪表板)
- Alertmanager (告警管理)
- Node Exporter (主备服务器安装，收集系统指标)
```

### 完整部署方案

**监控服务器部署**:
```yaml
# docker-compose.monitoring.yml
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
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - ./alerts.yml:/etc/prometheus/alerts.yml
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
      - ./alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
    networks:
      - monitoring

  # Grafana - 可视化面板
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=Admin_Monitor_2024
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-worldmap-panel
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    networks:
      - monitoring

volumes:
  prometheus_data:
  alertmanager_data:
  grafana_data:

networks:
  monitoring:
    driver: bridge
```

**Prometheus 配置**:
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
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

# 监控目标
scrape_configs:
  # 主服务器监控
  - job_name: 'master-server'
    static_configs:
      - targets: ['154.37.214.191:9100']  # Node Exporter
        labels:
          role: 'master'
          server: 'primary'
    metrics_path: '/metrics'
    
  # 主服务器 - MySQL
  - job_name: 'master-mysql'
    static_configs:
      - targets: ['154.37.214.191:9104']  # MySQL Exporter
        labels:
          role: 'database'
          server: 'master'
  
  # 主服务器 - Redis
  - job_name: 'master-redis'
    static_configs:
      - targets: ['154.37.214.191:9121']  # Redis Exporter
        labels:
          role: 'cache'
          server: 'master'
  
  # 主服务器 - 后端 API
  - job_name: 'master-backend'
    static_configs:
      - targets: ['154.37.214.191:8080']
        labels:
          role: 'backend'
          server: 'master'
    metrics_path: '/metrics'

  # 副服务器监控（配置相同，IP 和标签不同）
  - job_name: 'backup-server'
    static_configs:
      - targets: ['BACKUP_SERVER_IP:9100']
        labels:
          role: 'backup'
          server: 'secondary'
```

**告警规则配置**:
```yaml
# alerts.yml
groups:
  - name: server_alerts
    interval: 30s
    rules:
      # 主服务器宕机告警
      - alert: MasterServerDown
        expr: up{server="master"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "🔴 主服务器宕机！"
          description: "主服务器 {{ $labels.instance }} 已宕机超过1分钟，请立即检查！"

      # 副服务器宕机告警
      - alert: BackupServerDown
        expr: up{server="backup"} == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "⚠️ 副服务器宕机"
          description: "副服务器 {{ $labels.instance }} 已宕机超过5分钟"

      # MySQL 主从复制延迟
      - alert: MySQLReplicationLag
        expr: mysql_slave_status_seconds_behind_master > 60
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "MySQL 主从复制延迟"
          description: "MySQL 从库延迟 {{ $value }} 秒"

      # 磁盘空间不足
      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100 < 15
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "磁盘空间不足"
          description: "{{ $labels.instance }} 磁盘剩余空间低于 15%"
```

**Alertmanager 配置（邮件通知）**:
```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'smtp.qq.com:587'
  smtp_from: 'your-email@qq.com'
  smtp_auth_username: 'your-email@qq.com'
  smtp_auth_password: 'your-smtp-password'
  smtp_require_tls: true

route:
  receiver: 'admin'
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

  routes:
    # 严重告警立即发送
    - match:
        severity: critical
      receiver: 'admin-urgent'
      repeat_interval: 15m

receivers:
  - name: 'admin'
    email_configs:
      - to: 'admin@yourdomain.com'
        headers:
          Subject: '【志航密信】监控告警'
        html: |
          {{ range .Alerts }}
          <h3>{{ .Labels.alertname }}</h3>
          <p>{{ .Annotations.description }}</p>
          <p>时间: {{ .StartsAt.Format "2006-01-02 15:04:05" }}</p>
          {{ end }}

  - name: 'admin-urgent'
    email_configs:
      - to: 'admin@yourdomain.com'
        headers:
          Subject: '【紧急】志航密信严重告警！'
        html: |
          <h2 style="color:red;">严重告警！</h2>
          {{ range .Alerts }}
          <h3>{{ .Labels.alertname }}</h3>
          <p>{{ .Annotations.description }}</p>
          <p>时间: {{ .StartsAt.Format "2006-01-02 15:04:05" }}</p>
          {{ end }}
```

---

## 💰 成本估算

### 服务器成本

| 服务器 | 配置 | 用途 | 月成本 |
|--------|------|------|--------|
| **主服务器** | 8核16GB 100GB SSD | 日常提供服务 | 800 元 |
| **副服务器** | 8核16GB 100GB SSD | 实时备份+故障接管 | 800 元 |
| **监控服务器** | 4核8GB 50GB SSD | 监控+告警 | 400 元 |
| **合计** | - | - | **2000 元/月** |

### 额外成本（可选）

| 项目 | 月成本 | 说明 |
|------|--------|------|
| DNS 服务 | 0-50 元 | 阿里云DNS（免费版够用） |
| 云负载均衡 | 200-500 元 | 如果使用云SLB |
| 备份存储 | 50-100 元 | OSS 对象存储 |
| CDN 加速 | 100-300 元 | 静态资源加速 |
| 短信通知 | 10-50 元 | 告警短信通知 |

---

## 📅 实施计划

### 第 1 周：准备阶段
- [ ] Day 1: 购买副服务器和监控服务器
- [ ] Day 2-3: 配置副服务器基础环境（Docker, 网络）
- [ ] Day 4-5: 配置监控服务器（Prometheus, Grafana）
- [ ] Day 6-7: 测试服务器间网络连通性

### 第 2 周：数据同步配置
- [ ] Day 1-2: 配置 MySQL 主从复制
- [ ] Day 3: 配置 Redis 主从复制
- [ ] Day 4: 配置 MinIO 数据同步
- [ ] Day 5: 验证数据同步完整性
- [ ] Day 6-7: 压力测试数据同步性能

### 第 3 周：故障转移配置
- [ ] Day 1-2: 部署 Keepalived 或配置 DNS
- [ ] Day 3-4: 配置健康检查和自动切换
- [ ] Day 5: 编写故障通知脚本
- [ ] Day 6-7: 故障切换演练（模拟主服务器宕机）

### 第 4 周：监控和优化
- [ ] Day 1-2: 配置 Grafana 仪表板
- [ ] Day 3-4: 配置告警规则和通知
- [ ] Day 5: 编写运维文档
- [ ] Day 6: 团队培训（管理员培训）
- [ ] Day 7: 正式上线高可用架构

---

## 🎯 验收标准

### 功能验收
- [ ] ✅ 主服务器正常提供服务
- [ ] ✅ 副服务器实时同步数据（延迟 < 1秒）
- [ ] ✅ 监控系统正常收集和展示数据
- [ ] ✅ 故障切换时间 < 30秒
- [ ] ✅ 切换后用户无感知或短暂重连
- [ ] ✅ 告警通知正常发送

### 性能验收
- [ ] ✅ 主服务器响应时间 < 100ms
- [ ] ✅ 副服务器待命状态资源占用 < 30%
- [ ] ✅ 监控系统资源占用 < 20%

### 可靠性验收
- [ ] ✅ 模拟主服务器宕机，自动切换成功
- [ ] ✅ 模拟网络中断，告警正常发送
- [ ] ✅ 连续运行 7 天无故障

---

## 📞 给 Devin 的部署指令（阶段2启动时）

部署副服务器和监控服务器的完整指令我已经准备好了，当您决定启动阶段2时，我会生成详细的执行脚本。

**当前阶段（阶段1）**：
- ✅ 继续使用单机部署
- ✅ 立即部署自动备份系统
- ✅ 配置基础监控告警

**下一步（阶段2启动条件）**：
- 注册用户 > 100
- 日活跃用户 > 50
- 出现过停机事故
- 您决定正式商业运营

---

**总结**：
- ✅ **现在**: 单机 + 自动备份（已完成）
- ⏰ **1-2个月后**: 主备架构（本文档方案）
- 🚀 **长期**: 可根据需要扩展为多地多活

这个架构完全符合您的需求：主服务器日常服务，副服务器实时备份，故障时用户无感切换！🎉

