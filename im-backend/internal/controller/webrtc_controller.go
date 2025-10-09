package controller

import (
	"net/http"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// WebRTCController WebRTC控制器
type WebRTCController struct {
	webrtcService *service.WebRTCService
}

// NewWebRTCController 创建WebRTC控制器
func NewWebRTCController(webrtcService *service.WebRTCService) *WebRTCController {
	return &WebRTCController{
		webrtcService: webrtcService,
	}
}

// CreateCall 创建通话
// @Summary 创建通话
// @Description 发起音频或视频通话
// @Tags WebRTC
// @Accept json
// @Produce json
// @Param request body object true "通话请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls [post]
func (c *WebRTCController) CreateCall(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req struct {
		CalleeID uint   `json:"callee_id" binding:"required"`
		Type     string `json:"type" binding:"required"` // audio, video
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证通话类型
	if req.Type != "audio" && req.Type != "video" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "通话类型必须是 audio 或 video"})
		return
	}

	session, err := c.webrtcService.CreateCall(userID, req.CalleeID, req.Type)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
		"message": "通话创建成功",
	})
}

// EndCall 结束通话
// @Summary 结束通话
// @Description 结束当前通话
// @Tags WebRTC
// @Param call_id path string true "通话ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/end [post]
func (c *WebRTCController) EndCall(ctx *gin.Context) {
	callID := ctx.Param("call_id")

	if err := c.webrtcService.EndCall(callID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "通话已结束",
	})
}

// GetCallStats 获取通话统计
// @Summary 获取通话统计
// @Description 获取通话的详细统计信息
// @Tags WebRTC
// @Param call_id path string true "通话ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/stats [get]
func (c *WebRTCController) GetCallStats(ctx *gin.Context) {
	callID := ctx.Param("call_id")

	stats, err := c.webrtcService.GetCallStats(callID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ToggleMute 切换静音
// @Summary 切换静音
// @Description 切换通话中的静音状态
// @Tags WebRTC
// @Param call_id path string true "通话ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/mute [post]
func (c *WebRTCController) ToggleMute(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	callID := ctx.Param("call_id")

	if err := c.webrtcService.ToggleMute(callID, userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "静音状态已切换",
	})
}

// ToggleVideo 切换视频
// @Summary 切换视频
// @Description 切换通话中的视频状态
// @Tags WebRTC
// @Param call_id path string true "通话ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/video [post]
func (c *WebRTCController) ToggleVideo(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	callID := ctx.Param("call_id")

	if err := c.webrtcService.ToggleVideo(callID, userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "视频状态已切换",
	})
}

// StartScreenShare 开始屏幕共享
// @Summary 开始屏幕共享
// @Description 在通话中开始共享屏幕
// @Tags WebRTC
// @Accept json
// @Produce json
// @Param call_id path string true "通话ID"
// @Param request body object true "屏幕共享请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/screen-share/start [post]
func (c *WebRTCController) StartScreenShare(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	callID := ctx.Param("call_id")

	var req struct {
		UserName  string `json:"user_name"`
		Quality   string `json:"quality"`    // high, medium, low
		WithAudio bool   `json:"with_audio"` // 是否包含音频
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.Quality == "" {
		req.Quality = "medium"
	}

	// 验证质量参数
	if req.Quality != "high" && req.Quality != "medium" && req.Quality != "low" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "质量参数必须是: high, medium, low"})
		return
	}

	if err := c.webrtcService.StartScreenShare(callID, userID, req.UserName, req.Quality, req.WithAudio); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "屏幕共享已开始",
	})
}

// StopScreenShare 停止屏幕共享
// @Summary 停止屏幕共享
// @Description 停止当前的屏幕共享
// @Tags WebRTC
// @Param call_id path string true "通话ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/screen-share/stop [post]
func (c *WebRTCController) StopScreenShare(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	callID := ctx.Param("call_id")

	if err := c.webrtcService.StopScreenShare(callID, userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "屏幕共享已停止",
	})
}

// GetScreenShareStatus 获取屏幕共享状态
// @Summary 获取屏幕共享状态
// @Description 获取当前通话的屏幕共享状态
// @Tags WebRTC
// @Param call_id path string true "通话ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/screen-share/status [get]
func (c *WebRTCController) GetScreenShareStatus(ctx *gin.Context) {
	callID := ctx.Param("call_id")

	status, err := c.webrtcService.GetScreenShareStatus(callID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if status == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": map[string]interface{}{
				"is_active": false,
				"message":   "当前没有屏幕共享",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// ChangeScreenShareQuality 更改屏幕共享质量
// @Summary 更改屏幕共享质量
// @Description 动态调整屏幕共享的质量
// @Tags WebRTC
// @Accept json
// @Produce json
// @Param call_id path string true "通话ID"
// @Param request body object true "质量请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/calls/:call_id/screen-share/quality [post]
func (c *WebRTCController) ChangeScreenShareQuality(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	callID := ctx.Param("call_id")

	var req struct {
		Quality string `json:"quality" binding:"required"` // high, medium, low
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.webrtcService.ChangeScreenShareQuality(callID, userID, req.Quality); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "屏幕共享质量已更改为: " + req.Quality,
	})
}


