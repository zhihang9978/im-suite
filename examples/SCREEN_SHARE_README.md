# 屏幕共享功能使用指南

**志航密信 v1.6.0** - WebRTC 屏幕共享功能

---

## 📋 目录

1. [功能概述](#功能概述)
2. [技术架构](#技术架构)
3. [快速开始](#快速开始)
4. [API 文档](#api-文档)
5. [前端集成](#前端集成)
6. [使用示例](#使用示例)
7. [常见问题](#常见问题)
8. [浏览器兼容性](#浏览器兼容性)

---

## 🎯 功能概述

屏幕共享功能允许用户在音视频通话中实时共享自己的屏幕内容。支持以下特性：

### ✨ 核心功能

- ✅ **多质量级别**：高清（1080p）、标准（720p）、流畅（480p）
- ✅ **系统音频共享**：可选择是否包含系统音频
- ✅ **动态质量调整**：通话中随时调整共享质量
- ✅ **单人共享限制**：同一通话中同时只允许一人共享
- ✅ **自动断开检测**：用户停止共享时自动通知
- ✅ **状态查询**：实时查询当前共享状态
- ✅ **独立连接**：屏幕共享使用独立的 PeerConnection

### 📊 质量配置

| 质量级别 | 分辨率 | 帧率 | 适用场景 |
|---------|--------|------|---------|
| **High** | 1920×1080 | 30fps | 演示设计稿、视频播放 |
| **Medium** | 1280×720 | 24fps | 文档展示、PPT演示 |
| **Low** | 640×480 | 15fps | 网络较差、低配设备 |

---

## 🏗️ 技术架构

### 后端架构

```
im-backend/
├── internal/
│   ├── service/
│   │   └── webrtc_service.go          # WebRTC服务（含屏幕共享逻辑）
│   └── controller/
│       └── webrtc_controller.go       # WebRTC控制器（API端点）
└── main.go                            # 路由配置
```

### 前端架构

```
examples/
├── screen-share-example.js            # 屏幕共享管理器
├── screen-share-demo.html             # 演示页面
└── SCREEN_SHARE_README.md            # 使用文档
```

### 数据模型

#### CallSession（通话会话）
```go
type CallSession struct {
    ID              string
    Type            string              // audio, video, screen_share
    ScreenSharing   *ScreenShareInfo    // 屏幕共享信息
    Peers           map[uint]*PeerConnection
}
```

#### ScreenShareInfo（屏幕共享信息）
```go
type ScreenShareInfo struct {
    SharerID    uint       // 共享者用户ID
    SharerName  string     // 共享者名称
    IsActive    bool       // 是否正在共享
    StartTime   time.Time  // 开始时间
    Quality     string     // 质量: high, medium, low
    WithAudio   bool       // 是否包含音频
}
```

#### PeerConnection（对等连接）
```go
type PeerConnection struct {
    PC               *webrtc.PeerConnection  // 主连接（音视频）
    ScreenSharePC    *webrtc.PeerConnection  // 屏幕共享连接
    IsScreenSharing  bool                    // 是否正在共享
}
```

---

## 🚀 快速开始

### 1. 启动后端服务

```bash
# 进入后端目录
cd im-backend

# 运行服务
go run main.go
```

后端将在 `http://localhost:8080` 启动。

### 2. 打开演示页面

```bash
# 在浏览器中打开
examples/screen-share-demo.html
```

或者使用本地服务器：

```bash
# 使用Python
python -m http.server 8000

# 然后访问
http://localhost:8000/examples/screen-share-demo.html
```

### 3. 测试屏幕共享

1. 点击 **"开始共享屏幕"** 按钮
2. 浏览器会提示选择共享内容（整个屏幕/窗口/标签页）
3. 选择后开始共享
4. 可随时调整质量或停止共享

---

## 📚 API 文档

### 后端 REST API

#### 1. 开始屏幕共享

```http
POST /api/calls/:call_id/screen-share/start
Content-Type: application/json
Authorization: Bearer {token}

{
  "user_name": "张三",
  "quality": "medium",
  "with_audio": false
}
```

**响应：**
```json
{
  "success": true,
  "message": "屏幕共享已开始"
}
```

**错误响应：**
```json
{
  "error": "已有用户正在共享屏幕，共享者: 李四"
}
```

---

#### 2. 停止屏幕共享

```http
POST /api/calls/:call_id/screen-share/stop
Authorization: Bearer {token}
```

**响应：**
```json
{
  "success": true,
  "message": "屏幕共享已停止"
}
```

---

#### 3. 查询屏幕共享状态

```http
GET /api/calls/:call_id/screen-share/status
Authorization: Bearer {token}
```

**响应（正在共享）：**
```json
{
  "success": true,
  "data": {
    "sharer_id": 123,
    "sharer_name": "张三",
    "is_active": true,
    "start_time": "2025-10-09T10:30:00Z",
    "quality": "medium",
    "with_audio": false
  }
}
```

**响应（未共享）：**
```json
{
  "success": true,
  "data": {
    "is_active": false,
    "message": "当前没有屏幕共享"
  }
}
```

---

#### 4. 更改屏幕共享质量

```http
POST /api/calls/:call_id/screen-share/quality
Content-Type: application/json
Authorization: Bearer {token}

{
  "quality": "high"
}
```

**响应：**
```json
{
  "success": true,
  "message": "屏幕共享质量已更改为: high"
}
```

---

### WebRTC 信令

#### 屏幕共享 Offer
```javascript
{
  "type": "screen_share_offer",
  "call_id": "call_123456",
  "user_id": 123,
  "payload": {
    "type": "offer",
    "sdp": "..."
  }
}
```

#### 屏幕共享 Answer
```javascript
{
  "type": "screen_share_answer",
  "call_id": "call_123456",
  "user_id": 456,
  "payload": {
    "type": "answer",
    "sdp": "..."
  }
}
```

#### 屏幕共享 ICE Candidate
```javascript
{
  "type": "screen_share_ice_candidate",
  "call_id": "call_123456",
  "user_id": 123,
  "payload": {
    "candidate": "...",
    "sdpMid": "...",
    "sdpMLineIndex": 0
  }
}
```

---

## 💻 前端集成

### 基础使用

#### 1. 引入脚本

```html
<script src="screen-share-example.js"></script>
```

#### 2. 初始化管理器

```javascript
const callId = 'call_123456';
const screenShare = new ScreenShareManager(callId);
```

#### 3. 开始共享

```javascript
try {
  const stream = await screenShare.startScreenShare({
    quality: 'medium',
    withAudio: false
  });

  // 显示在video元素中
  const videoElement = document.getElementById('screenShareVideo');
  videoElement.srcObject = stream;
  videoElement.play();

  console.log('✅ 屏幕共享已开始');
} catch (error) {
  console.error('屏幕共享失败:', error.message);
  alert('屏幕共享失败: ' + error.message);
}
```

#### 4. 停止共享

```javascript
try {
  await screenShare.stopScreenShare();

  // 清除video元素
  const videoElement = document.getElementById('screenShareVideo');
  videoElement.srcObject = null;

  console.log('✅ 屏幕共享已停止');
} catch (error) {
  console.error('停止失败:', error);
}
```

---

### 高级功能

#### 动态调整质量

```javascript
// 根据网络情况调整质量
try {
  const newStream = await screenShare.changeQuality('low');
  
  // 更新video元素
  videoElement.srcObject = newStream;
  
  console.log('✅ 质量已降低为 low（网络不佳）');
} catch (error) {
  console.error('更改质量失败:', error);
}
```

#### 监听用户停止共享

```javascript
const stream = await screenShare.startScreenShare({ quality: 'medium' });

// 监听轨道结束事件
stream.getVideoTracks()[0].addEventListener('ended', () => {
  console.log('🛑 用户主动停止了屏幕共享');
  handleScreenShareStopped();
});
```

#### 查询共享状态

```javascript
const status = await screenShare.getStatus();

if (status && status.is_active) {
  console.log('📺 正在共享屏幕');
  console.log('  共享者:', status.sharer_name);
  console.log('  质量:', status.quality);
  console.log('  音频:', status.with_audio ? '是' : '否');
} else {
  console.log('⏸️ 当前没有屏幕共享');
}
```

---

## 📝 使用示例

### 示例 1：基础屏幕共享

```javascript
// 创建管理器
const screenShare = new ScreenShareManager('call_123456');

// 开始共享（中等质量，不含音频）
const stream = await screenShare.startScreenShare({
  quality: 'medium',
  withAudio: false
});

// 显示在页面上
document.getElementById('video').srcObject = stream;
```

### 示例 2：带音频的高质量共享

```javascript
// 高质量 + 系统音频
const stream = await screenShare.startScreenShare({
  quality: 'high',
  withAudio: true
});

document.getElementById('video').srcObject = stream;
```

### 示例 3：根据网络状况自适应

```javascript
// 开始时使用中等质量
await screenShare.startScreenShare({ quality: 'medium' });

// 监听网络状况
navigator.connection.addEventListener('change', async () => {
  const effectiveType = navigator.connection.effectiveType;
  
  let quality;
  if (effectiveType === '4g') {
    quality = 'high';
  } else if (effectiveType === '3g') {
    quality = 'medium';
  } else {
    quality = 'low';
  }
  
  // 调整质量
  await screenShare.changeQuality(quality);
  console.log(`网络: ${effectiveType}, 质量: ${quality}`);
});
```

### 示例 4：完整的UI集成

```javascript
class CallPage {
  constructor(callId) {
    this.screenShare = new ScreenShareManager(callId);
    this.setupUI();
  }

  setupUI() {
    // 开始按钮
    document.getElementById('startBtn').onclick = async () => {
      try {
        const quality = document.getElementById('quality').value;
        const withAudio = document.getElementById('audio').checked;
        
        const stream = await this.screenShare.startScreenShare({
          quality,
          withAudio
        });
        
        document.getElementById('video').srcObject = stream;
        this.showStopButton();
        this.showStatus('正在共享屏幕...');
      } catch (error) {
        alert('屏幕共享失败: ' + error.message);
      }
    };

    // 停止按钮
    document.getElementById('stopBtn').onclick = async () => {
      await this.screenShare.stopScreenShare();
      document.getElementById('video').srcObject = null;
      this.showStartButton();
      this.hideStatus();
    };
  }

  showStartButton() {
    document.getElementById('startBtn').style.display = 'block';
    document.getElementById('stopBtn').style.display = 'none';
  }

  showStopButton() {
    document.getElementById('startBtn').style.display = 'none';
    document.getElementById('stopBtn').style.display = 'block';
  }

  showStatus(message) {
    document.getElementById('status').textContent = message;
  }

  hideStatus() {
    document.getElementById('status').textContent = '';
  }
}

// 使用
const callPage = new CallPage('call_123456');
```

---

## ❓ 常见问题

### Q1: 为什么无法开始屏幕共享？

**A:** 可能的原因：

1. **浏览器不支持**：检查浏览器版本和兼容性
2. **权限被拒绝**：用户需要授予屏幕共享权限
3. **HTTPS 要求**：屏幕共享需要在 HTTPS 环境下运行（localhost 除外）
4. **已有人在共享**：同一通话中同时只允许一人共享

### Q2: 如何共享系统音频？

**A:** 设置 `withAudio: true`：

```javascript
const stream = await screenShare.startScreenShare({
  quality: 'medium',
  withAudio: true  // 包含系统音频
});
```

**注意**：系统音频共享需要浏览器支持，部分浏览器可能不支持。

### Q3: 如何优化屏幕共享性能？

**A:** 优化建议：

1. **选择合适的质量**：
   - 网络良好：`high`
   - 一般网络：`medium`
   - 网络较差：`low`

2. **动态调整质量**：
   ```javascript
   // 根据网络状况调整
   if (networkSpeed < 1000) {  // < 1Mbps
     await screenShare.changeQuality('low');
   }
   ```

3. **关闭不必要的音频**：
   ```javascript
   withAudio: false  // 如果不需要系统音频
   ```

### Q4: 屏幕共享突然中断怎么办？

**A:** 可能的原因和解决方案：

1. **用户主动停止**：用户点击了浏览器的"停止共享"按钮
   ```javascript
   // 监听停止事件
   stream.getVideoTracks()[0].addEventListener('ended', () => {
     console.log('用户停止了共享');
     handleStopShare();
   });
   ```

2. **网络中断**：检查网络连接，重新建立连接

3. **权限被撤销**：用户可能在系统设置中撤销了权限

### Q5: 如何实现观众端接收屏幕共享？

**A:** 观众端需要：

1. **监听屏幕共享事件**（通过WebSocket）
2. **创建接收端 PeerConnection**
3. **处理远程流**

```javascript
// 监听屏幕共享开始事件
websocket.on('screen_share_started', async (data) => {
  console.log(`${data.sharer_name} 开始共享屏幕`);
  
  // 创建接收端连接
  const pc = new RTCPeerConnection(config);
  
  // 处理远程流
  pc.ontrack = (event) => {
    const remoteVideo = document.getElementById('remoteScreenShare');
    remoteVideo.srcObject = event.streams[0];
    remoteVideo.play();
  };
  
  // 设置远程描述
  await pc.setRemoteDescription(data.offer);
  
  // 创建answer
  const answer = await pc.createAnswer();
  await pc.setLocalDescription(answer);
  
  // 发送answer给共享者
  websocket.send({
    type: 'screen_share_answer',
    answer: answer
  });
});
```

### Q6: 如何限制只有特定用户才能共享？

**A:** 在后端添加权限检查：

```go
// 在 StartScreenShare 方法中添加
func (s *WebRTCService) StartScreenShare(...) error {
    // 检查用户权限
    if !s.checkScreenSharePermission(userID) {
        return errors.New("您没有屏幕共享权限")
    }
    
    // ... 其余代码
}
```

---

## 🌐 浏览器兼容性

### 支持的浏览器

| 浏览器 | 最低版本 | 屏幕共享 | 系统音频 |
|--------|---------|---------|---------|
| **Chrome** | 72+ | ✅ | ✅ |
| **Edge** | 79+ | ✅ | ✅ |
| **Firefox** | 66+ | ✅ | ✅ |
| **Safari** | 13+ | ✅ | ❌ |
| **Opera** | 60+ | ✅ | ✅ |

### 功能支持说明

- ✅ **屏幕共享**：所有现代浏览器都支持
- ✅ **系统音频**：Chrome、Edge、Firefox、Opera 支持
- ❌ **Safari**：不支持系统音频共享

### 检测浏览器支持

```javascript
function checkScreenShareSupport() {
  // 检查 getDisplayMedia API
  if (!navigator.mediaDevices || !navigator.mediaDevices.getDisplayMedia) {
    alert('您的浏览器不支持屏幕共享功能');
    return false;
  }
  
  return true;
}

// 使用
if (checkScreenShareSupport()) {
  // 可以使用屏幕共享
  await screenShare.startScreenShare();
} else {
  // 提示用户升级浏览器
  alert('请使用 Chrome 72+ / Firefox 66+ / Safari 13+ 等浏览器');
}
```

---

## 📞 技术支持

如有问题，请联系：

- 📧 邮箱：support@zhihang-messenger.com
- 📝 文档：https://docs.zhihang-messenger.com
- 💬 社区：https://community.zhihang-messenger.com

---

## 📄 许可证

版权所有 © 2025 志航密信 (ZhiHang Messenger)

保留所有权利。



