#!/bin/bash

###############################################################################
# 志航密信 - 压力测试脚本
# 用途：对API进行压力测试，验证SLO目标
# 使用：bash ops/loadtest.sh [OPTIONS]
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
BASE_URL="${BASE_URL:-http://localhost:8080}"
REPORT_DIR="./loadtest-reports"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
REPORT_FILE="${REPORT_DIR}/report-${TIMESTAMP}.json"

mkdir -p "$REPORT_DIR"

# SLO目标
TARGET_P95_MS=200
TARGET_SUCCESS_RATE=99.9
TARGET_CONCURRENT_USERS=300

log_info "========================================="
log_info "志航密信 - 压力测试"
log_info "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
log_info "========================================="

###############################################################################
# 检查依赖
###############################################################################

check_dependencies() {
    log_info "检查测试工具..."
    
    MISSING_TOOLS=""
    
    # 检查 ab (Apache Bench)
    if ! command -v ab &> /dev/null; then
        MISSING_TOOLS="$MISSING_TOOLS apache2-utils(ab)"
    fi
    
    # 检查 wrk（可选）
    if ! command -v wrk &> /dev/null; then
        log_warning "wrk未安装（可选工具）"
    fi
    
    # 检查 jq
    if ! command -v jq &> /dev/null; then
        MISSING_TOOLS="$MISSING_TOOLS jq"
    fi
    
    if [ -n "$MISSING_TOOLS" ]; then
        log_error "缺少依赖:$MISSING_TOOLS"
        log_info "安装命令: sudo apt-get install apache2-utils jq"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

###############################################################################
# 健康检查端点测试
###############################################################################

test_health_endpoint() {
    log_info "测试健康检查端点..."
    
    REQUESTS=1000
    CONCURRENCY=50
    
    ab -n $REQUESTS -c $CONCURRENCY -g "${REPORT_DIR}/health-${TIMESTAMP}.tsv" \
        "${BASE_URL}/health" > "${REPORT_DIR}/health-${TIMESTAMP}.txt" 2>&1
    
    # 解析结果
    TIME_PER_REQUEST=$(grep "Time per request:" "${REPORT_DIR}/health-${TIMESTAMP}.txt" | head -1 | awk '{print $4}')
    REQUESTS_PER_SEC=$(grep "Requests per second:" "${REPORT_DIR}/health-${TIMESTAMP}.txt" | awk '{print $4}')
    FAILED_REQUESTS=$(grep "Failed requests:" "${REPORT_DIR}/health-${TIMESTAMP}.txt" | awk '{print $3}')
    
    log_info "结果: ${TIME_PER_REQUEST}ms/请求, ${REQUESTS_PER_SEC}req/s, ${FAILED_REQUESTS}失败"
    
    echo "{\"endpoint\": \"/health\", \"avg_ms\": $TIME_PER_REQUEST, \"rps\": $REQUESTS_PER_SEC, \"failed\": $FAILED_REQUESTS}" >> "$REPORT_FILE.tmp"
}

###############################################################################
# 登录API测试
###############################################################################

test_login_endpoint() {
    log_info "测试登录API..."
    
    # 准备测试数据
    LOGIN_DATA='{"phone":"13800138000","password":"Test@123456"}'
    echo "$LOGIN_DATA" > "${REPORT_DIR}/login-data.json"
    
    REQUESTS=500
    CONCURRENCY=50
    
    # 使用ab进行POST测试
    ab -n $REQUESTS -c $CONCURRENCY \
        -p "${REPORT_DIR}/login-data.json" \
        -T "application/json" \
        -g "${REPORT_DIR}/login-${TIMESTAMP}.tsv" \
        "${BASE_URL}/api/auth/login" > "${REPORT_DIR}/login-${TIMESTAMP}.txt" 2>&1
    
    # 解析结果
    TIME_PER_REQUEST=$(grep "Time per request:" "${REPORT_DIR}/login-${TIMESTAMP}.txt" | head -1 | awk '{print $4}')
    REQUESTS_PER_SEC=$(grep "Requests per second:" "${REPORT_DIR}/login-${TIMESTAMP}.txt" | awk '{print $4}')
    FAILED_REQUESTS=$(grep "Failed requests:" "${REPORT_DIR}/login-${TIMESTAMP}.txt" | awk '{print $3}')
    P95=$(awk 'NR>1 {print $5}' "${REPORT_DIR}/login-${TIMESTAMP}.tsv" | sort -n | awk 'BEGIN{c=0} {total[c]=$1; c++} END{print total[int(c*0.95-0.5)]}')
    
    log_info "结果: ${TIME_PER_REQUEST}ms平均, P95=${P95}ms, ${REQUESTS_PER_SEC}req/s, ${FAILED_REQUESTS}失败"
    
    # 检查SLO
    if (( $(echo "$P95 > $TARGET_P95_MS" | bc -l) )); then
        log_warning "⚠️ P95延迟超标: ${P95}ms > ${TARGET_P95_MS}ms"
    else
        log_success "✓ P95延迟达标: ${P95}ms ≤ ${TARGET_P95_MS}ms"
    fi
    
    echo "{\"endpoint\": \"/api/auth/login\", \"avg_ms\": $TIME_PER_REQUEST, \"p95_ms\": $P95, \"rps\": $REQUESTS_PER_SEC, \"failed\": $FAILED_REQUESTS}" >> "$REPORT_FILE.tmp"
}

###############################################################################
# 发消息API测试
###############################################################################

test_message_endpoint() {
    log_info "测试发消息API..."
    
    # 需要先获取token
    TOKEN=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"phone":"13800138000","password":"Test@123456"}' \
        | jq -r '.data.token' 2>/dev/null || echo "")
    
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        log_warning "无法获取token，跳过消息API测试"
        return
    fi
    
    # 准备测试数据
    MESSAGE_DATA='{"receiver_id":2,"content":"压测消息","message_type":"text"}'
    echo "$MESSAGE_DATA" > "${REPORT_DIR}/message-data.json"
    
    REQUESTS=300
    CONCURRENCY=30
    
    # 使用wrk测试（如果可用）
    if command -v wrk &> /dev/null; then
        wrk -t10 -c30 -d30s \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -s <(cat <<'EOF'
wrk.method = "POST"
wrk.body = '{"receiver_id":2,"content":"压测消息","message_type":"text"}'
EOF
            ) \
            "${BASE_URL}/api/messages/send" > "${REPORT_DIR}/message-${TIMESTAMP}.txt"
        
        log_info "wrk测试完成，查看: ${REPORT_DIR}/message-${TIMESTAMP}.txt"
    else
        log_warning "wrk未安装，跳过发消息压测"
    fi
}

