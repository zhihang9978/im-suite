/**
 * 志航密信监控系统 - 性能指标收集
 * 收集系统性能指标，包括响应时间、内存使用、CPU使用等
 */

package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu                sync.RWMutex
	metrics           map[string]interface{}
	websocketClients  map[*websocket.Conn]bool
	register          chan *websocket.Conn
	unregister        chan *websocket.Conn
	broadcast         chan []byte
	stopChan          chan bool
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   uint64    `json:"memory_usage"`
	MemoryPercent float64   `json:"memory_percent"`
	GoroutineCount int      `json:"goroutine_count"`
	ResponseTime  float64   `json:"response_time"`
	RequestCount  int64     `json:"request_count"`
	ErrorCount    int64     `json:"error_count"`
	ActiveUsers   int       `json:"active_users"`
}

// APIMetrics API指标
type APIMetrics struct {
	Endpoint     string        `json:"endpoint"`
	Method       string        `json:"method"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
	Timestamp    time.Time     `json:"timestamp"`
	UserID       string        `json:"user_id,omitempty"`
	IP           string        `json:"ip"`
}

// ErrorMetrics 错误指标
type ErrorMetrics struct {
	Timestamp   time.Time `json:"timestamp"`
	ErrorType   string    `json:"error_type"`
	ErrorMessage string   `json:"error_message"`
	Stack       string    `json:"stack"`
	UserID      string    `json:"user_id,omitempty"`
	IP          string    `json:"ip"`
	Endpoint    string    `json:"endpoint"`
	Severity    string    `json:"severity"`
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:          make(map[string]interface{}),
		websocketClients: make(map[*websocket.Conn]bool),
		register:         make(chan *websocket.Conn),
		unregister:       make(chan *websocket.Conn),
		broadcast:        make(chan []byte),
		stopChan:         make(chan bool),
	}
}

// Start 启动监控
func (mc *MetricsCollector) Start() {
	go mc.collectMetrics()
	go mc.handleWebSocket()
}

// Stop 停止监控
func (mc *MetricsCollector) Stop() {
	mc.stopChan <- true
}

// collectMetrics 收集性能指标
func (mc *MetricsCollector) collectMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := mc.gatherMetrics()
			mc.updateMetrics(metrics)
			mc.broadcastMetrics(metrics)
		case <-mc.stopChan:
			return
		}
	}
}

// gatherMetrics 收集系统指标
func (mc *MetricsCollector) gatherMetrics() *PerformanceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 计算CPU使用率（简化版本）
	cpuUsage := mc.calculateCPUUsage()
	
	// 计算内存使用
	memoryUsage := m.Alloc
	memoryPercent := float64(m.Alloc) / float64(m.Sys) * 100

	// 获取Goroutine数量
	goroutineCount := runtime.NumGoroutine()

	// 获取请求统计
	requestCount := mc.getRequestCount()
	errorCount := mc.getErrorCount()
	activeUsers := mc.getActiveUsers()

	return &PerformanceMetrics{
		Timestamp:      time.Now(),
		CPUUsage:       cpuUsage,
		MemoryUsage:    memoryUsage,
		MemoryPercent:  memoryPercent,
		GoroutineCount: goroutineCount,
		ResponseTime:   mc.getAverageResponseTime(),
		RequestCount:   requestCount,
		ErrorCount:     errorCount,
		ActiveUsers:    activeUsers,
	}
}

// calculateCPUUsage 计算CPU使用率
func (mc *MetricsCollector) calculateCPUUsage() float64 {
	// 简化的CPU使用率计算
	// 实际项目中可以使用更精确的方法
	return float64(runtime.NumGoroutine()) * 0.1
}

// updateMetrics 更新指标
func (mc *MetricsCollector) updateMetrics(metrics *PerformanceMetrics) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics["performance"] = metrics
	mc.metrics["last_updated"] = time.Now()
}

// broadcastMetrics 广播指标
func (mc *MetricsCollector) broadcastMetrics(metrics *PerformanceMetrics) {
	data, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("序列化指标失败: %v", err)
		return
	}

	select {
	case mc.broadcast <- data:
	default:
		// 如果广播通道满了，跳过这次广播
	}
}

// handleWebSocket 处理WebSocket连接
func (mc *MetricsCollector) handleWebSocket() {
	for {
		select {
		case client := <-mc.register:
			mc.websocketClients[client] = true
			log.Printf("监控客户端连接: %v", client.RemoteAddr())

		case client := <-mc.unregister:
			if _, ok := mc.websocketClients[client]; ok {
				delete(mc.websocketClients, client)
				client.Close()
				log.Printf("监控客户端断开: %v", client.RemoteAddr())
			}

		case message := <-mc.broadcast:
			for client := range mc.websocketClients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("发送监控数据失败: %v", err)
					client.Close()
					delete(mc.websocketClients, client)
				}
			}
		}
	}
}

// RegisterClient 注册WebSocket客户端
func (mc *MetricsCollector) RegisterClient(conn *websocket.Conn) {
	mc.register <- conn
}

// UnregisterClient 注销WebSocket客户端
func (mc *MetricsCollector) UnregisterClient(conn *websocket.Conn) {
	mc.unregister <- conn
}

// GetMetrics 获取当前指标
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// 返回指标的副本
	metricsCopy := make(map[string]interface{})
	for k, v := range mc.metrics {
		metricsCopy[k] = v
	}
	return metricsCopy
}

// RecordAPIMetrics 记录API指标
func (mc *MetricsCollector) RecordAPIMetrics(endpoint, method string, responseTime time.Duration, statusCode int, userID, ip string) {
	apiMetrics := &APIMetrics{
		Endpoint:     endpoint,
		Method:       method,
		ResponseTime: responseTime,
		StatusCode:   statusCode,
		Timestamp:    time.Now(),
		UserID:       userID,
		IP:           ip,
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 更新API统计
	if mc.metrics["api_stats"] == nil {
		mc.metrics["api_stats"] = make(map[string]int64)
	}
	apiStats := mc.metrics["api_stats"].(map[string]int64)
	apiStats[fmt.Sprintf("%s_%s", method, endpoint)]++

	// 更新响应时间统计
	if mc.metrics["response_times"] == nil {
		mc.metrics["response_times"] = []time.Duration{}
	}
	responseTimes := mc.metrics["response_times"].([]time.Duration)
	if len(responseTimes) > 1000 {
		// 保持最近1000个响应时间
		responseTimes = responseTimes[1:]
	}
	mc.metrics["response_times"] = append(responseTimes, responseTime)
}

// RecordError 记录错误
func (mc *MetricsCollector) RecordError(errorType, errorMessage, stack, userID, ip, endpoint, severity string) {
	errorMetrics := &ErrorMetrics{
		Timestamp:    time.Now(),
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		Stack:        stack,
		UserID:       userID,
		IP:           ip,
		Endpoint:     endpoint,
		Severity:     severity,
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 更新错误统计
	if mc.metrics["error_stats"] == nil {
		mc.metrics["error_stats"] = make(map[string]int64)
	}
	errorStats := mc.metrics["error_stats"].(map[string]int64)
	errorStats[errorType]++

	// 记录错误详情
	if mc.metrics["errors"] == nil {
		mc.metrics["errors"] = []*ErrorMetrics{}
	}
	errors := mc.metrics["errors"].([]*ErrorMetrics)
	if len(errors) > 100 {
		// 保持最近100个错误
		errors = errors[1:]
	}
	mc.metrics["errors"] = append(errors, errorMetrics)
}

// getRequestCount 获取请求数量
func (mc *MetricsCollector) getRequestCount() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.metrics["request_count"] == nil {
		return 0
	}
	return mc.metrics["request_count"].(int64)
}

// getErrorCount 获取错误数量
func (mc *MetricsCollector) getErrorCount() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.metrics["error_count"] == nil {
		return 0
	}
	return mc.metrics["error_count"].(int64)
}

// getActiveUsers 获取活跃用户数
func (mc *MetricsCollector) getActiveUsers() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.metrics["active_users"] == nil {
		return 0
	}
	return mc.metrics["active_users"].(int)
}

// getAverageResponseTime 获取平均响应时间
func (mc *MetricsCollector) getAverageResponseTime() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.metrics["response_times"] == nil {
		return 0
	}
	responseTimes := mc.metrics["response_times"].([]time.Duration)
	if len(responseTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, rt := range responseTimes {
		total += rt
	}
	return float64(total) / float64(len(responseTimes)) / float64(time.Millisecond)
}

// IncrementRequestCount 增加请求计数
func (mc *MetricsCollector) IncrementRequestCount() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.metrics["request_count"] == nil {
		mc.metrics["request_count"] = int64(0)
	}
	mc.metrics["request_count"] = mc.metrics["request_count"].(int64) + 1
}

// IncrementErrorCount 增加错误计数
func (mc *MetricsCollector) IncrementErrorCount() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.metrics["error_count"] == nil {
		mc.metrics["error_count"] = int64(0)
	}
	mc.metrics["error_count"] = mc.metrics["error_count"].(int64) + 1
}

// SetActiveUsers 设置活跃用户数
func (mc *MetricsCollector) SetActiveUsers(count int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metrics["active_users"] = count
}

// 全局监控实例
var GlobalMetricsCollector *MetricsCollector

// InitMonitoring 初始化监控
func InitMonitoring() {
	GlobalMetricsCollector = NewMetricsCollector()
	GlobalMetricsCollector.Start()
}

// GetGlobalMetrics 获取全局指标
func GetGlobalMetrics() map[string]interface{} {
	if GlobalMetricsCollector == nil {
		return make(map[string]interface{})
	}
	return GlobalMetricsCollector.GetMetrics()
}
