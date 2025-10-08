package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SystemMonitorService 系统监控服务
type SystemMonitorService struct {
	db  *gorm.DB
	ctx context.Context
}

// NewSystemMonitorService 创建系统监控服务
func NewSystemMonitorService() *SystemMonitorService {
	return &SystemMonitorService{
		db:  config.DB,
		ctx: context.Background(),
	}
}

// StartMonitoring 启动监控
func (s *SystemMonitorService) StartMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logrus.Info("系统监控服务已启动")

	for {
		select {
		case <-ticker.C:
			s.collectMetrics()
		case <-s.ctx.Done():
			logrus.Info("系统监控服务已停止")
			return
		}
	}
}

// collectMetrics 收集系统指标
func (s *SystemMonitorService) collectMetrics() {
	// CPU使用率
	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 {
		logrus.Debugf("CPU使用率: %.2f%%", cpuPercent[0])
		s.checkCPUAlert(cpuPercent[0])
	}

	// 内存使用率
	memInfo, _ := mem.VirtualMemory()
	if memInfo != nil {
		logrus.Debugf("内存使用率: %.2f%%", memInfo.UsedPercent)
		s.checkMemoryAlert(memInfo.UsedPercent)
	}

	// 磁盘使用率
	diskInfo, _ := disk.Usage("/")
	if diskInfo != nil {
		logrus.Debugf("磁盘使用率: %.2f%%", diskInfo.UsedPercent)
		s.checkDiskAlert(diskInfo.UsedPercent)
	}

	// 网络IO统计
	netIO, _ := net.IOCounters(false)
	if len(netIO) > 0 {
		logrus.Debugf("网络发送: %d bytes, 接收: %d bytes", netIO[0].BytesSent, netIO[0].BytesRecv)
	}

	// Go runtime信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logrus.Debugf("Go内存分配: %d MB, GC次数: %d", m.Alloc/1024/1024, m.NumGC)
}

// checkCPUAlert 检查CPU告警
func (s *SystemMonitorService) checkCPUAlert(usage float64) {
	if usage > 80 {
		s.createAlert("cpu", "high", "CPU使用率过高", fmt.Sprintf("当前使用率: %.2f%%", usage))
	}
}

// checkMemoryAlert 检查内存告警
func (s *SystemMonitorService) checkMemoryAlert(usage float64) {
	if usage > 85 {
		s.createAlert("memory", "high", "内存使用率过高", fmt.Sprintf("当前使用率: %.2f%%", usage))
	}
}

// checkDiskAlert 检查磁盘告警
func (s *SystemMonitorService) checkDiskAlert(usage float64) {
	if usage > 90 {
		s.createAlert("disk", "critical", "磁盘空间不足", fmt.Sprintf("当前使用率: %.2f%%", usage))
	}
}

// createAlert 创建告警
func (s *SystemMonitorService) createAlert(alertType, level, message, details string) {
	alert := model.Alert{
		AlertType: alertType,
		Severity:  level,
		Title:     message,
		Message:   details,
	}

	if err := s.db.Create(&alert).Error; err != nil {
		logrus.Errorf("创建告警失败: %v", err)
	} else {
		logrus.Warnf("系统告警: [%s] %s - %s", level, message, details)
	}
}

// GetSystemStats 获取系统统计信息
func (s *SystemMonitorService) GetSystemStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// CPU信息
	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 {
		stats["cpu_usage"] = cpuPercent[0]
	}

	// 内存信息
	memInfo, _ := mem.VirtualMemory()
	if memInfo != nil {
		stats["memory_total"] = memInfo.Total
		stats["memory_used"] = memInfo.Used
		stats["memory_percent"] = memInfo.UsedPercent
	}

	// 磁盘信息
	diskInfo, _ := disk.Usage("/")
	if diskInfo != nil {
		stats["disk_total"] = diskInfo.Total
		stats["disk_used"] = diskInfo.Used
		stats["disk_percent"] = diskInfo.UsedPercent
	}

	// 数据库连接信息
	if config.DB != nil {
		sqlDB, err := config.DB.DB()
		if err == nil {
			dbStats := sqlDB.Stats()
			stats["db_open_connections"] = dbStats.OpenConnections
			stats["db_in_use"] = dbStats.InUse
			stats["db_idle"] = dbStats.Idle
		}
	}

	// Redis信息
	if config.Redis != nil {
		redisInfo, err := config.Redis.Info(s.ctx, "memory").Result()
		if err == nil {
			stats["redis_info"] = redisInfo
		}
	}

	return stats, nil
}

// GetActiveAlerts 获取活跃告警
func (s *SystemMonitorService) GetActiveAlerts() ([]model.Alert, error) {
	var alerts []model.Alert
	err := s.db.Where("status = ?", "active").
		Order("created_at DESC").
		Limit(100).
		Find(&alerts).Error
	return alerts, err
}
