#!/bin/bash

###############################################################################
# 志航密信 - TURN服务器安装配置脚本
# 用途：安装和配置coturn TURN服务器
# 使用：sudo bash ops/setup-turn.sh
###############################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查root权限
if [[ $EUID -ne 0 ]]; then
   log_error "此脚本需要root权限运行"
   exit 1
fi

log_info "========================================="
log_info "TURN服务器安装配置"
log_info "========================================="

# ============================================
# 1. 安装coturn
# ============================================
log_info "1. 安装coturn..."

if command -v apt-get &> /dev/null; then
    apt-get update
    apt-get install -y coturn
elif command -v yum &> /dev/null; then
    yum install -y epel-release
    yum install -y coturn
else
    log_error "不支持的包管理器"
    exit 1
fi

log_success "coturn安装完成: $(turnserver --version 2>&1 | head -1)"

# ============================================
# 2. 配置参数
# ============================================
log_info "2. 配置TURN服务器..."

# 获取配置
read -p "请输入域名 (例：example.com): " DOMAIN
read -p "请输入公网IP（多个用逗号分隔）: " PUBLIC_IPS
read -p "请输入TURN用户名 [turn_user]: " TURN_USER
TURN_USER=${TURN_USER:-turn_user}
read -sp "请输入TURN密码: " TURN_PASSWORD
echo ""

# 转换IP为数组
IFS=',' read -ra IP_ARRAY <<< "$PUBLIC_IPS"

# ============================================
# 3. 生成配置文件
# ============================================
log_info "3. 生成coturn配置..."

cat > /etc/turnserver.conf <<EOF
# ========================================
# 志航密信 - coturn配置
# ========================================

# 监听地址
listening-ip=0.0.0.0
relay-ip=$(echo ${IP_ARRAY[0]})

# 外部IP地址
EOF

for ip in "${IP_ARRAY[@]}"; do
    echo "external-ip=$ip" >> /etc/turnserver.conf
done

cat >> /etc/turnserver.conf <<EOF

# 监听端口
listening-port=3478
tls-listening-port=5349

# 端口范围（用于中继）
min-port=49152
max-port=65535

# Realm
realm=${DOMAIN}

# 认证
user=${TURN_USER}:${TURN_PASSWORD}

# 日志
log-file=/var/log/turnserver.log
verbose

# 性能优化
fingerprint
lt-cred-mech

# 安全设置
no-multicast-peers
no-cli
no-loopback-peers
no-tcp-relay

# WebRTC优化
stale-nonce=600
no-stdout-log

# 数据库（可选）
# userdb=/var/lib/turn/turndb

# TLS证书（如果有）
# cert=/etc/letsencrypt/live/${DOMAIN}/cert.pem
# pkey=/etc/letsencrypt/live/${DOMAIN}/privkey.pem

# 性能调优
max-allocate-lifetime=3600
max-allocate-timeout=60
EOF

log_success "coturn配置已生成: /etc/turnserver.conf"

# ============================================
# 4. 启用coturn
# ============================================
log_info "4. 启用coturn服务..."

# 修改systemd配置以启用coturn
sed -i 's/#TURNSERVER_ENABLED=1/TURNSERVER_ENABLED=1/' /etc/default/coturn 2>/dev/null || true

# 启动服务
systemctl enable coturn
systemctl restart coturn

sleep 2

if systemctl is-active --quiet coturn; then
    log_success "coturn服务运行正常"
else
    log_error "coturn服务启动失败"
    log_info "查看日志: journalctl -u coturn -n 50"
    exit 1
fi

# ============================================
# 5. 防火墙配置
# ============================================
log_info "5. 配置防火墙..."

if command -v ufw &> /dev/null; then
    ufw allow 3478/tcp comment "TURN TCP"
    ufw allow 3478/udp comment "TURN UDP"
    ufw allow 5349/tcp comment "TURN TLS"
    ufw allow 5349/udp comment "TURN DTLS"
    ufw allow 49152:65535/udp comment "TURN Relay"
    log_success "UFW规则已添加"
