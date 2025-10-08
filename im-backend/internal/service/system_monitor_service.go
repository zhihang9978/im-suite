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

// SystemMonitorService 绯荤粺鐩戞帶鏈嶅姟
type SystemMonitorService struct {
	db    *gorm.DB
	redis *redis.Client
	ctx   context.Context
}

// SystemMetrics 绯荤粺鎸囨爣
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

// DatabaseMetrics 鏁版嵁搴撴寚鏍?type DatabaseMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	Connections     int       `json:"connections"`
	MaxConnections  int       `json:"max_connections"`
	SlowQueries     int64     `json:"slow_queries"`
	QueriesPerSecond float64  `json:"queries_per_second"`
	DatabaseSize    int64     `json:"database_size"`
	TableCount      int       `json:"table_count"`
}

// RedisMetrics Redis鎸囨爣
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

// Alert 鍛婅淇℃伅
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

// NewSystemMonitorService 鍒涘缓绯荤粺鐩戞帶鏈嶅姟
func NewSystemMonitorService() *SystemMonitorService {
	return &SystemMonitorService{
		db:    config.DB,
		redis: config.Redis,
		ctx:   context.Background(),
	}
}

// GetSystemMetrics 鑾峰彇绯荤粺鎸囨爣
func (s *SystemMonitorService) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// CPU浣跨敤鐜?	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPUUsage = cpuPercent[0]
	}

	// 鍐呭瓨浣跨敤鎯呭喌
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemoryUsage = memInfo.UsedPercent
		metrics.MemoryTotal = memInfo.Total
		metrics.MemoryUsed = memInfo.Used
	}

	// 纾佺洏浣跨敤鎯呭喌
	diskInfo, err := disk.Usage("/")
	if err == nil {
		metrics.DiskUsage = diskInfo.UsedPercent
		metrics.DiskTotal = diskInfo.Total
		metrics.DiskUsed = diskInfo.Used
	}

	// 缃戠粶IO
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		metrics.NetworkIn = netIO[0].BytesRecv
		metrics.NetworkOut = netIO[0].BytesSent
	}

	// Go杩愯鏃舵寚鏍?	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.GoroutineCount = runtime.NumGoroutine()
	metrics.HeapAlloc = memStats.HeapAlloc
	metrics.HeapSys = memStats.HeapSys

	// 淇濆瓨鍒癛edis鐢ㄤ簬鍘嗗彶璁板綍
	s.saveMetricsHistory(metrics)

	return metrics, nil
}

// GetDatabaseMetrics 鑾峰彇鏁版嵁搴撴寚鏍?func (s *SystemMonitorService) GetDatabaseMetrics() (*DatabaseMetrics, error) {
	metrics := &DatabaseMetrics{
		Timestamp: time.Now(),
	}

	// 鑾峰彇鏁版嵁搴撹繛鎺ユ暟
	sqlDB, err := s.db.DB()
	if err == nil {
		stats := sqlDB.Stats()
		metrics.Connections = stats.OpenConnections
		metrics.MaxConnections = stats.MaxOpenConnections
	}

	// 鑾峰彇鏁版嵁搴撳ぇ灏?	var result struct {
		Size int64
	}
	s.db.Raw(`
		SELECT SUM(data_length + index_length) as size
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`).Scan(&result)
	metrics.DatabaseSize = result.Size

	// 鑾峰彇琛ㄦ暟閲?	var tableCount int
	s.db.Raw(`
		SELECT COUNT(*) as count
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`).Scan(&tableCount)
	metrics.TableCount = tableCount

	return metrics, nil
}

// GetRedisMetrics 鑾峰彇Redis鎸囨爣
func (s *SystemMonitorService) GetRedisMetrics() (*RedisMetrics, error) {
	metrics := &RedisMetrics{
		Timestamp: time.Now(),
	}

	// 鑾峰彇Redis淇℃伅
	info, err := s.redis.Info(s.ctx, "memory", "stats", "clients").Result()
	if err != nil {
		return nil, err
	}

	// 瑙ｆ瀽info瀛楃涓诧紙绠€鍖栫増鏈級
	// 瀹為檯搴旇瑙ｆ瀽info涓殑鍏蜂綋瀛楁
	metrics.UsedMemory = 0
	metrics.ConnectedClients = 0

	return metrics, nil
}

