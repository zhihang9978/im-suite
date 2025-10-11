package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/gorm"
)

// MessageService 消息服务
type MessageService struct {
	db *gorm.DB
}

// NewMessageService 创建消息服务实例
func NewMessageService() *MessageService {
	return &MessageService{
		db: config.DB,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChatID      uint   `json:"chat_id"`
	ReceiverID  *uint  `json:"receiver_id,omitempty"`
	Content     string `json:"content" binding:"required"`
	MessageType string `json:"message_type"` // text, image, video, audio, file
	ReplyToID   *uint  `json:"reply_to_id,omitempty"`
	IsSilent    bool   `json:"is_silent"`
	IsEncrypted bool   `json:"is_encrypted"`
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(senderID uint, req SendMessageRequest) (*model.Message, error) {
	// 验证消息类型
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// 创建消息
	message := &model.Message{
		SenderID:    senderID,
		Content:     req.Content,
		MessageType: req.MessageType,
		Status:      "sent",
		IsSilent:    req.IsSilent,
		IsEncrypted: req.IsEncrypted,
	}

	// 设置接收者或聊天ID
	if req.ReceiverID != nil {
		message.ReceiverID = req.ReceiverID
	} else {
		message.ChatID = &req.ChatID
	}

	// 设置回复消息
	if req.ReplyToID != nil {
		message.ReplyToID = req.ReplyToID
	}

	// 保存消息
	if err := s.db.Create(message).Error; err != nil {
		return nil, fmt.Errorf("创建消息失败: %w", err)
	}

	// 预加载关联数据
	s.db.Preload("Sender").Preload("Receiver").Preload("Chat").Preload("ReplyTo").First(message, message.ID)

	// 检查是否是发给机器人的消息，如果是则异步处理
	if message.ReceiverID != nil {
		go func() {
			botHandler := NewBotChatHandler()
			_ = botHandler.HandleMessage(context.Background(), message)
		}()
	}

	return message, nil
}

// GetMessages 获取消息列表
func (s *MessageService) GetMessages(userID uint, chatID *uint, receiverID *uint, limit, offset int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := s.db.Model(&model.Message{})

	// 根据不同条件查询
	if chatID != nil {
		// 群聊消息
		query = query.Where("chat_id = ?", *chatID)
	} else if receiverID != nil {
		// 私聊消息
		query = query.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, *receiverID, *receiverID, userID)
	} else {
		// 用户所有消息
		query = query.Where("sender_id = ? OR receiver_id = ?", userID, userID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计消息数失败: %w", err)
	}

	// 查询消息列表
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Sender").
		Preload("Receiver").
		Preload("Chat").
		Preload("ReplyTo").
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("查询消息失败: %w", err)
	}

	return messages, total, nil
}

// GetMessage 获取单条消息
func (s *MessageService) GetMessage(messageID, userID uint) (*model.Message, error) {
	var message model.Message

	// 查询消息并验证权限
	if err := s.db.Where("id = ?", messageID).
		Preload("Sender").
		Preload("Receiver").
		Preload("Chat").
		Preload("ReplyTo").
		First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("消息不存在")
		}
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	// 验证用户权限
	if message.SenderID != userID && (message.ReceiverID == nil || *message.ReceiverID != userID) {
		// 如果是群聊消息，检查用户是否在群里
		if message.ChatID != nil {
			var member model.ChatMember
			if err := s.db.Where("chat_id = ? AND user_id = ?", *message.ChatID, userID).First(&member).Error; err != nil {
				return nil, errors.New("无权访问此消息")
			}
		} else {
			return nil, errors.New("无权访问此消息")
		}
	}

	return &message, nil
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(messageID, userID uint) error {
	var message model.Message

	// 查找消息
	if err := s.db.Where("id = ? AND sender_id = ?", messageID, userID).First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("消息不存在或无权删除")
		}
		return fmt.Errorf("查询消息失败: %w", err)
	}

	// 软删除消息
	if err := s.db.Delete(&message).Error; err != nil {
		return fmt.Errorf("删除消息失败: %w", err)
	}

	return nil
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(messageID, userID uint) error {
	// 检查消息是否存在
	var message model.Message
	if err := s.db.First(&message, messageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("消息不存在")
		}
		return fmt.Errorf("查询消息失败: %w", err)
	}

	// 创建已读记录
	messageRead := &model.MessageRead{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}

	// 检查是否已读
	var existingRead model.MessageRead
	if err := s.db.Where("message_id = ? AND user_id = ?", messageID, userID).First(&existingRead).Error; err == nil {
		// 已经标记为已读
		return nil
	}

	// 创建已读记录
	if err := s.db.Create(messageRead).Error; err != nil {
		return fmt.Errorf("标记已读失败: %w", err)
	}

	// 更新消息状态为已读
	s.db.Model(&message).Update("status", "read")

	return nil
}

// RecallMessage 撤回消息
func (s *MessageService) RecallMessage(messageID, userID uint, reason string) error {
	var message model.Message

	// 查找消息
	if err := s.db.Where("id = ? AND sender_id = ?", messageID, userID).First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("消息不存在或无权撤回")
		}
		return fmt.Errorf("查询消息失败: %w", err)
	}

	// 检查撤回时间限制（2分钟内）
	if time.Since(message.CreatedAt) > 2*time.Minute {
		return errors.New("消息发送超过2分钟，无法撤回")
	}

	// 标记消息为已撤回
	now := time.Now()
	if err := s.db.Model(&message).Updates(map[string]interface{}{
		"is_recalled":   true,
		"recall_time":   &now,
		"recall_reason": reason,
	}).Error; err != nil {
		return fmt.Errorf("撤回消息失败: %w", err)
	}

	// 创建撤回记录
	recallRecord := &model.MessageRecall{
		MessageID:  messageID,
		RecallBy:   userID,
		Reason:     reason,
		RecallTime: now,
	}

	if err := s.db.Create(recallRecord).Error; err != nil {
		return fmt.Errorf("创建撤回记录失败: %w", err)
	}

	return nil
}

