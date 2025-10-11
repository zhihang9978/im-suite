#!/bin/bash

###############################################################################
# 志航密信 - SSL/HTTPS证书配置脚本
# 用途：使用Let's Encrypt申请和配置SSL证书
# 使用：sudo bash ops/setup-ssl.sh
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
log_info "SSL/HTTPS证书配置"
log_info "========================================="

# ============================================
# 1. 选择证书类型
# ============================================
echo "请选择证书类型："
echo "  1) Let's Encrypt 免费证书（推荐）"
echo "  2) 自签名证书（仅用于测试）"
echo "  3) 导入已有证书"
read -p "请选择 [1-3]: " CERT_TYPE

case $CERT_TYPE in
    1)
        # Let's Encrypt
        log_info "使用Let's Encrypt免费证书"
        
        read -p "请输入域名 (例：im.example.com): " DOMAIN
        read -p "请输入邮箱地址: " EMAIL
        
        # 安装certbot
        log_info "安装certbot..."
        if command -v apt-get &> /dev/null; then
            apt-get update
            apt-get install -y certbot python3-certbot-nginx
        elif command -v yum &> /dev/null; then
            yum install -y certbot python3-certbot-nginx
        fi
        
        # 检查Nginx是否运行
        if systemctl is-active --quiet nginx; then
            log_warning "Nginx正在运行，将临时停止以申请证书"
            systemctl stop nginx
            RESTART_NGINX=1
        fi
        
        # 申请证书（standalone模式）
        log_info "申请证书..."
        certbot certonly --standalone \
            --preferred-challenges http \
            --email $EMAIL \
            --agree-tos \
            --no-eff-email \
            -d $DOMAIN
        
        CERT_PATH="/etc/letsencrypt/live/${DOMAIN}/fullchain.pem"
        KEY_PATH="/etc/letsencrypt/live/${DOMAIN}/privkey.pem"
        
        # 配置自动续期
        log_info "配置自动续期..."
        cat > /etc/cron.d/certbot-renew <<EOF
# Let's Encrypt 自动续期
0 3 * * * root certbot renew --quiet --deploy-hook "systemctl reload nginx"
EOF
        
        log_success "Let's Encrypt证书申请成功"
        log_info "证书将在60天后自动续期"
        ;;
        
    2)
        # 自签名证书
        log_warning "生成自签名证书（仅用于测试，不安全）"
        
        read -p "请输入域名 (例：im.example.com): " DOMAIN
        
        CERT_DIR="/opt/im-suite/ssl"
        mkdir -p $CERT_DIR
        
        # 生成私钥和证书
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ${CERT_DIR}/selfsigned.key \
            -out ${CERT_DIR}/selfsigned.crt \
            -subj "/C=CN/ST=State/L=City/O=Organization/CN=${DOMAIN}"
        
        CERT_PATH="${CERT_DIR}/selfsigned.crt"
        KEY_PATH="${CERT_DIR}/selfsigned.key"
        
        log_success "自签名证书生成成功"
        log_warning "浏览器将显示安全警告，这是正常的"
        ;;
        
    3)
        # 导入证书
        log_info "导入已有证书"
        
        read -p "请输入证书文件路径 (.crt或.pem): " CERT_FILE
        read -p "请输入私钥文件路径 (.key): " KEY_FILE
        
        if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
            log_error "证书或私钥文件不存在"
            exit 1
        fi
        
        CERT_DIR="/opt/im-suite/ssl"
        mkdir -p $CERT_DIR
        
        cp "$CERT_FILE" "${CERT_DIR}/certificate.crt"
        cp "$KEY_FILE" "${CERT_DIR}/private.key"
        
        CERT_PATH="${CERT_DIR}/certificate.crt"
        KEY_PATH="${CERT_DIR}/private.key"
        
        log_success "证书导入成功"
        ;;
        
    *)
        log_error "无效的选择"
        exit 1
        ;;
esac

# ============================================
# 2. 配置Nginx
# ============================================
log_info "配置Nginx HTTPS..."

