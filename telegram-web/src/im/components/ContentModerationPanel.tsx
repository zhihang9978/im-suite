import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface ContentModerationPanelProps {
    isAdmin: boolean;
}

interface Report {
    id: number;
    content_type: string;
    content_id: number;
    content_text: string;
    content_user: {
        id: number;
        nickname: string;
    };
    reporter: {
        id: number;
        nickname: string;
    };
    report_reason: string;
    report_detail: string;
    auto_detected: boolean;
    detection_score: number;
    detection_keywords: string;
    status: string;
    priority: string;
    created_at: string;
}

const ContentModerationPanel: React.FC<ContentModerationPanelProps> = ({ isAdmin }) => {
    const [reports, setReports] = useState<Report[]>([]);
    const [selectedReport, setSelectedReport] = useState<Report | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const [priority Filter, setPriorityFilter] = useState<string>('');
    const [handleAction, setHandleAction] = useState<string>('warn');
    const [handleComment, setHandleComment] = useState<string>('');

    const priorityLabels = {
        'urgent': { label: '紧急', color: '#f44336', icon: '🚨' },
        'high': { label: '高', color: '#ff9800', icon: '⚠️' },
        'normal': { label: '普通', color: '#2196f3', icon: 'ℹ️' },
        'low': { label: '低', color: '#4caf50', icon: '✓' }
    };

    const reasonLabels = {
        'spam': '垃圾信息',
        'porn': '色情内容',
        'violence': '暴力内容',
        'politics': '政治敏感',
        'harassment': '骚扰辱骂',
        'fraud': '诈骗',
        'other': '其他'
    };

    // 获取待处理举报列表
    const fetchPendingReports = async () => {
        if (!isAdmin) return;

        setIsLoading(true);
        setError(null);
        try {
            const params = new URLSearchParams();
            if (priorityFilter) {
                params.append('priority', priorityFilter);
            }

            const response = await axios.get(`/api/moderation/reports/pending?${params.toString()}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setReports(response.data.reports || []);
        } catch (err: any) {
            console.error('获取举报列表失败:', err);
            setError(err.response?.data?.error || '获取举报列表失败');
        } finally {
            setIsLoading(false);
        }
    };

    // 处理举报
    const handleReport = async () => {
        if (!selectedReport) return;

        try {
            await axios.post('/api/moderation/reports/handle', {
                report_id: selectedReport.id,
                handle_action: handleAction,
                handle_comment: handleComment
            }, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });

            alert('举报处理成功');
            setSelectedReport(null);
            setHandleComment('');
            fetchPendingReports();
        } catch (err: any) {
            console.error('处理举报失败:', err);
            alert(err.response?.data?.error || '处理举报失败');
        }
    };

    useEffect(() => {
        fetchPendingReports();
    }, [priorityFilter]);

    if (!isAdmin) {
        return (
            <div className="content-moderation-panel">
                <p>您没有权限访问此页面</p>
            </div>
        );
    }

    return (
        <div className="content-moderation-panel">
            <div className="panel-header">
                <h2>📋 内容审核管理</h2>
                <div className="filter-controls">
                    <select 
                        value={priorityFilter}
                        onChange={(e) => setPriorityFilter(e.target.value)}
                    >
                        <option value="">全部优先级</option>
                        <option value="urgent">🚨 紧急</option>
                        <option value="high">⚠️ 高</option>
                        <option value="normal">ℹ️ 普通</option>
                        <option value="low">✓ 低</option>
                    </select>
                    <button onClick={fetchPendingReports} className="refresh-btn">
                        🔄 刷新
                    </button>
                </div>
            </div>

            {error && (
                <div className="error-message" style={{ color: 'red', padding: '10px' }}>
                    {error}
                </div>
            )}

            <div className="panel-content">
                <div className="reports-list">
                    <h3>待处理举报 ({reports.length})</h3>
                    {isLoading ? (
                        <p>加载中...</p>
                    ) : reports.length === 0 ? (
                        <p className="no-reports">暂无待处理举报</p>
                    ) : (
                        <div className="report-items">
                            {reports.map((report) => {
                                const priorityInfo = priorityLabels[report.priority] || priorityLabels['normal'];
                                return (
                                    <div 
                                        key={report.id} 
                                        className={`report-item ${selectedReport?.id === report.id ? 'selected' : ''}`}
                                        onClick={() => setSelectedReport(report)}
                                        style={{ borderLeft: `4px solid ${priorityInfo.color}` }}
                                    >
                                        <div className="report-header">
                                            <span className="priority-badge" style={{ background: priorityInfo.color }}>
                                                {priorityInfo.icon} {priorityInfo.label}
                                            </span>
                                            {report.auto_detected && (
                                                <span className="auto-badge">🤖 自动检测</span>
                                            )}
                                            <span className="report-reason">
                                                {reasonLabels[report.report_reason] || report.report_reason}
                                            </span>
                                        </div>
                                        <div className="report-content">
                                            <p className="content-text">{report.content_text.substring(0, 100)}...</p>
                                            {report.auto_detected && report.detection_keywords && (
                                                <p className="detection-info">
                                                    🔍 检测关键词: {report.detection_keywords}
                                                </p>
                                            )}
                                        </div>
                                        <div className="report-footer">
                                            <span>举报人: {report.reporter?.nickname || '系统'}</span>
                                            <span>涉事用户: {report.content_user.nickname}</span>
                                            <span>{new Date(report.created_at).toLocaleString()}</span>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>

                {selectedReport && (
                    <div className="report-detail">
                        <h3>举报详情</h3>
                        <div className="detail-content">
                            <div className="detail-section">
                                <h4>基本信息</h4>
                                <p><strong>举报ID:</strong> {selectedReport.id}</p>
                                <p><strong>内容类型:</strong> {selectedReport.content_type}</p>
                                <p><strong>内容ID:</strong> {selectedReport.content_id}</p>
                                <p><strong>举报原因:</strong> {reasonLabels[selectedReport.report_reason]}</p>
                                <p><strong>优先级:</strong> {priorityLabels[selectedReport.priority].label}</p>
                            </div>

                            <div className="detail-section">
                                <h4>内容详情</h4>
                                <div className="content-box">
                                    {selectedReport.content_text}
                                </div>
                            </div>

                            {selectedReport.report_detail && (
                                <div className="detail-section">
                                    <h4>举报说明</h4>
                                    <p>{selectedReport.report_detail}</p>
                                </div>
                            )}

                            {selectedReport.auto_detected && (
                                <div className="detail-section">
                                    <h4>自动检测信息</h4>
                                    <p><strong>检测类型:</strong> {selectedReport.detection_type || 'N/A'}</p>
                                    <p><strong>置信度:</strong> {(selectedReport.detection_score * 100).toFixed(1)}%</p>
                                    <p><strong>匹配关键词:</strong> {selectedReport.detection_keywords}</p>
                                </div>
                            )}

                            <div className="detail-section">
                                <h4>处理操作</h4>
                                <div className="handle-form">
                                    <div className="form-group">
                                        <label>处理动作:</label>
                                        <select 
                                            value={handleAction}
                                            onChange={(e) => setHandleAction(e.target.value)}
                                        >
                                            <option value="warn">⚠️ 发出警告（不拦截）</option>
                                            <option value="delete">🗑️ 标记删除（需手动执行）</option>
                                            <option value="ban">🚫 标记封禁（需手动执行）</option>
                                            <option value="ignore">✓ 忽略举报</option>
                                        </select>
                                    </div>

                                    <div className="form-group">
                                        <label>处理备注:</label>
                                        <textarea 
                                            value={handleComment}
                                            onChange={(e) => setHandleComment(e.target.value)}
                                            placeholder="请输入处理备注..."
                                            rows={4}
                                        />
                                    </div>

                                    <div className="action-buttons">
                                        <button onClick={handleReport} className="submit-btn">
                                            提交处理
                                        </button>
                                        <button onClick={() => setSelectedReport(null)} className="cancel-btn">
                                            取消
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            <style>{`
                .content-moderation-panel {
                    padding: 20px;
                    max-width: 1400px;
                    margin: 0 auto;
                }

                .panel-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 20px;
                }

                .filter-controls {
                    display: flex;
                    gap: 10px;
                }

                .panel-content {
                    display: grid;
                    grid-template-columns: 1fr 1fr;
                    gap: 20px;
                }

                .reports-list {
                    background: #f5f5f5;
                    padding: 15px;
                    border-radius: 8px;
                    max-height: 800px;
                    overflow-y: auto;
                }

                .report-item {
                    background: white;
                    padding: 15px;
                    margin-bottom: 10px;
                    border-radius: 6px;
                    cursor: pointer;
                    transition: all 0.2s;
                }

                .report-item:hover {
                    transform: translateX(5px);
                    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
                }

                .report-item.selected {
                    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
                }

                .priority-badge, .auto-badge {
                    display: inline-block;
                    padding: 2px 8px;
                    border-radius: 12px;
                    color: white;
                    font-size: 12px;
                    margin-right: 8px;
                }

                .auto-badge {
                    background: #9c27b0;
                }

                .report-detail {
                    background: white;
                    padding: 20px;
                    border-radius: 8px;
                    max-height: 800px;
                    overflow-y: auto;
                }

                .detail-section {
                    margin-bottom: 20px;
                    padding-bottom: 20px;
                    border-bottom: 1px solid #eee;
                }

                .content-box {
                    background: #f9f9f9;
                    padding: 15px;
                    border-radius: 4px;
                    white-space: pre-wrap;
                    word-break: break-word;
                }

                .handle-form .form-group {
                    margin-bottom: 15px;
                }

                .handle-form label {
                    display: block;
                    margin-bottom: 5px;
                    font-weight: bold;
                }

                .handle-form select,
                .handle-form textarea {
                    width: 100%;
                    padding: 8px;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }

                .action-buttons {
                    display: flex;
                    gap: 10px;
                }

                .submit-btn, .cancel-btn {
                    padding: 10px 20px;
                    border: none;
                    border-radius: 4px;
                    cursor: pointer;
                    font-weight: bold;
                }

                .submit-btn {
                    background: #4caf50;
                    color: white;
                }

                .cancel-btn {
                    background: #999;
                    color: white;
                }
            `}</style>
        </div>
    );
};

export default ContentModerationPanel;
