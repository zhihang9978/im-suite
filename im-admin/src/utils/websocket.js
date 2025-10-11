/**
 * WebSocket管理器 - 支持自动重连
 */

class WebSocketManager {
  constructor(url, options = {}) {
    this.url = url
    this.ws = null
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = options.maxReconnectAttempts || 10
    this.reconnectInterval = options.reconnectInterval || 1000
    this.heartbeatInterval = options.heartbeatInterval || 30000
    this.heartbeatTimer = null
    this.reconnectTimer = null
    this.isManualClose = false
    this.messageHandlers = []
    this.errorHandlers = []
    this.closeHandlers = []
    this.openHandlers = []
    
    this.connect()
  }

  connect() {
    try {
      this.ws = new WebSocket(this.url)
      
      this.ws.onopen = () => {
        console.log('WebSocket连接已建立')
        this.reconnectAttempts = 0
        this.isManualClose = false
        this.startHeartbeat()
        
        // 触发open事件
        this.openHandlers.forEach(handler => handler())
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          
          // 处理心跳响应
          if (data.type === 'pong') {
            console.log('收到心跳响应')
            return
          }
          
          // 触发消息处理器
          this.messageHandlers.forEach(handler => handler(data))
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
        this.errorHandlers.forEach(handler => handler(error))
      }

      this.ws.onclose = (event) => {
        console.log('WebSocket连接已关闭:', event.code, event.reason)
        this.stopHeartbeat()
        
        // 触发close事件
        this.closeHandlers.forEach(handler => handler(event))
        
        // 如果不是手动关闭，则尝试重连
        if (!this.isManualClose) {
          this.reconnect()
        }
      }
    } catch (error) {
      console.error('创建WebSocket连接失败:', error)
      this.reconnect()
    }
  }

  reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('WebSocket重连次数已达上限')
      return
    }

    // 指数退避重连策略
    const delay = Math.min(
      this.reconnectInterval * Math.pow(2, this.reconnectAttempts),
      30000 // 最大30秒
    )

    console.log(`${delay/1000}秒后尝试第${this.reconnectAttempts + 1}次重连...`)

    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++
      this.connect()
    }, delay)
  }

  startHeartbeat() {
    this.stopHeartbeat()
    
    this.heartbeatTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping' })
      }
    }, this.heartbeatInterval)
  }

  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
      return true
    } else {
      console.warn('WebSocket未连接，消息未发送')
      return false
    }
  }

  close() {
    this.isManualClose = true
    this.stopHeartbeat()
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  onMessage(handler) {
    this.messageHandlers.push(handler)
  }

  onError(handler) {
    this.errorHandlers.push(handler)
  }

  onClose(handler) {
    this.closeHandlers.push(handler)
  }

  onOpen(handler) {
    this.openHandlers.push(handler)
  }

  getReadyState() {
    return this.ws ? this.ws.readyState : WebSocket.CLOSED
  }

  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN
  }
}

export default WebSocketManager

