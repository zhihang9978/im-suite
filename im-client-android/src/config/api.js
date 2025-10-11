export const API_CONFIG = {
  // 后端API地址
  BASE_URL: 'http://154.37.214.191:8080',
  
  // WebSocket地址
  WS_URL: 'ws://154.37.214.191:8080/ws',
  
  // 请求超时时间
  TIMEOUT: 30000,
  
  // 上传文件大小限制（50MB）
  MAX_FILE_SIZE: 50 * 1024 * 1024,
  
  // 心跳间隔（30秒）
  HEARTBEAT_INTERVAL: 30000,
  
  // 重连最大次数
  MAX_RECONNECT_ATTEMPTS: 5,
  
  // 重连间隔（3秒）
  RECONNECT_INTERVAL: 3000,
};

