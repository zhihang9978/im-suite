/**
 * IM-Suite WebSocket 适配层
 * 提供实时通讯功能，替代原有的 MTProto 实时连接
 */

// 获取 WebSocket 基础地址
const WS_BASE = window.VITE_WS_BASE_URL || 'ws://localhost:8080/ws';

/**
 * WebSocket 连接管理器
 * 处理实时消息推送和事件订阅
 */
class IMSocket {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectInterval = 3000;
        this.listeners = new Map();
        this.isConnecting = false;
        this.isConnected = false;
    }

    /**
     * 连接到 WebSocket 服务器
     * @returns {Promise} 连接结果
     */
    async connect() {
        if (this.isConnecting || this.isConnected) {
            return Promise.resolve();
        }

        this.isConnecting = true;

        return new Promise((resolve, reject) => {
            try {
                this.ws = new WebSocket(WS_BASE);
                
                this.ws.onopen = () => {
                    console.log('WebSocket 连接已建立');
                    this.isConnected = true;
                    this.isConnecting = false;
                    this.reconnectAttempts = 0;
                    resolve();
                };

                this.ws.onmessage = (event) => {
                    this.handleMessage(event);
                };

                this.ws.onclose = (event) => {
                    console.log('WebSocket 连接已关闭:', event.code, event.reason);
                    this.isConnected = false;
                    this.isConnecting = false;
                    
                    // 如果不是主动关闭，尝试重连
                    if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
                        this.scheduleReconnect();
                    }
                };

                this.ws.onerror = (error) => {
                    console.error('WebSocket 连接错误:', error);
                    this.isConnecting = false;
                    reject(error);
                };

            } catch (error) {
                this.isConnecting = false;
                reject(error);
            }
        });
    }

    /**
     * 断开 WebSocket 连接
     */
    disconnect() {
        if (this.ws) {
            this.ws.close(1000, '主动断开连接');
            this.ws = null;
        }
        this.isConnected = false;
        this.isConnecting = false;
    }

    /**
     * 发送消息
     * @param {Object} data - 要发送的数据
     * @returns {Promise} 发送结果
     */
    async send(data) {
        if (!this.isConnected || !this.ws) {
            throw new Error('WebSocket 未连接');
        }

        return new Promise((resolve, reject) => {
            try {
                const message = JSON.stringify(data);
                this.ws.send(message);
                resolve();
            } catch (error) {
                reject(error);
            }
        });
    }

    /**
     * 订阅事件
     * @param {string} event - 事件名称
     * @param {Function} callback - 回调函数
     */
    subscribe(event, callback) {
        if (!this.listeners.has(event)) {
            this.listeners.set(event, []);
        }
        this.listeners.get(event).push(callback);
    }

    /**
     * 取消订阅事件
     * @param {string} event - 事件名称
     * @param {Function} callback - 回调函数
     */
    unsubscribe(event, callback) {
        if (this.listeners.has(event)) {
            const callbacks = this.listeners.get(event);
            const index = callbacks.indexOf(callback);
            if (index > -1) {
                callbacks.splice(index, 1);
            }
        }
    }

    /**
     * 处理接收到的消息
     * @param {MessageEvent} event - WebSocket 消息事件
     */
    handleMessage(event) {
        try {
            const data = JSON.parse(event.data);
            const { type, payload } = data;

            // 触发对应的事件监听器
            if (this.listeners.has(type)) {
                this.listeners.get(type).forEach(callback => {
                    try {
                        callback(payload);
                    } catch (error) {
                        console.error('事件回调执行失败:', error);
                    }
                });
            }

            // 触发通用消息事件
            if (this.listeners.has('message')) {
                this.listeners.get('message').forEach(callback => {
                    try {
                        callback(data);
                    } catch (error) {
                        console.error('消息回调执行失败:', error);
                    }
                });
            }

        } catch (error) {
            console.error('解析 WebSocket 消息失败:', error);
        }
    }

    /**
     * 安排重连
     */
    scheduleReconnect() {
        this.reconnectAttempts++;
        console.log(`WebSocket 重连尝试 ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
        
        setTimeout(() => {
            if (this.reconnectAttempts <= this.maxReconnectAttempts) {
                this.connect().catch(error => {
                    console.error('WebSocket 重连失败:', error);
                });
            }
        }, this.reconnectInterval);
    }

    // ==================== 实时事件处理 ====================

    /**
     * 发送新消息
     * @param {Object} message - 消息对象
     */
    async sendNewMessage(message) {
        return await this.send({
            type: 'message.new',
            payload: message
        });
    }

    /**
     * 发送消息编辑事件
     * @param {Object} message - 编辑后的消息
     */
    async sendMessageEdit(message) {
        return await this.send({
            type: 'message.edit',
            payload: message
        });
    }

    /**
     * 发送消息删除事件
     * @param {number} messageId - 消息ID
     */
    async sendMessageDelete(messageId) {
        return await this.send({
            type: 'message.delete',
            payload: { messageId }
        });
    }

    /**
     * 发送消息已读事件
     * @param {number} messageId - 消息ID
     * @param {number} userId - 用户ID
     */
    async sendMessageRead(messageId, userId) {
        return await this.send({
            type: 'message.read',
            payload: { messageId, userId }
        });
    }

    /**
     * 发送正在输入事件
     * @param {number} chatId - 聊天ID
     * @param {boolean} isTyping - 是否正在输入
     */
    async sendTyping(chatId, isTyping = true) {
        return await this.send({
            type: 'typing',
            payload: { chatId, isTyping }
        });
    }

    /**
     * 发送在线状态事件
     * @param {boolean} isOnline - 是否在线
     */
    async sendPresence(isOnline) {
        return await this.send({
            type: 'presence',
            payload: { isOnline }
        });
    }

    // ==================== 通话相关事件 ====================

    /**
     * 发送通话邀请
     * @param {Object} callData - 通话数据
     */
    async sendCallOffer(callData) {
        return await this.send({
            type: 'call.offer',
            payload: callData
        });
    }

    /**
     * 发送通话应答
     * @param {Object} callData - 通话数据
     */
    async sendCallAnswer(callData) {
        return await this.send({
            type: 'call.answer',
            payload: callData
        });
    }

    /**
     * 发送 ICE 候选
     * @param {Object} iceData - ICE 数据
     */
    async sendCallIce(iceData) {
        return await this.send({
            type: 'call.ice',
            payload: iceData
        });
    }

    /**
     * 发送通话结束
     * @param {Object} callData - 通话数据
     */
    async sendCallEnd(callData) {
        return await this.send({
            type: 'call.end',
            payload: callData
        });
    }
}

// 创建全局 WebSocket 实例
window.IMSocket = new IMSocket();

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = IMSocket;
}
