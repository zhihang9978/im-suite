# 志航密信 - Android客户端

**React Native跨平台IM客户端（支持iOS和Android）**

## 创建和构建（给Devin）

### 第1步：初始化项目

```bash
cd /home/ubuntu

# 创建React Native项目
npx react-native init ZhihangIM --skip-install

cd ZhihangIM
npm install
```

### 第2步：安装依赖

```bash
npm install \
  axios \
  @react-native-async-storage/async-storage \
  react-native-webrtc \
  @react-navigation/native \
  @react-navigation/stack \
  react-native-gesture-handler \
  react-native-reanimated \
  react-native-screens \
  react-native-safe-area-context
```

### 第3步：配置API

创建 `src/config/api.js`:

```javascript
export const API_CONFIG = {
  BASE_URL: 'http://154.37.214.191:8080',
  WS_URL: 'ws://154.37.214.191:8080/ws',
  TIMEOUT: 30000,
};
```

### 第4步：创建核心组件

需要创建的文件：

```
src/
├── screens/
│   ├── LoginScreen.js          # 登录界面
│   ├── ChatListScreen.js       # 聊天列表
│   ├── ChatScreen.js           # 聊天界面
│   ├── ContactsScreen.js       # 联系人
│   └── SettingsScreen.js       # 设置
├── components/
│   ├── MessageBubble.js        # 消息气泡
│   ├── ChatInput.js            # 输入框
│   └── Avatar.js               # 头像
├── services/
│   ├── api.js                  # API调用
│   └── websocket.js            # WebSocket
├── stores/
│   └── index.js                # 状态管理
└── App.js                      # 主应用
```

### 第5步：构建APK

```bash
cd android

# 清理
./gradlew clean

# 构建Release版本
./gradlew assembleRelease

# APK位置
ls -lh app/build/outputs/apk/release/app-release.apk
```

## 签名APK（生产环境必需）

### 生成密钥

```bash
keytool -genkey -v \
  -keystore ~/zhihang-im-release.keystore \
  -alias zhihang-im \
  -keyalg RSA \
  -keysize 2048 \
  -validity 10000
```

### 配置签名

编辑 `android/gradle.properties`:

```properties
MYAPP_RELEASE_STORE_FILE=zhihang-im-release.keystore
MYAPP_RELEASE_KEY_ALIAS=zhihang-im
MYAPP_RELEASE_STORE_PASSWORD=your_store_password
MYAPP_RELEASE_KEY_PASSWORD=your_key_password
```

编辑 `android/app/build.gradle`:

```gradle
android {
    signingConfigs {
        release {
            if (project.hasProperty('MYAPP_RELEASE_STORE_FILE')) {
                storeFile file(MYAPP_RELEASE_STORE_FILE)
                storePassword MYAPP_RELEASE_STORE_PASSWORD
                keyAlias MYAPP_RELEASE_KEY_ALIAS
                keyPassword MYAPP_RELEASE_KEY_PASSWORD
            }
        }
    }
    buildTypes {
        release {
            signingConfig signingConfigs.release
        }
    }
}
```

### 构建签名APK

```bash
cd android
./gradlew assembleRelease

# 签名APK位置
ls -lh app/build/outputs/apk/release/app-release.apk
```

## 功能清单

### ✅ 基础功能
- ✅ 用户登录/注册
- ✅ 实时消息（WebSocket）
- ✅ 文本消息收发
- ✅ 联系人列表
- ✅ 聊天列表

### ⏳ 待实现
- ⏳ 图片/文件消息
- ⏳ 语音视频通话
- ⏳ 群组聊天
- ⏳ 推送通知
- ⏳ 本地数据库
- ⏳ 消息加密

## 测试

```bash
# 在模拟器运行
npx react-native run-android

# 在真机运行（USB调试）
adb devices
npx react-native run-android

# 查看日志
npx react-native log-android
```

## 最终产物

- `app-release.apk` - 未签名版本（测试用）
- `app-release-signed.apk` - 签名版本（生产用）

## 技术栈

- **框架**: React Native
- **导航**: React Navigation
- **HTTP**: Axios
- **WebSocket**: 原生WebSocket
- **状态管理**: React Context/Redux
- **音视频**: react-native-webrtc

---

**这是真正的Android IM客户端，可以安装到手机使用！**

