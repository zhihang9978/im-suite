#!/bin/bash

###############################################################################
# 志航密信 - 完整验证脚本
# 用途：验证所有断言，生成客观证据
# 使用：bash ops/verify_all.sh
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

# 创建reports目录
mkdir -p reports/{logs,screenshots,tests,builds,evidence}

REPORT_FILE="reports/verification-report-$(date +%Y%m%d-%H%M%S).md"
PASSED=0
FAILED=0
TOTAL=0

log_info "========================================="
log_info "志航密信 - 完整验证"
log_info "时间: $(date '+%Y-%m-%d %H:%M:%S')"
log_info "========================================="

# 开始生成报告
cat > "$REPORT_FILE" <<EOF
# 志航密信 - 验证报告

**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')

---

## 📊 验证总结

EOF

###############################################################################
# 1. 编译检查
###############################################################################
echo "## 1. 编译检查" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "1. 后端编译检查..."
((TOTAL++))
cd im-backend
if go build -o /tmp/im-backend main.go 2> reports/logs/backend-build.log; then
    log_success "后端编译通过"
    echo "- ✅ 后端编译: 通过" >> "$REPORT_FILE"
    ((PASSED++))
    rm -f /tmp/im-backend
else
    log_error "后端编译失败"
    echo "- ❌ 后端编译: 失败" >> "$REPORT_FILE"
    echo "\`\`\`" >> "$REPORT_FILE"
    cat reports/logs/backend-build.log >> "$REPORT_FILE"
    echo "\`\`\`" >> "$REPORT_FILE"
    ((FAILED++))
fi
cd ..

log_info "2. 前端构建检查..."
((TOTAL++))
cd im-admin
if npm run build > ../reports/logs/frontend-build.log 2>&1; then
    log_success "前端构建通过"
    echo "- ✅ 前端构建: 通过" >> "$REPORT_FILE"
    echo "  - 构建产物: \`im-admin/dist/\`" >> "$REPORT_FILE"
    ((PASSED++))
else
    log_error "前端构建失败"
    echo "- ❌ 前端构建: 失败" >> "$REPORT_FILE"
    ((FAILED++))
fi
cd ..

echo "" >> "$REPORT_FILE"

###############################################################################
# 2. 代码质量检查
###############################################################################
echo "## 2. 代码质量检查" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "3. Go代码格式检查..."
((TOTAL++))
cd im-backend
UNFORMATTED=$(gofmt -l . | grep -v vendor/ | grep -v telegram- || true)
if [ -z "$UNFORMATTED" ]; then
    log_success "Go代码格式正确"
    echo "- ✅ Go代码格式: 通过" >> "$REPORT_FILE"
    ((PASSED++))
else
    log_error "Go代码需要格式化"
    echo "- ❌ Go代码格式: 失败" >> "$REPORT_FILE"
    echo "  - 未格式化文件: $UNFORMATTED" >> "$REPORT_FILE"
    ((FAILED++))
fi

log_info "4. Go vet检查..."
((TOTAL++))
if go vet ./... > ../reports/logs/go-vet.log 2>&1; then
    log_success "Go vet通过"
    echo "- ✅ Go vet: 通过" >> "$REPORT_FILE"
    ((PASSED++))
else
    log_error "Go vet发现问题"
    echo "- ❌ Go vet: 失败" >> "$REPORT_FILE"
    ((FAILED++))
fi
cd ..

echo "" >> "$REPORT_FILE"

