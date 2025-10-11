# Telegram网络层深度分析

**日期**: 2025-10-12  
**目标**: 定位MTProto调用封装，为API适配层做准备

---

## 🎯 分析目标

### Android端
- 定位 `org.telegram.tgnet` 包
- 分析 `ConnectionsManager.cpp/h`
- 确认 Auth、消息、媒体调用路径

### Desktop端  
- 定位 `mtproto/` 目录
- 分析 `api.cpp` 和 `session.cpp`
- 确认网络请求封装

---

## 📁 Android网络层分析

### 核心文件位置

```
telegram-android/TMessagesProj/
├── jni/tgnet/                        # C++ 网络层（核心）
│   ├── ConnectionsManager.cpp        # 连接管理器（关键）
│   ├── ConnectionsManager.h          # 接口定义
│   ├── Datacenter.cpp                # 数据中心管理
│   ├── Connection.cpp                # TCP连接
│   ├── Request.cpp                   # 请求封装
│   ├── ApiScheme.cpp                 # API协议定义
│   └── MTProtoScheme.cpp             # MTProto协议
│
└── src/main/java/org/telegram/tgnet/ # Java 网络层（包装）
    ├── ConnectionsManager.java       # JNI包装
    ├── TLRPC.java                    # 协议对象
    └── TLObject.java                 # 基础对象
```

### 关键类：ConnectionsManager

**文件**: `TMessagesProj/jni/tgnet/ConnectionsManager.cpp`

#### 核心方法分析

```cpp
class ConnectionsManager {
public:
    // 发送请求（核心方法）
    int32_t sendRequest(
        TLObject *object,              // MTProto对象
        onCompleteFunc onComplete,     // 成功回调
        onQuickAckFunc onQuickAck,     // 快速确认
        uint32_t flags,                // 标志位
        uint32_t datacenterId,         // 数据中心ID
        ConnectionType conType,        // 连接类型
        bool immediate                 // 立即发送
    );
    
    // 处理服务器响应
    void processServerResponse(
        TL_message *message,           // 消息对象
        int64_t messageId,             // 消息ID
        int32_t messageSeqNo,          // 序列号
        int64_t messageSalt,           // 盐值
        Connection *connection,        // 连接对象
        int64_t innerMsgId,           // 内部消息ID
        int64_t containerMessageId     // 容器消息ID
    );
    
    // 初始化数据中心
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
        int32_t userId,                // 用户ID
        bool enablePushConnection,     // 启用推送
        bool hasNetwork,               // 网络状态
        int32_t networkType            // 网络类型
    );
};
```

### 数据流程分析

```
┌─────────────────────────────────────────────────────────────┐
│                    Telegram Android 网络流                    │
└─────────────────────────────────────────────────────────────┘

Java层（UI/业务逻辑）
    ↓
org.telegram.tgnet.ConnectionsManager (Java)
    ↓ JNI调用
    ↓
ConnectionsManager.cpp (C++)
    ↓
Request.cpp → 序列化MTProto对象
    ↓
Connection.cpp → 建立TCP连接
    ↓
加密层 → AES + RSA加密
    ↓
网络传输 → Socket发送到Telegram服务器
    ↓
DataCenter → 149.154.167.50:443 (DC1)
```

### 关键API调用示例

#### 1. 发送消息

```cpp
// Java层调用
TLRPC.TL_messages_sendMessage req = new TLRPC.TL_messages_sendMessage();
req.message = "Hello";
req.peer = peer;
ConnectionsManager.getInstance().sendRequest(req, (response, error) -> {
    // 处理响应
});

// ↓ 转换为C++层

// ConnectionsManager.cpp
int32_t requestId = sendRequest(
    tlObject,                    // TL_messages_sendMessage对象
    onCompleteCallback,          // 成功回调
    nullptr,                     // 快速确认
    RequestFlagWithoutLogin,     // 标志
    DEFAULT_DATACENTER_ID,       // DC ID
    ConnectionTypeGeneric,       // 连接类型
    true                         // 立即发送
);
```

