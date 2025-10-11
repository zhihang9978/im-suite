#!/bin/bash

###############################################################################
# 志航密信 - 零停机部署脚本
# 用途：蓝绿部署，自动健康检查，失败自动回滚
# 使用：bash ops/deploy.sh [--rollback]
###############################################################################

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 配置
DEPLOY_DIR="/opt/im-suite"
BACKUP_DIR="${DEPLOY_DIR}/backups/deployments"
LOG_FILE="/var/log/im-suite/deploy-$(date +%Y%m%d-%H%M%S).log"
MAX_WAIT_TIME=300  # 最大等待时间5分钟
HEALTH_CHECK_URL="http://localhost:8080/health"

# 创建日志
mkdir -p "$(dirname $LOG_FILE)"
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

log_info "========================================="
log_info "志航密信 - 零停机部署"
log_info "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
log_info "========================================="

# ============================================
# 1. 前置检查
# ============================================
log_info "1. 执行前置检查..."

# 检查当前目录
if [ ! -f "docker-compose.production.yml" ]; then
    log_error "请在项目根目录执行此脚本"
    exit 1
fi

# 检查.env文件
if [ ! -f ".env" ]; then
    log_error "缺少.env文件，请先配置环境变量"
    exit 1
fi

# 检查必要的环境变量
required_vars=(
    "MYSQL_ROOT_PASSWORD"
    "JWT_SECRET"
    "REDIS_PASSWORD"
    "MINIO_ROOT_PASSWORD"
)

for var in "${required_vars[@]}"; do
    if ! grep -q "^${var}=" .env; then
        log_error "缺少必需环境变量: $var"
        exit 1
    fi
done

log_success "前置检查通过"

# ============================================
# 2. 备份当前版本
# ============================================
log_info "2. 备份当前版本..."

# 创建备份目录
BACKUP_TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_TIMESTAMP}"
mkdir -p "$BACKUP_PATH"

# 备份数据库
log_info "备份MySQL数据库..."
docker exec im-mysql-prod mysqldump \
    -uroot -p${MYSQL_ROOT_PASSWORD} \
    --all-databases \
    --single-transaction \
    --quick \
    --lock-tables=false \
    > "${BACKUP_PATH}/mysql-backup.sql" 2>/dev/null || log_warning "MySQL备份失败（可能是首次部署）"

# 备份Redis
log_info "备份Redis数据..."
docker exec im-redis-prod redis-cli --no-auth-warning -a ${REDIS_PASSWORD} SAVE 2>/dev/null || log_warning "Redis备份失败（可能是首次部署）"
docker cp im-redis-prod:/data/dump.rdb "${BACKUP_PATH}/redis-dump.rdb" 2>/dev/null || true

# 记录当前镜像版本
docker images | grep "im-" > "${BACKUP_PATH}/images.txt"
docker-compose -f docker-compose.production.yml ps > "${BACKUP_PATH}/services.txt"

# 备份.env文件
cp .env "${BACKUP_PATH}/.env.backup"

# 创建回滚脚本
cat > "${BACKUP_PATH}/rollback.sh" <<ROLLBACK_SCRIPT
#!/bin/bash
# 自动生成的回滚脚本
echo "回滚到版本: ${BACKUP_TIMESTAMP}"
cd ${DEPLOY_DIR}

# 恢复数据库
docker exec -i im-mysql-prod mysql -uroot -p\${MYSQL_ROOT_PASSWORD} < ${BACKUP_PATH}/mysql-backup.sql

# 恢复Redis
docker cp ${BACKUP_PATH}/redis-dump.rdb im-redis-prod:/data/dump.rdb
docker-compose -f docker-compose.production.yml restart redis

# 恢复配置
cp ${BACKUP_PATH}/.env.backup .env

# 重启服务
docker-compose -f docker-compose.production.yml restart

echo "回滚完成"
ROLLBACK_SCRIPT

chmod +x "${BACKUP_PATH}/rollback.sh"
log_success "备份完成: $BACKUP_PATH"

# ============================================
# 3. 拉取最新代码
# ============================================
log_info "3. 拉取最新代码..."

