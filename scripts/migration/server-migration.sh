#!/bin/bash

# 志航密信 - 服务器迁移脚本
# 支持快速服务器迁移，确保数据完整性和服务连续性

set -e

# 配置参数
SOURCE_SERVER="192.168.1.10"
TARGET_SERVER="192.168.1.11"
DOMAIN="zhihang-messenger.com"
MYSQL_USER="root"
MYSQL_PASSWORD="your_password"
MYSQL_DATABASE="zhihang_messenger"
REDIS_PASSWORD="your_redis_password"
MINIO_ACCESS_KEY="your_access_key"
MINIO_SECRET_KEY="your_secret_key"
BACKUP_DIR="/data/migration_backup"
LOG_FILE="/var/log/server-migration.log"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 错误处理
error_exit() {
    log "错误: $1"
    exit 1
}

# 检查服务器连接
check_server_connection() {
    local server="$1"
    local server_name="$2"
    
    log "检查服务器连接: $server_name ($server)"
    
    if ! ping -c 3 "$server" > /dev/null 2>&1; then
        error_exit "无法连接到服务器: $server_name ($server)"
    fi
    
    # 检查SSH连接
    if ! ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no "root@$server" "echo 'SSH连接正常'" > /dev/null 2>&1; then
        error_exit "SSH连接失败: $server_name ($server)"
    fi
    
    log "服务器连接检查通过: $server_name"
}

# 创建备份目录
create_backup_dirs() {
    log "创建备份目录..."
    
    ssh "root@$SOURCE_SERVER" "mkdir -p $BACKUP_DIR/{mysql,redis,minio,config,logs}"
    ssh "root@$TARGET_SERVER" "mkdir -p $BACKUP_DIR/{mysql,redis,minio,config,logs}"
    
    log "备份目录创建完成"
}

# 停止源服务器服务
stop_source_services() {
    log "停止源服务器服务..."
    
    ssh "root@$SOURCE_SERVER" << 'EOF'
        # 停止应用服务
        docker-compose down || true
        
        # 停止MySQL服务
        systemctl stop mysql || true
        
        # 停止Redis服务
        systemctl stop redis || true
        
        # 停止MinIO服务
        systemctl stop minio || true
        
        # 停止Nginx服务
        systemctl stop nginx || true
        
        echo "源服务器服务已停止"
EOF
    
    log "源服务器服务停止完成"
}

# 备份MySQL数据
backup_mysql_data() {
    log "开始备份MySQL数据..."
    
    ssh "root@$SOURCE_SERVER" << EOF
        # 创建MySQL备份
        mysqldump -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" \
            --single-transaction \
            --routines \
            --triggers \
            --events \
            --hex-blob \
            --master-data=2 \
            --all-databases > "$BACKUP_DIR/mysql/full_backup_\$(date +%Y%m%d_%H%M%S).sql"
        
        # 压缩备份文件
        gzip "$BACKUP_DIR/mysql/full_backup_"*.sql
        
        echo "MySQL数据备份完成"
EOF
    
    log "MySQL数据备份完成"
}

# 备份Redis数据
backup_redis_data() {
    log "开始备份Redis数据..."
    
    ssh "root@$SOURCE_SERVER" << EOF
        # 触发Redis保存
        redis-cli -a "$REDIS_PASSWORD" BGSAVE
        
        # 等待保存完成
        while [ "\$(redis-cli -a "$REDIS_PASSWORD" LASTSAVE)" = "\$(redis-cli -a "$REDIS_PASSWORD" LASTSAVE)" ]; do
            sleep 1
        done
        
        # 复制RDB文件
        cp /var/lib/redis/dump.rdb "$BACKUP_DIR/redis/redis_backup_\$(date +%Y%m%d_%H%M%S).rdb"
        
        echo "Redis数据备份完成"
EOF
    
    log "Redis数据备份完成"
}

# 备份MinIO数据
backup_minio_data() {
    log "开始备份MinIO数据..."
    
    ssh "root@$SOURCE_SERVER" << EOF
        # 使用MinIO client备份数据
        mc mirror --overwrite --remove \
            /var/lib/minio/zhihang-messenger \
            "$BACKUP_DIR/minio/zhihang-messenger_backup_\$(date +%Y%m%d_%H%M%S)/"
        
        # 压缩备份
        tar -czf "$BACKUP_DIR/minio/minio_backup_\$(date +%Y%m%d_%H%M%S).tar.gz" \
            -C "$BACKUP_DIR/minio" "zhihang-messenger_backup_"*
        
        # 删除临时目录
        rm -rf "$BACKUP_DIR/minio/zhihang-messenger_backup_"*
        
        echo "MinIO数据备份完成"
EOF
    
    log "MinIO数据备份完成"
}

