#!/bin/bash

# 志航密信服务器部署脚本
# 服务器: 154.37.214.191
# 部署路径: /opt/im-suite
# 执行方式: bash server-deploy.sh

set -e  # 遇到错误立即退出

echo "========================================="
echo "志航密信 - 服务器部署脚本"
echo "版本: v1.3.1 - 超级管理后台版"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then 
   echo -e "${RED}请使用root用户运行此脚本${NC}"
   exit 1
fi

# 1. 检查并安装必要软件
echo -e "${YELLOW}[1/10] 检查系统环境...${NC}"
# 检查操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
    VER=$VERSION_ID
    echo "操作系统: $OS $VER"
else
    echo -e "${RED}无法识别操作系统${NC}"
    exit 1
fi

# 2. 安装Docker
echo -e "${YELLOW}[2/10] 安装Docker...${NC}"
if ! command -v docker &> /dev/null; then
    echo "正在安装Docker..."
    curl -fsSL https://get.docker.com | bash
    systemctl start docker
    systemctl enable docker
    echo -e "${GREEN}Docker安装成功${NC}"
else
    echo "Docker已安装: $(docker --version)"
fi

# 3. 安装Docker Compose
echo -e "${YELLOW}[3/10] 安装Docker Compose...${NC}"
if ! command -v docker-compose &> /dev/null; then
    echo "正在安装Docker Compose..."
    curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    echo -e "${GREEN}Docker Compose安装成功${NC}"
else
    echo "Docker Compose已安装: $(docker-compose --version)"
fi

# 4. 创建项目目录
echo -e "${YELLOW}[4/10] 创建项目目录...${NC}"
mkdir -p /opt/im-suite
cd /opt/im-suite

# 5. 克隆项目代码
echo -e "${YELLOW}[5/10] 克隆项目代码...${NC}"
if [ ! -d ".git" ]; then
    echo "正在克隆代码仓库..."
    git clone https://github.com/zhihang9978/im-suite.git .
    echo -e "${GREEN}代码克隆成功${NC}"
else
    echo "代码仓库已存在，正在更新..."
    git pull origin main
    echo -e "${GREEN}代码更新成功${NC}"
fi

# 6. 创建环境变量文件
echo -e "${YELLOW}[6/10] 配置环境变量...${NC}"
cat > .env.production <<EOF
# ========================================
# Docker Compose 配置
# ========================================
MYSQL_ROOT_PASSWORD=zhihang_im_2024_secure_password
MYSQL_DATABASE=zhihang_messenger
MYSQL_USER=zhihang
MYSQL_PASSWORD=zhihang_im_2024_secure_password

REDIS_PASSWORD=zhihang_redis_2024_secure_password

MINIO_ROOT_USER=zhihang_minio_admin
MINIO_ROOT_PASSWORD=zhihang_minio_2024_secure_key

JWT_SECRET=zhihang_jwt_super_secret_key_2024_production

ADMIN_API_BASE_URL=http://backend:8080
WEB_API_BASE_URL=http://backend:8080
WEB_WS_BASE_URL=ws://backend:8080/ws

WEBRTC_ICE_SERVERS=[{"urls":"stun:stun.l.google.com:19302"}]

GRAFANA_PASSWORD=zhihang_grafana_admin_2024

# ========================================
# 后端应用配置
# ========================================
DB_HOST=mysql
DB_PORT=3306
DB_NAME=zhihang_messenger
DB_USER=zhihang

REDIS_HOST=redis
REDIS_PORT=6379

MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=zhihang_minio_admin
MINIO_SECRET_KEY=zhihang_minio_2024_secure_key
MINIO_USE_SSL=false

JWT_EXPIRES_IN=24h

PORT=8080
GIN_MODE=release
LOG_LEVEL=info

MAX_FILE_SIZE=100MB
UPLOAD_PATH=/app/uploads

# 域名配置（请替换为您的实际域名）
DOMAIN=im.yourdomain.com

# 第三方服务配置（需要您自行申请）
# 短信服务
SMS_PROVIDER=aliyun
SMS_ACCESS_KEY=your_sms_access_key
SMS_SECRET_KEY=your_sms_secret_key

