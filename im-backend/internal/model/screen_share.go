package model

import (
	"time"

	"gorm.io/gorm"
)

// ScreenShareSession 屏幕共享会话记录
type ScreenShareSession struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	CallID           string     `json:"call_id" gorm:"index;not null"`   // 通话ID
	SharerID         uint       `json:"sharer_id" gorm:"index;not null"` // 共享者ID
	SharerName       string     `json:"sharer_name"`                     // 共享者名称
	StartTime        time.Time  `json:"start_time"`                      // 开始时间
	EndTime          *time.Time `json:"end_time"`                        // 结束时间
	Duration         int64      `json:"duration"`                        // 时长（秒）
	Quality          string     `json:"quality"`                         // 质量等级
	WithAudio        bool       `json:"with_audio"`                      // 是否包含音频
	InitialQuality   string     `json:"initial_quality"`                 // 初始质量
	QualityChanges   int        `json:"quality_changes"`                 // 质量调整次数
	ParticipantCount int        `json:"participant_count"`               // 参与者数量
	EndReason        string     `json:"end_reason"`                      // 结束原因
	Status           string     `json:"status" gorm:"default:'active'"`  // active, ended, interrupted

	// 关联关系
	Sharer User `json:"sharer" gorm:"foreignKey:SharerID"`
}

// ScreenShareQualityChange 屏幕共享质量变更记录
type ScreenShareQualityChange struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	SessionID    uint      `json:"session_id" gorm:"index;not null"` // 会话ID
	FromQuality  string    `json:"from_quality"`                     // 原质量
	ToQuality    string    `json:"to_quality"`                       // 新质量
	ChangeTime   time.Time `json:"change_time"`                      // 变更时间
	ChangeReason string    `json:"change_reason"`                    // 变更原因: manual, auto_network, auto_cpu
	NetworkSpeed float64   `json:"network_speed"`                    // 当时网速 (Kbps)
	CPUUsage     float64   `json:"cpu_usage"`                        // 当时CPU使用率

	// 关联关系
	Session ScreenShareSession `json:"session" gorm:"foreignKey:SessionID"`
}

// ScreenShareParticipant 屏幕共享参与者
type ScreenShareParticipant struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	SessionID    uint       `json:"session_id" gorm:"index;not null"` // 会话ID
	UserID       uint       `json:"user_id" gorm:"index;not null"`    // 用户ID
	UserName     string     `json:"user_name"`                        // 用户名
	JoinTime     time.Time  `json:"join_time"`                        // 加入时间
	LeaveTime    *time.Time `json:"leave_time"`                       // 离开时间
	ViewDuration int64      `json:"view_duration"`                    // 观看时长（秒）

	// 关联关系
	Session ScreenShareSession `json:"session" gorm:"foreignKey:SessionID"`
	User    User               `json:"user" gorm:"foreignKey:UserID"`
}

// ScreenShareStatistics 屏幕共享统计
type ScreenShareStatistics struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID             uint       `json:"user_id" gorm:"index:idx_screen_share_stats_user,unique;not null"` // 用户ID
	TotalSessions      int64      `json:"total_sessions"`                                                   // 总共享次数
	TotalDuration      int64      `json:"total_duration"`                                                   // 总共享时长（秒）
	AverageDuration    float64    `json:"average_duration"`                                                 // 平均时长（秒）
	TotalParticipants  int64      `json:"total_participants"`                                               // 总参与人次
	HighQualityCount   int64      `json:"high_quality_count"`                                               // 高清次数
	MediumQualityCount int64      `json:"medium_quality_count"`                                             // 标准次数
	LowQualityCount    int64      `json:"low_quality_count"`                                                // 流畅次数
	WithAudioCount     int64      `json:"with_audio_count"`                                                 // 包含音频次数
	LastShareTime      *time.Time `json:"last_share_time"`                                                  // 最后共享时间

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// ScreenShareRecording 屏幕共享录制
type ScreenShareRecording struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	SessionID     uint       `json:"session_id" gorm:"index;not null"`  // 会话ID
	RecorderID    uint       `json:"recorder_id" gorm:"index;not null"` // 录制者ID
	FileName      string     `json:"file_name"`                         // 文件名
	FilePath      string     `json:"file_path"`                         // 文件路径
	FileSize      int64      `json:"file_size"`                         // 文件大小（字节）
	Duration      int64      `json:"duration"`                          // 录制时长（秒）
	Format        string     `json:"format" gorm:"default:'webm'"`      // 格式: webm, mp4
	Quality       string     `json:"quality"`                           // 质量
	StartTime     time.Time  `json:"start_time"`                        // 开始时间
	EndTime       *time.Time `json:"end_time"`                          // 结束时间
	Status        string     `json:"status" gorm:"default:'recording'"` // recording, completed, failed
	DownloadCount int        `json:"download_count" gorm:"default:0"`   // 下载次数

	// 关联关系
	Session  ScreenShareSession `json:"session" gorm:"foreignKey:SessionID"`
	Recorder User               `json:"recorder" gorm:"foreignKey:RecorderID"`
}

// TableName 指定表名
func (ScreenShareSession) TableName() string {
	return "screen_share_sessions"
}

func (ScreenShareQualityChange) TableName() string {
	return "screen_share_quality_changes"
}

func (ScreenShareParticipant) TableName() string {
	return "screen_share_participants"
}

func (ScreenShareStatistics) TableName() string {
	return "screen_share_statistics"
}

func (ScreenShareRecording) TableName() string {
	return "screen_share_recordings"
}
