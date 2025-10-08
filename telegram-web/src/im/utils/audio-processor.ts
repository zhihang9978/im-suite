/**
 * 志航密信音频处理器
 * 提供音频预处理、降噪、回声消除、自动增益控制等功能
 */

// 音频处理器配置
export interface AudioProcessorConfig {
  enableNoiseReduction: boolean;
  enableEchoCancellation: boolean;
  enableAutoGainControl: boolean;
  enableHighPassFilter: boolean;
  enableLowPassFilter: boolean;
  noiseReductionLevel: number; // 0-100
  echoCancellationLevel: number; // 0-100
  gainControlLevel: number; // 0-100
  sampleRate: number;
  bufferSize: number;
  channels: number;
}

// 音频质量等级
export enum AudioQuality {
  HIGH = 'high',
  MEDIUM = 'medium',
  LOW = 'low',
  VERY_LOW = 'very_low'
}

// 音频统计信息
export interface AudioStats {
  volume: number;
  noiseLevel: number;
  echoLevel: number;
  quality: number;
  sampleRate: number;
  bitrate: number;
  latency: number;
}

// 音频处理器类
export class AudioProcessor {
  private static instance: AudioProcessor;
  private config: AudioProcessorConfig;
  private audioContext: AudioContext | null = null;
  private sourceNode: MediaStreamAudioSourceNode | null = null;
  private gainNode: GainNode | null = null;
  private highPassFilter: BiquadFilterNode | null = null;
  private lowPassFilter: BiquadFilterNode | null = null;
  private noiseReductionNode: ScriptProcessorNode | null = null;
  private echoCancellationNode: ScriptProcessorNode | null = null;
  private autoGainNode: DynamicsCompressorNode | null = null;
  private isProcessing = false;
  private audioStats: AudioStats;
  private volumeHistory: number[] = [];
  private noiseHistory: number[] = [];

  private constructor(config: Partial<AudioProcessorConfig> = {}) {
    this.config = {
      enableNoiseReduction: true,
      enableEchoCancellation: true,
      enableAutoGainControl: true,
      enableHighPassFilter: true,
      enableLowPassFilter: true,
      noiseReductionLevel: 70,
      echoCancellationLevel: 80,
      gainControlLevel: 60,
      sampleRate: 48000,
      bufferSize: 4096,
      channels: 1,
      ...config
    };

    this.audioStats = this.getDefaultAudioStats();
  }

  // 获取单例实例
  public static getInstance(config?: Partial<AudioProcessorConfig>): AudioProcessor {
    if (!AudioProcessor.instance) {
      AudioProcessor.instance = new AudioProcessor(config);
    }
    return AudioProcessor.instance;
  }

  // 获取默认音频统计
  private getDefaultAudioStats(): AudioStats {
    return {
      volume: 0,
      noiseLevel: 0,
      echoLevel: 0,
      quality: 100,
      sampleRate: this.config.sampleRate,
      bitrate: 128,
      latency: 0
    };
  }

  // 初始化音频处理器
  public async initialize(): Promise<void> {
    try {
      // 创建音频上下文
      this.audioContext = new (window.AudioContext || (window as any).webkitAudioContext)({
        sampleRate: this.config.sampleRate
      });

      // 创建音频节点
      this.createAudioNodes();

      console.log('音频处理器初始化成功');
    } catch (error) {
      console.error('音频处理器初始化失败:', error);
      throw error;
    }
  }

