#!/bin/bash

# 志航密信测试运行脚本
# 用于运行各种类型的测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
志航密信测试运行脚本

用法: $0 [选项] [测试类型]

选项:
    -h, --help          显示帮助信息
    -v, --verbose       显示详细输出
    -c, --coverage      生成测试覆盖率报告
    -r, --race          启用竞态条件检测
    -p, --parallel      并行运行测试
    -t, --timeout TIME  设置测试超时时间（秒）

测试类型:
    unit                 单元测试
    integration          集成测试
    e2e                  端到端测试
    performance          性能测试
    security             安全测试
    all                  所有测试

示例:
    $0 unit                    # 运行单元测试
    $0 integration --coverage  # 运行集成测试并生成覆盖率报告
    $0 all --verbose --race    # 运行所有测试，详细输出，启用竞态检测
    $0 --help                  # 显示帮助信息

EOF
}

# 默认参数
TEST_TYPE="unit"
VERBOSE=false
COVERAGE=false
RACE=false
PARALLEL=false
TIMEOUT=300

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -r|--race)
            RACE=true
            shift
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        unit|integration|e2e|performance|security|all)
            TEST_TYPE="$1"
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 构建测试命令
build_test_command() {
    local test_type="$1"
    local cmd="go test"
    
    # 添加超时
    cmd="$cmd -timeout ${TIMEOUT}s"
    
    # 添加详细输出
    if [ "$VERBOSE" = true ]; then
        cmd="$cmd -v"
    fi
    
    # 添加竞态检测
    if [ "$RACE" = true ]; then
        cmd="$cmd -race"
    fi
    
    # 添加并行测试
    if [ "$PARALLEL" = true ]; then
        cmd="$cmd -parallel 4"
    fi
    
    # 添加覆盖率
    if [ "$COVERAGE" = true ]; then
        cmd="$cmd -coverprofile=coverage.out -covermode=atomic"
    fi
    
    # 根据测试类型添加路径
    case $test_type in
        unit)
            cmd="$cmd ./internal/service/..."
            ;;
        integration)
            cmd="$cmd ./internal/..."
            ;;
        e2e)
            cmd="$cmd ./tests/e2e/..."
            ;;
        performance)
            cmd="$cmd -bench=. ./tests/performance/..."
            ;;
        security)
            cmd="$cmd ./tests/security/..."
            ;;
        all)
            cmd="$cmd ./..."
            ;;
    esac
    
    echo "$cmd"
}

# 运行测试
run_test() {
    local test_type="$1"
    local test_command="$2"
    
    log_info "开始运行 $test_type 测试..."
    log_info "命令: $test_command"
    
    # 进入后端目录
    cd im-backend
    
    # 运行测试
    if eval "$test_command"; then
        log_success "$test_type 测试通过"
        return 0
    else
        log_error "$test_type 测试失败"
        return 1
    fi
}

# 生成覆盖率报告
generate_coverage_report() {
    if [ "$COVERAGE" = true ]; then
        log_info "生成测试覆盖率报告..."
        
        cd im-backend
        
        # 生成 HTML 报告
        go tool cover -html=coverage.out -o coverage.html
        
        # 生成文本报告
        go tool cover -func=coverage.out > coverage.txt
        
        # 显示覆盖率统计
        log_info "测试覆盖率统计:"
        tail -1 coverage.txt
        
        # 移动报告文件到项目根目录
        mv coverage.out coverage.html coverage.txt ../
        
        log_success "覆盖率报告已生成:"
        log_info "  - HTML 报告: coverage.html"
        log_info "  - 文本报告: coverage.txt"
        log_info "  - 原始数据: coverage.out"
    fi
}

# 运行前端测试
run_frontend_tests() {
    log_info "运行前端测试..."
    
    # Web 端测试
    if [ -d "telegram-web" ]; then
        log_info "运行 Web 端测试..."
        cd telegram-web
        
        if command -v npm &> /dev/null; then
            if [ -f "package.json" ]; then
                npm test
                log_success "Web 端测试完成"
            else
                log_warning "Web 端没有测试配置"
            fi
        else
            log_warning "npm 未安装，跳过 Web 端测试"
        fi
        
        cd ..
    fi
    
    # 管理后台测试
    if [ -d "im-admin" ]; then
        log_info "运行管理后台测试..."
        cd im-admin
        
        if command -v npm &> /dev/null; then
            if [ -f "package.json" ]; then
                npm test
                log_success "管理后台测试完成"
            else
                log_warning "管理后台没有测试配置"
            fi
        else
            log_warning "npm 未安装，跳过管理后台测试"
        fi
        
        cd ..
    fi
}

# 运行 Android 测试
run_android_tests() {
    log_info "运行 Android 测试..."
    
    if [ -d "telegram-android" ]; then
        cd telegram-android
        
        if command -v ./gradlew &> /dev/null; then
            ./gradlew test
            log_success "Android 测试完成"
        else
            log_warning "Android 测试环境未配置"
        fi
        
        cd ..
    else
        log_warning "Android 项目目录不存在"
    fi
}

# 运行集成测试
run_integration_tests() {
    log_info "运行集成测试..."
    
    # 启动测试环境
    log_info "启动测试环境..."
    docker-compose -f docker-compose.dev.yml up -d mysql redis
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10
    
    # 运行集成测试
    cd im-backend
    go test -tags=integration ./tests/integration/...
    cd ..
    
    # 停止测试环境
    log_info "停止测试环境..."
    docker-compose -f docker-compose.dev.yml down
    
    log_success "集成测试完成"
}

# 运行性能测试
run_performance_tests() {
    log_info "运行性能测试..."
    
    cd im-backend
    
    # 运行基准测试
    go test -bench=. -benchmem ./tests/performance/...
    
    cd ..
    
    log_success "性能测试完成"
}

# 运行安全测试
run_security_tests() {
    log_info "运行安全测试..."
    
    cd im-backend
    
    # 运行安全测试
    go test ./tests/security/...
    
    cd ..
    
    log_success "安全测试完成"
}

# 主函数
main() {
    log_info "开始运行志航密信测试..."
    log_info "测试类型: $TEST_TYPE"
    
    # 检查 Go 环境
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查项目结构
    if [ ! -d "im-backend" ]; then
        log_error "后端项目目录不存在"
        exit 1
    fi
    
    # 运行测试
    case $TEST_TYPE in
        unit)
            test_command=$(build_test_command "unit")
            run_test "单元" "$test_command"
            ;;
        integration)
            run_integration_tests
            ;;
        e2e)
            test_command=$(build_test_command "e2e")
            run_test "端到端" "$test_command"
            ;;
        performance)
            run_performance_tests
            ;;
        security)
            run_security_tests
            ;;
        all)
            # 运行所有测试
            test_command=$(build_test_command "all")
            run_test "所有" "$test_command"
            run_frontend_tests
            run_android_tests
            ;;
    esac
    
    # 生成覆盖率报告
    generate_coverage_report
    
    log_success "🎉 所有测试完成！"
    
    # 显示测试结果摘要
    log_info "测试结果摘要:"
    if [ -f "coverage.txt" ]; then
        log_info "覆盖率报告:"
        tail -1 coverage.txt
    fi
}

# 脚本入口
main "$@"