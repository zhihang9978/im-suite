import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface GroupAdminManagerProps {
    chatId: number;
}

interface Admin {
    id: number;
    user: {
        id: number;
        nickname: string;
        avatar?: string;
    };
    role: {
        id: number;
        name: string;
        display_name: string;
        level: number;
    };
    title: string;
    promoted_by_user: {
        nickname: string;
    };
    created_at: string;
}

const GroupAdminManager: React.FC<GroupAdminManagerProps> = ({ chatId }) => {
    const [admins, setAdmins] = useState<Admin[]>([]);
    const [showPromoteDialog, setShowPromoteDialog] = useState<boolean>(false);
    const [promoteForm, setPromoteForm] = useState({
        user_id: 0,
        role_id: 0,
        title: ''
    });

    const roleOptions = [
        { id: 1, name: 'owner', display_name: '群主', level: 100 },
        { id: 2, name: 'admin', display_name: '管理员', level: 80 },
        { id: 3, name: 'moderator', display_name: '协管员', level: 50 }
    ];

    const fetchAdmins = async () => {
        try {
            const response = await axios.get(`/api/groups/${chatId}/admins`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            setAdmins(response.data.admins || []);
        } catch (err) {
            console.error('获取管理员列表失败:', err);
        }
    };

    const promoteMember = async () => {
        try {
            await axios.post('/api/groups/admins/promote', {
                chat_id: chatId,
                user_id: promoteForm.user_id,
                role_id: promoteForm.role_id,
                title: promoteForm.title
            }, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            setShowPromoteDialog(false);
            fetchAdmins();
        } catch (err: any) {
            alert(err.response?.data?.error || '提升管理员失败');
        }
    };

    const demoteMember = async (userId: number) => {
        if (!confirm('确定要降级此管理员吗？')) return;
        
        try {
            await axios.post(`/api/groups/${chatId}/admins/${userId}/demote`, {}, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('accessToken')}` }
            });
            fetchAdmins();
        } catch (err: any) {
            alert(err.response?.data?.error || '降级失败');
        }
    };

    useEffect(() => {
        fetchAdmins();
    }, [chatId]);

    const getRoleBadgeColor = (level: number) => {
        if (level >= 100) return '#f44336'; // 红色-群主
        if (level >= 80) return '#ff9800';  // 橙色-管理员
        return '#2196f3';                   // 蓝色-协管员
    };

    return (
        <div className="group-admin-manager">
            <div className="header">
                <h3>👥 管理员管理</h3>
                <button onClick={() => setShowPromoteDialog(true)} className="promote-btn">
                    + 提升管理员
                </button>
            </div>

            {showPromoteDialog && (
                <div className="dialog-overlay">
                    <div className="dialog">
                        <h4>提升管理员</h4>
                        <div className="form">
                            <label>
                                用户ID:
                                <input type="number" value={promoteForm.user_id}
                                    onChange={(e) => setPromoteForm({...promoteForm, user_id: parseInt(e.target.value)})} />
                            </label>
                            <label>
                                管理员角色:
                                <select value={promoteForm.role_id}
                                    onChange={(e) => setPromoteForm({...promoteForm, role_id: parseInt(e.target.value)})}>
                                    <option value={0}>选择角色</option>
                                    {roleOptions.map(role => (
                                        <option key={role.id} value={role.id}>{role.display_name}</option>
                                    ))}
                                </select>
                            </label>
                            <label>
                                自定义头衔（可选）:
                                <input type="text" value={promoteForm.title}
                                    onChange={(e) => setPromoteForm({...promoteForm, title: e.target.value})}
                                    placeholder="如：群管理" />
                            </label>
                            <div className="actions">
                                <button onClick={promoteMember}>提升</button>
                                <button onClick={() => setShowPromoteDialog(false)}>取消</button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            <div className="admins-list">
                {admins.map((admin) => (
                    <div key={admin.id} className="admin-item">
                        <div className="admin-info">
                            {admin.user.avatar && <img src={admin.user.avatar} alt={admin.user.nickname} className="avatar" />}
                            <div className="info-text">
                                <div className="name">{admin.user.nickname}</div>
                                <div className="role-info">
                                    <span className="role-badge" style={{ background: getRoleBadgeColor(admin.role.level) }}>
                                        {admin.role.display_name}
                                    </span>
                                    {admin.title && <span className="title">{admin.title}</span>}
                                </div>
                                <div className="meta">
                                    提升者: {admin.promoted_by_user.nickname} | 
                                    {new Date(admin.created_at).toLocaleDateString()}
                                </div>
                            </div>
                        </div>
                        <div className="admin-actions">
                            {admin.role.level < 100 && (
                                <button onClick={() => demoteMember(admin.user.id)} className="demote-btn">
                                    降级
                                </button>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default GroupAdminManager;
