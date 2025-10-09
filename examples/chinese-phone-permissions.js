/**
 * 中国手机品牌权限适配 - Web端
 * 志航密信 v1.6.0
 * 
 * 针对中国手机品牌的浏览器权限适配
 */

class ChinesePhonePermissionAdapter {
    constructor() {
        this.brand = this.detectBrand();
        this.browserInfo = this.detectBrowser();
    }

    /**
     * 检测手机品牌
     */
    detectBrand() {
        const ua = navigator.userAgent.toLowerCase();
        
        if (ua.includes('miui') || ua.includes('xiaomi') || ua.includes('redmi')) {
            return {
                name: 'MIUI',
                displayName: '小米/Redmi',
                needsSpecialGuide: true
            };
        }
        
        if (ua.includes('oppo') || ua.includes('realme')) {
            return {
                name: 'ColorOS',
                displayName: 'OPPO',
                needsSpecialGuide: true
            };
        }
        
        if (ua.includes('vivo') || ua.includes('iqoo')) {
            return {
                name: 'OriginOS',
                displayName: 'vivo',
                needsSpecialGuide: true
            };
        }
        
        if (ua.includes('huawei') || ua.includes('honor') || ua.includes('harmonyos')) {
            return {
                name: 'HarmonyOS',
                displayName: '华为/荣耀',
                needsSpecialGuide: true
            };
        }
        
        if (ua.includes('oneplus')) {
            return {
                name: 'OxygenOS',
                displayName: '一加',
                needsSpecialGuide: true
            };
        }
        
        if (ua.includes('meizu')) {
            return {
                name: 'Flyme',
                displayName: '魅族',
                needsSpecialGuide: true
            };
        }
        
        return {
            name: 'Android',
            displayName: 'Android',
            needsSpecialGuide: false
        };
    }

    /**
     * 检测浏览器
     */
    detectBrowser() {
        const ua = navigator.userAgent.toLowerCase();
        
        if (ua.includes('miuibrowser')) {
            return { name: 'MIUI浏览器', isSystemBrowser: true };
        }
        
        if (ua.includes('heytapbrowser') || ua.includes('oppobrowser')) {
            return { name: 'OPPO浏览器', isSystemBrowser: true };
        }
        
        if (ua.includes('vivobrowser')) {
            return { name: 'vivo浏览器', isSystemBrowser: true };
        }
        
        if (ua.includes('huaweibrowser')) {
            return { name: '华为浏览器', isSystemBrowser: true };
        }
        
        if (ua.includes('chrome')) {
            return { name: 'Chrome', isSystemBrowser: false };
        }
        
        if (ua.includes('firefox')) {
            return { name: 'Firefox', isSystemBrowser: false };
        }
        
        if (ua.includes('edge')) {
            return { name: 'Edge', isSystemBrowser: false };
        }
        
        return { name: '未知浏览器', isSystemBrowser: false };
    }

    /**
     * 检查屏幕共享支持
     */
    async checkScreenShareSupport() {
        const result = {
            supported: false,
            reason: '',
            needsGuide: false
        };

        // 检查API支持
        if (!navigator.mediaDevices || !navigator.mediaDevices.getDisplayMedia) {
            result.reason = '您的浏览器不支持屏幕共享功能';
            result.needsGuide = false;
            return result;
        }

        // 系统浏览器可能有限制
        if (this.browserInfo.isSystemBrowser) {
            result.supported = true;
            result.needsGuide = true;
            result.reason = `检测到您正在使用${this.brand.displayName}系统浏览器，可能需要额外权限设置`;
            return result;
        }

        result.supported = true;
        result.needsGuide = this.brand.needsSpecialGuide;
        
        return result;
    }

