# Telegram官方前端适配改造 - 详细方案

**创建时间**: 2025-10-12  
**状态**: ✅ 项目理解100%正确，开始技术实施

---

## ✅ 项目确认

### 用户视角（100%保持不变）
- ✅ 界面与官方Telegram完全一样
- ✅ 使用体验完全一样
- ✅ 用户无法察觉任何差异

### 技术视角（关键改造）
- ✅ 所有流量不走Telegram官方服务器
- ✅ 连接到自己的im-backend（Go服务器）
- ✅ 完全私有化部署
- ✅ 数据完全掌控

---

## 📊 源码确认完成

### ✅ Android端（官方）
- **位置**: `telegram-android/`
- **文件数**: 30,127个
- **大小**: 606.98 MB
- **关键文件已定位**: ✅
  - `TMessagesProj/jni/tgnet/ConnectionsManager.h`
  - `TMessagesProj/jni/tgnet/ConnectionsManager.cpp`
  - `TMessagesProj/src/main/java/org/telegram/tgnet/ConnectionsManager.java`

### ✅ Desktop端（官方）
- **位置**: `telegram-desktop/`
- **文件数**: 2,418个
- **关键文件已定位**: ✅
  - `Telegram/SourceFiles/mtproto/session.h`
  - `Telegram/SourceFiles/mtproto/session.cpp`
  - `Telegram/SourceFiles/apiwrap.h`
  - `Telegram/SourceFiles/apiwrap.cpp`

---

## 🔍 核心网络层分析

### Android端：ConnectionsManager

**关键方法签名**（从源码提取）:

```cpp
// 发送请求的核心方法
int32_t sendRequest(
    TLObject *object,              // MTProto TL对象
    onCompleteFunc onComplete,     // 完成回调
    onQuickAckFunc onQuickAck,     // 快速确认回调
    onRequestClearFunc onClear,    // 清理回调
    uint32_t flags,                // 请求标志
    uint32_t datacenterId,         // 数据中心ID
    ConnectionType connectionType, // 连接类型
    bool immediate                 // 是否立即发送
);

// 初始化方法
void init(
    uint32_t version,              // 版本号
    int32_t layer,                 // API层级
    int32_t apiId,                 // API ID
    std::string deviceModel,       // 设备型号
    std::string systemVersion,     // 系统版本
    std::string appVersion,        // 应用版本
    std::string langCode,          // 语言代码
    std::string systemLangCode,    // 系统语言
    std::string configPath,        // 配置路径
    std::string logPath,           // 日志路径
    std::string regId,             // 注册ID
    std::string cFingerprint,      // 证书指纹
    std::string installerId,       // 安装器ID
    std::string packageId,         // 包ID
    int32_t timezoneOffset,        // 时区偏移
    int64_t userId,                // 用户ID
    bool userPremium,              // 高级用户
    bool isPaused,                 // 是否暂停
    bool enablePushConnection,     // 启用推送
    bool hasNetwork,               // 网络状态
    int32_t networkType,           // 网络类型
    int32_t performanceClass       // 性能等级
);
```

**数据流程**:
```
Java层 → JNI → ConnectionsManager::sendRequest() 
    → Request对象序列化
    → Connection::sendData()
    → TCP/TLS传输
    → Telegram DataCenter
```

---

### Desktop端：Session

**关键方法签名**（从源码提取）:

```cpp
namespace MTP {
namespace details {

class Session {
public:
    // 发送准备好的请求
    void sendPrepared(
        const SerializedRequest &request,  // 已序列化的请求
        crl::time msCanWait = 0           // 可等待毫秒数
    );
    
    // 接收并处理响应
    void tryToReceive();
    
    // 认证密钥管理
    void setAuthKey(const AuthKeyPtr &key);
    AuthKeyPtr getTemporaryKey(TemporaryKeyType type) const;
    AuthKeyPtr getPersistentKey() const;
};

} // namespace details
} // namespace MTP
```

**数据流程**:
```
ApiWrap → api_request.cpp → Session::sendPrepared()
    → MTProto序列化
    → Connection传输
    → Telegram DataCenter
```

---

## 🎯 适配层架构设计

### 核心策略：拦截+转换

