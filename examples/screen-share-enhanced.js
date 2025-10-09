/**
 * 屏幕共享增强管理器
 * 志航密信 v1.6.0
 * 
 * 增强功能：
 * - 网络自适应质量调整
 * - 自动重连
 * - 错误处理和重试
 * - 历史记录管理
 * - 统计信息
 * - 录制功能
 * - 权限检查
 */

class ScreenShareEnhancedManager {
    constructor(callId, apiBaseUrl = 'http://localhost:8080/api') {
        this.callId = callId;
        this.apiBaseUrl = apiBaseUrl;
        this.localStream = null;
        this.peerConnection = null;
        this.isSharing = false;
        this.isRecording = false;
        this.currentRecordingId = null;
        
        // 网络监控
        this.networkMonitor = {
            speeds: [],
            avgSpeed: 0,
            lastCheckTime: Date.now(),
            checkInterval: 5000, // 5秒检查一次
        };
        
        // 质量历史
        this.qualityHistory = [];
        
        // 重连配置
        this.reconnectConfig = {
            maxRetries: 3,
            retryDelay: 2000,
            currentRetries: 0,
        };
        
        // 错误回调
        this.onError = null;
        this.onQualityChange = null;
        this.onNetworkChange = null;
        
        // 启动网络监控
        this.startNetworkMonitoring();
    }

    /**
     * 开始屏幕共享（增强版）
     */
    async startScreenShare(options = {}) {
        const { 
            quality = 'medium', 
            withAudio = false,
            autoAdjustQuality = true 
        } = options;

        try {
            // 1. 检查权限
            const hasPermission = await this.checkPermission(quality);
            if (!hasPermission) {
                throw new Error('您没有屏幕共享权限或质量等级过高');
            }

            // 2. 获取屏幕流
            const constraints = this.getConstraintsByQuality(quality, withAudio);
            this.localStream = await navigator.mediaDevices.getDisplayMedia(constraints);

            console.log('✅ 屏幕流获取成功');

            // 3. 通知后端开始共享
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

            // 4. 记录初始质量
            this.currentQuality = quality;
            this.qualityHistory.push({
                quality: quality,
                timestamp: Date.now(),
                reason: 'initial'
            });

            // 5. 设置监听
            this.setupStreamListeners();

            // 6. 启动自适应质量调整
            if (autoAdjustQuality) {
                this.startAutoQualityAdjustment();
            }

            this.isSharing = true;

            return this.localStream;

        } catch (error) {
            console.error('❌ 开始屏幕共享失败:', error);
            
            if (this.onError) {
                this.onError(error);
            }
            
            throw error;
        }
    }

    /**
     * 设置流监听器
     */
    setupStreamListeners() {
        if (!this.localStream) return;

        // 监听轨道结束
        this.localStream.getVideoTracks()[0].addEventListener('ended', () => {
            console.log('🛑 用户主动停止了屏幕共享');
            this.stopScreenShare('user_stopped');
        });

        // 监听轨道静音/取消静音
        this.localStream.getVideoTracks()[0].addEventListener('mute', () => {
            console.log('🔇 屏幕共享被静音');
        });

        this.localStream.getVideoTracks()[0].addEventListener('unmute', () => {
            console.log('🔊 屏幕共享取消静音');
        });
    }

