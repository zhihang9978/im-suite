import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface GroupJoinRequestManagerProps {
    chatId: number;
}

interface JoinRequest {
    id: number;
    user: {
        id: number;
        nickname: string;
        avatar?: string;
        email?: string;
    };
    message: string;
    status: string;
    created_at: string;
}

const GroupJoinRequestManager: React.FC<GroupJoinRequestManagerProps> = ({ chatId }) => {
    const [requests, setRequests] = useState<JoinRequest[]>([]);
    const [selectedRequest, setSelectedRequest] = useState<JoinRequest | null>(null);
    const [reviewNote, setReviewNote] = useState<string>('');

    const fetchRequests = async () => {
        try {
            const response = await axios.get(`/api/groups/${chatId}/join-requests`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            setRequests(response.data.requests || []);
        } catch (err) {
            console.error('获取入群申请失败:', err);
        }
    };

    const handleRequest = async (requestId: number, approved: boolean) => {
        try {
            await axios.post('/api/groups/join-requests/approve', {
                request_id: requestId,
                approved: approved,
                review_note: reviewNote
            }, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            setSelectedRequest(null);
            setReviewNote('');
            fetchRequests();
        } catch (err: any) {
            alert(err.response?.data?.error || '处理申请失败');
        }
    };

    useEffect(() => {
        fetchRequests();
    }, [chatId]);

    return (
        <div className="group-join-request-manager">
            <div className="header">
                <h3>🔍 入群审核 ({requests.length})</h3>
                <button onClick={fetchRequests} className="refresh-btn">刷新</button>
            </div>

            <div className="requests-list">
                {requests.length === 0 ? (
                    <div className="no-requests">暂无待审批的入群申请</div>
                ) : (
                    requests.map((request) => (
                        <div key={request.id} className="request-item">
                            <div className="request-info">
                                {request.user.avatar && (
                                    <img src={request.user.avatar} alt={request.user.nickname} className="avatar" />
                                )}
                                <div className="info-text">
                                    <div className="user-name">{request.user.nickname}</div>
                                    {request.user.email && <div className="user-email">{request.user.email}</div>}
                                    <div className="message">{request.message || '无申请消息'}</div>
                                    <div className="time">{new Date(request.created_at).toLocaleString()}</div>
                                </div>
                            </div>
                            <div className="request-actions">
                                <button 
                                    onClick={() => {
                                        setSelectedRequest(request);
                                    }}
                                    className="review-btn"
                                >
                                    审核
                                </button>
                            </div>

                            {selectedRequest?.id === request.id && (
                                <div className="review-panel">
                                    <textarea
                                        placeholder="审核备注（可选）"
                                        value={reviewNote}
                                        onChange={(e) => setReviewNote(e.target.value)}
                                        rows={3}
                                    />
                                    <div className="review-actions">
                                        <button 
                                            onClick={() => handleRequest(request.id, true)}
                                            className="approve-btn"
                                        >
                                            ✓ 通过
                                        </button>
                                        <button 
                                            onClick={() => handleRequest(request.id, false)}
                                            className="reject-btn"
                                        >
                                            ✗ 拒绝
                                        </button>
                                        <button 
                                            onClick={() => setSelectedRequest(null)}
                                            className="cancel-btn"
                                        >
                                            取消
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    ))
                )}
            </div>

            <style>{`
                .group-join-request-manager {
                    padding: 20px;
                }
                .header {
                    display: flex;
                    justify-content: space-between;
                    margin-bottom: 20px;
                }
                .request-item {
                    background: white;
                    padding: 15px;
                    margin-bottom: 10px;
                    border-radius: 8px;
                    border: 1px solid #e0e0e0;
                }
                .request-info {
                    display: flex;
                    gap: 15px;
                    margin-bottom: 10px;
                }
                .avatar {
                    width: 50px;
                    height: 50px;
                    border-radius: 50%;
                }
                .review-panel {
                    margin-top: 15px;
                    padding-top: 15px;
                    border-top: 1px solid #e0e0e0;
                }
                .review-actions {
                    display: flex;
                    gap: 10px;
                    margin-top: 10px;
                }
                .approve-btn { background: #4caf50; color: white; padding: 8px 20px; border: none; border-radius: 4px; }
                .reject-btn { background: #f44336; color: white; padding: 8px 20px; border: none; border-radius: 4px; }
                .cancel-btn { background: #999; color: white; padding: 8px 20px; border: none; border-radius: 4px; }
            `}</style>
        </div>
    );
};

export default GroupJoinRequestManager;
