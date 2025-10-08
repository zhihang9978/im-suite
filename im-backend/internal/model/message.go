package model

import (
	"time"

	"gorm.io/gorm"
)

// Message 消息模型
type Message struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SenderID      uint      `gorm:"not null;index" json:"sender_id"`
	ReceiverID    *uint     `gorm:"index" json:"receiver_id"` // 单聊时使用
	ChatID        *uint     `gorm:"index" json:"chat_id"`     // 群聊时使用
	Content       string    `gorm:"type:text;not null" json:"content"`
	MessageType   string    `gorm:"type:varchar(50);default:'text'" json:"message_type"` // text, image, video, audio, file, etc.
	Status        string    `gorm:"type:varchar(50);default:'sent'" json:"status"`       // sent, delivered, read, recalled, edited
	IsEdited      bool      `gorm:"default:false" json:"is_edited"`
	EditCount     int       `gorm:"default:0" json:"edit_count"`
	IsRecalled    bool      `gorm:"default:false" json:"is_recalled"`
	RecallTime    *time.Time `json:"recall_time"`
	RecallReason  string    `gorm:"type:varchar(255)" json:"recall_reason"`
	IsEncrypted   bool      `gorm:"default:false" json:"is_encrypted"`
	IsSelfDestruct bool     `gorm:"default:false" json:"is_self_destruct"`
	SelfDestructTime *time.Time `json:"self_destruct_time"`
	IsScheduled   bool      `gorm:"default:false" json:"is_scheduled"`
	ScheduledTime *time.Time `json:"scheduled_time"`
	IsSilent      bool      `gorm:"default:false" json:"is_silent"`
	IsPinned      bool      `gorm:"default:false" json:"is_pinned"` // 是否置顶
	PinTime       *time.Time `json:"pin_time"` // 置顶时间
	IsMarked      bool      `gorm:"default:false" json:"is_marked"` // 是否标记
	MarkType      string    `gorm:"type:varchar(50)" json:"mark_type"` // 标记类型：important, favorite, archive
	MarkTime      *time.Time `json:"mark_time"` // 标记时间
	ReplyToID     *uint     `gorm:"index" json:"reply_to_id"` // 回复的消息ID
	ForwardFromID *uint     `gorm:"index" json:"forward_from_id"` // 转发的原消息ID
	ShareCount    int       `gorm:"default:0" json:"share_count"` // 分享次数
	ViewCount     int       `gorm:"default:0" json:"view_count"` // 查看次数
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Sender     User    `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver   *User   `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	Chat       *Chat   `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	ReplyTo    *Message `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
	ForwardFrom *Message `gorm:"foreignKey:ForwardFromID" json:"forward_from,omitempty"`
}

// MessageEdit 消息编辑历史
type MessageEdit struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	OldContent string   `gorm:"type:text" json:"old_content"`
	NewContent string   `gorm:"type:text" json:"new_content"`
	EditTime  time.Time `json:"edit_time"`
	EditReason string   `gorm:"type:varchar(255)" json:"edit_reason"`

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

// MessageRecall 消息撤回记录
type MessageRecall struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	RecallBy  uint      `gorm:"not null" json:"recall_by"` // 撤回操作者ID
	Reason    string    `gorm:"type:varchar(255)" json:"reason"`
	RecallTime time.Time `json:"recall_time"`

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	RecallUser User `gorm:"foreignKey:RecallBy" json:"recall_user,omitempty"`
}

// MessageForward 消息转发记录
type MessageForward struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OriginalMessageID uint `gorm:"not null;index" json:"original_message_id"`
	ForwardMessageID  uint `gorm:"not null;index" json:"forward_message_id"`
	ForwardBy uint      `gorm:"not null" json:"forward_by"`
	ForwardTime time.Time `json:"forward_time"`

	// 关联
	OriginalMessage Message `gorm:"foreignKey:OriginalMessageID" json:"original_message,omitempty"`
	ForwardMessage  Message `gorm:"foreignKey:ForwardMessageID" json:"forward_message,omitempty"`
	ForwardUser     User    `gorm:"foreignKey:ForwardBy" json:"forward_user,omitempty"`
}

