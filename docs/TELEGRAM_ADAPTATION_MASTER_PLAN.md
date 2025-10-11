# Telegram客户端适配总体规划

## 📌 项目目标

**核心目标**：修改官方Telegram Android和Desktop客户端，使其连接到自建Go后端，同时保持100%原版UI和用户体验。

**技术本质**：
- ❌ 不是：重写Telegram客户端
- ✅ 是：添加API适配层，将MTProto协议调用转换为REST/WebSocket调用

---

## 🏗️ 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Telegram UI层（保持不变）                   │
│  • 聊天界面、联系人列表、设置页面                               │
│  • 所有动画、交互逻辑                                          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│              Telegram业务逻辑层（保持不变）                     │
│  • MessagesController, DialogsController                     │
│  • FileLoader, ContactsController                            │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│         🔥 API适配层（NEW - 我们要添加的）                      │
│  ┌───────────────────────────────────────────────────┐      │
│  │ TelegramMethodInterceptor                         │      │
│  │  - 拦截所有TLObject调用                            │      │
│  │  - 识别方法类型（Auth/Message/File/Contact）       │      │
│  └───────────────────────────────────────────────────┘      │
│                              ↓                               │
│  ┌───────────────────────────────────────────────────┐      │
│  │ ProtocolConverter                                 │      │
│  │  - MTProto TLObject → JSON                        │      │
│  │  - JSON Response → TLObject                       │      │
│  └───────────────────────────────────────────────────┘      │
│                              ↓                               │
│  ┌───────────────────────────────────────────────────┐      │
│  │ RestClient / WebSocketClient                      │      │
│  │  - HTTP请求到自建后端                              │      │
│  │  - WebSocket长连接（消息推送）                     │      │
│  └───────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│              自建Go后端（im-backend）                          │
│  • REST API: /api/auth/*, /api/messages/*, /api/files/*    │
│  • WebSocket: /ws (实时消息推送)                             │
│  • MySQL + Redis                                            │
└─────────────────────────────────────────────────────────────┘
```

---

## 📋 详细实施步骤

### 阶段1：准备与分析（已完成✅）

#### 1.1 源码获取
- ✅ 已获取Telegram Android源码（Telegram-master/）
- ✅ 已获取Telegram Desktop源码（tdesktop-dev/）
- ✅ 已移动到正确位置（telegram-android/, telegram-desktop/）

#### 1.2 网络层分析
- ✅ 定位Android关键文件：
  - `ConnectionsManager.cpp` - 网络连接管理
  - `NativeByteBuffer.cpp` - 数据序列化
  - `Request.cpp` - 请求封装
  - `TLObject.java` - TL对象基类
  - `TLRPC.java` - 所有API方法定义（18000+行）
  
- ✅ 定位Desktop关键文件：
  - `mtproto/session.cpp` - 会话管理
  - `api/api_request.cpp` - API请求
  - `mtproto/connection.cpp` - 底层连接
  - `api/*.tl` - TL Schema定义

---

### 阶段2：设计适配层架构（当前阶段🔥）

#### 2.1 Android适配层设计

**文件结构**：
```
telegram-android/
└── TMessagesProj/
    ├── jni/
    │   └── adapter/                    # C++ 适配层（JNI）
    │       ├── TelegramMethodInterceptor.h/cpp    # 方法拦截器
    │       ├── ProtocolConverter.h/cpp            # 协议转换器
    │       ├── RestClient.h/cpp                   # REST客户端
    │       ├── WebSocketClient.h/cpp              # WebSocket客户端
    │       └── AdapterConfig.h/cpp                # 配置管理
    │
    └── src/main/java/
        └── org/telegram/adapter/         # Java 适配层接口
            ├── AdapterManager.java       # 适配器管理类
            ├── BackendConfig.java        # 后端配置
            └── RequestCallback.java      # 回调接口
```

**关键类说明**：

##### `TelegramMethodInterceptor`
- **作用**：拦截所有Telegram API调用
- **输入**：TLObject对象（如TL_auth_sendCode, TL_messages_sendMessage）
- **输出**：路由到对应的转换方法
- **Hook点**：`ConnectionsManager::sendRequest()`

##### `ProtocolConverter`
- **作用**：协议转换核心
- **MTProto → REST**：
  ```cpp
  TL_auth_sendCode → POST /api/auth/send-code
  TL_messages_sendMessage → POST /api/messages/send
  TL_messages_getDialogs → GET /api/messages/dialogs
  ```
- **REST → MTProto**：
  ```cpp
  JSON Response → TL_auth_sentCode
  JSON Response → TL_messages_messages
  ```

##### `RestClient`
- **作用**：HTTP请求封装
- **功能**：
  - 支持GET/POST/PUT/DELETE
  - 自动添加JWT Token
  - 处理请求重试
  - 错误码映射

##### `WebSocketClient`
- **作用**：实时消息推送
- **功能**：
  - 连接到ws://backend:port/ws
  - 心跳保持
  - 消息接收 → 转换为TLObject → 通知UI层

##### `AdapterConfig`
- **作用**：配置管理
- **配置项**：
  ```json
  {
    "backend_url": "https://your-backend.com",
    "ws_url": "wss://your-backend.com/ws",
    "api_version": "v1",
    "timeout_ms": 30000,
    "enable_cache": true,
    "debug_mode": false
  }
  ```

#### 2.2 Desktop适配层设计

**文件结构**：
```
telegram-desktop/
└── Telegram/
    └── SourceFiles/
        └── adapter/                      # C++ 适配层
            ├── method_interceptor.h/cpp  # 方法拦截器
            ├── protocol_converter.h/cpp  # 协议转换器
            ├── rest_client.h/cpp         # REST客户端
            ├── websocket_client.h/cpp    # WebSocket客户端
            └── config.h/cpp              # 配置管理
```

**关键差异**：
- Desktop使用Qt框架 → 使用`QNetworkAccessManager`
- Desktop的TL Schema使用代码生成 → 需要修改生成器
- Desktop有独立的MTProto实现 → Hook点在`MTP::Instance::send()`

---

### 阶段3：实现适配层代码（下一步）

#### 3.1 Android - JNI C++层实现

**步骤**：
1. 创建adapter目录和文件
2. 实现RestClient（基于libcurl或Android NDK）
3. 实现WebSocketClient（基于libwebsockets）
4. 实现ProtocolConverter（关键转换逻辑）
5. 实现TelegramMethodInterceptor（Hook ConnectionsManager）
6. 修改CMakeLists.txt添加编译目标

#### 3.2 Android - Java层接口

**步骤**：
1. 创建AdapterManager单例
2. 提供配置接口给UI层
3. 实现JNI桥接方法
4. 添加初始化逻辑到Application类

#### 3.3 Desktop - C++层实现

**步骤**：
1. 创建adapter模块
2. 使用Qt网络库实现RestClient
3. 使用Qt WebSocket实现WebSocketClient
4. 实现协议转换器
5. Hook MTP层的请求发送
6. 修改CMakeLists.txt

---

### 阶段4：修改原有网络调用（关键步骤🔥）

#### 4.1 Android修改点

**文件**：`TMessagesProj/jni/tgnet/ConnectionsManager.cpp`

**原始代码**（第856行左右）：
```cpp
void ConnectionsManager::sendRequest(TLObject *object, ...) {
    // 原始：发送到Telegram服务器
    datacenter->connection->sendData(serializedData);
}
```

**修改后**：
```cpp
void ConnectionsManager::sendRequest(TLObject *object, ...) {
    // 🔥 拦截：路由到适配层
    if (AdapterConfig::isEnabled()) {
        IMAdapter::TelegramMethodInterceptor::intercept(object, onComplete, onError);
        return;
    }
    
    // 保留原始逻辑（用于调试/回退）
    datacenter->connection->sendData(serializedData);
}
```

#### 4.2 Desktop修改点

**文件**：`Telegram/SourceFiles/mtproto/session.cpp`

**原始代码**：
```cpp
void Session::send(mtpRequestId requestId, SerializedRequest &&request) {
    // 原始：发送MTProto数据包
    _connection->sendData(std::move(request));
}
```

**修改后**：
```cpp
void Session::send(mtpRequestId requestId, SerializedRequest &&request) {
    // 🔥 拦截：路由到适配层
    if (Adapter::Config::isEnabled()) {
        Adapter::MethodInterceptor::intercept(requestId, request);
        return;
    }
    
    _connection->sendData(std::move(request));
}
```

---

### 阶段5：协议转换映射表（核心逻辑）

#### 5.1 认证模块映射

| Telegram MTProto | REST API | 说明 |
|-----------------|----------|------|
| `TL_auth_sendCode` | `POST /api/auth/send-code` | 发送验证码 |
| `TL_auth_signIn` | `POST /api/auth/sign-in` | 登录 |
| `TL_auth_signUp` | `POST /api/auth/register` | 注册 |
| `TL_auth_logout` | `POST /api/auth/logout` | 登出 |

**示例转换**：
```cpp
// MTProto Request
TL_auth_sendCode request;
request.phone_number = "+8613800138000";
request.api_id = 6;
request.api_hash = "eb06d4abfb49dc3eeb1aeb98ae0f581e";

// ↓ 转换为 ↓

// REST Request
POST /api/auth/send-code
Content-Type: application/json

{
  "phone": "+8613800138000"
}

// ↓ 后端响应 ↓

// REST Response
{
  "success": true,
  "data": {
    "phone_code_hash": "abc123xyz",
    "timeout": 60
  }
}

// ↓ 转换回 ↓

// MTProto Response
TL_auth_sentCode response;
response.phone_code_hash = "abc123xyz";
response.timeout = 60;
```

#### 5.2 消息模块映射

| Telegram MTProto | REST API | 说明 |
|-----------------|----------|------|
| `TL_messages_sendMessage` | `POST /api/messages/send` | 发送消息 |
| `TL_messages_getDialogs` | `GET /api/messages/dialogs` | 获取会话列表 |
| `TL_messages_getHistory` | `GET /api/messages/history/{chat_id}` | 获取聊天记录 |
| `TL_messages_readHistory` | `POST /api/messages/read` | 标记已读 |

#### 5.3 文件模块映射

| Telegram MTProto | REST API | 说明 |
|-----------------|----------|------|
| `TL_upload_saveFilePart` | `POST /api/files/upload` | 上传文件分片 |
| `TL_upload_getFile` | `GET /api/files/download/{file_id}` | 下载文件 |

#### 5.4 联系人模块映射

| Telegram MTProto | REST API | 说明 |
|-----------------|----------|------|
| `TL_contacts_getContacts` | `GET /api/contacts` | 获取联系人列表 |
| `TL_contacts_search` | `GET /api/users/search?phone={phone}` | 搜索用户 |

---

### 阶段6：WebSocket实时消息对接

#### 6.1 连接建立

```cpp
// 客户端连接到WebSocket
ws://backend.com/ws?token={JWT_TOKEN}

// 后端验证Token后建立连接
```

#### 6.2 消息推送

```cpp
// 后端推送新消息
{
  "type": "new_message",
  "data": {
    "message_id": 12345,
    "from_user_id": 999,
    "to_chat_id": 888,
    "text": "你好",
    "timestamp": 1697000000
  }
}

// ↓ 适配层转换 ↓

// 转换为TL_updateNewMessage
TL_updateNewMessage update;
update.message = new TL_message();
update.message->id = 12345;
update.message->from_id = 999;
update.message->message = "你好";

// ↓ 发送到Telegram UI层 ↓

// 通知MessagesController处理
NotificationCenter.postNotification(NotificationCenter.didReceiveNewMessages);
```

---

### 阶段7：配置文件与开关

#### 7.1 配置文件位置

**Android**：
```
/sdcard/Android/data/org.telegram.messenger/files/im_config.json
```

**Desktop**：
```
~/.local/share/TelegramDesktop/im_config.json
```

#### 7.2 配置文件格式

```json
{
  "adapter_enabled": true,
  "backend": {
    "rest_api_url": "https://your-backend.com",
    "websocket_url": "wss://your-backend.com/ws",
    "api_version": "v1"
  },
  "auth": {
    "auto_refresh_token": true,
    "token_expire_hours": 24
  },
  "network": {
    "connect_timeout_ms": 10000,
    "read_timeout_ms": 30000,
    "retry_max_times": 3
  },
  "debug": {
    "log_enabled": true,
    "log_level": "INFO"
  }
}
```

#### 7.3 UI设置页面

在Telegram设置中添加"服务器设置"选项：
- 后端地址输入
- 启用/禁用自定义后端
- 连接状态显示
- 日志查看

---

### 阶段8：编译与测试

#### 8.1 Android编译

```bash
cd telegram-android
./gradlew assembleRelease

# 输出APK：
# TMessagesProj/build/outputs/apk/release/app-arm64-v8a-release.apk
```

#### 8.2 Desktop编译

```bash
cd telegram-desktop
mkdir build && cd build
cmake ..
make -j8

# 输出可执行文件：
# Telegram
```

#### 8.3 测试计划

**Phase 1 - 基础功能**：
- [ ] 登录（发送验证码 → 输入验证码 → 登录成功）
- [ ] 获取会话列表
- [ ] 查看聊天记录

**Phase 2 - 消息功能**：
- [ ] 发送文本消息
- [ ] 接收实时消息（WebSocket推送）
- [ ] 消息已读状态

**Phase 3 - 文件功能**：
- [ ] 发送图片
- [ ] 发送文件
- [ ] 下载文件

**Phase 4 - 高级功能**：
- [ ] 语音通话（需要TURN服务器）
- [ ] 视频通话
- [ ] 群组管理

---

## ⚠️ 关键风险与注意事项

### 风险1：TLObject序列化复杂度
- **问题**：Telegram有500+种TLObject类型
- **方案**：优先实现高频20种，逐步扩展

### 风险2：WebSocket断线重连
- **问题**：网络切换导致消息丢失
- **方案**：实现重连+消息队列+补发机制

### 风险3：文件分片上传
- **问题**：大文件上传需要分片管理
- **方案**：保持Telegram原有分片逻辑，适配层透传

### 风险4：加密兼容性
- **问题**：Telegram使用MTProto 2.0加密
- **方案**：
  - 在适配层做TLS加密（https/wss）
  - 后端做权限验证
  - 不需要实现MTProto加密

### 风险5：版本兼容性
- **问题**：Telegram频繁更新
- **方案**：
  - 固定使用当前版本源码
  - 适配层做成独立模块，便于后续升级

---

## 📊 工作量评估

| 阶段 | 工作量 | 优先级 | 状态 |
|-----|--------|-------|------|
| 1. 源码分析 | 2天 | P0 | ✅ 完成 |
| 2. 架构设计 | 1天 | P0 | 🔥 进行中 |
| 3. Android适配层实现 | 5天 | P0 | 待开始 |
| 4. Desktop适配层实现 | 5天 | P1 | 待开始 |
| 5. 协议转换器实现 | 3天 | P0 | 待开始 |
| 6. WebSocket对接 | 2天 | P0 | 待开始 |
| 7. 配置与UI | 2天 | P1 | 待开始 |
| 8. 测试与调试 | 5天 | P0 | 待开始 |
| **总计** | **25天** | - | **8%完成** |

---

## 🎯 下一步行动

1. **审阅本规划** ← 当前步骤
2. **确认技术方案可行性**
3. **开始编写Android适配层代码**
4. **编译测试最小功能（登录）**
5. **逐步扩展其他功能**

---

## 📝 技术决策记录

### 决策1：适配层位置
- ✅ 选择：在JNI C++层实现
- 原因：所有网络调用都经过C++层，修改点集中

### 决策2：协议转换方式
- ✅ 选择：显式映射表 + 转换函数
- 原因：清晰可维护，便于调试

### 决策3：配置管理
- ✅ 选择：JSON配置文件 + UI设置页面
- 原因：灵活性高，用户可自行配置

### 决策4：是否实现MTProto
- ✅ 选择：不实现，使用HTTPS+WSS替代
- 原因：降低复杂度，TLS同样安全

---

## 📚 参考资料

- [Telegram Android源码](https://github.com/DrKLO/Telegram)
- [Telegram Desktop源码](https://github.com/telegramdesktop/tdesktop)
- [MTProto协议文档](https://core.telegram.org/mtproto)
- [Telegram API方法列表](https://core.telegram.org/methods)
- [TL Language规范](https://core.telegram.org/mtproto/TL)

---

**文档版本**：v1.0
**创建时间**：2025-10-11
**最后更新**：2025-10-11

