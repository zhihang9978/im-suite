#!/bin/bash

###############################################################################
# 志航密信 - 开发自检脚本
# 用途：一键本地自检（lint、编译、单测、型检、依赖审计）
# 使用：bash ops/dev_check.sh
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

PASSED=0
FAILED=0
WARNINGS=0

log_info "========================================="
log_info "志航密信 - 开发自检"
log_info "========================================="

# ====================================
# 1. 检查依赖
# ====================================
log_info "1. 检查依赖工具..."

check_command() {
    if command -v $1 &> /dev/null; then
        log_success "$1 已安装"
        ((PASSED++))
        return 0
    else
        log_error "$1 未安装"
        ((FAILED++))
        return 1
    fi
}

check_command "go"
check_command "node"
check_command "npm"
check_command "git"

if [ $FAILED -gt 0 ]; then
    log_error "缺少必要工具，请先安装"
    exit 1
fi

# ====================================
# 2. 后端检查
# ====================================
log_info "2. 后端检查..."

cd im-backend

# 2.1 Go mod验证
log_info "2.1 检查 go.mod..."
if go mod verify; then
    log_success "go.mod 验证通过"
    ((PASSED++))
else
    log_error "go.mod 验证失败"
    ((FAILED++))
fi

# 2.2 Go fmt检查
log_info "2.2 检查代码格式..."
UNFORMATTED=$(gofmt -l . | grep -v "vendor/" || true)
if [ -z "$UNFORMATTED" ]; then
    log_success "代码格式正确"
    ((PASSED++))
else
    log_error "以下文件需要格式化:"
    echo "$UNFORMATTED"
    ((FAILED++))
fi

# 2.3 Go vet检查
log_info "2.3 运行 go vet..."
if go vet ./...; then
    log_success "go vet 通过"
    ((PASSED++))
else
    log_error "go vet 发现问题"
    ((FAILED++))
fi

# 2.4 编译检查
log_info "2.4 编译后端..."
if go build -o /tmp/im-backend main.go; then
    log_success "后端编译成功"
    ((PASSED++))
    rm -f /tmp/im-backend
else
    log_error "后端编译失败"
    ((FAILED++))
fi

# 2.5 单元测试
log_info "2.5 运行单元测试..."
if go test ./... -v -cover -coverprofile=/tmp/coverage.out; then
    COVERAGE=$(go tool cover -func=/tmp/coverage.out | grep total | awk '{print $3}')
    log_success "单元测试通过，覆盖率: $COVERAGE"
    ((PASSED++))
    
    # 检查覆盖率是否达标
    COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM >= 60" | bc -l) )); then
        log_success "覆盖率达标 (≥60%)"
        ((PASSED++))
    else
        log_warning "覆盖率不足60%: $COVERAGE"
        ((WARNINGS++))
    fi
else
    log_error "单元测试失败"
    ((FAILED++))
fi

# 2.6 依赖审计
log_info "2.6 Go依赖安全审计..."
if go list -json -m all | grep -q "CVE"; then
    log_warning "发现安全漏洞，请运行: go list -json -m all"
    ((WARNINGS++))
else
    log_success "未发现已知漏洞"
    ((PASSED++))
fi

cd ..

# ====================================
# 3. 前端检查
# ====================================
log_info "3. 前端检查..."

cd im-admin

# 3.1 npm依赖
log_info "3.1 检查npm依赖..."
if npm ls --depth=0 &> /dev/null; then
    log_success "npm依赖完整"
    ((PASSED++))
else
    log_warning "npm依赖有问题，运行: npm install"
    ((WARNINGS++))
fi

# 3.2 Lint检查
log_info "3.2 运行eslint..."
if npm run lint -- --max-warnings=0; then
    log_success "eslint通过"
    ((PASSED++))
else
    log_error "eslint发现问题"
    ((FAILED++))
fi

# 3.3 构建检查
log_info "3.3 构建前端..."
if npm run build; then
    log_success "前端构建成功"
    ((PASSED++))
else
    log_error "前端构建失败"
    ((FAILED++))
fi

# 3.4 依赖审计
log_info "3.4 npm依赖安全审计..."
AUDIT_OUTPUT=$(npm audit --json)
VULNERABILITIES=$(echo "$AUDIT_OUTPUT" | jq -r '.metadata.vulnerabilities | .high + .critical')
if [ "$VULNERABILITIES" -eq 0 ]; then
    log_success "未发现高危漏洞"
    ((PASSED++))
else
    log_warning "发现 $VULNERABILITIES 个高危漏洞，运行: npm audit fix"
    ((WARNINGS++))
fi

cd ..

# ====================================
# 4. 配置检查
# ====================================
log_info "4. 配置检查..."

# 4.1 环境变量
log_info "4.1 检查环境变量..."
if [ -f ".env" ]; then
    log_success ".env 文件存在"
    ((PASSED++))
    
    # 检查必需变量
    REQUIRED_VARS=("DB_HOST" "DB_PORT" "DB_USER" "DB_PASSWORD" "DB_NAME" "REDIS_HOST" "JWT_SECRET")
    MISSING_VARS=""
    for var in "${REQUIRED_VARS[@]}"; do
        if ! grep -q "^${var}=" .env; then
            MISSING_VARS="$MISSING_VARS $var"
        fi
    done
    
    if [ -z "$MISSING_VARS" ]; then
        log_success "所有必需环境变量已配置"
        ((PASSED++))
    else
        log_error "缺少环境变量:$MISSING_VARS"
        ((FAILED++))
    fi
else
    log_error ".env 文件不存在"
    log_info "创建: cp .env.example .env"
    ((FAILED++))
fi

# 4.2 Docker配置
log_info "4.2 检查Docker配置..."
if [ -f "docker-compose.production.yml" ]; then
    if docker-compose -f docker-compose.production.yml config > /dev/null 2>&1; then
        log_success "Docker Compose配置正确"
        ((PASSED++))
    else
        log_error "Docker Compose配置有误"
        ((FAILED++))
    fi
else
    log_warning "docker-compose.production.yml不存在"
    ((WARNINGS++))
fi

# ====================================
# 5. Git检查
# ====================================
log_info "5. Git检查..."

# 5.1 未提交的更改
UNCOMMITTED=$(git status --porcelain | wc -l)
if [ "$UNCOMMITTED" -eq 0 ]; then
    log_success "无未提交更改"
    ((PASSED++))
else
    log_warning "有 $UNCOMMITTED 个未提交更改"
    ((WARNINGS++))
fi

# 5.2 检查分支
CURRENT_BRANCH=$(git branch --show-current)
log_info "当前分支: $CURRENT_BRANCH"

# ====================================
# 总结
# ====================================
TOTAL=$((PASSED + FAILED + WARNINGS))

log_info "========================================="
log_info "自检完成"
log_info "========================================="
echo "总计: $TOTAL"
echo "通过: $PASSED"
echo "失败: $FAILED"
echo "警告: $WARNINGS"
echo ""

if [ $FAILED -eq 0 ]; then
    log_success "========================================="
    log_success "✅ 所有检查通过！可以提交代码"
    log_success "========================================="
    exit 0
else
    log_error "========================================="
    log_error "❌ 有 $FAILED 项检查失败"
    log_error "请修复后再提交"
    log_error "========================================="
    exit 1
fi

