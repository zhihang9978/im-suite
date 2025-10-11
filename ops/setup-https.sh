#!/bin/bash

###############################################################################
# 志航密信 - HTTPS/TLS自动配置脚本
# 用途：配置Let's Encrypt免费SSL证书
# 使用：sudo bash ops/setup-https.sh your-domain.com
###############################################################################

set -e

if [ $# -eq 0 ]; then
    echo "用法: sudo bash ops/setup-https.sh your-domain.com"
    exit 1
fi

DOMAIN=$1
EMAIL=${2:-"admin@${DOMAIN}"}

echo "========================================="
echo "HTTPS/TLS配置"
echo "域名: $DOMAIN"
echo "邮箱: $EMAIL"
echo "========================================="

# 检查是否为root
if [ "$EUID" -ne 0 ]; then
    echo "❌ 请使用sudo运行此脚本"
    exit 1
fi

# 安装certbot
if ! command -v certbot &> /dev/null; then
    echo "📦 安装certbot..."
    apt-get update
    apt-get install -y certbot python3-certbot-nginx
fi

# 检查Nginx
if ! command -v nginx &> /dev/null; then
    echo "📦 安装Nginx..."
    apt-get install -y nginx
fi

# 停止Nginx（certbot需要80端口）
systemctl stop nginx || true

# 获取证书
echo "🔐 申请Let's Encrypt证书..."
certbot certonly --standalone \
    -d $DOMAIN \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --force-renewal

# 配置Nginx HTTPS
echo "⚙️ 配置Nginx HTTPS..."
cat > /etc/nginx/sites-available/im-suite-https << EOF
# HTTP重定向到HTTPS
server {
    listen 80;
    server_name $DOMAIN;
    return 301 https://\$server_name\$request_uri;
}

# HTTPS服务器
server {
    listen 443 ssl http2;
    server_name $DOMAIN;

    # SSL证书
    ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/$DOMAIN/chain.pem;

    # SSL配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # HSTS
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # CSP
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' wss://$DOMAIN" always;

    # 后端API代理
    location /api {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # WebSocket代理
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # WebSocket超时
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }

    # Web客户端
    location / {
        root /var/www/im-client;
        index index.html;
        try_files \$uri \$uri/ /index.html;
        
        # 静态资源缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # 管理后台
    location /admin {
        alias /var/www/im-admin;
        index index.html;
        try_files \$uri \$uri/ /admin/index.html;
    }

    # 健康检查
    location /health {
        proxy_pass http://127.0.0.1:8080;
        access_log off;
    }

    # 禁止访问隐藏文件
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
}
EOF

# 启用站点
ln -sf /etc/nginx/sites-available/im-suite-https /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# 测试Nginx配置
nginx -t

# 启动Nginx
systemctl start nginx
systemctl enable nginx

# 设置自动续期
echo "🔄 设置证书自动续期..."
(crontab -l 2>/dev/null; echo "0 0,12 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'") | crontab -

echo ""
echo "✅ HTTPS配置完成！"
echo ""
echo "证书位置: /etc/letsencrypt/live/$DOMAIN/"
echo "Nginx配置: /etc/nginx/sites-available/im-suite-https"
echo ""
echo "访问地址:"
echo "  - Web客户端: https://$DOMAIN"
echo "  - 管理后台: https://$DOMAIN/admin"
echo "  - API端点: https://$DOMAIN/api"
echo "  - WebSocket: wss://$DOMAIN/ws"
echo ""
echo "自动续期: 已配置cron任务（每天0点和12点检查）"
echo ""

