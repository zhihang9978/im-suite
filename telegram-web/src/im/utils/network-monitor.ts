/**
 * 志航密信网络状态监控器
 * 提供网络连接状态、质量监控、断线重连等功能
 */

import { errorHandler } from './error-handler';

// 网络连接状态
export enum NetworkStatus {
  ONLINE = 'online',
  OFFLINE = 'offline',
  SLOW = 'slow',
  UNSTABLE = 'unstable',
  RECONNECTING = 'reconnecting'
}

// 网络类型
export enum NetworkType {
  WIFI = 'wifi',
  ETHERNET = 'ethernet',
  CELLULAR_4G = '4g',
  CELLULAR_3G = '3g',
  CELLULAR_2G = '2g',
  UNKNOWN = 'unknown'
}

// 网络质量等级
export enum NetworkQuality {
  EXCELLENT = 'excellent',
  GOOD = 'good',
  FAIR = 'fair',
  POOR = 'poor',
  VERY_POOR = 'very_poor'
}

// 网络监控配置
export interface NetworkMonitorConfig {
  checkInterval: number;
  timeoutThreshold: number;
  qualityCheckInterval: number;
  reconnectAttempts: number;
  reconnectDelay: number;
  qualityThresholds: {
    excellent: number;
    good: number;
    fair: number;
    poor: number;
  };
}

// 网络统计信息
export interface NetworkStats {
  status: NetworkStatus;
  type: NetworkType;
  quality: NetworkQuality;
  rtt: number;
  packetLoss: number;
  jitter: number;
  bandwidth: number;
  signalStrength: number;
  isStable: boolean;
  lastCheck: number;
  uptime: number;
  downtime: number;
}

// 网络事件
export interface NetworkEvent {
  type: 'status_change' | 'quality_change' | 'connection_lost' | 'connection_restored';
  timestamp: number;
  data: any;
}

// 网络监控器类
export class NetworkMonitor {
  private static instance: NetworkMonitor;
  private config: NetworkMonitorConfig;
  private currentStats: NetworkStats;
  private isMonitoring = false;
  private monitoringTimer: NodeJS.Timeout | null = null;
  private qualityTimer: NodeJS.Timeout | null = null;
  private eventListeners: Array<(event: NetworkEvent) => void> = [];
  private connectionStartTime: number = 0;
  private lastOnlineTime: number = 0;
  private lastOfflineTime: number = 0;
  private reconnectAttempts = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;

  private constructor(config: Partial<NetworkMonitorConfig> = {}) {
    this.config = {
      checkInterval: 5000,
      timeoutThreshold: 10000,
      qualityCheckInterval: 10000,
      reconnectAttempts: 5,
      reconnectDelay: 2000,
      qualityThresholds: {
        excellent: 90,
        good: 75,
        fair: 60,
        poor: 40
      },
      ...config
    };

    this.currentStats = this.getDefaultStats();
    this.setupEventListeners();
  }

  // 获取单例实例
  public static getInstance(config?: Partial<NetworkMonitorConfig>): NetworkMonitor {
    if (!NetworkMonitor.instance) {
      NetworkMonitor.instance = new NetworkMonitor(config);
    }
    return NetworkMonitor.instance;
  }

  // 获取默认统计信息
  private getDefaultStats(): NetworkStats {
    return {
      status: NetworkStatus.ONLINE,
      type: NetworkType.UNKNOWN,
      quality: NetworkQuality.GOOD,
      rtt: 50,
      packetLoss: 0,
      jitter: 5,
      bandwidth: 2000,
      signalStrength: 80,
      isStable: true,
      lastCheck: Date.now(),
      uptime: 0,
      downtime: 0
    };
  }

  // 设置事件监听器
  private setupEventListeners(): void {
    // 监听浏览器网络状态变化
    window.addEventListener('online', this.handleOnline.bind(this));
    window.addEventListener('offline', this.handleOffline.bind(this));

    // 监听页面可见性变化
    document.addEventListener('visibilitychange', this.handleVisibilityChange.bind(this));

    // 监听网络连接变化
    if ('connection' in navigator) {
      const connection = (navigator as any).connection;
      connection.addEventListener('change', this.handleConnectionChange.bind(this));
    }
  }

