package controller

import (
	"net/http"
	"strconv"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// BotUserController 机器人用户管理控制器
type BotUserController struct {
	botUserService *service.BotUserManagementService
}

// NewBotUserController 创建机器人用户管理控制器实例
func NewBotUserController() *BotUserController {
	return &BotUserController{
		botUserService: service.NewBotUserManagementService(),
	}
}

// CreateBotUser 创建机器人用户（super_admin）
func (c *BotUserController) CreateBotUser(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")

	var req service.CreateBotUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	botUser, err := c.botUserService.CreateBotUser(ctx.Request.Context(), adminID, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    botUser,
		"message": "机器人用户创建成功，现在可以在聊天中与机器人交互",
	})
}

// GrantPermission 授权用户使用机器人（admin/super_admin）
func (c *BotUserController) GrantPermission(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")

	var req service.GrantPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission, err := c.botUserService.GrantPermission(ctx.Request.Context(), adminID, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    permission,
		"message": "用户授权成功，现在可以使用机器人",
	})
}

// RevokePermission 撤销用户权限（admin/super_admin）
func (c *BotUserController) RevokePermission(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")

	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	botID, err := strconv.ParseUint(ctx.Param("bot_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}

	if err := c.botUserService.RevokePermission(ctx.Request.Context(), adminID, uint(userID), uint(botID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "权限已撤销",
	})
}

// GetBotUser 获取机器人用户信息
func (c *BotUserController) GetBotUser(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("bot_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}

	botUser, err := c.botUserService.GetBotUser(ctx.Request.Context(), uint(botID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    botUser,
	})
}

// GetUserPermissions 获取用户的机器人权限列表
func (c *BotUserController) GetUserPermissions(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	permissions, err := c.botUserService.GetUserPermissions(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    permissions,
		"total":   len(permissions),
	})
}

// GetBotPermissions 获取机器人的授权用户列表
func (c *BotUserController) GetBotPermissions(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("bot_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}

	permissions, err := c.botUserService.GetBotPermissions(ctx.Request.Context(), uint(botID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    permissions,
		"total":   len(permissions),
	})
}

// DeleteBotUser 删除机器人用户（super_admin）
func (c *BotUserController) DeleteBotUser(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")

	botID, err := strconv.ParseUint(ctx.Param("bot_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}

	if err := c.botUserService.DeleteBotUser(ctx.Request.Context(), adminID, uint(botID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "机器人用户已删除",
	})
}

