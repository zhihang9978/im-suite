package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// Prometheus指标（S+可观测性）
var (
	// HTTP请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求延迟
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// 当前在线用户数
	onlineUsersGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "online_users_current",
			Help: "Current number of online users",
		},
	)

	// 消息发送速率
	messagesSentTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "messages_sent_total",
			Help: "Total number of messages sent",
		},
	)

	// 数据库查询延迟
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	// Redis缓存命中率
	redisCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Total number of Redis cache hits",
		},
	)

	redisCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Total number of Redis cache misses",
		},
	)
)

// MetricsMiddleware Prometheus指标中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		// 记录请求指标
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)

		// 慢请求告警（>1秒）
		if duration > 1.0 {
			logrus.Warnf("慢请求: %s %s - %v秒", method, path, duration)
		}
	}
}

// RecordDBQuery 记录数据库查询性能
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit() {
	redisCacheHits.Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss() {
	redisCacheMisses.Inc()
}

// RecordMessageSent 记录消息发送
func RecordMessageSent() {
	messagesSentTotal.Inc()
}

// UpdateOnlineUsers 更新在线用户数
func UpdateOnlineUsers(count float64) {
	onlineUsersGauge.Set(count)
}

