package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

// ErrorCode 错误码类型
type ErrorCode int

// 业务错误码定义
const (
	// 用户相关错误 (10000-19999)
	ErrCodeUserNotFound     ErrorCode = 10001 // 用户不存在
	ErrCodeUserExists       ErrorCode = 10002 // 用户已存在
	ErrCodePasswordError    ErrorCode = 10003 // 密码错误
	ErrCodeCodeError        ErrorCode = 10004 // 验证码错误
	ErrCodeUserDisabled     ErrorCode = 10005 // 用户被禁用
	ErrCodePhoneExists      ErrorCode = 10006 // 手机号已存在
	ErrCodeUsernameExists   ErrorCode = 10007 // 用户名已存在
	ErrCodeTokenExpired     ErrorCode = 10008 // 令牌已过期
	ErrCodeTokenInvalid     ErrorCode = 10009 // 令牌无效
	ErrCodePermissionDenied ErrorCode = 10010 // 权限不足

	// 聊天相关错误 (20000-29999)
	ErrCodeChatNotFound    ErrorCode = 20001 // 聊天不存在
	ErrCodeNotChatMember   ErrorCode = 20002 // 用户不是聊天成员
	ErrCodeMessageNotFound ErrorCode = 20003 // 消息不存在
	ErrCodeNoPermission    ErrorCode = 20004 // 无权限操作此消息
	ErrCodeChatFull        ErrorCode = 20005 // 聊天已满
	ErrCodeInvalidChatType ErrorCode = 20006 // 无效的聊天类型

	// 文件相关错误 (30000-39999)
	ErrCodeFileNotFound     ErrorCode = 30001 // 文件不存在
	ErrCodeFileTypeInvalid  ErrorCode = 30002 // 文件类型不支持
	ErrCodeFileSizeExceeded ErrorCode = 30003 // 文件大小超限
	ErrCodeUploadFailed     ErrorCode = 30004 // 文件上传失败
	ErrCodeDownloadFailed   ErrorCode = 30005 // 文件下载失败

	// 系统相关错误 (40000-49999)
	ErrCodeTooManyRequests ErrorCode = 40001 // 请求过于频繁
	ErrCodeInvalidParams   ErrorCode = 40002 // 请求参数错误
	ErrCodeDatabaseError   ErrorCode = 40003 // 数据库错误
	ErrCodeCacheError      ErrorCode = 40004 // 缓存错误
	ErrCodeNetworkError    ErrorCode = 40005 // 网络错误

	// 服务器错误 (50000-59999)
	ErrCodeInternalError      ErrorCode = 50001 // 服务器内部错误
	ErrCodeServiceUnavailable ErrorCode = 50002 // 服务不可用
	ErrCodeTimeout            ErrorCode = 50003 // 请求超时
)