```
┌────────────────────────────────────────────────────────────┐
│                     原始流程                                 │
├────────────────────────────────────────────────────────────┤
│ UI层 → MTProto请求 → TCP加密传输 → Telegram服务器           │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│                     改造后流程                               │
├────────────────────────────────────────────────────────────┤
│ UI层 → MTProto请求 → 【ApiAdapter拦截】                      │
│                           ↓                                 │
│                      判断: 使用自建后端?                       │
│                           ↓                                 │
│                      转换为REST请求                           │
│                           ↓                                 │
│                      HTTP/WebSocket                         │
│                           ↓                                 │
│                      您的Go后端                              │
└────────────────────────────────────────────────────────────┘
```

---

## 📁 适配层文件结构

### Android端适配层

```
telegram-android/
└── TMessagesProj/
    ├── jni/adapter/                    # ← 新增C++适配层
    │   ├── ApiAdapter.h                # 适配器头文件
    │   ├── ApiAdapter.cpp              # 适配器实现
    │   ├── RestClient.h                # REST客户端头文件
    │   ├── RestClient.cpp              # REST客户端实现
    │   ├── ProtocolConverter.h         # 协议转换器
    │   ├── ProtocolConverter.cpp       # MTProto ↔ JSON转换
    │   └── Android.mk                  # NDK构建配置
    │
    └── src/main/java/org/telegram/
        └── adapter/                    # ← 新增Java适配层
            ├── BackendConfig.java      # 后端配置
            ├── ApiMode.java            # API模式枚举
            └── AdapterBridge.java      # JNI桥接
```

### Desktop端适配层

```
telegram-desktop/
└── Telegram/SourceFiles/
    └── adapter/                        # ← 新增C++适配层
        ├── api_adapter.h               # 适配器头文件
        ├── api_adapter.cpp             # 适配器实现
        ├── http_client.h               # HTTP客户端
        ├── http_client.cpp             # 客户端实现
        ├── protocol_converter.h        # 协议转换器
        ├── protocol_converter.cpp      # MTProto ↔ JSON转换
        └── config.h                    # 配置管理
```

---

## 🔧 核心改造点

### 改造点1：Android ConnectionsManager（关键）

**文件**: `TMessagesProj/jni/tgnet/ConnectionsManager.cpp`

**原始代码**（第XX行）:
```cpp
int32_t ConnectionsManager::sendRequest(TLObject *object, ...) {
    // 1. 创建Request对象
    Request *request = new Request(...);
    
    // 2. 序列化MTProto
    request->serializedData = object->serializeToByteBuffer();
    
    // 3. 添加到发送队列
    datacenter->addRequestToQueue(request);
    
    // 4. 触发网络发送
    wakeup();
}
```

**改造后**:
```cpp
int32_t ConnectionsManager::sendRequest(TLObject *object, ...) {
    // ← 添加适配层判断
    if (ApiAdapter::isCustomBackendEnabled()) {
        // 使用适配层
        return ApiAdapter::sendRestRequest(object, onComplete, ...);
    }
    
    // 原始MTProto流程（向后兼容）
    Request *request = new Request(...);
    request->serializedData = object->serializeToByteBuffer();
    datacenter->addRequestToQueue(request);
    wakeup();
}
```

---

### 改造点2：Desktop Session（关键）

**文件**: `Telegram/SourceFiles/mtproto/session.cpp`

**原始代码**（第422行）:
```cpp
void Session::sendPrepared(
        const SerializedRequest &request,
        crl::time msCanWait) {
    // 添加到发送队列
    {
        QWriteLocker locker(_data->toSendMutex());
        _data->toSendMap().emplace(request.requestId, request);
    }
    
    // 触发发送
    _data->queueSendAnything(msCanWait);
}
```

**改造后**:
```cpp
void Session::sendPrepared(
        const SerializedRequest &request,
        crl::time msCanWait) {
    // ← 添加适配层判断
    if (ApiAdapter::isCustomBackendEnabled()) {
        // 使用REST适配层
        ApiAdapter::sendHttpRequest(request, msCanWait);
        return;
    }
    
    // 原始MTProto流程
    {
        QWriteLocker locker(_data->toSendMutex());
        _data->toSendMap().emplace(request.requestId, request);
    }
    _data->queueSendAnything(msCanWait);
}
```

---

## 📋 协议转换设计

### TL Object → JSON转换

**示例1：发送消息**

**MTProto请求**（TL格式）:
```tl
messages.sendMessage#d9d75a4
    peer = {
        _: "inputPeerUser"
        user_id: 123456
        access_hash: 789012345
    }
    message = "Hello"
    random_id = 1234567890
```

