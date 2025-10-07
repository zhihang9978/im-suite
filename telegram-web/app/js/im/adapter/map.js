/**
 * IM-Suite 映射表
 * 将 Telegram 原有的调用点映射到我们的后端接口
 * 保持原有 UI 不变，只替换数据来源
 */

/**
 * Telegram API 到 IM-Suite API 的映射表
 * 这个映射表定义了如何将原有的 Telegram 调用转换为我们的 REST API 调用
 */
const TelegramToIMSuiteMap = {
    // ==================== 认证相关映射 ====================
    
    /**
     * 用户登录映射
     * 原: Telegram.auth.sendCode() -> 新: IMAPI.login()
     */
    auth: {
        sendCode: async (phone) => {
            // 发送验证码请求（这里简化为直接返回成功）
            console.log('发送验证码到:', phone);
            return { phone_code_hash: 'mock_hash_' + Date.now() };
        },
        
        signIn: async (phone, phone_code, phone_code_hash) => {
            // 使用验证码登录
            return await window.IMAPI.login(phone, phone_code);
        },
        
        signUp: async (phone, phone_code_hash, first_name, last_name) => {
            // 用户注册（仅后台创建，前端不支持）
            throw new Error('用户注册仅支持后台创建');
        },
        
        logOut: async () => {
            return await window.IMAPI.logout();
        }
    },

    // ==================== 用户相关映射 ====================
    
    /**
     * 用户信息映射
     * 原: Telegram.users.getFullUser() -> 新: IMAPI.getCurrentUser()
     */
    users: {
        getFullUser: async (id) => {
            if (id === 'self') {
                return await window.IMAPI.getCurrentUser();
            }
            // 获取其他用户信息
            return await window.IMAPI.get(`/users/${id}`);
        },
        
        updateProfile: async (first_name, last_name, about) => {
            return await window.IMAPI.updateUser({
                nickname: `${first_name} ${last_name}`.trim(),
                bio: about
            });
        },
        
        updateStatus: async (offline) => {
            return await window.IMSocket.sendPresence(!offline);
        }
    },

    // ==================== 联系人相关映射 ====================
    
    /**
     * 联系人映射
     * 原: Telegram.contacts.getContacts() -> 新: IMAPI.getContacts()
     */
    contacts: {
        getContacts: async () => {
            return await window.IMAPI.getContacts();
        },
        
        addContact: async (phone, first_name, last_name) => {
            return await window.IMAPI.addContact(phone, `${first_name} ${last_name}`.trim());
        },
        
        deleteContact: async (id) => {
            return await window.IMAPI.removeContact(id);
        },
        
        search: async (q, limit) => {
            return await window.IMAPI.get('/contacts', { q, limit });
        }
    },

    // ==================== 聊天相关映射 ====================
    
    /**
     * 聊天映射
     * 原: Telegram.messages.getDialogs() -> 新: IMAPI.getChats()
     */
    messages: {
        getDialogs: async (offset_date, offset_id, offset_peer, limit) => {
            return await window.IMAPI.getChats();
        },
        
        getHistory: async (peer, offset_id, offset_date, add_offset, limit, max_id, min_id) => {
            const chatId = peer.user_id || peer.chat_id || peer.channel_id;
            return await window.IMAPI.getMessages(chatId, {
                offset_id,
                offset_date,
                limit
            });
        },
        
        sendMessage: async (peer, message, random_id, reply_to_msg_id, reply_markup) => {
            const chatId = peer.user_id || peer.chat_id || peer.channel_id;
            return await window.IMAPI.sendMessage(chatId, {
                content: message,
                reply_to: reply_to_msg_id,
                random_id
            });
        },
        
        editMessage: async (peer, id, message, reply_markup) => {
            return await window.IMAPI.editMessage(id, message);
        },
        
        deleteMessages: async (peer, id, revoke) => {
            return await window.IMAPI.deleteMessage(id);
        },
        
        markAsRead: async (peer, max_id) => {
            const chatId = peer.user_id || peer.chat_id || peer.channel_id;
            return await window.IMAPI.markAsRead(max_id);
        },
        
        setTyping: async (peer, action) => {
            const chatId = peer.user_id || peer.chat_id || peer.channel_id;
            const isTyping = action._ === 'sendMessageTypingAction';
            return await window.IMSocket.sendTyping(chatId, isTyping);
        },
        
        search: async (peer, q, filter, min_date, max_date, offset_id, add_offset, limit) => {
            const chatId = peer.user_id || peer.chat_id || peer.channel_id;
            return await window.IMAPI.getMessages(chatId, {
                q,
                from: min_date,
                to: max_date,
                limit
            });
        }
    },

    // ==================== 媒体相关映射 ====================
    
    /**
     * 媒体上传映射
     * 原: Telegram.upload.getFile() -> 新: 直接上传到 MinIO
     */
    upload: {
        getFile: async (location, offset, limit) => {
            // 从我们的存储服务获取文件
            const fileId = location.volume_id + '_' + location.local_id;
            return await window.IMAPI.get(`/files/${fileId}`);
        },
        
        saveFilePart: async (file_id, file_part, bytes) => {
            // 上传文件到我们的存储服务
            const formData = new FormData();
            formData.append('file', bytes);
            formData.append('file_id', file_id);
            formData.append('part', file_part);
            
            return await window.IMAPI.post('/upload', formData);
        }
    },

    // ==================== 通话相关映射 ====================
    
    /**
     * 通话映射
     * 原: Telegram.phone.* -> 新: WebSocket 信令
     */
    phone: {
        requestCall: async (user_id, g_a_hash, protocol) => {
            // 发起通话邀请
            return await window.IMSocket.sendCallOffer({
                user_id,
                g_a_hash,
                protocol
            });
        },
        
        acceptCall: async (peer, g_b, protocol) => {
            // 接受通话
            return await window.IMSocket.sendCallAnswer({
                peer,
                g_b,
                protocol
            });
        },
        
        discardCall: async (peer, duration, reason, g_a_hash) => {
            // 结束通话
            return await window.IMSocket.sendCallEnd({
                peer,
                duration,
                reason,
                g_a_hash
            });
        }
    }
};

