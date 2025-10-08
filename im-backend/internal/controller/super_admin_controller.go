package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// SuperAdminController 瓒呯骇绠＄悊鍛樻帶鍒跺櫒
type SuperAdminController struct {
	superAdminService *service.SuperAdminService
}

// NewSuperAdminController 鍒涘缓瓒呯骇绠＄悊鍛樻帶鍒跺櫒
func NewSuperAdminController() *SuperAdminController {
	return &SuperAdminController{
		superAdminService: service.NewSuperAdminService(),
	}
}

// GetSystemStats 鑾峰彇绯荤粺缁熻淇℃伅
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

// GetOnlineUsers 鑾峰彇鍦ㄧ嚎鐢ㄦ埛鍒楄〃
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

// GetUserActivity 鑾峰彇鐢ㄦ埛娲诲姩璁板綍
func (c *SuperAdminController) GetUserActivity(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勭敤鎴稩D"})
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

// GetUserBehaviorAnalysis 鑾峰彇鐢ㄦ埛琛屼负鍒嗘瀽
func (c *SuperAdminController) GetUserBehaviorAnalysis(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勭敤鎴稩D"})
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

// ForceLogoutUser 寮哄埗鐢ㄦ埛涓嬬嚎
func (c *SuperAdminController) ForceLogoutUser(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勭敤鎴稩D"})
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
		"message": "鐢ㄦ埛宸插己鍒朵笅绾?,
	})
}

// BanUser 灏佺鐢ㄦ埛
func (c *SuperAdminController) BanUser(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勭敤鎴稩D"})
		return
	}

	var req struct {
		Duration int    `json:"duration" binding:"required"` // 灏佺鏃堕暱锛堝皬鏃讹級
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
		"message": "鐢ㄦ埛宸插皝绂?,
	})
}

// UnbanUser 瑙ｅ皝鐢ㄦ埛
func (c *SuperAdminController) UnbanUser(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勭敤鎴稩D"})
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
		"message": "鐢ㄦ埛宸茶В灏?,
	})
}

// DeleteUserAccount 鍒犻櫎鐢ㄦ埛璐﹀彿
func (c *SuperAdminController) DeleteUserAccount(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勭敤鎴稩D"})
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
		"message": "鐢ㄦ埛璐﹀彿宸插垹闄?,
	})
}

// GetContentModerationQueue 鑾峰彇鍐呭瀹℃牳闃熷垪
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

// ModerateContent 瀹℃牳鍐呭
func (c *SuperAdminController) ModerateContent(ctx *gin.Context) {
	contentIDStr := ctx.Param("content_id")
	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫唴瀹笽D"})
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
		"message": "鍐呭瀹℃牳瀹屾垚",
	})
}

// GetSystemLogs 鑾峰彇绯荤粺鏃ュ織
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

// BroadcastMessage 骞挎挱绯荤粺娑堟伅
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
		"message": "绯荤粺娑堟伅宸插箍鎾?,
	})
}

// GetServerHealth 鑾峰彇鏈嶅姟鍣ㄥ仴搴风姸鎬?func (c *SuperAdminController) GetServerHealth(ctx *gin.Context) {
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

// SetupRoutes 璁剧疆璺敱
func (c *SuperAdminController) SetupRoutes(r *gin.RouterGroup) {
	// 绯荤粺缁熻
	r.GET("/stats", c.GetSystemStats)
	r.GET("/health", c.GetServerHealth)

	// 鍦ㄧ嚎鐢ㄦ埛绠＄悊
	r.GET("/online-users", c.GetOnlineUsers)

	// 鐢ㄦ埛绠＄悊
	users := r.Group("/users")
	{
		users.GET("/:user_id/activity", c.GetUserActivity)
		users.GET("/:user_id/analysis", c.GetUserBehaviorAnalysis)
		users.POST("/:user_id/force-logout", c.ForceLogoutUser)
		users.POST("/:user_id/ban", c.BanUser)
		users.POST("/:user_id/unban", c.UnbanUser)
		users.DELETE("/:user_id", c.DeleteUserAccount)
	}

	// 鍐呭瀹℃牳
	moderation := r.Group("/moderation")
	{
		moderation.GET("/queue", c.GetContentModerationQueue)
		moderation.POST("/:content_id/moderate", c.ModerateContent)
	}

	// 绯荤粺绠＄悊
	system := r.Group("/system")
	{
		system.GET("/logs", c.GetSystemLogs)
		system.POST("/broadcast", c.BroadcastMessage)
	}
}
