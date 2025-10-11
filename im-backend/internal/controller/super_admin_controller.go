package controller

import (
	"net/http"
	"strconv"
	"time"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// SuperAdminController 超级管理员控制器
type SuperAdminController struct {
	service *service.SuperAdminService
	monitor *service.SystemMonitorService
}

// NewSuperAdminController 创建超级管理员控制器
func NewSuperAdminController() *SuperAdminController {
	return &SuperAdminController{
		service: service.NewSuperAdminService(),
		monitor: service.NewSystemMonitorService(),
	}
}

// SetupRoutes 设置路由
func (c *SuperAdminController) SetupRoutes(router *gin.RouterGroup) {
	// 系统统计
	router.GET("/stats", c.GetSystemStats)
	router.GET("/stats/system", c.GetSystemMetrics)

	// 用户列表和管理
	router.GET("/users", c.GetUserList) // 获取所有用户列表
	router.GET("/users/online", c.GetOnlineUsers)
	router.POST("/users/:id/logout", c.ForceLogout)

	// 用户管理
	router.POST("/users/:id/ban", c.BanUser)
	router.POST("/users/:id/unban", c.UnbanUser)
	router.DELETE("/users/:id", c.DeleteUser)
	router.GET("/users/:id/analysis", c.GetUserAnalysis)

	router.GET("/chats", c.GetChatList)
	router.GET("/messages", c.GetMessageList)

	// 系统管理
	router.GET("/alerts", c.GetAlerts)
	router.GET("/logs", c.GetAdminLogs)
	router.POST("/broadcast", c.BroadcastMessage)
}

// GetSystemStats 获取系统统计
func (c *SuperAdminController) GetSystemStats(ctx *gin.Context) {
	stats, err := c.service.GetSystemStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSystemMetrics 获取系统指标
func (c *SuperAdminController) GetSystemMetrics(ctx *gin.Context) {
	metrics, err := c.monitor.GetSystemStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetUserList 获取用户列表
func (c *SuperAdminController) GetUserList(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 搜索条件
	username := ctx.Query("username")
	phone := ctx.Query("phone")
	status := ctx.Query("status")

	users, total, err := c.service.GetUserList(page, pageSize, username, phone, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetOnlineUsers 获取在线用户
func (c *SuperAdminController) GetOnlineUsers(ctx *gin.Context) {
	users, err := c.service.GetOnlineUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"total":   len(users),
	})
}

// ForceLogout 强制用户下线
func (c *SuperAdminController) ForceLogout(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := c.service.ForceLogoutUser(adminID, uint(userID)); err != nil {
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
	adminID := ctx.GetUint("user_id")
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Duration int64  `json:"duration"` // 封禁时长（秒）
		Reason   string `json:"reason"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	duration := time.Duration(req.Duration) * time.Second
	if err := c.service.BanUser(adminID, uint(userID), duration, req.Reason); err != nil {
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
	adminID := ctx.GetUint("user_id")
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := c.service.UnbanUser(adminID, uint(userID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已解封",
	})
}

// DeleteUser 删除用户
func (c *SuperAdminController) DeleteUser(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := c.service.DeleteUser(adminID, uint(userID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已删除",
	})
}

// GetUserAnalysis 获取用户分析
func (c *SuperAdminController) GetUserAnalysis(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	analysis, err := c.service.GetUserAnalysis(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
	})
}

// GetAlerts 获取系统告警
func (c *SuperAdminController) GetAlerts(ctx *gin.Context) {
	alerts, err := c.monitor.GetActiveAlerts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerts,
		"total":   len(alerts),
	})
}

// GetAdminLogs 获取管理员操作日志
func (c *SuperAdminController) GetAdminLogs(ctx *gin.Context) {
	limit := 100
	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	activities, err := c.service.GetRecentActivities(limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
		"total":   len(activities),
	})
}

// BroadcastMessage 广播系统消息
func (c *SuperAdminController) BroadcastMessage(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")

	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.BroadcastMessage(adminID, req.Message); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "广播消息已发送",
	})
}

func (c *SuperAdminController) GetChatList(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	chatType := ctx.Query("type")
	
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	chats, total, err := c.service.GetChatList(page, pageSize, chatType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    chats,
		"total":   total,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (c *SuperAdminController) GetMessageList(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	msgType := ctx.Query("type")
	sender := ctx.Query("sender")
	
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	messages, total, err := c.service.GetMessageList(page, pageSize, msgType, sender)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
		"total":   total,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
