#!/bin/bash

# 志航密信部署脚本
# 支持 Docker Compose 和 Kubernetes 部署

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
志航密信部署脚本

用法: $0 [选项] [环境]

选项:
    -h, --help          显示帮助信息
    -e, --env ENV       指定环境 (dev|staging|prod)
    -m, --mode MODE     指定部署模式 (docker|k8s)
    -d, --dry-run       仅显示将要执行的命令，不实际执行
    -f, --force         强制部署，跳过确认
    -v, --verbose       显示详细输出

环境:
    dev                 开发环境
    staging             测试环境
    prod                生产环境

部署模式:
    docker             使用 Docker Compose 部署
    k8s                使用 Kubernetes 部署

示例:
    $0 -e dev -m docker              # 部署到开发环境
    $0 -e staging -m k8s             # 部署到测试环境
    $0 -e prod -m docker --force     # 强制部署到生产环境
    $0 --help                        # 显示帮助信息

EOF
}

# 默认参数
ENVIRONMENT="dev"
DEPLOY_MODE="docker"
DRY_RUN=false
FORCE=false
VERBOSE=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -m|--mode)
            DEPLOY_MODE="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证环境参数
case $ENVIRONMENT in
    dev|staging|prod)
        ;;
    *)
        log_error "无效的环境: $ENVIRONMENT"
        log_info "支持的环境: dev, staging, prod"
        exit 1
        ;;
esac

# 验证部署模式
case $DEPLOY_MODE in
    docker|k8s)
        ;;
    *)
        log_error "无效的部署模式: $DEPLOY_MODE"
        log_info "支持的部署模式: docker, k8s"
        exit 1
        ;;
esac

# 执行命令函数
execute_command() {
    local cmd="$1"
    local description="$2"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] $description"
        log_info "[DRY RUN] 命令: $cmd"
        return 0
    fi
    
    if [ "$VERBOSE" = true ]; then
        log_info "$description"
        log_info "执行命令: $cmd"
    fi
    
    if eval "$cmd"; then
        log_success "$description 完成"
        return 0
    else
        log_error "$description 失败"
        return 1
    fi
}

# 确认部署
confirm_deployment() {
    if [ "$FORCE" = true ]; then
        return 0
    fi
    
    echo
    log_warning "即将部署到 $ENVIRONMENT 环境，使用 $DEPLOY_MODE 模式"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "部署已取消"
        exit 0
    fi
}

# 检查依赖
check_dependencies() {
    log_info "检查部署依赖..."
    
    if [ "$DEPLOY_MODE" = "docker" ]; then
        if ! command -v docker &> /dev/null; then
            log_error "Docker 未安装"
            exit 1
        fi
        
        if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
            log_error "Docker Compose 未安装"
            exit 1
        fi
    fi
    
    if [ "$DEPLOY_MODE" = "k8s" ]; then
        if ! command -v kubectl &> /dev/null; then
            log_error "kubectl 未安装"
            exit 1
        fi
        
        if ! kubectl cluster-info &> /dev/null; then
            log_error "Kubernetes 集群连接失败"
            exit 1
        fi
    fi
    
    log_success "依赖检查完成"
}

# 准备环境配置
prepare_environment() {
    log_info "准备 $ENVIRONMENT 环境配置..."
    
    # 创建环境配置目录
    mkdir -p "configs/$ENVIRONMENT"
    
    # 复制环境配置文件
    if [ -f "configs/$ENVIRONMENT/.env" ]; then
        log_info "使用现有环境配置: configs/$ENVIRONMENT/.env"
    else
        log_warning "环境配置文件不存在: configs/$ENVIRONMENT/.env"
        log_info "请创建环境配置文件"
    fi
    
    # 生成 SSL 证书（如果需要）
    if [ ! -f "scripts/ssl/zhihang-messenger.crt" ]; then
        log_info "生成 SSL 证书..."
        execute_command "chmod +x scripts/ssl/generate-ssl.sh && scripts/ssl/generate-ssl.sh" "生成 SSL 证书"
    fi
    
    log_success "环境配置准备完成"
}

