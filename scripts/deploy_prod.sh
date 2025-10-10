#!/bin/bash
# 生产环境部署脚本（简化包装）

set -e  # 遇到错误立即退出

echo "========================================"
echo "部署志航密信生产环境"
echo "========================================"

# 检查当前目录
if [ ! -f "docker-compose.production.yml" ]; then
    echo "错误: 请在项目根目录执行此脚本"
    exit 1
fi

# 检查.env文件
if [ ! -f ".env" ]; then
    echo "❌ 错误: 缺少.env文件"
    echo "请复制ENV_TEMPLATE.md到.env并配置环境变量"
    exit 1
fi

# 验证必要的环境变量
echo "🔍 检查环境变量..."
required_vars=(
    "MYSQL_ROOT_PASSWORD"
    "MYSQL_DATABASE"
    "MYSQL_USER"
    "MYSQL_PASSWORD"
    "REDIS_PASSWORD"
    "MINIO_ROOT_USER"
    "MINIO_ROOT_PASSWORD"
    "JWT_SECRET"
)

missing_vars=()
for var in "${required_vars[@]}"; do
    if ! grep -q "^${var}=" .env; then
        missing_vars+=("$var")
    fi
done

if [ ${#missing_vars[@]} -gt 0 ]; then
    echo "❌ 错误: 缺少以下必要环境变量:"
    for var in "${missing_vars[@]}"; do
        echo "  - $var"
    done
    echo ""
    echo "请在.env文件中配置这些变量"
    exit 1
fi

echo "✅ 环境变量检查通过"
echo ""

# 部署服务
echo "🚀 启动生产环境服务..."
docker-compose -f docker-compose.production.yml up -d

echo ""
echo "⏳ 等待服务启动（120秒）..."
sleep 120

echo ""
echo "🔍 检查服务状态..."
docker-compose -f docker-compose.production.yml ps

echo ""
echo "✅ 部署完成！"
echo ""
echo "📊 访问地址:"
echo "  - 管理后台: http://your-server:3001"
echo "  - 后端API: http://your-server:8080"
echo "  - Grafana: http://your-server:3000"
echo ""
echo "📝 查看日志: docker-compose -f docker-compose.production.yml logs -f"
echo "🛑 停止服务: docker-compose -f docker-compose.production.yml down"

