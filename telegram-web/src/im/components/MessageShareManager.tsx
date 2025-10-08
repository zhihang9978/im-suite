import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface MessageShareManagerProps {
    messageId: number;
    onShareHistoryUpdate: (shares: any[]) => void;
}

interface ShareRecord {
    id: number;
    message_id: number;
    share_user: {
        id: number;
        nickname: string;
    };
    shared_to_user?: {
        id: number;
        nickname: string;
    };
    shared_to_chat?: {
        id: number;
        name: string;
    };
    share_type: string;
    share_time: string;
    share_data: string;
}

const MessageShareManager: React.FC<MessageShareManagerProps> = ({ messageId, onShareHistoryUpdate }) => {
    const [shareHistory, setShareHistory] = useState<ShareRecord[]>([]);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const [showShareDialog, setShowShareDialog] = useState<boolean>(false);
    const [shareForm, setShareForm] = useState({
        sharedTo: '',
        sharedToChatId: '',
        shareType: 'copy',
        shareData: ''
    });

    const shareTypes = [
        { value: 'copy', label: '复制链接', icon: '📋' },
        { value: 'forward', label: '转发', icon: '↗️' },
        { value: 'link', label: '生成链接', icon: '🔗' }
    ];

    // 获取分享历史
    const fetchShareHistory = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const response = await axios.get(`/api/messages/${messageId}/share-history`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setShareHistory(response.data);
            onShareHistoryUpdate(response.data);
        } catch (err: any) {
            console.error('获取分享历史失败:', err);
            setError(err.response?.data?.error || '获取分享历史失败');
        } finally {
            setIsLoading(false);
        }
    };

    // 分享消息
    const shareMessage = async () => {
        try {
            const shareData = {
                message_id: messageId,
                user_id: parseInt(localStorage.getItem('userId') || '0'),
                share_type: shareForm.shareType,
                share_data: shareForm.shareData
            };

            // 根据分享类型添加不同的目标
            if (shareForm.shareType === 'forward') {
                if (shareForm.sharedTo) {
                    shareData.shared_to = parseInt(shareForm.sharedTo);
                }
                if (shareForm.sharedToChatId) {
                    shareData.shared_to_chat_id = parseInt(shareForm.sharedToChatId);
                }
            }

            await axios.post('/api/messages/share', shareData, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });

            // 重新获取分享历史
            fetchShareHistory();
            setShowShareDialog(false);
            
            // 重置表单
            setShareForm({
                sharedTo: '',
                sharedToChatId: '',
                shareType: 'copy',
                shareData: ''
            });
        } catch (err: any) {
            console.error('分享消息失败:', err);
            setError(err.response?.data?.error || '分享消息失败');
        }
    };

    useEffect(() => {
        if (messageId) {
            fetchShareHistory();
        }
    }, [messageId]);

    const getShareTypeInfo = (type: string) => {
        return shareTypes.find(st => st.value === type) || { label: type, icon: '📤' };
    };

    const formatTime = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    if (isLoading) {
        return <div className="message-share-manager loading">加载分享历史中...</div>;
    }

    return (
        <div className="message-share-manager">
            <div className="share-header">
                <h3>消息分享</h3>
                <button onClick={() => setShowShareDialog(true)} className="share-btn">
                    📤 分享消息
                </button>
                <button onClick={fetchShareHistory} className="refresh-btn">刷新</button>
            </div>
            
            {error && (
                <div className="error-message" style={{ color: 'red', marginBottom: '10px' }}>
                    {error}
                </div>
            )}

            {/* 分享对话框 */}
            {showShareDialog && (
                <div className="share-dialog-overlay">
                    <div className="share-dialog">
                        <div className="dialog-header">
                            <h4>分享消息</h4>
                            <button onClick={() => setShowShareDialog(false)} className="close-btn">×</button>
                        </div>
                        
                        <div className="dialog-content">
                            <div className="form-group">
                                <label>分享类型:</label>
                                <select 
                                    value={shareForm.shareType}
                                    onChange={(e) => setShareForm({...shareForm, shareType: e.target.value})}
                                >
                                    {shareTypes.map(type => (
                                        <option key={type.value} value={type.value}>
                                            {type.icon} {type.label}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            {shareForm.shareType === 'forward' && (
                                <>
                                    <div className="form-group">
                                        <label>分享给用户ID:</label>
                                        <input 
                                            type="number"
                                            value={shareForm.sharedTo}
                                            onChange={(e) => setShareForm({...shareForm, sharedTo: e.target.value})}
                                            placeholder="输入用户ID"
                                        />
                                    </div>
                                    
                                    <div className="form-group">
                                        <label>分享到群聊ID:</label>
                                        <input 
                                            type="number"
                                            value={shareForm.sharedToChatId}
                                            onChange={(e) => setShareForm({...shareForm, sharedToChatId: e.target.value})}
                                            placeholder="输入群聊ID"
                                        />
                                    </div>
                                </>
                            )}

                            <div className="form-group">
                                <label>备注信息:</label>
                                <textarea 
                                    value={shareForm.shareData}
                                    onChange={(e) => setShareForm({...shareForm, shareData: e.target.value})}
                                    placeholder="可选的备注信息"
                                    rows={3}
                                />
                            </div>
                        </div>
                        
                        <div className="dialog-actions">
                            <button onClick={() => setShowShareDialog(false)} className="cancel-btn">
                                取消
                            </button>
                            <button onClick={shareMessage} className="confirm-btn">
                                分享
                            </button>
                        </div>
                    </div>
                </div>
            )}

            <div className="share-history-list">
                {shareHistory.length === 0 ? (
                    <div className="no-share-history">暂无分享记录</div>
                ) : (
                    shareHistory.map((share) => (
                        <div key={share.id} className="share-history-item">
                            <div className="share-info">
                                <div className="share-header-info">
                                    <span className="share-type">
                                        {getShareTypeInfo(share.share_type).icon} {getShareTypeInfo(share.share_type).label}
                                    </span>
                                    <span className="share-time">{formatTime(share.share_time)}</span>
                                </div>
                                
                                <div className="share-details">
                                    <div className="share-user">
                                        分享者: {share.share_user.nickname}
                                    </div>
                                    
                                    {share.shared_to_user && (
                                        <div className="share-target">
                                            分享给: {share.shared_to_user.nickname}
                                        </div>
                                    )}
                                    
                                    {share.shared_to_chat && (
                                        <div className="share-target">
                                            分享到群聊: {share.shared_to_chat.name}
                                        </div>
                                    )}
                                    
                                    {share.share_data && (
                                        <div className="share-data">
                                            备注: {share.share_data}
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default MessageShareManager;