  // 创建音频节点
  private createAudioNodes(): void {
    if (!this.audioContext) {
      throw new Error('音频上下文未初始化');
    }

    // 创建增益节点
    this.gainNode = this.audioContext.createGain();
    this.gainNode.gain.value = 1.0;

    // 创建高通滤波器
    if (this.config.enableHighPassFilter) {
      this.highPassFilter = this.audioContext.createBiquadFilter();
      this.highPassFilter.type = 'highpass';
      this.highPassFilter.frequency.value = 80; // 80Hz 高通滤波
    }

    // 创建低通滤波器
    if (this.config.enableLowPassFilter) {
      this.lowPassFilter = this.audioContext.createBiquadFilter();
      this.lowPassFilter.type = 'lowpass';
      this.lowPassFilter.frequency.value = 8000; // 8kHz 低通滤波
    }

    // 创建自动增益控制节点
    if (this.config.enableAutoGainControl) {
      this.autoGainNode = this.audioContext.createDynamicsCompressor();
      this.autoGainNode.threshold.value = -24;
      this.autoGainNode.knee.value = 30;
      this.autoGainNode.ratio.value = 12;
      this.autoGainNode.attack.value = 0.003;
      this.autoGainNode.release.value = 0.25;
    }

    // 创建降噪节点
    if (this.config.enableNoiseReduction) {
      this.noiseReductionNode = this.audioContext.createScriptProcessor(
        this.config.bufferSize,
        this.config.channels,
        this.config.channels
      );
      this.noiseReductionNode.onaudioprocess = this.processNoiseReduction.bind(this);
    }

    // 创建回声消除节点
    if (this.config.enableEchoCancellation) {
      this.echoCancellationNode = this.audioContext.createScriptProcessor(
        this.config.bufferSize,
        this.config.channels,
        this.config.channels
      );
      this.echoCancellationNode.onaudioprocess = this.processEchoCancellation.bind(this);
    }
  }

  // 开始处理音频流
  public async startProcessing(stream: MediaStream): Promise<MediaStream> {
    if (!this.audioContext) {
      await this.initialize();
    }

    if (!this.audioContext) {
      throw new Error('音频上下文初始化失败');
    }

    try {
      // 创建音频源节点
      this.sourceNode = this.audioContext.createMediaStreamSource(stream);

      // 连接音频节点链
      this.connectAudioNodes();

      // 创建输出流
      const destination = this.audioContext.createMediaStreamDestination();
      this.connectToDestination(destination);

      this.isProcessing = true;
      console.log('音频处理开始');

      return destination.stream;
    } catch (error) {
      console.error('开始音频处理失败:', error);
      throw error;
    }
  }

  // 连接音频节点
  private connectAudioNodes(): void {
    if (!this.sourceNode) {
      throw new Error('音频源节点未创建');
    }

    let currentNode: AudioNode = this.sourceNode;

    // 连接增益节点
    if (this.gainNode) {
      currentNode.connect(this.gainNode);
      currentNode = this.gainNode;
    }

    // 连接高通滤波器
    if (this.highPassFilter) {
      currentNode.connect(this.highPassFilter);
      currentNode = this.highPassFilter;
    }

    // 连接低通滤波器
    if (this.lowPassFilter) {
      currentNode.connect(this.lowPassFilter);
      currentNode = this.lowPassFilter;
    }

    // 连接降噪节点
    if (this.noiseReductionNode) {
      currentNode.connect(this.noiseReductionNode);
      currentNode = this.noiseReductionNode;
    }

    // 连接回声消除节点
    if (this.echoCancellationNode) {
      currentNode.connect(this.echoCancellationNode);
      currentNode = this.echoCancellationNode;
    }

    // 连接自动增益控制节点
    if (this.autoGainNode) {
      currentNode.connect(this.autoGainNode);
      currentNode = this.autoGainNode;
    }
  }

  // 连接到输出目标
  private connectToDestination(destination: MediaStreamAudioDestinationNode): void {
    let currentNode: AudioNode | null = null;

    // 找到最后一个节点
    if (this.autoGainNode) {
      currentNode = this.autoGainNode;
    } else if (this.echoCancellationNode) {
      currentNode = this.echoCancellationNode;
    } else if (this.noiseReductionNode) {
      currentNode = this.noiseReductionNode;
    } else if (this.lowPassFilter) {
      currentNode = this.lowPassFilter;
    } else if (this.highPassFilter) {
      currentNode = this.highPassFilter;
    } else if (this.gainNode) {
      currentNode = this.gainNode;
    } else if (this.sourceNode) {
      currentNode = this.sourceNode;
    }

    if (currentNode) {
      currentNode.connect(destination);
    }
  }

  // 停止处理
  public stopProcessing(): void {
    this.isProcessing = false;

    // 断开所有连接
    if (this.sourceNode) {
      this.sourceNode.disconnect();
    }

    if (this.gainNode) {
      this.gainNode.disconnect();
    }

    if (this.highPassFilter) {
      this.highPassFilter.disconnect();
    }

    if (this.lowPassFilter) {
      this.lowPassFilter.disconnect();
    }

    if (this.noiseReductionNode) {
      this.noiseReductionNode.disconnect();
    }

    if (this.echoCancellationNode) {
      this.echoCancellationNode.disconnect();
    }

    if (this.autoGainNode) {
      this.autoGainNode.disconnect();
    }

    console.log('音频处理停止');
  }

