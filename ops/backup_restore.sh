#!/bin/bash

###############################################################################
# 志航密信 - 备份恢复脚本
# 用途：自动化备份MySQL/Redis/MinIO，支持恢复和清理
# 使用：
#   备份：bash ops/backup_restore.sh backup
#   恢复：bash ops/backup_restore.sh restore [TIMESTAMP]
#   清理：bash ops/backup_restore.sh cleanup [DAYS]
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

# 配置
BACKUP_BASE_DIR="/opt/im-suite/backups"
MYSQL_BACKUP_DIR="${BACKUP_BASE_DIR}/mysql"
REDIS_BACKUP_DIR="${BACKUP_BASE_DIR}/redis"
MINIO_BACKUP_DIR="${BACKUP_BASE_DIR}/minio"
LOG_FILE="/var/log/im-suite/backup-$(date +%Y%m%d-%H%M%S).log"

# 保留策略
DAILY_RETENTION=7     # 保留7天的每日备份
WEEKLY_RETENTION=4    # 保留4周的每周备份
MONTHLY_RETENTION=3   # 保留3个月的每月备份

# 创建日志
mkdir -p "$(dirname $LOG_FILE)"
mkdir -p "$MYSQL_BACKUP_DIR" "$REDIS_BACKUP_DIR" "$MINIO_BACKUP_DIR"

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

###############################################################################
# 备份函数
###############################################################################

backup_mysql() {
    log_info "开始备份MySQL..."
    
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    BACKUP_FILE="${MYSQL_BACKUP_DIR}/mysql-${TIMESTAMP}.sql.gz"
    
    # 执行备份
    docker exec im-mysql-prod mysqldump \
        -uroot -p${MYSQL_ROOT_PASSWORD} \
        --all-databases \
        --single-transaction \
        --quick \
        --lock-tables=false \
        --routines \
        --triggers \
        --events \
        | gzip > "$BACKUP_FILE"
    
    # 验证备份
    if [ -s "$BACKUP_FILE" ]; then
        SIZE=$(du -h "$BACKUP_FILE" | awk '{print $1}')
        log_success "MySQL备份完成: $BACKUP_FILE ($SIZE)"
        echo "$BACKUP_FILE" > "${MYSQL_BACKUP_DIR}/latest.txt"
    else
        log_error "MySQL备份失败"
        rm -f "$BACKUP_FILE"
        return 1
    fi
}

backup_redis() {
    log_info "开始备份Redis..."
    
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    BACKUP_FILE="${REDIS_BACKUP_DIR}/redis-${TIMESTAMP}.rdb"
    
    # 触发BGSAVE
    docker exec im-redis-prod redis-cli \
        --no-auth-warning -a ${REDIS_PASSWORD} BGSAVE
    
    # 等待BGSAVE完成
    sleep 5
    while docker exec im-redis-prod redis-cli --no-auth-warning -a ${REDIS_PASSWORD} LASTSAVE | grep -q "$(date +%s)"; do
        sleep 1
    done
    
    # 复制RDB文件
    docker cp im-redis-prod:/data/dump.rdb "$BACKUP_FILE"
    
    if [ -s "$BACKUP_FILE" ]; then
        SIZE=$(du -h "$BACKUP_FILE" | awk '{print $1}')
        log_success "Redis备份完成: $BACKUP_FILE ($SIZE)"
        echo "$BACKUP_FILE" > "${REDIS_BACKUP_DIR}/latest.txt"
    else
        log_error "Redis备份失败"
        return 1
    fi
}