**REST请求**（JSON格式）:
```json
POST /api/messages
{
  "receiver_id": 123456,
  "content": "Hello",
  "message_type": "text"
}
```

**转换逻辑**（C++）:
```cpp
class ProtocolConverter {
public:
    static json mtprotoToRest(TLObject *object) {
        if (object->getObjectType() == TL_messages_sendMessage::ID) {
            TL_messages_sendMessage *req = (TL_messages_sendMessage*)object;
            return {
                {"receiver_id", req->peer->user_id},
                {"content", req->message},
                {"message_type", "text"}
            };
        }
        // ... 其他类型
    }
    
    static TLObject* restToMtproto(json response) {
        // REST响应转换为MTProto对象
    }
};
```

---

### 示例2：获取对话列表

**MTProto**:
```tl
messages.getDialogs#a0f4cb4f
    offset_date = 0
    offset_id = 0
    offset_peer = inputPeerEmpty
    limit = 20
```

**REST**:
```http
GET /api/dialogs?limit=20&offset=0
```

---

### 示例3：上传文件

**MTProto**:
```tl
upload.saveFilePart#b304a621
    file_id = 12345
    file_part = 0
    bytes = [binary data]
```

**REST**:
```http
POST /api/files/upload
Content-Type: multipart/form-data

file: [binary data]
chunk: 0
```

---

## 🚀 实施计划（分15个任务）

### Phase 1: 准备工作（Task 1-3）

#### Task 1: ✅ 网络层分析完成
- ✅ Android: ConnectionsManager定位完成
- ✅ Desktop: Session定位完成
- ✅ 核心方法签名已提取

#### Task 2: 设计适配层架构
- 创建adapter模块目录结构
- 定义ApiAdapter接口
- 设计ProtocolConverter

#### Task 3: 创建配置系统
- 创建`im_config.json`
- 定义配置结构（服务器地址、端口、模式）

---

### Phase 2: Android适配层实现（Task 4-7）

#### Task 4: 创建C++适配层
- `ApiAdapter.h/cpp` - 核心适配逻辑
- `RestClient.h/cpp` - HTTP客户端（libcurl）
- `ProtocolConverter.h/cpp` - 协议转换

#### Task 5: Hook ConnectionsManager
- 修改`sendRequest`方法添加拦截
- 添加模式判断逻辑

#### Task 6: 创建Java配置层
- `BackendConfig.java` - 读取配置
- `ApiMode.java` - 模式切换

#### Task 7: 修改Android.mk/CMakeLists
- 添加适配层到编译
- 链接libcurl或OkHttp

---

### Phase 3: Desktop适配层实现（Task 8-11）

#### Task 8: 创建C++适配层
- `api_adapter.h/cpp` - 核心适配
- `http_client.h/cpp` - Qt网络层

#### Task 9: Hook Session
- 修改`sendPrepared`添加拦截
- 添加配置判断

#### Task 10: 协议转换实现
- `protocol_converter.h/cpp` - TL ↔ JSON

#### Task 11: 更新CMakeLists.txt
- 添加adapter模块到编译
- 链接Qt Network库

---

### Phase 4: 测试验证（Task 12-14）

#### Task 12: 功能测试
- 登录流程测试
- 消息收发测试
- 文件上传下载测试

#### Task 13: 编译打包
- Android: 编译APK
- Desktop: 编译可执行文件

#### Task 14: UI一致性验证
- 对比官方Telegram
- 确认无视觉差异

---

### Phase 5: 文档输出（Task 15）

#### Task 15: 上线报告
- 改造说明文档
- 构建部署指南
- 测试验证报告

---

## 📝 配置文件设计

### im_config.json

```json
{
  "api_mode": "custom_backend",
  "backend_config": {
    "base_url": "http://your-server:8080",
    "api_endpoint": "/api",
    "ws_endpoint": "/ws",
    "upload_endpoint": "/api/files/upload",
    "download_endpoint": "/api/files/download"
  },
  "fallback": {
    "enable_mtproto_fallback": false,
    "official_api_id": 0,
    "official_api_hash": ""
  },
  "advanced": {
    "enable_logs": true,
    "log_level": "debug",
    "connection_timeout": 30,
    "request_timeout": 60
  }
}
```

---

## 🔧 第一步：立即开始实施

我现在立即开始创建适配层框架！

**状态**: 🟢 完全理解，准备就绪，立即开工！

