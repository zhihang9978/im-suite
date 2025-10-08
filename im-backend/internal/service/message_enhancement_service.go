package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"zhihang-messenger/im-backend/internal/model"
	
	"gorm.io/gorm"
)

// MessageEnhancementService 消息功能增强服务
type MessageEnhancementService struct {
	db *gorm.DB
}

// NewMessageEnhancementService 创建消息功能增强服务实例
func NewMessageEnhancementService(db *gorm.DB) *MessageEnhancementService {
	return &MessageEnhancementService{
		db: db,
	}
}

// PinMessageRequest 置顶消息请求
type PinMessageRequest struct {
	MessageID uint   `json:"message_id" binding:"required"`
	UserID    uint   `json:"user_id" binding:"required"`
	Reason    string `json:"reason"`
}

// MarkMessageRequest 标记消息请求
type MarkMessageRequest struct {
	MessageID uint   `json:"message_id" binding:"required"`
	UserID    uint   `json:"user_id" binding:"required"`
	MarkType  string `json:"mark_type" binding:"required"` // important, favorite, archive
}

// ReplyMessageRequest 回复消息请求
type ReplyMessageRequest struct {
	MessageID    uint   `json:"message_id" binding:"required"`
	ReplyToID    uint   `json:"reply_to_id" binding:"required"`
	UserID       uint   `json:"user_id" binding:"required"`
	Content      string `json:"content" binding:"required"`
	MessageType  string `json:"message_type" default:"text"`
}

// ShareMessageRequest 分享消息请求
type ShareMessageRequest struct {
	MessageID      uint   `json:"message_id" binding:"required"`
	UserID         uint   `json:"user_id" binding:"required"`
	SharedTo       *uint  `json:"shared_to"` // 分享给用户
	SharedToChatID *uint  `json:"shared_to_chat_id"` // 分享到群聊
	ShareType      string `json:"share_type" default:"copy"` // copy, forward, link
	ShareData      string `json:"share_data"` // 额外数据
}

// UpdateMessageStatusRequest 更新消息状态请求
type UpdateMessageStatusRequest struct {
	MessageID  uint   `json:"message_id" binding:"required"`
	UserID     uint   `json:"user_id" binding:"required"`
	Status     string `json:"status" binding:"required"` // sent, delivered, read
	DeviceID   string `json:"device_id"`
	IPAddress  string `json:"ip_address"`
}

// PinMessage 置顶消息
func (s *MessageEnhancementService) PinMessage(req PinMessageRequest) error {
	// 检查消息是否存在
	var message model.Message
	if err := s.db.Where("id = ?", req.MessageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}

	// 检查是否已经置顶
	var existingPin model.MessagePin
	if err := s.db.Where("message_id = ? AND is_active = ?", req.MessageID, true).First(&existingPin).Error; err == nil {
		return errors.New("消息已经置顶")
	}

	// 更新消息状态
	now := time.Now()
	if err := s.db.Model(&message).Updates(map[string]interface{}{
		"is_pinned": true,
		"pin_time":  now,
	}).Error; err != nil {
		return fmt.Errorf("更新消息状态失败: %v", err)
	}

	// 创建置顶记录
	pin := model.MessagePin{
		MessageID: req.MessageID,
		PinnedBy:  req.UserID,
		PinTime:   now,
		IsActive:  true,
	}

	if err := s.db.Create(&pin).Error; err != nil {
		return fmt.Errorf("创建置顶记录失败: %v", err)
	}

	return nil
}

// UnpinMessage 取消置顶消息
func (s *MessageEnhancementService) UnpinMessage(messageID, userID uint) error {
	// 检查置顶记录是否存在
	var pin model.MessagePin
	if err := s.db.Where("message_id = ? AND is_active = ?", messageID, true).First(&pin).Error; err != nil {
		return errors.New("消息未置顶")
	}

	// 更新置顶记录
	now := time.Now()
	if err := s.db.Model(&pin).Updates(map[string]interface{}{
		"is_active":   false,
		"unpin_time":  now,
		"unpin_by":    userID,
	}).Error; err != nil {
		return fmt.Errorf("取消置顶失败: %v", err)
	}

	// 更新消息状态
	if err := s.db.Model(&model.Message{}).Where("id = ?", messageID).Updates(map[string]interface{}{
		"is_pinned": false,
		"pin_time":  nil,
	}).Error; err != nil {
		return fmt.Errorf("更新消息状态失败: %v", err)
	}

	return nil
}

