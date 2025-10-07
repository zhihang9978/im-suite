/**
 * 志航密信 Web 端性能优化工具
 * 提供性能监控、资源优化、缓存管理等功能
 */

// 性能监控接口
export interface PerformanceMetrics {
  loadTime: number;
  renderTime: number;
  memoryUsage: number;
  networkRequests: number;
  errorCount: number;
  slowOperations: number;
}

// 性能监控器类
export class PerformanceMonitor {
  private static instance: PerformanceMonitor;
  private metrics: PerformanceMetrics;
  private observers: Array<(metrics: PerformanceMetrics) => void> = [];
  private isMonitoring = false;

  private constructor() {
    this.metrics = {
      loadTime: 0,
      renderTime: 0,
      memoryUsage: 0,
      networkRequests: 0,
      errorCount: 0,
      slowOperations: 0,
    };
  }

  // 获取单例实例
  public static getInstance(): PerformanceMonitor {
    if (!PerformanceMonitor.instance) {
      PerformanceMonitor.instance = new PerformanceMonitor();
    }
    return PerformanceMonitor.instance;
  }

  // 开始监控
  public startMonitoring(): void {
    if (this.isMonitoring) {
      return;
    }

    this.isMonitoring = true;
    this.initializeMetrics();
    this.setupPerformanceObserver();
    this.setupResourceObserver();
    this.setupErrorObserver();
  }

  // 停止监控
  public stopMonitoring(): void {
    this.isMonitoring = false;
  }

  // 初始化性能指标
  private initializeMetrics(): void {
    // 测量页面加载时间
    if (performance.timing) {
      const loadTime = performance.timing.loadEventEnd - performance.timing.navigationStart;
      this.metrics.loadTime = loadTime;
    }

    // 测量内存使用情况
    if ('memory' in performance) {
      const memory = (performance as any).memory;
      this.metrics.memoryUsage = memory.usedJSHeapSize;
    }
  }

  // 设置性能观察器
  private setupPerformanceObserver(): void {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        
        entries.forEach((entry) => {
          if (entry.entryType === 'measure') {
            // 测量自定义性能指标
            this.handleCustomMeasure(entry);
          } else if (entry.entryType === 'navigation') {
            // 测量页面导航性能
            this.handleNavigationTiming(entry);
          }
        });
      });

      observer.observe({ entryTypes: ['measure', 'navigation'] });
    }
  }

  // 设置资源观察器
  private setupResourceObserver(): void {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        
        entries.forEach((entry) => {
          if (entry.entryType === 'resource') {
            this.handleResourceTiming(entry);
          }
        });
      });

      observer.observe({ entryTypes: ['resource'] });
    }
  }

  // 设置错误观察器
  private setupErrorObserver(): void {
    window.addEventListener('error', (event) => {
      this.metrics.errorCount++;
      this.notifyObservers();
    });

    window.addEventListener('unhandledrejection', (event) => {
      this.metrics.errorCount++;
      this.notifyObservers();
    });
  }

  // 处理自定义测量
  private handleCustomMeasure(entry: PerformanceEntry): void {
    if (entry.name.includes('render')) {
      this.metrics.renderTime = entry.duration;
    }
    
    if (entry.duration > 100) {
      this.metrics.slowOperations++;
    }
    
    this.notifyObservers();
  }

  // 处理导航时序
  private handleNavigationTiming(entry: PerformanceEntry): void {
    const navEntry = entry as PerformanceNavigationTiming;
    this.metrics.loadTime = navEntry.loadEventEnd - navEntry.navigationStart;
    this.notifyObservers();
  }

  // 处理资源时序
  private handleResourceTiming(entry: PerformanceEntry): void {
    this.metrics.networkRequests++;
    
    if (entry.duration > 1000) {
      this.metrics.slowOperations++;
    }
    
    this.notifyObservers();
  }

  // 开始性能测量
  public startMeasure(name: string): void {
    performance.mark(`${name}-start`);
  }

  // 结束性能测量
  public endMeasure(name: string): number {
    performance.mark(`${name}-end`);
    performance.measure(name, `${name}-start`, `${name}-end`);
    
    const entries = performance.getEntriesByName(name, 'measure');
    if (entries.length > 0) {
      const duration = entries[entries.length - 1].duration;
      
      if (duration > 100) {
        this.metrics.slowOperations++;
      }
      
      this.notifyObservers();
      return duration;
    }
    
    return 0;
  }

  // 记录慢操作
  public recordSlowOperation(operation: string, duration: number): void {
    this.metrics.slowOperations++;
    
    console.warn(`Slow operation detected: ${operation} took ${duration}ms`);
    
    // 发送慢操作报告到服务器
    this.reportSlowOperation(operation, duration);
  }

  // 报告慢操作
  private async reportSlowOperation(operation: string, duration: number): Promise<void> {
    try {
      await fetch('/api/performance/slow-operation', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          operation,
          duration,
          timestamp: Date.now(),
          userAgent: navigator.userAgent,
          url: window.location.href,
        }),
      });
    } catch (error) {
      console.error('Failed to report slow operation:', error);
    }
  }

  // 获取性能指标
  public getMetrics(): PerformanceMetrics {
    return { ...this.metrics };
  }

  // 添加观察器
  public addObserver(observer: (metrics: PerformanceMetrics) => void): void {
    this.observers.push(observer);
  }

  // 移除观察器
  public removeObserver(observer: (metrics: PerformanceMetrics) => void): void {
    const index = this.observers.indexOf(observer);
    if (index > -1) {
      this.observers.splice(index, 1);
    }
  }

  // 通知观察器
  private notifyObservers(): void {
    this.observers.forEach(observer => {
      try {
        observer(this.getMetrics());
      } catch (error) {
        console.error('Performance observer error:', error);
      }
    });
  }

  // 重置指标
  public resetMetrics(): void {
    this.metrics = {
      loadTime: 0,
      renderTime: 0,
      memoryUsage: 0,
      networkRequests: 0,
      errorCount: 0,
      slowOperations: 0,
    };
  }
}

