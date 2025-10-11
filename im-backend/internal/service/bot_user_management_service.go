package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// BotUserManagementService 机器人用户管理服务
type BotUserManagementService struct {
	db *gorm.DB
}

// NewBotUserManagementService 创建机器人用户管理服务实例
func NewBotUserManagementService() *BotUserManagementService {
	return &BotUserManagementService{
		db: config.DB,
	}
}

// CreateBotUserRequest 创建机器人用户请求
type CreateBotUserRequest struct {
	BotID    uint   `json:"bot_id" binding:"required"`
	Username string `json:"username" binding:"required"` // 机器人的用户名
	Nickname string `json:"nickname" binding:"required"` // 机器人的昵称
	Avatar   string `json:"avatar"`                      // 机器人的头像
}

// GrantPermissionRequest 授权请求
type GrantPermissionRequest struct {
	UserID    uint       `json:"user_id" binding:"required"` // 被授权的用户ID
	BotID     uint       `json:"bot_id" binding:"required"`  // 机器人ID
	ExpiresAt *time.Time `json:"expires_at"`                 // 过期时间
}

// CreateBotUser 创建机器人用户（为机器人在系统中创建一个用户账号）
func (s *BotUserManagementService) CreateBotUser(ctx context.Context, adminID uint, req *CreateBotUserRequest) (*model.BotUser, error) {
	// 验证管理员权限
	var admin model.User
	if err := s.db.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return nil, errors.New("管理员不存在")
	}

	if admin.Role != "super_admin" {
		return nil, errors.New("只有超级管理员可以创建机器人用户")
	}

	// 验证机器人是否存在
	var bot model.Bot
	if err := s.db.WithContext(ctx).First(&bot, req.BotID).Error; err != nil {
		return nil, errors.New("机器人不存在")
	}

	// 检查机器人是否已关联用户
	var existing model.BotUser
	if err := s.db.WithContext(ctx).Where("bot_id = ?", req.BotID).First(&existing).Error; err == nil {
		return nil, errors.New("该机器人已关联用户账号")
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := s.db.WithContext(ctx).Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 创建系统用户
	user := model.User{
		Phone:    fmt.Sprintf("bot_%d", req.BotID), // 使用bot_id作为phone（唯一标识）
		Username: req.Username,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Password: "", // 机器人不需要密码
		Salt:     "",
		IsActive: true,
		Role:     "user",
		LastSeen: time.Now(),
		Online:   true, // 机器人始终在线
		Language: "zh-CN",
		Theme:    "auto",
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建机器人用户关联
	botUser := model.BotUser{
		BotID:    req.BotID,
		UserID:   user.ID,
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(&botUser).Error; err != nil {
		// 回滚用户创建
		s.db.WithContext(ctx).Delete(&user)
		return nil, fmt.Errorf("创建机器人用户关联失败: %w", err)
	}

	return &botUser, nil
}

// GrantPermission 授权用户使用机器人
func (s *BotUserManagementService) GrantPermission(ctx context.Context, adminID uint, req *GrantPermissionRequest) (*model.BotUserPermission, error) {
	// 验证管理员权限
	var admin model.User
	if err := s.db.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return nil, errors.New("管理员不存在")
	}

	if admin.Role != "admin" && admin.Role != "super_admin" {
		return nil, errors.New("需要管理员权限")
	}

	// 验证用户是否存在
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, req.UserID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 验证机器人是否存在
	var bot model.Bot
	if err := s.db.WithContext(ctx).First(&bot, req.BotID).Error; err != nil {
		return nil, errors.New("机器人不存在")
	}

	// 检查是否已授权
	var existing model.BotUserPermission
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND bot_id = ?", req.UserID, req.BotID).
		First(&existing).Error

	if err == nil {
		// 已存在，更新
		existing.IsActive = true
		existing.ExpiresAt = req.ExpiresAt
		existing.GrantedBy = adminID

		if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("更新权限失败: %w", err)
		}
		return &existing, nil
	}

	// 创建新授权
	permission := model.BotUserPermission{
		UserID:    req.UserID,
		BotID:     req.BotID,
		GrantedBy: adminID,
		IsActive:  true,
		ExpiresAt: req.ExpiresAt,
	}

	if err := s.db.WithContext(ctx).Create(&permission).Error; err != nil {
		return nil, fmt.Errorf("创建权限失败: %w", err)
	}

	return &permission, nil
}

