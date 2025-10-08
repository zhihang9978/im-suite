package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// ChatBackupService 群组备份服务
type ChatBackupService struct {
	db *gorm.DB
}

// NewChatBackupService 创建群组备份服务
func NewChatBackupService(db *gorm.DB) *ChatBackupService {
	return &ChatBackupService{
		db: db,
	}
}

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
	ChatID      uint   `json:"chat_id" binding:"required"`
	BackupType  string `json:"backup_type" binding:"required"` // full, messages, media, settings
	IsEncrypted bool   `json:"is_encrypted"`
	ExpiresIn   *int   `json:"expires_in,omitempty"` // 过期时间（小时）
}

// RestoreBackupRequest 恢复备份请求
type RestoreBackupRequest struct {
	BackupID uint `json:"backup_id" binding:"required"`
	ChatID   uint `json:"chat_id" binding:"required"`
}

// BackupData 备份数据结构
type BackupData struct {
	ChatInfo     *model.Chat                    `json:"chat_info,omitempty"`
	Members      []model.ChatMember             `json:"members,omitempty"`
	Messages     []model.Message                `json:"messages,omitempty"`
	Announcements []model.ChatAnnouncement      `json:"announcements,omitempty"`
	Rules        []model.ChatRule               `json:"rules,omitempty"`
	Permissions  *model.ChatPermission          `json:"permissions,omitempty"`
	Statistics   *model.ChatStatistics          `json:"statistics,omitempty"`
	CreatedAt    time.Time                      `json:"created_at"`
	CreatedBy    uint                           `json:"created_by"`
	BackupType   string                         `json:"backup_type"`
}

// CreateBackup 创建群组备份
func (s *ChatBackupService) CreateBackup(ctx context.Context, userID uint, req *CreateBackupRequest) (*model.ChatBackup, error) {
	// 检查用户权限
	if !s.hasPermission(ctx, req.ChatID, userID, "can_manage_chat") {
		return nil, fmt.Errorf("没有权限创建备份")
	}

	// 收集备份数据
	backupData, err := s.collectBackupData(ctx, req.ChatID, req.BackupType)
	if err != nil {
		return nil, fmt.Errorf("收集备份数据失败: %w", err)
	}

	// 序列化备份数据
	dataJSON, err := json.Marshal(backupData)
	if err != nil {
		return nil, fmt.Errorf("序列化备份数据失败: %w", err)
	}

	// 计算过期时间
	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		expireTime := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &expireTime
	}

	// 创建备份记录
	backup := &model.ChatBackup{
		ChatID:      req.ChatID,
		BackupType:  req.BackupType,
		BackupData:  string(dataJSON),
		BackupSize:  int64(len(dataJSON)),
		CreatedBy:   userID,
		IsEncrypted: req.IsEncrypted,
		ExpiresAt:   expiresAt,
	}

	if err := s.db.WithContext(ctx).Create(backup).Error; err != nil {
		return nil, fmt.Errorf("创建备份记录失败: %w", err)
	}

	return backup, nil
}