###############################################################################
# 3. 单元测试
###############################################################################
echo "## 3. 单元测试" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "5. 运行单元测试..."
((TOTAL++))
cd im-backend
if go test ./tests/unit/... -v -cover -coverprofile=../reports/tests/coverage.out > ../reports/tests/unit-test.log 2>&1; then
    COVERAGE=$(go tool cover -func=../reports/tests/coverage.out | grep total | awk '{print $3}')
    log_success "单元测试通过，覆盖率: $COVERAGE"
    echo "- ✅ 单元测试: 通过" >> "$REPORT_FILE"
    echo "  - 覆盖率: $COVERAGE" >> "$REPORT_FILE"
    echo "  - 报告: \`reports/tests/unit-test.log\`" >> "$REPORT_FILE"
    
    # 生成HTML覆盖率报告
    go tool cover -html=../reports/tests/coverage.out -o ../reports/tests/coverage.html
    echo "  - HTML报告: \`reports/tests/coverage.html\`" >> "$REPORT_FILE"
    ((PASSED++))
    
    # 检查覆盖率是否达标
    COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM >= 40" | bc -l 2>/dev/null || echo "0") )); then
        log_success "覆盖率达标 (≥40%)"
        echo "  - ✅ 覆盖率达标" >> "$REPORT_FILE"
    else
        log_warning "覆盖率不足40%: $COVERAGE"
        echo "  - ⚠️ 覆盖率不足40%" >> "$REPORT_FILE"
    fi
else
    log_error "单元测试失败"
    echo "- ❌ 单元测试: 失败" >> "$REPORT_FILE"
    ((FAILED++))
fi
cd ..

echo "" >> "$REPORT_FILE"

###############################################################################
# 4. 环境变量检查
###############################################################################
echo "## 4. 环境变量检查" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "6. 检查.env.example完整性..."
((TOTAL++))
if [ -f ".env.example" ]; then
    VAR_COUNT=$(grep -c "^[A-Z_]*=" .env.example || true)
    log_success ".env.example存在，包含 $VAR_COUNT 个变量"
    echo "- ✅ .env.example: 存在" >> "$REPORT_FILE"
    echo "  - 变量数量: $VAR_COUNT" >> "$REPORT_FILE"
    ((PASSED++))
    
    # 检查必需变量
    REQUIRED_VARS=("DB_HOST" "DB_PORT" "DB_USER" "DB_PASSWORD" "DB_NAME" "REDIS_HOST" "REDIS_PASSWORD" "JWT_SECRET")
    MISSING=""
    for var in "${REQUIRED_VARS[@]}"; do
        if ! grep -q "^${var}=" .env.example; then
            MISSING="$MISSING $var"
        fi
    done
    
    if [ -z "$MISSING" ]; then
        log_success "所有必需变量已包含"
        echo "  - ✅ 必需变量: 完整" >> "$REPORT_FILE"
    else
        log_warning "缺少变量:$MISSING"
        echo "  - ⚠️ 缺少变量:$MISSING" >> "$REPORT_FILE"
    fi
else
    log_error ".env.example不存在"
    echo "- ❌ .env.example: 不存在" >> "$REPORT_FILE"
    ((FAILED++))
fi

echo "" >> "$REPORT_FILE"

###############################################################################
# 5. Docker配置检查
###############################################################################
echo "## 5. Docker配置检查" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "7. 检查Docker Compose配置..."
((TOTAL++))
if docker-compose -f docker-compose.production.yml config > reports/logs/docker-config.log 2>&1; then
    log_success "Docker Compose配置正确"
    echo "- ✅ Docker Compose: 配置正确" >> "$REPORT_FILE"
    ((PASSED++))
else
    log_error "Docker Compose配置错误"
    echo "- ❌ Docker Compose: 配置错误" >> "$REPORT_FILE"
    ((FAILED++))
fi

echo "" >> "$REPORT_FILE"

###############################################################################
# 6. 安全检查
###############################################################################
echo "## 6. 安全检查" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "8. 检查硬编码密钥..."
((TOTAL++))
HARDCODED=$(grep -r "password.*=.*[\"'].*[\"']\|secret.*=.*[\"'].*[\"']" im-backend/internal/ | grep -v "test\|example\|TODO" || true)
if [ -z "$HARDCODED" ]; then
    log_success "未发现硬编码密钥"
    echo "- ✅ 硬编码检查: 通过" >> "$REPORT_FILE"
    ((PASSED++))