# 备份配置文件
backup_config_files() {
    log "开始备份配置文件..."
    
    ssh "root@$SOURCE_SERVER" << 'EOF'
        # 备份配置文件
        tar -czf "$BACKUP_DIR/config/config_backup_$(date +%Y%m%d_%H%M%S).tar.gz" \
            /etc/nginx/ \
            /etc/mysql/ \
            /etc/redis/ \
            /opt/zhihang-messenger/configs/ \
            /opt/zhihang-messenger/docker-compose.yml \
            /opt/zhihang-messenger/docker-stack.yml \
            2>/dev/null || true
        
        echo "配置文件备份完成"
EOF
    
    log "配置文件备份完成"
}

# 传输数据到目标服务器
transfer_data_to_target() {
    log "开始传输数据到目标服务器..."
    
    # 传输MySQL备份
    log "传输MySQL数据..."
    rsync -avz --progress \
        "root@$SOURCE_SERVER:$BACKUP_DIR/mysql/" \
        "root@$TARGET_SERVER:$BACKUP_DIR/mysql/"
    
    # 传输Redis备份
    log "传输Redis数据..."
    rsync -avz --progress \
        "root@$SOURCE_SERVER:$BACKUP_DIR/redis/" \
        "root@$TARGET_SERVER:$BACKUP_DIR/redis/"
    
    # 传输MinIO备份
    log "传输MinIO数据..."
    rsync -avz --progress \
        "root@$SOURCE_SERVER:$BACKUP_DIR/minio/" \
        "root@$TARGET_SERVER:$BACKUP_DIR/minio/"
    
    # 传输配置文件
    log "传输配置文件..."
    rsync -avz --progress \
        "root@$SOURCE_SERVER:$BACKUP_DIR/config/" \
        "root@$TARGET_SERVER:$BACKUP_DIR/config/"
    
    log "数据传输完成"
}

# 在目标服务器恢复数据
restore_data_on_target() {
    log "开始在目标服务器恢复数据..."
    
    ssh "root@$TARGET_SERVER" << EOF
        # 停止现有服务
        docker-compose down || true
        systemctl stop mysql || true
        systemctl stop redis || true
        systemctl stop minio || true
        
        # 恢复MySQL数据
        log "恢复MySQL数据..."
        LATEST_MYSQL_BACKUP=\$(ls -t "$BACKUP_DIR/mysql/full_backup_"*.gz | head -1)
        if [ -f "\$LATEST_MYSQL_BACKUP" ]; then
            gunzip < "\$LATEST_MYSQL_BACKUP" | mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD"
            echo "MySQL数据恢复完成"
        fi
        
        # 恢复Redis数据
        log "恢复Redis数据..."
        LATEST_REDIS_BACKUP=\$(ls -t "$BACKUP_DIR/redis/redis_backup_"*.rdb | head -1)
        if [ -f "\$LATEST_REDIS_BACKUP" ]; then
            cp "\$LATEST_REDIS_BACKUP" /var/lib/redis/dump.rdb
            chown redis:redis /var/lib/redis/dump.rdb
            echo "Redis数据恢复完成"
        fi
        
        # 恢复MinIO数据
        log "恢复MinIO数据..."
        LATEST_MINIO_BACKUP=\$(ls -t "$BACKUP_DIR/minio/minio_backup_"*.tar.gz | head -1)
        if [ -f "\$LATEST_MINIO_BACKUP" ]; then
            tar -xzf "\$LATEST_MINIO_BACKUP" -C /var/lib/minio/
            chown -R minio:minio /var/lib/minio/
            echo "MinIO数据恢复完成"
        fi
        
        # 恢复配置文件
        log "恢复配置文件..."
        LATEST_CONFIG_BACKUP=\$(ls -t "$BACKUP_DIR/config/config_backup_"*.tar.gz | head -1)
        if [ -f "\$LATEST_CONFIG_BACKUP" ]; then
            tar -xzf "\$LATEST_CONFIG_BACKUP" -C /
            echo "配置文件恢复完成"
        fi
        
        echo "数据恢复完成"
EOF
    
    log "目标服务器数据恢复完成"
}

