#!/bin/bash

# 志航密信更新脚本
# 用于更新已部署的服务

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
    -e, --env ENV       指定环境 (dev|staging|prod)
    -m, --mode MODE     指定部署模式 (docker|k8s)
    -s, --service SERVICE 指定要更新的服务 (backend|web|admin|all)
    -t, --tag TAG       指定镜像标签
    -d, --dry-run       仅显示将要执行的命令，不实际执行
    -f, --force         强制更新，跳过确认
    -v, --verbose       显示详细输出
    --no-backup         跳过备份
    --rollback          回滚到上一个版本

示例:
    $0 -e dev -m docker                    # 更新开发环境
    $0 -e prod -m k8s -s backend          # 更新生产环境后端服务
    $0 -e staging -m docker --rollback    # 回滚测试环境
    $0 --help                              # 显示帮助信息

EOF
}

# 默认参数
ENVIRONMENT="dev"
DEPLOY_MODE="docker"
SERVICE="all"
TAG="latest"
DRY_RUN=false
FORCE=false
VERBOSE=false
NO_BACKUP=false
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
        -s|--service)
            SERVICE="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
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
        --no-backup)
            NO_BACKUP=true
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
case $ENVIRONMENT in
    dev|staging|prod)
        ;;
    *)
        log_error "无效的环境: $ENVIRONMENT"
        exit 1
        ;;
esac

case $DEPLOY_MODE in
    docker|k8s)
        ;;
    *)
        log_error "无效的部署模式: $DEPLOY_MODE"
        exit 1
        ;;
esac

case $SERVICE in
    backend|web|admin|all)
        ;;
    *)
        log_error "无效的服务: $SERVICE"
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

# 创建备份
create_backup() {
    if [ "$NO_BACKUP" = true ]; then
        log_info "跳过备份"
        return 0
    fi
    
    log_info "创建备份..."
    
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    if [ "$DEPLOY_MODE" = "docker" ]; then
        # 备份 Docker 数据卷
        execute_command "docker-compose -f docker-compose.yml exec mysql mysqldump -u root -p\$MYSQL_ROOT_PASSWORD \$MYSQL_DATABASE > $backup_dir/database.sql" "备份数据库"
        execute_command "docker-compose -f docker-compose.yml exec redis redis-cli BGSAVE" "备份 Redis 数据"
    fi
    
    if [ "$DEPLOY_MODE" = "k8s" ]; then
        # 备份 Kubernetes 配置
        execute_command "kubectl get all -n zhihang-messenger -o yaml > $backup_dir/k8s-resources.yaml" "备份 Kubernetes 资源"
    fi
    
    log_success "备份创建完成: $backup_dir"
}

# 回滚操作
rollback() {
    log_info "执行回滚操作..."
    
    if [ "$DEPLOY_MODE" = "docker" ]; then
        # 回滚 Docker 服务
        execute_command "docker-compose -f docker-compose.yml down" "停止服务"
        execute_command "docker-compose -f docker-compose.yml up -d" "重新启动服务"
    fi
    
    if [ "$DEPLOY_MODE" = "k8s" ]; then
        # 回滚 Kubernetes 部署
        execute_command "kubectl rollout undo deployment/zhihang-messenger-backend -n zhihang-messenger" "回滚后端部署"
        execute_command "kubectl rollout undo deployment/zhihang-messenger-web -n zhihang-messenger" "回滚前端部署"
        execute_command "kubectl rollout undo deployment/zhihang-messenger-admin -n zhihang-messenger" "回滚管理后台部署"
    fi
    
    log_success "回滚完成"
}

# 更新 Docker 服务
update_docker() {
    log_info "更新 Docker 服务..."
    
    # 拉取最新镜像
    execute_command "docker-compose -f docker-compose.yml pull" "拉取最新镜像"
    
    # 重启指定服务
    if [ "$SERVICE" = "all" ]; then
        execute_command "docker-compose -f docker-compose.yml up -d" "重启所有服务"
    else
        execute_command "docker-compose -f docker-compose.yml up -d $SERVICE" "重启 $SERVICE 服务"
    fi
    
    # 清理旧镜像
    execute_command "docker image prune -f" "清理未使用的镜像"
    
    log_success "Docker 服务更新完成"
}

