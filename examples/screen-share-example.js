/**
 * 屏幕共享功能示例代码
 * 志航密信 v1.6.0
 * 
 * 使用方法：
 * 1. 在通话页面引入此文件
 * 2. 调用 ScreenShareManager 类的方法
 */

class ScreenShareManager {
    constructor(callId, apiBaseUrl = 'http://localhost:8080/api') {
        this.callId = callId;
        this.apiBaseUrl = apiBaseUrl;
        this.localStream = null;
        this.peerConnection = null;
        this.isSharing = false;
    }

    /**
     * 开始屏幕共享
     * @param {Object} options - 配置选项
     * @param {string} options.quality - 质量: 'high', 'medium', 'low'
     * @param {boolean} options.withAudio - 是否包含音频
     * @returns {Promise<MediaStream>}
     */
    async startScreenShare(options = {}) {
        const { quality = 'medium', withAudio = false } = options;

        try {
            // 1. 获取屏幕流
            const constraints = this.getConstraintsByQuality(quality, withAudio);
            this.localStream = await navigator.mediaDevices.getDisplayMedia(constraints);

            console.log('✅ 屏幕流获取成功');

            // 2. 通知后端开始屏幕共享
            const userName = localStorage.getItem('userName') || '用户';
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/calls/${this.callId}/screen-share/start`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    user_name: userName,
                    quality: quality,
                    with_audio: withAudio
                })
            });

            const result = await response.json();

            if (!result.success) {
                throw new Error(result.error || '开始屏幕共享失败');
            }

            console.log('✅ 后端通知成功');

            // 3. 设置停止监听
            this.localStream.getVideoTracks()[0].addEventListener('ended', () => {
                console.log('🛑 屏幕共享已停止（用户点击停止共享）');
                this.stopScreenShare();
            });

            this.isSharing = true;

            return this.localStream;

        } catch (error) {
            console.error('❌ 开始屏幕共享失败:', error);
            throw error;
        }
    }

    /**
     * 停止屏幕共享
     */
    async stopScreenShare() {
        try {
            // 1. 停止本地流
            if (this.localStream) {
                this.localStream.getTracks().forEach(track => track.stop());
                this.localStream = null;
            }

            // 2. 关闭PeerConnection
            if (this.peerConnection) {
                this.peerConnection.close();
                this.peerConnection = null;
            }

            // 3. 通知后端停止屏幕共享
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/calls/${this.callId}/screen-share/stop`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            const result = await response.json();

            if (!result.success) {
                console.warn('后端通知失败:', result.error);
            }

            this.isSharing = false;

            console.log('✅ 屏幕共享已停止');

        } catch (error) {
            console.error('❌ 停止屏幕共享失败:', error);
            throw error;
        }
    }

    /**
     * 更改屏幕共享质量
     * @param {string} quality - 质量: 'high', 'medium', 'low'
     */
    async changeQuality(quality) {
        if (!this.isSharing) {
            throw new Error('当前没有屏幕共享');
        }

        try {
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/calls/${this.callId}/screen-share/quality`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ quality })
            });

            const result = await response.json();

            if (!result.success) {
                throw new Error(result.error || '更改质量失败');
            }

            console.log('✅ 屏幕共享质量已更改为:', quality);

            // 重新获取屏幕流
            const constraints = this.getConstraintsByQuality(quality, false);
            const newStream = await navigator.mediaDevices.getDisplayMedia(constraints);

            // 替换旧流
            if (this.localStream) {
                this.localStream.getTracks().forEach(track => track.stop());
            }
            this.localStream = newStream;

            return newStream;

        } catch (error) {
            console.error('❌ 更改质量失败:', error);
            throw error;
        }
    }

    /**
     * 获取屏幕共享状态
     */
    async getStatus() {
        try {
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/calls/${this.callId}/screen-share/status`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            const result = await response.json();

            return result.data;

        } catch (error) {
            console.error('❌ 获取状态失败:', error);
            throw error;
        }
    }

    /**
     * 根据质量获取约束条件
     */
    getConstraintsByQuality(quality, withAudio) {
        const constraints = {
            video: {},
            audio: withAudio
        };

        switch (quality) {
            case 'high':
                constraints.video = {
                    width: { ideal: 1920 },
                    height: { ideal: 1080 },
                    frameRate: { ideal: 30 }
                };
                break;
            case 'medium':
                constraints.video = {
                    width: { ideal: 1280 },
                    height: { ideal: 720 },
                    frameRate: { ideal: 24 }
                };
                break;
            case 'low':
                constraints.video = {
                    width: { ideal: 640 },
                    height: { ideal: 480 },
                    frameRate: { ideal: 15 }
                };
                break;
            default:
                constraints.video = true;
        }

        return constraints;
    }
}

// ============================================
// 使用示例
// ============================================

/**
 * 示例1: 基础使用
 */
