#!/bin/bash

# 志航密信部署脚本
# 支持 Docker Swarm 和 Kubernetes 部署

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
    -e, --env ENV       指定环境 (staging|production)
    -m, --mode MODE     指定部署模式 (swarm|k8s)
    -v, --version VER   指定版本
    -f, --force         强制部署
    --dry-run          模拟部署
    --rollback          回滚到上一个版本

环境:
    staging             测试环境
    production          生产环境

部署模式:
    swarm              Docker Swarm 部署
    k8s                Kubernetes 部署

示例:
    $0 --env staging --mode swarm
    $0 --env production --mode k8s --version v1.0.0
    $0 --rollback
EOF
}

# 默认参数
ENVIRONMENT=""
DEPLOY_MODE=""
VERSION="latest"
FORCE=false
DRY_RUN=false
ROLLBACK=false

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
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --rollback)
            ROLLBACK=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证参数
if [[ -z "$ENVIRONMENT" ]]; then
    log_error "请指定环境 (--env staging|production)"
    exit 1
fi

if [[ -z "$DEPLOY_MODE" ]]; then
    log_error "请指定部署模式 (--mode swarm|k8s)"
    exit 1
fi

if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
    log_error "无效的环境: $ENVIRONMENT"
    exit 1
fi

if [[ "$DEPLOY_MODE" != "swarm" && "$DEPLOY_MODE" != "k8s" ]]; then
    log_error "无效的部署模式: $DEPLOY_MODE"
    exit 1
fi

# 设置环境变量
export ENVIRONMENT=$ENVIRONMENT
export VERSION=$VERSION
export DOCKER_REGISTRY="ghcr.io"
export IMAGE_NAME="zhihang-messenger"

# 检查依赖
check_dependencies() {
    log_info "检查部署依赖..."
    
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        if ! command -v docker &> /dev/null; then
            log_error "Docker 未安装"
            exit 1
        fi
        
        if ! docker info &> /dev/null; then
            log_error "Docker 服务未运行"
            exit 1
        fi
        
        if ! docker node ls &> /dev/null; then
            log_error "Docker Swarm 未初始化"
            exit 1
        fi
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        if ! command -v kubectl &> /dev/null; then
            log_error "kubectl 未安装"
            exit 1
        fi
        
        if ! kubectl cluster-info &> /dev/null; then
            log_error "Kubernetes 集群连接失败"
            exit 1
        fi
    fi
    
    log_success "依赖检查通过"
}

# 准备环境配置
prepare_config() {
    log_info "准备环境配置..."
    
    # 创建环境配置文件
    cat > .env << EOF
ENVIRONMENT=$ENVIRONMENT
VERSION=$VERSION
DOCKER_REGISTRY=$DOCKER_REGISTRY
IMAGE_NAME=$IMAGE_NAME
DB_PASSWORD=$(openssl rand -base64 32)
MYSQL_ROOT_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)
MINIO_ACCESS_KEY=$(openssl rand -base64 32)
MINIO_SECRET_KEY=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
EOF
    
    log_success "环境配置准备完成"
}

# Docker Swarm 部署
deploy_swarm() {
    log_info "开始 Docker Swarm 部署..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "模拟部署 (dry-run)"
        docker stack config docker-stack.yml
        return 0
    fi
    
    # 检查现有服务
    if docker stack ls | grep -q "zhihang-messenger"; then
        if [[ "$FORCE" != "true" ]]; then
            log_warning "服务已存在，使用 --force 强制更新"
            exit 1
        fi
        log_info "更新现有服务..."
        docker stack deploy -c docker-stack.yml zhihang-messenger
    else
        log_info "部署新服务..."
        docker stack deploy -c docker-stack.yml zhihang-messenger
    fi
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 30
    
    # 检查服务状态
    docker service ls --filter name=zhihang-messenger
    
    log_success "Docker Swarm 部署完成"
}

