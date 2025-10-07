/**
 * IM-Suite 调试页面
 * 提供隐藏的调试入口，用于测试 API 和 WebSocket 连接
 */

/**
 * 调试页面类
 * 提供各种测试功能，包括 API 测试、WebSocket 测试、消息发送测试等
 */
class IMDebugPage {
    constructor() {
        this.isVisible = false;
        this.testResults = [];
        this.init();
    }

    /**
     * 初始化调试页面
     */
    init() {
        this.createDebugPanel();
        this.bindEvents();
        this.addKeyboardShortcut();
    }

    /**
     * 创建调试面板
     */
    createDebugPanel() {
        // 创建调试面板容器
        this.debugContainer = document.createElement('div');
        this.debugContainer.id = 'im-debug-panel';
        this.debugContainer.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.8);
            z-index: 999999;
            display: none;
            font-family: Arial, sans-serif;
            color: white;
        `;

        // 创建调试面板内容
        this.debugContainer.innerHTML = `
            <div style="
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                background: #1a1a1a;
                border-radius: 8px;
                padding: 20px;
                width: 80%;
                max-width: 600px;
                max-height: 80%;
                overflow-y: auto;
            ">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                    <h2 style="margin: 0; color: #4CAF50;">IM-Suite 调试面板</h2>
                    <button id="close-debug" style="
                        background: #f44336;
                        color: white;
                        border: none;
                        padding: 8px 16px;
                        border-radius: 4px;
                        cursor: pointer;
                    ">关闭</button>
                </div>
                
                <div style="margin-bottom: 20px;">
                    <h3 style="color: #2196F3; margin-bottom: 10px;">连接状态</h3>
                    <div id="connection-status" style="
                        padding: 10px;
                        background: #333;
                        border-radius: 4px;
                        margin-bottom: 10px;
                    ">
                        <div>API 状态: <span id="api-status">检查中...</span></div>
                        <div>WebSocket 状态: <span id="ws-status">检查中...</span></div>
                    </div>
                </div>

                <div style="margin-bottom: 20px;">
                    <h3 style="color: #2196F3; margin-bottom: 10px;">API 测试</h3>
                    <div style="display: flex; gap: 10px; margin-bottom: 10px;">
                        <button id="test-ping" style="
                            background: #4CAF50;
                            color: white;
                            border: none;
                            padding: 8px 16px;
                            border-radius: 4px;
                            cursor: pointer;
                        ">测试 Ping</button>
                        <button id="test-user" style="
                            background: #2196F3;
                            color: white;
                            border: none;
                            padding: 8px 16px;
                            border-radius: 4px;
                            cursor: pointer;
                        ">测试用户信息</button>
                        <button id="test-contacts" style="
                            background: #FF9800;
                            color: white;
                            border: none;
                            padding: 8px 16px;
                            border-radius: 4px;
                            cursor: pointer;
                        ">测试联系人</button>
                    </div>
                </div>

                <div style="margin-bottom: 20px;">
                    <h3 style="color: #2196F3; margin-bottom: 10px;">WebSocket 测试</h3>
                    <div style="display: flex; gap: 10px; margin-bottom: 10px;">
                        <button id="test-ws-connect" style="
                            background: #4CAF50;
                            color: white;
                            border: none;
                            padding: 8px 16px;
                            border-radius: 4px;
                            cursor: pointer;
                        ">连接 WebSocket</button>
                        <button id="test-ws-disconnect" style="
                            background: #f44336;
                            color: white;
                            border: none;
                            padding: 8px 16px;
                            border-radius: 4px;
                            cursor: pointer;
                        ">断开 WebSocket</button>
                        <button id="test-ws-echo" style="
                            background: #9C27B0;
                            color: white;
                            border: none;
                            padding: 8px 16px;
                            border-radius: 4px;
                            cursor: pointer;
                        ">测试 Echo</button>
                    </div>
                </div>