  // 开始监控
  public startMonitoring(): void {
    if (this.isMonitoring) {
      return;
    }

    this.isMonitoring = true;
    this.connectionStartTime = Date.now();

    // 启动状态检查
    this.monitoringTimer = setInterval(() => {
      this.checkNetworkStatus();
    }, this.config.checkInterval);

    // 启动质量检查
    this.qualityTimer = setInterval(() => {
      this.checkNetworkQuality();
    }, this.config.qualityCheckInterval);

    console.log('网络监控启动');
  }

  // 停止监控
  public stopMonitoring(): void {
    if (!this.isMonitoring) {
      return;
    }

    this.isMonitoring = false;

    if (this.monitoringTimer) {
      clearInterval(this.monitoringTimer);
      this.monitoringTimer = null;
    }

    if (this.qualityTimer) {
      clearInterval(this.qualityTimer);
      this.qualityTimer = null;
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    console.log('网络监控停止');
  }

  // 检查网络状态
  private async checkNetworkStatus(): Promise<void> {
    try {
      const startTime = Date.now();
      
      // 使用 fetch 检查网络连接
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.config.timeoutThreshold);
      
      const response = await fetch('/api/ping', {
        method: 'GET',
        signal: controller.signal,
        cache: 'no-cache'
      });

      clearTimeout(timeoutId);
      
      const endTime = Date.now();
      const rtt = endTime - startTime;

      if (response.ok) {
        this.handleConnectionSuccess(rtt);
      } else {
        this.handleConnectionFailure();
      }

    } catch (error) {
      this.handleConnectionFailure();
    }
  }

  // 检查网络质量
  private async checkNetworkQuality(): Promise<void> {
    try {
      // 获取网络连接信息
      const connection = (navigator as any).connection || 
                        (navigator as any).mozConnection || 
                        (navigator as any).webkitConnection;

      if (connection) {
        this.updateNetworkInfo(connection);
      }

      // 执行网络质量测试
      await this.performQualityTest();

    } catch (error) {
      console.error('网络质量检查失败:', error);
      errorHandler.handleError(error, '网络质量监控');
    }
  }

  // 更新网络信息
  private updateNetworkInfo(connection: any): void {
    const oldType = this.currentStats.type;
    const oldBandwidth = this.currentStats.bandwidth;

    // 更新网络类型
    this.currentStats.type = this.getNetworkType(connection.effectiveType);
    
    // 更新带宽
    if (connection.downlink) {
      this.currentStats.bandwidth = connection.downlink * 1000; // 转换为 kbps
    }

    // 更新信号强度
    if (connection.rtt) {
      this.currentStats.rtt = connection.rtt;
    }

    // 检查是否有变化
    if (oldType !== this.currentStats.type || oldBandwidth !== this.currentStats.bandwidth) {
      this.emitEvent({
        type: 'quality_change',
        timestamp: Date.now(),
        data: {
          oldType,
          newType: this.currentStats.type,
          oldBandwidth,
          newBandwidth: this.currentStats.bandwidth
        }
      });
    }
  }

  // 获取网络类型
  private getNetworkType(effectiveType: string): NetworkType {
    switch (effectiveType) {
      case 'slow-2g':
      case '2g':
        return NetworkType.CELLULAR_2G;
      case '3g':
        return NetworkType.CELLULAR_3G;
      case '4g':
        return NetworkType.CELLULAR_4G;
      default:
        return NetworkType.WIFI;
    }
  }

  // 执行网络质量测试
  private async performQualityTest(): Promise<void> {
    try {
      // 执行多次 ping 测试
      const pingResults: number[] = [];
      const testCount = 3;

      for (let i = 0; i < testCount; i++) {
        const startTime = Date.now();
        
        try {
          const response = await fetch('/api/ping', {
            method: 'GET',
            cache: 'no-cache'
          });
          
          if (response.ok) {
            const endTime = Date.now();
            pingResults.push(endTime - startTime);
          }
        } catch (error) {
          // 忽略单次测试失败
        }
      }

      if (pingResults.length > 0) {
        // 计算平均 RTT
        const avgRTT = pingResults.reduce((sum, rtt) => sum + rtt, 0) / pingResults.length;
        this.currentStats.rtt = avgRTT;

        // 计算抖动
        const variance = pingResults.reduce((sum, rtt) => sum + Math.pow(rtt - avgRTT, 2), 0) / pingResults.length;
        this.currentStats.jitter = Math.sqrt(variance);

        // 更新网络质量
        this.updateNetworkQuality();
      }

    } catch (error) {
      console.error('网络质量测试失败:', error);
    }
  }

