#!/bin/bash

# 志航密信部署脚本
# 用于一键部署志航密信系统

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

# 检查 Docker 和 Docker Compose
check_requirements() {
    log_info "检查系统要求..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "系统要求检查通过"
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    mkdir -p scripts/ssl
    mkdir -p scripts/nginx/conf.d
    mkdir -p logs
    
    log_success "目录创建完成"
}

# 生成 SSL 证书
generate_ssl_certificates() {
    log_info "生成 SSL 证书..."
    
    if [ ! -f "scripts/ssl/cert.pem" ] || [ ! -f "scripts/ssl/key.pem" ]; then
        log_info "生成自签名 SSL 证书..."
        openssl req -x509 -newkey rsa:4096 -keyout scripts/ssl/key.pem -out scripts/ssl/cert.pem -days 365 -nodes -subj "/C=CN/ST=Beijing/L=Beijing/O=志航密信/OU=IT/CN=localhost"
        log_success "SSL 证书生成完成"
    else
        log_info "SSL 证书已存在，跳过生成"
    fi
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    # 构建后端镜像
    log_info "构建后端镜像..."
    docker-compose build backend
    
    # 构建 Web 镜像
    log_info "构建 Web 镜像..."
    docker-compose build web
    
    log_success "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 启动数据库和缓存服务
    log_info "启动数据库和缓存服务..."
    docker-compose up -d mysql redis minio
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    sleep 30
    
    # 启动后端服务
    log_info "启动后端服务..."
    docker-compose up -d backend
    
    # 等待后端启动
    log_info "等待后端服务启动..."
    sleep 10
    
    # 启动 Web 服务
    log_info "启动 Web 服务..."
    docker-compose up -d web
    
    # 启动 Nginx
    log_info "启动 Nginx 反向代理..."
    docker-compose up -d nginx
    
    log_success "所有服务启动完成"
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."
    
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

# 显示访问信息
show_access_info() {
    log_success "志航密信部署完成！"
    echo ""
    echo "访问信息："
    echo "  Web 端: https://localhost"
    echo "  API 文档: https://localhost/api/ping"
    echo "  MinIO 控制台: http://localhost:9001"
    echo ""
    echo "默认账号："
    echo "  MinIO: zhihang_messenger / zhihang_messenger_pass"
    echo "  MySQL: zhihang_messenger / zhihang_messenger_pass"
    echo ""
    echo "管理命令："
    echo "  查看日志: docker-compose logs -f"
    echo "  停止服务: docker-compose down"
    echo "  重启服务: docker-compose restart"
    echo "  更新服务: docker-compose pull && docker-compose up -d"
}

# 主函数
main() {
    log_info "开始部署志航密信系统..."
    
    check_requirements
    create_directories
    generate_ssl_certificates
    build_images
    start_services
    check_services
    show_access_info
    
    log_success "部署完成！"
}

# 执行主函数
main "$@"
