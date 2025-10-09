package controller

import (
	"net/http"
	"strconv"
	"time"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ScreenShareEnhancedController 屏幕共享增强控制器
type ScreenShareEnhancedController struct {
	screenShareService *service.ScreenShareEnhancedService
}

// NewScreenShareEnhancedController 创建屏幕共享增强控制器
func NewScreenShareEnhancedController(screenShareService *service.ScreenShareEnhancedService) *ScreenShareEnhancedController {
	return &ScreenShareEnhancedController{
		screenShareService: screenShareService,
	}
}

// GetSessionHistory 获取会话历史
// @Summary 获取屏幕共享历史
// @Description 获取用户的屏幕共享历史记录
// @Tags ScreenShare
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/history [get]
func (c *ScreenShareEnhancedController) GetSessionHistory(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	// 分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	sessions, total, err := c.screenShareService.GetSessionHistory(ctx, userID, pageSize, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sessions":    sessions,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUserStatistics 获取用户统计
// @Summary 获取屏幕共享统计
// @Description 获取用户的屏幕共享统计信息
// @Tags ScreenShare
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/statistics [get]
func (c *ScreenShareEnhancedController) GetUserStatistics(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	stats, err := c.screenShareService.GetUserStatistics(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSessionDetails 获取会话详情
// @Summary 获取会话详情
// @Description 获取屏幕共享会话的详细信息
// @Tags ScreenShare
// @Param session_id path int true "会话ID"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/sessions/:session_id [get]
func (c *ScreenShareEnhancedController) GetSessionDetails(ctx *gin.Context) {
	sessionID, err := strconv.ParseUint(ctx.Param("session_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	details, err := c.screenShareService.GetSessionDetails(ctx, uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    details,
	})
}

// StartRecording 开始录制
// @Summary 开始录制屏幕共享
// @Description 开始录制当前的屏幕共享
// @Tags ScreenShare
// @Accept json
// @Produce json
// @Param call_id path string true "通话ID"
// @Param request body object true "录制请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/:call_id/recording/start [post]
func (c *ScreenShareEnhancedController) StartRecording(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	callID := ctx.Param("call_id")

	var req struct {
		Format  string `json:"format"`  // webm, mp4
		Quality string `json:"quality"` // high, medium, low
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.Format == "" {
		req.Format = "webm"
	}
	if req.Quality == "" {
		req.Quality = "medium"
	}

	recording, err := c.screenShareService.StartRecording(ctx, callID, userID, req.Format, req.Quality)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recording,
		"message": "录制已开始",
	})
}

// EndRecording 结束录制
// @Summary 结束录制
// @Description 结束当前的屏幕共享录制
// @Tags ScreenShare
// @Accept json
// @Produce json
// @Param recording_id path int true "录制ID"
// @Param request body object true "录制结果"
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/recordings/:recording_id/end [post]
func (c *ScreenShareEnhancedController) EndRecording(ctx *gin.Context) {
	recordingID, err := strconv.ParseUint(ctx.Param("recording_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的录制ID"})
		return
	}

	var req struct {
		FilePath string `json:"file_path" binding:"required"`
		FileSize int64  `json:"file_size" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.screenShareService.EndRecording(ctx, uint(recordingID), req.FilePath, req.FileSize); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "录制已结束",
	})
}

// GetRecordings 获取录制列表
// @Summary 获取录制列表
// @Description 获取指定会话的所有录制
// @Tags ScreenShare
// @Param session_id path int true "会话ID"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/sessions/:session_id/recordings [get]
func (c *ScreenShareEnhancedController) GetRecordings(ctx *gin.Context) {
	sessionID, err := strconv.ParseUint(ctx.Param("session_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	recordings, err := c.screenShareService.GetRecordings(ctx, uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recordings,
	})
}

// ExportStatistics 导出统计数据
// @Summary 导出统计数据
// @Description 导出指定时间范围的统计数据
// @Tags ScreenShare
// @Produce json
// @Param start_time query string true "开始时间" format(date-time)
// @Param end_time query string true "结束时间" format(date-time)
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/export [get]
func (c *ScreenShareEnhancedController) ExportStatistics(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供起止时间"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始时间格式"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束时间格式"})
		return
	}

	jsonData, err := c.screenShareService.ExportStatistics(ctx, userID, startTime, endTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", "attachment; filename=screen_share_statistics.json")
	ctx.String(http.StatusOK, jsonData)
}

// CheckPermission 检查权限
// @Summary 检查屏幕共享权限
// @Description 检查用户是否有屏幕共享权限
// @Tags ScreenShare
// @Produce json
// @Param quality query string false "质量等级" default(medium)
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/check-permission [get]
func (c *ScreenShareEnhancedController) CheckPermission(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	quality := ctx.DefaultQuery("quality", "medium")

	err := c.screenShareService.CheckSharePermission(userID, quality)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"allowed": false,
				"message": err.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"allowed": true,
			"message": "您有屏幕共享权限",
		},
	})
}

// RecordQualityChange 记录质量变更
// @Summary 记录质量变更
// @Description 记录屏幕共享质量的变更
// @Tags ScreenShare
// @Accept json
// @Produce json
// @Param call_id path string true "通话ID"
// @Param request body object true "质量变更信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/screen-share/:call_id/quality-change [post]
func (c *ScreenShareEnhancedController) RecordQualityChange(ctx *gin.Context) {
	callID := ctx.Param("call_id")

	var req struct {
		FromQuality  string  `json:"from_quality" binding:"required"`
		ToQuality    string  `json:"to_quality" binding:"required"`
		Reason       string  `json:"reason" binding:"required"`
		NetworkSpeed float64 `json:"network_speed"`
		CPUUsage     float64 `json:"cpu_usage"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.screenShareService.RecordQualityChange(
		ctx,
		callID,
		req.FromQuality,
		req.ToQuality,
		req.Reason,
		req.NetworkSpeed,
		req.CPUUsage,
	); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "质量变更已记录",
	})
}