// saveMetricsHistory 淇濆瓨鎸囨爣鍘嗗彶
func (s *SystemMonitorService) saveMetricsHistory(metrics *SystemMetrics) {
	key := "system:metrics:history"
	data, _ := json.Marshal(metrics)
	
	s.redis.LPush(s.ctx, key, data)
	s.redis.LTrim(s.ctx, key, 0, 999) // 淇濈暀鏈€杩?000鏉¤褰?	s.redis.Expire(s.ctx, key, 7*24*time.Hour) // 淇濈暀7澶?}

// GetMetricsHistory 鑾峰彇鎸囨爣鍘嗗彶
func (s *SystemMonitorService) GetMetricsHistory(duration time.Duration) ([]SystemMetrics, error) {
	key := "system:metrics:history"
	
	// 璁＄畻闇€瑕佽幏鍙栫殑璁板綍鏁?	count := int64(duration.Minutes()) // 鍋囪姣忓垎閽熶竴鏉¤褰?	
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

// CheckThresholds 妫€鏌ラ槇鍊煎苟鐢熸垚鍛婅
func (s *SystemMonitorService) CheckThresholds(metrics *SystemMetrics) {
	// CPU浣跨敤鐜囧憡璀?	if metrics.CPUUsage > 80 {
		s.CreateAlert("cpu", "critical", "CPU浣跨敤鐜囪繃楂?, 
			fmt.Sprintf("褰撳墠CPU浣跨敤鐜? %.1f%%", metrics.CPUUsage),
			metrics.CPUUsage, 80)
	} else if metrics.CPUUsage > 60 {
		s.CreateAlert("cpu", "warning", "CPU浣跨敤鐜囪緝楂?,
			fmt.Sprintf("褰撳墠CPU浣跨敤鐜? %.1f%%", metrics.CPUUsage),
			metrics.CPUUsage, 60)
	}

	// 鍐呭瓨浣跨敤鐜囧憡璀?	if metrics.MemoryUsage > 85 {
		s.CreateAlert("memory", "critical", "鍐呭瓨浣跨敤鐜囪繃楂?,
			fmt.Sprintf("褰撳墠鍐呭瓨浣跨敤鐜? %.1f%%", metrics.MemoryUsage),
			metrics.MemoryUsage, 85)
	} else if metrics.MemoryUsage > 70 {
		s.CreateAlert("memory", "warning", "鍐呭瓨浣跨敤鐜囪緝楂?,
			fmt.Sprintf("褰撳墠鍐呭瓨浣跨敤鐜? %.1f%%", metrics.MemoryUsage),
			metrics.MemoryUsage, 70)
	}

	// 纾佺洏浣跨敤鐜囧憡璀?	if metrics.DiskUsage > 90 {
		s.CreateAlert("disk", "critical", "纾佺洏绌洪棿涓嶈冻",
			fmt.Sprintf("褰撳墠纾佺洏浣跨敤鐜? %.1f%%", metrics.DiskUsage),
			metrics.DiskUsage, 90)
	} else if metrics.DiskUsage > 80 {
		s.CreateAlert("disk", "warning", "纾佺洏绌洪棿杈冨皯",
			fmt.Sprintf("褰撳墠纾佺洏浣跨敤鐜? %.1f%%", metrics.DiskUsage),
			metrics.DiskUsage, 80)
	}
}

// CreateAlert 鍒涘缓鍛婅
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

	// 淇濆瓨鍒版暟鎹簱
	s.db.Create(alert)

	// 鍙戦€佸埌鍛婅闃熷垪
	key := "system:alerts:queue"
	data, _ := json.Marshal(alert)
	s.redis.LPush(s.ctx, key, data)

	logrus.WithFields(logrus.Fields{
		"alert_type": alertType,
		"severity":   severity,
		"title":      title,
		"value":      value,
		"threshold":  threshold,
	}).Warn("绯荤粺鍛婅")
}

// GetActiveAlerts 鑾峰彇娲昏穬鍛婅
func (s *SystemMonitorService) GetActiveAlerts() ([]Alert, error) {
	var alerts []Alert
	err := s.db.Where("resolved = ?", false).
		Order("created_at DESC").
		Limit(100).
		Find(&alerts).Error
	
	return alerts, err
}

// ResolveAlert 瑙ｅ喅鍛婅
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

// StartMonitoring 鍚姩鐩戞帶
func (s *SystemMonitorService) StartMonitoring() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	logrus.Info("绯荤粺鐩戞帶鏈嶅姟宸插惎鍔?)

	for {
		select {
		case <-ticker.C:
			metrics, err := s.GetSystemMetrics()
			if err != nil {
				logrus.Errorf("鑾峰彇绯荤粺鎸囨爣澶辫触: %v", err)
				continue
			}

			// 妫€鏌ラ槇鍊?			s.CheckThresholds(metrics)
		case <-s.ctx.Done():
			logrus.Info("绯荤粺鐩戞帶鏈嶅姟宸插仠姝?)
			return
		}
	}
}
