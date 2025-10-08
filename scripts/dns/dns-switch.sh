#!/bin/bash

# 志航密信 - DNS快速切换脚本
# 支持自动和手动DNS记录切换，确保服务连续性

set -e

# 配置参数
DOMAIN="zhihang-messenger.com"
DNS_PROVIDER="cloudflare"  # cloudflare, aliyun, tencent, aws
API_TOKEN="your_dns_api_token"
API_EMAIL="your_email@example.com"
PRIMARY_SERVER="192.168.1.10"
BACKUP_SERVER="192.168.1.11"
MONITOR_SERVER="192.168.1.12"
LOG_FILE="/var/log/dns-switch.log"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# 检查DNS记录
check_dns_record() {
    local record_type="$1"
    local record_name="$2"
    
    log "检查DNS记录: $record_name.$DOMAIN ($record_type)"
    
    case "$DNS_PROVIDER" in
        "cloudflare")
            curl -s -X GET "https://api.cloudflare.com/client/v4/zones?name=$DOMAIN" \
                -H "Authorization: Bearer $API_TOKEN" \
                -H "Content-Type: application/json" | \
                jq -r '.result[0].id' | \
                xargs -I {} curl -s -X GET "https://api.cloudflare.com/client/v4/zones/{}/dns_records?type=$record_type&name=$record_name.$DOMAIN" \
                -H "Authorization: Bearer $API_TOKEN" \
                -H "Content-Type: application/json"
            ;;
        "aliyun")
            # 阿里云DNS API
            curl -s "https://alidns.aliyuncs.com/?Action=DescribeDomainRecords&DomainName=$DOMAIN&Type=$record_type&RR=$record_name&AccessKeyId=your_access_key&Signature=your_signature"
            ;;
        "tencent")
            # 腾讯云DNS API
            curl -s "https://cns.tencentcloudapi.com/" \
                -H "Content-Type: application/json" \
                -d "{\"Action\":\"DescribeRecordList\",\"Domain\":\"$DOMAIN\",\"Subdomain\":\"$record_name\"}"
            ;;
        *)
            log "不支持的DNS提供商: $DNS_PROVIDER"
            return 1
            ;;
    esac
}

# 更新DNS记录
update_dns_record() {
    local record_type="$1"
    local record_name="$2"
    local new_value="$3"
    local ttl="${4:-300}"
    
    log "更新DNS记录: $record_name.$DOMAIN -> $new_value"
    
    case "$DNS_PROVIDER" in
        "cloudflare")
            # 获取Zone ID
            ZONE_ID=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones?name=$DOMAIN" \
                -H "Authorization: Bearer $API_TOKEN" \
                -H "Content-Type: application/json" | \
                jq -r '.result[0].id')
            
            # 获取Record ID
            RECORD_ID=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records?type=$record_type&name=$record_name.$DOMAIN" \
                -H "Authorization: Bearer $API_TOKEN" \
                -H "Content-Type: application/json" | \
                jq -r '.result[0].id')
            
            # 更新记录
            curl -s -X PUT "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records/$RECORD_ID" \
                -H "Authorization: Bearer $API_TOKEN" \
                -H "Content-Type: application/json" \
                --data "{\"type\":\"$record_type\",\"name\":\"$record_name.$DOMAIN\",\"content\":\"$new_value\",\"ttl\":$ttl}"
            ;;
        "aliyun")
            # 阿里云DNS更新
            curl -s "https://alidns.aliyuncs.com/?Action=UpdateDomainRecord&RecordId=your_record_id&RR=$record_name&Type=$record_type&Value=$new_value&TTL=$ttl&AccessKeyId=your_access_key&Signature=your_signature"
            ;;
        "tencent")
            # 腾讯云DNS更新
            curl -s "https://cns.tencentcloudapi.com/" \
                -H "Content-Type: application/json" \
                -d "{\"Action\":\"ModifyRecord\",\"Domain\":\"$DOMAIN\",\"Subdomain\":\"$record_name\",\"RecordType\":\"$record_type\",\"RecordLine\":\"默认\",\"Value\":\"$new_value\",\"TTL\":$ttl}"
            ;;
        *)
            log "不支持的DNS提供商: $DNS_PROVIDER"
            return 1
            ;;
    esac
    
    log "DNS记录更新完成"
}

# 切换到备用服务器
switch_to_backup() {
    log "开始切换到备用服务器..."
    
    # 更新A记录指向备用服务器
    update_dns_record "A" "@" "$BACKUP_SERVER" 60
    update_dns_record "A" "www" "$BACKUP_SERVER" 60
    update_dns_record "A" "api" "$BACKUP_SERVER" 60
    update_dns_record "A" "admin" "$BACKUP_SERVER" 60
    
    # 更新CNAME记录
    update_dns_record "CNAME" "backup" "$BACKUP_SERVER" 60
    
    log "DNS切换完成，指向备用服务器: $BACKUP_SERVER"
    
    # 等待DNS传播
    log "等待DNS传播..."
    sleep 30
    
    # 验证切换结果
    verify_dns_switch "$BACKUP_SERVER"
}

