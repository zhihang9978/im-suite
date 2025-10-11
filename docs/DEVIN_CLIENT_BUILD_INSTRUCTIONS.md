# 📋 Devin客户端构建指令

**创建时间**: 2025-10-12 00:30  
**状态**: ✅ **可立即执行**  
**目标**: 构建Web和Android客户端

---

## 🎯 构建目标

1. ✅ **Web客户端**: 基于im-admin的Vue3管理后台
2. ✅ **Android客户端**: React Native跨平台应用

---

## 🚀 方案1：一键构建（推荐）

### 执行命令

```bash
cd /home/ubuntu/repos/im-suite

# 拉取最新代码
git pull origin main

# 执行一键构建脚本
bash scripts/build-clients.sh

# 构建产物位置:
# builds/YYYYMMDD-HHMMSS/zhihang-im-web-*.zip
# builds/YYYYMMDD-HHMMSS/zhihang-im-android-*.apk
```

**预计时间**: 15-30分钟  
**产物**: Web ZIP包 + Android APK

---

## 🌐 方案2：单独构建Web客户端（快速）

### 简化版本（只构建Web）

```bash
cd /home/ubuntu/repos/im-suite/im-admin

# 1. 安装依赖
npm install

# 2. 配置生产环境
cat > .env.production << 'EOF'
VITE_API_BASE_URL=http://154.37.214.191:8080
VITE_WS_URL=ws://154.37.214.191:8080/ws
VITE_APP_TITLE=志航密信
EOF

# 3. 构建
npm run build

# 4. 打包
cd dist
zip -r ../zhihang-im-web-$(date +%Y%m%d).zip .
cd ..

echo "✅ Web客户端构建完成"
echo "📦 文件: zhihang-im-web-*.zip"
```

**预计时间**: 5-10分钟  
**产物**: Web ZIP包（约5-10MB）

---

## 📱 方案3：构建Android客户端选项

### 选项A: React Native（推荐）

**优势**:
- ✅ 一次开发，iOS+Android都支持
- ✅ 热更新支持
- ✅ 开发速度快
- ✅ 社区活跃

**构建命令**:
```bash
# 检查环境
echo "Android SDK: $ANDROID_HOME"
echo "Java版本: $(java -version 2>&1 | head -n 1)"

# 如果ANDROID_HOME未设置，需要先安装
if [ -z "$ANDROID_HOME" ]; then
    echo "需要安装Android SDK"
    echo "参考: https://reactnative.dev/docs/environment-setup"
    exit 1
fi

# 初始化项目
cd /home/ubuntu
npx react-native init ZhihangIM --skip-install
cd ZhihangIM

# 安装依赖
npm install
npm install axios react-native-webrtc @react-native-async-storage/async-storage

# 配置API
mkdir -p src/config
cat > src/config/api.js << 'EOF'
export const API_CONFIG = {
  BASE_URL: 'http://154.37.214.191:8080',
  WS_URL: 'ws://154.37.214.191:8080/ws',
};
EOF

# 构建Android
cd android
./gradlew assembleRelease

# APK位置
echo "APK: android/app/build/outputs/apk/release/app-release.apk"
```

**预计时间**: 30-60分钟（首次构建）  
**产物**: APK（约30-40MB）

---

### 选项B: Telegram Android修改

**优势**:
- ✅ 功能完整（基于成熟的Telegram）
- ✅ UI/UX优秀
- ✅ 性能优化好

**劣势**:
- ❌ 代码量巨大
- ❌ 编译时间长
- ❌ 需要Telegram API credentials

**构建命令**:
```bash
cd /home/ubuntu
mkdir -p telegram-builds
cd telegram-builds

# 克隆源码
git clone --depth 1 https://github.com/DrKLO/Telegram.git
cd Telegram

# 配置（需要先申请API）
# 访问 https://my.telegram.org/apps

# 编辑 TMessagesProj/src/main/java/org/telegram/messenger/BuildVars.java
# 设置 API_ID 和 API_HASH

# 构建
./gradlew assembleRelease

# APK位置
echo "APK: TMessagesProj/build/outputs/apk/release/app-release-unsigned.apk"
```

**预计时间**: 1-2小时（首次构建）  
**产物**: APK（约50-60MB）

---

## 📋 构建前置要求

### Web客户端
- ✅ Node.js 16+
- ✅ npm 8+
- ✅ 网络连接

### Android客户端（React Native）
- ✅ Node.js 16+
- ✅ npm 8+
- ✅ Java JDK 11+
- ✅ Android SDK (API 24+)
- ✅ Android Build Tools
- ✅ 环境变量 `ANDROID_HOME`

### 检查命令
```bash
# 检查所有依赖
node --version          # 应该 >= 16
npm --version           # 应该 >= 8
java -version           # 应该 >= 11
echo $ANDROID_HOME      # 应该有值
$ANDROID_HOME/tools/bin/sdkmanager --list | head -20
```

