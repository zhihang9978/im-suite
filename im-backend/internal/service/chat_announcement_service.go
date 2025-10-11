package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// ChatAnnouncementService 群组公告服务
type ChatAnnouncementService struct {
	db *gorm.DB
}

// NewChatAnnouncementService 创建群组公告服务
func NewChatAnnouncementService(db *gorm.DB) *ChatAnnouncementService {
	return &ChatAnnouncementService{
		db: db,
	}
}

// CreateAnnouncementRequest 创建公告请求
type CreateAnnouncementRequest struct {
	ChatID   uint   `json:"chat_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsPinned bool   `json:"is_pinned"`
}

// UpdateAnnouncementRequest 更新公告请求
type UpdateAnnouncementRequest struct {
	AnnouncementID uint   `json:"announcement_id" binding:"required"`
	Title          string `json:"title,omitempty"`
	Content        string `json:"content,omitempty"`
	IsPinned       *bool  `json:"is_pinned,omitempty"`
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	ChatID     uint   `json:"chat_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	RuleNumber int    `json:"rule_number" binding:"required"`
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	RuleID     uint   `json:"rule_id" binding:"required"`
	Title      string `json:"title,omitempty"`
	Content    string `json:"content,omitempty"`
	RuleNumber *int   `json:"rule_number,omitempty"`
}

// CreateAnnouncement 创建群组公告
func (s *ChatAnnouncementService) CreateAnnouncement(ctx context.Context, userID uint, req *CreateAnnouncementRequest) (*model.ChatAnnouncement, error) {
	// 检查用户权限
	if !s.hasPermission(ctx, req.ChatID, userID, "can_change_info") {
		return nil, fmt.Errorf("没有权限创建公告")
	}

	// 检查是否已有置顶公告
	if req.IsPinned {
		if err := s.db.WithContext(ctx).Model(&model.ChatAnnouncement{}).
			Where("chat_id = ? AND is_pinned = ? AND is_active = ?", req.ChatID, true, true).
			Update("is_pinned", false).Error; err != nil {
			return nil, fmt.Errorf("取消其他置顶公告失败: %w", err)
		}
	}

	// 创建公告
	announcement := &model.ChatAnnouncement{
		ChatID:   req.ChatID,
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
		IsPinned: req.IsPinned,
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(announcement).Error; err != nil {
		return nil, fmt.Errorf("创建公告失败: %w", err)
	}

	// 预加载关联数据
	if err := s.db.WithContext(ctx).Preload("Author").First(announcement, announcement.ID).Error; err != nil {
		return nil, fmt.Errorf("加载公告详情失败: %w", err)
	}

	return announcement, nil
}

// UpdateAnnouncement 更新群组公告
func (s *ChatAnnouncementService) UpdateAnnouncement(ctx context.Context, userID uint, req *UpdateAnnouncementRequest) error {
	var announcement model.ChatAnnouncement

	// 查找公告
	if err := s.db.WithContext(ctx).First(&announcement, req.AnnouncementID).Error; err != nil {
		return fmt.Errorf("公告不存在: %w", err)
	}

	// 检查用户权限
	if !s.hasPermission(ctx, announcement.ChatID, userID, "can_change_info") {
		return fmt.Errorf("没有权限修改公告")
	}

	// 检查是否为公告作者或管理员
	if announcement.AuthorID != userID && !s.isChatAdmin(ctx, announcement.ChatID, userID) {
		return fmt.Errorf("只有作者或管理员可以修改公告")
	}

	// 如果要置顶，先取消其他置顶公告
	if req.IsPinned != nil && *req.IsPinned {
		if err := s.db.WithContext(ctx).Model(&model.ChatAnnouncement{}).
			Where("chat_id = ? AND is_pinned = ? AND is_active = ? AND id != ?",
				announcement.ChatID, true, true, req.AnnouncementID).
			Update("is_pinned", false).Error; err != nil {
			return fmt.Errorf("取消其他置顶公告失败: %w", err)
		}
	}

	// 更新公告
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.IsPinned != nil {
		updates["is_pinned"] = *req.IsPinned
	}
	updates["updated_at"] = time.Now()

	if err := s.db.WithContext(ctx).Model(&announcement).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新公告失败: %w", err)
	}

	return nil
}

// DeleteAnnouncement 删除群组公告
func (s *ChatAnnouncementService) DeleteAnnouncement(ctx context.Context, userID uint, announcementID uint) error {
	var announcement model.ChatAnnouncement

	// 查找公告
	if err := s.db.WithContext(ctx).First(&announcement, announcementID).Error; err != nil {
		return fmt.Errorf("公告不存在: %w", err)
	}

	// 检查用户权限
	if !s.hasPermission(ctx, announcement.ChatID, userID, "can_change_info") {
		return fmt.Errorf("没有权限删除公告")
	}

	// 检查是否为公告作者或管理员
	if announcement.AuthorID != userID && !s.isChatAdmin(ctx, announcement.ChatID, userID) {
		return fmt.Errorf("只有作者或管理员可以删除公告")
	}

	// 软删除公告
	if err := s.db.WithContext(ctx).Model(&announcement).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("删除公告失败: %w", err)
	}

	return nil
}

