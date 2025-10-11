package controller

import (
	"net/http"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService      *service.AuthService
	twoFactorService *service.TwoFactorService
	deviceService    *service.DeviceManagementService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService:      authService,
		twoFactorService: service.NewTwoFactorService(),
		deviceService:    service.NewDeviceManagementService(),
	}
}

// LoginRequest 登录请求（支持phone或username登录）
type LoginRequest struct {
	Phone    string `json:"phone"`    // 手机号（可选）
	Username string `json:"username"` // 用户名（可选）
	Password string `json:"password" binding:"required"` // 密码（必需）
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`    // 手机号（必需）
	Username string `json:"username"`                    // 用户名（可选，为空时自动生成）
	Password string `json:"password" binding:"required"` // 密码（必需）
	Nickname string `json:"nickname"`                    // 昵称（可选）
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 验证：必须提供phone或username之一
	if req.Phone == "" && req.Username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": "必须提供phone或username之一",
		})
		return
	}

	// 调用服务层（优先使用phone，fallback到username）
	loginReq := service.LoginRequest{
		Username: req.Username,
		Phone:    req.Phone,
		Password: req.Password,
	}

	response, err := c.authService.Login(loginReq)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "登录失败",
			"details": err.Error(),
		})
		return
	}

	// 统一响应格式（兼容E2E测试）
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":         response.AccessToken, // E2E测试期望的token字段
			"access_token":  response.AccessToken,
			"refresh_token": response.RefreshToken,
			"expires_in":    response.ExpiresIn,
			"user":          response.User,
		},
	})
}

// Register 用户注册
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 调用服务层
	registerReq := service.RegisterRequest{
		Phone:    req.Phone,
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	}

	response, err := c.authService.Register(registerReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "注册失败",
			"details": err.Error(),
		})
		return
	}

	// 统一响应格式
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"token":         response.AccessToken,
			"access_token":  response.AccessToken,
			"refresh_token": response.RefreshToken,
			"expires_in":    response.ExpiresIn,
			"user":          response.User,
		},
	})
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少认证令牌",
		})
		return
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := c.authService.Logout(token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "登出失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}

// ValidateToken 验证令牌
func (c *AuthController) ValidateToken(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "缺少认证令牌",
		})
		return
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user, err := c.authService.ValidateToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "令牌验证失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  user,
		"valid": true,
	})
}

// RefreshToken 刷新令牌
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少认证令牌",
		})
		return
	}

	// 移除Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	req := service.RefreshRequest{RefreshToken: token}
	newToken, err := c.authService.RefreshToken(req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "令牌刷新失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}

// LoginWith2FA 使用2FA验证码完成登录
func (c *AuthController) LoginWith2FA(ctx *gin.Context) {
	var req struct {
		UserID      uint              `json:"user_id" binding:"required"`
		Code        string            `json:"code" binding:"required"`
		DeviceID    string            `json:"device_id"`
		DeviceInfo  map[string]string `json:"device_info"`
		TrustDevice bool              `json:"trust_device"` // 是否信任此设备
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 补充设备信息
	if req.DeviceInfo == nil {
		req.DeviceInfo = make(map[string]string)
	}
	req.DeviceInfo["ip"] = ctx.ClientIP()
	req.DeviceInfo["user_agent"] = ctx.GetHeader("User-Agent")

	// 调用服务层完成2FA登录
	response, err := c.authService.LoginWith2FA(req.UserID, req.Code, req.DeviceID, req.DeviceInfo)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "2FA验证失败",
			"details": err.Error(),
		})
		return
	}

	// 如果用户选择信任此设备
	if req.TrustDevice && req.DeviceID != "" {
		c.twoFactorService.AddTrustedDevice(
			ctx.Request.Context(),
			req.UserID,
			req.DeviceID,
			req.DeviceInfo["device_name"],
			req.DeviceInfo["device_type"],
			ctx.ClientIP(),
		)
	}

	ctx.JSON(http.StatusOK, response)
}
