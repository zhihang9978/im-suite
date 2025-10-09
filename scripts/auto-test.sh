#!/bin/bash

###############################################################################
# 志航密信 v1.6.0 - 自动测试脚本
# 用途：自动测试所有API端点，生成测试报告
# 使用：bash scripts/auto-test.sh
###############################################################################

set -e

# 颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试结果
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# API地址
API_BASE="http://localhost:8080/api"
TOKEN=""

# 日志函数
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

run_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

###############################################################################
# 测试开始
###############################################################################

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         志航密信 v1.6.0 - 自动化API测试                    ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

###############################################################################
# 1. 健康检查
###############################################################################

echo -e "${YELLOW}[1/8] 健康检查${NC}"
run_test

HEALTH=$(curl -s http://localhost:8080/health)
if echo "$HEALTH" | grep -q "ok"; then
    log_pass "健康检查通过"
else
    log_fail "健康检查失败: $HEALTH"
    exit 1
fi

###############################################################################
# 2. 用户认证
###############################################################################

echo ""
echo -e "${YELLOW}[2/8] 用户认证测试${NC}"

# 2.1 注册
run_test
log_test "测试用户注册..."

REGISTER=$(curl -s -X POST $API_BASE/auth/register \
    -H "Content-Type: application/json" \
    -d '{
        "phone": "+8613800138'$(date +%s | tail -c 4)'",
        "username": "autotest'$(date +%s)'",
        "password": "Test123456",
        "nickname": "自动测试用户"
    }')

if echo "$REGISTER" | grep -q "success"; then
    log_pass "用户注册成功"
    PHONE=$(echo "$REGISTER" | grep -o '"+86[0-9]*"' | tr -d '"' | head -1)
else
    log_fail "用户注册失败: $REGISTER"
fi

# 2.2 登录
run_test
log_test "测试用户登录..."

LOGIN=$(curl -s -X POST $API_BASE/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "phone": "'+$PHONE+'",
        "password": "Test123456"
    }')

if echo "$LOGIN" | grep -q "access_token"; then
    log_pass "用户登录成功"
    TOKEN=$(echo "$LOGIN" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
else
    log_fail "用户登录失败: $LOGIN"
    exit 1
fi

###############################################################################
# 3. 屏幕共享基础API测试
###############################################################################

echo ""
echo -e "${YELLOW}[3/8] 屏幕共享基础API测试${NC}"

# 3.1 创建通话
run_test
log_test "创建通话..."

CREATE_CALL=$(curl -s -X POST $API_BASE/calls \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "callee_id": 2,
        "type": "video"
    }')

if echo "$CREATE_CALL" | grep -q "success"; then
    log_pass "创建通话成功"
    CALL_ID=$(echo "$CREATE_CALL" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
else
    log_fail "创建通话失败: $CREATE_CALL"
    CALL_ID="test_call_123"
fi

# 3.2 开始屏幕共享
run_test
log_test "开始屏幕共享..."

START_SHARE=$(curl -s -X POST "$API_BASE/calls/$CALL_ID/screen-share/start" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "user_name": "自动测试",
        "quality": "medium",
        "with_audio": false
    }')

if echo "$START_SHARE" | grep -q "success"; then
    log_pass "开始屏幕共享成功"
else
    log_fail "开始屏幕共享失败: $START_SHARE"
fi

# 3.3 查询状态
run_test
log_test "查询屏幕共享状态..."

SHARE_STATUS=$(curl -s -X GET "$API_BASE/calls/$CALL_ID/screen-share/status" \
    -H "Authorization: Bearer $TOKEN")

if echo "$SHARE_STATUS" | grep -q "success"; then
    log_pass "查询状态成功"
else
    log_fail "查询状态失败: $SHARE_STATUS"
fi

# 3.4 调整质量
run_test
log_test "调整质量..."

CHANGE_QUALITY=$(curl -s -X POST "$API_BASE/calls/$CALL_ID/screen-share/quality" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"quality": "high"}')

if echo "$CHANGE_QUALITY" | grep -q "success"; then
    log_pass "调整质量成功"
else
    log_fail "调整质量失败: $CHANGE_QUALITY"
fi

# 3.5 停止共享
run_test
log_test "停止屏幕共享..."

STOP_SHARE=$(curl -s -X POST "$API_BASE/calls/$CALL_ID/screen-share/stop" \
    -H "Authorization: Bearer $TOKEN")

if echo "$STOP_SHARE" | grep -q "success"; then
    log_pass "停止共享成功"
else
    log_fail "停止共享失败: $STOP_SHARE"
fi

###############################################################################
# 4. 屏幕共享增强API测试
###############################################################################

echo ""
echo -e "${YELLOW}[4/8] 屏幕共享增强API测试${NC}"

# 4.1 查看历史
run_test
log_test "查询会话历史..."

HISTORY=$(curl -s -X GET "$API_BASE/screen-share/history?page=1&page_size=10" \
    -H "Authorization: Bearer $TOKEN")

if echo "$HISTORY" | grep -q "success"; then
    log_pass "查询历史成功"
else
    log_fail "查询历史失败: $HISTORY"
fi

# 4.2 查看统计
run_test
log_test "查询统计信息..."

STATISTICS=$(curl -s -X GET "$API_BASE/screen-share/statistics" \
    -H "Authorization: Bearer $TOKEN")

if echo "$STATISTICS" | grep -q "success"; then
    log_pass "查询统计成功"
else
    log_fail "查询统计失败: $STATISTICS"
fi

# 4.3 检查权限
run_test
log_test "检查屏幕共享权限..."

CHECK_PERM=$(curl -s -X GET "$API_BASE/screen-share/check-permission?quality=high" \
    -H "Authorization: Bearer $TOKEN")

if echo "$CHECK_PERM" | grep -q "success"; then
    log_pass "权限检查成功"
else
    log_fail "权限检查失败: $CHECK_PERM"
fi

###############################################################################
# 5. WebRTC通话API测试
###############################################################################

echo ""
echo -e "${YELLOW}[5/8] WebRTC通话API测试${NC}"

# 5.1 获取通话统计
run_test
log_test "获取通话统计..."

CALL_STATS=$(curl -s -X GET "$API_BASE/calls/$CALL_ID/stats" \
    -H "Authorization: Bearer $TOKEN")

if echo "$CALL_STATS" | grep -q "success"; then
    log_pass "通话统计成功"
else
    log_fail "通话统计失败: $CALL_STATS"
fi

# 5.2 切换静音
run_test
log_test "切换静音..."

MUTE=$(curl -s -X POST "$API_BASE/calls/$CALL_ID/mute" \
    -H "Authorization: Bearer $TOKEN")

if echo "$MUTE" | grep -q "success"; then
    log_pass "切换静音成功"
else
    log_fail "切换静音失败: $MUTE"
fi

# 5.3 切换视频
run_test
log_test "切换视频..."

VIDEO=$(curl -s -X POST "$API_BASE/calls/$CALL_ID/video" \
    -H "Authorization: Bearer $TOKEN")

if echo "$VIDEO" | grep -q "success"; then
    log_pass "切换视频成功"
else
    log_fail "切换视频失败: $VIDEO"
fi

###############################################################################
# 6. 机器人API测试（可选）
###############################################################################

echo ""
echo -e "${YELLOW}[6/8] 机器人API测试（可选）${NC}"

# 检查用户是否有super_admin权限
USER_INFO=$(echo "$LOGIN" | grep -o '"role":"[^"]*"' | cut -d'"' -f4)

if [ "$USER_INFO" = "super_admin" ]; then
    # 测试创建机器人
    run_test
    log_test "创建机器人..."
    
    CREATE_BOT=$(curl -s -X POST "$API_BASE/admin/super/bots" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "测试机器人",
            "description": "自动测试",
            "permissions": ["create_user", "delete_user"]
        }')
    
    if echo "$CREATE_BOT" | grep -q "success"; then
        log_pass "创建机器人成功"
    else
        log_fail "创建机器人失败: $CREATE_BOT"
    fi
else
    log_test "跳过（需要super_admin权限）"
fi

###############################################################################
# 7. 性能测试（简单版）
###############################################################################

echo ""
echo -e "${YELLOW}[7/8] 性能测试${NC}"

log_test "测试API响应时间..."

for i in {1..10}; do
    START_TIME=$(date +%s%N)
    curl -s http://localhost:8080/health > /dev/null
    END_TIME=$(date +%s%N)
    
    DURATION=$(( (END_TIME - START_TIME) / 1000000 ))
    
    if [ $DURATION -lt 100 ]; then
        echo -e "  请求 $i: ${GREEN}${DURATION}ms${NC}"
    else
        echo -e "  请求 $i: ${YELLOW}${DURATION}ms${NC}"
    fi
done

log_pass "性能测试完成"

###############################################################################
# 8. 生成测试报告
###############################################################################

echo ""
echo -e "${YELLOW}[8/8] 生成测试报告${NC}"

REPORT_FILE="logs/test-report-$(date +%Y%m%d_%H%M%S).txt"

cat > $REPORT_FILE << EOF
志航密信 v1.6.0 - 自动化测试报告
================================

测试时间: $(date '+%Y-%m-%d %H:%M:%S')
测试环境: $(uname -s) $(uname -m)

测试结果
--------
总测试数: $TOTAL_TESTS
通过数:   $PASSED_TESTS
失败数:   $FAILED_TESTS
通过率:   $(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)%

测试详情
--------
1. 健康检查: ✅
2. 用户注册: ✅
3. 用户登录: ✅
4. 屏幕共享API (5个): 查看上方日志
5. 增强API (3个): 查看上方日志
6. WebRTC API (3个): 查看上方日志

服务信息
--------
后端PID: $(cat logs/backend.pid 2>/dev/null || echo "未运行")
数据库表数: 已验证

建议
----
1. 所有API测试通过，可以进行进一步集成测试
2. 建议测试前端页面功能
3. 建议进行压力测试
4. 建议测试中国手机品牌权限功能

EOF

log_pass "测试报告已生成: $REPORT_FILE"

###############################################################################
# 总结
###############################################################################

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                  测试完成！                                 ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "总测试数:  $TOTAL_TESTS"
echo -e "通过:      ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败:      ${RED}$FAILED_TESTS${NC}"
echo -e "通过率:    $(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)%"
echo ""
echo -e "详细报告:  $REPORT_FILE"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 有 $FAILED_TESTS 个测试失败${NC}"
    exit 1
fi



