# Telegram适配层实施清单

## 📦 需要创建的文件清单

### Android端（telegram-android/）

#### C++ 适配层（JNI）
```
TMessagesProj/jni/adapter/
├── AdapterConfig.h              # 配置管理头文件
├── AdapterConfig.cpp            # 配置管理实现
├── RestClient.h                 # REST客户端头文件
├── RestClient.cpp               # REST客户端实现（使用libcurl）
├── WebSocketClient.h            # WebSocket客户端头文件
├── WebSocketClient.cpp          # WebSocket客户端实现（使用libwebsockets）
├── ProtocolConverter.h          # 协议转换器头文件
├── ProtocolConverter.cpp        # 协议转换器实现
├── TelegramMethodInterceptor.h  # 方法拦截器头文件
├── TelegramMethodInterceptor.cpp # 方法拦截器实现
└── CMakeLists.txt              # 编译配置
```

#### Java 接口层
```
TMessagesProj/src/main/java/org/telegram/adapter/
├── AdapterManager.java          # 适配器管理单例
├── BackendConfig.java           # 后端配置类
├── RequestCallback.java         # 请求回调接口
└── AdapterJNI.java             # JNI桥接类
```

#### UI 配置界面
```
TMessagesProj/src/main/java/org/telegram/ui/
└── BackendSettingsActivity.java # 后端设置页面
```

---

### Desktop端（telegram-desktop/）

#### C++ 适配层
```
Telegram/SourceFiles/adapter/
├── config.h                     # 配置管理头文件
├── config.cpp                   # 配置管理实现
├── rest_client.h                # REST客户端头文件（基于QNetworkAccessManager）
├── rest_client.cpp              # REST客户端实现
├── websocket_client.h           # WebSocket客户端头文件（基于QWebSocket）
├── websocket_client.cpp         # WebSocket客户端实现
├── protocol_converter.h         # 协议转换器头文件
├── protocol_converter.cpp       # 协议转换器实现
├── method_interceptor.h         # 方法拦截器头文件
└── method_interceptor.cpp       # 方法拦截器实现
```

#### UI 配置界面
```
Telegram/SourceFiles/settings/
└── settings_backend.h/cpp       # 后端设置页面
```

---

## 🔧 需要修改的文件清单

### Android端修改

#### 1. ConnectionsManager.cpp（核心Hook点）
**文件**：`telegram-android/TMessagesProj/jni/tgnet/ConnectionsManager.cpp`
**修改位置**：`sendRequest()` 方法（约856行）
**修改内容**：添加适配层拦截逻辑

#### 2. ConnectionsManager.h
**文件**：`telegram-android/TMessagesProj/jni/tgnet/ConnectionsManager.h`
**修改位置**：头文件引用部分
**修改内容**：添加adapter头文件引用

#### 3. CMakeLists.txt
**文件**：`telegram-android/TMessagesProj/jni/CMakeLists.txt`
**修改位置**：源文件列表
**修改内容**：添加adapter目录下的cpp文件

#### 4. build.gradle
**文件**：`telegram-android/TMessagesProj/build.gradle`
**修改位置**：dependencies部分
**修改内容**：
```gradle
dependencies {
    // 添加网络库
    implementation 'com.squareup.okhttp3:okhttp:4.10.0'
    implementation 'com.google.code.gson:gson:2.10.1'
}
```

#### 5. ApplicationLoader.java
**文件**：`telegram-android/TMessagesProj/src/main/java/org/telegram/messenger/ApplicationLoader.java`
**修改位置**：`onCreate()` 方法
**修改内容**：初始化AdapterManager

---

### Desktop端修改

#### 1. session.cpp（核心Hook点）
**文件**：`telegram-desktop/Telegram/SourceFiles/mtproto/session.cpp`
**修改位置**：`send()` 方法
**修改内容**：添加适配层拦截逻辑

#### 2. session.h
**文件**：`telegram-desktop/Telegram/SourceFiles/mtproto/session.h`
**修改位置**：头文件引用部分
**修改内容**：添加adapter头文件引用

#### 3. CMakeLists.txt
**文件**：`telegram-desktop/Telegram/CMakeLists.txt`
**修改位置**：源文件列表
**修改内容**：添加adapter目录下的cpp文件

#### 4. main.cpp
**文件**：`telegram-desktop/Telegram/SourceFiles/main/main.cpp`
**修改位置**：应用初始化部分
**修改内容**：初始化Adapter::Config

---

## 📋 实施步骤详细拆解

### 第1步：创建Android适配层框架（1天）

#### 1.1 创建目录结构
```bash
mkdir -p telegram-android/TMessagesProj/jni/adapter
mkdir -p telegram-android/TMessagesProj/src/main/java/org/telegram/adapter
```

