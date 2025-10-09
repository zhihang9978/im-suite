# 屏幕共享增强功能 - 快速开始

## 🚀 立即开始使用

### 第一步：了解文件

```bash
# 📚 文档（3个）
SCREEN_SHARE_FEATURE.md              # 基础功能介绍
SCREEN_SHARE_ENHANCED.md             # 增强功能详解
SCREEN_SHARE_ENHANCEMENT_SUMMARY.md  # 完成报告（推荐先看）

# 💻 后端代码（5个）
im-backend/internal/model/screen_share.go                     # 数据模型
im-backend/internal/service/screen_share_enhanced_service.go  # 增强服务
im-backend/internal/controller/screen_share_enhanced_controller.go  # 控制器
im-backend/internal/controller/webrtc_controller.go           # WebRTC控制器
im-backend/internal/service/webrtc_service.go                 # WebRTC服务(已修改)

# 🌐 前端代码（5个）
examples/screen-share-example.js       # 基础管理器
examples/screen-share-enhanced.js      # 增强管理器⭐
examples/screen-share-demo.html        # 演示页面
examples/SCREEN_SHARE_README.md        # 使用文档
examples/QUICK_TEST.md                 # 测试指南
```

---

## ⚡ 3分钟测试

### 1. 启动后端（1分钟）

```bash
cd im-backend
go run main.go
```

等待看到：`服务启动成功 http://localhost:8080`

### 2. 打开演示页面（30秒）

在浏览器打开：`examples/screen-share-demo.html`

### 3. 测试功能（1分30秒）

1. 点击"开始共享屏幕"
2. 选择要共享的屏幕/窗口
3. 观察视频显示和日志
4. 尝试切换质量
5. 点击"停止共享"

✅ **看到视频和日志 = 测试成功！**

---

## 📊 核心功能一览

### 1. 权限控制 ✅

```
角色          可共享  可录制  最大时长  最大质量
user          ✅      ❌      1小时    medium
admin         ✅      ✅      2小时    high
super_admin   ✅      ✅      无限     high
```

### 2. 智能质量调整 ✅

```javascript
// 自动根据网络调整
网速 > 3Mbps  →  High (1080p)
网速 > 1Mbps  →  Medium (720p)
网速 < 1Mbps  →  Low (480p)
```

### 3. 数据追溯 ✅

- 📝 会话历史记录
- 📊 质量变更追踪
- 👥 参与者管理
- 📈 统计信息

### 4. 录制功能 ✅

- 🎥 WebM/MP4格式
- 💾 本地+服务器双存储
- 📦 文件管理
- 🔐 权限控制

---

## 🎯 使用示例

### 基础使用

```javascript
const manager = new ScreenShareEnhancedManager('call_123');

// 开始共享（自动质量调整）
await manager.startScreenShare({
    quality: 'medium',
    autoAdjustQuality: true
});

// 停止共享
await manager.stopScreenShare();
```

### 带录制

```javascript
// 开始共享
await manager.startScreenShare();

// 开始录制
await manager.startRecording({
    format: 'webm',
    quality: 'high'
});

// 停止录制并下载
const blob = await manager.stopRecording();
downloadBlob(blob, 'recording.webm');
```

### 查看统计

```javascript
// 获取历史
const history = await manager.getHistory(1, 20);

// 获取统计
const stats = await manager.getStatistics();
console.log('总共享:', stats.total_sessions);
console.log('总时长:', stats.total_duration);
```

---

## 🔗 API速查

### 基础API

| API | 方法 | 说明 |
|-----|------|------|
| `/api/calls/:call_id/screen-share/start` | POST | 开始共享 |
| `/api/calls/:call_id/screen-share/stop` | POST | 停止共享 |
| `/api/calls/:call_id/screen-share/status` | GET | 查询状态 |
| `/api/calls/:call_id/screen-share/quality` | POST | 调整质量 |

### 增强API

| API | 方法 | 说明 |
|-----|------|------|
| `/api/screen-share/history` | GET | 历史记录 |
| `/api/screen-share/statistics` | GET | 统计信息 |
| `/api/screen-share/:call_id/recording/start` | POST | 开始录制 |
| `/api/screen-share/check-permission` | GET | 检查权限 |

**完整API列表**: 见 `SCREEN_SHARE_ENHANCED.md`

---

## 📖 详细文档

### 新手入门
👉 `examples/QUICK_TEST.md` - 5分钟快速测试

### 功能介绍  
👉 `SCREEN_SHARE_FEATURE.md` - 基础功能  
👉 `SCREEN_SHARE_ENHANCED.md` - 增强功能

### 开发指南
👉 `examples/SCREEN_SHARE_README.md` - 使用文档  
👉 `SCREEN_SHARE_ENHANCEMENT_SUMMARY.md` - 完成报告

---

## ❓ 常见问题

**Q: 权限不足怎么办？**  
A: 检查用户角色，user只能用medium质量，admin可用high

**Q: 自动质量调整不工作？**  
A: 确保 `autoAdjustQuality: true` 并检查Network API支持

**Q: 录制失败？**  
A: 检查用户角色（需要admin或super_admin）

**Q: 找不到历史记录？**  
A: 确保数据库表已创建，检查会话是否正常结束

---

## 🎉 开始使用

1. **阅读**: `SCREEN_SHARE_ENHANCEMENT_SUMMARY.md` (5分钟)
2. **测试**: 按本文档测试 (3分钟)
3. **集成**: 替换基础版为增强版
4. **部署**: 运行数据库迁移，部署服务

**祝您使用愉快！** 🚀