                <div style="margin-bottom: 20px;">
                    <h3 style="color: #2196F3; margin-bottom: 10px;">消息测试</h3>
                    <div style="margin-bottom: 10px;">
                        <input type="text" id="test-message" placeholder="输入测试消息" style="
                            width: 100%;
                            padding: 8px;
                            border: 1px solid #555;
                            border-radius: 4px;
                            background: #333;
                            color: white;
                            margin-bottom: 10px;
                        ">
                        <button id="test-send-message" style="
                            background: #4CAF50;
                            color: white;
                            border: none;
                            padding: 8px 16px;
                            border-radius: 4px;
                            cursor: pointer;
                        ">发送测试消息</button>
                    </div>
                </div>

                <div>
                    <h3 style="color: #2196F3; margin-bottom: 10px;">测试结果</h3>
                    <div id="test-results" style="
                        background: #333;
                        border-radius: 4px;
                        padding: 10px;
                        max-height: 200px;
                        overflow-y: auto;
                        font-family: monospace;
                        font-size: 12px;
                    "></div>
                </div>
            </div>
        `;

        // 添加到页面
        document.body.appendChild(this.debugContainer);
    }

    /**
     * 绑定事件
     */
    bindEvents() {
        // 关闭按钮
        document.getElementById('close-debug').addEventListener('click', () => {
            this.hide();
        });

        // API 测试按钮
        document.getElementById('test-ping').addEventListener('click', () => {
            this.testPing();
        });

        document.getElementById('test-user').addEventListener('click', () => {
            this.testUserInfo();
        });

        document.getElementById('test-contacts').addEventListener('click', () => {
            this.testContacts();
        });

        // WebSocket 测试按钮
        document.getElementById('test-ws-connect').addEventListener('click', () => {
            this.testWSConnect();
        });

        document.getElementById('test-ws-disconnect').addEventListener('click', () => {
            this.testWSDisconnect();
        });

        document.getElementById('test-ws-echo').addEventListener('click', () => {
            this.testWSEcho();
        });

        // 消息测试按钮
        document.getElementById('test-send-message').addEventListener('click', () => {
            this.testSendMessage();
        });

        // 点击背景关闭
        this.debugContainer.addEventListener('click', (e) => {
            if (e.target === this.debugContainer) {
                this.hide();
            }
        });
    }

    /**
     * 添加键盘快捷键
     */
    addKeyboardShortcut() {
        document.addEventListener('keydown', (e) => {
            // Ctrl + Shift + D 打开调试面板
            if (e.ctrlKey && e.shiftKey && e.key === 'D') {
                e.preventDefault();
                this.toggle();
            }
            
            // ESC 关闭调试面板
            if (e.key === 'Escape' && this.isVisible) {
                this.hide();
            }
        });
    }

    /**
     * 显示调试面板
     */
    show() {
        this.debugContainer.style.display = 'block';
        this.isVisible = true;
        this.checkConnectionStatus();
    }

    /**
     * 隐藏调试面板
     */
    hide() {
        this.debugContainer.style.display = 'none';
        this.isVisible = false;
    }

    /**
     * 切换调试面板显示状态
     */
    toggle() {
        if (this.isVisible) {
            this.hide();
        } else {
            this.show();
        }
    }

    /**
     * 检查连接状态
     */
    async checkConnectionStatus() {
        // 检查 API 状态
        try {
            const response = await window.IMAPI.get('/ping');
            document.getElementById('api-status').textContent = '✅ 连接正常';
            document.getElementById('api-status').style.color = '#4CAF50';
        } catch (error) {
            document.getElementById('api-status').textContent = '❌ 连接失败';
            document.getElementById('api-status').style.color = '#f44336';
        }

        // 检查 WebSocket 状态
        if (window.IMSocket && window.IMSocket.isConnected) {
            document.getElementById('ws-status').textContent = '✅ 已连接';
            document.getElementById('ws-status').style.color = '#4CAF50';
        } else {
            document.getElementById('ws-status').textContent = '❌ 未连接';
            document.getElementById('ws-status').style.color = '#f44336';
        }
    }

    /**
     * 添加测试结果
     */
    addTestResult(message, type = 'info') {
        const timestamp = new Date().toLocaleTimeString();
        const result = `[${timestamp}] ${message}`;
        this.testResults.push({ message: result, type });
        
        const resultsContainer = document.getElementById('test-results');
        const resultElement = document.createElement('div');
        resultElement.textContent = result;
        resultElement.style.color = type === 'error' ? '#f44336' : type === 'success' ? '#4CAF50' : '#2196F3';
        resultElement.style.marginBottom = '5px';
        
        resultsContainer.appendChild(resultElement);
        resultsContainer.scrollTop = resultsContainer.scrollHeight;
    }

    /**
     * 测试 Ping
     */
    async testPing() {
        try {
            this.addTestResult('正在测试 Ping...', 'info');
            const response = await window.IMAPI.get('/ping');
            this.addTestResult(`Ping 成功: ${JSON.stringify(response)}`, 'success');
        } catch (error) {
            this.addTestResult(`Ping 失败: ${error.message}`, 'error');
        }
    }

    /**
     * 测试用户信息
     */
    async testUserInfo() {
        try {
            this.addTestResult('正在获取用户信息...', 'info');
            const response = await window.IMAPI.getCurrentUser();
            this.addTestResult(`用户信息: ${JSON.stringify(response)}`, 'success');
        } catch (error) {
            this.addTestResult(`获取用户信息失败: ${error.message}`, 'error');
        }
    }

    /**
     * 测试联系人
     */
    async testContacts() {
        try {
            this.addTestResult('正在获取联系人列表...', 'info');
            const response = await window.IMAPI.getContacts();
            this.addTestResult(`联系人列表: ${JSON.stringify(response)}`, 'success');
        } catch (error) {
            this.addTestResult(`获取联系人失败: ${error.message}`, 'error');
        }
    }

    /**
     * 测试 WebSocket 连接
     */
    async testWSConnect() {
        try {
            this.addTestResult('正在连接 WebSocket...', 'info');
            await window.IMSocket.connect();
            this.addTestResult('WebSocket 连接成功', 'success');
            this.checkConnectionStatus();
        } catch (error) {
            this.addTestResult(`WebSocket 连接失败: ${error.message}`, 'error');
        }
    }

    /**
     * 测试 WebSocket 断开
     */
    testWSDisconnect() {
        try {
            this.addTestResult('正在断开 WebSocket...', 'info');
            window.IMSocket.disconnect();
            this.addTestResult('WebSocket 已断开', 'success');
            this.checkConnectionStatus();
        } catch (error) {
            this.addTestResult(`WebSocket 断开失败: ${error.message}`, 'error');
        }
    }

    /**
     * 测试 WebSocket Echo
     */
    async testWSEcho() {
        try {
            this.addTestResult('正在测试 WebSocket Echo...', 'info');
            const testMessage = { type: 'echo', data: 'Hello IM-Suite!' };
            await window.IMSocket.send(testMessage);
            this.addTestResult('Echo 消息已发送', 'success');
        } catch (error) {
            this.addTestResult(`Echo 测试失败: ${error.message}`, 'error');
        }
    }

    /**
     * 测试发送消息
     */
    async testSendMessage() {
        try {
            const messageInput = document.getElementById('test-message');
            const message = messageInput.value.trim();
            
            if (!message) {
                this.addTestResult('请输入测试消息', 'error');
                return;
            }

            this.addTestResult(`正在发送消息: ${message}`, 'info');
            
            // 这里需要实际的聊天ID，暂时使用测试ID
            const response = await window.IMAPI.sendMessage(1, {
                content: message,
                message_type: 'text'
            });
            
            this.addTestResult(`消息发送成功: ${JSON.stringify(response)}`, 'success');
            messageInput.value = '';
        } catch (error) {
            this.addTestResult(`消息发送失败: ${error.message}`, 'error');
        }
    }
}

// 创建全局调试页面实例
window.IMDebugPage = new IMDebugPage();

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = IMDebugPage;
}
