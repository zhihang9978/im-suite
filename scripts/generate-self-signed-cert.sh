#!/bin/bash

# 志航密信 - 自签名SSL证书生成脚本
# 用于开发和测试环境

set -e

echo "========================================="
echo "志航密信 - 自签名SSL证书生成"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 创建SSL目录
echo -e "${YELLOW}[1/3] 创建SSL目录...${NC}"
mkdir -p ssl

# 生成自签名证书（有效期365天）
echo -e "${YELLOW}[2/3] 生成自签名SSL证书（有效期365天）...${NC}"
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/key.pem \
  -out ssl/cert.pem \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=ZhiHang/OU=IT/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,DNS:*.localhost,DNS:127.0.0.1,IP:127.0.0.1,IP:::1"

# 设置权限
echo -e "${YELLOW}[3/3] 设置文件权限...${NC}"
chmod 600 ssl/key.pem
chmod 644 ssl/cert.pem

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}✅ 自签名SSL证书已生成${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "证书信息:"
echo "  📄 证书位置: ssl/cert.pem"
echo "  🔑 私钥位置: ssl/key.pem"
echo "  📅 有效期: 365天"
echo "  🌐 支持域名: localhost, *.localhost, 127.0.0.1"
echo ""
echo -e "${YELLOW}⚠️ 注意事项:${NC}"
echo "  1. 这是自签名证书，浏览器会显示安全警告（正常现象）"
echo "  2. 仅适用于开发和测试环境"
echo "  3. 生产环境请使用Let's Encrypt或购买商业SSL证书"
echo ""
echo "浏览器信任设置（可选）:"
echo "  Chrome: 设置 -> 隐私和安全 -> 管理证书 -> 导入 ssl/cert.pem"
echo "  Firefox: 设置 -> 隐私与安全 -> 证书 -> 导入 ssl/cert.pem"
echo ""
echo "生产环境证书申请:"
echo "  sudo certbot certonly --standalone -d yourdomain.com"
echo ""
echo -e "${GREEN}祝您使用愉快！${NC}"