    /**
     * 请求屏幕共享权限（带重试）
     */
    async requestScreenShare(options = {}) {
        const {
            quality = 'medium',
            withAudio = false,
            maxRetries = 2,
            onRetry = null
        } = options;

        let lastError = null;

        for (let i = 0; i <= maxRetries; i++) {
            try {
                if (i > 0 && onRetry) {
                    onRetry(i);
                    await this.delay(1000); // 等待1秒再重试
                }

                const stream = await this.getDisplayMedia({
                    video: this.getVideoConstraints(quality),
                    audio: withAudio
                });

                // 成功获取
                return {
                    success: true,
                    stream: stream,
                    message: '屏幕共享已启动'
                };

            } catch (error) {
                lastError = error;
                console.error(`屏幕共享请求失败 (尝试 ${i + 1}/${maxRetries + 1}):`, error);

                // 分析错误原因
                const errorInfo = this.analyzeError(error);
                
                // 如果是用户拒绝，不重试
                if (errorInfo.isUserDenied) {
                    break;
                }
            }
        }

        // 所有重试都失败
        const errorInfo = this.analyzeError(lastError);
        return {
            success: false,
            error: lastError,
            message: errorInfo.message,
            needsGuide: errorInfo.needsGuide,
            guideMessage: this.getErrorGuideMessage(errorInfo)
        };
    }

    /**
     * 获取显示媒体流
     */
    async getDisplayMedia(constraints) {
        // 某些手机浏览器可能需要特殊处理
        if (this.brand.name === 'MIUI') {
            // MIUI浏览器可能需要额外的权限提示
            return await navigator.mediaDevices.getDisplayMedia(constraints);
        }

        return await navigator.mediaDevices.getDisplayMedia(constraints);
    }

    /**
     * 获取视频约束
     */
    getVideoConstraints(quality) {
        const constraints = {
            displaySurface: 'monitor',
        };

        switch (quality) {
            case 'high':
                constraints.width = { ideal: 1920 };
                constraints.height = { ideal: 1080 };
                constraints.frameRate = { ideal: 30 };
                break;
            case 'medium':
                constraints.width = { ideal: 1280 };
                constraints.height = { ideal: 720 };
                constraints.frameRate = { ideal: 24 };
                break;
            case 'low':
                constraints.width = { ideal: 640 };
                constraints.height = { ideal: 480 };
                constraints.frameRate = { ideal: 15 };
                break;
        }

        return constraints;
    }

    /**
     * 分析错误
     */
    analyzeError(error) {
        const errorName = error.name || '';
        const errorMessage = error.message || '';

        // 用户拒绝
        if (errorName === 'NotAllowedError' || errorName === 'PermissionDeniedError') {
            return {
                type: 'user_denied',
                isUserDenied: true,
                needsGuide: false,
                message: '您拒绝了屏幕共享权限'
            };
        }

        // 没有找到屏幕
        if (errorName === 'NotFoundError') {
            return {
                type: 'not_found',
                isUserDenied: false,
                needsGuide: true,
                message: '未找到可共享的屏幕'
            };
        }

        // 安全错误（可能是权限问题）
        if (errorName === 'SecurityError') {
            return {
                type: 'security',
                isUserDenied: false,
                needsGuide: true,
                message: '安全限制，可能需要在设置中允许权限'
            };
        }

        // 不支持
        if (errorName === 'NotSupportedError') {
            return {
                type: 'not_supported',
                isUserDenied: false,
                needsGuide: true,
                message: '您的设备或浏览器不支持屏幕共享'
            };
        }

        // 其他错误
        return {
            type: 'unknown',
            isUserDenied: false,
            needsGuide: true,
            message: `屏幕共享失败: ${errorMessage}`
        };
    }

    /**
     * 获取错误引导信息
     */
    getErrorGuideMessage(errorInfo) {
        if (!errorInfo.needsGuide) {
            return '';
        }

        let guide = `\n\n📱 ${this.brand.displayName}手机用户请注意：\n\n`;

        switch (this.brand.name) {
            case 'MIUI':
                guide += this.getMiuiGuide(errorInfo.type);
                break;
            case 'ColorOS':
                guide += this.getOppoGuide(errorInfo.type);
                break;
            case 'OriginOS':
                guide += this.getVivoGuide(errorInfo.type);
                break;
            case 'HarmonyOS':
                guide += this.getHuaweiGuide(errorInfo.type);
                break;
            default:
                guide += this.getGenericGuide(errorInfo.type);
        }

        return guide;
    }

    /**
     * 小米引导
     */
    getMiuiGuide(errorType) {
        return `1️⃣ 打开浏览器设置\n` +
               `2️⃣ 找到「网站设置」或「权限管理」\n` +
               `3️⃣ 允许「屏幕录制」或「媒体」权限\n` +
               `4️⃣ 如使用MIUI浏览器，可能需要在系统设置中允许「显示悬浮窗」\n\n` +
               `💡 建议使用 Chrome 浏览器以获得最佳体验`;
    }

