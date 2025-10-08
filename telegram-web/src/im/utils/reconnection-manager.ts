/**
 * 志航密信重连管理器
 * 提供WebRTC连接断开后的自动重连、重连策略、连接恢复等功能
 */

import { errorHandler } from './error-handler';
import { networkMonitor } from './network-monitor';

// 重连状态
export enum ReconnectionStatus {
  CONNECTED = 'connected',
  CONNECTING = 'connecting',
  DISCONNECTED = 'disconnected',
  RECONNECTING = 'reconnecting',
  FAILED = 'failed',
  SUSPENDED = 'suspended'
}

// 重连策略
export enum ReconnectionStrategy {
  IMMEDIATE = 'immediate',           // 立即重连
  EXPONENTIAL_BACKOFF = 'exponential_backoff', // 指数退避
  LINEAR_BACKOFF = 'linear_backoff', // 线性退避
  FIXED_DELAY = 'fixed_delay',      // 固定延迟
  ADAPTIVE = 'adaptive'             // 自适应
}

// 重连配置
export interface ReconnectionConfig {
  maxAttempts: number;              // 最大重连次数
  baseDelay: number;                // 基础延迟时间 (ms)
  maxDelay: number;                 // 最大延迟时间 (ms)
  backoffMultiplier: number;        // 退避乘数
  strategy: ReconnectionStrategy;   // 重连策略
  enableAdaptiveDelay: boolean;     // 启用自适应延迟
  networkQualityThreshold: number;  // 网络质量阈值
  suspendOnPoorNetwork: boolean;    // 网络差时暂停重连
  maxSuspendTime: number;           // 最大暂停时间 (ms)
}

// 重连统计
export interface ReconnectionStats {
  totalAttempts: number;
  successfulReconnections: number;
  failedReconnections: number;
  averageReconnectTime: number;
  lastReconnectTime: number;
  consecutiveFailures: number;
  currentDelay: number;
  networkQualityAtFailure: number;
}

// 重连事件
export interface ReconnectionEvent {
  type: 'attempt' | 'success' | 'failure' | 'suspend' | 'resume';
  timestamp: number;
  attempt: number;
  delay: number;
  reason?: string;
  networkQuality?: number;
}

// 重连管理器类
export class ReconnectionManager {
  private static instance: ReconnectionManager;
  private config: ReconnectionConfig;
  private status: ReconnectionStatus = ReconnectionStatus.CONNECTED;
  private currentAttempt = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private suspendTimer: NodeJS.Timeout | null = null;
  private stats: ReconnectionStats;
  private eventListeners: Array<(event: ReconnectionEvent) => void> = [];
  private connectionStartTime: number = 0;
  private lastFailureTime: number = 0;
  private lastSuccessTime: number = 0;

  private constructor(config: Partial<ReconnectionConfig> = {}) {
    this.config = {
      maxAttempts: 10,
      baseDelay: 1000,
      maxDelay: 30000,
      backoffMultiplier: 2,
      strategy: ReconnectionStrategy.EXPONENTIAL_BACKOFF,
      enableAdaptiveDelay: true,
      networkQualityThreshold: 60,
      suspendOnPoorNetwork: true,
      maxSuspendTime: 60000,
      ...config
    };

    this.stats = this.getDefaultStats();
    this.setupNetworkMonitoring();
  }

  // 获取单例实例
  public static getInstance(config?: Partial<ReconnectionConfig>): ReconnectionManager {
    if (!ReconnectionManager.instance) {
      ReconnectionManager.instance = new ReconnectionManager(config);
    }
    return ReconnectionManager.instance;
  }

  // 获取默认统计
  private getDefaultStats(): ReconnectionStats {
    return {
      totalAttempts: 0,
      successfulReconnections: 0,
      failedReconnections: 0,
      averageReconnectTime: 0,
      lastReconnectTime: 0,
      consecutiveFailures: 0,
      currentDelay: this.config.baseDelay,
      networkQualityAtFailure: 0
    };
  }

  // 设置网络监控
  private setupNetworkMonitoring(): void {
    networkMonitor.addEventListener((event) => {
      if (event.type === 'status_change') {
        this.handleNetworkStatusChange(event.data);
      } else if (event.type === 'quality_change') {
        this.handleNetworkQualityChange(event.data);
      }
    });
  }

