package service

import (
	"context"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/sirupsen/logrus"
)

// OfflineMessageService 离线消息服务
type OfflineMessageService struct{}

// NewOfflineMessageService 创建离线消息服务
func NewOfflineMessageService() *OfflineMessageService {
	return &OfflineMessageService{}
}

// StoreOfflineMessage 存储离线消息
func (s *OfflineMessageService) StoreOfflineMessage(message *model.Message) error {
	if config.Redis == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("offline_msg:%d", message.ReceiverID)

	// 将消息ID存储到Redis List
	err := config.Redis.LPush(ctx, key, message.ID).Err()
	if err != nil {
		return fmt.Errorf("存储离线消息失败: %w", err)
	}

	// 设置过期时间7天
	config.Redis.Expire(ctx, key, 7*24*time.Hour)

	logrus.WithFields(logrus.Fields{
		"message_id":  message.ID,
		"receiver_id": message.ReceiverID,
	}).Info("存储离线消息")

	return nil
}

// GetOfflineMessages 获取用户的离线消息
func (s *OfflineMessageService) GetOfflineMessages(userID uint, limit int) ([]model.Message, error) {
	if config.Redis == nil {
		// Fallback：直接从数据库查询
		return s.getOfflineMessagesFromDB(userID, limit)
	}

	ctx := context.Background()
	key := fmt.Sprintf("offline_msg:%d", userID)

	// 从Redis获取消息ID列表
	messageIDs, err := config.Redis.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		logrus.WithError(err).Error("获取离线消息ID失败")
		// Fallback到数据库
		return s.getOfflineMessagesFromDB(userID, limit)
	}

	if len(messageIDs) == 0 {
		return []model.Message{}, nil
	}

	// 从数据库查询完整消息
	var messages []model.Message
	err = config.DB.Where("id IN ?", messageIDs).
		Order("created_at DESC").
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("查询离线消息失败: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(messages),
	}).Info("获取离线消息")

	return messages, nil
}

// getOfflineMessagesFromDB 从数据库获取离线消息（Fallback）
func (s *OfflineMessageService) getOfflineMessagesFromDB(userID uint, limit int) ([]model.Message, error) {
	var messages []model.Message

	// 查询最近7天的未读消息
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	err := config.DB.Where("receiver_id = ? AND status != ? AND created_at > ?",
		userID, "read", sevenDaysAgo).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("从数据库查询离线消息失败: %w", err)
	}

	return messages, nil
}

// ClearOfflineMessages 清除用户的离线消息记录
func (s *OfflineMessageService) ClearOfflineMessages(userID uint) error {
	if config.Redis == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("offline_msg:%d", userID)

	err := config.Redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("清除离线消息失败: %w", err)
	}

	logrus.WithField("user_id", userID).Info("清除离线消息")
	return nil
}

// GetOfflineMessageCount 获取离线消息数量
func (s *OfflineMessageService) GetOfflineMessageCount(userID uint) (int64, error) {
	if config.Redis == nil {
		return s.getOfflineMessageCountFromDB(userID)
	}

	ctx := context.Background()
	key := fmt.Sprintf("offline_msg:%d", userID)

	count, err := config.Redis.LLen(ctx, key).Result()
	if err != nil {
		logrus.WithError(err).Error("获取离线消息数量失败")
		return s.getOfflineMessageCountFromDB(userID)
	}

	return count, nil
}

// getOfflineMessageCountFromDB 从数据库获取离线消息数量
func (s *OfflineMessageService) getOfflineMessageCountFromDB(userID uint) (int64, error) {
	var count int64

	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	err := config.DB.Model(&model.Message{}).
		Where("receiver_id = ? AND status != ? AND created_at > ?",
			userID, "read", sevenDaysAgo).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("从数据库统计离线消息失败: %w", err)
	}

	return count, nil
}