// RestoreBackup 恢复群组备份
func (s *ChatBackupService) RestoreBackup(ctx context.Context, userID uint, req *RestoreBackupRequest) error {
	// 检查用户权限
	if !s.hasPermission(ctx, req.ChatID, userID, "can_manage_chat") {
		return fmt.Errorf("没有权限恢复备份")
	}

	// 查找备份记录
	var backup model.ChatBackup
	if err := s.db.WithContext(ctx).First(&backup, req.BackupID).Error; err != nil {
		return fmt.Errorf("备份不存在: %w", err)
	}

	// 检查备份是否过期
	if backup.ExpiresAt != nil && backup.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("备份已过期")
	}

	// 反序列化备份数据
	var backupData BackupData
	if err := json.Unmarshal([]byte(backup.BackupData), &backupData); err != nil {
		return fmt.Errorf("反序列化备份数据失败: %w", err)
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 根据备份类型恢复数据
	switch backup.BackupType {
	case "full":
		if err := s.restoreFullBackup(tx, req.ChatID, &backupData); err != nil {
			tx.Rollback()
			return fmt.Errorf("恢复完整备份失败: %w", err)
		}
	case "messages":
		if err := s.restoreMessagesBackup(tx, req.ChatID, &backupData); err != nil {
			tx.Rollback()
			return fmt.Errorf("恢复消息备份失败: %w", err)
		}
	case "media":
		if err := s.restoreMediaBackup(tx, req.ChatID, &backupData); err != nil {
			tx.Rollback()
			return fmt.Errorf("恢复媒体备份失败: %w", err)
		}
	case "settings":
		if err := s.restoreSettingsBackup(tx, req.ChatID, &backupData); err != nil {
			tx.Rollback()
			return fmt.Errorf("恢复设置备份失败: %w", err)
		}
	default:
		tx.Rollback()
		return fmt.Errorf("不支持的备份类型")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetBackupList 获取备份列表
func (s *ChatBackupService) GetBackupList(ctx context.Context, chatID uint, userID uint) ([]model.ChatBackup, error) {
	// 检查用户权限
	if !s.isChatMember(ctx, chatID, userID) {
		return nil, fmt.Errorf("不是群成员")
	}

	var backups []model.ChatBackup
	if err := s.db.WithContext(ctx).Preload("Creator").
		Where("chat_id = ? AND (expires_at IS NULL OR expires_at > ?)", chatID, time.Now()).
		Order("created_at DESC").
		Find(&backups).Error; err != nil {
		return nil, fmt.Errorf("获取备份列表失败: %w", err)
	}

	return backups, nil
}

// DeleteBackup 删除备份
func (s *ChatBackupService) DeleteBackup(ctx context.Context, userID uint, backupID uint) error {
	var backup model.ChatBackup
	
	// 查找备份记录
	if err := s.db.WithContext(ctx).First(&backup, backupID).Error; err != nil {
		return fmt.Errorf("备份不存在: %w", err)
	}

	// 检查用户权限
	if !s.hasPermission(ctx, backup.ChatID, userID, "can_manage_chat") {
		return fmt.Errorf("没有权限删除备份")
	}

	// 检查是否为备份创建者
	if backup.CreatedBy != userID && !s.isChatAdmin(ctx, backup.ChatID, userID) {
		return fmt.Errorf("只有创建者或管理员可以删除备份")
	}

	// 删除备份记录
	if err := s.db.WithContext(ctx).Delete(&backup).Error; err != nil {
		return fmt.Errorf("删除备份失败: %w", err)
	}

	return nil
}

// CleanupExpiredBackups 清理过期备份（定时任务调用）
func (s *ChatBackupService) CleanupExpiredBackups(ctx context.Context) error {
	now := time.Now()
	
	if err := s.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ?", now).
		Delete(&model.ChatBackup{}).Error; err != nil {
		return fmt.Errorf("清理过期备份失败: %w", err)
	}

	return nil
}

// collectBackupData 收集备份数据
func (s *ChatBackupService) collectBackupData(ctx context.Context, chatID uint, backupType string) (*BackupData, error) {
	backupData := &BackupData{
		ChatID:     chatID,
		BackupType: backupType,
		CreatedAt:  time.Now(),
	}

	switch backupType {
	case "full":
		// 完整备份
		if err := s.collectFullBackupData(ctx, chatID, backupData); err != nil {
			return nil, err
		}
	case "messages":
		// 消息备份
		if err := s.collectMessagesBackupData(ctx, chatID, backupData); err != nil {
			return nil, err
		}
	case "media":
		// 媒体备份
		if err := s.collectMediaBackupData(ctx, chatID, backupData); err != nil {
			return nil, err
		}
	case "settings":
		// 设置备份
		if err := s.collectSettingsBackupData(ctx, chatID, backupData); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的备份类型: %s", backupType)
	}

	return backupData, nil
}

// collectFullBackupData 收集完整备份数据
func (s *ChatBackupService) collectFullBackupData(ctx context.Context, chatID uint, backupData *BackupData) error {
	// 收集群组信息
	var chat model.Chat
	if err := s.db.WithContext(ctx).First(&chat, chatID).Error; err != nil {
		return fmt.Errorf("获取群组信息失败: %w", err)
	}
	backupData.ChatInfo = &chat

	// 收集成员信息
	if err := s.db.WithContext(ctx).Preload("User").
		Where("chat_id = ?", chatID).Find(&backupData.Members).Error; err != nil {
		return fmt.Errorf("获取成员信息失败: %w", err)
	}

	// 收集消息
	if err := s.db.WithContext(ctx).Preload("Sender").
		Where("chat_id = ?", chatID).Find(&backupData.Messages).Error; err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	// 收集公告
	if err := s.db.WithContext(ctx).Preload("Author").
		Where("chat_id = ? AND is_active = ?", chatID, true).Find(&backupData.Announcements).Error; err != nil {
		return fmt.Errorf("获取公告失败: %w", err)
	}

	// 收集规则
	if err := s.db.WithContext(ctx).Preload("Author").
		Where("chat_id = ? AND is_active = ?", chatID, true).Find(&backupData.Rules).Error; err != nil {
		return fmt.Errorf("获取规则失败: %w", err)
	}

	// 收集权限配置
	var permission model.ChatPermission
	if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&permission).Error; err == nil {
		backupData.Permissions = &permission
	}

	// 收集统计信息
	var statistics model.ChatStatistics
	if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&statistics).Error; err == nil {
		backupData.Statistics = &statistics
	}

	return nil
}

