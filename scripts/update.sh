#!/bin/bash

# 志航密信更新脚本
# 用于更新志航密信系统

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

# 备份数据
backup_data() {
    log_info "备份数据..."
    
    # 创建备份目录
    mkdir -p backups/$(date +%Y%m%d_%H%M%S)
    BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
    
    # 备份数据库
    log_info "备份 MySQL 数据库..."
    docker-compose exec mysql mysqldump -u zhihang_messenger -pzhihang_messenger_pass zhihang_messenger > "$BACKUP_DIR/database.sql"
    
    # 备份 Redis 数据
    log_info "备份 Redis 数据..."
    docker-compose exec redis redis-cli BGSAVE
    docker cp $(docker-compose ps -q redis):/data/dump.rdb "$BACKUP_DIR/redis.rdb"
    
    # 备份 MinIO 数据
    log_info "备份 MinIO 数据..."
    docker cp $(docker-compose ps -q minio):/data "$BACKUP_DIR/minio_data"
    
    log_success "数据备份完成: $BACKUP_DIR"
}

# 拉取最新镜像
pull_images() {
    log_info "拉取最新镜像..."
    
    docker-compose pull
    
    log_success "镜像拉取完成"
}

# 重新构建镜像
rebuild_images() {
    log_info "重新构建镜像..."
    
    # 构建后端镜像
    log_info "构建后端镜像..."
    docker-compose build --no-cache backend
    
    # 构建 Web 镜像
    log_info "构建 Web 镜像..."
    docker-compose build --no-cache web
    
    log_success "镜像构建完成"
}

# 更新服务
update_services() {
    log_info "更新服务..."
    
    # 停止服务
    log_info "停止服务..."
    docker-compose down
    
    # 启动服务
    log_info "启动服务..."
    docker-compose up -d
    
    log_success "服务更新完成"
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."
    
    # 等待服务启动
    sleep 30
    
    # 检查容器状态
    docker-compose ps
    
    # 检查后端健康状态
    log_info "检查后端服务健康状态..."
    if curl -f http://localhost:8080/api/ping > /dev/null 2>&1; then
        log_success "后端服务运行正常"
    else
        log_warning "后端服务可能未完全启动，请稍后检查"
    fi
    
    # 检查 Web 服务
    log_info "检查 Web 服务..."
    if curl -f http://localhost:3000 > /dev/null 2>&1; then
        log_success "Web 服务运行正常"
    else
        log_warning "Web 服务可能未完全启动，请稍后检查"
    fi
}

# 显示更新信息
show_update_info() {
    log_success "志航密信更新完成！"
    echo ""
    echo "访问信息："
    echo "  Web 端: https://localhost"
    echo "  API 文档: https://localhost/api/ping"
    echo "  MinIO 控制台: http://localhost:9001"
    echo ""
    echo "管理命令："
    echo "  查看日志: docker-compose logs -f"
    echo "  停止服务: docker-compose down"
    echo "  重启服务: docker-compose restart"
    echo ""
}

# 显示帮助信息
show_help() {
    echo "志航密信更新脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -b, --backup   更新前备份数据"
    echo "  -r, --rebuild  重新构建镜像"
    echo "  -f, --force    强制更新（不备份）"
    echo ""
}

# 主函数
main() {
    local backup=false
    local rebuild=false
    local force=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -b|--backup)
                backup=true
                shift
                ;;
            -r|--rebuild)
                rebuild=true
                shift
                ;;
            -f|--force)
                force=true
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    log_info "开始更新志航密信系统..."
    
    # 备份数据（如果需要）
    if [ "$backup" = true ] && [ "$force" = false ]; then
        backup_data
    fi
    
    # 拉取最新镜像
    pull_images
    
    # 重新构建镜像（如果需要）
    if [ "$rebuild" = true ]; then
        rebuild_images
    fi
    
    # 更新服务
    update_services
    
    # 检查服务状态
    check_services
    
    # 显示更新信息
    show_update_info
    
    log_success "更新完成！"
}

# 执行主函数
main "$@"