async function example1() {
    const callId = 'call_123456';
    const screenShare = new ScreenShareManager(callId);

    // 开始屏幕共享（中等质量，不含音频）
    try {
        const stream = await screenShare.startScreenShare({
            quality: 'medium',
            withAudio: false
        });

        // 显示在video元素中
        const videoElement = document.getElementById('screenShareVideo');
        videoElement.srcObject = stream;
        videoElement.play();

        console.log('✅ 屏幕共享已开始');
    } catch (error) {
        console.error('屏幕共享失败:', error.message);
        alert('屏幕共享失败: ' + error.message);
    }
}

/**
 * 示例2: 带音频的高质量共享
 */
async function example2() {
    const callId = 'call_123456';
    const screenShare = new ScreenShareManager(callId);

    try {
        // 高质量+系统音频
        const stream = await screenShare.startScreenShare({
            quality: 'high',
            withAudio: true
        });

        const videoElement = document.getElementById('screenShareVideo');
        videoElement.srcObject = stream;
        videoElement.play();

        // 显示停止按钮
        document.getElementById('stopShareBtn').style.display = 'block';

    } catch (error) {
        alert('开始屏幕共享失败: ' + error.message);
    }
}

/**
 * 示例3: 停止屏幕共享
 */
async function example3() {
    const callId = 'call_123456';
    const screenShare = new ScreenShareManager(callId);

    try {
        await screenShare.stopScreenShare();

        // 清除video元素
        const videoElement = document.getElementById('screenShareVideo');
        videoElement.srcObject = null;

        // 隐藏停止按钮
        document.getElementById('stopShareBtn').style.display = 'none';

        console.log('✅ 屏幕共享已停止');
    } catch (error) {
        console.error('停止失败:', error);
    }
}

/**
 * 示例4: 动态调整质量
 */
async function example4() {
    const callId = 'call_123456';
    const screenShare = new ScreenShareManager(callId);

    // 开始共享
    await screenShare.startScreenShare({ quality: 'medium' });

    // 根据网络情况调整质量
    setTimeout(async () => {
        try {
            const newStream = await screenShare.changeQuality('low');
            
            // 更新video元素
            const videoElement = document.getElementById('screenShareVideo');
            videoElement.srcObject = newStream;
            
            console.log('✅ 质量已降低为 low（网络不佳）');
        } catch (error) {
            console.error('更改质量失败:', error);
        }
    }, 10000); // 10秒后降低质量
}

/**
 * 示例5: 查询屏幕共享状态
 */
async function example5() {
    const callId = 'call_123456';
    const screenShare = new ScreenShareManager(callId);

    try {
        const status = await screenShare.getStatus();

        if (status && status.is_active) {
            console.log('📺 正在共享屏幕');
            console.log('  共享者:', status.sharer_name);
            console.log('  质量:', status.quality);
            console.log('  音频:', status.with_audio ? '是' : '否');
            console.log('  开始时间:', status.start_time);
        } else {
            console.log('⏸️ 当前没有屏幕共享');
        }
    } catch (error) {
        console.error('查询状态失败:', error);
    }
}

// ============================================
// 完整的UI集成示例
// ============================================

/**
 * 完整示例：在通话界面中集成屏幕共享
 */
class CallUIWithScreenShare {
    constructor(callId) {
        this.callId = callId;
        this.screenShare = new ScreenShareManager(callId);
        this.setupUI();
    }

    setupUI() {
        // 创建UI元素
        const container = document.getElementById('callContainer');

        // 屏幕共享按钮
        const shareBtn = document.createElement('button');
        shareBtn.id = 'startShareBtn';
        shareBtn.className = 'control-btn';
        shareBtn.innerHTML = '📺 共享屏幕';
        shareBtn.onclick = () => this.handleStartShare();

        // 停止共享按钮
        const stopBtn = document.createElement('button');
        stopBtn.id = 'stopShareBtn';
        stopBtn.className = 'control-btn danger';
        stopBtn.innerHTML = '🛑 停止共享';
        stopBtn.style.display = 'none';
        stopBtn.onclick = () => this.handleStopShare();

        // 质量选择
        const qualitySelect = document.createElement('select');
        qualitySelect.id = 'qualitySelect';
        qualitySelect.innerHTML = `
            <option value="low">流畅（低质量）</option>
            <option value="medium" selected>标准（中质量）</option>
            <option value="high">高清（高质量）</option>
        `;
        qualitySelect.onchange = (e) => this.handleChangeQuality(e.target.value);

        // 音频选项
        const audioCheck = document.createElement('label');
        audioCheck.innerHTML = `
            <input type="checkbox" id="shareAudioCheck"> 共享系统音频
        `;

        // 屏幕共享视频显示
        const video = document.createElement('video');
        video.id = 'screenShareVideo';
        video.className = 'screen-share-video';
        video.style.display = 'none';
        video.autoplay = true;

        // 状态显示
        const statusDiv = document.createElement('div');
        statusDiv.id = 'shareStatus';
        statusDiv.className = 'share-status';

        // 添加到容器
        container.appendChild(shareBtn);
        container.appendChild(stopBtn);
        container.appendChild(qualitySelect);
        container.appendChild(audioCheck);
        container.appendChild(video);
        container.appendChild(statusDiv);
    }

