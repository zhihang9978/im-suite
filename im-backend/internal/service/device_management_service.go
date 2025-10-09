package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/gorm"
)

// DeviceManagementService 设备管理服务
type DeviceManagementService struct {
	db *gorm.DB
}

// NewDeviceManagementService 创建设备管理服务实例
func NewDeviceManagementService() *DeviceManagementService {
	return &DeviceManagementService{
		db: config.DB,
	}
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"` // web, mobile, desktop
	Platform   string `json:"platform"`    // Windows, macOS, Linux, iOS, Android
	Browser    string `json:"browser"`     // Chrome, Firefox, Safari, etc.
	Version    string `json:"version"`     // 版本号
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
}

// RegisterDevice 注册设备
func (s *DeviceManagementService) RegisterDevice(ctx context.Context, userID uint, deviceInfo *DeviceInfo) (*model.DeviceSession, error) {
	// 生成设备ID（如果未提供）
	if deviceInfo.DeviceID == "" {
		deviceInfo.DeviceID = s.generateDeviceID(deviceInfo)
	}

	// 检查设备是否已注册
	var existing model.DeviceSession
	err := s.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", userID, deviceInfo.DeviceID).First(&existing).Error

	if err == nil {
		// 设备已存在，更新信息
		existing.LastActiveAt = time.Now()
		existing.IP = deviceInfo.IP
		existing.IsActive = true
		s.db.WithContext(ctx).Save(&existing)

		// 记录活动
		s.recordActivity(ctx, userID, deviceInfo.DeviceID, "device_reactivated", "设备重新激活", deviceInfo.IP, "")

		return &existing, nil
	}

	// 创建新设备会话
	session := model.DeviceSession{
		UserID:       userID,
		DeviceID:     deviceInfo.DeviceID,
		DeviceName:   deviceInfo.DeviceName,
		DeviceType:   deviceInfo.DeviceType,
		Platform:     deviceInfo.Platform,
		Browser:      deviceInfo.Browser,
		Version:      deviceInfo.Version,
		IP:           deviceInfo.IP,
		UserAgent:    deviceInfo.UserAgent,
		LastActiveAt: time.Now(),
		LoginAt:      time.Now(),
		ExpiresAt:    time.Now().AddDate(0, 1, 0), // 1个月后过期
		IsActive:     true,
		RiskScore:    0,
	}

	// 计算风险评分
	session.RiskScore = s.calculateRiskScore(ctx, userID, deviceInfo)
	session.IsSuspicious = session.RiskScore > 70

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, fmt.Errorf("注册设备失败: %w", err)
	}

	// 记录活动
	s.recordActivity(ctx, userID, deviceInfo.DeviceID, "device_registered", "新设备注册", deviceInfo.IP, "")

	return &session, nil
}

// GetUserDevices 获取用户的所有设备
func (s *DeviceManagementService) GetUserDevices(ctx context.Context, userID uint) ([]DeviceSession, error) {
	var devices []DeviceSession
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_active_at DESC").
		Find(&devices).Error
	return devices, err
}

// GetDeviceByID 获取设备详情
func (s *DeviceManagementService) GetDeviceByID(ctx context.Context, userID uint, deviceID string) (*DeviceSession, error) {
	var device DeviceSession
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		First(&device).Error
	if err != nil {
		return nil, fmt.Errorf("设备不存在: %w", err)
	}
	return &device, nil
}

// UpdateDeviceActivity 更新设备活动
func (s *DeviceManagementService) UpdateDeviceActivity(ctx context.Context, userID uint, deviceID string, ip string) error {
	var device DeviceSession
	err := s.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", userID, deviceID).First(&device).Error
	if err != nil {
		return fmt.Errorf("设备不存在: %w", err)
	}

	device.LastActiveAt = time.Now()
	device.IP = ip

	return s.db.WithContext(ctx).Save(&device).Error
}

// RevokeDevice 撤销设备（强制下线）
func (s *DeviceManagementService) RevokeDevice(ctx context.Context, userID uint, deviceID string) error {
	var device DeviceSession
	err := s.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", userID, deviceID).First(&device).Error
	if err != nil {
		return fmt.Errorf("设备不存在: %w", err)
	}

	device.IsActive = false
	if err := s.db.WithContext(ctx).Save(&device).Error; err != nil {
		return err
	}

	// 记录活动
	s.recordActivity(ctx, userID, deviceID, "device_revoked", "设备被撤销", device.IP, "")

	return nil
}

// RevokeAllDevices 撤销用户的所有设备（除当前设备）
func (s *DeviceManagementService) RevokeAllDevices(ctx context.Context, userID uint, exceptDeviceID string) error {
	result := s.db.WithContext(ctx).
		Model(&DeviceSession{}).
		Where("user_id = ? AND device_id != ? AND is_active = ?", userID, exceptDeviceID, true).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	// 记录活动
	s.recordActivity(ctx, userID, "all", "all_devices_revoked",
		fmt.Sprintf("撤销了%d台设备", result.RowsAffected), "")

	return nil
}

