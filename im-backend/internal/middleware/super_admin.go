package middleware

import (
	"net/http"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// SuperAdmin 超级管理员权限中间件
func SuperAdmin() gin.HandlerFunc {
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

		// 检查是否为超级管理员
		if user.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要超级管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
