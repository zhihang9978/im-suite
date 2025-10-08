/**
 * 志航密信 Web 端错误处理工具
 * 提供统一的错误处理、错误码映射和用户友好的错误提示
 */

// 错误码类型定义
export type ErrorCode = 
  // 用户相关错误 (10000-19999)
  | 10001 // 用户不存在
  | 10002 // 用户已存在
  | 10003 // 密码错误
  | 10004 // 验证码错误
  | 10005 // 用户被禁用
  | 10006 // 手机号已存在
  | 10007 // 用户名已存在
  | 10008 // 令牌已过期
  | 10009 // 令牌无效
  | 10010 // 权限不足
  
  // 聊天相关错误 (20000-29999)
  | 20001 // 聊天不存在
  | 20002 // 用户不是聊天成员
  | 20003 // 消息不存在
  | 20004 // 无权限操作此消息
  | 20005 // 聊天已满
  | 20006 // 无效的聊天类型
  
  // 文件相关错误 (30000-39999)
  | 30001 // 文件不存在
  | 30002 // 文件类型不支持
  | 30003 // 文件大小超限
  | 30004 // 文件上传失败
  | 30005 // 文件下载失败
  
  // 系统相关错误 (40000-49999)
  | 40001 // 请求过于频繁
  | 40002 // 请求参数错误
  | 40003 // 数据库错误
  | 40004 // 缓存错误
  | 40005 // 网络错误
  
  // 服务器错误 (50000-59999)
  | 50001 // 服务器内部错误
  | 50002 // 服务不可用
  | 50003; // 请求超时

// 应用错误接口
export interface AppError {
  code: ErrorCode;
  message: string;
  details?: string;
  request_id?: string;
  timestamp: string;
}

// 错误消息映射
const ERROR_MESSAGES: Record<ErrorCode, string> = {
  // 用户相关错误
  10001: '用户不存在',
  10002: '用户已存在',
  10003: '密码错误',
  10004: '验证码错误',
  10005: '用户已被禁用',
  10006: '手机号已存在',
  10007: '用户名已存在',
  10008: '令牌已过期',
  10009: '令牌无效',
  10010: '权限不足',

  // 聊天相关错误
  20001: '聊天不存在',
  20002: '用户不是聊天成员',
  20003: '消息不存在',
  20004: '无权限操作此消息',
  20005: '聊天已满',
  20006: '无效的聊天类型',

  // 文件相关错误
  30001: '文件不存在',
  30002: '文件类型不支持',
  30003: '文件大小超限',
  30004: '文件上传失败',
  30005: '文件下载失败',

  // 系统相关错误
  40001: '请求过于频繁',
  40002: '请求参数错误',
  40003: '数据库错误',
  40004: '缓存错误',
  40005: '网络错误',

  // 服务器错误
  50001: '服务器内部错误',
  50002: '服务不可用',
  50003: '请求超时',
};

// 错误处理类
export class ErrorHandler {
  private static instance: ErrorHandler;
  private errorListeners: Array<(error: AppError) => void> = [];
  private errorHistory: AppError[] = [];
  private maxHistorySize = 100;

  private constructor() {}