# 切换到主服务器
switch_to_primary() {
    log "开始切换到主服务器..."
    
    # 更新A记录指向主服务器
    update_dns_record "A" "@" "$PRIMARY_SERVER" 300
    update_dns_record "A" "www" "$PRIMARY_SERVER" 300
    update_dns_record "A" "api" "$PRIMARY_SERVER" 300
    update_dns_record "A" "admin" "$PRIMARY_SERVER" 300
    
    log "DNS切换完成，指向主服务器: $PRIMARY_SERVER"
    
    # 等待DNS传播
    log "等待DNS传播..."
    sleep 60
    
    # 验证切换结果
    verify_dns_switch "$PRIMARY_SERVER"
}

# 验证DNS切换结果
verify_dns_switch() {
    local expected_ip="$1"
    
    log "验证DNS切换结果..."
    
    # 检查多个DNS服务器
    local dns_servers=("8.8.8.8" "1.1.1.1" "114.114.114.114")
    
    for dns_server in "${dns_servers[@]}"; do
        local resolved_ip=$(nslookup "$DOMAIN" "$dns_server" | grep "Address:" | tail -1 | awk '{print $2}')
        
        if [ "$resolved_ip" = "$expected_ip" ]; then
            log "DNS服务器 $dns_server 解析正确: $DOMAIN -> $resolved_ip"
        else
            log "DNS服务器 $dns_server 解析异常: $DOMAIN -> $resolved_ip (期望: $expected_ip)"
        fi
    done
    
    # 检查HTTP响应
    local http_status=$(curl -s -o /dev/null -w "%{http_code}" "http://$expected_ip/health" || echo "000")
    
    if [ "$http_status" = "200" ]; then
        log "HTTP健康检查通过: $expected_ip"
    else
        log "HTTP健康检查失败: $expected_ip (状态码: $http_status)"
    fi
}

# 检查服务器健康状态
check_server_health() {
    local server_ip="$1"
    local server_name="$2"
    
    log "检查服务器健康状态: $server_name ($server_ip)"
    
    # 检查HTTP响应
    local http_status=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "http://$server_ip:8080/api/health" || echo "000")
    
    # 检查WebSocket连接
    local ws_status=$(timeout 5 nc -z "$server_ip" 8081 && echo "open" || echo "closed")
    
    # 检查数据库连接
    local db_status=$(timeout 5 nc -z "$server_ip" 3306 && echo "open" || echo "closed")
    
    if [ "$http_status" = "200" ] && [ "$ws_status" = "open" ] && [ "$db_status" = "open" ]; then
        log "服务器 $server_name 健康状态: 正常"
        return 0
    else
        log "服务器 $server_name 健康状态: 异常 (HTTP: $http_status, WS: $ws_status, DB: $db_status)"
        return 1
    fi
}

# 自动故障转移
auto_failover() {
    log "开始自动故障转移检查..."
    
    # 检查主服务器
    if check_server_health "$PRIMARY_SERVER" "主服务器"; then
        log "主服务器正常，无需切换"
        return 0
    fi
    
    log "主服务器故障，检查备用服务器..."
    
    # 检查备用服务器
    if check_server_health "$BACKUP_SERVER" "备用服务器"; then
        log "备用服务器正常，开始切换..."
        switch_to_backup
        return 0
    fi
    
    log "备用服务器也故障，检查监控服务器..."
    
    # 检查监控服务器
    if check_server_health "$MONITOR_SERVER" "监控服务器"; then
        log "监控服务器正常，切换到监控服务器..."
        update_dns_record "A" "@" "$MONITOR_SERVER" 60
        return 0
    fi
    
    log "所有服务器都故障，发送告警..."
    send_alert "所有服务器故障，需要人工干预"
    return 1
}

# 发送告警
send_alert() {
    local message="$1"
    
    log "发送告警: $message"
    
    # 发送邮件
    echo "$message" | mail -s "志航密信服务器告警" "admin@zhihang-messenger.com"
    
    # 发送短信 (需要配置短信服务)
    # curl -X POST "https://api.sms-provider.com/send" \
    #     -H "Content-Type: application/json" \
    #     -d "{\"to\":\"13800138000\",\"message\":\"$message\"}"
    
    # 发送钉钉/企业微信通知
    # curl -X POST "https://oapi.dingtalk.com/robot/send?access_token=your_token" \
    #     -H "Content-Type: application/json" \
    #     -d "{\"msgtype\":\"text\",\"text\":{\"content\":\"$message\"}}"
}

# 主函数
main() {
    log "DNS切换脚本启动"
    
    case "$1" in
        "switch-backup")
            switch_to_backup
            ;;
        "switch-primary")
            switch_to_primary
            ;;
        "check-health")
            check_server_health "$PRIMARY_SERVER" "主服务器"
            check_server_health "$BACKUP_SERVER" "备用服务器"
            check_server_health "$MONITOR_SERVER" "监控服务器"
            ;;
        "auto-failover")
            auto_failover
            ;;
        "verify")
            verify_dns_switch "$2"
            ;;
        *)
            echo "用法: $0 {switch-backup|switch-primary|check-health|auto-failover|verify <ip>}"
            echo "  switch-backup  - 切换到备用服务器"
            echo "  switch-primary - 切换到主服务器"
            echo "  check-health   - 检查所有服务器健康状态"
            echo "  auto-failover  - 自动故障转移"
            echo "  verify <ip>    - 验证DNS切换结果"
            exit 1
            ;;
    esac
    
    log "DNS切换脚本完成"
}

# 执行主函数
main "$@"
