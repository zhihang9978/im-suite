package model

import (
	"time"

	"gorm.io/gorm"
)

// Chat 聊天模型
type Chat struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255)" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Avatar      string    `gorm:"type:varchar(255)" json:"avatar"`
	Type        string    `gorm:"type:varchar(50);default:'private'" json:"type"` // private, group, channel
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Members []ChatMember `gorm:"foreignKey:ChatID" json:"members,omitempty"`
	Messages []Message `gorm:"foreignKey:ChatID" json:"messages,omitempty"`
}

// ChatMember 聊天成员
type ChatMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chat_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Role      string    `gorm:"type:varchar(50);default:'member'" json:"role"` // owner, admin, member
	JoinedAt  time.Time `json:"joined_at"`
	LeftAt    *time.Time `json:"left_at,omitempty"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	InvitedBy *uint     `json:"invited_by,omitempty"` // 邀请者ID
	InviteTime *time.Time `json:"invite_time,omitempty"` // 邀请时间
	LastSeen  time.Time `json:"last_seen"` // 最后活跃时间
	MuteUntil *time.Time `json:"mute_until,omitempty"` // 禁言到期时间
	IsBanned  bool      `gorm:"default:false" json:"is_banned"` // 是否被禁言
	BanReason string    `gorm:"type:varchar(255)" json:"ban_reason"` // 禁言原因

	// 关联
	Chat Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Inviter *User `gorm:"foreignKey:InvitedBy" json:"inviter,omitempty"`
}

// ChatPermission 群组权限配置
type ChatPermission struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	ChatID                uint      `gorm:"not null;index" json:"chat_id"`
	CanSendMessages       bool      `gorm:"default:true" json:"can_send_messages"`
	CanSendMedia          bool      `gorm:"default:true" json:"can_send_media"`
	CanSendStickers       bool      `gorm:"default:true" json:"can_send_stickers"`
	CanSendPolls          bool      `gorm:"default:true" json:"can_send_polls"`
	CanChangeInfo         bool      `gorm:"default:false" json:"can_change_info"`
	CanInviteUsers        bool      `gorm:"default:false" json:"can_invite_users"`
	CanPinMessages        bool      `gorm:"default:false" json:"can_pin_messages"`
	CanDeleteMessages     bool      `gorm:"default:false" json:"can_delete_messages"`
	CanEditMessages       bool      `gorm:"default:false" json:"can_edit_messages"`
	CanManageChat         bool      `gorm:"default:false" json:"can_manage_chat"`
	CanManageVoiceChats   bool      `gorm:"default:false" json:"can_manage_voice_chats"`
	CanRestrictMembers    bool      `gorm:"default:false" json:"can_restrict_members"`
	CanPromoteMembers     bool      `gorm:"default:false" json:"can_promote_members"`
	CanAddAdmins          bool      `gorm:"default:false" json:"can_add_admins"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// 关联
	Chat Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
}

// ChatAnnouncement 群组公告
type ChatAnnouncement struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chat_id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	AuthorID  uint      `gorm:"not null" json:"author_id"`
	IsPinned  bool      `gorm:"default:false" json:"is_pinned"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Chat   Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	Author User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// ChatRule 群组规则
type ChatRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chat_id"`
	RuleNumber int      `gorm:"not null" json:"rule_number"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	AuthorID  uint      `gorm:"not null" json:"author_id"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Chat   Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	Author User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// ChatStatistics 群组统计
type ChatStatistics struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	ChatID                uint      `gorm:"not null;index" json:"chat_id"`
	TotalMembers          int       `gorm:"default:0" json:"total_members"`
	ActiveMembers         int       `gorm:"default:0" json:"active_members"`
	TotalMessages         int       `gorm:"default:0" json:"total_messages"`
	MessagesToday         int       `gorm:"default:0" json:"messages_today"`
	MessagesThisWeek      int       `gorm:"default:0" json:"messages_this_week"`
	MessagesThisMonth     int       `gorm:"default:0" json:"messages_this_month"`
	TotalFiles            int       `gorm:"default:0" json:"total_files"`
	TotalImages           int       `gorm:"default:0" json:"total_images"`
	TotalVideos           int       `gorm:"default:0" json:"total_videos"`
	TotalAudios           int       `gorm:"default:0" json:"total_audios"`
	TotalVoiceCalls       int       `gorm:"default:0" json:"total_voice_calls"`
	TotalVideoCalls       int       `gorm:"default:0" json:"total_video_calls"`
	AverageMessageLength  float64   `gorm:"default:0" json:"average_message_length"`
	PeakActivityHour      int       `gorm:"default:0" json:"peak_activity_hour"`
	LastActivityAt        time.Time `json:"last_activity_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// 关联
	Chat Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
}

// ChatBackup 群组备份
type ChatBackup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ChatID      uint      `gorm:"not null;index" json:"chat_id"`
	BackupType  string    `gorm:"type:varchar(50);not null" json:"backup_type"` // full, messages, media, settings
	BackupData  string    `gorm:"type:longtext" json:"backup_data"` // JSON格式的备份数据
	BackupSize  int64     `gorm:"default:0" json:"backup_size"` // 备份大小（字节）
	CreatedBy   uint      `gorm:"not null" json:"created_by"`
	IsEncrypted bool      `gorm:"default:false" json:"is_encrypted"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	// 关联
	Chat Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
