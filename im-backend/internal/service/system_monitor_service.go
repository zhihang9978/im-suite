package service

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
)

// SystemMonitorService 系统监控服务
type SystemMonitorService struct {
	db    *gorm.DB
	redis *redis.Client
	ctx   context.Context
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	Timestamp        time.Time `json:"timestamp"`
	CPUUsage         float64   `json:"cpu_usage"`
	MemoryUsage      float64   `json:"memory_usage"`
	MemoryTotal      uint64    `json:"memory_total"`
	MemoryUsed       uint64    `json:"memory_used"`
	DiskUsage        float64   `json:"disk_usage"`
	DiskTotal        uint64    `json:"disk_total"`
	DiskUsed         uint64    `json:"disk_used"`
	NetworkIn        uint64    `json:"network_in"`
	NetworkOut       uint64    `json:"network_out"`
	GoroutineCount   int       `json:"goroutine_count"`
	HeapAlloc        uint64    `json:"heap_alloc"`
	HeapSys          uint64    `json:"heap_sys"`
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	Connections     int       `json:"connections"`
	MaxConnections  int       `json:"max_connections"`
	SlowQueries     int64     `json:"slow_queries"`
	QueriesPerSecond float64  `json:"queries_per_second"`
	DatabaseSize    int64     `json:"database_size"`
	TableCount      int       `json:"table_count"`
}

// RedisMetrics Redis指标
type RedisMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	UsedMemory      int64     `json:"used_memory"`
	UsedMemoryPeak  int64     `json:"used_memory_peak"`
	ConnectedClients int      `json:"connected_clients"`
	TotalCommands   int64     `json:"total_commands"`
	KeyspaceHits    int64     `json:"keyspace_hits"`
	KeyspaceMisses  int64     `json:"keyspace_misses"`
	EvictedKeys     int64     `json:"evicted_keys"`
}

