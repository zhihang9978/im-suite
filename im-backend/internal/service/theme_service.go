package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/gorm"
)

// ThemeService 主题服务
type ThemeService struct {
	db *gorm.DB
}

// NewThemeService 创建主题服务实例
func NewThemeService(db *gorm.DB) *ThemeService {
	return &ThemeService{
		db: db,
	}
}

// CreateThemeRequest 创建主题请求
type CreateThemeRequest struct {
	Name               string  `json:"name" binding:"required"`
	DisplayName        string  `json:"display_name" binding:"required"`
	Description        string  `json:"description"`
	ThemeType          string  `json:"theme_type" binding:"required"` // light, dark, auto
	PrimaryColor       string  `json:"primary_color"`
	SecondaryColor     string  `json:"secondary_color"`
	AccentColor        string  `json:"accent_color"`
	BackgroundColor    string  `json:"background_color"`
	SurfaceColor       string  `json:"surface_color"`
	TextPrimaryColor   string  `json:"text_primary_color"`
	TextSecondaryColor string  `json:"text_secondary_color"`
	BorderColor        string  `json:"border_color"`
	ErrorColor         string  `json:"error_color"`
	SuccessColor       string  `json:"success_color"`
	WarningColor       string  `json:"warning_color"`
	BorderRadius       string  `json:"border_radius"`
	Spacing            string  `json:"spacing"`
	FontSize           string  `json:"font_size"`
	FontFamily         string  `json:"font_family"`
	MessageBubbleRadius string `json:"message_bubble_radius"`
	CustomCSS          string  `json:"custom_css"`
	Preview            string  `json:"preview"`
	CreatorID          uint    `json:"creator_id"`
}

// UpdateUserThemeRequest 更新用户主题设置请求
type UpdateUserThemeRequest struct {
	UserID                uint    `json:"user_id" binding:"required"`
	ThemeID               uint    `json:"theme_id" binding:"required"`
	AutoDarkMode          bool    `json:"auto_dark_mode"`
	DarkModeStart         string  `json:"dark_mode_start"`
	DarkModeEnd           string  `json:"dark_mode_end"`
	FollowSystem          bool    `json:"follow_system"`
	CustomPrimaryColor    *string `json:"custom_primary_color"`
	CustomBackgroundColor *string `json:"custom_background_color"`
	CustomFontSize        *string `json:"custom_font_size"`
	CustomMessageBubble   *string `json:"custom_message_bubble"`
	EnableAnimations      bool    `json:"enable_animations"`
	ReducedMotion         bool    `json:"reduced_motion"`
	AnimationSpeed        string  `json:"animation_speed"`
	CompactMode           bool    `json:"compact_mode"`
	ShowAvatars           bool    `json:"show_avatars"`
	MessageGrouping       bool    `json:"message_grouping"`
}

// CreateTheme 创建主题
func (s *ThemeService) CreateTheme(req CreateThemeRequest) (*model.Theme, error) {
	// 验证主题类型
	validTypes := []string{"light", "dark", "auto"}
	if !contains(validTypes, req.ThemeType) {
		return nil, errors.New("无效的主题类型")
	}

	theme := model.Theme{
		Name:                req.Name,
		DisplayName:         req.DisplayName,
		Description:         req.Description,
		ThemeType:           req.ThemeType,
		IsBuiltIn:           false,
		IsEnabled:           true,
		PrimaryColor:        req.PrimaryColor,
		SecondaryColor:      req.SecondaryColor,
		AccentColor:         req.AccentColor,
		BackgroundColor:     req.BackgroundColor,
		SurfaceColor:        req.SurfaceColor,
		TextPrimaryColor:    req.TextPrimaryColor,
		TextSecondaryColor:  req.TextSecondaryColor,
		BorderColor:         req.BorderColor,
		ErrorColor:          req.ErrorColor,
		SuccessColor:        req.SuccessColor,
		WarningColor:        req.WarningColor,
		BorderRadius:        req.BorderRadius,
		Spacing:             req.Spacing,
		FontSize:            req.FontSize,
		FontFamily:          req.FontFamily,
		MessageBubbleRadius: req.MessageBubbleRadius,
		CustomCSS:           req.CustomCSS,
		Preview:             req.Preview,
		CreatorID:           req.CreatorID,
	}

	if err := s.db.Create(&theme).Error; err != nil {
		return nil, fmt.Errorf("创建主题失败: %v", err)
	}

	return &theme, nil
}

// GetTheme 获取主题
func (s *ThemeService) GetTheme(themeID uint) (*model.Theme, error) {
	var theme model.Theme
	if err := s.db.Preload("Creator").First(&theme, themeID).Error; err != nil {
		return nil, errors.New("主题不存在")
	}
	return &theme, nil
}

