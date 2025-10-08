package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/gorm"
)

// GroupManagementService 群组管理服务
type GroupManagementService struct {
	db *gorm.DB
}

// NewGroupManagementService 创建群组管理服务实例
func NewGroupManagementService(db *gorm.DB) *GroupManagementService {
	return &GroupManagementService{
		db: db,
	}
}

// CreateInviteRequest 创建邀请请求
type CreateInviteRequest struct {
	ChatID          uint       `json:"chat_id" binding:"required"`
	CreatorID       uint       `json:"creator_id" binding:"required"`
	MaxUses         int        `json:"max_uses"`
	ExpiresAt       *time.Time `json:"expires_at"`
	RequireApproval bool       `json:"require_approval"`
}

// ApproveJoinRequest 审批入群申请
type ApproveJoinRequestRequest struct {
	RequestID  uint   `json:"request_id" binding:"required"`
	ReviewerID uint   `json:"reviewer_id" binding:"required"`
	Approved   bool   `json:"approved"`
	ReviewNote string `json:"review_note"`
}

// PromoteAdminRequest 提升管理员请求
type PromoteAdminRequest struct {
	ChatID     uint   `json:"chat_id" binding:"required"`
	UserID     uint   `json:"user_id" binding:"required"`
	RoleID     uint   `json:"role_id" binding:"required"`
	Title      string `json:"title"`
	PromotedBy uint   `json:"promoted_by" binding:"required"`
}

// CreateInvite 创建邀请链接
func (s *GroupManagementService) CreateInvite(req CreateInviteRequest) (*model.GroupInvite, error) {
	// 验证群组是否存在
	var chat model.Chat
	if err := s.db.First(&chat, req.ChatID).Error; err != nil {
		return nil, errors.New("群组不存在")
	}

	// 生成唯一邀请码
	inviteCode := generateInviteCode()
	inviteLink := fmt.Sprintf("https://im.zhihang.com/invite/%s", inviteCode)

	invite := model.GroupInvite{
		ChatID:          req.ChatID,
		InviteCode:      inviteCode,
		InviteLink:      inviteLink,
		CreatorID:       req.CreatorID,
		MaxUses:         req.MaxUses,
		ExpiresAt:       req.ExpiresAt,
		RequireApproval: req.RequireApproval,
		IsEnabled:       true,
	}

	if err := s.db.Create(&invite).Error; err != nil {
		return nil, fmt.Errorf("创建邀请失败: %v", err)
	}

	// 记录审计日志
	s.logAudit(req.ChatID, req.CreatorID, "create_invite", "invite", invite.ID, 
		fmt.Sprintf("创建邀请链接: %s", inviteCode), "")

	return &invite, nil
}

// UseInvite 使用邀请链接
func (s *GroupManagementService) UseInvite(inviteCode string, userID uint, ipAddress string) error {
	// 查找邀请
	var invite model.GroupInvite
	if err := s.db.Preload("Chat").Where("invite_code = ?", inviteCode).First(&invite).Error; err != nil {
		return errors.New("邀请码无效")
	}

	// 验证邀请状态
	if !invite.IsEnabled {
		return errors.New("邀请已禁用")
	}
	if invite.IsRevoked {
		return errors.New("邀请已被撤销")
	}
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		return errors.New("邀请已过期")
	}
	if invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses {
		return errors.New("邀请已达到最大使用次数")
	}

	// 检查用户是否已在群组中
	var existingMember model.ChatMember
	if err := s.db.Where("chat_id = ? AND user_id = ?", invite.ChatID, userID).First(&existingMember).Error; err == nil {
		return errors.New("您已经在群组中")
	}

	// 记录使用
	usage := model.GroupInviteUsage{
		InviteID:  invite.ID,
		UserID:    userID,
		IPAddress: ipAddress,
		Status:    "pending",
	}

	if invite.RequireApproval {
		// 需要审批，创建入群申请
		joinRequest := model.GroupJoinRequest{
			ChatID:   invite.ChatID,
			UserID:   userID,
			InviteID: &invite.ID,
			Message:  "通过邀请链接申请加入",
			Status:   "pending",
		}
		if err := s.db.Create(&joinRequest).Error; err != nil {
			return fmt.Errorf("创建入群申请失败: %v", err)
		}
		usage.Status = "pending"
	} else {
		// 直接加入
		member := model.ChatMember{
			ChatID: invite.ChatID,
			UserID: userID,
			Role:   "member",
		}
		if err := s.db.Create(&member).Error; err != nil {
			return fmt.Errorf("加入群组失败: %v", err)
		}
		usage.Status = "approved"
	}

	// 保存使用记录
	s.db.Create(&usage)

	// 更新使用次数
	s.db.Model(&invite).Update("used_count", gorm.Expr("used_count + 1"))

	// 记录审计日志
	s.logAudit(invite.ChatID, userID, "use_invite", "invite", invite.ID, 
		fmt.Sprintf("使用邀请链接: %s", inviteCode), "")

	return nil
}