# 推送服务
PUSH_PROVIDER=jpush
PUSH_APP_KEY=your_push_app_key
PUSH_MASTER_SECRET=your_push_master_secret

# 对象存储（如果使用云服务）
OSS_PROVIDER=minio
OSS_ENDPOINT=http://minio:9000
OSS_ACCESS_KEY=zhihang_minio_admin
OSS_SECRET_KEY=zhihang_minio_2024_secure_key

# 监控配置
PROMETHEUS_ENABLED=true
GRAFANA_ENABLED=true
GRAFANA_ADMIN_PASSWORD=zhihang_grafana_admin_2024
EOF

echo -e "${GREEN}环境变量配置完成${NC}"
echo -e "${YELLOW}提示: 请修改 .env.production 中的密码和第三方服务配置${NC}"

# 7. 创建数据目录
echo -e "${YELLOW}[7/10] 创建数据目录...${NC}"
mkdir -p /opt/im-suite/data/{mysql,redis,minio,prometheus,grafana,logs}
chmod -R 777 /opt/im-suite/data
echo -e "${GREEN}数据目录创建成功${NC}"

# 8. 生成SSL证书（自签名）
echo -e "${YELLOW}[8/10] 生成SSL证书...${NC}"
if [ ! -f "ssl/cert.pem" ] || [ ! -f "ssl/key.pem" ]; then
    echo "正在生成自签名SSL证书..."
    mkdir -p ssl
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
      -keyout ssl/key.pem \
      -out ssl/cert.pem \
      -subj "/C=CN/ST=Beijing/L=Beijing/O=ZhiHang/OU=IT/CN=154.37.214.191" \
      -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:154.37.214.191,IP:127.0.0.1" 2>/dev/null || \
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
      -keyout ssl/key.pem \
      -out ssl/cert.pem \
      -subj "/C=CN/ST=Beijing/L=Beijing/O=ZhiHang/OU=IT/CN=154.37.214.191"
    chmod 600 ssl/key.pem
    chmod 644 ssl/cert.pem
    echo -e "${GREEN}SSL证书生成成功${NC}"
else
    echo "SSL证书已存在，跳过生成"
fi

# 9. 启动服务
echo -e "${YELLOW}[9/10] 启动Docker服务...${NC}"
echo "正在启动服务，这可能需要几分钟..."
docker-compose -f docker-compose.production.yml up -d

echo ""
echo "等待服务启动..."
sleep 10

# 10. 检查服务状态
echo -e "${YELLOW}[10/10] 检查服务状态...${NC}"
docker-compose -f docker-compose.production.yml ps

# 显示服务信息
echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}部署完成！${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "服务访问地址:"
echo "  - 后端API: http://154.37.214.191:8080"
echo "  - Web客户端: http://154.37.214.191:3002"
echo "  - 管理后台: http://154.37.214.191:3001"
echo "  - Grafana监控: http://154.37.214.191:3000"
echo "  - Prometheus: http://154.37.214.191:9090"
echo "  - Nginx负载均衡: http://154.37.214.191:80"
echo ""
echo "默认管理员账号:"
echo "  - 用户名: admin"
echo "  - 密码: 请在首次登录后修改"
echo ""
echo "查看日志:"
echo "  docker-compose -f docker-compose.production.yml logs -f [service_name]"
echo ""
echo "停止服务:"
echo "  docker-compose -f docker-compose.production.yml down"
echo ""
echo "重启服务:"
echo "  docker-compose -f docker-compose.production.yml restart"
echo ""
echo -e "${YELLOW}重要提示:${NC}"
echo "1. 请修改 .env.production 中的所有密码"
echo "2. 请配置第三方服务（短信/推送）的API密钥"
echo "3. 请配置您的域名并申请SSL证书"
echo "4. 建议配置防火墙规则，只开放必要端口"
echo "5. 定期备份数据库和文件存储"
echo ""
echo -e "${GREEN}祝您使用愉快！${NC}"


