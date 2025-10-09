# 屏幕共享功能增强版

**志航密信 v1.6.0 Enhanced** - 全面增强的屏幕共享功能

---

## 🎯 增强功能概述

在原有基础功能上，新增了以下高级特性：

### ✨ 核心增强

| 功能分类 | 具体功能 | 状态 |
|---------|---------|------|
| **权限控制** | 基于角色的权限管理 | ✅ 完成 |
| | 质量等级限制 | ✅ 完成 |
| | 时长限制 | ✅ 完成 |
| **数据管理** | 会话历史记录 | ✅ 完成 |
| | 质量变更记录 | ✅ 完成 |
| | 参与者管理 | ✅ 完成 |
| | 统计信息 | ✅ 完成 |
| **智能优化** | 网络自适应质量 | ✅ 完成 |
| | 实时质量监控 | ✅ 完成 |
| | 自动质量调整 | ✅ 完成 |
| **录制功能** | 屏幕录制 | ✅ 完成 |
| | 录制管理 | ✅ 完成 |
| | 录制权限控制 | ✅ 完成 |
| **前端增强** | 错误处理 | ✅ 完成 |
| | 自动重连 | ✅ 完成 |
| | 性能监控 | ✅ 完成 |

---

## 📦 新增文件清单

### 后端文件

```
im-backend/
├── internal/
│   ├── model/
│   │   └── screen_share.go                    ✨ 数据模型
│   ├── service/
│   │   └── screen_share_enhanced_service.go   ✨ 增强服务
│   └── controller/
│       └── screen_share_enhanced_controller.go ✨ 增强控制器
```

### 前端文件

```
examples/
└── screen-share-enhanced.js                    ✨ 增强管理器
```

### 文档文件

```
SCREEN_SHARE_ENHANCED.md                        ✨ 本文档
```

---

## 🗄️ 数据库设计

### 1. 会话记录表 (screen_share_sessions)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| call_id | string | 通话ID |
| sharer_id | uint | 共享者ID |
| sharer_name | string | 共享者名称 |
| start_time | timestamp | 开始时间 |
| end_time | timestamp | 结束时间 |
| duration | int64 | 时长（秒） |
| quality | string | 质量等级 |
| with_audio | bool | 是否包含音频 |
| initial_quality | string | 初始质量 |
| quality_changes | int | 质量调整次数 |
| participant_count | int | 参与者数量 |
| end_reason | string | 结束原因 |
| status | string | 状态 |

### 2. 质量变更记录表 (screen_share_quality_changes)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| session_id | uint | 会话ID |
| from_quality | string | 原质量 |
| to_quality | string | 新质量 |
| change_time | timestamp | 变更时间 |
| change_reason | string | 变更原因 |
| network_speed | float64 | 当时网速 (Kbps) |
| cpu_usage | float64 | 当时CPU使用率 |

### 3. 参与者表 (screen_share_participants)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| session_id | uint | 会话ID |
| user_id | uint | 用户ID |
| user_name | string | 用户名 |
| join_time | timestamp | 加入时间 |
| leave_time | timestamp | 离开时间 |
| view_duration | int64 | 观看时长（秒） |

### 4. 统计表 (screen_share_statistics)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| user_id | uint | 用户ID（唯一） |
| total_sessions | int64 | 总共享次数 |
| total_duration | int64 | 总共享时长（秒） |
| average_duration | float64 | 平均时长（秒） |
| total_participants | int64 | 总参与人次 |
| high_quality_count | int64 | 高清次数 |
| medium_quality_count | int64 | 标准次数 |
| low_quality_count | int64 | 流畅次数 |
| with_audio_count | int64 | 包含音频次数 |
| last_share_time | timestamp | 最后共享时间 |

### 5. 录制表 (screen_share_recordings)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint | 主键 |
| session_id | uint | 会话ID |
| recorder_id | uint | 录制者ID |
| file_name | string | 文件名 |
| file_path | string | 文件路径 |
| file_size | int64 | 文件大小（字节） |
| duration | int64 | 录制时长（秒） |
| format | string | 格式 (webm/mp4) |
| quality | string | 质量 |
| start_time | timestamp | 开始时间 |
| end_time | timestamp | 结束时间 |
| status | string | 状态 |
| download_count | int | 下载次数 |

---

## 🔐 权限管理