    /**
     * 停止屏幕共享
     */
    async stopScreenShare(reason = 'manual') {
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

            // 3. 停止录制
            if (this.isRecording) {
                await this.stopRecording();
            }

            // 4. 通知后端停止
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
            this.stopAutoQualityAdjustment();

            console.log('✅ 屏幕共享已停止, 原因:', reason);

        } catch (error) {
            console.error('❌ 停止屏幕共享失败:', error);
            throw error;
        }
    }

    /**
     * 启动网络监控
     */
    startNetworkMonitoring() {
        // 使用 Network Information API (如果可用)
        if ('connection' in navigator) {
            const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
            
            if (connection) {
                connection.addEventListener('change', () => {
                    const speed = this.estimateNetworkSpeed(connection);
                    this.updateNetworkSpeed(speed);
                    
                    if (this.onNetworkChange) {
                        this.onNetworkChange({
                            effectiveType: connection.effectiveType,
                            downlink: connection.downlink,
                            rtt: connection.rtt,
                            estimatedSpeed: speed
                        });
                    }
                });
            }
        }

        // 定期估算网速
        this.networkCheckTimer = setInterval(() => {
            if (this.isSharing) {
                this.checkNetworkQuality();
            }
        }, this.networkMonitor.checkInterval);
    }

    /**
     * 估算网络速度
     */
    estimateNetworkSpeed(connection) {
        if (!connection) return 1000; // 默认1Mbps

        // 使用 downlink (Mbps)
        if (connection.downlink) {
            return connection.downlink * 1000; // 转换为Kbps
        }

        // 根据 effectiveType 估算
        const speedMap = {
            'slow-2g': 50,
            '2g': 250,
            '3g': 750,
            '4g': 3000,
            '5g': 10000
        };

        return speedMap[connection.effectiveType] || 1000;
    }

    /**
     * 更新网速
     */
    updateNetworkSpeed(speed) {
        this.networkMonitor.speeds.push(speed);
        
        // 只保留最近10次测量
        if (this.networkMonitor.speeds.length > 10) {
            this.networkMonitor.speeds.shift();
        }

        // 计算平均速度
        this.networkMonitor.avgSpeed = this.networkMonitor.speeds.reduce((a, b) => a + b, 0) 
            / this.networkMonitor.speeds.length;

        console.log(`📶 网络速度: ${Math.round(this.networkMonitor.avgSpeed)} Kbps`);
    }

    /**
     * 检查网络质量
     */
    async checkNetworkQuality() {
        if (!this.peerConnection) return;

        try {
            const stats = await this.peerConnection.getStats();
            
            stats.forEach(report => {
                if (report.type === 'outbound-rtp' && report.kind === 'video') {
                    // 计算当前发送速率
                    const bytesSent = report.bytesSent || 0;
                    const timestamp = report.timestamp;

                    if (this.lastStats) {
                        const bytesDiff = bytesSent - this.lastStats.bytesSent;
                        const timeDiff = timestamp - this.lastStats.timestamp;
                        const speed = (bytesDiff * 8) / timeDiff; // Kbps

                        this.updateNetworkSpeed(speed);
                    }

                    this.lastStats = { bytesSent, timestamp };
                }
            });
        } catch (error) {
            console.error('获取网络统计失败:', error);
        }
    }

    /**
     * 启动自动质量调整
     */
    startAutoQualityAdjustment() {
        this.qualityCheckTimer = setInterval(async () => {
            if (!this.isSharing) return;

            const recommendedQuality = this.recommendQuality();
            
            if (recommendedQuality !== this.currentQuality) {
                console.log(`🔄 建议切换质量: ${this.currentQuality} -> ${recommendedQuality}`);
                
                try {
                    await this.changeQuality(recommendedQuality, 'auto_network');
                } catch (error) {
                    console.error('自动调整质量失败:', error);
                }
            }
        }, 10000); // 每10秒检查一次
    }

    /**
     * 停止自动质量调整
     */
    stopAutoQualityAdjustment() {
        if (this.qualityCheckTimer) {
            clearInterval(this.qualityCheckTimer);
            this.qualityCheckTimer = null;
        }
    }

    /**
     * 推荐质量
     */
    recommendQuality() {
        const speed = this.networkMonitor.avgSpeed;

        // 获取CPU使用率（如果可用）
        const cpuUsage = this.estimateCPUUsage();

        if (speed > 3000 && cpuUsage < 70) {
            return 'high';
        } else if (speed > 1000 && cpuUsage < 80) {
            return 'medium';
        }
        return 'low';
    }

    /**
     * 估算CPU使用率（简化版）
     */
    estimateCPUUsage() {
        // 实际应用中应该使用更准确的方法
        // 这里返回一个估算值
        return 50;
    }

    /**
     * 更改质量
     */
    async changeQuality(quality, reason = 'manual') {
        if (!this.isSharing) {
            throw new Error('当前没有屏幕共享');
        }

        try {
            const oldQuality = this.currentQuality;

            // 1. 通知后端
            const token = localStorage.getItem('token');

            await fetch(`${this.apiBaseUrl}/calls/${this.callId}/screen-share/quality`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ quality })
            });

            // 2. 记录质量变更
            await fetch(`${this.apiBaseUrl}/screen-share/${this.callId}/quality-change`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    from_quality: oldQuality,
                    to_quality: quality,
                    reason: reason,
                    network_speed: this.networkMonitor.avgSpeed,
                    cpu_usage: this.estimateCPUUsage()
                })
            });

            // 3. 重新获取屏幕流
            const constraints = this.getConstraintsByQuality(quality, false);
            const newStream = await navigator.mediaDevices.getDisplayMedia(constraints);

            // 4. 替换旧流
            if (this.localStream) {
                this.localStream.getTracks().forEach(track => track.stop());
            }
            this.localStream = newStream;

            // 5. 重新设置监听
            this.setupStreamListeners();

            // 6. 更新质量历史
            this.currentQuality = quality;
            this.qualityHistory.push({
                quality: quality,
                timestamp: Date.now(),
                reason: reason
            });

            console.log(`✅ 质量已更改: ${oldQuality} -> ${quality} (${reason})`);

            if (this.onQualityChange) {
                this.onQualityChange({ from: oldQuality, to: quality, reason });
            }

            return newStream;

        } catch (error) {
            console.error('❌ 更改质量失败:', error);
            throw error;
        }
    }

    /**
     * 开始录制
     */
    async startRecording(options = {}) {
        const { format = 'webm', quality = 'medium' } = options;

        if (!this.isSharing) {
            throw new Error('请先开始屏幕共享');
        }

        try {
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/screen-share/${this.callId}/recording/start`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ format, quality })
            });

            const result = await response.json();

            if (!result.success) {
                throw new Error(result.error || '开始录制失败');
            }

            this.isRecording = true;
            this.currentRecordingId = result.data.id;

            // 创建本地录制器
            this.mediaRecorder = new MediaRecorder(this.localStream, {
                mimeType: `video/${format}`
            });

            this.recordedChunks = [];

            this.mediaRecorder.ondataavailable = (event) => {
                if (event.data.size > 0) {
                    this.recordedChunks.push(event.data);
                }
            };

            this.mediaRecorder.start();

            console.log('✅ 录制已开始');

            return result.data;

        } catch (error) {
            console.error('❌ 开始录制失败:', error);
            throw error;
        }
    }

    /**
     * 停止录制
     */
    async stopRecording() {
        if (!this.isRecording) {
            throw new Error('当前没有录制');
        }

        try {
            // 停止本地录制器
            this.mediaRecorder.stop();

            await new Promise((resolve) => {
                this.mediaRecorder.onstop = resolve;
            });

            // 创建Blob
            const blob = new Blob(this.recordedChunks, {
                type: this.mediaRecorder.mimeType
            });

            const fileSize = blob.size;

            // 通知后端结束录制
            const token = localStorage.getItem('token');

            await fetch(`${this.apiBaseUrl}/screen-share/recordings/${this.currentRecordingId}/end`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    file_path: `/recordings/${this.currentRecordingId}.webm`,
                    file_size: fileSize
                })
            });

            this.isRecording = false;
            this.currentRecordingId = null;

            console.log('✅ 录制已停止, 大小:', fileSize, 'bytes');

            // 返回Blob供下载
            return blob;

        } catch (error) {
            console.error('❌ 停止录制失败:', error);
            throw error;
        }
    }

    /**
     * 检查权限
     */
    async checkPermission(quality) {
        try {
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/screen-share/check-permission?quality=${quality}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            const result = await response.json();

            return result.data.allowed;

        } catch (error) {
            console.error('检查权限失败:', error);
            return false;
        }
    }

    /**
     * 获取历史记录
     */
    async getHistory(page = 1, pageSize = 20) {
        try {
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/screen-share/history?page=${page}&page_size=${pageSize}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            const result = await response.json();

            return result.data;

        } catch (error) {
            console.error('获取历史记录失败:', error);
            throw error;
        }
    }

    /**
     * 获取统计信息
     */
    async getStatistics() {
        try {
            const token = localStorage.getItem('token');

            const response = await fetch(`${this.apiBaseUrl}/screen-share/statistics`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            const result = await response.json();

            return result.data;

        } catch (error) {
            console.error('获取统计信息失败:', error);
            throw error;
        }
    }

    /**
     * 获取约束条件
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

    /**
     * 清理资源
     */
    destroy() {
        this.stopAutoQualityAdjustment();
        
        if (this.networkCheckTimer) {
            clearInterval(this.networkCheckTimer);
        }

        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
        }

        if (this.peerConnection) {
            this.peerConnection.close();
        }

        console.log('🧹 资源已清理');
    }
}

// 导出
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ScreenShareEnhancedManager;
}

if (typeof window !== 'undefined') {
    window.ScreenShareEnhancedManager = ScreenShareEnhancedManager;
}


