/**
 * 志航密信 WebRTC 音视频优化器
 * 提供音视频质量优化、网络自适应、降级策略等功能
 */

import { errorHandler } from './error-handler';
import { performanceMonitor } from './performance';

// WebRTC 优化器接口
export interface WebRTCOptimizerConfig {
  enableAdaptiveBitrate: boolean;
  enableNoiseReduction: boolean;
  enableEchoCancellation: boolean;
  enableAutoGainControl: boolean;
  enableVideoOptimization: boolean;
  maxBitrate: number;
  minBitrate: number;
  qualityThreshold: number;
  networkCheckInterval: number;
  fallbackStrategy: 'aggressive' | 'conservative' | 'balanced';
}

// 网络质量等级
export enum NetworkQuality {
  EXCELLENT = 'excellent',
  GOOD = 'good',
  FAIR = 'fair',
  POOR = 'poor',
  VERY_POOR = 'very_poor'
}

// 音视频质量设置
export interface MediaQualitySettings {
  videoBitrate: number;
  audioBitrate: number;
  resolution: string;
  frameRate: number;
  audioCodec: string;
  videoCodec: string;
  qualityLevel: string;
}

// 网络统计信息
export interface NetworkStats {
  rtt: number;
  packetLoss: number;
  jitter: number;
  bandwidth: number;
  networkType: string;
  signalStrength: number;
  isStable: boolean;
}

// WebRTC 优化器类
export class WebRTCOptimizer {
  private static instance: WebRTCOptimizer;
  private config: WebRTCOptimizerConfig;
  private currentQuality: MediaQualitySettings;
  private networkStats: NetworkStats;
  private qualityHistory: Array<{ timestamp: number; quality: number }> = [];
  private adaptationTimer: NodeJS.Timeout | null = null;
  private networkMonitorTimer: NodeJS.Timeout | null = null;
  private isOptimizing = false;

  private constructor(config: Partial<WebRTCOptimizerConfig> = {}) {
    this.config = {
      enableAdaptiveBitrate: true,
      enableNoiseReduction: true,
      enableEchoCancellation: true,
      enableAutoGainControl: true,
      enableVideoOptimization: true,
      maxBitrate: 4000,
      minBitrate: 200,
      qualityThreshold: 70,
      networkCheckInterval: 5000,
      fallbackStrategy: 'balanced',
      ...config
    };

    this.currentQuality = this.getDefaultQualitySettings();
    this.networkStats = this.getDefaultNetworkStats();
  }

  // 获取单例实例
  public static getInstance(config?: Partial<WebRTCOptimizerConfig>): WebRTCOptimizer {
    if (!WebRTCOptimizer.instance) {
      WebRTCOptimizer.instance = new WebRTCOptimizer(config);
    }
    return WebRTCOptimizer.instance;
  }

  // 获取默认质量设置
  private getDefaultQualitySettings(): MediaQualitySettings {
    return {
      videoBitrate: 2000,
      audioBitrate: 128,
      resolution: '1280x720',
      frameRate: 30,
      audioCodec: 'opus',
      videoCodec: 'vp8',
      qualityLevel: 'medium'
    };
  }

  // 获取默认网络统计
  private getDefaultNetworkStats(): NetworkStats {
    return {
      rtt: 50,
      packetLoss: 0,
      jitter: 5,
      bandwidth: 2000,
      networkType: 'wifi',
      signalStrength: 80,
      isStable: true
    };
  }

  // 开始优化
  public startOptimization(): void {
    if (this.isOptimizing) {
      return;
    }

    this.isOptimizing = true;
    console.log('WebRTC 优化器启动');

    // 启动网络监控
    this.startNetworkMonitoring();

    // 启动质量自适应
    if (this.config.enableAdaptiveBitrate) {
      this.startAdaptiveBitrate();
    }
  }

  // 停止优化
  public stopOptimization(): void {
    if (!this.isOptimizing) {
      return;
    }

    this.isOptimizing = false;
    console.log('WebRTC 优化器停止');

    // 停止定时器
    if (this.adaptationTimer) {
      clearInterval(this.adaptationTimer);
      this.adaptationTimer = null;
    }

    if (this.networkMonitorTimer) {
      clearInterval(this.networkMonitorTimer);
      this.networkMonitorTimer = null;
    }
  }

  // 启动网络监控
  private startNetworkMonitoring(): void {
    this.networkMonitorTimer = setInterval(() => {
      this.updateNetworkStats();
    }, this.config.networkCheckInterval);
  }