// AppError 应用错误结构
type AppError struct {
	Code      ErrorCode `json:"code"`                 // 错误码
	Message   string    `json:"message"`              // 错误消息
	Details   string    `json:"details,omitempty"`    // 详细错误信息
	RequestID string    `json:"request_id,omitempty"` // 请求ID
	Timestamp string    `json:"timestamp"`            // 时间戳
	Stack     string    `json:"stack,omitempty"`      // 错误堆栈（仅开发环境）
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewAppError 创建新的应用错误
func NewAppError(code ErrorCode, message string, details ...string) *AppError {
	var detail string
	if len(details) > 0 {
		detail = details[0]
	}

	return &AppError{
		Code:      code,
		Message:   message,
		Details:   detail,
		Timestamp: getCurrentTimestamp(),
	}
}

// NewAppErrorWithStack 创建带堆栈信息的应用错误
func NewAppErrorWithStack(code ErrorCode, message string, details ...string) *AppError {
	err := NewAppError(code, message, details...)
	err.Stack = getStackTrace()
	return err
}

// SetRequestID 设置请求ID
func (e *AppError) SetRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// GetHTTPStatus 获取对应的HTTP状态码
func (e *AppError) GetHTTPStatus() int {
	switch {
	case e.Code >= 10000 && e.Code < 20000:
		return http.StatusUnauthorized
	case e.Code >= 20000 && e.Code < 30000:
		return http.StatusForbidden
	case e.Code >= 30000 && e.Code < 40000:
		return http.StatusBadRequest
	case e.Code >= 40000 && e.Code < 50000:
		return http.StatusTooManyRequests
	case e.Code >= 50000:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ToJSON 转换为JSON字符串
func (e *AppError) ToJSON() string {
	jsonData, _ := json.Marshal(e)
	return string(jsonData)
}

// ErrorMessages 错误消息映射
var ErrorMessages = map[ErrorCode]string{
	// 用户相关错误
	ErrCodeUserNotFound:     "用户不存在",
	ErrCodeUserExists:       "用户已存在",
	ErrCodePasswordError:    "密码错误",
	ErrCodeCodeError:        "验证码错误",
	ErrCodeUserDisabled:     "用户已被禁用",
	ErrCodePhoneExists:      "手机号已存在",
	ErrCodeUsernameExists:   "用户名已存在",
	ErrCodeTokenExpired:     "令牌已过期",
	ErrCodeTokenInvalid:     "令牌无效",
	ErrCodePermissionDenied: "权限不足",

	// 聊天相关错误
	ErrCodeChatNotFound:    "聊天不存在",
	ErrCodeNotChatMember:   "用户不是聊天成员",
	ErrCodeMessageNotFound: "消息不存在",
	ErrCodeNoPermission:    "无权限操作此消息",
	ErrCodeChatFull:        "聊天已满",
	ErrCodeInvalidChatType: "无效的聊天类型",

	// 文件相关错误
	ErrCodeFileNotFound:     "文件不存在",
	ErrCodeFileTypeInvalid:  "文件类型不支持",
	ErrCodeFileSizeExceeded: "文件大小超限",
	ErrCodeUploadFailed:     "文件上传失败",
	ErrCodeDownloadFailed:   "文件下载失败",

	// 系统相关错误
	ErrCodeTooManyRequests: "请求过于频繁",
	ErrCodeInvalidParams:   "请求参数错误",
	ErrCodeDatabaseError:   "数据库错误",
	ErrCodeCacheError:      "缓存错误",
	ErrCodeNetworkError:    "网络错误",

	// 服务器错误
	ErrCodeInternalError:      "服务器内部错误",
	ErrCodeServiceUnavailable: "服务不可用",
	ErrCodeTimeout:            "请求超时",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code ErrorCode) string {
	if message, exists := ErrorMessages[code]; exists {
		return message
	}
	return "未知错误"
}

// 错误处理工具函数

// WrapError 包装错误
func WrapError(err error, code ErrorCode, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewAppError(code, message, err.Error())
}

// WrapErrorWithStack 包装错误并添加堆栈信息
func WrapErrorWithStack(err error, code ErrorCode, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewAppErrorWithStack(code, message, err.Error())
}

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError 获取应用错误
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// 错误处理中间件相关

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// ToErrorResponse 转换为错误响应
func (e *AppError) ToErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Code:      e.Code,
		Message:   e.Message,
		Details:   e.Details,
		RequestID: e.RequestID,
		Timestamp: e.Timestamp,
	}
}

// 错误恢复函数

// RecoverPanic 恢复panic并转换为错误
func RecoverPanic() *AppError {
	if r := recover(); r != nil {
		var message string
		var details string

		switch v := r.(type) {
		case error:
			message = v.Error()
		case string:
			message = v
		default:
			message = fmt.Sprintf("%v", v)
		}

		// 获取堆栈信息
		stack := getStackTrace()
		details = fmt.Sprintf("Panic recovered: %s\nStack: %s", message, stack)

		return NewAppError(ErrCodeInternalError, "服务器内部错误", details)
	}
	return nil
}

// 工具函数

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", getCurrentTime().Unix())
}

// getStackTrace 获取堆栈信息
func getStackTrace() string {
	const size = 4096
	buf := make([]byte, size)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// getCurrentTime 获取当前时间（这里需要根据实际情况实现）
func getCurrentTime() Time {
	// 这里应该返回当前时间
	// 为了简化，返回一个时间值
	return Time{}
}

// Time 时间类型（需要根据实际情况定义）
type Time struct {
	// 时间字段
}

// Unix 返回Unix时间戳
func (t Time) Unix() int64 {
	// 返回当前时间的Unix时间戳
	return 0 // 实际实现中应该返回真实时间戳
}

// 错误日志记录

// LogError 记录错误日志
func LogError(err *AppError, context map[string]interface{}) {
	// 这里应该使用实际的日志库记录错误
	// 例如：logrus, zap 等
	fmt.Printf("Error logged: %s\nContext: %+v\n", err.ToJSON(), context)
}

// LogErrorWithFields 记录带字段的错误日志
func LogErrorWithFields(err *AppError, fields map[string]interface{}) {
	// 合并错误信息和字段
	logData := map[string]interface{}{
		"error_code":    err.Code,
		"error_message": err.Message,
		"error_details": err.Details,
		"request_id":    err.RequestID,
		"timestamp":     err.Timestamp,
	}

	// 添加额外字段
	for k, v := range fields {
		logData[k] = v
	}

	LogError(err, logData)
}
