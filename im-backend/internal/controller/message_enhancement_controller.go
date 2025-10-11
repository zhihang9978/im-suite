package controller

import (
	"net/http"
	"strconv"
	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// MessageEnhancementController 消息功能增强控制器
type MessageEnhancementController struct {
	messageEnhancementService *service.MessageEnhancementService
}

// NewMessageEnhancementController 创建消息功能增强控制器实例
func NewMessageEnhancementController(messageEnhancementService *service.MessageEnhancementService) *MessageEnhancementController {
	return &MessageEnhancementController{
		messageEnhancementService: messageEnhancementService,
	}
}

// PinMessage 置顶消息
// @Summary 置顶消息
// @Description 将消息置顶显示
// @Tags 消息功能增强
// @Accept json
// @Produce json
// @Param request body service.PinMessageRequest true "置顶请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/pin [post]
func (c *MessageEnhancementController) PinMessage(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析请求
	var req service.PinMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置用户ID
	req.UserID = userID.(uint)

	// 执行置顶操作
	if err := c.messageEnhancementService.PinMessage(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息置顶成功"})
}

// UnpinMessage 取消置顶消息
// @Summary 取消置顶消息
// @Description 取消消息置顶
// @Tags 消息功能增强
// @Produce json
// @Param message_id path int true "消息ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/{message_id}/unpin [post]
func (c *MessageEnhancementController) UnpinMessage(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取消息ID
	messageIDStr := ctx.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "消息ID格式错误"})
		return
	}

	// 执行取消置顶操作
	if err := c.messageEnhancementService.UnpinMessage(uint(messageID), userID.(uint)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "取消置顶成功"})
}

// MarkMessage 标记消息
// @Summary 标记消息
// @Description 标记消息为重要、收藏或归档
// @Tags 消息功能增强
// @Accept json
// @Produce json
// @Param request body service.MarkMessageRequest true "标记请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/mark [post]
func (c *MessageEnhancementController) MarkMessage(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析请求
	var req service.MarkMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置用户ID
	req.UserID = userID.(uint)

	// 执行标记操作
	if err := c.messageEnhancementService.MarkMessage(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息标记成功"})
}

// UnmarkMessage 取消标记消息
// @Summary 取消标记消息
// @Description 取消消息标记
// @Tags 消息功能增强
// @Produce json
// @Param message_id path int true "消息ID"
// @Param mark_type query string true "标记类型"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/{message_id}/unmark [post]
func (c *MessageEnhancementController) UnmarkMessage(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取消息ID
	messageIDStr := ctx.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "消息ID格式错误"})
		return
	}

	// 获取标记类型
	markType := ctx.Query("mark_type")
	if markType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "标记类型不能为空"})
		return
	}

	// 执行取消标记操作
	if err := c.messageEnhancementService.UnmarkMessage(uint(messageID), userID.(uint), markType); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "取消标记成功"})
}

// ReplyToMessage 回复消息
// @Summary 回复消息
// @Description 回复指定消息
// @Tags 消息功能增强
// @Accept json
// @Produce json
// @Param request body service.ReplyMessageRequest true "回复请求"
// @Success 200 {object} model.Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/reply [post]
func (c *MessageEnhancementController) ReplyToMessage(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析请求
	var req service.ReplyMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置用户ID
	req.UserID = userID.(uint)

	// 执行回复操作
	message, err := c.messageEnhancementService.ReplyToMessage(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, message)
}

// ShareMessage 分享消息
// @Summary 分享消息
// @Description 分享消息给其他用户或群聊
// @Tags 消息功能增强
// @Accept json
// @Produce json
// @Param request body service.ShareMessageRequest true "分享请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/share [post]
func (c *MessageEnhancementController) ShareMessage(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析请求
	var req service.ShareMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置用户ID
	req.UserID = userID.(uint)

	// 执行分享操作
	if err := c.messageEnhancementService.ShareMessage(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息分享成功"})
}

// UpdateMessageStatus 更新消息状态
// @Summary 更新消息状态
// @Description 更新消息的发送、送达、已读状态
// @Tags 消息功能增强
// @Accept json
// @Produce json
// @Param request body service.UpdateMessageStatusRequest true "状态更新请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/status [post]
func (c *MessageEnhancementController) UpdateMessageStatus(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 解析请求
	var req service.UpdateMessageStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置用户ID
	req.UserID = userID.(uint)

	// 执行状态更新操作
	if err := c.messageEnhancementService.UpdateMessageStatus(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息状态更新成功"})
}

// GetMessageReplyChain 获取消息回复链
// @Summary 获取消息回复链
// @Description 获取消息的完整回复链
// @Tags 消息功能增强
// @Produce json
// @Param message_id path int true "消息ID"
// @Success 200 {array} model.Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/{message_id}/reply-chain [get]
func (c *MessageEnhancementController) GetMessageReplyChain(ctx *gin.Context) {
	// 获取消息ID
	messageIDStr := ctx.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "消息ID格式错误"})
		return
	}

	// 获取回复链
	messages, err := c.messageEnhancementService.GetMessageReplyChain(uint(messageID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

// GetPinnedMessages 获取置顶消息列表
// @Summary 获取置顶消息列表
// @Description 获取聊天中的置顶消息
// @Tags 消息功能增强
// @Produce json
// @Param chat_id query int true "聊天ID"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} model.Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/pinned [get]
func (c *MessageEnhancementController) GetPinnedMessages(ctx *gin.Context) {
	// 获取查询参数
	chatIDStr := ctx.Query("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "聊天ID格式错误"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	// 获取置顶消息
	messages, err := c.messageEnhancementService.GetPinnedMessages(uint(chatID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

// GetMarkedMessages 获取标记消息列表
// @Summary 获取标记消息列表
// @Description 获取用户的标记消息
// @Tags 消息功能增强
// @Produce json
// @Param mark_type query string true "标记类型"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} model.Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/marked [get]
func (c *MessageEnhancementController) GetMarkedMessages(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取标记类型
	markType := ctx.Query("mark_type")
	if markType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "标记类型不能为空"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	// 获取标记消息
	messages, err := c.messageEnhancementService.GetMarkedMessages(userID.(uint), markType, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

// GetMessageStatus 获取消息状态
// @Summary 获取消息状态
// @Description 获取消息的状态追踪信息
// @Tags 消息功能增强
// @Produce json
// @Param message_id path int true "消息ID"
// @Success 200 {array} model.MessageStatus
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/{message_id}/status [get]
func (c *MessageEnhancementController) GetMessageStatus(ctx *gin.Context) {
	// 获取消息ID
	messageIDStr := ctx.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "消息ID格式错误"})
		return
	}

	// 获取消息状态
	statuses, err := c.messageEnhancementService.GetMessageStatus(uint(messageID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, statuses)
}

// GetMessageShareHistory 获取消息分享历史
// @Summary 获取消息分享历史
// @Description 获取消息的分享历史记录
// @Tags 消息功能增强
// @Produce json
// @Param message_id path int true "消息ID"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} model.MessageShare
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/messages/{message_id}/share-history [get]
func (c *MessageEnhancementController) GetMessageShareHistory(ctx *gin.Context) {
	// 获取消息ID
	messageIDStr := ctx.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "消息ID格式错误"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	// 获取分享历史
	shares, err := c.messageEnhancementService.GetMessageShareHistory(uint(messageID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, shares)
}