// collectMessagesBackupData 收集消息备份数据
func (s *ChatBackupService) collectMessagesBackupData(ctx context.Context, chatID uint, backupData *BackupData) error {
	// 只收集消息
	if err := s.db.WithContext(ctx).Preload("Sender").
		Where("chat_id = ?", chatID).Find(&backupData.Messages).Error; err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	return nil
}

// collectMediaBackupData 收集媒体备份数据
func (s *ChatBackupService) collectMediaBackupData(ctx context.Context, chatID uint, backupData *BackupData) error {
	// 收集媒体消息
	if err := s.db.WithContext(ctx).Preload("Sender").
		Where("chat_id = ? AND message_type IN (?)", chatID, []string{"image", "video", "audio", "file"}).
		Find(&backupData.Messages).Error; err != nil {
		return fmt.Errorf("获取媒体消息失败: %w", err)
	}

	return nil
}

// collectSettingsBackupData 收集设置备份数据
func (s *ChatBackupService) collectSettingsBackupData(ctx context.Context, chatID uint, backupData *BackupData) error {
	// 收集群组信息
	var chat model.Chat
	if err := s.db.WithContext(ctx).First(&chat, chatID).Error; err != nil {
		return fmt.Errorf("获取群组信息失败: %w", err)
	}
	backupData.ChatInfo = &chat

	// 收集权限配置
	var permission model.ChatPermission
	if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&permission).Error; err == nil {
		backupData.Permissions = &permission
	}

	// 收集公告
	if err := s.db.WithContext(ctx).Preload("Author").
		Where("chat_id = ? AND is_active = ?", chatID, true).Find(&backupData.Announcements).Error; err != nil {
		return fmt.Errorf("获取公告失败: %w", err)
	}

	// 收集规则
	if err := s.db.WithContext(ctx).Preload("Author").
		Where("chat_id = ? AND is_active = ?", chatID, true).Find(&backupData.Rules).Error; err != nil {
		return fmt.Errorf("获取规则失败: %w", err)
	}

	return nil
}

// restoreFullBackup 恢复完整备份
func (s *ChatBackupService) restoreFullBackup(tx *gorm.DB, chatID uint, backupData *BackupData) error {
	// 恢复群组信息
	if backupData.ChatInfo != nil {
		if err := tx.Model(&model.Chat{}).Where("id = ?", chatID).Updates(map[string]interface{}{
			"name":        backupData.ChatInfo.Name,
			"description": backupData.ChatInfo.Description,
			"avatar":      backupData.ChatInfo.Avatar,
			"updated_at":  time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("恢复群组信息失败: %w", err)
		}
	}

	// 恢复权限配置
	if backupData.Permissions != nil {
		var permission model.ChatPermission
		if err := tx.Where("chat_id = ?", chatID).First(&permission).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建新权限配置
				permission = *backupData.Permissions
				permission.ID = 0
				permission.ChatID = chatID
				if err := tx.Create(&permission).Error; err != nil {
					return fmt.Errorf("创建权限配置失败: %w", err)
				}
			} else {
				return fmt.Errorf("查询权限配置失败: %w", err)
			}
		} else {
			// 更新现有权限配置
			if err := tx.Model(&permission).Updates(map[string]interface{}{
				"can_send_messages":       backupData.Permissions.CanSendMessages,
				"can_send_media":          backupData.Permissions.CanSendMedia,
				"can_send_stickers":       backupData.Permissions.CanSendStickers,
				"can_send_polls":          backupData.Permissions.CanSendPolls,
				"can_change_info":         backupData.Permissions.CanChangeInfo,
				"can_invite_users":        backupData.Permissions.CanInviteUsers,
				"can_pin_messages":        backupData.Permissions.CanPinMessages,
				"can_delete_messages":     backupData.Permissions.CanDeleteMessages,
				"can_edit_messages":       backupData.Permissions.CanEditMessages,
				"can_manage_chat":         backupData.Permissions.CanManageChat,
				"can_manage_voice_chats":  backupData.Permissions.CanManageVoiceChats,
				"can_restrict_members":    backupData.Permissions.CanRestrictMembers,
				"can_promote_members":     backupData.Permissions.CanPromoteMembers,
				"can_add_admins":          backupData.Permissions.CanAddAdmins,
				"updated_at":              time.Now(),
			}).Error; err != nil {
				return fmt.Errorf("更新权限配置失败: %w", err)
			}
		}
	}

	return nil
}

