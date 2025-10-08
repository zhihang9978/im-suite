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
	ReplyToID     *uint     `gorm:"index" json:"reply_to_id"` // 回复的消息ID
	ForwardFromID *uint     `gorm:"index" json:"forward_from_id"` // 转发的原消息ID
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
