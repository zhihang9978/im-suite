package model

import (
	"time"

	"gorm.io/gorm"
)

// GroupInvite 群组邀请模型
type GroupInvite struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 邀请信息
	ChatID      uint   `gorm:"not null;index" json:"chat_id"`            // 群组ID
	InviteCode  string `gorm:"type:varchar(50);index:idx_group_invite_code,unique" json:"invite_code"` // 邀请码
	InviteLink  string `gorm:"type:varchar(500)" json:"invite_link"`     // 邀请链接
	CreatorID   uint   `gorm:"not null;index" json:"creator_id"`         // 创建者ID
	
	// 邀请设置
	MaxUses     int       `gorm:"default:0" json:"max_uses"`            // 最大使用次数，0表示无限制
	UsedCount   int       `gorm:"default:0" json:"used_count"`          // 已使用次数
	ExpiresAt   *time.Time `json:"expires_at"`                          // 过期时间
	RequireApproval bool  `gorm:"default:false" json:"require_approval"` // 是否需要审批
	
	// 邀请状态
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`          // 是否启用
	IsRevoked   bool   `gorm:"default:false" json:"is_revoked"`         // 是否已撤销
	RevokedBy   *uint  `json:"revoked_by"`                              // 撤销者ID
	RevokedAt   *time.Time `json:"revoked_at"`                          // 撤销时间
	RevokeReason string `gorm:"type:text" json:"revoke_reason"`         // 撤销原因
	
	// 关联
	Chat    Chat  `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	Creator User  `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	RevokedByUser *User `gorm:"foreignKey:RevokedBy" json:"revoked_by_user,omitempty"`
}

// GroupInviteUsage 邀请使用记录
type GroupInviteUsage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	InviteID  uint   `gorm:"not null;index" json:"invite_id"`    // 邀请ID
	UserID    uint   `gorm:"not null;index" json:"user_id"`      // 使用者ID
	IPAddress string `gorm:"type:varchar(45)" json:"ip_address"` // IP地址
	UserAgent string `gorm:"type:varchar(500)" json:"user_agent"` // User Agent
	Status    string `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, approved, rejected

	// 关联
	Invite GroupInvite `gorm:"foreignKey:InviteID" json:"invite,omitempty"`
	User   User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AdminRole 管理员角色模型
type AdminRole struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 角色信息
	Name        string `gorm:"type:varchar(50);not null" json:"name"`          // 角色名称
	DisplayName string `gorm:"type:varchar(100);not null" json:"display_name"` // 显示名称
	Description string `gorm:"type:text" json:"description"`                   // 角色描述
	Level       int    `gorm:"default:0" json:"level"`                         // 角色等级，数字越大权限越高
	IsBuiltIn   bool   `gorm:"default:false" json:"is_built_in"`               // 是否内置角色
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`                 // 是否启用

	// 权限配置
	CanManageMembers     bool `gorm:"default:false" json:"can_manage_members"`      // 管理成员
	CanDeleteMessages    bool `gorm:"default:false" json:"can_delete_messages"`     // 删除消息
	CanEditChat          bool `gorm:"default:false" json:"can_edit_chat"`           // 编辑群组信息
	CanInviteUsers       bool `gorm:"default:false" json:"can_invite_users"`        // 邀请用户
	CanBanUsers          bool `gorm:"default:false" json:"can_ban_users"`           // 封禁用户
	CanPromoteMembers    bool `gorm:"default:false" json:"can_promote_members"`     // 提升管理员
	CanManagePermissions bool `gorm:"default:false" json:"can_manage_permissions"`  // 管理权限
	CanManageInvites     bool `gorm:"default:false" json:"can_manage_invites"`      // 管理邀请
	CanPinMessages       bool `gorm:"default:false" json:"can_pin_messages"`        // 置顶消息
	CanManageAnnouncements bool `gorm:"default:false" json:"can_manage_announcements"` // 管理公告
	CanViewStatistics    bool `gorm:"default:false" json:"can_view_statistics"`     // 查看统计
	CanManageRoles       bool `gorm:"default:false" json:"can_manage_roles"`        // 管理角色
}

// ChatAdmin 群组管理员模型
type ChatAdmin struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 管理员信息
	ChatID    uint   `gorm:"not null;index:idx_chat_user" json:"chat_id"`        // 群组ID
	UserID    uint   `gorm:"not null;index:idx_chat_user" json:"user_id"`        // 用户ID
	RoleID    uint   `gorm:"not null" json:"role_id"`                            // 角色ID
	Title     string `gorm:"type:varchar(100)" json:"title"`                     // 自定义头衔
	PromotedBy uint  `gorm:"not null" json:"promoted_by"`                        // 提升者ID
	IsActive  bool   `gorm:"default:true" json:"is_active"`                      // 是否激活

	// 关联
	Chat       Chat      `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	User       User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role       AdminRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	PromotedByUser User   `gorm:"foreignKey:PromotedBy" json:"promoted_by_user,omitempty"`
}

// GroupJoinRequest 入群申请模型
type GroupJoinRequest struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 申请信息
	ChatID      uint   `gorm:"not null;index" json:"chat_id"`           // 群组ID
	UserID      uint   `gorm:"not null;index" json:"user_id"`           // 申请者ID
	InviteID    *uint  `gorm:"index" json:"invite_id"`                  // 邀请ID（通过邀请链接）
	Message     string `gorm:"type:text" json:"message"`                // 申请消息
	Status      string `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, approved, rejected
	
	// 审核信息
	ReviewedBy  *uint      `json:"reviewed_by"`                         // 审核者ID
	ReviewedAt  *time.Time `json:"reviewed_at"`                         // 审核时间
	ReviewNote  string     `gorm:"type:text" json:"review_note"`        // 审核备注
	
	// 关联
	Chat        Chat         `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Invite      *GroupInvite `gorm:"foreignKey:InviteID" json:"invite,omitempty"`
	ReviewedByUser *User     `gorm:"foreignKey:ReviewedBy" json:"reviewed_by_user,omitempty"`
}

// GroupAuditLog 群组审计日志
type GroupAuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 日志信息
	ChatID     uint   `gorm:"not null;index" json:"chat_id"`           // 群组ID
	OperatorID uint   `gorm:"not null;index" json:"operator_id"`       // 操作者ID
	Action     string `gorm:"type:varchar(50);not null" json:"action"` // 操作类型
	TargetType string `gorm:"type:varchar(50)" json:"target_type"`     // 目标类型
	TargetID   uint   `json:"target_id"`                               // 目标ID
	Details    string `gorm:"type:text" json:"details"`                // 详细信息
	IPAddress  string `gorm:"type:varchar(45)" json:"ip_address"`      // IP地址
	
	// 关联
	Chat     Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	Operator User `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// GroupPermissionTemplate 权限模板
type GroupPermissionTemplate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 模板信息
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Category    string `gorm:"type:varchar(50)" json:"category"` // strict, moderate, relaxed
	IsPublic    bool   `gorm:"default:true" json:"is_public"`
	UsageCount  int    `gorm:"default:0" json:"usage_count"`
	
	// 权限配置（JSON格式）
	Config string `gorm:"type:text;not null" json:"config"`
}
