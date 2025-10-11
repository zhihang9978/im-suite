package model

import (
	"time"
	"gorm.io/gorm"
)

// Alert 系统告警模型
type Alert struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	AlertType  string     `json:"alert_type" gorm:"type:varchar(50);not null"` // cpu, memory, disk, database, redis
	Severity   string     `json:"severity" gorm:"type:varchar(20);not null"`   // info, warning, error, critical
	Title      string     `json:"title" gorm:"type:varchar(200);not null"`
	Message    string     `json:"message" gorm:"type:text"`
	Value      float64    `json:"value"`
	Threshold  float64    `json:"threshold"`
	Resolved   bool       `json:"resolved" gorm:"default:false"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy uint       `json:"resolved_by,omitempty"`
}

// AdminOperationLog 管理员操作日志
type AdminOperationLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	AdminID       uint   `json:"admin_id" gorm:"not null"`               // 管理员ID
	AdminUsername string `json:"admin_username" gorm:"type:varchar(50)"` // 管理员用户名
	OperationType string `json:"operation_type" gorm:"type:varchar(50)"` // 操作类型
	TargetType    string `json:"target_type" gorm:"type:varchar(50)"`    // 目标类型: user, group, message, file
	TargetID      uint   `json:"target_id"`                              // 目标ID
	Action        string `json:"action" gorm:"type:varchar(50)"`         // 操作: ban, delete, approve, etc.
	Reason        string `json:"reason" gorm:"type:text"`                // 操作原因
	IPAddress     string `json:"ip_address" gorm:"type:varchar(45)"`     // IP地址
	UserAgent     string `json:"user_agent" gorm:"type:text"`            // 用户代理
	Details       string `json:"details" gorm:"type:text"`               // 详细信息（JSON）
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	Key         string `json:"key" gorm:"index:idx_system_config_key,unique;type:varchar(100);not null"` // 配置键
	Value       string `json:"value" gorm:"type:text"`                            // 配置值
	Category    string `json:"category" gorm:"type:varchar(50)"`                  // 配置分类
	Description string `json:"description" gorm:"type:text"`                      // 配置描述
	IsPublic    bool   `json:"is_public" gorm:"default:false"`                    // 是否公开
	DataType    string `json:"data_type" gorm:"type:varchar(20)"`                 // 数据类型: string, int, bool, json
}

// IPBlacklist IP黑名单
type IPBlacklist struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	IPAddress   string     `json:"ip_address" gorm:"index:idx_ip_blacklist_ip,unique;type:varchar(45);not null"` // IP地址
	Reason      string     `json:"reason" gorm:"type:text"`                                 // 封禁原因
	AddedBy     uint       `json:"added_by"`                                                // 添加管理员ID
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`                                    // 过期时间
	IsPermanent bool       `json:"is_permanent" gorm:"default:false"`                       // 是否永久
}

// UserBlacklist 用户黑名单
type UserBlacklist struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	UserID      uint       `json:"user_id" gorm:"not null;index"`               // 被拉黑的用户ID
	BlockedBy   uint       `json:"blocked_by"`                                  // 操作管理员ID
	Reason      string     `json:"reason" gorm:"type:text"`                     // 封禁原因
	ViolationType string   `json:"violation_type" gorm:"type:varchar(50)"`      // 违规类型
	Severity    string     `json:"severity" gorm:"type:varchar(20)"`            // 严重程度
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`                        // 过期时间
	IsPermanent bool       `json:"is_permanent" gorm:"default:false"`           // 是否永久
	IsActive    bool       `json:"is_active" gorm:"default:true"`               // 是否生效
}

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	
	Phone     string `json:"phone" gorm:"type:varchar(20);index"`      // 手机号
	IPAddress string `json:"ip_address" gorm:"type:varchar(45);index"` // IP地址
	Success   bool   `json:"success" gorm:"default:false"`             // 是否成功
	UserAgent string `json:"user_agent" gorm:"type:text"`              // 用户代理
	ErrorMsg  string `json:"error_msg" gorm:"type:varchar(200)"`       // 错误消息
}

// SuspiciousActivity 可疑活动记录
type SuspiciousActivity struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	UserID        uint   `json:"user_id" gorm:"not null;index"`           // 用户ID
	ActivityType  string `json:"activity_type" gorm:"type:varchar(50)"`   // 活动类型
	Description   string `json:"description" gorm:"type:text"`            // 描述
	Severity      string `json:"severity" gorm:"type:varchar(20)"`        // 严重程度
	IPAddress     string `json:"ip_address" gorm:"type:varchar(45)"`      // IP地址
	RiskScore     float64 `json:"risk_score"`                              // 风险分数
	AutoBlocked   bool   `json:"auto_blocked" gorm:"default:false"`       // 是否自动封禁
	Reviewed      bool   `json:"reviewed" gorm:"default:false"`           // 是否已审核
	ReviewedBy    uint   `json:"reviewed_by,omitempty"`                   // 审核管理员ID
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`                // 审核时间
	ReviewComment string `json:"review_comment" gorm:"type:text"`         // 审核评论
}
