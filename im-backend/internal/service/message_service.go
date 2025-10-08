package service

import (
	"errors"
	"fmt"
	"time"
	"zhihang-messenger/im-backend/internal/model"
	"zhihang-messenger/im-backend/config"
	
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
	ChatID      uint   `json:"chat_id" binding:"required"`
	Content     string `json:"content"`
	Type        string `json:"type" binding:"required"`
	FileName    string `json:"file_name,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
	FileURL     string `json:"file_url,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	ReplyToID   *uint  `json:"reply_to_id,omitempty"`
	ForwardFrom *uint  `json:"forward_from,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	SendAt      *time.Time `json:"send_at,omitempty"`
	IsSilent    bool   `json:"is_silent,omitempty"`
}

// GetMessagesRequest 获取消息请求
type GetMessagesRequest struct {
	ChatID  uint `form:"chat_id" binding:"required"`
	Limit   int  `form:"limit,default=50"`
	Offset  int  `form:"offset,default=0"`
	Before  *uint `form:"before"`
	After   *uint `form:"after"`
	Search  string `form:"search"`
	Type    string `form:"type"`
}

// GetMessagesResponse 获取消息响应
type GetMessagesResponse struct {
	Messages []model.Message `json:"messages"`
	Total    int64           `json:"total"`
	HasMore  bool            `json:"has_more"`
}

// CreateChatRequest 创建聊天请求
type CreateChatRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	Members     []uint `json:"members"`
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(senderID uint, req SendMessageRequest) (*model.Message, error) {
	// 验证聊天是否存在
	var chat model.Chat
	if err := s.db.Where("id = ?", req.ChatID).First(&chat).Error; err != nil {
		return nil, errors.New("聊天不存在")
	}
	
	// 验证用户是否为聊天成员
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ?", req.ChatID, senderID).First(&member).Error; err != nil {
		return nil, errors.New("您不是该聊天的成员")
	}
	
	// 检查定时发送
	if req.SendAt != nil && req.SendAt.After(time.Now()) {
		// 定时发送逻辑 (简化处理，实际应该使用任务队列)
		req.SendAt = nil
	}
	
	// 创建消息
	message := model.Message{
		ChatID:      req.ChatID,
		SenderID:    senderID,
		Content:     req.Content,
		Type:        req.Type,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileURL:     req.FileURL,
		Thumbnail:   req.Thumbnail,
		ReplyToID:   req.ReplyToID,
		ForwardFrom: req.ForwardFrom,
		TTL:         req.TTL,
		SendAt:      req.SendAt,
		IsSilent:    req.IsSilent,
		IsRead:      false,
		IsEdited:    false,
		IsDeleted:   false,
		IsPinned:    false,
	}
	
	if err := s.db.Create(&message).Error; err != nil {
		return nil, fmt.Errorf("创建消息失败: %v", err)
	}
	
	// 预加载关联数据
	s.db.Preload("Sender").Preload("Chat").Preload("ReplyTo").First(&message, message.ID)
	
	// 更新聊天最后更新时间
	chat.UpdatedAt = time.Now()
	s.db.Save(&chat)
	
	return &message, nil
}

// GetMessages 获取消息列表
func (s *MessageService) GetMessages(userID uint, req GetMessagesRequest) (*GetMessagesResponse, error) {
	// 验证用户是否为聊天成员
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ?", req.ChatID, userID).First(&member).Error; err != nil {
		return nil, errors.New("您不是该聊天的成员")
	}
	
	// 构建查询
	query := s.db.Model(&model.Message{}).Where("chat_id = ? AND is_deleted = ?", req.ChatID, false)
	
	// 添加过滤条件
	if req.Before != nil {
		query = query.Where("id < ?", *req.Before)
	}
	if req.After != nil {
		query = query.Where("id > ?", *req.After)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Search != "" {
		query = query.Where("content LIKE ?", "%"+req.Search+"%")
	}
	
	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取消息总数失败: %v", err)
	}
	
	// 获取消息列表
	var messages []model.Message
	query = query.Preload("Sender").Preload("ReplyTo").Preload("ForwardFromMsg")
	query = query.Order("created_at DESC").Limit(req.Limit).Offset(req.Offset)
	
	if err := query.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("获取消息列表失败: %v", err)
	}
	
	// 判断是否还有更多消息
	hasMore := int64(req.Offset+req.Limit) < total
	
	return &GetMessagesResponse{
		Messages: messages,
		Total:    total,
		HasMore:  hasMore,
	}, nil
}

// EditMessage 编辑消息
func (s *MessageService) EditMessage(messageID uint, userID uint, content string) (*model.Message, error) {
	var message model.Message
	if err := s.db.Where("id = ?", messageID).First(&message).Error; err != nil {
		return nil, errors.New("消息不存在")
	}
	
	// 验证权限
	if message.SenderID != userID {
		return nil, errors.New("您只能编辑自己的消息")
	}
	
	// 更新消息
	message.Content = content
	message.IsEdited = true
	message.UpdatedAt = time.Now()
	
	if err := s.db.Save(&message).Error; err != nil {
		return nil, fmt.Errorf("编辑消息失败: %v", err)
	}
	
	// 预加载关联数据
	s.db.Preload("Sender").Preload("Chat").First(&message, message.ID)
	
	return &message, nil
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(messageID uint, userID uint) error {
	var message model.Message
	if err := s.db.Where("id = ?", messageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}
	
	// 验证权限
	if message.SenderID != userID {
		return errors.New("您只能删除自己的消息")
	}
	
	// 软删除消息
	message.IsDeleted = true
	message.UpdatedAt = time.Now()
	
	if err := s.db.Save(&message).Error; err != nil {
		return fmt.Errorf("删除消息失败: %v", err)
	}
	
	return nil
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(messageID uint, userID uint) error {
	// 检查是否已标记为已读
	var readRecord model.MessageRead
	if err := s.db.Where("message_id = ? AND user_id = ?", messageID, userID).First(&readRecord).Error; err == nil {
		return nil // 已经标记为已读
	}
	
	// 创建已读记录
	readRecord = model.MessageRead{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}
	
	if err := s.db.Create(&readRecord).Error; err != nil {
		return fmt.Errorf("标记已读失败: %v", err)
	}
	
	// 更新消息已读状态
	var message model.Message
	if err := s.db.Where("id = ?", messageID).First(&message).Error; err == nil {
		message.IsRead = true
		message.UpdatedAt = time.Now()
		s.db.Save(&message)
	}
	
	return nil
}

// PinMessage 置顶消息
func (s *MessageService) PinMessage(messageID uint, userID uint) error {
	var message model.Message
	if err := s.db.Where("id = ?", messageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}
	
	// 验证用户是否为聊天管理员
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ? AND role IN ?", 
		message.ChatID, userID, []string{"owner", "admin"}).First(&member).Error; err != nil {
		return errors.New("您没有置顶消息的权限")
	}
	
	// 取消其他置顶消息
	s.db.Model(&model.Message{}).Where("chat_id = ? AND is_pinned = ?", message.ChatID, true).Update("is_pinned", false)
	
	// 置顶当前消息
	message.IsPinned = true
	message.UpdatedAt = time.Now()
	
	if err := s.db.Save(&message).Error; err != nil {
		return fmt.Errorf("置顶消息失败: %v", err)
	}
	
	return nil
}

// UnpinMessage 取消置顶消息
func (s *MessageService) UnpinMessage(messageID uint, userID uint) error {
	var message model.Message
	if err := s.db.Where("id = ?", messageID).First(&message).Error; err != nil {
		return errors.New("消息不存在")
	}
	
	// 验证用户是否为聊天管理员
	var member model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ? AND role IN ?", 
		message.ChatID, userID, []string{"owner", "admin"}).First(&member).Error; err != nil {
		return errors.New("您没有取消置顶消息的权限")
	}
	
	// 取消置顶
	message.IsPinned = false
	message.UpdatedAt = time.Now()
	
	if err := s.db.Save(&message).Error; err != nil {
		return fmt.Errorf("取消置顶消息失败: %v", err)
	}
	
	return nil
}

// GetChats 获取用户聊天列表
func (s *MessageService) GetChats(userID uint) ([]model.Chat, error) {
	var chats []model.Chat
	
	// 查询用户参与的聊天
	if err := s.db.Joins("JOIN chat_members ON chats.id = chat_members.chat_id").
		Where("chat_members.user_id = ? AND chats.is_active = ?", userID, true).
		Preload("Members").
		Order("chats.updated_at DESC").
		Find(&chats).Error; err != nil {
		return nil, fmt.Errorf("获取聊天列表失败: %v", err)
	}
	
	return chats, nil
}

// CreateChat 创建聊天
func (s *MessageService) CreateChat(creatorID uint, req CreateChatRequest) (*model.Chat, error) {
	// 创建聊天
	chat := model.Chat{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		IsActive:    true,
		IsPinned:    false,
		IsMuted:     false,
	}
	
	if err := s.db.Create(&chat).Error; err != nil {
		return nil, fmt.Errorf("创建聊天失败: %v", err)
	}
	
	// 添加创建者为成员
	creatorMember := model.ChatMember{
		ChatID:   chat.ID,
		UserID:   creatorID,
		Role:     "owner",
		JoinedAt: time.Now(),
	}
	
	if err := s.db.Create(&creatorMember).Error; err != nil {
		return nil, fmt.Errorf("添加创建者失败: %v", err)
	}
	
	// 添加其他成员
	for _, memberID := range req.Members {
		if memberID != creatorID {
			member := model.ChatMember{
				ChatID:   chat.ID,
				UserID:   memberID,
				Role:     "member",
				JoinedAt: time.Now(),
			}
			s.db.Create(&member)
		}
	}
	
	// 预加载关联数据
	s.db.Preload("Members").First(&chat, chat.ID)
	
	return &chat, nil
}


