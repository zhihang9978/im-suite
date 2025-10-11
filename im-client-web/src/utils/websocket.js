// WebSocket连接管理
let ws = null
let reconnectTimer = null
let reconnectAttempts = 0
const MAX_RECONNECT_ATTEMPTS = 5
const RECONNECT_INTERVAL = 3000

export function connectWebSocket(token, onMessage) {
  const wsUrl = (import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws') + `?token=${token}`
  
  ws = new WebSocket(wsUrl)
  
  ws.onopen = () => {
    console.log('WebSocket连接成功')
    reconnectAttempts = 0
    
    // 发送心跳
    setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)
  }
  
  ws.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data)
      if (onMessage) {
        onMessage(message)
      }
    } catch (error) {
      console.error('WebSocket消息解析失败:', error)
    }
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket错误:', error)
  }
  
  ws.onclose = () => {
    console.log('WebSocket连接关闭')
    
    // 自动重连
    if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      reconnectTimer = setTimeout(() => {
        reconnectAttempts++
        console.log(`尝试重连 (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`)
        connectWebSocket(token, onMessage)
      }, RECONNECT_INTERVAL)
    }
  }
  
  return ws
}

export function sendMessage(message) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message))
    return true
  }
  return false
}

export function closeWebSocket() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
  }
  if (ws) {
    ws.close()
    ws = null
  }
}

