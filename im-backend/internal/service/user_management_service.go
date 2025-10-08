package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// UserManagementService 用户管理服务
type UserManagementService struct {
	db *gorm.DB
}

// NewUserManagementService 创建用户管理服务
func NewUserManagementService(db *gorm.DB) *UserManagementService {
	return &UserManagementService{
		db: db,
	}
}

// BlacklistEntry 黑名单条目
type BlacklistEntry struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	Reason      string    `json:"reason" gorm:"type:text"`
	BlacklistType string  `json:"blacklist_type" gorm:"type:varchar(50)"` // temp, permanent, ip, device
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   uint      `json:"created_by"` // 管理员ID
}

// UserMgmtActivity 用户管理活动记录
type UserMgmtActivity struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Activity  string    `json:"activity" gorm:"type:varchar(100)"`
	IPAddress string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

// UserRestriction 用户限制
type UserRestriction struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	RestrictionType string `json:"restriction_type" gorm:"type:varchar(50)"` // message_limit, file_upload, voice_call, etc.
	LimitValue   int       `json:"limit_value"`
	CurrentUsage int       `json:"current_usage"`
	ResetTime    time.Time `json:"reset_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AddToBlacklist 添加用户到黑名单
func (s *UserManagementService) AddToBlacklist(ctx context.Context, userID uint, reason string, blacklistType string, expiresAt *time.Time, createdBy uint) error {
	entry := &BlacklistEntry{
		UserID:        userID,
		Reason:        reason,
		BlacklistType: blacklistType,
		ExpiresAt:     expiresAt,
		CreatedBy:     createdBy,
	}

	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("添加黑名单失败: %w", err)
	}

	// 记录用户活动
	s.LogUserMgmtActivity(ctx, userID, "added_to_blacklist", "", "")

	return nil
}

// RemoveFromBlacklist 从黑名单移除用户
func (s *UserManagementService) RemoveFromBlacklist(ctx context.Context, userID uint) error {
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&BlacklistEntry{}).Error; err != nil {
		return fmt.Errorf("移除黑名单失败: %w", err)
	}

	// 记录用户活动
	s.LogUserMgmtActivity(ctx, userID, "removed_from_blacklist", "", "")

	return nil
}

// IsBlacklisted 检查用户是否在黑名单中
func (s *UserManagementService) IsBlacklisted(ctx context.Context, userID uint) (bool, *BlacklistEntry, error) {
	var entry BlacklistEntry
	err := s.db.WithContext(ctx).Where("user_id = ? AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).First(&entry).Error
	
	if err == gorm.ErrRecordNotFound {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("检查黑名单失败: %w", err)
	}

	return true, &entry, nil
}

// GetBlacklist 获取黑名单列表
func (s *UserManagementService) GetBlacklist(ctx context.Context, page, pageSize int) ([]BlacklistEntry, int64, error) {
	var entries []BlacklistEntry
	var total int64

	// 获取总数
	if err := s.db.WithContext(ctx).Model(&BlacklistEntry{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取黑名单总数失败: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("CreatedByUser").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("获取黑名单列表失败: %w", err)
	}

	return entries, total, nil
}

// LogUserMgmtActivity 记录用户活动
func (s *UserManagementService) LogUserMgmtActivity(ctx context.Context, userID uint, activity, ipAddress, userAgent string) error {
	entry := &UserMgmtActivity{
		UserID:    userID,
		Activity:  activity,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("记录用户活动失败: %w", err)
	}

	return nil
}

// GetUserMgmtActivity 获取用户活动记录
func (s *UserManagementService) GetUserMgmtActivity(ctx context.Context, userID uint, page, pageSize int) ([]UserMgmtActivity, int64, error) {
	var activities []UserMgmtActivity
	var total int64

	// 获取总数
	if err := s.db.WithContext(ctx).Model(&UserMgmtActivity{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户活动总数失败: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&activities).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户活动失败: %w", err)
	}

	return activities, total, nil
}

// SetUserRestriction 设置用户限制
func (s *UserManagementService) SetUserRestriction(ctx context.Context, userID uint, restrictionType string, limitValue int) error {
	restriction := &UserRestriction{
		UserID:           userID,
		RestrictionType:  restrictionType,
		LimitValue:       limitValue,
		CurrentUsage:     0,
		ResetTime:        time.Now().Add(24 * time.Hour), // 默认24小时重置
	}

	if err := s.db.WithContext(ctx).Create(restriction).Error; err != nil {
		return fmt.Errorf("设置用户限制失败: %w", err)
	}

	// 记录用户活动
	s.LogUserMgmtActivity(ctx, userID, "restriction_set", "", "")

	return nil
}

// CheckUserRestriction 检查用户限制
func (s *UserManagementService) CheckUserRestriction(ctx context.Context, userID uint, restrictionType string) (bool, error) {
	var restriction UserRestriction
	err := s.db.WithContext(ctx).Where("user_id = ? AND restriction_type = ?", userID, restrictionType).First(&restriction).Error
	
	if err == gorm.ErrRecordNotFound {
		return true, nil // 没有限制
	}
	if err != nil {
		return false, fmt.Errorf("检查用户限制失败: %w", err)
	}

	// 检查是否超过限制
	if restriction.CurrentUsage >= restriction.LimitValue {
		return false, nil // 超过限制
	}

	return true, nil // 未超过限制
}

// IncrementUserRestriction 增加用户限制使用量
func (s *UserManagementService) IncrementUserRestriction(ctx context.Context, userID uint, restrictionType string) error {
	var restriction UserRestriction
	err := s.db.WithContext(ctx).Where("user_id = ? AND restriction_type = ?", userID, restrictionType).First(&restriction).Error
	
	if err == gorm.ErrRecordNotFound {
		return nil // 没有限制
	}
	if err != nil {
		return fmt.Errorf("获取用户限制失败: %w", err)
	}

	// 检查是否需要重置
	if time.Now().After(restriction.ResetTime) {
		restriction.CurrentUsage = 0
		restriction.ResetTime = time.Now().Add(24 * time.Hour)
	}

	// 增加使用量
	restriction.CurrentUsage++

	if err := s.db.WithContext(ctx).Save(&restriction).Error; err != nil {
		return fmt.Errorf("更新用户限制失败: %w", err)
	}

	return nil
}

// GetUserRestrictions 获取用户限制列表
func (s *UserManagementService) GetUserRestrictions(ctx context.Context, userID uint) ([]UserRestriction, error) {
	var restrictions []UserRestriction
	
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&restrictions).Error; err != nil {
		return nil, fmt.Errorf("获取用户限制失败: %w", err)
	}

	return restrictions, nil
}

// BanUser 封禁用户
func (s *UserManagementService) BanUser(ctx context.Context, userID uint, reason string, duration *time.Duration, createdBy uint) error {
	var expiresAt *time.Time
	if duration != nil {
		expiry := time.Now().Add(*duration)
		expiresAt = &expiry
	}

	// 添加到黑名单
	if err := s.AddToBlacklist(ctx, userID, reason, "permanent", expiresAt, createdBy); err != nil {
		return fmt.Errorf("封禁用户失败: %w", err)
	}

	// 更新用户状态
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("status", "banned").Error; err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	// 记录用户活动
	s.LogUserMgmtActivity(ctx, userID, "user_banned", "", "")

	return nil
}

// UnbanUser 解封用户
func (s *UserManagementService) UnbanUser(ctx context.Context, userID uint) error {
	// 从黑名单移除
	if err := s.RemoveFromBlacklist(ctx, userID); err != nil {
		return fmt.Errorf("解封用户失败: %w", err)
	}

	// 更新用户状态
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("status", "active").Error; err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	// 记录用户活动
	s.LogUserMgmtActivity(ctx, userID, "user_unbanned", "", "")

	return nil
}

// GetUserStats 获取用户统计信息
func (s *UserManagementService) GetUserStats(ctx context.Context, userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取用户基本信息
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	stats["user"] = user

	// 检查是否在黑名单中
	isBlacklisted, blacklistEntry, err := s.IsBlacklisted(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("检查黑名单状态失败: %w", err)
	}
	stats["is_blacklisted"] = isBlacklisted
	stats["blacklist_entry"] = blacklistEntry

	// 获取用户限制
	restrictions, err := s.GetUserRestrictions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户限制失败: %w", err)
	}
	stats["restrictions"] = restrictions

	// 获取最近活动
	activities, _, err := s.GetUserMgmtActivity(ctx, userID, 1, 10)
	if err != nil {
		return nil, fmt.Errorf("获取用户活动失败: %w", err)
	}
	stats["recent_activities"] = activities

	return stats, nil
}

// CleanupExpiredBlacklist 清理过期的黑名单条目
func (s *UserManagementService) CleanupExpiredBlacklist(ctx context.Context) error {
	if err := s.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&BlacklistEntry{}).Error; err != nil {
		return fmt.Errorf("清理过期黑名单失败: %w", err)
	}

	return nil
}

// GetSuspiciousUsers 获取可疑用户列表
func (s *UserManagementService) GetSuspiciousUsers(ctx context.Context, page, pageSize int) ([]model.User, error) {
	var users []model.User
	
	// 查询有异常活动的用户
	subQuery := s.db.WithContext(ctx).Model(&UserMgmtActivity{}).
		Select("user_id").
		Where("activity IN (?)", []string{"failed_login", "suspicious_activity", "spam_detected"}).
		Group("user_id").
		Having("COUNT(*) > ?", 5)

	if err := s.db.WithContext(ctx).
		Where("id IN (?)", subQuery).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("获取可疑用户失败: %w", err)
	}

	return users, nil
}
