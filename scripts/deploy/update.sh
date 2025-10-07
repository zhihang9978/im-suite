#!/bin/bash

# 志航密信更新脚本
# 支持滚动更新和蓝绿部署

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
志航密信更新脚本

用法: $0 [选项]

选项:
    -h, --help          显示帮助信息
    -v, --version VER   指定新版本
    -s, --strategy STR  更新策略 (rolling|blue-green)
    -e, --env ENV       指定环境 (staging|production)
    -m, --mode MODE     指定部署模式 (swarm|k8s)
    --dry-run          模拟更新
    --rollback          回滚更新

更新策略:
    rolling             滚动更新 (默认)
    blue-green          蓝绿部署

示例:
    $0 --version v1.1.0 --strategy rolling
    $0 --version v1.1.0 --strategy blue-green --env production
    $0 --rollback
EOF
}

# 默认参数
VERSION=""
STRATEGY="rolling"
ENVIRONMENT=""
DEPLOY_MODE=""
DRY_RUN=false
ROLLBACK=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -s|--strategy)
            STRATEGY="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -m|--mode)
            DEPLOY_MODE="$2"
            shift 2
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
if [[ -z "$VERSION" && "$ROLLBACK" != "true" ]]; then
    log_error "请指定版本 (--version)"
    exit 1
fi

if [[ "$STRATEGY" != "rolling" && "$STRATEGY" != "blue-green" ]]; then
    log_error "无效的更新策略: $STRATEGY"
    exit 1
fi

# 设置环境变量
export VERSION=$VERSION
export STRATEGY=$STRATEGY
export ENVIRONMENT=$ENVIRONMENT
export DEPLOY_MODE=$DEPLOY_MODE

# 检查当前版本
get_current_version() {
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        docker service inspect zhihang-messenger_backend --format '{{.Spec.Labels.version}}' 2>/dev/null || echo "unknown"
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        kubectl get deployment zhihang-messenger-backend -n zhihang-messenger -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null | cut -d: -f2 || echo "unknown"
    fi
}

# 滚动更新
rolling_update() {
    log_info "开始滚动更新到版本: $VERSION"
    
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        # Docker Swarm 滚动更新
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "模拟滚动更新 (dry-run)"
            docker service update --image ${DOCKER_REGISTRY}/zhihang-messenger/backend:$VERSION --dry-run zhihang-messenger_backend
        else
            docker service update --image ${DOCKER_REGISTRY}/zhihang-messenger/backend:$VERSION zhihang-messenger_backend
            docker service update --image ${DOCKER_REGISTRY}/zhihang-messenger/web:$VERSION zhihang-messenger_web
            docker service update --image ${DOCKER_REGISTRY}/zhihang-messenger/admin:$VERSION zhihang-messenger_admin
        fi
        
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        # Kubernetes 滚动更新
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "模拟滚动更新 (dry-run)"
            kubectl set image deployment/zhihang-messenger-backend backend=${DOCKER_REGISTRY}/zhihang-messenger/backend:$VERSION --dry-run=client -n zhihang-messenger
        else
            kubectl set image deployment/zhihang-messenger-backend backend=${DOCKER_REGISTRY}/zhihang-messenger/backend:$VERSION -n zhihang-messenger
            kubectl set image deployment/zhihang-messenger-web web=${DOCKER_REGISTRY}/zhihang-messenger/web:$VERSION -n zhihang-messenger
            kubectl set image deployment/zhihang-messenger-admin admin=${DOCKER_REGISTRY}/zhihang-messenger/admin:$VERSION -n zhihang-messenger
        fi
    fi
    
    log_success "滚动更新完成"
}

