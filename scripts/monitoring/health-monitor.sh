#!/bin/bash

# 志航密信 - 健康监控和自动切换系统
# 实时监控服务器状态，自动故障转移，确保服务高可用性

set -e

# 配置参数
PRIMARY_SERVER="192.168.1.10"
BACKUP_SERVER="192.168.1.11"
MONITOR_SERVER="192.168.1.12"
DOMAIN="zhihang-messenger.com"
CHECK_INTERVAL=30
FAILURE_THRESHOLD=3
RECOVERY_THRESHOLD=2
LOG_FILE="/var/log/health-monitor.log"
STATUS_FILE="/var/run/health-monitor.status"
ALERT_EMAIL="admin@zhihang-messenger.com"
ALERT_PHONE="13800138000"

# 服务器状态
declare -A SERVER_STATUS
declare -A FAILURE_COUNT
declare -A RECOVERY_COUNT
CURRENT_PRIMARY="$PRIMARY_SERVER"
LAST_FAILOVER_TIME=0
MIN_FAILOVER_INTERVAL=300  # 5分钟最小故障转移间隔

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 发送告警
send_alert() {
    local level="$1"
    local message="$2"
    local server="$3"
    
    log "告警 [$level]: $message (服务器: $server)"
    
    # 发送邮件告警
    echo "告警级别: $level
服务器: $server
消息: $message
时间: $(date)
当前主服务器: $CURRENT_PRIMARY" | mail -s "志航密信服务器告警 [$level]" "$ALERT_EMAIL"
    
    # 发送短信告警 (紧急级别)
    if [ "$level" = "CRITICAL" ]; then
        curl -X POST "https://api.sms-provider.com/send" \
            -H "Content-Type: application/json" \
            -d "{\"to\":\"$ALERT_PHONE\",\"message\":\"[紧急]志航密信服务器故障: $message\"}" \
            2>/dev/null || true
    fi
    
    # 发送钉钉/企业微信通知
    curl -X POST "https://oapi.dingtalk.com/robot/send?access_token=your_token" \
        -H "Content-Type: application/json" \
        -d "{\"msgtype\":\"text\",\"text\":{\"content\":\"[志航密信] $level: $message\"}}" \
        2>/dev/null || true
}

# 检查服务器健康状态
check_server_health() {
    local server="$1"
    local server_name="$2"
    local health_score=0
    
    log "检查服务器健康状态: $server_name ($server)"
    
    # 检查网络连通性
    if ping -c 3 -W 5 "$server" > /dev/null 2>&1; then
        health_score=$((health_score + 20))
        log "网络连通性检查通过: $server_name"
    else
        log "网络连通性检查失败: $server_name"
        return 1
    fi
    
    # 检查SSH连接
    if timeout 10 ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no "root@$server" "echo 'SSH OK'" > /dev/null 2>&1; then
        health_score=$((health_score + 20))
        log "SSH连接检查通过: $server_name"
    else
        log "SSH连接检查失败: $server_name"
        return 1
    fi
    
    # 检查HTTP服务
    local http_status=$(timeout 10 curl -s -o /dev/null -w "%{http_code}" "http://$server:8080/api/health" 2>/dev/null || echo "000")
    if [ "$http_status" = "200" ]; then
        health_score=$((health_score + 20))
        log "HTTP服务检查通过: $server_name (状态码: $http_status)"
    else
        log "HTTP服务检查失败: $server_name (状态码: $http_status)"
    fi
    
    # 检查WebSocket服务
    if timeout 5 nc -z "$server" 8081; then
        health_score=$((health_score + 20))
        log "WebSocket服务检查通过: $server_name"
    else
        log "WebSocket服务检查失败: $server_name"
    fi
    
    # 检查数据库服务
    if timeout 5 nc -z "$server" 3306; then
        health_score=$((health_score + 10))
        log "数据库服务检查通过: $server_name"
    else
        log "数据库服务检查失败: $server_name"
    fi
    
    # 检查Redis服务
    if timeout 5 nc -z "$server" 6379; then
        health_score=$((health_score + 10))
        log "Redis服务检查通过: $server_name"
    else
        log "Redis服务检查失败: $server_name"
    fi
    
    log "服务器 $server_name 健康评分: $health_score/100"
    
    if [ $health_score -ge 80 ]; then
        return 0  # 健康
    else
        return 1  # 不健康
    fi
}

