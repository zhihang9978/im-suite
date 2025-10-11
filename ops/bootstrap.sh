#!/bin/bash

###############################################################################
# 志航密信 - 生产环境系统初始化脚本
# 用途：一键初始化系统依赖、内核参数、安全配置
# 使用：sudo bash ops/bootstrap.sh
###############################################################################

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查root权限
if [[ $EUID -ne 0 ]]; then
   log_error "此脚本需要root权限运行"
   echo "请使用: sudo bash ops/bootstrap.sh"
   exit 1
fi

log_info "========================================="
log_info "志航密信 - 生产环境初始化"
log_info "========================================="

# ============================================
# 1. 检查操作系统
# ============================================
log_info "1. 检查操作系统..."

if [ -f /etc/os-release ]; then
    . /etc/os-release
    log_success "操作系统: $NAME $VERSION"
else
    log_error "无法识别操作系统"
    exit 1
fi

# ============================================
# 2. 安装系统依赖
# ============================================
log_info "2. 安装系统依赖..."

if command -v apt-get &> /dev/null; then
    # Debian/Ubuntu
    apt-get update
    apt-get install -y \
        curl wget git \
        build-essential \
        net-tools iptables \
        certbot python3-certbot-nginx \
        logrotate \
        htop iotop nethogs \
        jq yq
    log_success "Debian/Ubuntu依赖安装完成"
elif command -v yum &> /dev/null; then
    # CentOS/RHEL
    yum install -y epel-release
    yum install -y \
        curl wget git \
        gcc make \
        net-tools iptables \
        certbot python3-certbot-nginx \
        logrotate \
        htop iotop nethogs \
        jq
    log_success "CentOS/RHEL依赖安装完成"
else
    log_error "不支持的包管理器"
    exit 1
fi

# ============================================
# 3. 安装Docker和Docker Compose
# ============================================
log_info "3. 检查Docker..."

if ! command -v docker &> /dev/null; then
    log_warning "Docker未安装，正在安装..."
    curl -fsSL https://get.docker.com | bash
    systemctl enable docker
    systemctl start docker
    log_success "Docker安装完成"
else
    log_success "Docker已安装: $(docker --version)"
fi

if ! command -v docker-compose &> /dev/null; then
    log_warning "Docker Compose未安装，正在安装..."
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    log_success "Docker Compose安装完成"
else
    log_success "Docker Compose已安装: $(docker-compose --version)"
fi

# ============================================
# 4. 内核参数优化
# ============================================
log_info "4. 优化内核参数..."

cat > /etc/sysctl.d/99-im-suite.conf <<EOF
# 志航密信 - 内核参数优化

# ===== 网络性能优化 =====
# 启用BBR拥塞控制
net.core.default_qdisc=fq
net.ipv4.tcp_congestion_control=bbr

# TCP缓冲区大小
net.core.rmem_max=134217728
net.core.wmem_max=134217728
net.core.rmem_default=8388608
net.core.wmem_default=8388608
net.ipv4.tcp_rmem=4096 87380 67108864
net.ipv4.tcp_wmem=4096 65536 67108864

# UDP缓冲区（WebRTC重要）
net.core.optmem_max=65536
net.ipv4.udp_rmem_min=8192
net.ipv4.udp_wmem_min=8192

# 连接队列
net.core.somaxconn=65535
net.core.netdev_max_backlog=65535
net.ipv4.tcp_max_syn_backlog=8192

# TCP优化
net.ipv4.tcp_fin_timeout=30
net.ipv4.tcp_keepalive_time=1200
net.ipv4.tcp_max_tw_buckets=5000
net.ipv4.ip_local_port_range=10000 65000
net.ipv4.tcp_tw_reuse=1
net.ipv4.tcp_slow_start_after_idle=0

# ===== 文件系统优化 =====
fs.file-max=1048576
fs.inotify.max_user_watches=524288

# ===== 安全设置 =====
# SYN Flood防护
net.ipv4.tcp_syncookies=1
net.ipv4.tcp_max_syn_backlog=8192

# 禁用ICMP重定向
net.ipv4.conf.all.accept_redirects=0
net.ipv4.conf.default.accept_redirects=0
net.ipv6.conf.all.accept_redirects=0
net.ipv6.conf.default.accept_redirects=0

# IP转发（NAT穿透）
net.ipv4.ip_forward=1
net.ipv6.conf.all.forwarding=1

# ===== 内存优化 =====
vm.swappiness=10
vm.dirty_ratio=60
vm.dirty_background_ratio=2
EOF

# 应用配置
sysctl -p /etc/sysctl.d/99-im-suite.conf
log_success "内核参数优化完成"

# ============================================
# 5. 文件描述符限制
# ============================================
log_info "5. 配置文件描述符限制..."

cat > /etc/security/limits.d/99-im-suite.conf <<EOF
# 志航密信 - 文件描述符限制

* soft nofile 1048576
* hard nofile 1048576
* soft nproc 65535
* hard nproc 65535
root soft nofile 1048576
root hard nofile 1048576
EOF

log_success "文件描述符限制已设置为 1048576"

# ============================================
# 6. 配置防火墙
# ============================================
log_info "6. 配置防火墙规则..."

