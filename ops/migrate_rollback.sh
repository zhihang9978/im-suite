#!/bin/bash

###############################################################################
# 志航密信 - 数据库迁移回滚脚本
# 用途：安全地回滚数据库迁移
# 使用：bash ops/migrate_rollback.sh [--confirm]
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

# 加载环境变量
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
elif [ -f "/opt/im-suite/.env" ]; then
    cd /opt/im-suite
    export $(cat .env | grep -v '^#' | xargs)
else
    log_error "未找到.env文件"
    exit 1
fi

log_warning "========================================="
log_warning "⚠️  数据库迁移回滚警告"
log_warning "========================================="
echo ""
log_error "此操作将："
echo "  1. 删除所有数据库表"
echo "  2. 清除所有数据"
echo "  3. 无法撤销"
echo ""
log_warning "请确保："
echo "  ✅ 已备份数据库"
echo "  ✅ 了解此操作的后果"
echo "  ✅ 已获得授权"
echo ""

# 确认操作
if [ "$1" != "--confirm" ]; then
    read -p "确认继续回滚？输入 'YES I UNDERSTAND' 继续: " CONFIRM
    
    if [ "$CONFIRM" != "YES I UNDERSTAND" ]; then
        log_info "回滚已取消"
        exit 0
    fi
fi

# ====================================
# 1. 创建备份
# ====================================
log_info "1. 创建回滚前备份..."

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="/opt/im-suite/backups/migrations"
mkdir -p "$BACKUP_DIR"

BACKUP_FILE="${BACKUP_DIR}/pre-rollback-${TIMESTAMP}.sql.gz"

docker exec im-mysql-prod mysqldump \
    -uroot -p${MYSQL_ROOT_PASSWORD} \
    --all-databases \
    --single-transaction \
    --quick \
    | gzip > "$BACKUP_FILE"

log_success "备份完成: $BACKUP_FILE"

# ====================================
# 2. 执行回滚
# ====================================
log_info "2. 执行数据库回滚..."

docker exec -i im-mysql-prod mysql -uroot -p${MYSQL_ROOT_PASSWORD} ${DB_NAME} < config/database/migration_rollback.sql

log_success "回滚SQL执行完成"

# ====================================
# 3. 验证回滚
# ====================================
log_info "3. 验证回滚结果..."

TABLE_COUNT=$(docker exec im-mysql-prod mysql -uroot -p${MYSQL_ROOT_PASSWORD} ${DB_NAME} -e "SHOW TABLES;" | wc -l)

if [ $TABLE_COUNT -le 1 ]; then
    log_success "所有表已删除"
else
    log_warning "仍有 $((TABLE_COUNT - 1)) 个表存在"
    docker exec im-mysql-prod mysql -uroot -p${MYSQL_ROOT_PASSWORD} ${DB_NAME} -e "SHOW TABLES;"
fi

# ====================================
# 4. 可选：重新迁移
# ====================================
log_info "4. 是否重新迁移？"
read -p "输入 'yes' 重新创建所有表: " REMIGRATE

if [ "$REMIGRATE" = "yes" ]; then
    log_info "重新迁移数据库..."
    
    # 重启backend服务，触发自动迁移
    docker-compose -f docker-compose.production.yml restart backend
    
    # 等待迁移完成
    sleep 10
    
    # 验证表创建
    NEW_TABLE_COUNT=$(docker exec im-mysql-prod mysql -uroot -p${MYSQL_ROOT_PASSWORD} ${DB_NAME} -e "SHOW TABLES;" | wc -l)
    log_success "重新迁移完成，创建了 $((NEW_TABLE_COUNT - 1)) 个表"
fi

# ====================================
# 完成
# ====================================
log_success "========================================="
log_success "迁移回滚完成"
log_success "========================================="
echo ""
echo "备份位置: $BACKUP_FILE"
echo "表数量: $((TABLE_COUNT - 1))"
echo ""
echo "如需恢复备份:"
echo "  gunzip -c $BACKUP_FILE | docker exec -i im-mysql-prod mysql -uroot -p\${MYSQL_ROOT_PASSWORD}"
echo ""

