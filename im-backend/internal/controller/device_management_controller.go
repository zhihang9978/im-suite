package controller

import (
	"net/http"
	"strconv"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// DeviceManagementController 设备管理控制器
type DeviceManagementController struct {
	deviceService *service.DeviceManagementService
}

// NewDeviceManagementController 创建设备管理控制器实例
func NewDeviceManagementController() *DeviceManagementController {
	return &DeviceManagementController{
		deviceService: service.NewDeviceManagementService(),
	}
}

// RegisterDevice 注册设备
// @Summary 注册设备
// @Description 注册或更新用户设备信息
// @Tags Device
// @Accept json
// @Produce json
// @Param request body service.DeviceInfo true "设备信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/register [post]
func (c *DeviceManagementController) RegisterDevice(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var deviceInfo service.DeviceInfo
	if err := ctx.ShouldBindJSON(&deviceInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从请求中获取IP和User-Agent
	deviceInfo.IP = ctx.ClientIP()
	deviceInfo.UserAgent = ctx.GetHeader("User-Agent")

	session, err := c.deviceService.RegisterDevice(ctx.Request.Context(), userID.(uint), &deviceInfo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"device":  session,
	})
}

// GetUserDevices 获取用户设备列表
// @Summary 获取用户设备列表
// @Description 获取当前用户的所有活跃设备
// @Tags Device
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices [get]
func (c *DeviceManagementController) GetUserDevices(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	devices, err := c.deviceService.GetUserDevices(ctx.Request.Context(), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"devices": devices,
	})
}

// GetDeviceByID 获取设备详情
// @Summary 获取设备详情
// @Description 获取指定设备的详细信息
// @Tags Device
// @Produce json
// @Param device_id path string true "设备ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/:device_id [get]
func (c *DeviceManagementController) GetDeviceByID(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	deviceID := ctx.Param("device_id")
	if deviceID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备ID不能为空"})
		return
	}

	device, err := c.deviceService.GetDeviceByID(ctx.Request.Context(), userID.(uint), deviceID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"device":  device,
	})
}

// RevokeDevice 撤销设备
// @Summary 撤销设备
// @Description 撤销指定设备，强制其下线
// @Tags Device
// @Produce json
// @Param device_id path string true "设备ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/:device_id [delete]
func (c *DeviceManagementController) RevokeDevice(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	deviceID := ctx.Param("device_id")
	if deviceID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备ID不能为空"})
		return
	}

	if err := c.deviceService.RevokeDevice(ctx.Request.Context(), userID.(uint), deviceID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设备已撤销",
	})
}

// RevokeAllDevices 撤销所有设备
// @Summary 撤销所有设备
// @Description 撤销除当前设备外的所有设备
// @Tags Device
// @Produce json
// @Param except_device_id query string false "保留的设备ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/revoke-all [post]
func (c *DeviceManagementController) RevokeAllDevices(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	exceptDeviceID := ctx.Query("except_device_id")

	if err := c.deviceService.RevokeAllDevices(ctx.Request.Context(), userID.(uint), exceptDeviceID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "所有其他设备已撤销",
	})
}

// GetDeviceActivities 获取设备活动历史
// @Summary 获取设备活动历史
// @Description 获取设备的活动记录
// @Tags Device
// @Produce json
// @Param device_id query string false "设备ID（不提供则返回所有设备）"
// @Param limit query int false "限制数量"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/activities [get]
func (c *DeviceManagementController) GetDeviceActivities(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	deviceID := ctx.Query("device_id")
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	activities, err := c.deviceService.GetDeviceActivities(ctx.Request.Context(), userID.(uint), deviceID, limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"activities": activities,
	})
}

// GetSuspiciousDevices 获取可疑设备
// @Summary 获取可疑设备
// @Description 获取风险评分高的可疑设备
// @Tags Device
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/suspicious [get]
func (c *DeviceManagementController) GetSuspiciousDevices(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	devices, err := c.deviceService.GetSuspiciousDevices(ctx.Request.Context(), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"devices": devices,
	})
}

// GetDeviceStatistics 获取设备统计
// @Summary 获取设备统计
// @Description 获取设备的统计信息
// @Tags Device
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/statistics [get]
func (c *DeviceManagementController) GetDeviceStatistics(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	stats, err := c.deviceService.GetDeviceStatistics(ctx.Request.Context(), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// ExportDeviceData 导出设备数据
// @Summary 导出设备数据
// @Description 导出用户的所有设备数据（GDPR合规）
// @Tags Device
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/devices/export [get]
func (c *DeviceManagementController) ExportDeviceData(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	data, err := c.deviceService.ExportDeviceData(ctx.Request.Context(), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置下载头
	ctx.Header("Content-Type", "application/json")
	ctx.Header("Content-Disposition", "attachment; filename=device-data.json")
	ctx.String(http.StatusOK, data)
}
