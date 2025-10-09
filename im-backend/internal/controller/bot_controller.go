package controller

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// BotController 机器人控制器
type BotController struct {
	botService *service.BotService
}

// NewBotController 创建机器人控制器实例
func NewBotController() *BotController {
	return &BotController{
		botService: service.NewBotService(),
	}
}

// CreateBot 创建机器人（需要super_admin权限）
func (c *BotController) CreateBot(ctx *gin.Context) {
	adminID := ctx.GetUint("user_id")
	
	var req service.CreateBotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	resp, err := c.botService.CreateBot(ctx.Request.Context(), adminID, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
		"message": "机器人创建成功，请妥善保管API密钥",
		"warning": "API密钥只显示一次，请立即保存",
	})
}

// GetBotList 获取机器人列表
func (c *BotController) GetBotList(ctx *gin.Context) {
	bots, err := c.botService.GetBotList(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bots,
		"total":   len(bots),
	})
}

// GetBotByID 获取机器人详情
func (c *BotController) GetBotByID(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}
	
	bot, err := c.botService.GetBotByID(ctx.Request.Context(), uint(botID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bot,
	})
}

// UpdateBotPermissions 更新机器人权限
func (c *BotController) UpdateBotPermissions(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}
	
	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := c.botService.UpdateBotPermissions(ctx.Request.Context(), uint(botID), req.Permissions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "权限已更新",
	})
}

// ToggleBotStatus 切换机器人状态
func (c *BotController) ToggleBotStatus(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}
	
	var req struct {
		IsActive bool `json:"is_active"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := c.botService.ToggleBotStatus(ctx.Request.Context(), uint(botID), req.IsActive); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "状态已更新",
	})
}

// DeleteBot 删除机器人
func (c *BotController) DeleteBot(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}
	
	if err := c.botService.DeleteBot(ctx.Request.Context(), uint(botID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "机器人已删除",
	})
}

// GetBotLogs 获取机器人日志
func (c *BotController) GetBotLogs(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}
	
	limitStr := ctx.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	
	logs, err := c.botService.GetBotLogs(ctx.Request.Context(), uint(botID), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"total":   len(logs),
	})
}

// GetBotStats 获取机器人统计
func (c *BotController) GetBotStats(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}
	
	stats, err := c.botService.GetBotStats(ctx.Request.Context(), uint(botID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// RegenerateAPISecret 重新生成API密钥
func (c *BotController) RegenerateAPISecret(ctx *gin.Context) {
	botID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的机器人ID"})
		return
	}
	
	apiSecret, err := c.botService.RegenerateAPISecret(ctx.Request.Context(), uint(botID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"api_secret": apiSecret,
		"message":    "API密钥已重新生成，请立即保存",
		"warning":    "旧密钥将立即失效",
	})
}

// ========================================
// 机器人API端点（使用API Key认证）
// ========================================

// BotCreateUser 机器人创建用户
func (c *BotController) BotCreateUser(ctx *gin.Context) {
	bot, _ := ctx.Get("bot")
	botModel := bot.(*service.BotModel)
	
	var req service.BotCreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	user, err := c.botService.BotCreateUser(ctx.Request.Context(), botModel.Bot, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    user,
		"message": "用户创建成功",
	})
}

// BotBanUser 机器人封禁用户
func (c *BotController) BotBanUser(ctx *gin.Context) {
	bot, _ := ctx.Get("bot")
	botModel := bot.(*service.BotModel)
	
	var req service.BotBanUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := c.botService.BotBanUser(ctx.Request.Context(), botModel.Bot, &req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已封禁",
	})
}

// BotUnbanUser 机器人解封用户
func (c *BotController) BotUnbanUser(ctx *gin.Context) {
	bot, _ := ctx.Get("bot")
	botModel := bot.(*service.BotModel)
	
	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}
	
	if err := c.botService.BotUnbanUser(ctx.Request.Context(), botModel.Bot, uint(userID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已解封",
	})
}

// BotDeleteUser 机器人删除用户
func (c *BotController) BotDeleteUser(ctx *gin.Context) {
	bot, _ := ctx.Get("bot")
	botModel := bot.(*service.BotModel)
	
	var req service.BotDeleteUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := c.botService.BotDeleteUser(ctx.Request.Context(), botModel.Bot, &req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户已删除",
	})
}

