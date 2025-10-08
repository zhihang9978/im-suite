package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// MessageEncryptionController 消息加密控制器
type MessageEncryptionController struct {
	messageEncryptionService *service.MessageEncryptionService
}

// NewMessageEncryptionController 创建消息加密控制器
func NewMessageEncryptionController(messageEncryptionService *service.MessageEncryptionService) *MessageEncryptionController {
	return &MessageEncryptionController{
		messageEncryptionService: messageEncryptionService,
	}
}

// EncryptMessage 加密消息
func (c *MessageEncryptionController) EncryptMessage(ctx *gin.Context) {
	var req service.EncryptMessageRequest
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

	if err := c.messageEncryptionService.EncryptMessage(ctx, userID.(uint), &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息加密成功"})
}

// DecryptMessage 解密消息
func (c *MessageEncryptionController) DecryptMessage(ctx *gin.Context) {
	var req service.DecryptMessageRequest
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

	content, err := c.messageEncryptionService.DecryptMessage(ctx, userID.(uint), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"content": content,
		"message": "消息解密成功",
	})
}

// GetEncryptedMessageInfo 获取加密消息信息
func (c *MessageEncryptionController) GetEncryptedMessageInfo(ctx *gin.Context) {
	messageIDStr := ctx.Param("message_id")
	var messageID uint
	if _, err := fmt.Sscanf(messageIDStr, "%d", &messageID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的消息ID"})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	info, err := c.messageEncryptionService.GetEncryptedMessageInfo(ctx, messageID, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": info})
}

// SetMessageSelfDestruct 设置消息自毁时间
func (c *MessageEncryptionController) SetMessageSelfDestruct(ctx *gin.Context) {
	var req struct {
		MessageID    uint `json:"message_id" binding:"required"`
		DestructTime int  `json:"destruct_time" binding:"required,min=1"` // 秒数
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

	if err := c.messageEncryptionService.SetMessageSelfDestruct(ctx, req.MessageID, userID.(uint), req.DestructTime); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息自毁时间设置成功"})
}
