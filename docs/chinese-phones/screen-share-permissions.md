# 中国手机品牌屏幕共享权限适配指南

**志航密信** - 中国手机品牌权限兼容方案

---

## 📱 支持的手机品牌

| 品牌 | 市场份额 | 特殊系统 | 优先级 |
|------|---------|---------|--------|
| **小米/Redmi** | ~20% | MIUI | ⭐⭐⭐ |
| **OPPO** | ~18% | ColorOS | ⭐⭐⭐ |
| **vivo** | ~17% | OriginOS | ⭐⭐⭐ |
| **荣耀** | ~15% | MagicOS | ⭐⭐⭐ |
| **华为** | ~10% | HarmonyOS | ⭐⭐⭐ |
| **realme** | ~5% | realme UI | ⭐⭐ |
| **一加** | ~3% | ColorOS | ⭐⭐ |
| **魅族** | ~2% | Flyme | ⭐ |

---

## 🔐 需要的权限

### 1. 屏幕共享权限

```xml
<!-- AndroidManifest.xml -->
<!-- 屏幕录制/共享 -->
<uses-permission android:name="android.permission.CAPTURE_VIDEO_OUTPUT" />
<uses-permission android:name="android.permission.CAPTURE_AUDIO_OUTPUT" />

<!-- 悬浮窗权限（部分品牌需要） -->
<uses-permission android:name="android.permission.SYSTEM_ALERT_WINDOW" />

<!-- 后台运行权限 -->
<uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
<uses-permission android:name="android.permission.FOREGROUND_SERVICE_MEDIA_PROJECTION" />

<!-- 通知权限 -->
<uses-permission android:name="android.permission.POST_NOTIFICATIONS" />
```

### 2. 音视频权限

```xml
<!-- 相机 -->
<uses-permission android:name="android.permission.CAMERA" />
<uses-feature android:name="android.hardware.camera" android:required="false" />
<uses-feature android:name="android.hardware.camera.autofocus" android:required="false" />

<!-- 麦克风 -->
<uses-permission android:name="android.permission.RECORD_AUDIO" />

<!-- 存储（录制文件） -->
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.MANAGE_EXTERNAL_STORAGE" />
```

### 3. 网络权限

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
<uses-permission android:name="android.permission.ACCESS_WIFI_STATE" />
```

---

## 🎯 各品牌适配方案

### 小米/Redmi (MIUI)

#### 特殊问题
1. **自启动限制** - 后台进程容易被杀死
2. **省电模式** - 限制后台网络
3. **权限二次确认** - MIUI安全中心会二次询问
4. **悬浮窗限制** - 默认禁止悬浮窗

#### 解决方案

```java
// MiuiPermissionHelper.java
public class MiuiPermissionHelper {
    
    /**
     * 检测是否是MIUI系统
     */
    public static boolean isMiui() {
        return !TextUtils.isEmpty(getSystemProperty("ro.miui.ui.version.name"));
    }
    
    /**
     * 跳转到MIUI权限设置
     */
    public static void openMiuiPermissionSettings(Context context) {
        Intent intent = new Intent("miui.intent.action.APP_PERM_EDITOR");
        intent.setClassName("com.miui.securitycenter",
            "com.miui.permcenter.permissions.PermissionsEditorActivity");
        intent.putExtra("extra_pkgname", context.getPackageName());
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            // 备用方案：跳转到应用详情
            openAppSettings(context);
        }
    }
    
    /**
     * 跳转到MIUI自启动管理
     */
    public static void openMiuiAutoStartSettings(Context context) {
        Intent intent = new Intent();
        intent.setAction("miui.intent.action.OP_AUTO_START");
        intent.addCategory(Intent.CATEGORY_DEFAULT);
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            openAppSettings(context);
        }
    }
    
    /**
     * 跳转到MIUI省电优化设置
     */
    public static void openMiuiBatterySettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName("com.miui.powerkeeper",
            "com.miui.powerkeeper.ui.HiddenAppsConfigActivity");
        intent.putExtra("package_name", context.getPackageName());
        intent.putExtra("package_label", getAppName(context));
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            // 尝试另一个入口
            try {
                intent = new Intent("miui.intent.action.POWER_HIDE_MODE_APP_LIST");
                context.startActivity(intent);
            } catch (Exception ex) {
                openAppSettings(context);
            }
        }
    }
    
    /**
     * 检查屏幕共享权限
     */
    public static boolean checkScreenCapturePermission(Context context) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            // MIUI需要额外检查悬浮窗权限
            if (!Settings.canDrawOverlays(context)) {
                return false;
            }
        }
        return true;
    }
    
    /**
     * 请求悬浮窗权限（MIUI必需）
     */
    public static void requestOverlayPermission(Activity activity) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            if (!Settings.canDrawOverlays(activity)) {
                Intent intent = new Intent(Settings.ACTION_MANAGE_OVERLAY_PERMISSION);
                intent.setData(Uri.parse("package:" + activity.getPackageName()));
                activity.startActivityForResult(intent, REQUEST_CODE_OVERLAY);
            }
        }
    }
}
```

#### 用户引导文案

```
📱 小米/Redmi手机用户请注意：