backup_minio() {
    log_info "开始备份MinIO..."
    
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    BACKUP_DIR="${MINIO_BACKUP_DIR}/${TIMESTAMP}"
    mkdir -p "$BACKUP_DIR"
    
    # 安装mc（MinIO Client）如果未安装
    if ! command -v mc &> /dev/null; then
        log_info "安装MinIO Client..."
        wget -q https://dl.min.io/client/mc/release/linux-amd64/mc -O /usr/local/bin/mc
        chmod +x /usr/local/bin/mc
    fi
    
    # 配置mc
    mc alias set local http://localhost:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD} --insecure || true
    
    # 备份所有bucket
    BUCKETS=$(mc ls local --insecure | awk '{print $5}')
    
    for bucket in $BUCKETS; do
        log_info "备份bucket: $bucket"
        mc mirror --overwrite local/$bucket "${BACKUP_DIR}/$bucket" --insecure
    done
    
    # 打包备份
    tar -czf "${MINIO_BACKUP_DIR}/minio-${TIMESTAMP}.tar.gz" -C "$BACKUP_DIR" .
    rm -rf "$BACKUP_DIR"
    
    SIZE=$(du -h "${MINIO_BACKUP_DIR}/minio-${TIMESTAMP}.tar.gz" | awk '{print $1}')
    log_success "MinIO备份完成: ${MINIO_BACKUP_DIR}/minio-${TIMESTAMP}.tar.gz ($SIZE)"
    echo "${MINIO_BACKUP_DIR}/minio-${TIMESTAMP}.tar.gz" > "${MINIO_BACKUP_DIR}/latest.txt"
}

do_backup() {
    log_info "========================================="
    log_info "开始全量备份"
    log_info "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log_info "========================================="
    
    START_TIME=$(date +%s)
    
    # 执行备份
    backup_mysql
    backup_redis
    backup_minio
    
    END_TIME=$(date +%s)
    ELAPSED=$((END_TIME - START_TIME))
    
    log_success "========================================="
    log_success "备份完成！耗时: ${ELAPSED}秒"
    log_success "========================================="
    
    # 生成备份报告
    cat > "${BACKUP_BASE_DIR}/latest-backup.json" <<EOF
{
  "timestamp": "$(date -Iseconds)",
  "duration_seconds": $ELAPSED,
  "mysql_backup": "$(cat ${MYSQL_BACKUP_DIR}/latest.txt 2>/dev/null || echo 'N/A')",
  "redis_backup": "$(cat ${REDIS_BACKUP_DIR}/latest.txt 2>/dev/null || echo 'N/A')",
  "minio_backup": "$(cat ${MINIO_BACKUP_DIR}/latest.txt 2>/dev/null || echo 'N/A')"
}
EOF
    
    # 清理旧备份
    cleanup_old_backups $DAILY_RETENTION
}

###############################################################################
# 恢复函数
###############################################################################

restore_mysql() {
    BACKUP_FILE=$1
    
    log_info "开始恢复MySQL..."
    log_warning "此操作将覆盖当前数据库！"
    
    if [ ! -f "$BACKUP_FILE" ]; then
        log_error "备份文件不存在: $BACKUP_FILE"
        return 1
    fi
    
    # 解压并恢复
    gunzip -c "$BACKUP_FILE" | docker exec -i im-mysql-prod mysql -uroot -p${MYSQL_ROOT_PASSWORD}
    
    log_success "MySQL恢复完成"
}

restore_redis() {
    BACKUP_FILE=$1
    
    log_info "开始恢复Redis..."
    
    if [ ! -f "$BACKUP_FILE" ]; then
        log_error "备份文件不存在: $BACKUP_FILE"
        return 1
    fi
    
    # 停止Redis
    docker-compose -f docker-compose.production.yml stop redis
    
    # 复制备份文件
    docker cp "$BACKUP_FILE" im-redis-prod:/data/dump.rdb
    
    # 启动Redis
    docker-compose -f docker-compose.production.yml start redis
    
    log_success "Redis恢复完成"
}