---

## 🎯 推荐执行顺序

### 第1步：构建Web客户端（必需）

```bash
cd /home/ubuntu/repos/im-suite/im-admin
npm install
npm run build
cd dist && zip -r ../zhihang-im-web-$(date +%Y%m%d).zip . && cd ..
```

**时间**: 5-10分钟  
**产物**: Web ZIP包

---

### 第2步：测试Web客户端

```bash
# 解压到临时目录
cd /tmp
unzip /home/ubuntu/repos/im-suite/im-admin/zhihang-im-web-*.zip -d test-web

# 启动测试服务器
cd test-web
python3 -m http.server 8000

# 浏览器访问: http://154.37.214.191:8000
# 测试登录、消息等功能
```

---

### 第3步：构建Android客户端（可选）

```bash
# 如果环境已准备好
cd /home/ubuntu
npx react-native init ZhihangIM
cd ZhihangIM
npm install axios react-native-webrtc
cd android && ./gradlew assembleRelease
```

**时间**: 30-60分钟  
**产物**: Android APK

---

### 第4步：测试Android客户端

```bash
# 安装到Android设备或模拟器
adb install ZhihangIM/android/app/build/outputs/apk/release/app-release.apk

# 启动应用
adb shell am start -n com.zhihangim/.MainActivity

# 查看日志
adb logcat | grep ZhihangIM
```

---

## 📦 构建产物说明

### Web客户端
```
zhihang-im-web-20251012-0030.zip
  ├── index.html
  ├── assets/
  │   ├── index-[hash].js
  │   └── index-[hash].css
  └── favicon.ico
  
大小: 约5-10MB
部署: 解压到Nginx/Apache的www目录
访问: http://your-domain.com
```

### Android客户端
```
zhihang-im-android-20251012.apk

大小: 约30-50MB
安装: adb install 或直接在手机上安装
支持: Android 7.0+ (API 24+)
```

---

## 🎊 构建成功后

### Web客户端部署

**选项1: Docker部署**
```bash
cd /home/ubuntu/repos/im-suite
docker-compose -f docker-compose.production.yml up -d admin
```

**选项2: Nginx部署**
```bash
sudo mkdir -p /var/www/im-web
sudo unzip zhihang-im-web-*.zip -d /var/www/im-web

# Nginx配置
sudo nano /etc/nginx/sites-available/im-web
```

---

### Android客户端分发

**选项1: 直接分发APK**
```bash
# 上传到服务器
scp zhihang-im-android-*.apk user@server:/var/www/downloads/
```

**选项2: 通过Google Play或应用商店**
```bash
# 需要：
# 1. 开发者账号
# 2. 签名的APK
# 3. 应用描述、截图等
```

---

## 📌 注意事项

### ⚠️ 不要在本地处理

**严禁操作**:
- ❌ 不要在 `im-suite/telegram-web/` 操作
- ❌ 不要在 `im-suite/telegram-android/` 操作  
- ❌ 会导致网络错误

**正确方式**:
- ✅ 在独立目录（如 `/home/ubuntu/telegram-clients/`）操作
- ✅ 或使用im-admin作为Web客户端
- ✅ 或创建新的React Native项目

---

### ✅ 签名APK（Android）

**生产环境必需**:
```bash
# 生成密钥（首次）
keytool -genkey -v \
  -keystore ~/zhihang-im.keystore \
  -alias zhihang-im \
  -keyalg RSA \
  -keysize 2048 \
  -validity 10000

# 签名APK
jarsigner -verbose \
  -keystore ~/zhihang-im.keystore \
  app-release.apk \
  zhihang-im

# 对齐APK
zipalign -v 4 app-release.apk zhihang-im-signed.apk
```

---

## 🎯 Devin执行建议

### 最简单方案（推荐）

**只构建Web客户端**:
```bash
cd /home/ubuntu/repos/im-suite/im-admin
npm install && npm run build
cd dist && zip -r ../web-client.zip . && cd ..
```

**时间**: 5分钟  
**产物**: web-client.zip  
**用途**: 可立即部署使用

---

### 完整方案（需要时间）

**构建Web + Android**:
```bash
cd /home/ubuntu/repos/im-suite
bash scripts/build-clients.sh
```

**时间**: 30-60分钟  
**产物**: Web ZIP + Android APK  
**用途**: 完整的客户端套件

---

**🎉 客户端构建方案已准备完毕！Devin可以根据时间和需求选择方案执行！**

---

**准备人**: AI Assistant (Cursor)  
**准备时间**: 2025-10-12 00:30  
**方案数量**: 3个（简单、中等、完整）  
**推荐方案**: 使用im-admin作为Web客户端  
**状态**: ✅ **可立即执行**