为了确保屏幕共享功能正常使用，请进行以下设置：

1️⃣ 开启自启动权限
   设置 → 应用设置 → 应用管理 → 志航密信 → 自启动

2️⃣ 关闭省电优化
   设置 → 应用设置 → 应用管理 → 志航密信 → 省电策略 → 无限制

3️⃣ 允许后台弹出界面
   设置 → 应用设置 → 应用管理 → 志航密信 → 后台弹出界面

4️⃣ 允许显示悬浮窗
   设置 → 应用设置 → 应用管理 → 志航密信 → 显示悬浮窗
```

---

### OPPO (ColorOS)

#### 特殊问题
1. **严格的后台限制** - "纯净后台"机制
2. **关联启动限制** - 限制应用相互唤醒
3. **权限多次确认** - 安全检测中心会多次提示

#### 解决方案

```java
// OppoPermissionHelper.java
public class OppoPermissionHelper {
    
    /**
     * 检测是否是ColorOS
     */
    public static boolean isColorOS() {
        return !TextUtils.isEmpty(getSystemProperty("ro.build.version.opporom"));
    }
    
    /**
     * 跳转到ColorOS权限管理
     */
    public static void openOppoPermissionSettings(Context context) {
        Intent intent = new Intent();
        intent.putExtra("packageName", context.getPackageName());
        
        ComponentName comp = new ComponentName(
            "com.coloros.safecenter",
            "com.coloros.safecenter.permission.PermissionManagerActivity"
        );
        intent.setComponent(comp);
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            // ColorOS 7+
            try {
                intent = new Intent("android.settings.APPLICATION_DETAILS_SETTINGS");
                intent.setData(Uri.parse("package:" + context.getPackageName()));
                context.startActivity(intent);
            } catch (Exception ex) {
                openAppSettings(context);
            }
        }
    }
    
    /**
     * 跳转到关联启动管理
     */
    public static void openOppoStartupSettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName(
            "com.coloros.safecenter",
            "com.coloros.safecenter.startupapp.StartupAppListActivity"
        );
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            openAppSettings(context);
        }
    }
    
    /**
     * 跳转到后台管理
     */
    public static void openOppoBackgroundSettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName(
            "com.coloros.oppoguardelf",
            "com.coloros.powermanager.fuelgaue.PowerConsumptionActivity"
        );
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            openAppSettings(context);
        }
    }
}
```

#### 用户引导文案

```
📱 OPPO手机用户请注意：

为了确保屏幕共享功能正常使用，请进行以下设置：

1️⃣ 允许自启动
   设置 → 应用管理 → 志航密信 → 应用信息 → 自启动

2️⃣ 允许后台运行
   设置 → 电池 → 应用耗电管理 → 志航密信 → 允许后台运行

3️⃣ 允许关联启动
   设置 → 应用管理 → 关联启动 → 志航密信 → 允许

4️⃣ 关闭电池优化
   设置 → 电池 → 更多 → 应用冻结 → 将志航密信移出
```

---

### vivo (OriginOS/Funtouch OS)

#### 特殊问题
1. **iManager严格管控** - 智能管家会自动清理后台
2. **高耗电应用限制** - 自动限制"高耗电"应用
3. **后台高耗电提醒** - 频繁弹出提醒

#### 解决方案

```java
// VivoPermissionHelper.java
public class VivoPermissionHelper {
    
