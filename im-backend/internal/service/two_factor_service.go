package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// TwoFactorService 双因子认证服务
type TwoFactorService struct {
	db *gorm.DB
}

// NewTwoFactorService 创建双因子认证服务实例
func NewTwoFactorService() *TwoFactorService {
	return &TwoFactorService{
		db: config.DB,
	}
}

// check2FAFailedAttempts 检查2FA验证失败次数
func (s *TwoFactorService) check2FAFailedAttempts(ctx context.Context, userID uint) error {
	// 使用Redis记录失败次数（如果Redis可用）
	redis := config.GetRedis()
	if redis != nil {
		key := fmt.Sprintf("2fa:failed:%d", userID)
		count, _ := redis.Get(ctx, key).Int()
		
		// 超过5次失败，锁定10分钟
		if count >= 5 {
			ttl, _ := redis.TTL(ctx, key).Result()
			return fmt.Errorf("验证失败次数过多，请在%d秒后重试", int(ttl.Seconds()))
		}
	}
	return nil
}

// record2FAFailedAttempt 记录2FA验证失败
func (s *TwoFactorService) record2FAFailedAttempt(ctx context.Context, userID uint) {
	redis := config.GetRedis()
	if redis != nil {
		key := fmt.Sprintf("2fa:failed:%d", userID)
		redis.Incr(ctx, key)
		redis.Expire(ctx, key, 10*time.Minute) // 10分钟过期
	}
}

// reset2FAFailedAttempts 重置2FA验证失败次数
func (s *TwoFactorService) reset2FAFailedAttempts(ctx context.Context, userID uint) {
	redis := config.GetRedis()
	if redis != nil {
		key := fmt.Sprintf("2fa:failed:%d", userID)
		redis.Del(ctx, key)
	}
}

// EnableTwoFactorRequest 启用2FA请求
type EnableTwoFactorRequest struct {
	Password string `json:"password" binding:"required"` // 需要密码验证
}

// EnableTwoFactorResponse 启用2FA响应
type EnableTwoFactorResponse struct {
	Secret      string   `json:"secret"`       // TOTP密钥
	QRCode      string   `json:"qr_code"`      // 二维码URL
	BackupCodes []string `json:"backup_codes"` // 备用码
}

// VerifyTwoFactorRequest 验证2FA请求
type VerifyTwoFactorRequest struct {
	Code string `json:"code" binding:"required"` // 6位验证码
}

// DisableTwoFactorRequest 禁用2FA请求
type DisableTwoFactorRequest struct {
	Password string `json:"password" binding:"required"` // 需要密码验证
	Code     string `json:"code" binding:"required"`     // 2FA验证码
}

// ValidateTwoFactorRequest 验证2FA码请求
type ValidateTwoFactorRequest struct {
	UserID uint   `json:"user_id"`
	Code   string `json:"code" binding:"required"`
}

// RegeneratBackupCodesRequest 重新生成备用码请求
type RegeneratBackupCodesRequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// EnableTwoFactor 启用双因子认证
func (s *TwoFactorService) EnableTwoFactor(ctx context.Context, userID uint, req *EnableTwoFactorRequest) (*EnableTwoFactorResponse, error) {
	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 验证密码
	authService := NewAuthService()
	if err := authService.verifyPassword(user.Password, req.Password); err != nil {
		return nil, errors.New("密码错误")
	}

	// 检查是否已启用
	if user.TwoFactorEnabled {
		return nil, errors.New("双因子认证已启用")
	}

	// 生成TOTP密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "志航密信",
		AccountName: user.Username,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("生成TOTP密钥失败: %w", err)
	}

	// 生成备用码
	backupCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("生成备用码失败: %w", err)
	}

	// 保存到数据库（先不启用，等待验证）
	backupCodesJSON, _ := json.Marshal(backupCodes)
	user.TwoFactorSecret = key.Secret()
	user.BackupCodes = string(backupCodesJSON)
	user.TwoFactorEnabled = false // 等待验证后再启用

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("保存失败: %w", err)
	}

	return &EnableTwoFactorResponse{
		Secret:      key.Secret(),
		QRCode:      key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyAndEnableTwoFactor 验证并启用2FA
func (s *TwoFactorService) VerifyAndEnableTwoFactor(ctx context.Context, userID uint, req *VerifyTwoFactorRequest) error {
	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 验证TOTP码
	valid := totp.Validate(req.Code, user.TwoFactorSecret)
	if !valid {
		// 记录失败
		s.recordTwoFactorAttempt(ctx, userID, "totp", "failed", "")
		return errors.New("验证码错误")
	}

	// 启用2FA
	user.TwoFactorEnabled = true
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("启用失败: %w", err)
	}

	// 记录成功
	s.recordTwoFactorAttempt(ctx, userID, "totp", "success", "")

	return nil
}

