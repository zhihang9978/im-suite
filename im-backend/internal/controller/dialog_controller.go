package controller

import (
	"net/http"
	"strconv"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// DialogController 会话控制器
type DialogController struct {
	dialogService *service.DialogService
}

// NewDialogController 创建会话控制器
func NewDialogController(dialogService *service.DialogService) *DialogController {
	return &DialogController{
		dialogService: dialogService,
	}
}

// GetDialogs 获取会话列表（Telegram首屏核心API）
// 对应 Telegram MTProto: messages.getDialogs
func (c *DialogController) GetDialogs(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	// 解析分页参数
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100 // 最大100条
	}

	// 获取会话列表
	response, err := c.dialogService.GetDialogs(userID.(uint), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取会话列表失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetDialogByPeer 获取与指定用户/群组的会话
func (c *DialogController) GetDialogByPeer(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	peerID, err := strconv.ParseUint(ctx.Param("peer_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的peer_id",
		})
		return
	}

	peerType := ctx.DefaultQuery("peer_type", "user") // user/group/channel

	dialog, err := c.dialogService.GetDialogByPeer(userID.(uint), uint(peerID), peerType)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "会话不存在",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dialog,
	})
}

// PinDialog 置顶会话
func (c *DialogController) PinDialog(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	peerID, err := strconv.ParseUint(ctx.Param("peer_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的peer_id",
		})
		return
	}

	var req struct {
		PeerType string `json:"peer_type" binding:"required"` // user/group
		Pinned   bool   `json:"pinned"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	err = c.dialogService.SetDialogPin(userID.(uint), uint(peerID), req.PeerType, req.Pinned)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "设置置顶失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置成功",
	})
}

// MuteDialog 静音会话
func (c *DialogController) MuteDialog(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	peerID, err := strconv.ParseUint(ctx.Param("peer_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的peer_id",
		})
		return
	}

	var req struct {
		PeerType  string `json:"peer_type" binding:"required"` // user/group
		Muted     bool   `json:"muted"`
		MuteUntil int64  `json:"mute_until"` // Unix时间戳，0=永久
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	err = c.dialogService.SetDialogMute(userID.(uint), uint(peerID), req.PeerType, req.Muted, req.MuteUntil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "设置静音失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置成功",
	})
}
