#!/bin/bash

# 志航密信生产环境部署脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# 检查环境
check_prerequisites() {
    log_info "检查部署环境..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装"
        exit 1
    fi
    
    log_success "环境检查完成"
}

# 创建目录
create_directories() {
    log_info "创建目录结构..."
    
    mkdir -p config/{mysql/{init,conf.d},redis,nginx/conf.d,prometheus,grafana/provisioning/{datasources,dashboards},filebeat}
    mkdir -p logs/{nginx,backend,admin,web}
    mkdir -p backups/{mysql,redis,minio}
    mkdir -p ssl
    
    log_success "目录创建完成"
}

# 生成配置
generate_configs() {
    log_info "生成配置文件..."
    
    if [ ! -f ".env.production" ]; then
        if [ -f "env.production.example" ]; then
            cp env.production.example .env.production
            log_warning "请编辑.env.production文件"
        else
            log_error "未找到环境变量示例文件"
            exit 1
        fi
    fi
    
    log_success "配置生成完成"
}

# 启动服务
start_services() {
    log_info "启动生产服务..."
    docker-compose -f docker-compose.production.yml up -d
    log_success "服务启动完成"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker-compose -f docker-compose.production.yml down
    log_success "服务停止完成"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    stop_services
    sleep 5
    start_services
    log_success "服务重启完成"
}

# 查看状态
show_status() {
    log_info "服务状态:"
    docker-compose -f docker-compose.production.yml ps
}

# 显示帮助
show_help() {
    echo "志航密信生产环境部署脚本"
    echo "使用方法: $0 [操作]"
    echo ""
    echo "操作:"
    echo "  init        - 初始化部署环境"
    echo "  start       - 启动所有服务"
    echo "  stop        - 停止所有服务"
    echo "  restart     - 重启所有服务"
    echo "  status      - 查看服务状态"
    echo "  help        - 显示帮助信息"
}

# 主函数
main() {
    case "${1:-help}" in
        "init")
            check_prerequisites
            create_directories
            generate_configs
            log_success "初始化完成"
            ;;
        "start")
            check_prerequisites
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "status")
            show_status
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

main "$@"
