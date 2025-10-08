package controller

import (
	"net/http"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
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

	// 调用服务层
	loginReq := service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	response, err := c.authService.Login(loginReq)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "登录失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
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
			"error":   "注册失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, response)
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
			"error":   "登出失败",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
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
