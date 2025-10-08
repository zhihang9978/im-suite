#!/bin/bash

# 志航密信 - Docker容器清理脚本
# 清理所有im-*容器，为重新部署做准备

set -e

echo "========================================="
echo "志航密信 - Docker容器清理"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查是否有root权限
if [ "$EUID" -eq 0 ]; then 
   echo -e "${YELLOW}⚠️  正在以root用户运行${NC}"
fi

# 显示当前运行的im-*容器
echo -e "${YELLOW}[1/4] 检查当前运行的IM Suite容器...${NC}"
RUNNING_CONTAINERS=$(docker ps -a | grep "im-" | wc -l)
if [ $RUNNING_CONTAINERS -gt 0 ]; then
    echo "找到 $RUNNING_CONTAINERS 个IM Suite容器:"
    docker ps -a | grep "im-" | awk '{print "  -", $NF, "("$7")"}'
else
    echo "未找到IM Suite容器"
fi
echo ""

# 停止所有im-*容器
if [ $RUNNING_CONTAINERS -gt 0 ]; then
    echo -e "${YELLOW}[2/4] 停止所有IM Suite容器...${NC}"
    docker ps -a | grep "im-" | awk '{print $1}' | xargs -r docker stop 2>/dev/null || true
    echo -e "${GREEN}✅ 容器已停止${NC}"
else
    echo -e "${YELLOW}[2/4] 没有容器需要停止${NC}"
fi
echo ""

# 删除所有im-*容器
if [ $RUNNING_CONTAINERS -gt 0 ]; then
    echo -e "${YELLOW}[3/4] 删除所有IM Suite容器...${NC}"
    docker ps -a | grep "im-" | awk '{print $1}' | xargs -r docker rm 2>/dev/null || true
    echo -e "${GREEN}✅ 容器已删除${NC}"
else
    echo -e "${YELLOW}[3/4] 没有容器需要删除${NC}"
fi
echo ""

# 清理未使用的Docker资源
echo -e "${YELLOW}[4/4] 清理未使用的Docker资源...${NC}"
echo "  - 清理未使用的网络..."
docker network prune -f 2>/dev/null || true

echo "  - 清理悬空镜像..."
docker image prune -f 2>/dev/null || true

echo -e "${GREEN}✅ Docker资源已清理${NC}"
echo ""

# 可选：清理未使用的数据卷（会删除数据，需要用户确认）
echo -e "${RED}⚠️  是否清理未使用的数据卷？（这会删除未挂载的数据）${NC}"
read -p "输入 'yes' 确认清理数据卷: " confirm
if [ "$confirm" = "yes" ]; then
    echo "  - 清理未使用的数据卷..."
    docker volume prune -f
    echo -e "${GREEN}✅ 数据卷已清理${NC}"
else
    echo -e "${YELLOW}跳过数据卷清理${NC}"
fi
echo ""

# 显示清理后的状态
echo "========================================="
echo -e "${GREEN}✅ 清理完成${NC}"
echo "========================================="
echo ""
echo "当前Docker状态:"
echo "  容器数量: $(docker ps -a | wc -l)"
echo "  镜像数量: $(docker images | wc -l)"
echo "  数据卷数量: $(docker volume ls | wc -l)"
echo "  网络数量: $(docker network ls | wc -l)"
echo ""
echo "磁盘使用情况:"
docker system df
echo ""
echo -e "${GREEN}现在可以重新部署IM Suite了！${NC}"
echo "执行: docker-compose -f docker-compose.production.yml up -d"

