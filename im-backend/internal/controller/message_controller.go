package controller

import (
	"net/http"
	"strconv"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// MessageController 消息控制器
type MessageController struct {
	messageService *service.MessageService
}

// NewMessageController 创建消息控制器
func NewMessageController(messageService *service.MessageService) *MessageController {
	return &MessageController{
		messageService: messageService,
	}
}

// SendMessage 发送消息
func (c *MessageController) SendMessage(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req service.SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	message, err := c.messageService.SendMessage(userID.(uint), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "发送消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    message,
	})
}

// GetMessages 获取消息列表
func (c *MessageController) GetMessages(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	// 解析查询参数
	var chatID, receiverID *uint
	if chatIDStr := ctx.Query("chat_id"); chatIDStr != "" {
		if id, err := strconv.ParseUint(chatIDStr, 10, 32); err == nil {
			cid := uint(id)
			chatID = &cid
		}
	}

	if receiverIDStr := ctx.Query("receiver_id"); receiverIDStr != "" {
		if id, err := strconv.ParseUint(receiverIDStr, 10, 32); err == nil {
			rid := uint(id)
			receiverID = &rid
		}
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	messages, total, err := c.messageService.GetMessages(userID.(uint), chatID, receiverID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetMessage 获取单条消息
func (c *MessageController) GetMessage(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	messageID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的消息ID",
		})
		return
	}

	message, err := c.messageService.GetMessage(uint(messageID), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "获取消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    message,
	})
}

// DeleteMessage 删除消息
func (c *MessageController) DeleteMessage(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	messageID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的消息ID",
		})
		return
	}

	if err := c.messageService.DeleteMessage(uint(messageID), userID.(uint)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "消息已删除",
	})
}

// MarkAsRead 标记消息为已读
func (c *MessageController) MarkAsRead(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	messageID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的消息ID",
		})
		return
	}

	if err := c.messageService.MarkAsRead(uint(messageID), userID.(uint)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "标记已读失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已标记为已读",
	})
}

// RecallMessage 撤回消息
func (c *MessageController) RecallMessage(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	messageID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的消息ID",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	ctx.ShouldBindJSON(&req)

	if err := c.messageService.RecallMessage(uint(messageID), userID.(uint), req.Reason); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "撤回消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "消息已撤回",
	})
}

// EditMessage 编辑消息
func (c *MessageController) EditMessage(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	messageID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的消息ID",
		})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	message, err := c.messageService.EditMessage(uint(messageID), userID.(uint), req.Content)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "编辑消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    message,
	})
}

// SearchMessages 搜索消息
func (c *MessageController) SearchMessages(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	keyword := ctx.Query("keyword")
	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "搜索关键词不能为空",
		})
		return
	}

	var chatID *uint
	if chatIDStr := ctx.Query("chat_id"); chatIDStr != "" {
		if id, err := strconv.ParseUint(chatIDStr, 10, 32); err == nil {
			cid := uint(id)
			chatID = &cid
		}
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	messages, total, err := c.messageService.SearchMessages(userID.(uint), keyword, chatID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "搜索消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
		"total":   total,
		"keyword": keyword,
	})
}

// ForwardMessage 转发消息
func (c *MessageController) ForwardMessage(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req struct {
		MessageID    uint `json:"message_id" binding:"required"`
		TargetChatID uint `json:"target_chat_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	message, err := c.messageService.ForwardMessage(req.MessageID, userID.(uint), req.TargetChatID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "转发消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    message,
	})
}

// GetUnreadCount 获取未读消息数
func (c *MessageController) GetUnreadCount(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var chatID *uint
	if chatIDStr := ctx.Query("chat_id"); chatIDStr != "" {
		if id, err := strconv.ParseUint(chatIDStr, 10, 32); err == nil {
			cid := uint(id)
			chatID = &cid
		}
	}

	count, err := c.messageService.GetUnreadCount(userID.(uint), chatID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取未读数失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   count,
	})
}

// SetTyping 设置正在输入状态（Telegram核心功能）
// 对应 Telegram MTProto: messages.setTyping
func (c *MessageController) SetTyping(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	var req struct {
		ChatID     *uint  `json:"chat_id"`     // 群聊ID（可选）
		ReceiverID *uint  `json:"receiver_id"` // 私聊对方ID（可选）
		Action     string `json:"action"`      // typing/uploading_photo/recording_voice等
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 默认action
	if req.Action == "" {
		req.Action = "typing"
	}

	// 必须指定chat_id或receiver_id之一
	if req.ChatID == nil && req.ReceiverID == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "必须指定chat_id或receiver_id",
		})
		return
	}

	// 通过WebSocket广播typing事件
	err := c.messageService.BroadcastTypingStatus(userID.(uint), req.ChatID, req.ReceiverID, req.Action)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "广播失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
