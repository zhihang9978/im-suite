package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
	"zhihang-messenger/im-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var appErr *utils.AppError
		
		// 处理不同类型的错误
		switch err := recovered.(type) {
		case *utils.AppError:
			// 应用错误
			appErr = err
		case error:
			// 普通错误，转换为应用错误
			appErr = utils.WrapErrorWithStack(err, utils.ErrCodeInternalError, "服务器内部错误")
		case string:
			// 字符串错误
			appErr = utils.NewAppErrorWithStack(utils.ErrCodeInternalError, "服务器内部错误", err)
		default:
			// 其他类型错误
			appErr = utils.NewAppErrorWithStack(utils.ErrCodeInternalError, "服务器内部错误", fmt.Sprintf("%v", err))
		}
		
		// 设置请求ID
		if requestID, exists := c.Get("request_id"); exists {
			appErr.SetRequestID(fmt.Sprintf("%v", requestID))
		}
		
		// 记录错误日志
		logError(c, appErr)
		
		// 返回错误响应
		c.JSON(appErr.GetHTTPStatus(), appErr.ToErrorResponse())
		c.Abort()
	})
}

// ErrorHandlerWithLogger 带日志记录的错误处理中间件
func ErrorHandlerWithLogger(logger *logrus.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var appErr *utils.AppError
		
		// 处理不同类型的错误
		switch err := recovered.(type) {
		case *utils.AppError:
			appErr = err
		case error:
			appErr = utils.WrapErrorWithStack(err, utils.ErrCodeInternalError, "服务器内部错误")
		case string:
			appErr = utils.NewAppErrorWithStack(utils.ErrCodeInternalError, "服务器内部错误", err)
		default:
			appErr = utils.NewAppErrorWithStack(utils.ErrCodeInternalError, "服务器内部错误", fmt.Sprintf("%v", err))
		}
		
		// 设置请求ID
		if requestID, exists := c.Get("request_id"); exists {
			appErr.SetRequestID(fmt.Sprintf("%v", requestID))
		}
		
		// 记录错误日志
		logErrorWithLogger(c, appErr, logger)
		
		// 返回错误响应
		c.JSON(appErr.GetHTTPStatus(), appErr.ToErrorResponse())
		c.Abort()
	})
}

// ValidationErrorHandler 验证错误处理中间件
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// 检查是否有验证错误
		if len(c.Errors) > 0 {
			var validationErrors []string
			
			for _, err := range c.Errors {
				// 提取字段验证错误
				if fieldErr, ok := err.Err.(*FieldError); ok {
					validationErrors = append(validationErrors, fieldErr.Error())
				} else {
					validationErrors = append(validationErrors, err.Error())
				}
			}
			
			// 创建验证错误响应
			appErr := utils.NewAppError(
				utils.ErrCodeInvalidParams,
				"请求参数验证失败",
				strings.Join(validationErrors, "; "),
			)
			
			// 设置请求ID
			if requestID, exists := c.Get("request_id"); exists {
				appErr.SetRequestID(fmt.Sprintf("%v", requestID))
			}
			
			// 记录错误日志
			logError(c, appErr)
			
			// 返回错误响应
			c.JSON(appErr.GetHTTPStatus(), appErr.ToErrorResponse())
			c.Abort()
			return
		}
	}
}

// FieldError 字段验证错误
type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (e *FieldError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("字段 %s 验证失败: %s", e.Field, e.Tag)
}

// NewFieldError 创建字段错误
func NewFieldError(field, tag, value, message string) *FieldError {
	return &FieldError{
		Field:   field,
		Tag:     tag,
		Value:   value,
		Message: message,
	}
}

// 错误日志记录函数

// logError 记录错误日志
func logError(c *gin.Context, err *utils.AppError) {
	// 获取请求信息
	requestInfo := getRequestInfo(c)
	
	// 记录错误日志
	logrus.WithFields(logrus.Fields{
		"error_code":    err.Code,
		"error_message": err.Message,
		"error_details": err.Details,
		"request_id":    err.RequestID,
		"timestamp":     err.Timestamp,
		"request_info":  requestInfo,
	}).Error("Request error occurred")
}

// logErrorWithLogger 使用指定日志记录器记录错误
func logErrorWithLogger(c *gin.Context, err *utils.AppError, logger *logrus.Logger) {
	// 获取请求信息
	requestInfo := getRequestInfo(c)
	
	// 记录错误日志
	logger.WithFields(logrus.Fields{
		"error_code":    err.Code,
		"error_message": err.Message,
		"error_details": err.Details,
		"request_id":    err.RequestID,
		"timestamp":     err.Timestamp,
		"request_info":  requestInfo,
	}).Error("Request error occurred")
}

// getRequestInfo 获取请求信息
func getRequestInfo(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"method":     c.Request.Method,
		"url":        c.Request.URL.String(),
		"user_agent": c.Request.UserAgent(),
		"client_ip":  c.ClientIP(),
		"headers":    c.Request.Header,
	}
}