// EditMessage 编辑消息
func (s *MessageService) EditMessage(messageID, userID uint, newContent string) (*model.Message, error) {
	var message model.Message

	// 查找消息
	if err := s.db.Where("id = ? AND sender_id = ?", messageID, userID).First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("消息不存在或无权编辑")
		}
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	// 检查消息是否已撤回
	if message.IsRecalled {
		return nil, errors.New("已撤回的消息无法编辑")
	}

	// 保存编辑历史
	editRecord := &model.MessageEdit{
		MessageID:  messageID,
		OldContent: message.Content,
		NewContent: newContent,
		EditTime:   time.Now(),
	}

	if err := s.db.Create(editRecord).Error; err != nil {
		return nil, fmt.Errorf("保存编辑历史失败: %w", err)
	}

	// 更新消息内容
	message.Content = newContent
	message.IsEdited = true
	message.EditCount++

	if err := s.db.Save(&message).Error; err != nil {
		return nil, fmt.Errorf("更新消息失败: %w", err)
	}

	// 重新加载关联数据
	s.db.Preload("Sender").Preload("Receiver").Preload("Chat").First(&message, message.ID)

	return &message, nil
}

// SearchMessages 搜索消息
func (s *MessageService) SearchMessages(userID uint, keyword string, chatID *uint, limit, offset int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := s.db.Model(&model.Message{})

	// 内容搜索
	query = query.Where("content LIKE ?", "%"+keyword+"%")

	// 聊天范围限制
	if chatID != nil {
		query = query.Where("chat_id = ?", *chatID)
	} else {
		// 只搜索用户相关的消息
		query = query.Where("sender_id = ? OR receiver_id = ?", userID, userID)
	}

	// 排除已撤回的消息
	query = query.Where("is_recalled = ?", false)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计搜索结果失败: %w", err)
	}

	// 查询消息
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Sender").
		Preload("Receiver").
		Preload("Chat").
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("搜索消息失败: %w", err)
	}

	return messages, total, nil
}

// ForwardMessage 转发消息
func (s *MessageService) ForwardMessage(messageID, userID, targetChatID uint) (*model.Message, error) {
	var originalMessage model.Message

	// 查找原始消息
	if err := s.db.First(&originalMessage, messageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("原始消息不存在")
		}
		return nil, fmt.Errorf("查询原始消息失败: %w", err)
	}

	// 创建转发消息
	forwardedMessage := &model.Message{
		SenderID:      userID,
		ChatID:        &targetChatID,
		Content:       originalMessage.Content,
		MessageType:   originalMessage.MessageType,
		Status:        "sent",
		ForwardFromID: &messageID,
	}

	if err := s.db.Create(forwardedMessage).Error; err != nil {
		return nil, fmt.Errorf("创建转发消息失败: %w", err)
	}

	// 创建转发记录
	forwardRecord := &model.MessageForward{
		OriginalMessageID: messageID,
		ForwardMessageID:  forwardedMessage.ID,
		ForwardBy:         userID,
		ForwardTime:       time.Now(),
	}

	if err := s.db.Create(forwardRecord).Error; err != nil {
		return nil, fmt.Errorf("创建转发记录失败: %w", err)
	}

	// 更新原消息的分享次数
	s.db.Model(&originalMessage).UpdateColumn("share_count", gorm.Expr("share_count + 1"))

	// 重新加载关联数据
	s.db.Preload("Sender").Preload("Chat").Preload("ForwardFrom").First(forwardedMessage, forwardedMessage.ID)

	return forwardedMessage, nil
}

// GetUnreadCount 获取未读消息数
func (s *MessageService) GetUnreadCount(userID uint, chatID *uint) (int64, error) {
	var count int64

	query := s.db.Model(&model.Message{})

	if chatID != nil {
		// 特定聊天的未读数
		query = query.Where("chat_id = ?", *chatID)
	} else {
		// 所有未读消息
		query = query.Where("receiver_id = ?", userID)
	}

	// 排除已读的消息
	query = query.Where("id NOT IN (?)",
		s.db.Model(&model.MessageRead{}).Select("message_id").Where("user_id = ?", userID),
	)

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计未读消息失败: %w", err)
	}

	return count, nil
}

// BroadcastTypingStatus 广播正在输入状态（通过WebSocket）
func (s *MessageService) BroadcastTypingStatus(userID uint, chatID *uint, receiverID *uint, action string) error {
	// 构造typing事件消息
	typingEvent := map[string]interface{}{
		"type": "user_typing",
		"data": map[string]interface{}{
			"user_id": userID,
			"action":  action, // typing/uploading_photo/recording_voice/record_audio等
		},
	}

	// 添加聊天信息
	if chatID != nil {
		typingEvent["data"].(map[string]interface{})["chat_id"] = *chatID
	}
	if receiverID != nil {
		typingEvent["data"].(map[string]interface{})["receiver_id"] = *receiverID
	}

	// TODO: 通过WebSocket广播给相关用户
	// 这里需要与WebSocket服务集成，暂时打印日志
	fmt.Printf("⌨️  Typing事件: user_id=%d, chat_id=%v, receiver_id=%v, action=%s\n", 
		userID, chatID, receiverID, action)

	// 实际生产环境需要调用WebSocket管理器的广播方法
	// websocketManager.BroadcastToChat(chatID, typingEvent)
	// 或
	// websocketManager.SendToUser(receiverID, typingEvent)

	return nil
}