# 检查服务器资源使用情况
check_server_resources() {
    local server="$1"
    local server_name="$2"
    
    log "检查服务器资源使用情况: $server_name"
    
    ssh "root@$server" << 'EOF' | while read line; do
        echo "CPU使用率: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)"
        echo "内存使用率: $(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')"
        echo "磁盘使用率: $(df -h / | awk 'NR==2{print $5}' | cut -d'%' -f1)"
        echo "负载平均值: $(uptime | awk -F'load average:' '{print $2}')"
    done
EOF
    
    # 检查资源使用率
    local cpu_usage=$(ssh "root@$server" "top -bn1 | grep 'Cpu(s)' | awk '{print \$2}' | cut -d'%' -f1" 2>/dev/null || echo "0")
    local memory_usage=$(ssh "root@$server" "free | grep Mem | awk '{printf \"%.1f\", \$3/\$2 * 100.0}'" 2>/dev/null || echo "0")
    local disk_usage=$(ssh "root@$server" "df -h / | awk 'NR==2{print \$5}' | cut -d'%' -f1" 2>/dev/null || echo "0")
    
    # 资源告警
    if (( $(echo "$cpu_usage > 90" | bc -l) )); then
        send_alert "WARNING" "CPU使用率过高: ${cpu_usage}%" "$server_name"
    fi
    
    if (( $(echo "$memory_usage > 90" | bc -l) )); then
        send_alert "WARNING" "内存使用率过高: ${memory_usage}%" "$server_name"
    fi
    
    if [ "$disk_usage" -gt 90 ]; then
        send_alert "WARNING" "磁盘使用率过高: ${disk_usage}%" "$server_name"
    fi
    
    log "服务器 $server_name 资源使用情况 - CPU: ${cpu_usage}%, 内存: ${memory_usage}%, 磁盘: ${disk_usage}%"
}

# 检查应用服务状态
check_application_services() {
    local server="$1"
    local server_name="$2"
    
    log "检查应用服务状态: $server_name"
    
    # 检查Docker容器状态
    local container_status=$(ssh "root@$server" "docker-compose ps --services --filter 'status=running'" 2>/dev/null || echo "")
    
    if [ -n "$container_status" ]; then
        log "Docker容器状态正常: $server_name"
    else
        log "Docker容器状态异常: $server_name"
        send_alert "CRITICAL" "Docker容器服务异常" "$server_name"
    fi
    
    # 检查关键服务
    local services=("mysql" "redis" "nginx" "minio")
    for service in "${services[@]}"; do
        local service_status=$(ssh "root@$server" "systemctl is-active $service" 2>/dev/null || echo "inactive")
        
        if [ "$service_status" = "active" ]; then
            log "服务 $service 状态正常: $server_name"
        else
            log "服务 $service 状态异常: $server_name"
            send_alert "CRITICAL" "服务 $service 异常" "$server_name"
        fi
    done
}

# 执行故障转移
execute_failover() {
    local failed_server="$1"
    local target_server="$2"
    
    log "执行故障转移: $failed_server -> $target_server"
    
    # 检查最小故障转移间隔
    local current_time=$(date +%s)
    if [ $((current_time - LAST_FAILOVER_TIME)) -lt $MIN_FAILOVER_INTERVAL ]; then
        log "故障转移间隔过短，跳过本次转移"
        return 1
    fi
    
    # 更新DNS记录
    log "更新DNS记录指向目标服务器: $target_server"
    ./scripts/dns/dns-switch.sh switch-backup
    
    # 更新当前主服务器
    CURRENT_PRIMARY="$target_server"
    LAST_FAILOVER_TIME=$current_time
    
    # 发送故障转移通知
    send_alert "CRITICAL" "执行故障转移: $failed_server -> $target_server" "$target_server"
    
    log "故障转移完成: $target_server"
}

# 检查所有服务器
check_all_servers() {
    local servers=("$PRIMARY_SERVER:$BACKUP_SERVER:$MONITOR_SERVER")
    local server_names=("主服务器:备用服务器:监控服务器")
    local healthy_servers=()
    local failed_servers=()
    
    IFS=':' read -ra SERVER_LIST <<< "$servers"
    IFS=':' read -ra NAME_LIST <<< "$server_names"
    
    for i in "${!SERVER_LIST[@]}"; do
        local server="${SERVER_LIST[$i]}"
        local server_name="${NAME_LIST[$i]}"
        
        if check_server_health "$server" "$server_name"; then
            healthy_servers+=("$server")
            SERVER_STATUS["$server"]="healthy"
            FAILURE_COUNT["$server"]=0
            RECOVERY_COUNT["$server"]=$((RECOVERY_COUNT["$server"] + 1))
            
            # 检查资源使用情况
            check_server_resources "$server" "$server_name"
            
            # 检查应用服务状态
            check_application_services "$server" "$server_name"
            
            log "服务器 $server_name 状态: 健康"
        else
            failed_servers+=("$server")
            SERVER_STATUS["$server"]="failed"
            FAILURE_COUNT["$server"]=$((FAILURE_COUNT["$server"] + 1))
            RECOVERY_COUNT["$server"]=0
            
            log "服务器 $server_name 状态: 故障 (连续失败: ${FAILURE_COUNT["$server"]})"
            
            # 检查是否需要故障转移
            if [ "${FAILURE_COUNT["$server"]}" -ge "$FAILURE_THRESHOLD" ]; then
                if [ "$server" = "$CURRENT_PRIMARY" ]; then
                    # 主服务器故障，寻找健康的备用服务器
                    for healthy_server in "${healthy_servers[@]}"; do
                        if [ "$healthy_server" != "$server" ]; then
                            execute_failover "$server" "$healthy_server"
                            break
                        fi
                    done
                fi
            fi
        fi
    done
    
    # 更新状态文件
    update_status_file
}

