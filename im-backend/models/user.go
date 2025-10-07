package models

import (
	"time"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Phone     string         `json:"phone" gorm:"uniqueIndex;not null"`
	Username  string         `json:"username" gorm:"uniqueIndex"`
	Nickname  string         `json:"nickname"`
	Avatar    string         `json:"avatar"`
	Bio       string         `json:"bio"`
	Status    string         `json:"status" gorm:"default:'offline'"` // online, offline, away, busy
	LastSeen  *time.Time     `json:"last_seen"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Contact 联系人模型
type Contact struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	ContactID uint           `json:"contact_id" gorm:"not null"`
	Nickname  string         `json:"nickname"` // 备注名
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	User    User `json:"user" gorm:"foreignKey:UserID"`
	Contact User `json:"contact" gorm:"foreignKey:ContactID"`
}

// Chat 聊天模型
type Chat struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Type      string         `json:"type" gorm:"not null"` // private, group, channel
	Title     string         `json:"title"`
	Avatar    string         `json:"avatar"`
	LastMessageID *uint      `json:"last_message_id"`
	LastMessageAt *time.Time `json:"last_message_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Members   []ChatMember `json:"members" gorm:"foreignKey:ChatID"`
	Messages  []Message    `json:"messages" gorm:"foreignKey:ChatID"`
	LastMessage *Message   `json:"last_message" gorm:"foreignKey:LastMessageID"`
}

// ChatMember 聊天成员模型
type ChatMember struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ChatID    uint           `json:"chat_id" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Role      string         `json:"role" gorm:"default:'member'"` // admin, member
	JoinedAt  time.Time      `json:"joined_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Chat Chat `json:"chat" gorm:"foreignKey:ChatID"`
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Message 消息模型
type Message struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ChatID    uint           `json:"chat_id" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Type      string         `json:"type" gorm:"not null"` // text, image, video, audio, file, sticker
	Content   string         `json:"content"`
	MediaURL  string         `json:"media_url"`
	FileSize  int64          `json:"file_size"`
	Duration  int            `json:"duration"` // 音频/视频时长
	ReplyToID *uint          `json:"reply_to_id"`
	ForwardFromID *uint      `json:"forward_from_id"`
	IsEdited  bool           `json:"is_edited" gorm:"default:false"`
	IsDeleted bool           `json:"is_deleted" gorm:"default:false"`
	IsPinned  bool           `json:"is_pinned" gorm:"default:false"`
	TTL       int            `json:"ttl"` // 阅后即焚时间（秒）
	SendAt    *time.Time     `json:"send_at"` // 定时发送
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Chat       Chat    `json:"chat" gorm:"foreignKey:ChatID"`
	User       User    `json:"user" gorm:"foreignKey:UserID"`
	ReplyTo    *Message `json:"reply_to" gorm:"foreignKey:ReplyToID"`
	ForwardFrom *Message `json:"forward_from" gorm:"foreignKey:ForwardFromID"`
}

// MessageRead 消息已读模型
type MessageRead struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	MessageID uint           `json:"message_id" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	ReadAt    time.Time      `json:"read_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Message Message `json:"message" gorm:"foreignKey:MessageID"`
	User    User    `json:"user" gorm:"foreignKey:UserID"`
}

// Session 用户会话模型
type Session struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Token     string         `json:"token" gorm:"uniqueIndex;not null"`
	Device    string         `json:"device"`
	IP        string         `json:"ip"`
	UserAgent string         `json:"user_agent"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}
