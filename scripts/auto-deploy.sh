#!/bin/bash

###############################################################################
# 志航密信 v1.6.0 - 自动部署脚本
# 用途：一键部署所有服务，减少手动操作
# 使用：bash scripts/auto-deploy.sh
###############################################################################

set -e  # 遇到错误立即退出

# 颜色输出
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

# 步骤计数
STEP=1

print_step() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}步骤 $STEP: $1${NC}"
    echo -e "${GREEN}========================================${NC}"
    STEP=$((STEP + 1))
}

###############################################################################
# 1. 环境检查
###############################################################################

print_step "环境检查"

# 检查Go
if ! command -v go &> /dev/null; then
    log_error "未找到Go，请先安装Go 1.19+"
    exit 1
fi
log_success "Go版本: $(go version)"

# 检查Docker
if ! command -v docker &> /dev/null; then
    log_error "未找到Docker，请先安装Docker"
    exit 1
fi
log_success "Docker版本: $(docker --version)"

# 检查Docker Compose
if ! command -v docker-compose &> /dev/null; then
    log_error "未找到Docker Compose，请先安装"
    exit 1
fi
log_success "Docker Compose版本: $(docker-compose --version)"

###############################################################################
# 2. 环境变量配置
###############################################################################

print_step "配置环境变量"

if [ ! -f .env ]; then
    log_warning ".env文件不存在，从模板创建"
    
    if [ -f .env.example ]; then
        cp .env.example .env
        log_info "已从.env.example创建.env"
    else
        log_info "创建默认.env文件"
        cat > .env << 'EOF'
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=zhihang2025
DB_NAME=zhihang_messenger

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=zhihang_messenger_jwt_secret_key_2025
JWT_EXPIRES_IN=24h

# MinIO配置
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=zhihang-messenger
MINIO_USE_SSL=false

# 服务配置
PORT=8080
GIN_MODE=release

# WebSocket配置
WS_PORT=8081

# 文件上传配置
MAX_UPLOAD_SIZE=104857600
EOF
        log_success ".env文件已创建"
    fi
    
    log_warning "⚠️ 请编辑.env文件，修改数据库密码等配置！"
    log_info "按回车继续..."
    read
fi

log_success "环境变量配置完成"

###############################################################################
# 3. 启动依赖服务
###############################################################################

print_step "启动依赖服务（MySQL、Redis、MinIO）"

log_info "启动Docker服务..."
docker-compose -f docker-compose.production.yml up -d mysql redis minio

log_info "等待服务就绪（30秒）..."
sleep 30

# 检查服务状态
log_info "检查服务状态..."
if docker-compose -f docker-compose.production.yml ps | grep -q "Up"; then
    log_success "依赖服务启动成功"
    docker-compose -f docker-compose.production.yml ps
else
    log_error "服务启动失败"
    docker-compose -f docker-compose.production.yml logs
    exit 1
fi

###############################################################################
# 4. 编译后端
###############################################################################

print_step "编译后端服务"

cd im-backend

log_info "下载Go依赖..."
go mod download

log_info "编译后端..."
go build -o bin/im-backend main.go

if [ $? -eq 0 ]; then
    log_success "后端编译成功"
else
    log_error "后端编译失败"
    exit 1
fi

cd ..

###############################################################################
# 5. 启动后端
###############################################################################

print_step "启动后端服务"

log_info "启动后端（后台运行）..."
cd im-backend
nohup ./bin/im-backend > ../logs/backend.log 2>&1 &
BACKEND_PID=$!
cd ..

log_info "后端进程ID: $BACKEND_PID"
echo $BACKEND_PID > logs/backend.pid

log_info "等待后端启动（10秒）..."
sleep 10

# 检查进程是否运行
if ps -p $BACKEND_PID > /dev/null; then
    log_success "后端服务启动成功"
else
    log_error "后端服务启动失败，查看日志: logs/backend.log"
    exit 1
fi

