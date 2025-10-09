package model

import (
	"time"

	"gorm.io/gorm"
)

// TwoFactorAuth 双因子认证记录
type TwoFactorAuth struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	UserID    uint   `json:"user_id" gorm:"not null;index"` // 用户ID
	Method    string `json:"method" gorm:"not null"`        // 验证方式: totp, sms, email, backup_code
	Status    string `json:"status" gorm:"not null"`        // 状态: success, failed
	IP        string `json:"ip"`                            // 验证时的IP地址
	UserAgent string `json:"user_agent" gorm:"type:text"`   // 用户代理
	Device    string `json:"device"`                        // 设备信息

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TrustedDevice 受信任设备
type TrustedDevice struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	UserID         uint       `json:"user_id" gorm:"not null;index"`   // 用户ID
	DeviceID       string     `json:"device_id" gorm:"not null;index"` // 设备唯一标识
	DeviceName     string     `json:"device_name"`                     // 设备名称
	DeviceType     string     `json:"device_type"`                     // 设备类型: web, mobile, desktop
	IP             string     `json:"ip"`                              // 设备IP
	LastUsedAt     time.Time  `json:"last_used_at"`                    // 最后使用时间
	TrustExpiresAt *time.Time `json:"trust_expires_at"`                // 信任过期时间
	IsActive       bool       `json:"is_active" gorm:"default:true"`   // 是否激活

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (TwoFactorAuth) TableName() string {
	return "two_factor_auth"
}

// TableName 指定表名
func (TrustedDevice) TableName() string {
	return "trusted_devices"
}