  // 启动自适应码率
  private startAdaptiveBitrate(): void {
    this.adaptationTimer = setInterval(() => {
      this.adaptToNetworkConditions();
    }, 3000); // 每3秒检查一次
  }

  // 更新网络统计
  private async updateNetworkStats(): Promise<void> {
    try {
      // 获取网络连接信息
      const connection = (navigator as any).connection || 
                        (navigator as any).mozConnection || 
                        (navigator as any).webkitConnection;

      if (connection) {
        this.networkStats.networkType = this.getNetworkType(connection.effectiveType);
        this.networkStats.bandwidth = this.estimateBandwidth(connection);
      }

      // 通过 WebRTC 统计获取更准确的网络信息
      await this.updateWebRTCStats();

      // 更新网络稳定性
      this.updateNetworkStability();

    } catch (error) {
      console.error('更新网络统计失败:', error);
      errorHandler.handleError(error, 'WebRTC网络监控');
    }
  }

  // 更新 WebRTC 统计
  private async updateWebRTCStats(): Promise<void> {
    try {
      // 这里应该从实际的 WebRTC 连接获取统计信息
      // 由于我们无法直接访问 WebRTC 连接，这里提供模拟数据
      const mockStats = this.generateMockNetworkStats();
      
      this.networkStats.rtt = mockStats.rtt;
      this.networkStats.packetLoss = mockStats.packetLoss;
      this.networkStats.jitter = mockStats.jitter;
      this.networkStats.bandwidth = mockStats.bandwidth;

    } catch (error) {
      console.error('获取 WebRTC 统计失败:', error);
    }
  }

  // 生成模拟网络统计（实际应用中应该从 WebRTC 获取）
  private generateMockNetworkStats() {
    const baseRTT = 50;
    const basePacketLoss = 0.5;
    const baseJitter = 5;
    const baseBandwidth = 2000;

    // 添加一些随机变化来模拟网络波动
    const rttVariation = Math.random() * 20 - 10;
    const packetLossVariation = Math.random() * 1 - 0.5;
    const jitterVariation = Math.random() * 3 - 1.5;
    const bandwidthVariation = Math.random() * 500 - 250;

    return {
      rtt: Math.max(10, baseRTT + rttVariation),
      packetLoss: Math.max(0, basePacketLoss + packetLossVariation),
      jitter: Math.max(0, baseJitter + jitterVariation),
      bandwidth: Math.max(100, baseBandwidth + bandwidthVariation)
    };
  }

  // 获取网络类型
  private getNetworkType(effectiveType: string): string {
    switch (effectiveType) {
      case 'slow-2g':
      case '2g':
        return '2g';
      case '3g':
        return '3g';
      case '4g':
        return '4g';
      default:
        return 'wifi';
    }
  }

  // 估算带宽
  private estimateBandwidth(connection: any): number {
    if (connection.downlink) {
      return connection.downlink * 1000; // 转换为 kbps
    }
    return 2000; // 默认值
  }

  // 更新网络稳定性
  private updateNetworkStability(): void {
    const now = Date.now();
    const recentStats = this.qualityHistory.filter(
      stat => now - stat.timestamp < 30000 // 最近30秒
    );

    if (recentStats.length < 3) {
      this.networkStats.isStable = true;
      return;
    }

    // 计算质量分数的标准差
    const qualities = recentStats.map(stat => stat.quality);
    const avg = qualities.reduce((sum, q) => sum + q, 0) / qualities.length;
    const variance = qualities.reduce((sum, q) => sum + Math.pow(q - avg, 2), 0) / qualities.length;
    const stdDev = Math.sqrt(variance);

    // 标准差小于10认为稳定
    this.networkStats.isStable = stdDev < 10;
  }

  // 根据网络条件自适应
  private adaptToNetworkConditions(): void {
    const networkQuality = this.assessNetworkQuality();
    const currentQuality = this.calculateCurrentQuality();

    console.log(`网络质量: ${networkQuality}, 当前质量: ${currentQuality}`);

    // 记录质量历史
    this.qualityHistory.push({
      timestamp: Date.now(),
      quality: currentQuality
    });

    // 限制历史记录数量
    if (this.qualityHistory.length > 100) {
      this.qualityHistory = this.qualityHistory.slice(-100);
    }

    // 根据网络质量调整设置
    this.adjustQualitySettings(networkQuality, currentQuality);
  }

