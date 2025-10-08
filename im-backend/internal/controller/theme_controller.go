package controller

import (
	"net/http"
	"strconv"
	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ThemeController 主题控制器
type ThemeController struct {
	themeService *service.ThemeService
}

// NewThemeController 创建主题控制器实例
func NewThemeController(themeService *service.ThemeService) *ThemeController {
	return &ThemeController{
		themeService: themeService,
	}
}

// CreateTheme 创建主题
// @Summary 创建主题
// @Description 创建自定义主题
// @Tags 主题管理
// @Accept json
// @Produce json
// @Param request body service.CreateThemeRequest true "主题信息"
// @Success 200 {object} model.Theme
// @Failure 400 {object} map[string]interface{}
// @Router /api/themes [post]
func (c *ThemeController) CreateTheme(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateThemeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.CreatorID = userID.(uint)

	theme, err := c.themeService.CreateTheme(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "主题创建成功",
		"theme":   theme,
	})
}

// GetTheme 获取主题
// @Summary 获取主题
// @Description 获取主题详情
// @Tags 主题管理
// @Produce json
// @Param theme_id path int true "主题ID"
// @Success 200 {object} model.Theme
// @Failure 400 {object} map[string]interface{}
// @Router /api/themes/{theme_id} [get]
func (c *ThemeController) GetTheme(ctx *gin.Context) {
	themeIDStr := ctx.Param("theme_id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "主题ID格式错误"})
		return
	}

	theme, err := c.themeService.GetTheme(uint(themeID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, theme)
}

// ListThemes 获取主题列表
// @Summary 获取主题列表
// @Description 获取所有可用主题
// @Tags 主题管理
// @Produce json
// @Param theme_type query string false "主题类型"
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/themes [get]
func (c *ThemeController) ListThemes(ctx *gin.Context) {
	themeType := ctx.Query("theme_type")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	themes, total, err := c.themeService.ListThemes(themeType, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"themes": themes,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateUserTheme 更新用户主题设置
// @Summary 更新用户主题
// @Description 更新用户的主题设置
// @Tags 主题管理
// @Accept json
// @Produce json
// @Param request body service.UpdateUserThemeRequest true "主题设置"
// @Success 200 {object} model.UserThemeSetting
// @Failure 400 {object} map[string]interface{}
// @Router /api/themes/user [put]
func (c *ThemeController) UpdateUserTheme(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.UpdateUserThemeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.UserID = userID.(uint)

	setting, err := c.themeService.UpdateUserTheme(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "主题设置更新成功",
		"setting": setting,
	})
}

// GetUserTheme 获取用户主题设置
// @Summary 获取用户主题
// @Description 获取用户当前的主题设置
// @Tags 主题管理
// @Produce json
// @Success 200 {object} model.UserThemeSetting
// @Failure 400 {object} map[string]interface{}
// @Router /api/themes/user [get]
func (c *ThemeController) GetUserTheme(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	setting, err := c.themeService.GetUserTheme(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, setting)
}

// InitializeBuiltInThemes 初始化内置主题
// @Summary 初始化内置主题
// @Description 初始化系统内置的浅色和深色主题
// @Tags 主题管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/themes/initialize [post]
func (c *ThemeController) InitializeBuiltInThemes(ctx *gin.Context) {
	if err := c.themeService.InitializeBuiltInThemes(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "内置主题初始化成功",
	})
}
