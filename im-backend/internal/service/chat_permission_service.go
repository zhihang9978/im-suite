package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// ChatPermissionService 群组权限管理服务
type ChatPermissionService struct {
	db *gorm.DB
}

// NewChatPermissionService 创建群组权限管理服务
func NewChatPermissionService(db *gorm.DB) *ChatPermissionService {
	return &ChatPermissionService{
		db: db,
	}
}

// SetPermissionRequest 设置权限请求
type SetPermissionRequest struct {
	ChatID              uint  `json:"chat_id" binding:"required"`
	CanSendMessages     *bool `json:"can_send_messages,omitempty"`
	CanSendMedia        *bool `json:"can_send_media,omitempty"`
	CanSendStickers     *bool `json:"can_send_stickers,omitempty"`
	CanSendPolls        *bool `json:"can_send_polls,omitempty"`
	CanChangeInfo       *bool `json:"can_change_info,omitempty"`
	CanInviteUsers      *bool `json:"can_invite_users,omitempty"`
	CanPinMessages      *bool `json:"can_pin_messages,omitempty"`
	CanDeleteMessages   *bool `json:"can_delete_messages,omitempty"`
	CanEditMessages     *bool `json:"can_edit_messages,omitempty"`
	CanManageChat       *bool `json:"can_manage_chat,omitempty"`
	CanManageVoiceChats *bool `json:"can_manage_voice_chats,omitempty"`
	CanRestrictMembers  *bool `json:"can_restrict_members,omitempty"`
	CanPromoteMembers   *bool `json:"can_promote_members,omitempty"`
	CanAddAdmins        *bool `json:"can_add_admins,omitempty"`
}

// MuteMemberRequest 禁言成员请求
type MuteMemberRequest struct {
	ChatID   uint   `json:"chat_id" binding:"required"`
	UserID   uint   `json:"user_id" binding:"required"`
	Duration int    `json:"duration" binding:"required"` // 禁言时长（分钟）
	Reason   string `json:"reason,omitempty"`
}

// BanMemberRequest 踢出成员请求
type BanMemberRequest struct {
	ChatID uint   `json:"chat_id" binding:"required"`
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

// PromoteMemberRequest 提升成员权限请求
type PromoteMemberRequest struct {
	ChatID uint   `json:"chat_id" binding:"required"`
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"` // admin, owner
}

// SetChatPermissions 设置群组权限
func (s *ChatPermissionService) SetChatPermissions(ctx context.Context, userID uint, req *SetPermissionRequest) error {
	// 检查用户是否为群主或管理员
	if !s.hasPermission(ctx, req.ChatID, userID, "can_manage_chat") {
		return fmt.Errorf("没有权限管理群组设置")
	}

	// 查找或创建权限配置
	var permission model.ChatPermission
	if err := s.db.WithContext(ctx).Where("chat_id = ?", req.ChatID).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的权限配置
			permission = model.ChatPermission{
				ChatID: req.ChatID,
			}
		} else {
			return fmt.Errorf("查询权限配置失败: %w", err)
		}
	}

	// 更新权限设置
	updates := make(map[string]interface{})

	if req.CanSendMessages != nil {
		updates["can_send_messages"] = *req.CanSendMessages
	}
	if req.CanSendMedia != nil {
		updates["can_send_media"] = *req.CanSendMedia
	}
	if req.CanSendStickers != nil {
		updates["can_send_stickers"] = *req.CanSendStickers
	}
	if req.CanSendPolls != nil {
		updates["can_send_polls"] = *req.CanSendPolls
	}
	if req.CanChangeInfo != nil {
		updates["can_change_info"] = *req.CanChangeInfo
	}
	if req.CanInviteUsers != nil {
		updates["can_invite_users"] = *req.CanInviteUsers
	}
	if req.CanPinMessages != nil {
		updates["can_pin_messages"] = *req.CanPinMessages
	}
	if req.CanDeleteMessages != nil {
		updates["can_delete_messages"] = *req.CanDeleteMessages
	}
	if req.CanEditMessages != nil {
		updates["can_edit_messages"] = *req.CanEditMessages
	}
	if req.CanManageChat != nil {
		updates["can_manage_chat"] = *req.CanManageChat
	}
	if req.CanManageVoiceChats != nil {
		updates["can_manage_voice_chats"] = *req.CanManageVoiceChats
	}
	if req.CanRestrictMembers != nil {
		updates["can_restrict_members"] = *req.CanRestrictMembers
	}
	if req.CanPromoteMembers != nil {
		updates["can_promote_members"] = *req.CanPromoteMembers
	}
	if req.CanAddAdmins != nil {
		updates["can_add_admins"] = *req.CanAddAdmins
	}

	updates["updated_at"] = time.Now()

	if permission.ID == 0 {
		// 创建新权限配置
		if err := s.db.WithContext(ctx).Create(&permission).Error; err != nil {
			return fmt.Errorf("创建权限配置失败: %w", err)
		}
	} else {
		// 更新现有权限配置
		if err := s.db.WithContext(ctx).Model(&permission).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新权限配置失败: %w", err)
		}
	}

	return nil
}