  // 评估网络质量
  private assessNetworkQuality(): NetworkQuality {
    const { rtt, packetLoss, jitter, bandwidth } = this.networkStats;

    // 综合评分
    let score = 100;

    // RTT 评分
    if (rtt > 200) score -= 40;
    else if (rtt > 100) score -= 20;
    else if (rtt > 50) score -= 10;

    // 丢包率评分
    if (packetLoss > 10) score -= 30;
    else if (packetLoss > 5) score -= 20;
    else if (packetLoss > 1) score -= 10;

    // 抖动评分
    if (jitter > 50) score -= 20;
    else if (jitter > 20) score -= 10;

    // 带宽评分
    if (bandwidth < 500) score -= 30;
    else if (bandwidth < 1000) score -= 20;
    else if (bandwidth < 2000) score -= 10;

    // 确定质量等级
    if (score >= 90) return NetworkQuality.EXCELLENT;
    if (score >= 75) return NetworkQuality.GOOD;
    if (score >= 60) return NetworkQuality.FAIR;
    if (score >= 40) return NetworkQuality.POOR;
    return NetworkQuality.VERY_POOR;
  }

  // 计算当前质量
  private calculateCurrentQuality(): number {
    const { rtt, packetLoss, jitter, bandwidth } = this.networkStats;

    let quality = 100;

    // 基于各项指标计算质量分数
    quality -= Math.min(40, rtt * 0.2);
    quality -= Math.min(30, packetLoss * 3);
    quality -= Math.min(20, jitter * 0.4);
    quality -= Math.min(20, Math.max(0, 2000 - bandwidth) * 0.01);

    return Math.max(0, quality);
  }

  // 调整质量设置
  private adjustQualitySettings(networkQuality: NetworkQuality, currentQuality: number): void {
    const newSettings = this.getQualitySettingsForNetwork(networkQuality);

    // 检查是否需要调整
    if (this.shouldAdjustQuality(newSettings, currentQuality)) {
      this.applyQualitySettings(newSettings);
      
      // 发送质量调整事件
      this.emitQualityChange(newSettings);
    }
  }

  // 根据网络质量获取设置
  private getQualitySettingsForNetwork(networkQuality: NetworkQuality): MediaQualitySettings {
    switch (networkQuality) {
      case NetworkQuality.EXCELLENT:
        return {
          videoBitrate: 4000,
          audioBitrate: 192,
          resolution: '1920x1080',
          frameRate: 30,
          audioCodec: 'opus',
          videoCodec: 'vp8',
          qualityLevel: 'high'
        };

      case NetworkQuality.GOOD:
        return {
          videoBitrate: 2000,
          audioBitrate: 128,
          resolution: '1280x720',
          frameRate: 30,
          audioCodec: 'opus',
          videoCodec: 'vp8',
          qualityLevel: 'medium'
        };

      case NetworkQuality.FAIR:
        return {
          videoBitrate: 1000,
          audioBitrate: 96,
          resolution: '854x480',
          frameRate: 24,
          audioCodec: 'opus',
          videoCodec: 'vp8',
          qualityLevel: 'low'
        };

      case NetworkQuality.POOR:
        return {
          videoBitrate: 500,
          audioBitrate: 64,
          resolution: '640x360',
          frameRate: 15,
          audioCodec: 'opus',
          videoCodec: 'vp8',
          qualityLevel: 'very_low'
        };

      case NetworkQuality.VERY_POOR:
        return {
          videoBitrate: 0,
          audioBitrate: 64,
          resolution: '0x0',
          frameRate: 0,
          audioCodec: 'opus',
          videoCodec: '',
          qualityLevel: 'audio_only'
        };

      default:
        return this.currentQuality;
    }
  }

  // 判断是否需要调整质量
  private shouldAdjustQuality(newSettings: MediaQualitySettings, currentQuality: number): boolean {
    // 如果质量差异较大，需要调整
    const qualityDiff = Math.abs(this.calculateQualityScore(newSettings) - currentQuality);
    
    // 如果质量变化超过阈值，进行调整
    return qualityDiff > 15;
  }

  // 计算质量分数
  private calculateQualityScore(settings: MediaQualitySettings): number {
    let score = 0;

    // 基于码率计算分数
    score += Math.min(40, settings.videoBitrate * 0.01);
    score += Math.min(30, settings.audioBitrate * 0.3);

    // 基于分辨率计算分数
    const [width, height] = settings.resolution.split('x').map(Number);
    const pixels = width * height;
    score += Math.min(20, pixels / 10000);

    // 基于帧率计算分数
    score += Math.min(10, settings.frameRate * 0.3);

    return Math.min(100, score);
  }