# 启动目标服务器服务
start_target_services() {
    log "启动目标服务器服务..."
    
    ssh "root@$TARGET_SERVER" << 'EOF'
        # 启动基础服务
        systemctl start mysql
        systemctl start redis
        systemctl start minio
        
        # 等待服务启动
        sleep 10
        
        # 启动应用服务
        docker-compose up -d
        
        # 启动Nginx
        systemctl start nginx
        
        echo "目标服务器服务启动完成"
EOF
    
    log "目标服务器服务启动完成"
}

# 验证迁移结果
verify_migration() {
    log "验证迁移结果..."
    
    # 检查服务状态
    ssh "root@$TARGET_SERVER" << 'EOF'
        echo "检查服务状态:"
        systemctl status mysql --no-pager -l
        systemctl status redis --no-pager -l
        systemctl status minio --no-pager -l
        systemctl status nginx --no-pager -l
        
        echo "检查Docker容器状态:"
        docker-compose ps
EOF
    
    # 检查数据库连接
    log "检查数据库连接..."
    ssh "root@$TARGET_SERVER" "mysql -u$MYSQL_USER -p$MYSQL_PASSWORD -e 'SHOW DATABASES;'"
    
    # 检查Redis连接
    log "检查Redis连接..."
    ssh "root@$TARGET_SERVER" "redis-cli -a $REDIS_PASSWORD ping"
    
    # 检查HTTP服务
    log "检查HTTP服务..."
    local http_status=$(ssh "root@$TARGET_SERVER" "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/health" || echo "000")
    
    if [ "$http_status" = "200" ]; then
        log "HTTP服务检查通过"
    else
        error_exit "HTTP服务检查失败 (状态码: $http_status)"
    fi
    
    log "迁移验证完成"
}

# 更新DNS记录
update_dns_for_migration() {
    log "更新DNS记录指向目标服务器..."
    
    # 调用DNS切换脚本
    ./scripts/dns/dns-switch.sh switch-backup
    
    log "DNS记录更新完成"
}

# 清理备份文件
cleanup_backup_files() {
    log "清理备份文件..."
    
    # 清理源服务器备份文件
    ssh "root@$SOURCE_SERVER" "rm -rf $BACKUP_DIR"
    
    # 清理目标服务器备份文件
    ssh "root@$TARGET_SERVER" "rm -rf $BACKUP_DIR"
    
    log "备份文件清理完成"
}

# 发送迁移通知
send_migration_notification() {
    local status="$1"
    local message="$2"
    
    log "发送迁移通知: $status - $message"
    
    # 发送邮件通知
    echo "服务器迁移状态: $status
消息: $message
时间: $(date)
源服务器: $SOURCE_SERVER
目标服务器: $TARGET_SERVER" | mail -s "志航密信服务器迁移通知" "admin@zhihang-messenger.com"
    
    # 发送用户通知 (可选)
    # curl -X POST "http://$TARGET_SERVER:8080/api/notifications" \
    #     -H "Content-Type: application/json" \
    #     -d "{\"type\":\"migration\",\"status\":\"$status\",\"message\":\"$message\"}"
}

# 主函数
main() {
    log "开始服务器迁移流程..."
    
    # 检查参数
    if [ $# -lt 1 ]; then
        echo "用法: $0 {migrate|verify|cleanup}"
        echo "  migrate - 执行完整迁移"
        echo "  verify  - 验证迁移结果"
        echo "  cleanup - 清理备份文件"
        exit 1
    fi
    
    case "$1" in
        "migrate")
            log "执行完整迁移流程..."
            
            # 检查服务器连接
            check_server_connection "$SOURCE_SERVER" "源服务器"
            check_server_connection "$TARGET_SERVER" "目标服务器"
            
            # 创建备份目录
            create_backup_dirs
            
            # 停止源服务器服务
            stop_source_services
            
            # 备份数据
            backup_mysql_data
            backup_redis_data
            backup_minio_data
            backup_config_files
            
            # 传输数据
            transfer_data_to_target
            
            # 恢复数据
            restore_data_on_target
            
            # 启动服务
            start_target_services
            
            # 验证迁移
            verify_migration
            
            # 更新DNS
            update_dns_for_migration
            
            # 发送成功通知
            send_migration_notification "成功" "服务器迁移完成，服务已切换到目标服务器"
            
            log "服务器迁移完成"
            ;;
        "verify")
            verify_migration
            ;;
        "cleanup")
            cleanup_backup_files
            ;;
        *)
            echo "用法: $0 {migrate|verify|cleanup}"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