  // 获取单例实例
  public static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler();
    }
    return ErrorHandler.instance;
  }

  // 处理错误
  public handleError(error: any, context?: string): AppError {
    let appError: AppError;

    // 检查是否为应用错误
    if (this.isAppError(error)) {
      appError = error;
    } else if (error.response) {
      // HTTP 响应错误
      appError = this.handleHttpError(error);
    } else if (error instanceof Error) {
      // 普通错误
      appError = this.handleGenericError(error, context);
    } else {
      // 其他类型错误
      appError = this.handleUnknownError(error, context);
    }

    // 记录错误
    this.recordError(appError);

    // 通知监听器
    this.notifyListeners(appError);

    // 显示用户友好的错误提示
    this.showUserFriendlyMessage(appError);

    return appError;
  }

  // 检查是否为应用错误
  private isAppError(error: any): error is AppError {
    return (
      error &&
      typeof error.code === 'number' &&
      typeof error.message === 'string' &&
      typeof error.timestamp === 'string'
    );
  }

  // 处理 HTTP 错误
  private handleHttpError(error: any): AppError {
    const response = error.response;
    let code: ErrorCode = 50001;
    let message = '服务器内部错误';
    let details = '';

    if (response.data && this.isAppError(response.data)) {
      // 服务器返回的应用错误
      return response.data;
    }

    // 根据 HTTP 状态码确定错误类型
    switch (response.status) {
      case 400:
        code = 40002;
        message = '请求参数错误';
        details = response.data?.message || '请求参数格式不正确';
        break;
      case 401:
        code = 10009;
        message = '令牌无效';
        details = '请重新登录';
        break;
      case 403:
        code = 10010;
        message = '权限不足';
        details = '您没有权限执行此操作';
        break;
      case 404:
        code = 40002;
        message = '资源不存在';
        details = '请求的资源不存在';
        break;
      case 429:
        code = 40001;
        message = '请求过于频繁';
        details = '请稍后再试';
        break;
      case 500:
        code = 50001;
        message = '服务器内部错误';
        details = '服务器暂时无法处理请求';
        break;
      case 503:
        code = 50002;
        message = '服务不可用';
        details = '服务暂时不可用，请稍后再试';
        break;
      default:
        details = response.data?.message || `HTTP ${response.status}`;
    }

    return {
      code,
      message,
      details,
      timestamp: new Date().toISOString(),
    };
  }

  // 处理普通错误
  private handleGenericError(error: Error, context?: string): AppError {
    let code: ErrorCode = 50001;
    let message = '服务器内部错误';
    let details = error.message;

    // 根据错误类型确定错误码
    if (error.name === 'NetworkError' || error.message.includes('网络')) {
      code = 40005;
      message = '网络错误';
    } else if (error.name === 'TimeoutError' || error.message.includes('超时')) {
      code = 50003;
      message = '请求超时';
    } else if (error.name === 'ValidationError' || error.message.includes('验证')) {
      code = 40002;
      message = '请求参数错误';
    }

    if (context) {
      details = `[${context}] ${details}`;
    }

    return {
      code,
      message,
      details,
      timestamp: new Date().toISOString(),
    };
  }

  // 处理未知错误
  private handleUnknownError(error: any, context?: string): AppError {
    const details = context 
      ? `[${context}] ${String(error)}` 
      : String(error);

    return {
      code: 50001,
      message: '服务器内部错误',
      details,
      timestamp: new Date().toISOString(),
    };
  }

  // 记录错误
  private recordError(error: AppError): void {
    this.errorHistory.push(error);
    
    // 限制历史记录大小
    if (this.errorHistory.length > this.maxHistorySize) {
      this.errorHistory.shift();
    }

    // 记录到控制台（开发环境）
    if (process.env.NODE_ENV === 'development') {
      console.error('应用错误:', error);
    }
  }

  // 通知错误监听器
  private notifyListeners(error: AppError): void {
    this.errorListeners.forEach(listener => {
      try {
        listener(error);
      } catch (e) {
        console.error('错误监听器执行失败:', e);
      }
    });
  }

  // 显示用户友好的错误提示
  private showUserFriendlyMessage(error: AppError): void {
    const message = this.getUserFriendlyMessage(error);
    
    // 这里可以集成 Toast 组件或其他 UI 提示组件
    // 例如：toast.error(message);
    console.warn('用户提示:', message);
    
    // 如果是严重错误，可以显示更详细的提示
    if (this.isCriticalError(error.code)) {
      console.error('严重错误:', error);
    }
  }

  // 获取用户友好的错误消息
  public getUserFriendlyMessage(error: AppError): string {
    const baseMessage = ERROR_MESSAGES[error.code] || '未知错误';
    
    // 根据错误类型返回不同的用户提示
    switch (error.code) {
      case 10008:
        return '登录已过期，请重新登录';
      case 10009:
        return '登录状态无效，请重新登录';
      case 10010:
        return '您没有权限执行此操作';
      case 20002:
        return '您不是该聊天的成员';
      case 20004:
        return '您没有权限操作此消息';
      case 40001:
        return '操作过于频繁，请稍后再试';
      case 40005:
        return '网络连接异常，请检查网络设置';
      case 50001:
        return '服务器暂时无法处理请求，请稍后再试';
      case 50002:
        return '服务暂时不可用，请稍后再试';
      default:
        return baseMessage;
    }
  }

  // 检查是否为严重错误
  private isCriticalError(code: ErrorCode): boolean {
    return code >= 50000 || code === 10009 || code === 10008;
  }

  // 添加错误监听器
  public addErrorListener(listener: (error: AppError) => void): void {
    this.errorListeners.push(listener);
  }

  // 移除错误监听器
  public removeErrorListener(listener: (error: AppError) => void): void {
    const index = this.errorListeners.indexOf(listener);
    if (index > -1) {
      this.errorListeners.splice(index, 1);
    }
  }

  // 获取错误历史
  public getErrorHistory(): AppError[] {
    return [...this.errorHistory];
  }

  // 清空错误历史
  public clearErrorHistory(): void {
    this.errorHistory = [];
  }

  // 获取错误统计
  public getErrorStats(): Record<ErrorCode, number> {
    const stats: Record<ErrorCode, number> = {} as Record<ErrorCode, number>;
    
    this.errorHistory.forEach(error => {
      stats[error.code] = (stats[error.code] || 0) + 1;
    });
    
    return stats;
  }
}

