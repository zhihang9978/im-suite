import { API_CONFIG } from '../config/api';

let ws = null;
let reconnectTimer = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_INTERVAL = 3000;

// 连接WebSocket
export const connectWebSocket = (token, onMessage, onError) => {
  const wsUrl = `${API_CONFIG.WS_URL}?token=${token}`;
  
  ws = new WebSocket(wsUrl);
  
  ws.onopen = () => {
    console.log('WebSocket连接成功');
    reconnectAttempts = 0;
    
    // 发送心跳
    const heartbeatInterval = setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);
    
    // 保存interval用于清理
    ws.heartbeatInterval = heartbeatInterval;
  };
  
  ws.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data);
      if (onMessage) {
        onMessage(message);
      }
    } catch (error) {
      console.error('WebSocket消息解析失败:', error);
    }
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket错误:', error);
    if (onError) {
      onError(error);
    }
  };
  
  ws.onclose = () => {
    console.log('WebSocket连接关闭');
    
    // 清理心跳
    if (ws.heartbeatInterval) {
      clearInterval(ws.heartbeatInterval);
    }
    
    // 自动重连
    if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      reconnectTimer = setTimeout(() => {
        reconnectAttempts++;
        console.log(`尝试重连 (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`);
        connectWebSocket(token, onMessage, onError);
      }, RECONNECT_INTERVAL);
    }
  };
  
  return ws;
};

// 发送消息
export const sendWebSocketMessage = (message) => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message));
    return true;
  }
  return false;
};

// 关闭连接
export const closeWebSocket = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
  }
  
  if (ws) {
    if (ws.heartbeatInterval) {
      clearInterval(ws.heartbeatInterval);
    }
    ws.close();
    ws = null;
  }
};

// 获取连接状态
export const getWebSocketState = () => {
  return ws ? ws.readyState : WebSocket.CLOSED;
};