// GetChatAnnouncements 获取群组公告列表
func (s *ChatAnnouncementService) GetChatAnnouncements(ctx context.Context, chatID uint, userID uint) ([]model.ChatAnnouncement, error) {
	// 检查用户是否为群成员
	if !s.isChatMember(ctx, chatID, userID) {
		return nil, fmt.Errorf("不是群成员")
	}

	var announcements []model.ChatAnnouncement
	if err := s.db.WithContext(ctx).Preload("Author").
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Order("is_pinned DESC, created_at DESC").
		Find(&announcements).Error; err != nil {
		return nil, fmt.Errorf("获取公告列表失败: %w", err)
	}

	return announcements, nil
}

// GetPinnedAnnouncement 获取置顶公告
func (s *ChatAnnouncementService) GetPinnedAnnouncement(ctx context.Context, chatID uint, userID uint) (*model.ChatAnnouncement, error) {
	// 检查用户是否为群成员
	if !s.isChatMember(ctx, chatID, userID) {
		return nil, fmt.Errorf("不是群成员")
	}

	var announcement model.ChatAnnouncement
	if err := s.db.WithContext(ctx).Preload("Author").
		Where("chat_id = ? AND is_pinned = ? AND is_active = ?", chatID, true, true).
		First(&announcement).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 没有置顶公告
		}
		return nil, fmt.Errorf("获取置顶公告失败: %w", err)
	}

	return &announcement, nil
}

// CreateRule 创建群组规则
func (s *ChatAnnouncementService) CreateRule(ctx context.Context, userID uint, req *CreateRuleRequest) (*model.ChatRule, error) {
	// 检查用户权限
	if !s.hasPermission(ctx, req.ChatID, userID, "can_change_info") {
		return nil, fmt.Errorf("没有权限创建规则")
	}

	// 检查规则编号是否已存在
	var existingRule model.ChatRule
	if err := s.db.WithContext(ctx).Where("chat_id = ? AND rule_number = ? AND is_active = ?",
		req.ChatID, req.RuleNumber, true).First(&existingRule).Error; err == nil {
		return nil, fmt.Errorf("规则编号 %d 已存在", req.RuleNumber)
	}

	// 创建规则
	rule := &model.ChatRule{
		ChatID:     req.ChatID,
		RuleNumber: req.RuleNumber,
		Title:      req.Title,
		Content:    req.Content,
		AuthorID:   userID,
		IsActive:   true,
	}

	if err := s.db.WithContext(ctx).Create(rule).Error; err != nil {
		return nil, fmt.Errorf("创建规则失败: %w", err)
	}

	// 预加载关联数据
	if err := s.db.WithContext(ctx).Preload("Author").First(rule, rule.ID).Error; err != nil {
		return nil, fmt.Errorf("加载规则详情失败: %w", err)
	}

	return rule, nil
}

// UpdateRule 更新群组规则
func (s *ChatAnnouncementService) UpdateRule(ctx context.Context, userID uint, req *UpdateRuleRequest) error {
	var rule model.ChatRule

	// 查找规则
	if err := s.db.WithContext(ctx).First(&rule, req.RuleID).Error; err != nil {
		return fmt.Errorf("规则不存在: %w", err)
	}

	// 检查用户权限
	if !s.hasPermission(ctx, rule.ChatID, userID, "can_change_info") {
		return fmt.Errorf("没有权限修改规则")
	}

	// 检查是否为规则作者或管理员
	if rule.AuthorID != userID && !s.isChatAdmin(ctx, rule.ChatID, userID) {
		return fmt.Errorf("只有作者或管理员可以修改规则")
	}

	// 如果要修改规则编号，检查是否冲突
	if req.RuleNumber != nil && *req.RuleNumber != rule.RuleNumber {
		var existingRule model.ChatRule
		if err := s.db.WithContext(ctx).Where("chat_id = ? AND rule_number = ? AND is_active = ? AND id != ?",
			rule.ChatID, *req.RuleNumber, true, req.RuleID).First(&existingRule).Error; err == nil {
			return fmt.Errorf("规则编号 %d 已存在", *req.RuleNumber)
		}
	}

	// 更新规则
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.RuleNumber != nil {
		updates["rule_number"] = *req.RuleNumber
	}
	updates["updated_at"] = time.Now()

	if err := s.db.WithContext(ctx).Model(&rule).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新规则失败: %w", err)
	}

	return nil
}