// GetChatPermissions 获取群组权限配置
func (s *ChatPermissionService) GetChatPermissions(ctx context.Context, chatID uint, userID uint) (*model.ChatPermission, error) {
	// 检查用户是否为群成员
	if !s.isChatMember(ctx, chatID, userID) {
		return nil, fmt.Errorf("不是群成员")
	}

	var permission model.ChatPermission
	if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 返回默认权限配置
			return &model.ChatPermission{
				ChatID:              chatID,
				CanSendMessages:     true,
				CanSendMedia:        true,
				CanSendStickers:     true,
				CanSendPolls:        true,
				CanChangeInfo:       false,
				CanInviteUsers:      false,
				CanPinMessages:      false,
				CanDeleteMessages:   false,
				CanEditMessages:     false,
				CanManageChat:       false,
				CanManageVoiceChats: false,
				CanRestrictMembers:  false,
				CanPromoteMembers:   false,
				CanAddAdmins:        false,
			}, nil
		}
		return nil, fmt.Errorf("查询权限配置失败: %w", err)
	}

	return &permission, nil
}

// MuteMember 禁言成员
func (s *ChatPermissionService) MuteMember(ctx context.Context, userID uint, req *MuteMemberRequest) error {
	// 检查操作者权限
	if !s.hasPermission(ctx, req.ChatID, userID, "can_restrict_members") {
		return fmt.Errorf("没有权限禁言成员")
	}

	// 检查被禁言用户是否为群成员
	if !s.isChatMember(ctx, req.ChatID, req.UserID) {
		return fmt.Errorf("目标用户不是群成员")
	}

	// 检查被禁言用户是否为群主
	if s.isChatOwner(ctx, req.ChatID, req.UserID) {
		return fmt.Errorf("不能禁言群主")
	}

	// 检查操作者是否有权限禁言管理员
	if s.isChatAdmin(ctx, req.ChatID, req.UserID) && !s.isChatOwner(ctx, req.ChatID, userID) {
		return fmt.Errorf("只有群主才能禁言管理员")
	}

	// 计算禁言到期时间
	muteUntil := time.Now().Add(time.Duration(req.Duration) * time.Minute)

	// 更新成员状态
	if err := s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", req.ChatID, req.UserID).
		Updates(map[string]interface{}{
			"mute_until": &muteUntil,
			"is_banned":  false,
			"ban_reason": req.Reason,
		}).Error; err != nil {
		return fmt.Errorf("禁言成员失败: %w", err)
	}

	return nil
}

// UnmuteMember 解除禁言
func (s *ChatPermissionService) UnmuteMember(ctx context.Context, userID uint, chatID uint, targetUserID uint) error {
	// 检查操作者权限
	if !s.hasPermission(ctx, chatID, userID, "can_restrict_members") {
		return fmt.Errorf("没有权限解除禁言")
	}

	// 更新成员状态
	if err := s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, targetUserID).
		Updates(map[string]interface{}{
			"mute_until": nil,
			"is_banned":  false,
			"ban_reason": "",
		}).Error; err != nil {
		return fmt.Errorf("解除禁言失败: %w", err)
	}

	return nil
}

// BanMember 踢出成员
func (s *ChatPermissionService) BanMember(ctx context.Context, userID uint, req *BanMemberRequest) error {
	// 检查操作者权限
	if !s.hasPermission(ctx, req.ChatID, userID, "can_restrict_members") {
		return fmt.Errorf("没有权限踢出成员")
	}

	// 检查被踢出用户是否为群成员
	if !s.isChatMember(ctx, req.ChatID, req.UserID) {
		return fmt.Errorf("目标用户不是群成员")
	}

	// 检查被踢出用户是否为群主
	if s.isChatOwner(ctx, req.ChatID, req.UserID) {
		return fmt.Errorf("不能踢出群主")
	}

	// 检查操作者是否有权限踢出管理员
	if s.isChatAdmin(ctx, req.ChatID, req.UserID) && !s.isChatOwner(ctx, req.ChatID, userID) {
		return fmt.Errorf("只有群主才能踢出管理员")
	}

	// 更新成员状态
	if err := s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", req.ChatID, req.UserID).
		Updates(map[string]interface{}{
			"is_banned":  true,
			"ban_reason": req.Reason,
			"left_at":    time.Now(),
			"is_active":  false,
		}).Error; err != nil {
		return fmt.Errorf("踢出成员失败: %w", err)
	}

	return nil
}

// UnbanMember 解除封禁
func (s *ChatPermissionService) UnbanMember(ctx context.Context, userID uint, chatID uint, targetUserID uint) error {
	// 检查操作者权限
	if !s.hasPermission(ctx, chatID, userID, "can_restrict_members") {
		return fmt.Errorf("没有权限解除封禁")
	}

	// 更新成员状态
	if err := s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, targetUserID).
		Updates(map[string]interface{}{
			"is_banned":  false,
			"ban_reason": "",
			"is_active":  true,
		}).Error; err != nil {
		return fmt.Errorf("解除封禁失败: %w", err)
	}

	return nil
}