###############################################################################
# 并发用户测试
###############################################################################

test_concurrent_users() {
    log_info "测试并发用户..."
    
    CONCURRENT=$TARGET_CONCURRENT_USERS
    DURATION=60  # 60秒
    
    log_info "模拟 ${CONCURRENT} 个并发用户，持续 ${DURATION} 秒..."
    
    if command -v wrk &> /dev/null; then
        wrk -t20 -c${CONCURRENT} -d${DURATION}s \
            "${BASE_URL}/health" > "${REPORT_DIR}/concurrent-${TIMESTAMP}.txt"
        
        REQUESTS=$(grep "Requests/sec:" "${REPORT_DIR}/concurrent-${TIMESTAMP}.txt" | awk '{print $2}')
        LATENCY_AVG=$(grep "Latency" "${REPORT_DIR}/concurrent-${TIMESTAMP}.txt" | awk '{print $2}')
        
        log_info "结果: ${REQUESTS} req/s, 平均延迟 ${LATENCY_AVG}"
        
        echo "{\"test\": \"concurrent_users\", \"users\": $CONCURRENT, \"rps\": \"$REQUESTS\", \"latency\": \"$LATENCY_AVG\"}" >> "$REPORT_FILE.tmp"
    else
        log_warning "需要wrk进行并发测试"
    fi
}

###############################################################################
# WebRTC带宽测试（模拟）
###############################################################################

test_webrtc_bandwidth() {
    log_info "测试WebRTC带宽（模拟）..."
    
    # 这是一个简化的测试，实际应使用专门的WebRTC测试工具
    log_warning "WebRTC压测需要专门工具（如Jitsi Meet Torture）"
    log_info "跳过WebRTC压测，请手动验证"
    
    # 目标：20路720p并发，带宽<800Mbps
    echo "{\"test\": \"webrtc\", \"status\": \"manual_required\", \"target\": \"20 streams @ 720p < 800Mbps\"}" >> "$REPORT_FILE.tmp"
}

###############################################################################
# 生成报告
###############################################################################

generate_report() {
    log_info "生成测试报告..."
    
    # 合并所有结果
    echo "{" > "$REPORT_FILE"
    echo "  \"timestamp\": \"$(date -Iseconds)\"," >> "$REPORT_FILE"
    echo "  \"base_url\": \"$BASE_URL\"," >> "$REPORT_FILE"
    echo "  \"slo_targets\": {" >> "$REPORT_FILE"
    echo "    \"p95_ms\": $TARGET_P95_MS," >> "$REPORT_FILE"
    echo "    \"success_rate\": $TARGET_SUCCESS_RATE," >> "$REPORT_FILE"
    echo "    \"concurrent_users\": $TARGET_CONCURRENT_USERS" >> "$REPORT_FILE"
    echo "  }," >> "$REPORT_FILE"
    echo "  \"results\": [" >> "$REPORT_FILE"
    
    # 添加测试结果
    cat "$REPORT_FILE.tmp" | sed 's/$/,/' >> "$REPORT_FILE"
    
    # 删除最后一个逗号并关闭JSON
    sed -i '$ s/,$//' "$REPORT_FILE"
    echo "  ]" >> "$REPORT_FILE"
    echo "}" >> "$REPORT_FILE"
    
    rm -f "$REPORT_FILE.tmp"
    
    log_success "报告已生成: $REPORT_FILE"
    
    # 输出摘要
    echo ""
    log_info "========================================="
    log_info "测试摘要"
    log_info "========================================="
    jq '.' "$REPORT_FILE"
}

###############################################################################
# 主函数
###############################################################################

main() {
    check_dependencies
    
    # 检查服务是否运行
    if ! curl -f -s "${BASE_URL}/health" > /dev/null 2>&1; then
        log_error "服务未运行: $BASE_URL"
        exit 1
    fi
    
    log_success "服务运行正常，开始压测..."
    echo ""
    
    # 执行测试
    test_health_endpoint
    test_login_endpoint
    test_message_endpoint
    test_concurrent_users
    test_webrtc_bandwidth
    
    # 生成报告
    generate_report
    
    log_success "========================================="
    log_success "压力测试完成！"
    log_success "========================================="
    echo ""
    echo "报告位置: $REPORT_FILE"
    echo "详细日志: $REPORT_DIR"
    echo ""
    echo "查看报告: jq '.' $REPORT_FILE"
    echo ""
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --url)
            BASE_URL="$2"
            shift 2
            ;;
        --help)
            echo "用法: $0 [OPTIONS]"
            echo ""
            echo "选项:"
            echo "  --url URL    设置基础URL（默认: http://localhost:8080）"
            echo "  --help       显示帮助"
            echo ""
            echo "示例:"
            echo "  $0"
            echo "  $0 --url https://api.example.com"
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            exit 1
            ;;
    esac
done

main