  // 处理网络状态变化
  private handleNetworkStatusChange(data: any): void {
    if (data.status === 'offline') {
      this.handleDisconnection('网络断开');
    } else if (data.status === 'online' && this.status === ReconnectionStatus.DISCONNECTED) {
      this.startReconnection();
    }
  }

  // 处理网络质量变化
  private handleNetworkQualityChange(data: any): void {
    if (this.status === ReconnectionStatus.SUSPENDED && data.newQuality >= this.config.networkQualityThreshold) {
      this.resumeReconnection();
    }
  }

  // 开始重连
  public startReconnection(reason: string = '连接断开'): void {
    if (this.status === ReconnectionStatus.RECONNECTING) {
      return;
    }

    this.status = ReconnectionStatus.RECONNECTING;
    this.currentAttempt = 0;
    this.connectionStartTime = Date.now();

    console.log(`开始重连: ${reason}`);
    this.emitEvent({
      type: 'attempt',
      timestamp: Date.now(),
      attempt: 0,
      delay: 0,
      reason
    });

    this.attemptReconnection();
  }

  // 尝试重连
  private attemptReconnection(): void {
    if (this.currentAttempt >= this.config.maxAttempts) {
      this.handleMaxAttemptsReached();
      return;
    }

    // 检查网络质量
    if (this.shouldSuspendReconnection()) {
      this.suspendReconnection();
      return;
    }

    this.currentAttempt++;
    this.stats.totalAttempts++;

    // 计算重连延迟
    const delay = this.calculateReconnectionDelay();
    this.stats.currentDelay = delay;

    console.log(`重连尝试 ${this.currentAttempt}/${this.config.maxAttempts}, 延迟: ${delay}ms`);

    this.emitEvent({
      type: 'attempt',
      timestamp: Date.now(),
      attempt: this.currentAttempt,
      delay,
      networkQuality: this.getCurrentNetworkQuality()
    });

    // 设置重连定时器
    this.reconnectTimer = setTimeout(() => {
      this.executeReconnection();
    }, delay);
  }

  // 执行重连
  private async executeReconnection(): Promise<void> {
    try {
      const startTime = Date.now();
      
      // 执行实际的重连逻辑
      const success = await this.performReconnection();
      
      const reconnectTime = Date.now() - startTime;

      if (success) {
        this.handleReconnectionSuccess(reconnectTime);
      } else {
        this.handleReconnectionFailure(reconnectTime);
      }

    } catch (error) {
      console.error('重连执行失败:', error);
      errorHandler.handleError(error, '重连执行');
      this.handleReconnectionFailure(0);
    }
  }

  // 执行实际的重连操作
  private async performReconnection(): Promise<boolean> {
    try {
      // 这里应该实现实际的重连逻辑
      // 例如重新建立WebRTC连接、重新加入房间等
      
      // 模拟重连过程
      await this.simulateReconnection();
      
      return true;
    } catch (error) {
      console.error('重连操作失败:', error);
      return false;
    }
  }

