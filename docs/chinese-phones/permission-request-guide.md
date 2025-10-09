# 完整权限申请流程指南

**志航密信** - 确保功能正常运行的权限管理方案

---

## 📋 权限申请原则

### 核心原则

1. ✅ **使用前申请** - 在使用功能之前必须先获得权限
2. ✅ **系统弹窗** - 使用Android系统原生权限弹窗
3. ✅ **清晰说明** - 向用户解释为什么需要该权限
4. ✅ **优雅降级** - 权限被拒绝时提供替代方案或引导
5. ✅ **避免骚扰** - 不重复请求已拒绝的权限

---

## 🎯 权限类型和使用场景

### 1. 相机权限 (CAMERA)

**使用场景：**
- 📹 视频通话
- 📸 拍照发送
- 🎬 录制视频

**申请时机：**
- 用户点击视频通话按钮时
- 用户点击拍照按钮时
- 首次打开相机相关功能时

### 2. 麦克风权限 (RECORD_AUDIO)

**使用场景：**
- 🎤 语音通话
- 🎙️ 语音消息
- 🎵 录音

**申请时机：**
- 用户发起语音/视频通话时
- 用户按住录音按钮时
- 首次使用语音功能时

### 3. 存储权限 (READ/WRITE_EXTERNAL_STORAGE)

**使用场景：**
- 💾 保存文件
- 📁 读取文件
- 🖼️ 访问相册

**申请时机：**
- 用户选择发送文件时
- 用户保存文件时
- 访问相册时

### 4. 位置权限 (ACCESS_FINE_LOCATION)

**使用场景：**
- 📍 分享位置
- 🗺️ 附近的人

**申请时机：**
- 用户点击分享位置时
- 使用位置相关功能时

### 5. 通讯录权限 (READ_CONTACTS)

**使用场景：**
- 👥 同步联系人
- 🔍 查找好友

**申请时机：**
- 用户点击同步联系人时
- 搜索联系人时

### 6. 通知权限 (POST_NOTIFICATIONS - Android 13+)

**使用场景：**
- 📬 接收新消息通知
- 🔔 通话提醒
- ⚡ 系统通知

**申请时机：**
- 应用首次启动时
- 用户开启通知设置时

---

## 💻 Android端实现

### 第一步：在Activity中集成

```java
public class CallActivity extends AppCompatActivity {
    
    private PermissionManager permissionManager;
    
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_call);
        
        permissionManager = PermissionManager.getInstance();
        
        // 设置按钮点击事件
        findViewById(R.id.videoCallBtn).setOnClickListener(v -> {
            startVideoCall();
        });
        
        findViewById(R.id.voiceCallBtn).setOnClickListener(v -> {
            startVoiceCall();
        });
    }
    
    /**
     * 发起视频通话
     */
    private void startVideoCall() {
        // ✅ 使用前检查并请求权限
        permissionManager.requestCallPermissions(this, new PermissionManager.PermissionCallback() {
            @Override
            public void onPermissionGranted() {
                // ✅ 权限已授予，可以开始视频通话
                Toast.makeText(CallActivity.this, "开始视频通话...", Toast.LENGTH_SHORT).show();
                initVideoCall();
            }
            
            @Override
            public void onPermissionDenied(List<String> deniedPermissions) {
                // ❌ 权限被拒绝
                String permissions = String.join(", ", deniedPermissions);
                Toast.makeText(CallActivity.this, 
                    "需要" + permissions + "权限才能进行视频通话", 
                    Toast.LENGTH_LONG).show();
            }
            
            @Override
            public void onPermissionPermanentlyDenied(List<String> permanentlyDeniedPermissions) {
                // ⛔ 权限被永久拒绝
                showPermissionSettingsDialog();
            }
        });
    }
    
    /**
     * 发起语音通话
     */
    private void startVoiceCall() {
        // ✅ 只需要麦克风权限
        permissionManager.requestMicrophone(this, new PermissionManager.PermissionCallback() {
            @Override
            public void onPermissionGranted() {
                Toast.makeText(CallActivity.this, "开始语音通话...", Toast.LENGTH_SHORT).show();
                initVoiceCall();
            }
            
            @Override
            public void onPermissionDenied(List<String> deniedPermissions) {
                Toast.makeText(CallActivity.this, 
                    "需要麦克风权限才能进行语音通话", 
                    Toast.LENGTH_LONG).show();
            }
            
            @Override
            public void onPermissionPermanentlyDenied(List<String> permanentlyDeniedPermissions) {
                showPermissionSettingsDialog();
            }
        });
    }
    
    /**
     * 显示去设置的对话框
     */
    private void showPermissionSettingsDialog() {
        new AlertDialog.Builder(this)
            .setTitle("需要权限")
            .setMessage("请在设置中开启相关权限，以便正常使用此功能。")
            .setPositiveButton("去设置", (dialog, which) -> {
                permissionManager.openAppSettings(this);
            })
            .setNegativeButton("取消", null)
            .show();
    }
    
    /**
     * 处理权限请求结果
     * ⚠️ 重要：必须重写此方法
     */
    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        // 交给PermissionManager处理
        permissionManager.onRequestPermissionsResult(this, requestCode, permissions, grantResults);
    }
    
    @Override
    protected void onDestroy() {
        super.onDestroy();
        // 清理回调，防止内存泄漏
        permissionManager.clearCallbacks();
    }
}
```

