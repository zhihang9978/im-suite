package model

import (
	"time"
	"gorm.io/gorm"
)

// Bot 机器人模型
type Bot struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// 基本信息
	Name        string `json:"name" gorm:"not null;uniqueIndex"`          // 机器人名称
	Description string `json:"description"`                               // 描述
	Type        string `json:"type" gorm:"not null"`                      // 类型: internal, webhook, plugin
	
	// 认证信息
	APIKey      string `json:"api_key" gorm:"uniqueIndex;not null"`       // API密钥
	APISecret   string `json:"-" gorm:"not null"`                         // API密钥（加密存储）
	
	// 权限配置
	Permissions string `json:"permissions" gorm:"type:text"`              // 权限列表（JSON数组）
	
	// 状态信息
	IsActive    bool      `json:"is_active" gorm:"default:true"`          // 是否激活
	LastUsedAt  time.Time `json:"last_used_at"`                           // 最后使用时间
	
	// 限制配置
	RateLimit     int  `json:"rate_limit" gorm:"default:100"`            // 速率限制（请求/分钟）
	DailyLimit    int  `json:"daily_limit" gorm:"default:10000"`         // 每日限制
	
	// 创建者信息
	CreatedBy   uint `json:"created_by"`                                 // 创建者用户ID
	
	// 统计信息
	TotalCalls  int64 `json:"total_calls" gorm:"default:0"`              // 总调用次数
	SuccessCalls int64 `json:"success_calls" gorm:"default:0"`           // 成功次数
	FailedCalls int64 `json:"failed_calls" gorm:"default:0"`             // 失败次数
	
	// 关联关系
	Creator User `json:"creator" gorm:"foreignKey:CreatedBy"`
}

// BotPermission 机器人权限定义
type BotPermission string

const (
	// 用户管理权限
	PermissionCreateUser  BotPermission = "create_user"   // 创建用户
	PermissionDeleteUser  BotPermission = "delete_user"   // 删除用户
	PermissionBanUser     BotPermission = "ban_user"      // 封禁用户
	PermissionUnbanUser   BotPermission = "unban_user"    // 解封用户
	PermissionUpdateUser  BotPermission = "update_user"   // 更新用户信息
	PermissionListUsers   BotPermission = "list_users"    // 查看用户列表
	
	// 消息管理权限
	PermissionSendMessage   BotPermission = "send_message"    // 发送消息
	PermissionDeleteMessage BotPermission = "delete_message"  // 删除消息
	PermissionBroadcast     BotPermission = "broadcast"       // 广播消息
	
	// 群组管理权限
	PermissionCreateGroup BotPermission = "create_group"  // 创建群组
	PermissionDeleteGroup BotPermission = "delete_group"  // 删除群组
	PermissionManageGroup BotPermission = "manage_group"  // 管理群组
	
	// 内容审核权限
	PermissionModerateContent BotPermission = "moderate_content" // 内容审核
	
	// 系统管理权限
	PermissionViewStats  BotPermission = "view_stats"   // 查看统计
	PermissionViewLogs   BotPermission = "view_logs"    // 查看日志
)

// BotAPILog 机器人API调用日志
type BotAPILog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	BotID      uint      `json:"bot_id" gorm:"not null;index"`           // 机器人ID
	Endpoint   string    `json:"endpoint"`                               // 调用的端点
	Method     string    `json:"method"`                                 // HTTP方法
	StatusCode int       `json:"status_code"`                            // 响应状态码
	IPAddress  string    `json:"ip_address"`                             // 调用IP
	UserAgent  string    `json:"user_agent"`                             // User-Agent
	Duration   int64     `json:"duration"`                               // 耗时（毫秒）
	RequestBody string   `json:"request_body" gorm:"type:text"`          // 请求体
	ResponseBody string  `json:"response_body" gorm:"type:text"`         // 响应体
	Error      string    `json:"error"`                                  // 错误信息
	
	// 关联关系
	Bot Bot `json:"bot" gorm:"foreignKey:BotID"`
}

// TableName 指定表名
func (Bot) TableName() string {
	return "bots"
}

// TableName 指定表名
func (BotAPILog) TableName() string {
	return "bot_api_logs"
}