// RevokeInvite 撤销邀请
func (s *GroupManagementService) RevokeInvite(inviteID, revokerID uint, reason string) error {
	var invite model.GroupInvite
	if err := s.db.First(&invite, inviteID).Error; err != nil {
		return errors.New("邀请不存在")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_revoked":     true,
		"revoked_by":     revokerID,
		"revoked_at":     now,
		"revoke_reason":  reason,
	}

	if err := s.db.Model(&invite).Updates(updates).Error; err != nil {
		return fmt.Errorf("撤销邀请失败: %v", err)
	}

	// 记录审计日志
	s.logAudit(invite.ChatID, revokerID, "revoke_invite", "invite", inviteID, 
		fmt.Sprintf("撤销邀请: %s", reason), "")

	return nil
}

// GetChatInvites 获取群组邀请列表
func (s *GroupManagementService) GetChatInvites(chatID uint, limit, offset int) ([]model.GroupInvite, int64, error) {
	var invites []model.GroupInvite
	var total int64

	query := s.db.Preload("Creator").Where("chat_id = ?", chatID)

	// 获取总数
	query.Model(&model.GroupInvite{}).Count(&total)

	// 获取分页数据
	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&invites).Error; err != nil {
		return nil, 0, fmt.Errorf("获取邀请列表失败: %v", err)
	}

	return invites, total, nil
}

// ApproveJoinRequest 审批入群申请
func (s *GroupManagementService) ApproveJoinRequest(req ApproveJoinRequestRequest) error {
	var joinRequest model.GroupJoinRequest
	if err := s.db.First(&joinRequest, req.RequestID).Error; err != nil {
		return errors.New("申请不存在")
	}

	if joinRequest.Status != "pending" {
		return errors.New("申请已处理")
	}

	now := time.Now()
	status := "rejected"
	if req.Approved {
		status = "approved"
		
		// 添加用户到群组
		member := model.ChatMember{
			ChatID: joinRequest.ChatID,
			UserID: joinRequest.UserID,
			Role:   "member",
		}
		if err := s.db.Create(&member).Error; err != nil {
			return fmt.Errorf("添加成员失败: %v", err)
		}
	}

	// 更新申请状态
	updates := map[string]interface{}{
		"status":       status,
		"reviewed_by":  req.ReviewerID,
		"reviewed_at":  now,
		"review_note":  req.ReviewNote,
	}
	if err := s.db.Model(&joinRequest).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新申请状态失败: %v", err)
	}

	// 记录审计日志
	action := "reject_join_request"
	if req.Approved {
		action = "approve_join_request"
	}
	s.logAudit(joinRequest.ChatID, req.ReviewerID, action, "join_request", req.RequestID, 
		fmt.Sprintf("审批入群申请: %s", req.ReviewNote), "")

	return nil
}

