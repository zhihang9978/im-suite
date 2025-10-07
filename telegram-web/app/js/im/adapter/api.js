/**
 * IM-Suite API 适配层
 * 将 Telegram Web 的网络调用重定向到我们的 REST 后端
 */

// 获取 API 基础地址，默认指向本地后端
const API_BASE = window.VITE_API_BASE_URL || 'http://localhost:8080/api';
const WS_BASE = window.VITE_WS_BASE_URL || 'ws://localhost:8080/ws';

/**
 * API 请求封装类
 * 提供统一的 HTTP 请求接口，替代原有的 MTProto 调用
 */
class IMAPI {
    constructor() {
        this.baseURL = API_BASE;
        this.token = localStorage.getItem('im_token');
    }

    /**
     * 设置认证令牌
     * @param {string} token - JWT 令牌
     */
    setToken(token) {
        this.token = token;
        localStorage.setItem('im_token', token);
    }

    /**
     * 清除认证令牌
     */
    clearToken() {
        this.token = null;
        localStorage.removeItem('im_token');
    }

    /**
     * 获取请求头
     * @returns {Object} 请求头对象
     */
    getHeaders() {
        const headers = {
            'Content-Type': 'application/json',
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    }

    /**
     * 处理 API 响应
     * @param {Response} response - Fetch 响应对象
     * @returns {Promise} 解析后的数据
     */
    async handleResponse(response) {
        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: '网络请求失败' }));
            throw new Error(error.message || `HTTP ${response.status}`);
        }
        
        return await response.json();
    }

    /**
     * GET 请求
     * @param {string} endpoint - API 端点
     * @param {Object} params - 查询参数
     * @returns {Promise} 响应数据
     */
    async get(endpoint, params = {}) {
        const url = new URL(`${this.baseURL}${endpoint}`);
        Object.keys(params).forEach(key => {
            if (params[key] !== undefined && params[key] !== null) {
                url.searchParams.append(key, params[key]);
            }
        });

        try {
            const response = await fetch(url, {
                method: 'GET',
                headers: this.getHeaders(),
            });
            
            return await this.handleResponse(response);
        } catch (error) {
            console.error('API GET 请求失败:', error);
            throw error;
        }
    }

    /**
     * POST 请求
     * @param {string} endpoint - API 端点
     * @param {Object} data - 请求数据
     * @returns {Promise} 响应数据
     */
    async post(endpoint, data = {}) {
        try {
            const response = await fetch(`${this.baseURL}${endpoint}`, {
                method: 'POST',
                headers: this.getHeaders(),
                body: JSON.stringify(data),
            });
            
            return await this.handleResponse(response);
        } catch (error) {
            console.error('API POST 请求失败:', error);
            throw error;
        }
    }

    /**
     * PUT 请求
     * @param {string} endpoint - API 端点
     * @param {Object} data - 请求数据
     * @returns {Promise} 响应数据
     */
    async put(endpoint, data = {}) {
        try {
            const response = await fetch(`${this.baseURL}${endpoint}`, {
                method: 'PUT',
                headers: this.getHeaders(),
                body: JSON.stringify(data),
            });
            
            return await this.handleResponse(response);
        } catch (error) {
            console.error('API PUT 请求失败:', error);
            throw error;
        }
    }

    /**
     * DELETE 请求
     * @param {string} endpoint - API 端点
     * @returns {Promise} 响应数据
     */
    async delete(endpoint) {
        try {
            const response = await fetch(`${this.baseURL}${endpoint}`, {
                method: 'DELETE',
                headers: this.getHeaders(),
            });
            
            return await this.handleResponse(response);
        } catch (error) {
            console.error('API DELETE 请求失败:', error);
            throw error;
        }
    }

    // ==================== 认证相关接口 ====================

    /**
     * 用户登录
     * @param {string} phone - 手机号
     * @param {string} code - 验证码或密码
     * @returns {Promise} 登录结果
     */
    async login(phone, code) {
        const result = await this.post('/auth/login', { phone, code });
        if (result.token) {
            this.setToken(result.token);
        }
        return result;
    }

    /**
     * 刷新令牌
     * @returns {Promise} 新的令牌
     */
    async refreshToken() {
        const result = await this.post('/auth/refresh');
        if (result.token) {
            this.setToken(result.token);
        }
        return result;
    }

    /**
     * 用户登出
     * @returns {Promise} 登出结果
     */
    async logout() {
        try {
            await this.post('/auth/logout');
        } finally {
            this.clearToken();
        }
    }

    // ==================== 用户相关接口 ====================

    /**
     * 获取当前用户信息
     * @returns {Promise} 用户信息
     */
    async getCurrentUser() {
        return await this.get('/users/me');
    }

    /**
     * 更新用户信息
     * @param {Object} userData - 用户数据
     * @returns {Promise} 更新结果
     */
    async updateUser(userData) {
        return await this.put('/users/me', userData);
    }

    // ==================== 联系人相关接口 ====================

    /**
     * 获取联系人列表
     * @returns {Promise} 联系人列表
     */
    async getContacts() {
        return await this.get('/contacts');
    }

    /**
     * 添加联系人
     * @param {string} phone - 手机号
     * @param {string} nickname - 昵称
     * @returns {Promise} 添加结果
     */
    async addContact(phone, nickname) {
        return await this.post('/contacts', { phone, nickname });
    }

    /**
     * 删除联系人
     * @param {number} contactId - 联系人ID
     * @returns {Promise} 删除结果
     */
    async removeContact(contactId) {
        return await this.delete(`/contacts/${contactId}`);
    }

    // ==================== 聊天相关接口 ====================

    /**
     * 获取聊天列表
     * @returns {Promise} 聊天列表
     */
    async getChats() {
        return await this.get('/chats');
    }

    /**
     * 获取聊天消息
     * @param {number} chatId - 聊天ID
     * @param {Object} params - 查询参数
     * @returns {Promise} 消息列表
     */
    async getMessages(chatId, params = {}) {
        return await this.get(`/chats/${chatId}/messages`, params);
    }

    /**
     * 发送消息
     * @param {number} chatId - 聊天ID
     * @param {Object} messageData - 消息数据
     * @returns {Promise} 发送结果
     */
    async sendMessage(chatId, messageData) {
        return await this.post(`/chats/${chatId}/messages`, messageData);
    }

    /**
     * 编辑消息
     * @param {number} messageId - 消息ID
     * @param {string} content - 新内容
     * @returns {Promise} 编辑结果
     */
    async editMessage(messageId, content) {
        return await this.put(`/messages/${messageId}`, { content });
    }

    /**
     * 删除消息
     * @param {number} messageId - 消息ID
     * @returns {Promise} 删除结果
     */
    async deleteMessage(messageId) {
        return await this.delete(`/messages/${messageId}`);
    }

    /**
     * 标记消息为已读
     * @param {number} messageId - 消息ID
     * @returns {Promise} 标记结果
     */
    async markAsRead(messageId) {
        return await this.post(`/messages/${messageId}/read`);
    }

    /**
     * 置顶聊天
     * @param {number} chatId - 聊天ID
     * @returns {Promise} 置顶结果
     */
    async pinChat(chatId) {
        return await this.post(`/chats/${chatId}/pin`);
    }

    /**
     * 取消置顶聊天
     * @param {number} chatId - 聊天ID
     * @returns {Promise} 取消置顶结果
     */
    async unpinChat(chatId) {
        return await this.delete(`/chats/${chatId}/pin`);
    }
}

// 创建全局 API 实例
window.IMAPI = new IMAPI();

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = IMAPI;
}