### 第二步：屏幕共享权限申请

```java
public class ScreenShareActivity extends AppCompatActivity {
    
    private PermissionManager permissionManager;
    private MediaProjectionManager mediaProjectionManager;
    private static final int REQUEST_CODE_MEDIA_PROJECTION = 2001;
    
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_screen_share);
        
        permissionManager = PermissionManager.getInstance();
        mediaProjectionManager = (MediaProjectionManager) 
            getSystemService(Context.MEDIA_PROJECTION_SERVICE);
        
        findViewById(R.id.startShareBtn).setOnClickListener(v -> {
            startScreenShare();
        });
    }
    
    /**
     * 开始屏幕共享
     */
    private void startScreenShare() {
        // ✅ 第一步：请求基础权限（麦克风、通知）
        permissionManager.requestScreenSharePermissions(this, 
            new PermissionManager.PermissionCallback() {
                @Override
                public void onPermissionGranted() {
                    // ✅ 第二步：请求屏幕录制权限
                    requestMediaProjection();
                }
                
                @Override
                public void onPermissionDenied(List<String> deniedPermissions) {
                    Toast.makeText(ScreenShareActivity.this,
                        "需要相关权限才能进行屏幕共享",
                        Toast.LENGTH_LONG).show();
                }
                
                @Override
                public void onPermissionPermanentlyDenied(List<String> permanentlyDeniedPermissions) {
                    showPermissionSettingsDialog();
                }
            });
    }
    
    /**
     * 请求屏幕录制权限（系统弹窗）
     */
    private void requestMediaProjection() {
        // ✅ 弹出系统屏幕录制权限对话框
        Intent intent = mediaProjectionManager.createScreenCaptureIntent();
        startActivityForResult(intent, REQUEST_CODE_MEDIA_PROJECTION);
    }
    
    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        
        if (requestCode == REQUEST_CODE_MEDIA_PROJECTION) {
            if (resultCode == RESULT_OK) {
                // ✅ 用户允许了屏幕录制权限
                Toast.makeText(this, "开始屏幕共享...", Toast.LENGTH_SHORT).show();
                startScreenCaptureService(data);
            } else {
                // ❌ 用户拒绝了权限
                Toast.makeText(this, 
                    "需要屏幕录制权限才能共享屏幕", 
                    Toast.LENGTH_LONG).show();
                showScreenCaptureRationale();
            }
        }
    }
    
    /**
     * 显示屏幕录制权限说明
     */
    private void showScreenCaptureRationale() {
        new AlertDialog.Builder(this)
            .setTitle("屏幕共享权限")
            .setMessage("屏幕共享需要录制您的屏幕内容。\n\n" +
                       "这样其他参与者才能看到您的屏幕。\n\n" +
                       "您的屏幕内容只会发送给通话中的参与者。")
            .setPositiveButton("重试", (dialog, which) -> {
                requestMediaProjection();
            })
            .setNegativeButton("取消", null)
            .show();
    }
    
    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        permissionManager.onRequestPermissionsResult(this, requestCode, permissions, grantResults);
    }
}
```

---

## 🌐 Web端实现

### 权限申请示例