  // 更新网络质量
  private updateNetworkQuality(): void {
    const quality = this.calculateNetworkQuality();
    const oldQuality = this.currentStats.quality;

    this.currentStats.quality = quality;

    if (oldQuality !== quality) {
      this.emitEvent({
        type: 'quality_change',
        timestamp: Date.now(),
        data: {
          oldQuality,
          newQuality: quality,
          stats: { ...this.currentStats }
        }
      });
    }
  }

  // 计算网络质量
  private calculateNetworkQuality(): NetworkQuality {
    let score = 100;

    // RTT 评分
    if (this.currentStats.rtt > 500) score -= 40;
    else if (this.currentStats.rtt > 200) score -= 30;
    else if (this.currentStats.rtt > 100) score -= 20;
    else if (this.currentStats.rtt > 50) score -= 10;

    // 抖动评分
    if (this.currentStats.jitter > 100) score -= 30;
    else if (this.currentStats.jitter > 50) score -= 20;
    else if (this.currentStats.jitter > 20) score -= 10;

    // 带宽评分
    if (this.currentStats.bandwidth < 500) score -= 40;
    else if (this.currentStats.bandwidth < 1000) score -= 30;
    else if (this.currentStats.bandwidth < 2000) score -= 20;
    else if (this.currentStats.bandwidth < 5000) score -= 10;

    // 丢包率评分
    if (this.currentStats.packetLoss > 10) score -= 30;
    else if (this.currentStats.packetLoss > 5) score -= 20;
    else if (this.currentStats.packetLoss > 1) score -= 10;

    // 确定质量等级
    if (score >= this.config.qualityThresholds.excellent) return NetworkQuality.EXCELLENT;
    if (score >= this.config.qualityThresholds.good) return NetworkQuality.GOOD;
    if (score >= this.config.qualityThresholds.fair) return NetworkQuality.FAIR;
    if (score >= this.config.qualityThresholds.poor) return NetworkQuality.POOR;
    return NetworkQuality.VERY_POOR;
  }

  // 处理连接成功
  private handleConnectionSuccess(rtt: number): void {
    const wasOffline = this.currentStats.status === NetworkStatus.OFFLINE;
    
    this.currentStats.status = NetworkStatus.ONLINE;
    this.currentStats.rtt = rtt;
    this.currentStats.lastCheck = Date.now();

    if (wasOffline) {
      this.lastOnlineTime = Date.now();
      this.reconnectAttempts = 0;
      
      this.emitEvent({
        type: 'connection_restored',
        timestamp: Date.now(),
        data: { rtt }
      });
    }

    // 更新稳定性
    this.updateStability();
  }

  // 处理连接失败
  private handleConnectionFailure(): void {
    const wasOnline = this.currentStats.status === NetworkStatus.ONLINE;
    
    if (wasOnline) {
      this.currentStats.status = NetworkStatus.OFFLINE;
      this.lastOfflineTime = Date.now();
      
      this.emitEvent({
        type: 'connection_lost',
        timestamp: Date.now(),
        data: {}
      });

      // 开始重连
      this.startReconnection();
    }
  }

  // 开始重连
  private startReconnection(): void {
    if (this.reconnectAttempts >= this.config.reconnectAttempts) {
      console.log('重连次数已达上限，停止重连');
      return;
    }

    this.reconnectAttempts++;
    this.currentStats.status = NetworkStatus.RECONNECTING;

    this.reconnectTimer = setTimeout(() => {
      this.checkNetworkStatus();
    }, this.config.reconnectDelay * this.reconnectAttempts);

    console.log(`开始第 ${this.reconnectAttempts} 次重连`);
  }

