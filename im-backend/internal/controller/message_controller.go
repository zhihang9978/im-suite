package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

type MessageController struct {
	messageService *service.MessageService
}

func NewMessageController(messageService *service.MessageService) *MessageController {
	return &MessageController{
		messageService: messageService,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChatID      uint   `json:"chat_id" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Content     string `json:"content"`
	MediaURL    string `json:"media_url"`
	ReplyToID   *uint  `json:"reply_to_id"`
	ForwardFromID *uint `json:"forward_from_id"`
}

// GetMessagesRequest 获取消息请求
type GetMessagesRequest struct {
	ChatID uint `form:"chat_id" binding:"required"`
	Limit  int  `form:"limit,default=50"`
	Offset int  `form:"offset,default=0"`
}

// SendMessage 发送消息
func (c *MessageController) SendMessage(ctx *gin.Context) {
	var req SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
		})
		return
	}

	// 调用服务层
	serviceReq := service.SendMessageRequest{
		ChatID:        req.ChatID,
		Type:          req.Type,
		Content:       req.Content,
		MediaURL:      req.MediaURL,
		ReplyToID:     req.ReplyToID,
		ForwardFromID: req.ForwardFromID,
	}

	message, err := c.messageService.SendMessage(userID.(uint), serviceReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "发送消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, message)
}

// GetMessages 获取消息列表
func (c *MessageController) GetMessages(ctx *gin.Context) {
	var req GetMessagesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
		})
		return
	}

	// 调用服务层
	serviceReq := service.GetMessagesRequest{
		ChatID: req.ChatID,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	messages, err := c.messageService.GetMessages(userID.(uint), serviceReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count": len(messages),
	})
}

// GetMessage 获取单条消息
func (c *MessageController) GetMessage(ctx *gin.Context) {
	messageIDStr := ctx.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的消息ID",
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
		})
		return
	}

	message, err := c.messageService.GetMessage(userID.(uint), uint(messageID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "消息不存在",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, message)
}

// DeleteMessage 删除消息
func (c *MessageController) DeleteMessage(ctx *gin.Context) {
	messageIDStr := ctx.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的消息ID",
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
		})
		return
	}

	err = c.messageService.DeleteMessage(userID.(uint), uint(messageID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除消息失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "消息删除成功",
	})
}

// MarkAsRead 标记消息为已读
func (c *MessageController) MarkAsRead(ctx *gin.Context) {
	messageIDStr := ctx.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的消息ID",
		})
		return
	}

	// 从上下文获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未认证",
		})
		return
	}

	err = c.messageService.MarkAsRead(userID.(uint), uint(messageID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "标记已读失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "消息已标记为已读",
	})
}
