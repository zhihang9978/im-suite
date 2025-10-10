#!/bin/bash

# ========================================
# 志航密信 - 自动备份脚本
# ========================================
# 功能：自动备份 MySQL、Redis、MinIO 数据
# 使用：每日凌晨 2:00 执行
# Crontab: 0 2 * * * /root/im-suite/scripts/backup/auto-backup.sh
# ========================================

set -e

# 配置变量
BACKUP_DIR="/root/im-suite/backups"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7  # 保留最近7天的备份

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# ========================================
# 1. 创建备份目录
# ========================================
log "创建备份目录..."
mkdir -p ${BACKUP_DIR}/mysql
mkdir -p ${BACKUP_DIR}/redis
mkdir -p ${BACKUP_DIR}/minio
mkdir -p ${BACKUP_DIR}/config

# ========================================
# 2. 备份 MySQL 数据库
# ========================================
log "开始备份 MySQL 数据库..."
if docker ps | grep -q im-mysql-prod; then
    docker exec im-mysql-prod mysqldump \
        -u root \
        -p"ZhRoot2024SecurePass!@#" \
        --all-databases \
        --single-transaction \
        --quick \
        --lock-tables=false \
        > ${BACKUP_DIR}/mysql/full_backup_${DATE}.sql
    
    # 压缩备份文件
    gzip ${BACKUP_DIR}/mysql/full_backup_${DATE}.sql
    
    MYSQL_SIZE=$(du -h ${BACKUP_DIR}/mysql/full_backup_${DATE}.sql.gz | cut -f1)
    log "MySQL 备份完成！大小: ${MYSQL_SIZE}"
else
    error "MySQL 容器未运行，跳过备份"
fi

# ========================================
# 3. 备份 Redis 数据
# ========================================
log "开始备份 Redis 数据..."
if docker ps | grep -q im-redis-prod; then
    # 触发 Redis 保存
    docker exec im-redis-prod redis-cli -a "ZhRedis2024SecurePass!@#" BGSAVE
    sleep 5
    
    # 复制 RDB 文件
    docker cp im-redis-prod:/data/dump.rdb ${BACKUP_DIR}/redis/dump_${DATE}.rdb
    gzip ${BACKUP_DIR}/redis/dump_${DATE}.rdb
    
    REDIS_SIZE=$(du -h ${BACKUP_DIR}/redis/dump_${DATE}.rdb.gz | cut -f1)
    log "Redis 备份完成！大小: ${REDIS_SIZE}"
else
    error "Redis 容器未运行，跳过备份"
fi

# ========================================
# 4. 备份 MinIO 数据
# ========================================
log "开始备份 MinIO 数据..."
if docker ps | grep -q im-minio-prod; then
    tar -czf ${BACKUP_DIR}/minio/minio_data_${DATE}.tar.gz \
        -C /var/lib/docker/volumes/im-suite_minio_data/_data .
    
    MINIO_SIZE=$(du -h ${BACKUP_DIR}/minio/minio_data_${DATE}.tar.gz | cut -f1)
    log "MinIO 备份完成！大小: ${MINIO_SIZE}"
else
    error "MinIO 容器未运行，跳过备份"
fi

# ========================================
# 5. 备份配置文件
# ========================================
log "开始备份配置文件..."
cd /root/im-suite
tar -czf ${BACKUP_DIR}/config/config_${DATE}.tar.gz \
    .env \
    docker-compose.production.yml \
    docker-compose.partial.yml \
    config/ \
    nginx.conf 2>/dev/null || true

CONFIG_SIZE=$(du -h ${BACKUP_DIR}/config/config_${DATE}.tar.gz | cut -f1)
log "配置文件备份完成！大小: ${CONFIG_SIZE}"

# ========================================
# 6. 清理旧备份（保留最近N天）
# ========================================
log "清理 ${RETENTION_DAYS} 天前的备份..."

find ${BACKUP_DIR}/mysql -name "*.sql.gz" -mtime +${RETENTION_DAYS} -delete
find ${BACKUP_DIR}/redis -name "*.rdb.gz" -mtime +${RETENTION_DAYS} -delete
find ${BACKUP_DIR}/minio -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete
find ${BACKUP_DIR}/config -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete

log "旧备份清理完成"

# ========================================
# 7. 生成备份报告
# ========================================
cat > ${BACKUP_DIR}/backup_report_${DATE}.txt << EOF
========================================
志航密信备份报告
========================================
备份时间: $(date)
服务器: $(hostname)

备份文件:
- MySQL: ${BACKUP_DIR}/mysql/full_backup_${DATE}.sql.gz (${MYSQL_SIZE})
- Redis: ${BACKUP_DIR}/redis/dump_${DATE}.rdb.gz (${REDIS_SIZE})
- MinIO: ${BACKUP_DIR}/minio/minio_data_${DATE}.tar.gz (${MINIO_SIZE})
- Config: ${BACKUP_DIR}/config/config_${DATE}.tar.gz (${CONFIG_SIZE})

总大小: $(du -sh ${BACKUP_DIR} | cut -f1)
保留策略: ${RETENTION_DAYS} 天

========================================
备份完成！
========================================
EOF

log "========================================="
log "✅ 所有备份完成！"
log "备份目录: ${BACKUP_DIR}"
log "备份报告: ${BACKUP_DIR}/backup_report_${DATE}.txt"
log "========================================="

# ========================================
# 8. 可选：上传到远程服务器
# ========================================
# 取消下面的注释以启用远程备份
# log "上传备份到远程服务器..."
# rsync -avz --progress ${BACKUP_DIR}/ user@remote-server:/remote/backups/im-suite/
# log "远程备份完成"

# ========================================
# 9. 可选：上传到云存储（阿里云 OSS）
# ========================================
# 取消下面的注释以启用云备份
# log "上传备份到阿里云 OSS..."
# ossutil cp -r ${BACKUP_DIR}/mysql/full_backup_${DATE}.sql.gz oss://your-bucket/backups/mysql/
# ossutil cp -r ${BACKUP_DIR}/redis/dump_${DATE}.rdb.gz oss://your-bucket/backups/redis/
# log "云备份完成"

exit 0