// ListThemes 获取主题列表
func (s *ThemeService) ListThemes(themeType string, limit, offset int) ([]model.Theme, int64, error) {
	var themes []model.Theme
	var total int64

	query := s.db.Where("is_enabled = ?", true)
	if themeType != "" {
		query = query.Where("theme_type = ?", themeType)
	}

	// 获取总数
	query.Model(&model.Theme{}).Count(&total)

	// 获取分页数据
	if err := query.Order("usage_count DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&themes).Error; err != nil {
		return nil, 0, fmt.Errorf("获取主题列表失败: %v", err)
	}

	return themes, total, nil
}

// UpdateUserTheme 更新用户主题设置
func (s *ThemeService) UpdateUserTheme(req UpdateUserThemeRequest) (*model.UserThemeSetting, error) {
	// 验证主题是否存在
	var theme model.Theme
	if err := s.db.First(&theme, req.ThemeID).Error; err != nil {
		return nil, errors.New("主题不存在")
	}

	// 查找或创建用户主题设置
	var setting model.UserThemeSetting
	err := s.db.Where("user_id = ?", req.UserID).First(&setting).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新设置
			setting = model.UserThemeSetting{
				UserID:                req.UserID,
				ThemeID:               req.ThemeID,
				AutoDarkMode:          req.AutoDarkMode,
				DarkModeStart:         req.DarkModeStart,
				DarkModeEnd:           req.DarkModeEnd,
				FollowSystem:          req.FollowSystem,
				CustomPrimaryColor:    req.CustomPrimaryColor,
				CustomBackgroundColor: req.CustomBackgroundColor,
				CustomFontSize:        req.CustomFontSize,
				CustomMessageBubble:   req.CustomMessageBubble,
				EnableAnimations:      req.EnableAnimations,
				ReducedMotion:         req.ReducedMotion,
				AnimationSpeed:        req.AnimationSpeed,
				CompactMode:           req.CompactMode,
				ShowAvatars:           req.ShowAvatars,
				MessageGrouping:       req.MessageGrouping,
			}
			if err := s.db.Create(&setting).Error; err != nil {
				return nil, fmt.Errorf("创建用户主题设置失败: %v", err)
			}
		} else {
			return nil, fmt.Errorf("查询用户主题设置失败: %v", err)
		}
	} else {
		// 更新现有设置
		updates := map[string]interface{}{
			"theme_id":                 req.ThemeID,
			"auto_dark_mode":           req.AutoDarkMode,
			"dark_mode_start":          req.DarkModeStart,
			"dark_mode_end":            req.DarkModeEnd,
			"follow_system":            req.FollowSystem,
			"custom_primary_color":     req.CustomPrimaryColor,
			"custom_background_color":  req.CustomBackgroundColor,
			"custom_font_size":         req.CustomFontSize,
			"custom_message_bubble":    req.CustomMessageBubble,
			"enable_animations":        req.EnableAnimations,
			"reduced_motion":           req.ReducedMotion,
			"animation_speed":          req.AnimationSpeed,
			"compact_mode":             req.CompactMode,
			"show_avatars":             req.ShowAvatars,
			"message_grouping":         req.MessageGrouping,
		}
		if err := s.db.Model(&setting).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新用户主题设置失败: %v", err)
		}
	}

	// 更新主题使用次数
	s.db.Model(&theme).Update("usage_count", gorm.Expr("usage_count + 1"))

	// 预加载关联数据
	s.db.Preload("Theme").First(&setting, setting.ID)

	return &setting, nil
}

// GetUserTheme 获取用户主题设置
func (s *ThemeService) GetUserTheme(userID uint) (*model.UserThemeSetting, error) {
	var setting model.UserThemeSetting
	err := s.db.Preload("Theme").Where("user_id = ?", userID).First(&setting).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认主题
			return s.getDefaultThemeSetting(userID)
		}
		return nil, fmt.Errorf("获取用户主题设置失败: %v", err)
	}

	// 检查是否需要自动切换夜间模式
	if setting.AutoDarkMode && !setting.FollowSystem {
		currentHour := time.Now().Hour()
		startHour := parseTimeString(setting.DarkModeStart)
		endHour := parseTimeString(setting.DarkModeEnd)

		isDarkTime := false
		if startHour > endHour {
			// 跨天的情况，如 22:00 到 07:00
			isDarkTime = currentHour >= startHour || currentHour < endHour
		} else {
			isDarkTime = currentHour >= startHour && currentHour < endHour
		}

		if isDarkTime {
			// 应该使用夜间主题
			var darkTheme model.Theme
			if err := s.db.Where("theme_type = ? AND is_built_in = ?", "dark", true).First(&darkTheme).Error; err == nil {
				setting.ThemeID = darkTheme.ID
				setting.Theme = darkTheme
			}
		}
	}

	return &setting, nil
}

// getDefaultThemeSetting 获取默认主题设置
func (s *ThemeService) getDefaultThemeSetting(userID uint) (*model.UserThemeSetting, error) {
	// 获取默认light主题
	var defaultTheme model.Theme
	if err := s.db.Where("theme_type = ? AND is_built_in = ?", "light", true).First(&defaultTheme).Error; err != nil {
		return nil, errors.New("默认主题不存在")
	}

	setting := &model.UserThemeSetting{
		UserID:           userID,
		ThemeID:          defaultTheme.ID,
		AutoDarkMode:     false,
		DarkModeStart:    "22:00",
		DarkModeEnd:      "07:00",
		FollowSystem:     false,
		EnableAnimations: true,
		ReducedMotion:    false,
		AnimationSpeed:   "normal",
		CompactMode:      false,
		ShowAvatars:      true,
		MessageGrouping:  true,
		Theme:            defaultTheme,
	}

	return setting, nil
}