# 蓝绿部署
blue_green_deployment() {
    log_info "开始蓝绿部署到版本: $VERSION"
    
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        # Docker Swarm 蓝绿部署
        log_info "部署绿色环境..."
        
        # 创建绿色环境服务
        docker service create \
            --name zhihang-messenger-backend-green \
            --replicas 3 \
            --image ${DOCKER_REGISTRY}/zhihang-messenger/backend:$VERSION \
            --env-file .env \
            --network zhihang_net \
            zhihang-messenger-backend-green
        
        # 等待绿色环境启动
        log_info "等待绿色环境启动..."
        sleep 60
        
        # 健康检查
        if ! health_check_green; then
            log_error "绿色环境健康检查失败"
            docker service rm zhihang-messenger-backend-green
            exit 1
        fi
        
        # 切换流量到绿色环境
        log_info "切换流量到绿色环境..."
        # 这里需要更新负载均衡器配置
        
        # 删除蓝色环境
        log_info "删除蓝色环境..."
        docker service rm zhihang-messenger_backend
        
        # 重命名绿色环境为蓝色环境
        docker service update --name zhihang-messenger_backend zhihang-messenger-backend-green
        
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        # Kubernetes 蓝绿部署
        log_info "部署绿色环境..."
        
        # 创建绿色环境部署
        kubectl apply -f - << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zhihang-messenger-backend-green
  namespace: zhihang-messenger
  labels:
    app: zhihang-messenger-backend
    version: green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: zhihang-messenger-backend
      version: green
  template:
    metadata:
      labels:
        app: zhihang-messenger-backend
        version: green
    spec:
      containers:
      - name: backend
        image: ${DOCKER_REGISTRY}/zhihang-messenger/backend:$VERSION
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "mysql-service"
        - name: DB_PORT
          value: "3306"
        - name: DB_NAME
          value: "zhihang_messenger"
        - name: DB_USER
          value: "zhihang_messenger"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mysql-secret
              key: password
        - name: REDIS_HOST
          value: "redis-service"
        - name: REDIS_PORT
          value: "6379"
        - name: MINIO_ENDPOINT
          value: "minio-service:9000"
        - name: MINIO_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: minio-secret
              key: access-key
        - name: MINIO_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: minio-secret
              key: secret-key
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
EOF
        
        # 等待绿色环境启动
        log_info "等待绿色环境启动..."
        kubectl wait --for=condition=available --timeout=300s deployment/zhihang-messenger-backend-green -n zhihang-messenger
        
        # 健康检查
        if ! health_check_green; then
            log_error "绿色环境健康检查失败"
            kubectl delete deployment zhihang-messenger-backend-green -n zhihang-messenger
            exit 1
        fi
        
        # 切换流量到绿色环境
        log_info "切换流量到绿色环境..."
        kubectl patch service backend-service -n zhihang-messenger -p '{"spec":{"selector":{"version":"green"}}}'
        
        # 删除蓝色环境
        log_info "删除蓝色环境..."
        kubectl delete deployment zhihang-messenger-backend -n zhihang-messenger
        
        # 重命名绿色环境为蓝色环境
        kubectl patch deployment zhihang-messenger-backend-green -n zhihang-messenger -p '{"metadata":{"name":"zhihang-messenger-backend","labels":{"version":"blue"}},"spec":{"selector":{"matchLabels":{"version":"blue"}},"template":{"metadata":{"labels":{"version":"blue"}}}}}'
    fi
    
    log_success "蓝绿部署完成"
}

# 健康检查
health_check_green() {
    log_info "执行绿色环境健康检查..."
    
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        # 检查 Docker Swarm 服务状态
        if ! docker service ps zhihang-messenger-backend-green --filter desired-state=running --format "{{.CurrentState}}" | grep -q "Running"; then
            return 1
        fi
        
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        # 检查 Kubernetes 部署状态
        if ! kubectl get deployment zhihang-messenger-backend-green -n zhihang-messenger | grep -q "Available"; then
            return 1
        fi
    fi
    
    log_success "绿色环境健康检查通过"
    return 0
}

# 回滚更新
rollback_update() {
    log_info "开始回滚更新..."
    
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        # Docker Swarm 回滚
        docker service rollback zhihang-messenger_backend
        docker service rollback zhihang-messenger_web
        docker service rollback zhihang-messenger_admin
        
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        # Kubernetes 回滚
        kubectl rollout undo deployment/zhihang-messenger-backend -n zhihang-messenger
        kubectl rollout undo deployment/zhihang-messenger-web -n zhihang-messenger
        kubectl rollout undo deployment/zhihang-messenger-admin -n zhihang-messenger
    fi
    
    log_success "回滚完成"
}

# 验证更新
verify_update() {
    log_info "验证更新..."
    
    # 等待服务稳定
    sleep 30
    
    # 检查服务状态
    if [[ "$DEPLOY_MODE" == "swarm" ]]; then
        docker service ls --filter name=zhihang-messenger
        
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        kubectl get pods -n zhihang-messenger
    fi
    
    # 检查版本
    CURRENT_VERSION=$(get_current_version)
    log_info "当前版本: $CURRENT_VERSION"
    
    if [[ "$CURRENT_VERSION" == "$VERSION" ]]; then
        log_success "版本更新成功"
    else
        log_warning "版本可能未完全更新"
    fi
}

# 主函数
main() {
    log_info "开始志航密信更新..."
    log_info "目标版本: $VERSION"
    log_info "更新策略: $STRATEGY"
    log_info "环境: $ENVIRONMENT"
    log_info "部署模式: $DEPLOY_MODE"
    
    if [[ "$ROLLBACK" == "true" ]]; then
        rollback_update
        verify_update
        exit 0
    fi
    
    # 获取当前版本
    CURRENT_VERSION=$(get_current_version)
    log_info "当前版本: $CURRENT_VERSION"
    
    if [[ "$CURRENT_VERSION" == "$VERSION" ]]; then
        log_warning "版本已是最新，无需更新"
        exit 0
    fi
    
    # 执行更新
    if [[ "$STRATEGY" == "rolling" ]]; then
        rolling_update
    elif [[ "$STRATEGY" == "blue-green" ]]; then
        blue_green_deployment
    fi
    
    verify_update
    
    log_success "更新完成！"
}

# 执行主函数
main "$@"