/**
 * 消息类型映射
 * 将 Telegram 消息类型映射到我们的消息类型
 */
const MessageTypeMap = {
    'messageEmpty': 'text',
    'message': 'text',
    'messageService': 'system',
    'messageMediaPhoto': 'image',
    'messageMediaDocument': 'file',
    'messageMediaVideo': 'video',
    'messageMediaAudio': 'audio',
    'messageMediaVoice': 'voice',
    'messageMediaSticker': 'sticker',
    'messageMediaGif': 'gif'
};

/**
 * 聊天类型映射
 * 将 Telegram 聊天类型映射到我们的聊天类型
 */
const ChatTypeMap = {
    'user': 'private',
    'chat': 'group',
    'channel': 'channel'
};

/**
 * 状态映射
 * 将 Telegram 用户状态映射到我们的状态
 */
const StatusMap = {
    'userStatusOnline': 'online',
    'userStatusOffline': 'offline',
    'userStatusLastWeek': 'offline',
    'userStatusLastMonth': 'offline'
};

/**
 * 创建适配器函数
 * 这个函数将原有的 Telegram API 调用重定向到我们的映射表
 */
function createAdapter() {
    return {
        // 认证相关
        auth: TelegramToIMSuiteMap.auth,
        
        // 用户相关
        users: TelegramToIMSuiteMap.users,
        
        // 联系人相关
        contacts: TelegramToIMSuiteMap.contacts,
        
        // 消息相关
        messages: TelegramToIMSuiteMap.messages,
        
        // 媒体相关
        upload: TelegramToIMSuiteMap.upload,
        
        // 通话相关
        phone: TelegramToIMSuiteMap.phone,
        
        // 工具函数
        utils: {
            mapMessageType: (telegramType) => MessageTypeMap[telegramType] || 'text',
            mapChatType: (telegramType) => ChatTypeMap[telegramType] || 'private',
            mapStatus: (telegramStatus) => StatusMap[telegramStatus] || 'offline'
        }
    };
}

// 创建全局适配器
window.TelegramAdapter = createAdapter();

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        TelegramToIMSuiteMap,
        MessageTypeMap,
        ChatTypeMap,
        StatusMap,
        createAdapter
    };
}