elif command -v firewall-cmd &> /dev/null; then
    firewall-cmd --permanent --add-port=3478/tcp
    firewall-cmd --permanent --add-port=3478/udp
    firewall-cmd --permanent --add-port=5349/tcp
    firewall-cmd --permanent --add-port=5349/udp
    firewall-cmd --permanent --add-port=49152-65535/udp
    firewall-cmd --reload
    log_success "firewalld规则已添加"
else
    log_warning "请手动配置防火墙"
fi

# ============================================
# 6. 测试配置
# ============================================
log_info "6. 测试TURN服务器..."

# 检查端口监听
if netstat -tulpn | grep -q ":3478"; then
    log_success "✓ TURN TCP/UDP端口 3478 监听中"
else
    log_error "✗ 端口 3478 未监听"
fi

if netstat -tulpn | grep -q ":5349"; then
    log_success "✓ TURN TLS端口 5349 监听中"
else
    log_warning "⚠ 端口 5349 未监听（可能未配置TLS证书）"
fi

# ============================================
# 7. 生成客户端配置
# ============================================
log_info "7. 生成客户端配置..."

cat > /opt/im-suite/turn-config.json <<EOF
{
  "iceServers": [
    {
      "urls": "stun:${DOMAIN}:3478"
    },
    {
      "urls": "turn:${DOMAIN}:3478",
      "username": "${TURN_USER}",
      "credential": "${TURN_PASSWORD}"
    },
    {
      "urls": "turn:${DOMAIN}:3478?transport=tcp",
      "username": "${TURN_USER}",
      "credential": "${TURN_PASSWORD}"
    }
  ]
}
EOF

log_success "客户端配置已生成: /opt/im-suite/turn-config.json"

# ============================================
# 8. 创建监控脚本
# ============================================
log_info "8. 创建监控脚本..."

cat > /usr/local/bin/check-turn.sh <<'MONITOR'
#!/bin/bash
# TURN服务器健康检查

if ! systemctl is-active --quiet coturn; then
    echo "$(date): coturn服务已停止，正在重启..." >> /var/log/turn-monitor.log
    systemctl restart coturn
fi

# 检查端口
if ! netstat -tulpn | grep -q ":3478"; then
    echo "$(date): 端口3478未监听" >> /var/log/turn-monitor.log
fi
MONITOR

chmod +x /usr/local/bin/check-turn.sh

# 添加cron任务
(crontab -l 2>/dev/null; echo "*/5 * * * * /usr/local/bin/check-turn.sh") | crontab -

log_success "监控脚本已创建"

# ============================================
# 完成
# ============================================
log_success "========================================="
log_success "TURN服务器配置完成！"
log_success "========================================="
echo ""
echo "配置信息："
echo "  域名: ${DOMAIN}"
echo "  用户名: ${TURN_USER}"
echo "  密码: ${TURN_PASSWORD}"
echo "  监听端口: 3478 (TCP/UDP), 5349 (TLS)"
echo "  中继端口: 49152-65535 (UDP)"
echo ""
echo "测试TURN服务器："
echo "  在线工具: https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/"
echo "  配置: turn:${DOMAIN}:3478"
echo "  用户名: ${TURN_USER}"
echo "  密码: ${TURN_PASSWORD}"
echo ""
echo "查看日志："
echo "  journalctl -u coturn -f"
echo "  tail -f /var/log/turnserver.log"
echo ""
echo "环境变量配置（添加到.env）："
echo "TURN_ENABLED=true"
echo "TURN_SERVER=turn:${DOMAIN}:3478"
echo "TURN_USERNAME=${TURN_USER}"
echo "TURN_PASSWORD=${TURN_PASSWORD}"
echo "TURN_REALM=${DOMAIN}"
echo "TURN_PUBLIC_IPS=${PUBLIC_IPS}"
echo ""

