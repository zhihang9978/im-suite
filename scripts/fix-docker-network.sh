#!/bin/bash

# Docker网络问题一键修复脚本
# 适用于: Ubuntu/Debian/CentOS/RHEL
# 版本: v1.0.0
# 作者: 志航密信开发团队

set -e

echo "========================================="
echo " Docker网络问题一键修复脚本 v1.0.0"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否为root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}❌ 请使用root权限运行此脚本${NC}"
    echo "使用: sudo $0"
    exit 1
fi

echo "1️⃣  备份原配置..."
# 备份DNS配置
if [ -f /etc/resolv.conf ]; then
    cp /etc/resolv.conf /etc/resolv.conf.backup.$(date +%Y%m%d_%H%M%S)
    echo -e "${GREEN}✅ DNS配置已备份${NC}"
fi

# 备份Docker配置
if [ -f /etc/docker/daemon.json ]; then
    cp /etc/docker/daemon.json /etc/docker/daemon.json.backup.$(date +%Y%m%d_%H%M%S)
    echo -e "${GREEN}✅ Docker配置已备份${NC}"
fi

echo ""
echo "2️⃣  修复DNS配置..."

# 修改DNS
tee /etc/resolv.conf > /dev/null <<EOF
# Google Public DNS
nameserver 8.8.8.8
nameserver 8.8.4.4

# Cloudflare DNS
nameserver 1.1.1.1
nameserver 1.0.0.1

# 阿里云公共DNS
nameserver 223.5.5.5
nameserver 223.6.6.6

# 国内公共DNS
nameserver 114.114.114.114
nameserver 119.29.29.29
EOF

echo -e "${GREEN}✅ DNS配置已更新${NC}"

# 测试DNS
echo ""
echo "   测试DNS解析..."
if ping -c 2 -W 3 8.8.8.8 > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ Google DNS连接正常${NC}"
else
    echo -e "${YELLOW}   ⚠️  Google DNS连接失败（可能被防火墙拦截）${NC}"
fi

if ping -c 2 -W 3 114.114.114.114 > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ 国内DNS连接正常${NC}"
else
    echo -e "${RED}   ❌ 国内DNS连接失败${NC}"
fi

echo ""
echo "3️⃣  配置Docker镜像源..."

# 创建Docker配置目录
mkdir -p /etc/docker

# 配置Docker镜像源
tee /etc/docker/daemon.json > /dev/null <<'EOF'
{
  "registry-mirrors": [
    "https://hub-mirror.c.163.com",
    "https://mirror.ccs.tencentyun.com",
    "https://registry.docker-cn.com",
    "https://docker.mirrors.ustc.edu.cn",
    "https://dockerproxy.com"
  ],
  "dns": ["8.8.8.8", "114.114.114.114", "223.5.5.5", "1.1.1.1"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "live-restore": true,
  "userland-proxy": false
}
EOF

echo -e "${GREEN}✅ Docker镜像源已配置${NC}"

echo ""
echo "4️⃣  重启Docker服务..."

# 重新加载配置
systemctl daemon-reload

# 重启Docker
systemctl restart docker

# 等待Docker启动
sleep 3

# 检查Docker状态
if systemctl is-active --quiet docker; then
    echo -e "${GREEN}✅ Docker服务运行正常${NC}"
else
    echo -e "${RED}❌ Docker服务启动失败${NC}"
    echo "   查看日志: journalctl -u docker -n 50"
    exit 1
fi

echo ""
echo "5️⃣  验证修复结果..."

# 显示Docker信息
echo "   Docker镜像源:"
docker info 2>/dev/null | grep -A 10 "Registry Mirrors" || echo -e "${YELLOW}   未找到镜像源信息${NC}"

echo ""
echo "   测试DNS解析..."
if nslookup docker.io > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ docker.io DNS解析成功${NC}"
else
    echo -e "${YELLOW}   ⚠️  docker.io DNS解析失败${NC}"
fi

if nslookup hub-mirror.c.163.com > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ 网易镜像源DNS解析成功${NC}"
else
    echo -e "${YELLOW}   ⚠️  网易镜像源DNS解析失败${NC}"
fi

echo ""
echo "   测试Docker镜像拉取..."
echo "   正在拉取测试镜像 alpine:latest（约5MB）..."

if timeout 120 docker pull alpine:latest > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ Docker镜像拉取成功！${NC}"
    docker rmi alpine:latest > /dev/null 2>&1
    PULL_SUCCESS=true
else
    echo -e "${RED}   ❌ Docker镜像拉取失败${NC}"
    PULL_SUCCESS=false
fi

echo ""
echo "========================================="

if [ "$PULL_SUCCESS" = true ]; then
    echo -e "${GREEN}🎉 修复成功！${NC}"
    echo ""
    echo "下一步:"
    echo "1. 可以开始部署IM系统: cd /path/to/im-suite && docker-compose up -d"
    echo "2. 或继续拉取其他镜像"
    echo ""
else
    echo -e "${YELLOW}⚠️  修复未完全成功${NC}"
    echo ""
    echo "可能的原因:"
    echo "1. 云服务器安全组限制（最常见）"
    echo "   - 请在云服务器控制台检查安全组"
    echo "   - 确保允许出站: HTTPS(443), DNS(53)"
    echo ""
    echo "2. 服务器防火墙限制"
    echo "   检查: ufw status 或 iptables -L"
    echo ""
    echo "3. ISP网络限制"
    echo "   考虑使用VPN或代理"
    echo ""
    echo "4. 备用方案:"
    echo "   - 使用手动上传镜像方案"
    echo "   - 见文档: DEPLOYMENT_FOR_DEVIN_V1.6.0.md"
    echo ""
fi

echo "========================================="
echo ""

# 显示诊断信息
echo "📊 诊断信息:"
echo ""
echo "DNS配置:"
cat /etc/resolv.conf | head -5
echo ""
echo "Docker配置:"
cat /etc/docker/daemon.json 2>/dev/null || echo "无配置文件"
echo ""
echo "Docker状态:"
systemctl status docker | head -3
echo ""

exit 0