  // 降噪处理
  private processNoiseReduction(event: AudioProcessingEvent): void {
    const inputBuffer = event.inputBuffer;
    const outputBuffer = event.outputBuffer;

    for (let channel = 0; channel < inputBuffer.numberOfChannels; channel++) {
      const inputData = inputBuffer.getChannelData(channel);
      const outputData = outputBuffer.getChannelData(channel);

      // 简单的降噪算法 - 频谱减法
      this.spectralSubtraction(inputData, outputData);
    }

    // 更新音频统计
    this.updateAudioStats(inputBuffer);
  }

  // 频谱减法降噪
  private spectralSubtraction(input: Float32Array, output: Float32Array): void {
    const reductionLevel = this.config.noiseReductionLevel / 100;
    
    for (let i = 0; i < input.length; i++) {
      const sample = input[i];
      const magnitude = Math.abs(sample);
      
      // 计算噪声阈值
      const noiseThreshold = this.calculateNoiseThreshold(magnitude);
      
      // 应用降噪
      if (magnitude < noiseThreshold) {
        output[i] = sample * (1 - reductionLevel);
      } else {
        output[i] = sample;
      }
    }
  }

  // 计算噪声阈值
  private calculateNoiseThreshold(magnitude: number): number {
    // 更新噪声历史
    this.noiseHistory.push(magnitude);
    if (this.noiseHistory.length > 100) {
      this.noiseHistory.shift();
    }

    // 计算噪声水平
    const noiseLevel = this.noiseHistory.reduce((sum, n) => sum + n, 0) / this.noiseHistory.length;
    
    return noiseLevel * 1.5; // 噪声阈值为平均噪声水平的1.5倍
  }

  // 回声消除处理
  private processEchoCancellation(event: AudioProcessingEvent): void {
    const inputBuffer = event.inputBuffer;
    const outputBuffer = event.outputBuffer;

    for (let channel = 0; channel < inputBuffer.numberOfChannels; channel++) {
      const inputData = inputBuffer.getChannelData(channel);
      const outputData = outputBuffer.getChannelData(channel);

      // 简单的回声消除算法
      this.adaptiveEchoCancellation(inputData, outputData);
    }
  }

  // 自适应回声消除
  private adaptiveEchoCancellation(input: Float32Array, output: Float32Array): void {
    const cancellationLevel = this.config.echoCancellationLevel / 100;
    
    for (let i = 0; i < input.length; i++) {
      const sample = input[i];
      
      // 检测回声模式
      const echoPattern = this.detectEchoPattern(sample, i);
      
      if (echoPattern > 0.5) {
        // 应用回声消除
        output[i] = sample * (1 - cancellationLevel);
      } else {
        output[i] = sample;
      }
    }
  }

  // 检测回声模式
  private detectEchoPattern(sample: number, index: number): number {
    // 简单的回声检测算法
    // 实际应用中应该使用更复杂的算法
    
    const delay = 100; // 假设回声延迟100个采样点
    if (index >= delay) {
      const delayedSample = this.getDelayedSample(index - delay);
      const correlation = Math.abs(sample - delayedSample);
      return correlation > 0.1 ? 1 : 0;
    }
    
    return 0;
  }

  // 获取延迟采样
  private getDelayedSample(index: number): number {
    // 这里应该维护一个延迟缓冲区
    // 为了简化，返回0
    return 0;
  }

  // 更新音频统计
  private updateAudioStats(inputBuffer: AudioBuffer): void {
    // 计算音量
    let sum = 0;
    for (let channel = 0; channel < inputBuffer.numberOfChannels; channel++) {
      const data = inputBuffer.getChannelData(channel);
      for (let i = 0; i < data.length; i++) {
        sum += data[i] * data[i];
      }
    }
    
    const rms = Math.sqrt(sum / (inputBuffer.length * inputBuffer.numberOfChannels));
    const volume = Math.min(100, rms * 1000); // 转换为0-100的范围

    // 更新音量历史
    this.volumeHistory.push(volume);
    if (this.volumeHistory.length > 100) {
      this.volumeHistory.shift();
    }

    // 更新统计信息
    this.audioStats.volume = volume;
    this.audioStats.noiseLevel = this.calculateNoiseLevel();
    this.audioStats.quality = this.calculateAudioQuality();
    this.audioStats.sampleRate = this.audioContext?.sampleRate || 48000;
  }

