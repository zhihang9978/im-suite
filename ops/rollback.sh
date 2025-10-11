#!/bin/bash

###############################################################################
# 志航密信 - 快速回滚脚本
# 用途：回滚到指定版本，恢复数据和配置
# 使用：bash ops/rollback.sh [TIMESTAMP]
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

DEPLOY_DIR="/opt/im-suite"
BACKUP_DIR="${DEPLOY_DIR}/backups/deployments"

log_info "========================================="
log_info "志航密信 - 快速回滚"
log_info "========================================="

# ============================================
# 1. 选择回滚版本
# ============================================
if [ -z "$1" ]; then
    log_info "可用的备份版本："
    ls -lt "$BACKUP_DIR" | grep "^d" | head -5 | awk '{print $NF}'
    echo ""
    read -p "请输入要回滚的版本时间戳（格式：20251011-150000）: " TIMESTAMP
else
    TIMESTAMP=$1
fi

BACKUP_PATH="${BACKUP_DIR}/${TIMESTAMP}"

if [ ! -d "$BACKUP_PATH" ]; then
    log_error "备份不存在: $BACKUP_PATH"
    exit 1
fi

log_info "将回滚到版本: $TIMESTAMP"
log_warning "此操作将："
echo "  1. 停止当前服务"
echo "  2. 恢复数据库到备份点"
echo "  3. 恢复配置文件"
echo "  4. 重启服务"
echo ""
read -p "确认继续？(yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    log_info "回滚已取消"
    exit 0
fi

# ============================================
# 2. 停止当前服务
# ============================================
log_info "2. 停止当前服务..."
cd "$DEPLOY_DIR"
docker-compose -f docker-compose.production.yml down
log_success "服务已停止"

# ============================================
# 3. 恢复数据库
# ============================================
log_info "3. 恢复MySQL数据库..."

# 启动MySQL
docker-compose -f docker-compose.production.yml up -d mysql
sleep 10

# 加载.env
if [ -f "${BACKUP_PATH}/.env.backup" ]; then
    source "${BACKUP_PATH}/.env.backup"
fi

# 恢复数据库
if [ -f "${BACKUP_PATH}/mysql-backup.sql" ]; then
    docker exec -i im-mysql-prod mysql -uroot -p${MYSQL_ROOT_PASSWORD} < "${BACKUP_PATH}/mysql-backup.sql"
    log_success "MySQL数据库恢复完成"
else
    log_warning "未找到MySQL备份文件"
fi

# ============================================
# 4. 恢复Redis
# ============================================
log_info "4. 恢复Redis数据..."

docker-compose -f docker-compose.production.yml up -d redis
sleep 5

if [ -f "${BACKUP_PATH}/redis-dump.rdb" ]; then
    docker-compose -f docker-compose.production.yml stop redis
    docker cp "${BACKUP_PATH}/redis-dump.rdb" im-redis-prod:/data/dump.rdb
    docker-compose -f docker-compose.production.yml start redis
    log_success "Redis数据恢复完成"
else
    log_warning "未找到Redis备份文件"
fi

# ============================================
# 5. 恢复配置
# ============================================
log_info "5. 恢复配置文件..."

if [ -f "${BACKUP_PATH}/.env.backup" ]; then
    cp "${BACKUP_PATH}/.env.backup" .env
    log_success "配置文件恢复完成"
fi

# ============================================
# 6. 启动所有服务
# ============================================
log_info "6. 启动所有服务..."

docker-compose -f docker-compose.production.yml up -d

# ============================================
# 7. 健康检查
# ============================================
log_info "7. 执行健康检查..."

WAIT_TIME=0
MAX_WAIT=120

while [ $WAIT_TIME -lt $MAX_WAIT ]; do
    if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
        log_success "健康检查通过！"
        break
    fi
    
    sleep 5
    WAIT_TIME=$((WAIT_TIME + 5))
    log_info "等待服务启动... (${WAIT_TIME}s/${MAX_WAIT}s)"
done

if [ $WAIT_TIME -ge $MAX_WAIT ]; then
    log_error "健康检查失败！回滚可能不成功"
    log_error "请检查日志: docker-compose logs"
    exit 1
fi

# ============================================
# 8. 验证服务
# ============================================
log_info "8. 验证所有服务..."

docker-compose -f docker-compose.production.yml ps

# ============================================
# 完成
# ============================================
ELAPSED=$(($(date +%s) - $(date -d "$(stat -c %y $LOG_FILE | cut -d. -f1)" +%s) ))

log_success "========================================="
log_success "回滚完成！"
log_success "========================================="
echo ""
echo "回滚信息："
echo "  回滚版本: $TIMESTAMP"
echo "  执行时间: ${ELAPSED}秒"
echo "  日志文件: $LOG_FILE"
echo ""
echo "验证回滚："
echo "  curl http://localhost:8080/health"
echo "  docker-compose logs -f --tail=50"
echo ""

