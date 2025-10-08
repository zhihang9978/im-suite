package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "认证令牌格式错误"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// 验证令牌
		authService := service.NewAuthService()
		user, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "认证令牌无效"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", user.ID)
		c.Set("user", user)

		c.Next()
	}
}