#### 1.2 创建配置管理类
- `AdapterConfig.h/cpp` - 读取JSON配置文件
- 功能：
  - 读取`/sdcard/Android/data/org.telegram.messenger/files/im_config.json`
  - 提供`isEnabled()`, `getBackendUrl()`, `getWSUrl()`等接口

#### 1.3 创建REST客户端
- `RestClient.h/cpp` - 封装HTTP请求
- 依赖：libcurl（Android NDK自带）
- 功能：
  - `GET()`, `POST()`, `PUT()`, `DELETE()`
  - 自动添加JWT Token到Header
  - 超时处理和重试

#### 1.4 创建WebSocket客户端
- `WebSocketClient.h/cpp` - 封装WebSocket连接
- 依赖：libwebsockets
- 功能：
  - 连接到`ws://backend/ws?token=xxx`
  - 心跳保持（每30秒ping）
  - 消息接收回调

---

### 第2步：实现协议转换器（2天）

#### 2.1 分析TLRPC.java
**文件**：`telegram-android/TMessagesProj/src/main/java/org/telegram/tgnet/TLRPC.java`
**任务**：
- 识别高频使用的TLObject类型（前20个）
- 记录每个类的字段结构

**示例**：
```java
// TL_auth_sendCode
public static class TL_auth_sendCode extends TLObject {
    public String phone_number;
    public int api_id;
    public String api_hash;
    // ...
}
```

#### 2.2 实现ProtocolConverter
**核心函数**：
```cpp
RestRequest ProtocolConverter::tlToRest(TLObject *object) {
    RestRequest req;
    
    // 根据constructor判断类型
    switch (object->constructor) {
        case 0xa677244f:  // TL_auth_sendCode
            req.method = "POST";
            req.url = "/api/auth/send-code";
            req.body = convertAuthSendCode(object);
            break;
            
        case 0xbca9ae22:  // TL_auth_signIn
            req.method = "POST";
            req.url = "/api/auth/sign-in";
            req.body = convertAuthSignIn(object);
            break;
            
        // ... 其他20种类型
    }
    
    return req;
}
```

#### 2.3 实现JSON → TLObject转换
```cpp
TLObject* ProtocolConverter::restToTl(const std::string &json, TLObject *originalRequest) {
    // 解析JSON
    nlohmann::json j = nlohmann::json::parse(json);
    
    // 根据原始请求类型判断响应类型
    if (originalRequest->constructor == 0xa677244f) {
        // TL_auth_sentCode
        TL_auth_sentCode *response = new TL_auth_sentCode();
        response->phone_code_hash = j["data"]["phone_code_hash"];
        response->timeout = j["data"]["timeout"];
        return response;
    }
    
    // ... 其他类型
}
```

---

### 第3步：实现方法拦截器（1天）

#### 3.1 创建TelegramMethodInterceptor
```cpp
void TelegramMethodInterceptor::intercept(
    TLObject *object,
    onCompleteFunc onComplete,
    onErrorFunc onError
) {
    // 1. 转换TLObject → REST请求
    RestRequest req = ProtocolConverter::tlToRest(object);
    
    // 2. 发送HTTP请求
    RestClient::instance()->send(req, [=](const std::string &response) {
        // 3. 转换REST响应 → TLObject
        TLObject *tlResponse = ProtocolConverter::restToTl(response, object);
        
        // 4. 调用原始回调
        onComplete(tlResponse);
    }, [=](int errorCode, const std::string &errorMsg) {
        // 错误处理
        onError(errorCode, errorMsg);
    });
}
```

---

### 第4步：Hook ConnectionsManager（0.5天）

#### 4.1 修改ConnectionsManager.cpp
**位置**：`sendRequest()` 方法开始处

**添加代码**：
```cpp
void ConnectionsManager::sendRequest(TLObject *object, onCompleteFunc onComplete, ...) {
    // 🔥 添加适配层拦截
    if (AdapterConfig::isEnabled()) {
        DEBUG_D("Intercepting request: 0x%x", object->constructor);
        IMAdapter::TelegramMethodInterceptor::intercept(object, onComplete, onError);
        return;
    }
    
    // 原始逻辑保持不变
    // ...
}
```

---

### 第5步：创建Java接口层（0.5天）

#### 5.1 AdapterManager.java
```java
public class AdapterManager {
    private static AdapterManager instance;
    
    public static AdapterManager getInstance() {
        if (instance == null) {
            instance = new AdapterManager();
        }
        return instance;
    }
    
    public void initialize() {
        // 加载配置
        BackendConfig.load();
        
        // 初始化JNI
        AdapterJNI.init();
    }
    
    public boolean isEnabled() {
        return AdapterJNI.isAdapterEnabled();
    }
}
```