// RevokePermission 撤销用户权限
func (s *BotUserManagementService) RevokePermission(ctx context.Context, adminID uint, userID uint, botID uint) error {
	// 验证管理员权限
	var admin model.User
	if err := s.db.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return errors.New("管理员不存在")
	}

	if admin.Role != "admin" && admin.Role != "super_admin" {
		return errors.New("需要管理员权限")
	}

	// 撤销权限
	result := s.db.WithContext(ctx).
		Model(&model.BotUserPermission{}).
		Where("user_id = ? AND bot_id = ?", userID, botID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("撤销权限失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("权限不存在")
	}

	return nil
}

// GetBotUser 获取机器人用户信息
func (s *BotUserManagementService) GetBotUser(ctx context.Context, botID uint) (*model.BotUser, error) {
	var botUser model.BotUser
	err := s.db.WithContext(ctx).
		Preload("Bot").
		Preload("User").
		Where("bot_id = ? AND is_active = ?", botID, true).
		First(&botUser).Error

	if err != nil {
		return nil, errors.New("机器人用户不存在")
	}

	return &botUser, nil
}

// GetUserPermissions 获取用户的机器人权限列表
func (s *BotUserManagementService) GetUserPermissions(ctx context.Context, userID uint) ([]model.BotUserPermission, error) {
	var permissions []model.BotUserPermission
	err := s.db.WithContext(ctx).
		Preload("Bot").
		Preload("GrantedByUser").
		Where("user_id = ? AND is_active = ?", userID, true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&permissions).Error

	return permissions, err
}

// GetBotPermissions 获取机器人的授权用户列表
func (s *BotUserManagementService) GetBotPermissions(ctx context.Context, botID uint) ([]model.BotUserPermission, error) {
	var permissions []model.BotUserPermission
	err := s.db.WithContext(ctx).
		Preload("User").
		Preload("GrantedByUser").
		Where("bot_id = ? AND is_active = ?", botID, true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&permissions).Error

	return permissions, err
}

func (s *BotUserManagementService) GetAllBotUsers(ctx context.Context) ([]model.BotUser, error) {
	var botUsers []model.BotUser
	if err := s.db.WithContext(ctx).Find(&botUsers).Error; err != nil {
		return nil, err
	}
	return botUsers, nil
}


// DeleteBotUser 删除机器人用户
func (s *BotUserManagementService) DeleteBotUser(ctx context.Context, adminID uint, botID uint) error {
	// 验证管理员权限
	var admin model.User
	if err := s.db.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return errors.New("管理员不存在")
	}

	if admin.Role != "super_admin" {
		return errors.New("只有超级管理员可以删除机器人用户")
	}

	// 查找机器人用户
	var botUser model.BotUser
	if err := s.db.WithContext(ctx).Where("bot_id = ?", botID).First(&botUser).Error; err != nil {
		return errors.New("机器人用户不存在")
	}

	// 删除关联权限
	s.db.WithContext(ctx).Where("bot_id = ?", botID).Delete(&model.BotUserPermission{})

	// 删除机器人用户关联
	if err := s.db.WithContext(ctx).Delete(&botUser).Error; err != nil {
		return fmt.Errorf("删除机器人用户失败: %w", err)
	}

	// 删除系统用户
	s.db.WithContext(ctx).Delete(&model.User{}, botUser.UserID)

	return nil
}
