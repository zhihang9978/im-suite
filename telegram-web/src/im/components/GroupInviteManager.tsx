import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface GroupInviteManagerProps {
    chatId: number;
}

interface Invite {
    id: number;
    invite_code: string;
    invite_link: string;
    max_uses: number;
    used_count: number;
    expires_at: string | null;
    require_approval: boolean;
    is_enabled: boolean;
    is_revoked: boolean;
    creator: {
        id: number;
        nickname: string;
    };
    created_at: string;
}

const GroupInviteManager: React.FC<GroupInviteManagerProps> = ({ chatId }) => {
    const [invites, setInvites] = useState<Invite[]>([]);
    const [showCreateDialog, setShowCreateDialog] = useState<boolean>(false);
    const [createForm, setCreateForm] = useState({
        max_uses: 0,
        expires_at: '',
        require_approval: false
    });

    const fetchInvites = async () => {
        try {
            const response = await axios.get(`/api/groups/${chatId}/invites`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            setInvites(response.data.invites || []);
        } catch (err) {
            console.error('获取邀请列表失败:', err);
        }
    };

    const createInvite = async () => {
        try {
            await axios.post(`/api/groups/${chatId}/invites`, {
                chat_id: chatId,
                max_uses: createForm.max_uses,
                expires_at: createForm.expires_at || null,
                require_approval: createForm.require_approval
            }, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            setShowCreateDialog(false);
            fetchInvites();
        } catch (err: any) {
            alert(err.response?.data?.error || '创建邀请失败');
        }
    };

    const revokeInvite = async (inviteId: number) => {
        if (!confirm('确定要撤销此邀请吗？')) return;
        
        try {
            await axios.post(`/api/groups/invites/${inviteId}/revoke`, {
                reason: '管理员撤销'
            }, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            fetchInvites();
        } catch (err: any) {
            alert(err.response?.data?.error || '撤销失败');
        }
    };

    const copyInviteLink = (link: string) => {
        navigator.clipboard.writeText(link);
        alert('邀请链接已复制到剪贴板！');
    };

    useEffect(() => {
        fetchInvites();
    }, [chatId]);

    return (
        <div className="group-invite-manager">
            <div className="header">
                <h3>📨 邀请管理</h3>
                <button onClick={() => setShowCreateDialog(true)} className="create-btn">
                    + 创建邀请
                </button>
            </div>

            {showCreateDialog && (
                <div className="dialog-overlay">
                    <div className="dialog">
                        <h4>创建邀请链接</h4>
                        <div className="form">
                            <label>
                                最大使用次数 (0=无限制):
                                <input type="number" value={createForm.max_uses}
                                    onChange={(e) => setCreateForm({...createForm, max_uses: parseInt(e.target.value)})} />
                            </label>
                            <label>
                                过期时间:
                                <input type="datetime-local" value={createForm.expires_at}
                                    onChange={(e) => setCreateForm({...createForm, expires_at: e.target.value})} />
                            </label>
                            <label className="checkbox">
                                <input type="checkbox" checked={createForm.require_approval}
                                    onChange={(e) => setCreateForm({...createForm, require_approval: e.target.checked})} />
                                需要审批
                            </label>
                            <div className="actions">
                                <button onClick={createInvite}>创建</button>
                                <button onClick={() => setShowCreateDialog(false)}>取消</button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            <div className="invites-list">
                {invites.map((invite) => (
                    <div key={invite.id} className="invite-item">
                        <div className="invite-info">
                            <div className="invite-link">{invite.invite_link}</div>
                            <div className="invite-stats">
                                使用: {invite.used_count}/{invite.max_uses || '∞'} |
                                {invite.require_approval && ' 需审批 |'}
                                {invite.expires_at && ` 过期: ${new Date(invite.expires_at).toLocaleString()}`}
                            </div>
                            <div className="creator">创建者: {invite.creator.nickname}</div>
                        </div>
                        <div className="invite-actions">
                            <button onClick={() => copyInviteLink(invite.invite_link)}>📋 复制</button>
                            {!invite.is_revoked && (
                                <button onClick={() => revokeInvite(invite.id)} className="revoke-btn">撤销</button>
                            )}
                            {invite.is_revoked && <span className="revoked-badge">已撤销</span>}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default GroupInviteManager;
