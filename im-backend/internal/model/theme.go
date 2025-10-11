package model

import (
	"time"

	"gorm.io/gorm"
)

// Theme 主题模型
type Theme struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Name        string `gorm:"type:varchar(100);not null" json:"name"`         // 主题名称
	DisplayName string `gorm:"type:varchar(100);not null" json:"display_name"` // 显示名称
	Description string `gorm:"type:text" json:"description"`                   // 主题描述
	ThemeType   string `gorm:"type:varchar(20);not null" json:"theme_type"`    // light, dark, auto
	IsBuiltIn   bool   `gorm:"default:false" json:"is_built_in"`               // 是否内置主题
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`                 // 是否启用

	// 颜色配置
	PrimaryColor       string `gorm:"type:varchar(20)" json:"primary_color"`        // 主色调
	SecondaryColor     string `gorm:"type:varchar(20)" json:"secondary_color"`      // 副色调
	AccentColor        string `gorm:"type:varchar(20)" json:"accent_color"`         // 强调色
	BackgroundColor    string `gorm:"type:varchar(20)" json:"background_color"`     // 背景色
	SurfaceColor       string `gorm:"type:varchar(20)" json:"surface_color"`        // 表面色
	TextPrimaryColor   string `gorm:"type:varchar(20)" json:"text_primary_color"`   // 主文本色
	TextSecondaryColor string `gorm:"type:varchar(20)" json:"text_secondary_color"` // 副文本色
	BorderColor        string `gorm:"type:varchar(20)" json:"border_color"`         // 边框色
	ErrorColor         string `gorm:"type:varchar(20)" json:"error_color"`          // 错误色
	SuccessColor       string `gorm:"type:varchar(20)" json:"success_color"`        // 成功色
	WarningColor       string `gorm:"type:varchar(20)" json:"warning_color"`        // 警告色

	// 布局配置
	BorderRadius        string `gorm:"type:varchar(20)" json:"border_radius"`         // 圆角大小
	Spacing             string `gorm:"type:varchar(20)" json:"spacing"`               // 间距
	FontSize            string `gorm:"type:varchar(20)" json:"font_size"`             // 字体大小
	FontFamily          string `gorm:"type:varchar(100)" json:"font_family"`          // 字体族
	MessageBubbleRadius string `gorm:"type:varchar(20)" json:"message_bubble_radius"` // 消息气泡圆角

	// 高级配置
	CustomCSS string `gorm:"type:text" json:"custom_css"`      // 自定义CSS
	CustomJS  string `gorm:"type:text" json:"custom_js"`       // 自定义JS
	Preview   string `gorm:"type:varchar(500)" json:"preview"` // 预览图URL

	// 统计信息
	UsageCount int  `gorm:"default:0" json:"usage_count"` // 使用次数
	CreatorID  uint `gorm:"index" json:"creator_id"`      // 创建者ID

	// 关联
	Creator User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// UserThemeSetting 用户主题设置
type UserThemeSetting struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 用户设置
	UserID        uint   `gorm:"not null;index:idx_user_theme_settings_user,unique" json:"user_id"` // 用户ID
	ThemeID       uint   `gorm:"not null" json:"theme_id"`                                          // 主题ID
	AutoDarkMode  bool   `gorm:"default:false" json:"auto_dark_mode"`                               // 自动夜间模式
	DarkModeStart string `gorm:"type:varchar(5);default:'22:00'" json:"dark_mode_start"`            // 夜间模式开始时间
	DarkModeEnd   string `gorm:"type:varchar(5);default:'07:00'" json:"dark_mode_end"`              // 夜间模式结束时间
	FollowSystem  bool   `gorm:"default:false" json:"follow_system"`                                // 跟随系统

	// 个性化配置
	CustomPrimaryColor    *string `gorm:"type:varchar(20)" json:"custom_primary_color"`    // 自定义主色
	CustomBackgroundColor *string `gorm:"type:varchar(20)" json:"custom_background_color"` // 自定义背景色
	CustomFontSize        *string `gorm:"type:varchar(20)" json:"custom_font_size"`        // 自定义字体大小
	CustomMessageBubble   *string `gorm:"type:varchar(20)" json:"custom_message_bubble"`   // 自定义消息气泡样式

	// 动画设置
	EnableAnimations bool   `gorm:"default:true" json:"enable_animations"`                    // 启用动画
	ReducedMotion    bool   `gorm:"default:false" json:"reduced_motion"`                      // 减少动效
	AnimationSpeed   string `gorm:"type:varchar(20);default:'normal'" json:"animation_speed"` // 动画速度: slow, normal, fast

	// 布局设置
	CompactMode     bool `gorm:"default:false" json:"compact_mode"`    // 紧凑模式
	ShowAvatars     bool `gorm:"default:true" json:"show_avatars"`     // 显示头像
	MessageGrouping bool `gorm:"default:true" json:"message_grouping"` // 消息分组

	// 关联
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Theme Theme `gorm:"foreignKey:ThemeID" json:"theme,omitempty"`
}

// ThemeTemplate 主题模板（预设）
type ThemeTemplate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 模板信息
	Name          string `gorm:"type:varchar(100);not null" json:"name"`
	Category      string `gorm:"type:varchar(50)" json:"category"` // business, creative, minimal, colorful
	Preview       string `gorm:"type:varchar(500)" json:"preview"`
	IsPopular     bool   `gorm:"default:false" json:"is_popular"`
	DownloadCount int    `gorm:"default:0" json:"download_count"`

	// 主题配置（JSON格式）
	Config string `gorm:"type:text;not null" json:"config"`
}
