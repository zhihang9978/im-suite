package middleware

import (
	"net/http"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// Admin 管理员权限中间件（admin或super_admin）
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 查询用户角色
		var user model.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查是否为管理员或超级管理员
		if user.Role != "admin" && user.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "访问被拒绝",
				"message": "需要管理员权限才能访问此功能",
				"required_role": "admin或super_admin",
				"your_role": user.Role,
			})
			c.Abort()
			return
		}

		// 将用户角色存入上下文
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// RequireRole 要求特定角色权限
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		// 查询用户角色
		var user model.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查角色是否在允许列表中
		hasPermission := false
		for _, role := range roles {
			if user.Role == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "您的角色没有权限访问此功能",
				"required_roles": roles,
				"your_role": user.Role,
			})
			c.Abort()
			return
		}

		c.Set("user_role", user.Role)
		c.Next()
	}
}