    /**
     * 检测是否是vivo手机
     */
    public static boolean isVivo() {
        return !TextUtils.isEmpty(getSystemProperty("ro.vivo.os.version"));
    }
    
    /**
     * 跳转到vivo权限管理
     */
    public static void openVivoPermissionSettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName(
            "com.vivo.permissionmanager",
            "com.vivo.permissionmanager.activity.SoftPermissionDetailActivity"
        );
        intent.putExtra("packagename", context.getPackageName());
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            // 备用方案
            try {
                intent = new Intent("com.vivo.permissionmanager");
                context.startActivity(intent);
            } catch (Exception ex) {
                openAppSettings(context);
            }
        }
    }
    
    /**
     * 跳转到后台高耗电管理
     */
    public static void openVivoHighPowerSettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName(
            "com.vivo.abe",
            "com.vivo.applicationbehaviorengine.ui.ExcessivePowerManagerActivity"
        );
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            openAppSettings(context);
        }
    }
    
    /**
     * 跳转到自启动管理
     */
    public static void openVivoAutoStartSettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName(
            "com.iqoo.secure",
            "com.iqoo.secure.ui.phoneoptimize.AddWhiteListActivity"
        );
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            openAppSettings(context);
        }
    }
}
```

#### 用户引导文案

```
📱 vivo手机用户请注意：

为了确保屏幕共享功能正常使用，请进行以下设置：

1️⃣ 加入后台高耗电白名单
   i管家 → 应用管理 → 自启动管理 → 志航密信 → 允许

2️⃣ 关闭省电模式
   设置 → 电池 → 后台高耗电 → 志航密信 → 允许后台高耗电

3️⃣ 允许后台弹出界面
   设置 → 应用与权限 → 权限管理 → 悬浮窗 → 志航密信 → 允许

4️⃣ 锁定应用（防止被清理）
   近期任务 → 下拉应用卡片 → 锁定
```

---

### 华为/荣耀 (HarmonyOS/MagicOS)

#### 特殊问题
1. **HMS核心服务** - 需要HMS支持
2. **严格的通知管理** - 通知权限默认关闭
3. **精准后台管控** - 智能识别应用行为

#### 解决方案

```java
// HuaweiPermissionHelper.java
public class HuaweiPermissionHelper {
    
    /**
     * 检测是否是华为/荣耀
     */
    public static boolean isHuawei() {
        String manufacturer = Build.MANUFACTURER.toLowerCase();
        return manufacturer.contains("huawei") || manufacturer.contains("honor");
    }
    
    /**
     * 跳转到华为权限管理
     */
    public static void openHuaweiPermissionSettings(Context context) {
        Intent intent = new Intent();
        
        // HarmonyOS 2.0+
        intent.setClassName(
            "com.huawei.systemmanager",
            "com.huawei.permissionmanager.ui.MainActivity"
        );
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            // EMUI 旧版本
            try {
                intent.setClassName(
                    "com.huawei.systemmanager",
                    "com.huawei.permissionmanager.ui.SingleAppActivity"
                );
                intent.putExtra("packageName", context.getPackageName());
                context.startActivity(intent);
            } catch (Exception ex) {
                openAppSettings(context);
            }
        }
    }
    
    /**
     * 跳转到华为启动管理
     */
    public static void openHuaweiStartupSettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName(
            "com.huawei.systemmanager",
            "com.huawei.systemmanager.startupmgr.ui.StartupNormalAppListActivity"
        );
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            openAppSettings(context);
        }
    }
    
    /**
     * 跳转到华为电池优化
     */
    public static void openHuaweiBatterySettings(Context context) {
        Intent intent = new Intent();
        intent.setClassName(
            "com.huawei.systemmanager",
            "com.huawei.systemmanager.optimize.process.ProtectActivity"
        );
        
        try {
            context.startActivity(intent);
        } catch (Exception e) {
            openAppSettings(context);
        }
    }
    
    /**
     * 检查HMS可用性
     */
    public static boolean isHmsAvailable(Context context) {
        try {
            Class<?> hmsApiAvailability = Class.forName(
                "com.huawei.hms.api.HuaweiApiAvailability"
            );
            Method getInstance = hmsApiAvailability.getMethod("getInstance");
            Object instance = getInstance.invoke(null);
            Method isAvailable = hmsApiAvailability.getMethod(
                "isHuaweiMobileServicesAvailable", Context.class
            );
            int result = (int) isAvailable.invoke(instance, context);
            return result == 0; // ConnectionResult.SUCCESS
        } catch (Exception e) {
            return false;
        }
    }
}
```

#### 用户引导文案

```
📱 华为/荣耀手机用户请注意：

