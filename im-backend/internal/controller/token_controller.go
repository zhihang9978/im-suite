package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// TokenController Token控制器
type TokenController struct {
	authService         *service.AuthService
	tokenRefreshService *service.TokenRefreshService
}

// NewTokenController 创建Token控制器
func NewTokenController(authService *service.AuthService, tokenRefreshService *service.TokenRefreshService) *TokenController {
	return &TokenController{
		authService:         authService,
		tokenRefreshService: tokenRefreshService,
	}
}

// RefreshToken 刷新Token
// @Summary 刷新访问令牌
// @Description 使用Refresh Token获取新的Access Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} TokenResponse
// @Router /api/auth/refresh [post]
func (c *TokenController) RefreshToken(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证Refresh Token
	claims, err := c.tokenRefreshService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Refresh Token无效或已过期",
			"error":   err.Error(),
		})
		return
	}

	// 生成新的Access Token
	accessToken, err := c.authService.GenerateToken(claims.UserID, claims.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成Access Token失败",
			"error":   err.Error(),
		})
		return
	}

	// 生成新的Refresh Token（可选：延长有效期）
	newRefreshToken, err := c.tokenRefreshService.GenerateRefreshToken(claims.UserID, claims.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成Refresh Token失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token刷新成功",
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
			"token_type":    "Bearer",
			"expires_in":    86400, // 24小时
		},
	})
}

