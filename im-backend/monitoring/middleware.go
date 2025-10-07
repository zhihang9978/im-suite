/**
 * 志航密信监控系统 - 中间件
 * 提供监控中间件，用于收集API请求指标
 */

package monitoring

import (
	"time"

	"github.com/gin-gonic/gin"
)

// MonitoringMiddleware 监控中间件
func MonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 记录请求开始
		if GlobalMetricsCollector != nil {
			GlobalMetricsCollector.IncrementRequestCount()
		}

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start)

		// 记录API指标
		if GlobalMetricsCollector != nil {
			userID := ""
			if user, exists := c.Get("user_id"); exists {
				userID = user.(string)
			}

			GlobalMetricsCollector.RecordAPIMetrics(
				c.FullPath(),
				c.Request.Method,
				duration,
				c.Writer.Status(),
				userID,
				c.ClientIP(),
			)

			// 如果是错误状态码，增加错误计数
			if c.Writer.Status() >= 400 {
				GlobalMetricsCollector.IncrementErrorCount()
			}
		}
	}
}

// ErrorMonitoringMiddleware 错误监控中间件
func ErrorMonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				if GlobalMetricsCollector != nil {
					userID := ""
					if user, exists := c.Get("user_id"); exists {
						userID = user.(string)
					}

					GlobalMetricsCollector.RecordError(
						"gin_error",
						err.Error(),
						"", // 堆栈信息
						userID,
						c.ClientIP(),
						c.FullPath(),
						"error",
					)
				}
			}
		}
	}
}

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()

		// 记录性能指标
		duration := time.Since(start)
		
		// 如果响应时间超过阈值，记录警告
		if duration > 5*time.Second {
			if GlobalMetricsCollector != nil {
				userID := ""
				if user, exists := c.Get("user_id"); exists {
					userID = user.(string)
				}

				GlobalMetricsCollector.RecordError(
					"slow_response",
					"响应时间过长",
					"",
					userID,
					c.ClientIP(),
					c.FullPath(),
					"warning",
				)
			}
		}
	}
}