为了确保屏幕共享功能正常使用，请进行以下设置：

1️⃣ 允许手动管理
   设置 → 应用和服务 → 应用启动管理 → 志航密信 → 手动管理
   ✅ 自动启动 ✅ 关联启动 ✅ 后台活动

2️⃣ 加入锁屏清理白名单
   设置 → 电池 → 应用启动管理 → 志航密信 → 允许

3️⃣ 关闭省电模式
   设置 → 电池 → 更多电池设置 → 休眠时始终保持网络连接

4️⃣ 允许通知
   设置 → 通知 → 志航密信 → 允许通知
```

---

## 🎯 统一权限请求方案

### 1. 权限检查工具类

```java
// ChinesePhonePermissionManager.java
public class ChinesePhonePermissionManager {
    
    private Context context;
    private String brand;
    
    public ChinesePhonePermissionManager(Context context) {
        this.context = context;
        this.brand = detectBrand();
    }
    
    /**
     * 检测手机品牌
     */
    private String detectBrand() {
        String manufacturer = Build.MANUFACTURER.toLowerCase();
        
        if (manufacturer.contains("xiaomi") || manufacturer.contains("redmi")) {
            return "MIUI";
        } else if (manufacturer.contains("oppo") || manufacturer.contains("realme")) {
            return "ColorOS";
        } else if (manufacturer.contains("vivo") || manufacturer.contains("iqoo")) {
            return "OriginOS";
        } else if (manufacturer.contains("huawei") || manufacturer.contains("honor")) {
            return "HarmonyOS";
        } else if (manufacturer.contains("oneplus")) {
            return "OxygenOS";
        } else if (manufacturer.contains("meizu")) {
            return "Flyme";
        }
        
        return "Android";
    }
    
    /**
     * 检查所有必需权限
     */
    public PermissionCheckResult checkAllPermissions() {
        PermissionCheckResult result = new PermissionCheckResult();
        
        // 1. 屏幕录制权限（MediaProjection）
        result.screenCapture = checkScreenCapturePermission();
        
        // 2. 悬浮窗权限（部分品牌需要）
        if (needsOverlayPermission()) {
            result.overlay = checkOverlayPermission();
        }
        
        // 3. 通知权限
        result.notification = checkNotificationPermission();
        
        // 4. 后台运行权限（品牌特有）
        result.background = checkBackgroundPermission();
        
        // 5. 自启动权限（品牌特有）
        result.autoStart = checkAutoStartPermission();
        
        // 6. 电池优化豁免
        result.batteryOptimization = checkBatteryOptimization();
        
        return result;
    }
    
    /**
     * 是否需要悬浮窗权限
     */
    private boolean needsOverlayPermission() {
        return brand.equals("MIUI") || 
               brand.equals("ColorOS") || 
               brand.equals("OriginOS");
    }
    
    /**
     * 请求屏幕录制权限
     */
    public void requestScreenCapturePermission(Activity activity) {
        MediaProjectionManager projectionManager = 
            (MediaProjectionManager) context.getSystemService(
                Context.MEDIA_PROJECTION_SERVICE
            );
        
        Intent intent = projectionManager.createScreenCaptureIntent();
        activity.startActivityForResult(intent, REQUEST_CODE_SCREEN_CAPTURE);
    }
    
    /**
     * 打开权限设置页面
     */
    public void openPermissionSettings() {
        switch (brand) {
            case "MIUI":
                MiuiPermissionHelper.openMiuiPermissionSettings(context);
                break;
            case "ColorOS":
                OppoPermissionHelper.openOppoPermissionSettings(context);
                break;
            case "OriginOS":
                VivoPermissionHelper.openVivoPermissionSettings(context);
                break;
            case "HarmonyOS":
                HuaweiPermissionHelper.openHuaweiPermissionSettings(context);
                break;
            default:
                openAppSettings(context);
                break;
        }
    }
    