    async handleStartShare() {
        const quality = document.getElementById('qualitySelect').value;
        const withAudio = document.getElementById('shareAudioCheck').checked;

        try {
            const stream = await this.screenShare.startScreenShare({ quality, withAudio });

            // 显示屏幕共享
            const video = document.getElementById('screenShareVideo');
            video.srcObject = stream;
            video.style.display = 'block';
            video.play();

            // 切换按钮
            document.getElementById('startShareBtn').style.display = 'none';
            document.getElementById('stopShareBtn').style.display = 'block';

            // 更新状态
            this.updateStatus('正在共享屏幕...', 'active');

            console.log('✅ 屏幕共享已开始');
        } catch (error) {
            console.error('❌ 屏幕共享失败:', error);
            alert('屏幕共享失败: ' + error.message);
        }
    }

    async handleStopShare() {
        try {
            await this.screenShare.stopScreenShare();

            // 隐藏视频
            const video = document.getElementById('screenShareVideo');
            video.srcObject = null;
            video.style.display = 'none';

            // 切换按钮
            document.getElementById('startShareBtn').style.display = 'block';
            document.getElementById('stopShareBtn').style.display = 'none';

            // 更新状态
            this.updateStatus('', '');

            console.log('✅ 屏幕共享已停止');
        } catch (error) {
            console.error('❌ 停止失败:', error);
            alert('停止屏幕共享失败: ' + error.message);
        }
    }

    async handleChangeQuality(quality) {
        if (!this.screenShare.isSharing) {
            return;
        }

        try {
            const newStream = await this.screenShare.changeQuality(quality);

            // 更新视频流
            const video = document.getElementById('screenShareVideo');
            video.srcObject = newStream;

            console.log('✅ 质量已更改为:', quality);
            this.updateStatus(`正在共享屏幕（${this.getQualityLabel(quality)}）`, 'active');
        } catch (error) {
            console.error('❌ 更改质量失败:', error);
            alert('更改质量失败: ' + error.message);
        }
    }

    updateStatus(message, type) {
        const statusDiv = document.getElementById('shareStatus');
        statusDiv.textContent = message;
        statusDiv.className = 'share-status ' + type;
    }

    getQualityLabel(quality) {
        const labels = {
            'high': '高清',
            'medium': '标准',
            'low': '流畅'
        };
        return labels[quality] || quality;
    }

    // 定期检查屏幕共享状态
    async startStatusPolling() {
        setInterval(async () => {
            try {
                const status = await this.screenShare.getStatus();

                if (status && status.is_active) {
                    if (status.sharer_id !== this.getCurrentUserId()) {
                        // 其他人正在共享
                        this.showRemoteScreenShare(status);
                    }
                } else {
                    this.hideRemoteScreenShare();
                }
            } catch (error) {
                console.error('获取状态失败:', error);
            }
        }, 2000); // 每2秒检查一次
    }

    showRemoteScreenShare(status) {
        const statusDiv = document.getElementById('shareStatus');
        statusDiv.textContent = `${status.sharer_name} 正在共享屏幕`;
        statusDiv.className = 'share-status remote';
    }

    hideRemoteScreenShare() {
        const statusDiv = document.getElementById('shareStatus');
        if (statusDiv.className.includes('remote')) {
            statusDiv.textContent = '';
            statusDiv.className = 'share-status';
        }
    }

    getCurrentUserId() {
        return parseInt(localStorage.getItem('userId') || '0');
    }
}

// ============================================
// CSS样式（可选）
// ============================================

const styles = `
<style>
.screen-share-video {
    width: 100%;
    max-width: 1280px;
    height: auto;
    border: 2px solid #007bff;
    border-radius: 8px;
    margin: 10px 0;
    background: #000;
}

.control-btn {
    padding: 10px 20px;
    margin: 5px;
    border: none;
    border-radius: 4px;
    background: #007bff;
    color: white;
    cursor: pointer;
    font-size: 14px;
}

.control-btn:hover {
    background: #0056b3;
}

.control-btn.danger {
    background: #dc3545;
}

.control-btn.danger:hover {
    background: #bd2130;
}

.share-status {
    padding: 10px;
    margin: 10px 0;
    border-radius: 4px;
    font-size: 14px;
}

.share-status.active {
    background: #d4edda;
    color: #155724;
    border: 1px solid #c3e6cb;
}

.share-status.remote {
    background: #d1ecf1;
    color: #0c5460;
    border: 1px solid #bee5eb;
}

#qualitySelect {
    padding: 8px 12px;
    margin: 5px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
}
</style>
`;

// ============================================
// 导出
// ============================================

// 如果使用模块系统
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ScreenShareManager;
}

// 如果在浏览器中直接使用
if (typeof window !== 'undefined') {
    window.ScreenShareManager = ScreenShareManager;
    window.CallUIWithScreenShare = CallUIWithScreenShare;
}