// 资源优化器
export class ResourceOptimizer {
  private static instance: ResourceOptimizer;
  private imageCache = new Map<string, HTMLImageElement>();
  private scriptCache = new Set<string>();
  private styleCache = new Set<string>();

  private constructor() {}

  // 获取单例实例
  public static getInstance(): ResourceOptimizer {
    if (!ResourceOptimizer.instance) {
      ResourceOptimizer.instance = new ResourceOptimizer();
    }
    return ResourceOptimizer.instance;
  }

  // 预加载图片
  public preloadImage(src: string): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
      if (this.imageCache.has(src)) {
        resolve(this.imageCache.get(src)!);
        return;
      }

      const img = new Image();
      img.onload = () => {
        this.imageCache.set(src, img);
        resolve(img);
      };
      img.onerror = reject;
      img.src = src;
    });
  }

  // 预加载脚本
  public preloadScript(src: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.scriptCache.has(src)) {
        resolve();
        return;
      }

      const script = document.createElement('script');
      script.src = src;
      script.onload = () => {
        this.scriptCache.add(src);
        resolve();
      };
      script.onerror = reject;
      document.head.appendChild(script);
    });
  }

  // 预加载样式
  public preloadStyle(href: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.styleCache.has(href)) {
        resolve();
        return;
      }

      const link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = href;
      link.onload = () => {
        this.styleCache.add(href);
        resolve();
      };
      link.onerror = reject;
      document.head.appendChild(link);
    });
  }

  // 清理缓存
  public clearCache(): void {
    this.imageCache.clear();
    this.scriptCache.clear();
    this.styleCache.clear();
  }
}

// 缓存管理器
export class CacheManager {
  private static instance: CacheManager;
  private memoryCache = new Map<string, { data: any; timestamp: number; ttl: number }>();
  private maxCacheSize = 100;
  private defaultTTL = 5 * 60 * 1000; // 5分钟

  private constructor() {}

  // 获取单例实例
  public static getInstance(): CacheManager {
    if (!CacheManager.instance) {
      CacheManager.instance = new CacheManager();
    }
    return CacheManager.instance;
  }

  // 设置缓存
  public set(key: string, data: any, ttl?: number): void {
    const timestamp = Date.now();
    const cacheTTL = ttl || this.defaultTTL;

    // 如果缓存已满，删除最旧的条目
    if (this.memoryCache.size >= this.maxCacheSize) {
      this.evictOldest();
    }

    this.memoryCache.set(key, {
      data,
      timestamp,
      ttl: cacheTTL,
    });
  }

  // 获取缓存
  public get(key: string): any {
    const item = this.memoryCache.get(key);
    
    if (!item) {
      return null;
    }

    // 检查是否过期
    if (Date.now() - item.timestamp > item.ttl) {
      this.memoryCache.delete(key);
      return null;
    }

    return item.data;
  }

  // 删除缓存
  public delete(key: string): void {
    this.memoryCache.delete(key);
  }

  // 清空缓存
  public clear(): void {
    this.memoryCache.clear();
  }

