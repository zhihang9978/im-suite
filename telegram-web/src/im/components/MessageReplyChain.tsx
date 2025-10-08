import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface MessageReplyChainProps {
    messageId: number;
    onReplyChainUpdate: (messages: any[]) => void;
}

interface ReplyMessage {
    id: number;
    content: string;
    sender: {
        id: number;
        nickname: string;
        avatar?: string;
    };
    created_at: string;
    reply_to?: ReplyMessage;
}

const MessageReplyChain: React.FC<MessageReplyChainProps> = ({ messageId, onReplyChainUpdate }) => {
    const [replyChain, setReplyChain] = useState<ReplyMessage[]>([]);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    // 获取回复链
    const fetchReplyChain = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const response = await axios.get(`/api/messages/${messageId}/reply-chain`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setReplyChain(response.data);
            onReplyChainUpdate(response.data);
        } catch (err: any) {
            console.error('获取回复链失败:', err);
            setError(err.response?.data?.error || '获取回复链失败');
        } finally {
            setIsLoading(false);
        }
    };

    // 回复消息
    const replyToMessage = async (replyToId: number, content: string, messageType: string = 'text') => {
        try {
            const response = await axios.post('/api/messages/reply', {
                message_id: 0, // 新消息ID将由后端生成
                reply_to_id: replyToId,
                user_id: parseInt(localStorage.getItem('userId') || '0'),
                content: content,
                message_type: messageType
            }, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            
            // 重新获取回复链
            fetchReplyChain();
            return response.data;
        } catch (err: any) {
            console.error('回复消息失败:', err);
            setError(err.response?.data?.error || '回复消息失败');
            throw err;
        }
    };

    useEffect(() => {
        if (messageId) {
            fetchReplyChain();
        }
    }, [messageId]);

    const formatTime = (timestamp: string) => {
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now.getTime() - date.getTime();
        
        if (diff < 60000) { // 1分钟内
            return '刚刚';
        } else if (diff < 3600000) { // 1小时内
            return `${Math.floor(diff / 60000)}分钟前`;
        } else if (diff < 86400000) { // 24小时内
            return `${Math.floor(diff / 3600000)}小时前`;
        } else {
            return date.toLocaleDateString();
        }
    };

    if (isLoading) {
        return <div className="message-reply-chain loading">加载回复链中...</div>;
    }

    return (
        <div className="message-reply-chain">
            <div className="reply-chain-header">
                <h3>回复链 ({replyChain.length})</h3>
                <button onClick={fetchReplyChain} className="refresh-btn">刷新</button>
            </div>
            
            {error && (
                <div className="error-message" style={{ color: 'red', marginBottom: '10px' }}>
                    {error}
                </div>
            )}

            <div className="reply-chain-list">
                {replyChain.length === 0 ? (
                    <div className="no-reply-chain">暂无回复链</div>
                ) : (
                    replyChain.map((message, index) => (
                        <div key={message.id} className={`reply-chain-item ${index === replyChain.length - 1 ? 'latest' : ''}`}>
                            <div className="reply-chain-indicator">
                                {index < replyChain.length - 1 && <div className="reply-line"></div>}
                                <div className="reply-dot"></div>
                            </div>
                            <div className="reply-message-content">
                                <div className="message-header">
                                    <div className="sender-info">
                                        {message.sender.avatar && (
                                            <img 
                                                src={message.sender.avatar} 
                                                alt={message.sender.nickname}
                                                className="sender-avatar"
                                            />
                                        )}
                                        <span className="sender-name">{message.sender.nickname}</span>
                                    </div>
                                    <span className="message-time">{formatTime(message.created_at)}</span>
                                </div>
                                <div className="message-text">{message.content}</div>
                                
                                {/* 显示被回复的消息 */}
                                {message.reply_to && (
                                    <div className="quoted-message">
                                        <div className="quoted-header">
                                            <span className="quoted-label">回复:</span>
                                            <span className="quoted-sender">{message.reply_to.sender.nickname}</span>
                                        </div>
                                        <div className="quoted-content">{message.reply_to.content}</div>
                                    </div>
                                )}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default MessageReplyChain;