#### 5.2 在ApplicationLoader中初始化
```java
@Override
public void onCreate() {
    super.onCreate();
    
    // 初始化适配器
    AdapterManager.getInstance().initialize();
    
    // ... 原有初始化代码
}
```

---

### 第6步：实现WebSocket消息推送（1天）

#### 6.1 WebSocketClient实现
```cpp
class WebSocketClient {
public:
    void connect(const std::string &url, const std::string &token) {
        // 建立WebSocket连接
        ws_connect(url + "?token=" + token);
    }
    
    void onMessage(const std::string &message) {
        // 解析JSON消息
        nlohmann::json j = nlohmann::json::parse(message);
        
        if (j["type"] == "new_message") {
            // 转换为TL_updateNewMessage
            TL_updateNewMessage *update = new TL_updateNewMessage();
            update->message = parseMessage(j["data"]);
            
            // 通知Telegram核心层
            NotificationCenter::getInstance()->postNotification(
                NotificationCenter::didReceiveNewMessages,
                update
            );
        }
    }
};
```

---

### 第7步：编译测试（1天）

#### 7.1 修改CMakeLists.txt
```cmake
# 添加adapter源文件
set(ADAPTER_SRC
    adapter/AdapterConfig.cpp
    adapter/RestClient.cpp
    adapter/WebSocketClient.cpp
    adapter/ProtocolConverter.cpp
    adapter/TelegramMethodInterceptor.cpp
)

# 添加到编译目标
add_library(tmessages SHARED
    ${NATIVE_SRC}
    ${ADAPTER_SRC}  # 添加这行
)

# 添加libcurl和libwebsockets
target_link_libraries(tmessages
    curl
    websockets
)
```

#### 7.2 编译APK
```bash
cd telegram-android
./gradlew assembleDebug
```

#### 7.3 测试步骤
1. 安装APK到手机
2. 将配置文件推送到手机：
   ```bash
   adb push im_config.json /sdcard/Android/data/org.telegram.messenger/files/
   ```
3. 启动APP，查看日志：
   ```bash
   adb logcat | grep "IMAdapter"
   ```
4. 尝试登录，观察是否请求到自建后端

---

### 第8步：Desktop端实现（5天）

#### 8.1 创建适配层（与Android类似）
- 使用Qt网络库（QNetworkAccessManager, QWebSocket）
- Hook点在`mtproto/session.cpp`

#### 8.2 协议转换器（复用Android逻辑）

#### 8.3 编译测试
```bash
cd telegram-desktop
mkdir build && cd build
cmake ..
make -j8
./Telegram
```

---

## 🧪 测试用例

### 测试用例1：登录流程
```
步骤：
1. 启动APP
2. 点击"Start Messaging"
3. 输入手机号：+8613800138000
4. 点击"Next"

期望结果：
- 网络请求发送到自建后端（检查日志）
- 收到验证码短信
- 输入验证码后登录成功
```

### 测试用例2：发送消息
```
步骤：
1. 登录成功后进入聊天列表
2. 选择一个联系人
3. 输入文本消息
4. 点击发送

期望结果：
- 消息显示在聊天界面
- 对方（另一台设备）通过WebSocket收到实时消息
- 消息同步到MySQL数据库
```

### 测试用例3：接收消息
```
步骤：
1. 保持APP在后台
2. 从另一台设备发送消息

期望结果：
- WebSocket推送消息到客户端
- 显示系统通知
- 打开APP后消息已在聊天界面
```

---

## 📊 进度追踪表

| 步骤 | Android | Desktop | 完成度 |
|-----|---------|---------|--------|
| 1. 创建框架 | ⬜ 未开始 | ⬜ 未开始 | 0% |
| 2. 协议转换器 | ⬜ 未开始 | ⬜ 未开始 | 0% |
| 3. 方法拦截器 | ⬜ 未开始 | ⬜ 未开始 | 0% |
| 4. Hook网络层 | ⬜ 未开始 | ⬜ 未开始 | 0% |
| 5. Java/Qt接口 | ⬜ 未开始 | ⬜ 未开始 | 0% |
| 6. WebSocket | ⬜ 未开始 | ⬜ 未开始 | 0% |
| 7. 编译测试 | ⬜ 未开始 | ⬜ 未开始 | 0% |
| 8. 功能验证 | ⬜ 未开始 | ⬜ 未开始 | 0% |

---

## 🎯 里程碑

- **Milestone 1**：完成Android框架搭建（第1-3步）
- **Milestone 2**：实现登录功能（第4-5步）
- **Milestone 3**：实现消息收发（第6步）
- **Milestone 4**：完成Android全功能（第7步）
- **Milestone 5**：完成Desktop全功能（第8步）

---

**更新时间**：2025-10-11
**当前状态**：等待审阅