    /**
     * OPPO引导
     */
    getOppoGuide(errorType) {
        return `1️⃣ 打开浏览器设置\n` +
               `2️⃣ 找到「网站管理」\n` +
               `3️⃣ 允许「麦克风」和「相机」权限\n` +
               `4️⃣ 在系统设置中允许浏览器「后台运行」\n\n` +
               `💡 建议使用 Chrome 浏览器以获得最佳体验`;
    }

    /**
     * vivo引导
     */
    getVivoGuide(errorType) {
        return `1️⃣ 打开浏览器设置 → 隐私设置\n` +
               `2️⃣ 允许「媒体访问」权限\n` +
               `3️⃣ 在系统设置中将浏览器加入「后台高耗电」白名单\n` +
               `4️⃣ 关闭「省电模式」\n\n` +
               `💡 建议使用 Chrome 浏览器以获得最佳体验`;
    }

    /**
     * 华为引导
     */
    getHuaweiGuide(errorType) {
        return `1️⃣ 打开浏览器设置 → 网站设置\n` +
               `2️⃣ 允许「媒体权限」\n` +
               `3️⃣ 在系统设置中允许浏览器「自动管理」启动\n` +
               `4️⃣ 将浏览器加入「电池优化」白名单\n\n` +
               `💡 建议使用 Chrome 浏览器以获得最佳体验`;
    }

    /**
     * 通用引导
     */
    getGenericGuide(errorType) {
        return `1️⃣ 打开浏览器设置\n` +
               `2️⃣ 找到「网站权限」或「隐私设置」\n` +
               `3️⃣ 允许「屏幕录制」或「媒体」权限\n` +
               `4️⃣ 刷新页面重试\n\n` +
               `💡 建议使用 Chrome、Firefox 或 Edge 浏览器`;
    }

    /**
     * 显示引导对话框
     */
    showGuideDialog(message) {
        const dialog = document.createElement('div');
        dialog.className = 'chinese-phone-permission-dialog';
        dialog.innerHTML = `
            <div class="dialog-overlay"></div>
            <div class="dialog-content">
                <div class="dialog-header">
                    <h3>📱 权限设置指南</h3>
                    <button class="dialog-close">×</button>
                </div>
                <div class="dialog-body">
                    <p class="brand-info">
                        <strong>检测到：</strong>${this.brand.displayName} · ${this.browserInfo.name}
                    </p>
                    <pre class="guide-message">${message}</pre>
                </div>
                <div class="dialog-footer">
                    <button class="btn btn-secondary" onclick="this.closest('.chinese-phone-permission-dialog').remove()">
                        我知道了
                    </button>
                    <button class="btn btn-primary" onclick="window.location.reload()">
                        刷新页面重试
                    </button>
                </div>
            </div>
        `;

        // 添加样式
        if (!document.getElementById('chinese-phone-permission-styles')) {
            const styles = document.createElement('style');
            styles.id = 'chinese-phone-permission-styles';
            styles.textContent = this.getDialogStyles();
            document.head.appendChild(styles);
        }

        document.body.appendChild(dialog);

        // 关闭按钮
        dialog.querySelector('.dialog-close').onclick = () => {
            dialog.remove();
        };

        // 点击遮罩关闭
        dialog.querySelector('.dialog-overlay').onclick = () => {
            dialog.remove();
        };
    }