    /**
     * 显示品牌特定的引导对话框
     */
    public void showBrandSpecificGuide() {
        String message = getBrandSpecificGuideMessage();
        
        new AlertDialog.Builder(context)
            .setTitle("权限设置指南")
            .setMessage(message)
            .setPositiveButton("去设置", (dialog, which) -> {
                openPermissionSettings();
            })
            .setNegativeButton("稍后", null)
            .show();
    }
    
    /**
     * 获取品牌特定的引导文案
     */
    private String getBrandSpecificGuideMessage() {
        switch (brand) {
            case "MIUI":
                return "小米/Redmi手机需要：\n\n" +
                       "1. 开启自启动权限\n" +
                       "2. 关闭省电优化\n" +
                       "3. 允许后台弹出界面\n" +
                       "4. 允许显示悬浮窗\n\n" +
                       "点击「去设置」开始配置";
                       
            case "ColorOS":
                return "OPPO手机需要：\n\n" +
                       "1. 允许自启动\n" +
                       "2. 允许后台运行\n" +
                       "3. 允许关联启动\n" +
                       "4. 关闭电池优化\n\n" +
                       "点击「去设置」开始配置";
                       
            case "OriginOS":
                return "vivo手机需要：\n\n" +
                       "1. 加入后台高耗电白名单\n" +
                       "2. 关闭省电模式\n" +
                       "3. 允许后台弹出界面\n" +
                       "4. 锁定应用防止清理\n\n" +
                       "点击「去设置」开始配置";
                       
            case "HarmonyOS":
                return "华为/荣耀手机需要：\n\n" +
                       "1. 允许手动管理应用启动\n" +
                       "2. 加入锁屏清理白名单\n" +
                       "3. 关闭省电模式\n" +
                       "4. 允许通知\n\n" +
                       "点击「去设置」开始配置";
                       
            default:
                return "请授予以下权限：\n\n" +
                       "1. 屏幕录制权限\n" +
                       "2. 后台运行权限\n" +
                       "3. 通知权限\n\n" +
                       "点击「去设置」开始配置";
        }
    }
}

// 权限检查结果
public class PermissionCheckResult {
    public boolean screenCapture = false;
    public boolean overlay = false;
    public boolean notification = false;
    public boolean background = false;
    public boolean autoStart = false;
    public boolean batteryOptimization = false;
    
    public boolean isAllGranted() {
        return screenCapture && 
               (overlay || !needsOverlay()) && 
               notification && 
               background && 
               autoStart && 
               batteryOptimization;
    }
    
    public List<String> getMissingPermissions() {
        List<String> missing = new ArrayList<>();
        
        if (!screenCapture) missing.add("屏幕录制");
        if (!overlay && needsOverlay()) missing.add("悬浮窗");
        if (!notification) missing.add("通知");
        if (!background) missing.add("后台运行");
        if (!autoStart) missing.add("自启动");
        if (!batteryOptimization) missing.add("电池优化");
        
        return missing;
    }
}
```

---

## 📱 使用示例

### Activity中使用

```java
public class ScreenShareActivity extends AppCompatActivity {
    
    private ChinesePhonePermissionManager permissionManager;
    private static final int REQUEST_CODE_SCREEN_CAPTURE = 1001;
    
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_screen_share);
        
        permissionManager = new ChinesePhonePermissionManager(this);
        
        // 检查权限
        checkAndRequestPermissions();
    }
    
    private void checkAndRequestPermissions() {
        PermissionCheckResult result = permissionManager.checkAllPermissions();
        
        if (!result.isAllGranted()) {
            // 显示权限引导
            showPermissionGuideDialog(result);
        } else {
            // 可以开始屏幕共享
            startScreenShare();
        }
    }
    
    private void showPermissionGuideDialog(PermissionCheckResult result) {
        List<String> missing = result.getMissingPermissions();
        String message = "需要以下权限才能使用屏幕共享：\n\n" + 
                        String.join("\n", missing) + "\n\n" +
                        "点击「查看设置指南」了解详细步骤";
        
        new AlertDialog.Builder(this)
            .setTitle("权限设置")
            .setMessage(message)
            .setPositiveButton("查看设置指南", (dialog, which) -> {
                permissionManager.showBrandSpecificGuide();
            })
            .setNegativeButton("稍后", null)
            .setCancelable(false)
            .show();
    }
    
    private void startScreenShare() {
        // 请求屏幕录制权限
        permissionManager.requestScreenCapturePermission(this);
    }
    
    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        
        if (requestCode == REQUEST_CODE_SCREEN_CAPTURE) {
            if (resultCode == RESULT_OK) {
                // 权限授予成功，开始共享
                startScreenShareService(data);
            } else {
                // 权限被拒绝
                Toast.makeText(this, "需要屏幕录制权限才能共享屏幕", 
                    Toast.LENGTH_LONG).show();
            }
        }
    }
    
    private void startScreenShareService(Intent data) {
        Intent serviceIntent = new Intent(this, ScreenShareService.class);
        serviceIntent.putExtra("data", data);
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(serviceIntent);
        } else {
            startService(serviceIntent);
        }
    }
}
```

---

## 🔔 前台服务实现

```java
// ScreenShareService.java
public class ScreenShareService extends Service {
    
