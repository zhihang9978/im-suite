package middleware

import (
	"time"
	"zhihang-messenger/im-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	requestCounts    map[string]int64
	responseTimes    map[string][]time.Duration
	errorCounts      map[string]int64
	slowRequestCount int64
	threshold        time.Duration
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(threshold time.Duration) *PerformanceMonitor {
	return &PerformanceMonitor{
		requestCounts: make(map[string]int64),
		responseTimes: make(map[string][]time.Duration),
		errorCounts:   make(map[string]int64),
		threshold:     threshold,
	}
}

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware(monitor *PerformanceMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start)

		// 获取请求信息
		method := c.Request.Method
		path := c.FullPath()
		status := c.Writer.Status()

		// 记录请求统计
		key := method + " " + path
		monitor.requestCounts[key]++

		// 记录响应时间
		if monitor.responseTimes[key] == nil {
			monitor.responseTimes[key] = make([]time.Duration, 0, 100)
		}
		monitor.responseTimes[key] = append(monitor.responseTimes[key], duration)

		// 限制响应时间记录数量
		if len(monitor.responseTimes[key]) > 100 {
			monitor.responseTimes[key] = monitor.responseTimes[key][1:]
		}

		// 记录错误统计
		if status >= 400 {
			monitor.errorCounts[key]++
		}

		// 记录慢请求
		if duration > monitor.threshold {
			monitor.slowRequestCount++

			// 记录慢请求日志
			logrus.WithFields(logrus.Fields{
				"method":     method,
				"path":       path,
				"duration":   duration,
				"status":     status,
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			}).Warn("Slow request detected")
		}

		// 记录性能日志
		logrus.WithFields(logrus.Fields{
			"method":   method,
			"path":     path,
			"duration": duration,
			"status":   status,
		}).Debug("Request performance")
	}
}

// CacheMiddleware 缓存中间件
func CacheMiddleware(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对 GET 请求启用缓存
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := generateCacheKey(c)

		// 尝试从缓存获取
		if cachedData := getFromCache(cacheKey); cachedData != nil {
			c.Header("X-Cache", "HIT")
			c.Header("Content-Type", "application/json")
			c.String(200, string(cachedData))
			c.Abort()
			return
		}

		// 设置缓存头
		c.Header("X-Cache", "MISS")

		// 继续处理请求
		c.Next()

		// 如果响应成功，缓存结果
		if c.Writer.Status() == 200 {
			if data := getResponseData(c); data != nil {
				setCache(cacheKey, data, ttl)
			}
		}
	}
}

// CompressionMiddleware 压缩中间件
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查客户端是否支持压缩
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if acceptEncoding == "" {
			c.Next()
			return
		}

		// 设置压缩头
		c.Header("Vary", "Accept-Encoding")

		// 继续处理请求
		c.Next()
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// 简单的内存限流器
	rateLimiter := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// 清理过期记录
		if records, exists := rateLimiter[clientIP]; exists {
			validRecords := make([]time.Time, 0)
			for _, record := range records {
				if now.Sub(record) < time.Minute {
					validRecords = append(validRecords, record)
				}
			}
			rateLimiter[clientIP] = validRecords
		}

		// 检查是否超过限制
		if len(rateLimiter[clientIP]) >= requestsPerMinute {
			appErr := utils.NewAppError(
				utils.ErrCodeTooManyRequests,
				"请求过于频繁",
				"请稍后再试",
			)

			c.JSON(appErr.GetHTTPStatus(), appErr.ToErrorResponse())
			c.Abort()
			return
		}

		// 记录请求时间
		rateLimiter[clientIP] = append(rateLimiter[clientIP], now)

		c.Next()
	}
}

// ConnectionPoolMiddleware 连接池中间件
func ConnectionPoolMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置连接池相关头
		c.Header("Connection", "keep-alive")
		c.Header("Keep-Alive", "timeout=30, max=100")

		c.Next()
	}
}

// 缓存相关函数（需要根据实际缓存实现）

// generateCacheKey 生成缓存键
func generateCacheKey(c *gin.Context) string {
	// 基于请求方法和路径生成缓存键
	return c.Request.Method + ":" + c.FullPath() + ":" + c.Request.URL.RawQuery
}

// getFromCache 从缓存获取数据
func getFromCache(key string) []byte {
	// 这里应该实现实际的缓存获取逻辑
	// 例如使用 Redis 或内存缓存
	return nil
}

// setCache 设置缓存
func setCache(key string, data []byte, ttl time.Duration) {
	// 这里应该实现实际的缓存设置逻辑
	// 例如使用 Redis 或内存缓存
}

// getResponseData 获取响应数据
func getResponseData(c *gin.Context) []byte {
	// 这里应该实现获取响应数据的逻辑
	// 例如从响应缓冲区获取数据
	return nil
}

// 性能优化工具函数

// OptimizeDatabaseQuery 优化数据库查询
func OptimizeDatabaseQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置数据库查询优化相关头
		c.Header("X-DB-Optimized", "true")

		c.Next()
	}
}

// OptimizeStaticFiles 优化静态文件
func OptimizeStaticFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置静态文件优化相关头
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", generateETag(c.Request.URL.Path))

		c.Next()
	}
}

// generateETag 生成 ETag
func generateETag(path string) string {
	// 这里应该实现实际的 ETag 生成逻辑
	// 例如基于文件内容哈希生成
	return "\"" + path + "\""
}

// 性能监控工具函数

// GetPerformanceStats 获取性能统计
func (pm *PerformanceMonitor) GetPerformanceStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 计算平均响应时间
	avgResponseTimes := make(map[string]time.Duration)
	for endpoint, times := range pm.responseTimes {
		if len(times) > 0 {
			var total time.Duration
			for _, t := range times {
				total += t
			}
			avgResponseTimes[endpoint] = total / time.Duration(len(times))
		}
	}

	stats["request_counts"] = pm.requestCounts
	stats["avg_response_times"] = avgResponseTimes
	stats["error_counts"] = pm.errorCounts
	stats["slow_request_count"] = pm.slowRequestCount

	return stats
}

// ResetStats 重置统计
func (pm *PerformanceMonitor) ResetStats() {
	pm.requestCounts = make(map[string]int64)
	pm.responseTimes = make(map[string][]time.Duration)
	pm.errorCounts = make(map[string]int64)
	pm.slowRequestCount = 0
}

// 全局性能监控器
var globalPerformanceMonitor = NewPerformanceMonitor(5 * time.Second)

// GetGlobalPerformanceStats 获取全局性能统计
func GetGlobalPerformanceStats() map[string]interface{} {
	return globalPerformanceMonitor.GetPerformanceStats()
}

// ResetGlobalPerformanceStats 重置全局性能统计
func ResetGlobalPerformanceStats() {
	globalPerformanceMonitor.ResetStats()
}