  // 应用质量设置
  private applyQualitySettings(settings: MediaQualitySettings): void {
    console.log('应用新的质量设置:', settings);
    
    this.currentQuality = settings;

    // 这里应该实际应用到 WebRTC 连接
    // 例如调整编码器参数、分辨率等
    this.applyToWebRTCConnection(settings);
  }

  // 应用到 WebRTC 连接
  private applyToWebRTCConnection(settings: MediaQualitySettings): void {
    try {
      // 这里应该实现实际的 WebRTC 参数调整
      // 由于我们无法直接访问 WebRTC 连接，这里提供接口

      // 调整视频编码器
      if (settings.videoBitrate > 0) {
        this.adjustVideoEncoder(settings);
      }

      // 调整音频编码器
      this.adjustAudioEncoder(settings);

      // 调整分辨率
      this.adjustResolution(settings);

    } catch (error) {
      console.error('应用 WebRTC 设置失败:', error);
      errorHandler.handleError(error, 'WebRTC质量调整');
    }
  }

  // 调整视频编码器
  private adjustVideoEncoder(settings: MediaQualitySettings): void {
    console.log(`调整视频编码器: 码率=${settings.videoBitrate}kbps, 分辨率=${settings.resolution}, 帧率=${settings.frameRate}`);
    
    // 实际实现中，这里应该调用 WebRTC 的编码器设置方法
    // 例如：peerConnection.getSenders()[0].setParameters({...})
  }

  // 调整音频编码器
  private adjustAudioEncoder(settings: MediaQualitySettings): void {
    console.log(`调整音频编码器: 码率=${settings.audioBitrate}kbps, 编解码器=${settings.audioCodec}`);
    
    // 实际实现中，这里应该调用 WebRTC 的音频编码器设置方法
  }

  // 调整分辨率
  private adjustResolution(settings: MediaQualitySettings): void {
    console.log(`调整分辨率: ${settings.resolution}`);
    
    // 实际实现中，这里应该调用 WebRTC 的分辨率设置方法
  }

  // 发送质量变化事件
  private emitQualityChange(settings: MediaQualitySettings): void {
    const event = new CustomEvent('webrtc-quality-change', {
      detail: {
        settings,
        networkStats: this.networkStats,
        timestamp: Date.now()
      }
    });

    window.dispatchEvent(event);
  }

  // 获取当前质量设置
  public getCurrentQuality(): MediaQualitySettings {
    return { ...this.currentQuality };
  }

  // 获取网络统计
  public getNetworkStats(): NetworkStats {
    return { ...this.networkStats };
  }

  // 获取质量历史
  public getQualityHistory(): Array<{ timestamp: number; quality: number }> {
    return [...this.qualityHistory];
  }

  // 手动设置质量
  public setQuality(settings: Partial<MediaQualitySettings>): void {
    const newSettings = { ...this.currentQuality, ...settings };
    this.applyQualitySettings(newSettings);
  }

  // 获取优化建议
  public getOptimizationSuggestions(): string[] {
    const suggestions: string[] = [];
    const { rtt, packetLoss, jitter, bandwidth, networkType } = this.networkStats;

    if (rtt > 200) {
      suggestions.push('网络延迟较高，建议切换到更稳定的网络');
    }

    if (packetLoss > 5) {
      suggestions.push('网络丢包率较高，建议检查网络连接');
    }

    if (jitter > 50) {
      suggestions.push('网络抖动严重，建议使用有线网络');
    }

    if (bandwidth < 1000) {
      suggestions.push('带宽不足，建议关闭其他占用网络的应用程序');
    }

    if (networkType === '2g' || networkType === '3g') {
      suggestions.push('移动网络质量有限，建议在WiFi环境下使用视频通话');
    }

    if (suggestions.length === 0) {
      suggestions.push('网络质量良好');
    }

    return suggestions;
  }

  // 启用/禁用特定优化
  public toggleOptimization(optimization: keyof WebRTCOptimizerConfig, enabled: boolean): void {
    this.config[optimization] = enabled;

    if (optimization === 'enableAdaptiveBitrate') {
      if (enabled) {
        this.startAdaptiveBitrate();
      } else {
        if (this.adaptationTimer) {
          clearInterval(this.adaptationTimer);
          this.adaptationTimer = null;
        }
      }
    }
  }

  // 获取配置
  public getConfig(): WebRTCOptimizerConfig {
    return { ...this.config };
  }

  // 更新配置
  public updateConfig(newConfig: Partial<WebRTCOptimizerConfig>): void {
    this.config = { ...this.config, ...newConfig };
  }
}

// 创建全局实例
export const webrtcOptimizer = WebRTCOptimizer.getInstance();

// 默认导出
export default webrtcOptimizer;