###############################################################################
# 6. 健康检查
###############################################################################

print_step "健康检查"

log_info "测试健康检查API..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)

if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    log_success "健康检查通过: $HEALTH_RESPONSE"
else
    log_error "健康检查失败: $HEALTH_RESPONSE"
    log_info "查看日志: tail -f logs/backend.log"
    exit 1
fi

###############################################################################
# 7. 数据库验证
###############################################################################

print_step "验证数据库"

log_info "检查数据库表..."

# 获取数据库配置
source .env

TABLE_COUNT=$(mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME \
    -e "SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = '$DB_NAME';" \
    -s -N)

log_info "数据库表数量: $TABLE_COUNT"

if [ $TABLE_COUNT -ge 50 ]; then
    log_success "数据库表创建正常（共$TABLE_COUNT个表）"
else
    log_warning "数据库表数量偏少（$TABLE_COUNT个），可能有表未创建"
fi

# 检查关键表
REQUIRED_TABLES="users chats messages screen_share_sessions bots"
for table in $REQUIRED_TABLES; do
    EXISTS=$(mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME \
        -e "SHOW TABLES LIKE '$table';" -s -N)
    
    if [ -n "$EXISTS" ]; then
        log_success "表 $table 存在 ✅"
    else
        log_error "表 $table 不存在 ❌"
    fi
done

###############################################################################
# 8. 功能测试
###############################################################################

print_step "功能测试"

log_info "测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/register \
    -H "Content-Type: application/json" \
    -d '{
        "phone": "+8613800138000",
        "username": "testuser_'$(date +%s)'",
        "password": "Test123456",
        "nickname": "测试用户"
    }')

if echo "$REGISTER_RESPONSE" | grep -q "success"; then
    log_success "用户注册测试通过"
else
    log_warning "用户注册测试失败（可能用户已存在）: $REGISTER_RESPONSE"
fi

log_info "测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "phone": "+8613800138000",
        "password": "Test123456"
    }')

if echo "$LOGIN_RESPONSE" | grep -q "access_token"; then
    log_success "用户登录测试通过"
    
    # 提取token
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    log_info "Token: ${TOKEN:0:50}..."
    
    # 测试屏幕共享API
    log_info "测试屏幕共享状态API..."
    SHARE_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/calls/test_call/screen-share/status" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$SHARE_RESPONSE" | grep -q "success"; then
        log_success "屏幕共享API测试通过"
    else
        log_error "屏幕共享API测试失败: $SHARE_RESPONSE"
    fi
    
    # 测试统计API
    log_info "测试统计API..."
    STATS_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/screen-share/statistics" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$STATS_RESPONSE" | grep -q "success"; then
        log_success "统计API测试通过"
    else
        log_error "统计API测试失败: $STATS_RESPONSE"
    fi
else
    log_error "用户登录测试失败: $LOGIN_RESPONSE"
fi

###############################################################################
# 9. 完成
###############################################################################

print_step "部署完成"

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                  🎉 部署成功！                              ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}服务信息：${NC}"
echo -e "  后端API:        http://localhost:8080"
echo -e "  健康检查:       http://localhost:8080/health"
echo -e "  前端演示:       http://localhost:8000/examples/screen-share-demo.html"
echo ""
echo -e "${BLUE}进程信息：${NC}"
echo -e "  后端PID:        $BACKEND_PID"
echo -e "  日志文件:       logs/backend.log"
echo ""
echo -e "${BLUE}下一步：${NC}"
echo -e "  1. 查看日志：   tail -f logs/backend.log"
echo -e "  2. 测试API：    参考 DEPLOYMENT_FOR_DEVIN.md"
echo -e "  3. 前端测试：   打开 examples/screen-share-demo.html"
echo ""
echo -e "${BLUE}停止服务：${NC}"
echo -e "  kill $BACKEND_PID"
echo -e "  docker-compose -f docker-compose.production.yml down"
echo ""
log_success "部署脚本执行完毕！"