// GetPendingJoinRequests 获取待审批的入群申请
func (s *GroupManagementService) GetPendingJoinRequests(chatID uint, limit, offset int) ([]model.GroupJoinRequest, int64, error) {
	var requests []model.GroupJoinRequest
	var total int64

	query := s.db.Preload("User").Preload("Invite").
		Where("chat_id = ? AND status = ?", chatID, "pending")

	// 获取总数
	query.Model(&model.GroupJoinRequest{}).Count(&total)

	// 获取分页数据
	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&requests).Error; err != nil {
		return nil, 0, fmt.Errorf("获取入群申请失败: %v", err)
	}

	return requests, total, nil
}

// InitializeAdminRoles 初始化管理员角色
func (s *GroupManagementService) InitializeAdminRoles() error {
	// 检查是否已初始化
	var count int64
	s.db.Model(&model.AdminRole{}).Where("is_built_in = ?", true).Count(&count)
	if count > 0 {
		return nil
	}

	// 超级管理员（群主）
	owner := model.AdminRole{
		Name:                 "owner",
		DisplayName:          "群主",
		Description:          "群组创建者，拥有所有权限",
		Level:                100,
		IsBuiltIn:            true,
		IsEnabled:            true,
		CanManageMembers:     true,
		CanDeleteMessages:    true,
		CanEditChat:          true,
		CanInviteUsers:       true,
		CanBanUsers:          true,
		CanPromoteMembers:    true,
		CanManagePermissions: true,
		CanManageInvites:     true,
		CanPinMessages:       true,
		CanManageAnnouncements: true,
		CanViewStatistics:    true,
		CanManageRoles:       true,
	}

	// 管理员
	admin := model.AdminRole{
		Name:                 "admin",
		DisplayName:          "管理员",
		Description:          "群组管理员，拥有大部分管理权限",
		Level:                80,
		IsBuiltIn:            true,
		IsEnabled:            true,
		CanManageMembers:     true,
		CanDeleteMessages:    true,
		CanEditChat:          false,
		CanInviteUsers:       true,
		CanBanUsers:          true,
		CanPromoteMembers:    false,
		CanManagePermissions: false,
		CanManageInvites:     true,
		CanPinMessages:       true,
		CanManageAnnouncements: true,
		CanViewStatistics:    true,
		CanManageRoles:       false,
	}

	// 协管员
	moderator := model.AdminRole{
		Name:                 "moderator",
		DisplayName:          "协管员",
		Description:          "群组协管员，拥有基础管理权限",
		Level:                50,
		IsBuiltIn:            true,
		IsEnabled:            true,
		CanManageMembers:     false,
		CanDeleteMessages:    true,
		CanEditChat:          false,
		CanInviteUsers:       true,
		CanBanUsers:          false,
		CanPromoteMembers:    false,
		CanManagePermissions: false,
		CanManageInvites:     false,
		CanPinMessages:       true,
		CanManageAnnouncements: false,
		CanViewStatistics:    false,
		CanManageRoles:       false,
	}

	roles := []model.AdminRole{owner, admin, moderator}
	for _, role := range roles {
		if err := s.db.Create(&role).Error; err != nil {
			return fmt.Errorf("初始化管理员角色失败: %v", err)
		}
	}

	return nil
}

// PromoteMember 提升成员为管理员
func (s *GroupManagementService) PromoteMember(req PromoteAdminRequest) error {
	// 验证角色是否存在
	var role model.AdminRole
	if err := s.db.First(&role, req.RoleID).Error; err != nil {
		return errors.New("角色不存在")
	}

	// 检查是否已经是管理员
	var existingAdmin model.ChatAdmin
	if err := s.db.Where("chat_id = ? AND user_id = ? AND is_active = ?", 
		req.ChatID, req.UserID, true).First(&existingAdmin).Error; err == nil {
		return errors.New("该用户已经是管理员")
	}

	// 创建管理员记录
	admin := model.ChatAdmin{
		ChatID:     req.ChatID,
		UserID:     req.UserID,
		RoleID:     req.RoleID,
		Title:      req.Title,
		PromotedBy: req.PromotedBy,
		IsActive:   true,
	}

	if err := s.db.Create(&admin).Error; err != nil {
		return fmt.Errorf("提升管理员失败: %v", err)
	}

	// 记录审计日志
	s.logAudit(req.ChatID, req.PromotedBy, "promote_member", "user", req.UserID, 
		fmt.Sprintf("提升为%s", role.DisplayName), "")

	return nil
}