# Docker 部署
deploy_docker() {
    log_info "使用 Docker Compose 部署到 $ENVIRONMENT 环境..."
    
    # 构建镜像
    execute_command "docker-compose -f docker-compose.yml build" "构建 Docker 镜像"
    
    # 停止现有服务
    execute_command "docker-compose -f docker-compose.yml down" "停止现有服务"
    
    # 启动服务
    execute_command "docker-compose -f docker-compose.yml up -d" "启动服务"
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    execute_command "docker-compose -f docker-compose.yml ps" "检查服务状态"
    
    # 健康检查
    log_info "执行健康检查..."
    execute_command "curl -f http://localhost:8080/api/ping" "后端服务健康检查"
    
    log_success "Docker 部署完成"
}

# Kubernetes 部署
deploy_k8s() {
    log_info "使用 Kubernetes 部署到 $ENVIRONMENT 环境..."
    
    # 创建命名空间
    execute_command "kubectl apply -f k8s/namespace.yaml" "创建命名空间"
    
    # 应用配置
    execute_command "kubectl apply -f k8s/" "应用 Kubernetes 配置"
    
    # 等待部署完成
    log_info "等待部署完成..."
    execute_command "kubectl rollout status deployment/zhihang-messenger-backend -n zhihang-messenger" "等待后端部署完成"
    
    # 检查服务状态
    execute_command "kubectl get pods -n zhihang-messenger" "检查 Pod 状态"
    execute_command "kubectl get services -n zhihang-messenger" "检查服务状态"
    
    log_success "Kubernetes 部署完成"
}

# 部署后验证
post_deployment_check() {
    log_info "执行部署后验证..."
    
    if [ "$DEPLOY_MODE" = "docker" ]; then
        # 检查容器状态
        execute_command "docker-compose -f docker-compose.yml ps" "检查容器状态"
        
        # 检查日志
        execute_command "docker-compose -f docker-compose.yml logs --tail=50 backend" "检查后端日志"
        
        # API 健康检查
        execute_command "curl -f http://localhost:8080/api/ping" "API 健康检查"
        
        # Web 服务检查
        execute_command "curl -f http://localhost:3000" "Web 服务检查"
    fi
    
    if [ "$DEPLOY_MODE" = "k8s" ]; then
        # 检查 Pod 状态
        execute_command "kubectl get pods -n zhihang-messenger" "检查 Pod 状态"
        
        # 检查服务状态
        execute_command "kubectl get services -n zhihang-messenger" "检查服务状态"
        
        # 端口转发测试
        execute_command "kubectl port-forward -n zhihang-messenger svc/zhihang-messenger-backend 8080:8080 &" "设置端口转发"
        sleep 5
        execute_command "curl -f http://localhost:8080/api/ping" "API 健康检查"
        kill %1 2>/dev/null || true
    fi
    
    log_success "部署后验证完成"
}

# 主函数
main() {
    log_info "开始部署志航密信到 $ENVIRONMENT 环境..."
    log_info "部署模式: $DEPLOY_MODE"
    
    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN 模式 - 不会实际执行部署"
    fi
    
    # 确认部署
    confirm_deployment
    
    # 检查依赖
    check_dependencies
    
    # 准备环境
    prepare_environment
    
    # 执行部署
    case $DEPLOY_MODE in
        docker)
            deploy_docker
            ;;
        k8s)
            deploy_k8s
            ;;
    esac
    
    # 部署后验证
    post_deployment_check
    
    log_success "🎉 志航密信部署完成！"
    log_info "访问地址:"
    log_info "  - Web 端: http://localhost:3000"
    log_info "  - 管理后台: http://localhost:3001"
    log_info "  - API 文档: http://localhost:8080/api/ping"
    
    if [ "$ENVIRONMENT" = "prod" ]; then
        log_warning "⚠️  生产环境部署完成，请确保："
        log_warning "  1. 配置了正确的域名和 SSL 证书"
        log_warning "  2. 设置了防火墙规则"
        log_warning "  3. 配置了监控和日志收集"
        log_warning "  4. 设置了备份策略"
    fi
}

# 脚本入口
main "$@"