  // 计算噪声水平
  private calculateNoiseLevel(): number {
    if (this.noiseHistory.length === 0) {
      return 0;
    }
    
    const avgNoise = this.noiseHistory.reduce((sum, n) => sum + n, 0) / this.noiseHistory.length;
    return Math.min(100, avgNoise * 1000);
  }

  // 计算音频质量
  private calculateAudioQuality(): number {
    let quality = 100;

    // 基于音量调整质量
    const avgVolume = this.volumeHistory.reduce((sum, v) => sum + v, 0) / this.volumeHistory.length;
    if (avgVolume < 10) {
      quality -= 30; // 音量太低
    } else if (avgVolume > 90) {
      quality -= 20; // 音量太高
    }

    // 基于噪声水平调整质量
    const noiseLevel = this.audioStats.noiseLevel;
    quality -= noiseLevel * 0.5;

    return Math.max(0, quality);
  }

  // 设置音频质量
  public setAudioQuality(quality: AudioQuality): void {
    switch (quality) {
      case AudioQuality.HIGH:
        this.config.noiseReductionLevel = 80;
        this.config.echoCancellationLevel = 90;
        this.config.gainControlLevel = 70;
        this.config.sampleRate = 48000;
        this.config.bitrate = 192;
        break;

      case AudioQuality.MEDIUM:
        this.config.noiseReductionLevel = 60;
        this.config.echoCancellationLevel = 70;
        this.config.gainControlLevel = 50;
        this.config.sampleRate = 44100;
        this.config.bitrate = 128;
        break;

      case AudioQuality.LOW:
        this.config.noiseReductionLevel = 40;
        this.config.echoCancellationLevel = 50;
        this.config.gainControlLevel = 30;
        this.config.sampleRate = 22050;
        this.config.bitrate = 96;
        break;

      case AudioQuality.VERY_LOW:
        this.config.noiseReductionLevel = 20;
        this.config.echoCancellationLevel = 30;
        this.config.gainControlLevel = 20;
        this.config.sampleRate = 16000;
        this.config.bitrate = 64;
        break;
    }

    // 重新创建音频节点
    if (this.isProcessing) {
      this.stopProcessing();
      this.createAudioNodes();
    }
  }

  // 调整增益
  public setGain(gain: number): void {
    if (this.gainNode) {
      this.gainNode.gain.value = Math.max(0, Math.min(2, gain));
    }
  }

  // 调整降噪级别
  public setNoiseReductionLevel(level: number): void {
    this.config.noiseReductionLevel = Math.max(0, Math.min(100, level));
  }

  // 调整回声消除级别
  public setEchoCancellationLevel(level: number): void {
    this.config.echoCancellationLevel = Math.max(0, Math.min(100, level));
  }

  // 调整自动增益控制级别
  public setAutoGainLevel(level: number): void {
    this.config.gainControlLevel = Math.max(0, Math.min(100, level));
    
    if (this.autoGainNode) {
      const threshold = -24 + (level / 100) * 24; // -24 到 0
      this.autoGainNode.threshold.value = threshold;
    }
  }

  // 获取音频统计
  public getAudioStats(): AudioStats {
    return { ...this.audioStats };
  }

  // 获取配置
  public getConfig(): AudioProcessorConfig {
    return { ...this.config };
  }

  // 更新配置
  public updateConfig(newConfig: Partial<AudioProcessorConfig>): void {
    this.config = { ...this.config, ...newConfig };
  }

  // 检查浏览器支持
  public static isSupported(): boolean {
    return !!(window.AudioContext || (window as any).webkitAudioContext);
  }

  // 获取支持的音频格式
  public static getSupportedFormats(): string[] {
    const formats = ['audio/webm', 'audio/ogg', 'audio/mp4', 'audio/wav'];
    const supported: string[] = [];

    formats.forEach(format => {
      if (MediaRecorder.isTypeSupported(format)) {
        supported.push(format);
      }
    });

    return supported;
  }
}

// 创建全局实例
export const audioProcessor = AudioProcessor.getInstance();

// 默认导出
export default audioProcessor;