// 创建全局错误处理器实例
export const errorHandler = ErrorHandler.getInstance();

// 错误处理装饰器
export function withErrorHandling<T extends any[], R>(
  fn: (...args: T) => Promise<R>,
  context?: string
) {
  return async (...args: T): Promise<R> => {
    try {
      return await fn(...args);
    } catch (error) {
      throw errorHandler.handleError(error, context);
    }
  };
}

// 错误处理高阶函数
export function createErrorBoundary<T extends any[], R>(
  fn: (...args: T) => R,
  fallback?: (error: AppError) => R,
  context?: string
) {
  return (...args: T): R => {
    try {
      return fn(...args);
    } catch (error) {
      const appError = errorHandler.handleError(error, context);
      
      if (fallback) {
        return fallback(appError);
      }
      
      throw appError;
    }
  };
}

// 重试机制
export async function retry<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<T> {
  let lastError: any;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;
      
      // 检查是否为可重试的错误
      if (!isRetryableError(error)) {
        throw error;
      }
      
      // 如果不是最后一次重试，等待后重试
      if (i < maxRetries - 1) {
        await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)));
      }
    }
  }
  
  throw lastError;
}

// 检查是否为可重试的错误
function isRetryableError(error: any): boolean {
  if (error?.code) {
    // 网络错误、超时错误、服务器错误通常可重试
    return [40005, 50001, 50002, 50003].includes(error.code);
  }
  
  // 网络相关错误可重试
  return error?.message?.includes('网络') || 
         error?.message?.includes('超时') ||
         error?.message?.includes('timeout');
}

// 导出工具函数
export const errorUtils = {
  // 获取错误消息
  getErrorMessage: (code: ErrorCode): string => ERROR_MESSAGES[code] || '未知错误',
  
  // 检查是否为应用错误
  isAppError: (error: any): error is AppError => {
    return (
      error &&
      typeof error.code === 'number' &&
      typeof error.message === 'string' &&
      typeof error.timestamp === 'string'
    );
  },
  
  // 创建应用错误
  createAppError: (code: ErrorCode, message?: string, details?: string): AppError => ({
    code,
    message: message || ERROR_MESSAGES[code] || '未知错误',
    details,
    timestamp: new Date().toISOString(),
  }),
  
  // 格式化错误信息
  formatError: (error: AppError): string => {
    if (error.details) {
      return `${error.message}: ${error.details}`;
    }
    return error.message;
  },
};

// 默认导出
export default errorHandler;


