package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// SuperAdminController 超级管理员控制器
type SuperAdminController struct {
	superAdminService *service.SuperAdminService
}

// NewSuperAdminController 创建超级管理员控制器
func NewSuperAdminController() *SuperAdminController {
	return &SuperAdminController{
		superAdminService: service.NewSuperAdminService(),
	}
}

// GetSystemStats 获取系统统计信息
func (c *SuperAdminController) GetSystemStats(ctx *gin.Context) {
	stats, err := c.superAdminService.GetSystemStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetOnlineUsers 获取在线用户列表
func (c *SuperAdminController) GetOnlineUsers(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "50")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	users, total, err := c.superAdminService.GetOnlineUsers(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users":     users,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUserActivity 获取用户活动记录
func (c *SuperAdminController) GetUserActivity(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	limitStr := ctx.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	activities, err := c.superAdminService.GetUserActivity(uint(userID), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// GetUserBehaviorAnalysis 获取用户行为分析
func (c *SuperAdminController) GetUserBehaviorAnalysis(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	analysis, err := c.superAdminService.GetUserBehaviorAnalysis(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
	})
}

// ForceLogoutUser 强制用户下线
func (c *SuperAdminController) ForceLogoutUser(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.superAdminService.ForceLogoutUser(uint(userID), req.Reason)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已强制下线",
	})
}

// BanUser 封禁用户
func (c *SuperAdminController) BanUser(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Duration int    `json:"duration" binding:"required"` // 封禁时长（小时）
		Reason   string `json:"reason" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, _ := ctx.Get("user_id")
	duration := time.Duration(req.Duration) * time.Hour

	err = c.superAdminService.BanUser(uint(userID), duration, req.Reason, adminID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已封禁",
	})
}

// UnbanUser 解封用户
func (c *SuperAdminController) UnbanUser(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	adminID, _ := ctx.Get("user_id")

	err = c.superAdminService.UnbanUser(uint(userID), adminID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已解封",
	})
}

// DeleteUserAccount 删除用户账号
func (c *SuperAdminController) DeleteUserAccount(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, _ := ctx.Get("user_id")

	err = c.superAdminService.DeleteUserAccount(uint(userID), adminID.(uint), req.Reason)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户账号已删除",
	})
}

// GetContentModerationQueue 获取内容审核队列
func (c *SuperAdminController) GetContentModerationQueue(ctx *gin.Context) {
	status := ctx.DefaultQuery("status", "pending")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	records, total, err := c.superAdminService.GetContentModerationQueue(status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"records":   records,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ModerateContent 审核内容
func (c *SuperAdminController) ModerateContent(ctx *gin.Context) {
	contentIDStr := ctx.Param("content_id")
	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的内容ID"})
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"` // approve, reject, delete, warn, ban
		Reason string `json:"reason" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reviewerID, _ := ctx.Get("user_id")

	err = c.superAdminService.ModerateContent(uint(contentID), req.Action, req.Reason, reviewerID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "内容审核完成",
	})
}

// GetSystemLogs 获取系统日志
func (c *SuperAdminController) GetSystemLogs(ctx *gin.Context) {
	logType := ctx.DefaultQuery("type", "all")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "50")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	logs, total, err := c.superAdminService.GetSystemLogs(logType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// BroadcastMessage 广播系统消息
func (c *SuperAdminController) BroadcastMessage(ctx *gin.Context) {
	var req struct {
		Message    string `json:"message" binding:"required"`
		TargetType string `json:"target_type" binding:"required"` // all, users, groups
		TargetIDs  []uint `json:"target_ids"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.superAdminService.BroadcastMessage(req.Message, req.TargetType, req.TargetIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统消息已广播",
	})
}

// GetServerHealth 获取服务器健康状态
func (c *SuperAdminController) GetServerHealth(ctx *gin.Context) {
	health, err := c.superAdminService.GetServerHealth()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

// SetupRoutes 设置路由
func (c *SuperAdminController) SetupRoutes(r *gin.RouterGroup) {
	// 系统统计
	r.GET("/stats", c.GetSystemStats)
	r.GET("/health", c.GetServerHealth)

	// 在线用户管理
	r.GET("/online-users", c.GetOnlineUsers)

	// 用户管理
	users := r.Group("/users")
	{
		users.GET("/:user_id/activity", c.GetUserActivity)
		users.GET("/:user_id/analysis", c.GetUserBehaviorAnalysis)
		users.POST("/:user_id/force-logout", c.ForceLogoutUser)
		users.POST("/:user_id/ban", c.BanUser)
		users.POST("/:user_id/unban", c.UnbanUser)
		users.DELETE("/:user_id", c.DeleteUserAccount)
	}

	// 内容审核
	moderation := r.Group("/moderation")
	{
		moderation.GET("/queue", c.GetContentModerationQueue)
		moderation.POST("/:content_id/moderate", c.ModerateContent)
	}

	// 系统管理
	system := r.Group("/system")
	{
		system.GET("/logs", c.GetSystemLogs)
		system.POST("/broadcast", c.BroadcastMessage)
	}
}