restore_minio() {
    BACKUP_FILE=$1
    
    log_info "开始恢复MinIO..."
    
    if [ ! -f "$BACKUP_FILE" ]; then
        log_error "备份文件不存在: $BACKUP_FILE"
        return 1
    fi
    
    # 解压备份
    TEMP_DIR=$(mktemp -d)
    tar -xzf "$BACKUP_FILE" -C "$TEMP_DIR"
    
    # 恢复数据
    for bucket_dir in "$TEMP_DIR"/*; do
        bucket=$(basename "$bucket_dir")
        log_info "恢复bucket: $bucket"
        
        # 创建bucket
        mc mb local/$bucket --insecure || true
        
        # 恢复数据
        mc mirror --overwrite "$bucket_dir" local/$bucket --insecure
    done
    
    rm -rf "$TEMP_DIR"
    log_success "MinIO恢复完成"
}

do_restore() {
    TIMESTAMP=$1
    
    if [ -z "$TIMESTAMP" ]; then
        log_info "可用的备份："
        echo ""
        echo "MySQL备份:"
        ls -lht "$MYSQL_BACKUP_DIR"/*.sql.gz 2>/dev/null | head -5 | awk '{print $NF}'
        echo ""
        echo "Redis备份:"
        ls -lht "$REDIS_BACKUP_DIR"/*.rdb 2>/dev/null | head -5 | awk '{print $NF}'
        echo ""
        echo "MinIO备份:"
        ls -lht "$MINIO_BACKUP_DIR"/*.tar.gz 2>/dev/null | head -5 | awk '{print $NF}'
        echo ""
        read -p "请输入备份时间戳（格式：20251011-150000）: " TIMESTAMP
    fi
    
    log_info "========================================="
    log_info "开始恢复"
    log_info "目标时间: $TIMESTAMP"
    log_info "========================================="
    
    read -p "确认恢复？这将覆盖当前数据！(yes/no): " CONFIRM
    
    if [ "$CONFIRM" != "yes" ]; then
        log_info "恢复已取消"
        exit 0
    fi
    
    # 恢复数据
    restore_mysql "${MYSQL_BACKUP_DIR}/mysql-${TIMESTAMP}.sql.gz"
    restore_redis "${REDIS_BACKUP_DIR}/redis-${TIMESTAMP}.rdb"
    restore_minio "${MINIO_BACKUP_DIR}/minio-${TIMESTAMP}.tar.gz"
    
    log_success "========================================="
    log_success "恢复完成！"
    log_success "========================================="
}

###############################################################################
# 清理函数
###############################################################################

cleanup_old_backups() {
    DAYS=${1:-$DAILY_RETENTION}
    
    log_info "清理 ${DAYS} 天前的备份..."
    
    # 清理MySQL备份
    find "$MYSQL_BACKUP_DIR" -name "*.sql.gz" -mtime +$DAYS -delete
    MYSQL_DELETED=$(find "$MYSQL_BACKUP_DIR" -name "*.sql.gz" -mtime +$DAYS | wc -l)
    
    # 清理Redis备份
    find "$REDIS_BACKUP_DIR" -name "*.rdb" -mtime +$DAYS -delete
    REDIS_DELETED=$(find "$REDIS_BACKUP_DIR" -name "*.rdb" -mtime +$DAYS | wc -l)
    
    # 清理MinIO备份
    find "$MINIO_BACKUP_DIR" -name "*.tar.gz" -mtime +$DAYS -delete
    MINIO_DELETED=$(find "$MINIO_BACKUP_DIR" -name "*.tar.gz" -mtime +$DAYS | wc -l)
    
    log_success "清理完成: MySQL($MYSQL_DELETED), Redis($REDIS_DELETED), MinIO($MINIO_DELETED)"
}

###############################################################################
# 主函数
###############################################################################

case "$1" in
    backup)
        do_backup
        ;;
    restore)
        do_restore "$2"
        ;;
    cleanup)
        cleanup_old_backups "$2"
        ;;
    *)
        echo "用法: $0 {backup|restore|cleanup} [参数]"
        echo ""
        echo "命令："
        echo "  backup           - 执行全量备份"
        echo "  restore [时间]   - 恢复到指定时间点"
        echo "  cleanup [天数]   - 清理旧备份（默认${DAILY_RETENTION}天）"
        echo ""
        echo "示例："
        echo "  $0 backup"
        echo "  $0 restore 20251011-150000"
        echo "  $0 cleanup 30"
        exit 1
        ;;
esac

