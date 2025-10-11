# Telegram源码获取指南

**重要说明**：Telegram源码体积巨大（20000+文件），已从Git仓库中排除，仅在本地保留。

---

## 📥 开发者需要手动下载Telegram源码

### Android端

**步骤1：下载官方源码**
```bash
# 方式1：Git Clone（推荐）
git clone https://github.com/DrKLO/Telegram.git telegram-android

# 方式2：下载ZIP
# 访问：https://github.com/DrKLO/Telegram
# 点击 Code → Download ZIP
# 解压到 telegram-android/ 目录
```

**步骤2：验证目录结构**
```
telegram-android/
├── TMessagesProj/
│   ├── jni/
│   │   └── tgnet/
│   │       └── ConnectionsManager.cpp  # 我们要修改的核心文件
│   └── src/
│       └── main/
│           └── java/
│               └── org/telegram/
│                   ├── messenger/
│                   ├── tgnet/
│                   │   └── TLRPC.java   # TL对象定义
│                   └── ui/
└── build.gradle
```

---

### Desktop端

**步骤1：下载官方源码**
```bash
# 方式1：Git Clone（推荐）
git clone https://github.com/telegramdesktop/tdesktop.git telegram-desktop

# 方式2：下载ZIP
# 访问：https://github.com/telegramdesktop/tdesktop
# 点击 Code → Download ZIP
# 解压到 telegram-desktop/ 目录
```

**步骤2：验证目录结构**
```
telegram-desktop/
├── Telegram/
│   └── SourceFiles/
│       ├── mtproto/
│       │   ├── session.cpp           # 我们要修改的核心文件
│       │   └── connection.cpp
│       └── api/
│           └── api_request.cpp
└── CMakeLists.txt
```

---

## ⚠️ 重要提醒

1. **版本选择**：
   - 建议使用稳定版（不要用dev分支）
   - 当前测试版本：最新Release版本

2. **体积说明**：
   - Android源码：约500MB
   - Desktop源码：约1GB
   - **请确保有足够磁盘空间**

3. **不要提交到Git**：
   - Telegram源码已添加到`.gitignore`
   - 修改后也不要提交（我们的适配层代码单独管理）

4. **更新策略**：
   - 固定使用当前版本，不要频繁更新
   - Telegram更新时，需要重新验证适配层兼容性

---

## 📚 参考资料

- [Telegram Android官方仓库](https://github.com/DrKLO/Telegram)
- [Telegram Desktop官方仓库](https://github.com/telegramdesktop/tdesktop)
- [MTProto协议文档](https://core.telegram.org/mtproto)

---

**文档版本**：v1.0
**创建时间**：2025-10-11