# 更新状态文件
update_status_file() {
    cat > "$STATUS_FILE" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "current_primary": "$CURRENT_PRIMARY",
    "last_failover": "$LAST_FAILOVER_TIME",
    "servers": {
        "$PRIMARY_SERVER": {
            "status": "${SERVER_STATUS["$PRIMARY_SERVER"]:-unknown}",
            "failure_count": "${FAILURE_COUNT["$PRIMARY_SERVER"]:-0}",
            "recovery_count": "${RECOVERY_COUNT["$PRIMARY_SERVER"]:-0}"
        },
        "$BACKUP_SERVER": {
            "status": "${SERVER_STATUS["$BACKUP_SERVER"]:-unknown}",
            "failure_count": "${FAILURE_COUNT["$BACKUP_SERVER"]:-0}",
            "recovery_count": "${RECOVERY_COUNT["$BACKUP_SERVER"]:-0}"
        },
        "$MONITOR_SERVER": {
            "status": "${SERVER_STATUS["$MONITOR_SERVER"]:-unknown}",
            "failure_count": "${FAILURE_COUNT["$MONITOR_SERVER"]:-0}",
            "recovery_count": "${RECOVERY_COUNT["$MONITOR_SERVER"]:-0}"
        }
    }
}
EOF
}

# 生成监控报告
generate_monitoring_report() {
    local report_file="/var/log/monitoring-report-$(date +%Y%m%d).log"
    
    log "生成监控报告: $report_file"
    
    cat > "$report_file" << EOF
志航密信服务器监控报告
生成时间: $(date)
监控周期: 24小时

服务器状态摘要:
- 主服务器 ($PRIMARY_SERVER): ${SERVER_STATUS["$PRIMARY_SERVER"]:-unknown}
- 备用服务器 ($BACKUP_SERVER): ${SERVER_STATUS["$BACKUP_SERVER"]:-unknown}
- 监控服务器 ($MONITOR_SERVER): ${SERVER_STATUS["$MONITOR_SERVER"]:-unknown}

故障统计:
- 主服务器故障次数: ${FAILURE_COUNT["$PRIMARY_SERVER"]:-0}
- 备用服务器故障次数: ${FAILURE_COUNT["$BACKUP_SERVER"]:-0}
- 监控服务器故障次数: ${FAILURE_COUNT["$MONITOR_SERVER"]:-0}

恢复统计:
- 主服务器恢复次数: ${RECOVERY_COUNT["$PRIMARY_SERVER"]:-0}
- 备用服务器恢复次数: ${RECOVERY_COUNT["$BACKUP_SERVER"]:-0}
- 监控服务器恢复次数: ${RECOVERY_COUNT["$MONITOR_SERVER"]:-0}

当前主服务器: $CURRENT_PRIMARY
最后故障转移时间: $(date -d "@$LAST_FAILOVER_TIME" 2>/dev/null || echo "从未")

EOF
    
    # 发送报告邮件
    mail -s "志航密信服务器监控报告 - $(date +%Y-%m-%d)" "$ALERT_EMAIL" < "$report_file"
}

# 主监控循环
main_monitoring_loop() {
    log "启动健康监控系统..."
    
    # 初始化状态
    for server in "$PRIMARY_SERVER" "$BACKUP_SERVER" "$MONITOR_SERVER"; do
        SERVER_STATUS["$server"]="unknown"
        FAILURE_COUNT["$server"]=0
        RECOVERY_COUNT["$server"]=0
    done
    
    # 主循环
    while true; do
        log "开始健康检查循环..."
        
        check_all_servers
        
        # 每小时生成报告
        if [ $(($(date +%M) % 60)) -eq 0 ]; then
            generate_monitoring_report
        fi
        
        log "健康检查循环完成，等待 $CHECK_INTERVAL 秒..."
        sleep "$CHECK_INTERVAL"
    done
}

# 主函数
main() {
    case "$1" in
        "start")
            log "启动健康监控系统..."
            main_monitoring_loop
            ;;
        "check")
            log "执行单次健康检查..."
            check_all_servers
            ;;
        "status")
            if [ -f "$STATUS_FILE" ]; then
                cat "$STATUS_FILE" | jq .
            else
                echo "状态文件不存在"
            fi
            ;;
        "report")
            generate_monitoring_report
            ;;
        *)
            echo "用法: $0 {start|check|status|report}"
            echo "  start  - 启动监控系统"
            echo "  check  - 执行单次检查"
            echo "  status - 显示当前状态"
            echo "  report - 生成监控报告"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