    private static final int NOTIFICATION_ID = 1001;
    private static final String CHANNEL_ID = "screen_share_channel";
    
    @Override
    public void onCreate() {
        super.onCreate();
        createNotificationChannel();
    }
    
    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        // 启动前台服务（避免被杀死）
        Notification notification = createNotification();
        startForeground(NOTIFICATION_ID, notification);
        
        // 开始屏幕共享逻辑
        startScreenCapture(intent);
        
        return START_STICKY;
    }
    
    private void createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID,
                "屏幕共享",
                NotificationManager.IMPORTANCE_LOW
            );
            channel.setDescription("屏幕共享正在运行");
            channel.setShowBadge(false);
            
            NotificationManager manager = getSystemService(NotificationManager.class);
            manager.createNotificationChannel(channel);
        }
    }
    
    private Notification createNotification() {
        Intent intent = new Intent(this, ScreenShareActivity.class);
        PendingIntent pendingIntent = PendingIntent.getActivity(
            this, 0, intent, 
            PendingIntent.FLAG_IMMUTABLE
        );
        
        NotificationCompat.Builder builder = new NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("志航密信")
            .setContentText("屏幕共享进行中...")
            .setSmallIcon(R.drawable.ic_screen_share)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setPriority(NotificationCompat.PRIORITY_LOW);
        
        return builder.build();
    }
    
    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}
```

---

## ✅ 测试清单

### 小米/Redmi
- [ ] 自启动权限已开启
- [ ] 省电优化已关闭
- [ ] 后台弹出界面已允许
- [ ] 悬浮窗权限已允许
- [ ] 屏幕共享功能正常
- [ ] 锁屏后共享不中断

### OPPO
- [ ] 自启动权限已开启
- [ ] 后台运行已允许
- [ ] 关联启动已允许
- [ ] 电池优化已关闭
- [ ] 屏幕共享功能正常
- [ ] 后台不被清理

### vivo
- [ ] 后台高耗电白名单已加入
- [ ] 省电模式已关闭
- [ ] 后台弹出界面已允许
- [ ] 应用已锁定
- [ ] 屏幕共享功能正常
- [ ] i管家不清理应用

### 华为/荣耀
- [ ] 手动管理已开启
- [ ] 锁屏清理白名单已加入
- [ ] 省电模式已关闭
- [ ] 通知权限已允许
- [ ] 屏幕共享功能正常
- [ ] HMS服务正常

---

## 📝 注意事项

### 1. 用户体验优化

- ✅ 首次启动时显示清晰的权限引导
- ✅ 提供一键跳转到设置的功能
- ✅ 使用品牌特定的术语和截图
- ✅ 提供视频教程链接
- ✅ 定期检查权限状态并提醒

### 2. 代码健壮性

- ✅ 所有Intent跳转都要try-catch
- ✅ 提供多个备用跳转方案
- ✅ 检测系统版本差异
- ✅ 记录权限请求失败日志

### 3. 持续维护

- ✅ 关注各品牌系统更新
- ✅ 及时适配新版本变化
- ✅ 收集用户反馈
- ✅ 更新引导文案

---

## 🔗 相关资源

- [小米开发者文档](https://dev.mi.com/console/doc/detail?pId=2292)
- [OPPO开放平台](https://open.oppomobile.com/)
- [vivo开发者平台](https://dev.vivo.com.cn/)
- [华为开发者联盟](https://developer.huawei.com/)

---

**最后更新**: 2025年10月9日  
**维护者**: 志航密信团队