// PromoteMember 提升成员权限
func (s *ChatPermissionService) PromoteMember(ctx context.Context, userID uint, req *PromoteMemberRequest) error {
	// 检查操作者权限
	if req.Role == "owner" {
		if !s.isChatOwner(ctx, req.ChatID, userID) {
			return fmt.Errorf("只有群主可以转让群主权限")
		}
	} else if req.Role == "admin" {
		if !s.hasPermission(ctx, req.ChatID, userID, "can_promote_members") {
			return fmt.Errorf("没有权限提升成员为管理员")
		}
	} else {
		return fmt.Errorf("无效的角色")
	}

	// 检查目标用户是否为群成员
	if !s.isChatMember(ctx, req.ChatID, req.UserID) {
		return fmt.Errorf("目标用户不是群成员")
	}

	// 更新成员角色
	if err := s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", req.ChatID, req.UserID).
		Update("role", req.Role).Error; err != nil {
		return fmt.Errorf("提升成员权限失败: %w", err)
	}

	// 如果转让群主权限，需要更新原群主角色
	if req.Role == "owner" {
		if err := s.db.WithContext(ctx).Model(&model.ChatMember{}).
			Where("chat_id = ? AND user_id = ?", req.ChatID, userID).
			Update("role", "admin").Error; err != nil {
			return fmt.Errorf("更新原群主角色失败: %w", err)
		}
	}

	return nil
}

// DemoteMember 降级成员权限
func (s *ChatPermissionService) DemoteMember(ctx context.Context, userID uint, chatID uint, targetUserID uint) error {
	// 检查操作者权限
	if !s.hasPermission(ctx, chatID, userID, "can_promote_members") {
		return fmt.Errorf("没有权限降级成员")
	}

	// 检查目标用户是否为群成员
	if !s.isChatMember(ctx, chatID, targetUserID) {
		return fmt.Errorf("目标用户不是群成员")
	}

	// 检查目标用户是否为群主
	if s.isChatOwner(ctx, chatID, targetUserID) {
		return fmt.Errorf("不能降级群主")
	}

	// 更新成员角色
	if err := s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, targetUserID).
		Update("role", "member").Error; err != nil {
		return fmt.Errorf("降级成员权限失败: %w", err)
	}

	return nil
}

// GetChatMembers 获取群组成员列表
func (s *ChatPermissionService) GetChatMembers(ctx context.Context, chatID uint, userID uint) ([]model.ChatMember, error) {
	// 检查用户是否为群成员
	if !s.isChatMember(ctx, chatID, userID) {
		return nil, fmt.Errorf("不是群成员")
	}

	var members []model.ChatMember
	if err := s.db.WithContext(ctx).Preload("User").
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Order("role DESC, joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("获取群成员失败: %w", err)
	}

	return members, nil
}

// CheckPermission 检查用户权限
func (s *ChatPermissionService) CheckPermission(ctx context.Context, chatID uint, userID uint, permission string) bool {
	return s.hasPermission(ctx, chatID, userID, permission)
}

// 辅助方法

// hasPermission 检查用户是否有指定权限
func (s *ChatPermissionService) hasPermission(ctx context.Context, chatID uint, userID uint, permission string) bool {
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
	case "can_send_messages":
		return chatPermission.CanSendMessages
	case "can_send_media":
		return chatPermission.CanSendMedia
	case "can_send_stickers":
		return chatPermission.CanSendStickers
	case "can_send_polls":
		return chatPermission.CanSendPolls
	case "can_change_info":
		return chatPermission.CanChangeInfo
	case "can_invite_users":
		return chatPermission.CanInviteUsers
	case "can_pin_messages":
		return chatPermission.CanPinMessages
	case "can_delete_messages":
		return chatPermission.CanDeleteMessages
	case "can_edit_messages":
		return chatPermission.CanEditMessages
	case "can_manage_chat":
		return chatPermission.CanManageChat
	case "can_manage_voice_chats":
		return chatPermission.CanManageVoiceChats
	case "can_restrict_members":
		return chatPermission.CanRestrictMembers
	case "can_promote_members":
		return chatPermission.CanPromoteMembers
	case "can_add_admins":
		return chatPermission.CanAddAdmins
	default:
		return false
	}
}

// isChatMember 检查用户是否为群成员
func (s *ChatPermissionService) isChatMember(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND is_active = ?", chatID, userID, true).
		Count(&count)
	return count > 0
}

// isChatOwner 检查用户是否为群主
func (s *ChatPermissionService) isChatOwner(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND role = ? AND is_active = ?", chatID, userID, "owner", true).
		Count(&count)
	return count > 0
}

// isChatAdmin 检查用户是否为管理员
func (s *ChatPermissionService) isChatAdmin(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND role IN (?) AND is_active = ?", chatID, userID, []string{"owner", "admin"}, true).
		Count(&count)
	return count > 0
}

// getDefaultPermission 获取默认权限设置
func (s *ChatPermissionService) getDefaultPermission(permission string) bool {
	switch permission {
	case "can_send_messages", "can_send_media", "can_send_stickers", "can_send_polls":
		return true
	default:
		return false
	}
}