  // 删除最旧的条目
  private evictOldest(): void {
    let oldestKey = '';
    let oldestTimestamp = Date.now();

    for (const [key, item] of this.memoryCache) {
      if (item.timestamp < oldestTimestamp) {
        oldestTimestamp = item.timestamp;
        oldestKey = key;
      }
    }

    if (oldestKey) {
      this.memoryCache.delete(oldestKey);
    }
  }

  // 获取缓存统计
  public getStats(): { size: number; maxSize: number; hitRate: number } {
    return {
      size: this.memoryCache.size,
      maxSize: this.maxCacheSize,
      hitRate: 0, // 这里需要实现命中率统计
    };
  }
}

// 网络优化器
export class NetworkOptimizer {
  private static instance: NetworkOptimizer;
  private requestQueue: Array<() => Promise<any>> = [];
  private maxConcurrentRequests = 6;
  private activeRequests = 0;

  private constructor() {}

  // 获取单例实例
  public static getInstance(): NetworkOptimizer {
    if (!NetworkOptimizer.instance) {
      NetworkOptimizer.instance = new NetworkOptimizer();
    }
    return NetworkOptimizer.instance;
  }

  // 发送请求
  public async sendRequest<T>(request: () => Promise<T>): Promise<T> {
    return new Promise((resolve, reject) => {
      const executeRequest = async () => {
        try {
          const result = await request();
          resolve(result);
        } catch (error) {
          reject(error);
        } finally {
          this.activeRequests--;
          this.processQueue();
        }
      };

      if (this.activeRequests < this.maxConcurrentRequests) {
        this.activeRequests++;
        executeRequest();
      } else {
        this.requestQueue.push(executeRequest);
      }
    });
  }

  // 处理队列
  private processQueue(): void {
    if (this.requestQueue.length > 0 && this.activeRequests < this.maxConcurrentRequests) {
      const request = this.requestQueue.shift();
      if (request) {
        this.activeRequests++;
        request();
      }
    }
  }

  // 批量请求
  public async batchRequest<T>(requests: Array<() => Promise<T>>): Promise<T[]> {
    const results = await Promise.allSettled(
      requests.map(request => this.sendRequest(request))
    );

    return results.map(result => {
      if (result.status === 'fulfilled') {
        return result.value;
      } else {
        throw result.reason;
      }
    });
  }
}

// 性能优化工具函数
export const performanceUtils = {
  // 防抖函数
  debounce: <T extends (...args: any[]) => any>(
    func: T,
    delay: number
  ): ((...args: Parameters<T>) => void) => {
    let timeoutId: NodeJS.Timeout;
    return (...args: Parameters<T>) => {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => func(...args), delay);
    };
  },

  // 节流函数
  throttle: <T extends (...args: any[]) => any>(
    func: T,
    delay: number
  ): ((...args: Parameters<T>) => void) => {
    let lastCall = 0;
    return (...args: Parameters<T>) => {
      const now = Date.now();
      if (now - lastCall >= delay) {
        lastCall = now;
        func(...args);
      }
    };
  },

  // 懒加载
  lazyLoad: (element: HTMLElement, callback: () => void): void => {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          callback();
          observer.unobserve(element);
        }
      });
    });
    observer.observe(element);
  },

  // 虚拟滚动
  virtualScroll: (
    container: HTMLElement,
    items: any[],
    itemHeight: number,
    renderItem: (item: any, index: number) => HTMLElement
  ): void => {
    const visibleItems = Math.ceil(container.clientHeight / itemHeight) + 2;
    let startIndex = 0;

    const updateVisibleItems = () => {
      const scrollTop = container.scrollTop;
      startIndex = Math.floor(scrollTop / itemHeight);
      
      container.innerHTML = '';
      
      for (let i = startIndex; i < Math.min(startIndex + visibleItems, items.length); i++) {
        const item = renderItem(items[i], i);
        item.style.position = 'absolute';
        item.style.top = `${i * itemHeight}px`;
        container.appendChild(item);
      }
    };

    container.addEventListener('scroll', updateVisibleItems);
    updateVisibleItems();
  },
};

// 创建全局实例
export const performanceMonitor = PerformanceMonitor.getInstance();
export const resourceOptimizer = ResourceOptimizer.getInstance();
export const cacheManager = CacheManager.getInstance();
export const networkOptimizer = NetworkOptimizer.getInstance();

// 默认导出
export default {
  performanceMonitor,
  resourceOptimizer,
  cacheManager,
  networkOptimizer,
  performanceUtils,
};
