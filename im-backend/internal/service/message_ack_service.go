package service

import (
	"context"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/sirupsen/logrus"
)

// MessageACKService 消息确认和去重服务
type MessageACKService struct{}

// NewMessageACKService 创建消息ACK服务
func NewMessageACKService() *MessageACKService {
	return &MessageACKService{}
}

// GenerateMessageID 生成唯一的消息ID（UUID + 时间戳）
func (s *MessageACKService) GenerateMessageID(userID uint) string {
	// 使用用户ID + 时间戳 + 随机数生成唯一ID
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d-%d-%d", userID, timestamp, time.Now().Unix()%1000)
}

// CheckDuplicate 检查消息是否重复
func (s *MessageACKService) CheckDuplicate(messageID string) bool {
	if config.Redis == nil {
		return false
	}

	ctx := context.Background()
	key := "msg_dedup:" + messageID

	// 检查Redis中是否存在
	exists, err := config.Redis.Exists(ctx, key).Result()
	if err != nil {
		logrus.WithError(err).Error("检查消息去重失败")
		return false
	}

	return exists > 0
}

// MarkMessageSent 标记消息已发送（用于去重）
func (s *MessageACKService) MarkMessageSent(messageID string) error {
	if config.Redis == nil {
		return nil
	}

	ctx := context.Background()
	key := "msg_dedup:" + messageID

	// 设置24小时过期
	err := config.Redis.Set(ctx, key, "1", 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("标记消息失败: %w", err)
	}

	return nil
}

// SendACK 发送消息确认
func (s *MessageACKService) SendACK(messageID string, receiverID uint) error {
	// 在实际场景中，这里应该通过WebSocket发送ACK给发送方
	logrus.WithFields(logrus.Fields{
		"message_id":  messageID,
		"receiver_id": receiverID,
	}).Info("发送消息ACK")

	// 更新消息状态为已送达
	if config.DB != nil {
		err := config.DB.Model(&model.Message{}).
			Where("id = ?", messageID).
			Update("status", "delivered").Error
		if err != nil {
			return fmt.Errorf("更新消息状态失败: %w", err)
		}
	}

	return nil
}

// WaitForACK 等待消息确认（带超时）
func (s *MessageACKService) WaitForACK(messageID string, timeout time.Duration) error {
	if config.Redis == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 使用Redis pub/sub等待ACK
	pubsub := config.Redis.Subscribe(ctx, "ack:"+messageID)
	defer pubsub.Close()

	select {
	case <-pubsub.Channel():
		logrus.WithField("message_id", messageID).Info("收到消息ACK")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("等待ACK超时: %s", messageID)
	}
}

// PublishACK 发布消息ACK
func (s *MessageACKService) PublishACK(messageID string) error {
	if config.Redis == nil {
		return nil
	}

	ctx := context.Background()
	err := config.Redis.Publish(ctx, "ack:"+messageID, "ack").Err()
	if err != nil {
		return fmt.Errorf("发布ACK失败: %w", err)
	}

	return nil
}

// GetUnacknowledgedMessages 获取未确认的消息
func (s *MessageACKService) GetUnacknowledgedMessages(userID uint) ([]model.Message, error) {
	var messages []model.Message

	if config.DB == nil {
		return messages, nil
	}

	// 查询5分钟前发送但未确认的消息
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	err := config.DB.Where("sender_id = ? AND status = ? AND created_at < ?",
		userID, "sent", fiveMinutesAgo).
		Order("created_at ASC").
		Limit(100).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("查询未确认消息失败: %w", err)
	}

	return messages, nil
}

