package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

type MessageService struct {
	db *gorm.DB
}

func NewMessageService() *MessageService {
	return &MessageService{
		db: config.DB,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChatID       uint   `json:"chat_id" binding:"required"`
	Type         string `json:"type" binding:"required"`
	Content      string `json:"content"`
	MediaURL     string `json:"media_url"`
	FileSize     int64  `json:"file_size"`
	Duration     int    `json:"duration"`
	ReplyToID    *uint  `json:"reply_to_id"`
	ForwardFromID *uint `json:"forward_from_id"`
	TTL          int    `json:"ttl"` // 阅后即焚时间
	SendAt       *time.Time `json:"send_at"` // 定时发送
}

// GetMessagesRequest 获取消息请求
type GetMessagesRequest struct {
	ChatID uint `form:"chat_id" binding:"required"`
	Limit  int  `form:"limit"`
	Offset int  `form:"offset"`
	Before *uint `form:"before"`
	After  *uint `form:"after"`
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(userID uint, req SendMessageRequest) (*model.Message, error) {
	// 检查用户是否在聊天中
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ?", req.ChatID, userID).First(&member).Error; err != nil {
		return nil, errors.New("用户不在该聊天中")
	}

	// 创建消息
	chatID := req.ChatID
	message := model.Message{
		ChatID:        &chatID,
		SenderID:      userID,
		MessageType:   req.Type,
		Content:       req.Content,
		ReplyToID:     req.ReplyToID,
		ForwardFromID: req.ForwardFromID,
	}

	if err := s.db.Create(&message).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据
	s.db.Preload("User").Preload("Chat").Preload("ReplyTo").First(&message, message.ID)

	// 更新聊天的最后消息
	s.db.Model(&model.Chat{}).Where("id = ?", req.ChatID).Updates(map[string]interface{}{
		"last_message_id": message.ID,
		"last_message_at": time.Now(),
	})

	return &message, nil
}

// GetMessages 获取消息列表
func (s *MessageService) GetMessages(userID uint, req GetMessagesRequest) ([]model.Message, error) {
	// 检查用户是否在聊天中
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ?", req.ChatID, userID).First(&member).Error; err != nil {
		return nil, errors.New("用户不在该聊天中")
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	var messages []model.Message
	query := s.db.Where("chat_id = ? AND is_deleted = ?", req.ChatID, false).
		Preload("User").
		Preload("ReplyTo").
		Preload("ForwardFrom").
		Order("created_at DESC")

	// 添加分页条件
	if req.Before != nil {
		query = query.Where("id < ?", *req.Before)
	}
	if req.After != nil {
		query = query.Where("id > ?", *req.After)
	}

	if err := query.Limit(req.Limit).Offset(req.Offset).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// EditMessage 编辑消息
func (s *MessageService) EditMessage(userID uint, messageID uint, content string) (*model.Message, error) {
	var message model.Message
	
	// 查找消息
	if err := s.db.Where("id = ? AND user_id = ?", messageID, userID).First(&message).Error; err != nil {
		return nil, errors.New("消息不存在或无权限")
	}

	// 检查消息是否已删除
	if message.DeletedAt.Valid {
		return nil, errors.New("消息已删除")
	}
	
	// 标记消息为已删除（软删除）
	if err := s.db.Delete(&message).Error; err != nil {
		return err
	}

	// 更新消息
	message.Content = content
	message.IsEdited = true
	if err := s.db.Save(&message).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据
	s.db.Preload("User").Preload("Chat").First(&message, message.ID)

	return &message, nil
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(userID uint, messageID uint) error {
	var message model.Message
	
	// 查找消息
	if err := s.db.Where("id = ? AND user_id = ?", messageID, userID).First(&message).Error; err != nil {
		return errors.New("消息不存在或无权限")
	}

	// 软删除消息
	message.IsDeleted = true
	if err := s.db.Save(&message).Error; err != nil {
		return err
	}

	return nil
}

// PinMessage 置顶消息
func (s *MessageService) PinMessage(userID uint, messageID uint) error {
	// 检查用户权限（这里简化处理）
	var message model.Message
	if err := s.db.Where("id = ?", messageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}

	// 检查用户是否在聊天中
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ?", message.ChatID, userID).First(&member).Error; err != nil {
		return errors.New("用户不在该聊天中")
	}

	// 更新置顶状态
	message.IsPinned = !message.IsPinned
	if err := s.db.Save(&message).Error; err != nil {
		return err
	}

	return nil
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(userID uint, messageID uint) error {
	// 检查消息是否存在
	var message model.Message
	if err := s.db.Where("id = ?", messageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}

	// 检查用户是否在聊天中
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ?", message.ChatID, userID).First(&member).Error; err != nil {
		return errors.New("用户不在该聊天中")
	}

	// 创建或更新已读记录
	var readRecord model.MessageRead
	if err := s.db.Where("message_id = ? AND user_id = ?", messageID, userID).First(&readRecord).Error; err != nil {
		// 创建新的已读记录
		readRecord = model.MessageRead{
			MessageID: messageID,
			UserID:    userID,
			ReadAt:    time.Now(),
		}
		if err := s.db.Create(&readRecord).Error; err != nil {
			return err
		}
	} else {
		// 更新已读时间
		readRecord.ReadAt = time.Now()
		if err := s.db.Save(&readRecord).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetUnreadCount 获取未读消息数量
func (s *MessageService) GetUnreadCount(userID uint, chatID uint) (int64, error) {
	var count int64
	
	// 计算未读消息数量
	err := s.db.Model(&model.Message{}).
		Joins("LEFT JOIN message_reads ON messages.id = message_reads.message_id AND message_reads.user_id = ?", userID).
		Where("messages.chat_id = ? AND messages.user_id != ? AND message_reads.id IS NULL AND messages.is_deleted = ?", 
			chatID, userID, false).
		Count(&count).Error

	return count, err
}