// ScheduledMessage 定时消息
type ScheduledMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	ScheduledBy uint    `gorm:"not null" json:"scheduled_by"`
	ScheduledTime time.Time `gorm:"not null" json:"scheduled_time"`
	IsExecuted bool     `gorm:"default:false" json:"is_executed"`
	ExecuteTime *time.Time `json:"execute_time"`
	IsCancelled bool     `gorm:"default:false" json:"is_cancelled"`
	CancelTime *time.Time `json:"cancel_time"`
	CancelReason string  `gorm:"type:varchar(255)" json:"cancel_reason"`

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	ScheduledUser User `gorm:"foreignKey:ScheduledBy" json:"scheduled_user,omitempty"`
}

// MessageSearchIndex 消息搜索索引
type MessageSearchIndex struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	Content   string    `gorm:"type:text" json:"content"`
	Keywords  string    `gorm:"type:text" json:"keywords"` // 分词后的关键词
	Language  string    `gorm:"type:varchar(10);default:'zh'" json:"language"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

// MessagePin 消息置顶记录
type MessagePin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	PinnedBy  uint      `gorm:"not null" json:"pinned_by"` // 置顶操作者ID
	PinTime   time.Time `json:"pin_time"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	UnpinTime *time.Time `json:"unpin_time"`
	UnpinBy   *uint     `json:"unpin_by"` // 取消置顶操作者ID

	// 关联
	Message   Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	PinUser   User    `gorm:"foreignKey:PinnedBy" json:"pin_user,omitempty"`
	UnpinUser *User   `gorm:"foreignKey:UnpinBy" json:"unpin_user,omitempty"`
}

// MessageMark 消息标记记录
type MessageMark struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	MarkedBy  uint      `gorm:"not null" json:"marked_by"` // 标记操作者ID
	MarkType  string    `gorm:"type:varchar(50);not null" json:"mark_type"` // important, favorite, archive, etc.
	MarkTime  time.Time `json:"mark_time"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	UnmarkTime *time.Time `json:"unmark_time"`
	UnmarkBy  *uint     `json:"unmark_by"` // 取消标记操作者ID

	// 关联
	Message   Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	MarkUser  User    `gorm:"foreignKey:MarkedBy" json:"mark_user,omitempty"`
	UnmarkUser *User  `gorm:"foreignKey:UnmarkBy" json:"unmark_user,omitempty"`
}

// MessageStatus 消息状态追踪
type MessageStatus struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"` // 用户ID
	Status    string    `gorm:"type:varchar(50);not null" json:"status"` // sent, delivered, read, etc.
	StatusTime time.Time `json:"status_time"`
	DeviceID  string    `gorm:"type:varchar(255)" json:"device_id"` // 设备ID
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"` // IP地址

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// MessageShare 消息分享记录
type MessageShare struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	SharedBy  uint      `gorm:"not null" json:"shared_by"` // 分享者ID
	SharedTo  *uint     `gorm:"index" json:"shared_to"` // 分享给的用户ID（私聊）
	SharedToChatID *uint `gorm:"index" json:"shared_to_chat_id"` // 分享到的群聊ID
	ShareType string    `gorm:"type:varchar(50);default:'copy'" json:"share_type"` // copy, forward, link
	ShareTime time.Time `json:"share_time"`
	ShareData string    `gorm:"type:text" json:"share_data"` // 分享的额外数据（如链接、备注等）

	// 关联
	Message     Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	ShareUser   User    `gorm:"foreignKey:SharedBy" json:"share_user,omitempty"`
	SharedToUser *User  `gorm:"foreignKey:SharedTo" json:"shared_to_user,omitempty"`
	SharedToChat *Chat  `gorm:"foreignKey:SharedToChatID" json:"shared_to_chat,omitempty"`
}

// MessageReply 消息回复链
type MessageReply struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"` // 回复的消息ID
	ReplyToID uint      `gorm:"not null;index" json:"reply_to_id"` // 被回复的消息ID
	ReplyLevel int      `gorm:"default:1" json:"reply_level"` // 回复层级深度
	ReplyPath  string   `gorm:"type:text" json:"reply_path"` // 回复路径，如 "1,2,3"
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	ReplyTo Message `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
}

// MessageRead 消息已读记录
type MessageRead struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"` // 消息ID
	UserID    uint      `gorm:"not null;index" json:"user_id"`    // 用户ID
	ReadAt    time.Time `gorm:"not null" json:"read_at"`          // 阅读时间
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