  // 更新稳定性
  private updateStability(): void {
    // 简单的稳定性检测
    const now = Date.now();
    const timeSinceLastCheck = now - this.currentStats.lastCheck;
    
    // 如果检查间隔正常，认为网络稳定
    this.currentStats.isStable = timeSinceLastCheck < this.config.checkInterval * 2;
  }

  // 处理在线事件
  private handleOnline(): void {
    console.log('网络连接恢复');
    this.handleConnectionSuccess(0);
  }

  // 处理离线事件
  private handleOffline(): void {
    console.log('网络连接断开');
    this.handleConnectionFailure();
  }

  // 处理连接变化
  private handleConnectionChange(): void {
    console.log('网络连接类型变化');
    this.checkNetworkQuality();
  }

  // 处理页面可见性变化
  private handleVisibilityChange(): void {
    if (document.hidden) {
      // 页面隐藏时减少检查频率
      this.pauseMonitoring();
    } else {
      // 页面显示时恢复检查
      this.resumeMonitoring();
    }
  }

  // 暂停监控
  private pauseMonitoring(): void {
    if (this.monitoringTimer) {
      clearInterval(this.monitoringTimer);
      this.monitoringTimer = null;
    }
  }

  // 恢复监控
  private resumeMonitoring(): void {
    if (this.isMonitoring && !this.monitoringTimer) {
      this.monitoringTimer = setInterval(() => {
        this.checkNetworkStatus();
      }, this.config.checkInterval);
    }
  }

  // 发送事件
  private emitEvent(event: NetworkEvent): void {
    this.eventListeners.forEach(listener => {
      try {
        listener(event);
      } catch (error) {
        console.error('网络事件监听器执行失败:', error);
      }
    });
  }

  // 添加事件监听器
  public addEventListener(listener: (event: NetworkEvent) => void): void {
    this.eventListeners.push(listener);
  }

  // 移除事件监听器
  public removeEventListener(listener: (event: NetworkEvent) => void): void {
    const index = this.eventListeners.indexOf(listener);
    if (index > -1) {
      this.eventListeners.splice(index, 1);
    }
  }

  // 获取当前网络统计
  public getNetworkStats(): NetworkStats {
    const now = Date.now();
    const stats = { ...this.currentStats };
    
    // 更新运行时间
    if (this.connectionStartTime > 0) {
      if (stats.status === NetworkStatus.ONLINE) {
        stats.uptime = now - this.connectionStartTime;
      } else {
        stats.downtime = now - this.connectionStartTime;
      }
    }

    return stats;
  }

  // 获取网络质量建议
  public getNetworkRecommendations(): string[] {
    const recommendations: string[] = [];
    const { quality, type, rtt, packetLoss, bandwidth } = this.currentStats;

    if (quality === NetworkQuality.VERY_POOR || quality === NetworkQuality.POOR) {
      recommendations.push('网络质量较差，建议切换到更稳定的网络');
    }

    if (rtt > 200) {
      recommendations.push('网络延迟较高，建议使用有线网络');
    }

    if (packetLoss > 5) {
      recommendations.push('网络丢包率较高，建议检查网络连接');
    }

    if (bandwidth < 1000) {
      recommendations.push('网络带宽不足，建议关闭其他占用网络的应用程序');
    }

    if (type === NetworkType.CELLULAR_2G || type === NetworkType.CELLULAR_3G) {
      recommendations.push('移动网络质量有限，建议在WiFi环境下使用');
    }

    if (recommendations.length === 0) {
      recommendations.push('网络质量良好');
    }

    return recommendations;
  }

  // 强制检查网络状态
  public forceCheck(): void {
    this.checkNetworkStatus();
    this.checkNetworkQuality();
  }

  // 获取配置
  public getConfig(): NetworkMonitorConfig {
    return { ...this.config };
  }

  // 更新配置
  public updateConfig(newConfig: Partial<NetworkMonitorConfig>): void {
    this.config = { ...this.config, ...newConfig };
  }

  // 检查浏览器支持
  public static isSupported(): boolean {
    return 'fetch' in window && 'Promise' in window;
  }
}

// 创建全局实例
export const networkMonitor = NetworkMonitor.getInstance();

// 默认导出
export default networkMonitor;
