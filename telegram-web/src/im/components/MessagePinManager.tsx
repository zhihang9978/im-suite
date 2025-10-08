import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface MessagePinManagerProps {
    chatId: number;
    onPinnedMessagesUpdate: (messages: any[]) => void;
}

interface PinnedMessage {
    id: number;
    content: string;
    sender: {
        id: number;
        nickname: string;
    };
    pin_time: string;
    created_at: string;
}

const MessagePinManager: React.FC<MessagePinManagerProps> = ({ chatId, onPinnedMessagesUpdate }) => {
    const [pinnedMessages, setPinnedMessages] = useState<PinnedMessage[]>([]);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    // 获取置顶消息列表
    const fetchPinnedMessages = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const response = await axios.get('/api/messages/pinned', {
                params: { chat_id: chatId },
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setPinnedMessages(response.data);
            onPinnedMessagesUpdate(response.data);
        } catch (err: any) {
            console.error('获取置顶消息失败:', err);
            setError(err.response?.data?.error || '获取置顶消息失败');
        } finally {
            setIsLoading(false);
        }
    };

    // 置顶消息
    const pinMessage = async (messageId: number, reason?: string) => {
        try {
            await axios.post('/api/messages/pin', {
                message_id: messageId,
                user_id: parseInt(localStorage.getItem('userId') || '0'),
                reason: reason
            }, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            // 重新获取置顶消息列表
            fetchPinnedMessages();
        } catch (err: any) {
            console.error('置顶消息失败:', err);
            setError(err.response?.data?.error || '置顶消息失败');
        }
    };

    // 取消置顶消息
    const unpinMessage = async (messageId: number) => {
        try {
            await axios.post(`/api/messages/${messageId}/unpin`, {}, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            // 重新获取置顶消息列表
            fetchPinnedMessages();
        } catch (err: any) {
            console.error('取消置顶失败:', err);
            setError(err.response?.data?.error || '取消置顶失败');
        }
    };

    useEffect(() => {
        if (chatId) {
            fetchPinnedMessages();
        }
    }, [chatId]);

    if (isLoading) {
        return <div className="message-pin-manager loading">加载置顶消息中...</div>;
    }

    return (
        <div className="message-pin-manager">
            <div className="pin-header">
                <h3>置顶消息 ({pinnedMessages.length})</h3>
                <button onClick={fetchPinnedMessages} className="refresh-btn">刷新</button>
            </div>
            
            {error && (
                <div className="error-message" style={{ color: 'red', marginBottom: '10px' }}>
                    {error}
                </div>
            )}

            <div className="pinned-messages-list">
                {pinnedMessages.length === 0 ? (
                    <div className="no-pinned-messages">暂无置顶消息</div>
                ) : (
                    pinnedMessages.map((message) => (
                        <div key={message.id} className="pinned-message-item">
                            <div className="message-content">
                                <div className="message-header">
                                    <span className="sender-name">{message.sender.nickname}</span>
                                    <span className="pin-time">
                                        {new Date(message.pin_time).toLocaleString()}
                                    </span>
                                </div>
                                <div className="message-text">{message.content}</div>
                                <div className="original-time">
                                    原消息时间: {new Date(message.created_at).toLocaleString()}
                                </div>
                            </div>
                            <div className="message-actions">
                                <button 
                                    onClick={() => unpinMessage(message.id)}
                                    className="unpin-btn"
                                >
                                    取消置顶
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default MessagePinManager;
