/**
 * IM-Suite 数据映射层
 * 将 Telegram 的数据格式映射到我们的后端格式
 */

/**
 * 用户数据映射
 */
export class UserMapper {
    /**
     * 将 Telegram 用户数据映射到我们的用户格式
     */
    static fromTelegramUser(telegramUser: any): any {
        return {
            id: telegramUser.id || telegramUser.user_id,
            phone: telegramUser.phone,
            username: telegramUser.username,
            nickname: telegramUser.first_name || telegramUser.last_name || telegramUser.username,
            bio: telegramUser.about,
            avatar: telegramUser.photo ? this.getAvatarUrl(telegramUser.photo) : null,
            online: telegramUser.status === 'online',
            last_seen: telegramUser.status === 'online' ? new Date().toISOString() : null,
            language: 'zh-CN',
            theme: 'auto'
        };
    }

    /**
     * 将我们的用户数据映射到 Telegram 格式
     */
    static toTelegramUser(user: any): any {
        return {
            id: user.id,
            user_id: user.id,
            phone: user.phone,
            username: user.username,
            first_name: user.nickname,
            last_name: '',
            about: user.bio,
            photo: user.avatar ? { small_file_id: user.avatar } : null,
            status: user.online ? 'online' : 'offline',
            is_self: false,
            is_contact: true,
            is_mutual_contact: true,
            is_verified: false,
            is_premium: false,
            is_support: false,
            is_scam: false,
            is_fake: false,
            is_bot: false,
            restriction_reason: null,
            common_chats_count: 0
        };
    }

    /**
     * 获取头像URL
     */
    private static getAvatarUrl(photo: any): string {
        if (!photo) return '';
        
        // 简化处理，实际应该根据 Telegram 的规则生成URL
        if (photo.small_file_id) {
            return `/api/files/${photo.small_file_id}`;
        }
        return '';
    }
}

/**
 * 聊天数据映射
 */
export class ChatMapper {
    /**
     * 将 Telegram 聊天数据映射到我们的聊天格式
     */
    static fromTelegramChat(telegramChat: any): any {
        return {
            id: telegramChat.id,
            name: telegramChat.title || telegramChat.first_name || 'Unknown',
            description: telegramChat.about || '',
            avatar: telegramChat.photo ? this.getAvatarUrl(telegramChat.photo) : null,
            type: this.mapChatType(telegramChat.type),
            is_active: true,
            is_pinned: false,
            is_muted: telegramChat.muted || false,
            members_count: telegramChat.members_count || 0,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };
    }

    /**
     * 将我们的聊天数据映射到 Telegram 格式
     */
    static toTelegramChat(chat: any): any {
        return {
            id: chat.id,
            title: chat.name,
            about: chat.description,
            photo: chat.avatar ? { small_file_id: chat.avatar } : null,
            type: this.mapTelegramChatType(chat.type),
            members_count: chat.members_count || 0,
            muted: chat.is_muted,
            verified: false,
            scam: false,
            fake: false,
            gigagroup: false,
            megagroup: chat.type === 'group',
            broadcast: chat.type === 'channel',
            public: false,
            left: false,
            kicked: false,
            deactivated: false,
            call_active: false,
            call_not_empty: false,
            call_not_available: false,
            restriction_reason: null
        };
    }

    /**
     * 映射聊天类型
     */
    private static mapChatType(telegramType: string): string {
        switch (telegramType) {
            case 'private':
                return 'private';
            case 'group':
            case 'supergroup':
                return 'group';
            case 'channel':
                return 'channel';
            default:
                return 'private';
        }
    }

    /**
     * 映射到 Telegram 聊天类型
     */
    private static mapTelegramChatType(type: string): string {
        switch (type) {
            case 'private':
                return 'private';
            case 'group':
                return 'supergroup';
            case 'channel':
                return 'channel';
            default:
                return 'private';
        }
    }

    /**
     * 获取头像URL
     */
    private static getAvatarUrl(photo: any): string {
        if (!photo) return '';
        
        if (photo.small_file_id) {
            return `/api/files/${photo.small_file_id}`;
        }
        return '';
    }
}

/**
 * 消息数据映射
 */
export class MessageMapper {
    /**
     * 将 Telegram 消息数据映射到我们的消息格式
     */
    static fromTelegramMessage(telegramMessage: any): any {
        return {
            id: telegramMessage.id,
            chat_id: telegramMessage.peer_id?.channel_id || telegramMessage.peer_id?.chat_id || telegramMessage.peer_id?.user_id,
            sender_id: telegramMessage.from_id?.user_id || telegramMessage.from_id,
            content: this.getMessageContent(telegramMessage),
            type: this.mapMessageType(telegramMessage),
            file_name: this.getFileName(telegramMessage),
            file_size: this.getFileSize(telegramMessage),
            file_url: this.getFileUrl(telegramMessage),
            thumbnail: this.getThumbnail(telegramMessage),
            is_read: telegramMessage.out ? true : false,
            is_edited: telegramMessage.edit_date ? true : false,
            is_deleted: false,
            is_pinned: telegramMessage.pinned || false,
            reply_to_id: telegramMessage.reply_to?.reply_to_msg_id || null,
            forward_from: telegramMessage.fwd_from?.from_id || null,
            ttl: telegramMessage.ttl || 0,
            send_at: null,
            is_silent: telegramMessage.silent || false,
            created_at: new Date(telegramMessage.date * 1000).toISOString(),
            updated_at: new Date((telegramMessage.edit_date || telegramMessage.date) * 1000).toISOString()
        };
    }