// DisableTwoFactor 禁用双因子认证
func (s *TwoFactorService) DisableTwoFactor(ctx context.Context, userID uint, req *DisableTwoFactorRequest) error {
	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 检查是否已启用
	if !user.TwoFactorEnabled {
		return errors.New("双因子认证未启用")
	}

	// 验证密码
	authService := NewAuthService()
	if err := authService.verifyPassword(user.Password, req.Password); err != nil {
		return errors.New("密码错误")
	}

	// 验证2FA码
	if err := s.ValidateTwoFactorCode(ctx, userID, req.Code); err != nil {
		return err
	}

	// 禁用2FA
	user.TwoFactorEnabled = false
	user.TwoFactorSecret = ""
	user.BackupCodes = ""

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("禁用失败: %w", err)
	}

	// 删除所有受信任设备
	s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.TrustedDevice{})

	return nil
}

// ValidateTwoFactorCode 验证2FA码（用于登录）
func (s *TwoFactorService) ValidateTwoFactorCode(ctx context.Context, userID uint, code string) error {
	// 检查失败次数限制
	if err := s.check2FAFailedAttempts(ctx, userID); err != nil {
		return err
	}
	
	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 检查是否启用2FA
	if !user.TwoFactorEnabled {
		return errors.New("双因子认证未启用")
	}

	// 验证TOTP码
	valid := totp.Validate(code, user.TwoFactorSecret)
	if valid {
		s.recordTwoFactorAttempt(ctx, userID, "totp", "success", "")
		s.reset2FAFailedAttempts(ctx, userID) // 重置失败次数
		return nil
	}

	// 尝试备用码
	var backupCodes []string
	if err := json.Unmarshal([]byte(user.BackupCodes), &backupCodes); err == nil {
		for i, backupCode := range backupCodes {
			if backupCode == code && backupCode != "" {
				// 备用码匹配，使用后作废
				backupCodes[i] = ""
				backupCodesJSON, _ := json.Marshal(backupCodes)
				user.BackupCodes = string(backupCodesJSON)
				s.db.WithContext(ctx).Save(&user)

				s.recordTwoFactorAttempt(ctx, userID, "backup_code", "success", "")
				s.reset2FAFailedAttempts(ctx, userID) // 重置失败次数
				return nil
			}
		}
	}

	// 验证失败
	s.recordTwoFactorAttempt(ctx, userID, "totp", "failed", "")
	s.record2FAFailedAttempt(ctx, userID) // 记录失败次数
	return errors.New("验证码错误")
}

// RegenerateBackupCodes 重新生成备用码
func (s *TwoFactorService) RegenerateBackupCodes(ctx context.Context, userID uint, req *RegeneratBackupCodesRequest) ([]string, error) {
	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 验证密码
	authService := NewAuthService()
	if err := authService.verifyPassword(user.Password, req.Password); err != nil {
		return nil, errors.New("密码错误")
	}

	// 验证2FA码
	if err := s.ValidateTwoFactorCode(ctx, userID, req.Code); err != nil {
		return nil, err
	}

	// 生成新备用码
	backupCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("生成备用码失败: %w", err)
	}

	// 保存
	backupCodesJSON, _ := json.Marshal(backupCodes)
	user.BackupCodes = string(backupCodesJSON)

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("保存失败: %w", err)
	}

	return backupCodes, nil
}