// restoreMessagesBackup 恢复消息备份
func (s *ChatBackupService) restoreMessagesBackup(tx *gorm.DB, chatID uint, backupData *BackupData) error {
	// 注意：消息恢复需要谨慎处理，避免重复消息
	// 这里可以实现增量恢复或按时间范围恢复
	return fmt.Errorf("消息恢复功能待实现")
}

// restoreMediaBackup 恢复媒体备份
func (s *ChatBackupService) restoreMediaBackup(tx *gorm.DB, chatID uint, backupData *BackupData) error {
	// 媒体恢复通常涉及文件存储，需要额外处理
	return fmt.Errorf("媒体恢复功能待实现")
}

// restoreSettingsBackup 恢复设置备份
func (s *ChatBackupService) restoreSettingsBackup(tx *gorm.DB, chatID uint, backupData *BackupData) error {
	// 恢复群组设置
	if backupData.ChatInfo != nil {
		if err := tx.Model(&model.Chat{}).Where("id = ?", chatID).Updates(map[string]interface{}{
			"name":        backupData.ChatInfo.Name,
			"description": backupData.ChatInfo.Description,
			"avatar":      backupData.ChatInfo.Avatar,
			"updated_at":  time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("恢复群组设置失败: %w", err)
		}
	}

	// 恢复公告
	for _, announcement := range backupData.Announcements {
		announcement.ID = 0 // 重置ID，创建新记录
		announcement.ChatID = chatID
		if err := tx.Create(&announcement).Error; err != nil {
			return fmt.Errorf("恢复公告失败: %w", err)
		}
	}

	// 恢复规则
	for _, rule := range backupData.Rules {
		rule.ID = 0 // 重置ID，创建新记录
		rule.ChatID = chatID
		if err := tx.Create(&rule).Error; err != nil {
			return fmt.Errorf("恢复规则失败: %w", err)
		}
	}

	return nil
}

// 辅助方法

// hasPermission 检查用户是否有指定权限
func (s *ChatBackupService) hasPermission(ctx context.Context, chatID uint, userID uint, permission string) bool {
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
	case "can_manage_chat":
		return chatPermission.CanManageChat
	default:
		return false
	}
}

// isChatMember 检查用户是否为群成员
func (s *ChatBackupService) isChatMember(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND is_active = ?", chatID, userID, true).
		Count(&count)
	return count > 0
}

// isChatOwner 检查用户是否为群主
func (s *ChatBackupService) isChatOwner(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND role = ? AND is_active = ?", chatID, userID, "owner", true).
		Count(&count)
	return count > 0
}

// isChatAdmin 检查用户是否为管理员
func (s *ChatBackupService) isChatAdmin(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND role IN (?) AND is_active = ?", chatID, userID, []string{"owner", "admin"}, true).
		Count(&count)
	return count > 0
}

// getDefaultPermission 获取默认权限设置
func (s *ChatBackupService) getDefaultPermission(permission string) bool {
	switch permission {
	case "can_manage_chat":
		return false // 默认只有群主可以管理群组
	default:
		return false
	}
}