```javascript
class PermissionRequestManager {
    
    /**
     * 请求相机权限
     */
    async requestCamera() {
        try {
            // ✅ 浏览器会弹出系统权限请求
            const stream = await navigator.mediaDevices.getUserMedia({ 
                video: true 
            });
            
            console.log('✅ 相机权限已授予');
            return { granted: true, stream };
            
        } catch (error) {
            return this.handlePermissionError(error, '相机');
        }
    }
    
    /**
     * 请求麦克风权限
     */
    async requestMicrophone() {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ 
                audio: true 
            });
            
            console.log('✅ 麦克风权限已授予');
            return { granted: true, stream };
            
        } catch (error) {
            return this.handlePermissionError(error, '麦克风');
        }
    }
    
    /**
     * 请求音视频通话权限
     */
    async requestCallPermissions() {
        try {
            // ✅ 同时请求音频和视频
            const stream = await navigator.mediaDevices.getUserMedia({
                video: true,
                audio: true
            });
            
            console.log('✅ 音视频权限已授予');
            return { granted: true, stream };
            
        } catch (error) {
            return this.handlePermissionError(error, '音视频');
        }
    }
    
    /**
     * 请求屏幕共享权限
     */
    async requestScreenShare() {
        try {
            // ✅ 浏览器会弹出屏幕选择对话框
            const stream = await navigator.mediaDevices.getDisplayMedia({
                video: {
                    displaySurface: 'monitor'
                },
                audio: false
            });
            
            console.log('✅ 屏幕共享权限已授予');
            return { granted: true, stream };
            
        } catch (error) {
            return this.handlePermissionError(error, '屏幕共享');
        }
    }
    
    /**
     * 请求通知权限
     */
    async requestNotification() {
        try {
            if (!('Notification' in window)) {
                throw new Error('浏览器不支持通知');
            }
            
            if (Notification.permission === 'granted') {
                return { granted: true };
            }
            
            // ✅ 浏览器会弹出通知权限请求
            const permission = await Notification.requestPermission();
            
            if (permission === 'granted') {
                console.log('✅ 通知权限已授予');
                return { granted: true };
            } else {
                return { granted: false, reason: '用户拒绝了通知权限' };
            }
            
        } catch (error) {
            return this.handlePermissionError(error, '通知');
        }
    }
    
    /**
     * 处理权限错误
     */
    handlePermissionError(error, permissionName) {
        let reason = '';
        
        if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
            reason = `用户拒绝了${permissionName}权限`;
        } else if (error.name === 'NotFoundError') {
            reason = `未找到${permissionName}设备`;
        } else if (error.name === 'NotSupportedError') {
            reason = `浏览器不支持${permissionName}`;
        } else {
            reason = `${permissionName}权限请求失败: ${error.message}`;
        }
        
        console.error('❌', reason);
        
        return {
            granted: false,
            error: error,
            reason: reason
        };
    }
}

// 使用示例
const permissionManager = new PermissionRequestManager();

// 发起视频通话
async function startVideoCall() {
    const result = await permissionManager.requestCallPermissions();
    
    if (result.granted) {
        // ✅ 权限已授予，开始通话
        console.log('开始视频通话');
        setupVideoCall(result.stream);
    } else {
        // ❌ 权限被拒绝
        alert(result.reason);
        showPermissionGuide();
    }
}

// 开始屏幕共享
async function startScreenShare() {
    const result = await permissionManager.requestScreenShare();
    
    if (result.granted) {
        // ✅ 权限已授予，开始共享
        console.log('开始屏幕共享');
        setupScreenShare(result.stream);
    } else {
        // ❌ 权限被拒绝
        alert(result.reason);
    }
}
```

---

## 📱 完整的权限申请流程

### 流程图

```
用户点击功能按钮
     ↓
检查是否已有权限
     ↓
     ├─ 已有 → 直接使用功能
     │
     └─ 没有 → 显示权限说明对话框
                    ↓
               用户点击"授予权限"
                    ↓
           弹出系统权限请求窗口 ✨
                    ↓
          ┌─────────┴─────────┐
          │                   │
        允许                拒绝
          │                   │
     开始使用功能        记录拒绝状态
          │                   │
          │              是否永久拒绝?
          │                   │
          │          ┌────────┴────────┐
          │         是                 否
          │          │                  │
          │     引导去设置        提示原因+重试
          │          │                  │
          └──────────┴──────────────────┘
```

### 关键时机

#### 1. 视频通话

```
用户点击视频通话
    ↓
显示说明："需要相机和麦克风权限才能进行视频通话"
    ↓
用户点击"授予权限"
    ↓
✨ 系统弹窗：允许[应用名]访问您的相机和麦克风吗？
    ↓
用户点击"允许"
    ↓
开始视频通话
```

#### 2. 屏幕共享

```
用户点击屏幕共享
    ↓
检查麦克风权限
    ↓
✨ 系统弹窗：允许[应用名]录制音频吗？
    ↓
用户点击"允许"
    ↓
✨ 系统弹窗：选择要共享的内容（整个屏幕/窗口/标签页）
    ↓
用户选择并点击"共享"
    ↓
开始屏幕共享
```