// 错误响应工具函数

// HandleError 处理错误并返回响应
func HandleError(c *gin.Context, err error) {
	var appErr *utils.AppError
	
	// 检查是否为应用错误
	if utils.IsAppError(err) {
		appErr = utils.GetAppError(err)
	} else {
		// 转换为应用错误
		appErr = utils.WrapErrorWithStack(err, utils.ErrCodeInternalError, "服务器内部错误")
	}
	
	// 设置请求ID
	if requestID, exists := c.Get("request_id"); exists {
		appErr.SetRequestID(fmt.Sprintf("%v", requestID))
	}
	
	// 记录错误日志
	logError(c, appErr)
	
	// 返回错误响应
	c.JSON(appErr.GetHTTPStatus(), appErr.ToErrorResponse())
}

// HandleValidationError 处理验证错误
func HandleValidationError(c *gin.Context, field, tag, value, message string) {
	fieldErr := NewFieldError(field, tag, value, message)
	
	appErr := utils.NewAppError(
		utils.ErrCodeInvalidParams,
		"请求参数验证失败",
		fieldErr.Error(),
	)
	
	// 设置请求ID
	if requestID, exists := c.Get("request_id"); exists {
		appErr.SetRequestID(fmt.Sprintf("%v", requestID))
	}
	
	// 记录错误日志
	logError(c, appErr)
	
	// 返回错误响应
	c.JSON(appErr.GetHTTPStatus(), appErr.ToErrorResponse())
}

// HandleBusinessError 处理业务错误
func HandleBusinessError(c *gin.Context, code utils.ErrorCode, message string, details ...string) {
	appErr := utils.NewAppError(code, message, details...)
	
	// 设置请求ID
	if requestID, exists := c.Get("request_id"); exists {
		appErr.SetRequestID(fmt.Sprintf("%v", requestID))
	}
	
	// 记录错误日志
	logError(c, appErr)
	
	// 返回错误响应
	c.JSON(appErr.GetHTTPStatus(), appErr.ToErrorResponse())
}

// 错误监控和告警

// ErrorMonitor 错误监控器
type ErrorMonitor struct {
	errorCounts map[utils.ErrorCode]int
	lastReset   time.Time
}

// NewErrorMonitor 创建错误监控器
func NewErrorMonitor() *ErrorMonitor {
	return &ErrorMonitor{
		errorCounts: make(map[utils.ErrorCode]int),
		lastReset:   time.Now(),
	}
}

// RecordError 记录错误
func (em *ErrorMonitor) RecordError(code utils.ErrorCode) {
	em.errorCounts[code]++
}

// GetErrorStats 获取错误统计
func (em *ErrorMonitor) GetErrorStats() map[utils.ErrorCode]int {
	return em.errorCounts
}

// Reset 重置统计
func (em *ErrorMonitor) Reset() {
	em.errorCounts = make(map[utils.ErrorCode]int)
	em.lastReset = time.Now()
}

// ShouldAlert 检查是否需要告警
func (em *ErrorMonitor) ShouldAlert(threshold int) bool {
	totalErrors := 0
	for _, count := range em.errorCounts {
		totalErrors += count
	}
	return totalErrors >= threshold
}

// 全局错误监控器
var globalErrorMonitor = NewErrorMonitor()

// RecordGlobalError 记录全局错误
func RecordGlobalError(code utils.ErrorCode) {
	globalErrorMonitor.RecordError(code)
}

// GetGlobalErrorStats 获取全局错误统计
func GetGlobalErrorStats() map[utils.ErrorCode]int {
	return globalErrorMonitor.GetErrorStats()
}

// 错误恢复和重试

// RetryableError 可重试错误接口
type RetryableError interface {
	error
	IsRetryable() bool
	GetRetryDelay() time.Duration
}

// RetryableAppError 可重试的应用错误
type RetryableAppError struct {
	*utils.AppError
	retryDelay time.Duration
}

// IsRetryable 是否可重试
func (e *RetryableAppError) IsRetryable() bool {
	return true
}

// GetRetryDelay 获取重试延迟
func (e *RetryableAppError) GetRetryDelay() time.Duration {
	return e.retryDelay
}

// NewRetryableError 创建可重试错误
func NewRetryableError(code utils.ErrorCode, message string, retryDelay time.Duration) *RetryableAppError {
	return &RetryableAppError{
		AppError:   utils.NewAppError(code, message),
		retryDelay: retryDelay,
	}
}

// Retry 重试函数
func Retry(fn func() error, maxRetries int, delay time.Duration) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			
			// 检查是否可重试
			if retryableErr, ok := err.(RetryableError); ok {
				if !retryableErr.IsRetryable() {
					return err
				}
				time.Sleep(retryableErr.GetRetryDelay())
			} else {
				time.Sleep(delay)
			}
			continue
		}
		return nil
	}
	
	return lastErr
}
