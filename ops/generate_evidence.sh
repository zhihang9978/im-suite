#!/bin/bash

###############################################################################
# 志航密信 - 证据生成脚本
# 用途：生成所有验证证据（日志、截图、报告）
# 使用：bash ops/generate_evidence.sh
###############################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }

# 创建reports目录结构
mkdir -p reports/{logs,screenshots,tests,builds,evidence,security}

TIMESTAMP=$(date +%Y%m%d-%H%M%S)

log_info "========================================="
log_info "生成验证证据"
log_info "时间: $(date '+%Y-%m-%d %H:%M:%S')"
log_info "========================================="

# =====================================
# 1. 编译证据
# =====================================
log_info "1. 生成编译证据..."

cd im-backend
go build -v -o /tmp/im-backend main.go > ../reports/builds/backend-build-${TIMESTAMP}.log 2>&1
if [ $? -eq 0 ]; then
    log_success "后端编译成功"
    ls -lh /tmp/im-backend > ../reports/builds/backend-binary-info.txt
    rm -f /tmp/im-backend
fi
cd ..

cd im-admin
npm run build > ../reports/builds/frontend-build-${TIMESTAMP}.log 2>&1
if [ $? -eq 0 ]; then
    log_success "前端构建成功"
    du -sh dist/ > ../reports/builds/frontend-dist-size.txt
fi
cd ..

# =====================================
# 2. 测试证据
# =====================================
log_info "2. 生成测试证据..."

cd im-backend

# 单元测试
go test ./tests/unit/... -v -cover -coverprofile=../reports/tests/coverage-${TIMESTAMP}.out \
    -json > ../reports/tests/unit-test-${TIMESTAMP}.json 2>&1

# 覆盖率报告
go tool cover -func=../reports/tests/coverage-${TIMESTAMP}.out > ../reports/tests/coverage-summary-${TIMESTAMP}.txt
go tool cover -html=../reports/tests/coverage-${TIMESTAMP}.out -o ../reports/tests/coverage-${TIMESTAMP}.html

COVERAGE=$(go tool cover -func=../reports/tests/coverage-${TIMESTAMP}.out | grep total | awk '{print $3}')
log_success "测试覆盖率: $COVERAGE"

# Benchmark测试
go test ./tests/unit/... -bench=. -benchmem > ../reports/tests/benchmark-${TIMESTAMP}.txt 2>&1

cd ..

# =====================================
# 3. 安全证据
# =====================================
log_info "3. 生成安全证据..."

# Go依赖列表
cd im-backend
go list -json -m all > ../reports/security/go-dependencies-${TIMESTAMP}.json
cd ..

# npm审计
cd im-admin
npm audit --json > ../reports/security/npm-audit-${TIMESTAMP}.json 2>&1 || true
npm audit > ../reports/security/npm-audit-${TIMESTAMP}.txt 2>&1 || true
cd ..

# =====================================
# 4. 配置证据
# =====================================
log_info "4. 生成配置证据..."

# Docker配置
docker-compose -f docker-compose.production.yml config > reports/evidence/docker-compose-parsed-${TIMESTAMP}.yml 2>&1 || echo "需要Docker环境"

# 环境变量检查
if [ -f ".env.example" ]; then
    grep -c "^[A-Z_]*=" .env.example > reports/evidence/env-var-count.txt
    log_success ".env.example包含 $(cat reports/evidence/env-var-count.txt) 个变量"
fi

# =====================================
# 5. 文件清单
# =====================================
log_info "5. 生成文件清单..."

find im-backend -name "*.go" | wc -l > reports/evidence/go-file-count.txt
find im-admin/src -name "*.vue" -o -name "*.js" | wc -l > reports/evidence/frontend-file-count.txt
find docs -name "*.md" | wc -l > reports/evidence/doc-file-count.txt
find ops -name "*.sh" | wc -l > reports/evidence/script-file-count.txt

log_success "Go文件: $(cat reports/evidence/go-file-count.txt)"
log_success "前端文件: $(cat reports/evidence/frontend-file-count.txt)"
log_success "文档文件: $(cat reports/evidence/doc-file-count.txt)"
log_success "脚本文件: $(cat reports/evidence/script-file-count.txt)"

# =====================================
# 6. Git证据
# =====================================
log_info "6. 生成Git证据..."

git log --oneline -30 > reports/evidence/git-commits-${TIMESTAMP}.txt
git status > reports/evidence/git-status-${TIMESTAMP}.txt
git diff --stat HEAD~10 > reports/evidence/git-diff-stats-${TIMESTAMP}.txt

# =====================================
# 7. 生成索引文件
# =====================================
log_info "7. 生成证据索引..."

cat > reports/INDEX.md <<EOF
# 验证证据索引

**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')

## 📁 目录结构

\`\`\`
reports/
├── logs/ - 构建和测试日志
├── tests/ - 测试报告和覆盖率
├── builds/ - 编译产物信息
├── security/ - 安全审计报告
├── evidence/ - 其他证据文件
└── INDEX.md - 本文件
\`\`\`

## 📊 证据文件清单

### 编译证据
- ✅ backend-build-${TIMESTAMP}.log - 后端编译日志
- ✅ frontend-build-${TIMESTAMP}.log - 前端构建日志
- ✅ backend-binary-info.txt - 二进制文件信息
- ✅ frontend-dist-size.txt - 前端构建产物大小

### 测试证据
- ✅ coverage-${TIMESTAMP}.out - 覆盖率原始数据
- ✅ coverage-${TIMESTAMP}.html - 覆盖率HTML报告
- ✅ coverage-summary-${TIMESTAMP}.txt - 覆盖率摘要
- ✅ unit-test-${TIMESTAMP}.json - 单元测试JSON结果
- ✅ benchmark-${TIMESTAMP}.txt - 性能测试结果

### 安全证据
- ✅ go-dependencies-${TIMESTAMP}.json - Go依赖清单
- ✅ npm-audit-${TIMESTAMP}.json - npm审计JSON
- ✅ npm-audit-${TIMESTAMP}.txt - npm审计报告

### 配置证据
- ✅ docker-compose-parsed-${TIMESTAMP}.yml - Docker配置
- ✅ env-var-count.txt - 环境变量数量

### 文件统计
- ✅ go-file-count.txt - Go文件数量
- ✅ frontend-file-count.txt - 前端文件数量
- ✅ doc-file-count.txt - 文档文件数量
- ✅ script-file-count.txt - 脚本文件数量

### Git证据
- ✅ git-commits-${TIMESTAMP}.txt - 最近30次提交
- ✅ git-status-${TIMESTAMP}.txt - Git状态
- ✅ git-diff-stats-${TIMESTAMP}.txt - 代码变更统计

## 📊 验证总结

- **测试覆盖率**: $COVERAGE
- **Go文件数**: $(cat reports/evidence/go-file-count.txt)
- **文档数**: $(cat reports/evidence/doc-file-count.txt)
- **脚本数**: $(cat reports/evidence/script-file-count.txt)

---

**生成工具**: ops/generate_evidence.sh
EOF

log_success "========================================="
log_success "证据生成完成！"
log_success "========================================="
echo ""
echo "证据目录: reports/"
echo "索引文件: reports/INDEX.md"
echo ""
echo "查看测试覆盖率:"
echo "  open reports/tests/coverage-${TIMESTAMP}.html"
echo ""

