# Telegram官方前端适配改造 - 开始

**日期**: 2025-10-12  
**目标**: 将Telegram官方Android和Desktop前端改造为连接自己的后端

---

## ✅ 源码确认完成

### Android端（官方）
- **原目录**: `Telegram-master/`
- **目标目录**: `telegram-android/`
- **状态**: ✅ 已复制
- **文件数**: ~20,103个文件
- **技术栈**: Java + Kotlin + JNI(C++)
- **关键文件**:
  - `TMessagesProj/jni/tgnet/ConnectionsManager.cpp` - 网络管理器
  - `TMessagesProj/jni/tgnet/ApiScheme.cpp` - API协议
  - `TMessagesProj/src/main/java/org/telegram/tgnet/` - Java网络层

### Desktop端（官方）
- **原目录**: `tdesktop-dev/`
- **目标目录**: `telegram-desktop/`
- **状态**: ✅ 已就位
- **文件数**: ~2,418个文件
- **技术栈**: C++ + Qt
- **关键文件**:
  - `Telegram/SourceFiles/mtproto/` - MTProto协议实现
  - `Telegram/SourceFiles/api/` - API调用层
  - `Telegram/SourceFiles/apiwrap.cpp` - API包装器

---

## 🎯 改造策略

### 核心原则
**不改UI，只改网络层**

1. ✅ 保持原有UI和交互完全不变
2. ✅ 创建API适配层，拦截MTProto调用
3. ✅ 将MTProto请求转换为REST/WebSocket
4. ✅ 连接到自己的后端（Go服务器）

---

## 📋 改造计划

### Phase 1: Android端改造（3-5天）

#### 1.1 分析网络层架构
**文件**: `TMessagesProj/jni/tgnet/ConnectionsManager.cpp`

**关键类**:
```cpp
class ConnectionsManager {
    void sendRequest(TL_Object *object, 
                     onCompleteFunc onComplete,
                     onQuickAckFunc onQuickAck = nullptr);
    
    void processServerResponse(NativeByteBuffer *buffer, 
                              int64_t messageId);
}
```

**MTProto数据流**:
```
App → Java层 → JNI → ConnectionsManager → TCP/TLS → Telegram服务器
```

#### 1.2 创建适配层
**新增文件**:
1. `TMessagesProj/jni/adapter/ApiAdapter.cpp` - C++适配层
2. `TMessagesProj/jni/adapter/RestClient.cpp` - HTTP客户端
3. `TMessagesProj/src/main/java/org/telegram/adapter/BackendConfig.java` - 配置类

**适配层架构**:
```
App → Java层 → JNI → ApiAdapter → RestClient → 您的Go后端
                      ↓
                 (拦截MTProto)
                      ↓
              (转换为REST API)
```

#### 1.3 修改网络配置
**配置文件**: 创建 `backend_config.json`
```json
{
  "api_base_url": "http://your-server:8080/api",
  "ws_url": "ws://your-server:8080/ws",
  "enable_mtproto": false,
  "use_custom_backend": true
}
```

---

### Phase 2: Desktop端改造（3-5天）

#### 2.1 分析网络层架构
**文件**: `Telegram/SourceFiles/mtproto/session.cpp`

**关键类**:
```cpp
namespace MTP {
    class Session {
        void sendPrepared(
            const SerializedRequest &request,
            uint64 msCanWait = 0);
        
        void handleResponse(mtpBuffer &buffer);
    };
}
```

**MTProto数据流**:
```
App → Qt → Session → MTProto → TCP/TLS → Telegram服务器
```

#### 2.2 创建适配层
**新增文件**:
1. `Telegram/SourceFiles/adapter/api_adapter.h` - 适配层头文件
2. `Telegram/SourceFiles/adapter/api_adapter.cpp` - 适配层实现
3. `Telegram/SourceFiles/adapter/http_client.cpp` - HTTP客户端

**适配层架构**:
```
App → Qt → ApiAdapter → HttpClient → 您的Go后端
           ↓
      (拦截MTProto)
           ↓
      (转换为REST API)
```

