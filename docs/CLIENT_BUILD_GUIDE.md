#!/bin/bash

###############################################################################
# 志航密信 - 客户端构建脚本
# 用途：自动构建Web和Android客户端
# 使用：bash scripts/build-clients.sh
###############################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

BUILD_DATE=$(date +%Y%m%d-%H%M%S)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_OUTPUT="$PROJECT_ROOT/builds/$BUILD_DATE"

log_info "========================================="
log_info "志航密信 - 客户端构建"
log_info "构建时间: $BUILD_DATE"
log_info "========================================="

# 创建构建输出目录
mkdir -p "$BUILD_OUTPUT"

# ========================================
# 1. 构建Web客户端
# ========================================
build_web_client() {
    log_info "========================================="
    log_info "1. 构建Web客户端（im-admin）"
    log_info "========================================="
    
    cd "$PROJECT_ROOT/im-admin"
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js未安装"
        return 1
    fi
    
    log_info "Node版本: $(node --version)"
    log_info "npm版本: $(npm --version)"
    
    # 清理旧构建
    log_info "清理旧构建..."
    rm -rf dist/ node_modules/
    
    # 安装依赖
    log_info "安装依赖..."
    npm install
    
    # 配置生产环境
    log_info "配置生产环境..."
    cat > .env.production << EOF
VITE_API_BASE_URL=http://154.37.214.191:8080
VITE_WS_URL=ws://154.37.214.191:8080/ws
VITE_APP_TITLE=志航密信
VITE_APP_VERSION=1.0.0
EOF
    
    # 构建
    log_info "开始构建..."
    npm run build
    
    # 打包
    log_info "打包构建产物..."
    cd dist
    zip -r "$BUILD_OUTPUT/zhihang-im-web-$BUILD_DATE.zip" .
    cd ..
    
    # 统计
    WEB_SIZE=$(du -sh dist | cut -f1)
    ZIP_SIZE=$(du -sh "$BUILD_OUTPUT/zhihang-im-web-$BUILD_DATE.zip" | cut -f1)
    
    log_success "Web客户端构建完成"
    log_success "  - 构建目录大小: $WEB_SIZE"
    log_success "  - 压缩包大小: $ZIP_SIZE"
    log_success "  - 文件位置: $BUILD_OUTPUT/zhihang-im-web-$BUILD_DATE.zip"
}

# ========================================
# 2. 构建Android客户端（React Native）
# ========================================
build_android_client() {
    log_info "========================================="
    log_info "2. 构建Android客户端（React Native）"
    log_info "========================================="
    
    # 检查环境
    if ! command -v npx &> /dev/null; then
        log_error "npx未安装"
        return 1
    fi
    
    if [ -z "$ANDROID_HOME" ]; then
        log_warning "ANDROID_HOME未设置，跳过Android构建"
        log_info "安装Android SDK后设置: export ANDROID_HOME=/path/to/sdk"
        return 0
    fi
    
    # 创建React Native项目
    log_info "初始化React Native项目..."
    cd /tmp
    npx react-native init ZhihangIM --skip-install
    
    cd ZhihangIM
    npm install
    
    # 配置API
    log_info "配置API端点..."
    mkdir -p src/config
    cat > src/config/api.js << 'EOF'
export const API_CONFIG = {
  BASE_URL: 'http://154.37.214.191:8080',
  WS_URL: 'ws://154.37.214.191:8080/ws',
  TIMEOUT: 30000,
};
EOF
    
    # 安装依赖
    log_info "安装IM相关依赖..."
    npm install axios react-native-webrtc @react-native-async-storage/async-storage
    
    # 构建APK
    log_info "构建Android APK..."
    cd android
    ./gradlew assembleRelease
    
    # 复制APK到输出目录
    cp app/build/outputs/apk/release/app-release.apk \
       "$BUILD_OUTPUT/zhihang-im-android-$BUILD_DATE.apk"
    
    APK_SIZE=$(du -sh "$BUILD_OUTPUT/zhihang-im-android-$BUILD_DATE.apk" | cut -f1)
    
    log_success "Android客户端构建完成"
    log_success "  - APK大小: $APK_SIZE"
    log_success "  - 文件位置: $BUILD_OUTPUT/zhihang-im-android-$BUILD_DATE.apk"
}

# ========================================
# 主函数
# ========================================
main() {
    # 构建Web客户端
    build_web_client || log_error "Web客户端构建失败"
    
    # 构建Android客户端（可选）
    build_android_client || log_warning "Android客户端构建跳过或失败"
    
    # 生成构建报告
    log_info "========================================="
    log_info "构建完成报告"
    log_info "========================================="
    echo "构建时间: $BUILD_DATE"
    echo "输出目录: $BUILD_OUTPUT"
    echo ""
    echo "构建产物:"
    ls -lh "$BUILD_OUTPUT/"
    echo ""
    
    cat > "$BUILD_OUTPUT/BUILD_INFO.txt" << EOF
志航密信 - 客户端构建信息

构建时间: $BUILD_DATE
构建机器: $(hostname)
系统信息: $(uname -a)

Web客户端:
- 技术栈: Vue 3 + Vite + Element Plus
- 构建工具: npm + Vite
- 文件: zhihang-im-web-$BUILD_DATE.zip

Android客户端:
- 技术栈: React Native
- 构建工具: Gradle + Android SDK
- 文件: zhihang-im-android-$BUILD_DATE.apk

部署说明:
1. Web客户端: 解压zip到Web服务器
2. Android客户端: 直接安装APK到Android设备

API配置:
- 后端API: http://154.37.214.191:8080
- WebSocket: ws://154.37.214.191:8080/ws
EOF
    
    log_success "========================================="
    log_success "所有客户端构建完成！"
    log_success "========================================="
}

main