# Kubernetes 部署
deploy_k8s() {
    log_info "开始 Kubernetes 部署..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "模拟部署 (dry-run)"
        kubectl apply --dry-run=client -f k8s/
        return 0
    fi
    
    # 创建命名空间
    kubectl apply -f k8s/namespace.yaml
    
    # 创建密钥
    kubectl create secret generic mysql-secret \
        --from-literal=password=$(openssl rand -base64 32) \
        --namespace=zhihang-messenger \
        --dry-run=client -o yaml | kubectl apply -f -
    
    kubectl create secret generic redis-secret \
        --from-literal=password=$(openssl rand -base64 32) \
        --namespace=zhihang-messenger \
        --dry-run=client -o yaml | kubectl apply -f -
    
    kubectl create secret generic minio-secret \
        --from-literal=access-key=$(openssl rand -base64 32) \
        --from-literal=secret-key=$(openssl rand -base64 32) \
        --namespace=zhihang-messenger \
        --dry-run=client -o yaml | kubectl apply -f -
    
    kubectl create secret generic jwt-secret \
        --from-literal=secret=$(openssl rand -base64 64) \
        --namespace=zhihang-messenger \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # 部署服务
    kubectl apply -f k8s/
    
    # 等待部署完成
    log_info "等待部署完成..."
    kubectl wait --for=condition=available --timeout=300s deployment/zhihang-messenger-backend -n zhihang-messenger
    
    log_success "Kubernetes 部署完成"
}

# 回滚部署
rollback_deployment() {
    log_info "开始回滚部署..."
    
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        # 获取上一个版本
        PREVIOUS_VERSION=$(docker service inspect zhihang-messenger_backend --format '{{.Spec.Labels.version}}' 2>/dev/null || echo "unknown")
        
        if [[ "$PREVIOUS_VERSION" == "unknown" ]]; then
            log_error "无法获取上一个版本"
            exit 1
        fi
        
        log_info "回滚到版本: $PREVIOUS_VERSION"
        export VERSION=$PREVIOUS_VERSION
        docker stack deploy -c docker-stack.yml zhihang-messenger
        
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        # 回滚到上一个版本
        kubectl rollout undo deployment/zhihang-messenger-backend -n zhihang-messenger
        
        # 等待回滚完成
        kubectl rollout status deployment/zhihang-messenger-backend -n zhihang-messenger
    fi
    
    log_success "回滚完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待服务启动
    sleep 30
    
    # 检查服务状态
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        # 检查 Docker Swarm 服务状态
        docker service ls --filter name=zhihang-messenger
        
        # 检查服务健康状态
        for service in $(docker service ls --filter name=zhihang-messenger --format "{{.Name}}"); do
            if ! docker service ps $service --filter desired-state=running --format "{{.CurrentState}}" | grep -q "Running"; then
                log_error "服务 $service 未正常运行"
                exit 1
            fi
        done
        
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        # 检查 Kubernetes 部署状态
        kubectl get pods -n zhihang-messenger
        
        # 检查部署状态
        if ! kubectl get deployment zhihang-messenger-backend -n zhihang-messenger | grep -q "Available"; then
            log_error "后端服务未正常运行"
            exit 1
        fi
    fi
    
    log_success "健康检查通过"
}

# 清理资源
cleanup() {
    log_info "清理临时文件..."
    rm -f .env
    log_success "清理完成"
}

# 主函数
main() {
    log_info "开始志航密信部署..."
    log_info "环境: $ENVIRONMENT"
    log_info "部署模式: $DEPLOY_MODE"
    log_info "版本: $VERSION"
    
    if [[ "$ROLLBACK" == "true" ]]; then
        rollback_deployment
        health_check
        cleanup
        exit 0
    fi
    
    check_dependencies
    prepare_config
    
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        deploy_swarm
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        deploy_k8s
    fi
    
    health_check
    cleanup
    
    log_success "部署完成！"
}

# 设置错误处理
trap cleanup EXIT

# 执行主函数
main "$@"