else
    log_error "发现硬编码密钥"
    echo "- ❌ 硬编码检查: 发现问题" >> "$REPORT_FILE"
    echo "\`\`\`" >> "$REPORT_FILE"
    echo "$HARDCODED" >> "$REPORT_FILE"
    echo "\`\`\`" >> "$REPORT_FILE"
    ((FAILED++))
fi

echo "" >> "$REPORT_FILE"

###############################################################################
# 7. 文件检查
###############################################################################
echo "## 7. 关键文件检查" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "9. 检查关键文件存在性..."
CRITICAL_FILES=(
    "im-backend/main.go"
    "im-backend/go.mod"
    "im-admin/package.json"
    "docker-compose.production.yml"
    ".env.example"
    "ops/bootstrap.sh"
    "ops/deploy.sh"
    "ops/rollback.sh"
    "config/prometheus/alert-rules.yml"
    "config/grafana/dashboards/im-suite-dashboard.json"
)

ALL_EXISTS=true
for file in "${CRITICAL_FILES[@]}"; do
    ((TOTAL++))
    if [ -f "$file" ]; then
        log_success "✓ $file"
        echo "- ✅ $file: 存在" >> "$REPORT_FILE"
        ((PASSED++))
    else
        log_error "✗ $file 不存在"
        echo "- ❌ $file: 不存在" >> "$REPORT_FILE"
        ((FAILED++))
        ALL_EXISTS=false
    fi
done

echo "" >> "$REPORT_FILE"

###############################################################################
# 8. 脚本可执行性检查
###############################################################################
echo "## 8. 脚本可执行性检查" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log_info "10. 检查ops脚本语法..."
OPS_SCRIPTS=$(ls ops/*.sh 2>/dev/null || true)
for script in $OPS_SCRIPTS; do
    ((TOTAL++))
    if bash -n "$script" 2> /dev/null; then
        log_success "✓ $script 语法正确"
        echo "- ✅ $(basename $script): 语法正确" >> "$REPORT_FILE"
        ((PASSED++))
    else
        log_error "✗ $script 语法错误"
        echo "- ❌ $(basename $script): 语法错误" >> "$REPORT_FILE"
        ((FAILED++))
    fi
done

echo "" >> "$REPORT_FILE"

###############################################################################
# 总结
###############################################################################
echo "## 📊 验证总结" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "| 指标 | 数值 |" >> "$REPORT_FILE"
echo "|------|------|" >> "$REPORT_FILE"
echo "| 总检查项 | $TOTAL |" >> "$REPORT_FILE"
echo "| 通过 | $PASSED |" >> "$REPORT_FILE"
echo "| 失败 | $FAILED |" >> "$REPORT_FILE"
echo "| 通过率 | $(echo "scale=2; $PASSED * 100 / $TOTAL" | bc)% |" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

PASS_RATE=$(echo "scale=0; $PASSED * 100 / $TOTAL" | bc)

if [ $FAILED -eq 0 ]; then
    echo "**状态**: ✅ **全部通过**" >> "$REPORT_FILE"
    log_success "========================================="
    log_success "✅ 所有验证通过！($PASSED/$TOTAL)"
    log_success "========================================="
    EXIT_CODE=0
else
    echo "**状态**: ❌ **有失败项**" >> "$REPORT_FILE"
    log_error "========================================="
    log_error "❌ 验证失败: $FAILED 项"
    log_error "========================================="
    EXIT_CODE=1
fi

echo "" >> "$REPORT_FILE"
echo "---" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "**报告文件**: $REPORT_FILE" >> "$REPORT_FILE"
echo "**日志目录**: \`reports/logs/\`" >> "$REPORT_FILE"
echo "**测试报告**: \`reports/tests/\`" >> "$REPORT_FILE"

log_info "验证报告已生成: $REPORT_FILE"
log_info "证据文件位于: reports/"

exit $EXIT_CODE