# 检查Nginx是否安装
if ! command -v nginx &> /dev/null; then
    log_info "安装Nginx..."
    if command -v apt-get &> /dev/null; then
        apt-get install -y nginx
    elif command -v yum &> /dev/null; then
        yum install -y nginx
    fi
fi

# 生成Nginx配置
cat > /etc/nginx/sites-available/im-suite <<EOF
# ========================================
# 志航密信 - Nginx HTTPS配置
# ========================================

# HTTP -> HTTPS 重定向
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};
    
    # Let's Encrypt验证
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }
    
    # 其他请求重定向到HTTPS
    location / {
        return 301 https://\$host\$request_uri;
    }
}

# HTTPS主配置
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};
    
    # SSL证书
    ssl_certificate ${CERT_PATH};
    ssl_certificate_key ${KEY_PATH};
    
    # SSL配置（Mozilla现代配置）
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers off;
    
    # SSL会话缓存
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_session_tickets off;
    
    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;
    
    # 安全头
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' https:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';" always;
    
    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss application/rss+xml font/truetype font/opentype application/vnd.ms-fontobject image/svg+xml;
    
    # 后端API反向代理
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # WebSocket
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # WebSocket超时
        proxy_read_timeout 3600s;
    }
    
    # 管理后台
    location /admin/ {
        proxy_pass http://localhost:3001/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # 静态文件
    location /static/ {
        alias /opt/im-suite/static/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
    
    # 健康检查
    location /health {
        proxy_pass http://localhost:8080/health;
        access_log off;
    }
    
    # 默认location
    location / {
        root /var/www/html;
        index index.html index.htm;
        try_files \$uri \$uri/ =404;
    }
}
EOF

# 启用站点配置
if [ -d "/etc/nginx/sites-enabled" ]; then
    ln -sf /etc/nginx/sites-available/im-suite /etc/nginx/sites-enabled/
else
    # CentOS/RHEL
    ln -sf /etc/nginx/sites-available/im-suite /etc/nginx/conf.d/im-suite.conf
fi

# 测试Nginx配置
log_info "测试Nginx配置..."
if nginx -t; then
    log_success "Nginx配置测试通过"
else
    log_error "Nginx配置有误"
    exit 1
fi

# 启动/重启Nginx
systemctl enable nginx
systemctl restart nginx

log_success "Nginx HTTPS配置完成"

# ============================================
# 3. 验证SSL配置
# ============================================
log_info "验证SSL配置..."

sleep 2

# 检查端口
if netstat -tulpn | grep -q ":443"; then
    log_success "✓ HTTPS端口 443 监听中"
else
    log_error "✗ 端口 443 未监听"
fi

# 测试SSL证书
log_info "测试SSL证书..."
if curl -k -s -o /dev/null -w "%{http_code}" https://localhost | grep -q "200\|301\|302"; then
    log_success "✓ HTTPS服务响应正常"
else
    log_warning "⚠ HTTPS服务可能未正常启动"
fi

# ============================================
# 完成
# ============================================
log_success "========================================="
log_success "SSL/HTTPS配置完成！"
log_success "========================================="
echo ""
echo "配置信息："
echo "  域名: ${DOMAIN}"
echo "  证书路径: ${CERT_PATH}"
echo "  私钥路径: ${KEY_PATH}"
echo ""
echo "访问地址："
echo "  前端: https://${DOMAIN}/"
echo "  后端API: https://${DOMAIN}/api/"
echo "  管理后台: https://${DOMAIN}/admin/"
echo ""
echo "在线测试SSL："
echo "  https://www.ssllabs.com/ssltest/analyze.html?d=${DOMAIN}"
echo ""
echo "查看Nginx日志："
echo "  tail -f /var/log/nginx/access.log"
echo "  tail -f /var/log/nginx/error.log"
echo ""

if [ "$CERT_TYPE" = "1" ]; then
    echo "证书续期："
    echo "  自动续期已配置（每天凌晨3点）"
    echo "  手动续期: certbot renew"
    echo ""
fi

