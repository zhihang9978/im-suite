import axios from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { API_CONFIG } from '../config/api';

// 创建axios实例
const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  async (config) => {
    // 添加token
    const token = await AsyncStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    if (error.response) {
      const { status } = error.response;
      
      if (status === 401) {
        // 未授权，清除token并跳转登录
        await AsyncStorage.removeItem('token');
        // 触发导航到登录页（需要通过事件系统）
      }
    }
    
    return Promise.reject(error);
  }
);

// API方法
export const authAPI = {
  // 登录
  login: (phone, password) => 
    api.post('/api/auth/login', { phone, password }),
  
  // 注册
  register: (phone, password, nickname) => 
    api.post('/api/auth/register', { phone, password, nickname }),
  
  // 登出
  logout: () => 
    api.post('/api/auth/logout'),
  
  // 获取当前用户信息
  getMe: () => 
    api.get('/api/users/me'),
};

export const messageAPI = {
  // 发送消息
  send: (receiverId, content, messageType = 'text') => 
    api.post('/api/messages/send', {
      receiver_id: receiverId,
      content,
      message_type: messageType
    }),
  
  // 获取消息列表
  getMessages: (receiverId, limit = 50) => 
    api.get(`/api/messages?receiver_id=${receiverId}&limit=${limit}`),
  
  // 标记已读
  markAsRead: (messageId) => 
    api.post(`/api/messages/${messageId}/read`),
  
  // 撤回消息
  recall: (messageId, reason = '') => 
    api.post(`/api/messages/${messageId}/recall`, { reason }),
  
  // 删除消息
  delete: (messageId) => 
    api.delete(`/api/messages/${messageId}`),
};

export const contactAPI = {
  // 获取好友列表
  getFriends: () => 
    api.get('/api/users/friends'),
  
  // 搜索用户
  search: (keyword) => 
    api.get(`/api/users/search?keyword=${keyword}`),
};

export const fileAPI = {
  // 上传文件
  upload: (formData) => 
    api.post('/api/files/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    }),
};

export default api;