if command -v ufw &> /dev/null; then
    # 使用UFW
    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    
    # 基础服务
    ufw allow 22/tcp comment "SSH"
    ufw allow 80/tcp comment "HTTP"
    ufw allow 443/tcp comment "HTTPS"
    
    # WebRTC STUN/TURN
    ufw allow 3478/tcp comment "TURN TCP"
    ufw allow 3478/udp comment "TURN UDP"
    ufw allow 5349/tcp comment "TURN TLS"
    ufw allow 5349/udp comment "TURN DTLS"
    
    # WebRTC UDP端口范围（中继）
    ufw allow 49152:65535/udp comment "WebRTC UDP Range"
    
    # 管理后台（限制IP - 需要手动配置）
    # ufw allow from YOUR_ADMIN_IP to any port 3001 comment "Admin Backend"
    
    ufw --force enable
    log_success "UFW防火墙配置完成"
    
elif command -v firewall-cmd &> /dev/null; then
    # 使用firewalld
    systemctl enable firewalld
    systemctl start firewalld
    
    firewall-cmd --permanent --add-service=ssh
    firewall-cmd --permanent --add-service=http
    firewall-cmd --permanent --add-service=https
    
    # TURN端口
    firewall-cmd --permanent --add-port=3478/tcp
    firewall-cmd --permanent --add-port=3478/udp
    firewall-cmd --permanent --add-port=5349/tcp
    firewall-cmd --permanent --add-port=5349/udp
    
    # WebRTC UDP范围
    firewall-cmd --permanent --add-port=49152-65535/udp
    
    firewall-cmd --reload
    log_success "firewalld防火墙配置完成"
else
    log_warning "未检测到防火墙，请手动配置iptables"
fi

# ============================================
# 7. 配置日志轮转
# ============================================
log_info "7. 配置日志轮转..."

cat > /etc/logrotate.d/im-suite <<EOF
# 志航密信 - 日志轮转配置

/var/log/im-suite/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 root root
    sharedscripts
    postrotate
        docker-compose -f /opt/im-suite/docker-compose.production.yml restart backend > /dev/null 2>&1 || true
    endscript
}

/opt/im-suite/logs/**/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 root root
    maxsize 100M
}
EOF

log_success "日志轮转配置完成（保留30天）"

# ============================================
# 8. 创建必要目录
# ============================================
log_info "8. 创建必要目录..."

mkdir -p /opt/im-suite
mkdir -p /var/log/im-suite
mkdir -p /opt/im-suite/backups/{mysql,redis,minio}
mkdir -p /opt/im-suite/ssl
mkdir -p /opt/im-suite/logs/{backend,admin,nginx}

chmod 755 /opt/im-suite
chmod 750 /opt/im-suite/backups
chmod 750 /opt/im-suite/ssl
chmod 755 /var/log/im-suite

log_success "目录创建完成"

# ============================================
# 9. 配置系统时区
# ============================================
log_info "9. 配置系统时区..."

timedatectl set-timezone Asia/Shanghai
log_success "时区已设置为: $(timedatectl | grep 'Time zone')"

# ============================================
# 10. 安装音视频依赖（coturn）
# ============================================
log_info "10. 安装TURN服务器（coturn）..."

if ! command -v turnserver &> /dev/null; then
    if command -v apt-get &> /dev/null; then
        apt-get install -y coturn
    elif command -v yum &> /dev/null; then
        yum install -y coturn
    fi
    
    if command -v turnserver &> /dev/null; then
        log_success "coturn安装完成"
        systemctl stop coturn  # 暂时停止，等待配置
    else
        log_warning "coturn安装失败，请手动安装"
    fi
else
    log_success "coturn已安装"
fi

# ============================================
# 11. 系统健康检查
# ============================================
log_info "11. 系统健康检查..."

# 检查必要命令
commands=("docker" "docker-compose" "curl" "git" "openssl")
for cmd in "${commands[@]}"; do
    if command -v $cmd &> /dev/null; then
        log_success "✓ $cmd 已安装"
    else
        log_error "✗ $cmd 未安装"
    fi
done

# 检查端口占用
ports=(80 443 3478 5349 8080 3001 3306 6379 9000)
for port in "${ports[@]}"; do
    if netstat -tulpn | grep -q ":$port "; then
        log_warning "端口 $port 已被占用"
    else
        log_success "✓ 端口 $port 可用"
    fi
done

# 检查磁盘空间
available=$(df -h / | awk 'NR==2 {print $4}' | sed 's/G//')
if (( $(echo "$available < 20" | bc -l) )); then
    log_warning "磁盘可用空间不足20GB"
else
    log_success "✓ 磁盘空间充足: ${available}G"
fi

# 检查内存
mem_total=$(free -g | awk 'NR==2 {print $2}')
if [ $mem_total -lt 4 ]; then
    log_warning "内存小于4GB，建议至少4GB"
else
    log_success "✓ 内存充足: ${mem_total}GB"
fi

# ============================================
# 12. 创建systemd服务文件
# ============================================
log_info "12. 创建systemd服务..."

cat > /etc/systemd/system/im-suite.service <<EOF
[Unit]
Description=Zhihang Messenger Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/im-suite
ExecStart=/usr/local/bin/docker-compose -f docker-compose.production.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.production.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
log_success "systemd服务创建完成"

# ============================================
# 完成
# ============================================
log_success "========================================="
log_success "系统初始化完成！"
log_success "========================================="
echo ""
echo "下一步："
echo "1. 复制项目到 /opt/im-suite/"
echo "2. 配置环境变量 /opt/im-suite/.env"
echo "3. 运行: sudo bash ops/deploy.sh"
echo ""

# 输出系统信息
echo "系统信息："
echo "  操作系统: $NAME $VERSION"
echo "  内核版本: $(uname -r)"
echo "  内存: ${mem_total}GB"
echo "  磁盘: ${available}G 可用"
echo "  BBR状态: $(sysctl net.ipv4.tcp_congestion_control | awk '{print $3}')"
echo "  文件限制: $(ulimit -n)"
echo ""

