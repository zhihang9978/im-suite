package service

import (
	"context"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

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

// NewStorageOptimizationService 创建存储优化服务
func NewStorageOptimizationService() *StorageOptimizationService {
	return &StorageOptimizationService{
		db:    config.DB,
		redis: config.Redis,
		ctx:   context.Background(),
	}
}

// StartCleanupProcessor 启动清理处理器
func (s *StorageOptimizationService) StartCleanupProcessor() {
	go s.cleanupExpiredData()
	logrus.Info("存储优化服务已启动")
}

// cleanupExpiredData 清理过期数据
func (s *StorageOptimizationService) cleanupExpiredData() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.performCleanup()
	}
}

// performCleanup 执行清理
func (s *StorageOptimizationService) performCleanup() {
	logrus.Info("开始清理过期数据...")

	// 1. 清理过期的自毁消息
	s.cleanupSelfDestructMessages()

	// 2. 清理旧的已删除数据
	s.cleanupDeletedRecords()

	// 3. 清理过期的会话
	s.cleanupExpiredSessions()

	// 4. 清理过期的文件块
	s.cleanupOrphanedFileChunks()

	logrus.Info("数据清理完成")
}

// cleanupSelfDestructMessages 清理自毁消息
func (s *StorageOptimizationService) cleanupSelfDestructMessages() {
	result := s.db.Where("is_self_destruct = ? AND self_destruct_time < ?", true, time.Now()).
		Delete(&model.Message{})

	if result.Error == nil && result.RowsAffected > 0 {
		logrus.Infof("清理了 %d 条自毁消息", result.RowsAffected)
	}
}

// cleanupDeletedRecords 清理已删除记录
func (s *StorageOptimizationService) cleanupDeletedRecords() {
	// 清理30天前的软删除记录
	cutoffTime := time.Now().AddDate(0, 0, -30)

	// 清理消息
	s.db.Unscoped().Where("deleted_at < ?", cutoffTime).Delete(&model.Message{})

	// 清理文件
	s.db.Unscoped().Where("deleted_at < ?", cutoffTime).Delete(&model.File{})

	logrus.Info("清理了30天前的已删除记录")
}

// cleanupExpiredSessions 清理过期会话
func (s *StorageOptimizationService) cleanupExpiredSessions() {
	cutoffTime := time.Now().AddDate(0, 0, -7)

	result := s.db.Where("updated_at < ?", cutoffTime).Delete(&model.Session{})

	if result.Error == nil && result.RowsAffected > 0 {
		logrus.Infof("清理了 %d 个过期会话", result.RowsAffected)
	}
}

// cleanupOrphanedFileChunks 清理孤儿文件块
func (s *StorageOptimizationService) cleanupOrphanedFileChunks() {
	// 清理24小时前创建但未完成的文件块
	cutoffTime := time.Now().Add(-24 * time.Hour)

	result := s.db.Where("is_uploaded = ? AND created_at < ?", false, cutoffTime).
		Delete(&model.FileChunk{})

	if result.Error == nil && result.RowsAffected > 0 {
		logrus.Infof("清理了 %d 个孤儿文件块", result.RowsAffected)
	}
}

// GetStorageStats 获取存储统计
func (s *StorageOptimizationService) GetStorageStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计消息数量
	var messageCount int64
	s.db.Model(&model.Message{}).Count(&messageCount)
	stats["total_messages"] = messageCount

	// 统计文件数量和大小
	var fileCount int64
	var totalSize int64
	s.db.Model(&model.File{}).Count(&fileCount)
	s.db.Model(&model.File{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)
	stats["total_files"] = fileCount
	stats["total_file_size"] = totalSize

	// 统计用户数
	var userCount int64
	s.db.Model(&model.User{}).Count(&userCount)
	stats["total_users"] = userCount

	// 统计聊天数
	var chatCount int64
	s.db.Model(&model.Chat{}).Count(&chatCount)
	stats["total_chats"] = chatCount

	return stats, nil
}

// CompressOldMessages 压缩旧消息
func (s *StorageOptimizationService) CompressOldMessages(beforeDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -beforeDays)

	logrus.Infof("压缩 %d 天前的消息", beforeDays)

	// 这里可以实现消息内容压缩逻辑
	// 简化实现：标记为已压缩
	result := s.db.Model(&model.Message{}).
		Where("created_at < ? AND content != ''", cutoffTime).
		Update("content", gorm.Expr("COMPRESS(content)"))

	if result.Error != nil {
		return fmt.Errorf("压缩消息失败: %w", result.Error)
	}

	logrus.Infof("成功压缩 %d 条消息", result.RowsAffected)
	return nil
}