# 更新 Kubernetes 服务
update_k8s() {
    log_info "更新 Kubernetes 服务..."
    
    # 更新镜像标签
    if [ "$SERVICE" = "all" ] || [ "$SERVICE" = "backend" ]; then
        execute_command "kubectl set image deployment/zhihang-messenger-backend backend=zhihang-messenger/backend:$TAG -n zhihang-messenger" "更新后端镜像"
    fi
    
    if [ "$SERVICE" = "all" ] || [ "$SERVICE" = "web" ]; then
        execute_command "kubectl set image deployment/zhihang-messenger-web web=zhihang-messenger/web:$TAG -n zhihang-messenger" "更新前端镜像"
    fi
    
    if [ "$SERVICE" = "all" ] || [ "$SERVICE" = "admin" ]; then
        execute_command "kubectl set image deployment/zhihang-messenger-admin admin=zhihang-messenger/admin:$TAG -n zhihang-messenger" "更新管理后台镜像"
    fi
    
    # 等待部署完成
    log_info "等待部署完成..."
    execute_command "kubectl rollout status deployment/zhihang-messenger-backend -n zhihang-messenger" "等待后端部署完成"
    execute_command "kubectl rollout status deployment/zhihang-messenger-web -n zhihang-messenger" "等待前端部署完成"
    execute_command "kubectl rollout status deployment/zhihang-messenger-admin -n zhihang-messenger" "等待管理后台部署完成"
    
    log_success "Kubernetes 服务更新完成"
}

# 验证更新
verify_update() {
    log_info "验证更新结果..."
    
    # 等待服务启动
    sleep 10
    
    if [ "$DEPLOY_MODE" = "docker" ]; then
        # 检查容器状态
        execute_command "docker-compose -f docker-compose.yml ps" "检查容器状态"
        
        # 健康检查
        execute_command "curl -f http://localhost:8080/api/ping" "后端健康检查"
        execute_command "curl -f http://localhost:3000" "前端健康检查"
    fi
    
    if [ "$DEPLOY_MODE" = "k8s" ]; then
        # 检查 Pod 状态
        execute_command "kubectl get pods -n zhihang-messenger" "检查 Pod 状态"
        
        # 健康检查
        execute_command "kubectl port-forward -n zhihang-messenger svc/zhihang-messenger-backend 8080:8080 &" "设置端口转发"
        sleep 5
        execute_command "curl -f http://localhost:8080/api/ping" "后端健康检查"
        kill %1 2>/dev/null || true
    fi
    
    log_success "更新验证完成"
}

# 主函数
main() {
    log_info "开始更新志航密信..."
    log_info "环境: $ENVIRONMENT"
    log_info "部署模式: $DEPLOY_MODE"
    log_info "服务: $SERVICE"
    log_info "标签: $TAG"
    
    if [ "$ROLLBACK" = true ]; then
        log_warning "执行回滚操作"
        rollback
        verify_update
        log_success "🎉 回滚完成！"
        return 0
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN 模式 - 不会实际执行更新"
    fi
    
    # 确认更新
    if [ "$FORCE" != true ]; then
        echo
        log_warning "即将更新 $ENVIRONMENT 环境的 $SERVICE 服务"
        read -p "是否继续? (y/N): " -n 1 -r
        echo
        
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "更新已取消"
            exit 0
        fi
    fi
    
    # 创建备份
    create_backup
    
    # 执行更新
    case $DEPLOY_MODE in
        docker)
            update_docker
            ;;
        k8s)
            update_k8s
            ;;
    esac
    
    # 验证更新
    verify_update
    
    log_success "🎉 志航密信更新完成！"
    log_info "访问地址:"
    log_info "  - Web 端: http://localhost:3000"
    log_info "  - 管理后台: http://localhost:3001"
    log_info "  - API 文档: http://localhost:8080/api/ping"
}

# 脚本入口
main "$@"