package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// PerformanceOptimizationController 性能优化控制器
type PerformanceOptimizationController struct {
	messagePushService     *service.MessagePushService
	largeGroupService      *service.LargeGroupService
	storageOptimizationService *service.StorageOptimizationService
	networkOptimizationService *service.NetworkOptimizationService
}

// NewPerformanceOptimizationController 创建性能优化控制器
func NewPerformanceOptimizationController() *PerformanceOptimizationController {
	return &PerformanceOptimizationController{
		messagePushService:         service.NewMessagePushService(),
		largeGroupService:          service.NewLargeGroupService(),
		storageOptimizationService: service.NewStorageOptimizationService(),
		networkOptimizationService: service.NewNetworkOptimizationService(),
	}
}

// GetPushStats 获取推送统计
func (c *PerformanceOptimizationController) GetPushStats(ctx *gin.Context) {
	stats, err := c.messagePushService.GetPushStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// QueuePush 队列推送任务
func (c *PerformanceOptimizationController) QueuePush(ctx *gin.Context) {
	var req struct {
		MessageID uint   `json:"message_id" binding:"required"`
		ChatID    uint   `json:"chat_id" binding:"required"`
		SenderID  uint   `json:"sender_id" binding:"required"`
		Content   string `json:"content" binding:"required"`
		Type      string `json:"type" binding:"required"`
		Priority  int    `json:"priority"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认优先级
	if req.Priority == 0 {
		req.Priority = 2 // 普通优先级
	}

	err := c.messagePushService.QueuePush(req.MessageID, req.ChatID, req.SenderID, req.Content, req.Type, req.Priority)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "推送任务已加入队列",
	})
}

// GetChatInfo 获取聊天信息（优化版本）
func (c *PerformanceOptimizationController) GetChatInfo(ctx *gin.Context) {
	chatIDStr := ctx.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天ID"})
		return
	}

	chat, err := c.largeGroupService.GetChatInfo(uint(chatID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    chat,
	})
}

// GetChatMembers 分页获取群成员
func (c *PerformanceOptimizationController) GetChatMembers(ctx *gin.Context) {
	chatIDStr := ctx.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天ID"})
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	members, total, err := c.largeGroupService.GetMembers(uint(chatID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"members":   members,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetChatMessages 分页获取消息
func (c *PerformanceOptimizationController) GetChatMessages(ctx *gin.Context) {
	chatIDStr := ctx.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天ID"})
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	messages, total, err := c.largeGroupService.GetMessages(uint(chatID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"messages":  messages,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetChatStatistics 获取聊天统计信息
func (c *PerformanceOptimizationController) GetChatStatistics(ctx *gin.Context) {
	chatIDStr := ctx.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天ID"})
		return
	}

	stats, err := c.largeGroupService.GetChatStatistics(uint(chatID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// InvalidateChatCache 使聊天缓存失效
func (c *PerformanceOptimizationController) InvalidateChatCache(ctx *gin.Context) {
	chatIDStr := ctx.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天ID"})
		return
	}

	c.largeGroupService.InvalidateCache(uint(chatID))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "缓存已清理",
	})
}

// OptimizeDatabase 优化数据库
func (c *PerformanceOptimizationController) OptimizeDatabase(ctx *gin.Context) {
	err := c.largeGroupService.OptimizeDatabase()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "数据库优化完成",
	})
}

// CleanupInactiveMembers 清理不活跃成员
func (c *PerformanceOptimizationController) CleanupInactiveMembers(ctx *gin.Context) {
	chatIDStr := ctx.Param("id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天ID"})
		return
	}

	daysStr := ctx.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 30
	}

	err = c.largeGroupService.CleanupInactiveMembers(uint(chatID), days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "不活跃成员清理完成",
	})
}

// CompressTable 压缩表数据
func (c *PerformanceOptimizationController) CompressTable(ctx *gin.Context) {
	tableName := ctx.Param("table")

	if tableName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "表名不能为空"})
		return
	}

	stats, err := c.storageOptimizationService.CompressTable(tableName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// CreatePartitions 创建分区
func (c *PerformanceOptimizationController) CreatePartitions(ctx *gin.Context) {
	tableName := ctx.Param("table")

	if tableName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "表名不能为空"})
		return
	}

	err := c.storageOptimizationService.CreatePartitions(tableName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "分区创建完成",
	})
}

// GetPartitionInfo 获取分区信息
func (c *PerformanceOptimizationController) GetPartitionInfo(ctx *gin.Context) {
	tableName := ctx.Param("table")

	if tableName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "表名不能为空"})
		return
	}

	partitions, err := c.storageOptimizationService.GetPartitionInfo(tableName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    partitions,
	})
}

// CleanupOldMessages 清理旧消息
func (c *PerformanceOptimizationController) CleanupOldMessages(ctx *gin.Context) {
	daysStr := ctx.DefaultQuery("days", "90")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 90
	}

	err = c.storageOptimizationService.CleanupOldMessages(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "旧消息清理任务已调度",
	})
}

// CleanupInactiveSessions 清理不活跃会话
func (c *PerformanceOptimizationController) CleanupInactiveSessions(ctx *gin.Context) {
	daysStr := ctx.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 30
	}

	err = c.storageOptimizationService.CleanupInactiveSessions(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "不活跃会话清理任务已调度",
	})
}

// CleanupOrphanedFiles 清理孤立文件
func (c *PerformanceOptimizationController) CleanupOrphanedFiles(ctx *gin.Context) {
	err := c.storageOptimizationService.CleanupOrphanedFiles()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "孤立文件清理任务已调度",
	})
}

// GetStorageStats 获取存储统计
func (c *PerformanceOptimizationController) GetStorageStats(ctx *gin.Context) {
	stats, err := c.storageOptimizationService.GetStorageStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetNetworkStats 获取网络统计
func (c *PerformanceOptimizationController) GetNetworkStats(ctx *gin.Context) {
	stats, err := c.networkOptimizationService.GetNetworkStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetOptimizationRecommendations 获取优化建议
func (c *PerformanceOptimizationController) GetOptimizationRecommendations(ctx *gin.Context) {
	recommendations, err := c.networkOptimizationService.GetOptimizationRecommendations()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recommendations,
	})
}

// SetupRoutes 设置路由
func (c *PerformanceOptimizationController) SetupRoutes(r *gin.RouterGroup) {
	// 消息推送优化
	push := r.Group("/push")
	{
		push.GET("/stats", c.GetPushStats)
		push.POST("/queue", c.QueuePush)
	}

	// 大群组性能优化
	groups := r.Group("/groups")
	{
		groups.GET("/:id", c.GetChatInfo)
		groups.GET("/:id/members", c.GetChatMembers)
		groups.GET("/:id/messages", c.GetChatMessages)
		groups.GET("/:id/statistics", c.GetChatStatistics)
		groups.DELETE("/:id/cache", c.InvalidateChatCache)
		groups.POST("/:id/cleanup-members", c.CleanupInactiveMembers)
	}

	// 存储优化
	storage := r.Group("/storage")
	{
		storage.GET("/stats", c.GetStorageStats)
		storage.POST("/compress/:table", c.CompressTable)
		storage.POST("/partitions/:table", c.CreatePartitions)
		storage.GET("/partitions/:table", c.GetPartitionInfo)
		storage.POST("/cleanup/messages", c.CleanupOldMessages)
		storage.POST("/cleanup/sessions", c.CleanupInactiveSessions)
		storage.POST("/cleanup/files", c.CleanupOrphanedFiles)
	}

	// 网络优化
	network := r.Group("/network")
	{
		network.GET("/stats", c.GetNetworkStats)
		network.GET("/recommendations", c.GetOptimizationRecommendations)
	}

	// 数据库优化
	r.POST("/database/optimize", c.OptimizeDatabase)
}
