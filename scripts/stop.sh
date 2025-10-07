#!/bin/bash

# 志航密信停止脚本
# 用于停止志航密信系统

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

# 停止服务
stop_services() {
    log_info "停止志航密信服务..."
    
    # 停止所有服务
    docker-compose down
    
    log_success "服务已停止"
}

# 清理资源（可选）
cleanup() {
    log_info "清理 Docker 资源..."
    
    # 清理未使用的镜像
    docker image prune -f
    
    # 清理未使用的网络
    docker network prune -f
    
    log_success "资源清理完成"
}

# 显示帮助信息
show_help() {
    echo "志航密信停止脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -c, --cleanup  停止服务并清理资源"
    echo "  -v, --volumes  停止服务并删除数据卷（危险操作）"
    echo ""
}

# 删除数据卷（危险操作）
remove_volumes() {
    log_warning "这将删除所有数据卷，包括数据库数据！"
    read -p "确定要继续吗？(y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "删除数据卷..."
        docker-compose down -v
        log_success "数据卷已删除"
    else
        log_info "操作已取消"
    fi
}

# 主函数
main() {
    case "${1:-}" in
        -h|--help)
            show_help
            ;;
        -c|--cleanup)
            stop_services
            cleanup
            ;;
        -v|--volumes)
            remove_volumes
            ;;
        "")
            stop_services
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"