#### 2. 获取对话列表

```cpp
// Java层
TLRPC.TL_messages_getDialogs req = new TLRPC.TL_messages_getDialogs();
req.offset_date = 0;
req.offset_id = 0;
req.offset_peer = new TLRPC.TL_inputPeerEmpty();
req.limit = 20;

// C++层处理
sendRequest(req, onComplete, nullptr, flags, dcId, type, false);
```

#### 3. 上传文件

```cpp
// Java层
TLRPC.TL_upload_saveFilePart req = new TLRPC.TL_upload_saveFilePart();
req.file_id = fileId;
req.file_part = partNum;
req.bytes = data;

// C++层序列化并发送
```

---

## 📁 Desktop网络层分析

### 核心文件位置

```
telegram-desktop/Telegram/SourceFiles/
├── mtproto/                          # MTProto协议层
│   ├── session.h                     # 会话管理（关键）
│   ├── session.cpp                   # 会话实现
│   ├── connection.h                  # 连接管理
│   ├── connection.cpp                # 连接实现
│   ├── mtproto_auth_key.cpp         # 认证密钥
│   ├── mtproto_dc_options.cpp       # DC选项
│   └── mtproto_response.cpp         # 响应处理
│
├── api/                              # API调用层
│   ├── api_request.cpp               # 请求封装
│   ├── api_response.cpp              # 响应处理
│   └── api_text_entities.cpp        # 文本实体
│
└── apiwrap.cpp                       # API包装器（核心）
```

### 关键类：Session

**文件**: `Telegram/SourceFiles/mtproto/session.cpp`

#### 核心方法分析

```cpp
namespace MTP {

class Session {
public:
    // 发送准备好的请求
    void sendPrepared(
        const SerializedRequest &request,  // 序列化的请求
        uint64 msCanWait = 0               // 可等待毫秒数
    );
    
    // 处理响应
    void handleResponse(mtpBuffer &buffer);
    
    // 处理认证密钥
    void setAuthKey(const AuthKeyPtr &key);
    
    // 获取数据中心ID
    [[nodiscard]] DcId getDcId() const;
};

} // namespace MTP
```

### 数据流程分析

```
┌─────────────────────────────────────────────────────────────┐
│                   Telegram Desktop 网络流                     │
└─────────────────────────────────────────────────────────────┘

Qt GUI层（UI）
    ↓
ApiWrap.cpp (API包装器)
    ↓
api_request.cpp (请求构造)
    ↓
Session.cpp (会话管理)
    ↓
Connection.cpp (连接管理)
    ↓
MTProto加密 (AES-IGE + RSA)
    ↓
TCP/TLS传输
    ↓
Telegram DataCenter
```

---

## 🔍 MTProto协议分析

### MTProto请求结构

```
┌──────────────────────────────────────┐
│         MTProto Request              │
├──────────────────────────────────────┤
│ auth_key_id      (int64)             │ ← 认证密钥ID
│ msg_key          (int128)            │ ← 消息密钥
│ encrypted_data:                      │
│   ├─ salt        (int64)             │ ← 盐值
│   ├─ session_id  (int64)             │ ← 会话ID
│   ├─ msg_id      (int64)             │ ← 消息ID
│   ├─ seq_no      (int32)             │ ← 序列号
│   ├─ msg_len     (int32)             │ ← 消息长度
│   └─ message:                        │
│       ├─ constructor (int32)         │ ← TL构造器
│       └─ params     (TL-serialized)  │ ← TL参数
└──────────────────────────────────────┘
```

### TL (Type Language) 示例

```tl
// 发送消息
messages.sendMessage#d9d75a4 
    flags:# 
    no_webpage:flags.1?true 
    silent:flags.5?true 
    background:flags.6?true 
    clear_draft:flags.7?true 
    peer:InputPeer 
    reply_to_msg_id:flags.0?int 
    message:string 
    random_id:long 
    reply_markup:flags.2?ReplyMarkup 
    entities:flags.3?Vector<MessageEntity> 
    schedule_date:flags.10?int 
    = Updates;
```

