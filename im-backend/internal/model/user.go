package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 基本信息
	Phone    string `json:"phone" gorm:"uniqueIndex;not null"` // 手机号
	Username string `json:"username" gorm:"uniqueIndex"`       // 用户名
	Nickname string `json:"nickname"`                          // 昵称
	Bio      string `json:"bio"`                               // 个人简介
	Avatar   string `json:"avatar"`                            // 头像URL

	// 认证信息
	Password string `json:"-" gorm:"not null"` // 密码(加密)
	Salt     string `json:"-" gorm:"not null"` // 密码盐值

	// 双因子认证
	TwoFactorEnabled bool   `json:"two_factor_enabled" gorm:"default:false"` // 是否启用2FA
	TwoFactorSecret  string `json:"-" gorm:"type:text"`                      // TOTP密钥
	BackupCodes      string `json:"-" gorm:"type:text"`                      // 备用码(JSON数组)

	// 状态信息
	IsActive bool      `json:"is_active" gorm:"default:true"` // 是否激活
	LastSeen time.Time `json:"last_seen"`                     // 最后在线时间
	Online   bool      `json:"online" gorm:"default:false"`   // 是否在线

	// 权限信息
	Role      string     `json:"role" gorm:"default:'user'"`     // 用户角色: user, admin, super_admin
	IsBanned  bool       `json:"is_banned" gorm:"default:false"` // 是否被封禁
	BanUntil  *time.Time `json:"ban_until,omitempty"`            // 封禁到期时间
	BanReason string     `json:"ban_reason,omitempty"`           // 封禁原因

	// 设置信息
	Language string `json:"language" gorm:"default:'zh-CN'"` // 语言设置
	Theme    string `json:"theme" gorm:"default:'auto'"`     // 主题设置

	// 机器人管理信息
	CreatedByBotID *uint `json:"created_by_bot_id,omitempty"` // 创建该用户的机器人ID（null表示非机器人创建）
	BotManageable  bool  `json:"bot_manageable" gorm:"default:false"` // 是否允许机器人管理

	// 关联关系
	Contacts []Contact `json:"contacts" gorm:"foreignKey:UserID"`   // 联系人
	Chats    []Chat    `json:"chats" gorm:"many2many:chat_members"` // 参与的聊天
	Messages []Message `json:"messages" gorm:"foreignKey:SenderID"` // 发送的消息
}

// Contact 联系人模型
type Contact struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	UserID    uint   `json:"user_id" gorm:"not null"`         // 用户ID
	ContactID uint   `json:"contact_id" gorm:"not null"`      // 联系人ID
	Nickname  string `json:"nickname"`                        // 联系人昵称
	IsBlocked bool   `json:"is_blocked" gorm:"default:false"` // 是否拉黑
	IsMuted   bool   `json:"is_muted" gorm:"default:false"`   // 是否静音

	// 关联关系
	User    User `json:"user" gorm:"foreignKey:UserID"`
	Contact User `json:"contact" gorm:"foreignKey:ContactID"`
}

// Session 用户会话模型
type Session struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	UserID    uint      `json:"user_id" gorm:"not null"`           // 用户ID
	Token     string    `json:"token" gorm:"uniqueIndex;not null"` // 会话令牌
	Device    string    `json:"device"`                            // 设备信息
	IP        string    `json:"ip"`                                // IP地址
	UserAgent string    `json:"user_agent"`                        // 用户代理
	ExpiresAt time.Time `json:"expires_at"`                        // 过期时间

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}