// GetDeviceActivities 获取设备活动历史
func (s *DeviceManagementService) GetDeviceActivities(ctx context.Context, userID uint, deviceID string, limit int) ([]DeviceActivity, error) {
	var activities []DeviceActivity
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)

	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("timestamp DESC").Find(&activities).Error
	return activities, err
}

// GetSuspiciousDevices 获取可疑设备
func (s *DeviceManagementService) GetSuspiciousDevices(ctx context.Context, userID uint) ([]DeviceSession, error) {
	var devices []DeviceSession
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_suspicious = ? AND is_active = ?", userID, true, true).
		Order("risk_score DESC").
		Find(&devices).Error
	return devices, err
}

// GetDeviceStatistics 获取设备统计信息
func (s *DeviceManagementService) GetDeviceStatistics(ctx context.Context, userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 活跃设备数
	var activeCount int64
	s.db.WithContext(ctx).Model(&DeviceSession{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&activeCount)
	stats["active_devices"] = activeCount

	// 可疑设备数
	var suspiciousCount int64
	s.db.WithContext(ctx).Model(&DeviceSession{}).
		Where("user_id = ? AND is_suspicious = ? AND is_active = ?", userID, true, true).
		Count(&suspiciousCount)
	stats["suspicious_devices"] = suspiciousCount

	// 按类型统计
	var typeStats []struct {
		DeviceType string
		Count      int64
	}
	s.db.WithContext(ctx).Model(&DeviceSession{}).
		Select("device_type, count(*) as count").
		Where("user_id = ? AND is_active = ?", userID, true).
		Group("device_type").
		Scan(&typeStats)
	stats["by_type"] = typeStats

	// 最近活动
	var lastActivity DeviceActivity
	s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		First(&lastActivity)
	stats["last_activity"] = lastActivity

	return stats, nil
}

// generateDeviceID 生成设备ID
func (s *DeviceManagementService) generateDeviceID(info *DeviceInfo) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%s",
		info.DeviceType, info.Platform, info.Browser, info.Version, info.UserAgent)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// calculateRiskScore 计算风险评分
func (s *DeviceManagementService) calculateRiskScore(ctx context.Context, userID uint, deviceInfo *DeviceInfo) int {
	score := 0

	// 1. 检查是否是新IP
	var count int64
	s.db.WithContext(ctx).Model(&DeviceSession{}).
		Where("user_id = ? AND ip = ?", userID, deviceInfo.IP).
		Count(&count)
	if count == 0 {
		score += 20 // 新IP +20分
	}

	// 2. 检查是否是新设备类型
	s.db.WithContext(ctx).Model(&DeviceSession{}).
		Where("user_id = ? AND device_type = ?", userID, deviceInfo.DeviceType).
		Count(&count)
	if count == 0 {
		score += 15 // 新设备类型 +15分
	}

	// 3. 检查是否是新平台
	s.db.WithContext(ctx).Model(&DeviceSession{}).
		Where("user_id = ? AND platform = ?", userID, deviceInfo.Platform).
		Count(&count)
	if count == 0 {
		score += 10 // 新平台 +10分
	}

	// 4. 检查最近登录时间
	var lastDevice DeviceSession
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_active_at DESC").
		First(&lastDevice).Error
	if err == nil {
		timeSinceLastLogin := time.Since(lastDevice.LastActiveAt)
		if timeSinceLastLogin < 5*time.Minute {
			score += 25 // 5分钟内从不同设备登录 +25分
		}
	}

	// 5. 检查地理位置变化（简化版）
	// 实际应该使用IP地理位置库
	if err == nil && lastDevice.IP != deviceInfo.IP {
		score += 15 // IP变化 +15分
	}

	return score
}

// recordActivity 记录设备活动
func (s *DeviceManagementService) recordActivity(ctx context.Context, userID uint, deviceID string, activityType string, description string, ip string, location string) {
	activity := DeviceActivity{
		UserID:       userID,
		DeviceID:     deviceID,
		ActivityType: activityType,
		Description:  description,
		IP:           ip,
		Location:     location,
		Timestamp:    time.Now(),
	}
	s.db.WithContext(ctx).Create(&activity)
}

// ExportDeviceData 导出设备数据（GDPR合规）
func (s *DeviceManagementService) ExportDeviceData(ctx context.Context, userID uint) (string, error) {
	// 获取所有设备
	var devices []DeviceSession
	s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&devices)

	// 获取所有活动
	var activities []DeviceActivity
	s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&activities)

	// 组装数据
	data := map[string]interface{}{
		"devices":     devices,
		"activities":  activities,
		"exported_at": time.Now(),
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("导出失败: %w", err)
	}

	return string(jsonData), nil
}

// CleanupExpiredDevices 清理过期设备
func (s *DeviceManagementService) CleanupExpiredDevices(ctx context.Context) error {
	// 标记过期设备为不活跃
	result := s.db.WithContext(ctx).
		Model(&DeviceSession{}).
		Where("expires_at < ? AND is_active = ?", time.Now(), true).
		Update("is_active", false)

	return result.Error
}
