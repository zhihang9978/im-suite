/**
 * IM-Suite WebSocket 适配层
 * 将 Telegram Web 的 WebSocket 连接重定向到我们的后端
 */

// 获取 WebSocket 基础地址
const WS_BASE = window.VITE_WS_BASE_URL || 'ws://localhost:8080/ws';

/**
 * WebSocket 连接管理器
 * 封装 WebSocket 连接，提供统一的接口
 */
class IMWebSocket {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectInterval = 1000;
    private listeners: Map<string, Function[]> = new Map();
    private isConnecting = false;
    private heartbeatInterval: number | null = null;

    constructor() {
        this.connect();
    }

    /**
     * 连接到 WebSocket 服务器
     */
    connect(): void {
        if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
            return;
        }

        this.isConnecting = true;
        console.log('正在连接到 WebSocket 服务器:', WS_BASE);

        try {
            this.ws = new WebSocket(WS_BASE);
            this.setupEventHandlers();
        } catch (error) {
            console.error('WebSocket 连接失败:', error);
            this.handleReconnect();
        }
    }

    /**
     * 设置 WebSocket 事件处理器
     */
    private setupEventHandlers(): void {
        if (!this.ws) return;

        this.ws.onopen = () => {
            console.log('WebSocket 连接已建立');
            this.isConnecting = false;
            this.reconnectAttempts = 0;
            this.startHeartbeat();
            this.emit('connected');
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleMessage(data);
            } catch (error) {
                console.error('WebSocket 消息解析失败:', error);
            }
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket 连接已关闭:', event.code, event.reason);
            this.isConnecting = false;
            this.stopHeartbeat();
            this.emit('disconnected', { code: event.code, reason: event.reason });
            
            if (event.code !== 1000) { // 非正常关闭
                this.handleReconnect();
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket 错误:', error);
            this.isConnecting = false;
            this.emit('error', error);
        };
    }

    /**
     * 处理收到的消息
     */
    private handleMessage(data: any): void {
        console.log('收到 WebSocket 消息:', data);
        
        // 处理不同类型的消息
        switch (data.type) {
            case 'message.new':
                this.emit('message', data.data);
                break;
            case 'message.edit':
                this.emit('messageEdit', data.data);
                break;
            case 'message.delete':
                this.emit('messageDelete', data.data);
                break;
            case 'message.read':
                this.emit('messageRead', data.data);
                break;
            case 'typing':
                this.emit('typing', data.data);
                break;
            case 'presence':
                this.emit('presence', data.data);
                break;
            case 'call.offer':
                this.emit('callOffer', data.data);
                break;
            case 'call.answer':
                this.emit('callAnswer', data.data);
                break;
            case 'call.ice':
                this.emit('callIce', data.data);
                break;
            case 'call.end':
                this.emit('callEnd', data.data);
                break;
            case 'pong':
                // 心跳响应，无需处理
                break;
            default:
                this.emit('message', data);
        }
    }

    /**
     * 发送消息
     */
    send(data: any): void {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            try {
                const message = JSON.stringify(data);
                this.ws.send(message);
                console.log('发送 WebSocket 消息:', data);
            } catch (error) {
                console.error('WebSocket 发送失败:', error);
            }
        } else {
            console.warn('WebSocket 未连接，无法发送消息');
        }
    }

    /**
     * 发送文本消息
     */
    sendMessage(chatId: number, content: string, type: string = 'text'): void {
        this.send({
            type: 'message.send',
            data: {
                chat_id: chatId,
                content: content,
                message_type: type
            }
        });
    }

    /**
     * 发送文件消息
     */
    sendFile(chatId: number, file: File, type: string): void {
        // 创建文件读取器
        const reader = new FileReader();
        reader.onload = (e) => {
            const result = e.target?.result;
            if (result) {
                this.send({
                    type: 'message.send',
                    data: {
                        chat_id: chatId,
                        content: '',
                        message_type: type,
                        file_name: file.name,
                        file_size: file.size,
                        file_data: result
                    }
                });
            }
        };
        reader.readAsDataURL(file);
    }

    /**
     * 发送输入状态
     */
    sendTyping(chatId: number): void {
        this.send({
            type: 'typing',
            data: {
                chat_id: chatId
            }
        });
    }

    /**
     * 发送在线状态
     */
    sendPresence(status: 'online' | 'offline' | 'away'): void {
        this.send({
            type: 'presence',
            data: {
                status: status
            }
        });
    }

    /**
     * 发送通话信令
     */
    sendCallSignal(type: 'offer' | 'answer' | 'ice' | 'end', data: any): void {
        this.send({
            type: `call.${type}`,
            data: data
        });
    }

    /**
     * 订阅事件
     */
    subscribe(event: string, callback: Function): void {
        if (!this.listeners.has(event)) {
            this.listeners.set(event, []);
        }
        this.listeners.get(event)!.push(callback);
    }

    /**
     * 取消订阅事件
     */
    unsubscribe(event: string, callback: Function): void {
        const callbacks = this.listeners.get(event);
        if (callbacks) {
            const index = callbacks.indexOf(callback);
            if (index > -1) {
                callbacks.splice(index, 1);
            }
        }
    }

    /**
     * 触发事件
     */
    private emit(event: string, data?: any): void {
        const callbacks = this.listeners.get(event);
        if (callbacks) {
            callbacks.forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error(`WebSocket 事件处理器错误 (${event}):`, error);
                }
            });
        }
    }

    /**
     * 处理重连
     */
    private handleReconnect(): void {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('WebSocket 重连次数超限，停止重连');
            this.emit('reconnectFailed');
            return;
        }

        this.reconnectAttempts++;
        const delay = this.reconnectInterval * Math.pow(2, this.reconnectAttempts - 1);
        
        console.log(`WebSocket 将在 ${delay}ms 后重连 (第 ${this.reconnectAttempts} 次)`);
        
        setTimeout(() => {
            this.connect();
        }, delay);
    }

    /**
     * 开始心跳
     */
    private startHeartbeat(): void {
        this.heartbeatInterval = window.setInterval(() => {
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                this.send({ type: 'ping' });
            }
        }, 30000); // 每30秒发送一次心跳
    }

    /**
     * 停止心跳
     */
    private stopHeartbeat(): void {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
    }

    /**
     * 断开连接
     */
    disconnect(): void {
        this.stopHeartbeat();
        if (this.ws) {
            this.ws.close(1000, '主动断开连接');
            this.ws = null;
        }
    }

    /**
     * 获取连接状态
     */
    getReadyState(): number {
        return this.ws ? this.ws.readyState : WebSocket.CLOSED;
    }

    /**
     * 是否已连接
     */
    isConnected(): boolean {
        return this.ws ? this.ws.readyState === WebSocket.OPEN : false;
    }
}

// 创建全局 WebSocket 实例
window.IMWebSocket = new IMWebSocket();

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = IMWebSocket;
}