// MarkMessage 标记消息
func (s *MessageEnhancementService) MarkMessage(req MarkMessageRequest) error {
	// 验证标记类型
	validMarkTypes := []string{"important", "favorite", "archive"}
	if !contains(validMarkTypes, req.MarkType) {
		return errors.New("无效的标记类型")
	}

	// 检查消息是否存在
	var message model.Message
	if err := s.db.Where("id = ?", req.MessageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}

	// 检查是否已经标记
	var existingMark model.MessageMark
	if err := s.db.Where("message_id = ? AND mark_type = ? AND is_active = ?", 
		req.MessageID, req.MarkType, true).First(&existingMark).Error; err == nil {
		return errors.New("消息已经标记")
	}

	// 更新消息状态
	now := time.Now()
	if err := s.db.Model(&message).Updates(map[string]interface{}{
		"is_marked": true,
		"mark_type": req.MarkType,
		"mark_time": now,
	}).Error; err != nil {
		return fmt.Errorf("更新消息状态失败: %v", err)
	}

	// 创建标记记录
	mark := model.MessageMark{
		MessageID: req.MessageID,
		MarkedBy:  req.UserID,
		MarkType:  req.MarkType,
		MarkTime:  now,
		IsActive:  true,
	}

	if err := s.db.Create(&mark).Error; err != nil {
		return fmt.Errorf("创建标记记录失败: %v", err)
	}

	return nil
}

// UnmarkMessage 取消标记消息
func (s *MessageEnhancementService) UnmarkMessage(messageID, userID uint, markType string) error {
	// 检查标记记录是否存在
	var mark model.MessageMark
	if err := s.db.Where("message_id = ? AND mark_type = ? AND is_active = ?", 
		messageID, markType, true).First(&mark).Error; err != nil {
		return errors.New("消息未标记")
	}

	// 更新标记记录
	now := time.Now()
	if err := s.db.Model(&mark).Updates(map[string]interface{}{
		"is_active":   false,
		"unmark_time": now,
		"unmark_by":   userID,
	}).Error; err != nil {
		return fmt.Errorf("取消标记失败: %v", err)
	}

	// 检查是否还有其他活跃标记
	var activeMarks int64
	s.db.Model(&model.MessageMark{}).Where("message_id = ? AND is_active = ?", messageID, true).Count(&activeMarks)

	// 如果没有其他活跃标记，更新消息状态
	if activeMarks == 0 {
		if err := s.db.Model(&model.Message{}).Where("id = ?", messageID).Updates(map[string]interface{}{
			"is_marked": false,
			"mark_type": "",
			"mark_time": nil,
		}).Error; err != nil {
			return fmt.Errorf("更新消息状态失败: %v", err)
		}
	}

	return nil
}