// Alert 告警信息
type Alert struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
	AlertType   string    `json:"alert_type"` // cpu, memory, disk, database, redis
	Severity    string    `json:"severity"` // info, warning, error, critical
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Resolved    bool      `json:"resolved" gorm:"default:false"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy  uint      `json:"resolved_by,omitempty"`
}

// NewSystemMonitorService 创建系统监控服务
func NewSystemMonitorService() *SystemMonitorService {
	return &SystemMonitorService{
		db:    config.DB,
		redis: config.Redis,
		ctx:   context.Background(),
	}
}

// GetSystemMetrics 获取系统指标
func (s *SystemMonitorService) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPUUsage = cpuPercent[0]
	}

	// 内存使用情况
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemoryUsage = memInfo.UsedPercent
		metrics.MemoryTotal = memInfo.Total
		metrics.MemoryUsed = memInfo.Used
	}

	// 磁盘使用情况
	diskInfo, err := disk.Usage("/")
	if err == nil {
		metrics.DiskUsage = diskInfo.UsedPercent
		metrics.DiskTotal = diskInfo.Total
		metrics.DiskUsed = diskInfo.Used
	}

	// 网络IO
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		metrics.NetworkIn = netIO[0].BytesRecv
		metrics.NetworkOut = netIO[0].BytesSent
	}

	// Go运行时指标
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.GoroutineCount = runtime.NumGoroutine()
	metrics.HeapAlloc = memStats.HeapAlloc
	metrics.HeapSys = memStats.HeapSys

	// 保存到Redis用于历史记录
	s.saveMetricsHistory(metrics)

	return metrics, nil
}

// GetDatabaseMetrics 获取数据库指标
func (s *SystemMonitorService) GetDatabaseMetrics() (*DatabaseMetrics, error) {
	metrics := &DatabaseMetrics{
		Timestamp: time.Now(),
	}

	// 获取数据库连接数
	sqlDB, err := s.db.DB()
	if err == nil {
		stats := sqlDB.Stats()
		metrics.Connections = stats.OpenConnections
		metrics.MaxConnections = stats.MaxOpenConnections
	}

	// 获取数据库大小
	var result struct {
		Size int64
	}
	s.db.Raw(`
		SELECT SUM(data_length + index_length) as size
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`).Scan(&result)
	metrics.DatabaseSize = result.Size

	// 获取表数量
	var tableCount int
	s.db.Raw(`
		SELECT COUNT(*) as count
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`).Scan(&tableCount)
	metrics.TableCount = tableCount

	return metrics, nil
}

// GetRedisMetrics 获取Redis指标
func (s *SystemMonitorService) GetRedisMetrics() (*RedisMetrics, error) {
	metrics := &RedisMetrics{
		Timestamp: time.Now(),
	}

	// 获取Redis信息
	info, err := s.redis.Info(s.ctx, "memory", "stats", "clients").Result()
	if err != nil {
		return nil, err
	}

	// 解析info字符串（简化版本）
	// 实际应该解析info中的具体字段
	metrics.UsedMemory = 0
	metrics.ConnectedClients = 0

	return metrics, nil
}

// saveMetricsHistory 保存指标历史
func (s *SystemMonitorService) saveMetricsHistory(metrics *SystemMetrics) {
	key := "system:metrics:history"
	data, _ := json.Marshal(metrics)
	
	s.redis.LPush(s.ctx, key, data)
	s.redis.LTrim(s.ctx, key, 0, 999) // 保留最近1000条记录
	s.redis.Expire(s.ctx, key, 7*24*time.Hour) // 保留7天
}

// GetMetricsHistory 获取指标历史
func (s *SystemMonitorService) GetMetricsHistory(duration time.Duration) ([]SystemMetrics, error) {
	key := "system:metrics:history"
	
	// 计算需要获取的记录数
	count := int64(duration.Minutes()) // 假设每分钟一条记录
	
	records, err := s.redis.LRange(s.ctx, key, 0, count-1).Result()
	if err != nil {
		return nil, err
	}

	metrics := make([]SystemMetrics, 0, len(records))
	for _, record := range records {
		var metric SystemMetrics
		if err := json.Unmarshal([]byte(record), &metric); err == nil {
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

// CheckThresholds 检查阈值并生成告警
func (s *SystemMonitorService) CheckThresholds(metrics *SystemMetrics) {
	// CPU使用率告警
	if metrics.CPUUsage > 80 {
		s.CreateAlert("cpu", "critical", "CPU使用率过高", 
			fmt.Sprintf("当前CPU使用率: %.1f%%", metrics.CPUUsage),
			metrics.CPUUsage, 80)
	} else if metrics.CPUUsage > 60 {
		s.CreateAlert("cpu", "warning", "CPU使用率较高",
			fmt.Sprintf("当前CPU使用率: %.1f%%", metrics.CPUUsage),
			metrics.CPUUsage, 60)
	}

	// 内存使用率告警
	if metrics.MemoryUsage > 85 {
		s.CreateAlert("memory", "critical", "内存使用率过高",
			fmt.Sprintf("当前内存使用率: %.1f%%", metrics.MemoryUsage),
			metrics.MemoryUsage, 85)
	} else if metrics.MemoryUsage > 70 {
		s.CreateAlert("memory", "warning", "内存使用率较高",
			fmt.Sprintf("当前内存使用率: %.1f%%", metrics.MemoryUsage),
			metrics.MemoryUsage, 70)
	}

	// 磁盘使用率告警
	if metrics.DiskUsage > 90 {
		s.CreateAlert("disk", "critical", "磁盘空间不足",
			fmt.Sprintf("当前磁盘使用率: %.1f%%", metrics.DiskUsage),
			metrics.DiskUsage, 90)
	} else if metrics.DiskUsage > 80 {
		s.CreateAlert("disk", "warning", "磁盘空间较少",
			fmt.Sprintf("当前磁盘使用率: %.1f%%", metrics.DiskUsage),
			metrics.DiskUsage, 80)
	}
}

// CreateAlert 创建告警
func (s *SystemMonitorService) CreateAlert(alertType, severity, title, message string, value, threshold float64) {
	alert := &Alert{
		AlertType: alertType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Value:     value,
		Threshold: threshold,
		Resolved:  false,
	}

	// 保存到数据库
	s.db.Create(alert)

	// 发送到告警队列
	key := "system:alerts:queue"
	data, _ := json.Marshal(alert)
	s.redis.LPush(s.ctx, key, data)

	logrus.WithFields(logrus.Fields{
		"alert_type": alertType,
		"severity":   severity,
		"title":      title,
		"value":      value,
		"threshold":  threshold,
	}).Warn("系统告警")
}

// GetActiveAlerts 获取活跃告警
func (s *SystemMonitorService) GetActiveAlerts() ([]Alert, error) {
	var alerts []Alert
	err := s.db.Where("resolved = ?", false).
		Order("created_at DESC").
		Limit(100).
		Find(&alerts).Error
	
	return alerts, err
}

// ResolveAlert 解决告警
func (s *SystemMonitorService) ResolveAlert(alertID, adminID uint) error {
	now := time.Now()
	return s.db.Model(&Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"resolved":    true,
			"resolved_at": now,
			"resolved_by": adminID,
		}).Error
}

// StartMonitoring 启动监控
func (s *SystemMonitorService) StartMonitoring() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	logrus.Info("系统监控服务已启动")

	for {
		select {
		case <-ticker.C:
			metrics, err := s.GetSystemMetrics()
			if err != nil {
				logrus.Errorf("获取系统指标失败: %v", err)
				continue
			}

			// 检查阈值
			s.CheckThresholds(metrics)
		case <-s.ctx.Done():
			logrus.Info("系统监控服务已停止")
			return
		}
	}
}
