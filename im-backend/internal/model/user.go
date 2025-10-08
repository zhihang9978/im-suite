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
	Phone     string `json:"phone" gorm:"uniqueIndex;not null"`           // 手机号
	Username  string `json:"username" gorm:"uniqueIndex"`                 // 用户名
	Nickname  string `json:"nickname"`                                    // 昵称
	Bio       string `json:"bio"`                                         // 个人简介
	Avatar    string `json:"avatar"`                                      // 头像URL
	
	// 认证信息
	Password  string `json:"-" gorm:"not null"`                           // 密码(加密)
	Salt      string `json:"-" gorm:"not null"`                           // 密码盐值
	
	// 状态信息
	IsActive  bool      `json:"is_active" gorm:"default:true"`            // 是否激活
	LastSeen  time.Time `json:"last_seen"`                               // 最后在线时间
	Online    bool      `json:"online" gorm:"default:false"`             // 是否在线
	
	// 设置信息
	Language  string `json:"language" gorm:"default:'zh-CN'"`            // 语言设置
	Theme     string `json:"theme" gorm:"default:'auto'"`                // 主题设置
	
	// 关联关系
	Contacts  []Contact `json:"contacts" gorm:"foreignKey:UserID"`       // 联系人
	Chats     []Chat    `json:"chats" gorm:"many2many:chat_members"`     // 参与的聊天
	Messages  []Message `json:"messages" gorm:"foreignKey:SenderID"`     // 发送的消息
}

// Contact 联系人模型
type Contact struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	UserID      uint   `json:"user_id" gorm:"not null"`                  // 用户ID
	ContactID   uint   `json:"contact_id" gorm:"not null"`               // 联系人ID
	Nickname    string `json:"nickname"`                                 // 联系人昵称
	IsBlocked   bool   `json:"is_blocked" gorm:"default:false"`          // 是否拉黑
	IsMuted     bool   `json:"is_muted" gorm:"default:false"`            // 是否静音
	
	// 关联关系
	User    User `json:"user" gorm:"foreignKey:UserID"`
	Contact User `json:"contact" gorm:"foreignKey:ContactID"`
}

// Chat 聊天模型
type Chat struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// 基本信息
	Name        string `json:"name"`                                      // 聊天名称
	Description string `json:"description"`                               // 聊天描述
	Avatar      string `json:"avatar"`                                    // 聊天头像
	Type        string `json:"type" gorm:"default:'private'"`            // 聊天类型: private/group/channel
	
	// 设置信息
	IsActive    bool   `json:"is_active" gorm:"default:true"`            // 是否激活
	IsPinned    bool   `json:"is_pinned" gorm:"default:false"`           // 是否置顶
	IsMuted     bool   `json:"is_muted" gorm:"default:false"`            // 是否静音
	
	// 关联关系
	Members   []User    `json:"members" gorm:"many2many:chat_members"`   // 成员
	Messages  []Message `json:"messages" gorm:"foreignKey:ChatID"`       // 消息
}

// ChatMember 聊天成员关联表
type ChatMember struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	ChatID    uint   `json:"chat_id" gorm:"not null"`                    // 聊天ID
	UserID    uint   `json:"user_id" gorm:"not null"`                    // 用户ID
	Role      string `json:"role" gorm:"default:'member'"`               // 角色: owner/admin/member
	JoinedAt  time.Time `json:"joined_at"`                              // 加入时间
	
	// 关联关系
	Chat Chat `json:"chat" gorm:"foreignKey:ChatID"`
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Message 消息模型
type Message struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// 基本信息
	ChatID      uint   `json:"chat_id" gorm:"not null"`                  // 聊天ID
	SenderID    uint   `json:"sender_id" gorm:"not null"`                // 发送者ID
	Content     string `json:"content"`                                  // 消息内容
	Type        string `json:"type" gorm:"default:'text'"`               // 消息类型: text/image/video/file/audio/voice
	
	// 文件信息
	FileName    string `json:"file_name"`                                // 文件名
	FileSize    int64  `json:"file_size"`                                // 文件大小
	FileURL     string `json:"file_url"`                                 // 文件URL
	Thumbnail   string `json:"thumbnail"`                                // 缩略图URL
	
	// 消息状态
	IsRead      bool   `json:"is_read" gorm:"default:false"`             // 是否已读
	IsEdited    bool   `json:"is_edited" gorm:"default:false"`           // 是否编辑过
	IsDeleted   bool   `json:"is_deleted" gorm:"default:false"`          // 是否删除
	IsPinned    bool   `json:"is_pinned" gorm:"default:false"`           // 是否置顶
	
	// 特殊功能
	ReplyToID   *uint  `json:"reply_to_id"`                              // 回复的消息ID
	ForwardFrom *uint  `json:"forward_from"`                             // 转发的消息ID
	TTL         int    `json:"ttl"`                                      // 阅后即焚时间(秒)
	SendAt      *time.Time `json:"send_at"`                              // 定时发送时间
	IsSilent    bool   `json:"is_silent" gorm:"default:false"`           // 是否静默发送
	
	// 关联关系
	Chat      Chat   `json:"chat" gorm:"foreignKey:ChatID"`
	Sender    User   `json:"sender" gorm:"foreignKey:SenderID"`
	ReplyTo   *Message `json:"reply_to" gorm:"foreignKey:ReplyToID"`
	ForwardFromMsg *Message `json:"forward_from_msg" gorm:"foreignKey:ForwardFrom"`
}

// MessageRead 消息已读记录
type MessageRead struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	MessageID uint `json:"message_id" gorm:"not null"`                   // 消息ID
	UserID    uint `json:"user_id" gorm:"not null"`                      // 用户ID
	ReadAt    time.Time `json:"read_at"`                                 // 阅读时间
	
	// 关联关系
	Message Message `json:"message" gorm:"foreignKey:MessageID"`
	User    User    `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

func (Contact) TableName() string {
	return "contacts"
}

func (Chat) TableName() string {
	return "chats"
}

func (ChatMember) TableName() string {
	return "chat_members"
}

func (Message) TableName() string {
	return "messages"
}

func (MessageRead) TableName() string {
	return "message_reads"
}