#### 2.3 修改CMake配置
**文件**: `Telegram/CMakeLists.txt`

添加适配层编译:
```cmake
# 添加适配层源文件
set(ADAPTER_SOURCES
    SourceFiles/adapter/api_adapter.cpp
    SourceFiles/adapter/http_client.cpp
)

# 添加到编译目标
target_sources(Telegram PRIVATE ${ADAPTER_SOURCES})
```

---

## 🔧 API映射表

### MTProto → REST 映射关系

| MTProto方法 | REST API | 说明 |
|------------|----------|------|
| `auth.sendCode` | `POST /api/auth/send-code` | 发送验证码 |
| `auth.signIn` | `POST /api/auth/login` | 登录 |
| `messages.sendMessage` | `POST /api/messages` | 发送消息 |
| `messages.getHistory` | `GET /api/messages?chat_id=xxx` | 获取消息历史 |
| `upload.saveFilePart` | `POST /api/files/upload` | 上传文件 |
| `updates.getDifference` | WebSocket连接 | 实时更新 |

---

## 📝 实施步骤

### Step 1: 创建适配层框架（1天）
- [x] 确认源码位置
- [ ] 创建适配层目录结构
- [ ] 编写API映射配置
- [ ] 创建HTTP客户端基础类

### Step 2: Android适配层实现（2-3天）
- [ ] 实现ApiAdapter.cpp
- [ ] 实现RestClient.cpp
- [ ] 修改ConnectionsManager集成适配层
- [ ] 添加配置开关

### Step 3: Desktop适配层实现（2-3天）
- [ ] 实现api_adapter.cpp
- [ ] 实现http_client.cpp
- [ ] 修改Session集成适配层
- [ ] 更新CMake配置

### Step 4: 测试验证（1-2天）
- [ ] Android登录测试
- [ ] Android消息收发测试
- [ ] Desktop登录测试
- [ ] Desktop消息收发测试
- [ ] 文件上传下载测试

---

## 🛠️ 开发工具要求

### Android端
- Android Studio
- NDK r21+
- Gradle 7.0+
- JDK 11+

### Desktop端
- Visual Studio 2022 (Windows)
- CMake 3.16+
- Qt 6.2+
- vcpkg（依赖管理）

---

## ⚠️ 关键注意事项

### 1. 保持UI不变
- ❌ 不修改任何UI相关代码
- ✅ 只修改网络层和数据层
- ✅ 用户感觉不到差异

### 2. 数据格式转换
- MTProto使用TL序列化
- REST API使用JSON
- 需要完整的数据转换层

### 3. 认证机制
- Telegram使用phone + code认证
- 需要适配JWT token机制
- 保存token用于后续请求

### 4. 实时更新
- MTProto使用长轮询
- 改用WebSocket连接
- 保持消息实时性

---

## 📊 预计工作量

| 任务 | Android | Desktop | 总计 |
|------|---------|---------|------|
| 架构分析 | 0.5天 | 0.5天 | 1天 |
| 适配层开发 | 2天 | 2天 | 4天 |
| 集成测试 | 1天 | 1天 | 2天 |
| 调试优化 | 1天 | 1天 | 2天 |
| **总计** | **4.5天** | **4.5天** | **9天** |

**实际预留**: 12-15天（考虑意外情况）

---

## 🎯 成功标准

### Android端
✅ 编译成功，生成APK  
✅ 安装后能正常启动  
✅ 能使用手机号登录  
✅ 能收发文本消息  
✅ 能上传/下载文件  
✅ UI和原版一致  

### Desktop端
✅ 编译成功，生成可执行文件  
✅ 启动后能正常运行  
✅ 能使用手机号登录  
✅ 能收发文本消息  
✅ 能上传/下载文件  
✅ UI和原版一致  

---

**状态**: 🟢 准备就绪，即将开始改造！

**下一步**: 开始分析Android网络层架构

