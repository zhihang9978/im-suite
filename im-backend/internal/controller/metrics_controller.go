package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP请求总数
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求耗时
	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求耗时（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 活跃用户数
	ActiveUsersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "im_active_users_total",
			Help: "当前活跃用户数",
		},
	)

	// 消息发送总数
	MessagesSentTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "messages_sent_total",
			Help: "消息发送总数",
		},
	)

	// WebRTC连接数
	WebRTCConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "webrtc_connections_active",
			Help: "当前WebRTC连接数",
		},
	)

	// 数据库连接池
	MySQLConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_connections_active",
			Help: "MySQL活跃连接数",
		},
	)

	MySQLConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_connections_idle",
			Help: "MySQL空闲连接数",
		},
	)

	// Redis内存使用
	RedisMemoryUsedBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_memory_used_bytes",
			Help: "Redis内存使用（字节）",
		},
	)
)

func init() {
	// 注册所有指标
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(ActiveUsersTotal)
	prometheus.MustRegister(MessagesSentTotal)
	prometheus.MustRegister(WebRTCConnectionsActive)
	prometheus.MustRegister(MySQLConnectionsActive)
	prometheus.MustRegister(MySQLConnectionsIdle)
	prometheus.MustRegister(RedisMemoryUsedBytes)
}

// MetricsHandler Prometheus metrics处理器
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

