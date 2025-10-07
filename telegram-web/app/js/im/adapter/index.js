/**
 * IM-Suite 适配层初始化文件
 * 负责加载所有适配层模块并初始化
 */

/**
 * 适配层初始化器
 * 在页面加载时自动初始化所有适配层组件
 */
class IMAdapterInitializer {
    constructor() {
        this.isInitialized = false;
        this.init();
    }

    /**
     * 初始化适配层
     */
    async init() {
        if (this.isInitialized) {
            console.log('IM-Suite 适配层已初始化');
            return;
        }

        console.log('正在初始化 IM-Suite 适配层...');

        try {
            // 等待页面加载完成
            if (document.readyState === 'loading') {
                await new Promise(resolve => {
                    document.addEventListener('DOMContentLoaded', resolve);
                });
            }

            // 初始化 API 适配层
            this.initAPI();
            
            // 初始化 WebSocket 适配层
            this.initWebSocket();
            
            // 初始化映射表
            this.initMapping();
            
            // 初始化调试页面
            this.initDebug();
            
            // 替换原有的 Telegram API 调用
            this.replaceTelegramAPI();
            
            this.isInitialized = true;
            console.log('IM-Suite 适配层初始化完成');
            
            // 显示初始化成功提示
            this.showInitNotification();
            
        } catch (error) {
            console.error('IM-Suite 适配层初始化失败:', error);
            this.showErrorNotification('适配层初始化失败: ' + error.message);
        }
    }

    /**
     * 初始化 API 适配层
     */
    initAPI() {
        if (!window.IMAPI) {
            console.error('IMAPI 未加载');
            return;
        }
        
        console.log('API 适配层已初始化');
        
        // 设置全局错误处理
        window.addEventListener('unhandledrejection', (event) => {
            if (event.reason && event.reason.message && event.reason.message.includes('API')) {
                console.error('API 请求失败:', event.reason);
                this.showErrorNotification('API 请求失败: ' + event.reason.message);
            }
        });
    }

    /**
     * 初始化 WebSocket 适配层
     */
    initWebSocket() {
        if (!window.IMSocket) {
            console.error('IMSocket 未加载');
            return;
        }
        
        console.log('WebSocket 适配层已初始化');
        
        // 自动连接 WebSocket
        this.autoConnectWebSocket();
        
        // 监听连接状态变化
        this.monitorWebSocketStatus();
    }

    /**
     * 初始化映射表
     */
    initMapping() {
        if (!window.TelegramAdapter) {
            console.error('TelegramAdapter 未加载');
            return;
        }
        
        console.log('映射表已初始化');
    }

    /**
     * 初始化调试页面
     */
    initDebug() {
        if (!window.IMDebugPage) {
            console.error('IMDebugPage 未加载');
            return;
        }
        
        console.log('调试页面已初始化');
        
        // 添加调试入口到页面
        this.addDebugEntry();
    }

    /**
     * 自动连接 WebSocket
     */
    async autoConnectWebSocket() {
        try {
            await window.IMSocket.connect();
            console.log('WebSocket 自动连接成功');
        } catch (error) {
            console.warn('WebSocket 自动连接失败:', error);
            // 不显示错误提示，因为可能是服务器未启动
        }
    }

    /**
     * 监控 WebSocket 状态
     */
    monitorWebSocketStatus() {
        // 监听 WebSocket 连接事件
        window.IMSocket.subscribe('connect', () => {
            console.log('WebSocket 已连接');
            this.showSuccessNotification('WebSocket 连接成功');
        });

        window.IMSocket.subscribe('disconnect', () => {
            console.log('WebSocket 已断开');
            this.showWarningNotification('WebSocket 连接断开');
        });

        window.IMSocket.subscribe('error', (error) => {
            console.error('WebSocket 错误:', error);
            this.showErrorNotification('WebSocket 错误: ' + error.message);
        });
    }

    /**
     * 替换原有的 Telegram API 调用
     */
    replaceTelegramAPI() {
        // 检查是否存在原有的 Telegram API
        if (typeof window.Telegram !== 'undefined') {
            console.log('检测到原有 Telegram API，开始替换...');
            
            // 备份原有 API
            window.TelegramOriginal = window.Telegram;
            
            // 替换为适配器
            window.Telegram = window.TelegramAdapter;
            
            console.log('Telegram API 已替换为适配器');
        } else {
            console.log('未检测到原有 Telegram API，直接设置适配器');
            window.Telegram = window.TelegramAdapter;
        }
    }

    /**
     * 添加调试入口
     */
    addDebugEntry() {
        // 创建调试入口按钮
        const debugButton = document.createElement('div');
        debugButton.id = 'im-debug-entry';
        debugButton.style.cssText = `
            position: fixed;
            bottom: 20px;
            right: 20px;
            width: 50px;
            height: 50px;
            background: #4CAF50;
            border-radius: 50%;
            cursor: pointer;
            z-index: 999998;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.3);
            transition: all 0.3s ease;
        `;
        debugButton.innerHTML = '🐛';
        debugButton.title = 'IM-Suite 调试面板 (Ctrl+Shift+D)';
        
        // 添加悬停效果
        debugButton.addEventListener('mouseenter', () => {
            debugButton.style.transform = 'scale(1.1)';
            debugButton.style.background = '#45a049';
        });
        
        debugButton.addEventListener('mouseleave', () => {
            debugButton.style.transform = 'scale(1)';
            debugButton.style.background = '#4CAF50';
        });
        
        // 点击打开调试面板
        debugButton.addEventListener('click', () => {
            window.IMDebugPage.show();
        });
        
        // 添加到页面
        document.body.appendChild(debugButton);
    }

    /**
     * 显示初始化成功通知
     */
    showInitNotification() {
        this.showNotification('IM-Suite 适配层已加载', 'success');
    }

    /**
     * 显示成功通知
     */
    showSuccessNotification(message) {
        this.showNotification(message, 'success');
    }

    /**
     * 显示警告通知
     */
    showWarningNotification(message) {
        this.showNotification(message, 'warning');
    }

    /**
     * 显示错误通知
     */
    showErrorNotification(message) {
        this.showNotification(message, 'error');
    }

    /**
     * 显示通知
     */
    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 4px;
            color: white;
            font-family: Arial, sans-serif;
            font-size: 14px;
            z-index: 999999;
            max-width: 300px;
            word-wrap: break-word;
            box-shadow: 0 2px 10px rgba(0,0,0,0.3);
            animation: slideIn 0.3s ease;
        `;
        
        // 根据类型设置颜色
        switch (type) {
            case 'success':
                notification.style.background = '#4CAF50';
                break;
            case 'warning':
                notification.style.background = '#FF9800';
                break;
            case 'error':
                notification.style.background = '#f44336';
                break;
            default:
                notification.style.background = '#2196F3';
        }
        
        notification.textContent = message;
        document.body.appendChild(notification);
        
        // 3秒后自动移除
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 3000);
    }
}

// 添加 CSS 动画
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
`;
document.head.appendChild(style);

// 创建全局初始化器实例
window.IMAdapterInitializer = new IMAdapterInitializer();

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = IMAdapterInitializer;
}
