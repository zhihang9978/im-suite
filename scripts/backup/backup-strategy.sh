#!/bin/bash

# 志航密信 - 数据备份和同步策略
# 支持实时备份、增量备份、跨服务器同步

set -e

# 配置参数
BACKUP_DIR="/data/backups"
MYSQL_HOST="localhost"
MYSQL_USER="root"
MYSQL_PASSWORD="your_password"
MYSQL_DATABASE="zhihang_messenger"
REDIS_HOST="localhost"
REDIS_PORT="6379"
MINIO_ENDPOINT="localhost:9000"
MINIO_ACCESS_KEY="your_access_key"
MINIO_SECRET_KEY="your_secret_key"
SLAVE_SERVER="backup.example.com"
LOG_FILE="/var/log/backup.log"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 创建备份目录
create_backup_dirs() {
    log "创建备份目录..."
    mkdir -p "$BACKUP_DIR"/{mysql,redis,minio,config}
    mkdir -p "$BACKUP_DIR"/mysql/{full,incremental}
    mkdir -p "$BACKUP_DIR"/redis/{full,incremental}
    mkdir -p "$BACKUP_DIR"/minio/{full,incremental}
}

# MySQL 全量备份
mysql_full_backup() {
    log "开始 MySQL 全量备份..."
    BACKUP_FILE="$BACKUP_DIR/mysql/full/zhihang_messenger_$(date +%Y%m%d_%H%M%S).sql"
    
    mysqldump -h"$MYSQL_HOST" -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" \
        --single-transaction \
        --routines \
        --triggers \
        --events \
        --hex-blob \
        --master-data=2 \
        "$MYSQL_DATABASE" > "$BACKUP_FILE"
    
    # 压缩备份文件
    gzip "$BACKUP_FILE"
    log "MySQL 全量备份完成: ${BACKUP_FILE}.gz"
}

