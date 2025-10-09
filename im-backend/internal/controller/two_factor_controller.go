package controller

import (
	"net/http"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// TwoFactorController 双因子认证控制器
type TwoFactorController struct {
	twoFactorService *service.TwoFactorService
}

// NewTwoFactorController 创建双因子认证控制器实例
func NewTwoFactorController() *TwoFactorController {
	return &TwoFactorController{
		twoFactorService: service.NewTwoFactorService(),
	}
}

// Enable 启用双因子认证
// @Summary 启用双因子认证
// @Description 为用户账户启用双因子认证
// @Tags 2FA
// @Accept json
// @Produce json
// @Param request body service.EnableTwoFactorRequest true "启用请求"
// @Success 200 {object} service.EnableTwoFactorResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/2fa/enable [post]
func (c *TwoFactorController) Enable(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.EnableTwoFactorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.twoFactorService.EnableTwoFactor(ctx.Request.Context(), userID.(uint), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
		"message": "请使用验证器APP扫描二维码，并输入验证码完成启用",
	})
}

// Verify 验证并启用2FA
// @Summary 验证并启用2FA
// @Description 验证TOTP验证码并完成2FA启用
// @Tags 2FA
// @Accept json
// @Produce json
// @Param request body service.VerifyTwoFactorRequest true "验证请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/2fa/verify [post]
func (c *TwoFactorController) Verify(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.VerifyTwoFactorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.twoFactorService.VerifyAndEnableTwoFactor(ctx.Request.Context(), userID.(uint), &req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "双因子认证已成功启用",
	})
}

// Disable 禁用双因子认证
// @Summary 禁用双因子认证
// @Description 禁用用户账户的双因子认证
// @Tags 2FA
// @Accept json
// @Produce json
// @Param request body service.DisableTwoFactorRequest true "禁用请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/2fa/disable [post]
func (c *TwoFactorController) Disable(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.DisableTwoFactorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.twoFactorService.DisableTwoFactor(ctx.Request.Context(), userID.(uint), &req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "双因子认证已禁用",
	})
}

// GetStatus 获取2FA状态
// @Summary 获取2FA状态
// @Description 获取用户的双因子认证状态和统计信息
// @Tags 2FA
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/2fa/status [get]
func (c *TwoFactorController) GetStatus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	status, err := c.twoFactorService.GetTwoFactorStatus(ctx.Request.Context(), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// RegenerateBackupCodes 重新生成备用码
// @Summary 重新生成备用码
// @Description 重新生成2FA备用码
// @Tags 2FA
// @Accept json
// @Produce json
// @Param request body service.RegeneratBackupCodesRequest true "重新生成请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/2fa/backup-codes/regenerate [post]
func (c *TwoFactorController) RegenerateBackupCodes(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.RegeneratBackupCodesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	codes, err := c.twoFactorService.RegenerateBackupCodes(ctx.Request.Context(), userID.(uint), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":      true,
		"backup_codes": codes,
		"message":      "备用码已重新生成，请妥善保管",
	})
}

// GetTrustedDevices 获取受信任设备列表
// @Summary 获取受信任设备列表
// @Description 获取用户的所有受信任设备
// @Tags 2FA
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/2fa/trusted-devices [get]
func (c *TwoFactorController) GetTrustedDevices(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	devices, err := c.twoFactorService.GetTrustedDevices(ctx.Request.Context(), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"devices": devices,
	})
}

// RemoveTrustedDevice 移除受信任设备
// @Summary 移除受信任设备
// @Description 移除指定的受信任设备
// @Tags 2FA
// @Param device_id path string true "设备ID"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/2fa/trusted-devices/:device_id [delete]
func (c *TwoFactorController) RemoveTrustedDevice(ctx *gin.Context) {
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

	if err := c.twoFactorService.RemoveTrustedDevice(ctx.Request.Context(), userID.(uint), deviceID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设备已移除",
	})
}

// ValidateCode 验证2FA验证码（用于登录）
// @Summary 验证2FA验证码
// @Description 在登录时验证用户的2FA验证码
// @Tags 2FA
// @Accept json
// @Produce json
// @Param request body service.ValidateTwoFactorRequest true "验证请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/2fa/validate [post]
func (c *TwoFactorController) ValidateCode(ctx *gin.Context) {
	var req service.ValidateTwoFactorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.twoFactorService.ValidateTwoFactorCode(ctx.Request.Context(), req.UserID, req.Code); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证成功",
	})
}
