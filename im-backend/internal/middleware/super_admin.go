package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// SuperAdmin 超级管理员中间件
func SuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未认证",
			})
			c.Abort()
			return
		}

		// 查询用户信息
		var user model.User
		err := config.DB.First(&user, userID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "用户不存在",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "查询用户失败",
				})
			}
			c.Abort()
			return
		}

		// 检查是否为超级管理员
		if user.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "权限不足，需要超级管理员权限",
			})
			c.Abort()
			return
		}

		// 记录管理员操作
		c.Set("admin_id", user.ID)
		c.Set("admin_username", user.Username)

		c.Next()
	}
}