#### 3. 发送语音消息

```
用户按住录音按钮
    ↓
✨ 系统弹窗：允许[应用名]录制音频吗？
    ↓
用户点击"允许"
    ↓
开始录音
```

---

## ⚠️ 注意事项

### 1. 权限请求时机

✅ **正确**：
- 用户主动点击功能时
- 功能使用前的即时请求
- 有清晰的上下文说明

❌ **错误**：
- 应用启动时批量请求所有权限
- 没有说明就突然请求
- 重复骚扰已拒绝的用户

### 2. 权限说明

✅ **正确**：
```
需要相机权限才能进行视频通话

点击"授予权限"后，系统会弹出权限请求窗口。
```

❌ **错误**：
```
需要权限
```

### 3. 永久拒绝处理

✅ **正确**：
```
您已拒绝相机权限。

如需使用视频通话，请：
1. 点击"去设置"
2. 找到"权限管理"
3. 开启"相机"权限
```

❌ **错误**：
```
权限被拒绝，功能无法使用。
```

### 4. Android版本兼容

```java
// ✅ 检查Android版本
if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
    // Android 13+ 需要额外的通知权限
    permissions.add(Manifest.permission.POST_NOTIFICATIONS);
}

if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
    // Android 11+ 使用新的存储权限
    permissions.add(Manifest.permission.MANAGE_EXTERNAL_STORAGE);
}
```

---

## 📊 权限状态检查

### Android

```java
// 检查单个权限
boolean hasCamera = permissionManager.hasCameraPermission(context);

// 检查多个权限
boolean hasCallPermissions = permissionManager.hasCallPermissions(context);

// 在使用前检查
if (!permissionManager.hasCameraPermission(this)) {
    // 请求权限
    permissionManager.requestCamera(this, callback);
} else {
    // 直接使用
    startCamera();
}
```

### Web

```javascript
// 检查权限状态
async function checkPermissionStatus() {
    try {
        // 检查相机
        const cameraStatus = await navigator.permissions.query({ name: 'camera' });
        console.log('相机权限:', cameraStatus.state); // 'granted', 'denied', 'prompt'
        
        // 检查麦克风
        const micStatus = await navigator.permissions.query({ name: 'microphone' });
        console.log('麦克风权限:', micStatus.state);
        
        // 监听权限变化
        cameraStatus.addEventListener('change', () => {
            console.log('相机权限变化:', cameraStatus.state);
        });
        
    } catch (error) {
        console.error('检查权限失败:', error);
    }
}
```

---

## ✅ 最佳实践总结

### 1. 权限请求的黄金法则

1. **按需请求** - 只在需要时请求，不要一次性请求所有权限
2. **清晰说明** - 告诉用户为什么需要这个权限
3. **系统弹窗** - 使用系统原生权限对话框，不要自定义
4. **尊重选择** - 用户拒绝后不要反复骚扰
5. **提供引导** - 永久拒绝后提供去设置的明确路径

### 2. 用户体验优化

1. **上下文相关** - 在用户执行相关操作时请求权限
2. **分步请求** - 不要一次请求太多权限
3. **友好提示** - 权限被拒后给出友好的提示和解决方案
4. **优雅降级** - 部分权限缺失时提供替代功能

### 3. 代码健壮性

1. **错误处理** - 处理所有可能的权限状态
2. **版本兼容** - 考虑不同Android版本的权限差异
3. **内存管理** - 及时清理权限回调，防止内存泄漏
4. **状态检查** - 使用前始终检查权限状态

---

## 📝 检查清单

### 开发阶段

- [ ] 所有需要权限的功能都在使用前请求
- [ ] 所有权限请求都有清晰的说明
- [ ] 处理了用户拒绝权限的情况
- [ ] 处理了用户永久拒绝权限的情况
- [ ] 提供了去设置页面的入口
- [ ] 兼容不同Android版本
- [ ] 没有内存泄漏

### 测试阶段

- [ ] 首次使用时能正确弹出权限请求
- [ ] 允许权限后功能正常
- [ ] 拒绝权限后提示正确
- [ ] 永久拒绝后引导正确
- [ ] 在设置中授权后返回应用功能正常
- [ ] 不同品牌手机测试通过
- [ ] 不同Android版本测试通过

---

**最后更新**: 2025年10月9日  
**维护者**: 志航密信团队