    /**
     * 将我们的消息数据映射到 Telegram 格式
     */
    static toTelegramMessage(message: any): any {
        return {
            id: message.id,
            peer_id: this.getPeerId(message.chat_id, message.type),
            from_id: message.sender_id,
            message: message.content,
            date: Math.floor(new Date(message.created_at).getTime() / 1000),
            edit_date: message.is_edited ? Math.floor(new Date(message.updated_at).getTime() / 1000) : undefined,
            out: true,
            mentioned: false,
            media_unread: false,
            silent: message.is_silent,
            post: false,
            from_scheduled: false,
            legacy: false,
            edit_hide: false,
            pinned: message.is_pinned,
            noforwards: false,
            ttl: message.ttl || 0,
            reply_to: message.reply_to_id ? { reply_to_msg_id: message.reply_to_id } : undefined,
            fwd_from: message.forward_from ? { from_id: message.forward_from } : undefined,
            media: this.getTelegramMedia(message),
            reply_markup: null,
            entities: null,
            views: 0,
            forwards: 0,
            replies: null,
            reactions: null
        };
    }

    /**
     * 获取消息内容
     */
    private static getMessageContent(message: any): string {
        if (message.message) return message.message;
        if (message.media) {
            switch (message.media._) {
                case 'messageMediaPhoto':
                    return '[图片]';
                case 'messageMediaDocument':
                    return '[文件]';
                case 'messageMediaVideo':
                    return '[视频]';
                case 'messageMediaAudio':
                    return '[音频]';
                case 'messageMediaVoice':
                    return '[语音]';
                default:
                    return '[媒体]';
            }
        }
        return '';
    }

    /**
     * 映射消息类型
     */
    private static mapMessageType(message: any): string {
        if (message.media) {
            switch (message.media._) {
                case 'messageMediaPhoto':
                    return 'image';
                case 'messageMediaDocument':
                    return 'file';
                case 'messageMediaVideo':
                    return 'video';
                case 'messageMediaAudio':
                    return 'audio';
                case 'messageMediaVoice':
                    return 'voice';
                default:
                    return 'text';
            }
        }
        return 'text';
    }

    /**
     * 获取文件名
     */
    private static getFileName(message: any): string {
        if (message.media && message.media.document) {
            return message.media.document.attributes?.find(attr => attr._ === 'documentAttributeFilename')?.file_name || '';
        }
        return '';
    }

    /**
     * 获取文件大小
     */
    private static getFileSize(message: any): number {
        if (message.media && message.media.document) {
            return message.media.document.size || 0;
        }
        return 0;
    }

    /**
     * 获取文件URL
     */
    private static getFileUrl(message: any): string {
        if (message.media && message.media.document) {
            return `/api/files/${message.media.document.id}`;
        }
        return '';
    }

    /**
     * 获取缩略图
     */
    private static getThumbnail(message: any): string {
        if (message.media && message.media.photo) {
            return `/api/files/${message.media.photo.id}`;
        }
        return '';
    }

    /**
     * 获取 Peer ID
     */
    private static getPeerId(chatId: number, type: string): any {
        switch (type) {
            case 'private':
                return { user_id: chatId };
            case 'group':
                return { chat_id: chatId };
            case 'channel':
                return { channel_id: chatId };
            default:
                return { user_id: chatId };
        }
    }

    /**
     * 获取 Telegram 媒体对象
     */
    private static getTelegramMedia(message: any): any {
        switch (message.type) {
            case 'image':
                return {
                    _: 'messageMediaPhoto',
                    photo: { id: message.file_url },
                    ttl_seconds: message.ttl || 0
                };
            case 'video':
                return {
                    _: 'messageMediaDocument',
                    document: {
                        id: message.file_url,
                        size: message.file_size,
                        mime_type: 'video/mp4'
                    }
                };
            case 'file':
                return {
                    _: 'messageMediaDocument',
                    document: {
                        id: message.file_url,
                        size: message.file_size,
                        mime_type: 'application/octet-stream'
                    }
                };
            case 'audio':
                return {
                    _: 'messageMediaAudio',
                    audio: {
                        id: message.file_url,
                        size: message.file_size
                    }
                };
            case 'voice':
                return {
                    _: 'messageMediaVoice',
                    voice: {
                        id: message.file_url,
                        size: message.file_size
                    }
                };
            default:
                return null;
        }
    }
}

/**
 * 联系人数据映射
 */
export class ContactMapper {
    /**
     * 将 Telegram 联系人数据映射到我们的联系人格式
     */
    static fromTelegramContact(telegramContact: any): any {
        return {
            id: telegramContact.user_id,
            user_id: 0, // 当前用户ID，需要从上下文获取
            contact_id: telegramContact.user_id,
            nickname: telegramContact.first_name || telegramContact.last_name || '',
            is_blocked: false,
            is_muted: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
        };
    }

    /**
     * 将我们的联系人数据映射到 Telegram 格式
     */
    static toTelegramContact(contact: any): any {
        return {
            user_id: contact.contact_id,
            first_name: contact.nickname,
            last_name: '',
            phone: contact.phone || '',
            mutual_contact: true,
            deleted: false
        };
    }
}

// 导出映射器
export default {
    UserMapper,
    ChatMapper,
    MessageMapper,
    ContactMapper
};


