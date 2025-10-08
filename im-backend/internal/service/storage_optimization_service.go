package service

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zhihang-messenger/im-backend/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StorageOptimizationService 存储优化服务
type StorageOptimizationService struct {
	db    *gorm.DB
	redis *redis.Client
	ctx   context.Context
}

// CompressionStats 压缩统计
type CompressionStats struct {
	TableName       string    `json:"table_name"`
	OriginalSize    int64     `json:"original_size"`
	CompressedSize  int64     `json:"compressed_size"`
	CompressionRate float64   `json:"compression_rate"`
	RecordCount     int64     `json:"record_count"`
	CompressedAt    time.Time `json:"compressed_at"`
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	TableName     string    `json:"table_name"`
	PartitionName string    `json:"partition_name"`
	Rows          int64     `json:"rows"`
	Size          int64     `json:"size"`
	CreatedAt     time.Time `json:"created_at"`
}

// CleanupTask 清理任务
type CleanupTask struct {
	ID          uint       `json:"id"`
	TableName   string     `json:"table_name"`
	Condition   string     `json:"condition"`
	DeletedRows int64      `json:"deleted_rows"`
	Status      string     `json:"status"` // pending, running, completed, failed
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// NewStorageOptimizationService 创建存储优化服务
func NewStorageOptimizationService() *StorageOptimizationService {
	return &StorageOptimizationService{
		db:    config.DB,
		redis: config.Redis,
		ctx:   context.Background(),
	}
}

// CompressTable 压缩表数据
func (s *StorageOptimizationService) CompressTable(tableName string) (*CompressionStats, error) {
	start := time.Now()

	// 获取表信息
	tableInfo, err := s.getTableInfo(tableName)
	if err != nil {
		return nil, err
	}

	originalSize := tableInfo.Size

	// 执行压缩操作
	compressedSize, err := s.performCompression(tableName)
	if err != nil {
		return nil, err
	}

	compressionRate := 0.0
	if originalSize > 0 {
		compressionRate = float64(originalSize-compressedSize) / float64(originalSize) * 100
	}

	stats := &CompressionStats{
		TableName:       tableName,
		OriginalSize:    originalSize,
		CompressedSize:  compressedSize,
		CompressionRate: compressionRate,
		RecordCount:     tableInfo.Rows,
		CompressedAt:    start,
	}

	// 记录压缩统计
	s.recordCompressionStats(stats)

	logrus.WithFields(logrus.Fields{
		"table_name":       tableName,
		"original_size":    originalSize,
		"compressed_size":  compressedSize,
		"compression_rate": compressionRate,
		"duration":         time.Since(start).Seconds(),
	}).Info("表压缩完成")

	return stats, nil
}

// getTableInfo 获取表信息
func (s *StorageOptimizationService) getTableInfo(tableName string) (*struct {
	Size int64
	Rows int64
}, error) {
	var result struct {
		Size int64
		Rows int64
	}

	// 获取表大小和行数
	query := `
		SELECT 
			ROUND(((data_length + index_length) / 1024 / 1024), 2) AS size,
			table_rows AS rows
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() AND table_name = ?
	`

	err := s.db.Raw(query, tableName).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// performCompression 执行压缩操作
func (s *StorageOptimizationService) performCompression(tableName string) (int64, error) {
	// 优化表
	err := s.db.Exec(fmt.Sprintf("OPTIMIZE TABLE %s", tableName)).Error
	if err != nil {
		return 0, err
	}

	// 重新获取压缩后的大小
	tableInfo, err := s.getTableInfo(tableName)
	if err != nil {
		return 0, err
	}

	return tableInfo.Size, nil
}

// recordCompressionStats 记录压缩统计
func (s *StorageOptimizationService) recordCompressionStats(stats *CompressionStats) {
	key := fmt.Sprintf("storage:compression:%s:%d", stats.TableName, stats.CompressedAt.Unix())
	data, _ := json.Marshal(stats)

	s.redis.SetEX(s.ctx, key, data, 7*24*time.Hour) // 保留7天
}

// CreatePartitions 创建分区
func (s *StorageOptimizationService) CreatePartitions(tableName string) error {
	switch tableName {
	case "messages":
		return s.createMessagePartitions()
	case "chat_members":
		return s.createChatMemberPartitions()
	case "message_reads":
		return s.createMessageReadPartitions()
	default:
		return fmt.Errorf("不支持的表: %s", tableName)
	}
}

// createMessagePartitions 创建消息表分区
func (s *StorageOptimizationService) createMessagePartitions() error {
	// 按月份分区
	partitions := []string{
		"PARTITION p202401 VALUES LESS THAN (UNIX_TIMESTAMP('2024-02-01'))",
		"PARTITION p202402 VALUES LESS THAN (UNIX_TIMESTAMP('2024-03-01'))",
		"PARTITION p202403 VALUES LESS THAN (UNIX_TIMESTAMP('2024-04-01'))",
		"PARTITION p202404 VALUES LESS THAN (UNIX_TIMESTAMP('2024-05-01'))",
		"PARTITION p202405 VALUES LESS THAN (UNIX_TIMESTAMP('2024-06-01'))",
		"PARTITION p202406 VALUES LESS THAN (UNIX_TIMESTAMP('2024-07-01'))",
		"PARTITION p202407 VALUES LESS THAN (UNIX_TIMESTAMP('2024-08-01'))",
		"PARTITION p202408 VALUES LESS THAN (UNIX_TIMESTAMP('2024-09-01'))",
		"PARTITION p202409 VALUES LESS THAN (UNIX_TIMESTAMP('2024-10-01'))",
		"PARTITION p202410 VALUES LESS THAN (UNIX_TIMESTAMP('2024-11-01'))",
		"PARTITION p202411 VALUES LESS THAN (UNIX_TIMESTAMP('2024-12-01'))",
		"PARTITION p202412 VALUES LESS THAN (UNIX_TIMESTAMP('2025-01-01'))",
	}

	partitionSQL := fmt.Sprintf(`
		ALTER TABLE messages PARTITION BY RANGE (UNIX_TIMESTAMP(created_at)) (
			%s
		)
	`, strings.Join(partitions, ",\n"))

	return s.db.Exec(partitionSQL).Error
}

// createChatMemberPartitions 创建聊天成员表分区
func (s *StorageOptimizationService) createChatMemberPartitions() error {
	// 按聊天ID分区（每1000个聊天一个分区）
	partitions := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		end := (i + 1) * 1000
		partitionSQL := fmt.Sprintf("PARTITION p%d VALUES LESS THAN (%d)", i, end)
		partitions = append(partitions, partitionSQL)
	}

	partitionSQL := fmt.Sprintf(`
		ALTER TABLE chat_members PARTITION BY RANGE (chat_id) (
			%s
		)
	`, strings.Join(partitions, ",\n"))

	return s.db.Exec(partitionSQL).Error
}

// createMessageReadPartitions 创建消息已读表分区
func (s *StorageOptimizationService) createMessageReadPartitions() error {
	// 按月份分区
	partitions := []string{
		"PARTITION p202401 VALUES LESS THAN (UNIX_TIMESTAMP('2024-02-01'))",
		"PARTITION p202402 VALUES LESS THAN (UNIX_TIMESTAMP('2024-03-01'))",
		"PARTITION p202403 VALUES LESS THAN (UNIX_TIMESTAMP('2024-04-01'))",
		"PARTITION p202404 VALUES LESS THAN (UNIX_TIMESTAMP('2024-05-01'))",
		"PARTITION p202405 VALUES LESS THAN (UNIX_TIMESTAMP('2024-06-01'))",
		"PARTITION p202406 VALUES LESS THAN (UNIX_TIMESTAMP('2024-07-01'))",
		"PARTITION p202407 VALUES LESS THAN (UNIX_TIMESTAMP('2024-08-01'))",
		"PARTITION p202408 VALUES LESS THAN (UNIX_TIMESTAMP('2024-09-01'))",
		"PARTITION p202409 VALUES LESS THAN (UNIX_TIMESTAMP('2024-10-01'))",
		"PARTITION p202410 VALUES LESS THAN (UNIX_TIMESTAMP('2024-11-01'))",
		"PARTITION p202411 VALUES LESS THAN (UNIX_TIMESTAMP('2024-12-01'))",
		"PARTITION p202412 VALUES LESS THAN (UNIX_TIMESTAMP('2025-01-01'))",
	}

	partitionSQL := fmt.Sprintf(`
		ALTER TABLE message_reads PARTITION BY RANGE (UNIX_TIMESTAMP(read_at)) (
			%s
		)
	`, strings.Join(partitions, ",\n"))

	return s.db.Exec(partitionSQL).Error
}

// GetPartitionInfo 获取分区信息
func (s *StorageOptimizationService) GetPartitionInfo(tableName string) ([]PartitionInfo, error) {
	query := `
		SELECT 
			partition_name,
			table_rows as rows,
			ROUND(((data_length + index_length) / 1024 / 1024), 2) as size,
			create_time as created_at
		FROM information_schema.partitions 
		WHERE table_schema = DATABASE() 
		AND table_name = ? 
		AND partition_name IS NOT NULL
		ORDER BY partition_name
	`

	var partitions []PartitionInfo
	err := s.db.Raw(query, tableName).Scan(&partitions).Error
	if err != nil {
		return nil, err
	}

	// 设置表名
	for i := range partitions {
		partitions[i].TableName = tableName
	}

	return partitions, nil
}

// ScheduleCleanup 调度清理任务
func (s *StorageOptimizationService) ScheduleCleanup(tableName, condition string) (*CleanupTask, error) {
	task := &CleanupTask{
		TableName: tableName,
		Condition: condition,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// 保存到数据库
	err := s.db.Create(task).Error
	if err != nil {
		return nil, err
	}

	// 添加到清理队列
	s.addToCleanupQueue(task)

	return task, nil
}

// addToCleanupQueue 添加到清理队列
func (s *StorageOptimizationService) addToCleanupQueue(task *CleanupTask) {
	key := "storage:cleanup:queue"
	data, _ := json.Marshal(task)

	s.redis.LPush(s.ctx, key, data)
}

// ProcessCleanupQueue 处理清理队列
func (s *StorageOptimizationService) ProcessCleanupQueue() error {
	key := "storage:cleanup:queue"

	for {
		result, err := s.redis.BRPop(s.ctx, 5*time.Second, key).Result()
		if err != nil {
			if err.Error() == "redis: nil" {
				continue // 队列为空，继续等待
			}
			return err
		}

		if len(result) < 2 {
			continue
		}

		var task CleanupTask
		err = json.Unmarshal([]byte(result[1]), &task)
		if err != nil {
			logrus.Errorf("解析清理任务失败: %v", err)
			continue
		}

		s.executeCleanupTask(&task)
	}
}

// executeCleanupTask 执行清理任务
func (s *StorageOptimizationService) executeCleanupTask(task *CleanupTask) {
	// 更新任务状态为运行中
	s.updateTaskStatus(task.ID, "running", "")

	start := time.Now()

	// 执行清理
	deletedRows, err := s.performCleanup(task.TableName, task.Condition)
	if err != nil {
		s.updateTaskStatus(task.ID, "failed", err.Error())
		logrus.Errorf("清理任务失败: %v", err)
		return
	}

	// 更新任务状态为完成
	task.DeletedRows = deletedRows
	now := time.Now()
	task.CompletedAt = &now

	s.updateTaskStatus(task.ID, "completed", "")

	logrus.WithFields(logrus.Fields{
		"task_id":      task.ID,
		"table_name":   task.TableName,
		"deleted_rows": deletedRows,
		"duration":     time.Since(start).Seconds(),
	}).Info("清理任务完成")
}

// updateTaskStatus 更新任务状态
func (s *StorageOptimizationService) updateTaskStatus(taskID uint, status, errorMsg string) {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "completed" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	if errorMsg != "" {
		updates["error"] = errorMsg
	}

	s.db.Model(&CleanupTask{}).Where("id = ?", taskID).Updates(updates)
}

// performCleanup 执行清理操作
func (s *StorageOptimizationService) performCleanup(tableName, condition string) (int64, error) {
	// 构建删除SQL
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, condition)

	result := s.db.Exec(sql)
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// CleanupOldMessages 清理旧消息
func (s *StorageOptimizationService) CleanupOldMessages(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	task, err := s.ScheduleCleanup("messages", fmt.Sprintf("created_at < '%s'", cutoffDate.Format("2006-01-02")))
	if err != nil {
		return err
	}

	logrus.Infof("已调度清理 %d 天前的消息，任务ID: %d", days, task.ID)

	return nil
}

// CleanupInactiveSessions 清理不活跃会话
func (s *StorageOptimizationService) CleanupInactiveSessions(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	task, err := s.ScheduleCleanup("sessions", fmt.Sprintf("last_activity < '%s'", cutoffDate.Format("2006-01-02")))
	if err != nil {
		return err
	}

	logrus.Infof("已调度清理 %d 天前的不活跃会话，任务ID: %d", days, task.ID)

	return nil
}

// CleanupOrphanedFiles 清理孤立文件
func (s *StorageOptimizationService) CleanupOrphanedFiles() error {
	// 查找没有关联消息的文件
	condition := `
		id NOT IN (
			SELECT DISTINCT file_id FROM messages 
			WHERE file_id IS NOT NULL AND deleted_at IS NULL
		) AND created_at < DATE_SUB(NOW(), INTERVAL 7 DAY)
	`

	task, err := s.ScheduleCleanup("files", condition)
	if err != nil {
		return err
	}

	logrus.Infof("已调度清理孤立文件，任务ID: %d", task.ID)

	return nil
}

// GetStorageStats 获取存储统计
func (s *StorageOptimizationService) GetStorageStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取所有表的大小
	query := `
		SELECT 
			table_name,
			ROUND(((data_length + index_length) / 1024 / 1024), 2) AS size_mb,
			table_rows
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
		ORDER BY (data_length + index_length) DESC
	`

	var tableStats []struct {
		TableName string  `json:"table_name"`
		SizeMB    float64 `json:"size_mb"`
		Rows      int64   `json:"rows"`
	}

	err := s.db.Raw(query).Scan(&tableStats).Error
	if err != nil {
		return nil, err
	}

	stats["tables"] = tableStats

	// 计算总大小
	totalSize := 0.0
	for _, table := range tableStats {
		totalSize += table.SizeMB
	}
	stats["total_size_mb"] = totalSize

	// 获取最近的压缩统计
	recentCompressions, err := s.getRecentCompressions()
	if err == nil {
		stats["recent_compressions"] = recentCompressions
	}

	return stats, nil
}

// getRecentCompressions 获取最近的压缩统计
func (s *StorageOptimizationService) getRecentCompressions() ([]CompressionStats, error) {
	pattern := "storage:compression:*"
	keys, err := s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var compressions []CompressionStats
	for _, key := range keys {
		data, err := s.redis.Get(s.ctx, key).Result()
		if err != nil {
			continue
		}

		var stats CompressionStats
		err = json.Unmarshal([]byte(data), &stats)
		if err != nil {
			continue
		}

		compressions = append(compressions, stats)
	}

	return compressions, nil
}

// StartCleanupProcessor 启动清理处理器
func (s *StorageOptimizationService) StartCleanupProcessor() {
	go func() {
		logrus.Info("启动存储清理处理器")
		for {
			err := s.ProcessCleanupQueue()
			if err != nil {
				logrus.Errorf("处理清理队列失败: %v", err)
				time.Sleep(10 * time.Second)
			}
		}
	}()
}

// 实现driver.Valuer接口
func (c *CompressionStats) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// 实现sql.Scanner接口
func (c *CompressionStats) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("无法扫描到CompressionStats")
	}

	return json.Unmarshal(bytes, c)
}