### 权限配置

| 角色 | 可共享 | 可录制 | 最大时长 | 最大质量 | 需要审批 |
|------|-------|-------|---------|---------|---------|
| **user** | ✅ | ❌ | 1小时 | medium | ❌ |
| **admin** | ✅ | ✅ | 2小时 | high | ❌ |
| **super_admin** | ✅ | ✅ | 无限制 | high | ❌ |

### 权限检查示例

```go
// 检查屏幕共享权限
func (s *ScreenShareEnhancedService) CheckSharePermission(userID uint, quality string) error {
    var user model.User
    if err := s.db.First(&user, userID).Error; err != nil {
        return fmt.Errorf("用户不存在: %w", err)
    }

    permission, exists := defaultPermissions[user.Role]
    if !exists {
        permission = defaultPermissions["user"]
    }

    if !permission.CanShare {
        return errors.New("您没有屏幕共享权限")
    }

    if !s.isQualityAllowed(quality, permission.MaxQuality) {
        return fmt.Errorf("您的最高质量限制为: %s", permission.MaxQuality)
    }

    return nil
}
```

---

## 📡 新增 API 端点

### 会话历史

```http
GET /api/screen-share/history?page=1&page_size=20
Authorization: Bearer {token}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "sessions": [...],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

### 用户统计

```http
GET /api/screen-share/statistics
Authorization: Bearer {token}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "user_id": 1,
    "total_sessions": 50,
    "total_duration": 18000,
    "average_duration": 360,
    "total_participants": 250,
    "high_quality_count": 20,
    "medium_quality_count": 25,
    "low_quality_count": 5,
    "with_audio_count": 10,
    "last_share_time": "2025-10-09T10:30:00Z"
  }
}
```

### 会话详情

```http
GET /api/screen-share/sessions/:session_id
Authorization: Bearer {token}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "session": {...},
    "quality_changes": [...],
    "participants": [...],
    "recordings": [...]
  }
}
```

### 开始录制

```http
POST /api/screen-share/:call_id/recording/start
Authorization: Bearer {token}
Content-Type: application/json

{
  "format": "webm",
  "quality": "medium"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "session_id": 456,
    "format": "webm",
    "quality": "medium",
    "start_time": "2025-10-09T10:30:00Z",
    "status": "recording"
  },
  "message": "录制已开始"
}
```

### 结束录制

```http
POST /api/screen-share/recordings/:recording_id/end
Authorization: Bearer {token}
Content-Type: application/json

{
  "file_path": "/recordings/123.webm",
  "file_size": 1048576
}
```

### 获取录制列表

```http
GET /api/screen-share/sessions/:session_id/recordings
Authorization: Bearer {token}
```

### 权限检查

```http
GET /api/screen-share/check-permission?quality=high
Authorization: Bearer {token}
```

### 记录质量变更

```http
POST /api/screen-share/:call_id/quality-change
Authorization: Bearer {token}
Content-Type: application/json

{
  "from_quality": "medium",
  "to_quality": "low",
  "reason": "auto_network",
  "network_speed": 800.5,
  "cpu_usage": 75.2
}
```

### 导出统计数据

```http
GET /api/screen-share/export?start_time=2025-01-01T00:00:00Z&end_time=2025-12-31T23:59:59Z
Authorization: Bearer {token}
```

---

## 💻 前端增强功能

### 1. 网络自适应

```javascript
const manager = new ScreenShareEnhancedManager('call_123');

// 启用自动质量调整
await manager.startScreenShare({
    quality: 'medium',
    autoAdjustQuality: true
});

// 监听质量变化
manager.onQualityChange = ({ from, to, reason }) => {
    console.log(`质量已变更: ${from} -> ${to} (${reason})`);
};

// 监听网络变化
manager.onNetworkChange = (info) => {
    console.log('网络状态:', info);
};
```

### 2. 错误处理

```javascript
// 设置错误回调
manager.onError = (error) => {
    console.error('屏幕共享错误:', error);
    alert(`错误: ${error.message}`);
};

try {
    await manager.startScreenShare();
} catch (error) {
    // 错误已被捕获
}
```

### 3. 录制功能

```javascript
// 开始录制
await manager.startRecording({
    format: 'webm',
    quality: 'high'
});

// 停止录制并下载
const blob = await manager.stopRecording();

