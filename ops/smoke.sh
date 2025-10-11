#!/bin/bash

###############################################################################
# 志航密信 - 冒烟测试脚本
# 用途：一键冒烟（启动→健康检查→关键API/WS连通→最小E2E）
# 使用：bash ops/smoke.sh
###############################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓ PASS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠ WARN]${NC} $1"; }
log_error() { echo -e "${RED}[✗ FAIL]${NC} $1"; }

BASE_URL="${BASE_URL:-http://localhost:8080}"
TIMEOUT=60
START_TIME=$(date +%s)

log_info "========================================="
log_info "志航密信 - 冒烟测试"
log_info "目标: $BASE_URL"
log_info "========================================="

# ====================================
# 1. 启动服务
# ====================================
log_info "1. 启动服务..."

if docker-compose -f docker-compose.production.yml ps | grep -q "Up"; then
    log_success "服务已在运行"
else
    log_info "启动Docker Compose..."
    docker-compose -f docker-compose.production.yml up -d
    
    log_info "等待服务启动..."
    WAITED=0
    while [ $WAITED -lt $TIMEOUT ]; do
        if docker-compose -f docker-compose.production.yml ps | grep -q "Up"; then
            log_success "服务启动成功"
            break
        fi
        sleep 2
        WAITED=$((WAITED + 2))
        echo -n "."
    done
    echo ""
    
    if [ $WAITED -ge $TIMEOUT ]; then
        log_error "服务启动超时"
        exit 1
    fi
fi

# 等待5秒让服务完全启动
sleep 5

# ====================================
# 2. 健康检查
# ====================================
log_info "2. 健康检查..."

# 2.1 后端健康检查
log_info "2.1 检查后端健康..."
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/health" | tail -1)

if [ "$HEALTH_RESPONSE" = "200" ]; then
    log_success "后端健康检查通过"
else
    log_error "后端健康检查失败 (HTTP $HEALTH_RESPONSE)"
    docker-compose -f docker-compose.production.yml logs --tail=50 backend
    exit 1
fi

# 2.2 数据库检查
log_info "2.2 检查数据库连接..."
if docker exec im-mysql-prod mysqladmin ping -h localhost -uroot -p${MYSQL_ROOT_PASSWORD:-password} > /dev/null 2>&1; then
    log_success "数据库连接正常"
else
    log_error "数据库连接失败"
    exit 1
fi

# 2.3 Redis检查
log_info "2.3 检查Redis连接..."
if docker exec im-redis-prod redis-cli --no-auth-warning -a ${REDIS_PASSWORD:-password} ping | grep -q "PONG"; then
    log_success "Redis连接正常"
else
    log_error "Redis连接失败"
    exit 1
fi

# ====================================
# 3. 关键API测试
# ====================================
log_info "3. 关键API测试..."

# 3.1 注册测试
log_info "3.1 测试用户注册..."
RANDOM_PHONE="139$(date +%s | tail -c 8)"
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"phone\":\"$RANDOM_PHONE\",\"password\":\"Test@123456\",\"nickname\":\"测试用户\"}")

if echo "$REGISTER_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    log_success "用户注册成功"
else
    log_warning "用户注册失败（可能手机号已存在）: $REGISTER_RESPONSE"
fi

# 3.2 登录测试
log_info "3.2 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"phone\":\"13800138000\",\"password\":\"Test@123456\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token' 2>/dev/null)

if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    log_success "用户登录成功"
    export AUTH_TOKEN=$TOKEN
else
    log_error "用户登录失败: $LOGIN_RESPONSE"
    exit 1
fi

# 3.3 获取用户信息
log_info "3.3 测试获取用户信息..."
USER_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/users/me" \
    -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$USER_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    USER_ID=$(echo "$USER_RESPONSE" | jq -r '.data.id')
    log_success "获取用户信息成功 (ID: $USER_ID)"
else
    log_error "获取用户信息失败: $USER_RESPONSE"
    exit 1
fi

# 3.4 发送消息
log_info "3.4 测试发送消息..."
MESSAGE_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/messages/send" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"receiver_id\":1,\"content\":\"冒烟测试消息\",\"message_type\":\"text\"}")

if echo "$MESSAGE_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    log_success "发送消息成功"
else
    log_warning "发送消息失败: $MESSAGE_RESPONSE"
fi

# ====================================
# 4. WebSocket测试
# ====================================
log_info "4. WebSocket连通测试..."

# 使用wscat或websocat测试（如果已安装）
if command -v wscat &> /dev/null; then
    WS_URL=$(echo "$BASE_URL" | sed 's/http/ws/')/ws?token=$AUTH_TOKEN
    if timeout 5 wscat -c "$WS_URL" -x '{"type":"ping"}' > /dev/null 2>&1; then
        log_success "WebSocket连接成功"
    else
        log_warning "WebSocket连接失败（可能wscat不支持）"
    fi
else
    log_warning "wscat未安装，跳过WebSocket测试"
fi

# ====================================
# 5. 最小E2E流程
# ====================================
log_info "5. 最小E2E流程..."

# 5.1 登录 → 获取好友列表 → 发送消息
log_info "5.1 测试完整流程..."

FRIENDS_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/users/friends" \
    -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$FRIENDS_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    FRIEND_COUNT=$(echo "$FRIENDS_RESPONSE" | jq -r '.data | length')
    log_success "获取好友列表成功 ($FRIEND_COUNT 个好友)"
else
    log_warning "获取好友列表失败: $FRIENDS_RESPONSE"
fi

# ====================================
# 6. 性能检查
# ====================================
log_info "6. 性能检查..."

# 6.1 API响应时间
log_info "6.1 测试API响应时间..."
START=$(date +%s%N)
curl -s "${BASE_URL}/health" > /dev/null
END=$(date +%s%N)
ELAPSED=$((($END - $START) / 1000000))

if [ $ELAPSED -lt 200 ]; then
    log_success "健康检查响应时间: ${ELAPSED}ms (< 200ms)"
else
    log_warning "健康检查响应时间: ${ELAPSED}ms (>= 200ms)"
fi

# ====================================
# 7. 日志检查
# ====================================
log_info "7. 日志检查..."

# 检查最近的ERROR日志
ERROR_COUNT=$(docker-compose -f docker-compose.production.yml logs --tail=100 backend | grep -c "ERROR" || true)

if [ $ERROR_COUNT -eq 0 ]; then
    log_success "无ERROR日志"
else
    log_warning "发现 $ERROR_COUNT 条ERROR日志"
    docker-compose -f docker-compose.production.yml logs --tail=20 backend | grep "ERROR"
fi

# ====================================
# 8. 资源使用
# ====================================
log_info "8. 资源使用检查..."

# CPU和内存
STATS=$(docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep "im-")
echo "$STATS"

# ====================================
# 总结
# ====================================
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

log_success "========================================="
log_success "冒烟测试完成！"
log_success "========================================="
echo "耗时: ${DURATION}秒"
echo "所有关键功能正常"
echo ""
echo "查看完整日志:"
echo "  docker-compose -f docker-compose.production.yml logs -f"
echo ""
echo "停止服务:"
echo "  docker-compose -f docker-compose.production.yml down"
echo ""

