#!/bin/bash
# 构建管理后台脚本（简化包装）

set -e  # 遇到错误立即退出

echo "========================================"
echo "构建志航密信管理后台"
echo "========================================"

# 检查当前目录
if [ ! -f "docker-compose.production.yml" ]; then
    echo "错误: 请在项目根目录执行此脚本"
    exit 1
fi

# 构建管理后台镜像
echo "🔨 构建管理后台Docker镜像..."
docker-compose -f docker-compose.production.yml build admin

echo "✅ 管理后台构建完成！"
echo ""
echo "下一步: 运行 ./scripts/deploy_prod.sh 来部署"