// 下载文件
const url = URL.createObjectURL(blob);
const a = document.createElement('a');
a.href = url;
a.download = `screen_share_${Date.now()}.webm`;
a.click();
```

### 4. 历史记录

```javascript
// 获取历史记录
const history = await manager.getHistory(1, 20);

console.log('共享历史:', history.sessions);
console.log('总计:', history.total);
```

### 5. 统计信息

```javascript
// 获取统计
const stats = await manager.getStatistics();

console.log('总共享次数:', stats.total_sessions);
console.log('总时长:', stats.total_duration);
console.log('平均时长:', stats.average_duration);
```

---

## 🎨 使用示例

### 完整示例：智能屏幕共享

```javascript
class SmartScreenShareDemo {
    constructor(callId) {
        this.manager = new ScreenShareEnhancedManager(callId);
        this.setupCallbacks();
    }

    setupCallbacks() {
        // 错误处理
        this.manager.onError = (error) => {
            this.showError(error.message);
        };

        // 质量变化
        this.manager.onQualityChange = ({ from, to, reason }) => {
            this.showNotification(`质量已调整: ${this.getQualityLabel(to)}`);
            this.updateQualityIndicator(to);
        };

        // 网络变化
        this.manager.onNetworkChange = (info) => {
            this.updateNetworkIndicator(info);
        };
    }

    async start() {
        try {
            // 1. 检查权限
            const hasPermission = await this.manager.checkPermission('high');
            if (!hasPermission) {
                throw new Error('权限不足');
            }

            // 2. 开始共享
            const stream = await this.manager.startScreenShare({
                quality: 'medium',
                withAudio: false,
                autoAdjustQuality: true
            });

            // 3. 显示视频
            document.getElementById('video').srcObject = stream;

            // 4. 显示控制UI
            this.showControls();

        } catch (error) {
            this.showError(error.message);
        }
    }

    async stop() {
        await this.manager.stopScreenShare('manual');
        this.hideControls();
    }

    async startRecording() {
        try {
            await this.manager.startRecording({
                format: 'webm',
                quality: 'high'
            });
            this.showRecordingIndicator();
        } catch (error) {
            this.showError(error.message);
        }
    }

    async stopRecording() {
        try {
            const blob = await this.manager.stopRecording();
            this.downloadRecording(blob);
            this.hideRecordingIndicator();
        } catch (error) {
            this.showError(error.message);
        }
    }

    async showStatistics() {
        const stats = await this.manager.getStatistics();
        
        const html = `
            <div class="stats">
                <h3>您的屏幕共享统计</h3>
                <p>总共享次数: ${stats.total_sessions}</p>
                <p>总时长: ${this.formatDuration(stats.total_duration)}</p>
                <p>平均时长: ${this.formatDuration(stats.average_duration)}</p>
                <p>总参与人次: ${stats.total_participants}</p>
            </div>
        `;
        
        document.getElementById('stats-panel').innerHTML = html;
    }

    async showHistory() {
        const history = await this.manager.getHistory(1, 10);
        
        const html = history.sessions.map(session => `
            <div class="session-item">
                <span>${session.start_time}</span>
                <span>${this.formatDuration(session.duration)}</span>
                <span>${this.getQualityLabel(session.quality)}</span>
            </div>
        `).join('');
        
        document.getElementById('history-list').innerHTML = html;
    }

    getQualityLabel(quality) {
        const labels = {
            'high': '高清',
            'medium': '标准',
            'low': '流畅'
        };
        return labels[quality] || quality;
    }

    formatDuration(seconds) {
        const minutes = Math.floor(seconds / 60);
        const secs = Math.floor(seconds % 60);
        return `${minutes}分${secs}秒`;
    }

    updateQualityIndicator(quality) {
        const indicator = document.getElementById('quality-indicator');
        indicator.textContent = this.getQualityLabel(quality);
        indicator.className = `quality-${quality}`;
    }

    updateNetworkIndicator(info) {
        const indicator = document.getElementById('network-indicator');
        indicator.textContent = `${Math.round(info.estimatedSpeed)} Kbps`;
    }

    showNotification(message) {
        // 显示通知
        console.log('📢', message);
    }

    showError(message) {
        alert('错误: ' + message);
    }

    showControls() {
        document.getElementById('controls').style.display = 'block';
    }

