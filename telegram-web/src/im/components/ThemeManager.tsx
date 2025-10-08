import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface ThemeManagerProps {
    onThemeChange?: (theme: any) => void;
}

interface Theme {
    id: number;
    name: string;
    display_name: string;
    description: string;
    theme_type: string;
    primary_color: string;
    background_color: string;
    text_primary_color: string;
    is_built_in: boolean;
}

interface UserThemeSetting {
    theme_id: number;
    auto_dark_mode: boolean;
    dark_mode_start: string;
    dark_mode_end: string;
    follow_system: boolean;
    enable_animations: boolean;
    reduced_motion: boolean;
    animation_speed: string;
    compact_mode: boolean;
    show_avatars: boolean;
    message_grouping: boolean;
    theme: Theme;
}

const ThemeManager: React.FC<ThemeManagerProps> = ({ onThemeChange }) => {
    const [themes, setThemes] = useState<Theme[]>([]);
    const [currentSetting, setCurrentSetting] = useState<UserThemeSetting | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [activeTab, setActiveTab] = useState<string>('themes');

    // 临时设置状态
    const [tempSettings, setTempSettings] = useState({
        theme_id: 0,
        auto_dark_mode: false,
        dark_mode_start: '22:00',
        dark_mode_end: '07:00',
        follow_system: false,
        enable_animations: true,
        reduced_motion: false,
        animation_speed: 'normal',
        compact_mode: false,
        show_avatars: true,
        message_grouping: true,
    });

    // 获取主题列表
    const fetchThemes = async () => {
        try {
            const response = await axios.get('/api/themes', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setThemes(response.data.themes || []);
        } catch (err) {
            console.error('获取主题列表失败:', err);
        }
    };

    // 获取用户主题设置
    const fetchUserTheme = async () => {
        try {
            const response = await axios.get('/api/themes/user', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setCurrentSetting(response.data);
            setTempSettings({
                theme_id: response.data.theme_id,
                auto_dark_mode: response.data.auto_dark_mode,
                dark_mode_start: response.data.dark_mode_start,
                dark_mode_end: response.data.dark_mode_end,
                follow_system: response.data.follow_system,
                enable_animations: response.data.enable_animations,
                reduced_motion: response.data.reduced_motion,
                animation_speed: response.data.animation_speed,
                compact_mode: response.data.compact_mode,
                show_avatars: response.data.show_avatars,
                message_grouping: response.data.message_grouping,
            });
            applyTheme(response.data.theme);
        } catch (err) {
            console.error('获取用户主题失败:', err);
        }
    };

    // 应用主题
    const applyTheme = (theme: Theme) => {
        const root = document.documentElement;
        root.style.setProperty('--primary-color', theme.primary_color);
        root.style.setProperty('--background-color', theme.background_color);
        root.style.setProperty('--text-primary-color', theme.text_primary_color);
        
        // 设置深色模式类名
        if (theme.theme_type === 'dark') {
            document.body.classList.add('theme-dark');
            document.body.classList.remove('theme-light');
        } else {
            document.body.classList.add('theme-light');
            document.body.classList.remove('theme-dark');
        }

        if (onThemeChange) {
            onThemeChange(theme);
        }
    };

    // 保存主题设置
    const saveThemeSettings = async () => {
        setIsLoading(true);
        try {
            const response = await axios.put('/api/themes/user', tempSettings, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
                }
            });
            setCurrentSetting(response.data.setting);
            applyTheme(response.data.setting.theme);
            alert('主题设置已保存！');
        } catch (err: any) {
            console.error('保存主题设置失败:', err);
            alert(err.response?.data?.error || '保存失败');
        } finally {
            setIsLoading(false);
        }
    };

    // 切换主题
    const switchTheme = (themeId: number) => {
        setTempSettings({ ...tempSettings, theme_id: themeId });
    };

    useEffect(() => {
        fetchThemes();
        fetchUserTheme();
    }, []);

    return (
        <div className="theme-manager">
            <div className="theme-manager-header">
                <h2>🎨 主题设置</h2>
                <div className="tabs">
                    <button
                        className={activeTab === 'themes' ? 'active' : ''}
                        onClick={() => setActiveTab('themes')}
                    >
                        主题选择
                    </button>
                    <button
                        className={activeTab === 'customize' ? 'active' : ''}
                        onClick={() => setActiveTab('customize')}
                    >
                        个性化
                    </button>
                    <button
                        className={activeTab === 'advanced' ? 'active' : ''}
                        onClick={() => setActiveTab('advanced')}
                    >
                        高级设置
                    </button>
                </div>
            </div>

            <div className="theme-manager-content">
                {activeTab === 'themes' && (
                    <div className="themes-grid">
                        {themes.map((theme) => (
                            <div
                                key={theme.id}
                                className={`theme-card ${tempSettings.theme_id === theme.id ? 'selected' : ''}`}
                                onClick={() => switchTheme(theme.id)}
                                style={{
                                    background: theme.background_color,
                                    color: theme.text_primary_color,
                                    borderColor: theme.primary_color
                                }}
                            >
                                <div className="theme-preview" style={{ background: theme.primary_color }}></div>
                                <h3>{theme.display_name}</h3>
                                <p>{theme.description}</p>
                                {theme.is_built_in && <span className="badge">内置</span>}
                                {tempSettings.theme_id === theme.id && <span className="badge active">使用中</span>}
                            </div>
                        ))}
                    </div>
                )}

                {activeTab === 'customize' && (
                    <div className="customize-panel">
                        <div className="setting-group">
                            <h3>🌙 夜间模式</h3>
                            <label className="switch-label">
                                <input
                                    type="checkbox"
                                    checked={tempSettings.follow_system}
                                    onChange={(e) => setTempSettings({ ...tempSettings, follow_system: e.target.checked })}
                                />
                                <span>跟随系统设置</span>
                            </label>
                            <label className="switch-label">
                                <input
                                    type="checkbox"
                                    checked={tempSettings.auto_dark_mode}
                                    onChange={(e) => setTempSettings({ ...tempSettings, auto_dark_mode: e.target.checked })}
                                    disabled={tempSettings.follow_system}
                                />
                                <span>自动夜间模式</span>
                            </label>
                            {tempSettings.auto_dark_mode && !tempSettings.follow_system && (
                                <div className="time-range">
                                    <label>
                                        开始时间:
                                        <input
                                            type="time"
                                            value={tempSettings.dark_mode_start}
                                            onChange={(e) => setTempSettings({ ...tempSettings, dark_mode_start: e.target.value })}
                                        />
                                    </label>
                                    <label>
                                        结束时间:
                                        <input
                                            type="time"
                                            value={tempSettings.dark_mode_end}
                                            onChange={(e) => setTempSettings({ ...tempSettings, dark_mode_end: e.target.value })}
                                        />
                                    </label>
                                </div>
                            )}
                        </div>

                        <div className="setting-group">
                            <h3>✨ 动画效果</h3>
                            <label className="switch-label">
                                <input
                                    type="checkbox"
                                    checked={tempSettings.enable_animations}
                                    onChange={(e) => setTempSettings({ ...tempSettings, enable_animations: e.target.checked })}
                                />
                                <span>启用动画</span>
                            </label>
                            <label className="switch-label">
                                <input
                                    type="checkbox"
                                    checked={tempSettings.reduced_motion}
                                    onChange={(e) => setTempSettings({ ...tempSettings, reduced_motion: e.target.checked })}
                                />
                                <span>减少动效（适合性能较低设备）</span>
                            </label>
                            <label>
                                动画速度:
                                <select
                                    value={tempSettings.animation_speed}
                                    onChange={(e) => setTempSettings({ ...tempSettings, animation_speed: e.target.value })}
                                >
                                    <option value="slow">慢速</option>
                                    <option value="normal">正常</option>
                                    <option value="fast">快速</option>
                                </select>
                            </label>
                        </div>
                    </div>
                )}

                {activeTab === 'advanced' && (
                    <div className="advanced-panel">
                        <div className="setting-group">
                            <h3>📐 布局设置</h3>
                            <label className="switch-label">
                                <input
                                    type="checkbox"
                                    checked={tempSettings.compact_mode}
                                    onChange={(e) => setTempSettings({ ...tempSettings, compact_mode: e.target.checked })}
                                />
                                <span>紧凑模式</span>
                            </label>
                            <label className="switch-label">
                                <input
                                    type="checkbox"
                                    checked={tempSettings.show_avatars}
                                    onChange={(e) => setTempSettings({ ...tempSettings, show_avatars: e.target.checked })}
                                />
                                <span>显示头像</span>
                            </label>
                            <label className="switch-label">
                                <input
                                    type="checkbox"
                                    checked={tempSettings.message_grouping}
                                    onChange={(e) => setTempSettings({ ...tempSettings, message_grouping: e.target.checked })}
                                />
                                <span>消息分组显示</span>
                            </label>
                        </div>
                    </div>
                )}
            </div>

            <div className="theme-manager-footer">
                <button onClick={saveThemeSettings} disabled={isLoading} className="save-btn">
                    {isLoading ? '保存中...' : '保存设置'}
                </button>
            </div>

            <style>{`
                .theme-manager {
                    max-width: 900px;
                    margin: 0 auto;
                    padding: 20px;
                }

                .theme-manager-header {
                    margin-bottom: 30px;
                }

                .tabs {
                    display: flex;
                    gap: 10px;
                    margin-top: 20px;
                    border-bottom: 2px solid #e0e0e0;
                }

                .tabs button {
                    padding: 10px 20px;
                    border: none;
                    background: none;
                    cursor: pointer;
                    border-bottom: 2px solid transparent;
                    margin-bottom: -2px;
                }

                .tabs button.active {
                    border-bottom-color: #2196f3;
                    color: #2196f3;
                    font-weight: bold;
                }

                .themes-grid {
                    display: grid;
                    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
                    gap: 20px;
                }

                .theme-card {
                    padding: 20px;
                    border: 2px solid #e0e0e0;
                    border-radius: 12px;
                    cursor: pointer;
                    transition: all 0.3s;
                    position: relative;
                }

                .theme-card:hover {
                    transform: translateY(-5px);
                    box-shadow: 0 8px 16px rgba(0,0,0,0.1);
                }

                .theme-card.selected {
                    border-width: 3px;
                    box-shadow: 0 4px 12px rgba(33, 150, 243, 0.3);
                }

                .theme-preview {
                    height: 60px;
                    border-radius: 8px;
                    margin-bottom: 15px;
                }

                .badge {
                    position: absolute;
                    top: 10px;
                    right: 10px;
                    padding: 4px 8px;
                    background: #ff9800;
                    color: white;
                    border-radius: 12px;
                    font-size: 12px;
                }

                .badge.active {
                    background: #4caf50;
                }

                .setting-group {
                    margin-bottom: 30px;
                    padding: 20px;
                    background: #f5f5f5;
                    border-radius: 8px;
                }

                .switch-label {
                    display: flex;
                    align-items: center;
                    margin: 15px 0;
                    cursor: pointer;
                }

                .switch-label input[type="checkbox"] {
                    margin-right: 10px;
                    width: 20px;
                    height: 20px;
                }

                .time-range {
                    display: flex;
                    gap: 20px;
                    margin-top: 15px;
                }

                .time-range label {
                    display: flex;
                    flex-direction: column;
                    gap: 5px;
                }

                .time-range input[type="time"] {
                    padding: 8px;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }

                .theme-manager-footer {
                    margin-top: 30px;
                    text-align: center;
                }

                .save-btn {
                    padding: 12px 40px;
                    background: #2196f3;
                    color: white;
                    border: none;
                    border-radius: 6px;
                    font-size: 16px;
                    cursor: pointer;
                    transition: background 0.3s;
                }

                .save-btn:hover:not(:disabled) {
                    background: #1976d2;
                }

                .save-btn:disabled {
                    opacity: 0.6;
                    cursor: not-allowed;
                }
            `}</style>
        </div>
    );
};

export default ThemeManager;
