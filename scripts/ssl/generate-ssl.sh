#!/bin/bash

# 志航密信 SSL 证书生成脚本
# 用于生成自签名证书，适用于开发环境

set -e

# 配置变量
DOMAIN="localhost"
CERT_DIR="./scripts/ssl"
DAYS=365

echo "🔐 开始生成 SSL 证书..."

# 创建 SSL 目录
mkdir -p "$CERT_DIR"

# 生成私钥
echo "📝 生成私钥..."
openssl genrsa -out "$CERT_DIR/zhihang-messenger.key" 2048

# 生成证书签名请求
echo "📝 生成证书签名请求..."
openssl req -new -key "$CERT_DIR/zhihang-messenger.key" -out "$CERT_DIR/zhihang-messenger.csr" \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=Zhihang Messenger/OU=IT Department/CN=$DOMAIN"

# 生成自签名证书
echo "📝 生成自签名证书..."
openssl x509 -req -days $DAYS -in "$CERT_DIR/zhihang-messenger.csr" \
    -signkey "$CERT_DIR/zhihang-messenger.key" -out "$CERT_DIR/zhihang-messenger.crt" \
    -extensions v3_req -extfile <(
        echo "[req]"
        echo "distinguished_name = req_distinguished_name"
        echo "req_extensions = v3_req"
        echo "prompt = no"
        echo "[req_distinguished_name]"
        echo "C = CN"
        echo "ST = Beijing"
        echo "L = Beijing"
        echo "O = Zhihang Messenger"
        echo "OU = IT Department"
        echo "CN = $DOMAIN"
        echo "[v3_req]"
        echo "keyUsage = keyEncipherment, dataEncipherment"
        echo "extendedKeyUsage = serverAuth"
        echo "subjectAltName = @alt_names"
        echo "[alt_names]"
        echo "DNS.1 = $DOMAIN"
        echo "DNS.2 = localhost"
        echo "IP.1 = 127.0.0.1"
        echo "IP.2 = ::1"
    )

# 设置权限
chmod 600 "$CERT_DIR/zhihang-messenger.key"
chmod 644 "$CERT_DIR/zhihang-messenger.crt"

# 清理临时文件
rm -f "$CERT_DIR/zhihang-messenger.csr"

echo "✅ SSL 证书生成完成！"
echo "📁 证书文件位置:"
echo "   私钥: $CERT_DIR/zhihang-messenger.key"
echo "   证书: $CERT_DIR/zhihang-messenger.crt"
echo ""
echo "⚠️  注意: 这是自签名证书，浏览器会显示安全警告"
echo "   在开发环境中可以点击 '高级' -> '继续访问'"
echo ""
echo "🚀 生产环境建议使用宝塔面板自动获取 Let's Encrypt 证书"


