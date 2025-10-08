package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// UserManagementController 用户管理控制器
type UserManagementController struct {
	userManagementService *service.UserManagementService
}

// NewUserManagementController 创建用户管理控制器
func NewUserManagementController(userManagementService *service.UserManagementService) *UserManagementController {
	return &UserManagementController{
		userManagementService: userManagementService,
	}
}

// AddToBlacklistRequest 添加到黑名单请求
type AddToBlacklistRequest struct {
	UserID        uint       `json:"user_id" binding:"required"`
	Reason        string     `json:"reason" binding:"required"`
	BlacklistType string     `json:"blacklist_type" binding:"required"`
	Duration      *int64     `json:"duration,omitempty"` // 秒数，nil表示永久
}

// AddToBlacklist 添加用户到黑名单
func (c *UserManagementController) AddToBlacklist(ctx *gin.Context) {
	var req AddToBlacklistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前管理员ID
	adminID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var expiresAt *time.Time
	if req.Duration != nil {
		expiry := time.Now().Add(time.Duration(*req.Duration) * time.Second)
		expiresAt = &expiry
	}

	if err := c.userManagementService.AddToBlacklist(ctx, req.UserID, req.Reason, req.BlacklistType, expiresAt, adminID.(uint)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "用户已添加到黑名单"})
}

// RemoveFromBlacklist 从黑名单移除用户
func (c *UserManagementController) RemoveFromBlacklist(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := c.userManagementService.RemoveFromBlacklist(ctx, uint(userID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "用户已从黑名单移除"})
}

// GetBlacklist 获取黑名单列表
func (c *UserManagementController) GetBlacklist(ctx *gin.Context) {
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

	entries, total, err := c.userManagementService.GetBlacklist(ctx, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": entries,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUserActivity 获取用户活动记录
func (c *UserManagementController) GetUserActivity(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

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

	activities, total, err := c.userManagementService.GetUserActivity(ctx, uint(userID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": activities,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// SetUserRestrictionRequest 设置用户限制请求
type SetUserRestrictionRequest struct {
	UserID           uint   `json:"user_id" binding:"required"`
	RestrictionType  string `json:"restriction_type" binding:"required"`
	LimitValue       int    `json:"limit_value" binding:"required,min=1"`
}

// SetUserRestriction 设置用户限制
func (c *UserManagementController) SetUserRestriction(ctx *gin.Context) {
	var req SetUserRestrictionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.userManagementService.SetUserRestriction(ctx, req.UserID, req.RestrictionType, req.LimitValue); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "用户限制设置成功"})
}

// GetUserRestrictions 获取用户限制列表
func (c *UserManagementController) GetUserRestrictions(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	restrictions, err := c.userManagementService.GetUserRestrictions(ctx, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": restrictions})
}

// BanUserRequest 封禁用户请求
type BanUserRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
	Duration *int64 `json:"duration,omitempty"` // 秒数，nil表示永久封禁
}

// BanUser 封禁用户
func (c *UserManagementController) BanUser(ctx *gin.Context) {
	var req BanUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前管理员ID
	adminID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var duration *time.Duration
	if req.Duration != nil {
		d := time.Duration(*req.Duration) * time.Second
		duration = &d
	}

	if err := c.userManagementService.BanUser(ctx, req.UserID, req.Reason, duration, adminID.(uint)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "用户封禁成功"})
}

// UnbanUser 解封用户
func (c *UserManagementController) UnbanUser(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := c.userManagementService.UnbanUser(ctx, uint(userID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "用户解封成功"})
}

// GetUserStats 获取用户统计信息
func (c *UserManagementController) GetUserStats(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	stats, err := c.userManagementService.GetUserStats(ctx, uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetSuspiciousUsers 获取可疑用户列表
func (c *UserManagementController) GetSuspiciousUsers(ctx *gin.Context) {
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

	users, err := c.userManagementService.GetSuspiciousUsers(ctx, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": users})
}

// CleanupExpiredBlacklist 清理过期的黑名单条目
func (c *UserManagementController) CleanupExpiredBlacklist(ctx *gin.Context) {
	if err := c.userManagementService.CleanupExpiredBlacklist(ctx); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "过期黑名单清理完成"})
}

// CheckUserRestriction 检查用户限制
func (c *UserManagementController) CheckUserRestriction(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	restrictionType := ctx.Query("restriction_type")
	if restrictionType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "限制类型不能为空"})
		return
	}

	allowed, err := c.userManagementService.CheckUserRestriction(ctx, uint(userID), restrictionType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"allowed": allowed,
		"user_id": userID,
		"restriction_type": restrictionType,
	})
}

// IncrementUserRestriction 增加用户限制使用量
func (c *UserManagementController) IncrementUserRestriction(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	restrictionType := ctx.Query("restriction_type")
	if restrictionType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "限制类型不能为空"})
		return
	}

	if err := c.userManagementService.IncrementUserRestriction(ctx, uint(userID), restrictionType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "用户限制使用量已增加"})
}
