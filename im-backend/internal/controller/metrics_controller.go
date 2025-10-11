package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// IM特定指标（不与middleware/metrics.go冲突）
	
	// 活跃用户数
	ActiveUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "im_active_users_total",
			Help: "当前活跃用户数",
		},
	)

	// WebRTC连接数
	WebRTCConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "webrtc_connections_active",
			Help: "当前WebRTC连接数",
		},
	)

	// 数据库连接池
	MySQLConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_connections_active",
			Help: "MySQL活跃连接数",
		},
	)

	MySQLConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mysql_connections_idle",
			Help: "MySQL空闲连接数",
		},
	)

	// Redis内存使用
	RedisMemoryUsedBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_memory_used_bytes",
			Help: "Redis内存使用（字节）",
		},
	)
)

// MetricsHandler Prometheus metrics处理器
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

