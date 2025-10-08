package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// MessageAdvancedController 高级消息控制器
type MessageAdvancedController struct {
	messageAdvancedService *service.MessageAdvancedService
}

// NewMessageAdvancedController 创建高级消息控制器
func NewMessageAdvancedController(messageAdvancedService *service.MessageAdvancedService) *MessageAdvancedController {
	return &MessageAdvancedController{
		messageAdvancedService: messageAdvancedService,
	}
}

// RecallMessage 撤回消息
func (c *MessageAdvancedController) RecallMessage(ctx *gin.Context) {
	var req service.RecallMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if err := c.messageAdvancedService.RecallMessage(ctx, userID.(uint), &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息撤回成功"})
}

// EditMessage 编辑消息
func (c *MessageAdvancedController) EditMessage(ctx *gin.Context) {
	var req service.EditMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if err := c.messageAdvancedService.EditMessage(ctx, userID.(uint), &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息编辑成功"})
}

// ForwardMessage 转发消息
func (c *MessageAdvancedController) ForwardMessage(ctx *gin.Context) {
	var req service.ForwardMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if err := c.messageAdvancedService.ForwardMessage(ctx, userID.(uint), &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息转发成功"})
}

// ScheduleMessage 定时发送消息
func (c *MessageAdvancedController) ScheduleMessage(ctx *gin.Context) {
	var req service.ScheduleMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if err := c.messageAdvancedService.ScheduleMessage(ctx, userID.(uint), &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "定时消息设置成功"})
}

// SearchMessages 搜索消息
func (c *MessageAdvancedController) SearchMessages(ctx *gin.Context) {
	var req service.SearchMessagesRequest
	
	// 从查询参数获取搜索条件
	req.Query = ctx.Query("query")
	if req.Query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	// 聊天ID
	if chatIDStr := ctx.Query("chat_id"); chatIDStr != "" {
		if chatID, err := strconv.ParseUint(chatIDStr, 10, 32); err == nil {
			chatIDUint := uint(chatID)
			req.ChatID = &chatIDUint
		}
	}

	// 用户ID
	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDUint := uint(userID)
			req.UserID = &userIDUint
		}
	}

	// 消息类型
	req.MessageType = ctx.Query("message_type")

	// 分页参数
	if pageStr := ctx.DefaultQuery("page", "1"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}
	if pageSizeStr := ctx.DefaultQuery("page_size", "20"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	messages, total, err := c.messageAdvancedService.SearchMessages(ctx, userID.(uint), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": messages,
		"pagination": gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
			"total":     total,
			"pages":     (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		},
	})
}

// GetMessageEditHistory 获取消息编辑历史
func (c *MessageAdvancedController) GetMessageEditHistory(ctx *gin.Context) {
	messageIDStr := ctx.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的消息ID"})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	editHistory, err := c.messageAdvancedService.GetMessageEditHistory(ctx, uint(messageID), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": editHistory})
}

// CancelScheduledMessage 取消定时消息
func (c *MessageAdvancedController) CancelScheduledMessage(ctx *gin.Context) {
	messageIDStr := ctx.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的消息ID"})
		return
	}

	reason := ctx.Query("reason")

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if err := c.messageAdvancedService.CancelScheduledMessage(ctx, uint(messageID), userID.(uint), reason); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "定时消息取消成功"})
}

// GetScheduledMessages 获取用户的定时消息列表
func (c *MessageAdvancedController) GetScheduledMessages(ctx *gin.Context) {
	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 这里可以添加获取定时消息列表的逻辑
	ctx.JSON(http.StatusOK, gin.H{"message": "获取定时消息列表功能待实现"})
}

// ReplyToMessage 回复消息
func (c *MessageAdvancedController) ReplyToMessage(ctx *gin.Context) {
	var req struct {
		ReplyToID uint   `json:"reply_to_id" binding:"required"`
		Content   string `json:"content" binding:"required"`
		MessageType string `json:"message_type" binding:"required"`
		ChatID    *uint  `json:"chat_id,omitempty"`
		UserID    *uint  `json:"user_id,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 这里可以添加回复消息的逻辑
	ctx.JSON(http.StatusOK, gin.H{"message": "回复消息功能待实现"})
}