// GetTwoFactorStatus 获取2FA状态
func (s *TwoFactorService) GetTwoFactorStatus(ctx context.Context, userID uint) (map[string]interface{}, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 统计受信任设备
	var deviceCount int64
	s.db.WithContext(ctx).Model(&model.TrustedDevice{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&deviceCount)

	// 统计最近验证记录
	var recentAttempts []model.TwoFactorAuth
	s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&recentAttempts)

	// 检查备用码剩余数量
	backupCodesRemaining := 0
	if user.BackupCodes != "" {
		var backupCodes []string
		if err := json.Unmarshal([]byte(user.BackupCodes), &backupCodes); err == nil {
			for _, code := range backupCodes {
				if code != "" {
					backupCodesRemaining++
				}
			}
		}
	}

	return map[string]interface{}{
		"enabled":                user.TwoFactorEnabled,
		"trusted_devices_count":  deviceCount,
		"backup_codes_remaining": backupCodesRemaining,
		"recent_attempts":        recentAttempts,
	}, nil
}

// AddTrustedDevice 添加受信任设备
func (s *TwoFactorService) AddTrustedDevice(ctx context.Context, userID uint, deviceID, deviceName, deviceType, ip string) error {
	// 检查设备是否已存在
	var existing model.TrustedDevice
	err := s.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", userID, deviceID).First(&existing).Error

	if err == nil {
		// 已存在，更新最后使用时间
		existing.LastUsedAt = time.Now()
		existing.IsActive = true
		return s.db.WithContext(ctx).Save(&existing).Error
	}

	// 创建新设备
	device := model.TrustedDevice{
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		DeviceType: deviceType,
		IP:         ip,
		LastUsedAt: time.Now(),
		IsActive:   true,
		TrustExpiresAt: func() *time.Time {
			t := time.Now().AddDate(0, 0, 30) // 30天后过期
			return &t
		}(),
	}

	return s.db.WithContext(ctx).Create(&device).Error
}

// RemoveTrustedDevice 移除受信任设备
func (s *TwoFactorService) RemoveTrustedDevice(ctx context.Context, userID uint, deviceID string) error {
	return s.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&model.TrustedDevice{}).Error
}

// IsDeviceTrusted 检查设备是否受信任
func (s *TwoFactorService) IsDeviceTrusted(ctx context.Context, userID uint, deviceID string) bool {
	var device model.TrustedDevice
	err := s.db.WithContext(ctx).Where("user_id = ? AND device_id = ? AND is_active = ?", userID, deviceID, true).First(&device).Error

	if err != nil {
		return false
	}

	// 检查是否过期
	if device.TrustExpiresAt != nil && device.TrustExpiresAt.Before(time.Now()) {
		// 已过期，标记为不活跃
		device.IsActive = false
		s.db.WithContext(ctx).Save(&device)
		return false
	}

	return true
}

// GetTrustedDevices 获取受信任设备列表
func (s *TwoFactorService) GetTrustedDevices(ctx context.Context, userID uint) ([]model.TrustedDevice, error) {
	var devices []model.TrustedDevice
	err := s.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Order("last_used_at DESC").Find(&devices).Error
	return devices, err
}

// generateBackupCodes 生成备用码
func (s *TwoFactorService) generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// 生成8位随机码
		b := make([]byte, 5)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		code := base32.StdEncoding.EncodeToString(b)
		code = strings.TrimRight(code, "=")
		codes[i] = code[:8]
	}
	return codes, nil
}

// recordTwoFactorAttempt 记录2FA验证尝试
func (s *TwoFactorService) recordTwoFactorAttempt(ctx context.Context, userID uint, method, status, ip string) {
	attempt := model.TwoFactorAuth{
		UserID: userID,
		Method: method,
		Status: status,
		IP:     ip,
	}
	s.db.WithContext(ctx).Create(&attempt)
}