// InitializeBuiltInThemes 初始化内置主题
func (s *ThemeService) InitializeBuiltInThemes() error {
	// 检查是否已初始化
	var count int64
	s.db.Model(&model.Theme{}).Where("is_built_in = ?", true).Count(&count)
	if count > 0 {
		return nil // 已初始化
	}

	// Light主题
	lightTheme := model.Theme{
		Name:                "light",
		DisplayName:         "浅色主题",
		Description:         "清新明亮的浅色主题，适合白天使用",
		ThemeType:           "light",
		IsBuiltIn:           true,
		IsEnabled:           true,
		PrimaryColor:        "#2196F3",
		SecondaryColor:      "#03A9F4",
		AccentColor:         "#FF5722",
		BackgroundColor:     "#FFFFFF",
		SurfaceColor:        "#F5F5F5",
		TextPrimaryColor:    "#212121",
		TextSecondaryColor:  "#757575",
		BorderColor:         "#E0E0E0",
		ErrorColor:          "#F44336",
		SuccessColor:        "#4CAF50",
		WarningColor:        "#FF9800",
		BorderRadius:        "8px",
		Spacing:             "16px",
		FontSize:            "14px",
		FontFamily:          "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
		MessageBubbleRadius: "18px",
	}

	// Dark主题
	darkTheme := model.Theme{
		Name:                "dark",
		DisplayName:         "深色主题",
		Description:         "舒适护眼的深色主题，适合夜间使用",
		ThemeType:           "dark",
		IsBuiltIn:           true,
		IsEnabled:           true,
		PrimaryColor:        "#1E88E5",
		SecondaryColor:      "#0277BD",
		AccentColor:         "#FF6F00",
		BackgroundColor:     "#121212",
		SurfaceColor:        "#1E1E1E",
		TextPrimaryColor:    "#FFFFFF",
		TextSecondaryColor:  "#B0B0B0",
		BorderColor:         "#2C2C2C",
		ErrorColor:          "#EF5350",
		SuccessColor:        "#66BB6A",
		WarningColor:        "#FFA726",
		BorderRadius:        "8px",
		Spacing:             "16px",
		FontSize:            "14px",
		FontFamily:          "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
		MessageBubbleRadius: "18px",
	}

	// Auto主题（根据系统自动切换）
	autoTheme := model.Theme{
		Name:                "auto",
		DisplayName:         "自动主题",
		Description:         "根据系统设置自动切换浅色和深色主题",
		ThemeType:           "auto",
		IsBuiltIn:           true,
		IsEnabled:           true,
		PrimaryColor:        "#2196F3",
		SecondaryColor:      "#03A9F4",
		AccentColor:         "#FF5722",
		BackgroundColor:     "#FFFFFF",
		SurfaceColor:        "#F5F5F5",
		TextPrimaryColor:    "#212121",
		TextSecondaryColor:  "#757575",
		BorderColor:         "#E0E0E0",
		ErrorColor:          "#F44336",
		SuccessColor:        "#4CAF50",
		WarningColor:        "#FF9800",
		BorderRadius:        "8px",
		Spacing:             "16px",
		FontSize:            "14px",
		FontFamily:          "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
		MessageBubbleRadius: "18px",
	}

	themes := []model.Theme{lightTheme, darkTheme, autoTheme}
	for _, theme := range themes {
		if err := s.db.Create(&theme).Error; err != nil {
			return fmt.Errorf("初始化内置主题失败: %v", err)
		}
	}

	return nil
}

// CreateThemeFromTemplate 从模板创建主题
func (s *ThemeService) CreateThemeFromTemplate(templateID, userID uint) (*model.Theme, error) {
	// 获取模板
	var template model.ThemeTemplate
	if err := s.db.First(&template, templateID).Error; err != nil {
		return nil, errors.New("模板不存在")
	}

	// 解析配置
	var config CreateThemeRequest
	if err := json.Unmarshal([]byte(template.Config), &config); err != nil {
		return nil, fmt.Errorf("解析模板配置失败: %v", err)
	}

	config.CreatorID = userID
	config.Name = fmt.Sprintf("%s_user_%d_%d", template.Name, userID, time.Now().Unix())

	// 创建主题
	theme, err := s.CreateTheme(config)
	if err != nil {
		return nil, err
	}

	// 更新模板下载次数
	s.db.Model(&template).Update("download_count", gorm.Expr("download_count + 1"))

	return theme, nil
}

// parseTimeString 解析时间字符串
func parseTimeString(timeStr string) int {
	var hour int
	fmt.Sscanf(timeStr, "%d:", &hour)
	return hour
}
