package model

import (
	"time"

	"gorm.io/gorm"
)

// DeviceSession 设备会话
type DeviceSession struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	UserID       uint      `json:"user_id" gorm:"not null;index"`
	DeviceID     string    `json:"device_id" gorm:"not null;index"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	Platform     string    `json:"platform"`
	Browser      string    `json:"browser"`
	Version      string    `json:"version"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent" gorm:"type:text"`
	AccessToken  string    `json:"-" gorm:"type:text"`
	RefreshToken string    `json:"-" gorm:"type:text"`
	LastActiveAt time.Time `json:"last_active_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`

	// 安全信息
	LoginAt       time.Time `json:"login_at"`
	LoginLocation string    `json:"login_location"` // 登录位置（城市）
	IsSuspicious  bool      `json:"is_suspicious" gorm:"default:false"`
	RiskScore     int       `json:"risk_score" gorm:"default:0"` // 风险评分 0-100

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// DeviceActivity 设备活动记录
type DeviceActivity struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	UserID       uint      `json:"user_id" gorm:"not null;index"`
	DeviceID     string    `json:"device_id" gorm:"not null;index"`
	ActivityType string    `json:"activity_type"` // login, logout, token_refresh, action
	Description  string    `json:"description"`
	IP           string    `json:"ip"`
	Location     string    `json:"location"`
	Timestamp    time.Time `json:"timestamp"`

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (DeviceSession) TableName() string {
	return "device_sessions"
}

// TableName 指定表名
func (DeviceActivity) TableName() string {
	return "device_activities"
}