  // 模拟重连过程
  private async simulateReconnection(): Promise<void> {
    // 模拟网络请求
    const response = await fetch('/api/reconnect', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        attempt: this.currentAttempt,
        timestamp: Date.now()
      })
    });

    if (!response.ok) {
      throw new Error(`重连请求失败: ${response.status}`);
    }

    // 模拟连接建立时间
    await new Promise(resolve => setTimeout(resolve, 1000 + Math.random() * 2000));
  }

  // 处理重连成功
  private handleReconnectionSuccess(reconnectTime: number): void {
    this.status = ReconnectionStatus.CONNECTED;
    this.currentAttempt = 0;
    this.lastSuccessTime = Date.now();
    this.stats.successfulReconnections++;
    this.stats.consecutiveFailures = 0;
    this.stats.lastReconnectTime = reconnectTime;
    this.stats.currentDelay = this.config.baseDelay;

    // 更新平均重连时间
    const totalTime = this.stats.averageReconnectTime * (this.stats.successfulReconnections - 1) + reconnectTime;
    this.stats.averageReconnectTime = totalTime / this.stats.successfulReconnections;

    console.log(`重连成功，耗时: ${reconnectTime}ms`);

    this.emitEvent({
      type: 'success',
      timestamp: Date.now(),
      attempt: this.currentAttempt,
      delay: 0,
      reason: '重连成功'
    });
  }

  // 处理重连失败
  private handleReconnectionFailure(reconnectTime: number): void {
    this.stats.failedReconnections++;
    this.stats.consecutiveFailures++;
    this.lastFailureTime = Date.now();

    console.log(`重连失败，尝试 ${this.currentAttempt}/${this.config.maxAttempts}`);

    this.emitEvent({
      type: 'failure',
      timestamp: Date.now(),
      attempt: this.currentAttempt,
      delay: 0,
      reason: '重连失败'
    });

    // 继续尝试重连
    this.attemptReconnection();
  }

  // 处理达到最大重连次数
  private handleMaxAttemptsReached(): void {
    this.status = ReconnectionStatus.FAILED;

    console.log('达到最大重连次数，停止重连');

    this.emitEvent({
      type: 'failure',
      timestamp: Date.now(),
      attempt: this.currentAttempt,
      delay: 0,
      reason: '达到最大重连次数'
    });
  }

  // 计算重连延迟
  private calculateReconnectionDelay(): number {
    let delay: number;

    switch (this.config.strategy) {
      case ReconnectionStrategy.IMMEDIATE:
        delay = 0;
        break;

      case ReconnectionStrategy.FIXED_DELAY:
        delay = this.config.baseDelay;
        break;

      case ReconnectionStrategy.LINEAR_BACKOFF:
        delay = this.config.baseDelay * this.currentAttempt;
        break;

      case ReconnectionStrategy.EXPONENTIAL_BACKOFF:
        delay = this.config.baseDelay * Math.pow(this.config.backoffMultiplier, this.currentAttempt - 1);
        break;

      case ReconnectionStrategy.ADAPTIVE:
        delay = this.calculateAdaptiveDelay();
        break;

      default:
        delay = this.config.baseDelay;
    }

    // 限制最大延迟
    delay = Math.min(delay, this.config.maxDelay);

    // 应用网络质量调整
    if (this.config.enableAdaptiveDelay) {
      delay = this.adjustDelayByNetworkQuality(delay);
    }

    return Math.max(0, delay);
  }

  // 计算自适应延迟
  private calculateAdaptiveDelay(): number {
    const baseDelay = this.config.baseDelay;
    const networkQuality = this.getCurrentNetworkQuality();
    
    // 根据网络质量调整延迟
    let multiplier = 1;
    if (networkQuality < 30) {
      multiplier = 3; // 网络很差，延迟更长
    } else if (networkQuality < 60) {
      multiplier = 2; // 网络一般，延迟稍长
    } else if (networkQuality > 80) {
      multiplier = 0.5; // 网络很好，延迟更短
    }

    // 根据连续失败次数调整
    const failureMultiplier = Math.min(3, this.stats.consecutiveFailures * 0.5);

    return baseDelay * multiplier * (1 + failureMultiplier);
  }

  // 根据网络质量调整延迟
  private adjustDelayByNetworkQuality(delay: number): number {
    const networkQuality = this.getCurrentNetworkQuality();
    
    if (networkQuality < 30) {
      return delay * 2; // 网络很差，延迟翻倍
    } else if (networkQuality > 80) {
      return delay * 0.5; // 网络很好，延迟减半
    }
    
    return delay;
  }

  // 获取当前网络质量
  private getCurrentNetworkQuality(): number {
    const stats = networkMonitor.getNetworkStats();
    return stats.quality === 'excellent' ? 90 :
           stats.quality === 'good' ? 75 :
           stats.quality === 'fair' ? 60 :
           stats.quality === 'poor' ? 40 : 20;
  }

  // 判断是否应该暂停重连
  private shouldSuspendReconnection(): boolean {
    if (!this.config.suspendOnPoorNetwork) {
      return false;
    }

    const networkQuality = this.getCurrentNetworkQuality();
    return networkQuality < this.config.networkQualityThreshold;
  }

  // 暂停重连
  private suspendReconnection(): void {
    this.status = ReconnectionStatus.SUSPENDED;

    console.log('网络质量较差，暂停重连');

    this.emitEvent({
      type: 'suspend',
      timestamp: Date.now(),
      attempt: this.currentAttempt,
      delay: 0,
      reason: '网络质量较差'
    });

    // 设置恢复定时器
    this.suspendTimer = setTimeout(() => {
      this.resumeReconnection();
    }, this.config.maxSuspendTime);
  }

  // 恢复重连
  private resumeReconnection(): void {
    if (this.status !== ReconnectionStatus.SUSPENDED) {
      return;
    }

    this.status = ReconnectionStatus.RECONNECTING;

    console.log('网络质量改善，恢复重连');

    this.emitEvent({
      type: 'resume',
      timestamp: Date.now(),
      attempt: this.currentAttempt,
      delay: 0,
      reason: '网络质量改善'
    });

    this.attemptReconnection();
  }

  // 处理断开连接
  private handleDisconnection(reason: string): void {
    this.status = ReconnectionStatus.DISCONNECTED;
    this.lastFailureTime = Date.now();

    console.log(`连接断开: ${reason}`);

    // 清除定时器
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.suspendTimer) {
      clearTimeout(this.suspendTimer);
      this.suspendTimer = null;
    }
  }

  // 停止重连
  public stopReconnection(): void {
    this.status = ReconnectionStatus.DISCONNECTED;
    this.currentAttempt = 0;

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.suspendTimer) {
      clearTimeout(this.suspendTimer);
      this.suspendTimer = null;
    }

    console.log('停止重连');
  }

  // 重置重连状态
  public reset(): void {
    this.stopReconnection();
    this.status = ReconnectionStatus.CONNECTED;
    this.stats = this.getDefaultStats();
    this.currentAttempt = 0;
  }

  // 添加事件监听器
  public addEventListener(listener: (event: ReconnectionEvent) => void): void {
    this.eventListeners.push(listener);
  }

  // 移除事件监听器
  public removeEventListener(listener: (event: ReconnectionEvent) => void): void {
    const index = this.eventListeners.indexOf(listener);
    if (index > -1) {
      this.eventListeners.splice(index, 1);
    }
  }

  // 发送事件
  private emitEvent(event: ReconnectionEvent): void {
    this.eventListeners.forEach(listener => {
      try {
        listener(event);
      } catch (error) {
        console.error('重连事件监听器执行失败:', error);
      }
    });
  }

  // 获取当前状态
  public getStatus(): ReconnectionStatus {
    return this.status;
  }

  // 获取统计信息
  public getStats(): ReconnectionStats {
    return { ...this.stats };
  }

  // 获取配置
  public getConfig(): ReconnectionConfig {
    return { ...this.config };
  }

  // 更新配置
  public updateConfig(newConfig: Partial<ReconnectionConfig>): void {
    this.config = { ...this.config, ...newConfig };
  }

  // 获取重连建议
  public getReconnectionAdvice(): string[] {
    const advice: string[] = [];

    if (this.stats.consecutiveFailures > 5) {
      advice.push('连续重连失败次数较多，建议检查网络连接');
    }

    if (this.stats.averageReconnectTime > 10000) {
      advice.push('重连时间较长，建议优化网络环境');
    }

    const networkQuality = this.getCurrentNetworkQuality();
    if (networkQuality < 50) {
      advice.push('网络质量较差，建议切换到更稳定的网络');
    }

    if (this.status === ReconnectionStatus.SUSPENDED) {
      advice.push('重连已暂停，等待网络质量改善');
    }

    if (this.status === ReconnectionStatus.FAILED) {
      advice.push('重连失败，请手动检查连接');
    }

    if (advice.length === 0) {
      advice.push('重连状态正常');
    }

    return advice;
  }

  // 检查是否支持重连
  public static isSupported(): boolean {
    return 'fetch' in window && 'Promise' in window;
  }
}

// 创建全局实例
export const reconnectionManager = ReconnectionManager.getInstance();

// 默认导出
export default reconnectionManager;