// DemoteMember 降级管理员
func (s *GroupManagementService) DemoteMember(chatID, userID, demotedBy uint) error {
	var admin model.ChatAdmin
	if err := s.db.Where("chat_id = ? AND user_id = ? AND is_active = ?", 
		chatID, userID, true).First(&admin).Error; err != nil {
		return errors.New("该用户不是管理员")
	}

	if err := s.db.Model(&admin).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("降级管理员失败: %v", err)
	}

	// 记录审计日志
	s.logAudit(chatID, demotedBy, "demote_member", "user", userID, 
		"降级管理员", "")

	return nil
}

// GetChatAdmins 获取群组管理员列表
func (s *GroupManagementService) GetChatAdmins(chatID uint) ([]model.ChatAdmin, error) {
	var admins []model.ChatAdmin
	if err := s.db.Preload("User").Preload("Role").Preload("PromotedByUser").
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Order("role_id DESC").Find(&admins).Error; err != nil {
		return nil, fmt.Errorf("获取管理员列表失败: %v", err)
	}
	return admins, nil
}

// CheckPermission 检查用户权限
func (s *GroupManagementService) CheckPermission(chatID, userID uint, permission string) (bool, error) {
	var admin model.ChatAdmin
	if err := s.db.Preload("Role").
		Where("chat_id = ? AND user_id = ? AND is_active = ?", chatID, userID, true).
		First(&admin).Error; err != nil {
		return false, nil // 不是管理员，没有权限
	}

	role := admin.Role

	// 根据权限名称检查
	switch permission {
	case "can_manage_members":
		return role.CanManageMembers, nil
	case "can_delete_messages":
		return role.CanDeleteMessages, nil
	case "can_edit_chat":
		return role.CanEditChat, nil
	case "can_invite_users":
		return role.CanInviteUsers, nil
	case "can_ban_users":
		return role.CanBanUsers, nil
	case "can_promote_members":
		return role.CanPromoteMembers, nil
	case "can_manage_permissions":
		return role.CanManagePermissions, nil
	case "can_manage_invites":
		return role.CanManageInvites, nil
	case "can_pin_messages":
		return role.CanPinMessages, nil
	case "can_manage_announcements":
		return role.CanManageAnnouncements, nil
	case "can_view_statistics":
		return role.CanViewStatistics, nil
	case "can_manage_roles":
		return role.CanManageRoles, nil
	default:
		return false, errors.New("未知权限")
	}
}

// GetAuditLogs 获取审计日志
func (s *GroupManagementService) GetAuditLogs(chatID uint, limit, offset int) ([]model.GroupAuditLog, int64, error) {
	var logs []model.GroupAuditLog
	var total int64

	query := s.db.Preload("Operator").Where("chat_id = ?", chatID)

	// 获取总数
	query.Model(&model.GroupAuditLog{}).Count(&total)

	// 获取分页数据
	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("获取审计日志失败: %v", err)
	}

	return logs, total, nil
}

// logAudit 记录审计日志
func (s *GroupManagementService) logAudit(chatID, operatorID uint, action, targetType string, targetID uint, details, ipAddress string) {
	log := model.GroupAuditLog{
		ChatID:     chatID,
		OperatorID: operatorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
		IPAddress:  ipAddress,
	}
	s.db.Create(&log)
}

// generateInviteCode 生成邀请码
func generateInviteCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
