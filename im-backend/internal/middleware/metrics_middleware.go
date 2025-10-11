package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/controller"
)

// MetricsMiddleware Prometheus指标中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// HTTP请求总数
		controller.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		// HTTP请求耗时
		controller.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