// ReplyToMessage 回复消息
func (s *MessageEnhancementService) ReplyToMessage(req ReplyMessageRequest) (*model.Message, error) {
	// 检查被回复的消息是否存在
	var replyToMessage model.Message
	if err := s.db.Where("id = ?", req.ReplyToID).First(&replyToMessage).Error; err != nil {
		return nil, errors.New("被回复的消息不存在")
	}

	// 计算回复层级
	replyLevel := 1
	replyPath := strconv.Itoa(int(req.ReplyToID))

	// 如果被回复的消息也是回复消息，需要计算层级
	if replyToMessage.ReplyToID != nil {
		var replyChain model.MessageReply
		if err := s.db.Where("message_id = ?", req.ReplyToID).First(&replyChain).Error; err == nil {
			replyLevel = replyChain.ReplyLevel + 1
			replyPath = replyChain.ReplyPath + "," + strconv.Itoa(int(req.ReplyToID))
		}
	}

	// 创建新消息
	message := model.Message{
		SenderID:    req.UserID,
		ReceiverID:  replyToMessage.ReceiverID,
		ChatID:      replyToMessage.ChatID,
		Content:     req.Content,
		MessageType: req.MessageType,
		Status:      "sent",
		ReplyToID:   &req.ReplyToID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(&message).Error; err != nil {
		return nil, fmt.Errorf("创建消息失败: %v", err)
	}

	// 创建回复链记录
	reply := model.MessageReply{
		MessageID:  message.ID,
		ReplyToID:  req.ReplyToID,
		ReplyLevel: replyLevel,
		ReplyPath:  replyPath,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(&reply).Error; err != nil {
		return nil, fmt.Errorf("创建回复链记录失败: %v", err)
	}

	// 预加载关联数据
	s.db.Preload("Sender").Preload("ReplyTo").First(&message, message.ID)

	return &message, nil
}

// ShareMessage 分享消息
func (s *MessageEnhancementService) ShareMessage(req ShareMessageRequest) error {
	// 检查消息是否存在
	var message model.Message
	if err := s.db.Where("id = ?", req.MessageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}

	// 验证分享类型
	validShareTypes := []string{"copy", "forward", "link"}
	if !contains(validShareTypes, req.ShareType) {
		return errors.New("无效的分享类型")
	}

	// 创建分享记录
	share := model.MessageShare{
		MessageID:      req.MessageID,
		SharedBy:       req.UserID,
		SharedTo:       req.SharedTo,
		SharedToChatID: req.SharedToChatID,
		ShareType:      req.ShareType,
		ShareTime:      time.Now(),
		ShareData:      req.ShareData,
	}

	if err := s.db.Create(&share).Error; err != nil {
		return fmt.Errorf("创建分享记录失败: %v", err)
	}

	// 更新消息分享次数
	s.db.Model(&message).Update("share_count", gorm.Expr("share_count + 1"))

	return nil
}

// UpdateMessageStatus 更新消息状态
func (s *MessageEnhancementService) UpdateMessageStatus(req UpdateMessageStatusRequest) error {
	// 检查消息是否存在
	var message model.Message
	if err := s.db.Where("id = ?", req.MessageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}

	// 验证状态
	validStatuses := []string{"sent", "delivered", "read"}
	if !contains(validStatuses, req.Status) {
		return errors.New("无效的消息状态")
	}

	// 检查是否已存在该状态记录
	var existingStatus model.MessageStatus
	if err := s.db.Where("message_id = ? AND user_id = ? AND status = ?", 
		req.MessageID, req.UserID, req.Status).First(&existingStatus).Error; err == nil {
		return errors.New("状态已存在")
	}

	// 创建状态记录
	status := model.MessageStatus{
		MessageID:  req.MessageID,
		UserID:     req.UserID,
		Status:     req.Status,
		StatusTime: time.Now(),
		DeviceID:   req.DeviceID,
		IPAddress:  req.IPAddress,
	}

	if err := s.db.Create(&status).Error; err != nil {
		return fmt.Errorf("创建状态记录失败: %v", err)
	}

	// 更新消息状态
	s.db.Model(&message).Update("status", req.Status)

	return nil
}

// GetMessageReplyChain 获取消息回复链
func (s *MessageEnhancementService) GetMessageReplyChain(messageID uint) ([]model.Message, error) {
	// 获取回复链路径
	var reply model.MessageReply
	if err := s.db.Where("message_id = ?", messageID).First(&reply).Error; err != nil {
		return nil, errors.New("消息不是回复消息")
	}

	// 解析回复路径
	pathIDs := strings.Split(reply.ReplyPath, ",")
	var messageIDs []uint
	for _, idStr := range pathIDs {
		if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			messageIDs = append(messageIDs, uint(id))
		}
	}

	// 获取所有消息
	var messages []model.Message
	if err := s.db.Preload("Sender").Where("id IN ?", messageIDs).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("获取回复链失败: %v", err)
	}

	return messages, nil
}

// GetPinnedMessages 获取置顶消息列表
func (s *MessageEnhancementService) GetPinnedMessages(chatID uint, limit, offset int) ([]model.Message, error) {
	var messages []model.Message
	
	query := s.db.Preload("Sender").Preload("ReplyTo").
		Where("chat_id = ? AND is_pinned = ?", chatID, true).
		Order("pin_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("获取置顶消息失败: %v", err)
	}

	return messages, nil
}

// GetMarkedMessages 获取标记消息列表
func (s *MessageEnhancementService) GetMarkedMessages(userID uint, markType string, limit, offset int) ([]model.Message, error) {
	var messages []model.Message
	
	query := s.db.Preload("Sender").Preload("ReplyTo").
		Joins("JOIN message_marks ON messages.id = message_marks.message_id").
		Where("message_marks.marked_by = ? AND message_marks.mark_type = ? AND message_marks.is_active = ?", 
			userID, markType, true).
		Order("message_marks.mark_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("获取标记消息失败: %v", err)
	}

	return messages, nil
}

// GetMessageStatus 获取消息状态
func (s *MessageEnhancementService) GetMessageStatus(messageID uint) ([]model.MessageStatus, error) {
	var statuses []model.MessageStatus
	
	if err := s.db.Preload("User").Where("message_id = ?", messageID).
		Order("status_time ASC").Find(&statuses).Error; err != nil {
		return nil, fmt.Errorf("获取消息状态失败: %v", err)
	}

	return statuses, nil
}

// GetMessageShareHistory 获取消息分享历史
func (s *MessageEnhancementService) GetMessageShareHistory(messageID uint, limit, offset int) ([]model.MessageShare, error) {
	var shares []model.MessageShare
	
	query := s.db.Preload("ShareUser").Preload("SharedToUser").Preload("SharedToChat").
		Where("message_id = ?", messageID).
		Order("share_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("获取分享历史失败: %v", err)
	}

	return shares, nil
}

// 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