    hideControls() {
        document.getElementById('controls').style.display = 'none';
    }

    showRecordingIndicator() {
        document.getElementById('recording-indicator').style.display = 'block';
    }

    hideRecordingIndicator() {
        document.getElementById('recording-indicator').style.display = 'none';
    }

    downloadRecording(blob) {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `screen_share_${Date.now()}.webm`;
        a.click();
        URL.revokeObjectURL(url);
    }

    cleanup() {
        this.manager.destroy();
    }
}

// 使用
const demo = new SmartScreenShareDemo('call_123456');
demo.start();
```

---

## 🔧 配置说明

### 后端配置

在数据库迁移中添加新表：

```sql
-- 运行数据库迁移
-- 将自动创建所有新表
```

### 权限配置

可以在 `screen_share_enhanced_service.go` 中修改默认权限：

```go
var defaultPermissions = map[string]ScreenSharePermission{
    "user": {
        CanShare:         true,
        CanRecord:        false,
        MaxDuration:      3600, // 修改最大时长
        MaxQuality:       "medium", // 修改最大质量
        RequiresApproval: false,
    },
    // ... 更多角色
}
```

---

## 📊 性能优化

### 网络自适应策略

| 网速范围 | CPU使用率 | 推荐质量 |
|---------|----------|---------|
| > 3 Mbps | < 70% | high |
| > 1 Mbps | < 80% | medium |
| < 1 Mbps | 任意 | low |

### 质量调整频率

- **检查间隔**: 5秒
- **调整间隔**: 10秒
- **最大采样数**: 10次

---

## 🐛 故障排查

### 常见问题

**Q: 自动质量调整不工作？**

A: 检查：
1. 是否启用了 `autoAdjustQuality`
2. 浏览器是否支持 Network Information API
3. 查看控制台是否有错误

**Q: 录制功能不可用？**

A: 检查：
1. 用户角色是否有录制权限
2. 浏览器是否支持 MediaRecorder API
3. 是否正在共享屏幕

**Q: 历史记录为空？**

A: 确保：
1. 数据库迁移已完成
2. 共享会话已正确结束
3. 用户ID正确

---

## 📈 监控和分析

### 关键指标

1. **共享质量分布**: 高清/标准/流畅的使用占比
2. **平均共享时长**: 用户平均每次共享多久
3. **质量调整频率**: 自动调整的次数和原因
4. **网络质量**: 平均网速和波动情况
5. **录制使用率**: 录制功能的使用频率

### 统计查询示例

```sql
-- 查看质量分布
SELECT quality, COUNT(*) as count 
FROM screen_share_sessions 
GROUP BY quality;

-- 查看平均时长
SELECT AVG(duration) as avg_duration 
FROM screen_share_sessions 
WHERE status = 'ended';

-- 查看质量调整原因分布
SELECT change_reason, COUNT(*) as count 
FROM screen_share_quality_changes 
GROUP BY change_reason;
```

---

## 🚀 未来计划

### v1.7.0 计划

- [ ] AI 画质增强
- [ ] 实时字幕
- [ ] 画笔标注
- [ ] 多人同时共享
- [ ] 移动端支持

### v1.8.0 计划

- [ ] 虚拟背景
- [ ] 区域共享（只共享特定窗口）
- [ ] 水印功能
- [ ] 智能降噪
- [ ] 带宽预测

---

## 📝 更新日志

### v1.6.0 Enhanced - 2025-10-09

#### 新增功能 ✨
- 基于角色的权限管理
- 会话历史记录
- 质量变更追踪
- 参与者管理
- 用户统计信息
- 屏幕共享录制
- 网络自适应质量
- 实时性能监控
- 自动质量调整
- 增强的错误处理

#### 数据库 🗄️
- 5个新数据表
- 完整的关联关系
- 索引优化

#### API 📡
- 10+个新端点
- 完整的CRUD操作
- RESTful设计

#### 前端 💻
- 增强的管理器类
- 网络监控
- 自动重连
- 错误处理

---

## 📞 技术支持

如有问题或建议，请联系：

- 📧 邮箱：support@zhihang-messenger.com
- 📝 文档：https://docs.zhihang-messenger.com
- 💬 社区：https://community.zhihang-messenger.com
- 🐛 问题反馈：https://github.com/zhihang-messenger/issues

---

**志航密信团队**  
2025年10月9日