    /**
     * 对话框样式
     */
    getDialogStyles() {
        return `
            .chinese-phone-permission-dialog {
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                z-index: 10000;
            }
            
            .dialog-overlay {
                position: absolute;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: rgba(0, 0, 0, 0.5);
            }
            
            .dialog-content {
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                background: white;
                border-radius: 12px;
                max-width: 90%;
                max-height: 80vh;
                overflow: auto;
                box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
            }
            
            .dialog-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 20px;
                border-bottom: 1px solid #eee;
            }
            
            .dialog-header h3 {
                margin: 0;
                font-size: 18px;
                color: #333;
            }
            
            .dialog-close {
                background: none;
                border: none;
                font-size: 28px;
                color: #999;
                cursor: pointer;
                padding: 0;
                width: 32px;
                height: 32px;
                line-height: 1;
            }
            
            .dialog-body {
                padding: 20px;
            }
            
            .brand-info {
                background: #f0f7ff;
                padding: 12px;
                border-radius: 6px;
                margin-bottom: 15px;
                color: #0066cc;
            }
            
            .guide-message {
                background: #f5f5f5;
                padding: 15px;
                border-radius: 6px;
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
                font-size: 14px;
                line-height: 1.6;
                color: #333;
                white-space: pre-wrap;
                margin: 0;
            }
            
            .dialog-footer {
                display: flex;
                gap: 10px;
                padding: 20px;
                border-top: 1px solid #eee;
                justify-content: flex-end;
            }
            
            .btn {
                padding: 10px 20px;
                border: none;
                border-radius: 6px;
                font-size: 14px;
                cursor: pointer;
                transition: all 0.3s;
            }
            
            .btn-secondary {
                background: #f0f0f0;
                color: #666;
            }
            
            .btn-secondary:hover {
                background: #e0e0e0;
            }
            
            .btn-primary {
                background: #0066cc;
                color: white;
            }
            
            .btn-primary:hover {
                background: #0052a3;
            }
        `;
    }

    /**
     * 延迟函数
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * 获取品牌信息（供外部调用）
     */
    getBrandInfo() {
        return {
            brand: this.brand,
            browser: this.browserInfo
        };
    }
}

// ============================================
// 使用示例
// ============================================

/**
 * 示例：完整的屏幕共享流程
 */
async function startScreenShareWithGuide() {
    const adapter = new ChinesePhonePermissionAdapter();

    // 1. 检查支持
    const support = await adapter.checkScreenShareSupport();
    
    if (!support.supported) {
        alert(support.reason);
        return;
    }

    // 2. 显示品牌特定提示
    if (support.needsGuide) {
        console.log(`检测到${adapter.brand.displayName}，可能需要额外权限设置`);
    }

    // 3. 请求屏幕共享
    const result = await adapter.requestScreenShare({
        quality: 'medium',
        withAudio: false,
        maxRetries: 2,
        onRetry: (retryCount) => {
            console.log(`正在重试... (第${retryCount}次)`);
        }
    });

    // 4. 处理结果
    if (result.success) {
        console.log('✅ 屏幕共享成功');
        
        // 显示视频
        const videoElement = document.getElementById('screen-share-video');
        videoElement.srcObject = result.stream;
        videoElement.play();
        
    } else {
        console.error('❌ 屏幕共享失败:', result.message);
        
        // 显示引导
        if (result.needsGuide && result.guideMessage) {
            adapter.showGuideDialog(result.message + result.guideMessage);
        } else {
            alert(result.message);
        }
    }
}

/**
 * 示例：集成到现有管理器
 */
class ScreenShareManagerWithChinesePhoneSupport {
    constructor(callId) {
        this.callId = callId;
        this.adapter = new ChinesePhonePermissionAdapter();
        this.stream = null;
    }

    async start(options = {}) {
        try {
            // 使用适配器请求权限
            const result = await this.adapter.requestScreenShare({
                ...options,
                onRetry: (count) => {
                    this.showRetryNotification(count);
                }
            });

            if (result.success) {
                this.stream = result.stream;
                return result.stream;
            } else {
                if (result.needsGuide) {
                    this.adapter.showGuideDialog(result.message + result.guideMessage);
                }
                throw new Error(result.message);
            }

        } catch (error) {
            console.error('屏幕共享启动失败:', error);
            throw error;
        }
    }

    showRetryNotification(count) {
        console.log(`🔄 正在重试屏幕共享请求... (第${count}次)`);
        
        // 可以在这里显示UI提示
        const notification = document.getElementById('retry-notification');
        if (notification) {
            notification.textContent = `正在重试... (${count}/2)`;
            notification.style.display = 'block';
        }
    }

    async stop() {
        if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
            this.stream = null;
        }
    }

    getBrandInfo() {
        return this.adapter.getBrandInfo();
    }
}

// 导出
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        ChinesePhonePermissionAdapter,
        ScreenShareManagerWithChinesePhoneSupport
    };
}

if (typeof window !== 'undefined') {
    window.ChinesePhonePermissionAdapter = ChinesePhonePermissionAdapter;
    window.ScreenShareManagerWithChinesePhoneSupport = ScreenShareManagerWithChinesePhoneSupport;
}



