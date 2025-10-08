import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface MessageMarkManagerProps {
    userId: number;
    onMarkedMessagesUpdate: (messages: any[]) => void;
}

interface MarkedMessage {
    id: number;
    content: string;
    sender: {
        id: number;
        nickname: string;
    };
    mark_type: string;
    mark_time: string;
    created_at: string;
}

const MessageMarkManager: React.FC<MessageMarkManagerProps> = ({ userId, onMarkedMessagesUpdate }) => {
    const [markedMessages, setMarkedMessages] = useState<MarkedMessage[]>([]);
    const [selectedMarkType, setSelectedMarkType] = useState<string>('important');
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    const markTypes = [
        { value: 'important', label: '重要', icon: '⭐' },
        { value: 'favorite', label: '收藏', icon: '❤️' },
        { value: 'archive', label: '归档', icon: '📁' }
    ];

    // 获取标记消息列表
    const fetchMarkedMessages = async (markType: string) => {
        setIsLoading(true);
        setError(null);
        try {
            const response = await axios.get('/api/messages/marked', {
                params: { mark_type: markType },
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setMarkedMessages(response.data);
            onMarkedMessagesUpdate(response.data);
        } catch (err: any) {
            console.error('获取标记消息失败:', err);
            setError(err.response?.data?.error || '获取标记消息失败');
        } finally {
            setIsLoading(false);
        }
    };

    // 标记消息
    const markMessage = async (messageId: number, markType: string) => {
        try {
            await axios.post('/api/messages/mark', {
                message_id: messageId,
                user_id: userId,
                mark_type: markType
            }, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            // 重新获取标记消息列表
            fetchMarkedMessages(selectedMarkType);
        } catch (err: any) {
            console.error('标记消息失败:', err);
            setError(err.response?.data?.error || '标记消息失败');
        }
    };

    // 取消标记消息
    const unmarkMessage = async (messageId: number, markType: string) => {
        try {
            await axios.post(`/api/messages/${messageId}/unmark?mark_type=${markType}`, {}, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            // 重新获取标记消息列表
            fetchMarkedMessages(selectedMarkType);
        } catch (err: any) {
            console.error('取消标记失败:', err);
            setError(err.response?.data?.error || '取消标记失败');
        }
    };

    useEffect(() => {
        if (userId && selectedMarkType) {
            fetchMarkedMessages(selectedMarkType);
        }
    }, [userId, selectedMarkType]);

    const handleMarkTypeChange = (markType: string) => {
        setSelectedMarkType(markType);
    };

    const getMarkTypeInfo = (type: string) => {
        return markTypes.find(mt => mt.value === type) || { label: type, icon: '📌' };
    };

    if (isLoading) {
        return <div className="message-mark-manager loading">加载标记消息中...</div>;
    }

    return (
        <div className="message-mark-manager">
            <div className="mark-header">
                <h3>标记消息</h3>
                <div className="mark-type-tabs">
                    {markTypes.map((type) => (
                        <button
                            key={type.value}
                            className={`mark-type-tab ${selectedMarkType === type.value ? 'active' : ''}`}
                            onClick={() => handleMarkTypeChange(type.value)}
                        >
                            {type.icon} {type.label}
                        </button>
                    ))}
                </div>
                <button onClick={() => fetchMarkedMessages(selectedMarkType)} className="refresh-btn">
                    刷新
                </button>
            </div>
            
            {error && (
                <div className="error-message" style={{ color: 'red', marginBottom: '10px' }}>
                    {error}
                </div>
            )}

            <div className="marked-messages-list">
                {markedMessages.length === 0 ? (
                    <div className="no-marked-messages">
                        暂无{getMarkTypeInfo(selectedMarkType).label}消息
                    </div>
                ) : (
                    markedMessages.map((message) => (
                        <div key={message.id} className="marked-message-item">
                            <div className="message-content">
                                <div className="message-header">
                                    <span className="sender-name">{message.sender.nickname}</span>
                                    <span className="mark-type">
                                        {getMarkTypeInfo(message.mark_type).icon} {getMarkTypeInfo(message.mark_type).label}
                                    </span>
                                    <span className="mark-time">
                                        {new Date(message.mark_time).toLocaleString()}
                                    </span>
                                </div>
                                <div className="message-text">{message.content}</div>
                                <div className="original-time">
                                    原消息时间: {new Date(message.created_at).toLocaleString()}
                                </div>
                            </div>
                            <div className="message-actions">
                                <button 
                                    onClick={() => unmarkMessage(message.id, message.mark_type)}
                                    className="unmark-btn"
                                >
                                    取消标记
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default MessageMarkManager;