# MySQL 增量备份
mysql_incremental_backup() {
    log "开始 MySQL 增量备份..."
    BACKUP_FILE="$BACKUP_DIR/mysql/incremental/zhihang_messenger_inc_$(date +%Y%m%d_%H%M%S).sql"
    
    # 获取 binlog 位置
    BINLOG_POS=$(mysql -h"$MYSQL_HOST" -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" \
        -e "SHOW MASTER STATUS\G" | grep Position | awk '{print $2}')
    
    # 增量备份
    mysqlbinlog --start-position="$BINLOG_POS" \
        --stop-datetime="$(date '+%Y-%m-%d %H:%M:%S')" \
        /var/lib/mysql/mysql-bin.* > "$BACKUP_FILE"
    
    gzip "$BACKUP_FILE"
    log "MySQL 增量备份完成: ${BACKUP_FILE}.gz"
}

# Redis 备份
redis_backup() {
    log "开始 Redis 备份..."
    BACKUP_FILE="$BACKUP_DIR/redis/full/redis_$(date +%Y%m%d_%H%M%S).rdb"
    
    # 触发 Redis 保存
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" BGSAVE
    
    # 等待保存完成
    while [ "$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LASTSAVE)" = "$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LASTSAVE)" ]; do
        sleep 1
    done
    
    # 复制 RDB 文件
    cp /var/lib/redis/dump.rdb "$BACKUP_FILE"
    gzip "$BACKUP_FILE"
    log "Redis 备份完成: ${BACKUP_FILE}.gz"
}

# MinIO 备份
minio_backup() {
    log "开始 MinIO 备份..."
    BACKUP_DIR_MINIO="$BACKUP_DIR/minio/full/minio_$(date +%Y%m%d_%H%M%S)"
    
    mkdir -p "$BACKUP_DIR_MINIO"
    
    # 使用 MinIO client 同步数据
    mc mirror --overwrite --remove \
        "$MINIO_ENDPOINT/zhihang-messenger" \
        "$BACKUP_DIR_MINIO/"
    
    # 压缩备份
    tar -czf "${BACKUP_DIR_MINIO}.tar.gz" -C "$BACKUP_DIR/minio/full" "$(basename "$BACKUP_DIR_MINIO")"
    rm -rf "$BACKUP_DIR_MINIO"
    
    log "MinIO 备份完成: ${BACKUP_DIR_MINIO}.tar.gz"
}

# 配置文件备份
config_backup() {
    log "开始配置文件备份..."
    CONFIG_BACKUP="$BACKUP_DIR/config/config_$(date +%Y%m%d_%H%M%S).tar.gz"
    
    tar -czf "$CONFIG_BACKUP" \
        /etc/nginx/ \
        /etc/mysql/ \
        /etc/redis/ \
        /opt/zhihang-messenger/configs/ \
        /opt/zhihang-messenger/docker-compose.yml \
        /opt/zhihang-messenger/docker-stack.yml \
        2>/dev/null || true
    
    log "配置文件备份完成: $CONFIG_BACKUP"
}

# 数据同步到备用服务器
sync_to_slave() {
    log "开始同步数据到备用服务器..."
    
    # 同步 MySQL 数据
    rsync -avz --delete \
        "$BACKUP_DIR/mysql/" \
        "root@$SLAVE_SERVER:/data/backups/mysql/"
    
    # 同步 Redis 数据
    rsync -avz --delete \
        "$BACKUP_DIR/redis/" \
        "root@$SLAVE_SERVER:/data/backups/redis/"
    
    # 同步 MinIO 数据
    rsync -avz --delete \
        "$BACKUP_DIR/minio/" \
        "root@$SLAVE_SERVER:/data/backups/minio/"
    
    # 同步配置文件
    rsync -avz --delete \
        "$BACKUP_DIR/config/" \
        "root@$SLAVE_SERVER:/data/backups/config/"
    
    log "数据同步到备用服务器完成"
}

# 清理旧备份
cleanup_old_backups() {
    log "清理旧备份文件..."
    
    # 保留最近 7 天的全量备份
    find "$BACKUP_DIR/mysql/full" -name "*.gz" -mtime +7 -delete
    find "$BACKUP_DIR/redis/full" -name "*.gz" -mtime +7 -delete
    find "$BACKUP_DIR/minio/full" -name "*.tar.gz" -mtime +7 -delete
    
    # 保留最近 3 天的增量备份
    find "$BACKUP_DIR/mysql/incremental" -name "*.gz" -mtime +3 -delete
    
    # 保留最近 30 天的配置文件备份
    find "$BACKUP_DIR/config" -name "*.tar.gz" -mtime +30 -delete
    
    log "旧备份文件清理完成"
}

# 验证备份完整性
verify_backup() {
    log "验证备份完整性..."
    
    # 验证 MySQL 备份
    LATEST_MYSQL_BACKUP=$(find "$BACKUP_DIR/mysql/full" -name "*.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
    if [ -f "$LATEST_MYSQL_BACKUP" ]; then
        gunzip -t "$LATEST_MYSQL_BACKUP"
        log "MySQL 备份验证通过"
    fi
    
    # 验证 Redis 备份
    LATEST_REDIS_BACKUP=$(find "$BACKUP_DIR/redis/full" -name "*.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
    if [ -f "$LATEST_REDIS_BACKUP" ]; then
        gunzip -t "$LATEST_REDIS_BACKUP"
        log "Redis 备份验证通过"
    fi
    
    log "备份完整性验证完成"
}

# 主函数
main() {
    log "开始数据备份流程..."
    
    create_backup_dirs
    
    case "$1" in
        "full")
            mysql_full_backup
            redis_backup
            minio_backup
            config_backup
            ;;
        "incremental")
            mysql_incremental_backup
            ;;
        "sync")
            sync_to_slave
            ;;
        "cleanup")
            cleanup_old_backups
            ;;
        "verify")
            verify_backup
            ;;
        "all")
            mysql_full_backup
            redis_backup
            minio_backup
            config_backup
            sync_to_slave
            cleanup_old_backups
            verify_backup
            ;;
        *)
            echo "用法: $0 {full|incremental|sync|cleanup|verify|all}"
            echo "  full        - 全量备份"
            echo "  incremental - 增量备份"
            echo "  sync        - 同步到备用服务器"
            echo "  cleanup     - 清理旧备份"
            echo "  verify      - 验证备份完整性"
            echo "  all         - 执行所有操作"
            exit 1
            ;;
    esac
    
    log "数据备份流程完成"
}

# 执行主函数
main "$@"