// DeleteRule 删除群组规则
func (s *ChatAnnouncementService) DeleteRule(ctx context.Context, userID uint, ruleID uint) error {
	var rule model.ChatRule

	// 查找规则
	if err := s.db.WithContext(ctx).First(&rule, ruleID).Error; err != nil {
		return fmt.Errorf("规则不存在: %w", err)
	}

	// 检查用户权限
	if !s.hasPermission(ctx, rule.ChatID, userID, "can_change_info") {
		return fmt.Errorf("没有权限删除规则")
	}

	// 检查是否为规则作者或管理员
	if rule.AuthorID != userID && !s.isChatAdmin(ctx, rule.ChatID, userID) {
		return fmt.Errorf("只有作者或管理员可以删除规则")
	}

	// 软删除规则
	if err := s.db.WithContext(ctx).Model(&rule).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("删除规则失败: %w", err)
	}

	return nil
}

// GetChatRules 获取群组规则列表
func (s *ChatAnnouncementService) GetChatRules(ctx context.Context, chatID uint, userID uint) ([]model.ChatRule, error) {
	// 检查用户是否为群成员
	if !s.isChatMember(ctx, chatID, userID) {
		return nil, fmt.Errorf("不是群成员")
	}

	var rules []model.ChatRule
	if err := s.db.WithContext(ctx).Preload("Author").
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Order("rule_number ASC").
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("获取规则列表失败: %w", err)
	}

	return rules, nil
}

// PinAnnouncement 置顶公告
func (s *ChatAnnouncementService) PinAnnouncement(ctx context.Context, userID uint, announcementID uint) error {
	var announcement model.ChatAnnouncement

	// 查找公告
	if err := s.db.WithContext(ctx).First(&announcement, announcementID).Error; err != nil {
		return fmt.Errorf("公告不存在: %w", err)
	}

	// 检查用户权限
	if !s.hasPermission(ctx, announcement.ChatID, userID, "can_change_info") {
		return fmt.Errorf("没有权限置顶公告")
	}

	// 取消其他置顶公告
	if err := s.db.WithContext(ctx).Model(&model.ChatAnnouncement{}).
		Where("chat_id = ? AND is_pinned = ? AND is_active = ? AND id != ?",
			announcement.ChatID, true, true, announcementID).
		Update("is_pinned", false).Error; err != nil {
		return fmt.Errorf("取消其他置顶公告失败: %w", err)
	}

	// 置顶当前公告
	if err := s.db.WithContext(ctx).Model(&announcement).Update("is_pinned", true).Error; err != nil {
		return fmt.Errorf("置顶公告失败: %w", err)
	}

	return nil
}

// UnpinAnnouncement 取消置顶公告
func (s *ChatAnnouncementService) UnpinAnnouncement(ctx context.Context, userID uint, announcementID uint) error {
	var announcement model.ChatAnnouncement

	// 查找公告
	if err := s.db.WithContext(ctx).First(&announcement, announcementID).Error; err != nil {
		return fmt.Errorf("公告不存在: %w", err)
	}

	// 检查用户权限
	if !s.hasPermission(ctx, announcement.ChatID, userID, "can_change_info") {
		return fmt.Errorf("没有权限取消置顶公告")
	}

	// 取消置顶
	if err := s.db.WithContext(ctx).Model(&announcement).Update("is_pinned", false).Error; err != nil {
		return fmt.Errorf("取消置顶公告失败: %w", err)
	}

	return nil
}

// 辅助方法

// hasPermission 检查用户是否有指定权限
func (s *ChatAnnouncementService) hasPermission(ctx context.Context, chatID uint, userID uint, permission string) bool {
	// 检查是否为群主
	if s.isChatOwner(ctx, chatID, userID) {
		return true
	}

	// 检查是否为管理员
	if !s.isChatAdmin(ctx, chatID, userID) {
		return false
	}

	// 获取群组权限配置
	var chatPermission model.ChatPermission
	if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&chatPermission).Error; err != nil {
		// 如果没有权限配置，使用默认设置
		return s.getDefaultPermission(permission)
	}

	// 根据权限类型检查
	switch permission {
	case "can_change_info":
		return chatPermission.CanChangeInfo
	default:
		return false
	}
}

// isChatMember 检查用户是否为群成员
func (s *ChatAnnouncementService) isChatMember(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND is_active = ?", chatID, userID, true).
		Count(&count)
	return count > 0
}

// isChatOwner 检查用户是否为群主
func (s *ChatAnnouncementService) isChatOwner(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND role = ? AND is_active = ?", chatID, userID, "owner", true).
		Count(&count)
	return count > 0
}

// isChatAdmin 检查用户是否为管理员
func (s *ChatAnnouncementService) isChatAdmin(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND role IN (?) AND is_active = ?", chatID, userID, []string{"owner", "admin"}, true).
		Count(&count)
	return count > 0
}

// getDefaultPermission 获取默认权限设置
func (s *ChatAnnouncementService) getDefaultPermission(permission string) bool {
	switch permission {
	case "can_change_info":
		return false // 默认只有群主可以修改群信息
	default:
		return false
	}
}
