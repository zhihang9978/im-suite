#!/bin/bash

###############################################################################
# 志航密信 v1.6.0 - 项目完整性检查脚本
# 用途：检查所有必需的文件和配置是否齐全
# 使用：bash scripts/check-project-integrity.sh
###############################################################################

# 颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 统计
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

check_pass() {
    echo -e "  ${GREEN}✅${NC} $1"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
}

check_fail() {
    echo -e "  ${RED}❌${NC} $1"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
}

check_warn() {
    echo -e "  ${YELLOW}⚠️${NC} $1"
    WARNINGS=$((WARNINGS + 1))
}

run_check() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         志航密信 v1.6.0 - 项目完整性检查                   ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

###############################################################################
# 1. 后端文件检查
###############################################################################

echo -e "${BLUE}[1/10] 后端核心文件${NC}"

# 核心文件
files=(
    "im-backend/main.go"
    "im-backend/go.mod"
    "im-backend/go.sum"
    "im-backend/config/database.go"
    "im-backend/config/redis.go"
)

for file in "${files[@]}"; do
    run_check
    if [ -f "$file" ]; then
        check_pass "$file"
    else
        check_fail "$file 不存在"
    fi
done

###############################################################################
# 2. 屏幕共享相关文件
###############################################################################

echo ""
echo -e "${BLUE}[2/10] 屏幕共享功能文件${NC}"

files=(
    "im-backend/internal/model/screen_share.go"
    "im-backend/internal/service/webrtc_service.go"
    "im-backend/internal/service/screen_share_enhanced_service.go"
    "im-backend/internal/controller/webrtc_controller.go"
    "im-backend/internal/controller/screen_share_enhanced_controller.go"
)

for file in "${files[@]}"; do
    run_check
    if [ -f "$file" ]; then
        check_pass "$file"
    else
        check_fail "$file 不存在"
    fi
done

###############################################################################
# 3. 权限管理文件
###############################################################################

echo ""
echo -e "${BLUE}[3/10] 权限管理文件${NC}"

files=(
    "telegram-android/TMessagesProj/src/main/java/org/telegram/messenger/PermissionManager.java"
    "telegram-android/TMessagesProj/src/main/java/org/telegram/ui/PermissionExampleActivity.java"
)

for file in "${files[@]}"; do
    run_check
    if [ -f "$file" ]; then
        check_pass "$file"
    else
        check_warn "$file 不存在（Android端文件）"
    fi
done

###############################################################################
# 4. 前端示例文件
###############################################################################

echo ""
echo -e "${BLUE}[4/10] 前端示例文件${NC}"

files=(
    "examples/screen-share-example.js"
    "examples/screen-share-enhanced.js"
    "examples/chinese-phone-permissions.js"
    "examples/screen-share-demo.html"
)

for file in "${files[@]}"; do
    run_check
    if [ -f "$file" ]; then
        check_pass "$file"
    else
        check_fail "$file 不存在"
    fi
done

###############################################################################
# 5. 文档文件
###############################################################################

echo ""
echo -e "${BLUE}[5/10] 文档文件${NC}"

files=(
    "SCREEN_SHARE_FEATURE.md"
    "SCREEN_SHARE_ENHANCED.md"
    "SCREEN_SHARE_ENHANCEMENT_SUMMARY.md"
    "SCREEN_SHARE_QUICK_START.md"
    "PERMISSION_SYSTEM_COMPLETE.md"
    "COMPLETE_SUMMARY_v1.6.0.md"
    "DEPLOYMENT_FOR_DEVIN.md"
    "docs/chinese-phones/permission-request-guide.md"
    "docs/chinese-phones/screen-share-permissions.md"
    "examples/SCREEN_SHARE_README.md"
    "examples/QUICK_TEST.md"
)

for file in "${files[@]}"; do
    run_check
    if [ -f "$file" ]; then
        check_pass "$file"
    else
        check_fail "$file 不存在"
    fi
done

###############################################################################
# 6. 配置文件
###############################################################################

echo ""
echo -e "${BLUE}[6/10] 配置文件${NC}"

