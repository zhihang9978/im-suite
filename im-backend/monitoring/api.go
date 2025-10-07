/**
 * 志航密信监控系统 - API接口
 * 提供监控数据的API接口
 */

package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// MonitorAPI 监控API处理器
type MonitorAPI struct {
	collector *MetricsCollector
	upgrader  websocket.Upgrader
}

// NewMonitorAPI 创建监控API
func NewMonitorAPI() *MonitorAPI {
	return &MonitorAPI{
		collector: GlobalMetricsCollector,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
	}
}

// GetMetrics 获取监控指标
func (api *MonitorAPI) GetMetrics(c *gin.Context) {
	if api.collector == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "监控服务未初始化",
		})
		return
	}

	metrics := api.collector.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   metrics,
		"timestamp": time.Now(),
	})
}

// GetPerformanceMetrics 获取性能指标
func (api *MonitorAPI) GetPerformanceMetrics(c *gin.Context) {
	if api.collector == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "监控服务未初始化",
		})
		return
	}

	metrics := api.collector.GetMetrics()
	performance, exists := metrics["performance"]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "性能指标未找到",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   performance,
		"timestamp": time.Now(),
	})
}

// GetAPIMetrics 获取API指标
func (api *MonitorAPI) GetAPIMetrics(c *gin.Context) {
	if api.collector == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "监控服务未初始化",
		})
		return
	}

	metrics := api.collector.GetMetrics()
	apiStats, exists := metrics["api_stats"]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API指标未找到",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   apiStats,
		"timestamp": time.Now(),
	})
}

// GetErrorMetrics 获取错误指标
func (api *MonitorAPI) GetErrorMetrics(c *gin.Context) {
	if api.collector == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "监控服务未初始化",
		})
		return
	}

	metrics := api.collector.GetMetrics()
	errorStats, exists := metrics["error_stats"]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "错误指标未找到",
		})
		return
	}

	errors, exists := metrics["errors"]
	if !exists {
		errors = []*ErrorMetrics{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"error_stats": errorStats,
			"recent_errors": errors,
		},
		"timestamp": time.Now(),
	})
}

// GetHealthStatus 获取健康状态
func (api *MonitorAPI) GetHealthStatus(c *gin.Context) {
	if api.collector == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"message": "监控服务未初始化",
		})
		return
	}

	metrics := api.collector.GetMetrics()
	performance, exists := metrics["performance"]
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"status": "unknown",
			"message": "无法获取性能指标",
		})
		return
	}

	// 简化的健康检查逻辑
	healthStatus := "healthy"
	message := "系统运行正常"

	// 这里可以添加更复杂的健康检查逻辑
	// 例如检查CPU使用率、内存使用率等

	c.JSON(http.StatusOK, gin.H{
		"status": healthStatus,
		"message": message,
		"timestamp": time.Now(),
		"metrics": performance,
	})
}

// WebSocketMetrics WebSocket监控数据流
func (api *MonitorAPI) WebSocketMetrics(c *gin.Context) {
	conn, err := api.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "WebSocket升级失败",
		})
		return
	}
	defer conn.Close()

	// 注册客户端
	if api.collector != nil {
		api.collector.RegisterClient(conn)
		defer api.collector.UnregisterClient(conn)
	}

	// 发送初始数据
	if api.collector != nil {
		metrics := api.collector.GetMetrics()
		conn.WriteJSON(gin.H{
			"type": "initial",
			"data": metrics,
		})
	}

	// 保持连接
	for {
		// 读取客户端消息
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// GetMetricsHistory 获取指标历史
func (api *MonitorAPI) GetMetricsHistory(c *gin.Context) {
	// 获取查询参数
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的时间参数",
		})
		return
	}

	// 这里应该从数据库或缓存中获取历史数据
	// 目前返回模拟数据
	history := api.generateMockHistory(hours)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   history,
		"timestamp": time.Now(),
	})
}

// generateMockHistory 生成模拟历史数据
func (api *MonitorAPI) generateMockHistory(hours int) []gin.H {
	history := make([]gin.H, 0)
	now := time.Now()

	for i := 0; i < hours; i++ {
		timestamp := now.Add(-time.Duration(i) * time.Hour)
		history = append(history, gin.H{
			"timestamp": timestamp,
			"cpu_usage": 20 + float64(i%10),
			"memory_usage": 100 + uint64(i%50),
			"request_count": 1000 + i*100,
			"error_count": i%10,
		})
	}

	return history
}

// SetupMonitorRoutes 设置监控路由
func SetupMonitorRoutes(r *gin.Engine) {
	api := NewMonitorAPI()
	
	// 监控API组
	monitor := r.Group("/api/monitor")
	{
		// 基础监控接口
		monitor.GET("/metrics", api.GetMetrics)
		monitor.GET("/performance", api.GetPerformanceMetrics)
		monitor.GET("/api", api.GetAPIMetrics)
		monitor.GET("/errors", api.GetErrorMetrics)
		monitor.GET("/health", api.GetHealthStatus)
		monitor.GET("/history", api.GetMetricsHistory)
		
		// WebSocket监控流
		monitor.GET("/ws", api.WebSocketMetrics)
	}
}