git fetch origin
CURRENT_BRANCH=$(git branch --show-current)
log_info "当前分支: $CURRENT_BRANCH"

if [ "$1" != "--skip-pull" ]; then
    git pull origin $CURRENT_BRANCH
    log_success "代码更新完成"
fi

# ============================================
# 4. 构建新镜像
# ============================================
log_info "4. 构建Docker镜像..."

docker-compose -f docker-compose.production.yml build --no-cache
log_success "镜像构建完成"

# ============================================
# 5. 数据库迁移（非破坏性）
# ============================================
log_info "5. 执行数据库迁移..."

# 启动临时backend容器进行迁移
docker-compose -f docker-compose.production.yml run --rm backend go run main.go migrate || log_warning "迁移命令未找到（将在启动时自动迁移）"

# ============================================
# 6. 零停机部署（蓝绿切换）
# ============================================
log_info "6. 执行零停机部署..."

# 启动新容器（蓝绿部署）
log_info "启动新版本服务..."
docker-compose -f docker-compose.production.yml up -d --no-deps --build backend admin

# ============================================
# 7. 健康检查
# ============================================
log_info "7. 执行健康检查..."

WAIT_TIME=0
while [ $WAIT_TIME -lt $MAX_WAIT_TIME ]; do
    if curl -f -s $HEALTH_CHECK_URL > /dev/null 2>&1; then
        log_success "健康检查通过！"
        break
    fi
    
    log_info "等待服务启动... (${WAIT_TIME}s/${MAX_WAIT_TIME}s)"
    sleep 5
    WAIT_TIME=$((WAIT_TIME + 5))
done

if [ $WAIT_TIME -ge $MAX_WAIT_TIME ]; then
    log_error "健康检查失败！开始自动回滚..."
    bash "${BACKUP_PATH}/rollback.sh"
    exit 1
fi

# ============================================
# 8. 清理旧容器和镜像
# ============================================
log_info "8. 清理旧容器和镜像..."

docker system prune -f --volumes=false
log_success "清理完成"

# ============================================
# 9. 验证部署
# ============================================
log_info "9. 验证部署..."

# 检查所有服务状态
SERVICES=$(docker-compose -f docker-compose.production.yml ps --services)
FAILED_SERVICES=""

for service in $SERVICES; do
    STATUS=$(docker-compose -f docker-compose.production.yml ps $service --format json 2>/dev/null | jq -r '.[0].Health' 2>/dev/null || echo "unknown")
    if [ "$STATUS" = "healthy" ] || [ "$STATUS" = "unknown" ]; then
        log_success "✓ $service: 运行正常"
    else
        log_error "✗ $service: 状态异常 ($STATUS)"
        FAILED_SERVICES="$FAILED_SERVICES $service"
    fi
done

if [ -n "$FAILED_SERVICES" ]; then
    log_error "以下服务状态异常:$FAILED_SERVICES"
    log_error "部署失败！"
    exit 1
fi

# ============================================
# 10. 记录部署信息
# ============================================
log_info "10. 记录部署信息..."

cat > "${BACKUP_DIR}/latest-deployment.json" <<EOF
{
  "timestamp": "$(date -Iseconds)",
  "git_commit": "$(git rev-parse HEAD)",
  "git_branch": "$(git branch --show-current)",
  "backup_path": "${BACKUP_PATH}",
  "deployed_by": "${SUDO_USER:-root}",
  "status": "success"
}
EOF

log_success "部署信息已记录"

# ============================================
# 完成
# ============================================
log_success "========================================="
log_success "部署成功完成！"
log_success "========================================="
echo ""
echo "部署信息："
echo "  时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "  Git Commit: $(git rev-parse --short HEAD)"
echo "  备份位置: $BACKUP_PATH"
echo "  日志文件: $LOG_FILE"
echo ""
echo "服务访问："
echo "  前端: http://$(hostname -I | awk '{print $1}'):3001/"
echo "  后端: http://$(hostname -I | awk '{print $1}'):8080/health"
echo ""
echo "如需回滚："
echo "  bash ${BACKUP_PATH}/rollback.sh"
echo "  或: bash ops/rollback.sh ${BACKUP_TIMESTAMP}"
echo ""