files=(
    "config/mysql/conf.d/custom.cnf"
    "config/mysql/init/01-init.sql"
    "config/redis/redis.conf"
    "config/nginx/nginx.conf"
    "docker-compose.production.yml"
)

for file in "${files[@]}"; do
    run_check
    if [ -f "$file" ]; then
        check_pass "$file"
    else
        check_fail "$file 不存在"
    fi
done

###############################################################################
# 7. 环境变量检查
###############################################################################

echo ""
echo -e "${BLUE}[7/10] 环境变量${NC}"

run_check
if [ -f ".env" ]; then
    check_pass ".env 文件存在"
    
    # 检查必需的变量
    required_vars=("DB_HOST" "DB_USER" "DB_PASSWORD" "DB_NAME" "JWT_SECRET")
    for var in "${required_vars[@]}"; do
        run_check
        if grep -q "^$var=" .env; then
            check_pass "$var 已配置"
        else
            check_fail "$var 未配置"
        fi
    done
else
    check_warn ".env 文件不存在（首次运行会自动创建）"
fi

###############################################################################
# 8. Go依赖检查
###############################################################################

echo ""
echo -e "${BLUE}[8/10] Go依赖检查${NC}"

cd im-backend

run_check
if go mod verify > /dev/null 2>&1; then
    check_pass "Go依赖验证通过"
else
    check_fail "Go依赖验证失败"
fi

run_check
if go build -o /tmp/test-build main.go > /dev/null 2>&1; then
    check_pass "Go代码编译通过"
    rm -f /tmp/test-build
else
    check_fail "Go代码编译失败"
fi

cd ..

###############################################################################
# 9. 目录结构检查
###############################################################################

echo ""
echo -e "${BLUE}[9/10] 目录结构${NC}"

dirs=(
    "im-backend/internal/model"
    "im-backend/internal/service"
    "im-backend/internal/controller"
    "im-backend/internal/middleware"
    "examples"
    "docs/chinese-phones"
    "config/mysql"
    "config/redis"
    "logs"
)

for dir in "${dirs[@]}"; do
    run_check
    if [ -d "$dir" ]; then
        check_pass "$dir/"
    else
        if [ "$dir" = "logs" ]; then
            check_warn "$dir/ 不存在（首次运行会创建）"
            mkdir -p logs
        else
            check_fail "$dir/ 不存在"
        fi
    fi
done

###############################################################################
# 10. API路由检查
###############################################################################

echo ""
echo -e "${BLUE}[10/10] API路由检查${NC}"

run_check
if grep -q "webrtcController.StartScreenShare" im-backend/main.go; then
    check_pass "屏幕共享路由已配置"
else
    check_fail "屏幕共享路由未配置"
fi

run_check
if grep -q "screen_share_enhanced_controller" im-backend/main.go 2>/dev/null; then
    check_pass "增强API路由已配置"
else
    check_warn "增强API路由未配置（需要在main.go中添加）"
fi

###############################################################################
# 总结
###############################################################################

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                  检查完成！                                 ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "总检查数:  $TOTAL_CHECKS"
echo -e "通过:      ${GREEN}$PASSED_CHECKS${NC}"
echo -e "失败:      ${RED}$FAILED_CHECKS${NC}"
echo -e "警告:      ${YELLOW}$WARNINGS${NC}"
echo ""

# 计算完整性百分比
INTEGRITY=$(echo "scale=2; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc)
echo -e "完整性:    ${GREEN}$INTEGRITY%${NC}"
echo ""

if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "${GREEN}✅ 项目完整，可以开始部署！${NC}"
    echo ""
    echo -e "下一步："
    echo -e "  1. 运行部署脚本:   ${BLUE}bash scripts/auto-deploy.sh${NC}"
    echo -e "  2. 运行测试脚本:   ${BLUE}bash scripts/auto-test.sh${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}❌ 项目不完整，有 $FAILED_CHECKS 个文件缺失${NC}"
    echo ""
    echo -e "请检查缺失的文件并补充"
    echo ""
    exit 1
fi