---

## 🎯 适配层设计要点

### 需要拦截的关键点

#### Android端拦截点

1. **ConnectionsManager::sendRequest()**
   - 位置：`jni/tgnet/ConnectionsManager.cpp`
   - 作用：所有API请求的入口
   - 拦截：在此处添加适配层判断

2. **Connection::connect()**
   - 位置：`jni/tgnet/Connection.cpp`
   - 作用：建立TCP连接
   - 拦截：替换服务器地址和端口

3. **Datacenter配置**
   - 位置：`jni/tgnet/Datacenter.cpp`
   - 作用：数据中心地址配置
   - 拦截：替换为自己的服务器地址

#### Desktop端拦截点

1. **Session::sendPrepared()**
   - 位置：`mtproto/session.cpp`
   - 作用：发送准备好的请求
   - 拦截：在此处添加适配层

2. **ApiWrap方法**
   - 位置：`apiwrap.cpp`
   - 作用：高级API包装
   - 拦截：可选的高层拦截点

3. **Connection::connectToServer()**
   - 位置：`mtproto/connection.cpp`
   - 作用：连接到服务器
   - 拦截：替换服务器地址

---

## 📊 协议映射设计

### MTProto → REST映射

| MTProto方法 | TL Constructor | REST API | HTTP方法 |
|------------|----------------|----------|---------|
| auth.sendCode | `auth.sendCode#ccfd70cf` | `/api/auth/send-code` | POST |
| auth.signIn | `auth.signIn#8d52a951` | `/api/auth/login` | POST |
| messages.sendMessage | `messages.sendMessage#d9d75a4` | `/api/messages` | POST |
| messages.getDialogs | `messages.getDialogs#a0f4cb4f` | `/api/dialogs` | GET |
| messages.getHistory | `messages.getHistory#4423e6c5` | `/api/messages?chat_id=xxx` | GET |
| upload.saveFilePart | `upload.saveFilePart#b304a621` | `/api/files/upload` | POST |
| upload.getFile | `upload.getFile#be5335be` | `/api/files/download` | GET |
| updates.getState | `updates.getState#edd4882a` | WebSocket `/ws` | WS |

---

## 🔧 下一步：适配层架构设计

### 目录结构（计划）

```
telegram-android/
└── TMessagesProj/
    ├── jni/
    │   ├── tgnet/              # 原始网络层
    │   └── adapter/            # ← 新增：适配层
    │       ├── ApiAdapter.cpp   # API适配器
    │       ├── ApiAdapter.h
    │       ├── RestClient.cpp   # REST客户端
    │       ├── RestClient.h
    │       └── Config.cpp       # 配置管理
    │
    └── src/main/java/org/telegram/
        └── adapter/            # ← 新增：Java适配层
            ├── BackendConfig.java
            └── ApiMode.java

telegram-desktop/
└── Telegram/SourceFiles/
    ├── mtproto/                # 原始协议层
    ├── api/                    # 原始API层
    └── adapter/                # ← 新增：适配层
        ├── api_adapter.h        # 适配器头文件
        ├── api_adapter.cpp      # 适配器实现
        ├── http_client.h        # HTTP客户端
        ├── http_client.cpp      # 客户端实现
        └── config.h             # 配置管理
```

---

## 📋 分析总结

### ✅ 已完成

1. ✅ 定位Android网络层核心：`ConnectionsManager.cpp`
2. ✅ 定位Desktop网络层核心：`Session.cpp`
3. ✅ 分析MTProto请求流程
4. ✅ 确定拦截点和适配层位置

### 🎯 下一步任务

1. 设计适配层接口（`ApiAdapter`）
2. 实现MTProto → REST转换逻辑
3. 创建配置文件（`im_config.json`）
4. 开始Android端适配层编码

---

**状态**: 🟢 网络层分析完成，准备进入适配层设计阶段

