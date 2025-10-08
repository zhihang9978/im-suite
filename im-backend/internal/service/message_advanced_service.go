package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// MessageAdvancedService 高级消息服务
type MessageAdvancedService struct {
	db *gorm.DB
}

// NewMessageAdvancedService 创建高级消息服务
func NewMessageAdvancedService(db *gorm.DB) *MessageAdvancedService {
	return &MessageAdvancedService{
		db: db,
	}
}

// RecallMessageRequest 撤回消息请求
type RecallMessageRequest struct {
	MessageID uint   `json:"message_id" binding:"required"`
	Reason    string `json:"reason,omitempty"`
}

// EditMessageRequest 编辑消息请求
type EditMessageRequest struct {
	MessageID uint   `json:"message_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Reason    string `json:"reason,omitempty"`
}

// ForwardMessageRequest 转发消息请求
type ForwardMessageRequest struct {
	MessageID     uint   `json:"message_id" binding:"required"`
	TargetChatID  *uint  `json:"target_chat_id,omitempty"`
	TargetUserID  *uint  `json:"target_user_id,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

// ScheduleMessageRequest 定时消息请求
type ScheduleMessageRequest struct {
	Content       string    `json:"content" binding:"required"`
	MessageType   string    `json:"message_type" binding:"required"`
	TargetChatID  *uint     `json:"target_chat_id,omitempty"`
	TargetUserID  *uint     `json:"target_user_id,omitempty"`
	ScheduledTime time.Time `json:"scheduled_time" binding:"required"`
	IsSilent      bool      `json:"is_silent"`
}

// SearchMessagesRequest 搜索消息请求
type SearchMessagesRequest struct {
	Query      string `json:"query" binding:"required"`
	ChatID     *uint  `json:"chat_id,omitempty"`
	UserID     *uint  `json:"user_id,omitempty"`
	MessageType string `json:"message_type,omitempty"`
	DateFrom   *time.Time `json:"date_from,omitempty"`
	DateTo     *time.Time `json:"date_to,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
}

// RecallMessage 撤回消息
func (s *MessageAdvancedService) RecallMessage(ctx context.Context, userID uint, req *RecallMessageRequest) error {
	var message model.Message
	
	// 查找消息
	if err := s.db.WithContext(ctx).Preload("Sender").First(&message, req.MessageID).Error; err != nil {
		return fmt.Errorf("消息不存在: %w", err)
	}

	// 检查权限：只有发送者或群管理员可以撤回
	if message.SenderID != userID {
		// 检查是否为群管理员
		if message.ChatID != nil {
			var chatMember model.ChatMember
			if err := s.db.WithContext(ctx).Where("chat_id = ? AND user_id = ? AND role IN (?)", 
				*message.ChatID, userID, []string{"admin", "owner"}).First(&chatMember).Error; err != nil {
				return fmt.Errorf("没有权限撤回此消息")
			}
		} else {
			return fmt.Errorf("没有权限撤回此消息")
		}
	}

	// 检查消息是否已被撤回
	if message.IsRecalled {
		return fmt.Errorf("消息已被撤回")
	}

	// 检查消息时间（超过24小时不能撤回）
	if time.Since(message.CreatedAt) > 24*time.Hour {
		return fmt.Errorf("消息发送超过24小时，无法撤回")
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新消息状态
	now := time.Now()
	if err := tx.Model(&message).Updates(map[string]interface{}{
		"is_recalled":   true,
		"recall_time":   &now,
		"recall_reason": req.Reason,
		"status":        "recalled",
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// 创建撤回记录
	recall := &model.MessageRecall{
		MessageID:  message.ID,
		RecallBy:   userID,
		Reason:     req.Reason,
		RecallTime: now,
	}

	if err := tx.Create(recall).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建撤回记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// EditMessage 编辑消息
func (s *MessageAdvancedService) EditMessage(ctx context.Context, userID uint, req *EditMessageRequest) error {
	var message model.Message
	
	// 查找消息
	if err := s.db.WithContext(ctx).Preload("Sender").First(&message, req.MessageID).Error; err != nil {
		return fmt.Errorf("消息不存在: %w", err)
	}

	// 检查权限：只有发送者可以编辑
	if message.SenderID != userID {
		return fmt.Errorf("没有权限编辑此消息")
	}

	// 检查消息是否已被撤回
	if message.IsRecalled {
		return fmt.Errorf("已撤回的消息不能编辑")
	}

	// 检查编辑次数限制（最多5次）
	if message.EditCount >= 5 {
		return fmt.Errorf("消息编辑次数已达上限")
	}

	// 检查编辑时间限制（超过48小时不能编辑）
	if time.Since(message.CreatedAt) > 48*time.Hour {
		return fmt.Errorf("消息发送超过48小时，无法编辑")
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存编辑历史
	editHistory := &model.MessageEdit{
		MessageID:  message.ID,
		OldContent: message.Content,
		NewContent: req.Content,
		EditTime:   time.Now(),
		EditReason: req.Reason,
	}

	if err := tx.Create(editHistory).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存编辑历史失败: %w", err)
	}

	// 更新消息内容
	if err := tx.Model(&message).Updates(map[string]interface{}{
		"content":    req.Content,
		"is_edited":  true,
		"edit_count": gorm.Expr("edit_count + 1"),
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新消息内容失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// ForwardMessage 转发消息
func (s *MessageAdvancedService) ForwardMessage(ctx context.Context, userID uint, req *ForwardMessageRequest) error {
	var originalMessage model.Message
	
	// 查找原消息
	if err := s.db.WithContext(ctx).Preload("Sender").First(&originalMessage, req.MessageID).Error; err != nil {
		return fmt.Errorf("原消息不存在: %w", err)
	}

	// 检查权限：不能转发已撤回的消息
	if originalMessage.IsRecalled {
		return fmt.Errorf("不能转发已撤回的消息")
	}

	// 验证转发目标
	if req.TargetChatID == nil && req.TargetUserID == nil {
		return fmt.Errorf("必须指定转发目标")
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建转发消息
	forwardMessage := &model.Message{
		SenderID:       userID,
		ReceiverID:     req.TargetUserID,
		ChatID:         req.TargetChatID,
		Content:        req.Comment,
		MessageType:    originalMessage.MessageType,
		Status:         "sent",
		ForwardFromID:  &originalMessage.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 如果有评论，添加到内容中
	if req.Comment != "" {
		forwardMessage.Content = req.Comment + "\n\n" + "转发的消息: " + originalMessage.Content
	} else {
		forwardMessage.Content = "转发的消息: " + originalMessage.Content
	}

	if err := tx.Create(forwardMessage).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建转发消息失败: %w", err)
	}

	// 创建转发记录
	forwardRecord := &model.MessageForward{
		OriginalMessageID: originalMessage.ID,
		ForwardMessageID:  forwardMessage.ID,
		ForwardBy:         userID,
		ForwardTime:       time.Now(),
	}

	if err := tx.Create(forwardRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建转发记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// ScheduleMessage 定时发送消息
func (s *MessageAdvancedService) ScheduleMessage(ctx context.Context, userID uint, req *ScheduleMessageRequest) error {
	// 检查定时时间
	if req.ScheduledTime.Before(time.Now()) {
		return fmt.Errorf("定时时间不能早于当前时间")
	}

	// 验证目标
	if req.TargetChatID == nil && req.TargetUserID == nil {
		return fmt.Errorf("必须指定发送目标")
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建消息（初始状态为草稿）
	message := &model.Message{
		SenderID:        userID,
		ReceiverID:      req.TargetUserID,
		ChatID:          req.TargetChatID,
		Content:         req.Content,
		MessageType:     req.MessageType,
		Status:          "scheduled",
		IsScheduled:     true,
		ScheduledTime:   &req.ScheduledTime,
		IsSilent:        req.IsSilent,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := tx.Create(message).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建定时消息失败: %w", err)
	}

	// 创建定时记录
	scheduledRecord := &model.ScheduledMessage{
		MessageID:      message.ID,
		ScheduledBy:    userID,
		ScheduledTime:  req.ScheduledTime,
		IsExecuted:     false,
		IsCancelled:    false,
	}

	if err := tx.Create(scheduledMessage).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建定时记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// SearchMessages 搜索消息
func (s *MessageAdvancedService) SearchMessages(ctx context.Context, userID uint, req *SearchMessagesRequest) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	// 构建查询条件
	query := s.db.WithContext(ctx).Model(&model.Message{})
	
	// 权限过滤：只能搜索自己参与的消息
	query = query.Where("(sender_id = ? OR receiver_id = ? OR chat_id IN (SELECT chat_id FROM chat_members WHERE user_id = ?))", 
		userID, userID, userID)

	// 搜索内容
	if req.Query != "" {
		query = query.Where("content LIKE ? OR id IN (SELECT message_id FROM message_search_indexes WHERE content LIKE ? OR keywords LIKE ?)", 
			"%"+req.Query+"%", "%"+req.Query+"%", "%"+req.Query+"%")
	}

	// 聊天过滤
	if req.ChatID != nil {
		query = query.Where("chat_id = ?", *req.ChatID)
	}

	// 用户过滤
	if req.UserID != nil {
		query = query.Where("(sender_id = ? OR receiver_id = ?)", *req.UserID, *req.UserID)
	}

	// 消息类型过滤
	if req.MessageType != "" {
		query = query.Where("message_type = ?", req.MessageType)
	}

	// 时间过滤
	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", *req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("created_at <= ?", *req.DateTo)
	}

	// 排除已撤回的消息
	query = query.Where("is_recalled = ?", false)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取搜索结果总数失败: %w", err)
	}

	// 分页
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// 获取结果
	if err := query.Preload("Sender").
		Preload("ReplyTo").
		Preload("ForwardFrom").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("搜索消息失败: %w", err)
	}

	return messages, total, nil
}

// GetMessageEditHistory 获取消息编辑历史
func (s *MessageAdvancedService) GetMessageEditHistory(ctx context.Context, messageID uint, userID uint) ([]model.MessageEdit, error) {
	var message model.Message
	
	// 检查消息是否存在
	if err := s.db.WithContext(ctx).First(&message, messageID).Error; err != nil {
		return nil, fmt.Errorf("消息不存在: %w", err)
	}

	// 检查权限：只有发送者可以查看编辑历史
	if message.SenderID != userID {
		return nil, fmt.Errorf("没有权限查看编辑历史")
	}

	var editHistory []model.MessageEdit
	if err := s.db.WithContext(ctx).Where("message_id = ?", messageID).
		Order("edit_time DESC").
		Find(&editHistory).Error; err != nil {
		return nil, fmt.Errorf("获取编辑历史失败: %w", err)
	}

	return editHistory, nil
}

// CancelScheduledMessage 取消定时消息
func (s *MessageAdvancedService) CancelScheduledMessage(ctx context.Context, messageID uint, userID uint, reason string) error {
	var scheduledMessage model.ScheduledMessage
	
	// 查找定时消息
	if err := s.db.WithContext(ctx).Preload("Message").First(&scheduledMessage, "message_id = ?", messageID).Error; err != nil {
		return fmt.Errorf("定时消息不存在: %w", err)
	}

	// 检查权限
	if scheduledMessage.ScheduledBy != userID {
		return fmt.Errorf("没有权限取消此定时消息")
	}

	// 检查是否已执行
	if scheduledMessage.IsExecuted {
		return fmt.Errorf("定时消息已执行，无法取消")
	}

	// 检查是否已取消
	if scheduledMessage.IsCancelled {
		return fmt.Errorf("定时消息已取消")
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 取消定时消息
	now := time.Now()
	if err := tx.Model(&scheduledMessage).Updates(map[string]interface{}{
		"is_cancelled":  true,
		"cancel_time":   &now,
		"cancel_reason": reason,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("取消定时消息失败: %w", err)
	}

	// 更新消息状态
	if err := tx.Model(&scheduledMessage.Message).Updates(map[string]interface{}{
		"status": "cancelled",
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// ExecuteScheduledMessages 执行定时消息（定时任务调用）
func (s *MessageAdvancedService) ExecuteScheduledMessages(ctx context.Context) error {
	var scheduledMessages []model.ScheduledMessage
	now := time.Now()

	// 查找到期的定时消息
	if err := s.db.WithContext(ctx).Preload("Message").
		Where("is_executed = ? AND is_cancelled = ? AND scheduled_time <= ?", 
			false, false, now).
		Find(&scheduledMessages).Error; err != nil {
		return fmt.Errorf("查找定时消息失败: %w", err)
	}

	for _, scheduled := range scheduledMessages {
		// 开始事务
		tx := s.db.WithContext(ctx).Begin()

		// 更新定时记录
		if err := tx.Model(&scheduled).Updates(map[string]interface{}{
			"is_executed":  true,
			"execute_time": &now,
		}).Error; err != nil {
			tx.Rollback()
			continue
		}

		// 更新消息状态
		if err := tx.Model(&scheduled.Message).Updates(map[string]interface{}{
			"status": "sent",
		}).Error; err != nil {
			tx.Rollback()
			continue
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			continue
		}
	}

	return nil
}
