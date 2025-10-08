package model

import (
	"time"

	"gorm.io/gorm"
)

// ContentReport 内容举报模型
type ContentReport struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 举报内容信息
	ContentType   string `gorm:"type:varchar(50);not null" json:"content_type"` // message, user, chat, file
	ContentID     uint   `gorm:"not null;index" json:"content_id"`              // 被举报内容的ID
	ContentText   string `gorm:"type:text" json:"content_text"`                 // 内容文本（用于记录）
	ContentUserID uint   `gorm:"not null" json:"content_user_id"`               // 内容发布者ID

	// 举报信息
	ReporterID     uint   `gorm:"not null;index" json:"reporter_id"`              // 举报人ID
	ReportReason   string `gorm:"type:varchar(50);not null" json:"report_reason"` // 举报原因类型
	ReportDetail   string `gorm:"type:text" json:"report_detail"`                 // 举报详细说明
	ReportEvidence string `gorm:"type:text" json:"report_evidence"`               // 举报证据（截图链接等）

	// 自动检测信息
	AutoDetected      bool    `gorm:"default:false" json:"auto_detected"`     // 是否为自动检测
	DetectionType     string  `gorm:"type:varchar(50)" json:"detection_type"` // 检测类型
	DetectionScore    float64 `gorm:"default:0" json:"detection_score"`       // 检测置信度分数
	DetectionKeywords string  `gorm:"type:text" json:"detection_keywords"`    // 命中的关键词

	// 处理状态
	Status        string     `gorm:"type:varchar(50);default:'pending'" json:"status"`  // pending, reviewing, resolved, rejected
	Priority      string     `gorm:"type:varchar(20);default:'normal'" json:"priority"` // low, normal, high, urgent
	HandlerID     *uint      `gorm:"index" json:"handler_id"`                           // 处理人ID
	HandleTime    *time.Time `json:"handle_time"`                                       // 处理时间
	HandleAction  string     `gorm:"type:varchar(50)" json:"handle_action"`             // 处理动作：warn, delete, ban, ignore
	HandleComment string     `gorm:"type:text" json:"handle_comment"`                   // 处理备注

	// 关联
	Reporter    User  `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	ContentUser User  `gorm:"foreignKey:ContentUserID" json:"content_user,omitempty"`
	Handler     *User `gorm:"foreignKey:HandlerID" json:"handler,omitempty"`
}

// ContentFilter 内容过滤规则
type ContentFilter struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 规则信息
	Name        string `gorm:"type:varchar(100);not null" json:"name"`            // 规则名称
	Description string `gorm:"type:text" json:"description"`                      // 规则描述
	RuleType    string `gorm:"type:varchar(50);not null" json:"rule_type"`        // keyword, regex, ai, url
	RuleContent string `gorm:"type:text;not null" json:"rule_content"`            // 规则内容
	Category    string `gorm:"type:varchar(50);not null" json:"category"`         // 违规类别：spam, porn, violence, politics, etc.
	Severity    string `gorm:"type:varchar(20);default:'normal'" json:"severity"` // low, normal, high, critical

	// 规则配置
	IsEnabled  bool    `gorm:"default:true" json:"is_enabled"`                  // 是否启用
	Action     string  `gorm:"type:varchar(50);default:'report'" json:"action"` // report（仅上报）, warn（警告）
	Threshold  float64 `gorm:"default:0.8" json:"threshold"`                    // 触发阈值
	AutoReport bool    `gorm:"default:true" json:"auto_report"`                 // 是否自动上报

	// 统计信息
	HitCount    int        `gorm:"default:0" json:"hit_count"` // 命中次数
	LastHitTime *time.Time `json:"last_hit_time"`              // 最后命中时间

	// 创建人
	CreatorID uint `gorm:"not null" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// UserWarning 用户警告记录
type UserWarning struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 警告信息
	UserID       uint   `gorm:"not null;index" json:"user_id"`                       // 被警告用户ID
	WarningType  string `gorm:"type:varchar(50);not null" json:"warning_type"`       // 警告类型
	WarningLevel string `gorm:"type:varchar(20);default:'low'" json:"warning_level"` // low, medium, high, severe
	Reason       string `gorm:"type:text;not null" json:"reason"`                    // 警告原因
	Evidence     string `gorm:"type:text" json:"evidence"`                           // 证据信息
	ReportID     *uint  `gorm:"index" json:"report_id"`                              // 关联的举报ID

	// 警告处理
	IssuedBy       uint       `gorm:"not null" json:"issued_by"`            // 发出警告的管理员ID
	ExpiresAt      *time.Time `json:"expires_at"`                           // 警告过期时间
	IsAcknowledged bool       `gorm:"default:false" json:"is_acknowledged"` // 用户是否已确认
	AcknowledgedAt *time.Time `json:"acknowledged_at"`                      // 确认时间

	// 关联
	User         User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IssuedByUser User           `gorm:"foreignKey:IssuedBy" json:"issued_by_user,omitempty"`
	Report       *ContentReport `gorm:"foreignKey:ReportID" json:"report,omitempty"`
}

// ModerationLog 内容审核日志
type ModerationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 操作信息
	Action     string `gorm:"type:varchar(50);not null" json:"action"`      // 操作类型
	TargetType string `gorm:"type:varchar(50);not null" json:"target_type"` // 目标类型
	TargetID   uint   `gorm:"not null" json:"target_id"`                    // 目标ID
	OperatorID uint   `gorm:"not null;index" json:"operator_id"`            // 操作人ID
	Reason     string `gorm:"type:text" json:"reason"`                      // 操作原因
	Details    string `gorm:"type:text" json:"details"`                     // 详细信息
	IPAddress  string `gorm:"type:varchar(45)" json:"ip_address"`           // IP地址

	// 关联
	Operator User `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// ContentStatistics 内容统计
type ContentStatistics struct {
	ID   uint      `gorm:"primaryKey" json:"id"`
	Date time.Time `gorm:"type:date;not null;index" json:"date"` // 统计日期

	// 举报统计
	TotalReports    int `gorm:"default:0" json:"total_reports"`    // 总举报数
	PendingReports  int `gorm:"default:0" json:"pending_reports"`  // 待处理举报
	ResolvedReports int `gorm:"default:0" json:"resolved_reports"` // 已处理举报
	RejectedReports int `gorm:"default:0" json:"rejected_reports"` // 已拒绝举报

	// 自动检测统计
	AutoDetected   int `gorm:"default:0" json:"auto_detected"`   // 自动检测数
	ManualReported int `gorm:"default:0" json:"manual_reported"` // 人工举报数

	// 处理统计
	WarningsIssued int `gorm:"default:0" json:"warnings_issued"` // 发出警告数
	ContentDeleted int `gorm:"default:0" json:"content_deleted"` // 删除内容数
	UsersBanned    int `gorm:"default:0" json:"users_banned"`    // 封禁用户数

	// 分类统计
	SpamReports     int `gorm:"default:0" json:"spam_reports"`     // 垃圾信息
	PornReports     int `gorm:"default:0" json:"porn_reports"`     // 色情内容
	ViolenceReports int `gorm:"default:0" json:"violence_reports"` // 暴力内容
	PoliticsReports int `gorm:"default:0" json:"politics_reports"` // 政治敏感
	OtherReports    int `gorm:"default:0" json:"other_reports"`    // 其他类型
}
